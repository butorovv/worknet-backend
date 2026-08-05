package requestshttp

import (
	"github.com/butorovv/worknet-backend/internal/modules/requests/delivery/transport"
	"github.com/butorovv/worknet-backend/internal/modules/requests/usecase"
)

func toSpaceRequestsResponse(result usecase.ListSpaceRequestsResult) transport.SpaceRequestsResponse {
	items := make([]transport.RequestResponse, 0, len(result.Items))
	for _, request := range result.Items {
		items = append(items, transport.RequestResponse{
			ID:                 request.ID,
			SpaceID:            request.SpaceID,
			OwnerID:            request.OwnerID,
			Type:               string(request.Type),
			Status:             string(request.Status),
			Visibility:         string(request.Visibility),
			Title:              request.Title,
			Description:        request.Description,
			ExpectedCommitment: request.ExpectedCommitment,
			Format:             string(request.Format),
			ResponsesCount:     request.ResponsesCount,
			PublishedAt:        request.PublishedAt,
			ClosedAt:           request.ClosedAt,
			CreatedAt:          request.CreatedAt,
			UpdatedAt:          request.UpdatedAt,
		})
	}
	return transport.SpaceRequestsResponse{
		Items: items,
		Pagination: transport.PaginationResponse{
			Limit:  result.Pagination.Limit,
			Offset: result.Pagination.Offset,
			Total:  result.Pagination.Total,
		},
	}
}
