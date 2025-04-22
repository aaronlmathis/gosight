/*
SPDX-License-Identifier: GPL-3.0-or-later

Copyright (C) 2025 Aaron Mathis aaron.mathis@gmail.com

This file is part of GoSight.

GoSight is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

GoSight is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with GoSight. If not, see https://www.gnu.org/licenses/.
*/

// gosight/agent/internal/bootstrap/auth.go

package bootstrap

import (
	"fmt"

	gosightauth "github.com/aaronlmathis/gosight/server/internal/auth"
	"github.com/aaronlmathis/gosight/server/internal/config"
	"github.com/aaronlmathis/gosight/server/internal/store/userstore"
)

// InitAuth initializes the authentication providers for the GoSight server.
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

// buildAuthProviders builds the authentication providers based on the configuration.
// It returns a map of provider names to their respective AuthProvider implementations.

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
