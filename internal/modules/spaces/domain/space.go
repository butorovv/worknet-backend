package domain

import (
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

type SpaceID int64
type UserID int64

type SpaceStatus string

const (
	SpaceStatusActive   SpaceStatus = "active"
	SpaceStatusPaused   SpaceStatus = "paused"
	SpaceStatusArchived SpaceStatus = "archived"
)

type SpaceVisibility string

const (
	SpaceVisibilityPublic  SpaceVisibility = "public"
	SpaceVisibilityPrivate SpaceVisibility = "private"
)

type SpaceStage string

const (
	SpaceStageIdea      SpaceStage = "idea"
	SpaceStagePrototype SpaceStage = "prototype"
	SpaceStageMVP       SpaceStage = "mvp"
	SpaceStageBeta      SpaceStage = "beta"
	SpaceStageGrowth    SpaceStage = "growth"
	SpaceStageScale     SpaceStage = "scale"
)

type AccentColor string

const (
	AccentColorMint   AccentColor = "mint"
	AccentColorBlue   AccentColor = "blue"
	AccentColorPurple AccentColor = "purple"
	AccentColorOrange AccentColor = "orange"
	AccentColorGray   AccentColor = "gray"
)

type CoverIcon string

const (
	CoverIconLink      CoverIcon = "link"
	CoverIconSprout    CoverIcon = "sprout"
	CoverIconRocket    CoverIcon = "rocket"
	CoverIconCode      CoverIcon = "code"
	CoverIconDesign    CoverIcon = "design"
	CoverIconResearch  CoverIcon = "research"
	CoverIconCommunity CoverIcon = "community"
	CoverIconOther     CoverIcon = "other"
)

type Space struct {
	ID               SpaceID
	OwnerID          UserID
	Slug             string
	Title            string
	ShortDescription string
	Description      string
	WhyItMatters     string
	Status           SpaceStatus
	Visibility       SpaceVisibility
	Stage            SpaceStage
	CoverIcon        CoverIcon
	AccentColor      AccentColor
	CreatedAt        time.Time
	UpdatedAt        time.Time
	ArchivedAt       *time.Time
}

type SpaceListFilter struct {
	ViewerID   *UserID
	Status     SpaceStatus
	Visibility SpaceVisibility
	Limit      int
	Offset     int
}

var slugPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{2,49}$`)

func (s Space) Validate() error {
	if !validLength(strings.TrimSpace(s.Title), 2, 80) {
		return ErrInvalidInput
	}
	if !slugPattern.MatchString(s.Slug) {
		return ErrInvalidInput
	}
	if !validLength(strings.TrimSpace(s.ShortDescription), 10, 240) {
		return ErrInvalidInput
	}
	if utf8.RuneCountInString(s.Description) > 3000 {
		return ErrInvalidInput
	}
	if utf8.RuneCountInString(s.WhyItMatters) > 2000 {
		return ErrInvalidInput
	}
	if !s.Status.Valid() || !s.Visibility.Valid() || !s.Stage.Valid() {
		return ErrInvalidInput
	}
	if !s.CoverIcon.Valid() || !s.AccentColor.Valid() {
		return ErrInvalidInput
	}
	return nil
}

func (s SpaceStatus) Valid() bool {
	switch s {
	case SpaceStatusActive, SpaceStatusPaused, SpaceStatusArchived:
		return true
	default:
		return false
	}
}

func (v SpaceVisibility) Valid() bool {
	return v == SpaceVisibilityPublic || v == SpaceVisibilityPrivate
}

func (s SpaceStage) Valid() bool {
	switch s {
	case SpaceStageIdea, SpaceStagePrototype, SpaceStageMVP, SpaceStageBeta, SpaceStageGrowth, SpaceStageScale:
		return true
	default:
		return false
	}
}

func (c AccentColor) Valid() bool {
	switch c {
	case AccentColorMint, AccentColorBlue, AccentColorPurple, AccentColorOrange, AccentColorGray:
		return true
	default:
		return false
	}
}

func (i CoverIcon) Valid() bool {
	switch i {
	case CoverIconLink, CoverIconSprout, CoverIconRocket, CoverIconCode, CoverIconDesign, CoverIconResearch, CoverIconCommunity, CoverIconOther:
		return true
	default:
		return false
	}
}

func validLength(value string, minLength, maxLength int) bool {
	length := utf8.RuneCountInString(value)
	return length >= minLength && length <= maxLength
}
