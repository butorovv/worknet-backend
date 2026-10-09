package transport

import "time"

type RequestResponse struct {
	ID                 int64      `json:"id"`
	SpaceID            int64      `json:"space_id"`
	OwnerID            int64      `json:"owner_id"`
	Type               string     `json:"type"`
	Status             string     `json:"status"`
	Visibility         string     `json:"visibility"`
	Title              string     `json:"title"`
	Description        string     `json:"description"`
	ExpectedCommitment *string    `json:"expected_commitment"`
	Format             string     `json:"format"`
	ResponsesCount     int        `json:"responses_count"`
	PublishedAt        *time.Time `json:"published_at"`
	ClosedAt           *time.Time `json:"closed_at"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

type PaginationResponse struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
	Total  int `json:"total"`
}

type SpaceRequestsResponse struct {
	Items      []RequestResponse  `json:"items"`
	Pagination PaginationResponse `json:"pagination"`
}
