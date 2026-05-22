package entity

import "errors"

var (
	ErrNotFound                   = errors.New("user not found")
	ErrInvalidID                  = errors.New("invalid user id")
	ErrCannotDeleteSelf           = errors.New("cannot delete own account")
	ErrInvalidPassword            = errors.New("old password is incorrect")
	ErrForbidden                  = errors.New("forbidden")
	ErrInvalidRole                = errors.New("invalid role")
	ErrSuperAdminAlreadyExists    = errors.New("active super admin already exists")
	ErrCannotRemoveOnlySuperAdmin = errors.New("cannot remove only active super admin")
	ErrNoContactFields            = errors.New("no contact fields provided")
	ErrInvalidContactField        = errors.New("invalid contact field")
)

const (
	ContactSourceUser     = "user"
	ContactSourceExternal = "external_contact"

	ContactTypeUser            = "user"
	ContactTypeExternal        = "external"
	ContactTypeEmergency       = "emergency"
	ContactTypeOperationCenter = "operation_center"
)

type Actor struct {
	ID   uint
	Role string
}

type TeamInfo struct {
	ID   int64
	Name string
}

type UserInfo struct {
	ID                 uint
	Username           string
	Role               string
	TeamID             *int64
	Team               *TeamInfo
	DisplayName        *string
	Position           *string
	PhoneNumber        *string
	IsActive           bool
	MustChangePassword bool
	LastLogin          *string
	CreatedAt          string
}

type ListResult struct {
	Items []UserInfo
	Total int64
	Page  int
	Limit int
}

// CreateInput carries fields for creating a user — HashedPassword holds the bcrypt hash.
type CreateInput struct {
	Username           string
	HashedPassword     string
	Role               string
	TeamID             *int64
	DisplayName        *string
	IsActive           bool
	MustChangePassword bool
}

// UpdateInput carries optional fields for partial update.
type UpdateInput struct {
	Username *string
	Role     *string
	TeamID   *int64
	IsActive *bool
}

type ContactInfo struct {
	ID           uint
	ExternalID   int64
	Source       string
	Type         string
	Username     string
	DisplayName  *string
	Position     *string
	PhoneNumber  *string
	Organization *string
	Notes        *string
	Role         string
	TeamID       *int64
	Team         *TeamInfo
	IsActive     bool
	UpdatedAt    string
}

type ContactListQuery struct {
	Query           string
	TeamID          *int64
	Role            string
	Type            string
	IncludeInactive bool
	Page            int
	Limit           int
}

type ContactListResult struct {
	Items []ContactInfo
	Total int64
	Page  int
	Limit int
}

type UpdateContactInput struct {
	DisplayName *string
	Position    *string
	PhoneNumber *string
}

type ExternalContactInput struct {
	Type         string
	DisplayName  string
	PhoneNumber  string
	Organization *string
	Position     *string
	Notes        *string
	TeamID       *int64
	IsActive     *bool
}
