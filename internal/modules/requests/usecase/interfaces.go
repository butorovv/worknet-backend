package usecase

import (
	"context"

	"github.com/butorovv/worknet-backend/internal/modules/requests/domain"
)

type Repository interface {
	ListBySpace(ctx context.Context, params domain.ListBySpaceParams) (domain.Page, error)
}

type SpaceAccessResolver interface {
	ResolveAccessibleSpace(ctx context.Context, slug string, viewerID *int64) (SpaceAccess, error)
}
