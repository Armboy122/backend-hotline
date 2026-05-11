package entity

import "backend-hotlines3/internal/feature/auth/policy"

func (a Actor) IsPrivileged() bool {
	return a.Role == policy.RoleSuperAdmin || a.Role == policy.RoleAdmin
}

func (a Actor) IsTeamLead() bool {
	return a.Role == policy.RoleTeamLead
}

func (a Actor) CanCreateLargeWork(ownerTeamID *int64) bool {
	if a.IsPrivileged() {
		return true
	}
	return a.IsTeamLead() && a.TeamID != nil && ownerTeamID != nil && *a.TeamID == *ownerTeamID
}

func (a Actor) CanUpdateLargeWork(item *LargeWorkItem) bool {
	if item == nil {
		return false
	}
	return IsLargeWorkEditableStatus(item.Status) && a.canManageEditableLargeWork(item)
}

func (a Actor) CanCancelLargeWork(item *LargeWorkItem) bool {
	if item == nil {
		return false
	}
	return IsLargeWorkEditableStatus(item.Status) && a.canManageEditableLargeWork(item)
}

func (a Actor) CanAssignLargeWorkTasks(item *LargeWorkItem, _ int64) bool {
	if item == nil {
		return false
	}
	return IsLargeWorkEditableStatus(item.Status) && a.canManageEditableLargeWork(item)
}

func (a Actor) canManageEditableLargeWork(item *LargeWorkItem) bool {
	if a.IsPrivileged() {
		return true
	}
	return a.IsTeamLead() && a.TeamID != nil && *a.TeamID == item.OwnerTeamID
}

func IsLargeWorkEditableStatus(status string) bool {
	return status == "" || status == LargeWorkStatusDraft || status == LargeWorkStatusPlanned
}

func (a Actor) CanViewLargeWorkOverview() bool {
	switch a.Role {
	case policy.RoleSuperAdmin, policy.RoleAdmin, policy.RoleTeamLead, policy.RoleUser, policy.RoleViewer:
		return true
	default:
		return false
	}
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
