package external

import (
	"testing"
)

func TestParseSourceURL(t *testing.T) {
	tests := []struct {
		name     string
		rawURL   string
		wantHost string
		wantOrg  string
		wantRepo string
		wantRef  string
		wantErr  bool
	}{
		{
			name:     "github tag ref",
			rawURL:   "github.com/project-actions/aws-project-actions@v1",
			wantHost: "github.com",
			wantOrg:  "project-actions",
			wantRepo: "aws-project-actions",
			wantRef:  "v1",
		},
		{
			name:     "github branch ref",
			rawURL:   "github.com/myorg/my-project-actions@main",
			wantHost: "github.com",
			wantOrg:  "myorg",
			wantRepo: "my-project-actions",
			wantRef:  "main",
		},
		{
			name:    "missing ref",
			rawURL:  "github.com/myorg/my-repo",
			wantErr: true,
		},
		{
			name:    "too few path parts",
			rawURL:  "github.com/myorg@v1",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseSourceURL(tt.rawURL)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseSourceURL() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if got.Host != tt.wantHost {
				t.Errorf("Host = %q, want %q", got.Host, tt.wantHost)
			}
			if got.Org != tt.wantOrg {
				t.Errorf("Org = %q, want %q", got.Org, tt.wantOrg)
			}
			if got.Repo != tt.wantRepo {
				t.Errorf("Repo = %q, want %q", got.Repo, tt.wantRepo)
			}
			if got.Ref != tt.wantRef {
				t.Errorf("Ref = %q, want %q", got.Ref, tt.wantRef)
			}
		})
	}
}

func TestSourceCloneURL(t *testing.T) {
	s, _ := ParseSourceURL("github.com/project-actions/aws-project-actions@v1")
	want := "https://github.com/project-actions/aws-project-actions"
	if s.CloneURL() != want {
		t.Errorf("CloneURL() = %q, want %q", s.CloneURL(), want)
	}
}

func TestSourceBinaryURL(t *testing.T) {
	s, _ := ParseSourceURL("github.com/project-actions/aws-project-actions@v1")
	want := "https://github.com/project-actions/aws-project-actions/releases/download/v1/iam-role-setup-darwin-arm64"
	got := s.BinaryURL("iam-role-setup", "darwin", "arm64")
	if got != want {
		t.Errorf("BinaryURL() = %q, want %q", got, want)
	}
}
