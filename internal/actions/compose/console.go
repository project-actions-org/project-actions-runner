package compose

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

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
	cmd.Stdin = os.Stdin

	// With -it, the shell's stdout and stderr travel through the PTY and reach
	// the user via cmd.Stdout. cmd.Stderr here captures only docker compose's
	// own diagnostic messages (e.g. "Error: executing ... exit status 127"),
	// which we suppress when the shell simply exited non-zero, but preserve for
	// genuine docker errors (container not running, daemon unreachable, etc.).
	var dockerStderr bytes.Buffer
	cmd.Stderr = &dockerStderr

	err := cmd.Run()
	if err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			// Suppress docker's "Error: executing ... exit status N" noise.
			// Only surface docker stderr that looks like a real problem.
			msg := strings.TrimSpace(dockerStderr.String())
			if msg != "" && !strings.Contains(msg, "exit status") {
				fmt.Fprintln(os.Stderr, msg)
			}
			return nil
		}
		// Non-exit error (e.g. docker binary not found): show stderr and propagate.
		if msg := strings.TrimSpace(dockerStderr.String()); msg != "" {
			fmt.Fprintln(os.Stderr, msg)
		}
		return err
	}
	return nil
}

func (a *ConsoleAction) Validate(config map[string]interface{}) error {
	if s, ok := config["shell"]; ok {
		if shell := fmt.Sprint(s); shell == "" {
			return fmt.Errorf("shell must not be empty")
		}
	}
	return nil
}
