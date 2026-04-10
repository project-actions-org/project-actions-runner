package cli

import (
	"fmt"
	"sort"
	"strings"

	"github.com/project-actions/runner/internal/config"
	"github.com/project-actions/runner/internal/executor"
	"github.com/project-actions/runner/internal/logger"
	"github.com/project-actions/runner/internal/parser"
	"github.com/spf13/cobra"
)

// orderedCommands stores commands in registration order
var orderedCommands []*cobra.Command

// namespaceCommands maps namespace key (e.g. "build", "build:docker") to its
// commands in display order. Populated during RegisterProjectCommands.
var namespaceCommands map[string][]*cobra.Command

// RegisterProjectCommands dynamically registers all YAML commands as Cobra subcommands
func RegisterProjectCommands(rootCmd *cobra.Command, cfg *config.Config) error {
	// Reset package-level state (safe: called once per Execute)
	orderedCommands = nil
	namespaceCommands = make(map[string][]*cobra.Command)

	entries, err := cfg.ListCommands()
	if err != nil {
		return fmt.Errorf("failed to list commands: %w", err)
	}

	if err := validateSourceConsistency(entries); err != nil {
		return err
	}

	type cmdInfo struct {
		Entry config.CommandEntry
		Order int
		Cmd   *parser.Command
	}

	var commands []cmdInfo
	for _, entry := range entries {
		cmd, err := parser.ParseCommandFile(entry.FilePath, entry.Name)
		if err != nil {
			continue
		}
		commands = append(commands, cmdInfo{
			Entry: entry,
			Order: cmd.Help.Order,
			Cmd:   cmd,
		})
	}

	sort.Slice(commands, func(i, j int) bool {
		if commands[i].Order != commands[j].Order {
			return commands[i].Order < commands[j].Order
		}
		return commands[i].Entry.Name < commands[j].Entry.Name
	})

	// Register real commands; track which names are occupied
	registeredNames := make(map[string]bool)
	for _, info := range commands {
		cobraCmd := createDynamicCommand(info.Entry.Name, info.Cmd, cfg)
		rootCmd.AddCommand(cobraCmd)
		registeredNames[info.Entry.Name] = true

		if len(info.Entry.Namespace) == 0 {
			orderedCommands = append(orderedCommands, cobraCmd)
		} else {
			nsKey := strings.Join(info.Entry.Namespace, ":")
			namespaceCommands[nsKey] = append(namespaceCommands[nsKey], cobraCmd)
		}
	}

	// Collect all unique namespace paths that need synthetic commands.
	// For build:docker:image we need synthetics for "build" and "build:docker".
	syntheticNeeded := make(map[string]bool)
	for _, info := range commands {
		for depth := 1; depth <= len(info.Entry.Namespace); depth++ {
			nsKey := strings.Join(info.Entry.Namespace[:depth], ":")
			syntheticNeeded[nsKey] = true
		}
	}

	// Register synthetic namespace commands where no real command exists
	for nsKey := range syntheticNeeded {
		if registeredNames[nsKey] {
			continue
		}
		rootCmd.AddCommand(createNamespaceCommand(nsKey))
	}

	return nil
}

// createDynamicCommand creates a Cobra command from a parsed YAML command
func createDynamicCommand(name string, cmd *parser.Command, cfg *config.Config) *cobra.Command {
	use := name
	for _, p := range cmd.Params {
		use += " <" + p.Name + ">"
	}

	long := cmd.Help.Long
	if len(cmd.Params) > 0 {
		var sb strings.Builder
		if long != "" {
			sb.WriteString(long)
			sb.WriteString("\n\n")
		}
		sb.WriteString("Arguments:\n")
		for _, p := range cmd.Params {
			if p.Description != "" {
				sb.WriteString(fmt.Sprintf("  %-16s %s\n", p.Name, p.Description))
			} else {
				sb.WriteString(fmt.Sprintf("  %s\n", p.Name))
			}
		}
		long = sb.String()
	}

	cobraCmd := &cobra.Command{
		Use:                use,
		Short:              cmd.Help.Short,
		Long:               long,
		DisableFlagParsing: true, // Pass all args through as options
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			// Create logger
			log := logger.New()

			// Create and initialize engine
			engine := executor.NewEngine(cfg, log)

			// Execute the command
			if err := engine.ExecuteCommand(name, args); err != nil {
				return err
			}

			return nil
		},
	}

	// Mark this as a project command for help formatting
	cobraCmd.Annotations = map[string]string{
		"project-command": "true",
	}

	return cobraCmd
}

// createNamespaceCommand creates a synthetic Cobra command for a namespace
// directory that has no matching YAML file. Running it prints scoped help.
func createNamespaceCommand(nsKey string) *cobra.Command {
	cmd := &cobra.Command{
		Use:                nsKey + " [options...]",
		Short:              nsKey + " commands",
		DisableFlagParsing: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			printNamespaceHelp(cmd.CommandPath(), nsKey, cmd.OutOrStdout())
			return nil
		},
	}
	cmd.Annotations = map[string]string{
		"project-namespace": "true",
	}
	return cmd
}

// validateSourceConsistency collects sources from all command files and errors
// if the same source name is declared with different URLs or refs in different files.
func validateSourceConsistency(entries []config.CommandEntry) error {
	type sourceEntry struct {
		file   string
		rawURL string
	}
	seen := make(map[string]sourceEntry)

	for _, entry := range entries {
		cmd, err := parser.ParseCommandFile(entry.FilePath, entry.Name)
		if err != nil {
			return fmt.Errorf("failed to parse command file %s: %w", entry.FilePath, err)
		}
		for alias, rawURL := range cmd.Sources {
			if existing, conflict := seen[alias]; conflict {
				if existing.rawURL != rawURL {
					return fmt.Errorf(
						"conflicting source %q: %s declares %q but %s declares %q — all command files must agree on the same URL and ref",
						alias, existing.file, existing.rawURL, entry.FilePath, rawURL,
					)
				}
			} else {
				seen[alias] = sourceEntry{file: entry.FilePath, rawURL: rawURL}
			}
		}
	}
	return nil
}
