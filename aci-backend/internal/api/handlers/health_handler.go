package handlers

import (
	"net/http"
	"time"

	"github.com/phillipboles/aci-backend/internal/api/response"
)

const version = "1.0.0"

// HealthCheck returns the health status of the service
// GET /health
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	healthData := map[string]interface{}{
		"status":    "healthy",
		"version":   version,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	response.Success(w, healthData)
}

// ReadinessCheck returns the readiness status of the service
// GET /ready
func ReadinessCheck(w http.ResponseWriter, r *http.Request) {
	// TODO: Add actual health checks for dependencies (database, redis, etc.)
	// For now, return a basic ready status
	readinessData := map[string]interface{}{
		"status": "ready",
		"checks": map[string]string{
			"database": "ok",
			"redis":    "ok",
		},
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	response.Success(w, readinessData)
}
