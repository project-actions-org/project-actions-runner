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
			name: "mkdir step",
			raw: map[string]interface{}{
				"mkdir": "dist",
			},
			wantAction: "mkdir",
			wantErr:    false,
		},
		{
			name: "remove step",
			raw: map[string]interface{}{
				"remove": "dist",
			},
			wantAction: "remove",
			wantErr:    false,
		},
		{
			name: "link step with src and dest",
			raw: map[string]interface{}{
				"link": map[string]interface{}{
					"src":  "/from",
					"dest": "/to",
				},
			},
			wantAction: "link",
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

func TestParseStep_FilesystemPrimitiveConfigs(t *testing.T) {
	t.Run("mkdir config value", func(t *testing.T) {
		step, err := ParseStep(map[string]interface{}{"mkdir": "dist"})
		if err != nil {
			t.Fatalf("ParseStep() error = %v", err)
		}
		if step.ActionName != "mkdir" {
			t.Errorf("ActionName = %q, want %q", step.ActionName, "mkdir")
		}
		if step.Config["mkdir"] != "dist" {
			t.Errorf("config[mkdir] = %v, want dist", step.Config["mkdir"])
		}
	})

	t.Run("remove config value", func(t *testing.T) {
		step, err := ParseStep(map[string]interface{}{"remove": "dist"})
		if err != nil {
			t.Fatalf("ParseStep() error = %v", err)
		}
		if step.ActionName != "remove" {
			t.Errorf("ActionName = %q, want %q", step.ActionName, "remove")
		}
		if step.Config["remove"] != "dist" {
			t.Errorf("config[remove] = %v, want dist", step.Config["remove"])
		}
	})

	t.Run("link src and dest flattened into config", func(t *testing.T) {
		step, err := ParseStep(map[string]interface{}{
			"link": map[string]interface{}{
				"src":  "/from",
				"dest": "/to",
			},
		})
		if err != nil {
			t.Fatalf("ParseStep() error = %v", err)
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
	})
}

func TestParseStep_ForLoop(t *testing.T) {
	t.Run("string list", func(t *testing.T) {
		raw := map[string]interface{}{
			"for":   []interface{}{"a", "b"},
			"steps": []interface{}{map[string]interface{}{"echo": "hi"}},
		}
		step, err := ParseStep(raw)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if step.ForLoop == nil {
			t.Fatal("expected ForLoop to be set")
		}
		if step.ForLoop.VarName != "item" {
			t.Errorf("VarName = %q, want %q", step.ForLoop.VarName, "item")
		}
		if len(step.ForLoop.Items) != 2 {
			t.Errorf("Items len = %d, want 2", len(step.ForLoop.Items))
		}
		if len(step.ForLoop.Steps) != 1 {
			t.Errorf("Steps len = %d, want 1", len(step.ForLoop.Steps))
		}
	})

	t.Run("glob source", func(t *testing.T) {
		raw := map[string]interface{}{
			"for":   map[string]interface{}{"glob": "src/*/"},
			"steps": []interface{}{map[string]interface{}{"echo": "hi"}},
		}
		step, err := ParseStep(raw)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if step.ForLoop == nil {
			t.Fatal("expected ForLoop to be set")
		}
		if step.ForLoop.Glob != "src/*/" {
			t.Errorf("Glob = %q, want %q", step.ForLoop.Glob, "src/*/")
		}
	})

	t.Run("custom var name via as:", func(t *testing.T) {
		raw := map[string]interface{}{
			"for":   []interface{}{"x"},
			"as":    "entry",
			"steps": []interface{}{map[string]interface{}{"echo": "hi"}},
		}
		step, err := ParseStep(raw)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if step.ForLoop.VarName != "entry" {
			t.Errorf("VarName = %q, want %q", step.ForLoop.VarName, "entry")
		}
	})

	t.Run("empty as: is rejected", func(t *testing.T) {
		// as: "" should be ignored, VarName stays "item"
		raw := map[string]interface{}{
			"for":   []interface{}{"x"},
			"as":    "",
			"steps": []interface{}{},
		}
		step, err := ParseStep(raw)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if step.ForLoop.VarName != "item" {
			t.Errorf("VarName = %q, want %q (empty as: should be ignored)", step.ForLoop.VarName, "item")
		}
	})

	t.Run("for:{} with neither list nor glob returns error", func(t *testing.T) {
		raw := map[string]interface{}{
			"for":   map[string]interface{}{},
			"steps": []interface{}{},
		}
		_, err := ParseStep(raw)
		if err == nil {
			t.Error("expected error for for:{} with no list or glob")
		}
	})
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
	if err := os.WriteFile(path, []byte(yaml), 0644); err != nil {
		t.Fatal(err)
	}

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
	if err := os.WriteFile(path, []byte(yaml), 0644); err != nil {
		t.Fatal(err)
	}

	cmd, err := ParseCommandFile(path, "nosources")
	if err != nil {
		t.Fatalf("ParseCommandFile() error = %v", err)
	}
	if len(cmd.Sources) != 0 {
		t.Errorf("expected nil/empty sources, got %v", cmd.Sources)
	}
}
