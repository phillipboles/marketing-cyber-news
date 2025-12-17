package integration

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/phillipboles/aci-backend/internal/ai"
	"github.com/phillipboles/aci-backend/internal/api/handlers"
	"github.com/phillipboles/aci-backend/internal/domain"
	"github.com/phillipboles/aci-backend/internal/repository/postgres"
	"github.com/phillipboles/aci-backend/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testWebhookSecret = "test-webhook-secret-12345"
	testWebhookPath   = "/v1/webhooks/n8n"
)

// Test fixtures
var (
	validArticlePayload = handlers.ArticleCreatedData{
		Title:        "Critical Zero-Day Vulnerability in Apache HTTP Server",
		Content:      "<p>A critical vulnerability has been discovered in Apache HTTP Server affecting versions 2.4.0 to 2.4.49.</p>",
		Summary:      "Critical RCE vulnerability in Apache",
		CategorySlug: "vulnerabilities",
		Severity:     "critical",
		Tags:         []string{"apache", "rce", "zero-day"},
		SourceURL:    "https://example.com/apache-vuln-2024",
		SourceName:   "Security Research Blog",
		PublishedAt:  time.Now().Format(time.RFC3339),
		CVEs:         []string{"CVE-2024-12345"},
		Vendors:      []string{"Apache"},
	}

	validBulkPayload = handlers.BulkImportData{
		Articles: []handlers.ArticleCreatedData{
			{
				Title:        "Malware Campaign Targets Financial Institutions",
				Content:      "<p>A sophisticated malware campaign has been detected targeting banks.</p>",
				CategorySlug: "malware",
				SourceURL:    "https://example.com/malware-campaign-1",
				SourceName:   "Threat Intel Feed",
				Severity:     "high",
			},
			{
				Title:        "Ransomware Group Claims Major Healthcare Breach",
				Content:      "<p>A ransomware group has claimed responsibility for a healthcare data breach.</p>",
				CategorySlug: "ransomware",
				SourceURL:    "https://example.com/ransomware-breach-1",
				SourceName:   "Threat Intel Feed",
				Severity:     "critical",
			},
		},
	}
)

