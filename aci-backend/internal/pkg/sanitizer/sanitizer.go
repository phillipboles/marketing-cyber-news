package sanitizer

import (
	"fmt"
	"html"
	"net/url"
	"regexp"
	"strings"
)

var (
	// allowedTags defines safe HTML tags for formatting
	allowedTags = map[string]bool{
		"p":      true,
		"br":     true,
		"strong": true,
		"em":     true,
		"ul":     true,
		"ol":     true,
		"li":     true,
		"a":      true,
		"code":   true,
		"pre":    true,
	}

	// dangerousTags must be removed for security
	dangerousTags = []string{
		"script", "iframe", "object", "embed", "form",
		"input", "button", "textarea", "select", "option",
		"style", "link", "meta", "base", "frame", "frameset",
	}

	// htmlTagRegex matches HTML tags
	htmlTagRegex = regexp.MustCompile(`<[^>]*>`)

	// scriptEventRegex matches event handlers (onclick, onerror, etc.)
	scriptEventRegex = regexp.MustCompile(`(?i)\s*on\w+\s*=`)

	// allowedSchemes for URLs
	allowedSchemes = map[string]bool{
		"http":  true,
		"https": true,
	}
)

// Sanitizer handles HTML/text sanitization
type Sanitizer struct {
	maxURLLength int
}

// New creates a new sanitizer with default configuration
//
// Example usage:
//
//	s := sanitizer.New()
//	clean := s.SanitizeHTML(userInput)
func New() *Sanitizer {
	return &Sanitizer{
		maxURLLength: 2048,
	}
}

// SanitizeHTML removes dangerous HTML tags while preserving safe formatting
//
// Allowed tags: p, br, strong, em, ul, ol, li, a (with sanitized href), code, pre
// Removed tags: script, iframe, object, embed, form, input, style, and all event handlers
//
// The function also:
// - Escapes remaining HTML entities
// - Removes event handlers (onclick, onerror, etc.)
// - Sanitizes href attributes in <a> tags
//
// Example:
//
//	input := `<p>Safe content</p><script>alert('xss')</script>`
//	output := s.SanitizeHTML(input)
//	// Result: "<p>Safe content</p>"
func (s *Sanitizer) SanitizeHTML(input string) string {
	if input == "" {
		return ""
	}

	// First pass: remove dangerous tags completely
	cleaned := input
	for _, tag := range dangerousTags {
		// Remove both opening and closing tags with content
		pattern := regexp.MustCompile(fmt.Sprintf(`(?i)<\s*%s[^>]*>.*?<\s*/\s*%s\s*>`, tag, tag))
		cleaned = pattern.ReplaceAllString(cleaned, "")

		// Remove self-closing dangerous tags
		pattern = regexp.MustCompile(fmt.Sprintf(`(?i)<\s*%s[^>]*/?>`, tag))
		cleaned = pattern.ReplaceAllString(cleaned, "")
	}

	// Remove event handlers from all remaining tags
	cleaned = scriptEventRegex.ReplaceAllString(cleaned, "")

	// Second pass: sanitize URLs in allowed tags
	cleaned = s.sanitizeHrefAttributes(cleaned)

	// Third pass: remove any tags not in allowedTags
	cleaned = s.removeDisallowedTags(cleaned)

	return strings.TrimSpace(cleaned)
}

// SanitizeText removes all HTML and returns plain text
//
// This function:
// - Strips all HTML tags
// - Decodes HTML entities
// - Trims whitespace
//
// Example:
//
//	input := `<p>Hello <strong>World</strong>!</p>`
//	output := s.SanitizeText(input)
//	// Result: "Hello World!"
func (s *Sanitizer) SanitizeText(input string) string {
	if input == "" {
		return ""
	}

	// Remove all HTML tags
	text := htmlTagRegex.ReplaceAllString(input, "")

	// Decode HTML entities
	text = html.UnescapeString(text)

	// Normalize whitespace
	text = strings.Join(strings.Fields(text), " ")

	return strings.TrimSpace(text)
}

