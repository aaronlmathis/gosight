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
// Package model contains the data structures used in GoSight.
// agent/processes/processcollector/processes.go

package processcollector

import (
	"context"
	"time"

	"github.com/shirou/gopsutil/v4/process"

	"github.com/aaronlmathis/gosight/shared/model"
)

// Collector captures running processes
func CollectProcesses(ctx context.Context) (*model.ProcessSnapshot, error) {
	procs, err := process.ProcessesWithContext(ctx)
	if err != nil {
		return nil, err
	}

	snapshot := &model.ProcessSnapshot{
		Timestamp: time.Now(),
		Processes: make([]model.ProcessInfo, 0, len(procs)),
	}

	for _, p := range procs {
		info := model.ProcessInfo{PID: int(p.Pid)}

		if pp, err := p.PpidWithContext(ctx); err == nil {
			info.PPID = int(pp)
		}
		if exe, err := p.ExeWithContext(ctx); err == nil {
			info.Executable = exe
		}
		if cl, err := p.CmdlineWithContext(ctx); err == nil {
			info.Cmdline = cl
		}
		if u, err := p.UsernameWithContext(ctx); err == nil {
			info.User = u
		}
		if cpu, err := p.CPUPercentWithContext(ctx); err == nil {
			info.CPUPercent = cpu
		}
		if mem, err := p.MemoryPercentWithContext(ctx); err == nil {
			info.MemPercent = float64(mem)
		}
		if threads, err := p.NumThreadsWithContext(ctx); err == nil {
			info.Threads = int(threads)
		}
		if start, err := p.CreateTimeWithContext(ctx); err == nil {
			info.StartTime = time.UnixMilli(start)
		}

		snapshot.Processes = append(snapshot.Processes, info)
	}

	return snapshot, nil
}
