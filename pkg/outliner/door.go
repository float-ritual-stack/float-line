package outliner

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Door represents a pluggable interface that can be embedded in the outliner
type Door interface {
	// Name returns the door's identifier (e.g., "chat", "repl", "markdown")
	Name() string

	// Init initializes the door with optional parameters
	Init(params map[string]string) tea.Cmd

	// Update handles messages for this door
	Update(msg tea.Msg) (Door, tea.Cmd)

	// View renders the door's interface
	View(width, height int) string

	// IsActive returns whether this door is currently active/focused
	IsActive() bool

	// Activate/Deactivate control door focus
	Activate()
	Deactivate()

	// GetState returns serializable state for persistence
	GetState() map[string]interface{}

	// SetState restores door from serialized state
	SetState(state map[string]interface{})

	// OnConsciousnessCapture is called when consciousness patterns are detected
	OnConsciousnessCapture(patterns []ConsciousnessPattern)
}

// DoorRegistry manages available door types
type DoorRegistry struct {
	doors map[string]func() Door // door name -> constructor function
}

// NewDoorRegistry creates a new door registry with built-in doors
func NewDoorRegistry() *DoorRegistry {
	registry := &DoorRegistry{
		doors: make(map[string]func() Door),
	}

	// Register built-in doors
	registry.Register("chat", func() Door { return NewChatDoor() })
	registry.Register("repl", func() Door { return NewReplDoor() })
	registry.Register("markdown", func() Door { return NewMarkdownDoor() })
	registry.Register("consciousness", func() Door { return NewConsciousnessDoor() })

	return registry
}

// Register adds a new door type to the registry
func (dr *DoorRegistry) Register(name string, constructor func() Door) {
	dr.doors[name] = constructor
}

// Create creates a new door instance by name
func (dr *DoorRegistry) Create(name string) Door {
	if constructor, exists := dr.doors[name]; exists {
		return constructor()
	}
	return nil
}

// GetAvailable returns list of available door names
func (dr *DoorRegistry) GetAvailable() []string {
	var names []string
	for name := range dr.doors {
		names = append(names, name)
	}
	return names
}

// DoorInstance represents an active door in the outliner
type DoorInstance struct {
	ID       string                 // Unique instance ID
	DoorType string                 // Type of door (chat, repl, etc.)
	NodeID   string                 // ID of the node that spawned this door
	Door     Door                   // The actual door implementation
	Params   map[string]string      // Parameters passed to the door
	State    map[string]interface{} // Persistent state
}

// ChatDoor - Simple chat interface door
type ChatDoor struct {
	active   bool
	messages []string
	input    string
	style    lipgloss.Style
}

func NewChatDoor() Door {
	return &ChatDoor{
		messages: []string{"Welcome to consciousness chat!"},
		style:    lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1),
	}
}

func (cd *ChatDoor) Name() string { return "chat" }

func (cd *ChatDoor) Init(params map[string]string) tea.Cmd {
	return nil
}

func (cd *ChatDoor) Update(msg tea.Msg) (Door, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if !cd.active {
			return cd, nil
		}

		switch msg.String() {
		case "enter":
			if cd.input != "" {
				cd.messages = append(cd.messages, "> "+cd.input)
				cd.input = ""
			}
		case "backspace":
			if len(cd.input) > 0 {
				cd.input = cd.input[:len(cd.input)-1]
			}
		default:
			if len(msg.String()) == 1 {
				cd.input += msg.String()
			}
		}
	}

	return cd, nil
}

func (cd *ChatDoor) View(width, height int) string {
	content := ""

	// Show recent messages
	for _, msg := range cd.messages {
		content += msg + "\n"
	}

	// Show input line
	if cd.active {
		content += "> " + cd.input + "█"
	} else {
		content += "> " + cd.input
	}

	return cd.style.Width(width - 4).Height(height - 4).Render(content)
}

func (cd *ChatDoor) IsActive() bool { return cd.active }
func (cd *ChatDoor) Activate()      { cd.active = true }
func (cd *ChatDoor) Deactivate()    { cd.active = false }

func (cd *ChatDoor) GetState() map[string]interface{} {
	return map[string]interface{}{
		"messages": cd.messages,
		"input":    cd.input,
	}
}

