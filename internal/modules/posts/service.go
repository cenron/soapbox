package posts

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/radni/soapbox/internal/core/bus"
	"github.com/radni/soapbox/internal/core/db"
	"github.com/radni/soapbox/internal/core/types"
)

const maxBodyLength = 280
const maxMediaPerPost = 4

type Service struct {
	db     *db.DB
	store  *Store
	bus    bus.Bus
	logger *slog.Logger
}

func NewService(database *db.DB, store *Store, b bus.Bus, logger *slog.Logger) *Service {
	return &Service{
		db:     database,
		store:  store,
		bus:    b,
		logger: logger,
	}
}

// CreatePost creates a new post with optional media, link preview, and hashtags.
func (s *Service) CreatePost(ctx context.Context, authorID types.ID, req CreatePostRequest) (*PostResponse, error) {
	if err := validateCreatePostRequest(req); err != nil {
		return nil, err
	}

	if req.RepostOfID != nil {
		return nil, types.NewValidation("use the repost endpoint to create reposts")
	}

	author, err := s.getAuthorProfile(authorID)
	if err != nil {
		return nil, fmt.Errorf("service: create post: get author: %w", err)
	}

	now := time.Now().UTC()
	postID, err := types.NewID()
	if err != nil {
		return nil, fmt.Errorf("service: create post: generate id: %w", err)
	}

	post := &Post{
		ID:                postID,
		AuthorID:          authorID,
		AuthorUsername:    author.Username,
		AuthorDisplayName: author.DisplayName,
		AuthorAvatarURL:   author.AvatarURL,
		AuthorVerified:    author.Verified,
		Body:              req.Body,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	if req.ParentID != nil {
		parentID, err := types.ParseID(*req.ParentID)
		if err != nil {
			return nil, types.NewValidation("invalid parent_id")
		}

		if _, err := s.store.GetPostByID(ctx, parentID); err != nil {
			return nil, types.NewValidation("parent post does not exist")
		}

		post.ParentID = &parentID
	}

	mediaItems, err := s.resolveMedia(req.MediaIDs)
	if err != nil {
		return nil, err
	}

	hashtags := extractHashtags(req.Body)

	var preview *linkPreviewData
	if rawURL := extractFirstURL(req.Body); rawURL != "" && len(mediaItems) == 0 {
		preview = fetchLinkPreview(ctx, rawURL)
	}

	var linkPreviewRecord *LinkPreview
	var postMediaRecords []PostMedia

	err = s.db.WithTx(ctx, func(tx *sqlx.Tx) error {
		if err := s.store.CreatePost(ctx, tx, post); err != nil {
			return err
		}

		if post.ParentID != nil {
			if err := s.store.IncrementReplyCount(ctx, tx, *post.ParentID); err != nil {
				return err
			}
		}

		for i, item := range mediaItems {
			pmID, err := types.NewID()
			if err != nil {
				return fmt.Errorf("generate media id: %w", err)
			}
			pm := &PostMedia{
				ID:        pmID,
				PostID:    postID,
				MediaURL:  item.URL,
				MediaType: item.ContentType,
				Position:  i,
			}
			if err := s.store.CreatePostMedia(ctx, tx, pm); err != nil {
				return err
			}
			postMediaRecords = append(postMediaRecords, *pm)
		}

		if err := s.store.CreateHashtags(ctx, tx, postID, hashtags); err != nil {
			return err
		}

		if preview != nil {
			lpID, err := types.NewID()
			if err != nil {
				return fmt.Errorf("generate link preview id: %w", err)
			}
			linkPreviewRecord = &LinkPreview{
				ID:          lpID,
				PostID:      postID,
				URL:         preview.URL,
				Title:       preview.Title,
				Description: preview.Description,
				ImageURL:    preview.ImageURL,
			}
			if err := s.store.CreateLinkPreview(ctx, tx, linkPreviewRecord); err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	if err := s.bus.Publish(TopicCreated, PostCreatedEvent{
		PostID:         postID,
		AuthorID:       authorID,
		AuthorUsername: author.Username,
		Body:           req.Body,
		ParentID:       post.ParentID,
		CreatedAt:      now,
	}); err != nil {
		s.logger.Warn("service: create post: publish event failed", "error", err)
	}

	resp := postToResponse(post, postMediaRecords, linkPreviewRecord, hashtags, false, false)
	return &resp, nil
}

// GetPost retrieves a single post with all its related data.
func (s *Service) GetPost(ctx context.Context, id types.ID, viewerID *types.ID) (*PostResponse, error) {
	post, err := s.store.GetPostByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.enrichPost(ctx, post, viewerID)
}

// DeletePost deletes a post if the caller is the author.
func (s *Service) DeletePost(ctx context.Context, postID, callerID types.ID) error {
	post, err := s.store.GetPostByID(ctx, postID)
	if err != nil {
		return err
	}

	if post.AuthorID != callerID {
		return types.ErrForbidden()
	}

	parentID := post.ParentID
	repostOfID := post.RepostOfID
	authorID := post.AuthorID

	if err := s.store.DeletePost(ctx, postID); err != nil {
		return err
	}

	if parentID != nil {
		if err := s.store.DecrementReplyCount(ctx, *parentID); err != nil {
			s.logger.Warn("service: delete post: decrement reply count failed", "error", err)
		}
	}

	if repostOfID != nil {
		if _, err := s.store.DecrementRepostCount(ctx, *repostOfID); err != nil {
			s.logger.Warn("service: delete post: decrement repost count failed", "error", err)
		}
	}

	if err := s.bus.Publish(TopicDeleted, PostDeletedEvent{
		PostID:    postID,
		AuthorID:  authorID,
		DeletedAt: time.Now().UTC(),
	}); err != nil {
		s.logger.Warn("service: delete post: publish event failed", "error", err)
	}

	return nil
}

// GetReplies returns a cursor-paginated list of replies to a post.
func (s *Service) GetReplies(ctx context.Context, parentID types.ID, viewerID *types.ID, params types.CursorParams) (*types.CursorPage[PostResponse], error) {
	if _, err := s.store.GetPostByID(ctx, parentID); err != nil {
		return nil, err
	}

	posts, hasMore, err := s.store.GetReplies(ctx, parentID, params)
	if err != nil {
		return nil, err
	}

	return s.buildPostCursorPage(ctx, posts, hasMore, viewerID)
}

// GetPostsByAuthor returns a cursor-paginated list of root posts by an author.
func (s *Service) GetPostsByAuthor(ctx context.Context, authorID types.ID, viewerID *types.ID, params types.CursorParams) (*types.CursorPage[PostResponse], error) {
	posts, hasMore, err := s.store.GetPostsByAuthor(ctx, authorID, params)
	if err != nil {
		return nil, err
	}

	return s.buildPostCursorPage(ctx, posts, hasMore, viewerID)
}

// LikePost adds a like and returns updated counts.
func (s *Service) LikePost(ctx context.Context, postID, userID types.ID) (*LikeResponse, error) {
	post, err := s.store.GetPostByID(ctx, postID)
	if err != nil {
		return nil, err
	}

	var count int

	err = s.db.WithTx(ctx, func(tx *sqlx.Tx) error {
		if err := s.store.CreateLikeTx(ctx, tx, postID, userID); err != nil {
			return err
		}
		count, err = s.store.IncrementLikeCountTx(ctx, tx, postID)
		return err
	})
	if err != nil {
		return nil, err
	}

	if err := s.bus.Publish(TopicLiked, PostLikedEvent{
		PostID:    postID,
		UserID:    userID,
		AuthorID:  post.AuthorID,
		LikeCount: count,
		CreatedAt: time.Now().UTC(),
	}); err != nil {
		s.logger.Warn("service: like post: publish event failed", "error", err)
	}

	return &LikeResponse{
		PostID:    postID,
		LikeCount: count,
		LikedByMe: true,
	}, nil
}

// UnlikePost removes a like and returns updated counts.
func (s *Service) UnlikePost(ctx context.Context, postID, userID types.ID) (*LikeResponse, error) {
	var count int

	err := s.db.WithTx(ctx, func(tx *sqlx.Tx) error {
		if err := s.store.DeleteLikeTx(ctx, tx, postID, userID); err != nil {
			return err
		}
		var err error
		count, err = s.store.DecrementLikeCountTx(ctx, tx, postID)
		return err
	})
	if err != nil {
		return nil, err
	}

	return &LikeResponse{
		PostID:    postID,
		LikeCount: count,
		LikedByMe: false,
	}, nil
}

// RepostPost creates a repost and returns updated counts.
func (s *Service) RepostPost(ctx context.Context, postID, userID types.ID) (*RepostResponse, error) {
	original, err := s.store.GetPostByID(ctx, postID)
	if err != nil {
		return nil, err
	}

	if original.RepostOfID != nil {
		return nil, types.NewValidation("cannot repost a repost")
	}

	existing, _ := s.store.IsRepostedByUser(ctx, postID, userID)
	if existing {
		return nil, types.NewConflict("already reposted")
	}

	author, err := s.getAuthorProfile(userID)
	if err != nil {
		return nil, fmt.Errorf("service: repost: get author: %w", err)
	}

	now := time.Now().UTC()
	repostID, err := types.NewID()
	if err != nil {
		return nil, fmt.Errorf("service: repost: generate id: %w", err)
	}

	repost := &Post{
		ID:                repostID,
		AuthorID:          userID,
		AuthorUsername:    author.Username,
		AuthorDisplayName: author.DisplayName,
		AuthorAvatarURL:   author.AvatarURL,
		AuthorVerified:    author.Verified,
		Body:              original.Body,
		RepostOfID:        &postID,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	var count int

	err = s.db.WithTx(ctx, func(tx *sqlx.Tx) error {
		if err := s.store.CreatePost(ctx, tx, repost); err != nil {
			return err
		}
		count, err = s.store.IncrementRepostCountTx(ctx, tx, postID)
		return err
	})
	if err != nil {
		return nil, err
	}

	if err := s.bus.Publish(TopicReposted, PostRepostedEvent{
		PostID:      postID,
		UserID:      userID,
		AuthorID:    original.AuthorID,
		RepostCount: count,
		CreatedAt:   now,
	}); err != nil {
		s.logger.Warn("service: repost: publish event failed", "error", err)
	}

	return &RepostResponse{
		PostID:       postID,
		RepostCount:  count,
		RepostedByMe: true,
	}, nil
}

// UndoRepost removes a repost and returns updated counts.
func (s *Service) UndoRepost(ctx context.Context, postID, userID types.ID) (*RepostResponse, error) {
	repost, err := s.store.GetRepostByUser(ctx, postID, userID)
	if err != nil {
		return nil, err
	}

	if err := s.store.DeletePost(ctx, repost.ID); err != nil {
		return nil, err
	}

	count, err := s.store.DecrementRepostCount(ctx, postID)
	if err != nil {
		return nil, err
	}

	return &RepostResponse{
		PostID:       postID,
		RepostCount:  count,
		RepostedByMe: false,
	}, nil
}

// GetUserPosts returns root posts by a given username.
func (s *Service) GetUserPosts(ctx context.Context, username string, viewerID *types.ID, params types.CursorParams) (*types.CursorPage[PostResponse], error) {
	posts, hasMore, err := s.store.GetPostsByUsername(ctx, username, params)
	if err != nil {
		return nil, err
	}

	return s.buildPostCursorPage(ctx, posts, hasMore, viewerID)
}

// HandleProfileUpdated syncs denormalized author fields when a user updates their profile.
func (s *Service) HandleProfileUpdated(event userProfileUpdatedEvent) {
	ctx := context.Background()

	if err := s.store.UpdateAuthorDenorm(ctx, event.UserID, event.Username, event.DisplayName, event.AvatarURL, event.Verified); err != nil {
		s.logger.Error("service: handle profile updated: update denorm failed",
			"user_id", event.UserID,
			"error", err,
		)
	}
}

// --- helpers ---

func (s *Service) getAuthorProfile(userID types.ID) (*userProfileResponse, error) {
	result, err := s.bus.Query(usersQueryGetProfile, userGetProfileQuery{
		UserID: userID,
	})
	if err != nil {
		return nil, err
	}

	profile, err := bus.Convert[userProfileResponse](result)
	if err != nil {
		return nil, fmt.Errorf("service: convert profile response: %w", err)
	}
	return &profile, nil
}

type resolvedMedia struct {
	URL         string
	ContentType string
}

func (s *Service) resolveMedia(mediaIDs []string) ([]resolvedMedia, error) {
	if len(mediaIDs) == 0 {
		return nil, nil
	}

	if len(mediaIDs) > maxMediaPerPost {
		return nil, types.NewValidation(fmt.Sprintf("maximum %d images per post", maxMediaPerPost))
	}

	ids := make([]types.ID, len(mediaIDs))
	for i, raw := range mediaIDs {
		id, err := types.ParseID(raw)
		if err != nil {
			return nil, types.NewValidation(fmt.Sprintf("invalid media_id: %s", raw))
		}
		ids[i] = id
	}

	result, err := s.bus.Query(mediaQueryGetByIDs, mediaGetByIDsQuery{IDs: ids})
	if err != nil {
		return nil, fmt.Errorf("service: resolve media: %w", err)
	}

	uploads, err := bus.Convert[[]mediaUploadResponse](result)
	if err != nil {
		return nil, fmt.Errorf("service: resolve media: convert response: %w", err)
	}

	if len(uploads) != len(ids) {
		return nil, types.NewValidation("one or more media_ids not found")
	}

	items := make([]resolvedMedia, len(uploads))
	for i, u := range uploads {
		items[i] = resolvedMedia{
			URL:         u.URL,
			ContentType: u.ContentType,
		}
	}
	return items, nil
}

func (s *Service) enrichPost(ctx context.Context, post *Post, viewerID *types.ID) (*PostResponse, error) {
	postMedia, err := s.store.GetMediaByPostID(ctx, post.ID)
	if err != nil {
		return nil, err
	}

	linkPreview, err := s.store.GetLinkPreviewByPostID(ctx, post.ID)
	if err != nil {
		return nil, err
	}

	hashtags, err := s.store.GetHashtagsByPostID(ctx, post.ID)
	if err != nil {
		return nil, err
	}

	var likedByMe, repostedByMe bool
	if viewerID != nil {
		likedByMe, err = s.store.IsLikedByUser(ctx, post.ID, *viewerID)
		if err != nil {
			return nil, err
		}
		repostedByMe, err = s.store.IsRepostedByUser(ctx, post.ID, *viewerID)
		if err != nil {
			return nil, err
		}
	}

	resp := postToResponse(post, postMedia, linkPreview, hashtags, likedByMe, repostedByMe)
	return &resp, nil
}

func (s *Service) enrichPosts(ctx context.Context, posts []Post, viewerID *types.ID) ([]PostResponse, error) {
	if len(posts) == 0 {
		return []PostResponse{}, nil
	}

	ids := make([]types.ID, len(posts))
	for i := range posts {
		ids[i] = posts[i].ID
	}

	mediaMap, err := s.store.GetMediaByPostIDs(ctx, ids)
	if err != nil {
		return nil, err
	}

	previewMap, err := s.store.GetLinkPreviewsByPostIDs(ctx, ids)
	if err != nil {
		return nil, err
	}

	hashtagMap, err := s.store.GetHashtagsByPostIDs(ctx, ids)
	if err != nil {
		return nil, err
	}

	var likedMap, repostedMap map[types.ID]bool
	if viewerID != nil {
		likedMap, err = s.store.IsLikedByUserBatch(ctx, ids, *viewerID)
		if err != nil {
			return nil, err
		}
		repostedMap, err = s.store.IsRepostedByUserBatch(ctx, ids, *viewerID)
		if err != nil {
			return nil, err
		}
	}

	responses := make([]PostResponse, len(posts))
	for i := range posts {
		p := &posts[i]
		responses[i] = postToResponse(
			p,
			mediaMap[p.ID],
			previewMap[p.ID],
			hashtagMap[p.ID],
			likedMap[p.ID],
			repostedMap[p.ID],
		)
	}
	return responses, nil
}

func (s *Service) buildPostCursorPage(ctx context.Context, posts []Post, hasMore bool, viewerID *types.ID) (*types.CursorPage[PostResponse], error) {
	responses, err := s.enrichPosts(ctx, posts, viewerID)
	if err != nil {
		return nil, err
	}

	var nextCursor string
	if hasMore && len(posts) > 0 {
		nextCursor = posts[len(posts)-1].CreatedAt.UTC().Format(time.RFC3339Nano)
	}

	return &types.CursorPage[PostResponse]{
		Items:      responses,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, nil
}

func validateCreatePostRequest(req CreatePostRequest) error {
	if req.Body == "" {
		return types.NewValidation("body is required")
	}
	if len([]rune(req.Body)) > maxBodyLength {
		return types.NewValidation(fmt.Sprintf("body must be at most %d characters", maxBodyLength))
	}
	return nil
}

func postToResponse(p *Post, postMedia []PostMedia, lp *LinkPreview, hashtags []string, likedByMe, repostedByMe bool) PostResponse {
	mediaResp := make([]MediaResponse, len(postMedia))
	for i, m := range postMedia {
		mediaResp[i] = MediaResponse{
			ID:        m.ID,
			MediaURL:  m.MediaURL,
			MediaType: m.MediaType,
			Position:  m.Position,
		}
	}

	var linkPreviewResp *LinkPreviewResponse
	if lp != nil {
		linkPreviewResp = &LinkPreviewResponse{
			URL:         lp.URL,
			Title:       lp.Title,
			Description: lp.Description,
			ImageURL:    lp.ImageURL,
		}
	}

	if hashtags == nil {
		hashtags = []string{}
	}

	return PostResponse{
		ID:                p.ID,
		AuthorID:          p.AuthorID,
		AuthorUsername:    p.AuthorUsername,
		AuthorDisplayName: p.AuthorDisplayName,
		AuthorAvatarURL:   p.AuthorAvatarURL,
		AuthorVerified:    p.AuthorVerified,
		Body:              p.Body,
		ParentID:          p.ParentID,
		RepostOfID:        p.RepostOfID,
		Media:             mediaResp,
		LinkPreview:       linkPreviewResp,
		Hashtags:          hashtags,
		LikeCount:         p.LikeCount,
		RepostCount:       p.RepostCount,
		ReplyCount:        p.ReplyCount,
		LikedByMe:         likedByMe,
		RepostedByMe:      repostedByMe,
		CreatedAt:         p.CreatedAt,
	}
}
