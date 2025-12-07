package builtin

import (
	"fmt"
	"strings"

	"github.com/project-actions/runner/internal/actions"
	"github.com/project-actions/runner/internal/docker"
)

// ComposeExecAction executes a command in a running container
type ComposeExecAction struct{}

// Execute runs a command inside a docker-compose service
func (a *ComposeExecAction) Execute(ctx *actions.ExecutionContext, config map[string]interface{}) error {
	// Check if docker-compose is installed
	if !docker.IsDockerComposeInstalled() {
		return fmt.Errorf("docker-compose is not installed or not in PATH")
	}

	// Check if compose file exists
	if !docker.HasComposeFile(ctx.WorkingDir) {
		return fmt.Errorf("no docker-compose.yml file found in %s", ctx.WorkingDir)
	}

	// Get service name (required)
	service, ok := config["service"]
	if !ok {
		return fmt.Errorf("service name is required")
	}
	serviceStr := fmt.Sprint(service)

	// Get command (required)
	command, ok := config["command"]
	if !ok {
		return fmt.Errorf("command is required")
	}

	// Convert command to string array
	var commandParts []string
	switch cmd := command.(type) {
	case string:
		// Parse string command into parts
		commandParts = strings.Fields(cmd)
	case []interface{}:
		// Convert interface array to string array
		for _, part := range cmd {
			commandParts = append(commandParts, fmt.Sprint(part))
		}
	default:
		return fmt.Errorf("command must be a string or array")
	}

	if len(commandParts) == 0 {
		return fmt.Errorf("command cannot be empty")
	}

	// Check for interactive mode (default: true for TTY commands)
	interactive := true
	if i, ok := config["interactive"]; ok {
		if iBool, ok := i.(bool); ok {
			interactive = iBool
		}
	}

	ctx.Logger.Info("Executing command in service '%s': %s", serviceStr, strings.Join(commandParts, " "))

	// Execute the command
	if err := docker.Exec(ctx.WorkingDir, serviceStr, commandParts, interactive); err != nil {
		return fmt.Errorf("failed to execute command in service '%s': %w", serviceStr, err)
	}

	return nil
}

// Validate checks if the configuration is valid
func (a *ComposeExecAction) Validate(config map[string]interface{}) error {
	// Check for required service field
	if _, ok := config["service"]; !ok {
		return fmt.Errorf("compose-exec action requires a 'service' field")
	}

	// Check for required command field
	if _, ok := config["command"]; !ok {
		return fmt.Errorf("compose-exec action requires a 'command' field")
	}

	// Optional: interactive (bool)
	if i, ok := config["interactive"]; ok {
		if _, ok := i.(bool); !ok {
			return fmt.Errorf("interactive must be a boolean value")
		}
	}

	return nil
}
