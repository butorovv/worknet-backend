package spaceshttp

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/butorovv/worknet-backend/internal/modules/auth/delivery/middlewares"
	"github.com/butorovv/worknet-backend/internal/modules/spaces/delivery"
	"github.com/butorovv/worknet-backend/internal/modules/spaces/delivery/transport"
	"github.com/butorovv/worknet-backend/internal/modules/spaces/domain"
	"github.com/butorovv/worknet-backend/internal/modules/spaces/usecase"
	"github.com/butorovv/worknet-backend/internal/platform/httpx"
)

const maxJSONBodyBytes int64 = 1 << 20

type SpacesUseCaseInterface interface {
	CreateSpace(ctx context.Context, cmd usecase.CreateSpaceCommand) (usecase.SpaceDetails, error)
	GetSpace(ctx context.Context, cmd usecase.GetSpaceCommand) (usecase.SpaceDetails, error)
	ListSpaces(ctx context.Context, cmd usecase.ListSpacesCommand) ([]domain.Space, error)
	UpdateSpace(ctx context.Context, cmd usecase.UpdateSpaceCommand) (usecase.SpaceDetails, error)
	ArchiveSpace(ctx context.Context, cmd usecase.ArchiveSpaceCommand) error
	GetSpaceOverview(ctx context.Context, cmd usecase.GetSpaceOverviewCommand) (usecase.SpaceOverview, error)
}

type SpacesHandler struct {
	uc SpacesUseCaseInterface
}

func NewSpacesHandler(uc SpacesUseCaseInterface) *SpacesHandler {
	return &SpacesHandler{uc: uc}
}

func (h *SpacesHandler) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpx.MethodNotAllowed(w)
		return
	}

	var request transport.CreateSpaceRequest
	if !decodeJSON(w, r, &request) {
		return
	}

	result, err := h.uc.CreateSpace(r.Context(), usecase.CreateSpaceCommand{
		Viewer:           viewerFromContext(r.Context()),
		Title:            request.Title,
		Slug:             request.Slug,
		ShortDescription: request.ShortDescription,
		Description:      request.Description,
		WhyItMatters:     request.WhyItMatters,
		Visibility:       domain.SpaceVisibility(request.Visibility),
		Stage:            domain.SpaceStage(request.Stage),
		CoverIcon:        domain.CoverIcon(request.CoverIcon),
		AccentColor:      domain.AccentColor(request.AccentColor),
		Stack:            request.Stack,
		Links:            linksFromRequest(request.Links),
	})
	if err != nil {
		delivery.WriteUsecaseError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, toSpaceDetailsResponse(result))
}

func (h *SpacesHandler) List(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpx.MethodNotAllowed(w)
		return
	}

	limit, err := queryInt(r, "limit", 20)
	if err != nil {
		delivery.WriteUsecaseError(w, domain.ErrInvalidInput)
		return
	}
	offset, err := queryInt(r, "offset", 0)
	if err != nil {
		delivery.WriteUsecaseError(w, domain.ErrInvalidInput)
		return
	}
	status := r.URL.Query().Get("status")
	if status == "" {
		status = string(domain.SpaceStatusActive)
	}
	visibility := r.URL.Query().Get("visibility")
	if visibility == "" {
		visibility = string(domain.SpaceVisibilityPublic)
	}

	spaces, err := h.uc.ListSpaces(r.Context(), usecase.ListSpacesCommand{
		Viewer:     viewerFromContext(r.Context()),
		Status:     domain.SpaceStatus(status),
		Visibility: domain.SpaceVisibility(visibility),
		Limit:      limit,
		Offset:     offset,
	})
	if err != nil {
		delivery.WriteUsecaseError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, toSpacesResponse(spaces))
}

func (h *SpacesHandler) Get(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpx.MethodNotAllowed(w)
		return
	}

	result, err := h.uc.GetSpace(r.Context(), usecase.GetSpaceCommand{
		Viewer: viewerFromContext(r.Context()),
		Slug:   r.PathValue("slug"),
	})
	if err != nil {
		delivery.WriteUsecaseError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, toSpaceDetailsResponse(result))
}

func (h *SpacesHandler) Overview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpx.MethodNotAllowed(w)
		return
	}

	result, err := h.uc.GetSpaceOverview(r.Context(), usecase.GetSpaceOverviewCommand{
		Viewer: viewerFromContext(r.Context()),
		Slug:   r.PathValue("slug"),
	})
	if err != nil {
		delivery.WriteUsecaseError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, toOverviewResponse(result))
}

func (h *SpacesHandler) Update(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		httpx.MethodNotAllowed(w)
		return
	}

	var request transport.UpdateSpaceRequest
	if !decodeJSON(w, r, &request) {
		return
	}

	result, err := h.uc.UpdateSpace(r.Context(), usecase.UpdateSpaceCommand{
		Viewer:           viewerFromContext(r.Context()),
		Slug:             r.PathValue("slug"),
		Title:            request.Title,
		ShortDescription: request.ShortDescription,
		Description:      request.Description,
		WhyItMatters:     request.WhyItMatters,
		Visibility:       visibilityPointer(request.Visibility),
		Stage:            stagePointer(request.Stage),
		Status:           statusPointer(request.Status),
		CoverIcon:        coverIconPointer(request.CoverIcon),
		AccentColor:      accentColorPointer(request.AccentColor),
		Stack:            request.Stack,
		Links:            linksPointer(request.Links),
	})
	if err != nil {
		delivery.WriteUsecaseError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, toSpaceDetailsResponse(result))
}

func (h *SpacesHandler) Archive(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		httpx.MethodNotAllowed(w)
		return
	}

	if err := h.uc.ArchiveSpace(r.Context(), usecase.ArchiveSpaceCommand{
		Viewer: viewerFromContext(r.Context()),
		Slug:   r.PathValue("slug"),
	}); err != nil {
		delivery.WriteUsecaseError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func viewerFromContext(ctx context.Context) usecase.Viewer {
	user, ok := middlewares.CurrentUserFromContext(ctx)
	if !ok {
		return usecase.Viewer{}
	}
	return usecase.Viewer{
		UserID:          domain.UserID(user.ID),
		IsAuthenticated: true,
	}
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

func queryInt(r *http.Request, name string, defaultValue int) (int, error) {
	value := r.URL.Query().Get(name)
	if value == "" {
		return defaultValue, nil
	}
	return strconv.Atoi(value)
}

func visibilityPointer(value *string) *domain.SpaceVisibility {
	if value == nil {
		return nil
	}
	result := domain.SpaceVisibility(*value)
	return &result
}

func stagePointer(value *string) *domain.SpaceStage {
	if value == nil {
		return nil
	}
	result := domain.SpaceStage(*value)
	return &result
}

func statusPointer(value *string) *domain.SpaceStatus {
	if value == nil {
		return nil
	}
	result := domain.SpaceStatus(*value)
	return &result
}

func coverIconPointer(value *string) *domain.CoverIcon {
	if value == nil {
		return nil
	}
	result := domain.CoverIcon(*value)
	return &result
}

func accentColorPointer(value *string) *domain.AccentColor {
	if value == nil {
		return nil
	}
	result := domain.AccentColor(*value)
	return &result
}

func linksPointer(value *[]transport.SpaceLinkRequest) *[]domain.SpaceLink {
	if value == nil {
		return nil
	}
	result := linksFromRequest(*value)
	return &result
}
