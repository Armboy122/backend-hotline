package entity

import (
	"testing"

	"backend-hotlines3/internal/feature/auth/policy"
)

func TestLargeWorkActorPermissions(t *testing.T) {
	ownerTeamID := int64(10)
	otherTeamID := int64(11)

	tests := []struct {
		name      string
		actor     Actor
		itemTeam  *int64
		canCreate bool
		canUpdate bool
		canCancel bool
		canView   bool
	}{
		{name: "super admin full access", actor: Actor{Role: policy.RoleSuperAdmin}, itemTeam: &otherTeamID, canCreate: true, canUpdate: true, canCancel: true, canView: true},
		{name: "legacy admin cannot manage own team work", actor: Actor{Role: policy.RoleAdmin, TeamID: &ownerTeamID}, itemTeam: &ownerTeamID, canCreate: false, canUpdate: false, canCancel: false, canView: false},
		{name: "legacy admin cannot manage other team work", actor: Actor{Role: policy.RoleAdmin, TeamID: &ownerTeamID}, itemTeam: &otherTeamID, canCreate: false, canUpdate: false, canCancel: false, canView: false},
		{name: "team lead can create and manage own team planned work", actor: Actor{UserID: 101, Role: policy.RoleTeamLead, TeamID: &ownerTeamID}, itemTeam: &ownerTeamID, canCreate: true, canUpdate: true, canCancel: true, canView: true},
		{name: "team lead cannot manage other team", actor: Actor{Role: policy.RoleTeamLead, TeamID: &ownerTeamID}, itemTeam: &otherTeamID, canCreate: false, canUpdate: false, canCancel: false, canView: false},
		{name: "user can only view own team item", actor: Actor{Role: policy.RoleUser, TeamID: &ownerTeamID}, itemTeam: &ownerTeamID, canCreate: false, canUpdate: false, canCancel: false, canView: true},
		{name: "viewer can only view own team item", actor: Actor{Role: policy.RoleViewer, TeamID: &ownerTeamID}, itemTeam: &ownerTeamID, canCreate: false, canUpdate: false, canCancel: false, canView: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := &LargeWorkItem{OwnerTeamID: 99, Status: LargeWorkStatusPlanned}
			if tt.itemTeam != nil {
				item.OwnerTeamID = *tt.itemTeam
			}
			if got := tt.actor.CanCreateLargeWork(tt.itemTeam); got != tt.canCreate {
				t.Fatalf("CanCreateLargeWork() = %v, want %v", got, tt.canCreate)
			}
			if got := tt.actor.CanUpdateLargeWork(item); got != tt.canUpdate {
				t.Fatalf("CanUpdateLargeWork() = %v, want %v", got, tt.canUpdate)
			}
			if got := tt.actor.CanCancelLargeWork(item); got != tt.canCancel {
				t.Fatalf("CanCancelLargeWork() = %v, want %v", got, tt.canCancel)
			}
			if got := tt.actor.CanViewLargeWork(item); got != tt.canView {
				t.Fatalf("CanViewLargeWork() = %v, want %v", got, tt.canView)
			}
		})
	}
}

func TestLargeWorkActorTeamLeadCannotManageCreatedPlanForOtherTeam(t *testing.T) {
	actorTeamID := int64(10)
	otherTeamID := int64(11)
	actor := Actor{UserID: 101, Role: policy.RoleTeamLead, TeamID: &actorTeamID}
	item := &LargeWorkItem{OwnerTeamID: otherTeamID, CreatedByUserID: actor.UserID, Status: LargeWorkStatusPlanned}

	if actor.CanUpdateLargeWork(item) {
		t.Fatalf("team lead creator CanUpdateLargeWork() for another team = true, want false")
	}
	if actor.CanCancelLargeWork(item) {
		t.Fatalf("team lead creator CanCancelLargeWork() for another team = true, want false")
	}
}

