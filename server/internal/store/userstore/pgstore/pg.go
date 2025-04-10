// Package pgstore implements the userstore.Store interface using PostgreSQL
package pgstore

import (
	"context"
	"database/sql"

	"github.com/aaronlmathis/gosight/server/internal/usermodel"
)

type PGStore struct {
	db *sql.DB
}

func New(db *sql.DB) *PGStore {
	return &PGStore{db: db}
}

func (s *PGStore) GetUserByUsername(ctx context.Context, username string) (*usermodel.User, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, username, first_name, last_name, email, password_hash, mfa_secret FROM users WHERE username = $1
	`, username)

	u := &usermodel.User{}
	if err := row.Scan(&u.ID, &u.Username, &u.FirstName, &u.LastName, &u.Email, &u.PasswordHash, &u.TOTPSecret); err != nil {
		return nil, err
	}
	return u, nil
}

func (s *PGStore) GetUserByEmail(ctx context.Context, email string) (*usermodel.User, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, email, password_hash, mfa_secret FROM users WHERE email = $1
	`, email)

	u := &usermodel.User{}
	if err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.TOTPSecret); err != nil {
		return nil, err
	}
	return u, nil
}

func (s *PGStore) GetUserWithPermissions(ctx context.Context, userID string) (*usermodel.User, error) {
	u := &usermodel.User{ID: userID, Roles: []usermodel.Role{}}

	err := s.db.QueryRowContext(ctx, `
		SELECT email, password_hash, mfa_secret FROM users WHERE id = $1
	`, userID).Scan(&u.Email, &u.PasswordHash, &u.TOTPSecret)
	if err != nil {
		return nil, err
	}

	roleRows, err := s.db.QueryContext(ctx, `
		SELECT r.id, r.name, r.description
		FROM roles r
		JOIN user_roles ur ON ur.role_id = r.id
		WHERE ur.user_id = $1
	`, userID)
	if err != nil {
		return nil, err
	}
	defer roleRows.Close()

	for roleRows.Next() {
		var role usermodel.Role
		if err := roleRows.Scan(&role.ID, &role.Name, &role.Description); err != nil {
			return nil, err
		}

		permRows, err := s.db.QueryContext(ctx, `
			SELECT p.id, p.name, p.description
			FROM permissions p
			JOIN role_permissions rp ON rp.permission_id = p.id
			WHERE rp.role_id = $1
		`, role.ID)
		if err != nil {
			return nil, err
		}

		for permRows.Next() {
			var perm usermodel.Permission
			if err := permRows.Scan(&perm.ID, &perm.Name, &perm.Description); err != nil {
				permRows.Close()
				return nil, err
			}
			role.Permissions = append(role.Permissions, perm)
		}
		permRows.Close()

		u.Roles = append(u.Roles, role)
	}

	return u, nil
}

func (s *PGStore) AssignRoleToUser(ctx context.Context, userID, roleID string) error {
	return nil
}

func (s *PGStore) SaveUser(ctx context.Context, u *usermodel.User) error   { return nil }
func (s *PGStore) CreateRole(ctx context.Context, r *usermodel.Role) error { return nil }

func (s *PGStore) CreatePermission(ctx context.Context, p *usermodel.Permission) error { return nil }
func (s *PGStore) AttachPermissionToRole(ctx context.Context, roleID, permID string) error {
	return nil
}
