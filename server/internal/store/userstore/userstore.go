// Package userstore defines the RBAC storage interface and types
package userstore

import (
	"context"

	"github.com/aaronlmathis/gosight/server/internal/usermodel"
)

type UserStore interface {
	GetUserByEmail(ctx context.Context, email string) (*usermodel.User, error)
	GetUserByUsername(ctx context.Context, username string) (*usermodel.User, error)
	GetUserWithPermissions(ctx context.Context, userID string) (*usermodel.User, error)
	SaveUser(ctx context.Context, u *usermodel.User) error
	CreateRole(ctx context.Context, r *usermodel.Role) error
	AssignRoleToUser(ctx context.Context, userID, roleID string) error
	CreatePermission(ctx context.Context, p *usermodel.Permission) error
	AttachPermissionToRole(ctx context.Context, roleID, permID string) error
}
