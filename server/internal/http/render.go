package httpserver

import (
	"html/template"
	"log"
	"os"
	"path/filepath"
)

func LoadTemplates(root string, funcMap template.FuncMap) (*template.Template, error) {
	tmpl := template.New("layout").Funcs(funcMap) // Keep the initial name

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && filepath.Ext(path) == ".html" {
			log.Printf("🧩 Parsing: %s", path)

			// Create a template name based on the relative path
			relativePath, err := filepath.Rel(root, path)
			if err != nil {
				log.Printf("❌ Failed to get relative path for %s: %v", path, err)
				return err
			}
			templateName := filepath.ToSlash(relativePath) // Use forward slashes for consistency

			parsed, err := template.New(templateName).Funcs(funcMap).ParseFiles(path)
			if err != nil {
				log.Printf("❌ Failed to parse %s: %v", path, err)
				return err
			}

			// Merge into master template set using the relative path as the name
			_, err = tmpl.AddParseTree(parsed.Name(), parsed.Tree)
			if err != nil {
				log.Printf("❌ Failed to add parse tree for %s: %v", path, err)
				return err
			}
		}
		return nil
	})

	return tmpl, err
}
