package tui

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/evanschultz/float-rw-client/pkg/api"
	"github.com/evanschultz/float-rw-client/pkg/models"
)

type focusedPane int

const (
	focusBooks focusedPane = iota
	focusHighlights
	focusDetail // Simplified: just one detail focus instead of two
)

type editMode int

const (
	editNone editMode = iota
	editNote
	editBoth
)

const (
	minBookPaneWidth = 25
	maxBookPaneWidth = 35
	minPaneHeight    = 10
)

type ModelSplit struct {
	api    *api.Client
	width  int
	height int
	ready  bool

	// Fixed layout dimensions
	bookPaneWidth      int
	highlightPaneWidth int
	detailPaneWidth    int
	contentHeight      int

	// Components
	bookList        list.Model
	highlightList   list.Model
	highlightView   viewport.Model
	noteView        viewport.Model
	highlightEditor textarea.Model
	noteEditor      textarea.Model
	help            help.Model

	// Data
	books             []models.Book
	highlights        []models.Highlight
	currentBook       *models.Book
	currentHighlight  *models.Highlight
	originalHighlight *models.Highlight
	nextPageURL       string

	// UI state
	focusedPane     focusedPane
	editMode        editMode
	activeEditor    int // 0 = highlight, 1 = note
	loading         bool
	saving          bool
	err             error
	booksPaneHidden bool
	splitRatio      float64
}

func NewSplitModel(apiClient *api.Client) ModelSplit {
	m := ModelSplit{
		api:         apiClient,
		focusedPane: focusBooks,
		help:        help.New(),
		splitRatio:  0.5,
		editMode:    editNone,
	}

	// Initialize lists with custom delegates
	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = true
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(lipgloss.Color("170")).
		BorderForeground(lipgloss.Color("170"))
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(lipgloss.Color("241"))

	m.bookList = list.New([]list.Item{}, delegate, 0, 0)
	m.bookList.Title = "📚 Books"
	m.bookList.SetShowHelp(false)
	m.bookList.SetFilteringEnabled(true)
	m.bookList.DisableQuitKeybindings()

	// Custom delegate for highlights with more preview
	highlightDelegate := list.NewDefaultDelegate()
	highlightDelegate.ShowDescription = true
	highlightDelegate.SetHeight(5)
	highlightDelegate.Styles.SelectedTitle = highlightDelegate.Styles.SelectedTitle.
		Foreground(lipgloss.Color("170")).
		BorderForeground(lipgloss.Color("170"))

	m.highlightList = list.New([]list.Item{}, highlightDelegate, 0, 0)
	m.highlightList.Title = "📝 Highlights"
	m.highlightList.SetShowHelp(false)
	m.highlightList.SetFilteringEnabled(true)
	m.highlightList.DisableQuitKeybindings()

	// Initialize viewports with scrollbars
	m.highlightView = viewport.New(0, 0)
	m.highlightView.Style = lipgloss.NewStyle().PaddingRight(1)

	m.noteView = viewport.New(0, 0)
	m.noteView.Style = lipgloss.NewStyle().PaddingRight(1)

	// Initialize text areas for editing
	m.highlightEditor = textarea.New()
	m.highlightEditor.Placeholder = "Edit highlight text..."
	m.highlightEditor.CharLimit = 5000
	m.highlightEditor.SetHeight(10)
	m.highlightEditor.FocusedStyle.CursorLine = lipgloss.NewStyle()
	m.highlightEditor.ShowLineNumbers = false

	m.noteEditor = textarea.New()
	m.noteEditor.Placeholder = "Add your note..."
	m.noteEditor.CharLimit = 10000
	m.noteEditor.SetHeight(10)
	m.noteEditor.FocusedStyle.CursorLine = lipgloss.NewStyle()
	m.noteEditor.ShowLineNumbers = false

	return m
}

func (m ModelSplit) Init() tea.Cmd {
	m.loading = true
	return tea.Batch(
		m.loadBooks(),
		tea.EnterAltScreen,
	)
}

