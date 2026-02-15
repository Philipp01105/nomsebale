# Contributing to Noms

Thank you for your interest in contributing to Noms! This document provides guidelines and instructions for contributing.

## Code of Conduct

Be respectful and constructive in all interactions.

## How to Contribute

### Reporting Bugs

If you find a bug, please open an issue with:
- A clear description of the problem
- Steps to reproduce the issue
- Expected behavior
- Actual behavior
- Your environment (OS, Go version)

### Suggesting Features

Feature suggestions are welcome! Please open an issue describing:
- The feature you'd like to see
- Why it would be useful
- How it might work

### Pull Requests

1. **Fork the repository** and create a new branch for your feature or bugfix:
   ```bash
   git checkout -b feature/my-new-feature
   ```

2. **Make your changes** following the coding standards below

3. **Write or update tests** for your changes

4. **Run all checks** before submitting:
   ```bash
   make ci
   ```

5. **Commit your changes** with clear, descriptive commit messages:
   ```bash
   git commit -m "Add feature: description of what you did"
   ```

6. **Push to your fork** and create a pull request

## Coding Standards

### Go Code Style

- Follow standard Go conventions
- Run `gofmt` to format your code (or use `make fmt`)
- Run `go vet` to check for common mistakes (or use `make vet`)
- Follow the guidelines from [Effective Go](https://golang.org/doc/effective_go.html)

### Testing

- Write unit tests for new functionality
- Ensure all tests pass: `go test ./...`
- Aim for good test coverage
- Tests should be clear and maintainable

### Documentation

- Update README.md if you change user-facing functionality
- Add comments for exported functions and types
- Keep comments clear and concise

## Development Workflow

### Setting Up

```bash
# Clone your fork
git clone https://github.com/YOUR_USERNAME/nomsebale.git
cd nomsebale

# Add upstream remote
git remote add upstream https://github.com/Philipp01105/nomsebale.git

# Install dependencies
go mod download
```

### Making Changes

```bash
# Create a new branch
git checkout -b feature/my-feature

# Make changes and test
make test

# Check formatting and linting
make fmt
make vet

# Commit your changes
git add .
git commit -m "Descriptive commit message"

# Push to your fork
git push origin feature/my-feature
```

### Before Submitting

Run the full CI check suite:

```bash
make ci
```

This will:
1. Format your code
2. Run go vet
3. Run all tests

If you have golangci-lint installed, also run:

```bash
make lint
```

## CI/CD

All pull requests will automatically run through:

1. **Build Check**: Ensures code compiles on Go 1.21, 1.22, and 1.23
2. **Tests**: All tests must pass with race detection enabled
3. **Code Quality**: Checks formatting, vetting, and linting
4. **Coverage**: Generates code coverage reports

Failed checks will block merging, so ensure everything passes locally first.

## Project Structure

```
noms/
├── cmd/
│   └── main.go           # CLI entry point
├── pkg/
│   ├── vcs/              # Core VCS functionality
│   ├── branch/           # Branch operations
│   ├── commit/           # Commit operations
│   ├── checkout/         # Checkout operations
│   ├── log/              # Log operations
│   ├── status/           # Status operations
│   ├── initializer/      # Repository initialization
│   └── utils/            # Utility functions
├── .github/
│   └── workflows/        # GitHub Actions workflows
├── Makefile              # Build and test commands
└── README.md             # Project documentation
```

## Questions?

If you have questions about contributing, feel free to open an issue for discussion.
