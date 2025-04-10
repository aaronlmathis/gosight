// Package userstore defines the RBAC storage interface and types
package userstore

import "context"

type UserStore interface {
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	GetUserWithPermissions(ctx context.Context, userID string) (*User, error)
	SaveUser(ctx context.Context, u *User) error

	CreateRole(ctx context.Context, r *Role) error
	AssignRoleToUser(ctx context.Context, userID, roleID string) error

	CreatePermission(ctx context.Context, p *Permission) error
	AttachPermissionToRole(ctx context.Context, roleID, permID string) error
}

type User struct {
	ID           string
	Email        string
	PasswordHash string
	MFASecret    string
	Roles        []Role
}

type Role struct {
	ID          string
	Name        string
	Description string
	Permissions []Permission
}

type Permission struct {
	ID          string
	Name        string
	Description string
}
