package bootstrap

import (
	"context"
	"errors"
	"strings"
	"testing"

	"backend-hotlines3/internal/feature/auth/policy"
	"backend-hotlines3/internal/feature/user/entity"
)

func TestCreateSuperAdminRefusesWhenActiveSuperAdminExists(t *testing.T) {
	repo := &fakeRepo{activeSuperAdmins: 1}
	err := CreateSuperAdmin(context.Background(), repo, Input{Username: "123456", Password: "secret1"})
	if !errors.Is(err, ErrSuperAdminExists) {
		t.Fatalf("got %v, want ErrSuperAdminExists", err)
	}
	if repo.created.Role != "" {
		t.Fatalf("should not create duplicate super_admin")
	}
}

func TestCreateSuperAdminCreatesActiveSuperAdminWithHashedPassword(t *testing.T) {
	repo := &fakeRepo{}
	err := CreateSuperAdmin(context.Background(), repo, Input{Username: "123456", Password: "secret1"})
	if err != nil {
		t.Fatalf("CreateSuperAdmin: %v", err)
	}
	if repo.created.Role != policy.RoleSuperAdmin || !repo.created.IsActive {
		t.Fatalf("created = %+v, want active super_admin", repo.created)
	}
	if repo.created.HashedPassword == "" || strings.Contains(repo.created.HashedPassword, "secret1") {
		t.Fatalf("password not hashed: %q", repo.created.HashedPassword)
	}
}

type fakeRepo struct {
	activeSuperAdmins int64
	created           entity.CreateInput
}

func (r *fakeRepo) CountActiveSuperAdmins(context.Context, *uint) (int64, error) {
	return r.activeSuperAdmins, nil
}

func (r *fakeRepo) Create(_ context.Context, in entity.CreateInput) (entity.UserInfo, error) {
	r.created = in
	return entity.UserInfo{ID: 1, Username: in.Username, Role: in.Role, IsActive: in.IsActive}, nil
}
