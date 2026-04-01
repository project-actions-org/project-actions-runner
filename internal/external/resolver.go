package external

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// ActionType represents whether an action uses a shell script or Go binary.
type ActionType int

const (
	ShellAction  ActionType = iota
	BinaryAction ActionType = iota
)

// ResolvedAction holds everything needed to execute an external action.
type ResolvedAction struct {
	ActionDir  string
	ActionMeta *ActionMeta
	Type       ActionType
	ExecPath   string // path to run.sh (shell) or downloaded binary
}

// Resolver resolves source/action references to local executable paths,
// fetching sources on demand.
type Resolver struct {
	Fetcher *Fetcher
}

// Resolve looks up an action by source alias and action name, fetching the
// source if not cached, and downloading any required binary.
func (r *Resolver) Resolve(sourceName, actionName string, source *Source, projectRoot string) (*ResolvedAction, error) {
	// Fetch source if not already cached
	if !r.Fetcher.IsFetched(sourceName) {
		if err := r.Fetcher.EnsureGitignore(projectRoot); err != nil {
			// Non-fatal: gitignore management should not block execution
			fmt.Fprintf(os.Stderr, "warning: could not update .gitignore: %v\n", err)
		}
		if err := r.Fetcher.FetchSource(sourceName, source); err != nil {
			return nil, err
		}
	}

	// Locate the action directory
	actionDir := r.Fetcher.ActionDir(sourceName, actionName)
	if _, err := os.Stat(actionDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("action %q not found in source %q (%s)", actionName, sourceName, actionDir)
	}

	// Require action.yaml
	metaPath := filepath.Join(actionDir, "action.yaml")
	if _, err := os.Stat(metaPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("missing action.yaml in %s — action repositories must include action.yaml", actionDir)
	}

	meta, err := ParseActionMeta(metaPath)
	if err != nil {
		return nil, err
	}

	// Detect action type
	actionType, err := detectActionType(actionDir)
	if err != nil {
		return nil, err
	}

	resolved := &ResolvedAction{
		ActionDir:  actionDir,
		ActionMeta: meta,
		Type:       actionType,
	}

	switch actionType {
	case ShellAction:
		resolved.ExecPath = filepath.Join(actionDir, "run.sh")

	case BinaryAction:
		goos := runtime.GOOS
		goarch := runtime.GOARCH
		binaryPath := r.Fetcher.BinaryPath(sourceName, actionName, goos, goarch)

		if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
			// Download from GitHub Releases
			path, err := r.Fetcher.DownloadBinary(sourceName, actionName, source, goos, goarch)
			if err != nil {
				return nil, err
			}
			binaryPath = path
		}

		resolved.ExecPath = binaryPath
	}

	return resolved, nil
}

// detectActionType inspects an action directory for run.go (binary) or run.sh (shell).
// run.go takes priority over run.sh if both are present.
func detectActionType(actionDir string) (ActionType, error) {
	if _, err := os.Stat(filepath.Join(actionDir, "run.go")); err == nil {
		return BinaryAction, nil
	}
	if _, err := os.Stat(filepath.Join(actionDir, "run.sh")); err == nil {
		return ShellAction, nil
	}
	return 0, fmt.Errorf("no run.sh or run.go found in %s — action must include one of these files", actionDir)
}