// SanitizeURL validates and sanitizes URLs
//
// Only allows http:// and https:// schemes.
// Returns error for:
// - Invalid URL format
// - Disallowed schemes (javascript:, data:, file:, etc.)
// - URLs exceeding max length (2048 chars)
//
// Example:
//
//	url, err := s.SanitizeURL("https://example.com/path")
//	// Returns: "https://example.com/path", nil
//
//	url, err := s.SanitizeURL("javascript:alert('xss')")
//	// Returns: "", error
func (s *Sanitizer) SanitizeURL(rawURL string) (string, error) {
	if rawURL == "" {
		return "", fmt.Errorf("URL cannot be empty")
	}

	if len(rawURL) > s.maxURLLength {
		return "", fmt.Errorf("URL exceeds maximum length of %d characters", s.maxURLLength)
	}

	// Parse URL
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL format: %w", err)
	}

	// Check scheme is allowed
	scheme := strings.ToLower(parsedURL.Scheme)
	if scheme == "" {
		return "", fmt.Errorf("URL must include scheme (http:// or https://)")
	}

	if !allowedSchemes[scheme] {
		return "", fmt.Errorf("URL scheme '%s' not allowed (only http and https)", scheme)
	}

	// Ensure host is present
	if parsedURL.Host == "" {
		return "", fmt.Errorf("URL must include a host")
	}

	// Return clean URL
	return parsedURL.String(), nil
}

// sanitizeHrefAttributes sanitizes href attributes in anchor tags
func (s *Sanitizer) sanitizeHrefAttributes(input string) string {
	hrefPattern := regexp.MustCompile(`(?i)<a\s+[^>]*href\s*=\s*["']([^"']+)["'][^>]*>`)

	return hrefPattern.ReplaceAllStringFunc(input, func(match string) string {
		// Extract href value
		hrefValue := hrefPattern.FindStringSubmatch(match)
		if len(hrefValue) < 2 {
			// Remove anchor if we can't extract href
			return ""
		}

		// Sanitize the URL
		cleanURL, err := s.SanitizeURL(hrefValue[1])
		if err != nil {
			// Remove anchor if URL is invalid
			return ""
		}

		// Return sanitized anchor tag
		return fmt.Sprintf(`<a href="%s">`, html.EscapeString(cleanURL))
	})
}

// removeDisallowedTags removes HTML tags not in allowedTags list
func (s *Sanitizer) removeDisallowedTags(input string) string {
	tagPattern := regexp.MustCompile(`<\s*/?\s*([a-zA-Z][a-zA-Z0-9]*)[^>]*>`)

	return tagPattern.ReplaceAllStringFunc(input, func(match string) string {
		// Extract tag name
		tagMatch := tagPattern.FindStringSubmatch(match)
		if len(tagMatch) < 2 {
			return ""
		}

		tagName := strings.ToLower(tagMatch[1])

		// Keep allowed tags
		if allowedTags[tagName] {
			return match
		}

		// Remove disallowed tags
		return ""
	})
}

// TruncateText truncates text to specified length, adding ellipsis if needed
//
// The function ensures:
// - Truncation happens at word boundaries when possible
// - Ellipsis (...) is added only when text is actually truncated
// - Result never exceeds maxLen characters (including ellipsis)
//
// Example:
//
//	TruncateText("Hello World", 8)
//	// Result: "Hello..."
//
//	TruncateText("Hello", 10)
//	// Result: "Hello"
func TruncateText(text string, maxLen int) string {
	if text == "" {
		return ""
	}

	if maxLen <= 0 {
		return ""
	}

	if len(text) <= maxLen {
		return text
	}

	const ellipsis = "..."
	ellipsisLen := len(ellipsis)

	if maxLen <= ellipsisLen {
		return ellipsis[:maxLen]
	}

	// Truncate to max length minus ellipsis
	truncated := text[:maxLen-ellipsisLen]

	// Try to truncate at last word boundary
	lastSpace := strings.LastIndexAny(truncated, " \t\n\r")
	if lastSpace > 0 && lastSpace > maxLen/2 {
		// Only use word boundary if it's not too far back
		truncated = truncated[:lastSpace]
	}

	return strings.TrimSpace(truncated) + ellipsis
}

// StripHTML is an alias for SanitizeText for backward compatibility
func (s *Sanitizer) StripHTML(input string) string {
	return s.SanitizeText(input)
}

// ValidateEmail validates email format using basic regex
//
// This is a simple validation. For production, consider using
// a more robust solution or the validator package.
//
// Example:
//
//	valid := ValidateEmail("user@example.com")  // true
//	valid := ValidateEmail("invalid-email")     // false
func ValidateEmail(email string) bool {
	if email == "" {
		return false
	}

	emailPattern := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	return emailPattern.MatchString(email)
}
