package builtin

import (
	"fmt"

	"github.com/project-actions/runner/internal/actions"
	"github.com/project-actions/runner/internal/docker"
)

// ComposeStopAction stops docker-compose services
type ComposeStopAction struct{}

// Execute stops the docker-compose services
func (a *ComposeStopAction) Execute(ctx *actions.ExecutionContext, config map[string]interface{}) error {
	// Check if docker-compose is installed
	if !docker.IsDockerComposeInstalled() {
		return fmt.Errorf("docker-compose is not installed or not in PATH")
	}

	// Check if compose file exists
	if !docker.HasComposeFile(ctx.WorkingDir) {
		return fmt.Errorf("no docker-compose.yml file found in %s", ctx.WorkingDir)
	}

	ctx.Logger.Info("Stopping docker-compose services...")

	// Run docker-compose stop
	if err := docker.Stop(ctx.WorkingDir); err != nil {
		return fmt.Errorf("failed to stop docker-compose services: %w", err)
	}

	ctx.Logger.Info("Docker-compose services stopped")

	return nil
}

// Validate checks if the configuration is valid
func (a *ComposeStopAction) Validate(config map[string]interface{}) error {
	// No required fields for this action
	return nil
}
