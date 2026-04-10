package primitives

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/project-actions/runner/internal/actions"
)

// RemoveAction deletes files, directories, or glob-matched paths.
type RemoveAction struct{}

func (a *RemoveAction) Execute(ctx *actions.ExecutionContext, config map[string]interface{}) error {
	val, ok := config["remove"]
	if !ok {
		return fmt.Errorf("remove action requires a path")
	}

	patterns, err := toStringList(val)
	if err != nil {
		return fmt.Errorf("remove: %w", err)
	}

	for _, pattern := range patterns {
		if !filepath.IsAbs(pattern) {
			pattern = filepath.Join(ctx.WorkingDir, pattern)
		}

		if containsGlobChars(pattern) {
			matches, err := filepath.Glob(pattern)
			if err != nil {
				return fmt.Errorf("invalid glob %q: %w", pattern, err)
			}
			for _, m := range matches {
				if err := os.RemoveAll(m); err != nil {
					return fmt.Errorf("remove %s: %w", m, err)
				}
			}
		} else {
			if err := os.RemoveAll(pattern); err != nil {
				return fmt.Errorf("remove %s: %w", pattern, err)
			}
		}
	}
	return nil
}

func (a *RemoveAction) Validate(config map[string]interface{}) error {
	if _, ok := config["remove"]; !ok {
		return fmt.Errorf("remove action requires a 'remove' field")
	}
	return nil
}

func containsGlobChars(s string) bool {
	return strings.ContainsAny(s, "*?[")
}
