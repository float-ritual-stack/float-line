package outliner

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// InteractiveDebugPanel is an enhanced version of ConsciousnessDebugPanel
// that supports focus mode, object inspection, and filtering
type InteractiveDebugPanel struct {
	messages    []DebugMessage
	maxMessages int
	visible     bool
	focused     bool

	// View state
	selectedIndex int
	expandedMsg   *DebugMessage
	filterType    string
	searchQuery   string
	viewMode      DebugViewMode

	// UI components
	messageList list.Model
	detailView  viewport.Model

	// Styles
	panelStyle     lipgloss.Style
	headerStyle    lipgloss.Style
	infoStyle      lipgloss.Style
	successStyle   lipgloss.Style
	warningStyle   lipgloss.Style
	errorStyle     lipgloss.Style
	focusedStyle   lipgloss.Style
	unfocusedStyle lipgloss.Style
	keyStyle       lipgloss.Style
	valueStyle     lipgloss.Style
}

// DebugViewMode represents different view modes for the debug panel
type DebugViewMode int

const (
	ViewModeList DebugViewMode = iota
	ViewModeDetail
)

// DebugKeyMap defines keybindings for the debug panel
type DebugKeyMap struct {
	Up          key.Binding
	Down        key.Binding
	Enter       key.Binding
	Back        key.Binding
	Filter      key.Binding
	Search      key.Binding
	Copy        key.Binding
	Export      key.Binding
	ToggleFocus key.Binding
}

var DebugKeys = DebugKeyMap{
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
		key.WithHelp("enter", "inspect"),
	),
	Back: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "back"),
	),
	Filter: key.NewBinding(
		key.WithKeys("f"),
		key.WithHelp("f", "filter"),
	),
	Search: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "search"),
	),
	Copy: key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c", "copy"),
	),
	Export: key.NewBinding(
		key.WithKeys("e"),
		key.WithHelp("e", "export"),
	),
	ToggleFocus: key.NewBinding(
		key.WithKeys("ctrl+l"),
		key.WithHelp("ctrl+l", "toggle focus"),
	),
}

// debugMessageItem implements list.Item for the message list
type debugMessageItem struct {
	message DebugMessage
	styles  map[DebugLevel]lipgloss.Style
}

func (i debugMessageItem) FilterValue() string {
	return fmt.Sprintf("%s %s", i.message.Type, i.message.Content)
}

func (i debugMessageItem) Title() string {
	timestamp := i.message.Timestamp.Format("15:04:05")
	return fmt.Sprintf("[%s] %s", timestamp, i.message.Type)
}

func (i debugMessageItem) Description() string {
	return i.message.Content
}

// NewInteractiveDebugPanel creates a new interactive debug panel
func NewInteractiveDebugPanel() *InteractiveDebugPanel {
	idp := &InteractiveDebugPanel{
		messages:      []DebugMessage{},
		maxMessages:   100,   // Keep more messages for scrolling
		visible:       true,  // Start visible for debugging
		focused:       false, // Start unfocused
		selectedIndex: 0,
		viewMode:      ViewModeList,
		filterType:    "",
		searchQuery:   "",

		panelStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("8")).
			Padding(1).
			MarginTop(1),

		headerStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("15")).
			Bold(true).
			Underline(true),

		infoStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("14")), // cyan

		successStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("10")), // green

		warningStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("11")), // yellow

		errorStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("9")), // red

		focusedStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")),

		unfocusedStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")),

		keyStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("170")).
			Bold(true),

		valueStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("255")),
	}

	// Initialize list component
	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = true
	delegate.SetSpacing(1)
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(lipgloss.Color("170")).
		BorderForeground(lipgloss.Color("170"))
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(lipgloss.Color("255"))

	idp.messageList = list.New([]list.Item{}, delegate, 0, 0)
	idp.messageList.Title = "🧠 Consciousness Debug Messages"
	idp.messageList.SetShowHelp(false)
	idp.messageList.DisableQuitKeybindings()
	idp.messageList.SetFilteringEnabled(false) // We'll handle filtering ourselves

	// Initialize detail view
	idp.detailView = viewport.New(0, 0)
	idp.detailView.Style = lipgloss.NewStyle().
		PaddingRight(1).
		PaddingLeft(1)

	// Add startup message
	idp.AddMessage("SYSTEM", "🧠 Interactive Consciousness Debug Panel initialized", DebugLevelInfo)

	return idp
}

