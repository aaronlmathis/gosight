package httpserver

import (
	"fmt"

	gosightauth "github.com/aaronlmathis/gosight/server/internal/auth"
	"github.com/aaronlmathis/gosight/server/internal/config"
	"github.com/aaronlmathis/gosight/server/internal/store/userstore"
)

func BuildAuthProviders(cfg *config.Config, store userstore.UserStore) (map[string]gosightauth.AuthProvider, error) {
	providers := make(map[string]gosightauth.AuthProvider)

	for _, name := range cfg.Web.AuthProviders {
		switch name {
		case "local":
			providers["local"] = &gosightauth.LocalAuth{Store: store}

		case "google":
			providers["google"] = &gosightauth.GoogleAuth{
				OAuthConfig: cfg.Google.ToOAuthConfig(), // helper to build *oauth2.Config
				Store:       store,
			}

		default:
			return nil, fmt.Errorf("unsupported auth provider: %s", name)
		}
	}

	return providers, nil
}
