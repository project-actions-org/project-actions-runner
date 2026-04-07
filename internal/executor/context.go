package executor

import (
	"fmt"
	"strings"

	"github.com/project-actions/runner/internal/actions"
	"github.com/project-actions/runner/internal/docker"
	"github.com/project-actions/runner/internal/parser"
)

// ContextManager handles execution context (inside vs outside container)
type ContextManager struct {
	engine *Engine
}

// NewContextManager creates a new context manager
func NewContextManager(engine *Engine) *ContextManager {
	return &ContextManager{
		engine: engine,
	}
}

// ParseContext parses the context string from command YAML
// Format: "outside-container" or "inside-container:service-name"
type Context struct {
	Type        string // "outside-container" or "inside-container"
	ServiceName string // Only for inside-container context
}

// ParseContextString parses a context string
func ParseContextString(contextStr string) *Context {
	if contextStr == "" {
		return &Context{Type: "outside-container"}
	}

	parts := strings.Split(contextStr, ":")
	ctx := &Context{
		Type: parts[0],
	}

	if len(parts) > 1 {
		ctx.ServiceName = parts[1]
	}

	return ctx
}

// CheckContext verifies if we're in the correct context for the command
func (cm *ContextManager) CheckContext(cmd *parser.Command, execCtx *actions.ExecutionContext) error {
	if cmd.Context == "" {
		// No context specified, can run anywhere
		return nil
	}

	targetContext := ParseContextString(cmd.Context)
	insideContainer := docker.IsInsideContainer()

	switch targetContext.Type {
	case "outside-container":
		if insideContainer {
			cm.engine.logger.Warn("Command '%s' should run outside container, but we're inside one", cmd.Name)
			// For now, just warn - could implement re-execution later
		}

	case "inside-container":
		// Always set ServiceName so run: steps can route via docker compose exec
		// when the runner is outside the container.
		execCtx.ServiceName = targetContext.ServiceName
		if insideContainer {
			execCtx.ContainerMode = true
		}

	default:
		return fmt.Errorf("unknown context type: %s", targetContext.Type)
	}

	return nil
}

// ShouldReexecute determines if the command needs to be re-executed in different context
// This is a placeholder for future implementation
func (cm *ContextManager) ShouldReexecute(cmd *parser.Command) bool {
	if cmd.Context == "" {
		return false
	}

	targetContext := ParseContextString(cmd.Context)
	insideContainer := docker.IsInsideContainer()

	switch targetContext.Type {
	case "outside-container":
		return insideContainer
	case "inside-container":
		return !insideContainer
	default:
		return false
	}
}

// Reexecute re-executes the current command in the target context
// This is a placeholder for future implementation
func (cm *ContextManager) Reexecute(cmd *parser.Command, args []string) error {
	targetContext := ParseContextString(cmd.Context)

	switch targetContext.Type {
	case "inside-container":
		// Would implement docker-compose exec here to re-run the command
		return fmt.Errorf("context switching to inside-container not yet implemented")
	case "outside-container":
		// Would need to exec outside the container (tricky)
		return fmt.Errorf("context switching to outside-container not yet implemented")
	default:
		return fmt.Errorf("unknown context type: %s", targetContext.Type)
	}
}
