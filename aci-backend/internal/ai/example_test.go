package ai_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/phillipboles/aci-backend/internal/ai"
	"github.com/phillipboles/aci-backend/internal/domain"
)

// TestEnrichmentExample demonstrates how to use the AI enrichment system
// This is an example test that requires ANTHROPIC_API_KEY to be set
func TestEnrichmentExample(t *testing.T) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		t.Skip("ANTHROPIC_API_KEY not set, skipping integration test")
	}

	// Create AI client
	client, err := ai.NewClient(ai.Config{
		APIKey: apiKey,
		Model:  "claude-3-haiku-20240307",
	})
	if err != nil {
		t.Fatalf("failed to create AI client: %v", err)
	}

	// Create enricher
	enricher := ai.NewEnricher(client)

	// Example article
	article := &domain.Article{
		Title: "Critical Ransomware Attack Targets Healthcare Sector",
		Content: `A new ransomware campaign is targeting healthcare organizations worldwide.
The malware, dubbed "MediCrypt", uses phishing emails with malicious attachments to gain initial access.
Once inside the network, it spreads laterally and encrypts critical patient data.
Security researchers have identified the command-and-control server at malicious-domain.com.
The malware uses file hash a1b2c3d4e5f6 for the main payload.
Organizations are advised to implement strict email filtering and enable multi-factor authentication.`,
		CVEs:    []string{"CVE-2024-1234"},
		Vendors: []string{"Microsoft Exchange", "VMware"},
	}

	// Enrich the article
	ctx := context.Background()
	result, err := enricher.EnrichArticle(ctx, article)
	if err != nil {
		t.Fatalf("failed to enrich article: %v", err)
	}

	// Display results
	fmt.Printf("\n=== Enrichment Results ===\n")
	fmt.Printf("Threat Type: %s\n", result.ThreatType)
	fmt.Printf("Attack Vector: %s\n", result.AttackVector)
	fmt.Printf("Impact: %s\n", result.ImpactAssessment)
	fmt.Printf("Confidence: %.2f\n", result.ConfidenceScore)
	fmt.Printf("\nRecommended Actions:\n")
	for i, action := range result.RecommendedActions {
		fmt.Printf("  %d. %s\n", i+1, action)
	}
	fmt.Printf("\nIOCs Found: %d\n", len(result.IOCs))
	for _, ioc := range result.IOCs {
		fmt.Printf("  - %s: %s (%s)\n", ioc.Type, ioc.Value, ioc.Context)
	}

	// Validate results
	if result.ThreatType == "" {
		t.Error("threat_type should not be empty")
	}

	if result.ConfidenceScore < 0 || result.ConfidenceScore > 1 {
		t.Errorf("confidence_score out of range: %f", result.ConfidenceScore)
	}

	// Generate Armor CTA
	cta, err := enricher.GenerateArmorCTA(ctx, article)
	if err != nil {
		t.Logf("Warning: failed to generate CTA: %v", err)
	} else {
		fmt.Printf("\n=== Armor CTA ===\n")
		fmt.Printf("Type: %s\n", cta.Type)
		fmt.Printf("Title: %s\n", cta.Title)
		fmt.Printf("URL: %s\n", cta.URL)
	}
}

// Example usage without running a test
func ExampleEnricher_EnrichArticle() {
	client, _ := ai.NewClient(ai.Config{
		APIKey: "sk-ant-...",
	})

	enricher := ai.NewEnricher(client)

	article := &domain.Article{
		Title:   "Critical Vulnerability in Popular Software",
		Content: "A critical vulnerability has been discovered...",
		CVEs:    []string{"CVE-2024-1234"},
	}

	ctx := context.Background()
	result, err := enricher.EnrichArticle(ctx, article)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Threat Type: %s\n", result.ThreatType)
	fmt.Printf("Confidence: %.2f\n", result.ConfidenceScore)
}
