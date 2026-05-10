package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"backend-hotlines3/internal/feature/auth/policy"
	"backend-hotlines3/internal/feature/user/dto"
	"backend-hotlines3/internal/feature/user/entity"
)

func TestAdminCannotCreateAnyUser(t *testing.T) {
	svc := NewService(&fakeRepo{})
	actor := entity.Actor{ID: 10, Role: policy.RoleAdmin}
	for _, role := range []string{policy.RoleUser, policy.RoleTeamLead, policy.RoleViewer, policy.RoleAdmin, policy.RoleSuperAdmin} {
		t.Run(role, func(t *testing.T) {
			_, err := svc.Create(context.Background(), actor, dto.CreateUserRequest{Username: "123456", Password: "secret1", Role: role})
			if !errors.Is(err, entity.ErrForbidden) {
				t.Fatalf("Create role %s: got %v, want ErrForbidden", role, err)
			}
		})
	}
}

func TestSuperAdminCanCreateAdminButCannotCreateSecondActiveSuperAdmin(t *testing.T) {
	ctx := context.Background()
	repo := &fakeRepo{activeSuperAdmins: 1}
	svc := NewService(repo)
	actor := entity.Actor{ID: 1, Role: policy.RoleSuperAdmin}

	created, err := svc.Create(ctx, actor, dto.CreateUserRequest{Username: "123456", Password: "secret1", Role: policy.RoleAdmin})
	if err != nil {
		t.Fatalf("Create admin: %v", err)
	}
	if created.Role != policy.RoleAdmin {
		t.Fatalf("created role = %q, want admin", created.Role)
	}

	_, err = svc.Create(ctx, actor, dto.CreateUserRequest{Username: "234567", Password: "secret1", Role: policy.RoleSuperAdmin})
	if !errors.Is(err, entity.ErrSuperAdminAlreadyExists) {
		t.Fatalf("Create second super_admin: got %v, want ErrSuperAdminAlreadyExists", err)
	}
}

func TestAdminCannotPromoteUserToAdminOrSuperAdmin(t *testing.T) {
	ctx := context.Background()
	repo := &fakeRepo{users: map[uint]entity.UserInfo{2: {ID: 2, Role: policy.RoleUser, IsActive: true}}}
	svc := NewService(repo)
	actor := entity.Actor{ID: 10, Role: policy.RoleAdmin}
	for _, role := range []string{policy.RoleAdmin, policy.RoleSuperAdmin} {
		t.Run(role, func(t *testing.T) {
			_, err := svc.Update(ctx, actor, 2, dto.UpdateUserRequest{Role: &role})
			if !errors.Is(err, entity.ErrForbidden) {
				t.Fatalf("Update role %s: got %v, want ErrForbidden", role, err)
			}
		})
	}
}

func TestCannotDemoteOrDeactivateOnlyActiveSuperAdmin(t *testing.T) {
	ctx := context.Background()
	repo := &fakeRepo{activeSuperAdmins: 0, users: map[uint]entity.UserInfo{1: {ID: 1, Role: policy.RoleSuperAdmin, IsActive: true}}}
	svc := NewService(repo)
	actor := entity.Actor{ID: 1, Role: policy.RoleSuperAdmin}

	role := policy.RoleAdmin
	if _, err := svc.Update(ctx, actor, 1, dto.UpdateUserRequest{Role: &role}); !errors.Is(err, entity.ErrCannotRemoveOnlySuperAdmin) {
		t.Fatalf("demote only super_admin: got %v, want ErrCannotRemoveOnlySuperAdmin", err)
	}

	inactive := false
	if _, err := svc.Update(ctx, actor, 1, dto.UpdateUserRequest{IsActive: &inactive}); !errors.Is(err, entity.ErrCannotRemoveOnlySuperAdmin) {
		t.Fatalf("deactivate only super_admin: got %v, want ErrCannotRemoveOnlySuperAdmin", err)
	}
}

func TestOnlySuperAdminCanResetPasswordsForOthers(t *testing.T) {
	ctx := context.Background()
	repo := &fakeRepo{users: map[uint]entity.UserInfo{2: {ID: 2, Role: policy.RoleUser, IsActive: true}}}
	svc := NewService(repo)

	err := svc.ResetPassword(ctx, entity.Actor{ID: 10, Role: policy.RoleAdmin}, 2, dto.ResetPasswordRequest{NewPassword: "newsecret"})
	if !errors.Is(err, entity.ErrForbidden) {
		t.Fatalf("admin reset password: got %v, want ErrForbidden", err)
	}

	err = svc.ResetPassword(ctx, entity.Actor{ID: 1, Role: policy.RoleSuperAdmin}, 2, dto.ResetPasswordRequest{NewPassword: "newsecret"})
	if err != nil {
		t.Fatalf("super_admin reset password: %v", err)
	}
	if repo.updatedPasswordID != 2 || repo.updatedPasswordHash == "" || strings.Contains(repo.updatedPasswordHash, "newsecret") {
		t.Fatalf("password reset did not store a hash for target: id=%d hash=%q", repo.updatedPasswordID, repo.updatedPasswordHash)
	}
}

