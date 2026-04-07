package compose

import (
	"fmt"

	"github.com/project-actions/runner/internal/actions"
	"github.com/project-actions/runner/internal/docker"
)

// ComposeStopAction stops docker-compose services
type ComposeStopAction struct{}

func (a *ComposeStopAction) Execute(ctx *actions.ExecutionContext, config map[string]interface{}) error {
	if !docker.IsDockerComposeInstalled() {
		return fmt.Errorf("docker-compose is not installed or not in PATH")
	}
	if !docker.HasComposeFile(ctx.WorkingDir) {
		return fmt.Errorf("no docker-compose.yml file found in %s", ctx.WorkingDir)
	}
	ctx.Logger.Info("Stopping docker-compose services...")
	if err := docker.Stop(ctx.WorkingDir); err != nil {
		return fmt.Errorf("failed to stop docker-compose services: %w", err)
	}
	ctx.Logger.Info("Docker-compose services stopped")
	return nil
}

func (a *ComposeStopAction) Validate(config map[string]interface{}) error {
	return nil
}
