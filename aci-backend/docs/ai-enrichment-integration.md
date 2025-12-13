# AI Content Enrichment Integration Guide

## Overview

The AI Content Enrichment system uses Claude AI (Haiku model) to automatically analyze cybersecurity articles and generate:
- Threat intelligence analysis
- Indicators of Compromise (IOCs)
- Recommended security actions
- Armor.com marketing CTAs

## Architecture

```
┌─────────────────┐
│   Article       │
│   Created       │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  Enrichment     │
│  Service        │
└────────┬────────┘
         │
         ▼
┌─────────────────┐      ┌─────────────────┐
│   AI Enricher   │─────▶│  Claude API     │
│                 │      │  (Haiku)        │
└────────┬────────┘      └─────────────────┘
         │
         ▼
┌─────────────────┐
│  Article        │
│  Updated with   │
│  AI Analysis    │
└─────────────────┘
```

## Components

### 1. AI Client (`internal/ai/client.go`)

Wrapper around the Anthropic Claude SDK.

**Key Functions:**
- `NewClient(cfg Config) (*Client, error)` - Creates client with API key
- `Complete(ctx, systemPrompt, userMessage) (string, error)` - Text completion
- `CompleteWithJSON(ctx, systemPrompt, userMessage, result) error` - JSON response

**Configuration:**
```go
client, err := ai.NewClient(ai.Config{
    APIKey: os.Getenv("ANTHROPIC_API_KEY"),
    Model:  "claude-3-haiku-20240307", // Default
})
```

### 2. Prompt Templates (`internal/ai/prompts.go`)

Pre-configured prompts for threat analysis and CTA generation.

**System Prompts:**
- `ThreatAnalysisSystemPrompt` - Defines analyst role and JSON response format
- `ArmorCTASystemPrompt` - Defines marketing specialist role

**Builders:**
- `BuildThreatAnalysisPrompt(title, content, cves, vendors)` - Creates analysis prompt
- `BuildArmorCTAPrompt(title, content, threatType, attackVector)` - Creates CTA prompt

### 3. Enricher (`internal/ai/enrichment.go`)

Core enrichment logic.

**Types:**
```go
type EnrichmentResult struct {
    ThreatType         string
    AttackVector       string
    ImpactAssessment   string
    RecommendedActions []string
    IOCs               []IOC
    ConfidenceScore    float64
}

type IOC struct {
    Type    string // ip, domain, hash, url
    Value   string
    Context string
}
```

**Key Functions:**
- `EnrichArticle(ctx, article) (*EnrichmentResult, error)` - Analyzes article
- `GenerateArmorCTA(ctx, article) (*domain.ArmorCTA, error)` - Generates CTA

**Features:**
- 60-second timeout for threat analysis
- 30-second timeout for CTA generation
- Automatic result validation
- Error wrapping with context

### 4. Enrichment Service (`internal/service/enrichment_service.go`)

High-level service for batch enrichment.

**Key Functions:**
- `EnrichArticle(ctx, articleID) error` - Enriches single article
- `EnrichPendingArticles(ctx, limit) (int, error)` - Batch enrichment

**Features:**
- Skips already-enriched articles
- Calculates Armor relevance score
- 100ms delay between API calls (rate limiting)
- Graceful failure handling (continues on error)
- Atomic DB updates

## Integration Steps

### 1. Environment Configuration

Add to `.env`:
```bash
ANTHROPIC_API_KEY=sk-ant-...
ANTHROPIC_MODEL=claude-3-haiku-20240307  # Optional, defaults to Haiku
```

### 2. Initialize Services

```go
// In main.go or initialization code
aiClient, err := ai.NewClient(ai.Config{
    APIKey: os.Getenv("ANTHROPIC_API_KEY"),
    Model:  os.Getenv("ANTHROPIC_MODEL"),
})
if err != nil {
    log.Fatalf("failed to create AI client: %v", err)
}

enricher := ai.NewEnricher(aiClient)
enrichmentService := service.NewEnrichmentService(enricher, articleRepo)
```

### 3. Webhook Integration (Optional)

Add enrichment to webhook handler:

```go
// In internal/handler/webhook_handler.go
func (h *WebhookHandler) HandleArticleWebhook(w http.ResponseWriter, r *http.Request) {
    // ... existing article creation code ...

    // Enrich asynchronously (don't block webhook response)
    go func() {
        ctx := context.Background()
        if err := h.enrichmentService.EnrichArticle(ctx, article.ID); err != nil {
            log.Printf("failed to enrich article %s: %v", article.ID, err)
        }
    }()

    // ... return response ...
}
```

### 4. Scheduled Batch Processing

Run periodic enrichment for articles that failed or were skipped:

```go
// Background job (cron, systemd timer, etc.)
func enrichmentJob() {
    ctx := context.Background()
    count, err := enrichmentService.EnrichPendingArticles(ctx, 50)
    if err != nil {
        log.Printf("enrichment job failed: %v", err)
        return
    }
    log.Printf("enriched %d articles", count)
}
```

