package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/butorovv/worknet-backend/internal/modules/spaces/domain"
)

func TestResolveSpaceAccessAnonymousPublicSpace(t *testing.T) {
	repo := &fakeSpacesRepository{space: accessTestSpace(domain.SpaceVisibilityPublic, domain.SpaceStatusActive)}
	uc, _ := newSpacesUseCase(repo, fixedClock{})

	access, err := uc.ResolveSpaceAccess(context.Background(), ResolveSpaceAccessQuery{Slug: "worknet"})
	if err != nil {
		t.Fatal(err)
	}
	if access != (SpaceAccess{SpaceID: 1}) {
		t.Fatalf("access = %+v", access)
	}
}

func TestResolveSpaceAccessAuthenticatedFlags(t *testing.T) {
	tests := []struct {
		name      string
		viewerID  domain.UserID
		member    bool
		manage    bool
		wantOwner bool
		wantAdmin bool
	}{
		{name: "non-member", viewerID: 20},
		{name: "member", viewerID: 20, member: true},
		{name: "owner", viewerID: 10, member: true, manage: true, wantOwner: true},
		{name: "admin", viewerID: 20, member: true, manage: true, wantAdmin: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeSpacesRepository{
				space:     accessTestSpace(domain.SpaceVisibilityPublic, domain.SpaceStatusActive),
				member:    tt.member,
				canManage: tt.manage,
			}
			uc, _ := newSpacesUseCase(repo, fixedClock{})

			access, err := uc.ResolveSpaceAccess(context.Background(), ResolveSpaceAccessQuery{
				Viewer: Viewer{UserID: tt.viewerID, IsAuthenticated: true},
				Slug:   "worknet",
			})
			if err != nil {
				t.Fatal(err)
			}
			if access.SpaceID != 1 || access.IsMember != tt.member || access.IsOwner != tt.wantOwner || access.IsAdmin != tt.wantAdmin {
				t.Fatalf("access = %+v", access)
			}
		})
	}
}

func TestResolveSpaceAccessPrivateSpaceUsesCorePolicy(t *testing.T) {
	t.Run("active member", func(t *testing.T) {
		repo := &fakeSpacesRepository{
			space:  accessTestSpace(domain.SpaceVisibilityPrivate, domain.SpaceStatusActive),
			member: true,
		}
		uc, _ := newSpacesUseCase(repo, fixedClock{})

		access, err := uc.ResolveSpaceAccess(context.Background(), ResolveSpaceAccessQuery{
			Viewer: Viewer{UserID: 20, IsAuthenticated: true}, Slug: "worknet",
		})
		if err != nil {
			t.Fatal(err)
		}
		if !access.IsMember {
			t.Fatalf("access = %+v", access)
		}
	})

	t.Run("outsider", func(t *testing.T) {
		repo := &fakeSpacesRepository{space: accessTestSpace(domain.SpaceVisibilityPrivate, domain.SpaceStatusActive)}
		uc, _ := newSpacesUseCase(repo, fixedClock{})

		_, err := uc.ResolveSpaceAccess(context.Background(), ResolveSpaceAccessQuery{
			Viewer: Viewer{UserID: 20, IsAuthenticated: true}, Slug: "worknet",
		})
		if !errors.Is(err, domain.ErrForbidden) {
			t.Fatalf("error = %v", err)
		}
	})
}

func TestResolveSpaceAccessErrors(t *testing.T) {
	tests := []struct {
		name    string
		slug    string
		repo    *fakeSpacesRepository
		wantErr error
	}{
		{name: "invalid slug", slug: "INVALID", repo: &fakeSpacesRepository{}, wantErr: domain.ErrInvalidInput},
		{name: "unknown Space", slug: "unknown", repo: &fakeSpacesRepository{getBySlugErr: domain.ErrSpaceNotFound}, wantErr: domain.ErrSpaceNotFound},
		{name: "archived Space", slug: "worknet", repo: &fakeSpacesRepository{space: accessTestSpace(domain.SpaceVisibilityPublic, domain.SpaceStatusArchived)}, wantErr: domain.ErrSpaceArchived},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc, _ := newSpacesUseCase(tt.repo, fixedClock{})
			_, err := uc.ResolveSpaceAccess(context.Background(), ResolveSpaceAccessQuery{Slug: tt.slug})
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func accessTestSpace(visibility domain.SpaceVisibility, status domain.SpaceStatus) domain.Space {
	return domain.Space{
		ID: 1, OwnerID: 10, Slug: "worknet", Title: "WorkNet",
		ShortDescription: "A project collaboration workspace.",
		Status:           status, Visibility: visibility, Stage: domain.SpaceStageMVP,
		CoverIcon: domain.CoverIconLink, AccentColor: domain.AccentColorMint,
	}
}
