package gosightauth

// server/internal/auth/errors.go

import "errors"

var (
	ErrInvalidPassword = errors.New("invalid password")
	ErrInvalidTOTP     = errors.New("invalid TOTP code")
	ErrUserNotFound    = errors.New("user not found")
	ErrUnauthorized    = errors.New("unauthorized")
)
