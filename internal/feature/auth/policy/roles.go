package policy

const (
	RoleSuperAdmin = "super_admin"
	RoleAdmin      = "admin"
	RoleTeamLead   = "team_lead"
	RoleUser       = "user"
	RoleViewer     = "viewer"
)

func IsValidRole(role string) bool {
	switch role {
	case RoleSuperAdmin, RoleAdmin, RoleTeamLead, RoleUser, RoleViewer:
		return true
	default:
		return false
	}
}

func IsPrivilegedRole(role string) bool {
	return role == RoleSuperAdmin
}

func IsMonthlyPlanManagerRole(role string) bool {
	return role == RoleSuperAdmin || role == RoleAdmin
}

func CanManageRole(actorRole, targetRole string) bool {
	return actorRole == RoleSuperAdmin
}

func CanCreateRole(actorRole, targetRole string) bool {
	return actorRole == RoleSuperAdmin
}

func CanResetPasswords(actorRole string) bool {
	return actorRole == RoleSuperAdmin
}
