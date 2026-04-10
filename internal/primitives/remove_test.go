package primitives

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRemoveAction_Execute(t *testing.T) {
	t.Run("removes a file", func(t *testing.T) {
		ctx := makeCtx(t)
		f := filepath.Join(ctx.WorkingDir, "file.txt")
		_ = os.WriteFile(f, []byte("x"), 0644)
		err := (&RemoveAction{}).Execute(ctx, map[string]interface{}{"remove": "file.txt"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if _, err := os.Stat(f); !os.IsNotExist(err) {
			t.Error("expected file to be removed")
		}
	})

	t.Run("removes a directory recursively", func(t *testing.T) {
		ctx := makeCtx(t)
		d := filepath.Join(ctx.WorkingDir, "dist")
		_ = os.MkdirAll(filepath.Join(d, "sub"), 0755)
		err := (&RemoveAction{}).Execute(ctx, map[string]interface{}{"remove": "dist"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if _, err := os.Stat(d); !os.IsNotExist(err) {
			t.Error("expected dist to be removed")
		}
	})

	t.Run("removes files matching glob", func(t *testing.T) {
		ctx := makeCtx(t)
		_ = os.MkdirAll(filepath.Join(ctx.WorkingDir, "dist"), 0755)
		_ = os.WriteFile(filepath.Join(ctx.WorkingDir, "dist", "a.tmp"), []byte("x"), 0644)
		_ = os.WriteFile(filepath.Join(ctx.WorkingDir, "dist", "b.tmp"), []byte("x"), 0644)
		_ = os.WriteFile(filepath.Join(ctx.WorkingDir, "dist", "keep.txt"), []byte("x"), 0644)
		err := (&RemoveAction{}).Execute(ctx, map[string]interface{}{"remove": "dist/*.tmp"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if _, err := os.Stat(filepath.Join(ctx.WorkingDir, "dist", "a.tmp")); !os.IsNotExist(err) {
			t.Error("expected a.tmp to be removed")
		}
		if _, err := os.Stat(filepath.Join(ctx.WorkingDir, "dist", "keep.txt")); err != nil {
			t.Error("expected keep.txt to survive")
		}
	})

	t.Run("non-existent path does not error", func(t *testing.T) {
		ctx := makeCtx(t)
		err := (&RemoveAction{}).Execute(ctx, map[string]interface{}{"remove": "nonexistent"})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("missing remove field returns error", func(t *testing.T) {
		ctx := makeCtx(t)
		err := (&RemoveAction{}).Execute(ctx, map[string]interface{}{})
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestRemoveAction_Validate(t *testing.T) {
	if err := (&RemoveAction{}).Validate(map[string]interface{}{"remove": "dist"}); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if err := (&RemoveAction{}).Validate(map[string]interface{}{}); err == nil {
		t.Error("expected error for missing field")
	}
}
