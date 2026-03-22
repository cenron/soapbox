package httpkit

import (
	"context"

	"github.com/google/uuid"
)

type contextKey string

const (
	userIDKey   contextKey = "user_id"
	roleKey     contextKey = "role"
	verifiedKey contextKey = "verified"
)

func WithUserID(ctx context.Context, id uuid.UUID) context.Context {
	return context.WithValue(ctx, userIDKey, id)
}

func UserIDFrom(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(userIDKey).(uuid.UUID)
	return id, ok
}

func WithRole(ctx context.Context, role string) context.Context {
	return context.WithValue(ctx, roleKey, role)
}

func RoleFrom(ctx context.Context) (string, bool) {
	role, ok := ctx.Value(roleKey).(string)
	return role, ok
}

func WithVerified(ctx context.Context, verified bool) context.Context {
	return context.WithValue(ctx, verifiedKey, verified)
}

func VerifiedFrom(ctx context.Context) (verified, ok bool) {
	verified, ok = ctx.Value(verifiedKey).(bool)
	return verified, ok
}
