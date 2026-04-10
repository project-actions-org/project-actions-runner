package executor

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/project-actions/runner/internal/actions"
	"github.com/project-actions/runner/internal/actions/compose"
	"github.com/project-actions/runner/internal/config"
	"github.com/project-actions/runner/internal/external"
	"github.com/project-actions/runner/internal/logger"
	"github.com/project-actions/runner/internal/parser"
	"github.com/project-actions/runner/internal/primitives"
)

// Engine orchestrates command execution
type Engine struct {
	logger         *logger.Logger
	actionRegistry *actions.Registry
	config         *config.Config
}

// NewEngine creates a new execution engine
func NewEngine(cfg *config.Config, log *logger.Logger) *Engine {
	engine := &Engine{
		logger:         log,
		actionRegistry: actions.NewRegistry(log),
		config:         cfg,
	}

	// Register primitives
	engine.actionRegistry.Register("echo", &primitives.EchoAction{})
	engine.actionRegistry.Register("run", &primitives.RunAction{})
	engine.actionRegistry.Register("check-for", &primitives.CheckForAction{})
	engine.actionRegistry.Register("command", primitives.NewCommandAction(engine))
	engine.actionRegistry.Register("mkdir", &primitives.MkdirAction{})
	engine.actionRegistry.Register("remove", &primitives.RemoveAction{})
	engine.actionRegistry.Register("link", &primitives.LinkAction{})

	// Register docker-compose actions
	engine.actionRegistry.Register("compose-up", &compose.ComposeUpAction{})
	engine.actionRegistry.Register("compose-stop", &compose.ComposeStopAction{})
	engine.actionRegistry.Register("compose-down", &compose.ComposeDownAction{})
	engine.actionRegistry.Register("compose-exec", &compose.ComposeExecAction{})
	engine.actionRegistry.Register("console", &compose.ConsoleAction{})

	return engine
}

// ExecuteCommand loads and executes a command by name
func (e *Engine) ExecuteCommand(commandName string, args []string) error {
	// Find the command file
	cmdFile, err := e.config.FindCommandFile(commandName)
	if err != nil {
		// List available commands for helpful error message
		available, _ := e.config.ListCommands()
		if len(available) > 0 {
			names := make([]string, len(available))
			for i, entry := range available {
				names[i] = entry.Name
			}
			return fmt.Errorf("command '%s' not found\n\nAvailable commands: %s\n\nRun 'project' to see all available commands",
				commandName, strings.Join(names, ", "))
		}
		return fmt.Errorf("command '%s' not found: %w", commandName, err)
	}

	// Parse the command
	cmd, err := parser.ParseCommandFile(cmdFile, commandName)
	if err != nil {
		return fmt.Errorf("failed to parse command '%s': %w\n\nPlease check the YAML syntax in: %s", commandName, err, cmdFile)
	}

	// Parse options and positional args from args
	options, positionalArgs := parseOptions(args)

	// Check if verbose mode is enabled
	verbose := false
	if v, ok := options["verbose"]; ok && v == "true" {
		verbose = true
	}

	// Bind positional args to declared param names
	var namedArgs map[string]string
	if len(cmd.Params) > 0 {
		namedArgs = make(map[string]string, len(cmd.Params))
		for i, p := range cmd.Params {
			if i < len(positionalArgs) {
				namedArgs[p.Name] = positionalArgs[i]
			}
		}
	}

	// Create execution context
	ctx := &actions.ExecutionContext{
		WorkingDir:    e.config.ProjectRoot,
		Environment:   make(map[string]string),
		Options:       options,
		Args:          positionalArgs,
		NamedArgs:     namedArgs,
		ContainerMode: false,
		ServiceName:   "",
		Logger:        e.logger,
		Config:        e.config,
		Verbose:       verbose,
		Sources:       cmd.Sources,
		CommandName:   commandName,
	}

	// Check and handle context requirements
	contextMgr := NewContextManager(e)
	if err := contextMgr.CheckContext(cmd, ctx); err != nil {
		return fmt.Errorf("context check failed: %w", err)
	}

	// Copy environment variables
	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) == 2 {
			ctx.Environment[parts[0]] = parts[1]
		}
	}

	// Execute all steps
	for i, rawStep := range cmd.Steps {
		step, err := parser.ParseStep(rawStep)
		if err != nil {
			e.logger.Error("Failed to parse step %d: %v", i+1, err)
			return fmt.Errorf("failed to parse step %d in command '%s': %w", i+1, commandName, err)
		}

		if err := e.ExecuteStep(step, ctx); err != nil {
			return fmt.Errorf("command '%s' failed at step %d: %w", commandName, i+1, err)
		}
	}

	return nil
}

