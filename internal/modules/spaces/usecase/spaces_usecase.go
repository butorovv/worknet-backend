package usecase

import (
	"context"
	"errors"
	"strings"

	"github.com/butorovv/worknet-backend/internal/modules/spaces/domain"
)

const (
	defaultListLimit = 20
	maxListLimit     = 100
)

type spacesUseCase struct {
	repo  SpacesRepository
	clock Clock
}

func NewSpacesUseCase(repo SpacesRepository) (*spacesUseCase, error) {
	return newSpacesUseCase(repo, RealClock{})
}

func newSpacesUseCase(repo SpacesRepository, clock Clock) (*spacesUseCase, error) {
	if repo == nil {
		return nil, errors.New("spaces repository is required")
	}
	if clock == nil {
		clock = RealClock{}
	}
	return &spacesUseCase{repo: repo, clock: clock}, nil
}

func (s *spacesUseCase) CreateSpace(ctx context.Context, cmd CreateSpaceCommand) (SpaceDetails, error) {
	if !cmd.Viewer.IsAuthenticated || cmd.Viewer.UserID <= 0 {
		return SpaceDetails{}, domain.ErrForbidden
	}

	space := domain.Space{
		OwnerID:          cmd.Viewer.UserID,
		Slug:             strings.TrimSpace(cmd.Slug),
		Title:            strings.TrimSpace(cmd.Title),
		ShortDescription: strings.TrimSpace(cmd.ShortDescription),
		Description:      strings.TrimSpace(cmd.Description),
		WhyItMatters:     strings.TrimSpace(cmd.WhyItMatters),
		Status:           domain.SpaceStatusActive,
		Visibility:       cmd.Visibility,
		Stage:            cmd.Stage,
		CoverIcon:        cmd.CoverIcon,
		AccentColor:      cmd.AccentColor,
	}
	applyCreateDefaults(&space)
	if err := space.Validate(); err != nil {
		return SpaceDetails{}, err
	}

	stack, err := domain.NormalizeStack(cmd.Stack)
	if err != nil {
		return SpaceDetails{}, err
	}
	links, err := domain.NormalizeLinks(cmd.Links)
	if err != nil {
		return SpaceDetails{}, err
	}

	created, err := s.repo.Create(ctx, space, stack, links)
	if err != nil {
		return SpaceDetails{}, err
	}
	return s.loadDetails(ctx, created)
}

func (s *spacesUseCase) GetSpace(ctx context.Context, cmd GetSpaceCommand) (SpaceDetails, error) {
	space, err := s.getAccessibleSpace(ctx, strings.TrimSpace(cmd.Slug), cmd.Viewer)
	if err != nil {
		return SpaceDetails{}, err
	}
	return s.loadDetails(ctx, space)
}

func (s *spacesUseCase) ListSpaces(ctx context.Context, cmd ListSpacesCommand) ([]domain.Space, error) {
	status := cmd.Status
	if status == "" {
		status = domain.SpaceStatusActive
	}
	visibility := cmd.Visibility
	if visibility == "" {
		visibility = domain.SpaceVisibilityPublic
	}
	limit := cmd.Limit
	if limit == 0 {
		limit = defaultListLimit
	}

	if status == domain.SpaceStatusArchived || !status.Valid() || !visibility.Valid() || limit < 1 || limit > maxListLimit || cmd.Offset < 0 {
		return nil, domain.ErrInvalidInput
	}
	if visibility == domain.SpaceVisibilityPrivate && !cmd.Viewer.IsAuthenticated {
		return nil, domain.ErrForbidden
	}

	var viewerID *domain.UserID
	if cmd.Viewer.IsAuthenticated {
		viewerID = &cmd.Viewer.UserID
	}
	return s.repo.List(ctx, domain.SpaceListFilter{
		ViewerID:   viewerID,
		Status:     status,
		Visibility: visibility,
		Limit:      limit,
		Offset:     cmd.Offset,
	})
}

func (s *spacesUseCase) UpdateSpace(ctx context.Context, cmd UpdateSpaceCommand) (SpaceDetails, error) {
	space, err := s.repo.GetBySlug(ctx, strings.TrimSpace(cmd.Slug))
	if err != nil {
		return SpaceDetails{}, err
	}
	if space.Status == domain.SpaceStatusArchived {
		return SpaceDetails{}, domain.ErrSpaceArchived
	}
	if !cmd.Viewer.IsAuthenticated {
		return SpaceDetails{}, domain.ErrForbidden
	}
	canManage, err := s.repo.IsOwnerOrAdmin(ctx, space.ID, cmd.Viewer.UserID)
	if err != nil {
		return SpaceDetails{}, err
	}
	if !canManage {
		return SpaceDetails{}, domain.ErrForbidden
	}
	if !hasUpdate(cmd) {
		return SpaceDetails{}, domain.ErrInvalidInput
	}

	applyUpdate(&space, cmd)
	if space.Status == domain.SpaceStatusArchived {
		return SpaceDetails{}, domain.ErrInvalidInput
	}
	if err := space.Validate(); err != nil {
		return SpaceDetails{}, err
	}

	var stack *[]domain.SpaceStackItem
	if cmd.Stack != nil {
		normalized, err := domain.NormalizeStack(*cmd.Stack)
		if err != nil {
			return SpaceDetails{}, err
		}
		stack = &normalized
	}
	var links *[]domain.SpaceLink
	if cmd.Links != nil {
		normalized, err := domain.NormalizeLinks(*cmd.Links)
		if err != nil {
			return SpaceDetails{}, err
		}
		links = &normalized
	}

	updated, err := s.repo.Update(ctx, space, stack, links)
	if err != nil {
		return SpaceDetails{}, err
	}
	return s.loadDetails(ctx, updated)
}

