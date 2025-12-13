package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
)

// ParsePagination extracts pagination parameters from request
// Returns page (1-indexed), pageSize, and any validation error
func ParsePagination(r *http.Request) (page int, pageSize int, err error) {
	page = 1
	pageSize = 20

	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		p, parseErr := strconv.Atoi(pageStr)
		if parseErr != nil {
			return 0, 0, fmt.Errorf("invalid page parameter: %w", parseErr)
		}
		if p < 1 {
			return 0, 0, fmt.Errorf("page must be at least 1")
		}
		page = p
	}

	if pageSizeStr := r.URL.Query().Get("page_size"); pageSizeStr != "" {
		ps, parseErr := strconv.Atoi(pageSizeStr)
		if parseErr != nil {
			return 0, 0, fmt.Errorf("invalid page_size parameter: %w", parseErr)
		}
		if ps < 1 {
			return 0, 0, fmt.Errorf("page_size must be at least 1")
		}
		if ps > 100 {
			return 0, 0, fmt.Errorf("page_size cannot exceed 100")
		}
		pageSize = ps
	}

	return page, pageSize, nil
}

// ParseLimitOffset extracts limit/offset pagination from request
func ParseLimitOffset(r *http.Request) (limit, offset int) {
	limit = 50
	offset = 0

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	return limit, offset
}

// CalculateTotalPages calculates total pages from total count and page size
func CalculateTotalPages(total, pageSize int) int {
	if pageSize <= 0 {
		return 0
	}
	pages := total / pageSize
	if total%pageSize > 0 {
		pages++
	}
	return pages
}

// getRequestID extracts request ID from context
func getRequestID(ctx context.Context) string {
	if id, ok := ctx.Value("request_id").(string); ok {
		return id
	}
	return ""
}

// GetClientIP extracts client IP from request headers
func GetClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fallback to RemoteAddr
	return r.RemoteAddr
}
