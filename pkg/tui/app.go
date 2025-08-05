package tui

import (
	"fmt"
	"net/url"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/evanschultz/float-rw-client/pkg/api"
	"github.com/evanschultz/float-rw-client/pkg/models"
	"github.com/evanschultz/float-rw-client/pkg/tui/components"
)

type state int

const (
	stateBooks state = iota
	stateHighlights
	stateHighlightDetail
	stateEditNote
)

type Model struct {
	api    *api.Client
	state  state
	width  int
	height int

	// Components
	bookList      list.Model
	highlightList list.Model
	viewport      viewport.Model
	help          help.Model
	editor        components.MarkdownEditor

	// Data
	books            []models.Book
	highlights       []models.Highlight
	currentBook      *models.Book
	currentHighlight *models.Highlight
	nextPageURL      string

	// UI state
	loading bool
	saving  bool
	err     error
}

type keyMap struct {
	Up      key.Binding
	Down    key.Binding
	Enter   key.Binding
	Back    key.Binding
	Edit    key.Binding
	Refresh key.Binding
	Help    key.Binding
	Quit    key.Binding
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "down"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "select"),
	),
	Back: key.NewBinding(
		key.WithKeys("esc", "backspace"),
		key.WithHelp("esc", "back"),
	),
	Edit: key.NewBinding(
		key.WithKeys("e"),
		key.WithHelp("e", "edit note"),
	),
	Refresh: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "refresh"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "help"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
}

func NewModel(apiClient *api.Client) Model {
	m := Model{
		api:   apiClient,
		state: stateBooks,
		help:  help.New(),
	}

	// Initialize lists
	m.bookList = list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	m.bookList.Title = "Your Books"
	m.bookList.SetShowHelp(false)
	m.bookList.SetFilteringEnabled(true)

	m.highlightList = list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	m.highlightList.Title = "Highlights"
	m.highlightList.SetShowHelp(false)
	m.highlightList.SetFilteringEnabled(true)

	// Initialize viewport
	m.viewport = viewport.New(0, 0)

	// Initialize markdown editor
	m.editor = components.NewMarkdownEditor()

	return m
}

func (m Model) Init() tea.Cmd {
	return m.loadBooks()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateSizes()

	case tea.KeyMsg:
		switch m.state {
		case stateBooks:
			switch {
			case key.Matches(msg, keys.Quit):
				return m, tea.Quit
			case key.Matches(msg, keys.Enter):
				if i, ok := m.bookList.SelectedItem().(bookItem); ok {
					m.currentBook = &i.book
					return m, m.loadHighlights(i.book.ID)
				}
			case key.Matches(msg, keys.Refresh):
				return m, m.loadBooks()
			}

		case stateHighlights:
			switch {
			case key.Matches(msg, keys.Quit):
				return m, tea.Quit
			case key.Matches(msg, keys.Back):
				m.state = stateBooks
			case key.Matches(msg, keys.Enter):
				if i, ok := m.highlightList.SelectedItem().(highlightItem); ok {
					m.currentHighlight = &i.highlight
					m.state = stateHighlightDetail
					return m, m.renderHighlightDetail()
				}
			}

		case stateHighlightDetail:
			switch {
			case key.Matches(msg, keys.Quit):
				return m, tea.Quit
			case key.Matches(msg, keys.Back):
				m.state = stateHighlights
			case key.Matches(msg, keys.Edit):
				m.state = stateEditNote
				// Set the current note value in the editor
				if m.currentHighlight != nil {
					m.editor.SetValue(m.currentHighlight.Note)
				}
				return m, m.editor.Init()
			}

		case stateEditNote:
			// Handle editor-specific keys
			switch {
			case key.Matches(msg, components.DefaultKeyMap.Save):
				// Save the note
				if m.currentHighlight != nil {
					m.currentHighlight.Note = m.editor.Value()
					// TODO: Call API to update the highlight
				}
				m.state = stateHighlightDetail
				return m, m.renderHighlightDetail()
			case key.Matches(msg, components.DefaultKeyMap.Cancel):
				// Cancel editing
				m.state = stateHighlightDetail
			}
		}

	case booksLoadedMsg:
		m.loading = false
		m.books = msg.books
		items := make([]list.Item, len(m.books))
		for i, book := range m.books {
			items[i] = bookItem{book: book}
		}
		m.bookList.SetItems(items)

	case highlightsLoadedMsg:
		m.loading = false
		m.highlights = msg.highlights
		m.nextPageURL = msg.nextPageURL
		items := make([]list.Item, len(m.highlights))
		for i, highlight := range m.highlights {
			items[i] = highlightItem{highlight: highlight}
		}
		m.highlightList.SetItems(items)
		m.state = stateHighlights

	case highlightRenderedMsg:
		m.viewport.SetContent(msg.content)

	case errMsg:
		m.err = msg.err
		m.loading = false

	case components.SaveMsg:
		// Save the note
		if m.currentHighlight != nil {
			m.currentHighlight.Note = msg.Content
			m.saving = true
			cmds = append(cmds, m.updateHighlightNote())
		}

	case components.CancelMsg:
		// Cancel editing
		m.state = stateHighlightDetail

	case highlightSavedMsg:
		m.saving = false
		m.state = stateHighlightDetail
		cmds = append(cmds, m.renderHighlightDetail())
	}

	// Update components
	switch m.state {
	case stateBooks:
		newList, cmd := m.bookList.Update(msg)
		m.bookList = newList
		cmds = append(cmds, cmd)

	case stateHighlights:
		newList, cmd := m.highlightList.Update(msg)
		m.highlightList = newList
		cmds = append(cmds, cmd)

	case stateHighlightDetail:
		newViewport, cmd := m.viewport.Update(msg)
		m.viewport = newViewport
		cmds = append(cmds, cmd)

	case stateEditNote:
		newEditor, cmd := m.editor.Update(msg)
		m.editor = newEditor
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\n\nPress q to quit.", m.err)
	}

	if m.loading {
		return "Loading..."
	}

	if m.saving {
		return "Saving note..."
	}

	var content string
	var helpText string

	switch m.state {
	case stateBooks:
		content = m.bookList.View()
		helpText = "enter: select • /: search • r: refresh • q: quit"

	case stateHighlights:
		content = m.highlightList.View()
		helpText = "enter: view • /: search • esc: back • q: quit"
		if m.currentBook != nil {
			status := fmt.Sprintf("%d highlights", len(m.highlights))
			if m.nextPageURL != "" {
				status += " (more available)"
			}
			helpText = status + " • " + helpText
		}

	case stateHighlightDetail:
		content = m.viewport.View()
		helpText = "e: edit note • esc: back • q: quit"

	case stateEditNote:
		return m.editor.View()

	default:
		return "Unknown state"
	}

	// Add help text at the bottom
	if helpText != "" {
		helpStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Align(lipgloss.Center).
			Width(m.width)

		return lipgloss.JoinVertical(
			lipgloss.Top,
			content,
			helpStyle.Render(helpText),
		)
	}

	return content
}

