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
// server/web/utils/ui.js


export function debounce(fn, ms) {
  let timeout;
  return (...args) => {
    clearTimeout(timeout);
    timeout = setTimeout(() => fn.apply(this, args), ms);
  };
}

// Format metric value based on naming convention and scale
export function formatMetricValue(value, name = '') {
  const suffix = typeof name === 'string' ? name.toLowerCase() : '';

  if (suffix.endsWith("_percent") || suffix.endsWith("_usage") || (suffix.includes("cpu") && value <= 100)) {
    return `${value.toFixed(2)}%`;
  }

  if (suffix.endsWith("_bytes") || suffix.startsWith("disk_")) {
    const kb = value / 1024;
    const mb = kb / 1024;
    const gb = mb / 1024;
    return `${value.toFixed(2)} bytes (${kb.toFixed(2)} KB / ${mb.toFixed(2)} MB / ${gb.toFixed(2)} GB)`;
  }

  if (suffix.endsWith("_kb")) {
    const mb = value / 1024;
    const gb = mb / 1024;
    return `${value.toFixed(2)} KB (${mb.toFixed(2)} MB / ${gb.toFixed(2)} GB)`;
  }

  if (suffix.endsWith("_mb")) {
    const gb = value / 1024;
    return `${value.toFixed(2)} MB (${gb.toFixed(2)} GB)`;
  }

  return value.toFixed(2);
}

  // Format epoch or ISO timestamp into localized 12-hour date string
  export function formatTimestamp(ts) {
    const d = typeof ts === "string" ? new Date(ts) : new Date(ts * 1000);
    return d.toLocaleString("en-US", {
      year: "numeric",
      month: "long",
      day: "numeric",
      hour: "numeric",
      minute: "2-digit",
      hour12: true,
      timeZoneName: "short"
    });
  }
  
  // Format uptime seconds into D H M
  export function formatUptime(seconds) {
    const d = Math.floor(seconds / 86400);
    const h = Math.floor((seconds % 86400) / 3600);
    const m = Math.floor((seconds % 3600) / 60);
    return `${d}d ${h}h ${m}m`;
  }
  