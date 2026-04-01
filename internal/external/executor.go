package external

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ExecuteAction runs a resolved external action with the given parameters.
// withParams are the raw values from the step's "with:" block.
// It validates inputs, applies defaults, builds the environment, and
// streams stdin/stdout/stderr directly so interactive prompts work.
func ExecuteAction(resolved *ResolvedAction, projectRoot, commandName string, verbose bool, withParams map[string]interface{}) error {
	// Validate inputs and apply defaults
	resolvedInputs, err := resolved.ActionMeta.ValidateAndApplyDefaults(withParams)
	if err != nil {
		return err
	}

	// Build environment
	env := os.Environ()
	env = append(env, buildInputEnv(resolvedInputs)...)
	env = append(env, buildContextEnv(projectRoot, commandName, verbose)...)

	var cmd *exec.Cmd
	switch resolved.Type {
	case ShellAction:
		cmd = exec.Command("sh", resolved.ExecPath)
	case BinaryAction:
		cmd = exec.Command(resolved.ExecPath)
	default:
		return fmt.Errorf("unknown action type")
	}

	cmd.Dir = projectRoot
	cmd.Env = env
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// buildInputEnv converts a with-params map into INPUT_* environment variables.
func buildInputEnv(with map[string]interface{}) []string {
	var env []string
	for k, v := range with {
		env = append(env, fmt.Sprintf("%s=%v", inputEnvKey(k), v))
	}
	return env
}

// buildContextEnv returns the PROJECT_* context environment variables.
func buildContextEnv(projectRoot, commandName string, verbose bool) []string {
	env := []string{
		fmt.Sprintf("PROJECT_ROOT=%s", projectRoot),
		fmt.Sprintf("PROJECT_NAME=%s", filepath.Base(projectRoot)),
		fmt.Sprintf("PROJECT_COMMAND=%s", commandName),
	}
	if verbose {
		env = append(env, "PROJECT_VERBOSE=true")
	}
	return env
}

// inputEnvKey converts an input name to its INPUT_* env var name.
// "role-name" → "INPUT_ROLE_NAME"
func inputEnvKey(name string) string {
	upper := strings.ToUpper(name)
	return "INPUT_" + strings.ReplaceAll(upper, "-", "_")
}