func (s *spacesUseCase) ArchiveSpace(ctx context.Context, cmd ArchiveSpaceCommand) error {
	space, err := s.repo.GetBySlug(ctx, strings.TrimSpace(cmd.Slug))
	if err != nil {
		return err
	}
	if space.Status == domain.SpaceStatusArchived {
		return domain.ErrSpaceArchived
	}
	if !cmd.Viewer.IsAuthenticated {
		return domain.ErrForbidden
	}
	canManage, err := s.repo.IsOwnerOrAdmin(ctx, space.ID, cmd.Viewer.UserID)
	if err != nil {
		return err
	}
	if !canManage {
		return domain.ErrForbidden
	}
	_, err = s.repo.Archive(ctx, space.Slug, s.clock.Now())
	return err
}

func (s *spacesUseCase) getAccessibleSpace(ctx context.Context, slug string, viewer Viewer) (domain.Space, error) {
	space, err := s.repo.GetBySlug(ctx, slug)
	if err != nil {
		return domain.Space{}, err
	}
	if space.Status == domain.SpaceStatusArchived {
		return domain.Space{}, domain.ErrSpaceArchived
	}
	if space.Visibility == domain.SpaceVisibilityPublic {
		return space, nil
	}
	if !viewer.IsAuthenticated {
		return domain.Space{}, domain.ErrForbidden
	}
	isMember, err := s.repo.IsMember(ctx, space.ID, viewer.UserID)
	if err != nil {
		return domain.Space{}, err
	}
	if !isMember {
		return domain.Space{}, domain.ErrForbidden
	}
	return space, nil
}

func (s *spacesUseCase) loadDetails(ctx context.Context, space domain.Space) (SpaceDetails, error) {
	stack, err := s.repo.ListStack(ctx, space.ID)
	if err != nil {
		return SpaceDetails{}, err
	}
	links, err := s.repo.ListLinks(ctx, space.ID)
	if err != nil {
		return SpaceDetails{}, err
	}
	return SpaceDetails{
		Space: space,
		Stack: stack,
		Links: links,
	}, nil
}

func applyCreateDefaults(space *domain.Space) {
	if space.Visibility == "" {
		space.Visibility = domain.SpaceVisibilityPublic
	}
	if space.Stage == "" {
		space.Stage = domain.SpaceStageIdea
	}
	if space.CoverIcon == "" {
		space.CoverIcon = domain.CoverIconOther
	}
	if space.AccentColor == "" {
		space.AccentColor = domain.AccentColorGray
	}
}

func applyUpdate(space *domain.Space, cmd UpdateSpaceCommand) {
	if cmd.Title != nil {
		space.Title = strings.TrimSpace(*cmd.Title)
	}
	if cmd.ShortDescription != nil {
		space.ShortDescription = strings.TrimSpace(*cmd.ShortDescription)
	}
	if cmd.Description != nil {
		space.Description = strings.TrimSpace(*cmd.Description)
	}
	if cmd.WhyItMatters != nil {
		space.WhyItMatters = strings.TrimSpace(*cmd.WhyItMatters)
	}
	if cmd.Visibility != nil {
		space.Visibility = *cmd.Visibility
	}
	if cmd.Stage != nil {
		space.Stage = *cmd.Stage
	}
	if cmd.Status != nil {
		space.Status = *cmd.Status
	}
	if cmd.CoverIcon != nil {
		space.CoverIcon = *cmd.CoverIcon
	}
	if cmd.AccentColor != nil {
		space.AccentColor = *cmd.AccentColor
	}
}

func hasUpdate(cmd UpdateSpaceCommand) bool {
	return cmd.Title != nil ||
		cmd.ShortDescription != nil ||
		cmd.Description != nil ||
		cmd.WhyItMatters != nil ||
		cmd.Visibility != nil ||
		cmd.Stage != nil ||
		cmd.Status != nil ||
		cmd.CoverIcon != nil ||
		cmd.AccentColor != nil ||
		cmd.Stack != nil ||
		cmd.Links != nil
}
