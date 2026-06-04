package authhttp

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/butorovv/worknet-backend/internal/modules/auth/domain"
	"github.com/butorovv/worknet-backend/internal/modules/auth/usecase"
)

func TestLoginSetsSessionCookie(t *testing.T) {
	service := &fakeAuthService{
		loginResult: usecase.AuthResult{
			User: usecase.CurrentUser{
				ID:       1,
				Username: "georgiy",
				Email:    "georgiy@example.com",
				Status:   "active",
			},
			SessionID: "opaque-session",
			ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
		},
	}
	handler := NewHandler(service, CookieConfig{Name: "session_id", TTL: 30 * 24 * time.Hour})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`{"email":"georgiy@example.com","password":"password123"}`))
	rec := httptest.NewRecorder()

	handler.Login(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	cookies := rec.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("cookies len = %d, want 1", len(cookies))
	}
	if cookies[0].Name != "session_id" || cookies[0].Value != "opaque-session" {
		t.Fatalf("cookie = %s:%s", cookies[0].Name, cookies[0].Value)
	}
	if !cookies[0].HttpOnly {
		t.Fatal("session cookie must be HttpOnly")
	}
	if cookies[0].SameSite != http.SameSiteLaxMode {
		t.Fatalf("SameSite = %v, want Lax", cookies[0].SameSite)
	}
}

func TestLoginWritesErrorFormat(t *testing.T) {
	service := &fakeAuthService{loginErr: domain.ErrInvalidCredentials}
	handler := NewHandler(service, CookieConfig{Name: "session_id", TTL: 30 * 24 * time.Hour})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`{"email":"georgiy@example.com","password":"wrong-password"}`))
	rec := httptest.NewRecorder()

	handler.Login(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
	if body := rec.Body.String(); !strings.Contains(body, `"error"`) || !strings.Contains(body, `"invalid_credentials"`) {
		t.Fatalf("unexpected error body: %s", body)
	}
}

func TestMiddlewareRejectsMissingCookie(t *testing.T) {
	service := &fakeAuthService{}
	middleware := NewMiddleware(service, CookieConfig{Name: "session_id", TTL: 30 * 24 * time.Hour})
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next handler must not be called")
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	rec := httptest.NewRecorder()

	middleware.RequireAuth(next).ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

type fakeAuthService struct {
	loginResult usecase.AuthResult
	loginErr    error
}

func (f *fakeAuthService) Register(context.Context, usecase.RegisterCommand) (usecase.AuthResult, error) {
	return usecase.AuthResult{}, errors.New("not implemented")
}

func (f *fakeAuthService) Login(context.Context, usecase.LoginCommand) (usecase.AuthResult, error) {
	return f.loginResult, f.loginErr
}

func (f *fakeAuthService) Logout(context.Context, usecase.LogoutCommand) error {
	return nil
}

func (f *fakeAuthService) ValidateSession(context.Context, usecase.ValidateSessionCommand) (usecase.CurrentUser, error) {
	return usecase.CurrentUser{}, domain.ErrSessionNotFound
}
