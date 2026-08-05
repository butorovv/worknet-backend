package domain

import "time"

type RequestType string

const (
	RequestTypeLookingForContributor RequestType = "looking_for_contributor"
	RequestTypeLookingForFeedback    RequestType = "looking_for_feedback"
	RequestTypeLookingForCoauthor    RequestType = "looking_for_coauthor"
	RequestTypeLookingForReviewer    RequestType = "looking_for_reviewer"
	RequestTypeLookingForTeammate    RequestType = "looking_for_teammate"
	RequestTypeOfferingHelp          RequestType = "offering_help"
	RequestTypeOther                 RequestType = "other"
)

type RequestStatus string

const (
	RequestStatusDraft    RequestStatus = "draft"
	RequestStatusOpen     RequestStatus = "open"
	RequestStatusClosed   RequestStatus = "closed"
	RequestStatusArchived RequestStatus = "archived"
)

type RequestVisibility string

const (
	RequestVisibilityPublic     RequestVisibility = "public"
	RequestVisibilityRegistered RequestVisibility = "registered"
	RequestVisibilityMembers    RequestVisibility = "members"
	RequestVisibilityPrivate    RequestVisibility = "private"
)

type RequestFormat string

const (
	RequestFormatPortfolio         RequestFormat = "portfolio"
	RequestFormatPaid              RequestFormat = "paid"
	RequestFormatCoauthorship      RequestFormat = "coauthorship"
	RequestFormatFeedback          RequestFormat = "feedback"
	RequestFormatResearch          RequestFormat = "research"
	RequestFormatLearning          RequestFormat = "learning"
	RequestFormatEquity            RequestFormat = "equity"
	RequestFormatOpenCollaboration RequestFormat = "open_collaboration"
	RequestFormatOther             RequestFormat = "other"
)

type Request struct {
	ID                 int64
	SpaceID            int64
	OwnerID            int64
	Type               RequestType
	Status             RequestStatus
	Visibility         RequestVisibility
	Title              string
	Description        string
	ExpectedCommitment *string
	Format             RequestFormat
	ResponsesCount     int
	PublishedAt        *time.Time
	ClosedAt           *time.Time
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type ListBySpaceParams struct {
	SpaceID   int64
	ViewerID  *int64
	IsMember  bool
	CanManage bool
	Status    RequestStatus
	Limit     int
	Offset    int
}

type Page struct {
	Items []Request
	Total int
}

func (s RequestStatus) Valid() bool {
	switch s {
	case RequestStatusDraft, RequestStatusOpen, RequestStatusClosed, RequestStatusArchived:
		return true
	default:
		return false
	}
}
