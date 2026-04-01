package external

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBuildEnv(t *testing.T) {
	with := map[string]interface{}{
		"role-name": "my-role",
		"region":    "us-east-1",
	}
	env := buildInputEnv(with)

	found := map[string]bool{}
	for _, e := range env {
		found[e] = true
	}

	if !found["INPUT_ROLE_NAME=my-role"] {
		t.Error("expected INPUT_ROLE_NAME=my-role in env")
	}
	if !found["INPUT_REGION=us-east-1"] {
		t.Error("expected INPUT_REGION=us-east-1 in env")
	}
}

func TestBuildContextEnv(t *testing.T) {
	env := buildContextEnv("/home/user/myproject", "setup", true)

	found := map[string]bool{}
	for _, e := range env {
		found[e] = true
	}

	if !found["PROJECT_ROOT=/home/user/myproject"] {
		t.Error("expected PROJECT_ROOT=/home/user/myproject")
	}
	if !found["PROJECT_NAME=myproject"] {
		t.Error("expected PROJECT_NAME=myproject")
	}
	if !found["PROJECT_COMMAND=setup"] {
		t.Error("expected PROJECT_COMMAND=setup")
	}
	if !found["PROJECT_VERBOSE=true"] {
		t.Error("expected PROJECT_VERBOSE=true")
	}
}

func TestBuildContextEnvNotVerbose(t *testing.T) {
	env := buildContextEnv("/home/user/myproject", "setup", false)

	for _, e := range env {
		if e == "PROJECT_VERBOSE=true" {
			t.Error("PROJECT_VERBOSE should not be set when verbose=false")
		}
	}
}

func TestInputEnvKeyFormat(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"role-name", "INPUT_ROLE_NAME"},
		{"region", "INPUT_REGION"},
		{"my-long-key-name", "INPUT_MY_LONG_KEY_NAME"},
	}
	for _, tt := range tests {
		got := inputEnvKey(tt.input)
		if got != tt.want {
			t.Errorf("inputEnvKey(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

// TestExecuteShellAction runs a real shell script and verifies it runs.
func TestExecuteShellAction(t *testing.T) {
	tmpDir := t.TempDir()

	scriptPath := filepath.Join(tmpDir, "run.sh")
	outputFile := filepath.Join(tmpDir, "out.txt")
	script := "#!/bin/sh\necho \"${INPUT_NAME}\" > " + outputFile + "\n"
	if err := os.WriteFile(scriptPath, []byte(script), 0644); err != nil {
		t.Fatal(err)
	}

	resolved := &ResolvedAction{
		ActionDir: tmpDir,
		ActionMeta: &ActionMeta{
			Inputs: map[string]InputSpec{
				"name": {Required: true},
			},
		},
		Type:     ShellAction,
		ExecPath: scriptPath,
	}

	err := ExecuteAction(resolved, tmpDir, "setup", false, map[string]interface{}{"name": "world"})
	if err != nil {
		t.Fatalf("ExecuteAction() error = %v", err)
	}

	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("expected output file to exist: %v", err)
	}
	if string(content) != "world\n" {
		t.Errorf("output = %q, want %q", string(content), "world\n")
	}
}
