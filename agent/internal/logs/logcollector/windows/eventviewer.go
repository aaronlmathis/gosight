//go:build windows
// +build windows

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
package windowscollector

import (
	"context"
	"strings"
	"sync"
	"syscall"
	"time"
	"unicode/utf8"
	"unsafe"

	"github.com/aaronlmathis/gosight/agent/internal/config"
	"github.com/aaronlmathis/gosight/shared/model"
	"github.com/aaronlmathis/gosight/shared/utils"
	"golang.org/x/sys/windows"
)

var (
	modwevtapi     = windows.NewLazySystemDLL("wevtapi.dll")
	procEvtQuery   = modwevtapi.NewProc("EvtQuery")
	procEvtNext    = modwevtapi.NewProc("EvtNext")
	procEvtRender  = modwevtapi.NewProc("EvtRender")
	procEvtClose   = modwevtapi.NewProc("EvtClose")
)

const (
	EvtQueryChannelPath      = 0x1
	EvtQueryForwardDirection = 0x00000001
	EvtRenderEventXml        = 1
)

type EventViewerCollector struct {
	logName   string
	handle    syscall.Handle
	lines     chan model.LogEntry
	stop      chan struct{}
	wg        sync.WaitGroup
	batchSize int
	maxSize   int
}

func NewEventViewerCollector(cfg *config.Config, logName string) *EventViewerCollector {
	namePtr, err := syscall.UTF16PtrFromString(logName)
	if err != nil {
		utils.Error("Invalid log name: %v", err)
		return nil
	}

	h, _, callErr := procEvtQuery.Call(0, uintptr(unsafe.Pointer(namePtr)), 0, uintptr(EvtQueryChannelPath|EvtQueryForwardDirection))
	if h == 0 {
		utils.Error("EvtQuery failed: %v", callErr)
		return nil
	}

	c := &EventViewerCollector{
		logName:   logName,
		handle:    syscall.Handle(h),
		lines:     make(chan model.LogEntry, cfg.Agent.LogCollection.BatchSize*10),
		stop:      make(chan struct{}),
		batchSize: cfg.Agent.LogCollection.BatchSize,
		maxSize:   cfg.Agent.LogCollection.MessageMax,
	}

	c.wg.Add(1)
	go c.runReader()
	return c
}

func (e *EventViewerCollector) Name() string {
	return "eventviewer:" + e.logName
}

func (e *EventViewerCollector) runReader() {
	defer e.wg.Done()
	defer close(e.lines)

	buffer := make([]uint16, 65536) // 64KB

	for {
		select {
		case <-e.stop:
			procEvtClose.Call(uintptr(e.handle))
			return
		default:
		}

		var returned uint32
		eventHandles := make([]syscall.Handle, 10)
		r, _, _ := procEvtNext.Call(uintptr(e.handle), 10, uintptr(unsafe.Pointer(&eventHandles[0])), 1000, 0, uintptr(unsafe.Pointer(&returned)))
		if r == 0 || returned == 0 {
			time.Sleep(1 * time.Second)
			continue
		}

		for i := uint32(0); i < returned; i++ {
			var used, props uint32
			ret, _, _ := procEvtRender.Call(0, uintptr(eventHandles[i]), EvtRenderEventXml, uintptr(len(buffer)*2), uintptr(unsafe.Pointer(&buffer[0])), uintptr(unsafe.Pointer(&used)), uintptr(unsafe.Pointer(&props)))
			if ret == 0 {
				continue
			}
			xml := syscall.UTF16ToString(buffer[:used/2])
			entry := buildLogEntry(xml, e.maxSize)

			select {
			case e.lines <- entry:
			case <-e.stop:
				procEvtClose.Call(uintptr(eventHandles[i]))
				return
			default:
				utils.Warn("EventViewer log buffer full. Dropping entry.")
			}
			procEvtClose.Call(uintptr(eventHandles[i]))
		}
	}
}

func (e *EventViewerCollector) Collect(ctx context.Context) ([][]model.LogEntry, error) {
	var all [][]model.LogEntry
	var batch []model.LogEntry

collect:
	for {
		select {
		case log, ok := <-e.lines:
			if !ok {
				break collect
			}
			batch = append(batch, log)
			if len(batch) >= e.batchSize {
				all = append(all, batch)
				batch = nil
			}
		case <-ctx.Done():
			break collect
		default:
			break collect
		}
	}
	if len(batch) > 0 {
		all = append(all, batch)
	}
	return all, nil
}

func (e *EventViewerCollector) Close() error {
	close(e.stop)
	e.wg.Wait()
	return nil
}

func buildLogEntry(xml string, maxSize int) model.LogEntry {
	msg := xml
	if !utf8.ValidString(msg) {
		msg = strings.ToValidUTF8(msg, "\uFFFD")
	}
	if maxSize > 0 && len(msg) > maxSize {
		msg = msg[:maxSize] + " [truncated]"
	}
	return model.LogEntry{
		Timestamp: time.Now(),
		Level:     "info",
		Message:   msg,
		Category:  "eventviewer",
		Source:    "windows",
		Meta: &model.LogMeta{
			Platform: "eventviewer",
			Extra:    map[string]string{"raw_xml": msg},
		},
	}
}