func (m ModelSplit) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		m.calculateLayout()
		m.updateComponentSizes()
		if m.currentHighlight != nil {
			cmds = append(cmds, m.renderHighlightDetail())
		}
		return m, tea.Batch(cmds...)

	case tea.KeyMsg:
		// When in edit mode, handle editor keys first
		if m.editMode != editNone {
			switch msg.String() {
			case "ctrl+s":
				cmds = append(cmds, m.saveEdits())
				return m, tea.Batch(cmds...)
			case "ctrl+q":
				m.cancelEdit()
				return m, m.renderHighlightDetail()
			case "ctrl+w":
				if m.editMode == editBoth {
					m.activeEditor = 1 - m.activeEditor
					if m.activeEditor == 0 {
						m.highlightEditor.Focus()
						m.noteEditor.Blur()
					} else {
						m.noteEditor.Focus()
						m.highlightEditor.Blur()
					}
				}
				return m, nil
			default:
				// Pass through to the active editor
				if m.editMode == editBoth {
					if m.activeEditor == 0 {
						newEditor, cmd := m.highlightEditor.Update(msg)
						m.highlightEditor = newEditor
						return m, cmd
					} else {
						newEditor, cmd := m.noteEditor.Update(msg)
						m.noteEditor = newEditor
						return m, cmd
					}
				} else if m.editMode == editNote {
					newEditor, cmd := m.noteEditor.Update(msg)
					m.noteEditor = newEditor
					return m, cmd
				}
			}
			return m, nil
		}

		// Normal mode key handling
		switch msg.String() {
		case "ctrl+c", "ctrl+d":
			return m, tea.Quit

		case "ctrl+b":
			m.booksPaneHidden = !m.booksPaneHidden
			m.calculateLayout()
			m.updateComponentSizes()
			if m.currentHighlight != nil {
				cmds = append(cmds, m.renderHighlightDetail())
			}
			return m, tea.Batch(cmds...)

		case "tab":
			m.cycleFocus()
			// Debug: uncomment to see focus changes
			// fmt.Printf("Focus changed to: %d\n", m.focusedPane)
			return m, nil

		case "left", "h":
			m.navigateLeft()
			return m, nil

		case "right", "l":
			m.navigateRight()
			return m, nil

		case "q":
			if m.focusedPane == focusBooks || m.focusedPane == focusHighlights {
				return m, tea.Quit
			}
			return m, nil
		}

		// Pane-specific key handling
		switch m.focusedPane {
		case focusBooks:
			if !m.booksPaneHidden {
				switch msg.String() {
				case "enter":
					if i, ok := m.bookList.SelectedItem().(bookItem); ok {
						m.currentBook = &i.book
						m.currentHighlight = nil
						m.focusedPane = focusHighlights
						m.loading = true
						return m, m.loadHighlights(i.book.ID)
					}
				case "r":
					m.loading = true
					return m, m.loadBooks()
				default:
					newList, cmd := m.bookList.Update(msg)
					m.bookList = newList
					return m, cmd
				}
			}

		case focusHighlights:
			switch msg.String() {
			case "enter":
				if i, ok := m.highlightList.SelectedItem().(highlightItem); ok {
					m.currentHighlight = &i.highlight
					copy := *m.currentHighlight
					m.originalHighlight = &copy
					// IMPORTANT: Auto-focus the detail pane when highlight is selected
					m.focusedPane = focusDetail
					// Recalculate layout to ensure detail panel is visible
					m.calculateLayout()
					m.updateComponentSizes()
					return m, m.renderHighlightDetail()
				}
			case "esc":
				if !m.booksPaneHidden {
					m.focusedPane = focusBooks
				}
				return m, nil
			default:
				newList, cmd := m.highlightList.Update(msg)
				m.highlightList = newList
				return m, cmd
			}

		case focusDetail:
			switch msg.String() {
			case "e":
				m.startEdit(editBoth)
				return m, nil
			case "E":
				m.startEdit(editNote)
				return m, nil
			case "ctrl+e":
				return m, m.openExternalEditor()
			case "esc":
				// Go back to highlights pane
				m.focusedPane = focusHighlights
				return m, nil
			default:
				// Update both viewports - they handle scrolling
				newHighlightView, cmd1 := m.highlightView.Update(msg)
				m.highlightView = newHighlightView

				newNoteView, cmd2 := m.noteView.Update(msg)
				m.noteView = newNoteView

				return m, tea.Batch(cmd1, cmd2)
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

	case highlightRenderedMsg:
		m.highlightView.SetContent(msg.content)
		m.noteView.SetContent(msg.noteContent)

	case highlightSavedMsg:
		m.saving = false
		m.editMode = editNone
		items := m.highlightList.Items()
		for i, item := range items {
			if h, ok := item.(highlightItem); ok && h.highlight.ID == m.currentHighlight.ID {
				h.highlight = *m.currentHighlight
				items[i] = h
			}
		}
		m.highlightList.SetItems(items)
		cmds = append(cmds, m.renderHighlightDetail())

	case errMsg:
		m.err = msg.err
		m.loading = false
		m.saving = false

	case externalEditorFinishedMsg:
		if msg.err == nil {
			m.currentHighlight.Note = msg.content
			m.saving = true
			cmds = append(cmds, m.updateHighlightNote())
		}

	default:
		// ALWAYS update components - let them handle their own state
		// This fixes the race condition where components miss updates

		// Update book list
		if !m.booksPaneHidden {
			newList, cmd := m.bookList.Update(msg)
			m.bookList = newList
			cmds = append(cmds, cmd)
		}

		// Update highlight list
		newList, cmd := m.highlightList.Update(msg)
		m.highlightList = newList
		cmds = append(cmds, cmd)

		// Update viewports
		newHighlightView, cmd := m.highlightView.Update(msg)
		m.highlightView = newHighlightView
		cmds = append(cmds, cmd)

		newNoteView, cmd := m.noteView.Update(msg)
		m.noteView = newNoteView
		cmds = append(cmds, cmd)

		// Update editors (they handle their own focus state)
		newHighlightEditor, cmd := m.highlightEditor.Update(msg)
		m.highlightEditor = newHighlightEditor
		cmds = append(cmds, cmd)

		newNoteEditor, cmd := m.noteEditor.Update(msg)
		m.noteEditor = newNoteEditor
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m ModelSplit) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\n\nPress ctrl+c to quit.", m.err)
	}

	if !m.ready || m.width == 0 || m.height == 0 {
		return "Initializing..."
	}

	// Create styles
	focusedStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(0, 1)

	unfocusedStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 1)

	hiddenStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderLeft(false).
		BorderTop(false).
		BorderBottom(false).
		PaddingRight(1)

	// Build panes
	var panes []string

	// Books pane
	if m.booksPaneHidden {
		indicator := strings.Repeat("│\n", m.contentHeight-2)
		bookPane := hiddenStyle.
			Height(m.contentHeight).
			Render(indicator)
		panes = append(panes, bookPane)
	} else {
		bookContent := m.bookList.View()
		if m.loading && m.focusedPane == focusBooks {
			bookContent = "Loading books..."
		}

		bookPane := bookContent
		if m.focusedPane == focusBooks && m.editMode == editNone {
			bookPane = focusedStyle.
				Width(m.bookPaneWidth - 4).
				Height(m.contentHeight - 2).
				Render(bookContent)
		} else {
			bookPane = unfocusedStyle.
				Width(m.bookPaneWidth - 4).
				Height(m.contentHeight - 2).
				Render(bookContent)
		}
		panes = append(panes, bookPane)
	}

	// Highlights pane
	if m.currentBook != nil {
		highlightContent := m.highlightList.View()
		if m.loading && m.focusedPane == focusHighlights {
			highlightContent = fmt.Sprintf("Loading highlights for %s...", m.currentBook.Title)
		}

		highlightPane := highlightContent
		if m.focusedPane == focusHighlights && m.editMode == editNone {
			highlightPane = focusedStyle.
				Width(m.highlightPaneWidth - 4).
				Height(m.contentHeight - 2).
				Render(highlightContent)
		} else {
			highlightPane = unfocusedStyle.
				Width(m.highlightPaneWidth - 4).
				Height(m.contentHeight - 2).
				Render(highlightContent)
		}
		panes = append(panes, highlightPane)
	}

	// Detail pane - show whenever we have a highlight
	if m.currentHighlight != nil {
		var detailContent string

		if m.saving {
			detailContent = "Saving..."
		} else if m.editMode != editNone {
			detailContent = m.renderEditView()
		} else {
			detailContent = m.renderSplitView()
		}

		detailPane := detailContent
		if m.focusedPane == focusDetail || m.editMode != editNone {
			detailPane = focusedStyle.
				Width(m.detailPaneWidth - 4).
				Height(m.contentHeight - 2).
				Render(detailContent)
		} else {
			detailPane = unfocusedStyle.
				Width(m.detailPaneWidth - 4).
				Height(m.contentHeight - 2).
				Render(detailContent)
		}
		panes = append(panes, detailPane)
	}

	// Join panes horizontally
	content := lipgloss.JoinHorizontal(lipgloss.Top, panes...)

	// Add help text
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

