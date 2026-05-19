package entity

import (
	"testing"
	"time"

	"backend-hotlines3/internal/feature/auth/policy"
)

func TestActorPolicyForTeamPlanActions(t *testing.T) {
	ownTeamID := int64(7)
	otherTeamID := int64(8)
	creatorID := int64(91)

	tests := []struct {
		name       string
		actor      Actor
		teamID     int64
		plan       TeamPlan
		wantCreate bool
		wantDelete bool
	}{
		{
			name:       "super admin can manage any team",
			actor:      Actor{UserID: 1, Role: policy.RoleSuperAdmin},
			teamID:     otherTeamID,
			plan:       TeamPlan{TeamID: otherTeamID, CreatedByUserID: creatorID, Status: StatusPlanned},
			wantCreate: true,
			wantDelete: true,
		},
		{
			name:       "team lead can create own team and delete own team",
			actor:      Actor{UserID: 2, Role: policy.RoleTeamLead, TeamID: &ownTeamID},
			teamID:     ownTeamID,
			plan:       TeamPlan{TeamID: ownTeamID, CreatedByUserID: creatorID, Status: StatusPlanned},
			wantCreate: true,
			wantDelete: true,
		},
		{
			name:       "user can create and delete own team like team lead",
			actor:      Actor{UserID: 3, Role: policy.RoleUser, TeamID: &ownTeamID},
			teamID:     ownTeamID,
			plan:       TeamPlan{TeamID: ownTeamID, CreatedByUserID: 3, Status: StatusPlanned},
			wantCreate: true,
			wantDelete: true,
		},
		{
			name:       "legacy admin cannot create or delete team plan",
			actor:      Actor{UserID: 4, Role: policy.RoleAdmin, TeamID: &ownTeamID},
			teamID:     ownTeamID,
			plan:       TeamPlan{TeamID: ownTeamID, CreatedByUserID: creatorID, Status: StatusPlanned},
			wantCreate: false,
			wantDelete: false,
		},
		{
			name:       "viewer cannot create or delete",
			actor:      Actor{UserID: 5, Role: policy.RoleViewer, TeamID: &ownTeamID},
			teamID:     ownTeamID,
			plan:       TeamPlan{TeamID: ownTeamID, CreatedByUserID: creatorID, Status: StatusPlanned},
			wantCreate: false,
			wantDelete: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.actor.CanCreateTeamPlan(tt.teamID); got != tt.wantCreate {
				t.Fatalf("CanCreateTeamPlan() = %v, want %v", got, tt.wantCreate)
			}
			if got := tt.actor.CanDeleteTeamPlan(tt.plan); got != tt.wantDelete {
				t.Fatalf("CanDeleteTeamPlan() = %v, want %v", got, tt.wantDelete)
			}
		})
	}
}

func TestActorPolicyForTeamPlanUpdates(t *testing.T) {
	teamID := int64(7)
	otherTeamID := int64(8)
	creatorID := int64(91)
	otherCreatorID := int64(92)

	teamLead := Actor{UserID: 2, Role: policy.RoleTeamLead, TeamID: &teamID}
	user := Actor{UserID: creatorID, Role: policy.RoleUser, TeamID: &teamID}
	superAdmin := Actor{UserID: 1, Role: policy.RoleSuperAdmin}

	editable := TeamPlan{ID: 10, TeamID: teamID, CreatedByUserID: creatorID, Status: StatusPlanned}
	completed := TeamPlan{ID: 11, TeamID: teamID, CreatedByUserID: creatorID, Status: StatusCompleted}
	otherTeam := TeamPlan{ID: 12, TeamID: otherTeamID, CreatedByUserID: otherCreatorID, Status: StatusPlanned}

	if !superAdmin.CanUpdateTeamPlan(editable) {
		t.Fatalf("super admin should update any item")
	}
	if !teamLead.CanUpdateTeamPlan(editable) {
		t.Fatalf("team lead should update own team item")
	}
	if user.CanUpdateTeamPlan(otherTeam) {
		t.Fatalf("user should not update other team's item")
	}
	if !user.CanUpdateTeamPlan(editable) {
		t.Fatalf("user should update own-team item")
	}
	if !user.CanUpdateTeamPlan(completed) {
		t.Fatalf("user should update own-team completed item by team scope")
	}
}

func TestTeamPlanExpandedDates(t *testing.T) {
	start := time.Date(2026, 6, 3, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 6, 5, 0, 0, 0, 0, time.UTC)
	plan := TeamPlan{StartDate: &start, EndDate: &end}

	dates := plan.ExpandedDates()
	if len(dates) != 3 {
		t.Fatalf("ExpandedDates() len = %d, want 3", len(dates))
	}
	if !dates[0].Equal(start) || !dates[1].Equal(start.Add(24*time.Hour)) || !dates[2].Equal(end) {
		t.Fatalf("ExpandedDates() = %v", dates)
	}
}
