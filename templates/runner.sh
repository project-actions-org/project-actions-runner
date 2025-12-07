#!/bin/bash
#
# Project Actions Runner Bootstrap Script
#
# This script detects the platform and downloads/executes the appropriate
# command-runner binary. It caches the binary locally in .project/.runtime/
#

set -e

# Detect platform
detect_platform() {
    local os=$(uname -s | tr '[:upper:]' '[:lower:]')
    local arch=$(uname -m)

    # Normalize architecture names
    case "$arch" in
        x86_64|amd64)
            arch="amd64"
            ;;
        aarch64|arm64)
            arch="arm64"
            ;;
        *)
            echo "Error: Unsupported architecture: $arch"
            exit 1
            ;;
    esac

    echo "${os}-${arch}"
}

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
RUNTIME_DIR="$SCRIPT_DIR"

# Detect platform
PLATFORM=$(detect_platform)
RUNNER_NAME="command-runner-${PLATFORM}"
RUNNER_PATH="${RUNTIME_DIR}/${RUNNER_NAME}"

# Download URL (customize this to match your distribution location)
DOWNLOAD_URL="${PROJECT_ACTIONS_DOWNLOAD_URL:-https://project-actions.org/dist}/${RUNNER_NAME}"

# Check if binary exists and is executable
if [ ! -f "$RUNNER_PATH" ] || [ ! -x "$RUNNER_PATH" ]; then
    echo "Downloading Project Actions Runner for ${PLATFORM}..."
    echo "URL: ${DOWNLOAD_URL}"

    # Create runtime directory if it doesn't exist
    mkdir -p "${RUNTIME_DIR}"

    # Download the binary
    if command -v curl > /dev/null 2>&1; then
        if ! curl -fsSL "${DOWNLOAD_URL}" -o "${RUNNER_PATH}"; then
            echo "Error: Failed to download runner from ${DOWNLOAD_URL}"
            echo ""
            echo "Please ensure:"
            echo "  1. You have internet connectivity"
            echo "  2. The distribution server is available"
            echo "  3. Your platform (${PLATFORM}) is supported"
            exit 1
        fi
    elif command -v wget > /dev/null 2>&1; then
        if ! wget -q "${DOWNLOAD_URL}" -O "${RUNNER_PATH}"; then
            echo "Error: Failed to download runner from ${DOWNLOAD_URL}"
            exit 1
        fi
    else
        echo "Error: Neither curl nor wget is available"
        echo "Please install curl or wget and try again"
        exit 1
    fi

    # Make it executable
    chmod +x "${RUNNER_PATH}"
    echo "✓ Downloaded and installed runner"
fi

# Verify the binary is executable
if [ ! -x "$RUNNER_PATH" ]; then
    echo "Error: The command runner at $RUNNER_PATH is not executable"
    echo "Attempting to fix permissions..."
    chmod +x "${RUNNER_PATH}"
fi

# Execute the runner with all arguments
# Pass the script name via environment variable
PROJECT_SCRIPT_NAME="${PROJECT_SCRIPT_NAME:-$(basename "$0")}" exec "${RUNNER_PATH}" "$@"
