package httpserver

import (
	"context"

	gosightauth "github.com/aaronlmathis/gosight/server/internal/auth"
	"github.com/aaronlmathis/gosight/server/internal/contextutil"
	"github.com/aaronlmathis/gosight/server/internal/usermodel"
)

func InjectUserContext(ctx context.Context, user *usermodel.User) context.Context {
	ctx = contextutil.SetUserID(ctx, user.ID)
	roles := gosightauth.ExtractRoleNames(user.Roles)
	ctx = contextutil.SetUserRoles(ctx, roles)
	ctx = contextutil.SetUserPermissions(ctx, gosightauth.FlattenPermissions(user.Roles))
	return ctx
}
