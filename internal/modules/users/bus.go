package users

import (
	"context"
	"fmt"

	"github.com/radni/soapbox/internal/core/bus"
	"github.com/radni/soapbox/internal/core/types"
)

func RegisterQueries(b bus.Bus, store *Store) error {
	if err := b.RegisterQuery(QueryGetProfile, handleGetProfile(store)); err != nil {
		return fmt.Errorf("users: register query %s: %w", QueryGetProfile, err)
	}

	if err := b.RegisterQuery(QueryGetProfiles, handleGetProfiles(store)); err != nil {
		return fmt.Errorf("users: register query %s: %w", QueryGetProfiles, err)
	}

	if err := b.RegisterQuery(QueryGetFollowing, handleGetFollowing(store)); err != nil {
		return fmt.Errorf("users: register query %s: %w", QueryGetFollowing, err)
	}

	if err := b.RegisterQuery(QueryGetFollowerIDs, handleGetFollowerIDs(store)); err != nil {
		return fmt.Errorf("users: register query %s: %w", QueryGetFollowerIDs, err)
	}

	return nil
}

func handleGetProfile(store *Store) func(req any) (any, error) {
	return func(req any) (any, error) {
		q, err := bus.Convert[GetProfileQuery](req)
		if err != nil {
			return nil, fmt.Errorf("users: GetProfile: invalid request type: %w", err)
		}

		ctx := context.Background()

		profile, err := store.GetProfileByID(ctx, q.UserID)
		if err != nil {
			return nil, err
		}

		followerCount, err := store.GetFollowerCount(ctx, profile.ID)
		if err != nil {
			return nil, err
		}

		followingCount, err := store.GetFollowingCount(ctx, profile.ID)
		if err != nil {
			return nil, err
		}

		var isFollowing bool
		if q.ViewerID != nil && *q.ViewerID != types.ZeroID {
			isFollowing, err = store.IsFollowing(ctx, *q.ViewerID, profile.ID)
			if err != nil {
				return nil, err
			}
		}

		return profileToResponse(profile, followerCount, followingCount, isFollowing), nil
	}
}

func handleGetProfiles(store *Store) func(req any) (any, error) {
	return func(req any) (any, error) {
		q, err := bus.Convert[GetProfilesQuery](req)
		if err != nil {
			return nil, fmt.Errorf("users: GetProfiles: invalid request type: %w", err)
		}

		ctx := context.Background()

		profiles, err := store.GetProfilesByIDs(ctx, q.UserIDs)
		if err != nil {
			return nil, err
		}

		results := make([]ProfileResponse, len(profiles))
		for i, p := range profiles {
			results[i] = profileToResponse(&p, 0, 0, false)
		}

		return results, nil
	}
}

func handleGetFollowing(store *Store) func(req any) (any, error) {
	return func(req any) (any, error) {
		q, err := bus.Convert[GetFollowingQuery](req)
		if err != nil {
			return nil, fmt.Errorf("users: GetFollowing: invalid request type: %w", err)
		}

		ctx := context.Background()

		ids, err := store.GetFollowingIDs(ctx, q.UserID)
		if err != nil {
			return nil, err
		}

		return ids, nil
	}
}

func handleGetFollowerIDs(store *Store) func(req any) (any, error) {
	return func(req any) (any, error) {
		q, err := bus.Convert[GetFollowerIDsQuery](req)
		if err != nil {
			return nil, fmt.Errorf("users: GetFollowerIDs: invalid request type: %w", err)
		}

		ctx := context.Background()

		ids, err := store.GetFollowerIDs(ctx, q.UserID)
		if err != nil {
			return nil, err
		}

		return ids, nil
	}
}

// profileToResponse converts a Profile model to a ProfileResponse.
// Standalone function (not a Service method) for use by bus query handlers.
func profileToResponse(p *Profile, followerCount, followingCount int, isFollowing bool) ProfileResponse {
	return ProfileResponse{
		ID:             p.ID,
		Username:       p.Username,
		DisplayName:    p.DisplayName,
		Bio:            p.Bio,
		AvatarURL:      p.AvatarURL,
		Verified:       p.Verified,
		FollowerCount:  followerCount,
		FollowingCount: followingCount,
		IsFollowing:    isFollowing,
		CreatedAt:      p.CreatedAt,
	}
}
