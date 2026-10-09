package transport

type UserBriefResponse struct {
	ID        int64   `json:"id"`
	Username  string  `json:"username"`
	Name      *string `json:"name"`
	AvatarURL *string `json:"avatar_url"`
}

type SpaceMemberPreviewResponse struct {
	ID        int64   `json:"id"`
	Username  string  `json:"username"`
	Name      *string `json:"name"`
	Role      string  `json:"role"`
	Title     *string `json:"title,omitempty"`
	AvatarURL *string `json:"avatar_url"`
}

type SpaceStatsResponse struct {
	MembersCount                int `json:"members_count"`
	OpenRequestsCount           int `json:"open_requests_count"`
	ConfirmedContributionsCount int `json:"confirmed_contributions_count"`
	LinksCount                  int `json:"links_count"`
	ArtifactsCount              int `json:"artifacts_count"`
}

type PermissionsResponse struct {
	CanEditSpace          bool `json:"can_edit_space"`
	CanArchiveSpace       bool `json:"can_archive_space"`
	CanManageMembers      bool `json:"can_manage_members"`
	CanManageLinks        bool `json:"can_manage_links"`
	CanManageRoadmap      bool `json:"can_manage_roadmap"`
	CanCreateRequest      bool `json:"can_create_request"`
	CanCreateContribution bool `json:"can_create_contribution"`
	CanApply              bool `json:"can_apply"`
	CanFollow             bool `json:"can_follow"`
}

type ViewerResponse struct {
	IsAuthenticated bool                `json:"is_authenticated"`
	IsMember        bool                `json:"is_member"`
	Role            *string             `json:"role"`
	Permissions     PermissionsResponse `json:"permissions"`
}

type ActivityResponse struct {
	Range string `json:"range"`
	Days  []any  `json:"days"`
}

type SpaceOverviewResponse struct {
	Space                SpaceResponse                `json:"space"`
	Owner                UserBriefResponse            `json:"owner"`
	MembersPreview       []SpaceMemberPreviewResponse `json:"members_preview"`
	Stats                SpaceStatsResponse           `json:"stats"`
	Stack                []string                     `json:"stack"`
	Links                []SpaceLinkResponse          `json:"links"`
	Milestones           []any                        `json:"milestones"`
	ArtifactsPreview     []any                        `json:"artifacts_preview"`
	OpenRequestsPreview  []any                        `json:"open_requests_preview"`
	CompletedWorkPreview []any                        `json:"completed_work_preview"`
	Activity             ActivityResponse             `json:"activity"`
	Viewer               ViewerResponse               `json:"viewer"`
}
