package transport

import "time"

type SpaceLinkRequest struct {
	Type  string `json:"type"`
	Title string `json:"title"`
	URL   string `json:"url"`
}

type CreateSpaceRequest struct {
	Title            string             `json:"title"`
	Slug             string             `json:"slug"`
	ShortDescription string             `json:"short_description"`
	Description      string             `json:"description"`
	WhyItMatters     string             `json:"why_it_matters"`
	Visibility       string             `json:"visibility"`
	Stage            string             `json:"stage"`
	CoverIcon        string             `json:"cover_icon"`
	AccentColor      string             `json:"accent_color"`
	Stack            []string           `json:"stack"`
	Links            []SpaceLinkRequest `json:"links"`
}

type UpdateSpaceRequest struct {
	Title            *string             `json:"title"`
	ShortDescription *string             `json:"short_description"`
	Description      *string             `json:"description"`
	WhyItMatters     *string             `json:"why_it_matters"`
	Visibility       *string             `json:"visibility"`
	Stage            *string             `json:"stage"`
	Status           *string             `json:"status"`
	CoverIcon        *string             `json:"cover_icon"`
	AccentColor      *string             `json:"accent_color"`
	Stack            *[]string           `json:"stack"`
	Links            *[]SpaceLinkRequest `json:"links"`
}

type SpaceResponse struct {
	ID               int64      `json:"id"`
	OwnerID          int64      `json:"owner_id"`
	Slug             string     `json:"slug"`
	Title            string     `json:"title"`
	ShortDescription string     `json:"short_description"`
	Description      string     `json:"description"`
	WhyItMatters     string     `json:"why_it_matters"`
	Status           string     `json:"status"`
	Visibility       string     `json:"visibility"`
	Stage            string     `json:"stage"`
	CoverIcon        string     `json:"cover_icon"`
	AccentColor      string     `json:"accent_color"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	ArchivedAt       *time.Time `json:"archived_at,omitempty"`
}

type SpaceLinkResponse struct {
	ID    int64  `json:"id"`
	Type  string `json:"type"`
	Title string `json:"title"`
	URL   string `json:"url"`
}

type SpaceDetailsResponse struct {
	Space SpaceResponse       `json:"space"`
	Stack []string            `json:"stack"`
	Links []SpaceLinkResponse `json:"links"`
}

type SpacesResponse struct {
	Spaces []SpaceResponse `json:"spaces"`
}
