# AGENTS.md - float-rw-client

## Build Commands
- `go build -o float-rw ./cmd/float-rw` - Build binary
- `go test ./...` - Run all tests
- `go test ./pkg/api` - Run specific package tests
- `go mod download` - Download dependencies
- `go mod tidy` - Clean up dependencies

## Code Style Guidelines
- Use standard Go formatting (`gofmt`)
- Package imports: stdlib, third-party, local (separated by blank lines)
- Error handling: Always check and return errors, use `fmt.Errorf` for wrapping
- Naming: Use camelCase for unexported, PascalCase for exported
- Constants: Use ALL_CAPS with underscores for package-level constants
- Struct initialization: Use field names for clarity
- Interface naming: Single method interfaces end with -er (e.g., `Reader`)

## Architecture
- `/pkg/api` - Readwise API client
- `/pkg/models` - Data models
- `/pkg/tui` - Bubble Tea TUI components
- `/cmd/float-rw` - CLI entry point
- Uses Bubble Tea framework for TUI, Cobra for CLI, Viper for config

## Dependencies
- Bubble Tea ecosystem (bubbletea, bubbles, lipgloss, glamour)
- Cobra + Viper for CLI/config
- Standard library for HTTP client