// AddMessage adds a new debug message
func (idp *InteractiveDebugPanel) AddMessage(msgType, content string, level DebugLevel) {
	message := DebugMessage{
		Timestamp: time.Now(),
		Type:      msgType,
		Content:   content,
		Level:     level,
	}

	idp.messages = append(idp.messages, message)

	// Keep only the last maxMessages
	if len(idp.messages) > idp.maxMessages {
		idp.messages = idp.messages[len(idp.messages)-idp.maxMessages:]
	}

	// Update the list items
	idp.updateListItems()
}

// AddFloatDispatch adds a FLOAT dispatch message
func (idp *InteractiveDebugPanel) AddFloatDispatch(patternType, imprint, sigil, dispatchID string) {
	content := fmt.Sprintf("%s → %s [%s] %s", patternType, imprint, sigil, dispatchID)
	idp.AddMessage("FLOAT_DISPATCH", content, DebugLevelSuccess)
}

// AddConsciousnessCapture adds a consciousness capture message
func (idp *InteractiveDebugPanel) AddConsciousnessCapture(action, collection string) {
	content := fmt.Sprintf("%s → %s", action, collection)
	idp.AddMessage("CONSCIOUSNESS_CAPTURE", content, DebugLevelInfo)
}

// AddReducerCreated adds a reducer creation message
func (idp *InteractiveDebugPanel) AddReducerCreated(name, query string) {
	content := fmt.Sprintf("%s: %s", name, query)
	idp.AddMessage("FLOAT_REDUCER_CREATED", content, DebugLevelSuccess)
}

// AddSelectorCreated adds a selector creation message
func (idp *InteractiveDebugPanel) AddSelectorCreated(name, outputFormat string) {
	content := fmt.Sprintf("%s: %s", name, outputFormat)
	idp.AddMessage("FLOAT_SELECTOR_CREATED", content, DebugLevelSuccess)
}

// AddError adds an error message
func (idp *InteractiveDebugPanel) AddError(msgType, content string) {
	idp.AddMessage(msgType, content, DebugLevelError)
}

// Toggle toggles the visibility of the debug panel
func (idp *InteractiveDebugPanel) Toggle() {
	idp.visible = !idp.visible
	if !idp.visible {
		idp.focused = false
	}
}

// IsVisible returns whether the debug panel is visible
func (idp *InteractiveDebugPanel) IsVisible() bool {
	return idp.visible
}

// SetVisible sets the visibility of the debug panel
func (idp *InteractiveDebugPanel) SetVisible(visible bool) {
	idp.visible = visible
	if !idp.visible {
		idp.focused = false
	}
}

// Focus gives focus to the debug panel
func (idp *InteractiveDebugPanel) Focus() tea.Cmd {
	idp.focused = true
	idp.updateListItems() // Refresh list with current messages
	return nil
}

// Blur removes focus from the debug panel
func (idp *InteractiveDebugPanel) Blur() tea.Cmd {
	idp.focused = false
	return nil
}

// Focused returns whether the debug panel has focus
func (idp *InteractiveDebugPanel) Focused() bool {
	return idp.focused
}

// ToggleFocus toggles the focus state of the debug panel
func (idp *InteractiveDebugPanel) ToggleFocus() tea.Cmd {
	if idp.focused {
		return idp.Blur()
	}
	return idp.Focus()
}

