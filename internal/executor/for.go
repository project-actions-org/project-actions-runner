package executor

import (
	"fmt"
	"path/filepath"

	"github.com/project-actions/runner/internal/actions"
	"github.com/project-actions/runner/internal/parser"
)

func (e *Engine) executeForLoop(loop *parser.ForLoop, ctx *actions.ExecutionContext) error {
	items, err := resolveForItems(loop, ctx)
	if err != nil {
		return err
	}
	if len(items) == 0 {
		ctx.Logger.Debug("for: loop resolved to zero items, skipping")
		return nil
	}

	for _, item := range items {
		iterCtx := *ctx // shallow copy — isolates LoopVars per iteration
		iterCtx.LoopVars = map[string]interface{}{loop.VarName: item}

		for i := range loop.Steps {
			if err := e.ExecuteStep(&loop.Steps[i], &iterCtx); err != nil {
				return err
			}
		}
	}
	return nil
}

// resolveForItems returns the iteration list from either a literal list or a glob.
func resolveForItems(loop *parser.ForLoop, ctx *actions.ExecutionContext) ([]interface{}, error) {
	if loop.Glob != "" {
		pattern := loop.Glob
		if !filepath.IsAbs(pattern) {
			pattern = filepath.Join(ctx.WorkingDir, pattern)
		}
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return nil, fmt.Errorf("for: invalid glob %q: %w", loop.Glob, err)
		}
		items := make([]interface{}, len(matches))
		for i, m := range matches {
			rel, err := filepath.Rel(ctx.WorkingDir, m)
			if err != nil {
				rel = m
			}
			items[i] = rel
		}
		return items, nil
	}
	// loop.Items may be nil for an empty list — range over nil is a no-op in Go.
	return loop.Items, nil
}
