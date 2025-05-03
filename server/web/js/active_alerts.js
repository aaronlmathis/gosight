import { initTabs } from 'flowbite';

document.addEventListener("DOMContentLoaded", () => {
  loadActiveAlerts();

  document.getElementById("refresh-alerts")?.addEventListener("click", loadActiveAlerts);
  document.getElementById("alert-search")?.addEventListener("input", loadActiveAlerts);
  document.getElementById("filter-state")?.addEventListener("change", loadActiveAlerts);
  document.getElementById("filter-level")?.addEventListener("change", loadActiveAlerts);
});

async function loadActiveAlerts() {
  const tableBody = document.getElementById("active-alerts-table");
  tableBody.innerHTML = '<tr><td colspan="7" class="text-center py-6 text-gray-500">Loading...</td></tr>';

  const q = document.getElementById("alert-search")?.value.toLowerCase();
  const state = document.getElementById("filter-state")?.value;
  const level = document.getElementById("filter-level")?.value;

  let url = "/api/v1/alerts?state=ALARM&limit=1000";
  if (state) url += `&state=${state}`;
  if (level) url += `&level=${level}`;

  const res = await fetch(url);
  const data = await res.json();

  const filtered = data.filter(a => {
    const combined = `${a.rule_id} ${a.target} ${a.message}`.toLowerCase();
    return !q || combined.includes(q);
  });

  tableBody.innerHTML = "";

  for (const alert of filtered) {
    const row = document.createElement("tr");
    row.className = "hover:bg-blue-50 dark:hover:bg-gray-800 alert-row border-b border-gray-100 dark:border-gray-800";

    row.innerHTML = `
      <td class="px-4 py-2">${alert.rule_id}</td>
      <td class="px-4 py-2">${alert.state}</td>
      <td class="px-4 py-2 capitalize">${alert.level}</td>
      <td class="px-4 py-2">${alert.scope || '-'}</td>
      <td class="px-4 py-2">${alert.target || '-'}</td>
      <td class="px-4 py-2">${alert.fired_at}</td>
      <td class="px-4 py-2">
        <button class="text-sm text-blue-600 dark:text-blue-400 hover:underline" data-alert-id="${alert.id}">Expand</button>
      </td>
    `;

    const detailRow = document.createElement("tr");
    detailRow.className = "hidden";
    detailRow.id = `detail-${alert.id}`;
    detailRow.innerHTML = `
      <td colspan="7">
        <div class="p-4 bg-gray-50 dark:bg-gray-900 border-t border-gray-200 dark:border-gray-700">
          <h3 class="font-semibold mb-2">Message</h3>
          <p class="text-sm mb-4">${alert.message}</p>

          <div class="flex flex-wrap gap-2 mb-4">
            ${Object.entries(alert.tags || {}).map(([k, v]) => `
              <span class="px-2 py-1 text-xs font-semibold bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-300 rounded">${k}: ${v}</span>
            `).join('')}
          </div>

          <div id="incident-tabs-${alert.id}">
            <ul class="flex flex-wrap text-sm font-medium text-center border-b border-gray-200 dark:border-gray-700 mb-2" role="tablist">
              <li class="me-2"><button class="inline-block p-2 rounded-t-lg" data-tab-target="#logs-${alert.id}" type="button">Logs</button></li>
              <li class="me-2"><button class="inline-block p-2 rounded-t-lg" data-tab-target="#events-${alert.id}" type="button">Events</button></li>
              <li><button class="inline-block p-2 rounded-t-lg" data-tab-target="#chart-${alert.id}" type="button">Chart</button></li>
            </ul>
            <div class="border border-t-0 border-gray-200 dark:border-gray-700 p-4">
              <div id="logs-${alert.id}" class="hidden" role="tabpanel">
                <div class="text-gray-400 italic">Loading logs...</div>
              </div>
              <div id="events-${alert.id}" class="hidden" role="tabpanel">
                <div class="text-gray-400 italic">Loading events...</div>
              </div>
              <div id="chart-${alert.id}" class="hidden" role="tabpanel">
                <div class="text-gray-400 italic">Chart not implemented yet.</div>
              </div>
            </div>
          </div>
        </div>
      </td>
    `;

    row.querySelector("button[data-alert-id]").addEventListener("click", () => {
      const isVisible = !detailRow.classList.contains("hidden");
      document.querySelectorAll("tr[id^='detail-']").forEach(r => r.classList.add("hidden"));
      if (!isVisible) {
        detailRow.classList.remove("hidden");
        fetchIncidentContext(alert.id);
      }
    });

    tableBody.appendChild(row);
    tableBody.appendChild(detailRow);
  }

  if (filtered.length === 0) {
    tableBody.innerHTML = '<tr><td colspan="7" class="text-center py-6 text-gray-500">No active alerts found.</td></tr>';
  }
}

async function fetchIncidentContext(alertId) {
  const logsEl = document.getElementById(`logs-${alertId}`);
  const eventsEl = document.getElementById(`events-${alertId}`);

  const res = await fetch(`/api/v1/alerts/${alertId}/context`);
  const { logs, events } = await res.json();

  logsEl.innerHTML = logs.length
    ? logs.map(l => `<pre class="text-xs text-gray-700 dark:text-gray-300 font-mono">${l.Timestamp} ${l.Message}</pre>`).join('')
    : '<div class="text-gray-400 italic">No logs found in time window.</div>';

  eventsEl.innerHTML = events.length
    ? events.map(e => `<pre class="text-xs text-gray-700 dark:text-gray-300 font-mono">${e.Timestamp} ${e.Message}</pre>`).join('')
    : '<div class="text-gray-400 italic">No events found in time window.</div>';

  initTabs(); // re-initialize Flowbite tab logic
}