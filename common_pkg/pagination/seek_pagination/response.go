package seek_pagination

import (
	"go_project_structure/common_pkg/pagination/helper"
	"time"
)

// Entity is the constraint every paginated model must satisfy.
// The repo uses GetID / GetCreatedAt to build cursors without knowing the concrete type.
type Entity interface {
	GetID() uint
	GetCreatedAt() time.Time
}

// RailResult is the raw output from the repository — two slices + overflow flags
type RailResult[T Entity] struct {
	NewItems   []T
	OldItems   []T
	HasMoreNew bool
	HasMoreOld bool
}


// Response is the generic HTTP envelope returned to clients
type Response[T Entity] struct {
	Data       []T    `json:"data"`
	NewCount   int    `json:"new_count"`
	TotalNew   int64  `json:"total_new"`
	NextCursor string `json:"next_cursor,omitempty"`
	PrevCursor string `json:"prev_cursor,omitempty"`
	HasNext    bool   `json:"has_next"`
	HasPrev    bool   `json:"has_prev"`
	Limit      int    `json:"limit"`
}


// BuildResponse assembles the Response from raw rail data and builds cursors.
// Called by every handler — no domain knowledge required.
func BuildResponse[T Entity](rail RailResult[T], params Params, totalNew int64) Response[T] {

	merged := make([]T, 0, len(rail.NewItems)+len(rail.OldItems))
	merged = append(merged, rail.NewItems...)
	merged = append(merged, rail.OldItems...)

	resp := Response[T]{
		Data:     merged,
		NewCount: len(rail.NewItems),
		TotalNew: totalNew,
		HasNext:  rail.HasMoreNew || rail.HasMoreOld,
		HasPrev:  params.Cursor != nil && params.Direction == helper.DirectionNext,
		Limit:    params.Limit,
	}

	if len(merged) == 0 {
		return resp
	}

	// Resolve anchor
	var anchorID uint
	var anchorCreatedAt time.Time
	var newRailOffset int

	if params.Cursor == nil {
		// First page: anchor = the first (newest) item returned
		anchorID = merged[0].GetID()
		anchorCreatedAt = merged[0].GetCreatedAt()
	} else {
		anchorID = params.Cursor.AnchorID
		anchorCreatedAt = params.Cursor.AnchorCreatedAt
		newRailOffset = params.Cursor.NewRailOffset + len(rail.NewItems)
	}

	// Next cursor — last item in the merged slice
	if resp.HasNext {
		last := merged[len(merged)-1]
		next := Cursor{
			LastID:          last.GetID(),
			LastCreatedAt:   last.GetCreatedAt(),
			AnchorID:        anchorID,
			AnchorCreatedAt: anchorCreatedAt,
			NewRailOffset:   newRailOffset,
		}
		if token, err := next.EncodeCursor(); err == nil {
			resp.NextCursor = token
		}
	}

	// Prev cursor — first old (history) item
	if resp.HasPrev && len(rail.OldItems) > 0 {
		first := rail.OldItems[0]
		prev := Cursor{
			LastID:          first.GetID(),
			LastCreatedAt:   first.GetCreatedAt(),
			AnchorID:        anchorID,
			AnchorCreatedAt: anchorCreatedAt,
			NewRailOffset:   newRailOffset,
		}
		if token, err := prev.EncodeCursor(); err == nil {
			resp.PrevCursor = token
		}
	}

	return resp
}