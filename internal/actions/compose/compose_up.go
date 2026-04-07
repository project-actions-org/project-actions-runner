package compose

import (
	"fmt"

	"github.com/project-actions/runner/internal/actions"
	"github.com/project-actions/runner/internal/docker"
)

// ComposeUpAction starts docker-compose services
type ComposeUpAction struct{}

func (a *ComposeUpAction) Execute(ctx *actions.ExecutionContext, config map[string]interface{}) error {
	if !docker.IsDockerComposeInstalled() {
		return fmt.Errorf("docker-compose is not installed or not in PATH")
	}
	if !docker.IsDockerRunning() {
		return fmt.Errorf("Docker is not running. Please start Docker and try again")
	}
	if !docker.HasComposeFile(ctx.WorkingDir) {
		return fmt.Errorf("no docker-compose.yml file found in %s", ctx.WorkingDir)
	}
	detached := true
	if d, ok := config["detached"]; ok {
		if dBool, ok := d.(bool); ok {
			detached = dBool
		}
	}
	ctx.Logger.Info("Starting docker-compose services...")
	if err := docker.Up(ctx.WorkingDir, detached); err != nil {
		return fmt.Errorf("failed to start docker-compose services: %w", err)
	}
	if detached {
		ctx.Logger.Info("Docker-compose services started in detached mode")
	} else {
		ctx.Logger.Info("Docker-compose services started")
	}
	return nil
}

func (a *ComposeUpAction) Validate(config map[string]interface{}) error {
	if d, ok := config["detached"]; ok {
		if _, ok := d.(bool); !ok {
			return fmt.Errorf("detached must be a boolean value")
		}
	}
	return nil
}
