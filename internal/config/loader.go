package config

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Config holds the project configuration paths
type Config struct {
	ProjectRoot string
	ProjectDir  string
	CommandsDir string
	RuntimeDir  string
}

// CommandEntry represents a discovered command file with namespace information.
type CommandEntry struct {
	Name      string   // "setup", "build:all", "build:docker:image"
	FilePath  string   // absolute path to the YAML file
	Namespace []string // [] for root, ["build"] or ["build","docker"] for nested
}

// LoadConfig finds and loads the .project directory configuration
func LoadConfig() (*Config, error) {
	dir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current directory: %w", err)
	}

	for {
		projectDir := filepath.Join(dir, ".project")
		if _, err := os.Stat(projectDir); err == nil {
			config := &Config{
				ProjectRoot: dir,
				ProjectDir:  projectDir,
			}

			commandsDir := filepath.Join(projectDir, "commands")
			if _, err := os.Stat(commandsDir); err == nil {
				config.CommandsDir = commandsDir
			} else {
				config.CommandsDir = projectDir
			}

			config.RuntimeDir = filepath.Join(projectDir, ".runtime")
			return config, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return nil, fmt.Errorf(".project directory not found in current directory or any parent directory")
}

// FindCommandFile locates a command YAML file by name.
// Accepts colon-separated namespaced names: "build:all" resolves to build/all.yaml.
func (c *Config) FindCommandFile(commandName string) (string, error) {
	parts := strings.Split(commandName, ":")
	subPath := filepath.Join(parts...)

	yamlPath := filepath.Join(c.CommandsDir, subPath+".yaml")
	if _, err := os.Stat(yamlPath); err == nil {
		return yamlPath, nil
	}

	ymlPath := filepath.Join(c.CommandsDir, subPath+".yml")
	if _, err := os.Stat(ymlPath); err == nil {
		return ymlPath, nil
	}

	return "", fmt.Errorf("command '%s' not found in %s", commandName, c.CommandsDir)
}

// ListCommands recursively discovers all YAML command files under CommandsDir.
// Files in subdirectories become namespaced commands: build/all.yaml -> "build:all".
func (c *Config) ListCommands() ([]CommandEntry, error) {
	var entries []CommandEntry

	err := filepath.WalkDir(c.CommandsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		ext := filepath.Ext(d.Name())
		if ext != ".yaml" && ext != ".yml" {
			return nil
		}

		rel, err := filepath.Rel(c.CommandsDir, path)
		if err != nil {
			return err
		}

		// Convert OS path separators to forward slash for consistent splitting
		rel = filepath.ToSlash(rel)
		parts := strings.Split(rel, "/")

		// Strip extension from the last segment (the filename)
		last := parts[len(parts)-1]
		parts[len(parts)-1] = last[:len(last)-len(ext)]

		name := strings.Join(parts, ":")

		var namespace []string
		if len(parts) > 1 {
			namespace = make([]string, len(parts)-1)
			copy(namespace, parts[:len(parts)-1])
		}

		entries = append(entries, CommandEntry{
			Name:      name,
			FilePath:  path,
			Namespace: namespace,
		})
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to scan commands directory: %w", err)
	}

	return entries, nil
}
