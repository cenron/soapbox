package types

import "github.com/google/uuid"

type ID = uuid.UUID

var ZeroID = uuid.Nil

func NewID() (ID, error) {
	return uuid.NewV7()
}

func ParseID(s string) (ID, error) {
	return uuid.Parse(s)
}
