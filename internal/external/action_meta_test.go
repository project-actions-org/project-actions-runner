package external

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseActionMeta(t *testing.T) {
	tmpDir := t.TempDir()
	content := `name: IAM Role Setup
description: Interactively configure an AWS IAM role
inputs:
  role-name:
    description: Name for the IAM role
    required: true
  region:
    description: AWS region
    required: false
    default: us-east-1
`
	path := filepath.Join(tmpDir, "action.yaml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	meta, err := ParseActionMeta(path)
	if err != nil {
		t.Fatalf("ParseActionMeta() error = %v", err)
	}
	if meta.Name != "IAM Role Setup" {
		t.Errorf("Name = %q, want %q", meta.Name, "IAM Role Setup")
	}
	if len(meta.Inputs) != 2 {
		t.Fatalf("expected 2 inputs, got %d", len(meta.Inputs))
	}
	if !meta.Inputs["role-name"].Required {
		t.Error("role-name should be required")
	}
	if meta.Inputs["region"].Default != "us-east-1" {
		t.Errorf("region default = %q, want %q", meta.Inputs["region"].Default, "us-east-1")
	}
}

func TestValidateAndApplyDefaults(t *testing.T) {
	meta := &ActionMeta{
		Inputs: map[string]InputSpec{
			"role-name": {Required: true},
			"region":    {Required: false, Default: "us-east-1"},
		},
	}

	t.Run("missing required input", func(t *testing.T) {
		with := map[string]interface{}{}
		_, err := meta.ValidateAndApplyDefaults(with)
		if err == nil {
			t.Error("expected error for missing required input")
		}
	})

	t.Run("default applied", func(t *testing.T) {
		with := map[string]interface{}{"role-name": "my-role"}
		result, err := meta.ValidateAndApplyDefaults(with)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result["region"] != "us-east-1" {
			t.Errorf("region = %q, want %q", result["region"], "us-east-1")
		}
	})

	t.Run("provided value overrides default", func(t *testing.T) {
		with := map[string]interface{}{"role-name": "my-role", "region": "eu-west-1"}
		result, err := meta.ValidateAndApplyDefaults(with)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result["region"] != "eu-west-1" {
			t.Errorf("region = %q, want %q", result["region"], "eu-west-1")
		}
	})
}
