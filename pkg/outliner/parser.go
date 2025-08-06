package outliner

import (
	"regexp"
	"strings"
)

// StructuredContent represents parsed outliner content with semantic sections
type StructuredContent struct {
	Highlight         string
	Note              string
	Tags              []string
	Meta              map[string]string
	Raw               string // Original content
	ConsciousnessData []ConsciousnessPattern
}

// ConsciousnessPattern represents detected :: patterns for evna dispatch
type ConsciousnessPattern struct {
	Type    string // ctx, highlight, eureka, decision, etc.
	Content string
	Line    int
	Context map[string]string // parsed [key:: value] annotations
}

// AnnotationPattern represents a recognized annotation pattern
type AnnotationPattern struct {
	Name    string
	Pattern *regexp.Regexp
	Handler func(content string) interface{}
}

// Parser handles structured annotation parsing
type Parser struct {
	patterns []AnnotationPattern
}

// NewParser creates a new parser with default patterns
func NewParser() *Parser {
	p := &Parser{}

	// Register default patterns
	p.RegisterPattern("highlight", `^•\s*highlight::\s*(.+)$`, func(content string) interface{} {
		return strings.TrimSpace(content)
	})

	p.RegisterPattern("note", `^•\s*note::\s*(.*)$`, func(content string) interface{} {
		return strings.TrimSpace(content)
	})

	p.RegisterPattern("tags", `^•\s*tags::\s*(.+)$`, func(content string) interface{} {
		tags := strings.Split(content, ",")
		for i, tag := range tags {
			tags[i] = strings.TrimSpace(tag)
		}
		return tags
	})

	p.RegisterPattern("meta", `^•\s*meta::\s*$`, func(content string) interface{} {
		return make(map[string]string)
	})

	// Generic key-value pattern for meta items
	p.RegisterPattern("meta_item", `^\s*•\s*(\w+)::\s*(.+)$`, func(content string) interface{} {
		return content
	})

	return p
}

// RegisterPattern adds a new annotation pattern
func (p *Parser) RegisterPattern(name string, pattern string, handler func(string) interface{}) {
	compiled := regexp.MustCompile(pattern)
	p.patterns = append(p.patterns, AnnotationPattern{
		Name:    name,
		Pattern: compiled,
		Handler: handler,
	})
}

// Parse extracts structured content from outliner text
func (p *Parser) Parse(content string) *StructuredContent {
	result := &StructuredContent{
		Meta:              make(map[string]string),
		Raw:               content,
		ConsciousnessData: []ConsciousnessPattern{},
	}

	lines := strings.Split(content, "\n")
	currentSection := ""

	for lineNum, line := range lines {
		// Detect consciousness patterns first
		p.detectConsciousnessPatterns(line, lineNum+1, result)
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Check for main section headers
		if match := regexp.MustCompile(`^•\s*highlight::\s*(.+)$`).FindStringSubmatch(line); match != nil {
			result.Highlight = strings.TrimSpace(match[1])
			currentSection = "highlight"
			continue
		}

		if match := regexp.MustCompile(`^•\s*note::\s*(.*)$`).FindStringSubmatch(line); match != nil {
			noteContent := strings.TrimSpace(match[1])
			if noteContent != "" {
				result.Note = noteContent
			}
			currentSection = "note"
			continue
		}

		if match := regexp.MustCompile(`^•\s*tags::\s*(.+)$`).FindStringSubmatch(line); match != nil {
			tags := strings.Split(match[1], ",")
			for i, tag := range tags {
				tags[i] = strings.TrimSpace(tag)
			}
			result.Tags = tags
			currentSection = "tags"
			continue
		}

		if regexp.MustCompile(`^•\s*meta::\s*$`).MatchString(line) {
			currentSection = "meta"
			continue
		}

		// Handle sub-items based on current section
		if strings.HasPrefix(line, "  •") || strings.HasPrefix(line, "    •") {
			subContent := strings.TrimPrefix(line, "  •")
			subContent = strings.TrimPrefix(subContent, "    •")
			subContent = strings.TrimSpace(subContent)

			switch currentSection {
			case "note":
				if result.Note == "" {
					result.Note = subContent
				} else {
					result.Note += "\n" + subContent
				}

			case "meta":
				// Parse key-value pairs in meta section
				if match := regexp.MustCompile(`^(\w+)::\s*(.+)$`).FindStringSubmatch(subContent); match != nil {
					key := strings.TrimSpace(match[1])
					value := strings.TrimSpace(match[2])
					result.Meta[key] = value
				}
			}
		}
	}

	return result
}

