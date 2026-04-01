package external

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectActionType(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("shell action", func(t *testing.T) {
		dir := filepath.Join(tmpDir, "shell-action")
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "action.yaml"), []byte("name: Shell"), 0644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "run.sh"), []byte("#!/bin/sh\necho hello"), 0644); err != nil {
			t.Fatal(err)
		}

		actionType, err := detectActionType(dir)
		if err != nil {
			t.Fatalf("detectActionType() error = %v", err)
		}
		if actionType != ShellAction {
			t.Errorf("expected ShellAction, got %v", actionType)
		}
	})

	t.Run("binary action", func(t *testing.T) {
		dir := filepath.Join(tmpDir, "binary-action")
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "action.yaml"), []byte("name: Binary"), 0644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "run.go"), []byte("package main"), 0644); err != nil {
			t.Fatal(err)
		}

		actionType, err := detectActionType(dir)
		if err != nil {
			t.Fatalf("detectActionType() error = %v", err)
		}
		if actionType != BinaryAction {
			t.Errorf("expected BinaryAction, got %v", actionType)
		}
	})

	t.Run("missing action files", func(t *testing.T) {
		dir := filepath.Join(tmpDir, "empty-action")
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "action.yaml"), []byte("name: Empty"), 0644); err != nil {
			t.Fatal(err)
		}

		_, err := detectActionType(dir)
		if err == nil {
			t.Error("expected error for action with neither run.sh nor run.go")
		}
	})

	t.Run("run.go takes priority over run.sh", func(t *testing.T) {
		dir := filepath.Join(tmpDir, "both-action")
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "run.go"), []byte("package main"), 0644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "run.sh"), []byte("#!/bin/sh"), 0644); err != nil {
			t.Fatal(err)
		}

		actionType, err := detectActionType(dir)
		if err != nil {
			t.Fatalf("detectActionType() error = %v", err)
		}
		if actionType != BinaryAction {
			t.Errorf("expected BinaryAction (run.go takes priority), got %v", actionType)
		}
	})
}
