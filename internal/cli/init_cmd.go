package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/project-actions/runner/internal/config"
	"github.com/spf13/cobra"
)

const templatesManifestURL = "https://raw.githubusercontent.com/project-actions-org/templates/master/manifest.json"
const templatesBaseURL = "https://raw.githubusercontent.com/project-actions-org/templates/master"

// httpClient is used for all template downloads with a sensible timeout.
var httpClient = &http.Client{Timeout: 30 * time.Second}

// TemplateEntry describes a single template in the manifest.
type TemplateEntry struct {
	Description string   `json:"description"`
	Files       []string `json:"files"`
}

// TemplatesManifest is the parsed manifest.json from the templates repo.
type TemplatesManifest map[string]TemplateEntry

// newInitCommand creates the "project init" command using production GitHub URLs.
func newInitCommand(cfg *config.Config) *cobra.Command {
	return newInitCommandWithURLs(cfg, templatesManifestURL, templatesBaseURL)
}

// newInitCommandWithURLs creates the command with configurable URLs (used for testing).
func newInitCommandWithURLs(cfg *config.Config, manifestURL, baseURL string) *cobra.Command {
	return &cobra.Command{
		Use:   "init [template...]",
		Short: "Add starter commands from a template",
		Long: `Download starter command files from the project-actions templates repository.

Run without arguments to list available templates.

Examples:
  ./project init laravel
  ./project init django docker
  ./project init`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			manifest, err := fetchManifest(manifestURL)
			if err != nil {
				fmt.Fprintln(cmd.ErrOrStderr(), "Check your internet connection.")
				return fmt.Errorf("failed to fetch template manifest: %w", err)
			}

			if len(args) == 0 {
				printTemplateManifest(manifest)
				return nil
			}

			// Validate all template names before downloading anything
			for _, name := range args {
				if _, ok := manifest[name]; !ok {
					return fmt.Errorf("unknown template %q — run ./project init to see available templates", name)
				}
			}

			if err := os.MkdirAll(cfg.CommandsDir, 0755); err != nil {
				return fmt.Errorf("failed to create commands directory: %w", err)
			}

			for _, name := range args {
				entry := manifest[name]
				fmt.Printf("Adding %s template...\n", name)
				for _, file := range entry.Files {
					destPath := filepath.Join(cfg.CommandsDir, file)
					if _, err := os.Stat(destPath); err == nil {
						fmt.Printf("  — %s already exists, skipping\n", destPath)
						continue
					}
					fileURL := fmt.Sprintf("%s/%s/%s", baseURL, name, file)
					if err := downloadFile(fileURL, destPath); err != nil {
						return fmt.Errorf("failed to download %s for template %q: %w", file, name, err)
					}
					fmt.Printf("  ✓ %s\n", destPath)
				}
			}

			fmt.Printf("\nDone. Run ./project to see available commands.\n")
			return nil
		},
	}
}

// fetchManifest downloads and parses the templates manifest JSON.
func fetchManifest(url string) (TemplatesManifest, error) {
	resp, err := httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer func() {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned %d", resp.StatusCode)
	}
	var manifest TemplatesManifest
	if err := json.NewDecoder(resp.Body).Decode(&manifest); err != nil {
		return nil, fmt.Errorf("invalid manifest JSON: %w", err)
	}
	return manifest, nil
}

// downloadFile fetches a URL and writes the response body to destPath.
func downloadFile(url, destPath string) error {
	resp, err := httpClient.Get(url)
	if err != nil {
		return err
	}
	defer func() {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned %d", resp.StatusCode)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return os.WriteFile(destPath, data, 0644)
}

// printTemplateManifest prints available templates sorted alphabetically.
func printTemplateManifest(manifest TemplatesManifest) {
	fmt.Print("Available templates:\n\n")
	names := make([]string, 0, len(manifest))
	for name := range manifest {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		fmt.Printf("  %-12s %s\n", name, manifest[name].Description)
	}
	fmt.Println("\nUsage: ./project init <template> [template...]")
}
