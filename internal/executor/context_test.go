package executor

import (
	"testing"

	"github.com/project-actions/runner/internal/actions"
	"github.com/project-actions/runner/internal/logger"
	"github.com/project-actions/runner/internal/parser"
)

func TestParseContextString(t *testing.T) {
	tests := []struct {
		input       string
		wantType    string
		wantService string
	}{
		{"", "outside-container", ""},
		{"outside-container", "outside-container", ""},
		{"inside-container:web", "inside-container", "web"},
		{"inside-container:app", "inside-container", "app"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ParseContextString(tt.input)
			if got.Type != tt.wantType {
				t.Errorf("ParseContextString(%q).Type = %q, want %q", tt.input, got.Type, tt.wantType)
			}
			if got.ServiceName != tt.wantService {
				t.Errorf("ParseContextString(%q).ServiceName = %q, want %q", tt.input, got.ServiceName, tt.wantService)
			}
		})
	}
}

func TestCheckContext_ServiceNameAlwaysSet(t *testing.T) {
	engine := &Engine{logger: logger.New()}
	cm := NewContextManager(engine)

	cmd := &parser.Command{
		Name:    "run",
		Context: "inside-container:web",
	}
	execCtx := &actions.ExecutionContext{}

	err := cm.CheckContext(cmd, execCtx)
	if err != nil {
		t.Fatalf("CheckContext() unexpected error: %v", err)
	}
	if execCtx.ServiceName != "web" {
		t.Errorf("CheckContext() ServiceName = %q, want %q", execCtx.ServiceName, "web")
	}
}

func TestCheckContext_NoContext(t *testing.T) {
	engine := &Engine{logger: logger.New()}
	cm := NewContextManager(engine)

	cmd := &parser.Command{Name: "setup", Context: ""}
	execCtx := &actions.ExecutionContext{}

	err := cm.CheckContext(cmd, execCtx)
	if err != nil {
		t.Fatalf("CheckContext() unexpected error: %v", err)
	}
	if execCtx.ServiceName != "" {
		t.Errorf("CheckContext() ServiceName = %q, want empty", execCtx.ServiceName)
	}
}
