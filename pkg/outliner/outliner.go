package outliner

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// OutlineNode represents a single line in the outline
type OutlineNode struct {
	Text  string
	Level int // 0 = root level, 1 = indented once, etc.
}

// Outliner is the main component
type Outliner struct {
	lines     []OutlineNode
	cursor    int // which line we're editing
	cursorPos int // position within the current line
	width     int
	height    int
	focused   bool

	// Styles
	bulletStyle    lipgloss.Style
	textStyle      lipgloss.Style
	cursorStyle    lipgloss.Style
	focusedStyle   lipgloss.Style
	unfocusedStyle lipgloss.Style
}

// New creates a new outliner
func New() Outliner {
	return Outliner{
		lines: []OutlineNode{
			{Text: "", Level: 0}, // Start with one empty line
		},
		cursor:    0,
		cursorPos: 0,

		// Default styles
		bulletStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
		textStyle:   lipgloss.NewStyle(),
		cursorStyle: lipgloss.NewStyle().Background(lipgloss.Color("62")),
		focusedStyle: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(1),
		unfocusedStyle: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(1),
	}
}

// Focus gives focus to the outliner
func (o *Outliner) Focus() tea.Cmd {
	o.focused = true
	return nil
}

// Blur removes focus from the outliner
func (o *Outliner) Blur() tea.Cmd {
	o.focused = false
	return nil
}

// Focused returns whether the outliner has focus
func (o Outliner) Focused() bool {
	return o.focused
}

// SetSize sets the dimensions of the outliner
func (o *Outliner) SetSize(width, height int) {
	o.width = width
	o.height = height
}

// Update handles key presses and other messages
func (o Outliner) Update(msg tea.Msg) (Outliner, tea.Cmd) {
	if !o.focused {
		return o, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab":
			// CORE FEATURE: Indent current line
			if o.cursor < len(o.lines) {
				o.lines[o.cursor].Level++
				// Limit max indentation
				if o.lines[o.cursor].Level > 6 {
					o.lines[o.cursor].Level = 6
				}
			}

		case "shift+tab":
			// CORE FEATURE: Outdent current line
			if o.cursor < len(o.lines) && o.lines[o.cursor].Level > 0 {
				o.lines[o.cursor].Level--
			}

		case "enter":
			// Create new line at same level
			if o.cursor < len(o.lines) {
				currentLevel := o.lines[o.cursor].Level
				newNode := OutlineNode{Text: "", Level: currentLevel}

				// Insert after current line
				o.lines = append(o.lines[:o.cursor+1], append([]OutlineNode{newNode}, o.lines[o.cursor+1:]...)...)
				o.cursor++
				o.cursorPos = 0
			}

		case "up", "ctrl+p":
			// Move to previous line
			if o.cursor > 0 {
				o.cursor--
				// Adjust cursor position if new line is shorter
				if o.cursorPos > len(o.lines[o.cursor].Text) {
					o.cursorPos = len(o.lines[o.cursor].Text)
				}
			}

		case "down", "ctrl+n":
			// Move to next line
			if o.cursor < len(o.lines)-1 {
				o.cursor++
				// Adjust cursor position if new line is shorter
				if o.cursorPos > len(o.lines[o.cursor].Text) {
					o.cursorPos = len(o.lines[o.cursor].Text)
				}
			}

		case "left", "ctrl+b":
			// Move cursor left within line
			if o.cursorPos > 0 {
				o.cursorPos--
			}

		case "right", "ctrl+f":
			// Move cursor right within line
			if o.cursor < len(o.lines) && o.cursorPos < len(o.lines[o.cursor].Text) {
				o.cursorPos++
			}

		case "home", "ctrl+a":
			// Move to beginning of line
			o.cursorPos = 0

		case "end", "ctrl+e":
			// Move to end of line
			if o.cursor < len(o.lines) {
				o.cursorPos = len(o.lines[o.cursor].Text)
			}

		case "backspace", "ctrl+h":
			// Delete character before cursor
			if o.cursor < len(o.lines) {
				if o.cursorPos > 0 {
					// Delete character in current line
					line := &o.lines[o.cursor]
					line.Text = line.Text[:o.cursorPos-1] + line.Text[o.cursorPos:]
					o.cursorPos--
				} else if o.cursor > 0 {
					// Merge with previous line
					prevLine := &o.lines[o.cursor-1]
					currentLine := o.lines[o.cursor]
					o.cursorPos = len(prevLine.Text)
					prevLine.Text += currentLine.Text
					// Remove current line
					o.lines = append(o.lines[:o.cursor], o.lines[o.cursor+1:]...)
					o.cursor--
				}
			}

		case "delete", "ctrl+d":
			// Delete character at cursor
			if o.cursor < len(o.lines) {
				line := &o.lines[o.cursor]
				if o.cursorPos < len(line.Text) {
					line.Text = line.Text[:o.cursorPos] + line.Text[o.cursorPos+1:]
				}
			}

		default:
			// Handle regular character input
			if len(msg.String()) == 1 {
				char := msg.String()
				if o.cursor < len(o.lines) {
					line := &o.lines[o.cursor]
					// Insert character at cursor position
					line.Text = line.Text[:o.cursorPos] + char + line.Text[o.cursorPos:]
					o.cursorPos++
				}
			}
		}
	}

	return o, nil
}

