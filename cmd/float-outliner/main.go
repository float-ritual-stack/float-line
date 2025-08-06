package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/evanschultz/float-rw-client/pkg/outliner"
	"github.com/spf13/cobra"
)

var (
	testScenario string
)

var rootCmd = &cobra.Command{
	Use:   "float-outliner [file|directory]",
	Short: "A consciousness-enabled outliner with :: pattern detection",
	Long: `Float Outliner is a terminal-based outliner with built-in consciousness technology integration.
It automatically detects and captures :: patterns (ctx::, eureka::, decision::, etc.) for FLOAT ecosystem integration.

You can pass either a file to edit directly, or a directory to use as working directory.`,
	Args: cobra.MaximumNArgs(1),
	Run:  runOutliner,
}

func runOutliner(cmd *cobra.Command, args []string) {
	var path string
	if len(args) > 0 {
		path = args[0]
	}

	// Handle test scenarios
	if testScenario != "" {
		path = createTestScenario(testScenario)
		if path == "" {
			fmt.Printf("Unknown test scenario: %s\n", testScenario)
			fmt.Println("Available scenarios: reducer-basic, reducer-complex, patterns-all")
			os.Exit(1)
		}
	}

	app := NewOutlinerApp(path)

	p := tea.NewProgram(app, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running outliner: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringVar(&testScenario, "test", "", "Create test scenario (reducer-basic, reducer-complex, patterns-all)")
}

// createTestScenario creates a test file for the specified scenario
func createTestScenario(scenario string) string {
	var content string
	var filename string

	switch scenario {
	case "reducer-basic":
		filename = "test-reducer-basic.md"
		content = `# Reducer Basic Test

• reducer:: test collect all actions that mention test

• dispatch:: test pattern one
• dispatch:: test pattern two  
• eureka:: test breakthrough!
• decision:: use test approach [priority:: high]

• dispatch:: unrelated pattern (should not be collected)
`

	case "reducer-complex":
		filename = "test-reducer-complex.md"
		content = `# Reducer Complex Test

• reducer:: door_patterns collect all actions that are bridges or dispatches about door
• reducer:: tech_stuff collect all decisions and gotchas about technology

• bridge:: [[door]] connects to [[consciousness-tech]] [bridge-id:: DOOR-001]
• dispatch:: [[door]] system implementation
• decision:: implement [[technology]] stack [priority:: high]
• gotcha:: [[technology]] requires careful setup [fix:: documentation]
• eureka:: unrelated insight (should not be collected)

• selector:: (door_patterns, tech_stuff) => implementation guide for door tech
`

	case "patterns-all":
		filename = "test-patterns-all.md"
		content = `# All Patterns Test

• ctx:: 2025-08-05 6:00pm [project:: [[test-project]]] [mode:: testing]
• eureka:: All patterns working! [concept:: [[consciousness-tech]]]
• decision:: Test all pattern types [priority:: high]
• highlight:: This is important for testing [importance:: critical]
• gotcha:: Debug panel needs to be visible [fix:: check-visibility]
• bridge:: [[test-project]] connects to [[consciousness-tech]] [bridge-id:: TEST-001]
• dispatch:: raw consciousness fragment [sigil:: ⚡] [imprint:: techcraft]

• reducer:: test_patterns collect all actions about test
• selector:: (test_patterns) => test summary report
`

	default:
		return ""
	}

	// Write test file
	if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
		fmt.Printf("Error creating test file: %v\n", err)
		return ""
	}

	fmt.Printf("Created test scenario: %s\n", filename)
	return filename
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// OutlinerApp is the main application model
type OutlinerApp struct {
	outliner outliner.Outliner
	filename string
	width    int
	height   int
	saved    bool
}

// NewOutlinerApp creates a new outliner application
func NewOutlinerApp(filename string) *OutlinerApp {
	app := &OutlinerApp{
		outliner: outliner.New(),
		filename: filename,
		saved:    true,
	}

	// Load file if provided
	if filename != "" {
		app.loadFile()
	}

	app.outliner.Focus()
	return app
}

// Init initializes the application
func (a *OutlinerApp) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (a *OutlinerApp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.outliner.SetSize(a.width, a.height-2) // Leave room for status bar

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if !a.saved {
				// TODO: Add confirmation dialog
			}
			return a, tea.Quit

		case "ctrl+s":
			a.saveFile()
			a.saved = true
			return a, nil

		case "ctrl+o":
			// TODO: Add file open dialog
			return a, nil

		case "ctrl+t":
			// Toggle detail mode - pass to outliner
			newOutliner, cmd := a.outliner.Update(msg)
			a.outliner = newOutliner
			return a, cmd

		case "ctrl+l":
			// Toggle debug panel - pass to outliner
			newOutliner, cmd := a.outliner.Update(msg)
			a.outliner = newOutliner
			return a, cmd

		default:
			// Pass all other keys to outliner
			newOutliner, cmd := a.outliner.Update(msg)
			a.outliner = newOutliner
			a.saved = false // Mark as unsaved on any edit
			return a, cmd
		}
	}

	return a, nil
}

// View renders the application
func (a *OutlinerApp) View() string {
	if a.width == 0 || a.height == 0 {
		return "Loading..."
	}

	// Main outliner view
	content := a.outliner.View()

	// Status bar
	statusBar := a.renderStatusBar()

	return content + "\n" + statusBar
}

// renderStatusBar creates the bottom status bar
func (a *OutlinerApp) renderStatusBar() string {
	filename := a.filename
	if filename == "" {
		filename = "[untitled]"
	}

	saveStatus := ""
	if !a.saved {
		saveStatus = " [modified]"
	}

	detailMode := ""
	if a.outliner.IsDetailMode() {
		detailMode = " [DETAIL]"
	}

	debugMode := ""
	if a.outliner.IsDebugVisible() {
		debugMode = " [DEBUG]"
	}

	status := fmt.Sprintf(" %s%s%s%s | Ctrl+S: Save | Ctrl+T: Detail | Ctrl+L: Debug | Q: Quit", filename, saveStatus, detailMode, debugMode)

	// Pad to full width
	padding := a.width - len(status)
	if padding > 0 {
		status += fmt.Sprintf("%*s", padding, "")
	}

	return status
}

// loadFile loads content from the specified file
func (a *OutlinerApp) loadFile() {
	if a.filename == "" {
		return
	}

	content, err := os.ReadFile(a.filename)
	if err != nil {
		// File doesn't exist or can't be read - start with empty content
		return
	}

	a.outliner.SetContent(string(content))
	a.saved = true
}

// saveFile saves the current content to file
func (a *OutlinerApp) saveFile() {
	if a.filename == "" {
		// TODO: Add save-as dialog
		a.filename = "untitled.md"
	}

	content := a.outliner.GetContent()

	// Trigger consciousness capture before saving
	a.outliner.TriggerConsciousnessCapture()

	err := os.WriteFile(a.filename, []byte(content), 0644)
	if err != nil {
		// TODO: Show error message
		fmt.Printf("Error saving file: %v\n", err)
		return
	}

	a.saved = true
}
