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
	spaceshttp "github.com/butorovv/worknet-backend/internal/modules/spaces/delivery/http"
	spacesdomain "github.com/butorovv/worknet-backend/internal/modules/spaces/domain"
	spacesusecase "github.com/butorovv/worknet-backend/internal/modules/spaces/usecase"
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

func TestSpacesRoutesPublicProtectedAndPathValue(t *testing.T) {
	authUC := &routeAuthUseCase{}
	cookies := authdelivery.CookieConfig{Name: "session_id", TTL: 30 * 24 * time.Hour}
	authMiddleware := authmw.NewMiddleware(authUC, cookies)
	spacesUC := &routeSpacesUseCase{}
	handler := spaceshttp.NewSpacesHandler(spacesUC)

	mux := http.NewServeMux()
	registerSpacesRoutes(mux, handler, authMiddleware)

	tests := []struct {
		name       string
		method     string
		path       string
		wantStatus int
	}{
		{name: "public list", method: http.MethodGet, path: "/api/v1/spaces", wantStatus: http.StatusOK},
		{name: "public get", method: http.MethodGet, path: "/api/v1/spaces/worknet", wantStatus: http.StatusOK},
		{name: "public overview", method: http.MethodGet, path: "/api/v1/spaces/worknet/overview", wantStatus: http.StatusOK},
		{name: "protected create", method: http.MethodPost, path: "/api/v1/spaces", wantStatus: http.StatusUnauthorized},
		{name: "protected update", method: http.MethodPatch, path: "/api/v1/spaces/worknet", wantStatus: http.StatusUnauthorized},
		{name: "protected archive", method: http.MethodDelete, path: "/api/v1/spaces/worknet", wantStatus: http.StatusUnauthorized},
		{name: "wrong method", method: http.MethodPut, path: "/api/v1/spaces/worknet", wantStatus: http.StatusMethodNotAllowed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			rec := httptest.NewRecorder()

			mux.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body=%s", rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}
	if spacesUC.lastSlug != "worknet" {
		t.Fatalf("path slug = %q, want worknet", spacesUC.lastSlug)
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

type routeSpacesUseCase struct {
	lastSlug string
}

func (r *routeSpacesUseCase) CreateSpace(context.Context, spacesusecase.CreateSpaceCommand) (spacesusecase.SpaceDetails, error) {
	return spacesusecase.SpaceDetails{}, nil
}

func (r *routeSpacesUseCase) GetSpace(_ context.Context, cmd spacesusecase.GetSpaceCommand) (spacesusecase.SpaceDetails, error) {
	r.lastSlug = cmd.Slug
	return spacesusecase.SpaceDetails{
		Stack: []spacesdomain.SpaceStackItem{},
		Links: []spacesdomain.SpaceLink{},
	}, nil
}

func (r *routeSpacesUseCase) ListSpaces(context.Context, spacesusecase.ListSpacesCommand) ([]spacesdomain.Space, error) {
	return []spacesdomain.Space{}, nil
}

func (r *routeSpacesUseCase) UpdateSpace(context.Context, spacesusecase.UpdateSpaceCommand) (spacesusecase.SpaceDetails, error) {
	return spacesusecase.SpaceDetails{}, nil
}

func (r *routeSpacesUseCase) ArchiveSpace(context.Context, spacesusecase.ArchiveSpaceCommand) error {
	return nil
}

func (r *routeSpacesUseCase) GetSpaceOverview(_ context.Context, cmd spacesusecase.GetSpaceOverviewCommand) (spacesusecase.SpaceOverview, error) {
	r.lastSlug = cmd.Slug
	return spacesusecase.SpaceOverview{
		MembersPreview:       []spacesusecase.MemberPreview{},
		Stack:                []spacesdomain.SpaceStackItem{},
		Links:                []spacesdomain.SpaceLink{},
		Milestones:           []spacesusecase.MilestonePreview{},
		ArtifactsPreview:     []spacesusecase.ArtifactPreview{},
		OpenRequestsPreview:  []spacesusecase.RequestPreview{},
		CompletedWorkPreview: []spacesusecase.ContributionPreview{},
		Activity: spacesusecase.ActivitySummary{
			Range: "30d",
			Days:  []spacesusecase.ActivityDay{},
		},
	}, nil
}
