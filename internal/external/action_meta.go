package external

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// InputSpec describes a single input parameter for an action.
type InputSpec struct {
	Description string `yaml:"description"`
	Required    bool   `yaml:"required"`
	Default     string `yaml:"default"`
}

// ActionMeta holds the parsed contents of an action's action.yaml file.
type ActionMeta struct {
	Name        string               `yaml:"name"`
	Description string               `yaml:"description"`
	Inputs      map[string]InputSpec `yaml:"inputs"`
}

// ParseActionMeta reads and parses an action.yaml file.
func ParseActionMeta(path string) (*ActionMeta, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read action.yaml: %w", err)
	}

	var meta ActionMeta
	if err := yaml.Unmarshal(data, &meta); err != nil {
		return nil, fmt.Errorf("failed to parse action.yaml: %w", err)
	}

	return &meta, nil
}

// ValidateAndApplyDefaults checks that all required inputs are present in the
// provided with map, applies defaults for missing optional inputs, and returns
// the resolved inputs. The original map is not modified.
func (m *ActionMeta) ValidateAndApplyDefaults(with map[string]interface{}) (map[string]interface{}, error) {
	resolved := make(map[string]interface{})

	// Copy provided values
	for k, v := range with {
		resolved[k] = v
	}

	// Check required and apply defaults
	for name, spec := range m.Inputs {
		if _, provided := resolved[name]; !provided {
			if spec.Required {
				return nil, fmt.Errorf("missing required input %q", name)
			}
			if spec.Default != "" {
				resolved[name] = spec.Default
			}
		}
	}

	return resolved, nil
}
