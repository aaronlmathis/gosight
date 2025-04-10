package gosightauth

import "github.com/aaronlmathis/gosight/server/internal/usermodel"

func FlattenPermissions(roles []usermodel.Role) []string {
	perms := map[string]struct{}{}
	for _, role := range roles {
		for _, p := range role.Permissions {
			perms[p.Name] = struct{}{}
		}
	}
	var result []string
	for p := range perms {
		result = append(result, p)
	}
	return result
}

func ExtractRoleNames(roles []usermodel.Role) []string {
	names := make([]string, 0, len(roles))
	for _, r := range roles {
		names = append(names, r.Name)
	}
	return names
}
