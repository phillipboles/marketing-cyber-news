package handlers

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/phillipboles/aci-backend/internal/api/response"
	"github.com/phillipboles/aci-backend/internal/domain"
	"github.com/phillipboles/aci-backend/internal/repository"
	"github.com/phillipboles/aci-backend/internal/service"
)

// WebhookHandler handles n8n webhook events
type WebhookHandler struct {
	articleService *service.ArticleService
	webhookLogRepo repository.WebhookLogRepository
	webhookSecret  string
}

// WebhookPayload represents the incoming webhook payload from n8n
type WebhookPayload struct {
	EventType string           `json:"event_type"`
	Data      json.RawMessage  `json:"data"`
	Metadata  *WebhookMetadata `json:"metadata,omitempty"`
}

// WebhookMetadata contains metadata about the webhook event
type WebhookMetadata struct {
	WorkflowID  string `json:"workflow_id,omitempty"`
	ExecutionID string `json:"execution_id,omitempty"`
	Timestamp   string `json:"timestamp,omitempty"`
}

// ArticleCreatedData represents article.created event data
type ArticleCreatedData struct {
	Title          string   `json:"title"`
	Content        string   `json:"content"`
	Summary        string   `json:"summary,omitempty"`
	CategorySlug   string   `json:"category_slug"`
	Severity       string   `json:"severity,omitempty"`
	Tags           []string `json:"tags,omitempty"`
	SourceURL      string   `json:"source_url"`
	SourceName     string   `json:"source_name,omitempty"`
	PublishedAt    string   `json:"published_at,omitempty"`
	CVEs           []string `json:"cves,omitempty"`
	Vendors        []string `json:"vendors,omitempty"`
	SkipEnrichment bool     `json:"skip_enrichment,omitempty"`
}

