package docker

import (
	"os"
	"path/filepath"
	"testing"
)

func TestHasComposeFile(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(string)
		expected bool
	}{
		{
			name: "docker-compose.yml exists",
			setup: func(dir string) {
				os.WriteFile(filepath.Join(dir, "docker-compose.yml"), []byte("services:\n  test:\n    image: nginx"), 0644)
			},
			expected: true,
		},
		{
			name: "docker-compose.yaml exists",
			setup: func(dir string) {
				os.WriteFile(filepath.Join(dir, "docker-compose.yaml"), []byte("services:\n  test:\n    image: nginx"), 0644)
			},
			expected: true,
		},
		{
			name: "compose.yml exists",
			setup: func(dir string) {
				os.WriteFile(filepath.Join(dir, "compose.yml"), []byte("services:\n  test:\n    image: nginx"), 0644)
			},
			expected: true,
		},
		{
			name: "compose.yaml exists",
			setup: func(dir string) {
				os.WriteFile(filepath.Join(dir, "compose.yaml"), []byte("services:\n  test:\n    image: nginx"), 0644)
			},
			expected: true,
		},
		{
			name:     "no compose file",
			setup:    func(dir string) {},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir, err := os.MkdirTemp("", "compose-test")
			if err != nil {
				t.Fatal(err)
			}
			defer os.RemoveAll(tmpDir)

			tt.setup(tmpDir)

			result := HasComposeFile(tmpDir)
			if result != tt.expected {
				t.Errorf("HasComposeFile() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// Note: Testing actual docker-compose commands would require Docker to be running
// and a proper compose file, so we keep these as integration tests rather than unit tests
// The actual command execution is tested through integration tests
