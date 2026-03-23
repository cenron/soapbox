package feed

import (
	"context"
	"fmt"

	"github.com/radni/soapbox/internal/core/bus"
	"github.com/radni/soapbox/internal/core/types"
)

func RegisterQueries(b bus.Bus, service *Service) error {
	if err := b.RegisterQuery(QueryGetTimeline, handleGetTimeline(service)); err != nil {
		return fmt.Errorf("feed: register query %s: %w", QueryGetTimeline, err)
	}

	return nil
}

func handleGetTimeline(service *Service) func(req any) (any, error) {
	return func(req any) (any, error) {
		q, err := bus.Convert[GetTimelineQuery](req)
		if err != nil {
			return nil, fmt.Errorf("feed: GetTimeline: invalid request type: %w", err)
		}

		ctx := context.Background()
		params := types.CursorParams{Cursor: q.Cursor, Limit: q.Limit}
		if params.Limit == 0 {
			params.Limit = types.DefaultLimit
		}

		page, err := service.GetTimeline(ctx, q.UserID, q.ViewerID, params)
		if err != nil {
			return nil, err
		}

		return page, nil
	}
}
