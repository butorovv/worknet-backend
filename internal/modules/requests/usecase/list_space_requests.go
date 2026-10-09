package usecase

import (
	"context"
	"errors"

	"github.com/butorovv/worknet-backend/internal/modules/requests/domain"
)

const (
	defaultListLimit = 20
	maxListLimit     = 100
)

type requestsUseCase struct {
	repo          Repository
	spaceResolver SpaceAccessResolver
}

func NewRequestsUseCase(repo Repository, spaceResolver SpaceAccessResolver) (*requestsUseCase, error) {
	if repo == nil {
		return nil, errors.New("requests repository is required")
	}
	if spaceResolver == nil {
		return nil, errors.New("space access resolver is required")
	}
	return &requestsUseCase{repo: repo, spaceResolver: spaceResolver}, nil
}

func (s *requestsUseCase) ListSpaceRequests(ctx context.Context, query ListSpaceRequestsQuery) (ListSpaceRequestsResult, error) {
	status := query.Status
	if status == "" {
		status = domain.RequestStatusOpen
	}
	limit := query.Limit
	if limit == 0 {
		limit = defaultListLimit
	}
	if !status.Valid() || limit < 1 || limit > maxListLimit || query.Offset < 0 {
		return ListSpaceRequestsResult{}, domain.ErrInvalidInput
	}

	access, err := s.spaceResolver.ResolveAccessibleSpace(ctx, query.Slug, query.ViewerID)
	if err != nil {
		return ListSpaceRequestsResult{}, err
	}

	page, err := s.repo.ListBySpace(ctx, domain.ListBySpaceParams{
		SpaceID:   access.SpaceID,
		ViewerID:  query.ViewerID,
		IsMember:  access.IsMember,
		CanManage: access.IsOwner || access.IsAdmin,
		Status:    status,
		Limit:     limit,
		Offset:    query.Offset,
	})
	if err != nil {
		return ListSpaceRequestsResult{}, err
	}
	if page.Items == nil {
		page.Items = []domain.Request{}
	}

	return ListSpaceRequestsResult{
		Items: page.Items,
		Pagination: Pagination{
			Limit:  limit,
			Offset: query.Offset,
			Total:  page.Total,
		},
	}, nil
}
