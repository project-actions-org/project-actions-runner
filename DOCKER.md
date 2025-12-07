# Docker Integration (Phase 4)

Project Actions Runner includes comprehensive Docker and docker-compose support for managing containerized development environments.

## Docker Actions

### compose-up

Starts docker-compose services.

**Usage:**
```yaml
steps:
  - action: compose-up
```

**Configuration:**
- `detached` (bool, optional): Run services in detached mode (default: true)

**Example:**
```yaml
steps:
  - action: compose-up
    detached: true
```

### compose-stop

Stops docker-compose services without removing them.

**Usage:**
```yaml
steps:
  - action: compose-stop
```

**Example:**
```yaml
help:
  short: Bring down the project

steps:
  - action: compose-stop
  - echo: "Services stopped"
```

### compose-down

Stops and removes docker-compose services.

**Usage:**
```yaml
steps:
  - action: compose-down
```

**Configuration:**
- `volumes` (bool, optional): Remove volumes as well (default: false)

**Example:**
```yaml
steps:
  - if-option: destroy
    then:
      - action: compose-down
        volumes: true
```

### compose-exec

Executes a command in a running container service.

**Usage:**
```yaml
steps:
  - action: compose-exec
    service: web
    command: /bin/bash
```

**Configuration:**
- `service` (string, required): Service name to execute in
- `command` (string or array, required): Command to execute
- `interactive` (bool, optional): Enable interactive mode with TTY (default: true)

**Examples:**
```yaml
# Interactive shell
steps:
  - action: compose-exec
    service: web
    command: /bin/bash
    interactive: true

# Non-interactive command
steps:
  - action: compose-exec
    service: web
    command: "php artisan migrate"
    interactive: false

# Command with arguments (array syntax)
steps:
  - action: compose-exec
    service: web
    command: ["php", "artisan", "migrate", "--force"]
```

## Context Switching

Commands can specify their execution context to ensure they run in the correct environment.

### Context Types

**outside-container**: Command should run on the host machine
```yaml
context: outside-container

steps:
  - action: compose-up
  - echo: "Services started"
```

**inside-container:service**: Command should run inside a specific container
```yaml
context: inside-container:web

steps:
  - run: php artisan migrate
  - run: php artisan db:seed
```

### How It Works

1. The runner detects the current environment (inside or outside container)
2. If the command specifies a context, it checks if we're in the correct environment
3. If not in the correct context, a warning is logged
4. Future enhancement: Automatic re-execution in the correct context

**Example:**
```yaml
# Command that must run outside container
help:
  short: Bring the project up
context: outside-container

steps:
  - action: compose-up
  - echo: "Project is running"
```

## Docker Detection

The runner includes utilities to detect the Docker environment:

- **IsDockerInstalled()**: Checks if Docker is available
- **IsDockerComposeInstalled()**: Checks if docker-compose is available
- **IsDockerRunning()**: Verifies Docker daemon is running
- **IsInsideContainer()**: Detects if running inside a container
- **HasComposeFile()**: Checks for docker-compose.yml/yaml files

These are used internally by docker actions to provide helpful error messages.

## Example Project Structure

```
my-project/
├── docker-compose.yml
├── .project/
│   ├── .runtime/
│   │   └── command-runner
│   ├── setup.yaml
│   ├── up.yaml
│   ├── down.yaml
│   └── console.yaml
└── project (wrapper script)
```

### Example Commands

**up.yaml** - Start services:
```yaml
help:
  short: Bring the project up
  order: 2

context: outside-container

steps:
  - action: compose-up
  - echo: "Project is now up and running"
```

**down.yaml** - Stop services:
```yaml
help:
  short: Bring down the project
  order: 99

context: outside-container

steps:
  - action: compose-stop
  - echo: "Services stopped"
```

**console.yaml** - Get a shell:
```yaml
help:
  short: Get a shell inside the web container
  order: 50

context: outside-container

steps:
  - action: compose-exec
    service: web
    command: /bin/bash
```

## Error Handling

Docker actions provide helpful error messages:

- If docker-compose is not installed
- If Docker daemon is not running
- If no docker-compose.yml file is found
- If a service doesn't exist or isn't running

**Example error:**
```
Error: docker-compose is not installed or not in PATH

Please install docker-compose and ensure it's available in your PATH.
```

## Testing

Docker utilities include comprehensive tests:

```bash
# Run docker package tests
go test -v ./internal/docker/...

# Run all tests
./scripts/build.sh test
```

Tests cover:
- Docker and docker-compose detection
- Container detection
- Compose file detection
- Context parsing and switching logic

## Tips

1. **Always specify context** for docker commands to ensure correct execution
2. **Use detached mode** for compose-up to return control immediately
3. **Check for tools** with check-for action before running docker commands:
   ```yaml
   steps:
     - check-for: docker-compose
       if-missing: "Please install docker-compose first"
     - action: compose-up
   ```

4. **Combine with conditionals** for flexible workflows:
   ```yaml
   steps:
     - action: compose-up

     - if-option: fresh
       then:
         - action: compose-exec
           service: web
           command: "php artisan migrate:fresh --seed"
   ```

## What's Next?

Future enhancements for Phase 4:
- Automatic context switching (re-execute inside/outside containers)
- Support for Docker Compose v2 plugin syntax
- Build and image management actions
- Volume and network management
- Multi-file compose support
