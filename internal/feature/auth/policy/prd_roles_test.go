package policy

import "testing"

func TestPRDFinalRoleSetExcludesLegacyAdmin(t *testing.T) {
	finalRoles := []string{RoleSuperAdmin, RoleTeamLead, RoleUser, RoleViewer}
	for _, role := range finalRoles {
		t.Run(role, func(t *testing.T) {
			if !IsValidRole(role) {
				t.Fatalf("IsValidRole(%q) = false, want true for final PRD role", role)
			}
		})
	}

	if IsValidRole(RoleAdmin) {
		t.Fatalf("legacy admin role must not be valid in final PRD role model")
	}
	if IsPrivilegedRole(RoleAdmin) || IsMonthlyPlanManagerRole(RoleAdmin) {
		t.Fatalf("legacy admin must not receive privileged or monthly-plan-manager behavior")
	}
	if CanManageRole(RoleAdmin, RoleUser) || CanCreateRole(RoleAdmin, RoleUser) || CanResetPasswords(RoleAdmin) {
		t.Fatalf("legacy admin must not manage roles/users/passwords")
	}
}

func TestPRDOnlySuperAdminCanManageSystemRoles(t *testing.T) {
	for _, actorRole := range []string{RoleTeamLead, RoleUser, RoleViewer, RoleAdmin, ""} {
		t.Run(actorRole, func(t *testing.T) {
			if CanManageRole(actorRole, RoleUser) {
				t.Fatalf("CanManageRole(%q, user) = true, want false", actorRole)
			}
			if CanCreateRole(actorRole, RoleViewer) {
				t.Fatalf("CanCreateRole(%q, viewer) = true, want false", actorRole)
			}
			if CanResetPasswords(actorRole) {
				t.Fatalf("CanResetPasswords(%q) = true, want false", actorRole)
			}
		})
	}

	for _, targetRole := range []string{RoleSuperAdmin, RoleTeamLead, RoleUser, RoleViewer} {
		if !CanManageRole(RoleSuperAdmin, targetRole) {
			t.Fatalf("super_admin should manage target role %q", targetRole)
		}
		if !CanCreateRole(RoleSuperAdmin, targetRole) {
			t.Fatalf("super_admin should create target role %q", targetRole)
		}
	}
}
