package builtin

import (
	"fmt"

	"github.com/project-actions/runner/internal/actions"
	"github.com/project-actions/runner/internal/docker"
)

// ComposeUpAction starts docker-compose services
type ComposeUpAction struct{}

// Execute starts the docker-compose services
func (a *ComposeUpAction) Execute(ctx *actions.ExecutionContext, config map[string]interface{}) error {
	// Check if docker-compose is installed
	if !docker.IsDockerComposeInstalled() {
		return fmt.Errorf("docker-compose is not installed or not in PATH")
	}

	// Check if Docker is running
	if !docker.IsDockerRunning() {
		return fmt.Errorf("Docker is not running. Please start Docker and try again")
	}

	// Check if compose file exists
	if !docker.HasComposeFile(ctx.WorkingDir) {
		return fmt.Errorf("no docker-compose.yml file found in %s", ctx.WorkingDir)
	}

	// Check for detached mode (default: true)
	detached := true
	if d, ok := config["detached"]; ok {
		if dBool, ok := d.(bool); ok {
			detached = dBool
		}
	}

	ctx.Logger.Info("Starting docker-compose services...")

	// Run docker-compose up
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

// Validate checks if the configuration is valid
func (a *ComposeUpAction) Validate(config map[string]interface{}) error {
	// No required fields for this action
	// Optional: detached (bool)
	if d, ok := config["detached"]; ok {
		if _, ok := d.(bool); !ok {
			return fmt.Errorf("detached must be a boolean value")
		}
	}

	return nil
}