func (m ModelSplit) renderSplitView() string {
	innerWidth := max(1, m.detailPaneWidth-6)
	splitHeight := max(2, m.contentHeight-4)
	highlightHeight := max(1, int(float64(splitHeight)*m.splitRatio))
	noteHeight := max(1, splitHeight-highlightHeight-1)

	// Highlight section with border indicator
	highlightStyle := lipgloss.NewStyle().
		Width(innerWidth).
		Height(highlightHeight)

	if m.focusedPane == focusDetail {
		highlightStyle = highlightStyle.
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("62")).
			BorderLeft(false).
			BorderRight(false).
			BorderTop(false)
	} else {
		// Subtle unfocused border
		highlightStyle = highlightStyle.
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240")).
			BorderLeft(false).
			BorderRight(false).
			BorderTop(false)
	}
	highlightContent := m.highlightView.View()
	if m.highlightView.TotalLineCount() > m.highlightView.Height {
		denominator := m.highlightView.TotalLineCount() - m.highlightView.Height
		if denominator > 0 {
			scrollPercent := float64(m.highlightView.YOffset) / float64(denominator)
			highlightContent = m.addScrollbar(highlightContent, m.highlightView.Height, scrollPercent)
		}
	}

	highlightSection := highlightStyle.Render(highlightContent)

	// Separator
	separator := lipgloss.NewStyle().
		Width(innerWidth).
		Foreground(lipgloss.Color("240")).
		Render(strings.Repeat("─", innerWidth))

	// Note section with border indicator
	noteStyle := lipgloss.NewStyle().
		Width(innerWidth).
		Height(noteHeight)

	if m.focusedPane == focusDetail {
		noteStyle = noteStyle.
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("62")).
			BorderLeft(false).
			BorderRight(false).
			BorderBottom(false)
	} else {
		// Subtle unfocused border
		noteStyle = noteStyle.
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240")).
			BorderLeft(false).
			BorderRight(false).
			BorderBottom(false)
	}

	noteContent := m.noteView.View()
	if m.noteView.TotalLineCount() > m.noteView.Height {
		denominator := m.noteView.TotalLineCount() - m.noteView.Height
		if denominator > 0 {
			scrollPercent := float64(m.noteView.YOffset) / float64(denominator)
			noteContent = m.addScrollbar(noteContent, m.noteView.Height, scrollPercent)
		}
	}

	noteSection := noteStyle.Render(noteContent)

	return lipgloss.JoinVertical(
		lipgloss.Top,
		highlightSection,
		separator,
		noteSection,
	)
}