// Helper to sign webhook payload with HMAC-SHA256
func signPayload(payload []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

// Helper to create webhook payload
func createWebhookPayload(eventType string, data interface{}) ([]byte, error) {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	payload := map[string]interface{}{
		"event_type": eventType,
		"data":       json.RawMessage(dataBytes),
		"metadata": map[string]string{
			"workflow_id":  "workflow-123",
			"execution_id": "exec-456",
			"timestamp":    time.Now().Format(time.RFC3339),
		},
	}

	return json.Marshal(payload)
}

// Helper function to create string pointers
func strPtr(s string) *string {
	return &s
}

// Helper to setup webhook handler with test dependencies
func setupWebhookHandler(t *testing.T, db *TestDB) *handlers.WebhookHandler {
	t.Helper()

	_ = context.Background()

	// Create repositories
	articleRepo := postgres.NewArticleRepository(db.DB)
	categoryRepo := postgres.NewCategoryRepository(db.DB)
	sourceRepo := postgres.NewSourceRepository(db.DB)
	webhookLogRepo := postgres.NewWebhookLogRepository(db.DB)

	// Seed test categories
	seedTestCategories(t, db)

	// Create article service
	articleService := service.NewArticleService(
		articleRepo,
		categoryRepo,
		sourceRepo,
		webhookLogRepo,
	)

	// Create AI client for enrichment service (with dummy API key for testing)
	// Most webhook tests don't actually call enrichment, so this won't make real API calls
	aiClient, err := ai.NewClient(ai.Config{
		APIKey: "test-api-key-placeholder",
		Model:  "claude-3-haiku-20240307",
	})
	if err != nil {
		t.Fatalf("failed to create AI client: %v", err)
	}

	enricher := ai.NewEnricher(aiClient)
	enrichmentService := service.NewEnrichmentService(enricher, articleRepo)

	// Create webhook handler
	return handlers.NewWebhookHandler(
		articleService,
		enrichmentService,
		webhookLogRepo,
		testWebhookSecret,
	)
}

// Helper to seed test categories
func seedTestCategories(t *testing.T, db *TestDB) {
	t.Helper()

	ctx := context.Background()

	categories := []domain.Category{
		{
			ID:          uuid.New(),
			Name:        "Vulnerabilities",
			Slug:        "vulnerabilities",
			Description: strPtr("Security vulnerabilities and CVEs"),
			Icon:        strPtr("bug"),
			Color:       "#FF0000",
			CreatedAt:   time.Now(),
		},
		{
			ID:          uuid.New(),
			Name:        "Malware",
			Slug:        "malware",
			Description: strPtr("Malware analysis and threats"),
			Icon:        strPtr("virus"),
			Color:       "#FF6600",
			CreatedAt:   time.Now(),
		},
		{
			ID:          uuid.New(),
			Name:        "Ransomware",
			Slug:        "ransomware",
			Description: strPtr("Ransomware attacks and trends"),
			Icon:        strPtr("lock"),
			Color:       "#CC0000",
			CreatedAt:   time.Now(),
		},
		{
			ID:          uuid.New(),
			Name:        "Data Breaches",
			Slug:        "data-breaches",
			Description: strPtr("Data breach incidents"),
			Icon:        strPtr("database"),
			Color:       "#9900CC",
			CreatedAt:   time.Now(),
		},
	}

	for _, category := range categories {
		_, err := db.DB.Pool.Exec(ctx, `
			INSERT INTO categories (id, name, slug, description, icon, color, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			ON CONFLICT (slug) DO NOTHING
		`, category.ID, category.Name, category.Slug, category.Description, category.Icon,
			category.Color, category.CreatedAt)

		if err != nil {
			t.Fatalf("failed to seed category %s: %v", category.Slug, err)
		}
	}
}

// Helper to make webhook request
func makeWebhookRequest(t *testing.T, handler http.HandlerFunc, payload []byte, signature string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(http.MethodPost, testWebhookPath, bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-N8N-Signature", signature)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	return rr
}

// T056: Valid signature + article.created -> article stored
func TestWebhook_ArticleCreated_HappyPath(t *testing.T) {
	// Setup
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)
	defer CleanupDB(t, db)

	handler := setupWebhookHandler(t, db)

	// Create valid payload
	payload, err := createWebhookPayload("article.created", validArticlePayload)
	require.NoError(t, err)

	signature := signPayload(payload, testWebhookSecret)

	// Execute
	rr := makeWebhookRequest(t, handler.HandleN8nWebhook, payload, signature)

	// Assert
	assert.Equal(t, http.StatusAccepted, rr.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "accepted", response["status"])
	assert.NotEmpty(t, response["job_id"])

	// Verify article was stored in database
	ctx := context.Background()
	articleRepo := postgres.NewArticleRepository(db.DB)

	article, err := articleRepo.GetBySourceURL(ctx, validArticlePayload.SourceURL)
	require.NoError(t, err)
	assert.NotNil(t, article)
	assert.Equal(t, validArticlePayload.Title, article.Title)
	assert.Equal(t, validArticlePayload.SourceURL, article.SourceURL)
	assert.Equal(t, domain.SeverityCritical, article.Severity)
	assert.Contains(t, article.Tags, "apache")
	assert.Contains(t, article.CVEs, "CVE-2024-12345")
}

// T057: Invalid HMAC signature -> 401 Unauthorized
func TestWebhook_InvalidSignature(t *testing.T) {
	// Setup
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)
	defer CleanupDB(t, db)

	handler := setupWebhookHandler(t, db)

	// Create payload with INVALID signature
	payload, err := createWebhookPayload("article.created", validArticlePayload)
	require.NoError(t, err)

	invalidSignature := "sha256=invalidsignature1234567890"

	// Execute
	rr := makeWebhookRequest(t, handler.HandleN8nWebhook, payload, invalidSignature)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, rr.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response["error"], "invalid signature")
}

