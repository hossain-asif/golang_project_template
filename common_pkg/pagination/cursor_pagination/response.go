package cursor_pagination

import (
	"go_project_structure/common_pkg/pagination/helper"
	"time"
)

// Cursorable must be implemented by any model you want to paginate
type Cursorable interface {
	GetID() uint
	GetCreatedAt() time.Time
}

// PageInfo is returned alongside results so clients know what cursors to use
type PageInfo struct {
	HasNextPage bool    `json:"has_next_page"`
	HasPrevPage bool    `json:"has_prev_page"`
	StartCursor *string `json:"start_cursor"`
	EndCursor   *string `json:"end_cursor"`
	TotalCount  int64   `json:"total_count,omitempty"` // optional — expensive on large tables
}

// buildPageInfo is unexported — use BuildPage
func buildPageInfo[T Cursorable](data []T, params Params, hasMore bool) (PageInfo, error) {
	if len(data) == 0 {
		return PageInfo{}, nil
	}

	first, last := data[0], data[len(data)-1]

	startEncoded, err := (&Cursor{ID: first.GetID(), CreatedAt: first.GetCreatedAt()}).EncodeCursor()
	if err != nil {
		return PageInfo{}, err
	}
	endEncoded, err := (&Cursor{ID: last.GetID(), CreatedAt: last.GetCreatedAt()}).EncodeCursor()
	if err != nil {
		return PageInfo{}, err
	}

	info := PageInfo{
		StartCursor: &startEncoded,
		EndCursor:   &endEncoded,
	}

	switch params.Direction {
	case helper.DirectionNext:
		info.HasNextPage = hasMore
		info.HasPrevPage = params.Cursor != nil
	case helper.DirectionPrev:
		info.HasPrevPage = hasMore
		info.HasNextPage = params.Cursor != nil
	}

	return info, nil
}


// Page is the generic return type for any paginated list
type Page[T Cursorable] struct {
	Data     []T
	PageInfo PageInfo
}

// BuildPage slices the extra row, reverses if needed, and builds PageInfo.
// Call this from any repository — no domain-specific code needed.
func BuildPage[T Cursorable](data []T, params Params, hasMore bool) (Page[T], error) {
	// Reverse for prev-direction so callers always get newest-first
	if params.Direction == helper.DirectionPrev {
		for i, j := 0, len(data)-1; i < j; i, j = i+1, j-1 {
			data[i], data[j] = data[j], data[i]
		}
	}

	pageInfo, err := buildPageInfo(data, params, hasMore)
	if err != nil {
		return Page[T]{}, err
	}

	return Page[T]{Data: data, PageInfo: pageInfo}, nil
}