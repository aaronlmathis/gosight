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
// Package routestore provides functionality to load and manage action routes
package routestore

import (
	"os"

	"github.com/aaronlmathis/gosight/shared/model"
	"gopkg.in/yaml.v3"
)

// RouteStore is a structure that holds a list of action routes.
type RouteStore struct {
	Routes []model.ActionRoute
}

// LoadRoutesFromFile loads action routes from a YAML file.
func NewRouteStore(path string) (*RouteStore, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config model.ActionRouteSet
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &RouteStore{Routes: config.Routes}, nil
}
