package seek_pagination

import (
	"errors"
	"fmt"
	pagination "go_project_structure/common_pkg/pagination/helper"
	"net/http"
)

type Params struct {
	Limit     int
	Direction Direction
	Cursor    *Cursor // nil means first page
}

// Parse Params extracts and validates pagination params from request
func ParseParams(r *http.Request) ( Params, error) {
	query := r.URL.Query()

	// Limit
	limit := pagination.ParseInt(query.Get("limit"), DefaultLimit)
	if limit < 1 {
		limit = DefaultLimit
	}
	if limit > MaxLimit {
		limit = MaxLimit
	}

	// Direction
	direction := pagination.ParseString(query.Get("direction"), string(DirectionNext))

	if direction != string(DirectionNext) && direction != string(DirectionPrev) {
		return Params{}, errors.New("invalid direction")
	}

	// Cursor
	cursor, err := DecodeCursor(query.Get("cursor"))
	if err != nil {
		return Params{}, fmt.Errorf("cursor: %w", err)
	}

	return  Params{
		Limit:     limit,
		Direction: Direction(direction),
		Cursor:    cursor,
	}, nil
}
