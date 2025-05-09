package gosightauth

// google.go: Handles Google SSO login and callback.
// This file is part of the gosight project.

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"time"

	"github.com/aaronlmathis/gosight/server/internal/store/userstore"
	"github.com/aaronlmathis/gosight/server/internal/usermodel"
	"github.com/aaronlmathis/gosight/shared/utils"
	"golang.org/x/oauth2"
)

type GoogleAuth struct {
	OAuthConfig *oauth2.Config
	Store       userstore.UserStore
}

func (g *GoogleAuth) StartLogin(w http.ResponseWriter, r *http.Request) {
	next := r.URL.Query().Get("next")
	if next == "" {
		next = "/dashboard"
	}

	state := base64.URLEncoding.EncodeToString([]byte(next))
	url := g.OAuthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusFound)
}

func (g *GoogleAuth) HandleCallback(w http.ResponseWriter, r *http.Request) (*usermodel.User, error) {
	ctx := context.Background()
	code := r.URL.Query().Get("code")

	token, err := g.OAuthConfig.Exchange(ctx, code)
	if err != nil {
		return nil, err
	}

	client := g.OAuthConfig.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var userInfo struct {
		ID    string `json:"id"`
		Email string `json:"email"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, err
	}
	user, err := g.Store.GetUserByEmail(ctx, userInfo.Email)
	if err != nil {
		return nil, ErrUnauthorized 
	}

	//  If user has SSO info, verify match
	if user.SSOID != "" {
		if user.SSOID != userInfo.ID || user.SSOProvider != "google" {
			utils.Warn("SSO mismatch for %s: expected %s/%s, got %s/google",
				user.Email, user.SSOProvider, user.SSOID, userInfo.ID)
			return nil, ErrUnauthorized
		}
	} else {
		// First-time login — store SSO info
		user.SSOID = userInfo.ID
		user.SSOProvider = "google"
		utils.Info("First-time SSO link: %s → %s/%s", user.Email, user.SSOProvider, user.SSOID)
	}
	// Always update last_login
	user.LastLogin = time.Now()
	if err := g.Store.SaveUser(ctx, user); err != nil {
		utils.Warn("⚠Failed to update user login metadata: %v", err)
	}

	// Authorized user
	return user, nil
}
