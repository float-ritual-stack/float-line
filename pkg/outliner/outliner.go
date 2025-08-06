package outliner

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ReducerUpdateMsg represents a reducer collecting a new action
type ReducerUpdateMsg struct {
	ReducerName string
	Action      DispatchAction
}

// OutlineNode represents a single line in the outline with consciousness metadata
type OutlineNode struct {
	ID          string // Unique identifier for this node
	Text        string // Display text
	Level       int    // 0 = root level, 1 = indented once, etc.
	Collapsed   bool   // true if this node's children are hidden
	HasChildren bool   // true if this node has child nodes

	// Consciousness metadata
	CreatedAt   time.Time         // When this node was created
	ModifiedAt  time.Time         // When this node was last modified
	PatternType string            // ctx, eureka, decision, etc. (empty for plain text)
	Metadata    map[string]string // Additional key-value metadata
	Captured    bool              // Whether this node has been sent to consciousness system

	// Bidirectional linking
	Links     []string // [[concept]] links found in this node's text
	Backlinks []string // Node IDs that link to this node

	// Display state
	DetailMode bool // Whether to show full metadata in display
}

// Outliner is the main component
type Outliner struct {
	lines     []OutlineNode
	cursor    int // which line we're editing
	cursorPos int // position within the current line
	width     int
	height    int
	focused   bool

	// Consciousness integration
	parser     *Parser
	evna       *EvnaDispatcher
	detailMode bool // Global detail mode toggle

	// FLOAT.dispatch system
	dispatch   *FloatDispatchSystem
	debugPanel *InteractiveDebugPanel

	// Reducer update channel for Elm-style message passing
	reducerUpdates chan ReducerUpdateMsg

	// Bidirectional linking
	linkRegistry   map[string][]string // concept -> []nodeIDs that mention it	// Styles
	bulletStyle    lipgloss.Style
	textStyle      lipgloss.Style
	cursorStyle    lipgloss.Style
	focusedStyle   lipgloss.Style
	unfocusedStyle lipgloss.Style
	highlightStyle lipgloss.Style // For current row
	treeLineStyle  lipgloss.Style // For tree connection lines
}

// generateNodeID creates a unique ID for a node
func generateNodeID() string {
	bytes := make([]byte, 4)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// newNode creates a new OutlineNode with consciousness metadata
func newNode(text string, level int) OutlineNode {
	now := time.Now()
	return OutlineNode{
		ID:         generateNodeID(),
		Text:       text,
		Level:      level,
		CreatedAt:  now,
		ModifiedAt: now,
		Metadata:   make(map[string]string),
		Captured:   false,
		DetailMode: false,
		Links:      []string{},
		Backlinks:  []string{},
	}
}

// New creates a new outliner
func New() Outliner {
	o := Outliner{
		lines: []OutlineNode{
			newNode("", 0), // Start with one empty line
		},
		cursor:       0,
		cursorPos:    0,
		detailMode:   false,
		linkRegistry: make(map[string][]string),

		// Consciousness integration
		parser:     NewParser(),
		evna:       NewEvnaDispatcher(),
		dispatch:   NewFloatDispatchSystem(),
		debugPanel: NewInteractiveDebugPanel(),

		// Elm-style message channel
		reducerUpdates: make(chan ReducerUpdateMsg, 100),

		// Default styles
		bulletStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("62")),
		textStyle:   lipgloss.NewStyle(),
		cursorStyle: lipgloss.NewStyle().Background(lipgloss.Color("62")).Foreground(lipgloss.Color("15")),
		highlightStyle: lipgloss.NewStyle().
			Background(lipgloss.Color("236")).
			Foreground(lipgloss.Color("15")),
		treeLineStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
		focusedStyle: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(1),
		unfocusedStyle: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(1),
	}

	// Set up error logging callback for evna dispatcher
	o.evna.SetErrorLogger(func(msgType, content string) {
		o.debugPanel.AddError(msgType, content)
	})

	// Set up reducer update callback for Elm-style message passing
	o.dispatch.SetReducerUpdateCallback(func(reducerName string, action DispatchAction) {
		// Debug: Log callback firing
		o.debugPanel.AddMessage("CALLBACK_FIRED", fmt.Sprintf("Sending message for reducer '%s'", reducerName), DebugLevelInfo)

		// Send message through channel instead of direct mutation
		select {
		case o.reducerUpdates <- ReducerUpdateMsg{ReducerName: reducerName, Action: action}:
			// Message sent successfully
			o.debugPanel.AddMessage("MESSAGE_SENT", fmt.Sprintf("Message sent for reducer '%s'", reducerName), DebugLevelSuccess)
		default:
			// Channel full, skip this update (non-blocking)
			o.debugPanel.AddMessage("CHANNEL_FULL", fmt.Sprintf("Channel full, skipped update for reducer '%s'", reducerName), DebugLevelError)
		}
	})

	return o
}

