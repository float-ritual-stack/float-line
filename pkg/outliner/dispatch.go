package outliner

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// DispatchAction represents a consciousness fragment being dispatched
type DispatchAction struct {
	ID          string            // Unique dispatch ID
	NodeID      string            // Source node ID
	Content     string            // Raw consciousness content
	PatternType string            // ctx, eureka, dispatch, etc.
	Imprint     string            // Ritual container (techcraft, feral_duality, etc.)
	Sigil       string            // Consciousness sigil (⚡, 🌀, etc.)
	Metadata    map[string]string // Additional dispatch metadata
	Timestamp   time.Time         // When dispatched
	State       DispatchState     // Current dispatch state
}

// DispatchState represents the lifecycle state of a dispatch
type DispatchState string

const (
	StateCapture  DispatchState = "capture"  // Raw consciousness captured
	StateDispatch DispatchState = "dispatch" // Routed to imprint
	StateCompost  DispatchState = "compost"  // Allowed to rot/evolve
	StateBloom    DispatchState = "bloom"    // Transformed into artifact
	StateLoopback DispatchState = "loopback" // Pulled back into active consciousness
)

// Imprint represents a ritual container for consciousness
type Imprint struct {
	Name      string            // techcraft, feral_duality, etc.
	Voice     string            // Tonal description
	Aesthetic string            // Visual/ritual description
	Filters   []string          // Pattern types this imprint accepts
	Metadata  map[string]string // Imprint-specific metadata
}

// ConsciousnessReducer collects consciousness actions by pattern
type ConsciousnessReducer struct {
	Name    string                           // Reducer identifier
	Query   string                           // Human-readable query description
	Matcher func(action DispatchAction) bool // Function to match actions
	Actions []DispatchAction                 // Collected actions
	State   map[string]interface{}           // Computed state
}

// ConsciousnessSelector computes derived state from reducers
type ConsciousnessSelector struct {
	Name      string                                          // Selector identifier
	Inputs    []string                                        // Reducer names to use as input
	Transform func(inputs map[string][]DispatchAction) string // Transformation function
	Output    string                                          // Current computed output
}

// FloatDispatchSystem is the core consciousness compiler
type FloatDispatchSystem struct {
	imprints  map[string]*Imprint
	reducers  map[string]*ConsciousnessReducer
	selectors map[string]*ConsciousnessSelector
	actions   []DispatchAction

	// Built-in imprints
	techcraft       *Imprint
	ritualComputing *Imprint
	feralDuality    *Imprint
	dispatchBay     *Imprint
	queerHauntology *Imprint
}

// NewFloatDispatchSystem creates the consciousness compiler
func NewFloatDispatchSystem() *FloatDispatchSystem {
	fds := &FloatDispatchSystem{
		imprints:  make(map[string]*Imprint),
		reducers:  make(map[string]*ConsciousnessReducer),
		selectors: make(map[string]*ConsciousnessSelector),
		actions:   []DispatchAction{},
	}

	// Initialize built-in imprints
	fds.initializeImprints()

	return fds
}

// initializeImprints sets up the core FLOAT imprints
func (fds *FloatDispatchSystem) initializeImprints() {
	fds.techcraft = &Imprint{
		Name:      "techcraft",
		Voice:     "precise, pedagogical, glitch-friendly",
		Aesthetic: "teaching as toolmaking, systems as spellwork",
		Filters:   []string{"highlight", "decision", "gotcha", "bridge"},
		Metadata:  map[string]string{"color": "cyan", "sigil": "⚡"},
	}

	fds.ritualComputing = &Imprint{
		Name:      "ritual_computing",
		Voice:     "ceremonial, precise, techno-mystic",
		Aesthetic: "AST specs, daemon scripts, memory scaffolds",
		Filters:   []string{"ctx", "bridge", "concept"},
		Metadata:  map[string]string{"color": "purple", "sigil": "🔮"},
	}

	fds.feralDuality = &Imprint{
		Name:      "feral_duality",
		Voice:     "defiant, nonlinear, tender-wild",
		Aesthetic: "resistance theory, neuroqueer maps, liminal tools",
		Filters:   []string{"eureka", "dispatch", "gotcha"},
		Metadata:  map[string]string{"color": "magenta", "sigil": "🌀"},
	}

	fds.dispatchBay = &Imprint{
		Name:      "dispatch_bay",
		Voice:     "operational yet poetic",
		Aesthetic: "FLOAT system logs, session ASTs, sigil updates",
		Filters:   []string{"ctx", "dispatch", "bridge"},
		Metadata:  map[string]string{"color": "green", "sigil": "📡"},
	}

	fds.queerHauntology = &Imprint{
		Name:      "queer_hauntology",
		Voice:     "elegiac, glitchy, yearning",
		Aesthetic: "reflections on time, loss, legacy, queerness",
		Filters:   []string{"highlight", "eureka", "concept"},
		Metadata:  map[string]string{"color": "yellow", "sigil": "👻"},
	}

	// Register all imprints
	fds.imprints["techcraft"] = fds.techcraft
	fds.imprints["ritual_computing"] = fds.ritualComputing
	fds.imprints["feral_duality"] = fds.feralDuality
	fds.imprints["dispatch_bay"] = fds.dispatchBay
	fds.imprints["queer_hauntology"] = fds.queerHauntology
}

