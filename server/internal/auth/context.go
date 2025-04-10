package gosightauth

import (
	"context"

	"github.com/aaronlmathis/gosight/server/internal/contextutil"
	"github.com/aaronlmathis/gosight/server/internal/usermodel"
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

	return ctx
}
