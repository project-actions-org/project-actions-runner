package parser

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Command represents a parsed command from YAML
type Command struct {
	Name    string                   `yaml:"-"` // Set separately, not from YAML
	Help    HelpMetadata             `yaml:"help"`
	Context string                   `yaml:"context,omitempty"`
	Steps   []map[string]interface{} `yaml:"steps"`
}

// HelpMetadata contains help information for a command
type HelpMetadata struct {
	Short string `yaml:"short"`
	Long  string `yaml:"long,omitempty"`
	Order int    `yaml:"order,omitempty"`
}

// Step represents a parsed step with its action and configuration
type Step struct {
	ActionName  string
	Config      map[string]interface{}
	Conditional *Conditional
}

// Conditional represents conditional execution logic
type Conditional struct {
	Type      string // "if-option", "if-missing", "if-fails", "if-no-option"
	Value     string // The condition value
	ThenSteps []Step // Steps to execute if condition is true
}

// ParseCommandFile reads and parses a command YAML file
func ParseCommandFile(filePath, commandName string) (*Command, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read command file: %w", err)
	}

	var cmd Command
	if err := yaml.Unmarshal(data, &cmd); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	cmd.Name = commandName
	return &cmd, nil
}

// ParseStep converts a raw YAML step into a Step struct
func ParseStep(raw map[string]interface{}) (*Step, error) {
	step := &Step{
		Config: make(map[string]interface{}),
	}

	// Check for conditionals first
	if ifOption, ok := raw["if-option"]; ok {
		step.Conditional = &Conditional{
			Type:  "if-option",
			Value: fmt.Sprint(ifOption),
		}
		// Parse the 'then' steps
		if thenSteps, ok := raw["then"].([]interface{}); ok {
			for _, ts := range thenSteps {
				if tsMap, ok := ts.(map[string]interface{}); ok {
					thenStep, err := ParseStep(tsMap)
					if err != nil {
						return nil, err
					}
					step.Conditional.ThenSteps = append(step.Conditional.ThenSteps, *thenStep)
				}
			}
		}
		return step, nil
	}

	if ifNoOption, ok := raw["if-no-option"]; ok {
		step.Conditional = &Conditional{
			Type:  "if-no-option",
			Value: fmt.Sprint(ifNoOption),
		}
		if thenSteps, ok := raw["then"].([]interface{}); ok {
			for _, ts := range thenSteps {
				if tsMap, ok := ts.(map[string]interface{}); ok {
					thenStep, err := ParseStep(tsMap)
					if err != nil {
						return nil, err
					}
					step.Conditional.ThenSteps = append(step.Conditional.ThenSteps, *thenStep)
				}
			}
		}
		return step, nil
	}

	if ifMissing, ok := raw["if-missing"]; ok {
		step.Conditional = &Conditional{
			Type:  "if-missing",
			Value: fmt.Sprint(ifMissing),
		}
		if thenSteps, ok := raw["then"].([]interface{}); ok {
			for _, ts := range thenSteps {
				if tsMap, ok := ts.(map[string]interface{}); ok {
					thenStep, err := ParseStep(tsMap)
					if err != nil {
						return nil, err
					}
					step.Conditional.ThenSteps = append(step.Conditional.ThenSteps, *thenStep)
				}
			}
		}
		return step, nil
	}

	if ifFails, ok := raw["if-fails"]; ok {
		step.Conditional = &Conditional{
			Type:  "if-fails",
			Value: fmt.Sprint(ifFails),
		}
		if thenSteps, ok := raw["then"].([]interface{}); ok {
			for _, ts := range thenSteps {
				if tsMap, ok := ts.(map[string]interface{}); ok {
					thenStep, err := ParseStep(tsMap)
					if err != nil {
						return nil, err
					}
					step.Conditional.ThenSteps = append(step.Conditional.ThenSteps, *thenStep)
				}
			}
		}
		return step, nil
	}

	// Not a conditional, determine action type
	// Check for standard action types: run, echo, command, action
	if runCmd, ok := raw["run"]; ok {
		step.ActionName = "run"
		step.Config["run"] = runCmd
		// Copy any other keys
		for k, v := range raw {
			if k != "run" {
				step.Config[k] = v
			}
		}
		return step, nil
	}

	if echoMsg, ok := raw["echo"]; ok {
		step.ActionName = "echo"
		step.Config["echo"] = echoMsg
		return step, nil
	}

	if cmdName, ok := raw["command"]; ok {
		step.ActionName = "command"
		step.Config["command"] = cmdName
		return step, nil
	}

	if actionName, ok := raw["action"]; ok {
		step.ActionName = fmt.Sprint(actionName)
		// Copy all config
		for k, v := range raw {
			if k != "action" {
				step.Config[k] = v
			}
		}
		return step, nil
	}

	// Check for check-for
	if checkFor, ok := raw["check-for"]; ok {
		step.ActionName = "check-for"
		step.Config["check-for"] = checkFor
		// Copy any other keys like if-missing message
		for k, v := range raw {
			if k != "check-for" {
				step.Config[k] = v
			}
		}
		return step, nil
	}

	// Unknown step type
	return nil, fmt.Errorf("unknown step type: %v", raw)
}
