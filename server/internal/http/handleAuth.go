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

// File: server/internal/http/handleAuth.go
// Description: This file contains the authentication handlers for the GoSight server.

package httpserver

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	gosightauth "github.com/aaronlmathis/gosight/server/internal/auth"
	"github.com/aaronlmathis/gosight/server/internal/contextutil"
	"github.com/aaronlmathis/gosight/server/internal/usermodel"
	"github.com/aaronlmathis/gosight/shared/model"
	"github.com/aaronlmathis/gosight/shared/utils"
)

// startLogin handles the login/start route for various auth providers.
func (s *HttpServer) HandleLoginStart(w http.ResponseWriter, r *http.Request) {
	provider := r.URL.Query().Get("provider")
	if handler, ok := s.Sys.Auth[provider]; ok {
		handler.StartLogin(w, r)
		return
	}
	http.Error(w, "invalid provider", http.StatusBadRequest)
}

// handleCallback handles the callback route for various auth providers.
func (s *HttpServer) HandleCallback(w http.ResponseWriter, r *http.Request) {
	provider := r.URL.Query().Get("provider")

	handler, ok := s.Sys.Auth[provider]

	if !ok {
		http.Error(w, "invalid provider", http.StatusBadRequest)
		return
	}

	user, err := handler.HandleCallback(w, r)
	if err != nil {
		s.Sys.Tele.Emitter.Emit(r.Context(), model.EventEntry{
			Timestamp: time.Now(),
			Type:      "user.login.failed",
			Level:     "warning",
			Category:  "auth",
			Source:    fmt.Sprintf("auth:%s", provider),
			Scope:     "user",
			Target:    "", // no user yet
			Message:   fmt.Sprintf("Login failed via %s", provider),
			Meta: map[string]string{
				"provider":   provider,
				"ip":         utils.GetClientIP(r),
				"user_agent": r.UserAgent(),
				"reason":     "SSO callback error",
				"timestamp":  time.Now().Format(time.RFC3339),
			},
		})
		SetFlash(w, "Invalid username or password")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	user, err = s.Sys.Stores.Users.GetUserWithPermissions(r.Context(), user.ID)
	if err != nil {
		utils.Error("Failed to load roles for user %s: %v", user.Email, err)
		http.Error(w, "failed to load user roles", http.StatusInternalServerError)
		return
	}

	// Handle 2FA for local accounts
	if provider == "local" && user.TOTPSecret != "" {
		if gosightauth.CheckRememberMFA(r, user.ID) {
			ctx := InjectUserContext(r.Context(), user)
			s.createFinalSessionAndRedirect(w, r.WithContext(ctx), user)

			// Always emit login success event here (centralized)
			s.Sys.Tele.Emitter.Emit(r.Context(), model.EventEntry{
				Timestamp: time.Now(),
				Type:      "user.login.success",
				Level:     "info",
				Message:   fmt.Sprintf("User %s %s (%s) signed in", user.FirstName, user.LastName, user.Email),
				Target:    user.ID,
				Category:  "auth",
				Source:    "auth.session",
				Scope:     "user",
				Meta:      s.BuildAuthEventMeta(user, r),
			})
			return
		}

		gosightauth.SavePendingMFA(user.ID, w)

		// Optional: persist 'next' state
		if state := r.URL.Query().Get("state"); state != "" {
			http.SetCookie(w, &http.Cookie{
				Name:     "pending_next",
				Value:    state,
				Path:     "/",
				MaxAge:   300,
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
				Secure:   true, // TODO s.Config.Server.UseHTTPS,
			})
		}

		http.Redirect(w, r, "/mfa", http.StatusSeeOther)
		return
	}

	// Normal Login flow for SSO provider login
	ctx := InjectUserContext(r.Context(), user)
	utils.Debug("%s user: %s", provider, user.Email)
	utils.Debug("Roles: %v", user.Roles)
	utils.Debug("Permissions: %v", gosightauth.FlattenPermissions(user.Roles))

	// Set final session information and redirect user.
	s.createFinalSessionAndRedirect(w, r.WithContext(ctx), user)

	// Build Auth Event Entry
	eventEntry := model.EventEntry{
		Timestamp: time.Now(),
		Type:      "user.login.success",
		Level:     "info",
		Message:   fmt.Sprintf("User %s %s (%s) signed in", user.FirstName, user.LastName, user.Email),
		Target:    user.ID,
		Category:  "auth",
		Source:    fmt.Sprintf("auth:%s", provider),
		Scope:     "user",
		Meta:      s.BuildAuthEventMeta(user, r),
	}
	// Emit the auth event
	s.Sys.Tele.Emitter.Emit(ctx, eventEntry)
}

// HandleLogin renders the login page.
// It lists all available auth providers and handles flash messages.
func (s *HttpServer) HandleLogin(w http.ResponseWriter, r *http.Request) {
	next := r.URL.Query().Get("next")
	if next == "" {
		next = "/"
	}

	var providers []string
	for name := range s.Sys.Auth {
		providers = append(providers, name)
	}

	flash := GetFlash(w, r)

	data := map[string]any{
		"Next":      next,
		"Providers": providers,
		"Flash":     flash,
	}

	utils.Debug("Auth providers: %v", providers)
	utils.Debug("Template Data: %v", data)

	if err := s.Tmpl.RenderTemplate(w, "layout_auth", "login", data); err != nil {
		utils.Error("Failed to execute template: %v", err)
		http.Error(w, "template execution error", http.StatusInternalServerError)
	}
}

// MFA Handlers
// HandleMFA handles the MFA verification process.

func (s *HttpServer) HandleMFA(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		s.HandleMFAPage(w, r)
		return
	} else {
		ctx := r.Context()
		utils.Debug("Pending MFA for ")
		// POST - MFA verification
		userID, err := gosightauth.LoadPendingMFA(r)
		if err != nil || userID == "" {
			utils.Debug("No pending MFA for %s", userID)
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		utils.Debug("Code Check for %s", userID)
		code := r.FormValue("code")
		user, err := s.Sys.Stores.Users.GetUserByID(r.Context(), userID)
		if err != nil || !gosightauth.ValidateTOTP(user.TOTPSecret, code) {

			// Emit MFA failure event
			eventEntry := model.EventEntry{
				Timestamp: time.Now(),
				Type:      "user.mfa.failed",
				Level:     "warning",
				Message:   fmt.Sprintf("MFA failed for user ID %s", userID),
				Target:    user.ID,
				Category:  "auth",
				Source:    "auth.local",
				Scope:     "user",
				Meta:      s.BuildAuthEventMeta(user, r),
			}
			// Emit the auth event
			s.Sys.Tele.Emitter.Emit(ctx, eventEntry)
			http.Error(w, "Invalid TOTP code", http.StatusUnauthorized)
			return
		}
		utils.Debug("Role Check for %s", userID)
		// Load full roles/permissions
		user, err = s.Sys.Stores.Users.GetUserWithPermissions(r.Context(), user.ID)
		if err != nil {
			http.Error(w, "failed to load roles", http.StatusInternalServerError)
			return
		}
		utils.Debug("Got perms for %v", user)
		utils.Debug("MFA passed for %s", user.ID)

		if r.FormValue("remember") == "on" {
			utils.Debug("Setting remember_mfa cookie")
			gosightauth.SetRememberMFA(w, user.ID, r)
		}

		// Inject full context
		ctx = contextutil.SetUserID(r.Context(), user.ID)
		ctx = contextutil.SetUserRoles(ctx, gosightauth.ExtractRoleNames(user.Roles))
		ctx = contextutil.SetUserPermissions(ctx, gosightauth.FlattenPermissions(user.Roles))

		s.createFinalSessionAndRedirect(w, r.WithContext(ctx), user)

		// Emit MFA success event
		// Build Auth Event Entry
		eventEntry := model.EventEntry{
			Timestamp: time.Now(),
			Type:      "user.login.success",
			Level:     "info",
			Message:   fmt.Sprintf("User %s %s (%s) signed in", user.FirstName, user.LastName, user.Email),
			Target:    user.ID,
			Category:  "auth",
			Source:    "auth.local",
			Scope:     "user",
			Meta:      s.BuildAuthEventMeta(user, r),
		}
		// Emit the auth event
		s.Sys.Tele.Emitter.Emit(ctx, eventEntry)
	}
}

// HandleMFAPage renders the MFA page.

func (s *HttpServer) HandleMFAPage(w http.ResponseWriter, r *http.Request) {
	flash := GetFlash(w, r)

	data := map[string]any{
		"Flash": flash,
	}

	if err := s.Tmpl.RenderTemplate(w, "layout_auth", "mfa", data); err != nil {
		utils.Error("❌ Failed to execute template: %v", err)
		http.Error(w, "template execution error", http.StatusInternalServerError)
	}
}

// HandleLogout handles the logout process.
func (s *HttpServer) HandleLogout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "gosight_session",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   false, // TODO s.Config.Server.UseHTTPS,
	})
	ctx := r.Context()
	userID, _ := contextutil.GetUserID(ctx)

	s.Sys.Tele.Emitter.Emit(ctx, model.EventEntry{
		Timestamp: time.Now(),
		Type:      "user.logout",
		Level:     "info",
		Category:  "auth",
		Source:    "auth.logout",
		Scope:     "user",
		Target:    userID,
		Message:   fmt.Sprintf("User %s logged out", userID),
		Meta: map[string]string{
			"user_id":    userID,
			"ip":         utils.GetClientIP(r),
			"user_agent": r.UserAgent(),
		},
	})
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (s *HttpServer) createFinalSessionAndRedirect(w http.ResponseWriter, r *http.Request, user *usermodel.User) {
	traceID, _ := contextutil.GetTraceID(r.Context())
	roles := gosightauth.ExtractRoleNames(user.Roles)

	token, err := gosightauth.GenerateToken(user.ID, roles, traceID)
	if err != nil {
		utils.Error("Failed to generate session token: %v", err)
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	gosightauth.SetSessionCookie(w, token)

	// Clear any pending MFA/session leftovers
	gosightauth.ClearCookie(w, "pending_mfa")
	http.SetCookie(w, &http.Cookie{
		Name:     "pending_next",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
	})

	// handle redirect after login
	next := "/"
	if state := r.URL.Query().Get("state"); state != "" {
		if decoded, err := base64.URLEncoding.DecodeString(state); err == nil {
			next = string(decoded)
		}
	}
	if c, err := r.Cookie("pending_next"); err == nil {
		next = c.Value
	}
	http.Redirect(w, r, next, http.StatusSeeOther)
}