func (m ModelSplit) renderEditView() string {
	innerWidth := max(20, m.detailPaneWidth-6)
	innerHeight := max(10, m.contentHeight-4)

	if m.editMode == editNote {
		m.noteEditor.SetWidth(innerWidth)
		m.noteEditor.SetHeight(innerHeight)
		return m.noteEditor.View()
	} else if m.editMode == editBoth {
		editorHeight := (innerHeight - 3) / 2

		m.highlightEditor.SetWidth(innerWidth)
		m.highlightEditor.SetHeight(editorHeight)

		highlightStyle := lipgloss.NewStyle()
		if m.activeEditor == 0 {
			highlightStyle = highlightStyle.BorderForeground(lipgloss.Color("62"))
		} else {
			highlightStyle = highlightStyle.BorderForeground(lipgloss.Color("240"))
		}

		highlightSection := highlightStyle.Render(m.highlightEditor.View())

		m.noteEditor.SetWidth(innerWidth)
		m.noteEditor.SetHeight(editorHeight)

		noteStyle := lipgloss.NewStyle()
		if m.activeEditor == 1 {
			noteStyle = noteStyle.BorderForeground(lipgloss.Color("62"))
		} else {
			noteStyle = noteStyle.BorderForeground(lipgloss.Color("240"))
		}

		noteSection := noteStyle.Render(m.noteEditor.View())

		return lipgloss.JoinVertical(
			lipgloss.Top,
			lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("Highlight:"),
			highlightSection,
			lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("Note:"),
			noteSection,
		)
	}

	return ""
}

