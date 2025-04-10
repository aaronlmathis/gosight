package gosightauth

// google.go: Handles Google SSO login and callback.
// This file is part of the gosight project.

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/aaronlmathis/gosight/server/internal/store/userstore"
	"github.com/aaronlmathis/gosight/server/internal/usermodel"
	"golang.org/x/oauth2"
)

type GoogleAuth struct {
	OAuthConfig *oauth2.Config
	Store       userstore.UserStore
}

func (g *GoogleAuth) StartLogin(w http.ResponseWriter, r *http.Request) {
	url := g.OAuthConfig.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
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
		// 🛑 Optional: return unauthorized if not found
		return nil, ErrUnauthorized // you can define this in errors.go
	}

	// ✅ Authorized user
	return user, nil
}
