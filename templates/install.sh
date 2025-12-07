#!/bin/bash
#
# Project Actions Installation Script
#
# Usage:
#   curl -fsSL https://project-actions.org/install.sh | bash
#   curl -fsSL https://project-actions.org/install.sh | bash -s -- --with-starter-commands
#

set -e

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Configuration
PROJECT_DIR=".project"
RUNTIME_DIR="${PROJECT_DIR}/.runtime"
COMMANDS_DIR="${PROJECT_DIR}"
RUNNER_SCRIPT="runner.sh"
PROJECT_SCRIPT="project"
WITH_STARTER_COMMANDS=false

# Parse arguments
for arg in "$@"; do
    case $arg in
        --with-starter-commands)
            WITH_STARTER_COMMANDS=true
            shift
            ;;
    esac
done

# Detect platform
detect_platform() {
    local os=$(uname -s | tr '[:upper:]' '[:lower:]')
    local arch=$(uname -m)

    case "$arch" in
        x86_64|amd64)
            arch="amd64"
            ;;
        aarch64|arm64)
            arch="arm64"
            ;;
        *)
            echo -e "${RED}Error: Unsupported architecture: $arch${NC}"
            exit 1
            ;;
    esac

    echo "${os}-${arch}"
}

# Check if we're in a project directory
check_project() {
    if [ -d "${PROJECT_DIR}" ]; then
        echo -e "${YELLOW}Warning: ${PROJECT_DIR} directory already exists${NC}"
        read -p "Continue with installation? (y/N) " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            echo "Installation cancelled"
            exit 0
        fi
    fi
}

# Create directory structure
create_directories() {
    echo -e "${BLUE}Creating directory structure...${NC}"
    mkdir -p "${RUNTIME_DIR}"
    mkdir -p "${COMMANDS_DIR}"

    # Create .gitignore for runtime directory
    cat > "${RUNTIME_DIR}/.gitignore" << 'EOF'
# Ignore all binaries
command-runner*

# But keep this gitignore
!.gitignore
EOF

    echo -e "${GREEN}✓ Created ${PROJECT_DIR} directory structure${NC}"
}

# Download runner bootstrap script
download_runner() {
    echo -e "${BLUE}Downloading runner bootstrap script...${NC}"

    local runner_url="https://project-actions.org/dist/${RUNNER_SCRIPT}"
    local runner_path="${RUNTIME_DIR}/${RUNNER_SCRIPT}"

    if command -v curl > /dev/null 2>&1; then
        curl -fsSL "${runner_url}" -o "${runner_path}"
    elif command -v wget > /dev/null 2>&1; then
        wget -q "${runner_url}" -O "${runner_path}"
    else
        echo -e "${RED}Error: Neither curl nor wget is available${NC}"
        exit 1
    fi

    chmod +x "${runner_path}"
    echo -e "${GREEN}✓ Downloaded runner bootstrap script${NC}"
}

# Create project wrapper script
create_wrapper() {
    echo -e "${BLUE}Creating project wrapper script...${NC}"

    cat > "${PROJECT_SCRIPT}" << 'EOF'
#!/bin/bash
#
# Project Actions Wrapper Script
#
# This script provides a convenient entry point for running project commands.
# It delegates to the runner bootstrap script which handles downloading and
# executing the appropriate command-runner binary for your platform.
#

set -e

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Path to the runner bootstrap script
RUNNER="${SCRIPT_DIR}/.project/.runtime/runner.sh"

# Check if runner exists
if [ ! -f "${RUNNER}" ]; then
    echo "Error: Runner not found at ${RUNNER}"
    echo "Please run the installation script first:"
    echo "  curl -fsSL https://project-actions.org/install.sh | bash"
    exit 1
fi

# Execute the runner with the script name and all arguments
PROJECT_SCRIPT_NAME="$(basename "$0")" "${RUNNER}" "$@"
EOF

    chmod +x "${PROJECT_SCRIPT}"
    echo -e "${GREEN}✓ Created ${PROJECT_SCRIPT} wrapper script${NC}"
}

# Download starter commands
download_starter_commands() {
    echo -e "${BLUE}Downloading starter commands...${NC}"

    # For now, create simple example commands
    # In production, these would be downloaded from the website

    cat > "${COMMANDS_DIR}/hello.yaml" << 'EOF'
help:
  short: Example hello command
  long: |
    This is an example command that demonstrates basic Project Actions features.
  order: 1

steps:
  - echo: "Hello from Project Actions!"
  - run: "echo 'Current directory:' && pwd"
  - echo: "✓ Command completed successfully"
EOF

    echo -e "${GREEN}✓ Created starter commands${NC}"
}

# Main installation
main() {
    echo -e "${BLUE}================================${NC}"
    echo -e "${BLUE}Project Actions Installation${NC}"
    echo -e "${BLUE}================================${NC}"
    echo ""

    # Detect platform
    PLATFORM=$(detect_platform)
    echo -e "Platform: ${GREEN}${PLATFORM}${NC}"
    echo ""

    # Check existing project
    check_project

    # Create directories
    create_directories

    # Download runner
    download_runner

    # Create wrapper script
    create_wrapper

    # Download starter commands if requested
    if [ "$WITH_STARTER_COMMANDS" = true ]; then
        download_starter_commands
    fi

    echo ""
    echo -e "${GREEN}================================${NC}"
    echo -e "${GREEN}Installation Complete!${NC}"
    echo -e "${GREEN}================================${NC}"
    echo ""
    echo "To get started:"
    echo -e "  ${BLUE}./${PROJECT_SCRIPT}${NC}              - List available commands"
    echo -e "  ${BLUE}./${PROJECT_SCRIPT} hello${NC}        - Run the hello command"
    echo ""
    echo "Create your own commands:"
    echo -e "  Add YAML files to ${BLUE}${COMMANDS_DIR}/${NC}"
    echo ""
    echo "Documentation:"
    echo "  https://project-actions.org/docs"
    echo ""
}

main