// ArticleUpdatedData represents article.updated event data
type ArticleUpdatedData struct {
	ArticleID   string   `json:"article_id"`
	Title       *string  `json:"title,omitempty"`
	Content     *string  `json:"content,omitempty"`
	Summary     *string  `json:"summary,omitempty"`
	Severity    *string  `json:"severity,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	CVEs        []string `json:"cves,omitempty"`
	Vendors     []string `json:"vendors,omitempty"`
	IsPublished *bool    `json:"is_published,omitempty"`
}

// ArticleDeletedData represents article.deleted event data
type ArticleDeletedData struct {
	ArticleID string `json:"article_id"`
}

// BulkImportData represents bulk.import event data
type BulkImportData struct {
	Articles []ArticleCreatedData `json:"articles"`
}

// EnrichmentCompleteData represents enrichment.complete event data
type EnrichmentCompleteData struct {
	ArticleID          string   `json:"article_id"`
	ThreatType         *string  `json:"threat_type,omitempty"`
	AttackVector       *string  `json:"attack_vector,omitempty"`
	ImpactAssessment   *string  `json:"impact_assessment,omitempty"`
	RecommendedActions []string `json:"recommended_actions,omitempty"`
	IOCs               []IOC    `json:"iocs,omitempty"`
}

// IOC represents an Indicator of Compromise
type IOC struct {
	Type    string `json:"type"`
	Value   string `json:"value"`
	Context string `json:"context,omitempty"`
}

// NewWebhookHandler creates a new webhook handler
func NewWebhookHandler(
	articleService *service.ArticleService,
	webhookLogRepo repository.WebhookLogRepository,
	webhookSecret string,
) *WebhookHandler {
	return &WebhookHandler{
		articleService: articleService,
		webhookLogRepo: webhookLogRepo,
		webhookSecret:  webhookSecret,
	}
}

// HandleN8nWebhook handles POST /v1/webhooks/n8n
func (h *WebhookHandler) HandleN8nWebhook(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Read body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		response.BadRequest(w, "failed to read request body")
		return
	}
	defer r.Body.Close()

	// Verify HMAC signature
	signature := r.Header.Get("X-N8N-Signature")
	if !h.verifySignature(body, signature) {
		response.Unauthorized(w, "invalid signature")
		return
	}

	// Parse payload
	var payload WebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		response.BadRequest(w, "invalid JSON payload")
		return
	}

	if payload.EventType == "" {
		response.BadRequest(w, "event_type is required")
		return
	}

	// Create webhook log entry
	var workflowID, executionID *string
	if payload.Metadata != nil {
		if payload.Metadata.WorkflowID != "" {
			workflowID = &payload.Metadata.WorkflowID
		}
		if payload.Metadata.ExecutionID != "" {
			executionID = &payload.Metadata.ExecutionID
		}
	}

	webhookLog := domain.NewWebhookLog(
		payload.EventType,
		string(body),
		workflowID,
		executionID,
	)

	if err := h.webhookLogRepo.Create(ctx, webhookLog); err != nil {
		// Log error but don't fail the request
		fmt.Printf("Failed to create webhook log: %v\n", err)
	}

	// Mark as processing
	webhookLog.MarkProcessing()
	_ = h.webhookLogRepo.Update(ctx, webhookLog)

	// Route by event type
	var result interface{}
	var handlerErr error

	switch payload.EventType {
	case "article.created":
		result, handlerErr = h.handleArticleCreated(ctx, payload.Data)
	case "article.updated":
		result, handlerErr = h.handleArticleUpdated(ctx, payload.Data)
	case "article.deleted":
		result, handlerErr = h.handleArticleDeleted(ctx, payload.Data)
	case "bulk.import":
		result, handlerErr = h.handleBulkImport(ctx, payload.Data)
	case "enrichment.complete":
		result, handlerErr = h.handleEnrichmentComplete(ctx, payload.Data)
	default:
		webhookLog.MarkFailed(fmt.Sprintf("unsupported event type: %s", payload.EventType))
		_ = h.webhookLogRepo.Update(ctx, webhookLog)
		response.BadRequest(w, "unsupported event type")
		return
	}

	// Handle errors
	if handlerErr != nil {
		webhookLog.MarkFailed(handlerErr.Error())
		_ = h.webhookLogRepo.Update(ctx, webhookLog)
		response.InternalError(w, handlerErr.Error(), "")
		return
	}

	// Mark as success
	webhookLog.MarkSuccess()
	_ = h.webhookLogRepo.Update(ctx, webhookLog)

	// Return 202 Accepted with job_id
	response.JSON(w, http.StatusAccepted, map[string]interface{}{
		"job_id": webhookLog.ID.String(),
		"status": "accepted",
		"result": result,
	})
}

// handleArticleCreated handles article.created events
func (h *WebhookHandler) handleArticleCreated(ctx context.Context, data json.RawMessage) (interface{}, error) {
	var articleData ArticleCreatedData
	if err := json.Unmarshal(data, &articleData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal article data: %w", err)
	}

	// Convert to service data
	serviceData := service.ArticleCreatedData{
		Title:          articleData.Title,
		Content:        articleData.Content,
		Summary:        articleData.Summary,
		CategorySlug:   articleData.CategorySlug,
		Severity:       articleData.Severity,
		Tags:           articleData.Tags,
		SourceURL:      articleData.SourceURL,
		SourceName:     articleData.SourceName,
		PublishedAt:    articleData.PublishedAt,
		CVEs:           articleData.CVEs,
		Vendors:        articleData.Vendors,
		SkipEnrichment: articleData.SkipEnrichment,
	}

	article, err := h.articleService.CreateArticle(ctx, serviceData)
	if err != nil {
		return nil, fmt.Errorf("failed to create article: %w", err)
	}

	return map[string]interface{}{
		"article_id": article.ID.String(),
		"slug":       article.Slug,
	}, nil
}

// handleArticleUpdated handles article.updated events
func (h *WebhookHandler) handleArticleUpdated(ctx context.Context, data json.RawMessage) (interface{}, error) {
	var updateData ArticleUpdatedData
	if err := json.Unmarshal(data, &updateData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal update data: %w", err)
	}

	articleID, err := uuid.Parse(updateData.ArticleID)
	if err != nil {
		return nil, fmt.Errorf("invalid article ID: %w", err)
	}

	// Convert to service data
	serviceData := service.ArticleUpdatedData{
		Title:       updateData.Title,
		Content:     updateData.Content,
		Summary:     updateData.Summary,
		Severity:    updateData.Severity,
		Tags:        updateData.Tags,
		CVEs:        updateData.CVEs,
		Vendors:     updateData.Vendors,
		IsPublished: updateData.IsPublished,
	}

	article, err := h.articleService.UpdateArticle(ctx, articleID, serviceData)
	if err != nil {
		return nil, fmt.Errorf("failed to update article: %w", err)
	}

	return map[string]interface{}{
		"article_id": article.ID.String(),
		"updated_at": article.UpdatedAt,
	}, nil
}

// handleArticleDeleted handles article.deleted events
func (h *WebhookHandler) handleArticleDeleted(ctx context.Context, data json.RawMessage) (interface{}, error) {
	var deleteData ArticleDeletedData
	if err := json.Unmarshal(data, &deleteData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal delete data: %w", err)
	}

	articleID, err := uuid.Parse(deleteData.ArticleID)
	if err != nil {
		return nil, fmt.Errorf("invalid article ID: %w", err)
	}

	if err := h.articleService.DeleteArticle(ctx, articleID); err != nil {
		return nil, fmt.Errorf("failed to delete article: %w", err)
	}

	return map[string]interface{}{
		"article_id": deleteData.ArticleID,
		"deleted":    true,
	}, nil
}

// handleBulkImport handles bulk.import events
func (h *WebhookHandler) handleBulkImport(ctx context.Context, data json.RawMessage) (interface{}, error) {
	var bulkData BulkImportData
	if err := json.Unmarshal(data, &bulkData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal bulk data: %w", err)
	}

	// Convert to service data
	serviceArticles := make([]service.ArticleCreatedData, len(bulkData.Articles))
	for i, article := range bulkData.Articles {
		serviceArticles[i] = service.ArticleCreatedData{
			Title:          article.Title,
			Content:        article.Content,
			Summary:        article.Summary,
			CategorySlug:   article.CategorySlug,
			Severity:       article.Severity,
			Tags:           article.Tags,
			SourceURL:      article.SourceURL,
			SourceName:     article.SourceName,
			PublishedAt:    article.PublishedAt,
			CVEs:           article.CVEs,
			Vendors:        article.Vendors,
			SkipEnrichment: article.SkipEnrichment,
		}
	}

	successCount, errors := h.articleService.BulkImport(ctx, serviceArticles)

	errorMessages := make([]string, len(errors))
	for i, err := range errors {
		errorMessages[i] = err.Error()
	}

	return map[string]interface{}{
		"total":       len(bulkData.Articles),
		"success":     successCount,
		"failed":      len(errors),
		"errors":      errorMessages,
	}, nil
}

// handleEnrichmentComplete handles enrichment.complete events
func (h *WebhookHandler) handleEnrichmentComplete(ctx context.Context, data json.RawMessage) (interface{}, error) {
	var enrichmentData EnrichmentCompleteData
	if err := json.Unmarshal(data, &enrichmentData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal enrichment data: %w", err)
	}

	// TODO: Implement enrichment handling
	// This would update an article with AI-generated enrichment data

	return map[string]interface{}{
		"article_id": enrichmentData.ArticleID,
		"enriched":   true,
	}, nil
}

// verifySignature verifies the HMAC-SHA256 signature
func (h *WebhookHandler) verifySignature(payload []byte, signature string) bool {
	if signature == "" {
		return false
	}

	// Parse "sha256=<hex>" format
	parts := strings.SplitN(signature, "=", 2)
	if len(parts) != 2 {
		return false
	}

	algorithm := parts[0]
	receivedHex := parts[1]

	if algorithm != "sha256" {
		return false
	}

	// Compute HMAC-SHA256
	mac := hmac.New(sha256.New, []byte(h.webhookSecret))
	mac.Write(payload)
	expectedMAC := mac.Sum(nil)
	expectedHex := hex.EncodeToString(expectedMAC)

	// Compare using constant-time comparison
	return hmac.Equal([]byte(expectedHex), []byte(receivedHex))
}
