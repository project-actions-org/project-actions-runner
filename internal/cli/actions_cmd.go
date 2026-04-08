package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/project-actions/runner/internal/config"
	"github.com/project-actions/runner/internal/external"
	"github.com/project-actions/runner/internal/parser"
	"github.com/spf13/cobra"
)

// newActionsCommand creates the "project actions" parent command.
func newActionsCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "actions",
		Short: "Manage external action sources",
	}
	cmd.AddCommand(newActionsListCommand(cfg))
	cmd.AddCommand(newActionsUpdateCommand(cfg))
	cmd.AddCommand(newActionsInfoCommand(cfg))
	return cmd
}

// newActionsListCommand shows all declared sources and their cache status.
func newActionsListCommand(cfg *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List declared action sources and their cache status",
		RunE: func(cmd *cobra.Command, args []string) error {
			sources, err := collectAllSources(cfg)
			if err != nil {
				return err
			}
			if len(sources) == 0 {
				fmt.Println("No action sources declared. Add a sources: block to your command files.")
				return nil
			}

			// Get sorted keys for deterministic output
			aliases := make([]string, 0, len(sources))
			for alias := range sources {
				aliases = append(aliases, alias)
			}
			sort.Strings(aliases)

			fetcher := &external.Fetcher{ActionsDir: filepath.Join(cfg.RuntimeDir, "actions")}
			fmt.Print("Sources (from command files):\n\n")
			for _, alias := range aliases {
				rawURL := sources[alias]
				status := "not fetched"
				if fetcher.IsFetched(alias) {
					status = "cached  ✓"
				}
				fmt.Printf("  %-12s %-55s %s\n", alias, rawURL, status)
			}
			return nil
		},
	}
}

// newActionsUpdateCommand re-fetches one or all action sources.
func newActionsUpdateCommand(cfg *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "update [source]",
		Short: "Re-fetch action sources (all, or a specific one)",
		RunE: func(cmd *cobra.Command, args []string) error {
			sources, err := collectAllSources(cfg)
			if err != nil {
				return err
			}

			fetcher := &external.Fetcher{ActionsDir: filepath.Join(cfg.RuntimeDir, "actions")}

			toUpdate := sources
			if len(args) == 1 {
				rawURL, ok := sources[args[0]]
				if !ok {
					return fmt.Errorf("source %q not declared in any command file", args[0])
				}
				toUpdate = map[string]string{args[0]: rawURL}
			}

			// Sort for deterministic output
			aliases := make([]string, 0, len(toUpdate))
			for alias := range toUpdate {
				aliases = append(aliases, alias)
			}
			sort.Strings(aliases)

			for _, alias := range aliases {
				rawURL := toUpdate[alias]
				src, err := external.ParseSourceURL(rawURL)
				if err != nil {
					return fmt.Errorf("invalid source URL for %q: %w", alias, err)
				}
				src.Name = alias

				// Remove existing cache
				cacheDir := fetcher.SourceDir(alias)
				if err := os.RemoveAll(cacheDir); err != nil {
					return fmt.Errorf("failed to remove cached source %q: %w", alias, err)
				}

				fmt.Printf("Fetching %s from %s...\n", alias, src.CloneURL())
				if err := fetcher.FetchSource(alias, src); err != nil {
					return err
				}
				fmt.Printf("  ✓ %s updated\n", alias)
			}
			return nil
		},
	}
}

// newActionsInfoCommand prints metadata for a specific action.
func newActionsInfoCommand(cfg *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "info source/action",
		Short: "Show metadata and inputs for an action",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			parts := strings.SplitN(args[0], "/", 2)
			if len(parts) != 2 {
				return fmt.Errorf("argument must be source/action-name (e.g. aws/iam-role-setup)")
			}
			sourceName, actionName := parts[0], parts[1]

			sources, err := collectAllSources(cfg)
			if err != nil {
				return err
			}
			rawURL, ok := sources[sourceName]
			if !ok {
				return fmt.Errorf("source %q not declared in any command file", sourceName)
			}

			fetcher := &external.Fetcher{ActionsDir: filepath.Join(cfg.RuntimeDir, "actions")}
			if !fetcher.IsFetched(sourceName) {
				return fmt.Errorf("source %q is not fetched — run: project actions update %s", sourceName, sourceName)
			}

			metaPath := filepath.Join(fetcher.ActionDir(sourceName, actionName), "action.yaml")
			meta, err := external.ParseActionMeta(metaPath)
			if err != nil {
				return fmt.Errorf("failed to read action.yaml for %s/%s: %w", sourceName, actionName, err)
			}

			fmt.Printf("Action: %s/%s\n", sourceName, actionName)
			fmt.Printf("Source: %s\n", rawURL)
			fmt.Printf("Name:   %s\n", meta.Name)
			if meta.Description != "" {
				fmt.Printf("Desc:   %s\n", meta.Description)
			}
			if len(meta.Inputs) > 0 {
				fmt.Println("\nInputs:")
				for name, spec := range meta.Inputs {
					req := "optional"
					if spec.Required {
						req = "required"
					}
					defaultStr := ""
					if spec.Default != "" {
						defaultStr = fmt.Sprintf(" (default: %s)", spec.Default)
					}
					fmt.Printf("  %-20s %s%s\n", name, req, defaultStr)
					if spec.Description != "" {
						fmt.Printf("  %-20s   %s\n", "", spec.Description)
					}
				}
			}
			return nil
		},
	}
}

// collectAllSources reads all command files and returns a deduplicated source map.
func collectAllSources(cfg *config.Config) (map[string]string, error) {
	entries, err := cfg.ListCommands()
	if err != nil {
		return nil, fmt.Errorf("failed to list commands: %w", err)
	}

	sources := make(map[string]string)
	for _, entry := range entries {
		cmd, err := parser.ParseCommandFile(entry.FilePath, entry.Name)
		if err != nil {
			continue
		}
		for alias, rawURL := range cmd.Sources {
			sources[alias] = rawURL
		}
	}
	return sources, nil
}
