# CRUSH.md - float-rw-client

## Build/Test Commands
```bash
# Build
go build -o float-rw ./cmd/float-rw
go build -race -o float-rw ./cmd/float-rw  # with race detection

# Test
go test ./...
go test -v -run TestSpecificFunction ./pkg/...
go test -cover ./...
go test -race ./...

# Format and lint
go fmt ./...
go mod tidy
golangci-lint run

# Run the TUI
./float-rw tui
./float-rw tui --token="your-token"
READWISE_TOKEN="your-token" ./float-rw tui
```

## Code Style Guidelines

### Imports
- Group imports: stdlib, third-party, local packages
- Use absolute imports for clarity
- Sort alphabetically within groups

### Error Handling
- Go: Always check errors, wrap with context using fmt.Errorf
- Python: Use explicit exception types, avoid bare except
- Return early on errors to reduce nesting

### Naming Conventions
- Go: CamelCase for exports, camelCase for private
- Python: snake_case for functions/variables, PascalCase for classes
- Descriptive names over comments

### Type Annotations
- Python: Use type hints for all function signatures
- Go: Define interfaces before implementations
- Document complex types with comments

### Project Structure
- Keep packages focused and single-purpose
- Separate concerns: CLI, business logic, data models
- Use dependency injection for testability

### Git Workflow
- Branch format: type/issue-description (e.g., feat/42-add-parser)
- Commit format: type(scope): message (#issue)
- Small, focused PRs addressing single issues