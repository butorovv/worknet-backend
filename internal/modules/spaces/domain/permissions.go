package domain

type Permissions struct {
	CanEditSpace          bool
	CanArchiveSpace       bool
	CanManageMembers      bool
	CanManageLinks        bool
	CanManageRoadmap      bool
	CanCreateRequest      bool
	CanCreateContribution bool
	CanApply              bool
	CanFollow             bool
}

func PermissionsForRole(role *MemberRole) Permissions {
	if role == nil || !role.CanManageSpace() {
		return Permissions{}
	}
	return Permissions{
		CanEditSpace:     true,
		CanArchiveSpace:  true,
		CanManageMembers: true,
		CanManageLinks:   true,
	}
}
