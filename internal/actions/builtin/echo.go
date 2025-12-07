package builtin

import (
	"fmt"

	"github.com/project-actions/runner/internal/actions"
)

// EchoAction prints a message to stdout
type EchoAction struct{}

// Execute runs the echo action
func (a *EchoAction) Execute(ctx *actions.ExecutionContext, config map[string]interface{}) error {
	msg, ok := config["echo"]
	if !ok {
		return fmt.Errorf("echo action requires a message")
	}

	// Convert to string
	msgStr := fmt.Sprint(msg)

	// Print the message
	fmt.Println(msgStr)

	return nil
}

// Validate checks if the configuration is valid
func (a *EchoAction) Validate(config map[string]interface{}) error {
	if _, ok := config["echo"]; !ok {
		return fmt.Errorf("echo action requires a 'echo' field")
	}
	return nil
}
