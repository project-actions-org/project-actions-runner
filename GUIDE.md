# Writing Custom Commands Guide

This guide will teach you everything you need to know to write powerful custom commands for Project Actions.

## Table of Contents

1. [Command Basics](#command-basics)
2. [Help Metadata](#help-metadata)
3. [Steps and Actions](#steps-and-actions)
4. [Conditional Logic](#conditional-logic)
5. [Command-Line Options](#command-line-options)
6. [Docker Context](#docker-context)
7. [Best Practices](#best-practices)
8. [Complete Examples](#complete-examples)

## Command Basics

Commands are YAML files in your `.project/` directory (or `.project/commands/`).

**Minimal command:**
```yaml
help:
  short: Brief description

steps:
  - echo: "Hello, World!"
```

**File naming:**
- `setup.yaml` → `./project setup`
- `deploy.yaml` → `./project deploy`
- `run-tests.yaml` → `./project run-tests`

## Help Metadata

The `help` section defines how your command appears in the command list:

```yaml
help:
  short: One-line description (required)
  long: |
    Multi-line detailed description.
    Explain what the command does, when to use it,
    and any important notes.
  order: 10  # Controls position in list (optional)
```

**Order guidelines:**
- 1-10: Setup and initialization commands
- 11-50: Development commands
- 51-89: Maintenance commands
- 90-99: Teardown and cleanup commands

## Steps and Actions

Steps are executed sequentially. Each step uses an action:

### Built-in Actions

**echo** - Print messages
```yaml
steps:
  - echo: "Simple message"

  - echo: |
      Multi-line message
      with multiple lines
```

**run** - Execute shell commands
```yaml
steps:
  - run: "npm install"

  - run: |
      echo "Running tests"
      npm test
      echo "Tests complete"
```

**command** - Call other commands
```yaml
steps:
  - command: setup
  - echo: "Setup completed, starting server"
  - command: serve
```

**check-for** - Verify tools exist
```yaml
steps:
  - check-for: docker
    if-missing: |
      Docker is required but not found.
      Install from https://docker.com

  - check-for: node
```

### Docker Actions

**compose-up** - Start services
```yaml
steps:
  - action: compose-up
    detached: true  # Run in background (default)
```

**compose-stop** - Stop services
```yaml
steps:
  - action: compose-stop
```

**compose-down** - Stop and remove services
```yaml
steps:
  - action: compose-down
    volumes: false  # Keep volumes (default)

  # Or destroy everything
  - action: compose-down
    volumes: true
```

**compose-exec** - Execute in container
```yaml
steps:
  # Interactive shell
  - action: compose-exec
    service: web
    command: /bin/bash

  # Run non-interactive command
  - action: compose-exec
    service: web
    command: "php artisan migrate"
    interactive: false

  # Command with arguments (array form)
  - action: compose-exec
    service: web
    command: ["npm", "run", "build"]
```

## Conditional Logic

Conditionals let you execute steps based on conditions:

### if-option

Execute steps when a command-line option is provided:

```yaml
steps:
  - echo: "Starting deployment"

  - if-option: production
    then:
      - echo: "Production mode enabled"
      - run: "npm run build:production"

  - if-option: debug
    then:
      - echo: "Debug mode enabled"
      - run: "export DEBUG=*"
```

**OR logic** with pipe:
```yaml
steps:
  - if-option: debug|trace|verbose
    then:
      - echo: "Detailed logging enabled"
      - run: "export LOG_LEVEL=debug"
```

### if-no-option

Execute steps when an option is NOT provided:

```yaml
steps:
  - if-no-option: skip-tests
    then:
      - echo: "Running tests"
      - run: "npm test"

  - if-option: skip-tests
    then:
      - echo: "⚠️  Tests skipped"
```

### if-missing

Execute steps when a file or directory doesn't exist:

```yaml
steps:
  - if-missing: node_modules
    then:
      - echo: "Installing dependencies"
      - run: "npm install"

  - if-missing: .env
    then:
      - echo: "Creating .env from template"
      - run: "cp .env.example .env"
```

### Nested Conditionals

You can nest conditionals for complex logic:

```yaml
steps:
  - if-option: production
    then:
      - echo: "Production deployment"

      - if-missing: dist
        then:
          - echo: "Building for production"
          - run: "npm run build"

      - if-option: deploy
        then:
          - echo: "Deploying to server"
          - run: "./deploy.sh"
```

## Command-Line Options

Users pass options with `--flag` syntax:

```bash
# Boolean flags
./project deploy --production --verbose

# Value options
./project build --env=staging

# Multiple options
./project up --fresh --seed --debug
```

**Access in commands:**

```yaml
help:
  short: Build the application

steps:
  - echo: "Starting build"

  - if-option: production
    then:
      - run: "npm run build:prod"

  - if-no-option: production
    then:
      - run: "npm run build:dev"

  - if-option: watch
    then:
      - run: "npm run watch"

  - if-option: verbose
    then:
      - echo: "Build artifacts:"
      - run: "ls -lah dist/"
```

## Docker Context

Specify where commands should run:

### Outside Container

For commands that manage Docker:

```yaml
help:
  short: Start the application

context: outside-container

steps:
  - action: compose-up
  - echo: "Services started"
```

### Inside Container

For commands that run in a specific service:

```yaml
help:
  short: Run database migrations

context: inside-container:web

steps:
  - run: "php artisan migrate"
  - run: "php artisan db:seed"
```

**Note:** Full context switching (automatic re-execution) is planned for a future release. Currently, commands log a warning if run in the wrong context.

## Best Practices

### 1. Always Provide Help

```yaml
help:
  short: Clear one-line description
  long: |
    Detailed explanation of:
    - What the command does
    - When to use it
    - Any prerequisites
    - Available options
```

### 2. Check Prerequisites

```yaml
steps:
  - check-for: docker
    if-missing: "Please install Docker first"

  - check-for: node
    if-missing: "Please install Node.js"
```

### 3. Provide Feedback

```yaml
steps:
  - echo: "🚀 Starting deployment..."
  - run: "./deploy.sh"
  - echo: "✅ Deployment complete!"
```

### 4. Handle Missing Dependencies

```yaml
steps:
  - if-missing: node_modules
    then:
      - echo: "Installing dependencies..."
      - run: "npm install"
```

### 5. Use Meaningful Order Numbers

```yaml
# Setup commands
help:
  order: 1  # setup
  order: 5  # init

# Development commands
help:
  order: 10  # up
  order: 20  # dev
  order: 30  # test

# Cleanup commands
help:
  order: 99  # down
```

### 6. Document Options

```yaml
help:
  short: Deploy the application
  long: |
    Deploy the application to the specified environment.

    Options:
      --production  Deploy to production
      --staging     Deploy to staging
      --dry-run     Show what would be deployed
```

### 7. Fail Fast

```yaml
steps:
  - check-for: docker
  - check-for: docker-compose

  # Only proceed if above checks pass
  - action: compose-up
```

## Complete Examples

### Example 1: Simple Web Server

**`.project/dev.yaml`:**
```yaml
help:
  short: Start development server
  long: |
    Starts the development server with hot reloading.

    Options:
      --port=8080  Custom port (default: 3000)
      --open       Open browser automatically
  order: 10

steps:
  - echo: "Starting development server..."

  - check-for: node
    if-missing: "Please install Node.js from https://nodejs.org"

  - if-missing: node_modules
    then:
      - echo: "Installing dependencies..."
      - run: "npm install"

  - run: "npm run dev"

  - if-option: open
    then:
      - run: "open http://localhost:3000"
```

### Example 2: Docker Compose Project

**`.project/up.yaml`:**
```yaml
help:
  short: Bring the project up
  long: |
    Starts all services using docker-compose.

    Options:
      --fresh    Reset database and run migrations
      --build    Rebuild images before starting
  order: 10

context: outside-container

steps:
  - check-for: docker
    if-missing: "Please install Docker"

  - check-for: docker-compose
    if-missing: "Please install docker-compose"

  - if-option: build
    then:
      - echo: "Building images..."
      - run: "docker-compose build"

  - action: compose-up

  - echo: "⏳ Waiting for services to be ready..."
  - run: "sleep 3"

  - if-option: fresh
    then:
      - echo: "Resetting database..."
      - action: compose-exec
        service: web
        command: "php artisan migrate:fresh --seed"
        interactive: false

  - echo: |
      ✅ Project is up and running!

      Services:
      - Web: http://localhost:8000
      - Database: localhost:3306
```

**`.project/console.yaml`:**
```yaml
help:
  short: Get a shell inside the web container
  order: 50

context: outside-container

steps:
  - check-for: docker-compose

  - action: compose-exec
    service: web
    command: /bin/bash
    interactive: true
```

### Example 3: Testing with Options

**`.project/test.yaml`:**
```yaml
help:
  short: Run tests
  long: |
    Run the test suite.

    Options:
      --watch      Watch for changes
      --coverage   Generate coverage report
      --unit       Run only unit tests
      --e2e        Run only E2E tests
  order: 30

steps:
  - echo: "🧪 Running tests..."

  - check-for: npm

  - if-no-option: unit|e2e
    then:
      - echo: "Running all tests"
      - run: "npm test"

  - if-option: unit
    then:
      - echo: "Running unit tests"
      - run: "npm run test:unit"

  - if-option: e2e
    then:
      - echo: "Running E2E tests"
      - run: "npm run test:e2e"

  - if-option: coverage
    then:
      - echo: "Generating coverage report"
      - run: "npm run test:coverage"
      - echo: "Coverage report: ./coverage/index.html"

  - if-option: watch
    then:
      - echo: "Watching for changes..."
      - run: "npm run test:watch"

  - echo: "✅ Tests complete"
```

### Example 4: Multi-Environment Deployment

**`.project/deploy.yaml`:**
```yaml
help:
  short: Deploy the application
  long: |
    Deploy the application to the specified environment.

    Options:
      --production  Deploy to production (requires confirmation)
      --staging     Deploy to staging
      --dry-run     Show deployment plan without executing
  order: 90

steps:
  - echo: "🚀 Starting deployment..."

  - check-for: git
  - check-for: docker

  # Ensure working directory is clean
  - run: "git diff-index --quiet HEAD || echo 'Warning: uncommitted changes'"

  - if-option: dry-run
    then:
      - echo: "DRY RUN MODE - No changes will be made"

  - if-option: staging
    then:
      - echo: "Deploying to STAGING"
      - run: "git checkout main"
      - run: "git pull"
      - run: "./deploy-staging.sh"

  - if-option: production
    then:
      - echo: "⚠️  PRODUCTION DEPLOYMENT"
      - echo: "Tagging release..."
      - run: "./tag-release.sh"
      - run: "./deploy-production.sh"

  - if-no-option: staging|production
    then:
      - echo: "❌ Error: Must specify --staging or --production"

  - if-no-option: dry-run
    then:
      - echo: "✅ Deployment complete!"
```

## Tips and Tricks

### Combining Multiple Commands

```yaml
steps:
  - command: setup
  - command: up
  - command: seed
  - echo: "Environment ready!"
```

### Environment-Specific Configuration

```yaml
steps:
  - if-missing: .env.local
    then:
      - run: "cp .env.example .env.local"
      - echo: "⚠️  Please update .env.local with your configuration"
```

### Interactive Prompts

```yaml
steps:
  - run: |
      read -p "Are you sure? (y/n) " -n 1 -r
      echo
      if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Cancelled"
        exit 1
      fi
  - echo: "Proceeding..."
```

### Cleanup Actions

```yaml
steps:
  - echo: "Cleaning up..."
  - run: "rm -rf dist/"
  - run: "rm -rf node_modules/"
  - run: "rm -rf .cache/"
  - echo: "✅ Cleanup complete"
```

## Next Steps

- Explore the [examples directory](./examples/)
- Read the [Docker integration guide](./DOCKER.md)
- Check out [real-world examples](./examples/)
- Join the community discussions

## Getting Help

- Documentation: https://project-actions.org/docs
- GitHub Issues: https://github.com/project-actions/runner/issues
- Discussions: https://github.com/project-actions/runner/discussions