// View renders the outliner
func (o Outliner) View() string {
	if len(o.lines) == 0 {
		return ""
	}

	// Debug: show line count
	var content strings.Builder
	content.WriteString(fmt.Sprintf("Lines: %d, Cursor: %d\n", len(o.lines), o.cursor))

	for i, line := range o.lines {
		// Create indentation
		indent := strings.Repeat("  ", line.Level) // 2 spaces per level

		// Create bullet
		bullet := "•"
		if line.Level > 0 {
			bullet = "◦" // Different bullet for sub-items
		}

		// Build the line
		lineText := indent + o.bulletStyle.Render(bullet+" ") + line.Text

		// Add cursor if this is the current line
		if i == o.cursor && o.focused {
			// Simple cursor - just add a pipe character
			bulletAndSpace := o.bulletStyle.Render(bullet + " ")
			cursorPos := o.cursorPos
			if cursorPos > len(line.Text) {
				cursorPos = len(line.Text)
			}

			// Build line with cursor
			beforeCursor := line.Text[:cursorPos]
			afterCursor := line.Text[cursorPos:]
			lineText = indent + bulletAndSpace + beforeCursor + "|" + afterCursor
		}

		content.WriteString(lineText)
		if i < len(o.lines)-1 {
			content.WriteString("\n")
		}
	}

	// Apply border style based on focus
	if o.focused {
		return o.focusedStyle.Width(o.width - 4).Height(o.height - 4).Render(content.String())
	}
	return o.unfocusedStyle.Width(o.width - 4).Height(o.height - 4).Render(content.String())
}

// GetContent returns the current outline as a string
func (o Outliner) GetContent() string {
	var result strings.Builder
	for _, line := range o.lines {
		indent := strings.Repeat("  ", line.Level)
		result.WriteString(indent + "• " + line.Text + "\n")
	}
	return result.String()
}

// SetContent loads content into the outliner
func (o *Outliner) SetContent(content string) {
	lines := strings.Split(content, "\n")
	o.lines = make([]OutlineNode, 0, len(lines))

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		// Count leading spaces to determine level
		level := 0
		trimmed := line
		for strings.HasPrefix(trimmed, "  ") {
			level++
			trimmed = trimmed[2:]
		}

		// Remove bullet if present
		trimmed = strings.TrimPrefix(trimmed, "• ")
		trimmed = strings.TrimPrefix(trimmed, "◦ ")

		o.lines = append(o.lines, OutlineNode{
			Text:  trimmed,
			Level: level,
		})
	}

	// Ensure we have at least one line
	if len(o.lines) == 0 {
		o.lines = []OutlineNode{{Text: "", Level: 0}}
	}

	o.cursor = 0
	o.cursorPos = 0
}
