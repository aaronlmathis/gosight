package httpserver

import (
	"net/http"
	"path/filepath"
	"text/template"

	gosightauth "github.com/aaronlmathis/gosight/server/internal/auth"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func HandleLoginPage(w http.ResponseWriter, r *http.Request, authProviders map[string]gosightauth.AuthProvider, templateDir string) {
	next := r.URL.Query().Get("next")
	if next == "" {
		next = "/dashboard"
	}

	// Extract the list of enabled providers
	var providers []string
	for name := range authProviders {
		providers = append(providers, name)
	}
	tmpl := template.Must(
		template.New("login.html").
			Funcs(template.FuncMap{
				"title": cases.Title(language.English).String,
			}).
			ParseFiles(filepath.Join(templateDir, "login.html")),
	)
	err := tmpl.Execute(w, map[string]any{
		"Next":      next,
		"Providers": providers,
	})
	if err != nil {
		http.Error(w, "template error", http.StatusInternalServerError)
		return
	}
}