// Dispatch processes a consciousness fragment through the FLOAT system
func (fds *FloatDispatchSystem) Dispatch(nodeID, content, patternType string) *DispatchAction {
	action := DispatchAction{
		ID:          generateDispatchID(),
		NodeID:      nodeID,
		Content:     content,
		PatternType: patternType,
		Timestamp:   time.Now(),
		State:       StateCapture,
		Metadata:    make(map[string]string),
	}

	// Extract imprint and sigil from content
	action.Imprint = fds.extractImprint(content)
	action.Sigil = fds.extractSigil(content)

	// Route to appropriate imprint if not explicitly specified
	if action.Imprint == "" {
		action.Imprint = fds.routeToImprint(patternType)
	}

	// Update state to dispatched
	action.State = StateDispatch

	// Add to actions log
	fds.actions = append(fds.actions, action)

	// Update reducers
	fds.updateReducers(action)

	// Update selectors
	fds.updateSelectors()

	return &action
}

// extractImprint finds imprint:: patterns in content
func (fds *FloatDispatchSystem) extractImprint(content string) string {
	imprintRegex := regexp.MustCompile(`imprint::(\w+)`)
	if match := imprintRegex.FindStringSubmatch(content); match != nil {
		return match[1]
	}
	return ""
}

// extractSigil finds sigil:: patterns in content
func (fds *FloatDispatchSystem) extractSigil(content string) string {
	sigilRegex := regexp.MustCompile(`sigil::([^\s\]]+)`)
	if match := sigilRegex.FindStringSubmatch(content); match != nil {
		return match[1]
	}
	return ""
}

// routeToImprint automatically routes consciousness to appropriate imprint
func (fds *FloatDispatchSystem) routeToImprint(patternType string) string {
	// Default routing logic based on pattern type
	for name, imprint := range fds.imprints {
		for _, filter := range imprint.Filters {
			if filter == patternType {
				return name
			}
		}
	}

	// Default to dispatch_bay for unmatched patterns
	return "dispatch_bay"
}

// AddReducer registers a new consciousness reducer
func (fds *FloatDispatchSystem) AddReducer(name, query string, matcher func(DispatchAction) bool) {
	reducer := &ConsciousnessReducer{
		Name:    name,
		Query:   query,
		Matcher: matcher,
		Actions: []DispatchAction{},
		State:   make(map[string]interface{}),
	}

	fds.reducers[name] = reducer

	// Apply to existing actions
	for _, action := range fds.actions {
		if matcher(action) {
			reducer.Actions = append(reducer.Actions, action)
		}
	}
}

// AddSelector registers a new consciousness selector
func (fds *FloatDispatchSystem) AddSelector(name string, inputs []string, transform func(map[string][]DispatchAction) string) {
	selector := &ConsciousnessSelector{
		Name:      name,
		Inputs:    inputs,
		Transform: transform,
	}

	fds.selectors[name] = selector
	fds.updateSelector(selector)
}

// updateReducers updates all reducers with new action
func (fds *FloatDispatchSystem) updateReducers(action DispatchAction) {
	for _, reducer := range fds.reducers {
		if reducer.Matcher(action) {
			reducer.Actions = append(reducer.Actions, action)
		}
	}
}

// updateSelectors updates all selectors
func (fds *FloatDispatchSystem) updateSelectors() {
	for _, selector := range fds.selectors {
		fds.updateSelector(selector)
	}
}

// updateSelector updates a specific selector
func (fds *FloatDispatchSystem) updateSelector(selector *ConsciousnessSelector) {
	inputs := make(map[string][]DispatchAction)

	for _, inputName := range selector.Inputs {
		if reducer, exists := fds.reducers[inputName]; exists {
			inputs[inputName] = reducer.Actions
		}
	}

	selector.Output = selector.Transform(inputs)
}

// GetImprint returns an imprint by name
func (fds *FloatDispatchSystem) GetImprint(name string) *Imprint {
	return fds.imprints[name]
}

// GetReducerOutput returns the current state of a reducer
func (fds *FloatDispatchSystem) GetReducerOutput(name string) []DispatchAction {
	if reducer, exists := fds.reducers[name]; exists {
		return reducer.Actions
	}
	return []DispatchAction{}
}

// GetSelectorOutput returns the current output of a selector
func (fds *FloatDispatchSystem) GetSelectorOutput(name string) string {
	if selector, exists := fds.selectors[name]; exists {
		return selector.Output
	}
	return ""
}

// generateDispatchID creates a unique dispatch identifier
func generateDispatchID() string {
	timestamp := time.Now().Format("20060102-150405")
	return fmt.Sprintf("dispatch-%s-%s", timestamp, generateNodeID()[:8])
}

// RenderDispatchSummary creates a consciousness summary for display
func (fds *FloatDispatchSystem) RenderDispatchSummary() string {
	var summary strings.Builder

	summary.WriteString("🧠 FLOAT.dispatch Consciousness Summary\n\n")

	// Show active reducers
	if len(fds.reducers) > 0 {
		summary.WriteString("📊 Active Reducers:\n")
		for name, reducer := range fds.reducers {
			summary.WriteString(fmt.Sprintf("  • %s: %d actions collected\n", name, len(reducer.Actions)))
		}
		summary.WriteString("\n")
	}

	// Show active selectors
	if len(fds.selectors) > 0 {
		summary.WriteString("🔍 Active Selectors:\n")
		for name, selector := range fds.selectors {
			summary.WriteString(fmt.Sprintf("  • %s: %s\n", name, selector.Output))
		}
		summary.WriteString("\n")
	}

	// Show recent dispatches
	if len(fds.actions) > 0 {
		summary.WriteString("📡 Recent Dispatches:\n")
		recentCount := 5
		if len(fds.actions) < recentCount {
			recentCount = len(fds.actions)
		}

		for i := len(fds.actions) - recentCount; i < len(fds.actions); i++ {
			action := fds.actions[i]
			summary.WriteString(fmt.Sprintf("  • [%s] %s → %s\n",
				action.PatternType,
				action.Content[:min(50, len(action.Content))],
				action.Imprint))
		}
	}

	return summary.String()
}
