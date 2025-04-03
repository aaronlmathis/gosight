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

// gosight/server/internal/store/store.go


// server/web/js/init.js
import { toggleMenu } from './ui.js';
import { fetchAgents } from './api.js';
import { renderAgentTable } from './tables.js';
import { renderCPUGraph } from './charts.js';

window.toggleMenu = toggleMenu; // Global exposure for inline onclick

document.addEventListener('DOMContentLoaded', async () => {
  try {
    const agents = await fetchAgents();
    renderAgentTable(agents);

    const ctx = document.getElementById('cpuChart');
    renderCPUGraph(ctx, agents); // or fetch time-series from another API
  } catch (err) {
    console.error("Frontend init failed:", err);
  }
});