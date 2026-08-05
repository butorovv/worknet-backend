package domain

import (
	"net/url"
	"strings"
	"time"
)

type LinkType string

const (
	LinkTypeGitHub LinkType = "github"
	LinkTypeFigma  LinkType = "figma"
	LinkTypeNotion LinkType = "notion"
	LinkTypeSite   LinkType = "site"
	LinkTypeDocs   LinkType = "docs"
	LinkTypeOther  LinkType = "other"
)

type SpaceLink struct {
	ID        int64
	SpaceID   SpaceID
	Type      LinkType
	Title     string
	URL       string
	Position  int
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (l SpaceLink) Validate() error {
	if !l.Type.Valid() || !validLength(strings.TrimSpace(l.Title), 1, 80) || l.Position < 0 {
		return ErrInvalidInput
	}

	parsed, err := url.ParseRequestURI(strings.TrimSpace(l.URL))
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return ErrInvalidInput
	}
	return nil
}

func (t LinkType) Valid() bool {
	switch t {
	case LinkTypeGitHub, LinkTypeFigma, LinkTypeNotion, LinkTypeSite, LinkTypeDocs, LinkTypeOther:
		return true
	default:
		return false
	}
}

func NormalizeLinks(links []SpaceLink) ([]SpaceLink, error) {
	if len(links) > 8 {
		return nil, ErrInvalidInput
	}

	result := make([]SpaceLink, 0, len(links))
	for index, link := range links {
		link.Title = strings.TrimSpace(link.Title)
		link.URL = strings.TrimSpace(link.URL)
		link.Position = index
		if err := link.Validate(); err != nil {
			return nil, err
		}
		result = append(result, link)
	}
	return result, nil
}
