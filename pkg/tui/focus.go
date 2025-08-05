package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// FocusableComponent represents a component that can receive focus
type FocusableComponent interface {
	// Focus gives focus to the component
	Focus() tea.Cmd
	// Blur removes focus from the component
	Blur() tea.Cmd
	// Focused returns whether the component currently has focus
	Focused() bool
}

// FocusManager handles focus state for multiple components
type FocusManager struct {
	components []FocusableComponent
	current    int
	enabled    bool
}

// NewFocusManager creates a new focus manager
func NewFocusManager(components ...FocusableComponent) *FocusManager {
	fm := &FocusManager{
		components: components,
		current:    0,
		enabled:    true,
	}

	// Focus the first component if available
	if len(components) > 0 {
		components[0].Focus()
	}

	return fm
}

// Next moves focus to the next component
func (fm *FocusManager) Next() tea.Cmd {
	if !fm.enabled || len(fm.components) == 0 {
		return nil
	}

	// Blur current
	if fm.current < len(fm.components) {
		fm.components[fm.current].Blur()
	}

	// Move to next
	fm.current = (fm.current + 1) % len(fm.components)

	// Focus new current
	return fm.components[fm.current].Focus()
}

// Previous moves focus to the previous component
func (fm *FocusManager) Previous() tea.Cmd {
	if !fm.enabled || len(fm.components) == 0 {
		return nil
	}

	// Blur current
	if fm.current < len(fm.components) {
		fm.components[fm.current].Blur()
	}

	// Move to previous
	fm.current = (fm.current - 1 + len(fm.components)) % len(fm.components)

	// Focus new current
	return fm.components[fm.current].Focus()
}

// SetFocus sets focus to a specific component index
func (fm *FocusManager) SetFocus(index int) tea.Cmd {
	if !fm.enabled || index < 0 || index >= len(fm.components) {
		return nil
	}

	// Blur current
	if fm.current < len(fm.components) {
		fm.components[fm.current].Blur()
	}

	// Set new focus
	fm.current = index
	return fm.components[fm.current].Focus()
}

// Current returns the currently focused component index
func (fm *FocusManager) Current() int {
	return fm.current
}

// Enable/Disable focus management
func (fm *FocusManager) SetEnabled(enabled bool) {
	fm.enabled = enabled
	if !enabled && fm.current < len(fm.components) {
		fm.components[fm.current].Blur()
	}
}

// FocusStyles defines styles for focused/unfocused states
type FocusStyles struct {
	Focused   lipgloss.Style
	Unfocused lipgloss.Style
}

// DefaultFocusStyles returns sensible default focus styles
func DefaultFocusStyles() FocusStyles {
	return FocusStyles{
		Focused: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")),
		Unfocused: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")),
	}
}
