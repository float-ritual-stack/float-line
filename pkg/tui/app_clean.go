package tui

import (
	"fmt"
	"net/url"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/evanschultz/float-rw-client/pkg/api"
	"github.com/evanschultz/float-rw-client/pkg/models"
	"github.com/evanschultz/float-rw-client/pkg/outliner"
)

// Simple focus states - just 3 panels
type Focus int

const (
	FocusBooks Focus = iota
	FocusHighlights
	FocusDetail
)

// Edit modes
type EditMode int

const (
	ModeView EditMode = iota
	ModeEdit
)

// Clean model with minimal state
type CleanModel struct {
	api    *api.Client
	width  int
	height int

	// Focus
	focus Focus

	// Data
	books            []models.Book
	highlights       []models.Highlight
	currentBook      *models.Book
	currentHighlight *models.Highlight

	// Components
	bookList      list.Model
	highlightList list.Model
	detailView    viewport.Model
	noteOutliner  outliner.Outliner

	// UI state
	loading  bool
	err      error
	editMode EditMode
}

func NewCleanModel(apiClient *api.Client) CleanModel {
	// Book list
	bookList := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	bookList.Title = "📚 Books"
	bookList.SetShowHelp(false)
	bookList.SetFilteringEnabled(true)
	bookList.DisableQuitKeybindings()

	// Highlight list
	highlightList := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	highlightList.Title = "📝 Highlights"
	highlightList.SetShowHelp(false)
	highlightList.SetFilteringEnabled(true)
	highlightList.DisableQuitKeybindings()

	// Detail viewport
	detailView := viewport.New(0, 0)

	// Note outliner
	noteOutliner := outliner.New()

	return CleanModel{
		api:           apiClient,
		focus:         FocusBooks,
		bookList:      bookList,
		highlightList: highlightList,
		detailView:    detailView,
		noteOutliner:  noteOutliner,
		editMode:      ModeView,
	}
}

func (m CleanModel) Init() tea.Cmd {
	return m.loadBooks()
}

func (m CleanModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateSizes()

	case tea.KeyMsg:
		// In edit mode, handle only specific keys and pass everything else to outliner
		if m.editMode == ModeEdit {
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			case "esc":
				// Exit edit mode
				m.editMode = ModeView
				m.noteOutliner.Blur()
			case "ctrl+s":
				// Save note
				if m.currentHighlight != nil {
					m.currentHighlight.Note = m.noteOutliner.GetContent()
					m.editMode = ModeView
					m.noteOutliner.Blur()
					// TODO: Save to API
					return m, m.renderHighlightDetail()
				}
			default:
				// ALL other keys go to the outliner
				newOutliner, cmd := m.noteOutliner.Update(msg)
				m.noteOutliner = newOutliner
				cmds = append(cmds, cmd)
			}
		} else {
			// View mode - normal key handling
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit

			case "tab":
				m.cycleFocus()

			case "left", "h":
				m.focusLeft()

			case "right", "l":
				m.focusRight()

			case "enter":
				return m.handleEnter()

			case "esc":
				m.focusLeft()

			case "e":
				// Enter edit mode when in detail panel
				if m.focus == FocusDetail && m.currentHighlight != nil {
					m.editMode = ModeEdit
					m.noteOutliner.Focus()
					// Load current note into outliner
					if m.currentHighlight.Note != "" {
						m.noteOutliner.SetContent(m.currentHighlight.Note)
					}
				}

			default:
				// Let the focused component handle other keys
				switch m.focus {
				case FocusBooks:
					newList, cmd := m.bookList.Update(msg)
					m.bookList = newList
					cmds = append(cmds, cmd)

				case FocusHighlights:
					newList, cmd := m.highlightList.Update(msg)
					m.highlightList = newList
					cmds = append(cmds, cmd)

				case FocusDetail:
					newView, cmd := m.detailView.Update(msg)
					m.detailView = newView
					cmds = append(cmds, cmd)
				}
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
		items := make([]list.Item, len(m.highlights))
		for i, highlight := range m.highlights {
			items[i] = highlightItem{highlight: highlight}
		}
		m.highlightList.SetItems(items)
		// Don't auto-focus - let user navigate manually

	case highlightRenderedMsg:
		m.detailView.SetContent(msg.content)

	case errMsg:
		m.err = msg.err
		m.loading = false
	}

	return m, tea.Batch(cmds...)
}

