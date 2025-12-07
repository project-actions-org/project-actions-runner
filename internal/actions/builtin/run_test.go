package builtin

import (
	"testing"

	"github.com/project-actions/runner/internal/actions"
	"github.com/project-actions/runner/internal/logger"
)

func TestRunAction_Execute(t *testing.T) {
	tests := []struct {
		name    string
		config  map[string]interface{}
		wantErr bool
	}{
		{
			name: "simple echo command",
			config: map[string]interface{}{
				"run": "echo 'test'",
			},
			wantErr: false,
		},
		{
			name: "pwd command",
			config: map[string]interface{}{
				"run": "pwd",
			},
			wantErr: false,
		},
		{
			name: "command with exit 0",
			config: map[string]interface{}{
				"run": "exit 0",
			},
			wantErr: false,
		},
		{
			name: "command that fails",
			config: map[string]interface{}{
				"run": "exit 1",
			},
			wantErr: true,
		},
		{
			name:    "missing run field",
			config:  map[string]interface{}{},
			wantErr: true,
		},
		{
			name: "invalid command",
			config: map[string]interface{}{
				"run": "nonexistent-command-xyz",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			action := &RunAction{}
			ctx := &actions.ExecutionContext{
				Logger:     logger.New(),
				WorkingDir: "/tmp",
			}

			err := action.Execute(ctx, tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("RunAction.Execute() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRunAction_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  map[string]interface{}
		wantErr bool
	}{
		{
			name: "valid config",
			config: map[string]interface{}{
				"run": "echo test",
			},
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
