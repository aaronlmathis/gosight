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

package cache

import (
	"sync"

	"github.com/aaronlmathis/gosight/shared/model"
)

type LogCache interface {
	Add(entry model.StoredLog)
	Get(logID string) (*model.LogEntry, bool)
}

type logCache struct {
	mu    sync.RWMutex
	store map[string]model.StoredLog
}

func NewLogCache() LogCache {
	return &logCache{
		store: make(map[string]model.StoredLog),
	}
}

func (c *logCache) Add(entry model.StoredLog) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.store[entry.LogID] = entry
}

func (c *logCache) Get(logID string) (*model.LogEntry, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	entry, ok := c.store[logID]
	if !ok {
		return nil, false
	}
	return &entry.Log, true
}
