package posts

import (
	"context"
	"fmt"

	"github.com/radni/soapbox/internal/core/bus"
	"github.com/radni/soapbox/internal/core/types"
)

func RegisterQueries(b bus.Bus, service *Service) error {
	if err := b.RegisterQuery(QueryGetByIDs, handleGetByIDs(service)); err != nil {
		return fmt.Errorf("posts: register query %s: %w", QueryGetByIDs, err)
	}

	if err := b.RegisterQuery(QueryGetByAuthor, handleGetByAuthor(service)); err != nil {
		return fmt.Errorf("posts: register query %s: %w", QueryGetByAuthor, err)
	}

	if err := b.RegisterQuery(QueryGetThread, handleGetThread(service)); err != nil {
		return fmt.Errorf("posts: register query %s: %w", QueryGetThread, err)
	}

	if err := b.RegisterQuery(QuerySearch, handleSearch(service)); err != nil {
		return fmt.Errorf("posts: register query %s: %w", QuerySearch, err)
	}

	if err := b.RegisterQuery(QuerySearchHashtag, handleSearchHashtag(service)); err != nil {
		return fmt.Errorf("posts: register query %s: %w", QuerySearchHashtag, err)
	}

	return nil
}

func handleGetByIDs(service *Service) func(req any) (any, error) {
	return func(req any) (any, error) {
		q, ok := req.(GetByIDsQuery)
		if !ok {
			return nil, fmt.Errorf("posts: GetByIDs: invalid request type")
		}

		ctx := context.Background()
		posts, err := service.store.GetPostsByIDs(ctx, q.PostIDs)
		if err != nil {
			return nil, err
		}

		responses, err := service.enrichPosts(ctx, posts, q.ViewerID)
		if err != nil {
			return nil, err
		}
		return responses, nil
	}
}

func handleGetByAuthor(service *Service) func(req any) (any, error) {
	return func(req any) (any, error) {
		q, ok := req.(GetByAuthorQuery)
		if !ok {
			return nil, fmt.Errorf("posts: GetByAuthor: invalid request type")
		}

		ctx := context.Background()
		params := types.CursorParams{Cursor: q.Cursor, Limit: q.Limit}
		if params.Limit == 0 {
			params.Limit = types.DefaultLimit
		}

		page, err := service.GetPostsByAuthor(ctx, q.AuthorID, q.ViewerID, params)
		if err != nil {
			return nil, err
		}
		return page, nil
	}
}

func handleGetThread(service *Service) func(req any) (any, error) {
	return func(req any) (any, error) {
		q, ok := req.(GetThreadQuery)
		if !ok {
			return nil, fmt.Errorf("posts: GetThread: invalid request type")
		}

		ctx := context.Background()
		params := types.CursorParams{Limit: 100}

		page, err := service.GetReplies(ctx, q.RootPostID, q.ViewerID, params)
		if err != nil {
			return nil, err
		}
		return page, nil
	}
}

func handleSearch(service *Service) func(req any) (any, error) {
	return func(req any) (any, error) {
		q, ok := req.(SearchPostsQuery)
		if !ok {
			return nil, fmt.Errorf("posts: Search: invalid request type")
		}

		ctx := context.Background()
		params := types.CursorParams{Cursor: q.Cursor, Limit: q.Limit}
		if params.Limit == 0 {
			params.Limit = types.DefaultLimit
		}

		page, err := service.SearchPosts(ctx, q.Q, q.ViewerID, params)
		if err != nil {
			return nil, err
		}
		return page, nil
	}
}

func handleSearchHashtag(service *Service) func(req any) (any, error) {
	return func(req any) (any, error) {
		q, ok := req.(SearchHashtagQuery)
		if !ok {
			return nil, fmt.Errorf("posts: SearchHashtag: invalid request type")
		}

		ctx := context.Background()
		params := types.CursorParams{Cursor: q.Cursor, Limit: q.Limit}
		if params.Limit == 0 {
			params.Limit = types.DefaultLimit
		}

		page, err := service.SearchByHashtag(ctx, q.Tag, q.ViewerID, params)
		if err != nil {
			return nil, err
		}
		return page, nil
	}
}
