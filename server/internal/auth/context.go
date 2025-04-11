package gosightauth

import (
	"context"

	"github.com/aaronlmathis/gosight/server/internal/contextutil"
	"github.com/aaronlmathis/gosight/server/internal/usermodel"
	"github.com/aaronlmathis/gosight/shared/utils"
)

func InjectSessionContext(ctx context.Context, user *usermodel.User) context.Context {
	ctx = contextutil.SetUserID(ctx, user.ID)

	roleNames := ExtractRoleNames((user.Roles))
	ctx = contextutil.SetUserRoles(ctx, roleNames)

	var permNames []string
	seen := map[string]struct{}{}
	for _, role := range user.Roles {
		for _, perm := range role.Permissions {
			if _, exists := seen[perm.Name]; !exists {
				permNames = append(permNames, perm.Name)
				seen[perm.Name] = struct{}{}
			}
		}
	}
	ctx = contextutil.SetUserPermissions(ctx, permNames)
	utils.Debug("🔐 Injected user: %s", user.ID)
	utils.Debug("🔐 Roles: %v", roleNames)
	utils.Debug("🔐 Permissions: %v", permNames)
	return ctx
}