// ExecuteStep executes a single step
func (e *Engine) ExecuteStep(step *parser.Step, ctx *actions.ExecutionContext) error {
	// Handle conditionals
	if step.Conditional != nil {
		e.logger.Debug("Evaluating conditional: %s = %s", step.Conditional.Type, step.Conditional.Value)

		shouldExecute, err := e.evaluateConditional(step.Conditional, ctx)
		if err != nil {
			return fmt.Errorf("conditional evaluation failed: %w", err)
		}

		if shouldExecute {
			e.logger.Debug("Conditional %s evaluated to true, executing %d then step(s)",
				step.Conditional.Type, len(step.Conditional.ThenSteps))

			// Execute the 'then' steps
			for _, thenStep := range step.Conditional.ThenSteps {
				if err := e.ExecuteStep(&thenStep, ctx); err != nil {
					return err
				}
			}
		} else {
			e.logger.Debug("Conditional %s evaluated to false, skipping then steps", step.Conditional.Type)
		}
		return nil
	}

	// Handle for: loops
	if step.ForLoop != nil {
		return e.executeForLoop(step.ForLoop, ctx)
	}

	// Handle if: expressions
	if step.IfExpr != nil {
		return e.executeIfExpr(step.IfExpr, ctx)
	}

	// Interpolate <args> tokens in step config before dispatch
	var interpolatedConfig map[string]interface{}
	if step.Config != nil {
		var err error
		interpolatedConfig, err = interpolateConfig(step.Config, ctx.Args, ctx.LoopVars, ctx.NamedArgs)
		if err != nil {
			return fmt.Errorf("argument interpolation failed: %w", err)
		}
	}

	// Handle external actions (source/action-name format)
	if strings.Contains(step.ActionName, "/") {
		return e.executeExternalAction(step, ctx, interpolatedConfig)
	}

	// Get the action
	action, err := e.actionRegistry.Get(step.ActionName)
	if err != nil {
		return fmt.Errorf("action '%s' not found: %w", step.ActionName, err)
	}

	// Validate action configuration
	if err := action.Validate(interpolatedConfig); err != nil {
		return fmt.Errorf("invalid configuration for action '%s': %w", step.ActionName, err)
	}

	detail := stepDetail(step.ActionName, interpolatedConfig)
	e.logger.StepStart(step.ActionName, detail)

	// Execute the action
	if err := action.Execute(ctx, interpolatedConfig); err != nil {
		e.logger.StepFail(step.ActionName, err)
		return fmt.Errorf("action '%s' failed: %w", step.ActionName, err)
	}

	e.logger.StepSuccess(step.ActionName, detail)

	return nil
}

// evaluateConditional determines if a conditional should execute
func (e *Engine) evaluateConditional(cond *parser.Conditional, ctx *actions.ExecutionContext) (bool, error) {
	switch cond.Type {
	case "if-option":
		// Support pipe syntax for OR: "option1|option2"
		options := strings.Split(cond.Value, "|")
		for _, opt := range options {
			optName := strings.TrimSpace(opt)
			if _, exists := ctx.Options[optName]; exists {
				e.logger.Debug("Option '%s' found", optName)
				return true, nil
			}
		}
		e.logger.Debug("None of the options [%s] found", cond.Value)
		return false, nil

	case "if-no-option":
		// Inverse of if-option
		options := strings.Split(cond.Value, "|")
		for _, opt := range options {
			optName := strings.TrimSpace(opt)
			if _, exists := ctx.Options[optName]; exists {
				e.logger.Debug("Option '%s' found, if-no-option condition is false", optName)
				return false, nil
			}
		}
		e.logger.Debug("None of the options [%s] found, if-no-option condition is true", cond.Value)
		return true, nil

	case "if-missing":
		// Check if file/directory doesn't exist
		path := cond.Value
		// Make path relative to project root if not absolute
		if !strings.HasPrefix(path, "/") {
			path = fmt.Sprintf("%s/%s", ctx.WorkingDir, path)
		}
		_, err := os.Stat(path)
		isMissing := os.IsNotExist(err)
		if isMissing {
			e.logger.Debug("Path '%s' does not exist, if-missing condition is true", path)
		} else {
			e.logger.Debug("Path '%s' exists, if-missing condition is false", path)
		}
		return isMissing, nil

	case "if-fails":
		// TODO: Implement proper failure tracking
		// if-fails requires tracking whether previous steps failed
		// For now, this is a placeholder that always returns false
		// Future implementation should:
		// 1. Track step execution results in context
		// 2. Check if the named step/action failed
		// 3. Execute then steps if failure occurred
		e.logger.Debug("if-fails conditional not fully implemented yet")
		return false, nil

	default:
		return false, fmt.Errorf("unknown conditional type: %s", cond.Type)
	}
}

