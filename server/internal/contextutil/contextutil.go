// Package contextutil provides safe helpers for working with context values
package contextutil

import (
	"context"
)

// private key type to avoid collisions
type ctxKey string

const (
	userIDKey  ctxKey = "user_id"
	roleKey    ctxKey = "user_roles"
	traceIDKey ctxKey = "trace_id"
)

// SetUserID returns a new context with the user ID stored
func SetUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// GetUserID retrieves the user ID from the context
func GetUserID(ctx context.Context) (string, bool) {
	val := ctx.Value(userIDKey)
	if id, ok := val.(string); ok {
		return id, true
	}
	return "", false
}

// SetUserRoles stores a slice of role names in the context
func SetUserRoles(ctx context.Context, roles []string) context.Context {
	return context.WithValue(ctx, roleKey, roles)
}

// GetUserRoles retrieves the roles slice from context
func GetUserRoles(ctx context.Context) ([]string, bool) {
	val := ctx.Value(roleKey)
	if roles, ok := val.([]string); ok {
		return roles, true
	}
	return nil, false
}

// SetTraceID stores a trace ID for request correlation
func SetTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceIDKey, traceID)
}

// GetTraceID retrieves the trace ID if set
func GetTraceID(ctx context.Context) (string, bool) {
	val := ctx.Value(traceIDKey)
	if id, ok := val.(string); ok {
		return id, true
	}
	return "", false
}
