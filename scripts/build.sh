#!/bin/bash
set -e

# Build script for Project Actions Runner

VERSION="0.3.3"
BUILD_TIME=$(date +%Y%m%d%H%M)
BINARY_NAME="command-runner"
DIST_DIR="dist"
EXAMPLES_DIR="examples"

# Build flags
LDFLAGS="-X github.com/project-actions/runner/internal/cli.Version=${VERSION} -X github.com/project-actions/runner/internal/cli.BuildTime=${BUILD_TIME} -s -w"

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to show usage
usage() {
    echo "Usage: $0 [command]"
    echo ""
    echo "Commands:"
    echo "  build          Build for current platform (default)"
    echo "  dev            Build without version (dev mode)"
    echo "  all            Build for all platforms"
    echo "  clean          Remove build artifacts"
    echo "  update-examples Update example project links"
    echo "  test           Run tests"
    echo "  version        Show version"
    echo ""
    exit 1
}

# Ad-hoc sign a binary if on macOS (required for arm64 binaries)
sign_if_macos() {
    local binary="$1"
    if [ "$(uname)" = "Darwin" ]; then
        codesign --sign - "${binary}" 2>/dev/null || true
    fi
}

# Build for current platform
build() {
    echo -e "${BLUE}Building ${BINARY_NAME} for current platform...${NC}"
    echo "Version: ${VERSION}+${BUILD_TIME}"

    mkdir -p "${DIST_DIR}"
    go build -ldflags "${LDFLAGS}" -o "${DIST_DIR}/${BINARY_NAME}" ./cmd/runner
    sign_if_macos "${DIST_DIR}/${BINARY_NAME}"

    echo -e "${GREEN}Build complete: ${DIST_DIR}/${BINARY_NAME}${NC}"
}

# Build in dev mode (no version)
build_dev() {
    echo -e "${BLUE}Building ${BINARY_NAME} in dev mode...${NC}"

    mkdir -p "${DIST_DIR}"
    go build -o "${DIST_DIR}/${BINARY_NAME}" ./cmd/runner
    sign_if_macos "${DIST_DIR}/${BINARY_NAME}"

    echo -e "${GREEN}Build complete: ${DIST_DIR}/${BINARY_NAME}${NC}"
}

# Build for all platforms
build_all() {
    echo -e "${BLUE}Building ${BINARY_NAME} for all platforms...${NC}"
    echo "Version: ${VERSION}+${BUILD_TIME}"

    mkdir -p "${DIST_DIR}"

    platforms=("darwin-amd64" "darwin-arm64" "linux-amd64" "linux-arm64")

    for platform in "${platforms[@]}"; do
        OS=$(echo $platform | cut -d'-' -f1)
        ARCH=$(echo $platform | cut -d'-' -f2)
        OUTPUT="${DIST_DIR}/${BINARY_NAME}-${platform}"

        echo "Building for ${OS}/${ARCH} -> ${OUTPUT}"
        GOOS=$OS GOARCH=$ARCH go build -ldflags "${LDFLAGS}" -o "${OUTPUT}" ./cmd/runner
        if [ "$OS" = "darwin" ]; then
            sign_if_macos "${OUTPUT}"
        fi
    done

    echo -e "${GREEN}All builds complete!${NC}"
    ls -lh "${DIST_DIR}/"
}

# Clean build artifacts
clean() {
    echo -e "${BLUE}Cleaning...${NC}"
    rm -rf "${DIST_DIR}"
    echo -e "${GREEN}Clean complete${NC}"
}

# Update example project links
update_examples() {
    echo -e "${BLUE}Updating example project runner links...${NC}"

    # First make sure we have a build
    if [ ! -f "${DIST_DIR}/${BINARY_NAME}" ]; then
        echo "No build found, building first..."
        build
    fi

    for example in ${EXAMPLES_DIR}/*; do
        if [ -d "$example/.project/.runtime" ]; then
            runtime_dir="$example/.project/.runtime"
            echo "Updating $example..."

            rm -f "$runtime_dir/${BINARY_NAME}"
            ln -s "../../../../${DIST_DIR}/${BINARY_NAME}" "$runtime_dir/${BINARY_NAME}"

            echo "  ✓ Linked ${BINARY_NAME} -> ../../../../${DIST_DIR}/${BINARY_NAME}"
        fi
    done

    echo -e "${GREEN}Example projects updated!${NC}"
}

# Run tests
run_tests() {
    echo -e "${BLUE}Running tests...${NC}"
    go test -v ./...
}

# Show version
show_version() {
    echo "Version: ${VERSION}+${BUILD_TIME}"
}

# Main script logic
case "${1:-build}" in
    build)
        build
        ;;
    dev)
        build_dev
        ;;
    all)
        build_all
        ;;
    clean)
        clean
        ;;
    update-examples)
        update_examples
        ;;
    test)
        run_tests
        ;;
    version)
        show_version
        ;;
    help|--help|-h)
        usage
        ;;
    *)
        echo "Unknown command: $1"
        usage
        ;;
esac