func (m ModelSplit) addScrollbar(content string, height int, scrollPercent float64) string {
	if height <= 0 {
		return content
	}
	scrollbarStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))

	scrollbarActiveStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("62"))

	lines := strings.Split(content, "\n")
	if len(lines) > height {
		lines = lines[:height]
	}

	// Calculate scrollbar position
	scrollbarHeight := height
	thumbHeight := max(1, scrollbarHeight/10)
	thumbPosition := int(float64(scrollbarHeight-thumbHeight) * scrollPercent)

	// Add scrollbar to each line
	for i := range lines {
		if i >= thumbPosition && i < thumbPosition+thumbHeight {
			lines[i] += scrollbarActiveStyle.Render(" ▐")
		} else {
			lines[i] += scrollbarStyle.Render(" │")
		}
	}

	return strings.Join(lines, "\n")
}

func (m *ModelSplit) calculateLayout() {
	if m.width == 0 || m.height == 0 {
		return
	}

	helpHeight := 2
	m.contentHeight = m.height - helpHeight

	// Debug: uncomment to see layout calculations
	// fmt.Printf("Layout: width=%d, highlight=%v, bookWidth=%d, highlightWidth=%d, detailWidth=%d\n",
	//	m.width, m.currentHighlight != nil, m.bookPaneWidth, m.highlightPaneWidth, m.detailPaneWidth) // PRIORITY: If we have a highlight, detail panel MUST be visible
	// This ensures the highlight/note view is always accessible
	if m.currentHighlight != nil {
		// Force minimum detail panel width
		minDetailWidth := 50
		if m.width < 100 {
			minDetailWidth = m.width / 3 // At least 1/3 of screen
		}

		// Calculate remaining space for books + highlights
		availableWidth := m.width - minDetailWidth

		if m.booksPaneHidden {
			m.bookPaneWidth = 3
			m.highlightPaneWidth = availableWidth - m.bookPaneWidth
			m.detailPaneWidth = minDetailWidth
		} else {
			// Books get minimum space, highlights get the rest
			m.bookPaneWidth = minBookPaneWidth
			if availableWidth > 80 && m.width > 120 {
				m.bookPaneWidth = maxBookPaneWidth
			}

			m.highlightPaneWidth = availableWidth - m.bookPaneWidth
			m.detailPaneWidth = minDetailWidth

			// Ensure highlight pane isn't too small
			if m.highlightPaneWidth < 25 {
				m.bookPaneWidth = availableWidth - 25
				m.highlightPaneWidth = 25
			}
		}
	} else {
		// No highlight selected - use original logic
		if m.booksPaneHidden {
			m.bookPaneWidth = 3
			m.highlightPaneWidth = m.width - m.bookPaneWidth
			m.detailPaneWidth = 0
		} else {
			m.bookPaneWidth = minBookPaneWidth
			if m.width > 120 {
				m.bookPaneWidth = maxBookPaneWidth
			}

			m.highlightPaneWidth = m.width - m.bookPaneWidth
			m.detailPaneWidth = 0
		}
	}
}
func (m *ModelSplit) updateComponentSizes() {
	// Update list sizes
	if !m.booksPaneHidden {
		m.bookList.SetSize(m.bookPaneWidth-6, m.contentHeight-2)
	}
	m.highlightList.SetSize(m.highlightPaneWidth-6, m.contentHeight-2)

	// Update viewport sizes
	if m.detailPaneWidth > 0 {
		splitHeight := m.contentHeight - 4
		highlightHeight := int(float64(splitHeight) * m.splitRatio)
		noteHeight := splitHeight - highlightHeight - 1

		// Ensure minimum heights
		if highlightHeight < minPaneHeight {
			highlightHeight = minPaneHeight
			noteHeight = splitHeight - highlightHeight - 1
		}
		if noteHeight < minPaneHeight {
			noteHeight = minPaneHeight
			highlightHeight = splitHeight - noteHeight - 1
		}

		m.highlightView.Width = m.detailPaneWidth - 8 // Account for padding and scrollbar
		m.highlightView.Height = highlightHeight

		m.noteView.Width = m.detailPaneWidth - 8
		m.noteView.Height = noteHeight
	}
}

