package offsetpagination

import (
	"math"
)

const (
	DefaultPage  = 1
	DefaultLimit = 10
	MaxLimit     = 50
)

// Meta holds pagination metadata for API responses
type Meta struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalItems int64 `json:"total_items"`
	TotalPages int   `json:"total_pages"`
	HasNext    bool  `json:"has_next"`
	HasPrev    bool  `json:"has_prev"`
}

// NewMeta constructs a Meta struct from pagination params and total count.
func NewMeta(p Params, totalItems int64) Meta {
	totalPages := int(math.Ceil(float64(totalItems) / float64(p.Limit)))
	if totalPages == 0 {
		totalPages = 1
	}

	return Meta{
		Page:       p.Page,
		Limit:      p.Limit,
		TotalItems: totalItems,
		TotalPages: totalPages,
		HasNext:    p.Page < totalPages,
		HasPrev:    p.Page > 1,
	}
}
