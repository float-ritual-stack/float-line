# float-rw-client

A beautiful TUI (Terminal User Interface) client for Readwise built with Go and the Bubble Tea framework.

## Features

- 📚 Browse your books and articles
- 🔍 View and search highlights
- ✏️ Edit highlight notes with a rich markdown editor
- 🎨 Beautiful terminal UI with syntax highlighting
- ⚡ Fast and responsive
- 🖼️ Split panel view for better context while browsing and editing

## Installation

```bash
go install github.com/evanschultz/float-rw-client/cmd/float-rw@latest
```

Or build from source:

```bash
git clone https://github.com/evanschultz/float-rw-client
cd float-rw-client
go build -o float-rw ./cmd/float-rw
```

## Configuration

### API Token

Get your Readwise API token from: https://readwise.io/access_token

Set it via environment variable:
```bash
export READWISE_TOKEN="your-token-here"
```

Or create a config file at `~/.float-rw.yaml`:
```yaml
token: your-token-here
```

## Usage

Launch the TUI:
```bash
float-rw tui
```

Or with a token flag:
```bash
float-rw tui --token="your-token-here"
```

Use the classic single-pane view:
```bash
float-rw tui --split=false
```

Export highlights:
```bash
# Export all highlights to markdown
float-rw export -o highlights.md

# Export highlights from a specific book to JSON
float-rw export --book-id=123 -f json -o book-highlights.json

# Export to stdout
float-rw export -f markdown
```

### Keyboard Shortcuts

- `↑/k` - Move up
- `↓/j` - Move down
- `←/h` - Focus left pane
- `→/l` - Focus right pane
- `Tab` - Cycle through panes
- `Enter` - Select item
- `Esc` - Go back
- `e` - Edit note (in highlight detail view)
- `/` - Search in lists
- `r` - Refresh current view
- `?` - Show help
- `q` - Quit

In the editor:
- `Ctrl+S` - Save note
- `Esc` - Cancel editing

## Development

### Prerequisites

- Go 1.22+
- Readwise API token

### Building

```bash
go mod download
go build -o float-rw ./cmd/float-rw
```

### Testing

```bash
go test ./...
```

## Architecture

The project uses:
- **Bubble Tea** - Terminal UI framework
- **Bubbles** - Pre-built TUI components
- **Glamour** - Markdown rendering
- **Lipgloss** - Terminal styling
- **Cobra** - CLI framework
- **Viper** - Configuration management

## Roadmap

- [x] Basic book and highlight browsing
- [x] Highlight detail view with markdown rendering
- [x] Rich markdown editor for notes
- [ ] Search functionality
- [ ] Tag management
- [x] Export highlights
- [ ] Daily review integration
- [ ] Offline caching

## License

MIT