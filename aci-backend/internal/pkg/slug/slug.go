package slug

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

var (
	// nonAlphanumericRegex matches any character that is not alphanumeric or hyphen
	nonAlphanumericRegex = regexp.MustCompile(`[^a-z0-9-]+`)

	// consecutiveHyphensRegex matches multiple consecutive hyphens
	consecutiveHyphensRegex = regexp.MustCompile(`-+`)

	// validSlugRegex validates final slug format
	validSlugRegex = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
)

const (
	maxSlugLength     = 200
	uniqueSuffixStart = 2
)

// Generate creates a URL-friendly slug from text
//
// Transformations:
// - Converts to lowercase
// - Replaces spaces with hyphens
// - Removes special characters
// - Removes consecutive hyphens
// - Truncates to max 200 characters
// - Removes leading/trailing hyphens
//
// Example:
//
//	Generate("Hello World!")           // "hello-world"
//	Generate("Cyber Security & AI")    // "cyber-security-ai"
//	Generate("CVE-2024-12345 Analysis") // "cve-2024-12345-analysis"
func Generate(text string) string {
	if text == "" {
		return ""
	}

	// Convert to lowercase
	slug := strings.ToLower(text)

	// Replace spaces with hyphens
	slug = strings.ReplaceAll(slug, " ", "-")

	// Remove non-alphanumeric characters (except hyphens)
	slug = nonAlphanumericRegex.ReplaceAllString(slug, "")

	// Replace consecutive hyphens with single hyphen
	slug = consecutiveHyphensRegex.ReplaceAllString(slug, "-")

	// Remove leading and trailing hyphens
	slug = strings.Trim(slug, "-")

	// Truncate to max length
	if len(slug) > maxSlugLength {
		slug = slug[:maxSlugLength]
		// Remove trailing hyphen if truncation created one
		slug = strings.TrimRight(slug, "-")
	}

	return slug
}

// GenerateUnique creates a unique slug by appending a numeric suffix if needed
//
// The existsFn function should return true if the slug already exists.
// If the base slug exists, appends "-2", "-3", etc. until a unique slug is found.
//
// Example:
//
//	existsFn := func(slug string) bool {
//	    return db.SlugExists(slug)
//	}
//	slug := GenerateUnique("hello-world", existsFn)
//	// Returns "hello-world" if available
//	// Returns "hello-world-2" if "hello-world" exists
//	// Returns "hello-world-3" if both exist, etc.
func GenerateUnique(text string, existsFn func(slug string) bool) string {
	if existsFn == nil {
		return Generate(text)
	}

	baseSlug := Generate(text)
	if baseSlug == "" {
		return ""
	}

	// If base slug doesn't exist, use it
	if !existsFn(baseSlug) {
		return baseSlug
	}

	// Try appending numbers until we find a unique slug
	suffix := uniqueSuffixStart
	for {
		candidateSlug := fmt.Sprintf("%s-%d", baseSlug, suffix)

		// Ensure candidate doesn't exceed max length
		if len(candidateSlug) > maxSlugLength {
			// Truncate base slug to make room for suffix
			suffixLen := len(fmt.Sprintf("-%d", suffix))
			truncatedBase := baseSlug[:maxSlugLength-suffixLen]
			truncatedBase = strings.TrimRight(truncatedBase, "-")
			candidateSlug = fmt.Sprintf("%s-%d", truncatedBase, suffix)
		}

		if !existsFn(candidateSlug) {
			return candidateSlug
		}

		suffix++

		// Safety check to prevent infinite loop
		if suffix > 10000 {
			// This should never happen in practice
			return fmt.Sprintf("%s-%d", baseSlug[:maxSlugLength-10], suffix)
		}
	}
}

// IsValid checks if a string is a valid slug
//
// Valid slug requirements:
// - Only lowercase letters, numbers, and hyphens
// - No leading or trailing hyphens
// - No consecutive hyphens
// - Not empty
// - Maximum 200 characters
//
// Example:
//
//	IsValid("hello-world")        // true
//	IsValid("hello--world")       // false (consecutive hyphens)
//	IsValid("-hello-world")       // false (leading hyphen)
//	IsValid("Hello-World")        // false (uppercase)
//	IsValid("hello_world")        // false (underscore)
func IsValid(s string) bool {
	if s == "" {
		return false
	}

	if len(s) > maxSlugLength {
		return false
	}

	return validSlugRegex.MatchString(s)
}

// Sanitize ensures a string is a valid slug, generating one if necessary
//
// If the input is already a valid slug, returns it unchanged.
// Otherwise, generates a new slug from the input.
//
// Example:
//
//	Sanitize("hello-world")       // "hello-world" (already valid)
//	Sanitize("Hello World!")      // "hello-world" (generated)
//	Sanitize("--hello--")         // "hello" (sanitized)
func Sanitize(s string) string {
	if IsValid(s) {
		return s
	}
	return Generate(s)
}

// ToTitle converts a slug back to a human-readable title
//
// Transformations:
// - Replaces hyphens with spaces
// - Capitalizes first letter of each word
//
// Example:
//
//	ToTitle("hello-world")              // "Hello World"
//	ToTitle("cyber-security-analysis")  // "Cyber Security Analysis"
func ToTitle(slug string) string {
	if slug == "" {
		return ""
	}

	// Replace hyphens with spaces
	words := strings.Split(slug, "-")

	// Capitalize first letter of each word
	for i, word := range words {
		if word == "" {
			continue
		}

		// Convert first rune to uppercase
		runes := []rune(word)
		runes[0] = unicode.ToUpper(runes[0])
		words[i] = string(runes)
	}

	return strings.Join(words, " ")
}
