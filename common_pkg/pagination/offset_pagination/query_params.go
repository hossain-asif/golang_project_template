package offset_pagination

import (
	"go_project_structure/common_pkg/pagination/helper"
	"net/http"
)

// Params holds parsed pagination query parameters
type Params struct {
	Page   int `json:"page"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

// ParseParams extracts and validates pagination params from request query string.
// Accepts ?page=1&limit=20
func ParseParams(r *http.Request) Params {
	q := r.URL.Query()

	page := helper.ParseInt(q.Get("page"), helper.DefaultPage)
	limit := helper.ParseInt(q.Get("limit"), helper.DefaultLimit)

	// Clamp values to safe ranges
	if page < 1 {
		page = helper.DefaultPage
	}
	if limit < 1 {
		limit = helper.DefaultLimit
	}
	if limit > helper.MaxLimit {
		limit = helper.MaxLimit
	}

	offset := (page - 1) * limit

	return Params{
		Page:   page,
		Limit:  limit,
		Offset: offset,
	}
}