// listenForReducerUpdates creates a command that listens for reducer updates
func (o *Outliner) listenForReducerUpdates() tea.Cmd {
	return func() tea.Msg {
		return <-o.reducerUpdates
	}
}

// Focus gives focus to the outliner
func (o *Outliner) Focus() tea.Cmd {
	o.focused = true
	return o.listenForReducerUpdates() // Start listening for reducer updates
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
	// Always handle window size messages
	if _, ok := msg.(tea.WindowSizeMsg); ok {
		if o.debugPanel.IsVisible() {
			cmd := o.debugPanel.Update(msg)
			return o, cmd
		}
	}

	// If debug panel is focused, send all messages to it first
	if o.debugPanel.IsVisible() && o.debugPanel.Focused() {
		cmd := o.debugPanel.Update(msg)
		return o, cmd
	}

	// Otherwise, only process messages when outliner is focused
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
				newNodeObj := newNode("", currentLevel)

				// Insert after current line
				o.lines = append(o.lines[:o.cursor+1], append([]OutlineNode{newNodeObj}, o.lines[o.cursor+1:]...)...)
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

		case "ctrl+t":
			// Toggle detail mode
			o.detailMode = !o.detailMode

		case "ctrl+l":
			// Toggle debug panel (log)
			o.debugPanel.Toggle()
			o.debugPanel.AddMessage("KEY_BINDING", "Ctrl+L pressed - toggled debug panel", DebugLevelInfo)

		case "ctrl+shift+l":
			// Toggle focus on debug panel
			o.debugPanel.AddMessage("KEY_BINDING", "Ctrl+Shift+L pressed", DebugLevelInfo)
			if o.debugPanel.IsVisible() {
				if o.debugPanel.Focused() {
					o.debugPanel.Blur()
					o.debugPanel.AddMessage("FOCUS", "Debug panel blurred", DebugLevelInfo)
				} else {
					o.debugPanel.Focus()
					o.debugPanel.AddMessage("FOCUS", "Debug panel focused", DebugLevelSuccess)
				}
			} else {
				o.debugPanel.AddMessage("FOCUS", "Debug panel not visible, cannot focus", DebugLevelError)
			}
		default:
			// Handle regular character input
			if len(msg.String()) == 1 {
				char := msg.String()
				if o.cursor < len(o.lines) {
					line := &o.lines[o.cursor]
					// Insert character at cursor position
					line.Text = line.Text[:o.cursorPos] + char + line.Text[o.cursorPos:]
					line.ModifiedAt = time.Now()
					line.Captured = false // Mark as needing re-capture
					o.cursorPos++

					// Update links when text changes
					o.updateNodeLinks(o.cursor)
				}
			}
		}

	case ReducerUpdateMsg:
		// Debug: Log message received
		o.debugPanel.AddMessage("MESSAGE_RECEIVED", fmt.Sprintf("Processing reducer update for '%s'", msg.ReducerName), DebugLevelInfo)

		// Handle reducer update message (Elm-style)
		o.handleReducerUpdateMessage(msg)
		return o, o.listenForReducerUpdates() // Continue listening
	}

	return o, nil
}

