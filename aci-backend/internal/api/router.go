package api

import (
	"net/http"

	"github.com/phillipboles/aci-backend/internal/api/handlers"
	"github.com/phillipboles/aci-backend/internal/api/middleware"
	"github.com/phillipboles/aci-backend/internal/api/response"

	"github.com/go-chi/chi/v5"
)

// WebSocketHandler interface for WebSocket upgrade handling
type WebSocketHandler interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}

// SetupRoutes configures all API routes and middleware
func (s *Server) SetupRoutes() {
	s.setupRoutesWithWebSocket(nil)
}

// SetupRoutesWithWebSocket configures all API routes with optional WebSocket handler
func (s *Server) SetupRoutesWithWebSocket(wsHandler WebSocketHandler) {
	s.setupRoutesWithWebSocket(wsHandler)
}

// setupRoutesWithWebSocket is the internal implementation for route setup
func (s *Server) setupRoutesWithWebSocket(wsHandler WebSocketHandler) {
	// Apply global middleware in order
	s.router.Use(middleware.RequestID)
	s.router.Use(middleware.Logger)
	s.router.Use(middleware.Recoverer)
	s.router.Use(middleware.CORS)

	// Health endpoints (no authentication required)
	s.router.Get("/health", handlers.HealthCheck)
	s.router.Get("/ready", handlers.ReadinessCheck)

	// WebSocket endpoint (authentication handled in handler via query param token)
	if wsHandler != nil {
		s.router.Get("/ws", wsHandler.ServeHTTP)
	}

	// API v1 routes
	s.router.Route("/v1", func(r chi.Router) {
		// Auth routes (no authentication required)
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", s.handlers.Auth.Register)
			r.Post("/login", s.handlers.Auth.Login)
			r.Post("/refresh", s.handlers.Auth.Refresh)
			r.Post("/logout", s.handlers.Auth.Logout)
		})

		// Category routes (no authentication required)
		r.Route("/categories", func(r chi.Router) {
			r.Get("/", s.handlers.Category.List)
			r.Get("/{slug}", s.handlers.Category.GetBySlug)
		})

		// Webhook routes (HMAC validation handled in handler)
		r.Route("/webhooks", func(r chi.Router) {
			r.Post("/n8n", s.handlers.Webhook.HandleN8nWebhook)
		})

		// Protected routes (authentication required)
		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(s.jwtService))

			// Article routes
			r.Route("/articles", func(r chi.Router) {
				r.Get("/", s.handlers.Article.List)
				r.Get("/search", s.handlers.Article.Search)
				r.Get("/{id}", s.handlers.Article.GetByID)
				r.Get("/slug/{slug}", s.handlers.Article.GetBySlug)

				// Article engagement routes
				r.Post("/{id}/bookmark", s.handlers.Article.AddBookmark)
				r.Delete("/{id}/bookmark", s.handlers.Article.RemoveBookmark)
				r.Post("/{id}/read", s.handlers.Article.MarkRead)
			})

			// Alert routes
			r.Route("/alerts", func(r chi.Router) {
				r.Get("/", s.handlers.Alert.List)
				r.Post("/", s.handlers.Alert.Create)
				r.Get("/{id}", s.handlers.Alert.GetByID)
				r.Patch("/{id}", s.handlers.Alert.Update)
				r.Delete("/{id}", s.handlers.Alert.Delete)
				r.Get("/{id}/matches", s.handlers.Alert.ListMatches)
			})

			// User routes
			r.Route("/users", func(r chi.Router) {
				r.Get("/me", s.handlers.User.GetCurrentUser)
				r.Patch("/me", s.handlers.User.UpdateCurrentUser)
				r.Get("/me/bookmarks", s.handlers.User.GetBookmarks)
				r.Get("/me/history", s.handlers.User.GetReadingHistory)
				r.Get("/me/stats", s.handlers.User.GetStats)
			})

			// Admin routes (require admin role)
			r.Route("/admin", func(r chi.Router) {
				r.Use(middleware.RequireAdmin())

				// Handle case where Admin handler is not initialized
				if s.handlers.Admin == nil {
					r.HandleFunc("/*", func(w http.ResponseWriter, req *http.Request) {
						response.ServiceUnavailable(w, "Admin service is not available")
					})
					return
				}

				// Article management
				r.Put("/articles/{id}", s.handlers.Admin.UpdateArticle)
				r.Delete("/articles/{id}", s.handlers.Admin.DeleteArticle)

				// Source management
				r.Get("/sources", s.handlers.Admin.ListSources)
				r.Post("/sources", s.handlers.Admin.CreateSource)
				r.Put("/sources/{id}", s.handlers.Admin.UpdateSource)
				r.Delete("/sources/{id}", s.handlers.Admin.DeleteSource)

				// User management
				r.Get("/users", s.handlers.Admin.ListUsers)
				r.Put("/users/{id}", s.handlers.Admin.UpdateUser)
				r.Delete("/users/{id}", s.handlers.Admin.DeleteUser)

				// Audit logs
				r.Get("/audit-logs", s.handlers.Admin.ListAuditLogs)
			})
		})
	})
}
