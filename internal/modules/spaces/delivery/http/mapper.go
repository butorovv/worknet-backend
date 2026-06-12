package spaceshttp

import (
	"github.com/butorovv/worknet-backend/internal/modules/spaces/delivery/transport"
	"github.com/butorovv/worknet-backend/internal/modules/spaces/domain"
	"github.com/butorovv/worknet-backend/internal/modules/spaces/usecase"
)

func toSpaceDetailsResponse(details usecase.SpaceDetails) transport.SpaceDetailsResponse {
	return transport.SpaceDetailsResponse{
		Space: toSpaceResponse(details.Space),
		Stack: toStackResponse(details.Stack),
		Links: toLinksResponse(details.Links),
	}
}

func toSpaceResponse(space domain.Space) transport.SpaceResponse {
	return transport.SpaceResponse{
		ID:               int64(space.ID),
		OwnerID:          int64(space.OwnerID),
		Slug:             space.Slug,
		Title:            space.Title,
		ShortDescription: space.ShortDescription,
		Description:      space.Description,
		WhyItMatters:     space.WhyItMatters,
		Status:           string(space.Status),
		Visibility:       string(space.Visibility),
		Stage:            string(space.Stage),
		CoverIcon:        string(space.CoverIcon),
		AccentColor:      string(space.AccentColor),
		CreatedAt:        space.CreatedAt,
		UpdatedAt:        space.UpdatedAt,
		ArchivedAt:       space.ArchivedAt,
	}
}

func toSpacesResponse(spaces []domain.Space) transport.SpacesResponse {
	response := make([]transport.SpaceResponse, 0, len(spaces))
	for _, space := range spaces {
		response = append(response, toSpaceResponse(space))
	}
	return transport.SpacesResponse{Spaces: response}
}

func toLinksResponse(links []domain.SpaceLink) []transport.SpaceLinkResponse {
	response := make([]transport.SpaceLinkResponse, 0, len(links))
	for _, link := range links {
		response = append(response, transport.SpaceLinkResponse{
			ID:    link.ID,
			Type:  string(link.Type),
			Title: link.Title,
			URL:   link.URL,
		})
	}
	return response
}

func toStackResponse(stack []domain.SpaceStackItem) []string {
	response := make([]string, 0, len(stack))
	for _, item := range stack {
		response = append(response, item.Name)
	}
	return response
}

func toOverviewResponse(overview usecase.SpaceOverview) transport.SpaceOverviewResponse {
	members := make([]transport.SpaceMemberPreviewResponse, 0, len(overview.MembersPreview))
	for _, member := range overview.MembersPreview {
		members = append(members, transport.SpaceMemberPreviewResponse{
			ID:        int64(member.ID),
			Username:  member.Username,
			Name:      member.Name,
			Role:      string(member.Role),
			Title:     member.Title,
			AvatarURL: member.AvatarURL,
		})
	}

	var role *string
	if overview.Viewer.Role != nil {
		value := string(*overview.Viewer.Role)
		role = &value
	}

	permissions := overview.Viewer.Permissions
	return transport.SpaceOverviewResponse{
		Space: toSpaceResponse(overview.Space),
		Owner: transport.UserBriefResponse{
			ID:        int64(overview.Owner.ID),
			Username:  overview.Owner.Username,
			Name:      overview.Owner.Name,
			AvatarURL: overview.Owner.AvatarURL,
		},
		MembersPreview: members,
		Stats: transport.SpaceStatsResponse{
			MembersCount:                overview.Stats.MembersCount,
			OpenRequestsCount:           overview.Stats.OpenRequestsCount,
			ConfirmedContributionsCount: overview.Stats.ConfirmedContributionsCount,
			LinksCount:                  overview.Stats.LinksCount,
			ArtifactsCount:              overview.Stats.ArtifactsCount,
		},
		Stack:                toStackResponse(overview.Stack),
		Links:                toLinksResponse(overview.Links),
		Milestones:           []any{},
		ArtifactsPreview:     []any{},
		OpenRequestsPreview:  []any{},
		CompletedWorkPreview: []any{},
		Activity: transport.ActivityResponse{
			Range: overview.Activity.Range,
			Days:  []any{},
		},
		Viewer: transport.ViewerResponse{
			IsAuthenticated: overview.Viewer.IsAuthenticated,
			IsMember:        overview.Viewer.IsMember,
			Role:            role,
			Permissions: transport.PermissionsResponse{
				CanEditSpace:          permissions.CanEditSpace,
				CanArchiveSpace:       permissions.CanArchiveSpace,
				CanManageMembers:      permissions.CanManageMembers,
				CanManageLinks:        permissions.CanManageLinks,
				CanManageRoadmap:      permissions.CanManageRoadmap,
				CanCreateRequest:      permissions.CanCreateRequest,
				CanCreateContribution: permissions.CanCreateContribution,
				CanApply:              permissions.CanApply,
				CanFollow:             permissions.CanFollow,
			},
		},
	}
}

func linksFromRequest(requests []transport.SpaceLinkRequest) []domain.SpaceLink {
	links := make([]domain.SpaceLink, 0, len(requests))
	for _, request := range requests {
		links = append(links, domain.SpaceLink{
			Type:  domain.LinkType(request.Type),
			Title: request.Title,
			URL:   request.URL,
		})
	}
	return links
}
