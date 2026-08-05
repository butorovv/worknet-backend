package app

import (
	"context"
	"errors"

	requestsdomain "github.com/butorovv/worknet-backend/internal/modules/requests/domain"
	requestsusecase "github.com/butorovv/worknet-backend/internal/modules/requests/usecase"
	spacesdomain "github.com/butorovv/worknet-backend/internal/modules/spaces/domain"
	spacesusecase "github.com/butorovv/worknet-backend/internal/modules/spaces/usecase"
)

type spacesAccessService interface {
	ResolveSpaceAccess(ctx context.Context, query spacesusecase.ResolveSpaceAccessQuery) (spacesusecase.SpaceAccess, error)
}

type requestsSpaceAccessAdapter struct {
	spaces spacesAccessService
}

func newRequestsSpaceAccessAdapter(spaces spacesAccessService) *requestsSpaceAccessAdapter {
	return &requestsSpaceAccessAdapter{spaces: spaces}
}

func (a *requestsSpaceAccessAdapter) ResolveAccessibleSpace(ctx context.Context, slug string, viewerID *int64) (requestsusecase.SpaceAccess, error) {
	viewer := spacesusecase.Viewer{}
	if viewerID != nil {
		viewer = spacesusecase.Viewer{
			UserID:          spacesdomain.UserID(*viewerID),
			IsAuthenticated: true,
		}
	}

	access, err := a.spaces.ResolveSpaceAccess(ctx, spacesusecase.ResolveSpaceAccessQuery{
		Viewer: viewer,
		Slug:   slug,
	})
	if err != nil {
		return requestsusecase.SpaceAccess{}, mapSpaceAccessError(err)
	}
	return requestsusecase.SpaceAccess{
		SpaceID:  int64(access.SpaceID),
		IsMember: access.IsMember,
		IsOwner:  access.IsOwner,
		IsAdmin:  access.IsAdmin,
	}, nil
}

func mapSpaceAccessError(err error) error {
	switch {
	case errors.Is(err, spacesdomain.ErrInvalidInput):
		return requestsdomain.ErrInvalidInput
	case errors.Is(err, spacesdomain.ErrSpaceNotFound):
		return requestsdomain.ErrSpaceNotFound
	case errors.Is(err, spacesdomain.ErrForbidden):
		return requestsdomain.ErrSpaceForbidden
	case errors.Is(err, spacesdomain.ErrSpaceArchived):
		return requestsdomain.ErrSpaceArchived
	default:
		return err
	}
}
