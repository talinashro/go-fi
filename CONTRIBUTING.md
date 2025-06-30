# Contributing to go-fi

Thank you for your interest in contributing to go-fi! This document provides guidelines for contributing to the project.

## How to Contribute

### 1. Fork the Repository

1. Fork the repository on GitHub
2. Clone your fork locally:
   ```bash
   git clone https://github.com/your-username/go-fi.git
   cd go-fi
   ```

### 2. Create a Feature Branch

Create a new branch for your feature or fix:

```bash
git checkout -b feature/amazing-feature
# or
git checkout -b fix/bug-fix
```

### 3. Make Your Changes

- Write clean, well-documented code
- Follow Go conventions and best practices
- Add tests for new functionality
- Update documentation as needed

### 4. Test Your Changes

Run the test suite to ensure everything works:

```bash
go test ./...
go test -tags testing ./...
```

### 5. Commit Your Changes

Write clear, descriptive commit messages:

```bash
git commit -m "Add new feature: environment-based fault injection control"
```

### 6. Push to Your Fork

Push your changes to your fork:

```bash
git push origin feature/amazing-feature
```

### 7. Create a Pull Request

1. Go to your fork on GitHub
2. Click "New Pull Request"
3. Select the branch with your changes
4. Write a clear description of your changes
5. Submit the pull request

## Development Guidelines

### Code Style

- Follow Go formatting standards (`gofmt`)
- Use meaningful variable and function names
- Add comments for complex logic
- Keep functions small and focused

### Testing

- Write unit tests for new functionality
- Ensure tests pass with both production and testing build tags
- Test edge cases and error conditions

### Documentation

- Update README.md for user-facing changes
- Add inline comments for complex code
- Update examples if API changes

## Environment Setup

### Prerequisites

- Go 1.21 or later
- Git

### Local Development

1. Clone the repository
2. Install dependencies:
   ```bash
   go mod download
   ```
3. Run tests:
   ```bash
   go test ./...
   go test -tags testing ./...
   ```

## Issue Reporting

Before creating an issue, please:

1. Check existing issues to avoid duplicates
2. Use the issue template if available
3. Provide clear steps to reproduce the problem
4. Include relevant error messages and logs

## Pull Request Guidelines

- Keep PRs focused on a single feature or fix
- Write clear commit messages
- Include tests for new functionality
- Update documentation as needed
- Ensure all tests pass

## Code of Conduct

- Be respectful and inclusive
- Focus on constructive feedback
- Help others learn and grow

## License

By contributing to go-fi, you agree that your contributions will be licensed under the Apache License 2.0.

Thank you for contributing to go-fi! 🚀 