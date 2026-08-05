package usecase

import (
	"context"
	"strings"

	"github.com/butorovv/worknet-backend/internal/modules/spaces/domain"
)

func (s *spacesUseCase) ResolveSpaceAccess(ctx context.Context, query ResolveSpaceAccessQuery) (SpaceAccess, error) {
	slug := strings.TrimSpace(query.Slug)
	if !domain.ValidSlug(slug) {
		return SpaceAccess{}, domain.ErrInvalidInput
	}

	space, err := s.getAccessibleSpace(ctx, slug, query.Viewer)
	if err != nil {
		return SpaceAccess{}, err
	}

	access := SpaceAccess{SpaceID: space.ID}
	if !query.Viewer.IsAuthenticated || query.Viewer.UserID <= 0 {
		return access, nil
	}

	if space.Visibility == domain.SpaceVisibilityPrivate {
		access.IsMember = true
	} else {
		access.IsMember, err = s.repo.IsMember(ctx, space.ID, query.Viewer.UserID)
		if err != nil {
			return SpaceAccess{}, err
		}
	}

	access.IsOwner = space.OwnerID == query.Viewer.UserID
	canManage, err := s.repo.IsOwnerOrAdmin(ctx, space.ID, query.Viewer.UserID)
	if err != nil {
		return SpaceAccess{}, err
	}
	access.IsAdmin = canManage && !access.IsOwner
	return access, nil
}
