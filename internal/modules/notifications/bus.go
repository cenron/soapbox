package notifications

import (
	"context"
	"fmt"

	"github.com/radni/soapbox/internal/core/bus"
	"github.com/radni/soapbox/internal/core/types"
)

func RegisterQueries(b bus.Bus, service *Service) error {
	if err := b.RegisterQuery(QueryGetForUser, handleGetForUser(service)); err != nil {
		return fmt.Errorf("notifications: register query %s: %w", QueryGetForUser, err)
	}

	return nil
}

func handleGetForUser(service *Service) func(req any) (any, error) {
	return func(req any) (any, error) {
		q, err := bus.Convert[GetForUserQuery](req)
		if err != nil {
			return nil, fmt.Errorf("notifications: GetForUser: invalid request type: %w", err)
		}

		ctx := context.Background()
		params := types.CursorParams{Cursor: q.Cursor, Limit: q.Limit}
		if params.Limit == 0 {
			params.Limit = types.DefaultLimit
		}

		page, err := service.ListNotifications(ctx, q.UserID, params)
		if err != nil {
			return nil, err
		}

		return page, nil
	}
}
