package monthlyplan

// Actor represents the authenticated user making a request.
type Actor struct {
	UserID int64
	Role   string // "admin", "manager", "staff", etc.
	TeamID *int64 // nil if user has no team assignment
}

// IsAdmin returns true if the actor has admin role.
func (a Actor) IsAdmin() bool {
	return a.Role == "admin"
}

// CanUploadAfterLock checks if actor can bypass the lock.
// Admins always can. Non-admins cannot.
func (a Actor) CanUploadAfterLock(adminCanUpload bool) bool {
	return a.IsAdmin() && adminCanUpload
}

// CanUploadMasterPlan only admins can upload master plans.
func (a Actor) CanUploadMasterPlan() bool {
	return a.IsAdmin()
}

// CanUploadForTeam checks if actor can upload on behalf of a specific team.
// Admin can upload for any team; non-admin must match their own team.
func (a Actor) CanUploadForTeam(teamID *int64) bool {
	if a.IsAdmin() {
		return true
	}
	if teamID == nil || a.TeamID == nil {
		return false
	}
	return *a.TeamID == *teamID
}

// CanAccessFile checks if actor can access a non-master-plan file.
// Admin can access any; others can only access files from their team.
func (a Actor) CanAccessFile(fileTeamID *int64) bool {
	if a.IsAdmin() {
		return true
	}
	if a.TeamID == nil || fileTeamID == nil {
		return false
	}
	return *a.TeamID == *fileTeamID
}
