package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/phillipboles/aci-backend/internal/ai"
	"github.com/phillipboles/aci-backend/internal/domain"
	"github.com/phillipboles/aci-backend/internal/repository"
)

// EnrichmentService handles AI enrichment of articles
type EnrichmentService struct {
	enricher    *ai.Enricher
	articleRepo repository.ArticleRepository
}

// NewEnrichmentService creates a new enrichment service instance
func NewEnrichmentService(enricher *ai.Enricher, articleRepo repository.ArticleRepository) *EnrichmentService {
	if enricher == nil {
		panic("enricher cannot be nil")
	}

	if articleRepo == nil {
		panic("articleRepo cannot be nil")
	}

	return &EnrichmentService{
		enricher:    enricher,
		articleRepo: articleRepo,
	}
}

// EnrichArticle enriches an article with AI analysis and saves to DB
func (s *EnrichmentService) EnrichArticle(ctx context.Context, articleID uuid.UUID) error {
	if articleID == uuid.Nil {
		return fmt.Errorf("article id is required")
	}

	// Retrieve the article
	article, err := s.articleRepo.GetByID(ctx, articleID)
	if err != nil {
		return fmt.Errorf("failed to get article: %w", err)
	}

	if article == nil {
		return fmt.Errorf("article not found: %s", articleID)
	}

	// Check if already enriched
	if article.EnrichedAt != nil {
		log.Printf("article %s already enriched at %s, skipping", articleID, article.EnrichedAt.Format(time.RFC3339))
		return nil
	}

	// Perform threat analysis
	enrichmentResult, err := s.enricher.EnrichArticle(ctx, article)
	if err != nil {
		return fmt.Errorf("failed to enrich article: %w", err)
	}

	// Update article with enrichment data
	article.ThreatType = &enrichmentResult.ThreatType
	article.AttackVector = &enrichmentResult.AttackVector
	article.ImpactAssessment = &enrichmentResult.ImpactAssessment
	article.RecommendedActions = enrichmentResult.RecommendedActions

	// Convert AI IOCs to domain IOCs
	article.IOCs = make([]domain.IOC, len(enrichmentResult.IOCs))
	for i, ioc := range enrichmentResult.IOCs {
		article.IOCs[i] = domain.IOC{
			Type:    ioc.Type,
			Value:   ioc.Value,
			Context: ioc.Context,
		}
	}

	// Generate Armor CTA
	armorCTA, err := s.enricher.GenerateArmorCTA(ctx, article)
	if err != nil {
		// Log error but don't fail - CTA is optional
		log.Printf("failed to generate armor cta for article %s: %v", articleID, err)
	} else {
		article.ArmorCTA = armorCTA

		// Calculate armor relevance based on threat type and severity
		article.ArmorRelevance = calculateArmorRelevance(
			enrichmentResult.ThreatType,
			article.Severity,
			enrichmentResult.ConfidenceScore,
		)
	}

	// Set enrichment timestamp
	now := time.Now()
	article.EnrichedAt = &now

	// Save updated article
	if err := s.articleRepo.Update(ctx, article); err != nil {
		return fmt.Errorf("failed to update article: %w", err)
	}

	log.Printf("successfully enriched article %s (threat_type=%s, confidence=%.2f)",
		articleID, enrichmentResult.ThreatType, enrichmentResult.ConfidenceScore)

	return nil
}

// EnrichPendingArticles processes articles that haven't been enriched
func (s *EnrichmentService) EnrichPendingArticles(ctx context.Context, limit int) (int, error) {
	if limit < 1 {
		return 0, fmt.Errorf("limit must be at least 1")
	}

	if limit > 100 {
		return 0, fmt.Errorf("limit cannot exceed 100")
	}

	// Create filter for unenriched articles
	filter := &domain.ArticleFilter{
		Page:     1,
		PageSize: limit,
	}

	articles, _, err := s.articleRepo.List(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to list articles: %w", err)
	}

	enrichedCount := 0
	for _, article := range articles {
		// Skip already enriched articles
		if article.EnrichedAt != nil {
			continue
		}

		// Enrich article
		if err := s.EnrichArticle(ctx, article.ID); err != nil {
			// Log error but continue with other articles
			log.Printf("failed to enrich article %s: %v", article.ID, err)
			continue
		}

		enrichedCount++

		// Add small delay to respect API rate limits
		select {
		case <-ctx.Done():
			return enrichedCount, ctx.Err()
		case <-time.After(100 * time.Millisecond):
			// Continue
		}
	}

	return enrichedCount, nil
}

// calculateArmorRelevance calculates how relevant the article is to Armor's services
func calculateArmorRelevance(threatType string, severity domain.Severity, confidenceScore float64) float64 {
	baseScore := 0.5

	// Increase relevance for high-priority threats
	threatMultiplier := 1.0
	switch threatType {
	case "ransomware", "apt", "data breach", "supply chain":
		threatMultiplier = 1.5
	case "malware", "phishing", "vulnerability":
		threatMultiplier = 1.2
	case "ddos", "social engineering":
		threatMultiplier = 1.0
	default:
		threatMultiplier = 0.8
	}

	// Increase relevance based on severity
	severityMultiplier := 1.0
	switch severity {
	case domain.SeverityCritical:
		severityMultiplier = 1.5
	case domain.SeverityHigh:
		severityMultiplier = 1.3
	case domain.SeverityMedium:
		severityMultiplier = 1.1
	case domain.SeverityLow:
		severityMultiplier = 0.9
	case domain.SeverityInformational:
		severityMultiplier = 0.7
	}

	// Calculate final score
	score := baseScore * threatMultiplier * severityMultiplier * confidenceScore

	// Clamp to [0, 1]
	if score > 1.0 {
		score = 1.0
	}

	if score < 0.0 {
		score = 0.0
	}

	return score
}
