package builtin

import (
	"testing"

	"github.com/project-actions/runner/internal/actions"
	"github.com/project-actions/runner/internal/logger"
)

func TestEchoAction_Execute(t *testing.T) {
	tests := []struct {
		name    string
		config  map[string]interface{}
		wantErr bool
	}{
		{
			name: "simple string message",
			config: map[string]interface{}{
				"echo": "Hello, World!",
			},
			wantErr: false,
		},
		{
			name: "empty string",
			config: map[string]interface{}{
				"echo": "",
			},
			wantErr: false,
		},
		{
			name: "numeric value",
			config: map[string]interface{}{
				"echo": 123,
			},
			wantErr: false,
		},
		{
			name:    "missing echo field",
			config:  map[string]interface{}{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			action := &EchoAction{}
			ctx := &actions.ExecutionContext{
				Logger: logger.New(),
			}

			err := action.Execute(ctx, tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("EchoAction.Execute() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEchoAction_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  map[string]interface{}
		wantErr bool
	}{
		{
			name: "valid config",
			config: map[string]interface{}{
				"echo": "message",
			},
			wantErr: false,
		},
		{
			name:    "missing echo field",
			config:  map[string]interface{}{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			action := &EchoAction{}
			err := action.Validate(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("EchoAction.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