## API Response Format

### Enriched Article Fields

```json
{
  "id": "uuid",
  "title": "Article Title",
  "content": "Article content...",

  // AI-enriched fields
  "threat_type": "ransomware",
  "attack_vector": "phishing email",
  "impact_assessment": "High risk of data encryption...",
  "recommended_actions": [
    "Block sender domains immediately",
    "Enable MFA on all accounts",
    "Review email gateway rules"
  ],
  "iocs": [
    {
      "type": "domain",
      "value": "malicious-site.com",
      "context": "C2 server identified in campaign"
    },
    {
      "type": "hash",
      "value": "a1b2c3d4e5f6...",
      "context": "Malware payload SHA-256"
    }
  ],

  // Marketing
  "armor_relevance": 0.85,
  "armor_cta": {
    "type": "service",
    "title": "Get 24/7 Ransomware Protection",
    "url": "https://armor.com/managed-detection-response"
  },

  // Metadata
  "enriched_at": "2024-01-15T10:30:00Z"
}
```

## Performance Characteristics

### Cost Efficiency
- **Model**: Claude 3 Haiku (lowest cost)
- **Average tokens per analysis**: ~2,000 input, ~500 output
- **Estimated cost**: $0.0003 per article

### Latency
- **Threat analysis**: 2-5 seconds
- **CTA generation**: 1-2 seconds
- **Total enrichment**: 3-7 seconds per article

### Rate Limiting
- Built-in 100ms delay between requests
- Respects Anthropic API rate limits
- Graceful degradation on failures

## Error Handling

### Enrichment Failures
Articles remain available even if enrichment fails:

```go
// Enrichment errors are logged but don't block article creation
if err := enrichmentService.EnrichArticle(ctx, articleID); err != nil {
    log.Printf("enrichment failed for %s: %v", articleID, err)
    // Article is still published without enrichment
}
```

### Timeout Protection
All API calls have timeouts to prevent hanging:
- Threat analysis: 60 seconds
- CTA generation: 30 seconds

### Validation
All AI responses are validated before saving:
- Required fields checked
- IOC types validated
- Confidence scores clamped to [0, 1]
- Invalid responses rejected with error

## Monitoring

### Key Metrics to Track
1. **Enrichment success rate**: Articles enriched / total articles
2. **API latency**: Time per Claude API call
3. **Error rate**: Failed enrichments / total attempts
4. **Cost**: Total API costs per day/week/month

### Logging
All enrichment operations log:
```
INFO: successfully enriched article {id} (threat_type={type}, confidence={score})
ERROR: failed to enrich article {id}: {error}
WARN: failed to generate armor cta for article {id}: {error}
```

## Testing

### Unit Tests
Test AI components with mocked responses:

```go
// Mock client for testing
type MockAIClient struct {
    Response string
    Error    error
}

func (m *MockAIClient) Complete(ctx, system, user) (string, error) {
    if m.Error != nil {
        return "", m.Error
    }
    return m.Response, nil
}
```

### Integration Tests
Test with real API (use test API key):

```bash
export ANTHROPIC_API_KEY=sk-ant-test-...
go test ./internal/ai/... -integration
```

## Security Considerations

1. **API Key Protection**: Never commit API keys to version control
2. **Input Sanitization**: Article content is not sanitized (Claude handles safely)
3. **Output Validation**: All AI responses validated before DB storage
4. **Rate Limiting**: Built-in delays prevent API abuse
5. **Timeout Protection**: Prevents resource exhaustion from slow API calls

## Future Enhancements

1. **Caching**: Cache enrichment results for similar articles
2. **Batch API**: Use Anthropic's batch API for cost savings
3. **Model Selection**: Auto-select model based on article complexity
4. **Confidence Thresholds**: Skip low-confidence enrichments
5. **Multi-language**: Support non-English articles
6. **Custom Prompts**: Allow per-category custom prompts
7. **A/B Testing**: Test different prompts for better results

## Troubleshooting

### Enrichment Not Working

1. **Check API key**: Verify `ANTHROPIC_API_KEY` is set
2. **Check network**: Ensure outbound HTTPS to api.anthropic.com
3. **Check logs**: Look for error messages
4. **Check article state**: Verify `enriched_at IS NULL`

### Slow Enrichment

1. **Reduce batch size**: Lower limit in `EnrichPendingArticles`
2. **Add concurrency**: Process articles in parallel (add sync.WaitGroup)
3. **Increase timeout**: Adjust context deadlines if needed

### High Costs

1. **Switch to batching**: Use batch API for lower costs
2. **Add caching**: Cache similar enrichments
3. **Filter articles**: Only enrich high-priority articles
4. **Reduce token usage**: Shorten prompts or article content

## Support

For issues or questions:
1. Check logs in `./logs/`
2. Verify API key and connectivity
3. Review Anthropic API status: https://status.anthropic.com
4. Consult Anthropic documentation: https://docs.anthropic.com
