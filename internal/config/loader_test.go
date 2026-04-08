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

	entries, err := cfg.ListCommands()
	if err != nil {
		t.Fatalf("ListCommands() error = %v", err)
	}

	if len(entries) != 3 {
		t.Errorf("ListCommands() found %d commands, want 3", len(entries))
	}

	expected := map[string]bool{"test": true, "setup": true, "deploy": true}
	for _, e := range entries {
		if !expected[e.Name] {
			t.Errorf("Unexpected command: %s", e.Name)
		}
	}
}

func TestListCommands_Recursive(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, ".project")
	buildDir := filepath.Join(projectDir, "build")
	dockerDir := filepath.Join(projectDir, "build", "docker")
	if err := os.MkdirAll(dockerDir, 0755); err != nil {
		t.Fatal(err)
	}

	files := map[string]string{
		filepath.Join(projectDir, "setup.yaml"):  "help:\n  short: Setup",
		filepath.Join(buildDir, "all.yaml"):      "help:\n  short: Build all",
		filepath.Join(buildDir, "release.yaml"):  "help:\n  short: Build release",
		filepath.Join(dockerDir, "image.yaml"):   "help:\n  short: Build image",
		filepath.Join(projectDir, "readme.txt"):  "not a command",
	}
	for path, content := range files {
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	cfg := &Config{
		ProjectRoot: tmpDir,
		ProjectDir:  projectDir,
		CommandsDir: projectDir,
	}

	entries, err := cfg.ListCommands()
	if err != nil {
		t.Fatalf("ListCommands() error = %v", err)
	}

	byName := make(map[string]CommandEntry)
	for _, e := range entries {
		byName[e.Name] = e
	}

	if len(entries) != 4 {
		t.Errorf("want 4 entries, got %d: %v", len(entries), entries)
	}

	e, ok := byName["setup"]
	if !ok {
		t.Fatal("missing 'setup'")
	}
	if len(e.Namespace) != 0 {
		t.Errorf("setup namespace want [], got %v", e.Namespace)
	}

	e, ok = byName["build:all"]
	if !ok {
		t.Fatal("missing 'build:all'")
	}
	if len(e.Namespace) != 1 || e.Namespace[0] != "build" {
		t.Errorf("build:all namespace want [build], got %v", e.Namespace)
	}

	e, ok = byName["build:docker:image"]
	if !ok {
		t.Fatal("missing 'build:docker:image'")
	}
	if len(e.Namespace) != 2 || e.Namespace[0] != "build" || e.Namespace[1] != "docker" {
		t.Errorf("build:docker:image namespace want [build docker], got %v", e.Namespace)
	}
}

func TestFindCommandFile_Namespaced(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, ".project")
	buildDir := filepath.Join(projectDir, "build")
	if err := os.MkdirAll(buildDir, 0755); err != nil {
		t.Fatal(err)
	}
	allPath := filepath.Join(buildDir, "all.yaml")
	if err := os.WriteFile(allPath, []byte("help:\n  short: Build all"), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := &Config{
		ProjectRoot: tmpDir,
		ProjectDir:  projectDir,
		CommandsDir: projectDir,
	}

	got, err := cfg.FindCommandFile("build:all")
	if err != nil {
		t.Fatalf("FindCommandFile() error = %v", err)
	}
	if got != allPath {
		t.Errorf("FindCommandFile() = %v, want %v", got, allPath)
	}

	_, err = cfg.FindCommandFile("build:missing")
	if err == nil {
		t.Error("expected error for missing namespaced command, got nil")
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
