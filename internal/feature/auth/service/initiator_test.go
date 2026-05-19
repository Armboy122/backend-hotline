package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"backend-hotlines3/internal/feature/auth/entity"
)

type fakeAuthRepo struct {
	findByUsernameErr error
}

func (r fakeAuthRepo) FindByUsername(ctx context.Context, username string) (*entity.AuthUser, error) {
	return nil, r.findByUsernameErr
}

func (r fakeAuthRepo) FindByID(ctx context.Context, id uint) (*entity.AuthUser, error) {
	return nil, entity.ErrUserNotFound
}

func (r fakeAuthRepo) FindByIDWithTeam(ctx context.Context, id uint) (entity.UserInfo, error) {
	return entity.UserInfo{}, entity.ErrUserNotFound
}

func (r fakeAuthRepo) Create(ctx context.Context, in entity.CreateUserInput) (entity.UserInfo, error) {
	return entity.UserInfo{}, nil
}

func (r fakeAuthRepo) UpdateLastLogin(ctx context.Context, id uint, t time.Time) error {
	return nil
}

func TestLoginPreservesRepositoryErrors(t *testing.T) {
	repoErr := errors.New("schema unavailable")
	svc := NewService(fakeAuthRepo{findByUsernameErr: repoErr}, nil)

	_, err := svc.Login(context.Background(), "900001", "secret")

	if !errors.Is(err, repoErr) {
		t.Fatalf("Login error = %v, want repository error", err)
	}
}

func TestLoginKeepsInvalidCredentialsOpaque(t *testing.T) {
	svc := NewService(fakeAuthRepo{findByUsernameErr: entity.ErrInvalidCredentials}, nil)

	_, err := svc.Login(context.Background(), "900001", "secret")

	if !errors.Is(err, entity.ErrInvalidCredentials) {
		t.Fatalf("Login error = %v, want ErrInvalidCredentials", err)
	}
}
