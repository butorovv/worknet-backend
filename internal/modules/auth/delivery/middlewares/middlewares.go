package middlewares

import (
	"context"
	"net/http"

	authdelivery "github.com/butorovv/worknet-backend/internal/modules/auth/delivery"
	"github.com/butorovv/worknet-backend/internal/modules/auth/usecase"
	"github.com/butorovv/worknet-backend/internal/platform/httpx"
)

type AuthUsecase interface {
	ValidateSession(ctx context.Context, cmd usecase.ValidateSessionCommand) (usecase.CurrentUser, error)
}

type Middleware struct {
	service AuthUsecase
	cookies authdelivery.CookieConfig
}

func NewMiddleware(service AuthUsecase, cookies authdelivery.CookieConfig) *Middleware {
	return &Middleware{
		service: service,
		cookies: cookies,
	}
}

func (m *Middleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(m.cookies.Name)
		if err != nil || cookie.Value == "" {
			httpx.WriteError(w, http.StatusUnauthorized, "unauthorized", "Unauthorized")
			return
		}

		user, err := m.service.ValidateSession(r.Context(), usecase.ValidateSessionCommand{
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
