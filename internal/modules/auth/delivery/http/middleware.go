package authhttp

import (
	"net/http"

	"github.com/butorovv/worknet-backend/internal/modules/auth/domain"
	"github.com/butorovv/worknet-backend/internal/modules/auth/usecase"
	"github.com/butorovv/worknet-backend/internal/platform/httpx"
)

type Middleware struct {
	service AuthService
	cookies CookieConfig
}

func NewMiddleware(service AuthService, cookies CookieConfig) *Middleware {
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
			writeUsecaseError(w, err)
			return
		}

		ctx := withAuthContext(r.Context(), cookie.Value, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func usecaseSessionNotFound() error {
	return domain.ErrSessionNotFound
}
