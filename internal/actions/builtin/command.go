package builtin

import (
	"fmt"

	"github.com/project-actions/runner/internal/actions"
)

// CommandAction executes another command
type CommandAction struct {
	commandExecutor CommandExecutor
}

// CommandExecutor is an interface for executing commands
type CommandExecutor interface {
	ExecuteCommand(commandName string, args []string) error
}

// NewCommandAction creates a new command action with a command executor
func NewCommandAction(executor CommandExecutor) *CommandAction {
	return &CommandAction{
		commandExecutor: executor,
	}
}

// Execute runs another command
func (a *CommandAction) Execute(ctx *actions.ExecutionContext, config map[string]interface{}) error {
	cmdName, ok := config["command"]
	if !ok {
		return fmt.Errorf("command action requires a command name")
	}

	cmdNameStr := fmt.Sprint(cmdName)

	ctx.Logger.Info("→ Calling command: %s", cmdNameStr)

	// Execute the command through the engine
	if a.commandExecutor == nil {
		return fmt.Errorf("command executor not available")
	}

	// Pass empty args - command action doesn't support passing options yet
	return a.commandExecutor.ExecuteCommand(cmdNameStr, []string{})
}

// Validate checks if the configuration is valid
func (a *CommandAction) Validate(config map[string]interface{}) error {
	if _, ok := config["command"]; !ok {
		return fmt.Errorf("command action requires a 'command' field")
	}
	return nil
}