// getAvailablePanes returns the list of currently available panes
func (m *ModelSplit) getAvailablePanes() []focusedPane {
	panes := []focusedPane{}

	// Books pane (if not hidden)
	if !m.booksPaneHidden {
		panes = append(panes, focusBooks)
	}

	// Highlights pane (if we have a book)
	if m.currentBook != nil {
		panes = append(panes, focusHighlights)
	}

	// Detail pane (if we have a highlight)
	if m.currentHighlight != nil {
		panes = append(panes, focusDetail)
	}

	return panes
}

// findPaneIndex returns the index of the current pane in available panes
func (m *ModelSplit) findPaneIndex() int {
	panes := m.getAvailablePanes()
	for i, pane := range panes {
		if pane == m.focusedPane {
			return i
		}
	}
	return 0 // Default to first pane if not found
}

func (m *ModelSplit) navigateLeft() {
	if m.editMode != editNone {
		return
	}

	panes := m.getAvailablePanes()
	if len(panes) <= 1 {
		return
	}

	currentIndex := m.findPaneIndex()
	if currentIndex > 0 {
		m.focusedPane = panes[currentIndex-1]
	}
}

func (m *ModelSplit) navigateRight() {
	if m.editMode != editNone {
		return
	}

	panes := m.getAvailablePanes()
	if len(panes) <= 1 {
		return
	}

	currentIndex := m.findPaneIndex()
	if currentIndex < len(panes)-1 {
		m.focusedPane = panes[currentIndex+1]
	}
}

func (m *ModelSplit) cycleFocus() {
	if m.editMode != editNone {
		return
	}

	panes := m.getAvailablePanes()
	if len(panes) <= 1 {
		return
	}

	oldFocus := m.focusedPane
	currentIndex := m.findPaneIndex()
	m.focusedPane = panes[(currentIndex+1)%len(panes)]

	// Debug: uncomment to see focus transitions
	// fmt.Printf("Focus: %d -> %d (available: %v)\n", oldFocus, m.focusedPane, panes)
	_ = oldFocus // Prevent unused variable warning
}

func (m *ModelSplit) startEdit(mode editMode) {
	m.editMode = mode

	// Blur all viewports when entering edit mode
	// (viewports don't have Focus/Blur methods, but this is conceptually what we want)

	if mode == editBoth {
		m.highlightEditor.SetValue(m.currentHighlight.Text)
		m.noteEditor.SetValue(m.currentHighlight.Note)
		m.activeEditor = 1 // Start with note editor
		m.noteEditor.Focus()
		m.highlightEditor.Blur()
	} else if mode == editNote {
		m.noteEditor.SetValue(m.currentHighlight.Note)
		m.noteEditor.Focus()
		// Make sure highlight editor is blurred
		m.highlightEditor.Blur()
	}
}
func (m *ModelSplit) cancelEdit() {
	m.editMode = editNone

	// Blur editors when exiting edit mode
	m.highlightEditor.Blur()
	m.noteEditor.Blur()

	// Restore original content
	if m.originalHighlight != nil {
		m.currentHighlight.Text = m.originalHighlight.Text
		m.currentHighlight.Note = m.originalHighlight.Note
	}
}

func (m ModelSplit) getHelpText() string {
	var parts []string

	if m.booksPaneHidden {
		parts = append(parts, "ctrl+b: show books")
	} else {
		parts = append(parts, "ctrl+b: hide books")
	}

	if m.editMode != editNone {
		parts = append(parts, "ctrl+s: save • ctrl+q: cancel")
		if m.editMode == editBoth {
			parts = append(parts, "ctrl+w: switch editor")
		}
	} else {
		switch m.focusedPane {
		case focusBooks:
			parts = append(parts, "enter: select • /: search • r: refresh")
		case focusHighlights:
			parts = append(parts, "enter: view • /: search • esc: back")
			if m.currentBook != nil {
				status := fmt.Sprintf("%d highlights", len(m.highlights))
				if m.nextPageURL != "" {
					status += " (more available)"
				}
				parts = append([]string{status}, parts...)
			}
		case focusDetail:
			parts = append(parts, "e: edit both • E: edit note • ctrl+e: external • ↑↓: scroll • esc: back")
		}

		parts = append(parts, "tab/←→: navigate • ctrl+c: quit")
	}

	return strings.Join(parts, " • ")
}

