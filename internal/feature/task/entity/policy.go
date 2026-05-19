package entity

import "backend-hotlines3/internal/feature/auth/policy"

// Actor represents the authenticated user making a task request.
type Actor struct {
	UserID int64
	Role   string
	TeamID *int64
}

func (a Actor) IsPrivileged() bool {
	return policy.IsPrivilegedRole(a.Role)
}

func (a Actor) CanReadTeam(teamID int64) bool {
	if a.IsPrivileged() {
		return true
	}
	switch a.Role {
	case policy.RoleTeamLead, policy.RoleUser, policy.RoleViewer:
	default:
		return false
	}
	if a.TeamID == nil {
		return false
	}
	return *a.TeamID == teamID
}

func (a Actor) CanWriteTeam(teamID int64) bool {
	if a.IsPrivileged() {
		return true
	}
	switch a.Role {
	case policy.RoleTeamLead, policy.RoleUser:
	default:
		return false
	}
	if a.TeamID == nil {
		return false
	}
	return *a.TeamID == teamID
}

func (a Actor) ScopedTeamID(requested *int64) *int64 {
	if a.IsPrivileged() {
		return requested
	}
	switch a.Role {
	case policy.RoleTeamLead, policy.RoleUser, policy.RoleViewer:
	default:
		noTeam := int64(-1)
		return &noTeam
	}
	if a.TeamID != nil {
		return a.TeamID
	}
	noTeam := int64(-1)
	return &noTeam
}