func TestDeleteForbidsAdminDeletingAdminAndOnlySuperAdmin(t *testing.T) {
	ctx := context.Background()
	repo := &fakeRepo{activeSuperAdmins: 0, users: map[uint]entity.UserInfo{
		1: {ID: 1, Role: policy.RoleSuperAdmin, IsActive: true},
		2: {ID: 2, Role: policy.RoleAdmin, IsActive: true},
	}}
	svc := NewService(repo)

	if err := svc.Delete(ctx, entity.Actor{ID: 10, Role: policy.RoleAdmin}, 2); !errors.Is(err, entity.ErrForbidden) {
		t.Fatalf("admin delete admin: got %v, want ErrForbidden", err)
	}
	if err := svc.Delete(ctx, entity.Actor{ID: 1, Role: policy.RoleSuperAdmin}, 1); !errors.Is(err, entity.ErrCannotDeleteSelf) {
		t.Fatalf("delete self: got %v, want ErrCannotDeleteSelf", err)
	}
	if err := svc.Delete(ctx, entity.Actor{ID: 99, Role: policy.RoleSuperAdmin}, 1); !errors.Is(err, entity.ErrCannotRemoveOnlySuperAdmin) {
		t.Fatalf("delete only super_admin: got %v, want ErrCannotRemoveOnlySuperAdmin", err)
	}
}

func TestListContactsPassesSearchAndIncludesAllActiveContacts(t *testing.T) {
	ctx := context.Background()
	repo := &fakeRepo{}
	svc := NewService(repo)

	result, err := svc.ListContacts(ctx, entity.Actor{ID: 7, Role: policy.RoleUser}, entity.ContactListQuery{
		Query:  "somchai",
		TeamID: ptrInt64(2),
		Page:   0,
		Limit:  500,
	})
	if err != nil {
		t.Fatalf("ListContacts: %v", err)
	}
	if repo.capturedContactQuery.Query != "somchai" || repo.capturedContactQuery.TeamID == nil || *repo.capturedContactQuery.TeamID != 2 {
		t.Fatalf("repo query = %+v, want search somchai team 2", repo.capturedContactQuery)
	}
	if repo.capturedContactQuery.IncludeInactive {
		t.Fatalf("non-super-admin should not request inactive contacts")
	}
	if result.Page != 1 || result.Limit != 50 {
		t.Fatalf("pagination = page %d limit %d, want normalized 1/50", result.Page, result.Limit)
	}
}

func TestUpdateOwnContactAllowsUserToEditOwnDisplayPositionPhone(t *testing.T) {
	ctx := context.Background()
	repo := &fakeRepo{users: map[uint]entity.UserInfo{7: {ID: 7, Role: policy.RoleUser, IsActive: true}}}
	svc := NewService(repo)
	displayName := "Somchai"
	position := "Line technician"
	phone := "0812345678"

	updated, err := svc.UpdateContact(ctx, entity.Actor{ID: 7, Role: policy.RoleUser}, 7, dto.UpdateContactRequest{
		DisplayName: &displayName,
		Position:    &position,
		PhoneNumber: &phone,
	})
	if err != nil {
		t.Fatalf("UpdateContact own profile: %v", err)
	}
	if repo.updatedContactID != 7 || repo.updatedContact.DisplayName == nil || *repo.updatedContact.DisplayName != displayName {
		t.Fatalf("contact update not captured: id=%d input=%+v", repo.updatedContactID, repo.updatedContact)
	}
	if updated.DisplayName == nil || *updated.DisplayName != displayName || updated.PhoneNumber == nil || *updated.PhoneNumber != phone {
		t.Fatalf("updated contact = %+v, want display and phone", updated)
	}
}

func TestUpdateContactForbidsUserUpdatingAnotherUser(t *testing.T) {
	ctx := context.Background()
	repo := &fakeRepo{users: map[uint]entity.UserInfo{8: {ID: 8, Role: policy.RoleUser, IsActive: true}}}
	svc := NewService(repo)
	displayName := "Other"

	_, err := svc.UpdateContact(ctx, entity.Actor{ID: 7, Role: policy.RoleUser}, 8, dto.UpdateContactRequest{DisplayName: &displayName})
	if !errors.Is(err, entity.ErrForbidden) {
		t.Fatalf("UpdateContact other user: got %v, want ErrForbidden", err)
	}
	if repo.updatedContactID != 0 {
		t.Fatalf("repo update should not be called, got id %d", repo.updatedContactID)
	}
}

