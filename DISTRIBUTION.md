# Distribution Guide

This document explains how to distribute Project Actions Runner to end users.

## Build Process

### 1. Build All Platforms

```bash
./scripts/build.sh all
```

This creates binaries in `dist/`:
- `command-runner-darwin-amd64` - macOS Intel
- `command-runner-darwin-arm64` - macOS Apple Silicon
- `command-runner-linux-amd64` - Linux x64
- `command-runner-linux-arm64` - Linux ARM64

The version format is `X.Y.Z+YYYYMMDDHHMM` (e.g., `1.0.0+202512051709`).

### 2. Verify Builds

```bash
# Check all binaries exist
ls -lh dist/command-runner-*

# Test each binary (if on matching platform)
./dist/command-runner-darwin-arm64 --version
```

## Distribution Files

### Files to Distribute

From the `dist/` directory:
- All `command-runner-*` binaries

From the `templates/` directory:
- `runner.sh` - Bootstrap script that downloads the appropriate binary
- `install.sh` - Installation script for end users

### File Locations on Server

Upload to your distribution server (e.g., `project-actions.org/dist/`):

```
https://project-actions.org/dist/
├── command-runner-darwin-amd64
├── command-runner-darwin-arm64
├── command-runner-linux-amd64
├── command-runner-linux-arm64
├── runner.sh
└── install.sh
```

## Installation Flow

### User Installation

Users run:
```bash
curl -fsSL https://project-actions.org/install.sh | bash
```

The install script:
1. Detects the platform (OS and architecture)
2. Creates `.project/` directory structure
3. Downloads `runner.sh` bootstrap script
4. Creates `project` wrapper script
5. Optionally installs starter commands

### First Run

When user runs `./project`:
1. `project` script calls `.project/.runtime/runner.sh`
2. `runner.sh` detects platform
3. If binary doesn't exist, downloads `command-runner-{platform}`
4. Executes the binary with arguments
5. Binary loads and executes commands

## Binary Versioning

### Version Information

Each binary includes:
- Semantic version (e.g., `1.0.0`)
- Build timestamp (e.g., `202512051709`)
- Combined format: `1.0.0+202512051709`

Check version:
```bash
./dist/command-runner-darwin-arm64 --version
# Output: project version 1.0.0+202512051709
```

### Update Strategy

Users get updates when:
1. They delete the cached binary in `.project/.runtime/`
2. We implement version checking in runner.sh (future enhancement)

## Environment Variables

### PROJECT_ACTIONS_DOWNLOAD_URL

Users can override the download URL:

```bash
export PROJECT_ACTIONS_DOWNLOAD_URL="https://custom-cdn.example.com/dist"
curl -fsSL https://project-actions.org/install.sh | bash
```

### PROJECT_SCRIPT_NAME

The wrapper script name (default: "project"):

```bash
# Using custom script name
mv project myapp
./myapp help  # Shows "myapp" in usage
```

## Testing Distribution

### 1. Local Testing

Test the entire flow locally:

```bash
# Start a simple HTTP server in dist/
cd dist
python3 -m http.server 8080 &

# In another terminal, test installation
cd /tmp/test-project
export PROJECT_ACTIONS_DOWNLOAD_URL="http://localhost:8080"
curl -fsSL http://localhost:8080/install.sh | bash

# Test it works
./project --version
./project help
```

### 2. Platform Testing

Test on each platform:

- macOS Intel (darwin-amd64)
- macOS ARM (darwin-arm64)
- Linux x64 (linux-amd64)
- Linux ARM (linux-arm64)

### 3. Integration Testing

Test with real projects:

```bash
# Test with example projects
cd examples/simple-project
./project
./project test

cd examples/docker-compose-project
./project
./project up
```

## Release Checklist

- [ ] Update version in `internal/cli/version.go`
- [ ] Run all tests: `./scripts/build.sh test`
- [ ] Build all platforms: `./scripts/build.sh all`
- [ ] Verify all binaries: `ls -lh dist/`
- [ ] Test on macOS
- [ ] Test on Linux
- [ ] Upload binaries to distribution server
- [ ] Upload `runner.sh` to distribution server
- [ ] Upload `install.sh` to distribution server
- [ ] Test installation flow from distribution server
- [ ] Update website with new version
- [ ] Create GitHub release
- [ ] Update documentation
- [ ] Announce release

## CDN Configuration

### Headers

Set appropriate headers for distribution files:

```
command-runner-*:
  Content-Type: application/octet-stream
  Cache-Control: public, max-age=3600

runner.sh:
  Content-Type: text/x-shellscript
  Cache-Control: public, max-age=3600

install.sh:
  Content-Type: text/x-shellscript
  Cache-Control: public, max-age=3600
```

### CORS

If hosting on a CDN, enable CORS:

```
Access-Control-Allow-Origin: *
Access-Control-Allow-Methods: GET, HEAD, OPTIONS
```

## Troubleshooting

### Binary Download Fails

Common issues:
1. **Platform not supported**: Check user's OS and architecture
2. **Network issue**: Verify distribution server is accessible
3. **Binary corrupted**: Verify file integrity with checksums (future enhancement)

### Permission Denied

If users can't execute the binary:
```bash
chmod +x .project/.runtime/command-runner-*
```

### Wrong Binary Downloaded

If wrong platform is detected:
```bash
# Check detection
uname -s  # Should be Darwin or Linux
uname -m  # Should be x86_64, amd64, aarch64, or arm64
```

## Future Enhancements

### Checksums

Generate and verify SHA256 checksums:

```bash
# Generate checksums
cd dist
sha256sum command-runner-* > checksums.txt

# Verify in runner.sh
curl -fsSL .../checksums.txt | grep $RUNNER_NAME | sha256sum -c
```

### Version Checking

Auto-update when new version available:

```bash
# In runner.sh
CURRENT_VERSION=$(${RUNNER_PATH} --version 2>/dev/null || echo "0.0.0")
LATEST_VERSION=$(curl -fsSL .../version.txt)

if [ "$CURRENT_VERSION" != "$LATEST_VERSION" ]; then
    echo "New version available, downloading..."
    # Re-download
fi
```

### Compression

Compress binaries to reduce download size:

```bash
# Compress
gzip -9 command-runner-darwin-arm64

# Update runner.sh to decompress
gunzip command-runner-darwin-arm64.gz
```

## Support

- GitHub Issues: https://github.com/project-actions/runner/issues
- Documentation: https://project-actions.org/docs
- Email: support@project-actions.org
