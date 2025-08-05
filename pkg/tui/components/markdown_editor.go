package components

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

type EditorMode int

const (
	ModeEdit EditorMode = iota
	ModePreview
)

type MarkdownEditor struct {
	textarea textarea.Model
	preview  string
	mode     EditorMode
	width    int
	height   int
	renderer *glamour.TermRenderer

	// Styles
	borderStyle lipgloss.Style
	titleStyle  lipgloss.Style
	modeStyle   lipgloss.Style
	helpStyle   lipgloss.Style
}

type KeyMap struct {
	Save         key.Binding
	Cancel       key.Binding
	ToggleMode   key.Binding
	Bold         key.Binding
	Italic       key.Binding
	Link         key.Binding
	Quote        key.Binding
	Code         key.Binding
	BulletList   key.Binding
	NumberedList key.Binding
}

var DefaultKeyMap = KeyMap{
	Save: key.NewBinding(
		key.WithKeys("ctrl+s"),
		key.WithHelp("ctrl+s", "save"),
	),
	Cancel: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "cancel"),
	),
	ToggleMode: key.NewBinding(
		key.WithKeys("ctrl+p"),
		key.WithHelp("ctrl+p", "toggle preview"),
	),
	Bold: key.NewBinding(
		key.WithKeys("ctrl+b"),
		key.WithHelp("ctrl+b", "bold"),
	),
	Italic: key.NewBinding(
		key.WithKeys("ctrl+i"),
		key.WithHelp("ctrl+i", "italic"),
	),
	Link: key.NewBinding(
		key.WithKeys("ctrl+k"),
		key.WithHelp("ctrl+k", "link"),
	),
	Quote: key.NewBinding(
		key.WithKeys("ctrl+q"),
		key.WithHelp("ctrl+q", "quote"),
	),
	Code: key.NewBinding(
		key.WithKeys("ctrl+e"),
		key.WithHelp("ctrl+e", "code"),
	),
	BulletList: key.NewBinding(
		key.WithKeys("ctrl+u"),
		key.WithHelp("ctrl+u", "bullet list"),
	),
	NumberedList: key.NewBinding(
		key.WithKeys("ctrl+o"),
		key.WithHelp("ctrl+o", "numbered list"),
	),
}

func NewMarkdownEditor() MarkdownEditor {
	ta := textarea.New()
	ta.Placeholder = "Write your note here..."
	ta.ShowLineNumbers = false
	ta.Focus()

	renderer, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(80),
	)

	return MarkdownEditor{
		textarea: ta,
		mode:     ModeEdit,
		renderer: renderer,
		borderStyle: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")),
		titleStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("230")),
		modeStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("244")).
			Italic(true),
		helpStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")),
	}
}

func (m MarkdownEditor) Init() tea.Cmd {
	return textarea.Blink
}

func (m MarkdownEditor) Update(msg tea.Msg) (MarkdownEditor, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateSize()

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, DefaultKeyMap.ToggleMode):
			if m.mode == ModeEdit {
				m.mode = ModePreview
				m.updatePreview()
			} else {
				m.mode = ModeEdit
			}

		case key.Matches(msg, DefaultKeyMap.Bold):
			if m.mode == ModeEdit {
				m.insertMarkdown("**", "**")
			}

		case key.Matches(msg, DefaultKeyMap.Italic):
			if m.mode == ModeEdit {
				m.insertMarkdown("*", "*")
			}

		case key.Matches(msg, DefaultKeyMap.Code):
			if m.mode == ModeEdit {
				m.insertMarkdown("`", "`")
			}

		case key.Matches(msg, DefaultKeyMap.Quote):
			if m.mode == ModeEdit {
				m.insertLinePrefix("> ")
			}

		case key.Matches(msg, DefaultKeyMap.BulletList):
			if m.mode == ModeEdit {
				m.insertLinePrefix("- ")
			}

		case key.Matches(msg, DefaultKeyMap.NumberedList):
			if m.mode == ModeEdit {
				m.insertLinePrefix("1. ")
			}

		case key.Matches(msg, DefaultKeyMap.Link):
			if m.mode == ModeEdit {
				m.insertMarkdown("[", "](url)")
			}

		default:
			if m.mode == ModeEdit {
				var cmd tea.Cmd
				m.textarea, cmd = m.textarea.Update(msg)
				cmds = append(cmds, cmd)
			}
		}
	}

	return m, tea.Batch(cmds...)
}

func (m MarkdownEditor) View() string {
	var content string
	var modeText string

	if m.mode == ModeEdit {
		content = m.textarea.View()
		modeText = "Edit Mode"
	} else {
		content = m.preview
		modeText = "Preview Mode"
	}

	// Title bar
	title := m.titleStyle.Render("Markdown Editor")
	mode := m.modeStyle.Render(modeText)
	titleBar := lipgloss.JoinHorizontal(
		lipgloss.Top,
		title,
		strings.Repeat(" ", max(0, m.width-lipgloss.Width(title)-lipgloss.Width(mode)-4)),
		mode,
	)

	// Help text
	var helpText string
	if m.mode == ModeEdit {
		helpText = m.helpStyle.Render(
			"ctrl+p: preview • ctrl+s: save • esc: cancel • ctrl+b: bold • ctrl+i: italic",
		)
	} else {
		helpText = m.helpStyle.Render(
			"ctrl+p: edit • ctrl+s: save • esc: cancel",
		)
	}

	// Combine everything
	return lipgloss.JoinVertical(
		lipgloss.Left,
		titleBar,
		m.borderStyle.Width(m.width-2).Height(m.height-4).Render(content),
		helpText,
	)
}

func (m *MarkdownEditor) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.updateSize()
}

func (m *MarkdownEditor) SetValue(value string) {
	m.textarea.SetValue(value)
	if m.mode == ModePreview {
		m.updatePreview()
	}
}

func (m MarkdownEditor) Value() string {
	return m.textarea.Value()
}

func (m *MarkdownEditor) updateSize() {
	m.textarea.SetWidth(m.width - 4)
	m.textarea.SetHeight(m.height - 6)

	if m.renderer != nil {
		m.renderer, _ = glamour.NewTermRenderer(
			glamour.WithAutoStyle(),
			glamour.WithWordWrap(m.width-4),
		)
	}
}

func (m *MarkdownEditor) updatePreview() {
	if m.renderer != nil {
		rendered, err := m.renderer.Render(m.textarea.Value())
		if err != nil {
			m.preview = "Error rendering markdown: " + err.Error()
		} else {
			m.preview = rendered
		}
	}
}

func (m *MarkdownEditor) insertMarkdown(prefix, suffix string) {
	// Get current cursor position
	value := m.textarea.Value()
	// For now, just append at the end since CursorOffset is not available
	pos := len(value)

	// Insert markdown syntax
	newValue := value[:pos] + prefix + suffix + value[pos:]
	m.textarea.SetValue(newValue)

	// Move cursor between the markers
	m.textarea.SetCursor(pos + len(prefix))
}

func (m *MarkdownEditor) insertLinePrefix(prefix string) {
	value := m.textarea.Value()
	// For now, just prepend to the current line
	lines := strings.Split(value, "\n")
	if len(lines) > 0 {
		lines[len(lines)-1] = prefix + lines[len(lines)-1]
		newValue := strings.Join(lines, "\n")
		m.textarea.SetValue(newValue)
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Messages
type SaveMsg struct {
	Content string
}

type CancelMsg struct{}