// Lint checks for common issues in structured content
func (p *Parser) Lint(content string) []LintIssue {
	var issues []LintIssue

	lines := strings.Split(content, "\n")
	hasHighlight := false
	hasNote := false

	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Check for required sections
		if regexp.MustCompile(`^•\s*highlight::`).MatchString(line) {
			hasHighlight = true
		}
		if regexp.MustCompile(`^•\s*note::`).MatchString(line) {
			hasNote = true
		}

		// Check for malformed annotations
		if strings.Contains(line, "::") && !regexp.MustCompile(`^\s*•.*::`).MatchString(line) {
			issues = append(issues, LintIssue{
				Line:     i + 1,
				Type:     "format",
				Message:  "Annotation should start with bullet point",
				Severity: "warning",
			})
		}

		// Check for empty annotation values
		if match := regexp.MustCompile(`^•\s*(\w+)::\s*$`).FindStringSubmatch(line); match != nil {
			if match[1] != "note" && match[1] != "meta" { // These can be empty
				issues = append(issues, LintIssue{
					Line:     i + 1,
					Type:     "content",
					Message:  "Empty annotation: " + match[1],
					Severity: "info",
				})
			}
		}
	}

	// Check for missing required sections
	if !hasHighlight {
		issues = append(issues, LintIssue{
			Line:     0,
			Type:     "structure",
			Message:  "Missing highlight:: section",
			Severity: "error",
		})
	}

	if !hasNote {
		issues = append(issues, LintIssue{
			Line:     0,
			Type:     "structure",
			Message:  "Missing note:: section",
			Severity: "warning",
		})
	}

	return issues
}

// LintIssue represents a problem found during linting
type LintIssue struct {
	Line     int    // 0 for general issues
	Type     string // "format", "content", "structure"
	Message  string
	Severity string // "error", "warning", "info"
}

// detectConsciousnessPatterns finds :: patterns for evna dispatch
func (p *Parser) detectConsciousnessPatterns(line string, lineNum int, result *StructuredContent) {
	// Common consciousness patterns
	patterns := map[string]*regexp.Regexp{
		"ctx":       regexp.MustCompile(`ctx::\s*(.+)`),
		"highlight": regexp.MustCompile(`highlight::\s*(.+)`),
		"eureka":    regexp.MustCompile(`eureka::\s*(.+)`),
		"decision":  regexp.MustCompile(`decision::\s*(.+)`),
		"gotcha":    regexp.MustCompile(`gotcha::\s*(.+)`),
		"bridge":    regexp.MustCompile(`bridge::\s*(.+)`),
		"mode":      regexp.MustCompile(`mode::\s*(.+)`),
		"project":   regexp.MustCompile(`project::\s*(.+)`),
		"concept":   regexp.MustCompile(`concept::\s*(.+)`),
		"aka":       regexp.MustCompile(`aka::\s*(.+)`),
		// FLOAT system patterns
		"dispatch": regexp.MustCompile(`dispatch::\s*(.+)`),
		"reducer":  regexp.MustCompile(`reducer::\s*(.+)`),
		"selector": regexp.MustCompile(`selector::\s*(.+)`),
		"imprint":  regexp.MustCompile(`imprint::\s*(.+)`),
	}

	for patternType, regex := range patterns {
		if match := regex.FindStringSubmatch(line); match != nil {
			// Extract context annotations [key:: value]
			context := p.extractContextAnnotations(line)

			pattern := ConsciousnessPattern{
				Type:    patternType,
				Content: strings.TrimSpace(match[1]),
				Line:    lineNum,
				Context: context,
			}
			result.ConsciousnessData = append(result.ConsciousnessData, pattern)
		}
	}
}

// extractContextAnnotations finds [key:: value] patterns in text
func (p *Parser) extractContextAnnotations(text string) map[string]string {
	context := make(map[string]string)

	// Match [key:: value] patterns
	contextRegex := regexp.MustCompile(`\[(\w+)::\s*([^\]]+)\]`)
	matches := contextRegex.FindAllStringSubmatch(text, -1)

	for _, match := range matches {
		if len(match) >= 3 {
			key := strings.TrimSpace(match[1])
			value := strings.TrimSpace(match[2])
			context[key] = value
		}
	}

	return context
}

// ToReadwiseFormat converts structured content back to Readwise API format
func (sc *StructuredContent) ToReadwiseFormat() (highlight string, note string, tags []string) {
	return sc.Highlight, sc.Note, sc.Tags
}
