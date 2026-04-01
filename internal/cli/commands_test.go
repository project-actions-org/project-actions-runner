package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/project-actions/runner/internal/config"
)

func TestValidateSourceConsistency_Conflict(t *testing.T) {
	tmpDir := t.TempDir()
	commandsDir := filepath.Join(tmpDir, ".project", "commands")
	if err := os.MkdirAll(commandsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Two command files with same source name but different refs
	yaml1 := `sources:
  aws: github.com/project-actions/aws-project-actions@v1
help:
  short: Command one
steps: []
`
	yaml2 := `sources:
  aws: github.com/project-actions/aws-project-actions@v2
help:
  short: Command two
steps: []
`
	if err := os.WriteFile(filepath.Join(commandsDir, "one.yaml"), []byte(yaml1), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(commandsDir, "two.yaml"), []byte(yaml2), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		ProjectRoot: tmpDir,
		ProjectDir:  filepath.Join(tmpDir, ".project"),
		CommandsDir: commandsDir,
		RuntimeDir:  filepath.Join(tmpDir, ".project", ".runtime"),
	}

	err := validateSourceConsistency(cfg)
	if err == nil {
		t.Error("expected conflict error, got nil")
	} else if !strings.Contains(err.Error(), "aws") {
		t.Errorf("expected error to mention source alias 'aws', got: %v", err)
	}
}

func TestValidateSourceConsistency_NoConflict(t *testing.T) {
	tmpDir := t.TempDir()
	commandsDir := filepath.Join(tmpDir, ".project", "commands")
	if err := os.MkdirAll(commandsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Same source name and same URL in both files — no conflict
	yaml1 := `sources:
  aws: github.com/project-actions/aws-project-actions@v1
help:
  short: Command one
steps: []
`
	yaml2 := `sources:
  aws: github.com/project-actions/aws-project-actions@v1
help:
  short: Command two
steps: []
`
	if err := os.WriteFile(filepath.Join(commandsDir, "one.yaml"), []byte(yaml1), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(commandsDir, "two.yaml"), []byte(yaml2), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		ProjectRoot: tmpDir,
		ProjectDir:  filepath.Join(tmpDir, ".project"),
		CommandsDir: commandsDir,
		RuntimeDir:  filepath.Join(tmpDir, ".project", ".runtime"),
	}

	if err := validateSourceConsistency(cfg); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCollectAllSources_Basic(t *testing.T) {
	tmpDir := t.TempDir()
	commandsDir := filepath.Join(tmpDir, ".project", "commands")
	if err := os.MkdirAll(commandsDir, 0755); err != nil {
		t.Fatal(err)
	}

	yaml1 := `sources:
  aws: github.com/project-actions/aws-project-actions@v1
help:
  short: Command one
steps: []
`
	yaml2 := `sources:
  docker: github.com/project-actions/docker-project-actions@v1
help:
  short: Command two
steps: []
`
	if err := os.WriteFile(filepath.Join(commandsDir, "one.yaml"), []byte(yaml1), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(commandsDir, "two.yaml"), []byte(yaml2), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		ProjectRoot: tmpDir,
		ProjectDir:  filepath.Join(tmpDir, ".project"),
		CommandsDir: commandsDir,
		RuntimeDir:  filepath.Join(tmpDir, ".project", ".runtime"),
	}

	sources, err := collectAllSources(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sources["aws"] != "github.com/project-actions/aws-project-actions@v1" {
		t.Errorf("expected aws source, got %v", sources["aws"])
	}
	if sources["docker"] != "github.com/project-actions/docker-project-actions@v1" {
		t.Errorf("expected docker source, got %v", sources["docker"])
	}
}

func TestValidateSourceConsistency_NoSources(t *testing.T) {
	tmpDir := t.TempDir()
	commandsDir := filepath.Join(tmpDir, ".project", "commands")
	if err := os.MkdirAll(commandsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Command files with no sources: block at all
	yaml1 := `help:
  short: Command one
steps: []
`
	if err := os.WriteFile(filepath.Join(commandsDir, "one.yaml"), []byte(yaml1), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		ProjectRoot: tmpDir,
		ProjectDir:  filepath.Join(tmpDir, ".project"),
		CommandsDir: commandsDir,
		RuntimeDir:  filepath.Join(tmpDir, ".project", ".runtime"),
	}

	if err := validateSourceConsistency(cfg); err != nil {
		t.Errorf("unexpected error for command with no sources: %v", err)
	}
}
