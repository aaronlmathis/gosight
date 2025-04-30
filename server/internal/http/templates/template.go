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

// server/internal/http/template.go
// Handle loading of template files

package templates

import (
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/aaronlmathis/gosight/server/internal/config"
	"github.com/aaronlmathis/gosight/server/internal/store/metricindex"
	"github.com/aaronlmathis/gosight/server/internal/store/metricstore"
	"github.com/aaronlmathis/gosight/server/internal/store/userstore"
	"github.com/aaronlmathis/gosight/server/internal/usermodel"
	"github.com/aaronlmathis/gosight/shared/model"
	"github.com/aaronlmathis/gosight/shared/utils"
	"github.com/fsnotify/fsnotify"
)

var (
	mu   sync.RWMutex
	Tmpl *template.Template
	fmap template.FuncMap
)

type TemplateData struct {
	Title       string
	User        *usermodel.User
	UserData    usermodel.SafeUser
	Permissions []string
	Metrics     map[string]float64
	Timeseries  map[string][]model.MetricPoint
	Tags        map[string]string
	Labels      map[string]string
	Meta        model.Meta
	MetricStore metricstore.MetricStore
	MetricIndex *metricindex.MetricIndex
	UserStore   userstore.UserStore
	Breadcrumbs []Breadcrumb
}

type Breadcrumb struct {
	Label string
	URL   string
}

func InitTemplates(cfg *config.Config, funcMap template.FuncMap) error {
	fmap = funcMap // store for reload use
	err := loadTemplates(cfg.Web.TemplateDir)
	if err != nil {
		return err
	}
	go watchForChanges(cfg.Web.TemplateDir)
	return nil
}

func loadTemplates(baseDir string) error {
	mu.Lock()
	defer mu.Unlock()

	newTmpl := template.New("layout").Funcs(fmap)
	counter := 0

	layoutDirs := []string{
		filepath.Join(baseDir, "layouts"),
		filepath.Join(baseDir, "partials"),
	}

	for _, dir := range layoutDirs {
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() || !strings.HasSuffix(path, ".html") {
				return nil
			}

			_, err = newTmpl.ParseFiles(path)
			if err != nil {
				utils.Error("error parsing template %v: %v", path, err)
				return err
			}
			counter++
			return nil
		})
		if err != nil {
			utils.Error("error walking templates in %s: %v", dir, err)
			return err
		}
	}

	Tmpl = newTmpl
	//utils.Debug("Loaded %d base layout/partial templates from %s", counter, baseDir)
	return nil
}

func watchForChanges(baseDir string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		utils.Error("fsnotify error: %v", err)
		return
	}
	defer watcher.Close()

	err = filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
		if info != nil && info.IsDir() {
			return watcher.Add(path)
		}
		return nil
	})
	if err != nil {
		utils.Error("Watcher setup failed: %v", err)
		return
	}

	for {
		select {
		case event := <-watcher.Events:
			if filepath.Ext(event.Name) == ".html" {
				//utils.Debug("Template changed: %s", event.Name)
				loadTemplates(baseDir)
			}
		case err := <-watcher.Errors:
			utils.Error("Watcher error: %v", err)
		}
	}
}

func RenderTemplate(w http.ResponseWriter, layout, page string, data any) error {
	mu.RLock()
	base := Tmpl
	mu.RUnlock()

	tmpl, err := base.Clone()
	if err != nil {
		utils.Error("Clone failed: %v", err)
		http.Error(w, "Template error", http.StatusInternalServerError)
		return err
	}

	pagePath := filepath.Join("web/templates/pages", page+".html") // TODO use config.
	tmpl, err = tmpl.ParseFiles(pagePath)
	if err != nil {
		utils.Error("❌ ParseFiles failed for %s: %v", pagePath, err)
		http.Error(w, "Template parse error", http.StatusInternalServerError)
		return err
	}
	utils.Debug("Parsed page template: %s", pagePath)

	layoutPath := "layouts/" + layout
	err = tmpl.ExecuteTemplate(w, layoutPath, data)
	if err != nil {
		utils.Error("ExecuteTemplate failed: %v", err)
		http.Error(w, "Render error", http.StatusInternalServerError)
		return err
	}

	return nil
}
