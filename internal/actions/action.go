package actions

import (
	"github.com/project-actions/runner/internal/config"
	"github.com/project-actions/runner/internal/logger"
)

// ExecutionContext holds the .runtime context for executing steps
type ExecutionContext struct {
	WorkingDir    string
	Environment   map[string]string
	Options       map[string]string
	ContainerMode bool
	ServiceName   string
	Logger        *logger.Logger
	Config        *config.Config
	Verbose       bool // Show subprocess output
	Sources       map[string]string // from command file sources: front-matter
	CommandName   string            // name of the running command
}

// Action represents an executable action
type Action interface {
	Execute(ctx *ExecutionContext, config map[string]interface{}) error
	Validate(config map[string]interface{}) error
}
