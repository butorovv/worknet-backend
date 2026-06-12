package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/butorovv/worknet-backend/internal/modules/spaces/domain"
)

func TestCreateSpaceRequiresAuthenticatedViewer(t *testing.T) {
	uc, err := newSpacesUseCase(&fakeSpacesRepository{}, fixedClock{})
	if err != nil {
		t.Fatal(err)
	}

	_, err = uc.CreateSpace(context.Background(), CreateSpaceCommand{})
	if err != domain.ErrForbidden {
		t.Fatalf("error = %v, want %v", err, domain.ErrForbidden)
	}
}

func TestPrivateSpaceRequiresMembership(t *testing.T) {
	repo := &fakeSpacesRepository{
		space: domain.Space{
			ID:               1,
			OwnerID:          10,
			Slug:             "private-space",
			Title:            "Private Space",
			ShortDescription: "A private project workspace.",
			Status:           domain.SpaceStatusActive,
			Visibility:       domain.SpaceVisibilityPrivate,
			Stage:            domain.SpaceStageIdea,
			CoverIcon:        domain.CoverIconOther,
			AccentColor:      domain.AccentColorGray,
		},
	}
	uc, _ := newSpacesUseCase(repo, fixedClock{})

	_, err := uc.GetSpace(context.Background(), GetSpaceCommand{
		Viewer: Viewer{UserID: 20, IsAuthenticated: true},
		Slug:   "private-space",
	})
	if err != domain.ErrForbidden {
		t.Fatalf("error = %v, want %v", err, domain.ErrForbidden)
	}
}

func TestOverviewOwnerPermissionsAndEmptyPreviews(t *testing.T) {
	repo := &fakeSpacesRepository{
		space: domain.Space{
			ID:               1,
			OwnerID:          10,
			Slug:             "worknet",
			Title:            "WorkNet",
			ShortDescription: "A public network for project work.",
			Status:           domain.SpaceStatusActive,
			Visibility:       domain.SpaceVisibilityPublic,
			Stage:            domain.SpaceStageMVP,
			CoverIcon:        domain.CoverIconLink,
			AccentColor:      domain.AccentColorMint,
		},
		members: []domain.SpaceMember{{
			SpaceID:  1,
			UserID:   10,
			Username: "owner",
			Role:     domain.MemberRoleOwner,
		}},
		member: true,
	}
	uc, _ := newSpacesUseCase(repo, fixedClock{})

	overview, err := uc.GetSpaceOverview(context.Background(), GetSpaceOverviewCommand{
		Viewer: Viewer{UserID: 10, IsAuthenticated: true},
		Slug:   "worknet",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !overview.Viewer.Permissions.CanEditSpace || !overview.Viewer.IsMember {
		t.Fatalf("viewer = %+v", overview.Viewer)
	}
	if overview.Milestones == nil || overview.ArtifactsPreview == nil || overview.OpenRequestsPreview == nil || overview.CompletedWorkPreview == nil || overview.Activity.Days == nil {
		t.Fatal("future preview slices must be non-nil")
	}
}

type fixedClock struct{}

func (fixedClock) Now() time.Time {
	return time.Date(2026, 6, 12, 12, 0, 0, 0, time.UTC)
}

type fakeSpacesRepository struct {
	space     domain.Space
	members   []domain.SpaceMember
	member    bool
	canManage bool
}

func (f *fakeSpacesRepository) Create(context.Context, domain.Space, []domain.SpaceStackItem, []domain.SpaceLink) (domain.Space, error) {
	return f.space, nil
}
func (f *fakeSpacesRepository) GetByID(context.Context, domain.SpaceID) (domain.Space, error) {
	return f.space, nil
}
func (f *fakeSpacesRepository) GetBySlug(context.Context, string) (domain.Space, error) {
	return f.space, nil
}
func (f *fakeSpacesRepository) List(context.Context, domain.SpaceListFilter) ([]domain.Space, error) {
	return []domain.Space{f.space}, nil
}
func (f *fakeSpacesRepository) Update(context.Context, domain.Space, *[]domain.SpaceStackItem, *[]domain.SpaceLink) (domain.Space, error) {
	return f.space, nil
}
func (f *fakeSpacesRepository) Archive(context.Context, string, time.Time) (domain.Space, error) {
	return f.space, nil
}
func (f *fakeSpacesRepository) AddMember(context.Context, domain.SpaceMember) error {
	return nil
}
func (f *fakeSpacesRepository) ListMembers(context.Context, domain.SpaceID) ([]domain.SpaceMember, error) {
	return f.members, nil
}
func (f *fakeSpacesRepository) CountMembers(context.Context, domain.SpaceID) (int, error) {
	return len(f.members), nil
}
func (f *fakeSpacesRepository) IsMember(context.Context, domain.SpaceID, domain.UserID) (bool, error) {
	return f.member, nil
}
func (f *fakeSpacesRepository) IsOwnerOrAdmin(context.Context, domain.SpaceID, domain.UserID) (bool, error) {
	return f.canManage, nil
}
func (f *fakeSpacesRepository) ReplaceStack(context.Context, domain.SpaceID, []domain.SpaceStackItem) error {
	return nil
}
func (f *fakeSpacesRepository) ListStack(context.Context, domain.SpaceID) ([]domain.SpaceStackItem, error) {
	return []domain.SpaceStackItem{}, nil
}
func (f *fakeSpacesRepository) ReplaceLinks(context.Context, domain.SpaceID, []domain.SpaceLink) error {
	return nil
}
func (f *fakeSpacesRepository) ListLinks(context.Context, domain.SpaceID) ([]domain.SpaceLink, error) {
	return []domain.SpaceLink{}, nil
}
