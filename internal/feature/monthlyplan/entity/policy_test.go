package entity

import (
	"testing"

	"backend-hotlines3/internal/feature/auth/policy"
)

func TestMonthlyPlanUploadPolicyAllowsManagersToUploadMasterPlansOnly(t *testing.T) {
	teamID := int64(7)

	tests := []struct {
		name  string
		actor Actor
		want  bool
	}{
		{name: "super admin can upload master plan", actor: Actor{Role: policy.RoleSuperAdmin}, want: true},
		{name: "team lead with upload capability can upload master plan", actor: Actor{Role: policy.RoleTeamLead, TeamID: &teamID, Capabilities: []string{"can_upload_approved_monthly_plan"}}, want: true},
		{name: "legacy admin cannot upload master plan", actor: Actor{Role: policy.RoleAdmin}, want: false},
		{name: "team lead without capability cannot upload master plan", actor: Actor{Role: policy.RoleTeamLead, TeamID: &teamID}, want: false},
		{name: "user cannot upload master plan", actor: Actor{Role: policy.RoleUser, TeamID: &teamID}, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.actor.CanUploadMasterPlan(); got != tt.want {
				t.Fatalf("CanUploadMasterPlan() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMonthlyPlanUploadPolicyAllowsTeamSubmittersOnlyForOwnTeam(t *testing.T) {
	teamID := int64(7)
	otherTeamID := int64(8)

	tests := []struct {
		name   string
		actor  Actor
		teamID *int64
		want   bool
	}{
		{name: "team lead can upload own team plan", actor: Actor{Role: policy.RoleTeamLead, TeamID: &teamID}, teamID: &teamID, want: true},
		{name: "user can upload own team plan", actor: Actor{Role: policy.RoleUser, TeamID: &teamID}, teamID: &teamID, want: true},
		{name: "team lead cannot upload another team plan", actor: Actor{Role: policy.RoleTeamLead, TeamID: &teamID}, teamID: &otherTeamID, want: false},
		{name: "user cannot upload another team plan", actor: Actor{Role: policy.RoleUser, TeamID: &teamID}, teamID: &otherTeamID, want: false},
		{name: "viewer cannot upload own team plan", actor: Actor{Role: policy.RoleViewer, TeamID: &teamID}, teamID: &teamID, want: false},
		{name: "user without team cannot upload team plan", actor: Actor{Role: policy.RoleUser}, teamID: nil, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.actor.CanUploadForTeam(tt.teamID); got != tt.want {
				t.Fatalf("CanUploadForTeam() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestViewerCannotDownloadMonthlyPlanFiles(t *testing.T) {
	teamID := int64(7)

	tests := []struct {
		name  string
		actor Actor
		want  bool
	}{
		{name: "super admin can download", actor: Actor{Role: policy.RoleSuperAdmin}, want: true},
		{name: "legacy admin cannot download", actor: Actor{Role: policy.RoleAdmin}, want: false},
		{name: "team lead can download", actor: Actor{Role: policy.RoleTeamLead, TeamID: &teamID}, want: true},
		{name: "user can download", actor: Actor{Role: policy.RoleUser, TeamID: &teamID}, want: true},
		{name: "viewer cannot download", actor: Actor{Role: policy.RoleViewer, TeamID: &teamID}, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.actor.CanDownloadFile(); got != tt.want {
				t.Fatalf("CanDownloadFile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSuperAdminCanBypassMonthlyPlanLockEvenWhenAdminOverrideDisabled(t *testing.T) {
	if !((Actor{Role: policy.RoleSuperAdmin}).CanUploadAfterLock(false)) {
		t.Fatalf("expected super_admin to bypass lock regardless of admin override setting")
	}
	if (Actor{Role: policy.RoleTeamLead, Capabilities: []string{"can_upload_approved_monthly_plan"}}).CanUploadAfterLock(false) {
		t.Fatalf("expected capability holder to respect disabled override setting")
	}
	if !((Actor{Role: policy.RoleTeamLead, Capabilities: []string{"can_upload_approved_monthly_plan"}}).CanUploadAfterLock(true)) {
		t.Fatalf("expected capability holder to bypass lock when override setting is enabled")
	}
	if (Actor{Role: policy.RoleAdmin, Capabilities: []string{"can_upload_approved_monthly_plan"}}).CanUploadAfterLock(true) {
		t.Fatalf("expected legacy admin to be denied by role validation even with stale capability")
	}
}
