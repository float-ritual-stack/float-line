package tui

import (
	"fmt"
	"strings"

	"github.com/evanschultz/float-rw-client/pkg/models"
)

// Messages
type booksLoadedMsg struct {
	books []models.Book
}

type highlightsLoadedMsg struct {
	highlights  []models.Highlight
	nextPageURL string
}

type highlightRenderedMsg struct {
	content     string
	noteContent string
}

type highlightSavedMsg struct{}

type errMsg struct {
	err error
}

// List items
type bookItem struct {
	book models.Book
}

func (i bookItem) FilterValue() string { return i.book.Title }
func (i bookItem) Title() string       { return i.book.Title }
func (i bookItem) Description() string {
	return fmt.Sprintf("%s • %d highlights", i.book.Author, i.book.NumHighlights)
}

type highlightItem struct {
	highlight models.Highlight
}

func (i highlightItem) FilterValue() string { return i.highlight.Text }
func (i highlightItem) Title() string {
	// Show much more of the highlight text
	text := i.highlight.Text
	// Remove excessive whitespace and newlines for display
	text = strings.ReplaceAll(text, "\n", " ")
	text = strings.Join(strings.Fields(text), " ")
	
	if len(text) > 200 {
		return text[:197] + "..."
	}
	return text
}

func (i highlightItem) Description() string {
	parts := []string{}
	
	// Show note preview if present
	if i.highlight.Note != "" {
		note := i.highlight.Note
		// Clean up note for display
		note = strings.ReplaceAll(note, "\n", " ")
		note = strings.Join(strings.Fields(note), " ")
		
		if len(note) > 150 {
			note = note[:147] + "..."
		}
		parts = append(parts, "📝 "+note)
	}
	
	// Add metadata
	metadata := []string{}
	if i.highlight.URL != "" {
		metadata = append(metadata, "🔗 Source")
	}
	if i.highlight.HighlightedAt != nil {
		metadata = append(metadata, i.highlight.HighlightedAt.Format("Jan 2, 2006"))
	}
	
	if len(metadata) > 0 && len(parts) == 0 {
		parts = append(parts, strings.Join(metadata, " • "))
	}
	
	return strings.Join(parts, "\n")
}
