package sanitizer

import (
	"html"
	"regexp"
	"strings"
)

// Sanitizer sanitizes HTML content
type Sanitizer struct {
	allowedTags map[string]bool
}

// NewSanitizer creates a new HTML sanitizer
func NewSanitizer() *Sanitizer {
	return &Sanitizer{
		allowedTags: map[string]bool{
			"p":          true,
			"br":         true,
			"strong":     true,
			"em":         true,
			"u":          true,
			"h1":         true,
			"h2":         true,
			"h3":         true,
			"h4":         true,
			"h5":         true,
			"h6":         true,
			"ul":         true,
			"ol":         true,
			"li":         true,
			"a":          true,
			"code":       true,
			"pre":        true,
			"blockquote": true,
		},
	}
}

// SanitizeHTML sanitizes HTML content by removing dangerous tags and attributes
func (s *Sanitizer) SanitizeHTML(content string) string {
	if content == "" {
		return ""
	}

	// Remove script tags and content
	scriptRegex := regexp.MustCompile(`(?i)<script[^>]*>.*?</script>`)
	content = scriptRegex.ReplaceAllString(content, "")

	// Remove style tags and content
	styleRegex := regexp.MustCompile(`(?i)<style[^>]*>.*?</style>`)
	content = styleRegex.ReplaceAllString(content, "")

	// Remove event handlers (onclick, onload, etc.)
	eventRegex := regexp.MustCompile(`(?i)\s+on\w+\s*=\s*["'][^"']*["']`)
	content = eventRegex.ReplaceAllString(content, "")

	// Remove javascript: protocol in href
	jsProtocolRegex := regexp.MustCompile(`(?i)javascript:`)
	content = jsProtocolRegex.ReplaceAllString(content, "")

	// Remove data: protocol (can be used for XSS)
	dataProtocolRegex := regexp.MustCompile(`(?i)data:`)
	content = dataProtocolRegex.ReplaceAllString(content, "")

	// Escape HTML entities
	content = html.UnescapeString(content)
	content = html.EscapeString(content)

	return content
}

// StripHTML removes all HTML tags from content
func (s *Sanitizer) StripHTML(content string) string {
	if content == "" {
		return ""
	}

	// Remove all HTML tags
	tagRegex := regexp.MustCompile(`<[^>]*>`)
	content = tagRegex.ReplaceAllString(content, "")

	// Decode HTML entities
	content = html.UnescapeString(content)

	// Trim whitespace
	content = strings.TrimSpace(content)

	return content
}

// TruncateText truncates text to a maximum length, adding ellipsis if needed
func (s *Sanitizer) TruncateText(text string, maxLength int) string {
	if text == "" {
		return ""
	}

	if maxLength <= 0 {
		return text
	}

	if len(text) <= maxLength {
		return text
	}

	// Truncate and add ellipsis
	truncated := text[:maxLength]

	// Try to break at last space to avoid cutting words
	lastSpace := strings.LastIndex(truncated, " ")
	if lastSpace > 0 && lastSpace > maxLength-20 {
		truncated = truncated[:lastSpace]
	}

	return truncated + "..."
}

// CalculateReadingTime estimates reading time in minutes based on word count
// Assumes average reading speed of 200 words per minute
func (s *Sanitizer) CalculateReadingTime(content string) int {
	if content == "" {
		return 0
	}

	// Strip HTML first
	plainText := s.StripHTML(content)

	// Count words
	words := strings.Fields(plainText)
	wordCount := len(words)

	// Calculate minutes (minimum 1 minute)
	readingTime := wordCount / 200
	if readingTime == 0 {
		readingTime = 1
	}

	return readingTime
}

// ExtractPlainText extracts plain text from HTML content
func (s *Sanitizer) ExtractPlainText(content string) string {
	return s.StripHTML(content)
}
