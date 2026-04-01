package integration

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/project-actions/runner/internal/config"
	"github.com/project-actions/runner/internal/executor"
	"github.com/project-actions/runner/internal/logger"
)

// TestExternalShellAction tests the full pipeline with a local shell action.
// It simulates an action source by creating the expected directory structure
// in the .project/.runtime/actions/ cache (bypassing git clone).
func TestExternalShellAction(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .project/commands directory
	commandsDir := filepath.Join(tmpDir, ".project", "commands")
	if err := os.MkdirAll(commandsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Pre-populate the actions cache to avoid real git clone in tests
	actionDir := filepath.Join(tmpDir, ".project", ".runtime", "actions", "test-src", "greet")
	if err := os.MkdirAll(actionDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Write action.yaml
	actionYAML := `name: Greeter
description: Says hello
inputs:
  name:
    description: Who to greet
    required: true
  greeting:
    description: The greeting word
    required: false
    default: Hello
`
	if err := os.WriteFile(filepath.Join(actionDir, "action.yaml"), []byte(actionYAML), 0644); err != nil {
		t.Fatal(err)
	}

	// Write run.sh — writes to a file so we can verify execution
	outputFile := filepath.Join(tmpDir, "output.txt")
	script := fmt.Sprintf("#!/bin/sh\necho \"${INPUT_GREETING}, ${INPUT_NAME}!\" > %s\n", outputFile)
	if err := os.WriteFile(filepath.Join(actionDir, "run.sh"), []byte(script), 0755); err != nil {
		t.Fatal(err)
	}

	// Write command YAML referencing the external action
	// Note: sources block is present but fetch is skipped since cache is pre-populated
	commandYAML := `sources:
  test-src: github.com/test/test-project-actions@v1

help:
  short: Test external action

steps:
  - action: test-src/greet
    with:
      name: World
`
	if err := os.WriteFile(filepath.Join(commandsDir, "greet.yaml"), []byte(commandYAML), 0644); err != nil {
		t.Fatal(err)
	}

	// Change to tmpDir so config.LoadConfig() finds .project
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	// WARNING: os.Chdir is process-global. This test is not safe to run with t.Parallel().
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	log := logger.New()
	eng := executor.NewEngine(cfg, log)

	if err := eng.ExecuteCommand("greet", []string{}); err != nil {
		t.Fatalf("ExecuteCommand() error = %v", err)
	}

	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("expected output file to exist: %v", err)
	}
	if string(content) != "Hello, World!\n" {
		t.Errorf("output = %q, want %q", string(content), "Hello, World!\n")
	}
}

// TestExternalShellActionDefault verifies that a default input value is used
// when the with: block does not provide it.
func TestExternalShellActionDefault(t *testing.T) {
	tmpDir := t.TempDir()

	commandsDir := filepath.Join(tmpDir, ".project", "commands")
	if err := os.MkdirAll(commandsDir, 0755); err != nil {
		t.Fatal(err)
	}

	actionDir := filepath.Join(tmpDir, ".project", ".runtime", "actions", "test-src", "greet")
	if err := os.MkdirAll(actionDir, 0755); err != nil {
		t.Fatal(err)
	}

	actionYAML := `name: Greeter
inputs:
  name:
    required: true
  greeting:
    required: false
    default: Howdy
`
	if err := os.WriteFile(filepath.Join(actionDir, "action.yaml"), []byte(actionYAML), 0644); err != nil {
		t.Fatal(err)
	}

	outputFile := filepath.Join(tmpDir, "output.txt")
	script := fmt.Sprintf("#!/bin/sh\necho \"${INPUT_GREETING}, ${INPUT_NAME}!\" > %s\n", outputFile)
	if err := os.WriteFile(filepath.Join(actionDir, "run.sh"), []byte(script), 0755); err != nil {
		t.Fatal(err)
	}

	commandYAML := `sources:
  test-src: github.com/test/test-project-actions@v1

help:
  short: Test default input

steps:
  - action: test-src/greet
    with:
      name: Partner
`
	if err := os.WriteFile(filepath.Join(commandsDir, "greet-default.yaml"), []byte(commandYAML), 0644); err != nil {
		t.Fatal(err)
	}

	// WARNING: os.Chdir is process-global. This test is not safe to run with t.Parallel().
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	log := logger.New()
	eng := executor.NewEngine(cfg, log)

	if err := eng.ExecuteCommand("greet-default", []string{}); err != nil {
		t.Fatalf("ExecuteCommand() error = %v", err)
	}

	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("expected output file: %v", err)
	}
	if string(content) != "Howdy, Partner!\n" {
		t.Errorf("output = %q, want %q", string(content), "Howdy, Partner!\n")
	}
}
