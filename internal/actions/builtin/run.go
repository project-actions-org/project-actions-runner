package builtin

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/project-actions/runner/internal/actions"
	"github.com/project-actions/runner/internal/docker"
)

// RunAction executes a shell command.
// It supports @$ interpolation (replaced by positional args) and context-aware
// routing: when ctx.ServiceName is set and the runner is outside the container,
// the command is forwarded through `docker-compose exec` instead of running locally.
type RunAction struct {
	isInsideContainer func() bool
}

func (a *RunAction) containerCheck() bool {
	if a.isInsideContainer != nil {
		return a.isInsideContainer()
	}
	return docker.IsInsideContainer()
}

func (a *RunAction) Execute(ctx *actions.ExecutionContext, config map[string]interface{}) error {
	cmdStr, ok := config["run"]
	if !ok {
		return fmt.Errorf("run action requires a command")
	}

	cmdString := fmt.Sprint(cmdStr)

	// Expand @$ → space-joined positional args
	if strings.Contains(cmdString, "@$") {
		if len(ctx.Args) == 0 {
			return fmt.Errorf("no arguments given — usage: ./project run <script> [args...]")
		}
		cmdString = strings.ReplaceAll(cmdString, "@$", strings.Join(ctx.Args, " "))
	}

	ctx.Logger.CommandStart(cmdString)

	var cmd *exec.Cmd
	if ctx.ServiceName != "" && !a.containerCheck() {
		// Outside the container: route through docker-compose exec
		cmd = exec.Command("docker-compose", "exec", ctx.ServiceName, "sh", "-c", cmdString)
	} else {
		// Inside the container (or no service specified): run locally
		cmd = exec.Command("sh", "-c", cmdString)
	}

	cmd.Dir = ctx.WorkingDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	cmd.Env = os.Environ()
	for k, v := range ctx.Environment {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	return cmd.Run()
}

// Validate checks if the configuration is valid
func (a *RunAction) Validate(config map[string]interface{}) error {
	if _, ok := config["run"]; !ok {
		return fmt.Errorf("run action requires a 'run' field")
	}
	return nil
}
