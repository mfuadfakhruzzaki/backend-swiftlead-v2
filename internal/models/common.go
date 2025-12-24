package models

// Common types used across the application

// PaginationParams for query parameters
type PaginationParams struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
}

// PaginationMeta for response metadata
type PaginationMeta struct {
	Total      int `json:"total"`
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	TotalPages int `json:"total_pages"`
}

// NewPaginationMeta creates pagination metadata
func NewPaginationMeta(total, page, limit int) *PaginationMeta {
	totalPages := total / limit
	if total%limit > 0 {
		totalPages++
	}
	return &PaginationMeta{
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}
}

// DefaultPagination returns default pagination params
func DefaultPagination() PaginationParams {
	return PaginationParams{
		Page:  1,
		Limit: 20,
	}
}
