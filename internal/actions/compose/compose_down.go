package compose

import (
	"fmt"

	"github.com/project-actions/runner/internal/actions"
	"github.com/project-actions/runner/internal/docker"
)

// ComposeDownAction stops and removes docker-compose services
type ComposeDownAction struct{}

func (a *ComposeDownAction) Execute(ctx *actions.ExecutionContext, config map[string]interface{}) error {
	if !docker.IsDockerComposeInstalled() {
		return fmt.Errorf("docker-compose is not installed or not in PATH")
	}
	if !docker.HasComposeFile(ctx.WorkingDir) {
		return fmt.Errorf("no docker-compose.yml file found in %s", ctx.WorkingDir)
	}
	removeVolumes := false
	if v, ok := config["volumes"]; ok {
		if vBool, ok := v.(bool); ok {
			removeVolumes = vBool
		}
	}
	ctx.Logger.Info("Stopping and removing docker-compose services...")
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

func (a *ComposeDownAction) Validate(config map[string]interface{}) error {
	if v, ok := config["volumes"]; ok {
		if _, ok := v.(bool); !ok {
			return fmt.Errorf("volumes must be a boolean value")
		}
	}
	return nil
}
