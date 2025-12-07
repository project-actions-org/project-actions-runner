# Project Actions Runner

![Version](https://img.shields.io/badge/version-1.0.0-blue)
![License](https://img.shields.io/badge/license-MIT-green)
![Go](https://img.shields.io/badge/go-1.21+-00ADD8?logo=go)

**Project Actions** is a flexible workflow system for local projects. Write commands in YAML and run them via a simple CLI, similar to GitHub Actions but running locally on your machine.

## Features

- 📝 **YAML-based commands** - Define workflows in simple, readable YAML files
- 🎯 **Built-in actions** - echo, run, command, check-for, and more
- 🐳 **Docker integration** - Full docker-compose support with context switching
- 🔄 **Conditional logic** - if-option, if-no-option, if-missing, if-fails
- 🚀 **Easy installation** - One-command install for macOS and Linux
- 🔧 **Extensible** - Create custom actions and share them
- 💻 **Cross-platform** - Supports macOS (Intel/ARM) and Linux (x64/ARM64)

## Quick Start

### Installation

```bash
# Install Project Actions in your project
curl -fsSL https://project-actions.org/install.sh | bash

# Or with starter commands
curl -fsSL https://project-actions.org/install.sh | bash -s -- --with-starter-commands
```

This creates:
- `.project/` directory for your commands
- `.project/.runtime/` for the runner binary
- `project` wrapper script

### Your First Command

Create `.project/hello.yaml`:

```yaml
help:
  short: Say hello
  order: 1

steps:
  - echo: "Hello from Project Actions!"
  - run: "date"
```

Run it:

```bash
./project hello
```

## Command Structure

Commands are YAML files in `.project/` (or `.project/commands/`):

```yaml
help:
  short: Brief description (shown in command list)
  long: |
    Detailed help text that explains what this command does
    and how to use it.
  order: 10  # Controls display order

context: outside-container  # Optional: outside-container or inside-container:service

steps:
  - echo: "Starting workflow..."

  - check-for: docker
    if-missing: "Please install Docker first"

  - run: "npm install"

  - if-option: production
    then:
      - run: "npm run build"
      - echo: "Production build complete"

  - if-no-option: production
    then:
      - run: "npm run dev"

  - command: deploy  # Call another command
```

## Built-in Actions

### Basic Actions

**echo** - Print a message
```yaml
- echo: "Hello, World!"
```

**run** - Execute a shell command
```yaml
- run: "npm install"
- run: |
    echo "Multi-line"
    echo "commands work too"
```

**command** - Call another command
```yaml
- command: setup
```

**check-for** - Verify a tool exists
```yaml
- check-for: docker
  if-missing: "Please install Docker from https://docker.com"
```

### Docker Actions

**compose-up** - Start docker-compose services
```yaml
- action: compose-up
  detached: true
```

**compose-stop** - Stop services
```yaml
- action: compose-stop
```

**compose-down** - Stop and remove services
```yaml
- action: compose-down
  volumes: true  # Also remove volumes
```

**compose-exec** - Execute in container
```yaml
- action: compose-exec
  service: web
  command: /bin/bash
  interactive: true
```

### Conditionals

**if-option** - Check for command-line option
```yaml
- if-option: verbose
  then:
    - echo: "Verbose mode enabled"

# OR logic with pipe
- if-option: debug|trace
  then:
    - echo: "Debug mode active"
```

**if-no-option** - Inverse of if-option
```yaml
- if-no-option: skip
  then:
    - run: "npm test"
```

**if-missing** - Check if file/directory doesn't exist
```yaml
- if-missing: node_modules
  then:
    - run: "npm install"
```

**if-fails** - Execute on failure (partial implementation)
```yaml
- if-fails: previous-step
  then:
    - echo: "Something went wrong"
```

## Command-Line Options

Pass options to commands with `--flag` syntax:

```bash
# Boolean flag
./project deploy --production

# Value
./project build --env=staging

# Multiple options
./project up --refresh --verbose
```

Access in commands with `if-option`:

```yaml
- if-option: production
  then:
    - echo: "Production mode"
```

### Verbose Mode

The `--verbose` flag is special: it shows subprocess output from `run` actions.

**Without `--verbose` (default):**
- Shows what commands are running
- Hides command output for cleaner logs
- Perfect for CI/CD and production

**With `--verbose`:**
- Shows what commands are running
- Shows all command output
- Perfect for debugging

```bash
# Clean output (commands shown, output hidden)
./project build

# Full output (commands and their output shown)
./project build --verbose
```

Example output:
```
==> npm install
✓ Completed: run
==> npm test
✓ Completed: run
```

With `--verbose`, you'll see all npm output between the `==>` and `✓` lines.

## Docker Integration

Project Actions includes comprehensive Docker support:

### Context Switching

Specify where commands should run:

```yaml
context: outside-container
# or
context: inside-container:web
```

### Example Docker Workflow

```yaml
help:
  short: Start the application

context: outside-container

steps:
  - check-for: docker-compose
    if-missing: "Please install docker-compose"

  - action: compose-up

  - echo: "Services are starting..."

  - action: compose-exec
    service: web
    command: "php artisan migrate"
    interactive: false

  - echo: |
      Application is ready!
      Visit: http://localhost:8000
```

## Project Structure

```
my-project/
├── .project/
│   ├── .runtime/
│   │   ├── .gitignore
│   │   ├── runner.sh
│   │   └── command-runner-*  # Platform-specific binaries (downloaded)
│   ├── setup.yaml
│   ├── up.yaml
│   ├── down.yaml
│   └── deploy.yaml
├── project  # Wrapper script
└── ... (your project files)
```

## Development

### Building from Source

```bash
# Install dependencies
go mod download

# Build for current platform
./scripts/build.sh build

# Build for all platforms
./scripts/build.sh all

# Run tests
./scripts/build.sh test

# Update example projects
./scripts/build.sh update-examples
```

### Running Tests

```bash
# All tests
go test -v ./...

# Specific package
go test -v ./internal/parser/...

# With coverage
go test -v -cover ./...
```

### Project Structure

```
project-actions-runner/
├── cmd/runner/          # Main entry point
├── internal/
│   ├── actions/         # Action interface and registry
│   │   └── builtin/     # Built-in actions
│   ├── cli/             # Cobra CLI commands
│   ├── config/          # Configuration loading
│   ├── docker/          # Docker utilities
│   ├── executor/        # Execution engine
│   ├── logger/          # Logging
│   └── parser/          # YAML parsing
├── examples/            # Example projects
├── scripts/             # Build scripts
├── templates/           # Installation templates
└── tests/               # Integration tests
```

## Documentation

- **[BUILD.md](BUILD.md)** - Building and development
- **[DOCKER.md](DOCKER.md)** - Docker integration details
- **[Website](https://project-actions.org/)** - Full documentation

## Examples

### Simple Web Project

```yaml
help:
  short: Start development server
  order: 1

steps:
  - check-for: node
  - if-missing: node_modules
    then:
      - run: "npm install"
  - run: "npm run dev"
```

### Docker Compose Project

```yaml
help:
  short: Bring the project up
  order: 2

context: outside-container

steps:
  - action: compose-up
  - echo: "Services started"

  - if-option: fresh
    then:
      - action: compose-exec
        service: web
        command: "php artisan migrate:fresh --seed"
        interactive: false

  - echo: "Visit http://localhost:8000"
```

### Multi-Step Deployment

```yaml
help:
  short: Deploy to production
  order: 99

steps:
  - echo: "🚀 Starting deployment..."

  - check-for: git
  - check-for: docker

  - run: "git fetch origin"
  - run: "git checkout main"
  - run: "git pull"

  - run: "docker build -t myapp:latest ."

  - if-option: push
    then:
      - run: "docker push myapp:latest"
      - echo: "✓ Image pushed"

  - echo: "✓ Deployment complete"
```

## Troubleshooting

### Command not found

```bash
# Make sure project script is executable
chmod +x ./project

# Check runner is installed
ls -la .project/.runtime/
```

### Binary download fails

The runner bootstrap script downloads binaries from `https://project-actions.org/dist/`. If downloads fail:

1. Check internet connectivity
2. Verify the distribution server is accessible
3. Manually download the binary for your platform
4. Place it in `.project/.runtime/`

### Docker commands fail

```bash
# Verify Docker is installed and running
docker info

# Check docker-compose
docker-compose --version

# Verify compose file exists
ls docker-compose.yml
```

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Write tests for new features
4. Ensure all tests pass
5. Submit a pull request

## License

MIT License - see LICENSE file for details

## Links

- **Website**: https://project-actions.org
- **GitHub**: https://github.com/project-actions/runner
- **Issues**: https://github.com/project-actions/runner/issues
- **Discussions**: https://github.com/project-actions/runner/discussions

---

Made with ❤️ by the Project Actions team
