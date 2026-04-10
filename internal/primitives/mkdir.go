package primitives

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/project-actions/runner/internal/actions"
)

// MkdirAction creates one or more directories (mkdir -p semantics).
type MkdirAction struct{}

func (a *MkdirAction) Execute(ctx *actions.ExecutionContext, config map[string]interface{}) error {
	val, ok := config["mkdir"]
	if !ok {
		return fmt.Errorf("mkdir action requires a path")
	}

	paths, err := toStringList(val)
	if err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}

	for _, p := range paths {
		if !filepath.IsAbs(p) {
			p = filepath.Join(ctx.WorkingDir, p)
		}
		if err := os.MkdirAll(p, 0755); err != nil {
			return fmt.Errorf("mkdir %s: %w", p, err)
		}
	}
	return nil
}

func (a *MkdirAction) Validate(config map[string]interface{}) error {
	if _, ok := config["mkdir"]; !ok {
		return fmt.Errorf("mkdir action requires a 'mkdir' field")
	}
	return nil
}

// toStringList converts a string or []interface{} to []string.
func toStringList(val interface{}) ([]string, error) {
	switch v := val.(type) {
	case string:
		return []string{v}, nil
	case []interface{}:
		result := make([]string, len(v))
		for i, item := range v {
			result[i] = fmt.Sprint(item)
		}
		return result, nil
	default:
		return nil, fmt.Errorf("expected string or list of strings, got %T", val)
	}
}
