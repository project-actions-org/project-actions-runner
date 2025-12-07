package builtin

import (
	"fmt"

	"github.com/project-actions/runner/internal/actions"
	"github.com/project-actions/runner/internal/docker"
)

// ComposeDownAction stops and removes docker-compose services
type ComposeDownAction struct{}

// Execute stops and removes the docker-compose services
func (a *ComposeDownAction) Execute(ctx *actions.ExecutionContext, config map[string]interface{}) error {
	// Check if docker-compose is installed
	if !docker.IsDockerComposeInstalled() {
		return fmt.Errorf("docker-compose is not installed or not in PATH")
	}

	// Check if compose file exists
	if !docker.HasComposeFile(ctx.WorkingDir) {
		return fmt.Errorf("no docker-compose.yml file found in %s", ctx.WorkingDir)
	}

	// Check for remove volumes option (default: false)
	removeVolumes := false
	if v, ok := config["volumes"]; ok {
		if vBool, ok := v.(bool); ok {
			removeVolumes = vBool
		}
	}

	ctx.Logger.Info("Stopping and removing docker-compose services...")

	// Run docker-compose down
	if err := docker.Down(ctx.WorkingDir, removeVolumes); err != nil {
		return fmt.Errorf("failed to stop and remove docker-compose services: %w", err)
	}

	if removeVolumes {
		ctx.Logger.Info("Docker-compose services stopped and removed (including volumes)")
	} else {
		ctx.Logger.Info("Docker-compose services stopped and removed")
	}

	return nil
}

// Validate checks if the configuration is valid
func (a *ComposeDownAction) Validate(config map[string]interface{}) error {
	// Optional: volumes (bool)
	if v, ok := config["volumes"]; ok {
		if _, ok := v.(bool); !ok {
			return fmt.Errorf("volumes must be a boolean value")
		}
	}

	return nil
}
