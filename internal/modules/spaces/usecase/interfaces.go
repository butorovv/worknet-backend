package usecase

import (
	"context"
	"time"

	"github.com/butorovv/worknet-backend/internal/modules/spaces/domain"
)

type SpacesRepository interface {
	Create(ctx context.Context, space domain.Space, stack []domain.SpaceStackItem, links []domain.SpaceLink) (domain.Space, error)
	GetByID(ctx context.Context, id domain.SpaceID) (domain.Space, error)
	GetBySlug(ctx context.Context, slug string) (domain.Space, error)
	List(ctx context.Context, filter domain.SpaceListFilter) ([]domain.Space, error)
	Update(ctx context.Context, space domain.Space, stack *[]domain.SpaceStackItem, links *[]domain.SpaceLink) (domain.Space, error)
	Archive(ctx context.Context, slug string, archivedAt time.Time) (domain.Space, error)
	AddMember(ctx context.Context, member domain.SpaceMember) error
	ListMembers(ctx context.Context, spaceID domain.SpaceID) ([]domain.SpaceMember, error)
	CountMembers(ctx context.Context, spaceID domain.SpaceID) (int, error)
	IsMember(ctx context.Context, spaceID domain.SpaceID, userID domain.UserID) (bool, error)
	IsOwnerOrAdmin(ctx context.Context, spaceID domain.SpaceID, userID domain.UserID) (bool, error)
	ReplaceStack(ctx context.Context, spaceID domain.SpaceID, stack []domain.SpaceStackItem) error
	ListStack(ctx context.Context, spaceID domain.SpaceID) ([]domain.SpaceStackItem, error)
	ReplaceLinks(ctx context.Context, spaceID domain.SpaceID, links []domain.SpaceLink) error
	ListLinks(ctx context.Context, spaceID domain.SpaceID) ([]domain.SpaceLink, error)
}

type Clock interface {
	Now() time.Time
}
