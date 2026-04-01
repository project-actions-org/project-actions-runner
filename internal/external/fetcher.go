package external

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Fetcher handles cloning action source repos and downloading platform binaries.
type Fetcher struct {
	ActionsDir string // absolute path to .project/.runtime/actions/
}

// SourceDir returns the local directory for a fetched source.
func (f *Fetcher) SourceDir(sourceName string) string {
	return filepath.Join(f.ActionsDir, sourceName)
}

// ActionDir returns the local directory for a specific action within a source.
func (f *Fetcher) ActionDir(sourceName, actionName string) string {
	return filepath.Join(f.ActionsDir, sourceName, actionName)
}

// BinaryPath returns the expected path of a platform binary in the action dir.
func (f *Fetcher) BinaryPath(sourceName, actionName, goos, goarch string) string {
	binaryName := fmt.Sprintf("%s-%s-%s", actionName, goos, goarch)
	return filepath.Join(f.ActionDir(sourceName, actionName), binaryName)
}

// IsFetched returns true if the source repo has already been cloned.
func (f *Fetcher) IsFetched(sourceName string) bool {
	_, err := os.Stat(f.SourceDir(sourceName))
	return err == nil
}

// FetchSource performs a shallow git clone of the source repo into the cache dir.
func (f *Fetcher) FetchSource(sourceName string, source *Source) error {
	destDir := f.SourceDir(sourceName)

	if err := os.MkdirAll(filepath.Dir(destDir), 0755); err != nil {
		return fmt.Errorf("failed to create actions cache directory: %w", err)
	}

	cmd := exec.Command("git", "clone", "--depth", "1", "--branch", source.Ref, source.CloneURL(), destDir)
	cmd.Stdout = os.Stderr // progress to stderr
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to fetch source %q from %s: %w", sourceName, source.CloneURL(), err)
	}

	return nil
}

// DownloadBinary downloads the platform binary for an action from GitHub Releases.
func (f *Fetcher) DownloadBinary(sourceName, actionName string, source *Source, goos, goarch string) (string, error) {
	url := source.BinaryURL(actionName, goos, goarch)
	destPath := f.BinaryPath(sourceName, actionName, goos, goarch)

	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return "", fmt.Errorf("failed to create directory for binary: %w", err)
	}

	resp, err := http.Get(url) //nolint:gosec
	if err != nil {
		return "", fmt.Errorf("failed to download binary from %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("binary not found at %s (HTTP %d) — ensure the release tag %q exists and includes a %s-%s asset",
			url, resp.StatusCode, source.Ref, goos, goarch)
	}

	file, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return "", fmt.Errorf("failed to create binary file: %w", err)
	}
	defer file.Close()

	if _, err := io.Copy(file, resp.Body); err != nil {
		return "", fmt.Errorf("failed to write binary: %w", err)
	}

	return destPath, nil
}

// EnsureGitignore adds .project/.runtime/actions/ to the project .gitignore
// if not already present. Creates the .gitignore file if it does not exist.
func (f *Fetcher) EnsureGitignore(projectRoot string) error {
	const entry = ".project/.runtime/actions/"
	gitignorePath := filepath.Join(projectRoot, ".gitignore")

	// Read existing content
	existing, err := os.ReadFile(gitignorePath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to read .gitignore: %w", err)
	}

	// Check if entry already present
	for _, line := range strings.Split(string(existing), "\n") {
		if strings.TrimSpace(line) == entry {
			return nil
		}
	}

	// Append entry
	file, err := os.OpenFile(gitignorePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open .gitignore: %w", err)
	}
	defer file.Close()

	prefix := ""
	if len(existing) > 0 && !strings.HasSuffix(string(existing), "\n") {
		prefix = "\n"
	}

	_, err = fmt.Fprintf(file, "%s%s\n", prefix, entry)
	return err
}
