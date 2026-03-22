package types

import (
	"net/http"
	"strconv"
)

const (
	DefaultLimit = 20
	MaxLimit     = 100
)

type CursorParams struct {
	Cursor string
	Limit  int
}

type CursorPage[T any] struct {
	Items      []T    `json:"items"`
	NextCursor string `json:"next_cursor,omitempty"`
	HasMore    bool   `json:"has_more"`
}

func ParseCursorParams(r *http.Request) CursorParams {
	cursor := r.URL.Query().Get("cursor")

	limit := DefaultLimit
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	if limit > MaxLimit {
		limit = MaxLimit
	}

	return CursorParams{
		Cursor: cursor,
		Limit:  limit,
	}
}