func TestLargeWorkOverviewVisibleToAllAuthenticatedTeamRoles(t *testing.T) {
	teamID := int64(10)
	for _, role := range []string{policy.RoleSuperAdmin, policy.RoleTeamLead, policy.RoleUser, policy.RoleViewer} {
		t.Run(role, func(t *testing.T) {
			actor := Actor{Role: role, TeamID: &teamID}
			if !actor.CanViewLargeWorkOverview() {
				t.Fatalf("CanViewLargeWorkOverview() = false, want true for %s", role)
			}
		})
	}
	if (Actor{Role: ""}).CanViewLargeWorkOverview() {
		t.Fatalf("anonymous CanViewLargeWorkOverview() = true, want false")
	}
}

func TestLargeWorkActorTeamLeadCanAssignTasksForOwnPlanOnly(t *testing.T) {
	ownerTeamID := int64(10)
	otherTeamID := int64(11)
	actor := Actor{UserID: 101, Role: policy.RoleTeamLead, TeamID: &ownerTeamID}

	if !actor.CanAssignLargeWorkTasks(&LargeWorkItem{OwnerTeamID: ownerTeamID, Status: LargeWorkStatusPlanned}, otherTeamID) {
		t.Fatalf("team lead CanAssignLargeWorkTasks() for own-team plan = false, want true")
	}
	if actor.CanAssignLargeWorkTasks(&LargeWorkItem{OwnerTeamID: otherTeamID, CreatedByUserID: actor.UserID, Status: LargeWorkStatusPlanned}, otherTeamID) {
		t.Fatalf("team lead CanAssignLargeWorkTasks() for another-team created plan = true, want false")
	}
	if actor.CanAssignLargeWorkTasks(&LargeWorkItem{OwnerTeamID: otherTeamID, Status: LargeWorkStatusPlanned}, otherTeamID) {
		t.Fatalf("team lead CanAssignLargeWorkTasks() for unrelated plan = true, want false")
	}
}

func TestLargeWorkActorManageActionsRequirePrivilegedRoleAndEditableState(t *testing.T) {
	for _, status := range []string{LargeWorkStatusDraft, LargeWorkStatusPlanned} {
		t.Run("team lead can manage own team "+status, func(t *testing.T) {
			item := &LargeWorkItem{OwnerTeamID: 7, Status: status}
			teamID := int64(7)
			actor := Actor{Role: policy.RoleTeamLead, TeamID: &teamID}

			if !actor.CanUpdateLargeWork(item) {
				t.Fatalf("CanUpdateLargeWork(%s) = false, want true", status)
			}
			if !actor.CanCancelLargeWork(item) {
				t.Fatalf("CanCancelLargeWork(%s) = false, want true", status)
			}
		})
	}

	for _, status := range []string{LargeWorkStatusInProgress, LargeWorkStatusCompleted, LargeWorkStatusCancelled} {
		t.Run("team lead cannot manage own team "+status, func(t *testing.T) {
			item := &LargeWorkItem{OwnerTeamID: 7, Status: status}
			teamID := int64(7)
			actor := Actor{Role: policy.RoleTeamLead, TeamID: &teamID}

			if actor.CanUpdateLargeWork(item) {
				t.Fatalf("CanUpdateLargeWork(%s) = true, want false", status)
			}
			if actor.CanCancelLargeWork(item) {
				t.Fatalf("CanCancelLargeWork(%s) = true, want false", status)
			}
		})
	}
}

func TestLargeWorkActorScopesRequestedTeamToOwnTeamForNonPrivilegedActors(t *testing.T) {
	ownerTeamID := int64(10)
	requested := int64(99)
	actor := Actor{Role: policy.RoleTeamLead, TeamID: &ownerTeamID}

	if got := actor.ScopedTeamID(&requested); got == nil || *got != ownerTeamID {
		t.Fatalf("ScopedTeamID() = %#v, want own team %d", got, ownerTeamID)
	}
	if got := (Actor{Role: policy.RoleSuperAdmin}).ScopedTeamID(&requested); got == nil || *got != requested {
		t.Fatalf("privileged ScopedTeamID() = %#v, want requested team %d", got, requested)
	}
}