func (cd *ChatDoor) SetState(state map[string]interface{}) {
	if messages, ok := state["messages"].([]string); ok {
		cd.messages = messages
	}
	if input, ok := state["input"].(string); ok {
		cd.input = input
	}
}

func (cd *ChatDoor) OnConsciousnessCapture(patterns []ConsciousnessPattern) {
	for _, pattern := range patterns {
		cd.messages = append(cd.messages, "🧠 "+pattern.Type+":: "+pattern.Content)
	}
}

// ReplDoor - Code execution door (placeholder)
type ReplDoor struct {
	active bool
	style  lipgloss.Style
}

func NewReplDoor() Door {
	return &ReplDoor{
		style: lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1),
	}
}

func (rd *ReplDoor) Name() string                          { return "repl" }
func (rd *ReplDoor) Init(params map[string]string) tea.Cmd { return nil }
func (rd *ReplDoor) Update(msg tea.Msg) (Door, tea.Cmd)    { return rd, nil }
func (rd *ReplDoor) View(width, height int) string {
	return rd.style.Width(width - 4).Height(height - 4).Render("REPL Door - Coming Soon!")
}
func (rd *ReplDoor) IsActive() bool                                         { return rd.active }
func (rd *ReplDoor) Activate()                                              { rd.active = true }
func (rd *ReplDoor) Deactivate()                                            { rd.active = false }
func (rd *ReplDoor) GetState() map[string]interface{}                       { return map[string]interface{}{} }
func (rd *ReplDoor) SetState(state map[string]interface{})                  {}
func (rd *ReplDoor) OnConsciousnessCapture(patterns []ConsciousnessPattern) {}

// MarkdownDoor - Rich markdown rendering door (placeholder)
type MarkdownDoor struct {
	active bool
	style  lipgloss.Style
}

func NewMarkdownDoor() Door {
	return &MarkdownDoor{
		style: lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1),
	}
}

func (md *MarkdownDoor) Name() string                          { return "markdown" }
func (md *MarkdownDoor) Init(params map[string]string) tea.Cmd { return nil }
func (md *MarkdownDoor) Update(msg tea.Msg) (Door, tea.Cmd)    { return md, nil }
func (md *MarkdownDoor) View(width, height int) string {
	return md.style.Width(width - 4).Height(height - 4).Render("Markdown Door - Coming Soon!")
}
func (md *MarkdownDoor) IsActive() bool                                         { return md.active }
func (md *MarkdownDoor) Activate()                                              { md.active = true }
func (md *MarkdownDoor) Deactivate()                                            { md.active = false }
func (md *MarkdownDoor) GetState() map[string]interface{}                       { return map[string]interface{}{} }
func (md *MarkdownDoor) SetState(state map[string]interface{})                  {}
func (md *MarkdownDoor) OnConsciousnessCapture(patterns []ConsciousnessPattern) {}

// ConsciousnessDoor - Consciousness pattern visualization door (placeholder)
type ConsciousnessDoor struct {
	active bool
	style  lipgloss.Style
}

func NewConsciousnessDoor() Door {
	return &ConsciousnessDoor{
		style: lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1),
	}
}

func (cd *ConsciousnessDoor) Name() string                          { return "consciousness" }
func (cd *ConsciousnessDoor) Init(params map[string]string) tea.Cmd { return nil }
func (cd *ConsciousnessDoor) Update(msg tea.Msg) (Door, tea.Cmd)    { return cd, nil }
func (cd *ConsciousnessDoor) View(width, height int) string {
	return cd.style.Width(width - 4).Height(height - 4).Render("Consciousness Door - Pattern Visualization Coming Soon!")
}
func (cd *ConsciousnessDoor) IsActive() bool                                         { return cd.active }
func (cd *ConsciousnessDoor) Activate()                                              { cd.active = true }
func (cd *ConsciousnessDoor) Deactivate()                                            { cd.active = false }
func (cd *ConsciousnessDoor) GetState() map[string]interface{}                       { return map[string]interface{}{} }
func (cd *ConsciousnessDoor) SetState(state map[string]interface{})                  {}
func (cd *ConsciousnessDoor) OnConsciousnessCapture(patterns []ConsciousnessPattern) {}
