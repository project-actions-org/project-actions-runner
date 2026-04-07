package primitives

import (
	"bytes"
	"testing"

	"github.com/project-actions/runner/internal/actions"
	"github.com/project-actions/runner/internal/logger"
)

func TestCheckForAction_Execute(t *testing.T) {
	tests := []struct {
		name    string
		config  map[string]interface{}
		wantErr bool
	}{
		{name: "existing command - sh", config: map[string]interface{}{"check-for": "sh"}, wantErr: false},
		{name: "existing command - echo", config: map[string]interface{}{"check-for": "echo"}, wantErr: false},
		{name: "nonexistent command", config: map[string]interface{}{"check-for": "nonexistent-tool-12345"}, wantErr: true},
		{name: "nonexistent command with custom message", config: map[string]interface{}{"check-for": "nonexistent-tool-12345", "if-missing": "Please install the required tool"}, wantErr: true},
		{name: "missing check-for field", config: map[string]interface{}{}, wantErr: true},
		{name: "check-for with integer value", config: map[string]interface{}{"check-for": 123}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			action := &CheckForAction{}
			buf := &bytes.Buffer{}
			log := logger.NewWithWriter(buf)
			ctx := &actions.ExecutionContext{
				WorkingDir:  "/tmp",
				Environment: make(map[string]string),
				Options:     make(map[string]string),
				Logger:      log,
			}
			err := action.Execute(ctx, tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckForAction.Execute() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCheckForAction_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  map[string]interface{}
		wantErr bool
	}{
		{name: "valid config", config: map[string]interface{}{"check-for": "docker"}, wantErr: false},
		{name: "valid config with if-missing", config: map[string]interface{}{"check-for": "docker", "if-missing": "Please install Docker"}, wantErr: false},
		{name: "missing check-for field", config: map[string]interface{}{}, wantErr: true},
		{name: "only if-missing field", config: map[string]interface{}{"if-missing": "Some message"}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			action := &CheckForAction{}
			err := action.Validate(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckForAction.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
