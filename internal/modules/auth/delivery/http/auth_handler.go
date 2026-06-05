package authhttp

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	authdelivery "github.com/butorovv/worknet-backend/internal/modules/auth/delivery"
	"github.com/butorovv/worknet-backend/internal/modules/auth/delivery/middlewares"
	"github.com/butorovv/worknet-backend/internal/modules/auth/delivery/transport"
	"github.com/butorovv/worknet-backend/internal/modules/auth/domain"
	"github.com/butorovv/worknet-backend/internal/modules/auth/usecase"
	"github.com/butorovv/worknet-backend/internal/platform/httpx"
)

const maxJSONBodyBytes int64 = 1 << 20

type AuthUseCaseInterface interface {
	Register(ctx context.Context, cmd usecase.RegisterCommand) (usecase.AuthResult, error)
	Login(ctx context.Context, cmd usecase.LoginCommand) (usecase.AuthResult, error)
	Logout(ctx context.Context, cmd usecase.LogoutCommand) error
}

type AuthHandler struct {
	uc      AuthUseCaseInterface
	cookies authdelivery.CookieConfig
}

func NewAuthHandler(uc AuthUseCaseInterface, cookies authdelivery.CookieConfig) *AuthHandler {
	return &AuthHandler{
		uc:      uc,
		cookies: cookies,
	}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpx.MethodNotAllowed(w)
		return
	}

	var request transport.RegisterRequest
	if !decodeJSON(w, r, &request) {
		return
	}

	result, err := h.uc.Register(r.Context(), usecase.RegisterCommand{
		Username: request.Username,
		Email:    request.Email,
		Password: request.Password,
	})
	if err != nil {
		authdelivery.WriteUsecaseError(w, err)
		return
	}

	h.setSessionCookie(w, result.SessionID)
	httpx.WriteJSON(w, http.StatusCreated, transport.ToAuthResponse(result))
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpx.MethodNotAllowed(w)
		return
	}

	var request transport.LoginRequest
	if !decodeJSON(w, r, &request) {
		return
	}

	result, err := h.uc.Login(r.Context(), usecase.LoginCommand{
		Email:    request.Email,
		Password: request.Password,
	})
	if err != nil {
		authdelivery.WriteUsecaseError(w, err)
		return
	}

	h.setSessionCookie(w, result.SessionID)
	httpx.WriteJSON(w, http.StatusOK, transport.ToAuthResponse(result))
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpx.MethodNotAllowed(w)
		return
	}

	sessionID, ok := middlewares.SessionIDFromContext(r.Context())
	if !ok {
		authdelivery.WriteUsecaseError(w, domain.ErrSessionNotFound)
		return
	}

	if err := h.uc.Logout(r.Context(), usecase.LogoutCommand{SessionID: sessionID}); err != nil {
		authdelivery.WriteUsecaseError(w, err)
		return
	}

	h.clearSessionCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpx.MethodNotAllowed(w)
		return
	}

	user, ok := middlewares.CurrentUserFromContext(r.Context())
	if !ok {
		authdelivery.WriteUsecaseError(w, domain.ErrSessionNotFound)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, transport.AuthResponse{User: transport.ToUserResponse(user)})
}

func (h *AuthHandler) setSessionCookie(w http.ResponseWriter, sessionID string) {
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

func (h *AuthHandler) clearSessionCookie(w http.ResponseWriter) {
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
