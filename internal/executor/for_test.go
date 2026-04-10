package executor

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/project-actions/runner/internal/config"
	"github.com/project-actions/runner/internal/logger"
)

func TestForLoop_StringList(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, ".project")
	commandsDir := filepath.Join(projectDir, "commands")
	_ = os.MkdirAll(commandsDir, 0755)

	yaml := `help:
  short: test for loop
steps:
  - for: [hello, world]
    steps:
      - mkdir: out/<item>
`
	_ = os.WriteFile(filepath.Join(commandsDir, "test.yaml"), []byte(yaml), 0644)

	cfg := &config.Config{
		ProjectRoot: tmpDir,
		ProjectDir:  projectDir,
		CommandsDir: commandsDir,
		RuntimeDir:  filepath.Join(projectDir, ".runtime"),
	}
	engine := NewEngine(cfg, logger.New())
	if err := engine.ExecuteCommand("test", nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, d := range []string{"out/hello", "out/world"} {
		if _, err := os.Stat(filepath.Join(tmpDir, d)); err != nil {
			t.Errorf("expected %s to be created: %v", d, err)
		}
	}
}

func TestForLoop_ObjectList(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, ".project")
	commandsDir := filepath.Join(projectDir, "commands")
	_ = os.MkdirAll(commandsDir, 0755)

	yaml := `help:
  short: test for loop objects
steps:
  - for:
      - {name: alpha}
      - {name: beta}
    steps:
      - mkdir: out/<item.name>
`
	_ = os.WriteFile(filepath.Join(commandsDir, "test.yaml"), []byte(yaml), 0644)

	cfg := &config.Config{
		ProjectRoot: tmpDir,
		ProjectDir:  projectDir,
		CommandsDir: commandsDir,
		RuntimeDir:  filepath.Join(projectDir, ".runtime"),
	}
	engine := NewEngine(cfg, logger.New())
	if err := engine.ExecuteCommand("test", nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, d := range []string{"out/alpha", "out/beta"} {
		if _, err := os.Stat(filepath.Join(tmpDir, d)); err != nil {
			t.Errorf("expected %s to be created: %v", d, err)
		}
	}
}

func TestForLoop_CustomVarName(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, ".project")
	commandsDir := filepath.Join(projectDir, "commands")
	_ = os.MkdirAll(commandsDir, 0755)

	yaml := `help:
  short: test for loop custom var
steps:
  - for: [foo, bar]
    as: entry
    steps:
      - mkdir: out/<entry>
`
	_ = os.WriteFile(filepath.Join(commandsDir, "test.yaml"), []byte(yaml), 0644)

	cfg := &config.Config{
		ProjectRoot: tmpDir,
		ProjectDir:  projectDir,
		CommandsDir: commandsDir,
		RuntimeDir:  filepath.Join(projectDir, ".runtime"),
	}
	engine := NewEngine(cfg, logger.New())
	if err := engine.ExecuteCommand("test", nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, d := range []string{"out/foo", "out/bar"} {
		if _, err := os.Stat(filepath.Join(tmpDir, d)); err != nil {
			t.Errorf("expected %s to be created: %v", d, err)
		}
	}
}

func TestForLoop_GlobSource(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, ".project")
	commandsDir := filepath.Join(projectDir, "commands")
	_ = os.MkdirAll(commandsDir, 0755)

	// Create some directories to glob
	_ = os.MkdirAll(filepath.Join(tmpDir, "src", "alpha"), 0755)
	_ = os.MkdirAll(filepath.Join(tmpDir, "src", "beta"), 0755)
	_ = os.MkdirAll(filepath.Join(tmpDir, "out"), 0755)

	yaml := `help:
  short: test for glob
steps:
  - for:
      glob: src/*/
    steps:
      - mkdir: out/<item>
`
	_ = os.WriteFile(filepath.Join(commandsDir, "test.yaml"), []byte(yaml), 0644)

	cfg := &config.Config{
		ProjectRoot: tmpDir,
		ProjectDir:  projectDir,
		CommandsDir: commandsDir,
		RuntimeDir:  filepath.Join(projectDir, ".runtime"),
	}
	engine := NewEngine(cfg, logger.New())
	if err := engine.ExecuteCommand("test", nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, d := range []string{"out/src/alpha", "out/src/beta"} {
		if _, err := os.Stat(filepath.Join(tmpDir, d)); err != nil {
			t.Errorf("expected %s to be created: %v", d, err)
		}
	}
}
