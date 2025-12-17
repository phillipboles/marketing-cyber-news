package integration

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"database/sql"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/stdlib"
	"github.com/phillipboles/aci-backend/internal/ai"
	"github.com/phillipboles/aci-backend/internal/api"
	"github.com/phillipboles/aci-backend/internal/api/handlers"
	"github.com/phillipboles/aci-backend/internal/pkg/jwt"
	"github.com/phillipboles/aci-backend/internal/repository/postgres"
	"github.com/phillipboles/aci-backend/internal/service"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/testcontainers/testcontainers-go"
	testcontainerspostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	// Use pgvector-enabled postgres image for vector search support
	defaultPostgresImage = "pgvector/pgvector:pg17"
	defaultDatabase      = "aci_backend_test"
	defaultUsername      = "test"
	defaultPassword      = "test"
	testTimeout          = 60 * time.Second
)

// TestDB holds the test database container and connection details
type TestDB struct {
	Container testcontainers.Container
	DSN       string
	DB        *postgres.DB
	SqlDB     *sql.DB // For repositories that still use database/sql
}

// TestKeys holds test RSA key pairs for JWT
type TestKeys struct {
	PrivateKeyPath string
	PublicKeyPath  string
	TempDir        string
}

// TestServer holds the test HTTP server and dependencies
type TestServer struct {
	Server     *httptest.Server
	DB         *TestDB
	Keys       *TestKeys
	JWTService jwt.Service
	BaseURL    string
}

// SetupTestDB creates a PostgreSQL container for testing
func SetupTestDB(t *testing.T) *TestDB {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	// Create PostgreSQL container
	container, err := testcontainerspostgres.Run(ctx,
		defaultPostgresImage,
		testcontainerspostgres.WithDatabase(defaultDatabase),
		testcontainerspostgres.WithUsername(defaultUsername),
		testcontainerspostgres.WithPassword(defaultPassword),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(testTimeout),
		),
	)
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}

	// Get connection string
	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		container.Terminate(ctx)
		t.Fatalf("failed to get connection string: %v", err)
	}

	// Parse the DSN to extract host and port from the container
	host, port, err := parseHostPortFromDSN(dsn)
	if err != nil {
		container.Terminate(ctx)
		t.Fatalf("failed to parse DSN: %v", err)
	}

	// Create database connection using the container's host and port
	db, err := postgres.NewDB(ctx, postgres.Config{
		Host:        host,
		Port:        port,
		User:        defaultUsername,
		Password:    defaultPassword,
		Database:    defaultDatabase,
		SSLMode:     "disable",
		MaxConns:    10,
		MinConns:    2,
		MaxConnLife: time.Hour,
		MaxConnIdle: 30 * time.Minute,
	})
	if err != nil {
		container.Terminate(ctx)
		t.Fatalf("failed to create database connection: %v", err)
	}

	// Create sql.DB connection for repositories that still use database/sql
	// Register the pgx connection config with stdlib
	connString := stdlib.RegisterConnConfig(db.Pool.Config().ConnConfig)
	sqlDB, err := sql.Open("pgx", connString)
	if err != nil {
		db.Close()
		container.Terminate(ctx)
		t.Fatalf("failed to create sql.DB connection: %v", err)
	}

	if err := sqlDB.Ping(); err != nil {
		sqlDB.Close()
		db.Close()
		container.Terminate(ctx)
		t.Fatalf("failed to ping sql.DB connection: %v", err)
	}

	testDB := &TestDB{
		Container: container,
		DSN:       dsn,
		DB:        db,
		SqlDB:     sqlDB,
	}

	// Run migrations
	if err := runMigrations(ctx, testDB); err != nil {
		TeardownTestDB(t, testDB)
		t.Fatalf("failed to run migrations: %v", err)
	}

	return testDB
}

// parseHostPortFromDSN extracts host and port from a PostgreSQL DSN
// Format: postgres://user:password@host:port/database?params
func parseHostPortFromDSN(dsn string) (string, int, error) {
	parsedURL, err := url.Parse(dsn)
	if err != nil {
		return "", 0, fmt.Errorf("failed to parse DSN URL: %w", err)
	}

	host := parsedURL.Hostname()
	if host == "" {
		host = "localhost"
	}

	portStr := parsedURL.Port()
	if portStr == "" {
		return host, 5432, nil
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return "", 0, fmt.Errorf("failed to parse port: %w", err)
	}

	return host, port, nil
}


