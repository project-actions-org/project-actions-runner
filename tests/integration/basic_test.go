package integration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/project-actions/runner/internal/config"
	"github.com/project-actions/runner/internal/executor"
	"github.com/project-actions/runner/internal/logger"
)

func TestBasicCommandExecution(t *testing.T) {
	// Create a temporary project directory
	tmpDir, err := os.MkdirTemp("", "integration-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .project/commands directory
	commandsDir := filepath.Join(tmpDir, ".project", "commands")
	if err := os.MkdirAll(commandsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create a test command YAML file
	testYAML := `help:
  short: Test command
  long: Integration test command
  order: 1

steps:
  - echo: "Starting test"
  - run: "echo 'Hello from shell'"
  - echo: "Test complete"
`
	testPath := filepath.Join(commandsDir, "test.yaml")
	if err := os.WriteFile(testPath, []byte(testYAML), 0644); err != nil {
		t.Fatal(err)
	}

	// Change to the temporary directory
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(tmpDir)

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Create engine and execute command
	log := logger.New()
	engine := executor.NewEngine(cfg, log)

	err = engine.ExecuteCommand("test", []string{})
	if err != nil {
		t.Errorf("ExecuteCommand() error = %v", err)
	}
}

func TestCommandWithOptions(t *testing.T) {
	// Create a temporary project directory
	tmpDir, err := os.MkdirTemp("", "integration-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .project/commands directory
	commandsDir := filepath.Join(tmpDir, ".project", "commands")
	if err := os.MkdirAll(commandsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create a test command with conditionals
	testYAML := `help:
  short: Test command with options
  long: Test conditional logic

steps:
  - echo: "Running tests"

  - if-option: verbose
    then:
      - echo: "Verbose mode enabled"

  - if-option: debug|trace
    then:
      - echo: "Debug or trace mode enabled"

  - echo: "Done"
`
	testPath := filepath.Join(commandsDir, "test-options.yaml")
	if err := os.WriteFile(testPath, []byte(testYAML), 0644); err != nil {
		t.Fatal(err)
	}

	// Change to the temporary directory
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(tmpDir)

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	tests := []struct {
		name string
		args []string
	}{
		{
			name: "no options",
			args: []string{},
		},
		{
			name: "with verbose option",
			args: []string{"--verbose"},
		},
		{
			name: "with debug option",
			args: []string{"--debug"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := logger.New()
			engine := executor.NewEngine(cfg, log)

			err := engine.ExecuteCommand("test-options", tt.args)
			if err != nil {
				t.Errorf("ExecuteCommand() error = %v", err)
			}
		})
	}
}

func TestCommandAction(t *testing.T) {
	// Create a temporary project directory
	tmpDir, err := os.MkdirTemp("", "integration-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .project/commands directory
	commandsDir := filepath.Join(tmpDir, ".project", "commands")
	if err := os.MkdirAll(commandsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create a simple command
	simpleYAML := `help:
  short: Simple command

steps:
  - echo: "Simple command executed"
`
	simplePath := filepath.Join(commandsDir, "simple.yaml")
	if err := os.WriteFile(simplePath, []byte(simpleYAML), 0644); err != nil {
		t.Fatal(err)
	}

	// Create a command that calls another command
	callerYAML := `help:
  short: Caller command

steps:
  - echo: "Before calling simple"
  - command: simple
  - echo: "After calling simple"
`
	callerPath := filepath.Join(commandsDir, "caller.yaml")
	if err := os.WriteFile(callerPath, []byte(callerYAML), 0644); err != nil {
		t.Fatal(err)
	}

	// Change to the temporary directory
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(tmpDir)

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Execute the caller command
	log := logger.New()
	engine := executor.NewEngine(cfg, log)

	err = engine.ExecuteCommand("caller", []string{})
	if err != nil {
		t.Errorf("ExecuteCommand() error = %v", err)
	}
}

func TestCheckForAction(t *testing.T) {
	// Create a temporary project directory
	tmpDir, err := os.MkdirTemp("", "integration-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .project/commands directory
	commandsDir := filepath.Join(tmpDir, ".project", "commands")
	if err := os.MkdirAll(commandsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create a command that checks for required tools
	testYAML := `help:
  short: Test check-for action

steps:
  - check-for: sh
  - check-for: echo
  - echo: "All required tools found"
`
	testPath := filepath.Join(commandsDir, "test-check.yaml")
	if err := os.WriteFile(testPath, []byte(testYAML), 0644); err != nil {
		t.Fatal(err)
	}

	// Change to the temporary directory
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(tmpDir)

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Execute the command
	log := logger.New()
	engine := executor.NewEngine(cfg, log)

	err = engine.ExecuteCommand("test-check", []string{})
	if err != nil {
		t.Errorf("ExecuteCommand() error = %v", err)
	}
}

func TestIfMissingConditional(t *testing.T) {
	// Create a temporary project directory
	tmpDir, err := os.MkdirTemp("", "integration-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .project/commands directory
	commandsDir := filepath.Join(tmpDir, ".project", "commands")
	if err := os.MkdirAll(commandsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create a command with if-missing conditional
	testYAML := `help:
  short: Test if-missing conditional

steps:
  - echo: "Checking for file"

  - if-missing: nonexistent-file.txt
    then:
      - echo: "File is missing (as expected)"

  - echo: "Done"
`
	testPath := filepath.Join(commandsDir, "test-if-missing.yaml")
	if err := os.WriteFile(testPath, []byte(testYAML), 0644); err != nil {
		t.Fatal(err)
	}

	// Change to the temporary directory
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(tmpDir)

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Execute the command
	log := logger.New()
	engine := executor.NewEngine(cfg, log)

	err = engine.ExecuteCommand("test-if-missing", []string{})
	if err != nil {
		t.Errorf("ExecuteCommand() error = %v", err)
	}
}
