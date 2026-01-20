# Contributing to bujo

Thank you for your interest in contributing to bujo.

## Development Setup

1. Clone the repository:
   ```bash
   git clone https://github.com/typingincolor/bujo.git
   cd bujo
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Build and test:
   ```bash
   go build -o bujo ./cmd/bujo
   go test ./...
   ```

For desktop app development, see [BUILD.md](BUILD.md).

## Test-Driven Development

**TDD is mandatory for all contributions.** Every line of production code must be written in response to a failing test.

### The TDD Cycle

1. **RED** - Write a failing test first
2. **GREEN** - Write the minimum code to pass
3. **REFACTOR** - Improve code quality while tests pass

### Example

```go
// 1. RED - Write failing test
func TestEntry_MarkDone(t *testing.T) {
    entry := Entry{Type: Task}
    done := entry.MarkDone()
    assert.Equal(t, Done, done.Type)
}

// 2. GREEN - Implement minimum code
func (e Entry) MarkDone() Entry {
    e.Type = Done
    return e
}

// 3. REFACTOR - Improve if needed
```

### Running Tests

```bash
# All tests
go test ./...

# With coverage
go test -cover ./...

# Specific package
go test -v ./internal/domain/...

# Single test
go test -v -run TestEntry_MarkDone ./internal/domain/...
```

## Code Style

### Go Conventions

- Follow standard Go formatting (`go fmt`)
- Use `go vet` for static analysis
- Prefer early returns over nested conditionals
- Keep functions small and focused

### Domain Layer

The `internal/domain` package must:
- Have 100% test coverage
- Use value receivers and return new instances (immutability)
- Have no external dependencies
- Contain pure business logic only

```go
// Good - immutable
func (e Entry) MarkDone() Entry {
    e.Type = Done
    return e
}

// Avoid - mutation
func (e *Entry) MarkDone() {
    e.Type = Done
}
```

### Repository Pattern

All data mutations use event sourcing:

1. Close current version (`SET valid_to = NOW()`)
2. Insert new version with incremented version number
3. Use transactions for atomicity

Never do in-place updates:
```go
// Wrong - destroys history
db.Exec("UPDATE entries SET content = ? WHERE id = ?", content, id)

// Correct - event sourcing
tx.Exec("UPDATE entries SET valid_to = ? WHERE entity_id = ? AND valid_to IS NULL", now, entityID)
tx.Exec("INSERT INTO entries (..., version, op_type) VALUES (..., ?, 'UPDATE')", nextVersion)
```

## Database Safety

**Never run the application without specifying a test database.**

```bash
# Correct
./bujo --db-path ./test.db add "test entry"

# Wrong - may corrupt production data
./bujo add "test entry"
```

Tests should use in-memory databases (`:memory:`).

## Architecture

bujo uses hexagonal architecture:

```
cmd/bujo/cmd/          CLI commands (Cobra adapter)
internal/
  domain/              Core business logic (100% TDD)
  service/             Application services
  repository/sqlite/   SQLite implementations
  adapter/             External integrations
  tui/                 Terminal UI (Bubbletea)
frontend/              Desktop app (React/TypeScript)
```

See [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) for details.

## Submitting Changes

1. Create a feature branch:
   ```bash
   git checkout -b feature/your-feature
   ```

2. Make changes following TDD

3. Ensure all tests pass:
   ```bash
   go test ./...
   ```

4. Commit with clear messages:
   ```bash
   git commit -m "Add entry cancellation feature"
   ```

5. Push and create a pull request

## Pull Request Guidelines

- Include tests for all new functionality
- Update documentation if behavior changes
- Keep changes focused and atomic
- Reference related issues in PR description

## Reporting Issues

Open an issue on GitHub with:
- Clear description of the problem
- Steps to reproduce
- Expected vs actual behavior
- Version information (`bujo version`)

## Questions

For questions about contributing, open a GitHub issue or discussion.
