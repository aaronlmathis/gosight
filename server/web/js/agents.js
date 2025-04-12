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

// gosight/server/internal/web/js/agents.js
// server/web/js/agents.js
import { fetchAgents } from './api.js';
import { renderAgentTable } from './tables.js';

async function updateAgents() {
  try {
    const agents = await fetchAgents();
    renderAgentTable(agents);

    const lastUpdateEl = document.getElementById('lastUpdate');
    if (lastUpdateEl) {
      lastUpdateEl.textContent = `Last updated: ${new Date().toLocaleTimeString()}`;
    }
    console.log("⏱ Refreshed agent data at", new Date().toLocaleTimeString());
  } catch (err) {
    console.error("❌ Agent update failed:", err);
  }
}

function filterTable() {
  const search = document.getElementById("search").value.toLowerCase();
  const status = document.getElementById("statusFilter").value;

  document.querySelectorAll("#agentTableBody tr").forEach((row) => {
    const text = row.textContent.toLowerCase();
    const matchesSearch = text.includes(search);
    const matchesStatus = !status || row.innerText.includes(status);

    row.style.display = matchesSearch && matchesStatus ? "" : "none";
  });
}

function exportCSV() {
  const rows = Array.from(document.querySelectorAll("#agentTableBody tr"))
    .map((tr) =>
      Array.from(tr.children).map((td) => `"${td.textContent.trim()}"`).join(",")
    );

  const csv = ["Hostname,IP,OS,Last Seen,Status", ...rows].join("\n");
  const blob = new Blob([csv], { type: "text/csv" });
  const url = URL.createObjectURL(blob);

  const a = document.createElement("a");
  a.href = url;
  a.download = "agents.csv";
  a.click();
  URL.revokeObjectURL(url);
}

document.addEventListener("DOMContentLoaded", () => {
  updateAgents();
  setInterval(updateAgents, 5000); // auto-refresh

  document.getElementById("search").addEventListener("input", filterTable);
  document.getElementById("statusFilter").addEventListener("change", filterTable);
  document.getElementById("viewToggle").addEventListener("click", () => {
    document.querySelector("#agentTableBody").classList.toggle("text-xs");
  });
  document.getElementById("exportBtn").addEventListener("click", exportCSV);
});
