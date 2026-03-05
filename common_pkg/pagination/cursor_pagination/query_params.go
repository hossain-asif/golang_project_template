package cursor_pagination

import (
	"errors"
	"fmt"
	pagination "go_project_structure/common_pkg/pagination/helper"
	"net/http"
)

// Params are the parsed, validated pagination inputs from a request
type Params struct {
	Limit     int
	Cursor    *Cursor
	Direction Direction
}

func ParseParams(r *http.Request) (Params, error) {
	query := r.URL.Query()

	limit := pagination.ParseInt(query.Get("limit"), DefaultLimit)
	if limit < 1 {
		limit = DefaultLimit
	}
	if limit > MaxLimit {
		limit = MaxLimit
	}

	direction := pagination.ParseString(query.Get("direction"), string(DirectionNext))

	if direction != string(DirectionNext) && direction != string(DirectionPrev) {
		return Params{}, errors.New("invalid direction")
	}

	cursor, err := DecodeCursor(query.Get("cursor"))
	if err != nil {
		return Params{}, fmt.Errorf("cursor: %w", err)
	}

	return Params{
		Limit:     limit,
		Cursor:    cursor,
		Direction: Direction(direction),
	}, nil
}
