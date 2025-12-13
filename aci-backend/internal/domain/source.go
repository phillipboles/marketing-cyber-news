package domain

import (
	"fmt"
	"net/url"
	"time"

	"github.com/google/uuid"
)

// Source represents a news source in the system
type Source struct {
	ID            uuid.UUID  `json:"id"`
	Name          string     `json:"name"`
	URL           string     `json:"url"`
	Description   *string    `json:"description,omitempty"`
	IsActive      bool       `json:"is_active"`
	TrustScore    float64    `json:"trust_score"`
	LastScrapedAt *time.Time `json:"last_scraped_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

// Validate validates the source entity
func (s *Source) Validate() error {
	if s.ID == uuid.Nil {
		return fmt.Errorf("source ID is required")
	}

	if s.Name == "" {
		return fmt.Errorf("name is required")
	}

	if len(s.Name) > 200 {
		return fmt.Errorf("name must not exceed 200 characters")
	}

	if s.URL == "" {
		return fmt.Errorf("URL is required")
	}

	if err := validateURL(s.URL); err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	if s.TrustScore < 0.0 || s.TrustScore > 1.0 {
		return fmt.Errorf("trust score must be between 0.0 and 1.0, got: %f", s.TrustScore)
	}

	if s.Description != nil && len(*s.Description) > 1000 {
		return fmt.Errorf("description must not exceed 1000 characters")
	}

	if s.CreatedAt.IsZero() {
		return fmt.Errorf("created_at is required")
	}

	return nil
}

// validateURL checks if the URL is valid
func validateURL(rawURL string) error {
	if rawURL == "" {
		return fmt.Errorf("URL cannot be empty")
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("failed to parse URL: %w", err)
	}

	if parsedURL.Scheme == "" {
		return fmt.Errorf("URL must have a scheme (http or https)")
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("URL scheme must be http or https, got: %s", parsedURL.Scheme)
	}

	if parsedURL.Host == "" {
		return fmt.Errorf("URL must have a host")
	}

	return nil
}

// UpdateLastScraped updates the last scraped timestamp
func (s *Source) UpdateLastScraped() {
	now := time.Now()
	s.LastScrapedAt = &now
}

// Activate sets the source as active
func (s *Source) Activate() {
	s.IsActive = true
}

// Deactivate sets the source as inactive
func (s *Source) Deactivate() {
	s.IsActive = false
}

// UpdateTrustScore updates the trust score with validation
func (s *Source) UpdateTrustScore(score float64) error {
	if score < 0.0 || score > 1.0 {
		return fmt.Errorf("trust score must be between 0.0 and 1.0, got: %f", score)
	}
	s.TrustScore = score
	return nil
}

// NewSource creates a new source with default values
func NewSource(name, rawURL string, description *string) (*Source, error) {
	if err := validateURL(rawURL); err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	now := time.Now()
	return &Source{
		ID:          uuid.New(),
		Name:        name,
		URL:         rawURL,
		Description: description,
		IsActive:    true,
		TrustScore:  0.5, // Default neutral trust score
		CreatedAt:   now,
	}, nil
}
