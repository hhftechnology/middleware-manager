package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

// PaginationParams holds pagination parameters
type PaginationParams struct {
	Page     int
	PageSize int
	Offset   int
}

// PaginatedResponse represents a paginated API response
type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Total      int         `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalPages int         `json:"total_pages"`
}

// DefaultPageSize is the default number of items per page
const DefaultPageSize = 50

// MaxPageSize is the maximum number of items per page
const MaxPageSize = 100

// GetPaginationParams extracts pagination parameters from the request
func GetPaginationParams(c *gin.Context) PaginationParams {
	// Get page number (default: 1)
	page := 1
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	// Get page size (default: DefaultPageSize)
	pageSize := DefaultPageSize
	if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 {
			pageSize = ps
		}
	}

	// Enforce max page size
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}

	// Calculate offset
	offset := (page - 1) * pageSize

	return PaginationParams{
		Page:     page,
		PageSize: pageSize,
		Offset:   offset,
	}
}

// NewPaginatedResponse creates a new paginated response
func NewPaginatedResponse(data interface{}, total int, params PaginationParams) PaginatedResponse {
	totalPages := total / params.PageSize
	if total%params.PageSize > 0 {
		totalPages++
	}

	return PaginatedResponse{
		Data:       data,
		Total:      total,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: totalPages,
	}
}

// IsPaginationRequested checks if pagination is requested
func IsPaginationRequested(c *gin.Context) bool {
	return c.Query("page") != "" || c.Query("page_size") != ""
}