// View renders the outliner with enhanced visual feedback
func (o Outliner) View() string {
	if len(o.lines) == 0 {
		return ""
	}

	var content strings.Builder

	// Debug info (can be removed later)
	content.WriteString(fmt.Sprintf("Lines: %d, Cursor: %d\n", len(o.lines), o.cursor))

	for i, line := range o.lines {
		isCurrentLine := i == o.cursor && o.focused

		// Build tree structure with connection lines
		var treePrefix strings.Builder

		// Add tree connection lines for nested items
		for level := 0; level < line.Level; level++ {
			if level == line.Level-1 {
				// Last level - show branch
				treePrefix.WriteString(o.treeLineStyle.Render("├─ "))
			} else {
				// Intermediate levels - show vertical line
				treePrefix.WriteString(o.treeLineStyle.Render("│  "))
			}
		}

		// Choose bullet based on level and expand state
		var bullet string
		if line.HasChildren {
			// Show expand/collapse indicator for nodes with children
			if line.Collapsed {
				bullet = "▶" // Collapsed (children hidden)
			} else {
				bullet = "▼" // Expanded (children visible)
			}
		} else {
			// Regular bullets for leaf nodes
			switch line.Level {
			case 0:
				bullet = "●" // Solid bullet for root items
			case 1:
				bullet = "○" // Hollow bullet for level 1
			case 2:
				bullet = "◦" // Small bullet for level 2
			default:
				bullet = "·" // Tiny bullet for deeper levels
			}
		}

		// Style the bullet
		styledBullet := o.bulletStyle.Render(bullet + " ")

		// Build the text content with consciousness metadata
		textContent := o.renderNodeContent(line)

		// Add cursor if this is the current line
		if isCurrentLine {
			cursorPos := o.cursorPos
			if cursorPos > len(line.Text) {
				cursorPos = len(line.Text)
			}

			// Insert cursor character
			beforeCursor := line.Text[:cursorPos]
			afterCursor := line.Text[cursorPos:]
			textContent = beforeCursor + o.cursorStyle.Render("│") + afterCursor
		}

		// Combine all parts
		lineContent := treePrefix.String() + styledBullet + textContent

		// Apply row highlighting for current line
		if isCurrentLine {
			// Pad to full width and highlight entire row
			padding := o.width - lipgloss.Width(lineContent) - 6 // Account for border
			if padding > 0 {
				lineContent += strings.Repeat(" ", padding)
			}
			lineContent = o.highlightStyle.Render(lineContent)
		}

		content.WriteString(lineContent)
		if i < len(o.lines)-1 {
			content.WriteString("\n")
		}
	}

	// Calculate heights based on debug panel visibility
	var mainHeight int
	var mainContent string

	if o.debugPanel.IsVisible() {
		debugPanelHeight := o.height / 3
		mainHeight = o.height - debugPanelHeight - 4

		// Style the main content based on focus state
		if o.focused && !o.debugPanel.Focused() {
			mainContent = o.focusedStyle.Width(o.width - 4).Height(mainHeight).Render(content.String())
		} else {
			mainContent = o.unfocusedStyle.Width(o.width - 4).Height(mainHeight).Render(content.String())
		}

		// Render debug panel with appropriate focus
		debugContent := o.debugPanel.View(o.width, debugPanelHeight)
		return mainContent + "\n" + debugContent
	} else {
		// Full height when debug panel is hidden
		if o.focused {
			mainContent = o.focusedStyle.Width(o.width - 4).Height(o.height - 4).Render(content.String())
		} else {
			mainContent = o.unfocusedStyle.Width(o.width - 4).Height(o.height - 4).Render(content.String())
		}
		return mainContent
	}
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

// handleReducerUpdateMessage handles reducer update messages (Elm-style)
func (o *Outliner) handleReducerUpdateMessage(msg ReducerUpdateMsg) {
	// Debug: Log that message was received
	o.debugPanel.AddMessage("REDUCER_UPDATE", fmt.Sprintf("Reducer '%s' collected: %s", msg.ReducerName, msg.Action.Content), DebugLevelSuccess)

	// Find the reducer node in the outline
	for i, line := range o.lines {
		if line.PatternType == "reducer" && strings.Contains(line.Text, msg.ReducerName) {
			// Mark reducer as having children
			o.lines[i].HasChildren = true
			o.lines[i].Collapsed = false // Start expanded so user can see collection happening

			// Create child node for the collected action
			childNode := OutlineNode{
				ID:          generateNodeID(),
				Text:        fmt.Sprintf("%s: %s", msg.Action.PatternType, msg.Action.Content),
				Level:       line.Level + 1,
				Collapsed:   false,
				HasChildren: false,
				CreatedAt:   time.Now(),
				ModifiedAt:  time.Now(),
				PatternType: msg.Action.PatternType,
				Metadata:    msg.Action.Metadata,
				Captured:    true, // Already captured by reducer
			}

			// Insert child node after the reducer (expand downward for now)
			// TODO: Implement upward expansion to avoid cursor displacement
			insertIndex := i + 1

			// Skip existing children to insert at the end
			for insertIndex < len(o.lines) && o.lines[insertIndex].Level > line.Level {
				insertIndex++
			}

			// Insert the new child
			o.lines = append(o.lines[:insertIndex], append([]OutlineNode{childNode}, o.lines[insertIndex:]...)...)

			break
		}
	}
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

		node := newNode(trimmed, level)
		// Detect if this is a consciousness pattern and mark it
		if patternType := o.detectPatternType(trimmed); patternType != "" {
			node.PatternType = patternType
		}
		o.lines = append(o.lines, node)
	}

	// Ensure we have at least one line
	if len(o.lines) == 0 {
		o.lines = []OutlineNode{newNode("", 0)}
	}

	o.cursor = 0
	o.cursorPos = 0

	// Update all links after loading content
	for i := range o.lines {
		o.updateNodeLinks(i)
	}

	// Trigger consciousness capture on content load
	o.captureConsciousness("content_load")
}

