package response

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog/log"
)

// Response represents a standard API response
type Response struct {
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
}

// Meta contains pagination metadata
type Meta struct {
	Page       int `json:"page,omitempty"`
	PageSize   int `json:"page_size,omitempty"`
	TotalCount int `json:"total_count,omitempty"`
	TotalPages int `json:"total_pages,omitempty"`
}

// JSON sends a JSON response with the specified status code and data
func JSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			log.Error().
				Err(err).
				Msg("Failed to encode JSON response")
		}
	}
}

// Success sends a successful JSON response with data
func Success(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusOK, Response{Data: data})
}

// SuccessWithMessage sends a successful JSON response with data and message
func SuccessWithMessage(w http.ResponseWriter, data interface{}, message string) {
	JSON(w, http.StatusOK, Response{
		Data:    data,
		Message: message,
	})
}

// SuccessWithMeta sends a successful JSON response with data and pagination metadata
func SuccessWithMeta(w http.ResponseWriter, data interface{}, meta *Meta) {
	JSON(w, http.StatusOK, Response{
		Data: data,
		Meta: meta,
	})
}

// Created sends a 201 Created response with data
func Created(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusCreated, Response{Data: data})
}

// NoContent sends a 204 No Content response
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// Error sends an error response with the specified status code, error code, and message
func Error(w http.ResponseWriter, status int, code, message string) {
	ErrorWithDetails(w, status, code, message, nil, "")
}
