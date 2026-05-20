package service

import (
	"context"
	"errors"
	"testing"

	"backend-hotlines3/internal/feature/auth/policy"
	"backend-hotlines3/internal/feature/user/entity"
)

func TestExternalContactsSupportRoleScopedCRUD(t *testing.T) {
	ctx := context.Background()
	teamID := int64(7)
	repo := &fakeRepo{}
	svc := NewService(repo)

	created, err := svc.CreateExternalContact(ctx, entity.Actor{ID: 2, Role: policy.RoleTeamLead}, entity.ExternalContactInput{
		Type:         entity.ContactTypeEmergency,
		DisplayName:  "ศูนย์แก้ไฟฉุกเฉิน",
		PhoneNumber:  "1129",
		TeamID:       &teamID,
		Organization: strPtr("PEA"),
	})
	if err != nil {
		t.Fatalf("team lead creates scoped emergency contact: %v", err)
	}
	if repo.createdExternalContact.Type != entity.ContactTypeEmergency || repo.createdExternalContact.TeamID == nil || *repo.createdExternalContact.TeamID != teamID {
		t.Fatalf("created external contact = %+v", repo.createdExternalContact)
	}
	if created.Source != entity.ContactSourceExternal || created.Type != entity.ContactTypeEmergency {
		t.Fatalf("created contact source/type = %s/%s", created.Source, created.Type)
	}

	_, err = svc.CreateExternalContact(ctx, entity.Actor{ID: 3, Role: policy.RoleViewer}, entity.ExternalContactInput{Type: entity.ContactTypeExternal, DisplayName: "หจก. คู่สัญญา", PhoneNumber: "0811111111", TeamID: &teamID})
	if !errors.Is(err, entity.ErrForbidden) {
		t.Fatalf("viewer create external contact = %v, want forbidden", err)
	}

	list, err := svc.ListContacts(ctx, entity.Actor{ID: 4, Role: policy.RoleUser}, entity.ContactListQuery{TeamID: &teamID, Page: 1, Limit: 20, Type: entity.ContactTypeEmergency})
	if err != nil {
		t.Fatalf("list contacts: %v", err)
	}
	if repo.capturedContactQuery.Type != entity.ContactTypeEmergency || list.Items[0].Source != entity.ContactSourceExternal {
		t.Fatalf("unified contact list/query = %+v item=%+v", repo.capturedContactQuery, list.Items)
	}
}

func strPtr(v string) *string { return &v }
