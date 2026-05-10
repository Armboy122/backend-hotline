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
	if a.TeamID == nil {
		return false
	}
	return *a.TeamID == teamID
}

func (a Actor) CanWriteTeam(teamID int64) bool {
	if a.IsPrivileged() {
		return true
	}
	if a.Role == policy.RoleViewer || a.TeamID == nil {
		return false
	}
	return *a.TeamID == teamID
}

func (a Actor) ScopedTeamID(requested *int64) *int64 {
	if a.IsPrivileged() {
		return requested
	}
	if a.TeamID != nil {
		return a.TeamID
	}
	noTeam := int64(-1)
	return &noTeam
}
