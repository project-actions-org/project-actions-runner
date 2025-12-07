package actions

import (
	"fmt"

	"github.com/project-actions/runner/internal/logger"
)

// Registry manages available actions
type Registry struct {
	actions map[string]Action
	logger  *logger.Logger
}

// NewRegistry creates a new action registry
func NewRegistry(log *logger.Logger) *Registry {
	r := &Registry{
		actions: make(map[string]Action),
		logger:  log,
	}

	return r
}

// Register adds an action to the registry
func (r *Registry) Register(name string, action Action) {
	r.actions[name] = action
}

// Get retrieves an action by name
func (r *Registry) Get(name string) (Action, error) {
	action, exists := r.actions[name]
	if !exists {
		return nil, fmt.Errorf("action '%s' not found", name)
	}
	return action, nil
}

// List returns all registered action names
func (r *Registry) List() []string {
	names := make([]string, 0, len(r.actions))
	for name := range r.actions {
		names = append(names, name)
	}
	return names
}
