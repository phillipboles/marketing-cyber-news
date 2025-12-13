package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/phillipboles/aci-backend/internal/ai"
	"github.com/phillipboles/aci-backend/internal/api"
	"github.com/phillipboles/aci-backend/internal/api/handlers"
	"github.com/phillipboles/aci-backend/internal/config"
	"github.com/phillipboles/aci-backend/internal/pkg/jwt"
	"github.com/phillipboles/aci-backend/internal/repository/postgres"
	"github.com/phillipboles/aci-backend/internal/service"
	"github.com/phillipboles/aci-backend/internal/websocket"
)

func main() {
	// Configure logger
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	ctx := context.Background()

	// Load configuration from environment
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	log.Info().
		Int("port", cfg.Server.Port).
		Str("log_level", cfg.Logger.Level).
		Msg("Configuration loaded")

	// Initialize database connection using pgxpool
	poolConfig, err := pgxpool.ParseConfig(cfg.Database.URL)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to parse database URL")
	}

	poolConfig.MaxConns = 25
	poolConfig.MinConns = 5
	poolConfig.MaxConnLifetime = time.Hour
	poolConfig.MaxConnIdleTime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create database pool")
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatal().Err(err).Msg("Failed to ping database")
	}

	log.Info().Msg("Database connection established")

	// Create postgres.DB wrapper for pgx-based repositories
	db := &postgres.DB{Pool: pool}

	// Create database/sql connection for repositories that still require it
	// (article_read_repo, audit_log_repo, bookmark_repo)
	connString := stdlib.RegisterConnConfig(poolConfig.ConnConfig)
	sqlDB, err := sql.Open("pgx", connString)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to open sql.DB connection")
	}
	defer sqlDB.Close()

	if err := sqlDB.Ping(); err != nil {
		log.Fatal().Err(err).Msg("Failed to ping sql.DB connection")
	}

	// Initialize JWT service
	jwtService, err := jwt.NewService(&jwt.Config{
		PrivateKeyPath: cfg.JWT.PrivateKeyPath,
		PublicKeyPath:  cfg.JWT.PublicKeyPath,
		Issuer:         "aci-backend",
	})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize JWT service")
	}

	log.Info().Msg("JWT service initialized")

	// Initialize AI client and enricher
	aiClient, err := ai.NewClient(ai.Config{
		APIKey: cfg.AI.AnthropicAPIKey,
		Model:  "claude-3-haiku-20240307",
	})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize AI client")
	}

	enricher := ai.NewEnricher(aiClient)
	log.Info().Msg("AI enrichment service initialized")

	// Initialize repositories
	// Repositories using postgres.DB (pgx-based)
	userRepo := postgres.NewUserRepository(db)
	tokenRepo := postgres.NewRefreshTokenRepository(db)
	articleRepo := postgres.NewArticleRepository(db)
	categoryRepo := postgres.NewCategoryRepository(db)
	sourceRepo := postgres.NewSourceRepository(db)
	webhookLogRepo := postgres.NewWebhookLogRepository(db)
	alertRepo := postgres.NewAlertRepository(db)
	alertMatchRepo := postgres.NewAlertMatchRepository(db)

	// Repositories still using *sql.DB
	bookmarkRepo := postgres.NewBookmarkRepository(sqlDB)
	articleReadRepo := postgres.NewArticleReadRepository(sqlDB)
	_ = postgres.NewAuditLogRepository(sqlDB) // TODO: Wire into AdminService once UserRepository type mismatch is resolved

	log.Info().Msg("Repositories initialized")

	// Initialize WebSocket hub
	hub := websocket.NewHub(&websocket.HubConfig{
		MaxConnectionsPerUser: 5,
		MaxChannelsPerClient:  50,
	})

	// Start hub in background
	go hub.Run()
	log.Info().Msg("WebSocket hub started")

	// Initialize services
	authService := service.NewAuthService(userRepo, tokenRepo, jwtService)
	articleService := service.NewArticleService(articleRepo, categoryRepo, sourceRepo, webhookLogRepo)
	alertService := service.NewAlertService(alertRepo, alertMatchRepo, articleRepo)
	searchService := service.NewSearchService(articleRepo)
	engagementService := service.NewEngagementService(bookmarkRepo, articleReadRepo, articleRepo)
	enrichmentService := service.NewEnrichmentService(enricher, articleRepo)

	// NOTE: AdminService initialization blocked due to interface mismatch
	// UserRepository expects domain.User but postgres.UserRepository uses entities.User
	// This needs to be resolved before AdminService can be initialized
	// adminService := service.NewAdminService(articleRepo, sourceRepo, userRepo, auditLogRepo)

	notificationService, err := service.NewNotificationService(hub)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize notification service")
	}

	log.Info().Msg("Services initialized")

	// Initialize WebSocket handler
	wsHandler, err := websocket.NewHandler(hub, jwtService)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize WebSocket handler")
	}

	// Initialize HTTP handlers
	authHandler := handlers.NewAuthHandler(authService)
	articleHandler := handlers.NewArticleHandler(articleRepo, searchService, engagementService)
	alertHandler := handlers.NewAlertHandler(alertService)
	categoryHandler := handlers.NewCategoryHandler(categoryRepo, articleRepo)
	userHandler := handlers.NewUserHandler(engagementService, userRepo)
	webhookHandler := handlers.NewWebhookHandler(articleService, webhookLogRepo, cfg.N8N.WebhookSecret)

	// NOTE: AdminHandler blocked until AdminService interface issue is resolved
	// adminHandler := handlers.NewAdminHandler(adminService)

	log.Info().Msg("Handlers initialized")

	// Create HTTP server
	// TODO: Router agent needs to wire handlers into SetupRoutes()
	// Services available: notificationService, enrichmentService
	// NOTE: adminHandler not available until UserRepository interface mismatch resolved
	handlers := &api.Handlers{
		Auth:     authHandler,
		Article:  articleHandler,
		Alert:    alertHandler,
		Webhook:  webhookHandler,
		User:     userHandler,
		Admin:    nil, // TODO: Wire AdminHandler once UserRepository type mismatch is resolved
		Category: categoryHandler,
	}

	serverConfig := api.Config{
		Port:         cfg.Server.Port,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Create server with WebSocket handler wired
	server := api.NewServerWithWebSocket(serverConfig, handlers, jwtService, wsHandler)

	// Prevent unused variable warnings until services are wired
	_ = notificationService
	_ = enrichmentService

	log.Info().Msg("ACI Backend server starting...")

	// Start HTTP server in background
	serverErrChan := make(chan error, 1)
	go func() {
		if err := server.Start(); err != nil {
			serverErrChan <- err
		}
	}()

	// Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrChan:
		log.Fatal().Err(err).Msg("Server failed to start")
	case sig := <-sigChan:
		log.Info().Str("signal", sig.String()).Msg("Shutdown signal received, gracefully stopping...")
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Shutdown HTTP server
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("Server shutdown failed")
	}

	// Close database connections
	pool.Close()
	sqlDB.Close()
	log.Info().Msg("Database connections closed")

	// Hub cleanup happens automatically when goroutines finish

	<-shutdownCtx.Done()
	if shutdownCtx.Err() == context.DeadlineExceeded {
		log.Warn().Msg("Shutdown deadline exceeded")
	}

	log.Info().Msg("Server stopped")
	fmt.Println("Goodbye!")
}
