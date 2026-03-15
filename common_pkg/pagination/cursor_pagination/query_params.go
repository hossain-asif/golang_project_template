package cursor_pagination

import (
	"errors"
	"fmt"
	"go_project_structure/common_pkg/pagination/helper"
	"net/http"
)

// Params are the parsed, validated pagination inputs from a request
type Params struct {
	Limit     int
	Cursor    *Cursor
	Direction helper.Direction
}

func ParseParams(r *http.Request) (Params, error) {
	query := r.URL.Query()

	limit := helper.ParseInt(query.Get("limit"), helper.DefaultLimit)
	if limit < 1 {
		limit = helper.DefaultLimit
	}
	if limit > helper.MaxLimit {
		limit = helper.MaxLimit
	}

	direction := helper.ParseString(query.Get("direction"), string(helper.DirectionNext))

	if direction != string(helper.DirectionNext) && direction != string(helper.DirectionPrev) {
		return Params{}, errors.New("invalid direction")
	}

	cursor, err := DecodeCursor(query.Get("cursor"))
	if err != nil {
		return Params{}, fmt.Errorf("cursor: %w", err)
	}

	return Params{
		Limit:     limit,
		Cursor:    cursor,
		Direction: helper.Direction(direction),
	}, nil
}