// Update handles messages and user input
func (idp *InteractiveDebugPanel) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	// Only process messages if visible and focused
	if !idp.visible {
		return nil
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Update component sizes
		idp.updateComponentSizes(msg.Width, msg.Height)

	case tea.KeyMsg:
		// Handle global keys regardless of focus
		switch {
		case key.Matches(msg, DebugKeys.ToggleFocus):
			return idp.ToggleFocus()
		}

		// Only process other keys if focused
		if !idp.focused {
			return nil
		}

		switch idp.viewMode {
		case ViewModeList:
			// List view key handling
			switch {
			case key.Matches(msg, DebugKeys.Enter):
				if len(idp.messageList.Items()) > 0 {
					selectedItem := idp.messageList.SelectedItem().(debugMessageItem)
					idp.expandedMsg = &selectedItem.message
					idp.viewMode = ViewModeDetail
					idp.updateDetailView()
				}
				return nil

			case key.Matches(msg, DebugKeys.Filter):
				// Toggle between filter types
				switch idp.filterType {
				case "":
					idp.filterType = "FLOAT_DISPATCH"
				case "FLOAT_DISPATCH":
					idp.filterType = "CONSCIOUSNESS_CAPTURE"
				case "CONSCIOUSNESS_CAPTURE":
					idp.filterType = "FLOAT_REDUCER_CREATED"
				case "FLOAT_REDUCER_CREATED":
					idp.filterType = "FLOAT_SELECTOR_CREATED"
				default:
					idp.filterType = ""
				}
				idp.updateListItems()
				return nil

			case key.Matches(msg, DebugKeys.Back):
				return idp.Blur()

			default:
				// Pass to list component
				var cmd tea.Cmd
				idp.messageList, cmd = idp.messageList.Update(msg)
				return cmd
			}

		case ViewModeDetail:
			// Detail view key handling
			switch {
			case key.Matches(msg, DebugKeys.Back):
				idp.viewMode = ViewModeList
				idp.expandedMsg = nil
				return nil

			case key.Matches(msg, DebugKeys.Copy):
				// TODO: Implement copy to clipboard
				return nil

			default:
				// Pass to viewport component
				var cmd tea.Cmd
				idp.detailView, cmd = idp.detailView.Update(msg)
				return cmd
			}
		}
	}

	return tea.Batch(cmds...)
}

// View renders the debug panel
func (idp *InteractiveDebugPanel) View(width, height int) string {
	if !idp.visible {
		return ""
	}

	// Update component sizes
	idp.updateComponentSizes(width, height)

	var content string
	var style lipgloss.Style

	if idp.focused {
		style = idp.focusedStyle
	} else {
		style = idp.unfocusedStyle
	}

	// Render appropriate view based on mode
	switch idp.viewMode {
	case ViewModeList:
		content = idp.messageList.View()
	case ViewModeDetail:
		content = idp.detailView.View()
	}

	// Add help text based on view mode and focus state
	var helpText string
	if idp.focused {
		switch idp.viewMode {
		case ViewModeList:
			helpText = "↑/↓: navigate • enter: inspect • f: filter • esc: exit focus"
		case ViewModeDetail:
			helpText = "↑/↓: scroll • c: copy • esc: back"
		}
	} else {
		helpText = "ctrl+l: focus debug panel"
	}

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Align(lipgloss.Center)

	// Calculate available space
	availableWidth := width - 4   // Account for padding and borders
	availableHeight := height - 4 // Account for padding, borders, and help text

	// Render panel with content and help text
	return style.
		Width(availableWidth).
		Height(availableHeight).
		Render(lipgloss.JoinVertical(
			lipgloss.Left,
			content,
			helpStyle.Render(helpText),
		))
}

// updateListItems refreshes the list items based on current messages and filters
func (idp *InteractiveDebugPanel) updateListItems() {
	var filteredMessages []DebugMessage

	// Apply type filter if set
	if idp.filterType != "" {
		for _, msg := range idp.messages {
			if msg.Type == idp.filterType {
				filteredMessages = append(filteredMessages, msg)
			}
		}
	} else {
		filteredMessages = idp.messages
	}

	// Apply search query if set
	if idp.searchQuery != "" {
		var searchResults []DebugMessage
		query := strings.ToLower(idp.searchQuery)
		for _, msg := range filteredMessages {
			if strings.Contains(strings.ToLower(msg.Content), query) ||
				strings.Contains(strings.ToLower(msg.Type), query) {
				searchResults = append(searchResults, msg)
			}
		}
		filteredMessages = searchResults
	}

	// Create list items
	items := make([]list.Item, len(filteredMessages))
	styles := map[DebugLevel]lipgloss.Style{
		DebugLevelInfo:    idp.infoStyle,
		DebugLevelSuccess: idp.successStyle,
		DebugLevelWarning: idp.warningStyle,
		DebugLevelError:   idp.errorStyle,
	}

	for i, msg := range filteredMessages {
		items[i] = debugMessageItem{
			message: msg,
			styles:  styles,
		}
	}

	// Update list
	idp.messageList.SetItems(items)
}

