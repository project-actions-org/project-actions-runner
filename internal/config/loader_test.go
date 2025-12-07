package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Create a temporary directory with .project
	tmpDir, err := os.MkdirTemp("", "config-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	projectDir := filepath.Join(tmpDir, ".project")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Change to the temp directory
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(tmpDir)

	// Load config
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	// Resolve symlinks for comparison (macOS /var -> /private/var)
	tmpDirResolved, _ := filepath.EvalSymlinks(tmpDir)
	projectDirResolved, _ := filepath.EvalSymlinks(projectDir)

	cfgRootResolved, _ := filepath.EvalSymlinks(cfg.ProjectRoot)
	cfgProjectDirResolved, _ := filepath.EvalSymlinks(cfg.ProjectDir)
	cfgCommandsDirResolved, _ := filepath.EvalSymlinks(cfg.CommandsDir)
	cfgRuntimeDirResolved, _ := filepath.EvalSymlinks(cfg.RuntimeDir)

	if cfgRootResolved != tmpDirResolved {
		t.Errorf("ProjectRoot = %v, want %v", cfgRootResolved, tmpDirResolved)
	}

	if cfgProjectDirResolved != projectDirResolved {
		t.Errorf("ProjectDir = %v, want %v", cfgProjectDirResolved, projectDirResolved)
	}

	if cfgCommandsDirResolved != projectDirResolved {
		t.Errorf("CommandsDir = %v, want %v", cfgCommandsDirResolved, projectDirResolved)
	}

	expectedRuntimeDir, _ := filepath.EvalSymlinks(filepath.Join(projectDir, ".runtime"))
	if cfgRuntimeDirResolved != expectedRuntimeDir {
		t.Errorf("RuntimeDir = %v, want %v", cfgRuntimeDirResolved, expectedRuntimeDir)
	}
}

func TestLoadConfig_WithCommandsSubdir(t *testing.T) {
	// Create a temporary directory with .project/commands
	tmpDir, err := os.MkdirTemp("", "config-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	projectDir := filepath.Join(tmpDir, ".project")
	commandsDir := filepath.Join(projectDir, "commands")
	if err := os.MkdirAll(commandsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Change to the temp directory
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(tmpDir)

	// Load config
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	// Resolve symlinks for comparison
	commandsDirResolved, _ := filepath.EvalSymlinks(commandsDir)
	cfgCommandsDirResolved, _ := filepath.EvalSymlinks(cfg.CommandsDir)

	if cfgCommandsDirResolved != commandsDirResolved {
		t.Errorf("CommandsDir = %v, want %v", cfgCommandsDirResolved, commandsDirResolved)
	}
}

func TestLoadConfig_NoProjectDir(t *testing.T) {
	// Create a temporary directory without .project
	tmpDir, err := os.MkdirTemp("", "config-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Change to the temp directory
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(tmpDir)

	// Load config should fail
	_, err = LoadConfig()
	if err == nil {
		t.Error("LoadConfig() should return error when .project dir doesn't exist")
	}
}

func TestListCommands(t *testing.T) {
	// Create a temporary directory with .project
	tmpDir, err := os.MkdirTemp("", "config-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	projectDir := filepath.Join(tmpDir, ".project")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create some test command files
	testFiles := []string{"test.yaml", "setup.yaml", "deploy.yml", "readme.txt"}
	for _, file := range testFiles {
		path := filepath.Join(projectDir, file)
		if err := os.WriteFile(path, []byte("help:\n  short: Test"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	cfg := &Config{
		ProjectRoot: tmpDir,
		ProjectDir:  projectDir,
		CommandsDir: projectDir,
	}

	commands, err := cfg.ListCommands()
	if err != nil {
		t.Fatalf("ListCommands() error = %v", err)
	}

	// Should find 3 YAML files (test, setup, deploy)
	if len(commands) != 3 {
		t.Errorf("ListCommands() found %d commands, want 3", len(commands))
	}

	// Check that only YAML files are included
	expected := map[string]bool{"test": true, "setup": true, "deploy": true}
	for _, cmd := range commands {
		if !expected[cmd] {
			t.Errorf("Unexpected command: %s", cmd)
		}
	}
}

func TestFindCommandFile(t *testing.T) {
	// Create a temporary directory with .project
	tmpDir, err := os.MkdirTemp("", "config-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	projectDir := filepath.Join(tmpDir, ".project")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create test files
	testPath := filepath.Join(projectDir, "test.yaml")
	if err := os.WriteFile(testPath, []byte("help:\n  short: Test"), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := &Config{
		ProjectRoot: tmpDir,
		ProjectDir:  projectDir,
		CommandsDir: projectDir,
	}

	tests := []struct {
		name    string
		command string
		wantErr bool
	}{
		{
			name:    "existing command",
			command: "test",
			wantErr: false,
		},
		{
			name:    "nonexistent command",
			command: "nonexistent",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, err := cfg.FindCommandFile(tt.command)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindCommandFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && path != testPath {
				t.Errorf("FindCommandFile() = %v, want %v", path, testPath)
			}
		})
	}
}
