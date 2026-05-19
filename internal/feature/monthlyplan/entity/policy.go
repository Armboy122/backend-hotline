package entity

import "backend-hotlines3/internal/feature/auth/policy"

// Actor represents the authenticated user making a request.
type Actor struct {
	UserID int64
	Role   string // "admin", "manager", "staff", etc.
	TeamID *int64 // nil if user has no team assignment
}

// IsAdmin returns true if the actor has monthly-plan manager capability.
func (a Actor) IsAdmin() bool {
	return policy.IsMonthlyPlanManagerRole(a.Role)
}

func (a Actor) IsTeamSubmitter() bool {
	return a.Role == policy.RoleAdmin || a.Role == policy.RoleTeamLead || a.Role == policy.RoleUser
}

// CanUploadAfterLock checks if actor can bypass the lock.
// Super admin always can; admin follows the DB adminCanUploadAfterLock setting.
func (a Actor) CanUploadAfterLock(adminCanUpload bool) bool {
	if a.Role == policy.RoleSuperAdmin {
		return true
	}
	return a.Role == policy.RoleAdmin && adminCanUpload
}

// CanUploadMasterPlan only admins can upload master plans.
func (a Actor) CanUploadMasterPlan() bool {
	return a.IsAdmin()
}

// CanUploadForTeam checks if actor can upload on behalf of a specific team.
// Super admin can upload for any team; others are scoped to their own team.
func (a Actor) CanUploadForTeam(teamID *int64) bool {
	if a.Role == policy.RoleSuperAdmin {
		return true
	}
	if !a.IsTeamSubmitter() || teamID == nil || a.TeamID == nil {
		return false
	}
	return *a.TeamID == *teamID
}

// CanDownloadFile checks if actor can download a monthly plan file.
// Viewer role is always blocked from downloading per Requirement A.
func (a Actor) CanDownloadFile() bool {
	return a.Role != policy.RoleViewer
}

// CanAccessFile checks if actor can access a non-master-plan file.
// Super admin can access any file; others can only access files from their team.
func (a Actor) CanAccessFile(fileTeamID *int64) bool {
	if a.Role == policy.RoleSuperAdmin {
		return true
	}
	if a.TeamID == nil || fileTeamID == nil {
		return false
	}
	return *a.TeamID == *fileTeamID
}
