package seek_pagination

import (
	"errors"
	"fmt"
	"go_project_structure/common_pkg/pagination/helper"
	"net/http"
)

type Params struct {
	Limit     int
	Direction helper.Direction
	Cursor    *Cursor // nil means first page
}

// Parse Params extracts and validates pagination params from request
func ParseParams(r *http.Request) ( Params, error) {
	query := r.URL.Query()

	// Limit
	limit := helper.ParseInt(query.Get("limit"), helper.DefaultLimit)
	if limit < 1 {
		limit = helper.DefaultLimit
	}
	if limit > helper.MaxLimit {
		limit = helper.MaxLimit
	}

	// Direction
	direction := helper.ParseString(query.Get("direction"), string(helper.DirectionNext))

	if direction != string(helper.DirectionNext) && direction != string(helper.DirectionPrev) {
		return Params{}, errors.New("invalid direction")
	}

	// Cursor
	cursor, err := DecodeCursor(query.Get("cursor"))
	if err != nil {
		return Params{}, fmt.Errorf("cursor: %w", err)
	}

	return  Params{
		Limit:     limit,
		Direction: helper.Direction(direction),
		Cursor:    cursor,
	}, nil
}
