package service

import (
	"strings"
)

// CompetitorFilter detects and scores competitor mentions in articles
type CompetitorFilter struct {
	keywords map[string]float64
}

// NewCompetitorFilter creates a new competitor filter with default keywords
func NewCompetitorFilter() *CompetitorFilter {
	return &CompetitorFilter{
		keywords: map[string]float64{
			// Major competitors (higher weight)
			"crowdstrike":    1.0,
			"palo alto":      1.0,
			"palo alto networks": 1.0,
			"fortinet":       0.9,
			"sentinelone":    0.9,
			"sentinel one":   0.9,
			"mcafee":         0.8,
			"symantec":       0.8,
			"broadcom":       0.7,
			"trend micro":    0.8,
			"sophos":         0.7,
			"kaspersky":      0.7,
			"bitdefender":    0.7,
			"f-secure":       0.6,
			"eset":           0.6,
			"avast":          0.5,
			"avg":            0.5,
			"norton":         0.7,

			// Cloud security competitors
			"cloudflare":     0.8,
			"akamai":         0.7,
			"zscaler":        0.8,
			"netskope":       0.7,
			"proofpoint":     0.7,
			"mimecast":       0.6,

			// EDR/XDR competitors
			"carbon black":   0.8,
			"cylance":        0.7,
			"cybereason":     0.7,
			"tanium":         0.7,
			"rapid7":         0.6,
			"qualys":         0.6,
			"tenable":        0.6,

			// SIEM competitors
			"splunk":         0.9,
			"elastic security": 0.8,
			"logrhythm":      0.6,
			"sumo logic":     0.6,
			"datadog security": 0.7,
		},
	}
}

// Score calculates the competitor score for an article
// Returns score (0-1) and whether the article is competitor-favorable
func (f *CompetitorFilter) Score(title, content string) (score float64, isFavorable bool) {
	if title == "" && content == "" {
		return 0.0, false
	}

	lowerTitle := strings.ToLower(title)
	lowerContent := strings.ToLower(content)
	combinedText := lowerTitle + " " + lowerContent

	var totalWeight float64
	var matchedWeight float64

	// Check for keyword matches
	for keyword, weight := range f.keywords {
		if strings.Contains(combinedText, keyword) {
			matchedWeight += weight
			totalWeight += weight
		}
	}

	// No competitors mentioned
	if totalWeight == 0 {
		return 0.0, false
	}

	// Calculate normalized score (0-1)
	score = matchedWeight / float64(len(f.keywords))
	if score > 1.0 {
		score = 1.0
	}

	// Determine if favorable to competitors
	// Check for positive sentiment indicators
	positiveIndicators := []string{
		"announced",
		"launches",
		"introduces",
		"unveils",
		"releases",
		"partnership",
		"acquired",
		"innovation",
		"breakthrough",
		"leading",
		"award",
		"recognized",
		"success",
		"growth",
		"expansion",
	}

	// Check for negative sentiment indicators
	negativeIndicators := []string{
		"breach",
		"hacked",
		"vulnerability",
		"flaw",
		"exploit",
		"failure",
		"outage",
		"lawsuit",
		"investigation",
		"criticized",
		"backdoor",
		"compromised",
		"bypassed",
	}

	positiveCount := 0
	negativeCount := 0

	for _, indicator := range positiveIndicators {
		if strings.Contains(combinedText, indicator) {
			positiveCount++
		}
	}

	for _, indicator := range negativeIndicators {
		if strings.Contains(combinedText, indicator) {
			negativeCount++
		}
	}

	// Article is competitor-favorable if more positive than negative sentiment
	isFavorable = positiveCount > negativeCount

	return score, isFavorable
}

// AddKeyword adds a new competitor keyword with weight
func (f *CompetitorFilter) AddKeyword(keyword string, weight float64) {
	if keyword == "" {
		return
	}

	if weight < 0.0 {
		weight = 0.0
	}

	if weight > 1.0 {
		weight = 1.0
	}

	f.keywords[strings.ToLower(keyword)] = weight
}

// RemoveKeyword removes a competitor keyword
func (f *CompetitorFilter) RemoveKeyword(keyword string) {
	if keyword == "" {
		return
	}

	delete(f.keywords, strings.ToLower(keyword))
}

// GetKeywords returns all configured keywords
func (f *CompetitorFilter) GetKeywords() map[string]float64 {
	// Return a copy to prevent external modification
	keywords := make(map[string]float64, len(f.keywords))
	for k, v := range f.keywords {
		keywords[k] = v
	}
	return keywords
}
