package usecase

import (
	"time"

	"github.com/butorovv/worknet-backend/internal/modules/auth/domain"
)

type RegisterCommand struct {
	Username string
	Email    string
	Password string
}

type LoginCommand struct {
	Email    string
	Password string
}

type LogoutCommand struct {
	SessionID string
}

type ValidateSessionCommand struct {
	SessionID string
}

type CurrentUser struct {
	ID       int64
	Username string
	Email    string
	Status   string
}

type AuthResult struct {
	User      CurrentUser
	SessionID string
	ExpiresAt time.Time
}

type CreatePasswordUserParams struct {
	Username     string
	Email        string
	PasswordHash string
}

func currentUserFromDomain(user domain.User) CurrentUser {
	return CurrentUser{
		ID:       int64(user.ID),
		Username: user.Username,
		Email:    user.Email,
		Status:   string(user.Status),
	}
}
