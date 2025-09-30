package types

// PaginationInfo represents pagination metadata
type PaginationInfo struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// MetaInfo represents metadata for API responses
type MetaInfo struct {
	Pagination PaginationInfo `json:"pagination"`
	Sort       string         `json:"sort,omitempty"`
	Filter     string         `json:"filter,omitempty"`
}

// APIResponse represents a standard API response structure
type APIResponse struct {
	Error   bool        `json:"error"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Meta    *MetaInfo   `json:"meta,omitempty"`
}
