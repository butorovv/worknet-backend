package delivery

import (
	"errors"
	"net/http"

	"github.com/butorovv/worknet-backend/internal/modules/spaces/domain"
	"github.com/butorovv/worknet-backend/internal/platform/httpx"
)

func WriteUsecaseError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidInput):
		httpx.WriteError(w, http.StatusBadRequest, "space_invalid_input", "Invalid space input")
	case errors.Is(err, domain.ErrForbidden):
		httpx.WriteError(w, http.StatusForbidden, "space_forbidden", "Space access forbidden")
	case errors.Is(err, domain.ErrSpaceNotFound):
		httpx.WriteError(w, http.StatusNotFound, "space_not_found", "Space not found")
	case errors.Is(err, domain.ErrSlugAlreadyExists):
		httpx.WriteError(w, http.StatusConflict, "space_slug_already_exists", "Space slug already exists")
	case errors.Is(err, domain.ErrSpaceArchived):
		httpx.WriteError(w, http.StatusGone, "space_archived", "Space is archived")
	default:
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "Internal server error")
	}
}
