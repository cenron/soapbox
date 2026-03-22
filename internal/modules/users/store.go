package users

import "github.com/radni/soapbox/internal/core/db"

type Store struct {
	db *db.DB
}

func NewStore(database *db.DB) *Store {
	return &Store{db: database}
}
