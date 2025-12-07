package docker

import (
	"testing"
)

func TestIsDockerInstalled(t *testing.T) {
	// This test will pass if docker is in PATH
	// We can't mock exec.LookPath easily, so this is an integration-style test
	result := IsDockerInstalled()

	// We don't assert true/false since it depends on the environment
	// Just make sure the function doesn't panic
	t.Logf("IsDockerInstalled returned: %v", result)
}

func TestIsDockerComposeInstalled(t *testing.T) {
	result := IsDockerComposeInstalled()
	t.Logf("IsDockerComposeInstalled returned: %v", result)
}

func TestIsInsideContainer(t *testing.T) {
	// This test checks if we're running inside a container
	result := IsInsideContainer()
	t.Logf("IsInsideContainer returned: %v", result)

	// In normal test environments, this should be false
	// unless tests are running in a container
	if result {
		t.Log("Tests are running inside a Docker container")
	} else {
		t.Log("Tests are running outside a Docker container")
	}
}

func TestDetect(t *testing.T) {
	env := Detect()

	if env == nil {
		t.Fatal("Detect() returned nil")
	}

	t.Logf("Docker Environment:")
	t.Logf("  DockerInstalled: %v", env.DockerInstalled)
	t.Logf("  DockerComposeInstalled: %v", env.DockerComposeInstalled)
	t.Logf("  DockerRunning: %v", env.DockerRunning)
	t.Logf("  InsideContainer: %v", env.InsideContainer)
	t.Logf("  DockerVersion: %s", env.DockerVersion)
	t.Logf("  DockerComposeVersion: %s", env.DockerComposeVersion)

	// Basic sanity checks
	if env.DockerInstalled && env.DockerVersion == "" {
		t.Error("Docker is installed but version is empty")
	}

	if env.DockerComposeInstalled && env.DockerComposeVersion == "" {
		t.Error("Docker Compose is installed but version is empty")
	}
}