func (m *Model) updateSizes() {
	h, v := lipgloss.NewStyle().GetFrameSize()
	m.bookList.SetSize(m.width-h, m.height-v)
	m.highlightList.SetSize(m.width-h, m.height-v)
	m.viewport.Width = m.width
	m.viewport.Height = m.height
	m.editor.SetSize(m.width, m.height)
}

// Commands
func (m Model) loadBooks() tea.Cmd {
	return func() tea.Msg {
		books, err := m.api.GetBooks(nil)
		if err != nil {
			return errMsg{err}
		}
		return booksLoadedMsg{books: books.Results}
	}
}

func (m Model) loadHighlights(bookID int) tea.Cmd {
	return func() tea.Msg {
		params := url.Values{}
		params.Set("book_id", fmt.Sprintf("%d", bookID))
		highlights, err := m.api.GetHighlights(params)
		if err != nil {
			return errMsg{err}
		}
		return highlightsLoadedMsg{
			highlights:  highlights.Results,
			nextPageURL: highlights.Next,
		}
	}
}

func (m Model) renderHighlightDetail() tea.Cmd {
	return func() tea.Msg {
		if m.currentHighlight == nil {
			return nil
		}

		content := fmt.Sprintf("# Highlight\n\n%s\n\n", m.currentHighlight.Text)

		if m.currentHighlight.Note != "" {
			content += fmt.Sprintf("## Note\n\n%s\n\n", m.currentHighlight.Note)
		}

		if m.currentBook != nil {
			content += fmt.Sprintf("---\n\n**Book:** %s\n\n**Author:** %s\n\n",
				m.currentBook.Title, m.currentBook.Author)
		}

		// Render markdown
		renderer, _ := glamour.NewTermRenderer(
			glamour.WithAutoStyle(),
			glamour.WithWordWrap(m.width),
		)

		rendered, err := renderer.Render(content)
		if err != nil {
			return errMsg{err}
		}

		return highlightRenderedMsg{content: rendered}
	}
}

func (m Model) updateHighlightNote() tea.Cmd {
	return func() tea.Msg {
		if m.currentHighlight == nil {
			return nil
		}

		// Update the highlight via API
		update := models.HighlightUpdate{
			Note: m.currentHighlight.Note,
		}
		_, err := m.api.UpdateHighlight(m.currentHighlight.ID, update)
		if err != nil {
			return errMsg{err}
		}

		// Update the highlight in the local list
		for i, h := range m.highlights {
			if h.ID == m.currentHighlight.ID {
				m.highlights[i].Note = m.currentHighlight.Note
				break
			}
		}

		return highlightSavedMsg{}
	}
}
