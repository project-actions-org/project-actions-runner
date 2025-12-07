package docker

import (
	"os/exec"
	"strings"
)

// IsDockerInstalled checks if Docker is installed and available
func IsDockerInstalled() bool {
	_, err := exec.LookPath("docker")
	return err == nil
}

// IsDockerComposeInstalled checks if docker-compose is installed and available
func IsDockerComposeInstalled() bool {
	_, err := exec.LookPath("docker-compose")
	return err == nil
}

// IsDockerRunning checks if Docker daemon is running
func IsDockerRunning() bool {
	cmd := exec.Command("docker", "info")
	err := cmd.Run()
	return err == nil
}

// GetDockerVersion returns the Docker version string
func GetDockerVersion() (string, error) {
	cmd := exec.Command("docker", "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// GetDockerComposeVersion returns the docker-compose version string
func GetDockerComposeVersion() (string, error) {
	cmd := exec.Command("docker-compose", "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// IsInsideContainer checks if we're running inside a Docker container
func IsInsideContainer() bool {
	// Check for .dockerenv file (common indicator)
	_, err := exec.Command("test", "-f", "/.dockerenv").Output()
	if err == nil {
		return true
	}

	// Check cgroup (another indicator)
	cmd := exec.Command("grep", "-q", "docker", "/proc/1/cgroup")
	err = cmd.Run()
	return err == nil
}

// DetectEnvironment returns information about the Docker environment
type Environment struct {
	DockerInstalled        bool
	DockerComposeInstalled bool
	DockerRunning          bool
	InsideContainer        bool
	DockerVersion          string
	DockerComposeVersion   string
}

// Detect returns comprehensive Docker environment information
func Detect() *Environment {
	env := &Environment{
		DockerInstalled:        IsDockerInstalled(),
		DockerComposeInstalled: IsDockerComposeInstalled(),
		InsideContainer:        IsInsideContainer(),
	}

	if env.DockerInstalled {
		env.DockerRunning = IsDockerRunning()
		if version, err := GetDockerVersion(); err == nil {
			env.DockerVersion = version
		}
	}

	if env.DockerComposeInstalled {
		if version, err := GetDockerComposeVersion(); err == nil {
			env.DockerComposeVersion = version
		}
	}

	return env
}
