# Building Project Actions Runner

## Version Format

The runner uses semantic versioning with a build timestamp:

```
Version: X.Y.Z+YYYYMMDDHHMM
Example: 1.0.0+202512051658
```

The build timestamp is automatically generated when building and represents the year, month, day, hour, and minute of the build.

## Quick Build

### Using Build Script

```bash
# Build for current platform
./scripts/build.sh build

# Build for all platforms
./scripts/build.sh all

# Update example project links
./scripts/build.sh update-examples

# Clean build artifacts
./scripts/build.sh clean

# Run tests
./scripts/build.sh test

# Show version
./scripts/build.sh version

# Development build (no version)
./scripts/build.sh dev
```

### Using Go directly

```bash
# Set build time
BUILD_TIME=$(date +%Y%m%d%H%M)

# Build with version
go build \
  -ldflags "-X github.com/project-actions/runner/internal/cli.Version=1.0.0 \
            -X github.com/project-actions/runner/internal/cli.BuildTime=$BUILD_TIME \
            -s -w" \
  -o dist/command-runner \
  ./cmd/runner
```

## Build Targets

- **darwin-amd64**: macOS Intel (x86_64)
- **darwin-arm64**: macOS Apple Silicon (arm64)
- **linux-amd64**: Linux x86_64
- **linux-arm64**: Linux ARM64

## Development Builds

For development, you can build without the version information:

```bash
go build -o dist/command-runner ./cmd/runner
```

This will show version as `1.0.0+dev`.

## Example Projects

The example projects in `examples/` have their `.project/.runtime/command-runner` files as symlinks pointing to `../../../../dist/command-runner`. This means when you build the runner, the examples automatically use the latest build.

### Updating Example Links

After building, update the symlinks:

```bash
./scripts/build.sh update-examples
```

Or manually:

```bash
for example in examples/*; do
  if [ -d "$example/.project/.runtime" ]; then
    rm -f "$example/.project/.runtime/command-runner"
    ln -s "../../../../dist/command-runner" "$example/.project/.runtime/command-runner"
  fi
done
```

## Testing

```bash
# Run all tests
./scripts/build.sh test
```

Or using go directly:

```bash
go test -v ./...
go test -v -cover ./...
```

## Verifying the Build

Check the version:

```bash
# From project root
./dist/command-runner --version

# From an example project
cd examples/simple-project
./.project/.runtime/command-runner --version
```

Output should be:
```
project version 1.0.0+202512051651
```

## Clean Build

Remove all build artifacts:

```bash
./scripts/build.sh clean
# or
rm -rf dist/
```

## Build Script Commands

The `scripts/build.sh` script supports the following commands:

- `build` - Build for current platform with version (default)
- `dev` - Build without version (development mode)
- `all` - Build for all platforms (darwin/linux, amd64/arm64)
- `clean` - Remove build artifacts
- `update-examples` - Update example project symlinks to dist/
- `test` - Run all tests
- `version` - Show current version
- `help` - Show help message
