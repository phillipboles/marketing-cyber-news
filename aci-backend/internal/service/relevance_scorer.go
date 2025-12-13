package service

import (
	"strings"

	"github.com/phillipboles/aci-backend/internal/domain"
)

// RelevanceScorer calculates Armor.com relevance for articles
type RelevanceScorer struct {
	productKeywords  []string
	industryKeywords []string
}

// NewRelevanceScorer creates a new relevance scorer with default keywords
func NewRelevanceScorer() *RelevanceScorer {
	return &RelevanceScorer{
		productKeywords: []string{
			// Armor products and services
			"managed security",
			"security operations center",
			"soc",
			"threat detection",
			"threat response",
			"incident response",
			"security monitoring",
			"cloud security",
			"compliance",
			"pci dss",
			"pci compliance",
			"hipaa",
			"gdpr",
			"vulnerability management",
			"penetration testing",
			"security assessment",
			"managed cloud",
			"cloud hosting",
			"dedicated hosting",
			"hybrid cloud",
			"disaster recovery",
			"business continuity",
			"backup",
			"security automation",
			"threat intelligence",
			"log management",
			"siem",
		},
		industryKeywords: []string{
			// Target industries
			"healthcare",
			"financial services",
			"fintech",
			"banking",
			"e-commerce",
			"retail",
			"payment processing",
			"credit card",
			"payment card industry",
			"online payments",
			"saas",
			"software as a service",
			"enterprise",
			"small business",
			"medium business",
			"smb",
		},
	}
}

// Score calculates the Armor relevance score for an article (0-1)
func (s *RelevanceScorer) Score(article *domain.Article) float64 {
	if article == nil {
		return 0.0
	}

	combinedText := strings.ToLower(article.Title + " " + article.Content)
	if article.Summary != nil {
		combinedText += " " + strings.ToLower(*article.Summary)
	}

	// Calculate product keyword matches
	productMatches := 0
	for _, keyword := range s.productKeywords {
		if strings.Contains(combinedText, keyword) {
			productMatches++
		}
	}

	// Calculate industry keyword matches
	industryMatches := 0
	for _, keyword := range s.industryKeywords {
		if strings.Contains(combinedText, keyword) {
			industryMatches++
		}
	}

	// Weighted scoring: product keywords are more important
	productWeight := 0.7
	industryWeight := 0.3

	productScore := float64(productMatches) / float64(len(s.productKeywords))
	industryScore := float64(industryMatches) / float64(len(s.industryKeywords))

	// Combine scores with weights
	score := (productScore * productWeight) + (industryScore * industryWeight)

	// Normalize to 0-1 range
	if score > 1.0 {
		score = 1.0
	}

	// Boost score for high severity articles (critical/high)
	if article.Severity == domain.SeverityCritical {
		score *= 1.2
	} else if article.Severity == domain.SeverityHigh {
		score *= 1.1
	}

	// Cap at 1.0
	if score > 1.0 {
		score = 1.0
	}

	return score
}

// GenerateCTA generates a call-to-action if relevance is high enough
func (s *RelevanceScorer) GenerateCTA(article *domain.Article) *domain.ArmorCTA {
	if article == nil {
		return nil
	}

	// Only generate CTA if relevance is > 0.5
	if article.ArmorRelevance <= 0.5 {
		return nil
	}

	combinedText := strings.ToLower(article.Title + " " + article.Content)

	// Determine CTA type based on content
	if s.containsAny(combinedText, []string{"managed security", "soc", "threat detection", "incident response"}) {
		return &domain.ArmorCTA{
			Type:  "service",
			Title: "Protect Your Business with Armor Managed Security",
			URL:   "https://www.armor.com/services/managed-security",
		}
	}

	if s.containsAny(combinedText, []string{"cloud security", "cloud hosting", "aws", "azure", "gcp"}) {
		return &domain.ArmorCTA{
			Type:  "service",
			Title: "Secure Your Cloud Infrastructure with Armor",
			URL:   "https://www.armor.com/services/cloud-security",
		}
	}

	if s.containsAny(combinedText, []string{"compliance", "pci", "hipaa", "gdpr"}) {
		return &domain.ArmorCTA{
			Type:  "service",
			Title: "Achieve Compliance with Armor's Expert Guidance",
			URL:   "https://www.armor.com/services/compliance",
		}
	}

	if s.containsAny(combinedText, []string{"vulnerability", "penetration test", "security assessment"}) {
		return &domain.ArmorCTA{
			Type:  "service",
			Title: "Schedule a Security Assessment with Armor",
			URL:   "https://www.armor.com/services/security-assessment",
		}
	}

	if article.Severity == domain.SeverityCritical || article.Severity == domain.SeverityHigh {
		return &domain.ArmorCTA{
			Type:  "consultation",
			Title: "Speak with an Armor Security Expert",
			URL:   "https://www.armor.com/contact",
		}
	}

	// Default CTA
	return &domain.ArmorCTA{
		Type:  "product",
		Title: "Learn How Armor Can Protect Your Business",
		URL:   "https://www.armor.com/solutions",
	}
}

// containsAny checks if the text contains any of the keywords
func (s *RelevanceScorer) containsAny(text string, keywords []string) bool {
	for _, keyword := range keywords {
		if strings.Contains(text, keyword) {
			return true
		}
	}
	return false
}

// AddProductKeyword adds a product keyword
func (s *RelevanceScorer) AddProductKeyword(keyword string) {
	if keyword == "" {
		return
	}
	s.productKeywords = append(s.productKeywords, strings.ToLower(keyword))
}

// AddIndustryKeyword adds an industry keyword
func (s *RelevanceScorer) AddIndustryKeyword(keyword string) {
	if keyword == "" {
		return
	}
	s.industryKeywords = append(s.industryKeywords, strings.ToLower(keyword))
}
