# Contributing to Project Actions

Thank you for your interest in contributing to Project Actions! This document will help you get started.

## Code of Conduct

Be respectful, inclusive, and considerate of others. We're all here to learn and build something useful together.

## How to Contribute

There are many ways to contribute:

### 1. Report Bugs

Found a bug? Please [open an issue](https://github.com/project-actions/runner/issues/new) with:
- Clear description of the problem
- Steps to reproduce
- Expected vs actual behavior
- Your environment (OS, Go version, Docker version if applicable)
- Relevant logs or error messages

### 2. Suggest Features

Have an idea? [Start a discussion](https://github.com/project-actions/runner/discussions) or open an issue with:
- Clear description of the feature
- Use cases and benefits
- Potential implementation approach (if you have ideas)

### 3. Improve Documentation

Documentation is crucial! You can:
- Fix typos and grammar
- Clarify confusing sections
- Add more examples
- Improve code comments
- Create tutorials

### 4. Write Code

Ready to code? Great! Here's how:

## Development Setup

### Prerequisites

- Go 1.21 or later
- Docker and docker-compose (for docker-related features)
- Git

### Getting Started

1. **Fork and clone the repository:**
```bash
git fetch https://github.com/project-actions/runner.git
cd runner
```

2. **Install dependencies:**
```bash
go mod download
```

3. **Build the project:**
```bash
./scripts/build.sh build
```

4. **Run tests:**
```bash
./scripts/build.sh test
# or
go test -v ./...
```

5. **Try it out:**
```bash
cd examples/simple-project
./project
```

## Project Structure

```
project-actions-runner/
├── cmd/runner/              # Main entry point
├── internal/
│   ├── actions/             # Action interface and registry
│   │   └── builtin/         # Built-in actions (echo, run, etc)
│   ├── cli/                 # CLI commands (Cobra)
│   ├── config/              # Configuration loading
│   ├── docker/              # Docker utilities
│   ├── executor/            # Execution engine
│   ├── logger/              # Logging
│   └── parser/              # YAML parsing
├── examples/                # Example projects
├── scripts/                 # Build scripts
├── templates/               # Installation templates
└── tests/                   # Integration tests
```

## Making Changes

### 1. Create a Branch

```bash
git checkout -b feature/your-feature-name
# or
git checkout -b fix/bug-description
```

### 2. Write Code

Follow these guidelines:

**Code Style:**
- Use `gofmt` to format code
- Follow Go best practices and idioms
- Keep functions small and focused
- Add comments for complex logic
- Use meaningful variable names

**Testing:**
- Write tests for new features
- Update tests when changing behavior
- Aim for >80% code coverage
- Include both unit and integration tests

**Commits:**
- Write clear commit messages
- Use conventional commit format:
  - `feat: add new action for X`
  - `fix: resolve issue with Y`
  - `docs: update Z documentation`
  - `test: add tests for W`
  - `refactor: improve V implementation`

### 3. Run Tests

```bash
# Run all tests
go test -v ./...

# Run tests with coverage
go test -v -cover ./...

# Run specific package tests
go test -v ./internal/parser/...

# Run integration tests
go test -v ./tests/integration/...
```

### 4. Update Documentation

If your changes affect:
- User-facing features → Update README.md
- Command writing → Update GUIDE.md
- Docker features → Update DOCKER.md
- Building/contributing → Update BUILD.md or CONTRIBUTING.md

### 5. Submit a Pull Request

1. Push your branch:
```bash
git push origin feature/your-feature-name
```

2. Open a pull request on GitHub

3. Fill out the PR template with:
   - Description of changes
   - Related issues
   - Testing performed
   - Documentation updates

## Pull Request Guidelines

**Before submitting:**
- [ ] Tests pass (`go test ./...`)
- [ ] Code is formatted (`go fmt ./...`)
- [ ] No linting errors
- [ ] Documentation updated
- [ ] Commit messages are clear
- [ ] Branch is up to date with main

**PR Description should include:**
- What problem does this solve?
- How does it solve it?
- Any breaking changes?
- Screenshots/examples (if applicable)

## Adding New Features

### Adding a New Built-in Action

1. **Create the action file:**
```go
// internal/actions/builtin/your_action.go
package builtin

import (
    "github.com/project-actions/runner/internal/actions"
)

type YourAction struct{}

func (a *YourAction) Execute(ctx *actions.ExecutionContext, config map[string]interface{}) error {
    // Implementation
    return nil
}

func (a *YourAction) Validate(config map[string]interface{}) error {
    // Validation
    return nil
}
```

2. **Register the action:**
```go
// internal/executor/engine.go
engine.actionRegistry.Register("your-action", &builtin.YourAction{})
```

3. **Write tests:**
```go
// internal/actions/builtin/your_action_test.go
func TestYourAction_Execute(t *testing.T) {
    // Test implementation
}

func TestYourAction_Validate(t *testing.T) {
    // Test validation
}
```

4. **Update documentation:**
```markdown
<!-- README.md -->
**your-action** - Description
\```yaml
- action: your-action
  option: value
\```
```

### Adding New Conditional Types

1. **Update parser:**
```go
// internal/parser/command.go
// Add new conditional type to Conditional struct
```

2. **Implement evaluation:**
```go
// internal/executor/engine.go
case "your-conditional":
    // Evaluation logic
    return result, nil
```

3. **Add tests:**
```go
// internal/parser/command_test.go
// internal/executor/...
```

4. **Document:**
```markdown
<!-- GUIDE.md -->
### your-conditional

Description and examples
```

## Testing Guidelines

### Unit Tests

- Test individual functions
- Mock external dependencies
- Cover edge cases and error conditions
- Keep tests fast

**Example:**
```go
func TestParseStep(t *testing.T) {
    tests := []struct {
        name       string
        raw        map[string]interface{}
        wantAction string
        wantErr    bool
    }{
        {
            name: "echo step",
            raw: map[string]interface{}{
                "echo": "Hello",
            },
            wantAction: "echo",
            wantErr:    false,
        },
        // More test cases...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            step, err := ParseStep(tt.raw)
            if (err != nil) != tt.wantErr {
                t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
            }
            // More assertions...
        })
    }
}
```

### Integration Tests

- Test complete workflows
- Use temporary directories
- Clean up after tests
- Test real command execution

**Example:**
```go
func TestCommandExecution(t *testing.T) {
    tmpDir, _ := os.MkdirTemp("", "test")
    defer os.RemoveAll(tmpDir)

    // Setup test project
    commandsDir := filepath.Join(tmpDir, ".project")
    os.MkdirAll(commandsDir, 0755)

    // Create test command
    testYAML := `help:
  short: Test
steps:
  - echo: "test"
`
    os.WriteFile(filepath.Join(commandsDir, "test.yaml"), []byte(testYAML), 0644)

    // Test execution
    // ...
}
```

## Documentation Style

- Use clear, simple language
- Include code examples
- Show expected output
- Explain "why" not just "what"
- Use proper markdown formatting

## Release Process

(For maintainers)

1. Update version in `internal/cli/version.go`
2. Run all tests
3. Build all platforms: `./scripts/build.sh all`
4. Update CHANGELOG.md
5. Create git tag: `git tag v1.0.0`
6. Push tag: `git push origin v1.0.0`
7. Create GitHub release
8. Upload binaries
9. Update website

## Getting Help

Need help with your contribution?

- Ask in [Discussions](https://github.com/project-actions/runner/discussions)
- Join our community chat
- Tag maintainers in your PR

## Recognition

Contributors will be:
- Listed in the project README
- Mentioned in release notes
- Forever appreciated! 🎉

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

---

Thank you for contributing to Project Actions! 🚀