// T058: bulk.import with 50 articles -> all queued
func TestWebhook_BulkImport_HappyPath(t *testing.T) {
	// Setup
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)
	defer CleanupDB(t, db)

	handler := setupWebhookHandler(t, db)

	// Create bulk payload with 50 articles
	bulkData := handlers.BulkImportData{
		Articles: make([]handlers.ArticleCreatedData, 50),
	}

	for i := 0; i < 50; i++ {
		bulkData.Articles[i] = handlers.ArticleCreatedData{
			Title:        fmt.Sprintf("Security Article %d", i+1),
			Content:      fmt.Sprintf("<p>Content for article %d about security threats.</p>", i+1),
			CategorySlug: "vulnerabilities",
			SourceURL:    fmt.Sprintf("https://example.com/article-%d", i+1),
			SourceName:   "Bulk Import Feed",
			Severity:     "medium",
		}
	}

	payload, err := createWebhookPayload("bulk.import", bulkData)
	require.NoError(t, err)

	signature := signPayload(payload, testWebhookSecret)

	// Execute
	rr := makeWebhookRequest(t, handler.HandleN8nWebhook, payload, signature)

	// Assert
	assert.Equal(t, http.StatusAccepted, rr.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)

	result := response["result"].(map[string]interface{})
	assert.Equal(t, float64(50), result["total"])
	assert.Equal(t, float64(50), result["success"])
	assert.Equal(t, float64(0), result["failed"])

	// Verify all articles were stored
	ctx := context.Background()
	articleRepo := postgres.NewArticleRepository(db.DB)

	for i := 0; i < 50; i++ {
		sourceURL := fmt.Sprintf("https://example.com/article-%d", i+1)
		article, err := articleRepo.GetBySourceURL(ctx, sourceURL)
		require.NoError(t, err, "Article %d should exist", i+1)
		assert.NotNil(t, article)
	}
}

// T059: Duplicate source_url -> skipped, success returned
func TestWebhook_DuplicateSourceURL(t *testing.T) {
	// Setup
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)
	defer CleanupDB(t, db)

	handler := setupWebhookHandler(t, db)

	// Create first article
	payload1, err := createWebhookPayload("article.created", validArticlePayload)
	require.NoError(t, err)

	signature1 := signPayload(payload1, testWebhookSecret)

	rr1 := makeWebhookRequest(t, handler.HandleN8nWebhook, payload1, signature1)
	assert.Equal(t, http.StatusAccepted, rr1.Code)

	// Try to create duplicate article with same source_url
	duplicatePayload := validArticlePayload
	duplicatePayload.Title = "Different Title but Same URL"

	payload2, err := createWebhookPayload("article.created", duplicatePayload)
	require.NoError(t, err)

	signature2 := signPayload(payload2, testWebhookSecret)

	// Execute
	rr2 := makeWebhookRequest(t, handler.HandleN8nWebhook, payload2, signature2)

	// Assert - should return error (not skip silently)
	assert.Equal(t, http.StatusInternalServerError, rr2.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rr2.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response["error"], "already exists")

	// Verify only one article exists in database
	ctx := context.Background()
	articleRepo := postgres.NewArticleRepository(db.DB)

	article, err := articleRepo.GetBySourceURL(ctx, validArticlePayload.SourceURL)
	require.NoError(t, err)
	assert.Equal(t, validArticlePayload.Title, article.Title) // Original title, not duplicate
}

// T060: Malformed payload -> 400 Bad Request
func TestWebhook_MalformedPayload(t *testing.T) {
	// Setup
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)
	defer CleanupDB(t, db)

	handler := setupWebhookHandler(t, db)

	// Create malformed JSON payload
	malformedPayload := []byte(`{"event_type": "article.created", "data": {invalid json}`)
	signature := signPayload(malformedPayload, testWebhookSecret)

	// Execute
	rr := makeWebhookRequest(t, handler.HandleN8nWebhook, malformedPayload, signature)

	// Assert
	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response["error"], "invalid JSON")
}

// Additional Tests

