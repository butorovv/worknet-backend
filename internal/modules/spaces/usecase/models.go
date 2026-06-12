package usecase

import (
	"time"

	"github.com/butorovv/worknet-backend/internal/modules/spaces/domain"
)

type Viewer struct {
	UserID          domain.UserID
	IsAuthenticated bool
}

type CreateSpaceCommand struct {
	Viewer           Viewer
	Title            string
	Slug             string
	ShortDescription string
	Description      string
	WhyItMatters     string
	Visibility       domain.SpaceVisibility
	Stage            domain.SpaceStage
	CoverIcon        domain.CoverIcon
	AccentColor      domain.AccentColor
	Stack            []string
	Links            []domain.SpaceLink
}

type GetSpaceCommand struct {
	Viewer Viewer
	Slug   string
}

type ListSpacesCommand struct {
	Viewer     Viewer
	Status     domain.SpaceStatus
	Visibility domain.SpaceVisibility
	Limit      int
	Offset     int
}

type UpdateSpaceCommand struct {
	Viewer           Viewer
	Slug             string
	Title            *string
	ShortDescription *string
	Description      *string
	WhyItMatters     *string
	Visibility       *domain.SpaceVisibility
	Stage            *domain.SpaceStage
	Status           *domain.SpaceStatus
	CoverIcon        *domain.CoverIcon
	AccentColor      *domain.AccentColor
	Stack            *[]string
	Links            *[]domain.SpaceLink
}

type ArchiveSpaceCommand struct {
	Viewer Viewer
	Slug   string
}

type GetSpaceOverviewCommand struct {
	Viewer Viewer
	Slug   string
}

type SpaceDetails struct {
	Space domain.Space
	Stack []domain.SpaceStackItem
	Links []domain.SpaceLink
}

type UserBrief struct {
	ID        domain.UserID
	Username  string
	Name      *string
	AvatarURL *string
}

type MemberPreview struct {
	ID        domain.UserID
	Username  string
	Name      *string
	Role      domain.MemberRole
	Title     *string
	AvatarURL *string
}

type SpaceStats struct {
	MembersCount                int
	OpenRequestsCount           int
	ConfirmedContributionsCount int
	LinksCount                  int
	ArtifactsCount              int
}

type MilestonePreview struct{}
type ArtifactPreview struct{}
type RequestPreview struct{}
type ContributionPreview struct{}
type ActivityDay struct{}

type ActivitySummary struct {
	Range string
	Days  []ActivityDay
}

type ViewerInfo struct {
	IsAuthenticated bool
	IsMember        bool
	Role            *domain.MemberRole
	Permissions     domain.Permissions
}

type SpaceOverview struct {
	Space                domain.Space
	Owner                UserBrief
	MembersPreview       []MemberPreview
	Stats                SpaceStats
	Stack                []domain.SpaceStackItem
	Links                []domain.SpaceLink
	Milestones           []MilestonePreview
	ArtifactsPreview     []ArtifactPreview
	OpenRequestsPreview  []RequestPreview
	CompletedWorkPreview []ContributionPreview
	Activity             ActivitySummary
	Viewer               ViewerInfo
}

type RealClock struct{}

func (RealClock) Now() time.Time {
	return time.Now().UTC()
}
