package primitives

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/project-actions/runner/internal/parser"
)

func TestLinkAction_Execute(t *testing.T) {
	t.Run("creates a symlink", func(t *testing.T) {
		ctx := makeCtx(t)
		src := filepath.Join(ctx.WorkingDir, "src.txt")
		_ = os.WriteFile(src, []byte("x"), 0644)
		err := (&LinkAction{}).Execute(ctx, map[string]interface{}{
			"src":  src,
			"dest": "link.txt",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		dest := filepath.Join(ctx.WorkingDir, "link.txt")
		info, err := os.Lstat(dest)
		if err != nil {
			t.Fatalf("expected link.txt to exist: %v", err)
		}
		if info.Mode()&os.ModeSymlink == 0 {
			t.Error("expected a symlink")
		}
	})

	t.Run("overwrites existing file at dest", func(t *testing.T) {
		ctx := makeCtx(t)
		src := filepath.Join(ctx.WorkingDir, "src.txt")
		_ = os.WriteFile(src, []byte("x"), 0644)
		dest := filepath.Join(ctx.WorkingDir, "link.txt")
		_ = os.WriteFile(dest, []byte("old"), 0644)
		err := (&LinkAction{}).Execute(ctx, map[string]interface{}{
			"src":  src,
			"dest": "link.txt",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("missing src returns error", func(t *testing.T) {
		ctx := makeCtx(t)
		err := (&LinkAction{}).Execute(ctx, map[string]interface{}{"dest": "link.txt"})
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("missing dest returns error", func(t *testing.T) {
		ctx := makeCtx(t)
		err := (&LinkAction{}).Execute(ctx, map[string]interface{}{"src": "src.txt"})
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestLinkStep_ParsedFromYAML(t *testing.T) {
	// Simulates: - link:\n    src: /from\n    dest: /to
	raw := map[string]interface{}{
		"link": map[string]interface{}{
			"src":  "/from",
			"dest": "/to",
		},
	}
	step, err := parser.ParseStep(raw)
	if err != nil {
		t.Fatalf("ParseStep error: %v", err)
	}
	if step.ActionName != "link" {
		t.Errorf("ActionName = %q, want %q", step.ActionName, "link")
	}
	if step.Config["src"] != "/from" {
		t.Errorf("config[src] = %v, want /from", step.Config["src"])
	}
	if step.Config["dest"] != "/to" {
		t.Errorf("config[dest] = %v, want /to", step.Config["dest"])
	}
}

func TestLinkAction_Validate(t *testing.T) {
	if err := (&LinkAction{}).Validate(map[string]interface{}{"src": "a", "dest": "b"}); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if err := (&LinkAction{}).Validate(map[string]interface{}{"src": "a"}); err == nil {
		t.Error("expected error for missing dest")
	}
	if err := (&LinkAction{}).Validate(map[string]interface{}{"dest": "b"}); err == nil {
		t.Error("expected error for missing src")
	}
}
