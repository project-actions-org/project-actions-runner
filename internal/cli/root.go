package cli

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/project-actions/runner/internal/config"
	"github.com/spf13/cobra"
)

// ANSI color codes
const (
	colorReset  = "\033[0m"
	colorBold   = "\033[1m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
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

	// Customize usage template (used for error output)
	rootCmd.SetUsageTemplate(usageTemplate)

	// Set custom help function to preserve command order
	rootCmd.SetHelpFunc(customHelpFunc)

	return rootCmd.Execute()
}

func customHelpFunc(cmd *cobra.Command, args []string) {
	out := cmd.OutOrStdout()

	// Separate global commands from project/namespace commands.
	// Synthetic namespace commands are excluded from all sections.
	var globalCmds []*cobra.Command
	for _, c := range cmd.Commands() {
		ann := c.Annotations
		if ann != nil && (ann["project-command"] == "true" || ann["project-namespace"] == "true") {
			continue
		}
		if c.IsAvailableCommand() || c.Name() == "help" {
			globalCmds = append(globalCmds, c)
		}
	}

	// Namespace sections sorted alphabetically by key
	var nsKeys []string
	for k := range namespaceCommands {
		nsKeys = append(nsKeys, k)
	}
	sort.Strings(nsKeys)

	fmt.Fprintf(out, "%sProject Actions%s v%s\n\n", colorBold, colorReset, GetVersion())
	fmt.Fprintf(out, "%sUsage:%s\n", colorBold, colorReset)
	if cmd.Runnable() {
		fmt.Fprintf(out, "  %s\n", cmd.UseLine())
	}
	if cmd.HasAvailableSubCommands() {
		fmt.Fprintf(out, "  %s [command]\n", cmd.CommandPath())
	}
	fmt.Fprintf(out, "\n")

	if len(orderedCommands) > 0 {
		fmt.Fprintf(out, "%sAvailable Commands:%s\n", colorYellow, colorReset)
		for _, c := range orderedCommands {
			fmt.Fprintf(out, "  %s%-20s%s %s\n", colorGreen, c.Name(), colorReset, c.Short)
		}
		fmt.Fprintf(out, "\n")
	}

	for _, ns := range nsKeys {
		fmt.Fprintf(out, "%s[%s]%s\n", colorCyan, ns, colorReset)
		for _, c := range namespaceCommands[ns] {
			fmt.Fprintf(out, "  %s%-20s%s %s\n", colorGreen, c.Name(), colorReset, c.Short)
		}
		fmt.Fprintf(out, "\n")
	}

	if len(globalCmds) > 0 {
		fmt.Fprintf(out, "%sGlobal Commands:%s\n", colorYellow, colorReset)
		for _, c := range globalCmds {
			fmt.Fprintf(out, "  %s%-20s%s %s\n", colorGreen, c.Name(), colorReset, c.Short)
		}
		fmt.Fprintf(out, "\n")
	}

	if cmd.HasAvailableLocalFlags() {
		fmt.Fprintf(out, "%sFlags:%s\n", colorYellow, colorReset)
		fmt.Fprintf(out, "%s\n", cmd.LocalFlags().FlagUsages())
	}

	if cmd.HasAvailableSubCommands() {
		fmt.Fprintf(out, "Use \"%s [command] --help\" for more information about a command.\n", cmd.CommandPath())
	}
}

// printNamespaceHelp prints a scoped help view for a synthetic namespace command.
// It shows all namespace sections whose key starts with the given namespace prefix.
// namespace must be a non-empty colon-separated string (e.g. "build" or "build:docker").
func printNamespaceHelp(commandPath string, namespace string, w io.Writer) {
	nsParts := strings.Split(namespace, ":")

	var relevantKeys []string
	for k := range namespaceCommands {
		kParts := strings.Split(k, ":")
		if len(kParts) < len(nsParts) {
			continue
		}
		match := true
		for i, p := range nsParts {
			if kParts[i] != p {
				match = false
				break
			}
		}
		if match {
			relevantKeys = append(relevantKeys, k)
		}
	}
	sort.Strings(relevantKeys)

	fmt.Fprintf(w, "%sUsage:%s\n  %s:<command>\n\n", colorBold, colorReset, commandPath)
	for _, k := range relevantKeys {
		fmt.Fprintf(w, "%s[%s]%s\n", colorCyan, k, colorReset)
		for _, c := range namespaceCommands[k] {
			fmt.Fprintf(w, "  %s%-20s%s %s\n", colorGreen, c.Name(), colorReset, c.Short)
		}
		fmt.Fprintf(w, "\n")
	}
}

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