func (m ModelSplit) saveEdits() tea.Cmd {
	return func() tea.Msg {
		if m.currentHighlight == nil {
			return nil
		}

		if m.editMode == editNote || m.editMode == editBoth {
			m.currentHighlight.Note = m.noteEditor.Value()
		}
		if m.editMode == editBoth {
			m.currentHighlight.Text = m.highlightEditor.Value()
		}

		m.saving = true
		return highlightSavedMsg{}
	}
}

func (m ModelSplit) openExternalEditor() tea.Cmd {
	return func() tea.Msg {
		tea.ClearScreen()

		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = "vim"
		}

		tmpfile, err := os.CreateTemp("", "readwise-note-*.md")
		if err != nil {
			return errMsg{err}
		}

		content := fmt.Sprintf("# Note for Highlight\n\n> %s\n\n---\n\n%s",
			m.currentHighlight.Text, m.currentHighlight.Note)

		if _, err := tmpfile.Write([]byte(content)); err != nil {
			tmpfile.Close()
			os.Remove(tmpfile.Name())
			return errMsg{err}
		}
		tmpfile.Close()

		cmd := exec.Command(editor, tmpfile.Name())
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			os.Remove(tmpfile.Name())
			return errMsg{err}
		}

		edited, err := os.ReadFile(tmpfile.Name())
		os.Remove(tmpfile.Name())
		if err != nil {
			return errMsg{err}
		}

		parts := strings.Split(string(edited), "---\n\n")
		if len(parts) > 1 {
			return externalEditorFinishedMsg{content: strings.TrimSpace(parts[1])}
		}

		return externalEditorFinishedMsg{content: string(edited)}
	}
}

// Commands
func (m ModelSplit) loadBooks() tea.Cmd {
	return func() tea.Msg {
		books, err := m.api.GetBooks(nil)
		if err != nil {
			return errMsg{err}
		}
		return booksLoadedMsg{books: books.Results}
	}
}

func (m ModelSplit) loadHighlights(bookID int) tea.Cmd {
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

func (m ModelSplit) renderHighlightDetail() tea.Cmd {
	return func() tea.Msg {
		if m.currentHighlight == nil {
			return nil
		}

		highlightContent := fmt.Sprintf("# Highlight\n\n> %s\n\n", m.currentHighlight.Text)

		if m.currentBook != nil {
			highlightContent += fmt.Sprintf("**Book:** %s by %s\n\n",
				m.currentBook.Title, m.currentBook.Author)
		}

		if m.currentHighlight.URL != "" {
			highlightContent += fmt.Sprintf("**Source:** [Link](%s)\n\n", m.currentHighlight.URL)
		}

		noteContent := "## Note\n\n"
		if m.currentHighlight.Note != "" {
			noteContent += m.currentHighlight.Note
		} else {
			noteContent += "*No note yet. Press 'e' to add one.*"
		}

		detailWidth := max(50, m.detailPaneWidth-10)
		if detailWidth < 40 {
			detailWidth = 40
		}

		renderer, _ := glamour.NewTermRenderer(
			glamour.WithAutoStyle(),
			glamour.WithWordWrap(detailWidth),
		)

		renderedHighlight, err := renderer.Render(highlightContent)
		if err != nil {
			renderedHighlight = highlightContent
		}

		renderedNote, err := renderer.Render(noteContent)
		if err != nil {
			renderedNote = noteContent
		}

		return highlightRenderedMsg{
			content:     renderedHighlight,
			noteContent: renderedNote,
		}
	}
}

func (m ModelSplit) updateHighlightNote() tea.Cmd {
	return func() tea.Msg {
		if m.currentHighlight == nil {
			return nil
		}

		update := models.HighlightUpdate{
			Note: m.currentHighlight.Note,
		}

		if m.editMode == editBoth {
			update.Text = m.currentHighlight.Text
		}

		_, err := m.api.UpdateHighlight(m.currentHighlight.ID, update)
		if err != nil {
			return errMsg{err}
		}

		for i, h := range m.highlights {
			if h.ID == m.currentHighlight.ID {
				m.highlights[i] = *m.currentHighlight
				break
			}
		}

		return highlightSavedMsg{}
	}
}

// Helper functions
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Additional message types
type externalEditorFinishedMsg struct {
	content string
	err     error
}
