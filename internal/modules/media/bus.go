package media

import (
	"context"
	"fmt"

	"github.com/radni/soapbox/internal/core/bus"
)

func RegisterQueries(b bus.Bus, service *Service) error {
	if err := b.RegisterQuery(QueryGetByIDs, handleGetByIDs(service)); err != nil {
		return fmt.Errorf("media: register query %s: %w", QueryGetByIDs, err)
	}

	return nil
}

func handleGetByIDs(service *Service) func(req any) (any, error) {
	return func(req any) (any, error) {
		q, ok := req.(GetByIDsQuery)
		if !ok {
			return nil, fmt.Errorf("media: GetByIDs: invalid request type")
		}

		ctx := context.Background()

		return service.GetByIDs(ctx, q.IDs)
	}
}