// captureConsciousness analyzes content for :: patterns and dispatches through FLOAT system
func (o *Outliner) captureConsciousness(trigger string) {
	if o.parser == nil || o.evna == nil || o.dispatch == nil {
		return
	}

	content := o.GetContent()
	parsed := o.parser.Parse(content)

	if len(parsed.ConsciousnessData) > 0 {
		// Process through FLOAT.dispatch system
		for _, pattern := range parsed.ConsciousnessData {
			// Find the corresponding node
			nodeID := ""
			if pattern.Line <= len(o.lines) {
				nodeID = o.lines[pattern.Line-1].ID
			}

			// Handle special FLOAT patterns
			o.handleFloatPattern(pattern, nodeID)

			// Dispatch through FLOAT system
			action := o.dispatch.Dispatch(nodeID, pattern.Content, pattern.Type)

			// Also send to evna for external consciousness integration
			source := fmt.Sprintf("float-dispatch:%s", trigger)
			if err := o.evna.DispatchPatterns([]ConsciousnessPattern{pattern}, source); err != nil {
				o.debugPanel.AddError("EVNA_DISPATCH_ERROR", err.Error())
			} else {
				o.debugPanel.AddConsciousnessCapture(action.PatternType, "evna")
			}

			// Log the FLOAT dispatch
			o.debugPanel.AddFloatDispatch(action.PatternType, action.Imprint, action.Sigil, action.ID)
		}

		// Mark nodes as captured after successful dispatch
		o.markNodesAsCaptured(parsed.ConsciousnessData)
	}
}

// markNodesAsCaptured updates node capture status after successful consciousness dispatch
func (o *Outliner) markNodesAsCaptured(patterns []ConsciousnessPattern) {
	// Create a map of line numbers that were captured
	capturedLines := make(map[int]bool)
	for _, pattern := range patterns {
		capturedLines[pattern.Line] = true
	}

	// Mark corresponding nodes as captured
	lineNum := 1
	for i := range o.lines {
		node := &o.lines[i]
		if node.Text != "" { // Only count non-empty lines
			if capturedLines[lineNum] {
				node.Captured = true
			}
			lineNum++
		}
	}
}

// TriggerConsciousnessCapture manually triggers consciousness pattern analysis
func (o *Outliner) TriggerConsciousnessCapture() {
	o.captureConsciousness("manual_trigger")
}

// IsDetailMode returns whether detail mode is enabled
func (o *Outliner) IsDetailMode() bool {
	return o.detailMode
}

// IsDebugVisible returns whether debug panel is visible
func (o *Outliner) IsDebugVisible() bool {
	return o.debugPanel.IsVisible()
}

// handleFloatPattern processes special FLOAT patterns (reducer::, selector::)
func (o *Outliner) handleFloatPattern(pattern ConsciousnessPattern, nodeID string) {
	switch pattern.Type {
	case "reducer":
		o.handleReducerPattern(pattern, nodeID)
	case "selector":
		o.handleSelectorPattern(pattern, nodeID)
	}
}

// handleReducerPattern creates a new consciousness reducer
func (o *Outliner) handleReducerPattern(pattern ConsciousnessPattern, nodeID string) {
	// Parse reducer definition: "reducer::name collect all actions that are bridges about rangle"
	parts := strings.SplitN(pattern.Content, " ", 2)
	if len(parts) < 2 {
		return
	}

	reducerName := parts[0]
	query := parts[1]

	// Create matcher based on query (simplified for now)
	matcher := func(action DispatchAction) bool {
		content := strings.ToLower(action.Content)
		queryLower := strings.ToLower(query)

		// Parse query for keywords and pattern types
		// Example: "collect all actions that mention test" -> look for "test" in content
		// Example: "collect all bridges about rangle" -> look for bridges with "rangle"

		// Extract keywords after "about" or "that mention"
		var keywords []string
		if strings.Contains(queryLower, "about ") {
			parts := strings.Split(queryLower, "about ")
			if len(parts) > 1 {
				keywords = strings.Fields(parts[1])
			}
		} else if strings.Contains(queryLower, "that mention ") {
			parts := strings.Split(queryLower, "that mention ")
			if len(parts) > 1 {
				keywords = strings.Fields(parts[1])
			}
		}

		// Check if content contains any of the keywords
		for _, keyword := range keywords {
			if strings.Contains(content, keyword) {
				return true
			}
		}

		// Legacy hardcoded patterns for backward compatibility
		if strings.Contains(queryLower, "bridges") && action.PatternType == "bridge" {
			return true
		}
		if strings.Contains(queryLower, "rangle") && strings.Contains(content, "rangle") {
			return true
		}

		return false
	}

	o.dispatch.AddReducer(reducerName, query, matcher)
	o.debugPanel.AddReducerCreated(reducerName, query)
}

