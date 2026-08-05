package usecase

import (
	"context"
	"fmt"
	"strings"

	"github.com/butorovv/worknet-backend/internal/modules/spaces/domain"
)

const membersPreviewLimit = 8

func (s *spacesUseCase) GetSpaceOverview(ctx context.Context, cmd GetSpaceOverviewCommand) (SpaceOverview, error) {
	space, err := s.getAccessibleSpace(ctx, strings.TrimSpace(cmd.Slug), cmd.Viewer)
	if err != nil {
		return SpaceOverview{}, err
	}

	members, err := s.repo.ListMembers(ctx, space.ID)
	if err != nil {
		return SpaceOverview{}, err
	}
	membersCount, err := s.repo.CountMembers(ctx, space.ID)
	if err != nil {
		return SpaceOverview{}, err
	}
	stack, err := s.repo.ListStack(ctx, space.ID)
	if err != nil {
		return SpaceOverview{}, err
	}
	links, err := s.repo.ListLinks(ctx, space.ID)
	if err != nil {
		return SpaceOverview{}, err
	}

	var owner UserBrief
	var viewerRole *domain.MemberRole
	memberPreviews := make([]MemberPreview, 0, min(len(members), membersPreviewLimit))
	for _, member := range members {
		if member.Role == domain.MemberRoleOwner {
			owner = UserBrief{ID: member.UserID, Username: member.Username}
		}
		if cmd.Viewer.IsAuthenticated && member.UserID == cmd.Viewer.UserID {
			role := member.Role
			viewerRole = &role
		}
		if len(memberPreviews) < membersPreviewLimit {
			memberPreviews = append(memberPreviews, MemberPreview{
				ID:       member.UserID,
				Username: member.Username,
				Role:     member.Role,
				Title:    member.Title,
			})
		}
	}
	if owner.ID == 0 {
		return SpaceOverview{}, fmt.Errorf("spaces overview: owner membership not found")
	}

	return SpaceOverview{
		Space:          space,
		Owner:          owner,
		MembersPreview: memberPreviews,
		Stats: SpaceStats{
			MembersCount: membersCount,
			LinksCount:   len(links),
		},
		Stack:                stack,
		Links:                links,
		Milestones:           []MilestonePreview{},
		ArtifactsPreview:     []ArtifactPreview{},
		OpenRequestsPreview:  []RequestPreview{},
		CompletedWorkPreview: []ContributionPreview{},
		Activity: ActivitySummary{
			Range: "30d",
			Days:  []ActivityDay{},
		},
		Viewer: ViewerInfo{
			IsAuthenticated: cmd.Viewer.IsAuthenticated,
			IsMember:        viewerRole != nil,
			Role:            viewerRole,
			Permissions:     domain.PermissionsForRole(viewerRole),
		},
	}, nil
}
