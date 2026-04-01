package cli

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/project-actions/runner/internal/config"
)

func TestFetchManifest(t *testing.T) {
	manifest := TemplatesManifest{
		"laravel": {Description: "Laravel starter", Files: []string{"setup.yaml"}},
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(manifest)
	}))
	defer ts.Close()

	got, err := fetchManifest(ts.URL)
	if err != nil {
		t.Fatalf("fetchManifest() error = %v", err)
	}
	if got["laravel"].Description != "Laravel starter" {
		t.Errorf("expected laravel description, got %v", got["laravel"])
	}
}

func TestFetchManifestServerError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	_, err := fetchManifest(ts.URL)
	if err == nil {
		t.Error("expected error for 500 response, got nil")
	}
}

func TestDownloadFile(t *testing.T) {
	content := "help:\n  short: Test\nsteps: []\n"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(content))
	}))
	defer ts.Close()

	tmpDir := t.TempDir()
	dest := filepath.Join(tmpDir, "test.yaml")

	if err := downloadFile(ts.URL, dest); err != nil {
		t.Fatalf("downloadFile() error = %v", err)
	}

	got, err := os.ReadFile(dest)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != content {
		t.Errorf("file content = %q, want %q", string(got), content)
	}
}

func TestDownloadFileServerError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	tmpDir := t.TempDir()
	err := downloadFile(ts.URL, filepath.Join(tmpDir, "out.yaml"))
	if err == nil {
		t.Error("expected error for 404 response, got nil")
	}
}

func TestInitCommandDownloadsFiles(t *testing.T) {
	mux := http.NewServeMux()
	manifest := TemplatesManifest{
		"testfw": {Description: "Test framework", Files: []string{"setup.yaml", "test.yaml"}},
	}
	mux.HandleFunc("/manifest.json", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(manifest)
	})
	mux.HandleFunc("/testfw/setup.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("help:\n  short: Setup\nsteps: []\n"))
	})
	mux.HandleFunc("/testfw/test.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("help:\n  short: Test\nsteps: []\n"))
	})
	ts := httptest.NewServer(mux)
	defer ts.Close()

	tmpDir := t.TempDir()
	commandsDir := filepath.Join(tmpDir, "commands")
	cfg := &config.Config{
		ProjectRoot: tmpDir,
		CommandsDir: commandsDir,
	}

	cmd := newInitCommandWithURLs(cfg, ts.URL+"/manifest.json", ts.URL)
	cmd.SetArgs([]string{"testfw"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("init command error = %v", err)
	}

	expectedContents := map[string]string{
		"setup.yaml": "help:\n  short: Setup\nsteps: []\n",
		"test.yaml":  "help:\n  short: Test\nsteps: []\n",
	}
	for file, want := range expectedContents {
		got, err := os.ReadFile(filepath.Join(commandsDir, file))
		if err != nil {
			t.Fatalf("expected %s to exist: %v", file, err)
		}
		if string(got) != want {
			t.Errorf("%s content = %q, want %q", file, string(got), want)
		}
	}
}

func TestInitCommandSkipsExistingFiles(t *testing.T) {
	mux := http.NewServeMux()
	manifest := TemplatesManifest{
		"testfw": {Description: "Test", Files: []string{"setup.yaml"}},
	}
	mux.HandleFunc("/manifest.json", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(manifest)
	})
	mux.HandleFunc("/testfw/setup.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("help:\n  short: New\nsteps: []\n"))
	})
	ts := httptest.NewServer(mux)
	defer ts.Close()

	tmpDir := t.TempDir()
	commandsDir := filepath.Join(tmpDir, "commands")
	if err := os.MkdirAll(commandsDir, 0755); err != nil {
		t.Fatal(err)
	}

	existingContent := "help:\n  short: Existing\nsteps: []\n"
	if err := os.WriteFile(filepath.Join(commandsDir, "setup.yaml"), []byte(existingContent), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{ProjectRoot: tmpDir, CommandsDir: commandsDir}
	cmd := newInitCommandWithURLs(cfg, ts.URL+"/manifest.json", ts.URL)
	cmd.SetArgs([]string{"testfw"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("init command error = %v", err)
	}

	got, err := os.ReadFile(filepath.Join(commandsDir, "setup.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != existingContent {
		t.Errorf("file was overwritten, got %q, want %q", string(got), existingContent)
	}
}

func TestInitCommandUnknownTemplate(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(TemplatesManifest{
			"laravel": {Description: "Laravel", Files: []string{"setup.yaml"}},
		})
	}))
	defer ts.Close()

	tmpDir := t.TempDir()
	cfg := &config.Config{ProjectRoot: tmpDir, CommandsDir: filepath.Join(tmpDir, "commands")}
	cmd := newInitCommandWithURLs(cfg, ts.URL, ts.URL)
	cmd.SetArgs([]string{"doesnotexist"})
	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for unknown template, got nil")
	}
}

func TestInitCommandNoArgs(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(TemplatesManifest{
			"laravel": {Description: "Laravel starter", Files: []string{"setup.yaml"}},
			"docker":  {Description: "Docker Compose", Files: []string{"up.yaml"}},
		})
	}))
	defer ts.Close()

	tmpDir := t.TempDir()
	cfg := &config.Config{ProjectRoot: tmpDir, CommandsDir: filepath.Join(tmpDir, "commands")}
	cmd := newInitCommandWithURLs(cfg, ts.URL, ts.URL)
	cmd.SetArgs([]string{})
	if err := cmd.Execute(); err != nil {
		t.Errorf("init with no args error = %v", err)
	}
}
