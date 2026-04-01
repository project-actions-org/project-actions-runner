package external

import (
	"fmt"
	"strings"
)

// Source represents a parsed external action source declaration.
type Source struct {
	Name string // short alias used in command YAML (e.g. "aws")
	Host string // e.g. "github.com"
	Org  string // e.g. "project-actions"
	Repo string // e.g. "aws-project-actions"
	Ref  string // tag, branch, or SHA (e.g. "v1", "main")
}

// ParseSourceURL parses a raw source URL of the form "host/org/repo@ref".
// Example: "github.com/project-actions/aws-project-actions@v1"
func ParseSourceURL(rawURL string) (*Source, error) {
	parts := strings.SplitN(rawURL, "@", 2)
	if len(parts) != 2 || parts[1] == "" {
		return nil, fmt.Errorf("source URL must include @ref (e.g. github.com/org/repo@v1): %q", rawURL)
	}
	ref := parts[1]

	pathParts := strings.Split(parts[0], "/")
	if len(pathParts) != 3 {
		return nil, fmt.Errorf("source URL must be host/org/repo@ref (e.g. github.com/org/repo@v1): %q", rawURL)
	}

	return &Source{
		Host: pathParts[0],
		Org:  pathParts[1],
		Repo: pathParts[2],
		Ref:  ref,
	}, nil
}

// CloneURL returns the HTTPS URL used to clone the source repository.
func (s *Source) CloneURL() string {
	return fmt.Sprintf("https://%s/%s/%s", s.Host, s.Org, s.Repo)
}

// BinaryURL returns the GitHub Releases download URL for a platform-specific binary.
func (s *Source) BinaryURL(actionName, goos, goarch string) string {
	asset := fmt.Sprintf("%s-%s-%s", actionName, goos, goarch)
	return fmt.Sprintf("https://%s/%s/%s/releases/download/%s/%s",
		s.Host, s.Org, s.Repo, s.Ref, asset)
}
