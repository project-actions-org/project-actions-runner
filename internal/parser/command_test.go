package parser

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseStep(t *testing.T) {
	tests := []struct {
		name       string
		raw        map[string]interface{}
		wantAction string
		wantErr    bool
	}{
		{
			name: "echo step",
			raw: map[string]interface{}{
				"echo": "Hello",
			},
			wantAction: "echo",
			wantErr:    false,
		},
		{
			name: "run step",
			raw: map[string]interface{}{
				"run": "pwd",
			},
			wantAction: "run",
			wantErr:    false,
		},
		{
			name: "command step",
			raw: map[string]interface{}{
				"command": "setup",
			},
			wantAction: "command",
			wantErr:    false,
		},
		{
			name: "action step",
			raw: map[string]interface{}{
				"action": "compose-up",
			},
			wantAction: "compose-up",
			wantErr:    false,
		},
		{
			name: "check-for step",
			raw: map[string]interface{}{
				"check-for": "docker",
			},
			wantAction: "check-for",
			wantErr:    false,
		},
		{
			name: "if-option conditional",
			raw: map[string]interface{}{
				"if-option": "refresh",
				"then": []interface{}{
					map[string]interface{}{
						"echo": "Refreshing...",
					},
				},
			},
			wantAction: "",
			wantErr:    false,
		},
		{
			name: "if-no-option conditional",
			raw: map[string]interface{}{
				"if-no-option": "skip",
				"then": []interface{}{
					map[string]interface{}{
						"run": "some-command",
					},
				},
			},
			wantAction: "",
			wantErr:    false,
		},
		{
			name: "if-missing conditional",
			raw: map[string]interface{}{
				"if-missing": "./file.txt",
				"then": []interface{}{
					map[string]interface{}{
						"echo": "File not found",
					},
				},
			},
			wantAction: "",
			wantErr:    false,
		},
		{
			name: "if-fails conditional",
			raw: map[string]interface{}{
				"if-fails": "previous-step",
				"then": []interface{}{
					map[string]interface{}{
						"echo": "Handling failure",
					},
				},
			},
			wantAction: "",
			wantErr:    false,
		},
		{
			name:       "unknown step type",
			raw:        map[string]interface{}{"unknown": "value"},
			wantAction: "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			step, err := ParseStep(tt.raw)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseStep() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				// For conditionals, ActionName should be empty
				if step.Conditional != nil {
					if step.ActionName != "" {
						t.Errorf("ParseStep() conditional step has ActionName = %v, want empty", step.ActionName)
					}
				} else if tt.wantAction != "" && step.ActionName != tt.wantAction {
					t.Errorf("ParseStep() ActionName = %v, want %v", step.ActionName, tt.wantAction)
				}
			}
		})
	}
}

func TestParseCommandFile(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "parser-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a valid test YAML file
	validYAML := `help:
  short: Test command
  long: This is a test command
  order: 1

context: outside-container

steps:
  - echo: "Hello"
  - run: "pwd"
`
	validPath := filepath.Join(tmpDir, "valid.yaml")
	if err := os.WriteFile(validPath, []byte(validYAML), 0644); err != nil {
		t.Fatal(err)
	}

	// Create an invalid YAML file
	invalidYAML := `this is not: valid: yaml:`
	invalidPath := filepath.Join(tmpDir, "invalid.yaml")
	if err := os.WriteFile(invalidPath, []byte(invalidYAML), 0644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name        string
		filePath    string
		commandName string
		wantErr     bool
	}{
		{
			name:        "valid command file",
			filePath:    validPath,
			commandName: "valid",
			wantErr:     false,
		},
		{
			name:        "invalid YAML",
			filePath:    invalidPath,
			commandName: "invalid",
			wantErr:     true,
		},
		{
			name:        "nonexistent file",
			filePath:    filepath.Join(tmpDir, "nonexistent.yaml"),
			commandName: "nonexistent",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := ParseCommandFile(tt.filePath, tt.commandName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseCommandFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				if cmd.Name != tt.commandName {
					t.Errorf("ParseCommandFile() Name = %v, want %v", cmd.Name, tt.commandName)
				}
				if cmd.Help.Short == "" {
					t.Errorf("ParseCommandFile() Help.Short is empty")
				}
			}
		})
	}
}

func TestParseCommandFileWithSources(t *testing.T) {
	tmpDir := t.TempDir()
	yaml := `sources:
  aws: github.com/project-actions/aws-project-actions@v1
  utils: github.com/myorg/my-project-actions@main

help:
  short: Test command

steps:
  - echo: "hello"
`
	path := filepath.Join(tmpDir, "test.yaml")
	os.WriteFile(path, []byte(yaml), 0644)

	cmd, err := ParseCommandFile(path, "test")
	if err != nil {
		t.Fatalf("ParseCommandFile() error = %v", err)
	}

	if len(cmd.Sources) != 2 {
		t.Fatalf("expected 2 sources, got %d", len(cmd.Sources))
	}
	if cmd.Sources["aws"] != "github.com/project-actions/aws-project-actions@v1" {
		t.Errorf("aws source = %q, want %q", cmd.Sources["aws"], "github.com/project-actions/aws-project-actions@v1")
	}
	if cmd.Sources["utils"] != "github.com/myorg/my-project-actions@main" {
		t.Errorf("utils source = %q, want %q", cmd.Sources["utils"], "github.com/myorg/my-project-actions@main")
	}
}

func TestParseCommandFileWithoutSources(t *testing.T) {
	tmpDir := t.TempDir()
	yaml := `help:
  short: Test command

steps:
  - echo: "hello"
`
	path := filepath.Join(tmpDir, "nosources.yaml")
	os.WriteFile(path, []byte(yaml), 0644)

	cmd, err := ParseCommandFile(path, "nosources")
	if err != nil {
		t.Fatalf("ParseCommandFile() error = %v", err)
	}
	if cmd.Sources != nil && len(cmd.Sources) != 0 {
		t.Errorf("expected nil/empty sources, got %v", cmd.Sources)
	}
}