func ptrInt64(v int64) *int64 { return &v }

type fakeRepo struct {
	users                map[uint]entity.UserInfo
	activeSuperAdmins    int64
	created              entity.CreateInput
	updated              entity.UpdateInput
	capturedContactQuery entity.ContactListQuery
	updatedContactID     uint
	updatedContact       entity.UpdateContactInput
	deletedID            uint
	updatedPasswordID    uint
	updatedPasswordHash  string
}

func (r *fakeRepo) List(context.Context, int, int) ([]entity.UserInfo, int64, error) {
	items := make([]entity.UserInfo, 0, len(r.users))
	for _, user := range r.users {
		items = append(items, user)
	}
	return items, int64(len(items)), nil
}

func (r *fakeRepo) GetByID(_ context.Context, id uint) (entity.UserInfo, error) {
	if user, ok := r.users[id]; ok {
		return user, nil
	}
	return entity.UserInfo{}, entity.ErrNotFound
}

func (r *fakeRepo) Create(_ context.Context, in entity.CreateInput) (entity.UserInfo, error) {
	r.created = in
	return entity.UserInfo{ID: 100, Username: in.Username, Role: in.Role, TeamID: in.TeamID, IsActive: in.IsActive}, nil
}

func (r *fakeRepo) Update(_ context.Context, id uint, in entity.UpdateInput) (entity.UserInfo, error) {
	r.updated = in
	user, ok := r.users[id]
	if !ok {
		return entity.UserInfo{}, entity.ErrNotFound
	}
	if in.Username != nil {
		user.Username = *in.Username
	}
	if in.Role != nil {
		user.Role = *in.Role
	}
	if in.TeamID != nil {
		user.TeamID = in.TeamID
	}
	if in.IsActive != nil {
		user.IsActive = *in.IsActive
	}
	return user, nil
}

func (r *fakeRepo) Delete(_ context.Context, id uint) error {
	r.deletedID = id
	return nil
}

func (r *fakeRepo) GetPasswordHash(context.Context, uint) (string, error) {
	return "$2a$12$BL9bVdgl.6mnHeYvuewIEeJcDObTvWNiyzxkSvjKw8aUP3XvR2lRm", nil // bcrypt for oldsecret
}

func (r *fakeRepo) UpdatePassword(_ context.Context, id uint, hashed string) error {
	r.updatedPasswordID = id
	r.updatedPasswordHash = hashed
	return nil
}

func (r *fakeRepo) CountActiveSuperAdmins(context.Context, *uint) (int64, error) {
	return r.activeSuperAdmins, nil
}

func (r *fakeRepo) ListContacts(_ context.Context, q entity.ContactListQuery) ([]entity.ContactInfo, int64, error) {
	r.capturedContactQuery = q
	items := []entity.ContactInfo{}
	for _, user := range r.users {
		items = append(items, entity.ContactInfo{ID: user.ID, Username: user.Username, Role: user.Role, TeamID: user.TeamID, DisplayName: user.DisplayName, Position: user.Position, PhoneNumber: user.PhoneNumber, IsActive: user.IsActive})
	}
	return items, int64(len(items)), nil
}

func (r *fakeRepo) GetContactByID(_ context.Context, id uint) (entity.ContactInfo, error) {
	if user, ok := r.users[id]; ok {
		return entity.ContactInfo{ID: user.ID, Username: user.Username, Role: user.Role, TeamID: user.TeamID, DisplayName: user.DisplayName, Position: user.Position, PhoneNumber: user.PhoneNumber, IsActive: user.IsActive}, nil
	}
	return entity.ContactInfo{}, entity.ErrNotFound
}

func (r *fakeRepo) UpdateContact(_ context.Context, id uint, in entity.UpdateContactInput) (entity.ContactInfo, error) {
	r.updatedContactID = id
	r.updatedContact = in
	user, ok := r.users[id]
	if !ok {
		return entity.ContactInfo{}, entity.ErrNotFound
	}
	if in.DisplayName != nil {
		user.DisplayName = in.DisplayName
	}
	if in.Position != nil {
		user.Position = in.Position
	}
	if in.PhoneNumber != nil {
		user.PhoneNumber = in.PhoneNumber
	}
	return entity.ContactInfo{ID: user.ID, Username: user.Username, Role: user.Role, TeamID: user.TeamID, DisplayName: user.DisplayName, Position: user.Position, PhoneNumber: user.PhoneNumber, IsActive: user.IsActive}, nil
}
