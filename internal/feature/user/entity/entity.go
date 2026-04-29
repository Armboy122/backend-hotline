package entity

import "errors"

var (
	ErrNotFound         = errors.New("user not found")
	ErrInvalidID        = errors.New("invalid user id")
	ErrCannotDeleteSelf = errors.New("cannot delete own account")
	ErrInvalidPassword  = errors.New("old password is incorrect")
)

type UserInfo struct {
	ID        uint
	Username  string
	Role      string
	TeamID    *int64
	IsActive  bool
	LastLogin *string
	CreatedAt string
}

type ListResult struct {
	Items []UserInfo
	Total int64
	Page  int
	Limit int
}

// CreateInput carries fields for creating a user — HashedPassword holds the bcrypt hash.
type CreateInput struct {
	Username       string
	HashedPassword string
	Role           string
	TeamID         *int64
	IsActive       bool
}

// UpdateInput carries optional fields for partial update.
type UpdateInput struct {
	Username *string
	Role     *string
	TeamID   *int64
	IsActive *bool
}
