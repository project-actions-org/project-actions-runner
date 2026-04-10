package primitives

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/project-actions/runner/internal/actions"
	"github.com/project-actions/runner/internal/logger"
)

func makeCtx(t *testing.T) *actions.ExecutionContext {
	t.Helper()
	return &actions.ExecutionContext{
		WorkingDir: t.TempDir(),
		Logger:     logger.New(),
	}
}

func TestMkdirAction_Execute(t *testing.T) {
	t.Run("creates a single directory", func(t *testing.T) {
		ctx := makeCtx(t)
		err := (&MkdirAction{}).Execute(ctx, map[string]interface{}{"mkdir": "dist"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if _, err := os.Stat(filepath.Join(ctx.WorkingDir, "dist")); err != nil {
			t.Errorf("expected dist to exist: %v", err)
		}
	})

	t.Run("creates nested directories", func(t *testing.T) {
		ctx := makeCtx(t)
		err := (&MkdirAction{}).Execute(ctx, map[string]interface{}{"mkdir": "a/b/c"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if _, err := os.Stat(filepath.Join(ctx.WorkingDir, "a/b/c")); err != nil {
			t.Errorf("expected a/b/c to exist: %v", err)
		}
	})

	t.Run("creates multiple directories from list", func(t *testing.T) {
		ctx := makeCtx(t)
		err := (&MkdirAction{}).Execute(ctx, map[string]interface{}{
			"mkdir": []interface{}{"dist", "tmp"},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		for _, d := range []string{"dist", "tmp"} {
			if _, err := os.Stat(filepath.Join(ctx.WorkingDir, d)); err != nil {
				t.Errorf("expected %s to exist: %v", d, err)
			}
		}
	})

	t.Run("existing directory does not error", func(t *testing.T) {
		ctx := makeCtx(t)
		_ = os.MkdirAll(filepath.Join(ctx.WorkingDir, "dist"), 0755)
		err := (&MkdirAction{}).Execute(ctx, map[string]interface{}{"mkdir": "dist"})
		if err != nil {
			t.Errorf("unexpected error on existing dir: %v", err)
		}
	})

	t.Run("missing mkdir field returns error", func(t *testing.T) {
		ctx := makeCtx(t)
		err := (&MkdirAction{}).Execute(ctx, map[string]interface{}{})
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestMkdirAction_Validate(t *testing.T) {
	if err := (&MkdirAction{}).Validate(map[string]interface{}{"mkdir": "dist"}); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if err := (&MkdirAction{}).Validate(map[string]interface{}{}); err == nil {
		t.Error("expected error for missing field")
	}
}
