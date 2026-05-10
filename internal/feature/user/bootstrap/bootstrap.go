package bootstrap

import (
	"context"
	"errors"

	"backend-hotlines3/internal/feature/auth/policy"
	"backend-hotlines3/internal/feature/user/entity"
	"backend-hotlines3/pkg/password"
)

var ErrSuperAdminExists = errors.New("active super_admin already exists")

type Input struct {
	Username string
	Password string
}

type Repository interface {
	CountActiveSuperAdmins(ctx context.Context, excludeID *uint) (int64, error)
	Create(ctx context.Context, in entity.CreateInput) (entity.UserInfo, error)
}

func CreateSuperAdmin(ctx context.Context, repo Repository, input Input) error {
	count, err := repo.CountActiveSuperAdmins(ctx, nil)
	if err != nil {
		return err
	}
	if count > 0 {
		return ErrSuperAdminExists
	}
	hashed, err := password.HashPassword(input.Password)
	if err != nil {
		return err
	}
	_, err = repo.Create(ctx, entity.CreateInput{
		Username:       input.Username,
		HashedPassword: hashed,
		Role:           policy.RoleSuperAdmin,
		IsActive:       true,
	})
	return err
}
