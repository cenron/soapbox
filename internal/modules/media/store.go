package media

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/radni/soapbox/internal/core/db"
	"github.com/radni/soapbox/internal/core/types"
)

// Upload is the database model for a media upload.
type Upload struct {
	ID          types.ID  `db:"id"`
	UserID      types.ID  `db:"user_id"`
	FileKey     string    `db:"file_key"`
	ContentType string    `db:"content_type"`
	Size        int64     `db:"size"`
	Status      string    `db:"status"`
	CreatedAt   time.Time `db:"created_at"`
}

type Store struct {
	db *db.DB
}

func NewStore(database *db.DB) *Store {
	return &Store{db: database}
}

func (s *Store) Create(ctx context.Context, upload *Upload) error {
	const q = `INSERT INTO media.uploads (id, user_id, file_key, content_type, size, status, created_at)
	           VALUES (:id, :user_id, :file_key, :content_type, :size, :status, :created_at)`

	if _, err := s.db.Conn.NamedExecContext(ctx, q, upload); err != nil {
		return fmt.Errorf("store: create upload: %w", err)
	}

	return nil
}

func (s *Store) GetByID(ctx context.Context, id types.ID) (*Upload, error) {
	var upload Upload

	const q = `SELECT id, user_id, file_key, content_type, size, status, created_at
	           FROM media.uploads WHERE id = $1`

	if err := s.db.Conn.GetContext(ctx, &upload, q, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, types.NewNotFound("upload")
		}
		return nil, fmt.Errorf("store: get upload: %w", err)
	}

	return &upload, nil
}

func (s *Store) GetByIDs(ctx context.Context, ids []types.ID) ([]Upload, error) {
	if len(ids) == 0 {
		return []Upload{}, nil
	}

	query, args, err := sqlx.In(
		`SELECT id, user_id, file_key, content_type, size, status, created_at
		 FROM media.uploads WHERE id IN (?)`,
		ids,
	)
	if err != nil {
		return nil, fmt.Errorf("store: get uploads by ids: build query: %w", err)
	}

	query = s.db.Conn.Rebind(query)

	var uploads []Upload
	if err := s.db.Conn.SelectContext(ctx, &uploads, query, args...); err != nil {
		return nil, fmt.Errorf("store: get uploads by ids: %w", err)
	}

	return uploads, nil
}

func (s *Store) UpdateStatus(ctx context.Context, id types.ID, status string, size int64) error {
	const q = `UPDATE media.uploads SET status = $1, size = $2 WHERE id = $3`

	res, err := s.db.Conn.ExecContext(ctx, q, status, size, id)
	if err != nil {
		return fmt.Errorf("store: update upload status: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("store: update upload status: rows affected: %w", err)
	}

	if rows == 0 {
		return types.NewNotFound("upload")
	}

	return nil
}
