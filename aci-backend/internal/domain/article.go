package domain

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Severity string

const (
	SeverityCritical      Severity = "critical"
	SeverityHigh          Severity = "high"
	SeverityMedium        Severity = "medium"
	SeverityLow           Severity = "low"
	SeverityInformational Severity = "informational"
)

// IsValid validates the severity value
func (s Severity) IsValid() bool {
	switch s {
	case SeverityCritical, SeverityHigh, SeverityMedium, SeverityLow, SeverityInformational:
		return true
	default:
		return false
	}
}

// IOC represents an Indicator of Compromise
type IOC struct {
	Type    string `json:"type"`              // ip, domain, hash, url
	Value   string `json:"value"`             // The actual IOC value
	Context string `json:"context,omitempty"` // Additional context
}

// IsValid validates the IOC structure
func (i *IOC) IsValid() bool {
	if i.Type == "" {
		return false
	}

	if i.Value == "" {
		return false
	}

	validTypes := map[string]bool{
		"ip":     true,
		"domain": true,
		"hash":   true,
		"url":    true,
	}

	return validTypes[i.Type]
}

// ArmorCTA represents a call to action for Armor.com marketing
type ArmorCTA struct {
	Type  string `json:"type"`  // product, service, consultation
	Title string `json:"title"` // Display title
	URL   string `json:"url"`   // Target URL
}

// IsValid validates the ArmorCTA structure
func (a *ArmorCTA) IsValid() bool {
	if a.Type == "" || a.Title == "" || a.URL == "" {
		return false
	}

	validTypes := map[string]bool{
		"product":      true,
		"service":      true,
		"consultation": true,
	}

	return validTypes[a.Type]
}

// Article represents a cybersecurity news article
type Article struct {
	ID         uuid.UUID `json:"id"`
	Title      string    `json:"title"`
	Slug       string    `json:"slug"`
	Content    string    `json:"content"`
	Summary    *string   `json:"summary,omitempty"`
	CategoryID uuid.UUID `json:"category_id"`
	Category   *Category `json:"category,omitempty"`
	SourceID   uuid.UUID `json:"source_id"`
	Source     *Source   `json:"source,omitempty"`
	SourceURL  string    `json:"source_url"`
	Severity   Severity  `json:"severity"`
	Tags       []string  `json:"tags"`
	CVEs       []string  `json:"cves"`
	Vendors    []string  `json:"vendors"`

	// AI Enrichment fields
	ThreatType         *string  `json:"threat_type,omitempty"`
	AttackVector       *string  `json:"attack_vector,omitempty"`
	ImpactAssessment   *string  `json:"impact_assessment,omitempty"`
	RecommendedActions []string `json:"recommended_actions,omitempty"`
	IOCs               []IOC    `json:"iocs,omitempty"`

	// Armor marketing
	ArmorRelevance float64    `json:"armor_relevance"`
	ArmorCTA       *ArmorCTA  `json:"armor_cta,omitempty"`

	// Internal scoring (not exposed to API)
	CompetitorScore       float64 `json:"-"`
	IsCompetitorFavorable bool    `json:"-"`

	// Metadata
	ReadingTimeMinutes int        `json:"reading_time_minutes"`
	ViewCount          int        `json:"view_count"`
	IsPublished        bool       `json:"is_published"`
	PublishedAt        time.Time  `json:"published_at"`
	EnrichedAt         *time.Time `json:"enriched_at,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

// Validate performs validation on the Article
func (a *Article) Validate() error {
	if a.Title == "" {
		return fmt.Errorf("title is required")
	}

	if a.Slug == "" {
		return fmt.Errorf("slug is required")
	}

	if a.Content == "" {
		return fmt.Errorf("content is required")
	}

	if a.CategoryID == uuid.Nil {
		return fmt.Errorf("category_id is required")
	}

	if a.SourceID == uuid.Nil {
		return fmt.Errorf("source_id is required")
	}

	if a.SourceURL == "" {
		return fmt.Errorf("source_url is required")
	}

	if !a.Severity.IsValid() {
		return fmt.Errorf("invalid severity value")
	}

	if a.ArmorRelevance < 0 || a.ArmorRelevance > 1 {
		return fmt.Errorf("armor_relevance must be between 0 and 1")
	}

	if a.ArmorCTA != nil && !a.ArmorCTA.IsValid() {
		return fmt.Errorf("invalid armor_cta")
	}

	for _, ioc := range a.IOCs {
		if !ioc.IsValid() {
			return fmt.Errorf("invalid IOC entry")
		}
	}

	if a.ReadingTimeMinutes < 0 {
		return fmt.Errorf("reading_time_minutes cannot be negative")
	}

	if a.ViewCount < 0 {
		return fmt.Errorf("view_count cannot be negative")
	}

	return nil
}

// ContainsKeyword checks if the article contains the given keyword in title, content, or summary
func (a *Article) ContainsKeyword(keyword string) bool {
	if keyword == "" {
		return false
	}

	lowerKeyword := strings.ToLower(keyword)

	if strings.Contains(strings.ToLower(a.Title), lowerKeyword) {
		return true
	}

	if strings.Contains(strings.ToLower(a.Content), lowerKeyword) {
		return true
	}

	if a.Summary != nil && strings.Contains(strings.ToLower(*a.Summary), lowerKeyword) {
		return true
	}

	return false
}

// HasCVE checks if the article mentions the given CVE
func (a *Article) HasCVE(cve string) bool {
	if cve == "" {
		return false
	}

	upperCVE := strings.ToUpper(cve)
	for _, articleCVE := range a.CVEs {
		if strings.ToUpper(articleCVE) == upperCVE {
			return true
		}
	}

	return false
}

// HasVendor checks if the article mentions the given vendor
func (a *Article) HasVendor(vendor string) bool {
	if vendor == "" {
		return false
	}

	lowerVendor := strings.ToLower(vendor)
	for _, articleVendor := range a.Vendors {
		if strings.ToLower(articleVendor) == lowerVendor {
			return true
		}
	}

	return false
}

// ArticleFilter represents query parameters for filtering articles
type ArticleFilter struct {
	CategoryID  *uuid.UUID
	SourceID    *uuid.UUID
	Severity    *Severity
	Tags        []string
	CVE         *string
	Vendor      *string
	DateFrom    *time.Time
	DateTo      *time.Time
	SearchQuery *string
	Page        int
	PageSize    int
}

// NewArticleFilter returns a filter with default values
func NewArticleFilter() *ArticleFilter {
	return &ArticleFilter{
		Page:     1,
		PageSize: 20,
	}
}

// Validate validates the filter parameters
func (f *ArticleFilter) Validate() error {
	if f.Page < 1 {
		return fmt.Errorf("page must be at least 1")
	}

	if f.PageSize < 1 {
		return fmt.Errorf("page_size must be at least 1")
	}

	if f.PageSize > 100 {
		return fmt.Errorf("page_size cannot exceed 100")
	}

	if f.Severity != nil && !f.Severity.IsValid() {
		return fmt.Errorf("invalid severity value")
	}

	if f.DateFrom != nil && f.DateTo != nil && f.DateFrom.After(*f.DateTo) {
		return fmt.Errorf("date_from cannot be after date_to")
	}

	return nil
}

// Offset calculates the offset for pagination
func (f *ArticleFilter) Offset() int {
	return (f.Page - 1) * f.PageSize
}
