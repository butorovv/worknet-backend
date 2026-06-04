package authhttp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	authdelivery "github.com/butorovv/worknet-backend/internal/modules/auth/delivery"
	authmw "github.com/butorovv/worknet-backend/internal/modules/auth/delivery/middlewares"
	"github.com/butorovv/worknet-backend/internal/modules/auth/domain"
	"github.com/butorovv/worknet-backend/internal/modules/auth/usecase"
)

func TestRegisterCreatesSessionCookie(t *testing.T) {
	service := &fakeAuthService{
		registerResult: usecase.AuthResult{
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
	handler := NewHandler(service, authdelivery.CookieConfig{Name: "session_id", TTL: 30 * 24 * time.Hour})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", strings.NewReader(`{"username":"georgiy","email":"georgiy@example.com","password":"password123"}`))
	rec := httptest.NewRecorder()

	handler.Register(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusCreated)
	}
	cookies := rec.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("cookies len = %d, want 1", len(cookies))
	}
	if cookies[0].Name != "session_id" || cookies[0].Value != "opaque-session" {
		t.Fatalf("cookie = %s:%s", cookies[0].Name, cookies[0].Value)
	}
	if body := rec.Body.String(); !strings.Contains(body, `"username":"georgiy"`) {
		t.Fatalf("unexpected body: %s", body)
	}
}

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
	handler := NewHandler(service, authdelivery.CookieConfig{Name: "session_id", TTL: 30 * 24 * time.Hour})

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
	handler := NewHandler(service, authdelivery.CookieConfig{Name: "session_id", TTL: 30 * 24 * time.Hour})

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

func TestLogoutClearsSessionCookie(t *testing.T) {
	service := &fakeAuthService{
		validateUser: usecase.CurrentUser{
			ID:       1,
			Username: "georgiy",
			Email:    "georgiy@example.com",
			Status:   "active",
		},
	}
	cookies := authdelivery.CookieConfig{Name: "session_id", TTL: 30 * 24 * time.Hour}
	handler := NewHandler(service, cookies)
	middleware := authmw.NewMiddleware(service, cookies)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	req.AddCookie(&http.Cookie{Name: "session_id", Value: "opaque-session"})
	rec := httptest.NewRecorder()

	middleware.RequireAuth(http.HandlerFunc(handler.Logout)).ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
	cookiesSet := rec.Result().Cookies()
	if len(cookiesSet) != 1 {
		t.Fatalf("cookies len = %d, want 1", len(cookiesSet))
	}
	if cookiesSet[0].Name != "session_id" || cookiesSet[0].Value != "" {
		t.Fatalf("cookie = %s:%s", cookiesSet[0].Name, cookiesSet[0].Value)
	}
	if cookiesSet[0].MaxAge != -1 {
		t.Fatalf("MaxAge = %d, want -1", cookiesSet[0].MaxAge)
	}
	if service.logoutCommand.SessionID != "opaque-session" {
		t.Fatalf("logout session id = %q", service.logoutCommand.SessionID)
	}
}

func TestMeReturnsCurrentUser(t *testing.T) {
	service := &fakeAuthService{
		validateUser: usecase.CurrentUser{
			ID:       1,
			Username: "georgiy",
			Email:    "georgiy@example.com",
			Status:   "active",
		},
	}
	cookies := authdelivery.CookieConfig{Name: "session_id", TTL: 30 * 24 * time.Hour}
	handler := NewHandler(service, cookies)
	middleware := authmw.NewMiddleware(service, cookies)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	req.AddCookie(&http.Cookie{Name: "session_id", Value: "opaque-session"})
	rec := httptest.NewRecorder()

	middleware.RequireAuth(http.HandlerFunc(handler.Me)).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if body := rec.Body.String(); !strings.Contains(body, `"email":"georgiy@example.com"`) {
		t.Fatalf("unexpected body: %s", body)
	}
}

