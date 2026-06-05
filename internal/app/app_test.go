package app

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	authdelivery "github.com/butorovv/worknet-backend/internal/modules/auth/delivery"
	authhttp "github.com/butorovv/worknet-backend/internal/modules/auth/delivery/http"
	authmw "github.com/butorovv/worknet-backend/internal/modules/auth/delivery/middlewares"
	"github.com/butorovv/worknet-backend/internal/modules/auth/usecase"
)

func TestProtectedAuthRoutesCheckMethodBeforeAuth(t *testing.T) {
	mux := newTestAuthMux()

	tests := []struct {
		name       string
		method     string
		path       string
		wantStatus int
	}{
		{name: "logout wrong method without cookie", method: http.MethodGet, path: "/api/v1/auth/logout", wantStatus: http.StatusMethodNotAllowed},
		{name: "me wrong method without cookie", method: http.MethodPost, path: "/api/v1/auth/me", wantStatus: http.StatusMethodNotAllowed},
		{name: "logout allowed method without cookie", method: http.MethodPost, path: "/api/v1/auth/logout", wantStatus: http.StatusUnauthorized},
		{name: "me allowed method without cookie", method: http.MethodGet, path: "/api/v1/auth/me", wantStatus: http.StatusUnauthorized},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			rec := httptest.NewRecorder()

			mux.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
		})
	}
}

func newTestAuthMux() *http.ServeMux {
	uc := &routeAuthUseCase{}
	cookies := authdelivery.CookieConfig{Name: "session_id", TTL: 30 * 24 * time.Hour}
	handler := authhttp.NewAuthHandler(uc, cookies)
	middleware := authmw.NewMiddleware(uc, cookies)

	mux := http.NewServeMux()
	registerAuthRoutes(mux, handler, middleware)
	return mux
}

type routeAuthUseCase struct{}

func (routeAuthUseCase) Register(context.Context, usecase.RegisterCommand) (usecase.AuthResult, error) {
	return usecase.AuthResult{}, nil
}

func (routeAuthUseCase) Login(context.Context, usecase.LoginCommand) (usecase.AuthResult, error) {
	return usecase.AuthResult{}, nil
}

func (routeAuthUseCase) Logout(context.Context, usecase.LogoutCommand) error {
	return nil
}

func (routeAuthUseCase) ValidateSession(context.Context, usecase.ValidateSessionCommand) (usecase.CurrentUser, error) {
	return usecase.CurrentUser{}, nil
}
