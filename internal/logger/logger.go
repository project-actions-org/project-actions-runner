package logger

import (
	"fmt"
	"io"
	"os"

	"github.com/fatih/color"
)

// LogLevel represents the severity of a log message
type LogLevel int

const (
	LogDebug LogLevel = iota
	LogInfo
	LogWarn
	LogError
)

// Logger handles formatted output for the CLI
type Logger struct {
	level  LogLevel
	writer io.Writer
}

// New creates a new Logger instance
func New() *Logger {
	return &Logger{
		level:  LogInfo,
		writer: os.Stdout,
	}
}

// NewWithLevel creates a logger with a specific level
func NewWithLevel(level LogLevel) *Logger {
	return &Logger{
		level:  level,
		writer: os.Stdout,
	}
}

// NewWithWriter creates a logger with a custom writer (useful for testing)
func NewWithWriter(writer io.Writer) *Logger {
	return &Logger{
		level:  LogInfo,
		writer: writer,
	}
}

// SetLevel changes the log level
func (l *Logger) SetLevel(level LogLevel) {
	l.level = level
}

// Info logs an informational message
func (l *Logger) Info(format string, args ...interface{}) {
	if l.level <= LogInfo {
		fmt.Fprintf(l.writer, format+"\n", args...)
	}
}

// Debug logs a debug message
func (l *Logger) Debug(format string, args ...interface{}) {
	if l.level <= LogDebug {
		fmt.Fprintf(l.writer, "[DEBUG] "+format+"\n", args...)
	}
}

// Warn logs a warning message
func (l *Logger) Warn(format string, args ...interface{}) {
	if l.level <= LogWarn {
		fmt.Fprintf(l.writer, "[WARN] "+format+"\n", args...)
	}
}

// Error logs an error message
func (l *Logger) Error(format string, args ...interface{}) {
	if l.level <= LogError {
		fmt.Fprintf(os.Stderr, "[ERROR] "+format+"\n", args...)
	}
}

// StepStart logs the start of a step
func (l *Logger) StepStart(actionName string) {
	if l.level <= LogInfo {
		green := color.New(color.FgGreen).SprintFunc()
		fmt.Fprintf(l.writer, "%s Running: %s\n", green("→"), actionName)
	}
}

// StepSuccess logs successful step completion
func (l *Logger) StepSuccess(actionName string) {
	if l.level <= LogInfo {
		green := color.New(color.FgGreen).SprintFunc()
		fmt.Fprintf(l.writer, "%s Completed: %s\n", green("✓"), actionName)
	}
}

// StepFail logs a failed step
func (l *Logger) StepFail(actionName string, err error) {
	if l.level <= LogError {
		red := color.New(color.FgRed).SprintFunc()
		fmt.Fprintf(os.Stderr, "%s Failed: %s\n", red("✗"), actionName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  Error: %v\n", err)
		}
	}
}

// CommandStart logs the start of a command execution with a nice format
func (l *Logger) CommandStart(cmdStr string) {
	if l.level <= LogInfo {
		green := color.New(color.FgGreen, color.Bold).SprintFunc()
		gray := color.New(color.FgHiBlack).SprintFunc()
		fmt.Fprintf(l.writer, "%s %s\n", green("==>"), gray(cmdStr))
	}
}

// Print outputs a message without formatting
func (l *Logger) Print(format string, args ...interface{}) {
	fmt.Fprintf(l.writer, format, args...)
}

// Println outputs a message with a newline
func (l *Logger) Println(format string, args ...interface{}) {
	fmt.Fprintf(l.writer, format+"\n", args...)
}