func TestHandlerRejectsInvalidMethod(t *testing.T) {
	service := &fakeAuthService{}
	handler := NewHandler(service, authdelivery.CookieConfig{Name: "session_id", TTL: 30 * 24 * time.Hour})

	tests := []struct {
		name   string
		method string
		path   string
		call   func(http.ResponseWriter, *http.Request)
	}{
		{name: "register", method: http.MethodGet, path: "/api/v1/auth/register", call: handler.Register},
		{name: "login", method: http.MethodGet, path: "/api/v1/auth/login", call: handler.Login},
		{name: "logout", method: http.MethodGet, path: "/api/v1/auth/logout", call: handler.Logout},
		{name: "me", method: http.MethodPost, path: "/api/v1/auth/me", call: handler.Me},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			rec := httptest.NewRecorder()

			tt.call(rec, req)

			if rec.Code != http.StatusMethodNotAllowed {
				t.Fatalf("status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
			}
		})
	}
}

func TestHandlerRejectsInvalidJSON(t *testing.T) {
	service := &fakeAuthService{}
	handler := NewHandler(service, authdelivery.CookieConfig{Name: "session_id", TTL: 30 * 24 * time.Hour})

	tests := []struct {
		name string
		body string
	}{
		{name: "invalid", body: `{`},
		{name: "unknown field", body: `{"email":"georgiy@example.com","password":"password123","unexpected":true}`},
		{name: "trailing token", body: `{"email":"georgiy@example.com","password":"password123"} {"extra":true}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(tt.body))
			rec := httptest.NewRecorder()

			handler.Login(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
			}
			if body := rec.Body.String(); !strings.Contains(body, `"invalid_json"`) {
				t.Fatalf("unexpected body: %s", body)
			}
		})
	}
}

func TestMiddlewareRejectsMissingCookie(t *testing.T) {
	service := &fakeAuthService{}
	middleware := authmw.NewMiddleware(service, authdelivery.CookieConfig{Name: "session_id", TTL: 30 * 24 * time.Hour})
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

func TestRegisterRoutesConnectsAuthEndpoints(t *testing.T) {
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
	cookies := authdelivery.CookieConfig{Name: "session_id", TTL: 30 * 24 * time.Hour}
	mux := http.NewServeMux()
	RegisterRoutes(mux, NewHandler(service, cookies), authmw.NewMiddleware(service, cookies))

	loginReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`{"email":"georgiy@example.com","password":"password123"}`))
	loginRec := httptest.NewRecorder()

	mux.ServeHTTP(loginRec, loginReq)

	if loginRec.Code != http.StatusOK {
		t.Fatalf("login status = %d, want %d", loginRec.Code, http.StatusOK)
	}

	meReq := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	meRec := httptest.NewRecorder()

	mux.ServeHTTP(meRec, meReq)

	if meRec.Code != http.StatusUnauthorized {
		t.Fatalf("me status = %d, want %d", meRec.Code, http.StatusUnauthorized)
	}
}

type fakeAuthService struct {
	registerResult usecase.AuthResult
	registerErr    error
	loginResult    usecase.AuthResult
	loginErr       error
	logoutCommand  usecase.LogoutCommand
	logoutErr      error
	validateUser   usecase.CurrentUser
	validateErr    error
}

func (f *fakeAuthService) Register(context.Context, usecase.RegisterCommand) (usecase.AuthResult, error) {
	return f.registerResult, f.registerErr
}

func (f *fakeAuthService) Login(context.Context, usecase.LoginCommand) (usecase.AuthResult, error) {
	return f.loginResult, f.loginErr
}

func (f *fakeAuthService) Logout(_ context.Context, cmd usecase.LogoutCommand) error {
	f.logoutCommand = cmd
	return f.logoutErr
}

func (f *fakeAuthService) ValidateSession(_ context.Context, _ usecase.ValidateSessionCommand) (usecase.CurrentUser, error) {
	return f.validateUser, f.validateErr
}
