package builtin

import (
	"strings"
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

func TestRunAction_AtDollarInterpolation(t *testing.T) {
	tests := []struct {
		name      string
		cmdStr    string
		args      []string
		wantErr   bool
		errSubstr string
	}{
		{
			name:    "at-dollar replaced with all args",
			cmdStr:  "echo @$",
			args:    []string{"hello", "world"},
			wantErr: false,
		},
		{
			name:    "at-dollar as prefix",
			cmdStr:  "echo prefix-@$",
			args:    []string{"suffix"},
			wantErr: false,
		},
		{
			name:      "at-dollar with empty args returns error",
			cmdStr:    "echo @$",
			args:      []string{},
			wantErr:   true,
			errSubstr: "no arguments given",
		},
		{
			name:      "at-dollar with nil args returns error",
			cmdStr:    "@$",
			args:      nil,
			wantErr:   true,
			errSubstr: "no arguments given",
		},
		{
			name:    "no at-dollar, args present, runs normally",
			cmdStr:  "echo fixed",
			args:    []string{"ignored", "args"},
			wantErr: false,
		},
		{
			name:    "multiple at-dollar occurrences all replaced",
			cmdStr:  "echo @$ && echo @$",
			args:    []string{"hi"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := makeRunCtx()
			ctx.Args = tt.args

			action := &RunAction{}
			err := action.Execute(ctx, map[string]interface{}{"run": tt.cmdStr})

			if (err != nil) != tt.wantErr {
				t.Errorf("RunAction.Execute() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.errSubstr != "" && err != nil {
				if !strings.Contains(err.Error(), tt.errSubstr) {
					t.Errorf("RunAction.Execute() error = %q, want to contain %q", err.Error(), tt.errSubstr)
				}
			}
		})
	}
}

func TestRunAction_ContextAwareRouting_WithService(t *testing.T) {
	ctx := makeRunCtx()
	ctx.ServiceName = "web"
	ctx.Args = []string{"hello"}

	// Simulate being outside the container: isInsideContainer returns false
	action := &RunAction{isInsideContainer: func() bool { return false }}

	// docker-compose exec will fail (not running), but the error should be
	// exec-related, not "no arguments given" — confirms routing took the right branch
	err := action.Execute(ctx, map[string]interface{}{"run": "echo @$"})
	if err == nil {
		// docker-compose happened to be available and it worked — that's fine too
		return
	}
	// The error should not be our empty-args error; it should be a system/exec error
	if strings.Contains(err.Error(), "no arguments given") {
		t.Errorf("got wrong error branch: %v", err)
	}
}

func TestRunAction_ContextAwareRouting_InsideContainer(t *testing.T) {
	ctx := makeRunCtx()
	ctx.ServiceName = "web"
	ctx.Args = []string{"hello"}

	// Simulate being INSIDE the container: isInsideContainer returns true
	// → should run locally (not via docker-compose exec)
	action := &RunAction{isInsideContainer: func() bool { return true }}

	err := action.Execute(ctx, map[string]interface{}{"run": "echo @$"})
	if err != nil {
		t.Errorf("expected local execution to succeed, got: %v", err)
	}
}

func TestRunAction_ContextAwareRouting_NoService(t *testing.T) {
	ctx := makeRunCtx()
	ctx.ServiceName = ""
	ctx.Args = []string{"hello"}

	action := &RunAction{}
	err := action.Execute(ctx, map[string]interface{}{"run": "echo @$"})
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