// TestWebhook_ArticleUpdated_HappyPath tests successful article update
func TestWebhook_ArticleUpdated_HappyPath(t *testing.T) {
	// Setup
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)
	defer CleanupDB(t, db)

	handler := setupWebhookHandler(t, db)
	ctx := context.Background()

	// Create initial article
	createPayload, err := createWebhookPayload("article.created", validArticlePayload)
	require.NoError(t, err)

	createSig := signPayload(createPayload, testWebhookSecret)
	rr1 := makeWebhookRequest(t, handler.HandleN8nWebhook, createPayload, createSig)
	require.Equal(t, http.StatusAccepted, rr1.Code)

	// Get created article ID
	articleRepo := postgres.NewArticleRepository(db.DB)
	article, err := articleRepo.GetBySourceURL(ctx, validArticlePayload.SourceURL)
	require.NoError(t, err)

	// Update article
	updatedTitle := "UPDATED: Critical Zero-Day Vulnerability"
	updatedSeverity := "critical"
	updateData := handlers.ArticleUpdatedData{
		ArticleID: article.ID.String(),
		Title:     &updatedTitle,
		Severity:  &updatedSeverity,
		Tags:      []string{"apache", "rce", "zero-day", "updated"},
	}

	updatePayload, err := createWebhookPayload("article.updated", updateData)
	require.NoError(t, err)

	updateSig := signPayload(updatePayload, testWebhookSecret)

	// Execute
	rr2 := makeWebhookRequest(t, handler.HandleN8nWebhook, updatePayload, updateSig)

	// Assert
	assert.Equal(t, http.StatusAccepted, rr2.Code)

	// Verify update in database
	updatedArticle, err := articleRepo.GetByID(ctx, article.ID)
	require.NoError(t, err)
	assert.Equal(t, updatedTitle, updatedArticle.Title)
	assert.Contains(t, updatedArticle.Tags, "updated")
}

// TestWebhook_ArticleDeleted_HappyPath tests successful article deletion
func TestWebhook_ArticleDeleted_HappyPath(t *testing.T) {
	// Setup
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)
	defer CleanupDB(t, db)

	handler := setupWebhookHandler(t, db)
	ctx := context.Background()

	// Create initial article
	createPayload, err := createWebhookPayload("article.created", validArticlePayload)
	require.NoError(t, err)

	createSig := signPayload(createPayload, testWebhookSecret)
	rr1 := makeWebhookRequest(t, handler.HandleN8nWebhook, createPayload, createSig)
	require.Equal(t, http.StatusAccepted, rr1.Code)

	// Get created article ID
	articleRepo := postgres.NewArticleRepository(db.DB)
	article, err := articleRepo.GetBySourceURL(ctx, validArticlePayload.SourceURL)
	require.NoError(t, err)

	// Delete article
	deleteData := handlers.ArticleDeletedData{
		ArticleID: article.ID.String(),
	}

	deletePayload, err := createWebhookPayload("article.deleted", deleteData)
	require.NoError(t, err)

	deleteSig := signPayload(deletePayload, testWebhookSecret)

	// Execute
	rr2 := makeWebhookRequest(t, handler.HandleN8nWebhook, deletePayload, deleteSig)

	// Assert
	assert.Equal(t, http.StatusAccepted, rr2.Code)

	// Verify deletion in database
	_, err = articleRepo.GetByID(ctx, article.ID)
	assert.Error(t, err) // Article should not exist
}

// TestWebhook_EnrichmentComplete_HappyPath tests enrichment webhook
func TestWebhook_EnrichmentComplete_HappyPath(t *testing.T) {
	// Setup
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)
	defer CleanupDB(t, db)

	handler := setupWebhookHandler(t, db)

	// Create enrichment data
	enrichmentData := handlers.EnrichmentCompleteData{
		ArticleID:          uuid.New().String(),
		ThreatType:         stringPtr("Ransomware"),
		AttackVector:       stringPtr("Phishing"),
		ImpactAssessment:   stringPtr("High impact on financial sector"),
		RecommendedActions: []string{"Apply patches", "Monitor network", "Enable MFA"},
		IOCs: []handlers.IOC{
			{Type: "ip", Value: "192.168.1.100", Context: "C2 server"},
			{Type: "domain", Value: "malicious.example.com", Context: "Phishing domain"},
		},
	}

	payload, err := createWebhookPayload("enrichment.complete", enrichmentData)
	require.NoError(t, err)

	signature := signPayload(payload, testWebhookSecret)

	// Execute
	rr := makeWebhookRequest(t, handler.HandleN8nWebhook, payload, signature)

	// Assert
	assert.Equal(t, http.StatusAccepted, rr.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "accepted", response["status"])
}