// handleSelectorPattern creates a new consciousness selector
func (o *Outliner) handleSelectorPattern(pattern ConsciousnessPattern, nodeID string) {
	// Parse selector definition: "selector:: (name_a, name_b) => toc for tech craft zine"
	content := pattern.Content

	// Extract inputs and output format (simplified parsing)
	if strings.Contains(content, "=>") {
		parts := strings.Split(content, "=>")
		if len(parts) == 2 {
			inputPart := strings.TrimSpace(parts[0])
			outputFormat := strings.TrimSpace(parts[1])

			// Extract reducer names from (name_a, name_b) format
			inputPart = strings.Trim(inputPart, "()")
			inputs := strings.Split(inputPart, ",")
			for i, input := range inputs {
				inputs[i] = strings.TrimSpace(input)
			}

			// Create transform function
			transform := func(reducerInputs map[string][]DispatchAction) string {
				var result strings.Builder
				result.WriteString(fmt.Sprintf("# %s\n\n", outputFormat))

				for reducerName, actions := range reducerInputs {
					result.WriteString(fmt.Sprintf("## From %s (%d items)\n", reducerName, len(actions)))
					for _, action := range actions {
						result.WriteString(fmt.Sprintf("- %s: %s\n", action.PatternType, action.Content))
					}
					result.WriteString("\n")
				}

				return result.String()
			}

			selectorName := fmt.Sprintf("selector_%s", generateNodeID()[:8])
			o.dispatch.AddSelector(selectorName, inputs, transform)
			o.debugPanel.AddSelectorCreated(selectorName, outputFormat)
		}
	}
}

// renderNodeContent renders node text with consciousness metadata based on detail mode
func (o *Outliner) renderNodeContent(node OutlineNode) string {
	baseText := o.renderLinksInText(node.Text)

	// Detect pattern type from text
	patternType := o.detectPatternType(baseText)

	if !o.detailMode {
		// Simple mode - show text with color coding and capture indicators
		if patternType != "" {
			var style lipgloss.Style

			// Add color coding for different pattern types
			switch patternType {
			case "ctx":
				style = lipgloss.NewStyle().Foreground(lipgloss.Color("14")) // cyan
			case "eureka":
				style = lipgloss.NewStyle().Foreground(lipgloss.Color("11")) // yellow
			case "decision":
				style = lipgloss.NewStyle().Foreground(lipgloss.Color("9")) // red
			case "highlight":
				style = lipgloss.NewStyle().Foreground(lipgloss.Color("10")) // green
			case "gotcha":
				style = lipgloss.NewStyle().Foreground(lipgloss.Color("13")) // magenta
			case "bridge":
				style = lipgloss.NewStyle().Foreground(lipgloss.Color("12")) // blue
			case "dispatch":
				style = lipgloss.NewStyle().Foreground(lipgloss.Color("15")).Bold(true) // bright white, bold
			case "reducer":
				style = lipgloss.NewStyle().Foreground(lipgloss.Color("6")).Bold(true) // bright cyan, bold
			case "selector":
				style = lipgloss.NewStyle().Foreground(lipgloss.Color("5")).Bold(true) // bright magenta, bold
			case "imprint":
				style = lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Bold(true) // bright yellow, bold
			default:
				style = lipgloss.NewStyle().Foreground(lipgloss.Color("8")) // gray
			}

			// Add subtle capture indicator
			text := baseText
			if !node.Captured {
				text += " ●" // Uncaptured indicator
			} else {
				text += " ○" // Captured indicator
			}

			return style.Render(text)
		}
		return baseText
	}

	// Detail mode - show full metadata
	var details strings.Builder
	details.WriteString(baseText)

	if patternType != "" {
		details.WriteString(fmt.Sprintf(" [%s]", patternType))
	}

	if !node.Captured {
		details.WriteString(" [uncaptured]")
	}

	details.WriteString(fmt.Sprintf(" [id:%s]", node.ID[:8])) // Show short ID
	details.WriteString(fmt.Sprintf(" [%s]", node.ModifiedAt.Format("15:04")))

	return details.String()
}

