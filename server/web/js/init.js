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

console.log("🚀 init.js loaded");
// server/web/js/init.js
import { toggleMenu } from './ui.js';
import { fetchAgents } from './api.js';
import { renderAgentTable } from './tables.js';
import { renderCPUGraph } from './charts.js';

window.toggleMenu = toggleMenu; // Global exposure for inline onclick

async function updateAgents() {
  try {
    const agents = await fetchAgents();
    renderAgentTable(agents);
    renderCPUGraph(document.getElementById('cpuChart'), agents);

    const lastUpdateEl = document.getElementById('lastUpdate');
    if (lastUpdateEl) {
      lastUpdateEl.textContent = `Last updated: ${new Date().toLocaleTimeString()}`;
    }
    console.log("⏱ Refreshing at", new Date().toLocaleTimeString());
  } catch (err) {
    console.error("❌ Agent update failed:", err);
  }
}

document.addEventListener('DOMContentLoaded', () => {
  updateAgents(); // initial load
  setInterval(updateAgents, 5000); // 🔁 refresh every 5s
});
