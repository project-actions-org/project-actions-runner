package logger

import (
	"bytes"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	log := New()
	if log == nil {
		t.Fatal("New() returned nil")
	}

	if log.level != LogInfo {
		t.Errorf("Default log level = %v, want %v", log.level, LogInfo)
	}
}

func TestNewWithLevel(t *testing.T) {
	log := NewWithLevel(LogDebug)
	if log == nil {
		t.Fatal("NewWithLevel() returned nil")
	}

	if log.level != LogDebug {
		t.Errorf("Log level = %v, want %v", log.level, LogDebug)
	}
}

func TestNewWithWriter(t *testing.T) {
	buf := &bytes.Buffer{}
	log := NewWithWriter(buf)

	if log == nil {
		t.Fatal("NewWithWriter() returned nil")
	}

	log.Info("test message")
	output := buf.String()

	if !strings.Contains(output, "test message") {
		t.Errorf("Expected output to contain 'test message', got: %s", output)
	}
}

func TestSetLevel(t *testing.T) {
	log := New()
	log.SetLevel(LogWarn)

	if log.level != LogWarn {
		t.Errorf("Log level = %v, want %v", log.level, LogWarn)
	}
}

func TestInfo(t *testing.T) {
	buf := &bytes.Buffer{}
	log := NewWithWriter(buf)

	log.Info("info message")
	output := buf.String()

	if !strings.Contains(output, "info message") {
		t.Errorf("Expected output to contain 'info message', got: %s", output)
	}

	if !strings.HasSuffix(strings.TrimSpace(output), "info message") {
		t.Errorf("Expected output to end with 'info message', got: %s", output)
	}
}

func TestDebug(t *testing.T) {
	buf := &bytes.Buffer{}
	log := NewWithWriter(buf)
	log.SetLevel(LogDebug)

	log.Debug("debug message")
	output := buf.String()

	if !strings.Contains(output, "[DEBUG] debug message") {
		t.Errorf("Expected output to contain '[DEBUG] debug message', got: %s", output)
	}
}

func TestDebug_NotShownAtInfoLevel(t *testing.T) {
	buf := &bytes.Buffer{}
	log := NewWithWriter(buf)
	log.SetLevel(LogInfo)

	log.Debug("debug message")
	output := buf.String()

	if output != "" {
		t.Errorf("Expected no output at Info level, got: %s", output)
	}
}

func TestWarn(t *testing.T) {
	buf := &bytes.Buffer{}
	log := NewWithWriter(buf)

	log.Warn("warning message")
	output := buf.String()

	if !strings.Contains(output, "[WARN] warning message") {
		t.Errorf("Expected output to contain '[WARN] warning message', got: %s", output)
	}
}

func TestStepStart(t *testing.T) {
	buf := &bytes.Buffer{}
	log := NewWithWriter(buf)

	log.StepStart("test-action")
	output := buf.String()

	if !strings.Contains(output, "→ Running: test-action") {
		t.Errorf("Expected output to contain '→ Running: test-action', got: %s", output)
	}
}

func TestStepSuccess(t *testing.T) {
	buf := &bytes.Buffer{}
	log := NewWithWriter(buf)

	log.StepSuccess("test-action")
	output := buf.String()

	if !strings.Contains(output, "✓ Completed: test-action") {
		t.Errorf("Expected output to contain '✓ Completed: test-action', got: %s", output)
	}
}

func TestPrint(t *testing.T) {
	buf := &bytes.Buffer{}
	log := NewWithWriter(buf)

	log.Print("test %s", "message")
	output := buf.String()

	if output != "test message" {
		t.Errorf("Expected 'test message', got: %s", output)
	}
}

func TestPrintln(t *testing.T) {
	buf := &bytes.Buffer{}
	log := NewWithWriter(buf)

	log.Println("test message")
	output := buf.String()

	if output != "test message\n" {
		t.Errorf("Expected 'test message\\n', got: %s", output)
	}
}

func TestLogLevels(t *testing.T) {
	tests := []struct {
		name      string
		level     LogLevel
		shouldLog map[string]bool
	}{
		{
			name:  "Debug level",
			level: LogDebug,
			shouldLog: map[string]bool{
				"debug": true,
				"info":  true,
				"warn":  true,
			},
		},
		{
			name:  "Info level",
			level: LogInfo,
			shouldLog: map[string]bool{
				"debug": false,
				"info":  true,
				"warn":  true,
			},
		},
		{
			name:  "Warn level",
			level: LogWarn,
			shouldLog: map[string]bool{
				"debug": false,
				"info":  false,
				"warn":  true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			log := NewWithWriter(buf)
			log.SetLevel(tt.level)

			// Test debug
			buf.Reset()
			log.Debug("debug")
			hasOutput := len(buf.String()) > 0
			if hasOutput != tt.shouldLog["debug"] {
				t.Errorf("Debug log at %v level: got output=%v, want %v", tt.level, hasOutput, tt.shouldLog["debug"])
			}

			// Test info
			buf.Reset()
			log.Info("info")
			hasOutput = len(buf.String()) > 0
			if hasOutput != tt.shouldLog["info"] {
				t.Errorf("Info log at %v level: got output=%v, want %v", tt.level, hasOutput, tt.shouldLog["info"])
			}

			// Test warn
			buf.Reset()
			log.Warn("warn")
			hasOutput = len(buf.String()) > 0
			if hasOutput != tt.shouldLog["warn"] {
				t.Errorf("Warn log at %v level: got output=%v, want %v", tt.level, hasOutput, tt.shouldLog["warn"])
			}
		})
	}
}
