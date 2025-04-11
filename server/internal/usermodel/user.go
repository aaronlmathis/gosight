package usermodel

import "time"

type User struct {
	ID            string
	Username      string
	Email         string
	PasswordHash  string
	FirstName     string
	LastName      string
	MFAEnabled    bool
	MFAMethod     string // "totp", "webauthn"
	TOTPSecret    string
	WebAuthnCreds []byte
	SSOProvider   string
	SSOID         string
	Roles         []Role
	CreatedAt     time.Time
	LastLogin     time.Time
	Scopes        map[string][]string
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