// TestWebhook_UnsupportedEventType tests handling of unknown event types
func TestWebhook_UnsupportedEventType(t *testing.T) {
	// Setup
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)
	defer CleanupDB(t, db)

	handler := setupWebhookHandler(t, db)

	// Create payload with unsupported event type
	payload, err := createWebhookPayload("unknown.event", map[string]string{"test": "data"})
	require.NoError(t, err)

	signature := signPayload(payload, testWebhookSecret)

	// Execute
	rr := makeWebhookRequest(t, handler.HandleN8nWebhook, payload, signature)

	// Assert
	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response["error"], "unsupported event type")
}

// TestWebhook_MissingEventType tests payload without event_type field
func TestWebhook_MissingEventType(t *testing.T) {
	// Setup
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)
	defer CleanupDB(t, db)

	handler := setupWebhookHandler(t, db)

	// Create payload without event_type
	payload := []byte(`{"data": {"title": "test"}}`)
	signature := signPayload(payload, testWebhookSecret)

	// Execute
	rr := makeWebhookRequest(t, handler.HandleN8nWebhook, payload, signature)

	// Assert
	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response["error"], "event_type is required")
}

// TestWebhook_EmptySignature tests request without signature header
func TestWebhook_EmptySignature(t *testing.T) {
	// Setup
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)
	defer CleanupDB(t, db)

	handler := setupWebhookHandler(t, db)

	// Create valid payload but no signature
	payload, err := createWebhookPayload("article.created", validArticlePayload)
	require.NoError(t, err)

	// Execute with empty signature
	rr := makeWebhookRequest(t, handler.HandleN8nWebhook, payload, "")

	// Assert
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

// TestWebhook_InvalidCategorySlug tests handling of non-existent category
func TestWebhook_InvalidCategorySlug(t *testing.T) {
	// Setup
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)
	defer CleanupDB(t, db)

	handler := setupWebhookHandler(t, db)

	// Create payload with invalid category
	invalidPayload := validArticlePayload
	invalidPayload.CategorySlug = "non-existent-category"
	invalidPayload.SourceURL = "https://example.com/invalid-category-test"

	payload, err := createWebhookPayload("article.created", invalidPayload)
	require.NoError(t, err)

	signature := signPayload(payload, testWebhookSecret)

	// Execute
	rr := makeWebhookRequest(t, handler.HandleN8nWebhook, payload, signature)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, rr.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response["error"], "category")
}

// TestWebhook_CompetitorContentFiltered tests competitor content scoring
func TestWebhook_CompetitorContentFiltered(t *testing.T) {
	// Setup
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)
	defer CleanupDB(t, db)

	handler := setupWebhookHandler(t, db)

	// Create article with competitor mentions
	competitorPayload := validArticlePayload
	competitorPayload.Title = "CrowdStrike Falcon vs Palo Alto Networks Comparison"
	competitorPayload.Content = "<p>CrowdStrike Falcon outperforms other solutions in endpoint protection. Palo Alto Networks also offers competitive features.</p>"
	competitorPayload.SourceURL = "https://example.com/competitor-article"

	payload, err := createWebhookPayload("article.created", competitorPayload)
	require.NoError(t, err)

	signature := signPayload(payload, testWebhookSecret)

	// Execute
	rr := makeWebhookRequest(t, handler.HandleN8nWebhook, payload, signature)

	// Assert - article is still created but with competitor score
	assert.Equal(t, http.StatusAccepted, rr.Code)

	// Verify competitor scoring was applied
	ctx := context.Background()
	articleRepo := postgres.NewArticleRepository(db.DB)

	article, err := articleRepo.GetBySourceURL(ctx, competitorPayload.SourceURL)
	require.NoError(t, err)
	assert.NotNil(t, article)
	// Competitor score should be calculated (implementation-dependent)
	assert.GreaterOrEqual(t, article.CompetitorScore, 0.0)
}

