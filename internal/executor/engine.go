package executor

import (
	"fmt"
	"os"
	"strings"

	"github.com/project-actions/runner/internal/actions"
	"github.com/project-actions/runner/internal/actions/builtin"
	"github.com/project-actions/runner/internal/config"
	"github.com/project-actions/runner/internal/logger"
	"github.com/project-actions/runner/internal/parser"
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

	// Register built-in actions
	engine.actionRegistry.Register("echo", &builtin.EchoAction{})
	engine.actionRegistry.Register("run", &builtin.RunAction{})
	engine.actionRegistry.Register("check-for", &builtin.CheckForAction{})
	// Pass the engine itself so command action can execute other commands
	engine.actionRegistry.Register("command", builtin.NewCommandAction(engine))

	// Register docker-compose actions
	engine.actionRegistry.Register("compose-up", &builtin.ComposeUpAction{})
	engine.actionRegistry.Register("compose-stop", &builtin.ComposeStopAction{})
	engine.actionRegistry.Register("compose-down", &builtin.ComposeDownAction{})
	engine.actionRegistry.Register("compose-exec", &builtin.ComposeExecAction{})

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
			return fmt.Errorf("command '%s' not found\n\nAvailable commands: %s\n\nRun 'project' to see all available commands",
				commandName, strings.Join(available, ", "))
		}
		return fmt.Errorf("command '%s' not found: %w", commandName, err)
	}

	// Parse the command
	cmd, err := parser.ParseCommandFile(cmdFile, commandName)
	if err != nil {
		return fmt.Errorf("failed to parse command '%s': %w\n\nPlease check the YAML syntax in: %s", commandName, err, cmdFile)
	}

	// Parse options from args
	options := parseOptions(args)

	// Check if verbose mode is enabled
	verbose := false
	if v, ok := options["verbose"]; ok && v == "true" {
		verbose = true
	}

	// Create execution context
	ctx := &actions.ExecutionContext{
		WorkingDir:    e.config.ProjectRoot,
		Environment:   make(map[string]string),
		Options:       options,
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

	// Get the action
	action, err := e.actionRegistry.Get(step.ActionName)
	if err != nil {
		return fmt.Errorf("action '%s' not found: %w", step.ActionName, err)
	}

	// Validate action configuration
	if err := action.Validate(step.Config); err != nil {
		return fmt.Errorf("invalid configuration for action '%s': %w", step.ActionName, err)
	}

	// Log step start
	e.logger.StepStart(step.ActionName)

	// Execute the action
	if err := action.Execute(ctx, step.Config); err != nil {
		e.logger.StepFail(step.ActionName, err)
		return fmt.Errorf("action '%s' failed: %w", step.ActionName, err)
	}

	// Log step success
	e.logger.StepSuccess(step.ActionName)

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

// parseOptions extracts options from command-line arguments
func parseOptions(args []string) map[string]string {
	options := make(map[string]string)

	for _, arg := range args {
		if strings.HasPrefix(arg, "--") {
			// Strip the -- prefix
			opt := strings.TrimPrefix(arg, "--")

			// Check for =value syntax
			if strings.Contains(opt, "=") {
				parts := strings.SplitN(opt, "=", 2)
				options[parts[0]] = parts[1]
			} else {
				// Boolean flag
				options[opt] = "true"
			}
		}
	}

	return options
}
