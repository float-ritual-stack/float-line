package outliner

import (
	"strings"
	"testing"
)

// createReducerMatcher extracts the matcher logic for testing
func createReducerMatcher(query string) func(DispatchAction) bool {
	return func(action DispatchAction) bool {
		content := strings.ToLower(action.Content)
		queryLower := strings.ToLower(query)

		// Parse query for keywords and pattern types
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
}

func TestReducerMatching(t *testing.T) {
	tests := []struct {
		name          string
		query         string
		actionContent string
		actionType    string
		shouldMatch   bool
	}{
		{
			name:          "simple keyword match",
			query:         "collect all actions that mention test",
			actionContent: "test pattern one",
			actionType:    "dispatch",
			shouldMatch:   true,
		},
		{
			name:          "keyword not found",
			query:         "collect all actions that mention test",
			actionContent: "unrelated pattern",
			actionType:    "dispatch",
			shouldMatch:   false,
		},
		{
			name:          "multiple keywords",
			query:         "collect all actions that mention door patterns",
			actionContent: "door system implementation",
			actionType:    "dispatch",
			shouldMatch:   true,
		},
		{
			name:          "about syntax",
			query:         "collect all bridges about rangle",
			actionContent: "rangle team collaboration",
			actionType:    "bridge",
			shouldMatch:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the matcher logic directly
			matcher := createReducerMatcher(tt.query)

			// Create a test action
			action := DispatchAction{
				Content:     tt.actionContent,
				PatternType: tt.actionType,
			}

			// Test the matcher
			result := matcher(action)

			if result != tt.shouldMatch {
				t.Errorf("Expected match=%v for query='%s' content='%s', got %v",
					tt.shouldMatch, tt.query, tt.actionContent, result)
			}
		})
	}
}

func TestConsciousnessCapture(t *testing.T) {
	content := `# Test File
• reducer:: test collect all actions that mention test
• dispatch:: test pattern one
• dispatch:: unrelated pattern
• eureka:: test breakthrough!`

	outliner := New()
	outliner.SetContent(content)

	// Capture callback to track reducer updates
	var collectedActions []DispatchAction
	outliner.dispatch.SetReducerUpdateCallback(func(reducerName string, action DispatchAction) {
		collectedActions = append(collectedActions, action)
	})

	// Debug: Check content was set
	content_check := outliner.GetContent()
	t.Logf("Content set: %s", content_check)

	// Debug: Check what patterns are parsed
	parsed := outliner.parser.Parse(content_check)
	t.Logf("Parsed %d patterns", len(parsed.ConsciousnessData))
	for _, pattern := range parsed.ConsciousnessData {
		t.Logf("Pattern: %s -> %s", pattern.Type, pattern.Content)
	}

	// Trigger consciousness processing
	outliner.TriggerConsciousnessCapture()

	// Debug: Check if reducer was created
	reducers := outliner.dispatch.GetReducers()
	if len(reducers) == 0 {
		t.Error("No reducers were created from content")
		return
	}

	t.Logf("Created %d reducers", len(reducers))
	for name := range reducers {
		t.Logf("Reducer: %s", name)
	}

	// Debug: Check if actions were dispatched
	actions := outliner.dispatch.GetActions()
	t.Logf("Dispatched %d actions", len(actions))
	for _, action := range actions {
		t.Logf("Action: %s -> %s", action.PatternType, action.Content)
	}

	// Verify results
	if len(collectedActions) == 0 {
		t.Error("Expected reducer to collect actions, but none were collected")
	}

	// Check that the right actions were collected
	expectedMatches := 0
	for _, action := range collectedActions {
		if strings.Contains(action.Content, "test") {
			expectedMatches++
		}
	}

	if expectedMatches == 0 {
		t.Error("Expected to collect actions containing 'test', but none found")
		for _, action := range collectedActions {
			t.Logf("Collected: %s", action.Content)
		}
	}
}
