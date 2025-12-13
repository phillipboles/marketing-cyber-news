package domain

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	// hexColorRegex validates hex color codes (#RRGGBB or #RGB)
	hexColorRegex = regexp.MustCompile(`^#([A-Fa-f0-9]{6}|[A-Fa-f0-9]{3})$`)

	// slugRegex validates slugs (lowercase alphanumeric with hyphens)
	slugRegex = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
)

// Category represents a news category in the system
type Category struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description *string   `json:"description,omitempty"`
	Color       string    `json:"color"`
	Icon        *string   `json:"icon,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// Validate validates the category entity
func (c *Category) Validate() error {
	if c.ID == uuid.Nil {
		return fmt.Errorf("category ID is required")
	}

	if c.Name == "" {
		return fmt.Errorf("name is required")
	}

	if len(c.Name) > 100 {
		return fmt.Errorf("name must not exceed 100 characters")
	}

	if c.Slug == "" {
		return fmt.Errorf("slug is required")
	}

	if !slugRegex.MatchString(c.Slug) {
		return fmt.Errorf("slug must contain only lowercase letters, numbers, and hyphens")
	}

	if len(c.Slug) > 100 {
		return fmt.Errorf("slug must not exceed 100 characters")
	}

	if c.Color == "" {
		return fmt.Errorf("color is required")
	}

	if !hexColorRegex.MatchString(c.Color) {
		return fmt.Errorf("color must be a valid hex color code (e.g., #FF5733)")
	}

	if c.Description != nil && len(*c.Description) > 500 {
		return fmt.Errorf("description must not exceed 500 characters")
	}

	if c.Icon != nil && len(*c.Icon) > 100 {
		return fmt.Errorf("icon must not exceed 100 characters")
	}

	if c.CreatedAt.IsZero() {
		return fmt.Errorf("created_at is required")
	}

	return nil
}

// GenerateSlug creates a URL-friendly slug from the category name
func GenerateSlug(name string) string {
	// Convert to lowercase
	slug := strings.ToLower(name)

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

	return slug
}

// NewCategory creates a new category with generated slug
func NewCategory(name, color string, description *string, icon *string) *Category {
	now := time.Now()
	return &Category{
		ID:          uuid.New(),
		Name:        name,
		Slug:        GenerateSlug(name),
		Description: description,
		Color:       color,
		Icon:        icon,
		CreatedAt:   now,
	}
}
