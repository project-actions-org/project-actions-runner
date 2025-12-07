package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// Config holds the project configuration paths
type Config struct {
	ProjectRoot string
	ProjectDir  string
	CommandsDir string
	RuntimeDir  string
}

// LoadConfig finds and loads the .project directory configuration
func LoadConfig() (*Config, error) {
	// Start from current directory and walk up to find .project
	dir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current directory: %w", err)
	}

	// Walk up the directory tree
	for {
		projectDir := filepath.Join(dir, ".project")
		if _, err := os.Stat(projectDir); err == nil {
			// Found .project directory
			config := &Config{
				ProjectRoot: dir,
				ProjectDir:  projectDir,
			}

			// Check for commands in both locations:
			// 1. .project/commands/ (preferred, nested structure)
			// 2. .project/ (flat structure, for backwards compatibility)
			commandsDir := filepath.Join(projectDir, "commands")
			if _, err := os.Stat(commandsDir); err == nil {
				config.CommandsDir = commandsDir
			} else {
				config.CommandsDir = projectDir
			}

			config.RuntimeDir = filepath.Join(projectDir, ".runtime")

			return config, nil
		}

		// Move up one directory
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root without finding .project
			break
		}
		dir = parent
	}

	return nil, fmt.Errorf(".project directory not found in current directory or any parent directory")
}

// FindCommandFile locates a command YAML file by name
func (c *Config) FindCommandFile(commandName string) (string, error) {
	// Try with .yaml extension
	yamlPath := filepath.Join(c.CommandsDir, commandName+".yaml")
	if _, err := os.Stat(yamlPath); err == nil {
		return yamlPath, nil
	}

	// Try with .yml extension
	ymlPath := filepath.Join(c.CommandsDir, commandName+".yml")
	if _, err := os.Stat(ymlPath); err == nil {
		return ymlPath, nil
	}

	return "", fmt.Errorf("command '%s' not found in %s", commandName, c.CommandsDir)
}

// ListCommands returns all available command files
func (c *Config) ListCommands() ([]string, error) {
	entries, err := os.ReadDir(c.CommandsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read commands directory: %w", err)
	}

	var commands []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		// Check for .yaml or .yml extension
		if filepath.Ext(name) == ".yaml" || filepath.Ext(name) == ".yml" {
			// Strip extension to get command name
			cmdName := name[:len(name)-len(filepath.Ext(name))]
			commands = append(commands, cmdName)
		}
	}

	return commands, nil
}
