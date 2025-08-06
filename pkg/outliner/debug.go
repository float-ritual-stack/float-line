package outliner

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// DebugMessage represents a consciousness debug message
type DebugMessage struct {
	Timestamp time.Time
	Type      string // FLOAT_DISPATCH, CONSCIOUSNESS_CAPTURE, FLOAT_REDUCER_CREATED, etc.
	Content   string
	Level     DebugLevel
}

// DebugLevel represents the importance/type of debug message
type DebugLevel string

const (
	DebugLevelInfo    DebugLevel = "info"
	DebugLevelSuccess DebugLevel = "success"
	DebugLevelWarning DebugLevel = "warning"
	DebugLevelError   DebugLevel = "error"
)

// ConsciousnessDebugPanel manages consciousness debug messages
type ConsciousnessDebugPanel struct {
	messages    []DebugMessage
	maxMessages int
	visible     bool

	// Styles
	panelStyle   lipgloss.Style
	headerStyle  lipgloss.Style
	infoStyle    lipgloss.Style
	successStyle lipgloss.Style
	warningStyle lipgloss.Style
	errorStyle   lipgloss.Style
}

// NewConsciousnessDebugPanel creates a new debug panel
func NewConsciousnessDebugPanel() *ConsciousnessDebugPanel {
	cdp := &ConsciousnessDebugPanel{
		messages:    []DebugMessage{},
		maxMessages: 50,   // Keep last 50 messages
		visible:     true, // Start visible for debugging

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
	}

	// Add startup message
	cdp.AddMessage("SYSTEM", "🧠 Consciousness Debug Panel initialized", DebugLevelInfo)

	return cdp
}

// AddMessage adds a new debug message
func (cdp *ConsciousnessDebugPanel) AddMessage(msgType, content string, level DebugLevel) {
	message := DebugMessage{
		Timestamp: time.Now(),
		Type:      msgType,
		Content:   content,
		Level:     level,
	}

	cdp.messages = append(cdp.messages, message)

	// Keep only the last maxMessages
	if len(cdp.messages) > cdp.maxMessages {
		cdp.messages = cdp.messages[len(cdp.messages)-cdp.maxMessages:]
	}
}

// AddFloatDispatch adds a FLOAT dispatch message
func (cdp *ConsciousnessDebugPanel) AddFloatDispatch(patternType, imprint, sigil, dispatchID string) {
	content := fmt.Sprintf("%s → %s [%s] %s", patternType, imprint, sigil, dispatchID)
	cdp.AddMessage("FLOAT_DISPATCH", content, DebugLevelSuccess)
}

// AddConsciousnessCapture adds a consciousness capture message
func (cdp *ConsciousnessDebugPanel) AddConsciousnessCapture(action, collection string) {
	content := fmt.Sprintf("%s → %s", action, collection)
	cdp.AddMessage("CONSCIOUSNESS_CAPTURE", content, DebugLevelInfo)
}

// AddReducerCreated adds a reducer creation message
func (cdp *ConsciousnessDebugPanel) AddReducerCreated(name, query string) {
	content := fmt.Sprintf("%s: %s", name, query)
	cdp.AddMessage("FLOAT_REDUCER_CREATED", content, DebugLevelSuccess)
}

// AddSelectorCreated adds a selector creation message
func (cdp *ConsciousnessDebugPanel) AddSelectorCreated(name, outputFormat string) {
	content := fmt.Sprintf("%s: %s", name, outputFormat)
	cdp.AddMessage("FLOAT_SELECTOR_CREATED", content, DebugLevelSuccess)
}

// AddError adds an error message
func (cdp *ConsciousnessDebugPanel) AddError(msgType, content string) {
	cdp.AddMessage(msgType, content, DebugLevelError)
}

// Toggle toggles the visibility of the debug panel
func (cdp *ConsciousnessDebugPanel) Toggle() {
	cdp.visible = !cdp.visible
}

// IsVisible returns whether the debug panel is visible
func (cdp *ConsciousnessDebugPanel) IsVisible() bool {
	return cdp.visible
}

// SetVisible sets the visibility of the debug panel
func (cdp *ConsciousnessDebugPanel) SetVisible(visible bool) {
	cdp.visible = visible
}

// View renders the debug panel
func (cdp *ConsciousnessDebugPanel) View(width, height int) string {
	if !cdp.visible {
		return ""
	}

	var content strings.Builder

	// Header
	header := cdp.headerStyle.Render("🧠 Consciousness Debug Panel")
	content.WriteString(header + "\n\n")

	// Show recent messages (last 10 for display)
	displayCount := 10
	if len(cdp.messages) < displayCount {
		displayCount = len(cdp.messages)
	}

	if displayCount == 0 {
		content.WriteString("No consciousness activity yet...")
	} else {
		startIndex := len(cdp.messages) - displayCount
		for i := startIndex; i < len(cdp.messages); i++ {
			msg := cdp.messages[i]
			content.WriteString(cdp.renderMessage(msg) + "\n")
		}
	}

	// Panel styling
	panelContent := content.String()
	availableWidth := width - 6   // Account for padding and borders
	availableHeight := height - 4 // Account for padding and borders

	return cdp.panelStyle.
		Width(availableWidth).
		Height(availableHeight).
		Render(panelContent)
}

// renderMessage renders a single debug message
func (cdp *ConsciousnessDebugPanel) renderMessage(msg DebugMessage) string {
	timestamp := msg.Timestamp.Format("15:04:05")

	var style lipgloss.Style
	switch msg.Level {
	case DebugLevelSuccess:
		style = cdp.successStyle
	case DebugLevelWarning:
		style = cdp.warningStyle
	case DebugLevelError:
		style = cdp.errorStyle
	default:
		style = cdp.infoStyle
	}

	// Format: [15:04:05] TYPE: content
	return fmt.Sprintf("[%s] %s: %s",
		timestamp,
		style.Render(msg.Type),
		msg.Content)
}

// GetMessageCount returns the total number of messages
func (cdp *ConsciousnessDebugPanel) GetMessageCount() int {
	return len(cdp.messages)
}

// Clear clears all debug messages
func (cdp *ConsciousnessDebugPanel) Clear() {
	cdp.messages = []DebugMessage{}
}
