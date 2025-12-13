package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"

	"github.com/phillipboles/aci-backend/internal/api/handlers"
	"github.com/phillipboles/aci-backend/internal/pkg/jwt"
)

// Server represents the HTTP API server
type Server struct {
	httpServer *http.Server
	router     *chi.Mux
	handlers   *Handlers
	jwtService jwt.Service
}

// Handlers holds all HTTP handlers
type Handlers struct {
	Auth     *handlers.AuthHandler
	Article  *handlers.ArticleHandler
	Alert    *handlers.AlertHandler
	Webhook  *handlers.WebhookHandler
	User     *handlers.UserHandler
	Admin    *handlers.AdminHandler
	Category *handlers.CategoryHandler
}

// Config holds server configuration
type Config struct {
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// NewServer creates a new API server with the provided configuration
func NewServer(cfg Config, h *Handlers, jwtService jwt.Service) *Server {
	return NewServerWithWebSocket(cfg, h, jwtService, nil)
}

// NewServerWithWebSocket creates a new API server with WebSocket support
func NewServerWithWebSocket(cfg Config, h *Handlers, jwtService jwt.Service, wsHandler WebSocketHandler) *Server {
	if h == nil {
		panic("handlers cannot be nil")
	}
	if jwtService == nil {
		panic("jwtService cannot be nil")
	}

	router := chi.NewRouter()

	server := &Server{
		router:     router,
		handlers:   h,
		jwtService: jwtService,
		httpServer: &http.Server{
			Addr:         fmt.Sprintf(":%d", cfg.Port),
			Handler:      router,
			ReadTimeout:  cfg.ReadTimeout,
			WriteTimeout: cfg.WriteTimeout,
			IdleTimeout:  cfg.IdleTimeout,
		},
	}

	// Setup all routes and middleware with optional WebSocket handler
	server.SetupRoutesWithWebSocket(wsHandler)

	return server
}

// Start begins listening for HTTP requests
func (s *Server) Start() error {
	log.Info().
		Str("addr", s.httpServer.Addr).
		Msg("Starting HTTP server")

	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}

// Shutdown gracefully shuts down the server without interrupting active connections
func (s *Server) Shutdown(ctx context.Context) error {
	log.Info().Msg("Shutting down HTTP server gracefully")

	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	log.Info().Msg("HTTP server shut down successfully")
	return nil
}

// ServeHTTP implements the http.Handler interface
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

// Router returns the underlying chi router for testing purposes
func (s *Server) Router() *chi.Mux {
	return s.router
}
