package requestshttp

import (
	"context"
	"net/http"
	"strconv"

	"github.com/butorovv/worknet-backend/internal/modules/auth/delivery/middlewares"
	"github.com/butorovv/worknet-backend/internal/modules/requests/delivery"
	"github.com/butorovv/worknet-backend/internal/modules/requests/domain"
	"github.com/butorovv/worknet-backend/internal/modules/requests/usecase"
	"github.com/butorovv/worknet-backend/internal/platform/httpx"
)

const (
	defaultRequestsLimit = 20
	maxRequestsLimit     = 100
)

type RequestsUseCase interface {
	ListSpaceRequests(ctx context.Context, query usecase.ListSpaceRequestsQuery) (usecase.ListSpaceRequestsResult, error)
}

type RequestsHandler struct {
	uc RequestsUseCase
}

func NewRequestsHandler(uc RequestsUseCase) *RequestsHandler {
	return &RequestsHandler{uc: uc}
}

func (h *RequestsHandler) List(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpx.MethodNotAllowed(w)
		return
	}

	limit, err := queryInt(r, "limit", defaultRequestsLimit)
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
		status = string(domain.RequestStatusOpen)
	}
	requestStatus := domain.RequestStatus(status)
	if limit < 1 || limit > maxRequestsLimit || offset < 0 || !requestStatus.Valid() {
		delivery.WriteUsecaseError(w, domain.ErrInvalidInput)
		return
	}

	result, err := h.uc.ListSpaceRequests(r.Context(), usecase.ListSpaceRequestsQuery{
		ViewerID: viewerIDFromContext(r.Context()),
		Slug:     r.PathValue("slug"),
		Status:   requestStatus,
		Limit:    limit,
		Offset:   offset,
	})
	if err != nil {
		delivery.WriteUsecaseError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, toSpaceRequestsResponse(result))
}

func viewerIDFromContext(ctx context.Context) *int64 {
	user, ok := middlewares.CurrentUserFromContext(ctx)
	if !ok || user.ID <= 0 {
		return nil
	}
	viewerID := user.ID
	return &viewerID
}

func queryInt(r *http.Request, name string, defaultValue int) (int, error) {
	value := r.URL.Query().Get(name)
	if value == "" {
		return defaultValue, nil
	}
	return strconv.Atoi(value)
}
