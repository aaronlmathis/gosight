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

package gosighttemplate

import (
	"context"
	"fmt"
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
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type GoSightTemplate struct {
	ctx         context.Context
	Cfg         *config.Config
	mu          sync.RWMutex
	Tmpl        *template.Template
	fmap        *template.FuncMap
	MetricStore metricstore.MetricStore
	MetricIndex *metricindex.MetricIndex
	UserStore   userstore.UserStore
}

// TemplateData holds the data passed to templates for rendering.
type TemplateData struct {
	Title       string
	User        *usermodel.User
	UserData    usermodel.SafeUser
	Permissions []string
	Metrics     map[string]float64 // TODO: Revaluate need.
	Timeseries  map[string][]model.MetricPoint
	Tags        map[string]string
	Labels      map[string]string
	Meta        *model.Meta
	Breadcrumbs []Breadcrumb
	CurrentPath string
}

// Breadcrumb represents a single breadcrumb in the navigation trail.
type Breadcrumb struct {
	Label string
	URL   string
}

// NewGoSightTemplate creates a new GoSightTemplate instance, loading templates and setting up file watchers.

func NewGoSightTemplate(
	ctx context.Context,
	cfg *config.Config,
	metricStore metricstore.MetricStore,
	metricIndex *metricindex.MetricIndex,
	userStore userstore.UserStore) (*GoSightTemplate, error) {

	fmap := createTemplateFunctionMap()
	t := &GoSightTemplate{
		ctx:         ctx,
		Cfg:         cfg,
		mu:          sync.RWMutex{},
		fmap:        fmap,
		MetricStore: metricStore,
		MetricIndex: metricIndex,
		UserStore:   userStore,
	}

	if err := t.loadTemplates(); err != nil {
		return nil, err
	}
	t.watchForChanges()

	return t, nil
}

// createTemplateFunctionMap initializes the function map for templates.
func createTemplateFunctionMap() *template.FuncMap {
	return &template.FuncMap{
		"hasPermission": HasPermission,
		"safeHTML":      SafeHTML,
		"title":         cases.Title(language.English).String,
		"marshal":       Marshal,
		"since":         Since,
		"uptime":        FormatUptime,
		"trim":          strings.TrimSpace,
		"div":           Div,
		"seq":           Seq,
		"hasPrefix":     HasPrefix,
	}
}

// loadTemplates loads all HTML templates from the configured directory.

func (t *GoSightTemplate) loadTemplates() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	newTmpl := template.New("layout").Funcs(*t.fmap)

	// Load layout/partials
	layoutDirs := []string{
		filepath.Join(t.Cfg.Web.TemplateDir, "layouts"),
		filepath.Join(t.Cfg.Web.TemplateDir, "partials"),
	}
	for _, dir := range layoutDirs {
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() || !strings.HasSuffix(path, ".html") {
				return nil
			}
			_, err = newTmpl.ParseFiles(path)
			return err
		})
		if err != nil {
			return fmt.Errorf("error parsing templates in %s: %w", dir, err)
		}
	}

	t.Tmpl = newTmpl
	return nil
}

// watchForChanges sets up a file watcher to reload templates when changes are detected.

func (t *GoSightTemplate) watchForChanges() {
	go func() {
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			utils.Error("failed to create template watcher: %v", err)
			return
		}
		defer watcher.Close()

		err = filepath.Walk(t.Cfg.Web.TemplateDir, func(path string, info os.FileInfo, err error) error {
			if info != nil && info.IsDir() {
				return watcher.Add(path)
			}
			return nil
		})
		if err != nil {
			utils.Error("watcher setup failed: %v", err)
			return
		}

		for {
			select {
			case event := <-watcher.Events:
				if filepath.Ext(event.Name) == ".html" {
					utils.Debug("Reloading templates due to change: %s", event.Name)
					if err := t.loadTemplates(); err != nil {
						utils.Error("template reload failed: %v", err)
					}
				}
			case err := <-watcher.Errors:
				utils.Error("template watcher error: %v", err)
			case <-t.ctx.Done():
				utils.Info("Template watcher shutting down")
				return
			}
		}
	}()
}

// BuildPageData constructs the TemplateData for rendering a page.
func (t *GoSightTemplate) BuildPageData(user *usermodel.User, breadcrumbs, labels map[string]string, path, title string,
	meta *model.Meta, permissions []string) *TemplateData {

	safeUser := usermodel.SafeUser{}
	if user != nil {
		safeUser = usermodel.SafeUser{
			Username:  user.Username,
			Email:     user.Email,
			FirstName: user.FirstName,
			LastName:  user.LastName,
		}
	} else {
		utils.Debug("No user provided, using empty SafeUser")
	}

	breadCrumbs := make([]Breadcrumb, 0, len(breadcrumbs))
	for name, url := range breadcrumbs {
		bc := Breadcrumb{
			Label: name,
		}
		if url != "" {
			bc.URL = url
		}

		breadCrumbs = append(breadCrumbs, bc)
	}

	if labels == nil {
		labels = make(map[string]string)
	}

	return &TemplateData{
		Title:       title,
		User:        user,
		UserData:    safeUser,
		Permissions: permissions,
		Metrics:     make(map[string]float64),
		Timeseries:  make(map[string][]model.MetricPoint),
		Tags:        make(map[string]string),
		Labels:      labels,
		Meta:        meta,
		CurrentPath: path,
		Breadcrumbs: breadCrumbs,
	}
}

// RenderTemplate renders a template with the given layout and page, passing in the provided data.
func (t *GoSightTemplate) RenderTemplate(w http.ResponseWriter, layout, page string, data any) error {
	t.mu.RLock()
	base := t.Tmpl
	t.mu.RUnlock()

	tmpl, err := base.Clone()
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return fmt.Errorf("failed to clone template: %w", err)
	}

	pagePath := filepath.Join(t.Cfg.Web.TemplateDir, "pages", page+".html")
	tmpl, err = tmpl.ParseFiles(pagePath)
	if err != nil {
		http.Error(w, "Template parse error", http.StatusInternalServerError)
		return fmt.Errorf("failed to parse page template: %w", err)
	}

	layoutPath := filepath.Join("layouts", layout)
	err = tmpl.ExecuteTemplate(w, layoutPath, data)
	if err != nil {
		http.Error(w, "Render error", http.StatusInternalServerError)
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return nil
}