func (m CleanModel) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\n\nPress q to quit.", m.err)
	}

	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	// Calculate layout - always 3 columns when we have data
	bookWidth := 30
	highlightWidth := 40
	detailWidth := m.width - bookWidth - highlightWidth - 6 // Account for borders

	// Ensure minimum widths
	if detailWidth < 40 {
		bookWidth = 25
		highlightWidth = 35
		detailWidth = m.width - bookWidth - highlightWidth - 6
	}

	contentHeight := m.height - 3 // Account for help text

	// Styles
	focusedStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(0, 1)

	unfocusedStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 1)

	// Book panel
	bookContent := m.bookList.View()
	if m.loading && m.focus == FocusBooks {
		bookContent = "Loading books..."
	}

	var bookPanel string
	if m.focus == FocusBooks {
		bookPanel = focusedStyle.Width(bookWidth - 4).Height(contentHeight - 2).Render(bookContent)
	} else {
		bookPanel = unfocusedStyle.Width(bookWidth - 4).Height(contentHeight - 2).Render(bookContent)
	}

	// Highlight panel (show if we have a book)
	var highlightPanel string
	if m.currentBook != nil {
		highlightContent := m.highlightList.View()
		if m.loading && m.focus == FocusHighlights {
			highlightContent = fmt.Sprintf("Loading highlights for %s...", m.currentBook.Title)
		}

		if m.focus == FocusHighlights {
			highlightPanel = focusedStyle.Width(highlightWidth - 4).Height(contentHeight - 2).Render(highlightContent)
		} else {
			highlightPanel = unfocusedStyle.Width(highlightWidth - 4).Height(contentHeight - 2).Render(highlightContent)
		}
	} else {
		// Empty placeholder
		highlightPanel = unfocusedStyle.Width(highlightWidth - 4).Height(contentHeight - 2).Render("Select a book to see highlights")
	}

	// Detail panel (show if we have a highlight)
	var detailPanel string
	if m.currentHighlight != nil {
		var detailContent string

		if m.editMode == ModeEdit {
			// Show outliner for editing
			m.noteOutliner.SetSize(detailWidth-4, contentHeight-2)
			detailContent = m.noteOutliner.View()
		} else {
			// Show rendered view
			detailContent = m.detailView.View()
		}

		if m.focus == FocusDetail || m.editMode == ModeEdit {
			detailPanel = focusedStyle.Width(detailWidth - 4).Height(contentHeight - 2).Render(detailContent)
		} else {
			detailPanel = unfocusedStyle.Width(detailWidth - 4).Height(contentHeight - 2).Render(detailContent)
		}
	} else {
		// Empty placeholder
		detailPanel = unfocusedStyle.Width(detailWidth - 4).Height(contentHeight - 2).Render("Select a highlight to see details")
	}

	// Join panels
	content := lipgloss.JoinHorizontal(lipgloss.Top, bookPanel, highlightPanel, detailPanel)

	// Help text
	helpText := m.getHelpText()
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

// Focus management
func (m *CleanModel) cycleFocus() {
	switch m.focus {
	case FocusBooks:
		if m.currentBook != nil {
			m.focus = FocusHighlights
		}
	case FocusHighlights:
		if m.currentHighlight != nil {
			m.focus = FocusDetail
		} else {
			m.focus = FocusBooks
		}
	case FocusDetail:
		m.focus = FocusBooks
	}
}

func (m *CleanModel) focusLeft() {
	switch m.focus {
	case FocusHighlights:
		m.focus = FocusBooks
	case FocusDetail:
		if m.currentBook != nil {
			m.focus = FocusHighlights
		} else {
			m.focus = FocusBooks
		}
	}
}

func (m *CleanModel) focusRight() {
	switch m.focus {
	case FocusBooks:
		if m.currentBook != nil {
			m.focus = FocusHighlights
		}
	case FocusHighlights:
		if m.currentHighlight != nil {
			m.focus = FocusDetail
		}
	}
}

func (m CleanModel) handleEnter() (tea.Model, tea.Cmd) {
	switch m.focus {
	case FocusBooks:
		if i, ok := m.bookList.SelectedItem().(bookItem); ok {
			m.currentBook = &i.book
			m.currentHighlight = nil // Clear previous highlight
			m.loading = true
			return m, m.loadHighlights(i.book.ID)
		}

	case FocusHighlights:
		if i, ok := m.highlightList.SelectedItem().(highlightItem); ok {
			m.currentHighlight = &i.highlight
			// Don't auto-focus detail - just load it
			return m, m.renderHighlightDetail()
		}
	}

	return m, nil
}

func (m *CleanModel) updateSizes() {
	bookWidth := 30
	highlightWidth := 40
	detailWidth := m.width - bookWidth - highlightWidth - 6

	if detailWidth < 40 {
		bookWidth = 25
		highlightWidth = 35
		detailWidth = m.width - bookWidth - highlightWidth - 6
	}

	contentHeight := m.height - 3

	m.bookList.SetSize(bookWidth-6, contentHeight-2)
	m.highlightList.SetSize(highlightWidth-6, contentHeight-2)
	m.detailView.Width = detailWidth - 6
	m.detailView.Height = contentHeight - 2
}

func (m CleanModel) getHelpText() string {
	if m.editMode == ModeEdit {
		return "tab: indent • shift+tab: outdent • enter: new line • ctrl+s: save • esc: cancel"
	}

	switch m.focus {
	case FocusBooks:
		return "enter: select • /: search • tab/→: next • q: quit"
	case FocusHighlights:
		return "enter: view • /: search • ←→: navigate • tab: next • q: quit"
	case FocusDetail:
		return "e: edit note • ↑↓: scroll • ←: back • tab: next • q: quit"
	}
	return "tab/←→: navigate • q: quit"
}

// Commands (reuse existing ones)
func (m CleanModel) loadBooks() tea.Cmd {
	return func() tea.Msg {
		books, err := m.api.GetBooks(nil)
		if err != nil {
			return errMsg{err}
		}
		return booksLoadedMsg{books: books.Results}
	}
}

func (m CleanModel) loadHighlights(bookID int) tea.Cmd {
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

func (m CleanModel) renderHighlightDetail() tea.Cmd {
	return func() tea.Msg {
		if m.currentHighlight == nil {
			return nil
		}

		content := fmt.Sprintf("# Highlight\n\n> %s\n\n", m.currentHighlight.Text)

		if m.currentBook != nil {
			content += fmt.Sprintf("**Book:** %s by %s\n\n", m.currentBook.Title, m.currentBook.Author)
		}

		if m.currentHighlight.Note != "" {
			content += fmt.Sprintf("## Note\n\n%s\n\n", m.currentHighlight.Note)
		} else {
			content += "## Note\n\n*No note yet.*\n\n"
		}

		// Render markdown
		renderer, _ := glamour.NewTermRenderer(
			glamour.WithAutoStyle(),
			glamour.WithWordWrap(m.detailView.Width),
		)

		rendered, err := renderer.Render(content)
		if err != nil {
			rendered = content
		}

		return highlightRenderedMsg{content: rendered}
	}
}
