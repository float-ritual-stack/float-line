package outliner

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// EvnaDispatcher handles consciousness pattern dispatch to evna collections
type EvnaDispatcher struct {
	enabled  bool
	logError func(string, string) // callback for logging errors
}

// NewEvnaDispatcher creates a new evna dispatcher
func NewEvnaDispatcher() *EvnaDispatcher {
	return &EvnaDispatcher{
		enabled:  true,                    // TODO: make configurable
		logError: func(string, string) {}, // no-op by default
	}
}

// SetErrorLogger sets the error logging callback
func (ed *EvnaDispatcher) SetErrorLogger(logError func(string, string)) {
	ed.logError = logError
}

// DispatchPatterns sends consciousness patterns to evna collections
func (ed *EvnaDispatcher) DispatchPatterns(patterns []ConsciousnessPattern, source string) error {
	if !ed.enabled {
		return nil
	}

	for _, pattern := range patterns {
		if err := ed.dispatchSinglePattern(pattern, source); err != nil {
			// Log error but continue with other patterns
			ed.logError("EVNA_DISPATCH_WARNING", fmt.Sprintf("Failed to dispatch pattern %s: %v", pattern.Type, err))
		}
	}

	return nil
}

// dispatchSinglePattern sends a single pattern to appropriate evna collection
func (ed *EvnaDispatcher) dispatchSinglePattern(pattern ConsciousnessPattern, source string) error {
	// Build the dispatch text in FLOAT format
	timestamp := time.Now().Format("2006-01-02 3:04pm")

	var dispatchText strings.Builder
	dispatchText.WriteString(fmt.Sprintf("%s:: %s", pattern.Type, pattern.Content))

	// Add context annotations
	if len(pattern.Context) > 0 {
		for key, value := range pattern.Context {
			dispatchText.WriteString(fmt.Sprintf(" [%s:: %s]", key, value))
		}
	}

	// Add source metadata
	dispatchText.WriteString(fmt.Sprintf(" [source:: %s] [timestamp:: %s]", source, timestamp))

	// Route to appropriate collection based on pattern type
	collection := ed.routeToCollection(pattern.Type)

	// Use evna MCP to capture the pattern
	return ed.callEvnaMCP(dispatchText.String(), collection)
}

// routeToCollection determines which evna collection to use for a pattern type
func (ed *EvnaDispatcher) routeToCollection(patternType string) string {
	routing := map[string]string{
		"ctx":       "active_context_stream",
		"highlight": "float_highlights",
		"eureka":    "float_highlights",
		"decision":  "float_dispatch_bay",
		"gotcha":    "active_context_stream",
		"bridge":    "float_bridges",
		"mode":      "active_context_stream",
		"project":   "active_context_stream",
		"concept":   "float_highlights",
		"aka":       "float_highlights",
	}

	if collection, exists := routing[patternType]; exists {
		return collection
	}

	// Default to active_context_stream for unknown patterns
	return "active_context_stream"
}

// callEvnaMCP invokes evna pattern capture via structured output
func (ed *EvnaDispatcher) callEvnaMCP(text string, collection string) error {
	// Create the evna capture payload in FLOAT format
	payload := map[string]interface{}{
		"action":     "evna_capture",
		"text":       text,
		"collection": collection,
		"source":     "float-rw-client",
		"timestamp":  time.Now().Unix(),
		"iso_time":   time.Now().Format(time.RFC3339),
	}

	_, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal evna payload: %w", err)
	}

	// Output structured data for external processing (commented out to avoid console spam)
	// This can be captured by shell scripts, log processors, or MCP bridges
	// fmt.Printf("CONSCIOUSNESS_CAPTURE: %s\n", string(jsonPayload))

	// Note: Consciousness capture is now logged via the debug panel in the outliner
	return nil
}

// callEvnaCommand is a fallback method using command line evna tools
func (ed *EvnaDispatcher) callEvnaCommand(text string) error {
	// Try to call evna command line tool if available
	cmd := exec.Command("evna", "capture", text)
	output, err := cmd.CombinedOutput()

	if err != nil {
		return fmt.Errorf("evna command failed: %w, output: %s", err, string(output))
	}

	return nil
}
