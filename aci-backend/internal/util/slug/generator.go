package slug

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Generator generates unique URL-friendly slugs
type Generator struct {
	maxLength int
}

// NewGenerator creates a new slug generator
func NewGenerator() *Generator {
	return &Generator{
		maxLength: 100,
	}
}

// Generate creates a URL-friendly slug from text
func (g *Generator) Generate(text string) string {
	if text == "" {
		return g.generateFallback()
	}

	// Convert to lowercase
	slug := strings.ToLower(text)

	// Replace spaces with hyphens
	slug = strings.ReplaceAll(slug, " ", "-")

	// Remove special characters (keep only alphanumeric and hyphens)
	reg := regexp.MustCompile(`[^a-z0-9-]+`)
	slug = reg.ReplaceAllString(slug, "")

	// Remove consecutive hyphens
	reg = regexp.MustCompile(`-+`)
	slug = reg.ReplaceAllString(slug, "-")

	// Trim hyphens from start and end
	slug = strings.Trim(slug, "-")

	// Truncate to max length
	if len(slug) > g.maxLength {
		slug = slug[:g.maxLength]
		// Ensure it doesn't end with a hyphen after truncation
		slug = strings.TrimRight(slug, "-")
	}

	// If empty after cleaning, use fallback
	if slug == "" {
		return g.generateFallback()
	}

	return slug
}

// GenerateUnique creates a unique slug by appending a timestamp and short UUID
func (g *Generator) GenerateUnique(text string) string {
	baseSlug := g.Generate(text)

	// Append timestamp and short UUID for uniqueness
	timestamp := time.Now().Unix()
	shortUUID := uuid.New().String()[:8]

	return fmt.Sprintf("%s-%d-%s", baseSlug, timestamp, shortUUID)
}

// generateFallback creates a fallback slug when text is empty or invalid
func (g *Generator) generateFallback() string {
	timestamp := time.Now().Unix()
	shortUUID := uuid.New().String()[:8]
	return fmt.Sprintf("article-%d-%s", timestamp, shortUUID)
}

// SetMaxLength sets the maximum slug length
func (g *Generator) SetMaxLength(maxLength int) {
	if maxLength > 0 {
		g.maxLength = maxLength
	}
}
