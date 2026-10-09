package domain

import "testing"

func TestSpaceValidation(t *testing.T) {
	valid := Space{
		Slug:             "worknet",
		Title:            "WorkNet",
		ShortDescription: "A public network for project work.",
		Status:           SpaceStatusActive,
		Visibility:       SpaceVisibilityPublic,
		Stage:            SpaceStageMVP,
		CoverIcon:        CoverIconLink,
		AccentColor:      AccentColorMint,
	}
	if err := valid.Validate(); err != nil {
		t.Fatalf("valid space: %v", err)
	}

	invalid := valid
	invalid.Slug = "WorkNet"
	if err := invalid.Validate(); err != ErrInvalidInput {
		t.Fatalf("invalid slug error = %v, want %v", err, ErrInvalidInput)
	}
}

func TestNormalizeStackRejectsDuplicates(t *testing.T) {
	_, err := NormalizeStack([]string{"Go", " go "})
	if err != ErrInvalidInput {
		t.Fatalf("error = %v, want %v", err, ErrInvalidInput)
	}
}

func TestNormalizeLinksValidatesURL(t *testing.T) {
	_, err := NormalizeLinks([]SpaceLink{{
		Type:  LinkTypeGitHub,
		Title: "GitHub",
		URL:   "javascript:alert(1)",
	}})
	if err != ErrInvalidInput {
		t.Fatalf("error = %v, want %v", err, ErrInvalidInput)
	}
}

func TestPermissionsForRole(t *testing.T) {
	owner := MemberRoleOwner
	permissions := PermissionsForRole(&owner)
	if !permissions.CanEditSpace || !permissions.CanArchiveSpace || !permissions.CanManageMembers || !permissions.CanManageLinks {
		t.Fatalf("owner permissions = %+v", permissions)
	}
	if permissions.CanManageRoadmap || permissions.CanCreateRequest || permissions.CanCreateContribution || permissions.CanApply || permissions.CanFollow {
		t.Fatalf("future permissions must be disabled: %+v", permissions)
	}
}
