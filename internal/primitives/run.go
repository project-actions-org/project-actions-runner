package primitives

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/mattn/go-isatty"
	"github.com/project-actions/runner/internal/actions"
	"github.com/project-actions/runner/internal/docker"
)

// RunAction executes a shell command.
// When ctx.ServiceName is set and the runner is outside the container,
// the command is forwarded through `docker compose exec` instead of running locally.
// Argument interpolation (<args>, <args.N>, <args.length>) is handled upstream
// by the executor engine before Execute is called.
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

	var cmd *exec.Cmd
	if ctx.ServiceName != "" && !a.containerCheck() {
		if ctx.Verbose {
			ctx.Logger.Info("routing via docker compose exec %s: %s", ctx.ServiceName, cmdString)
		}
		args := []string{"compose", "exec"}
		if isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd()) {
			args = append(args, "-it")
		}
		args = append(args, ctx.ServiceName, "sh", "-c", cmdString)
		cmd = exec.Command("docker", args...)
	} else {
		if ctx.Verbose {
			ctx.Logger.Info("run: %s", cmdString)
		}
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
