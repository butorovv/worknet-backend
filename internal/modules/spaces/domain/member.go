package domain

import "time"

type MemberRole string

const (
	MemberRoleOwner       MemberRole = "owner"
	MemberRoleAdmin       MemberRole = "admin"
	MemberRoleMember      MemberRole = "member"
	MemberRoleContributor MemberRole = "contributor"
)

type SpaceMember struct {
	SpaceID   SpaceID
	UserID    UserID
	Username  string
	Role      MemberRole
	Title     *string
	CreatedAt time.Time
}

func (r MemberRole) Valid() bool {
	switch r {
	case MemberRoleOwner, MemberRoleAdmin, MemberRoleMember, MemberRoleContributor:
		return true
	default:
		return false
	}
}

func (r MemberRole) CanManageSpace() bool {
	return r == MemberRoleOwner || r == MemberRoleAdmin
}
