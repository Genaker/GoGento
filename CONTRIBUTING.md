# Contributing to GoGento

Thank you for your interest in contributing to GoGento! This document provides guidelines and instructions for contributing to this project.

## Code of Conduct

Please be respectful and constructive in all interactions with other contributors.

## Getting Started

### Prerequisites

- Go 1.23 or higher
- MySQL 8.0 or higher
- Redis (optional)
- Make (optional but recommended)

### Development Setup

1. **Clone the repository**
   ```bash
   git clone https://github.com/Genaker/GoGento.git
   cd GoGento
   ```

2. **Install dependencies**
   ```bash
   make deps
   # or
   go mod download
   ```

3. **Set up environment**
   ```bash
   cp .env.example .env
   # Edit .env with your local database credentials
   ```

4. **Start local services with Docker**
   ```bash
   make docker-up
   # or
   docker-compose up -d
   ```

5. **Run the application**
   ```bash
   make run
   # or
   go run magento.go
   ```

## Development Workflow

### Before Making Changes

1. Create a new branch from `main`:
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. Make sure you're up to date:
   ```bash
   git pull origin main
   ```

### Making Changes

1. **Write clean, idiomatic Go code**
   - Follow the [Effective Go](https://golang.org/doc/effective_go) guidelines
   - Use meaningful variable and function names
   - Keep functions small and focused
   - Add comments for exported functions and complex logic

2. **Format your code**
   ```bash
   make fmt
   # or
   go fmt ./...
   ```

3. **Run linters**
   ```bash
   make lint
   # or
   golangci-lint run ./...
   ```

4. **Run tests**
   ```bash
   make test
   # or
   go test -v ./...
   ```

### Code Structure

Follow the existing project structure:

```
magento.GO/
├── api/            # HTTP handlers
├── cmd/            # CLI commands
├── config/         # Configuration
├── core/           # Core utilities (cache, log, registry)
├── cron/           # Scheduled jobs
├── html/           # HTML templates and handlers
├── model/
│   ├── entity/     # Data models
│   └── repository/ # Data access layer
└── service/        # Business logic
```

### Commit Messages

Write clear, descriptive commit messages:

```
feat: add new product search endpoint
fix: resolve race condition in cache
docs: update API documentation
refactor: simplify order service logic
test: add unit tests for product repository
```

Use conventional commit format:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks

## Testing

### Writing Tests

- Place test files next to the code they test (e.g., `product_service.go` → `product_service_test.go`)
- Use table-driven tests for multiple test cases
- Mock external dependencies
- Aim for high test coverage

Example test structure:
```go
func TestProductService_GetProduct(t *testing.T) {
    tests := []struct {
        name    string
        id      uint
        want    *Product
        wantErr bool
    }{
        // test cases
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // test implementation
        })
    }
}
```

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run tests for a specific package
go test -v ./service/product/...
```

## Pull Request Process

1. **Update documentation** if you've changed APIs or added features
2. **Add tests** for new functionality
3. **Ensure all tests pass** locally
4. **Run linters** and fix any issues
5. **Update README.md** if needed
6. **Create a pull request** with a clear description of changes

### PR Checklist

- [ ] Code follows project style guidelines
- [ ] Tests pass locally
- [ ] New code is covered by tests
- [ ] Documentation is updated
- [ ] Commit messages are clear and descriptive
- [ ] No unnecessary dependencies added

## Adding New Features

### Adding a New API Endpoint

1. **Create the entity model** in `model/entity/`
2. **Create the repository** in `model/repository/`
3. **Create the service** in `service/`
4. **Create the API handler** in `api/`
5. **Register routes** in `magento.go`
6. **Add tests** for each layer
7. **Update documentation**

### Adding a New Cron Job

1. **Create job implementation** in `cron/jobs/`
2. **Register the job** in `config/cron.go`
3. **Add CLI command** in `cmd/cron.go`
4. **Document the job** in README.md

## Code Review

All submissions require review. We use GitHub pull requests for this purpose. Reviewers will check for:

- Code quality and style
- Test coverage
- Documentation
- Performance implications
- Security considerations

## Questions?

If you have questions about contributing, please:

1. Check existing issues and discussions
2. Open a new issue with your question
3. Reach out to maintainers

## License

By contributing to GoGento, you agree that your contributions will be licensed under the MIT License.
