package gosightauth

import (
	"context"

	"github.com/aaronlmathis/gosight/server/internal/contextutil"
	"github.com/aaronlmathis/gosight/shared/utils"
)

// HasPermission checks if the current context includes the given permission
func HasPermission(ctx context.Context, required string) bool {
	perms, ok := contextutil.GetUserPermissions(ctx)
	if !ok {
		utils.Debug("No permissions in context")
		return false
	}
	utils.Debug("🔍 Checking for permission: %s", required)
	utils.Debug("🧾 Available: %v", perms)

	for _, p := range perms {
		if p == required {
			utils.Debug("Permission matched: %s", p)
			return true
		}
	}
	utils.Debug("❌ Permission missing: %s", required)
	return false
}

func HasAnyPermission(ctx context.Context, required ...string) bool {
	perms, ok := contextutil.GetUserPermissions(ctx)
	if !ok {
		return false
	}
	permSet := make(map[string]struct{}, len(perms))
	for _, p := range perms {
		permSet[p] = struct{}{}
	}
	for _, r := range required {
		if _, ok := permSet[r]; ok {
			return true
		}
	}
	return false
}

func HasRole(roles []string, required string) bool {
	for _, r := range roles {
		if r == required {
			return true
		}
	}
	return false
}

func HasAnyRole(roles []string, requiredRoles ...string) bool {
	roleSet := make(map[string]struct{}, len(roles))
	for _, r := range roles {
		roleSet[r] = struct{}{}
	}
	for _, required := range requiredRoles {
		if _, ok := roleSet[required]; ok {
			return true
		}
	}
	return false
}
