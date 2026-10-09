package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/butorovv/worknet-backend/internal/modules/requests/domain"
)

func TestListSpaceRequestsDefaultsAndEmptyItems(t *testing.T) {
	repo := &fakeRepository{}
	resolver := &fakeSpaceAccessResolver{access: SpaceAccess{SpaceID: 7}}
	uc, err := NewRequestsUseCase(repo, resolver)
	if err != nil {
		t.Fatal(err)
	}

	result, err := uc.ListSpaceRequests(context.Background(), ListSpaceRequestsQuery{Slug: "worknet"})
	if err != nil {
		t.Fatal(err)
	}
	if result.Items == nil || len(result.Items) != 0 {
		t.Fatalf("items = %#v", result.Items)
	}
	if result.Pagination != (Pagination{Limit: 20, Offset: 0, Total: 0}) {
		t.Fatalf("pagination = %+v", result.Pagination)
	}
	if repo.lastParams.SpaceID != 7 || repo.lastParams.Status != domain.RequestStatusOpen || repo.lastParams.Limit != 20 {
		t.Fatalf("params = %+v", repo.lastParams)
	}
	if resolver.lastSlug != "worknet" || resolver.lastViewerID != nil {
		t.Fatalf("resolver input = slug %q viewer %v", resolver.lastSlug, resolver.lastViewerID)
	}
}

func TestListSpaceRequestsPassesViewerAccessAndPagination(t *testing.T) {
	viewerID := int64(42)
	tests := []struct {
		name      string
		access    SpaceAccess
		canManage bool
	}{
		{name: "authenticated non-member", access: SpaceAccess{SpaceID: 7}},
		{name: "active member", access: SpaceAccess{SpaceID: 7, IsMember: true}},
		{name: "space owner", access: SpaceAccess{SpaceID: 7, IsMember: true, IsOwner: true}, canManage: true},
		{name: "space admin", access: SpaceAccess{SpaceID: 7, IsMember: true, IsAdmin: true}, canManage: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeRepository{page: domain.Page{Items: []domain.Request{{ID: 9}}, Total: 1}}
			resolver := &fakeSpaceAccessResolver{access: tt.access}
			uc, _ := NewRequestsUseCase(repo, resolver)

			result, err := uc.ListSpaceRequests(context.Background(), ListSpaceRequestsQuery{
				ViewerID: &viewerID,
				Slug:     "worknet",
				Status:   domain.RequestStatusClosed,
				Limit:    5,
				Offset:   2,
			})
			if err != nil {
				t.Fatal(err)
			}
			if len(result.Items) != 1 || result.Pagination != (Pagination{Limit: 5, Offset: 2, Total: 1}) {
				t.Fatalf("result = %+v", result)
			}
			if repo.lastParams.ViewerID == nil || *repo.lastParams.ViewerID != viewerID {
				t.Fatalf("viewer params = %+v", repo.lastParams)
			}
			if repo.lastParams.IsMember != tt.access.IsMember || repo.lastParams.CanManage != tt.canManage {
				t.Fatalf("access params = %+v", repo.lastParams)
			}
			if repo.lastParams.Status != domain.RequestStatusClosed || repo.lastParams.Limit != 5 || repo.lastParams.Offset != 2 {
				t.Fatalf("query params = %+v", repo.lastParams)
			}
		})
	}
}

func TestListSpaceRequestsRejectsInvalidRequestQueryBeforeResolvingSpace(t *testing.T) {
	tests := []ListSpaceRequestsQuery{
		{Slug: "worknet", Status: "pending"},
		{Slug: "worknet", Limit: -1},
		{Slug: "worknet", Limit: 101},
		{Slug: "worknet", Offset: -1},
	}
	for _, query := range tests {
		repo := &fakeRepository{}
		resolver := &fakeSpaceAccessResolver{}
		uc, _ := NewRequestsUseCase(repo, resolver)

		_, err := uc.ListSpaceRequests(context.Background(), query)
		if !errors.Is(err, domain.ErrInvalidInput) {
			t.Fatalf("query = %+v, error = %v", query, err)
		}
		if resolver.calls != 0 || repo.calls != 0 {
			t.Fatalf("resolver calls = %d, repository calls = %d", resolver.calls, repo.calls)
		}
	}
}

func TestListSpaceRequestsPropagatesSpaceAccessErrors(t *testing.T) {
	tests := []error{
		domain.ErrInvalidInput,
		domain.ErrSpaceNotFound,
		domain.ErrSpaceForbidden,
		domain.ErrSpaceArchived,
		errors.New("spaces infrastructure unavailable"),
	}
	for _, wantErr := range tests {
		repo := &fakeRepository{}
		resolver := &fakeSpaceAccessResolver{err: wantErr}
		uc, _ := NewRequestsUseCase(repo, resolver)

		_, err := uc.ListSpaceRequests(context.Background(), ListSpaceRequestsQuery{Slug: "worknet"})
		if !errors.Is(err, wantErr) {
			t.Fatalf("error = %v, want %v", err, wantErr)
		}
		if repo.calls != 0 {
			t.Fatal("repository must not be called when Space access fails")
		}
	}
}

func TestListSpaceRequestsPropagatesRepositoryError(t *testing.T) {
	repositoryErr := errors.New("database unavailable")
	repo := &fakeRepository{err: repositoryErr}
	resolver := &fakeSpaceAccessResolver{access: SpaceAccess{SpaceID: 7}}
	uc, _ := NewRequestsUseCase(repo, resolver)

	_, err := uc.ListSpaceRequests(context.Background(), ListSpaceRequestsQuery{Slug: "worknet"})
	if !errors.Is(err, repositoryErr) {
		t.Fatalf("error = %v", err)
	}
}

type fakeRepository struct {
	page       domain.Page
	err        error
	lastParams domain.ListBySpaceParams
	calls      int
}

func (f *fakeRepository) ListBySpace(_ context.Context, params domain.ListBySpaceParams) (domain.Page, error) {
	f.calls++
	f.lastParams = params
	return f.page, f.err
}

type fakeSpaceAccessResolver struct {
	access       SpaceAccess
	err          error
	lastSlug     string
	lastViewerID *int64
	calls        int
}

func (f *fakeSpaceAccessResolver) ResolveAccessibleSpace(_ context.Context, slug string, viewerID *int64) (SpaceAccess, error) {
	f.calls++
	f.lastSlug = slug
	f.lastViewerID = viewerID
	return f.access, f.err
}
