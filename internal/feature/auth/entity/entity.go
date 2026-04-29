package entity

import "errors"

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrAccountDisabled    = errors.New("account disabled")
	ErrUserExists         = errors.New("username already taken")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidToken       = errors.New("invalid or expired token")
	ErrTeamRequired       = errors.New("teamId required for role user")
)

type Actor struct {
	UserID uint
	Role   string
}

type UserInfo struct {
	ID        uint
	Username  string
	Role      string
	TeamID    *int64
	IsActive  bool
	LastLogin *string
	CreatedAt string
}

// AuthUser carries credential fields needed during login/refresh in addition to UserInfo fields.
type AuthUser struct {
	ID           uint
	Username     string
	Role         string
	TeamID       *int64
	IsActive     bool
	LastLogin    *string
	CreatedAt    string
	PasswordHash string
}

func (a AuthUser) ToUserInfo() UserInfo {
	return UserInfo{
		ID:        a.ID,
		Username:  a.Username,
		Role:      a.Role,
		TeamID:    a.TeamID,
		IsActive:  a.IsActive,
		LastLogin: a.LastLogin,
		CreatedAt: a.CreatedAt,
	}
}

// CreateUserInput is passed from service to repository — Password field carries the hash.
type CreateUserInput struct {
	Username       string
	HashedPassword string
	Role           string
	TeamID         *int64
	IsActive       bool
}

type LoginResult struct {
	AccessToken  string
	RefreshToken string
	User         UserInfo
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

type RegisterInput struct {
	Username string
	Password string
	Role     string
	TeamID   *int64
	IsActive *bool
}
