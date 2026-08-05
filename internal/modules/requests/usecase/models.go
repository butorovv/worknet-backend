package usecase

import "github.com/butorovv/worknet-backend/internal/modules/requests/domain"

type SpaceAccess struct {
	SpaceID  int64
	IsMember bool
	IsOwner  bool
	IsAdmin  bool
}

type ListSpaceRequestsQuery struct {
	ViewerID *int64
	Slug     string
	Status   domain.RequestStatus
	Limit    int
	Offset   int
}

type Pagination struct {
	Limit  int
	Offset int
	Total  int
}

type ListSpaceRequestsResult struct {
	Items      []domain.Request
	Pagination Pagination
}