// updateDetailView refreshes the detail view with the expanded message
func (idp *InteractiveDebugPanel) updateDetailView() {
	if idp.expandedMsg == nil {
		return
	}

	// Format message as JSON for detailed inspection
	var detailContent strings.Builder

	// Header
	detailContent.WriteString(idp.headerStyle.Render(fmt.Sprintf("Message Details: %s\n\n", idp.expandedMsg.Type)))

	// Basic info
	detailContent.WriteString(fmt.Sprintf("%s %s\n",
		idp.keyStyle.Render("Timestamp:"),
		idp.valueStyle.Render(idp.expandedMsg.Timestamp.Format("2006-01-02 15:04:05.000"))))

	detailContent.WriteString(fmt.Sprintf("%s %s\n",
		idp.keyStyle.Render("Type:"),
		idp.valueStyle.Render(idp.expandedMsg.Type)))

	detailContent.WriteString(fmt.Sprintf("%s %s\n",
		idp.keyStyle.Render("Level:"),
		idp.valueStyle.Render(string(idp.expandedMsg.Level))))

	detailContent.WriteString(fmt.Sprintf("%s %s\n\n",
		idp.keyStyle.Render("Content:"),
		idp.valueStyle.Render(idp.expandedMsg.Content)))

	// For FLOAT_DISPATCH messages, parse and display structured data
	if idp.expandedMsg.Type == "FLOAT_DISPATCH" {
		// Parse the content to extract pattern type, imprint, sigil, and dispatch ID
		parts := strings.Split(idp.expandedMsg.Content, " → ")
		if len(parts) == 2 {
			patternType := parts[0]
			rest := parts[1]

			// Extract imprint and sigil
			imprintSigilParts := strings.Split(rest, " [")
			imprint := imprintSigilParts[0]

			// Extract sigil and dispatch ID if available
			var sigil, dispatchID string
			if len(imprintSigilParts) > 1 {
				sigilPart := imprintSigilParts[1]
				sigilParts := strings.Split(sigilPart, "] ")
				if len(sigilParts) > 1 {
					sigil = sigilParts[0]
					dispatchID = sigilParts[1]
				}
			}

			// Display structured data
			detailContent.WriteString(idp.headerStyle.Render("Structured Data:\n\n"))
			detailContent.WriteString(fmt.Sprintf("%s %s\n",
				idp.keyStyle.Render("Pattern Type:"),
				idp.valueStyle.Render(patternType)))
			detailContent.WriteString(fmt.Sprintf("%s %s\n",
				idp.keyStyle.Render("Imprint:"),
				idp.valueStyle.Render(imprint)))
			detailContent.WriteString(fmt.Sprintf("%s %s\n",
				idp.keyStyle.Render("Sigil:"),
				idp.valueStyle.Render(sigil)))
			detailContent.WriteString(fmt.Sprintf("%s %s\n\n",
				idp.keyStyle.Render("Dispatch ID:"),
				idp.valueStyle.Render(dispatchID)))
		}
	}

	// Add JSON representation for advanced inspection
	jsonData, err := json.MarshalIndent(idp.expandedMsg, "", "  ")
	if err == nil {
		detailContent.WriteString(idp.headerStyle.Render("JSON Representation:\n\n"))
		detailContent.WriteString(idp.valueStyle.Render(string(jsonData)))
	}

	// Set content to viewport
	idp.detailView.SetContent(detailContent.String())
	idp.detailView.GotoTop()
}

// updateComponentSizes updates the sizes of UI components
func (idp *InteractiveDebugPanel) updateComponentSizes(width, height int) {
	availableWidth := width - 6   // Account for padding and borders
	availableHeight := height - 6 // Account for padding, borders, and help text

	// Update list size
	idp.messageList.SetSize(availableWidth, availableHeight)

	// Update viewport size
	idp.detailView.Width = availableWidth
	idp.detailView.Height = availableHeight
}

// GetMessageCount returns the total number of messages
func (idp *InteractiveDebugPanel) GetMessageCount() int {
	return len(idp.messages)
}

// Clear clears all debug messages
func (idp *InteractiveDebugPanel) Clear() {
	idp.messages = []DebugMessage{}
	idp.updateListItems()
}

// SetFilter sets the message type filter
func (idp *InteractiveDebugPanel) SetFilter(filterType string) {
	idp.filterType = filterType
	idp.updateListItems()
}

// SetSearchQuery sets the search query
func (idp *InteractiveDebugPanel) SetSearchQuery(query string) {
	idp.searchQuery = query
	idp.updateListItems()
}

// ExportMessages exports all messages to a string (for saving to file)
func (idp *InteractiveDebugPanel) ExportMessages() string {
	jsonData, err := json.MarshalIndent(idp.messages, "", "  ")
	if err != nil {
		return fmt.Sprintf("Error exporting messages: %v", err)
	}
	return string(jsonData)
}
