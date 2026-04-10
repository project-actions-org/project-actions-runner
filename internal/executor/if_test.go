package executor

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/project-actions/runner/internal/actions"
	"github.com/project-actions/runner/internal/config"
	"github.com/project-actions/runner/internal/logger"
)

// --- Unit tests for the DSL evaluator ---

func makeIfCtx() *actions.ExecutionContext {
	return &actions.ExecutionContext{
		Options:     map[string]string{"verbose": "true"},
		Environment: map[string]string{"CI": "true", "EMPTY": ""},
		LoopVars:    map[string]interface{}{"item": map[string]interface{}{"os": "darwin"}},
		Logger:      logger.New(),
	}
}

func TestEvalIfExpr(t *testing.T) {
	ctx := makeIfCtx()
	tests := []struct {
		expr string
		want bool
	}{
		{"option.verbose", true},
		{"option.missing", false},
		{"!option.verbose", false},
		{"!option.missing", true},
		{"env.CI", true},
		{"env.EMPTY", false},
		{"env.CI != \"\"", true},
		{"env.EMPTY != \"\"", false},
		{"env.CI == true", true},
		{"option.verbose == true", true},
		{"option.missing == true", false},
		{"option.verbose && env.CI", true},
		{"option.verbose && option.missing", false},
		{"option.missing || env.CI", true},
		{"option.missing || option.missing", false},
		{"item.os == darwin", true},
		{"item.os == linux", false},
	}
	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			got, err := evalIfExpr(tt.expr, ctx)
			if err != nil {
				t.Fatalf("evalIfExpr(%q) error: %v", tt.expr, err)
			}
			if got != tt.want {
				t.Errorf("evalIfExpr(%q) = %v, want %v", tt.expr, got, tt.want)
			}
		})
	}
}

// --- Integration test: if: then:/else: in a command ---

func TestIfExpr_ThenExecuted(t *testing.T) {
	yaml := `help:
  short: test if
steps:
  - if: option.verbose
    then:
      - mkdir: created
`
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, ".project")
	commandsDir := filepath.Join(projectDir, "commands")
	_ = os.MkdirAll(commandsDir, 0755)
	_ = os.WriteFile(filepath.Join(commandsDir, "test.yaml"), []byte(yaml), 0644)
	cfg := &config.Config{
		ProjectRoot: tmpDir,
		ProjectDir:  projectDir,
		CommandsDir: commandsDir,
		RuntimeDir:  filepath.Join(projectDir, ".runtime"),
	}
	engine := NewEngine(cfg, logger.New())
	if err := engine.ExecuteCommand("test", []string{"--verbose"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := os.Stat(filepath.Join(tmpDir, "created")); err != nil {
		t.Error("expected 'created' dir to exist (then branch executed)")
	}
}

func TestIfExpr_ElseExecuted(t *testing.T) {
	yaml := `help:
  short: test if else
steps:
  - if: option.verbose
    then:
      - mkdir: then-branch
    else:
      - mkdir: else-branch
`
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, ".project")
	commandsDir := filepath.Join(projectDir, "commands")
	_ = os.MkdirAll(commandsDir, 0755)
	_ = os.WriteFile(filepath.Join(commandsDir, "test.yaml"), []byte(yaml), 0644)
	cfg := &config.Config{
		ProjectRoot: tmpDir,
		ProjectDir:  projectDir,
		CommandsDir: commandsDir,
		RuntimeDir:  filepath.Join(projectDir, ".runtime"),
	}
	engine := NewEngine(cfg, logger.New())
	// no --verbose flag → else branch
	if err := engine.ExecuteCommand("test", nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := os.Stat(filepath.Join(tmpDir, "else-branch")); err != nil {
		t.Error("expected 'else-branch' dir to exist")
	}
	if _, err := os.Stat(filepath.Join(tmpDir, "then-branch")); !os.IsNotExist(err) {
		t.Error("did not expect 'then-branch' to exist")
	}
}
