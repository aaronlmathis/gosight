package httpserver

import (
	"fmt"

	gosightauth "github.com/aaronlmathis/gosight/server/internal/auth"
	"github.com/aaronlmathis/gosight/server/internal/config"
	"github.com/aaronlmathis/gosight/server/internal/store/userstore"
)

func InitAuth(cfg *config.Config, userStore userstore.UserStore) (map[string]gosightauth.AuthProvider, error) {
	// Decode and store JWTSecret
	err := gosightauth.InitJWTSecret(cfg.Auth.JWTSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to decode JWT secret: %v", err)
	}

	// Decode and store MFASecret and JWTSecret
	err = gosightauth.InitMFAKey(cfg.Auth.MFASecret)
	if err != nil {
		return nil, fmt.Errorf("failed to decode MFA secret: %v", err)
	}

	// Build Auth Providers
	authProviders, err := buildAuthProviders(cfg, userStore)
	if err != nil {
		return nil, fmt.Errorf("failed to build auth providers: %v", err)
	}

	return authProviders, nil
}

func buildAuthProviders(cfg *config.Config, store userstore.UserStore) (map[string]gosightauth.AuthProvider, error) {
	providers := make(map[string]gosightauth.AuthProvider)

	for _, name := range cfg.Web.AuthProviders {
		switch name {
		case "local":
			providers["local"] = &gosightauth.LocalAuth{Store: store}

		case "google":
			providers["google"] = &gosightauth.GoogleAuth{
				OAuthConfig: cfg.Auth.Google.ToOAuthConfig(), // helper to build *oauth2.Config
				Store:       store,
			}

		default:
			return nil, fmt.Errorf("unsupported auth provider: %s", name)
		}
	}

	return providers, nil
}
