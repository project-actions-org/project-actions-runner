package builtin

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/project-actions/runner/internal/actions"
)

// RunAction executes a shell command
type RunAction struct{}

// Execute runs the shell command
func (a *RunAction) Execute(ctx *actions.ExecutionContext, config map[string]interface{}) error {
	cmdStr, ok := config["run"]
	if !ok {
		return fmt.Errorf("run action requires a command")
	}

	// Convert to string
	cmdString := fmt.Sprint(cmdStr)

	// Log the command start
	ctx.Logger.CommandStart(cmdString)

	// Create the command
	cmd := exec.Command("sh", "-c", cmdString)
	cmd.Dir = ctx.WorkingDir

	// If verbose, show output directly. Otherwise, capture it silently
	if ctx.Verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	// Note: We don't set Stdout/Stderr if not verbose, so output is discarded

	cmd.Stdin = os.Stdin

	// Set environment variables
	cmd.Env = os.Environ()
	for k, v := range ctx.Environment {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	// Execute the command
	return cmd.Run()
}

// Validate checks if the configuration is valid
func (a *RunAction) Validate(config map[string]interface{}) error {
	if _, ok := config["run"]; !ok {
		return fmt.Errorf("run action requires a 'run' field")
	}
	return nil
}
