package models

// PaginationParams holds pagination parameters from query strings.
type PaginationParams struct {
	Limit  int
	Offset int
}

// PaginatedResponse wraps paginated list responses.
type PaginatedResponse struct {
	Total  int64 `json:"total"`
	Limit  int   `json:"limit"`
	Offset int   `json:"offset"`
}

// DefaultPagination returns default pagination parameters.
func DefaultPagination() PaginationParams {
	return PaginationParams{
		Limit:  100,
		Offset: 0,
	}
}

// Validate ensures pagination parameters are within acceptable bounds.
func (p *PaginationParams) Validate() {
	if p.Limit <= 0 {
		p.Limit = 100
	}
	if p.Limit > 1000 {
		p.Limit = 1000 // Max limit to prevent DoS
	}
	if p.Offset < 0 {
		p.Offset = 0
	}
}