// runMigrations applies database migrations
func runMigrations(ctx context.Context, testDB *TestDB) error {
	// Get migrations directory
	migrationsDir := filepath.Join("..", "..", "migrations")

	// Read and execute migrations in order
	migrations := []string{
		"000001_initial_schema.up.sql",
		"000002_content_schema.up.sql",
		"000003_alerts_schema.up.sql",
		"000004_engagement_schema.up.sql",
		"000005_audit_schema.up.sql",
	}

	for _, migration := range migrations {
		migrationPath := filepath.Join(migrationsDir, migration)
		content, err := os.ReadFile(migrationPath)
		if err != nil {
			return fmt.Errorf("failed to read migration %s: %w", migration, err)
		}

		if _, err := testDB.DB.Pool.Exec(ctx, string(content)); err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", migration, err)
		}
	}

	return nil
}

// TeardownTestDB cleans up the database container
func TeardownTestDB(t *testing.T, testDB *TestDB) {
	t.Helper()

	if testDB == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	// Close sql.DB connection
	if testDB.SqlDB != nil {
		testDB.SqlDB.Close()
	}

	// Close database connection
	if testDB.DB != nil {
		testDB.DB.Close()
	}

	// Terminate container
	if testDB.Container != nil {
		if err := testDB.Container.Terminate(ctx); err != nil {
			t.Logf("failed to terminate container: %v", err)
		}
	}
}

// SetupTestKeys generates temporary RSA keys for testing
func SetupTestKeys(t *testing.T) *TestKeys {
	t.Helper()

	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "aci-test-keys-*")
	if err != nil {
		t.Fatalf("failed to create temp directory: %v", err)
	}

	// Generate RSA private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("failed to generate private key: %v", err)
	}

	// Save private key
	privateKeyPath := filepath.Join(tempDir, "private.pem")
	privateKeyFile, err := os.Create(privateKeyPath)
	if err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("failed to create private key file: %v", err)
	}
	defer privateKeyFile.Close()

	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	if err := pem.Encode(privateKeyFile, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	}); err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("failed to encode private key: %v", err)
	}

	// Save public key
	publicKeyPath := filepath.Join(tempDir, "public.pem")
	publicKeyFile, err := os.Create(publicKeyPath)
	if err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("failed to create public key file: %v", err)
	}
	defer publicKeyFile.Close()

	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("failed to marshal public key: %v", err)
	}

	if err := pem.Encode(publicKeyFile, &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	}); err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("failed to encode public key: %v", err)
	}

	return &TestKeys{
		PrivateKeyPath: privateKeyPath,
		PublicKeyPath:  publicKeyPath,
		TempDir:        tempDir,
	}
}

// TeardownTestKeys removes temporary keys
func TeardownTestKeys(t *testing.T, keys *TestKeys) {
	t.Helper()

	if keys != nil && keys.TempDir != "" {
		if err := os.RemoveAll(keys.TempDir); err != nil {
			t.Logf("failed to remove temp directory: %v", err)
		}
	}
}

