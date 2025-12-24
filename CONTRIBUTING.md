# Contributing to Flowfull Go Starter

Thank you for your interest in contributing to the Flowfull Go Starter! 🎉

## Code of Conduct

Please be respectful and constructive in all interactions.

## How to Contribute

### Reporting Bugs

1. Check if the bug has already been reported in [Issues](https://github.com/yourrepo/issues)
2. If not, create a new issue with:
   - Clear title and description
   - Steps to reproduce
   - Expected vs actual behavior
   - Go version, OS, and other relevant info
   - Code samples or error messages

### Suggesting Features

1. Check existing [Issues](https://github.com/yourrepo/issues) and [Discussions](https://github.com/yourrepo/discussions)
2. Create a new issue with:
   - Clear use case
   - Proposed solution
   - Alternative solutions considered
   - Impact on existing functionality

### Pull Requests

1. **Fork the repository**
2. **Create a feature branch**
   ```bash
   git checkout -b feature/amazing-feature
   ```

3. **Make your changes**
   - Follow the code style (see below)
   - Add tests for new functionality
   - Update documentation as needed

4. **Run tests**
   ```bash
   make test
   make lint
   ```

5. **Commit your changes**
   ```bash
   git commit -m "feat: add amazing feature"
   ```
   
   Use conventional commits:
   - `feat:` - New feature
   - `fix:` - Bug fix
   - `docs:` - Documentation changes
   - `test:` - Test changes
   - `refactor:` - Code refactoring
   - `chore:` - Maintenance tasks

6. **Push to your fork**
   ```bash
   git push origin feature/amazing-feature
   ```

7. **Open a Pull Request**
   - Describe your changes
   - Reference related issues
   - Include screenshots if applicable

## Development Setup

```bash
# Clone your fork
git clone https://github.com/yourusername/flowfull-go-starter.git
cd flowfull-go-starter

# Install dependencies
make install

# Copy environment file
cp .env.example .env

# Run development server
make dev
```

## Code Style

- Follow standard Go conventions
- Use `gofmt` for formatting
- Run `golangci-lint` before committing
- Write clear, descriptive variable names
- Add comments for complex logic
- Keep functions small and focused

### Example

```go
// Good
func (bv *BridgeValidator) ValidateSession(
    ctx context.Context,
    sessionID string,
    opts *ValidationOptions,
) (*SessionData, error) {
    // Implementation
}

// Bad
func validate(s string) interface{} {
    // Implementation
}
```

## Testing

- Write tests for new features
- Maintain or improve code coverage
- Use table-driven tests where appropriate
- Mock external dependencies

### Example Test

```go
func TestValidateSession_Success(t *testing.T) {
    // Arrange
    validator := setupValidator()
    
    // Act
    session, err := validator.ValidateSession(ctx, "session-id", nil)
    
    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, session)
}
```

## Documentation

- Update README.md for user-facing changes
- Update docs/ for architectural changes
- Add inline comments for complex code
- Include examples in documentation

## Project Structure

```
flowfull-go-starter/
├── cmd/server/          # Application entry point
├── internal/
│   ├── config/          # Configuration
│   ├── lib/             # Core libraries
│   ├── models/          # Database models
│   └── routes/          # HTTP routes
├── scripts/             # Utility scripts
├── docs/                # Documentation
└── tests/               # Integration tests
```

## Questions?

- Open a [Discussion](https://github.com/yourrepo/discussions)
- Join our [Discord](https://discord.gg/yourserver)
- Email: support@pubflow.com

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

---

Thank you for contributing! 🙏

