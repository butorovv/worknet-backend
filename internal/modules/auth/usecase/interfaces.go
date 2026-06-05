package usecase

import (
	"context"
	"time"

	"github.com/butorovv/worknet-backend/internal/modules/auth/domain"
)

type AuthRepository interface {
	CreatePasswordUser(ctx context.Context, username, email, passwordHash string) (domain.User, error)
	FindByEmail(ctx context.Context, email string) (domain.User, error)
	FindByID(ctx context.Context, id domain.UserID) (domain.User, error)
	PasswordHashByUserID(ctx context.Context, userID domain.UserID) (string, error)
	TouchPasswordIdentity(ctx context.Context, userID domain.UserID, usedAt time.Time) error
}

type SessionStore interface {
	Create(ctx context.Context, session domain.Session) error
	Get(ctx context.Context, sessionID string) (domain.Session, error)
	Delete(ctx context.Context, sessionID string) error
}

type PasswordHasher interface {
	Hash(ctx context.Context, password string) (string, error)
	Compare(ctx context.Context, hash, password string) error
}

type SessionTokenGenerator interface {
	Generate(ctx context.Context) (string, error)
}

type Clock interface {
	Now() time.Time
}

type RealClock struct{}

func (RealClock) Now() time.Time {
	return time.Now().UTC()
}
