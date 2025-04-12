package httpserver

import (
	"net/http"
	"time"

	gosightauth "github.com/aaronlmathis/gosight/server/internal/auth"
	"github.com/aaronlmathis/gosight/server/internal/http/templates"
	"github.com/aaronlmathis/gosight/shared/utils"
)

func HandleLoginPage(w http.ResponseWriter, r *http.Request, authProviders map[string]gosightauth.AuthProvider, templateDir string) {
	next := r.URL.Query().Get("next")
	if next == "" {
		next = "/"
	}

	var providers []string
	for name := range authProviders {
		providers = append(providers, name)
	}
	data := map[string]any{
		"Next":      next,
		"Providers": providers,
	}
	utils.Debug("Auth providers: %v", authProviders)
	utils.Debug("Template Data: %v", data)
	err := templates.RenderTemplate(w, "dashboard/login", data)
	if err != nil {
		utils.Error("❌ Failed to execute template: %v", err)
		http.Error(w, "template execution error", http.StatusInternalServerError)
	}
}

func HandleLogout(w http.ResponseWriter, r *http.Request) {
	// Overwrite cookie with expired one
	http.SetCookie(w, &http.Cookie{
		Name:     "gosight_session",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0), // force expiration
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   false, // or true if you're using HTTPS
	})

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
