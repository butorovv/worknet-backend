package authhttp

import (
	"errors"
	"net/http"

	"github.com/butorovv/worknet-backend/internal/modules/auth/domain"
	"github.com/butorovv/worknet-backend/internal/platform/httpx"
)

func writeUsecaseError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidInput):
		httpx.WriteError(w, http.StatusBadRequest, "invalid_request", "Invalid request")
	case errors.Is(err, domain.ErrUserAlreadyExists):
		httpx.WriteError(w, http.StatusConflict, "user_already_exists", "User already exists")
	case errors.Is(err, domain.ErrInvalidCredentials):
		httpx.WriteError(w, http.StatusUnauthorized, "invalid_credentials", "Invalid credentials")
	case errors.Is(err, domain.ErrSessionNotFound), errors.Is(err, domain.ErrUserNotFound):
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized", "Unauthorized")
	case errors.Is(err, domain.ErrUserInactive):
		httpx.WriteError(w, http.StatusForbidden, "user_inactive", "User is inactive")
	default:
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "Internal server error")
	}
}
