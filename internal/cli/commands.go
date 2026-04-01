package cli

import (
	"fmt"
	"sort"

	"github.com/project-actions/runner/internal/config"
	"github.com/project-actions/runner/internal/executor"
	"github.com/project-actions/runner/internal/logger"
	"github.com/project-actions/runner/internal/parser"
	"github.com/spf13/cobra"
)

// orderedCommands stores commands in registration order
var orderedCommands []*cobra.Command

// RegisterProjectCommands dynamically registers all YAML commands as Cobra subcommands
func RegisterProjectCommands(rootCmd *cobra.Command, cfg *config.Config) error {
	// Validate that sources are consistent across all command files
	if err := validateSourceConsistency(cfg); err != nil {
		return err
	}

	// Get list of all command files
	commandNames, err := cfg.ListCommands()
	if err != nil {
		return fmt.Errorf("failed to list commands: %w", err)
	}

	// Sort commands by order (if available) or alphabetically
	type cmdInfo struct {
		Name  string
		Order int
		Cmd   *parser.Command
	}

	var commands []cmdInfo
	for _, name := range commandNames {
		cmdFile, err := cfg.FindCommandFile(name)
		if err != nil {
			continue
		}

		cmd, err := parser.ParseCommandFile(cmdFile, name)
		if err != nil {
			// Skip commands that can't be parsed
			continue
		}

		commands = append(commands, cmdInfo{
			Name:  name,
			Order: cmd.Help.Order,
			Cmd:   cmd,
		})
	}

	// Sort by order, then alphabetically
	sort.Slice(commands, func(i, j int) bool {
		if commands[i].Order != commands[j].Order {
			return commands[i].Order < commands[j].Order
		}
		return commands[i].Name < commands[j].Name
	})

	// Register each command in sorted order
	for _, cmdInfo := range commands {
		cobraCmd := createDynamicCommand(cmdInfo.Name, cmdInfo.Cmd, cfg)
		rootCmd.AddCommand(cobraCmd)
		// Store in ordered list for help display
		orderedCommands = append(orderedCommands, cobraCmd)
	}

	return nil
}

// createDynamicCommand creates a Cobra command from a parsed YAML command
func createDynamicCommand(name string, cmd *parser.Command, cfg *config.Config) *cobra.Command {
	cobraCmd := &cobra.Command{
		Use:                name + " [options...]",
		Short:              cmd.Help.Short,
		Long:               cmd.Help.Long,
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

// validateSourceConsistency collects sources from all command files and errors
// if the same source name is declared with different URLs or refs in different files.
func validateSourceConsistency(cfg *config.Config) error {
	commandNames, err := cfg.ListCommands()
	if err != nil {
		return fmt.Errorf("failed to list commands: %w", err)
	}

	// Map from source alias to "file:rawURL" for error reporting
	type sourceEntry struct {
		file   string
		rawURL string
	}
	seen := make(map[string]sourceEntry)

	for _, name := range commandNames {
		cmdFile, err := cfg.FindCommandFile(name)
		if err != nil {
			continue
		}
		cmd, err := parser.ParseCommandFile(cmdFile, name)
		if err != nil {
			continue
		}
		for alias, rawURL := range cmd.Sources {
			if existing, conflict := seen[alias]; conflict {
				if existing.rawURL != rawURL {
					return fmt.Errorf(
						"conflicting source %q: %s declares %q but %s declares %q — all command files must agree on the same URL and ref",
						alias, existing.file, existing.rawURL, cmdFile, rawURL,
					)
				}
			} else {
				seen[alias] = sourceEntry{file: cmdFile, rawURL: rawURL}
			}
		}
	}
	return nil
}