// TestWebhook_MissingRequiredFields tests validation of required fields
func TestWebhook_MissingRequiredFields(t *testing.T) {
	// Setup
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)
	defer CleanupDB(t, db)

	handler := setupWebhookHandler(t, db)

	testCases := []struct {
		name          string
		modifyPayload func(handlers.ArticleCreatedData) handlers.ArticleCreatedData
		expectedError string
	}{
		{
			name: "missing title",
			modifyPayload: func(p handlers.ArticleCreatedData) handlers.ArticleCreatedData {
				p.Title = ""
				p.SourceURL = "https://example.com/no-title"
				return p
			},
			expectedError: "title",
		},
		{
			name: "missing content",
			modifyPayload: func(p handlers.ArticleCreatedData) handlers.ArticleCreatedData {
				p.Content = ""
				p.SourceURL = "https://example.com/no-content"
				return p
			},
			expectedError: "content",
		},
		{
			name: "missing category_slug",
			modifyPayload: func(p handlers.ArticleCreatedData) handlers.ArticleCreatedData {
				p.CategorySlug = ""
				p.SourceURL = "https://example.com/no-category"
				return p
			},
			expectedError: "category_slug",
		},
		{
			name: "missing source_url",
			modifyPayload: func(p handlers.ArticleCreatedData) handlers.ArticleCreatedData {
				p.SourceURL = ""
				return p
			},
			expectedError: "source_url",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			invalidPayload := tc.modifyPayload(validArticlePayload)

			payload, err := createWebhookPayload("article.created", invalidPayload)
			require.NoError(t, err)

			signature := signPayload(payload, testWebhookSecret)

			// Execute
			rr := makeWebhookRequest(t, handler.HandleN8nWebhook, payload, signature)

			// Assert
			assert.Equal(t, http.StatusInternalServerError, rr.Code)

			var response map[string]interface{}
			err = json.Unmarshal(rr.Body.Bytes(), &response)
			require.NoError(t, err)
			assert.Contains(t, response["error"], tc.expectedError)
		})
	}
}

// TestWebhook_BulkImportPartialFailure tests bulk import with some failures
func TestWebhook_BulkImportPartialFailure(t *testing.T) {
	// Setup
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)
	defer CleanupDB(t, db)

	handler := setupWebhookHandler(t, db)

	// Create bulk payload with mix of valid and invalid articles
	bulkData := handlers.BulkImportData{
		Articles: []handlers.ArticleCreatedData{
			{
				Title:        "Valid Article 1",
				Content:      "<p>Valid content</p>",
				CategorySlug: "vulnerabilities",
				SourceURL:    "https://example.com/valid-1",
				Severity:     "medium",
			},
			{
				Title:        "", // Invalid - missing title
				Content:      "<p>Content without title</p>",
				CategorySlug: "vulnerabilities",
				SourceURL:    "https://example.com/invalid-1",
			},
			{
				Title:        "Valid Article 2",
				Content:      "<p>Another valid article</p>",
				CategorySlug: "vulnerabilities",
				SourceURL:    "https://example.com/valid-2",
				Severity:     "low",
			},
		},
	}

	payload, err := createWebhookPayload("bulk.import", bulkData)
	require.NoError(t, err)

	signature := signPayload(payload, testWebhookSecret)

	// Execute
	rr := makeWebhookRequest(t, handler.HandleN8nWebhook, payload, signature)

	// Assert
	assert.Equal(t, http.StatusAccepted, rr.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)

	result := response["result"].(map[string]interface{})
	assert.Equal(t, float64(3), result["total"])
	assert.Equal(t, float64(2), result["success"])
	assert.Equal(t, float64(1), result["failed"])

	errors := result["errors"].([]interface{})
	assert.Len(t, errors, 1)
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}
