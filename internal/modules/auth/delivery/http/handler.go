package authhttp

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/butorovv/worknet-backend/internal/modules/auth/usecase"
	"github.com/butorovv/worknet-backend/internal/platform/httpx"
)

const maxJSONBodyBytes int64 = 1 << 20

type AuthService interface {
	Register(ctx context.Context, cmd usecase.RegisterCommand) (usecase.AuthResult, error)
	Login(ctx context.Context, cmd usecase.LoginCommand) (usecase.AuthResult, error)
	Logout(ctx context.Context, cmd usecase.LogoutCommand) error
	ValidateSession(ctx context.Context, cmd usecase.ValidateSessionCommand) (usecase.CurrentUser, error)
}

type CookieConfig struct {
	Name   string
	TTL    time.Duration
	Secure bool
}

type Handler struct {
	service AuthService
	cookies CookieConfig
}

func NewHandler(service AuthService, cookies CookieConfig) *Handler {
	return &Handler{
		service: service,
		cookies: cookies,
	}
}

func RegisterRoutes(mux *http.ServeMux, handler *Handler, middleware *Middleware) {
	mux.HandleFunc("/api/v1/auth/register", handler.Register)
	mux.HandleFunc("/api/v1/auth/login", handler.Login)
	mux.Handle("/api/v1/auth/logout", middleware.RequireAuth(http.HandlerFunc(handler.Logout)))
	mux.Handle("/api/v1/auth/me", middleware.RequireAuth(http.HandlerFunc(handler.Me)))
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpx.MethodNotAllowed(w)
		return
	}

	var request registerRequest
	if !decodeJSON(w, r, &request) {
		return
	}

	result, err := h.service.Register(r.Context(), usecase.RegisterCommand{
		Username: request.Username,
		Email:    request.Email,
		Password: request.Password,
	})
	if err != nil {
		writeUsecaseError(w, err)
		return
	}

	h.setSessionCookie(w, result.SessionID)
	httpx.WriteJSON(w, http.StatusCreated, authResultResponse(result))
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpx.MethodNotAllowed(w)
		return
	}

	var request loginRequest
	if !decodeJSON(w, r, &request) {
		return
	}

	result, err := h.service.Login(r.Context(), usecase.LoginCommand{
		Email:    request.Email,
		Password: request.Password,
	})
	if err != nil {
		writeUsecaseError(w, err)
		return
	}

	h.setSessionCookie(w, result.SessionID)
	httpx.WriteJSON(w, http.StatusOK, authResultResponse(result))
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpx.MethodNotAllowed(w)
		return
	}

	sessionID, ok := SessionIDFromContext(r.Context())
	if !ok {
		writeUsecaseError(w, usecaseSessionNotFound())
		return
	}

	if err := h.service.Logout(r.Context(), usecase.LogoutCommand{SessionID: sessionID}); err != nil {
		writeUsecaseError(w, err)
		return
	}

	h.clearSessionCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpx.MethodNotAllowed(w)
		return
	}

	user, ok := CurrentUserFromContext(r.Context())
	if !ok {
		writeUsecaseError(w, usecaseSessionNotFound())
		return
	}

	httpx.WriteJSON(w, http.StatusOK, authResponse{User: userResponseFromCurrent(user)})
}

func (h *Handler) setSessionCookie(w http.ResponseWriter, sessionID string) {
	http.SetCookie(w, &http.Cookie{
		Name:     h.cookies.Name,
		Value:    sessionID,
		Path:     "/",
		MaxAge:   int(h.cookies.TTL.Seconds()),
		HttpOnly: true,
		Secure:   h.cookies.Secure,
		SameSite: http.SameSiteLaxMode,
	})
}

func (h *Handler) clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     h.cookies.Name,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   h.cookies.Secure,
		SameSite: http.SameSiteLaxMode,
	})
}

func decodeJSON(w http.ResponseWriter, r *http.Request, target any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, maxJSONBodyBytes)
	defer r.Body.Close()

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_json", "Invalid JSON request body")
		return false
	}
	var extra struct{}
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_json", "Invalid JSON request body")
		return false
	}
	return true
}
