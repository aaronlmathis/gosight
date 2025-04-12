package httpserver

import (
	"net/http"
	"path/filepath"
	"text/template"
	"time"

	gosightauth "github.com/aaronlmathis/gosight/server/internal/auth"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
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

	tmpl, err := template.New("login.html").
		Funcs(template.FuncMap{
			"title": cases.Title(language.English).String,
		}).
		ParseFiles(filepath.Join(templateDir, "login.html"))
	if err != nil {
		http.Error(w, "template parse error", http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, map[string]any{
		"Next":      next,
		"Providers": providers,
	})
	if err != nil {
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
