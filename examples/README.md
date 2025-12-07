# Project Actions Examples

This directory contains example projects demonstrating various Project Actions features.

## Available Examples

### [simple-project](./simple-project/)

A basic example demonstrating:
- Simple commands (echo, run)
- Command calling other commands
- Conditional logic (if-option, if-no-option, if-missing)
- Option handling

**Try it:**
```bash
cd examples/simple-project
./project
./project test
./project test --verbose
```

### [docker-compose-project](./docker-compose-project/)

A Docker Compose example demonstrating:
- Docker actions (compose-up, compose-stop, compose-exec)
- Container context switching
- Interactive container shells
- Service management

**Try it:**
```bash
cd examples/docker-compose-project
./project
./project up
./project console
./project down
```

## Creating Your Own Project

### Quick Start

1. Copy one of the example projects:
```bash
cp -r examples/simple-project my-project
cd my-project
```

2. Create your commands in `.project/`:
```bash
cat > .project/hello.yaml << 'EOF'
help:
  short: Say hello
  order: 1

steps:
  - echo: "Hello from Project Actions!"
  - run: "date"
EOF
```

3. Run your command:
```bash
./project hello
```

## Command Structure

All commands follow this structure:

```yaml
help:
  short: Brief description (shown in list)
  long: |
    Detailed explanation (shown in help)
  order: 10  # Display order

context: outside-container  # Optional

steps:
  - echo: "Step 1"
  - run: "echo 'Step 2'"
  - command: other-command
```

## Common Patterns

### Setup Command

```yaml
help:
  short: Setup the project
  order: 1

steps:
  - check-for: docker
    if-missing: "Please install Docker"

  - if-missing: .env
    then:
      - run: "cp .env.example .env"
      - echo: "Created .env file - please update it"

  - command: up
```

### Test Command with Options

```yaml
help:
  short: Run tests
  order: 20

steps:
  - if-option: watch
    then:
      - run: "npm run test:watch"

  - if-no-option: watch
    then:
      - run: "npm test"

  - if-option: coverage
    then:
      - run: "npm run coverage"
```

### Docker Workflow

```yaml
help:
  short: Start services
  order: 10

context: outside-container

steps:
  - action: compose-up

  - if-option: migrate
    then:
      - action: compose-exec
        service: web
        command: "php artisan migrate"
        interactive: false
```

## Learning Path

1. **Start with simple-project**
   - Learn basic actions
   - Understand conditionals
   - Practice with options

2. **Move to docker-compose-project**
   - Learn Docker integration
   - Understand context switching
   - Practice service management

3. **Read the guides**
   - [Writing Custom Commands (GUIDE.md)](../GUIDE.md)
   - [Docker Integration (DOCKER.md)](../DOCKER.md)
   - [Main README](../README.md)

4. **Create your own project**
   - Apply what you've learned
   - Customize for your needs
   - Share your commands!

## More Examples

Looking for more? Check out:
- Real-world projects using Project Actions
- Community-contributed examples
- Template repositories

Visit: https://project-actions.org/examples

## Need Help?

- Read the [Guide](../GUIDE.md)
- Check [Documentation](https://project-actions.org/docs)
- Ask in [Discussions](https://github.com/project-actions/runner/discussions)
- Report issues on [GitHub](https://github.com/project-actions/runner/issues)
