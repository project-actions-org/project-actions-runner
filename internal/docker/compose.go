package docker

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// ComposeCommand represents a docker-compose command
type ComposeCommand struct {
	WorkingDir string
	Args       []string
	Env        map[string]string
}

// Up starts docker-compose services
func Up(workingDir string, detached bool) error {
	args := []string{"up"}
	if detached {
		args = append(args, "-d")
	}

	cmd := exec.Command("docker-compose", args...)
	cmd.Dir = workingDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}

// Down stops and removes docker-compose services
func Down(workingDir string, removeVolumes bool) error {
	args := []string{"down"}
	if removeVolumes {
		args = append(args, "-v")
	}

	cmd := exec.Command("docker-compose", args...)
	cmd.Dir = workingDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// Stop stops docker-compose services without removing them
func Stop(workingDir string) error {
	cmd := exec.Command("docker-compose", "stop")
	cmd.Dir = workingDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// Exec executes a command in a running container
func Exec(workingDir, service string, command []string, interactive bool) error {
	args := []string{"exec"}

	if !interactive {
		args = append(args, "-T")
	}

	args = append(args, service)
	args = append(args, command...)

	cmd := exec.Command("docker-compose", args...)
	cmd.Dir = workingDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}

// IsRunning checks if docker-compose services are running
func IsRunning(workingDir string) (bool, error) {
	cmd := exec.Command("docker-compose", "ps", "-q")
	cmd.Dir = workingDir

	output, err := cmd.Output()
	if err != nil {
		return false, err
	}

	// If there's output, services are running
	return len(strings.TrimSpace(string(output))) > 0, nil
}

// GetRunningServices returns a list of running service names
func GetRunningServices(workingDir string) ([]string, error) {
	cmd := exec.Command("docker-compose", "ps", "--services", "--filter", "status=running")
	cmd.Dir = workingDir

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	services := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(services) == 1 && services[0] == "" {
		return []string{}, nil
	}

	return services, nil
}

// Pull pulls the latest images for all services
func Pull(workingDir string) error {
	cmd := exec.Command("docker-compose", "pull")
	cmd.Dir = workingDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// Build builds or rebuilds services
func Build(workingDir string, noCache bool) error {
	args := []string{"build"}
	if noCache {
		args = append(args, "--no-cache")
	}

	cmd := exec.Command("docker-compose", args...)
	cmd.Dir = workingDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// Logs shows logs from services
func Logs(workingDir string, follow bool, tail string) error {
	args := []string{"logs"}
	if follow {
		args = append(args, "-f")
	}
	if tail != "" {
		args = append(args, "--tail", tail)
	}

	cmd := exec.Command("docker-compose", args...)
	cmd.Dir = workingDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}

// HasComposeFile checks if a docker-compose.yml file exists
func HasComposeFile(workingDir string) bool {
	composeFiles := []string{
		"docker-compose.yml",
		"docker-compose.yaml",
		"compose.yml",
		"compose.yaml",
	}

	for _, file := range composeFiles {
		path := fmt.Sprintf("%s/%s", workingDir, file)
		if _, err := os.Stat(path); err == nil {
			return true
		}
	}

	return false
}
