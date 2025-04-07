package httpserver

import (
	"net/http"
	"path/filepath"
	"text/template"
)

func RenderMockupPage(w http.ResponseWriter, r *http.Request, templateDir string) {
	containerTemplate := filepath.Join(templateDir, "new.html")
	layoutTemplate := filepath.Join(templateDir, "layout2.html")
	tmpl, err := template.ParseFiles(containerTemplate, layoutTemplate)

	if err != nil {
		http.Error(w, "Template parsing error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// You can pass real data here if needed
	err = tmpl.ExecuteTemplate(w, "layout2.html", nil)
	if err != nil {
		http.Error(w, "Template execution error: "+err.Error(), http.StatusInternalServerError)
	}
}
