# Contributing to gofs

Thank you for your interest in contributing to gofs! This document provides guidelines and instructions for contributing to this project.

## Getting Started

1. Fork the repository on GitHub
2. Clone your fork locally:
   ```bash
   git clone https://github.com/YOUR_USERNAME/gofs.git
   cd gofs
   ```
3. Add the upstream repository:
   ```bash
   git remote add upstream https://github.com/absfs/gofs.git
   ```

## Development Setup

### Prerequisites

- Go 1.22 or later
- Git

### Install Dependencies

```bash
go mod download
```

### Running Tests

Run the full test suite:
```bash
go test -v -race ./...
```

Run tests with coverage:
```bash
go test -v -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

Run benchmarks:
```bash
go test -bench=. -benchmem ./...
```

Run example tests:
```bash
go test -v -run=Example
```

## Making Changes

### Before You Start

- Check existing issues and pull requests to avoid duplicating work
- For major changes, open an issue first to discuss your proposal
- Keep changes focused - one feature or fix per pull request

### Code Style

- Follow standard Go conventions and idioms
- Run `gofmt` to format your code:
  ```bash
  gofmt -w .
  ```
- Run `go vet` to check for common issues:
  ```bash
  go vet ./...
  ```
- Ensure all exported functions, types, and methods have documentation comments
- Documentation comments should be complete sentences starting with the name being documented

### Writing Tests

- Add tests for all new functionality
- Maintain or improve code coverage
- Include table-driven tests where appropriate
- Add examples for new public APIs
- Benchmark performance-critical code

Example test structure:
```go
func TestNewFeature(t *testing.T) {
    t.Run("success case", func(t *testing.T) {
        // Test code
    })

    t.Run("error case", func(t *testing.T) {
        // Test code
    })
}
```

### Commit Messages

Write clear, descriptive commit messages:

```
Short summary (50 chars or less)

More detailed explanation if needed. Wrap at 72 characters.
- Explain what changed and why
- Reference related issues: Fixes #123

Multiple paragraphs are okay.
```

## Submitting Changes

1. Create a feature branch:
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. Make your changes and commit them:
   ```bash
   git add .
   git commit -m "Add feature description"
   ```

3. Keep your branch up to date with upstream:
   ```bash
   git fetch upstream
   git rebase upstream/master
   ```

4. Push to your fork:
   ```bash
   git push origin feature/your-feature-name
   ```

5. Create a Pull Request on GitHub

### Pull Request Guidelines

- Fill out the pull request template completely
- Link to any related issues
- Ensure all tests pass
- Ensure code is properly formatted (`gofmt`)
- Update documentation if needed
- Add or update examples if adding new features
- Keep pull requests focused on a single change

## Code Review Process

- All submissions require review before merging
- Reviewers may request changes or improvements
- Address review feedback promptly
- Once approved, a maintainer will merge your PR

## Testing Standards

All code contributions should include appropriate tests:

- **Unit Tests**: Test individual functions and methods
- **Integration Tests**: Test component interactions
- **Example Tests**: Demonstrate usage (appear in godoc)
- **Benchmarks**: For performance-critical code

Aim for high test coverage, but focus on meaningful tests over achieving a percentage.

## Documentation

- Update README.md if adding user-facing features
- Add godoc comments for all exported identifiers
- Include runnable examples for new public APIs
- Keep examples simple and focused

## Performance Considerations

- Benchmark any performance-critical changes
- Avoid unnecessary allocations
- Consider memory usage and efficiency
- Profile before optimizing

## Reporting Issues

When reporting issues, please include:

- Go version (`go version`)
- Operating system and version
- Minimal code to reproduce the issue
- Expected vs actual behavior
- Any error messages or stack traces

## Questions?

If you have questions about contributing, feel free to:

- Open an issue with the question label
- Check existing documentation and issues first

## License

By contributing to gofs, you agree that your contributions will be licensed under the MIT License.

Thank you for contributing! 🎉
