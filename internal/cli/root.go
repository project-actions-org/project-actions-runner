package cli

import (
	"fmt"
	"os"

	"github.com/project-actions/runner/internal/config"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:     "project",
	Short:   "Project Actions - Run workflows from YAML",
	Version: GetVersion(),
	Long: `Project Actions is a flexible workflow system for running commands
defined in YAML files. It allows you to create simple or complex workflows
for your projects, similar to GitHub Actions but running locally.`,
}

// Execute runs the root command
func Execute() error {
	// Get script name from environment (set by wrapper script)
	scriptName := os.Getenv("PROJECT_SCRIPT_NAME")
	if scriptName != "" {
		rootCmd.Use = scriptName
	}

	// Load configuration to find .project directory
	cfg, err := config.LoadConfig()
	if err != nil {
		// If no .project directory found, show error and exit
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		fmt.Fprintf(os.Stderr, "\nProject Actions requires a .project directory in the current directory or a parent directory.\n")
		os.Exit(1)
	}

	// Register all project commands dynamically
	if err := RegisterProjectCommands(rootCmd, cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error registering commands: %v\n", err)
		os.Exit(1)
	}

	// Register built-in management commands
	rootCmd.AddCommand(newActionsCommand(cfg))
	rootCmd.AddCommand(newInitCommand(cfg))

	// Customize help template
	rootCmd.SetHelpTemplate(helpTemplate)
	rootCmd.SetUsageTemplate(usageTemplate)

	// Set custom help function to preserve command order
	rootCmd.SetHelpFunc(customHelpFunc)

	return rootCmd.Execute()
}

// customHelpFunc is a custom help function that preserves command ordering
func customHelpFunc(cmd *cobra.Command, args []string) {
	// Use the ordered commands list that was built during registration
	// This preserves the YAML order field
	projectCmds := orderedCommands

	// Get global commands (completion, help, etc)
	var globalCmds []*cobra.Command
	for _, c := range cmd.Commands() {
		if c.Annotations == nil || c.Annotations["project-command"] != "true" {
			if c.IsAvailableCommand() || c.Name() == "help" {
				globalCmds = append(globalCmds, c)
			}
		}
	}

	// Print custom help
	fmt.Fprintf(cmd.OutOrStdout(), "Project Actions v%s\n\n", GetVersion())
	fmt.Fprintf(cmd.OutOrStdout(), "Usage:\n")
	if cmd.Runnable() {
		fmt.Fprintf(cmd.OutOrStdout(), "  %s\n", cmd.UseLine())
	}
	if cmd.HasAvailableSubCommands() {
		fmt.Fprintf(cmd.OutOrStdout(), "  %s [command]\n", cmd.CommandPath())
	}
	fmt.Fprintf(cmd.OutOrStdout(), "\n")

	// Print project commands
	if len(projectCmds) > 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "Available Commands:\n")
		for _, c := range projectCmds {
			fmt.Fprintf(cmd.OutOrStdout(), "  %-20s %s\n", c.Name(), c.Short)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "\n")
	}

	// Print global commands
	if len(globalCmds) > 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "Global Commands:\n")
		for _, c := range globalCmds {
			fmt.Fprintf(cmd.OutOrStdout(), "  %-20s %s\n", c.Name(), c.Short)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "\n")
	}

	// Print flags
	if cmd.HasAvailableLocalFlags() {
		fmt.Fprintf(cmd.OutOrStdout(), "Flags:\n")
		fmt.Fprintf(cmd.OutOrStdout(), "%s\n", cmd.LocalFlags().FlagUsages())
	}

	if cmd.HasAvailableSubCommands() {
		fmt.Fprintf(cmd.OutOrStdout(), "Use \"%s [command] --help\" for more information about a command.\n", cmd.CommandPath())
	}
}

const helpTemplate = `Project Actions v{{.Version}}

Usage:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}

Available Commands:{{range .Commands}}{{if and (or .IsAvailableCommand (eq .Name "help")) .Annotations}}{{if index .Annotations "project-command"}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}

Global Commands:{{range .Commands}}{{if and (or .IsAvailableCommand (eq .Name "help")) (not (index .Annotations "project-command"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`

const usageTemplate = `Usage:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}{{$cmds := .Commands}}{{if eq (len .Groups) 0}}

Available Commands:{{range $cmds}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{else}}{{range $group := .Groups}}

{{.Title}}{{range $cmds}}{{if (and (eq .GroupID $group.ID) (or .IsAvailableCommand (eq .Name "help")))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if not .AllChildCommandsHaveGroup}}

Additional Commands:{{range $cmds}}{{if (and (eq .GroupID "") (or .IsAvailableCommand (eq .Name "help")))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`
