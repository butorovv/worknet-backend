package middlewares

import (
	"context"
	"errors"
	"net/http"

	authdelivery "github.com/butorovv/worknet-backend/internal/modules/auth/delivery"
	authdomain "github.com/butorovv/worknet-backend/internal/modules/auth/domain"
	"github.com/butorovv/worknet-backend/internal/modules/auth/usecase"
	"github.com/butorovv/worknet-backend/internal/platform/httpx"
)

type SessionValidator interface {
	ValidateSession(ctx context.Context, cmd usecase.ValidateSessionCommand) (usecase.CurrentUser, error)
}

type Middleware struct {
	validator SessionValidator
	cookies   authdelivery.CookieConfig
}

func NewMiddleware(validator SessionValidator, cookies authdelivery.CookieConfig) *Middleware {
	return &Middleware{
		validator: validator,
		cookies:   cookies,
	}
}

func (m *Middleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(m.cookies.Name)
		if err != nil || cookie.Value == "" {
			httpx.WriteError(w, http.StatusUnauthorized, "unauthorized", "Unauthorized")
			return
		}

		user, err := m.validator.ValidateSession(r.Context(), usecase.ValidateSessionCommand{
			SessionID: cookie.Value,
		})
		if err != nil {
			authdelivery.WriteUsecaseError(w, err)
			return
		}

		ctx := withAuthContext(r.Context(), cookie.Value, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *Middleware) OptionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(m.cookies.Name)
		if err != nil || cookie.Value == "" {
			next.ServeHTTP(w, r)
			return
		}

		user, err := m.validator.ValidateSession(r.Context(), usecase.ValidateSessionCommand{
			SessionID: cookie.Value,
		})
		if err != nil {
			if isAnonymousSessionError(err) {
				next.ServeHTTP(w, r)
				return
			}
			authdelivery.WriteUsecaseError(w, err)
			return
		}

		ctx := withAuthContext(r.Context(), cookie.Value, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func isAnonymousSessionError(err error) bool {
	return errors.Is(err, authdomain.ErrSessionNotFound) ||
		errors.Is(err, authdomain.ErrUserNotFound) ||
		errors.Is(err, authdomain.ErrUserInactive)
}
