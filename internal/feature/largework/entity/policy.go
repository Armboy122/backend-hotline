package entity

import "backend-hotlines3/internal/feature/auth/policy"

func (a Actor) IsPrivileged() bool {
	return a.Role == policy.RoleSuperAdmin || a.Role == policy.RoleAdmin
}

func (a Actor) CanCreateLargeWork(_ *int64) bool {
	return a.IsPrivileged()
}

func (a Actor) CanUpdateLargeWork(item *LargeWorkItem) bool {
	if item == nil || !a.IsPrivileged() {
		return false
	}
	return IsLargeWorkEditableStatus(item.Status)
}

func (a Actor) CanCancelLargeWork(item *LargeWorkItem) bool {
	if item == nil || !a.IsPrivileged() {
		return false
	}
	return IsLargeWorkEditableStatus(item.Status)
}

func IsLargeWorkEditableStatus(status string) bool {
	return status == "" || status == LargeWorkStatusDraft || status == LargeWorkStatusPlanned
}

func (a Actor) CanViewLargeWork(item *LargeWorkItem) bool {
	if item == nil {
		return false
	}
	if a.IsPrivileged() {
		return true
	}
	if a.TeamID == nil {
		return false
	}
	if *a.TeamID == item.OwnerTeamID {
		return true
	}
	for _, team := range item.Teams {
		if team.ID == *a.TeamID {
			return true
		}
	}
	return false
}

func (a Actor) ScopedTeamID(requested *int64) *int64 {
	if a.IsPrivileged() {
		return requested
	}
	if a.TeamID != nil {
		return a.TeamID
	}
	return nil
}
