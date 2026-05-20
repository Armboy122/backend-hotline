package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"backend-hotlines3/internal/feature/auth/entity"
	"backend-hotlines3/pkg/jwt"
	"backend-hotlines3/pkg/password"
)

type fakeAuthRepo struct {
	findByUsernameErr error
	user              *entity.AuthUser
	me                entity.UserInfo
	created           entity.CreateUserInput
}

func (r fakeAuthRepo) FindByUsername(ctx context.Context, username string) (*entity.AuthUser, error) {
	if r.findByUsernameErr != nil {
		return nil, r.findByUsernameErr
	}
	if r.user != nil {
		return r.user, nil
	}
	return nil, entity.ErrInvalidCredentials
}

func (r fakeAuthRepo) FindByID(ctx context.Context, id uint) (*entity.AuthUser, error) {
	return nil, entity.ErrUserNotFound
}

func (r fakeAuthRepo) FindByIDWithTeam(ctx context.Context, id uint) (entity.UserInfo, error) {
	if r.me.ID != 0 {
		return r.me, nil
	}
	return entity.UserInfo{}, entity.ErrUserNotFound
}

func (r fakeAuthRepo) Create(ctx context.Context, in entity.CreateUserInput) (entity.UserInfo, error) {
	return entity.UserInfo{ID: 10, Username: in.Username, Role: in.Role, TeamID: in.TeamID, IsActive: in.IsActive, MustChangePassword: in.MustChangePassword}, nil
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

func TestLoginIncludesMustChangePasswordFlag(t *testing.T) {
	hash, err := password.HashPassword("secret")
	if err != nil {
		t.Fatalf("hash fixture password: %v", err)
	}
	svc := NewService(fakeAuthRepo{user: &entity.AuthUser{
		ID:                 9,
		Username:           "900001",
		Role:               "user",
		IsActive:           true,
		MustChangePassword: true,
		PasswordHash:       hash,
	}}, jwt.NewJWTManager("secret", time.Hour, 24*time.Hour))

	result, err := svc.Login(context.Background(), "900001", "secret")
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if !result.User.MustChangePassword {
		t.Fatalf("login user MustChangePassword = false, want true")
	}
}

func TestMeIncludesMustChangePasswordFlag(t *testing.T) {
	svc := NewService(fakeAuthRepo{me: entity.UserInfo{ID: 9, Username: "900001", Role: "user", IsActive: true, MustChangePassword: true}}, nil)

	info, err := svc.Me(context.Background(), 9)
	if err != nil {
		t.Fatalf("Me: %v", err)
	}
	if !info.MustChangePassword {
		t.Fatalf("me MustChangePassword = false, want true")
	}
}
