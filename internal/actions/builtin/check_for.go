package builtin

import (
	"fmt"
	"os/exec"

	"github.com/project-actions/runner/internal/actions"
)

// CheckForAction checks if a command/tool exists in PATH
type CheckForAction struct{}

// Execute checks if the specified command exists
func (a *CheckForAction) Execute(ctx *actions.ExecutionContext, config map[string]interface{}) error {
	tool, ok := config["check-for"]
	if !ok {
		return fmt.Errorf("check-for action requires a tool name")
	}

	toolStr := fmt.Sprint(tool)

	// Check if the tool exists in PATH
	_, err := exec.LookPath(toolStr)
	if err != nil {
		// Tool not found - check for custom error message
		if msg, ok := config["if-missing"]; ok {
			ctx.Logger.Error("%v", msg)
		}
		return fmt.Errorf("required tool not found: %s", toolStr)
	}

	ctx.Logger.Info("✓ Found: %s", toolStr)
	return nil
}

// Validate checks if the configuration is valid
func (a *CheckForAction) Validate(config map[string]interface{}) error {
	if _, ok := config["check-for"]; !ok {
		return fmt.Errorf("check-for action requires a 'check-for' field")
	}
	return nil
}