// SetupTestServer creates a test API server with all dependencies
func SetupTestServer(t *testing.T, testDB *TestDB) *TestServer {
	t.Helper()

	// Setup test keys
	keys := SetupTestKeys(t)

	// Create JWT service
	jwtService, err := jwt.NewService(&jwt.Config{
		PrivateKeyPath: keys.PrivateKeyPath,
		PublicKeyPath:  keys.PublicKeyPath,
		Issuer:         "aci-backend-test",
	})
	if err != nil {
		TeardownTestKeys(t, keys)
		t.Fatalf("failed to create JWT service: %v", err)
	}

	// Configure zerolog for testing (silent by default, change to DebugLevel for debugging)
	zerolog.SetGlobalLevel(zerolog.Disabled)
	log.Logger = zerolog.Nop()

	// Create repositories using pgx pool
	userRepo := postgres.NewUserRepository(testDB.DB)
	tokenRepo := postgres.NewRefreshTokenRepository(testDB.DB)
	articleRepo := postgres.NewArticleRepository(testDB.DB)
	categoryRepo := postgres.NewCategoryRepository(testDB.DB)
	sourceRepo := postgres.NewSourceRepository(testDB.DB)
	webhookLogRepo := postgres.NewWebhookLogRepository(testDB.DB)
	alertRepo := postgres.NewAlertRepository(testDB.DB)
	alertMatchRepo := postgres.NewAlertMatchRepository(testDB.DB)

	// Create repositories using sql.DB (for engagement service)
	bookmarkRepo := postgres.NewBookmarkRepository(testDB.SqlDB)
	articleReadRepo := postgres.NewArticleReadRepository(testDB.SqlDB)

	// Create services
	authService := service.NewAuthService(userRepo, tokenRepo, jwtService)
	articleService := service.NewArticleService(articleRepo, categoryRepo, sourceRepo, webhookLogRepo)
	alertService := service.NewAlertService(alertRepo, alertMatchRepo, articleRepo)
	searchService := service.NewSearchService(articleRepo)
	engagementService := service.NewEngagementService(bookmarkRepo, articleReadRepo, articleRepo)

	// Create AI client for enrichment service (with dummy API key for testing)
	// Most integration tests don't actually call enrichment, so this won't make real API calls
	aiClient, err := ai.NewClient(ai.Config{
		APIKey: "test-api-key-placeholder",
		Model:  "claude-3-haiku-20240307",
	})
	if err != nil {
		TeardownTestKeys(t, keys)
		t.Fatalf("failed to create AI client: %v", err)
	}

	enricher := ai.NewEnricher(aiClient)
	enrichmentService := service.NewEnrichmentService(enricher, articleRepo)

	// Create handlers
	authHandler := handlers.NewAuthHandler(authService)
	articleHandler := handlers.NewArticleHandler(articleRepo, searchService, engagementService)
	alertHandler := handlers.NewAlertHandler(alertService)
	categoryHandler := handlers.NewCategoryHandler(categoryRepo, articleRepo)
	userHandler := handlers.NewUserHandler(engagementService, userRepo)
	webhookHandler := handlers.NewWebhookHandler(articleService, enrichmentService, webhookLogRepo, "test-webhook-secret")

	// Create Handlers struct
	h := &api.Handlers{
		Auth:     authHandler,
		Article:  articleHandler,
		Alert:    alertHandler,
		Webhook:  webhookHandler,
		User:     userHandler,
		Admin:    nil,
		Category: categoryHandler,
	}

	// Create API server with new signature
	server := api.NewServer(api.Config{
		Port:         8080,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}, h, jwtService)

	// Create test HTTP server
	testServer := httptest.NewServer(server)

	return &TestServer{
		Server:     testServer,
		DB:         testDB,
		Keys:       keys,
		JWTService: jwtService,
		BaseURL:    testServer.URL,
	}
}


// TeardownTestServer cleans up the test server
func TeardownTestServer(t *testing.T, testServer *TestServer) {
	t.Helper()

	if testServer == nil {
		return
	}

	// Close HTTP server
	if testServer.Server != nil {
		testServer.Server.Close()
	}

	// Cleanup keys
	TeardownTestKeys(t, testServer.Keys)
}

// PostJSON makes a POST request with JSON body
func PostJSON(t *testing.T, url string, body interface{}) *http.Response {
	t.Helper()

	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("failed to marshal request body: %v", err)
		}
		bodyReader = strings.NewReader(string(jsonBody))
	}

	req, err := http.NewRequest(http.MethodPost, url, bodyReader)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("failed to execute request: %v", err)
	}

	return resp
}

// GetJSON makes a GET request with optional authorization token
func GetJSON(t *testing.T, url string, token string) *http.Response {
	t.Helper()

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("failed to execute request: %v", err)
	}

	return resp
}

// PutJSON makes a PUT request with JSON body and optional authorization
func PutJSON(t *testing.T, url string, body interface{}, token string) *http.Response {
	t.Helper()

	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("failed to marshal request body: %v", err)
		}
		bodyReader = strings.NewReader(string(jsonBody))
	}

	req, err := http.NewRequest(http.MethodPut, url, bodyReader)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("failed to execute request: %v", err)
	}

	return resp
}

// DeleteJSON makes a DELETE request with optional authorization
func DeleteJSON(t *testing.T, url string, token string) *http.Response {
	t.Helper()

	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("failed to execute request: %v", err)
	}

	return resp
}

// ReadJSONResponse reads and unmarshals JSON response body
func ReadJSONResponse(t *testing.T, resp *http.Response, v interface{}) {
	t.Helper()

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}

	if err := json.Unmarshal(body, v); err != nil {
		t.Fatalf("failed to unmarshal response body: %v\nBody: %s", err, string(body))
	}
}

// ReadResponseBody reads response body as string
func ReadResponseBody(t *testing.T, resp *http.Response) string {
	t.Helper()

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}

	return string(body)
}

// CleanupDB truncates all tables for test isolation
func CleanupDB(t *testing.T, testDB *TestDB) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	// Truncate tables in reverse order of dependencies
	tables := []string{
		"alert_matches",
		"alerts",
		"bookmarks",
		"read_history",
		"article_categories",
		"articles",
		"categories",
		"sources",
		"refresh_tokens",
		"user_preferences",
		"users",
		"audit_logs",
	}

	for _, table := range tables {
		_, err := testDB.DB.Pool.Exec(ctx, fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table))
		if err != nil {
			// Table might not exist, continue
			t.Logf("warning: failed to truncate table %s: %v", table, err)
		}
	}
}
