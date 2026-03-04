package offsetpagination

import (
	"go_project_structure/common_pkg/pagination"
	"net/http"
)

// Params holds parsed pagination query parameters
type Params struct {
	Page   int `json:"page"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

// Parse extracts and validates pagination params from request query string.
// Accepts ?page=1&limit=20
func Parse(r *http.Request) Params {
	q := r.URL.Query()

	page := pagination.ParseInt(q.Get("page"), DefaultPage)
	limit := pagination.ParseInt(q.Get("limit"), DefaultLimit)

	// Clamp values to safe ranges
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = DefaultLimit
	}
	if limit > MaxLimit {
		limit = MaxLimit
	}

	offset := (page - 1) * limit

	return Params{
		Page:   page,
		Limit:  limit,
		Offset: offset,
	}
}
