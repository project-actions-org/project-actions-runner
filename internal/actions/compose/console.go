package compose

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/project-actions/runner/internal/actions"
	"github.com/project-actions/runner/internal/docker"
)

// ConsoleAction opens an interactive shell in a Docker Compose service.
// The command must declare context: inside-container:<service> so that
// ctx.ServiceName is set. Running from inside the container is not supported.
type ConsoleAction struct{}

func (a *ConsoleAction) Execute(ctx *actions.ExecutionContext, config map[string]interface{}) error {
	if ctx.ServiceName == "" {
		return fmt.Errorf("console action requires a service name — set 'context: inside-container:<service>' in your command")
	}

	if docker.IsInsideContainer() {
		return fmt.Errorf("console action must be run from outside the container")
	}

	if !docker.IsDockerComposeInstalled() {
		return fmt.Errorf("docker compose is not installed or not in PATH")
	}

	if !docker.HasComposeFile(ctx.WorkingDir) {
		return fmt.Errorf("no docker-compose.yml file found in %s", ctx.WorkingDir)
	}

	shell := "bash"
	if s, ok := config["shell"]; ok {
		shell = fmt.Sprint(s)
	}

	cmd := exec.Command("docker", "compose", "exec", "-it", ctx.ServiceName, shell)
	cmd.Dir = ctx.WorkingDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}

func (a *ConsoleAction) Validate(config map[string]interface{}) error {
	if s, ok := config["shell"]; ok {
		if shell := fmt.Sprint(s); shell == "" {
			return fmt.Errorf("shell must not be empty")
		}
	}
	return nil
}
