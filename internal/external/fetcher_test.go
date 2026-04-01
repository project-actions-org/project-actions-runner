package external

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEnsureGitignore(t *testing.T) {
	tmpDir := t.TempDir()
	f := &Fetcher{ActionsDir: filepath.Join(tmpDir, ".project", ".runtime", "actions")}

	// When no .gitignore exists, it should be created
	if err := f.EnsureGitignore(tmpDir); err != nil {
		t.Fatalf("EnsureGitignore() error = %v", err)
	}

	content, err := os.ReadFile(filepath.Join(tmpDir, ".gitignore"))
	if err != nil {
		t.Fatalf("expected .gitignore to be created: %v", err)
	}
	if !containsLine(string(content), ".project/.runtime/actions/") {
		t.Error("expected .gitignore to contain .project/.runtime/actions/")
	}

	// Calling again should not duplicate the entry
	if err := f.EnsureGitignore(tmpDir); err != nil {
		t.Fatalf("EnsureGitignore() second call error = %v", err)
	}

	content2, err := os.ReadFile(filepath.Join(tmpDir, ".gitignore"))
	if err != nil {
		t.Fatal(err)
	}
	count := countOccurrences(string(content2), ".project/.runtime/actions/")
	if count != 1 {
		t.Errorf("expected exactly 1 entry, got %d", count)
	}
}

func TestBinaryPath(t *testing.T) {
	f := &Fetcher{ActionsDir: "/project/.runtime/actions"}
	got := f.BinaryPath("aws", "iam-role-setup", "darwin", "arm64")
	want := "/project/.runtime/actions/aws/iam-role-setup/iam-role-setup-darwin-arm64"
	if got != want {
		t.Errorf("BinaryPath() = %q, want %q", got, want)
	}
}

func TestIsFetched(t *testing.T) {
	tmpDir := t.TempDir()
	f := &Fetcher{ActionsDir: filepath.Join(tmpDir, "actions")}

	if f.IsFetched("aws") {
		t.Error("expected IsFetched to return false for non-existent source")
	}

	// Create the source directory
	if err := os.MkdirAll(filepath.Join(tmpDir, "actions", "aws"), 0755); err != nil {
		t.Fatal(err)
	}

	if !f.IsFetched("aws") {
		t.Error("expected IsFetched to return true after directory created")
	}
}

func containsLine(content, line string) bool {
	for _, l := range splitLines(content) {
		if strings.TrimSpace(l) == line {
			return true
		}
	}
	return false
}

func countOccurrences(content, substr string) int {
	count := 0
	for _, l := range splitLines(content) {
		if strings.TrimSpace(l) == substr {
			count++
		}
	}
	return count
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i, c := range s {
		if c == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}
