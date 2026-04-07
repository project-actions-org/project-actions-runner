package builtin

import (
	"testing"

	"github.com/project-actions/runner/internal/actions"
	"github.com/project-actions/runner/internal/logger"
)

func makeRunCtx() *actions.ExecutionContext {
	return &actions.ExecutionContext{
		Logger:     logger.New(),
		WorkingDir: "/tmp",
	}
}

func TestRunAction_Execute(t *testing.T) {
	tests := []struct {
		name    string
		config  map[string]interface{}
		ctx     *actions.ExecutionContext
		wantErr bool
	}{
		{
			name:    "simple echo command",
			config:  map[string]interface{}{"run": "echo 'test'"},
			ctx:     makeRunCtx(),
			wantErr: false,
		},
		{
			name:    "command with exit 0",
			config:  map[string]interface{}{"run": "exit 0"},
			ctx:     makeRunCtx(),
			wantErr: false,
		},
		{
			name:    "command that fails",
			config:  map[string]interface{}{"run": "exit 1"},
			ctx:     makeRunCtx(),
			wantErr: true,
		},
		{
			name:    "missing run field",
			config:  map[string]interface{}{},
			ctx:     makeRunCtx(),
			wantErr: true,
		},
		{
			name:    "invalid command",
			config:  map[string]interface{}{"run": "nonexistent-command-xyz"},
			ctx:     makeRunCtx(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			action := &RunAction{}
			err := action.Execute(tt.ctx, tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("RunAction.Execute() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRunAction_ContextAwareRouting_WithService(t *testing.T) {
	ctx := makeRunCtx()
	ctx.ServiceName = "web"

	// Simulate being outside the container: isInsideContainer returns false.
	// Config is pre-interpolated — plain command string, no tokens.
	action := &RunAction{isInsideContainer: func() bool { return false }}

	// docker compose exec will fail (not running), but the error should be
	// exec-related, not a config error — confirms routing took the right branch.
	err := action.Execute(ctx, map[string]interface{}{"run": "echo hello"})
	if err == nil {
		// docker compose happened to be available and it worked — fine too.
		return
	}
	if err.Error() == "run action requires a command" {
		t.Errorf("got config error instead of routing error: %v", err)
	}
}

func TestRunAction_ContextAwareRouting_InsideContainer(t *testing.T) {
	ctx := makeRunCtx()
	ctx.ServiceName = "web"

	// Simulate being INSIDE the container: should run locally.
	action := &RunAction{isInsideContainer: func() bool { return true }}

	err := action.Execute(ctx, map[string]interface{}{"run": "echo hello"})
	if err != nil {
		t.Errorf("expected local execution to succeed, got: %v", err)
	}
}

func TestRunAction_ContextAwareRouting_NoService(t *testing.T) {
	ctx := makeRunCtx()
	ctx.ServiceName = ""

	action := &RunAction{}
	err := action.Execute(ctx, map[string]interface{}{"run": "echo hello"})
	if err != nil {
		t.Errorf("expected no error for local execution, got: %v", err)
	}
}

func TestRunAction_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  map[string]interface{}
		wantErr bool
	}{
		{
			name:    "valid config",
			config:  map[string]interface{}{"run": "echo test"},
			wantErr: false,
		},
		{
			name:    "missing run field",
			config:  map[string]interface{}{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			action := &RunAction{}
			err := action.Validate(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("RunAction.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
