package ai

import (
	"context"
	"fmt"
	"time"

	"github.com/phillipboles/aci-backend/internal/domain"
)

// EnrichmentResult contains AI-generated analysis
type EnrichmentResult struct {
	ThreatType         string   `json:"threat_type"`
	AttackVector       string   `json:"attack_vector"`
	ImpactAssessment   string   `json:"impact_assessment"`
	RecommendedActions []string `json:"recommended_actions"`
	IOCs               []IOC    `json:"iocs"`
	ConfidenceScore    float64  `json:"confidence_score"`
}

// IOC represents an Indicator of Compromise
type IOC struct {
	Type    string `json:"type"`              // ip, domain, hash, url
	Value   string `json:"value"`             // The actual IOC value
	Context string `json:"context,omitempty"` // Additional context
}

// Validate validates the enrichment result
func (r *EnrichmentResult) Validate() error {
	if r.ThreatType == "" {
		return fmt.Errorf("threat_type is required")
	}

	if r.AttackVector == "" {
		return fmt.Errorf("attack_vector is required")
	}

	if r.ImpactAssessment == "" {
		return fmt.Errorf("impact_assessment is required")
	}

	if len(r.RecommendedActions) == 0 {
		return fmt.Errorf("at least one recommended action is required")
	}

	if r.ConfidenceScore < 0 || r.ConfidenceScore > 1 {
		return fmt.Errorf("confidence_score must be between 0 and 1")
	}

	for i, ioc := range r.IOCs {
		if err := validateIOC(&ioc); err != nil {
			return fmt.Errorf("invalid ioc at index %d: %w", i, err)
		}
	}

	return nil
}

// validateIOC validates an IOC entry
func validateIOC(ioc *IOC) error {
	if ioc.Type == "" {
		return fmt.Errorf("type is required")
	}

	if ioc.Value == "" {
		return fmt.Errorf("value is required")
	}

	validTypes := map[string]bool{
		"ip":     true,
		"domain": true,
		"hash":   true,
		"url":    true,
	}

	if !validTypes[ioc.Type] {
		return fmt.Errorf("invalid type: %s (must be ip, domain, hash, or url)", ioc.Type)
	}

	return nil
}

// Enricher performs AI enrichment on articles
type Enricher struct {
	client *Client
}

// NewEnricher creates a new enricher instance
func NewEnricher(client *Client) *Enricher {
	if client == nil {
		panic("client cannot be nil")
	}

	return &Enricher{
		client: client,
	}
}

// EnrichArticle analyzes an article and returns enrichment data
func (e *Enricher) EnrichArticle(ctx context.Context, article *domain.Article) (*EnrichmentResult, error) {
	if article == nil {
		return nil, fmt.Errorf("article cannot be nil")
	}

	if article.Title == "" {
		return nil, fmt.Errorf("article title is required")
	}

	if article.Content == "" {
		return nil, fmt.Errorf("article content is required")
	}

	// Add timeout to prevent long-running requests
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	// Build the prompt
	userPrompt := BuildThreatAnalysisPrompt(
		article.Title,
		article.Content,
		article.CVEs,
		article.Vendors,
	)

	// Call Claude API
	var result EnrichmentResult
	if err := e.client.CompleteWithJSON(ctx, ThreatAnalysisSystemPrompt, userPrompt, &result); err != nil {
		return nil, fmt.Errorf("failed to analyze article: %w", err)
	}

	// Validate the result
	if err := result.Validate(); err != nil {
		return nil, fmt.Errorf("invalid enrichment result: %w", err)
	}

	return &result, nil
}

// GenerateArmorCTA generates Armor.com call-to-action based on content
func (e *Enricher) GenerateArmorCTA(ctx context.Context, article *domain.Article) (*domain.ArmorCTA, error) {
	if article == nil {
		return nil, fmt.Errorf("article cannot be nil")
	}

	if article.Title == "" {
		return nil, fmt.Errorf("article title is required")
	}

	if article.Content == "" {
		return nil, fmt.Errorf("article content is required")
	}

	// Add timeout to prevent long-running requests
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Get threat context if available
	threatType := ""
	if article.ThreatType != nil {
		threatType = *article.ThreatType
	}

	attackVector := ""
	if article.AttackVector != nil {
		attackVector = *article.AttackVector
	}

	// Build the prompt
	userPrompt := BuildArmorCTAPrompt(
		article.Title,
		article.Content,
		threatType,
		attackVector,
	)

	// Call Claude API
	var cta domain.ArmorCTA
	if err := e.client.CompleteWithJSON(ctx, ArmorCTASystemPrompt, userPrompt, &cta); err != nil {
		return nil, fmt.Errorf("failed to generate armor cta: %w", err)
	}

	// Validate the result
	if !cta.IsValid() {
		return nil, fmt.Errorf("invalid armor cta generated")
	}

	return &cta, nil
}
