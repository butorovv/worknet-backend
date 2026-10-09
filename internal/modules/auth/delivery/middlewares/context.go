package middlewares

import (
	"context"

	"github.com/butorovv/worknet-backend/internal/modules/auth/usecase"
)

type contextKey string

const (
	currentUserContextKey contextKey = "auth.current_user"
	sessionIDContextKey   contextKey = "auth.session_id"
)

func withAuthContext(ctx context.Context, sessionID string, user usecase.CurrentUser) context.Context {
	ctx = context.WithValue(ctx, sessionIDContextKey, sessionID)
	return context.WithValue(ctx, currentUserContextKey, user)
}

func CurrentUserFromContext(ctx context.Context) (usecase.CurrentUser, bool) {
	user, ok := ctx.Value(currentUserContextKey).(usecase.CurrentUser)
	return user, ok
}

func SessionIDFromContext(ctx context.Context) (string, bool) {
	sessionID, ok := ctx.Value(sessionIDContextKey).(string)
	return sessionID, ok
}
