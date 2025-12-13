package websocket

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
	jwtPkg "github.com/phillipboles/aci-backend/internal/pkg/jwt"
	"github.com/rs/zerolog/log"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// TODO: Configure allowed origins in production
		// For now, allow all origins
		return true
	},
}

// Handler handles WebSocket upgrade requests
type Handler struct {
	hub        *Hub
	jwtService jwtPkg.Service
}

// NewHandler creates a new WebSocket handler
func NewHandler(hub *Hub, jwtService jwtPkg.Service) (*Handler, error) {
	if hub == nil {
		return nil, fmt.Errorf("hub is required")
	}

	if jwtService == nil {
		return nil, fmt.Errorf("jwt service is required")
	}

	return &Handler{
		hub:        hub,
		jwtService: jwtService,
	}, nil
}

// ServeWS handles WebSocket upgrade requests
// GET /ws?token=<jwt>
func (h *Handler) ServeWS(w http.ResponseWriter, r *http.Request) {
	// Extract JWT from query parameter
	tokenString := r.URL.Query().Get("token")
	if tokenString == "" {
		log.Warn().
			Str("remote_addr", r.RemoteAddr).
			Msg("WebSocket connection attempt without token")
		http.Error(w, "Token is required", http.StatusUnauthorized)
		return
	}

	// Validate JWT
	claims, err := h.jwtService.ValidateAccessToken(tokenString)
	if err != nil {
		log.Warn().
			Err(err).
			Str("remote_addr", r.RemoteAddr).
			Msg("Invalid JWT token for WebSocket")
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// Check connection limit before upgrading
	if h.hub.GetConnectionCount(claims.UserID) >= h.hub.maxConnectionsPerUser {
		log.Warn().
			Str("user_id", claims.UserID.String()).
			Int("current_connections", h.hub.GetConnectionCount(claims.UserID)).
			Msg("Max connections per user reached")
		http.Error(w, "Max connections reached", http.StatusTooManyRequests)
		return
	}

	// Upgrade connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error().
			Err(err).
			Str("remote_addr", r.RemoteAddr).
			Msg("Failed to upgrade WebSocket connection")
		return
	}

	// Extract token expiration
	tokenExp := claims.ExpiresAt.Time

	// Create new client
	client := NewClient(h.hub, conn, claims.UserID, claims.Email, claims.Role, tokenExp)

	// Register client with hub
	if err := h.hub.RegisterClient(client); err != nil {
		log.Error().
			Err(err).
			Str("user_id", claims.UserID.String()).
			Msg("Failed to register client")
		conn.Close()
		return
	}

	log.Info().
		Str("user_id", claims.UserID.String()).
		Str("email", claims.Email).
		Str("role", claims.Role).
		Str("remote_addr", r.RemoteAddr).
		Msg("WebSocket connection established")

	// Start read and write pumps in separate goroutines
	go client.WritePump()
	go client.ReadPump()
}

// ServeHTTP implements http.Handler interface for routing integration
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.ServeWS(w, r)
}