// executeExternalAction handles steps with "source/action-name" format.
func (e *Engine) executeExternalAction(step *parser.Step, ctx *actions.ExecutionContext, stepConfig map[string]interface{}) error {
	parts := strings.SplitN(step.ActionName, "/", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid external action format %q: expected source/action-name", step.ActionName)
	}
	sourceName, actionName := parts[0], parts[1]

	// Look up the source URL from the command's declared sources
	rawURL, ok := ctx.Sources[sourceName]
	if !ok {
		if len(ctx.Sources) == 0 {
			return fmt.Errorf("no sources declared in this command file — add a sources: block to use external actions")
		}
		return fmt.Errorf("unknown action source %q — add it to the sources: block in your command file", sourceName)
	}

	source, err := external.ParseSourceURL(rawURL)
	if err != nil {
		return fmt.Errorf("invalid source URL for %q: %w", sourceName, err)
	}
	source.Name = sourceName

	// Build fetcher and resolver
	fetcher := &external.Fetcher{
		ActionsDir: filepath.Join(e.config.RuntimeDir, "actions"),
	}
	resolver := &external.Resolver{Fetcher: fetcher}

	// Resolve (fetch if needed, detect type, download binary if needed)
	e.logger.StepStart(step.ActionName, "")
	resolved, err := resolver.Resolve(sourceName, actionName, source, e.config.ProjectRoot)
	if err != nil {
		e.logger.StepFail(step.ActionName, err)
		return fmt.Errorf("action %q failed to resolve: %w", step.ActionName, err)
	}

	// Extract the "with:" params from config
	var withParams map[string]interface{}
	if w, ok := stepConfig["with"]; ok {
		if wMap, ok := w.(map[string]interface{}); ok {
			withParams = wMap
		}
	}
	if withParams == nil {
		withParams = make(map[string]interface{})
	}

	// Execute the action
	if err := external.ExecuteAction(resolved, e.config.ProjectRoot, ctx.CommandName, ctx.Verbose, withParams); err != nil {
		e.logger.StepFail(step.ActionName, err)
		return fmt.Errorf("action %q failed: %w", step.ActionName, err)
	}

	e.logger.StepSuccess(step.ActionName, "")
	return nil
}

// stepDetail extracts a human-readable summary from an action's interpolated config.
func stepDetail(actionName string, config map[string]interface{}) string {
	switch actionName {
	case "run":
		cmd, _ := config["run"].(string)
		cmd = strings.TrimSpace(cmd)
		if cmd == "" {
			return ""
		}
		lines := strings.Split(cmd, "\n")
		var firstLine string
		for _, l := range lines {
			l = strings.TrimSpace(l)
			if l != "" {
				firstLine = l
				break
			}
		}
		if len(lines) > 1 {
			return firstLine + " ..."
		}
		return firstLine
	case "mkdir":
		return configValueSummary(config["mkdir"])
	case "remove":
		return configValueSummary(config["remove"])
	case "link":
		src, _ := config["src"].(string)
		dest, _ := config["dest"].(string)
		if src != "" && dest != "" {
			return dest + " -> " + src
		}
	case "echo":
		v, _ := config["echo"].(string)
		return v
	}
	return ""
}

// configValueSummary returns a short string representation of a string or string-list config value.
func configValueSummary(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case []interface{}:
		parts := make([]string, 0, len(val))
		for _, item := range val {
			parts = append(parts, fmt.Sprint(item))
		}
		if len(parts) > 3 {
			return strings.Join(parts[:3], ", ") + ", ..."
		}
		return strings.Join(parts, ", ")
	}
	return ""
}

// parseOptions extracts --flag options and positional arguments from command-line args.
// Flags stop at the first non-"--" token or at an explicit "--" separator.
func parseOptions(args []string) (map[string]string, []string) {
	options := make(map[string]string)
	var positional []string

	i := 0
	for i < len(args) {
		arg := args[i]
		if arg == "--" {
			// Everything after -- is positional
			positional = args[i+1:]
			break
		}
		if !strings.HasPrefix(arg, "--") {
			// First non-flag token starts positional args
			positional = args[i:]
			break
		}
		// Parse --key or --key=value
		opt := strings.TrimPrefix(arg, "--")
		if strings.Contains(opt, "=") {
			parts := strings.SplitN(opt, "=", 2)
			options[parts[0]] = parts[1]
		} else {
			options[opt] = "true"
		}
		i++
	}

	if positional == nil {
		positional = []string{}
	}

	return options, positional
}
