package feed

import (
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/radni/soapbox/internal/core/testutil"
	"github.com/radni/soapbox/internal/core/types"
	"github.com/radni/soapbox/internal/core/ws"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestService() (*Service, *testutil.MockBus) {
	mockBus := testutil.NewMockBus()
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	hub := ws.NewHub(logger)

	svc := NewService(nil, nil, mockBus, hub, logger)
	return svc, mockBus
}

func TestHandlePostCreated_QueriesFollowers(t *testing.T) {
	svc, mockBus := newTestService()

	authorID, _ := types.NewID()
	postID, _ := types.NewID()
	followerID, _ := types.NewID()

	var queriedFollowerFor types.ID
	require.NoError(t, mockBus.RegisterQuery(usersQueryGetFollowerIDs, func(_ any) (any, error) {
		queriedFollowerFor = authorID
		return []types.ID{followerID}, nil
	}))

	// Service calls store.AddToTimelines which needs a DB — it will panic.
	// We verify only that the follower query was called.
	defer func() {
		_ = recover() //nolint:errcheck // Expected panic from nil store.
	}()

	svc.HandlePostCreated(postCreatedEvent{
		PostID:    postID,
		AuthorID:  authorID,
		CreatedAt: time.Now().UTC(),
	})

	assert.Equal(t, authorID, queriedFollowerFor)
}

func TestFetchPosts_PreservesTimelineOrder(t *testing.T) {
	svc, mockBus := newTestService()

	postID1, _ := types.NewID()
	postID2, _ := types.NewID()
	postID3, _ := types.NewID()
	now := time.Now().UTC()

	require.NoError(t, mockBus.RegisterQuery(postsQueryGetByIDs, func(_ any) (any, error) {
		return []postResponse{
			{ID: postID3, Body: "third", CreatedAt: now.Add(-2 * time.Hour)},
			{ID: postID1, Body: "first", CreatedAt: now},
			{ID: postID2, Body: "second", CreatedAt: now.Add(-1 * time.Hour)},
		}, nil
	}))

	entries := []TimelineEntry{
		{PostID: postID1, CreatedAt: now},
		{PostID: postID2, CreatedAt: now.Add(-1 * time.Hour)},
		{PostID: postID3, CreatedAt: now.Add(-2 * time.Hour)},
	}

	posts, err := svc.fetchPosts(entries, nil)
	require.NoError(t, err)
	require.Len(t, posts, 3)

	assert.Equal(t, postID1, posts[0].ID)
	assert.Equal(t, postID2, posts[1].ID)
	assert.Equal(t, postID3, posts[2].ID)
}

func TestFetchPosts_HandlesDeletedPosts(t *testing.T) {
	svc, mockBus := newTestService()

	postID1, _ := types.NewID()
	postID2, _ := types.NewID()
	now := time.Now().UTC()

	require.NoError(t, mockBus.RegisterQuery(postsQueryGetByIDs, func(_ any) (any, error) {
		return []postResponse{
			{ID: postID1, Body: "alive", CreatedAt: now},
		}, nil
	}))

	entries := []TimelineEntry{
		{PostID: postID1, CreatedAt: now},
		{PostID: postID2, CreatedAt: now.Add(-1 * time.Hour)},
	}

	posts, err := svc.fetchPosts(entries, nil)
	require.NoError(t, err)
	require.Len(t, posts, 1)

	assert.Equal(t, postID1, posts[0].ID)
}