// detectPatternType identifies the consciousness pattern type from text
func (o *Outliner) detectPatternType(text string) string {
	patterns := []string{
		"ctx::", "eureka::", "decision::", "highlight::", "gotcha::", "bridge::", "concept::", "mode::", "project::",
		"dispatch::", "reducer::", "selector::", "imprint::", "sigil::",
	}

	for _, pattern := range patterns {
		if strings.Contains(text, pattern) {
			return strings.TrimSuffix(pattern, "::")
		}
	}

	return ""
}

// extractLinks finds all [[concept]] links in text
func (o *Outliner) extractLinks(text string) []string {
	linkRegex := regexp.MustCompile(`\[\[([^\]]+)\]\]`)
	matches := linkRegex.FindAllStringSubmatch(text, -1)

	var links []string
	for _, match := range matches {
		if len(match) >= 2 {
			concept := strings.TrimSpace(match[1])
			if concept != "" {
				links = append(links, concept)
			}
		}
	}

	return links
}

// updateNodeLinks updates a node's links and the global link registry
func (o *Outliner) updateNodeLinks(nodeIndex int) {
	if nodeIndex >= len(o.lines) {
		return
	}

	node := &o.lines[nodeIndex]

	// Remove old links from registry
	for _, oldLink := range node.Links {
		o.removeLinkFromRegistry(oldLink, node.ID)
	}

	// Extract new links
	newLinks := o.extractLinks(node.Text)
	node.Links = newLinks

	// Add new links to registry
	for _, link := range newLinks {
		o.addLinkToRegistry(link, node.ID)
	}

	// Update backlinks for all nodes
	o.updateBacklinks()
}

// addLinkToRegistry adds a node ID to a concept's registry
func (o *Outliner) addLinkToRegistry(concept, nodeID string) {
	if o.linkRegistry[concept] == nil {
		o.linkRegistry[concept] = []string{}
	}

	// Check if already exists
	for _, id := range o.linkRegistry[concept] {
		if id == nodeID {
			return
		}
	}

	o.linkRegistry[concept] = append(o.linkRegistry[concept], nodeID)
}

// removeLinkFromRegistry removes a node ID from a concept's registry
func (o *Outliner) removeLinkFromRegistry(concept, nodeID string) {
	if o.linkRegistry[concept] == nil {
		return
	}

	for i, id := range o.linkRegistry[concept] {
		if id == nodeID {
			o.linkRegistry[concept] = append(o.linkRegistry[concept][:i], o.linkRegistry[concept][i+1:]...)
			break
		}
	}

	// Clean up empty entries
	if len(o.linkRegistry[concept]) == 0 {
		delete(o.linkRegistry, concept)
	}
}

// updateBacklinks updates backlink information for all nodes
func (o *Outliner) updateBacklinks() {
	// Clear all backlinks
	for i := range o.lines {
		o.lines[i].Backlinks = []string{}
	}

	// Rebuild backlinks from link registry
	for concept, nodeIDs := range o.linkRegistry {
		// Find nodes that contain this concept as text (not just links)
		for i, node := range o.lines {
			if strings.Contains(strings.ToLower(node.Text), strings.ToLower(concept)) {
				// This node mentions the concept, so all nodes linking to this concept
				// should have this node as a backlink
				for _, linkingNodeID := range nodeIDs {
					// Find the linking node and add this node as a backlink
					for j := range o.lines {
						if o.lines[j].ID == linkingNodeID {
							o.lines[i].Backlinks = append(o.lines[i].Backlinks, linkingNodeID)
							break
						}
					}
				}
			}
		}
	}
}

// renderLinksInText applies visual styling to [[links]] in text
func (o *Outliner) renderLinksInText(text string) string {
	linkRegex := regexp.MustCompile(`\[\[([^\]]+)\]\]`)

	// Style for links
	linkStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("4")). // Blue
		Underline(true)

	return linkRegex.ReplaceAllStringFunc(text, func(match string) string {
		// Extract the concept name
		concept := strings.Trim(match, "[]")
		return linkStyle.Render("[[" + concept + "]]")
	})
}
