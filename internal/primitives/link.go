package primitives

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/project-actions/runner/internal/actions"
)

// LinkAction creates a symlink (ln -sf semantics).
type LinkAction struct{}

func (a *LinkAction) Execute(ctx *actions.ExecutionContext, config map[string]interface{}) error {
	src, ok := config["src"]
	if !ok {
		return fmt.Errorf("link action requires a 'src' field")
	}
	dest, ok := config["dest"]
	if !ok {
		return fmt.Errorf("link action requires a 'dest' field")
	}

	srcStr := fmt.Sprint(src)
	destStr := fmt.Sprint(dest)

	// dest is resolved relative to WorkingDir if not absolute.
	// src is passed verbatim to os.Symlink: if relative, it is resolved from
	// dest's directory (the directory containing the symlink), not from WorkingDir.
	// This matches ln -sf semantics.
	if !filepath.IsAbs(destStr) {
		destStr = filepath.Join(ctx.WorkingDir, destStr)
	}

	// Remove existing file or symlink at dest (ln -f semantics)
	if fi, err := os.Lstat(destStr); err == nil {
		if fi.IsDir() {
			return fmt.Errorf("link: dest %s is an existing directory, cannot replace with symlink", destStr)
		}
		if err := os.Remove(destStr); err != nil {
			return fmt.Errorf("link: remove existing dest %s: %w", destStr, err)
		}
	}

	if err := os.Symlink(srcStr, destStr); err != nil {
		return fmt.Errorf("link %s -> %s: %w", destStr, srcStr, err)
	}
	return nil
}

func (a *LinkAction) Validate(config map[string]interface{}) error {
	if _, ok := config["src"]; !ok {
		return fmt.Errorf("link action requires a 'src' field")
	}
	if _, ok := config["dest"]; !ok {
		return fmt.Errorf("link action requires a 'dest' field")
	}
	return nil
}
