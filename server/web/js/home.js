
import { escapeHTML } from "./format.js";
import { gosightFetch } from './api.js';
//
// AGENT CARD
//
export async function loadAgentCard() {
  try {
    const res = await gosightFetch("/api/v1/endpoints/hosts");
    const agents = await res.json();

    console.log("Agents:", agents);

    if (!Array.isArray(agents) || agents.length === 0) {
      console.warn("⚠️ No agent data available");
      return;
    }

    const online = agents.filter(a => a.status === "Online").length;
    const total = agents.length;
    const offline = total - online;

    // Update text
    document.getElementById("agent-health-status").textContent = `${online} Online / ${offline} Offline`;

    // Render radial chart
    const percent = total > 0 ? (online / total) * 100 : 0;
    new ApexCharts(document.getElementById("radial-agents"), {
      chart: { type: "radialBar", height: 200 },
      series: [percent],
      labels: ["Online %"],
      colors: ["#10b981"],
      plotOptions: {
        radialBar: {
          hollow: { size: "60%" },
          dataLabels: {
            name: { fontSize: "12px" },
            value: { fontSize: "16px", fontWeight: 600, formatter: v => `${v.toFixed(0)}%` },
          },
        },
      },
    }).render();
  } catch (err) {
    console.error("Agent summary failed:", err);
    document.getElementById("agent-health-status").textContent = "Unavailable";
  }
}
//

//
// ALERT RADIAL
//
export async function loadAlertRadial() {
  try {
    const res = await gosightFetch("/api/v1/alerts");
    const alerts = await res.json();
    const count = Array.isArray(alerts) ? alerts.length : 0;

    new ApexCharts(document.getElementById("radial-alerts"), {
      chart: { type: "radialBar", height: 200 },
      series: [Math.min(count * 10, 100)], // scale for visual only
      labels: ["Alerts"],
      colors: ["#ef4444"],
      plotOptions: {
        radialBar: {
          hollow: { size: "60%" },
          dataLabels: {
            name: { fontSize: "12px" },
            value: { fontSize: "16px", fontWeight: 600, formatter: v => `${count}` },
          },
        },
      },
    }).render();
  } catch (err) {
    console.error("Alert radial failed:", err);
  }
}

//
// LOAD CONTAINER RADIAL
//
export async function loadContainerCard() {
  try {
    const runtimes = ["podman", "docker"];
    const metricNames = ["cpu_percent", "uptime_seconds"];
    const metrics = runtimes.flatMap(rt => metricNames.map(name => `container.${rt}.${name}`));
    const query = metrics.map(m => `metric=${m}`).join("&");

    const res = await gosightFetch(`/api/v1/query?${query}`);
    const rows = await res.json();

    const containers = new Map();

    for (const row of rows) {
      const tags = row.tags || {};
      const id = tags.container_id;
      if (!id) continue;

      if (!containers.has(id)) {
        containers.set(id, { id, status: tags.status || "unknown" });
      }
    }

    const total = containers.size;
    const running = [...containers.values()].filter(c => c.status === "running").length;
    const percent = total > 0 ? (running / total) * 100 : 0;

    // Render radial chart
    new ApexCharts(document.getElementById("radial-containers"), {
      chart: { type: "radialBar", height: 200 },
      series: [percent],
      labels: ["Containers"],
      colors: ["#3b82f6"],
      plotOptions: {
        radialBar: {
          hollow: { size: "60%" },
          dataLabels: {
            name: { fontSize: "12px" },
            value: {
              fontSize: "16px",
              fontWeight: 600,
              formatter: () => `${running} / ${total}`
            }
          }
        }
      }
    }).render();
  } catch (err) {
    console.error("Failed to load container radial:", err);
  }
}
//
// LOAD SYSTEM LOAD RADIAL
//
export async function loadSystemLoadRadial() {
  const now = new Date().toISOString();
  const start = new Date(Date.now() - 10 * 60 * 1000).toISOString(); // 10 minutes ago

  const url = `/api/v1/query?metric=system.cpu.usage_percent&scope=total&start=${encodeURIComponent(start)}&end=${encodeURIComponent(now)}`;
  console.log(url)
  try {
    const rows = await fetchJson(url);
    console.log("Raw metric rows: ", rows);

    if (!Array.isArray(rows) || rows.length === 0) throw new Error("No CPU data");

    const values = rows.map(r => r.value).filter(v => typeof v === "number");
    const avg = values.reduce((sum, v) => sum + v, 0) / values.length;

    const chartEl = document.getElementById("radial-load");
    if (!chartEl) throw new Error("Chart element not found");

    new ApexCharts(chartEl, {
      chart: { type: "radialBar", height: 200 },
      series: [avg],
      labels: ["CPU Load"],
      colors: ["#f59e0b"],
      plotOptions: {
        radialBar: {
          hollow: { size: "60%" },
          dataLabels: {
            name: { fontSize: "12px" },
            value: { fontSize: "10px", fontWeight: 600, formatter: v => `${v.toFixed(1)}%` },
          },
        },
      },
    }).render();
  } catch (err) {
    console.error("loadSystemLoadRadial failed:", err);
  }
}

//
// load Top Containers by CPU
export async function loadTopContainersByCpu() {
  try {
    const url = "/api/v1/query?metric=container.podman.cpu_percent&metric=container.docker.cpu_percent&sort=desc&limit=5";
    const rows = await fetchJson(url);
    if (!rows || !Array.isArray(rows)) return;

    const labels = rows.map(row =>
      row.tags?.name ||
      row.tags?.instance ||
      row.tags?.container_id?.slice(0, 12) ||
      "unknown"
    );
    const values = rows.map(row => parseFloat(row.value.toFixed(2)));

    new ApexCharts(document.getElementById("top-cpu-containers"), {
      chart: {
        type: "bar",
        height: 180
      },
      series: [{
        name: "CPU %",
        data: values
      }],
      xaxis: {
        categories: labels
      },
      plotOptions: {
        bar: {
          horizontal: true
        }
      },
      colors: ["#f97316"]
    }).render();
  } catch (err) {
    console.error("Failed to load top containers:", err);
  }
}
// 
// load Top Containers by Memory
// 

export async function loadTopContainersByMemory() {
  try {
    const url = "/api/v1/query?metric=container.podman.mem_usage_bytes&sort=desc&limit=5";
    const rows = await fetchJson(url);
    if (!rows || !Array.isArray(rows)) return;

    const labels = rows.map(row =>
      row.tags?.name ||
      row.tags?.instance ||
      row.tags?.container_id?.slice(0, 12) ||
      "unknown"
    );

    const values = rows.map(row => (row.value / 1024 / 1024).toFixed(1)); // Convert bytes to MB

    new ApexCharts(document.getElementById("top-mem-containers"), {
      chart: {
        type: "bar",
        height: 180
      },
      series: [{
        name: "Memory (MB)",
        data: values
      }],
      xaxis: {
        categories: labels
      },
      plotOptions: {
        bar: {
          horizontal: true
        }
      },
      colors: ["#6366f1"]
    }).render();
  } catch (err) {
    console.error("Failed to load top memory containers:", err);
  }
}


async function fetchJson(url) {
  try {
    const res = await gosightFetch(url);
    if (!res.ok) throw new Error("Failed to fetch");
    return await res.json();
  } catch (e) {
    console.error("Fetch error", e);
    return null;
  }
}

// 
// LOAD TREND
//

export async function loadTrend(url, elId, label) {
  const points = await fetchJson(url);
  if (!points || !Array.isArray(points)) return;
  const values = points.map(p => p.value);
  new ApexCharts(document.getElementById(elId), {
    chart: { type: "area", height: 100, sparkline: { enabled: true } },
    series: [{ name: label, data: values }],
    stroke: { curve: "smooth" },
    colors: ["#3b82f6"],
    tooltip: { enabled: true },
  }).render();
}
//
// UPDATE ALERT SUMMARY
//
export async function updateAlertSummary() {
  try {
    const res = await gosightFetch("/api/v1/alerts");
    const alerts = await res.json();
    console.log("Alerts:", alerts);

    if (!Array.isArray(alerts)) {
      console.warn("Expected array, got:", alerts);
      document.getElementById("alert-status").textContent = "Unavailable";
      return;
    }

    const counts = alerts.reduce((acc, alert) => {
      const level = alert.level || "unknown";
      acc[level] = (acc[level] || 0) + 1;
      return acc;
    }, {});

    const summary = Object.entries(counts)
      .map(([level, count]) => `${count} ${capitalize(level)}`)
      .join(", ") || "None";

    const el = document.getElementById("alert-status");
    if (el) el.textContent = summary;
    else console.warn("alert-status element not found");

  } catch (err) {
    console.error("Failed to fetch alerts:", err);
    const el = document.getElementById("alert-status");
    if (el) el.textContent = "Unavailable";
  }
}


function capitalize(str) {
  return str.charAt(0).toUpperCase() + str.slice(1);
}
/*
async function loadTop(url, elId, label, unit) {
    const items = await fetchJson(url);
    if (!items || !Array.isArray(items)) return;
    const categories = items.map(i => i.labels?.instance || i.labels?.name || i.labels?.id || "unknown");
    const data = items.map(i => parseFloat(i.value.toFixed(1)));
    new ApexCharts(document.getElementById(elId), {
        chart: { type: "bar", height: 180 },
        series: [{ name: label, data }],
        xaxis: { categories },
        plotOptions: { bar: { horizontal: true } },
        colors: ["#6366f1"],
    }).render();
}*/

//
// render Event Log
//
function renderEventLog(events) {
  const el = document.getElementById("event-log");
  el.innerHTML = "";

  events.forEach((e, idx) => {
    const level = (e.level || "info").toLowerCase();
    const levelLabel = level.charAt(0).toUpperCase() + level.slice(1);
    const scope = e.scope || "unknown";
    const timestamp = new Date(e.timestamp).toLocaleTimeString();
    const message = e.message || e.summary || "—";
    const id = e.id || "";

    // Context badge based on scope
    const contextHint = (() => {
      switch (scope) {
        case "agent": return e.meta?.hostname;
        case "container": return e.meta?.name || e.meta?.image;
        case "user": return e.meta?.email || e.meta?.user_id;
        case "rule": return e.meta?.rule_id || e.meta?.metric_name;
        default: return "";
      }
    })();

    const levelClass = {
      info: "bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200",
      warning: "bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200",
      error: "bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200",
    }[level] || "bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-200";

    const scopeClass = {
      agent: "bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200",
      container: "bg-purple-100 text-purple-800 dark:bg-purple-900 dark:text-purple-200",
      user: "bg-pink-100 text-pink-800 dark:bg-pink-900 dark:text-pink-200",
      rule: "bg-indigo-100 text-indigo-800 dark:bg-indigo-900 dark:text-indigo-200",
      system: "bg-gray-200 text-gray-800 dark:bg-gray-700 dark:text-gray-100",
    }[scope] || "bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-200";

    const rowShade = idx % 2 === 0
      ? "bg-gray-50 dark:bg-gray-800"
      : "bg-white dark:bg-gray-900";

    el.innerHTML += `
      <a href="/events/${id}" class="block px-4 py-3 ${rowShade} hover:bg-gray-100 dark:hover:bg-gray-700 transition" title="${escapeHTML(message)}">
        <div class="flex items-center gap-3 text-sm text-gray-800 dark:text-gray-200">
          <span class="inline-block px-2 py-0.5 rounded text-xs font-semibold ${levelClass}">${levelLabel}</span>
          <span class="inline-block px-2 py-0.5 rounded text-xs font-medium ${scopeClass}">${escapeHTML(scope)}</span>
          <span class="text-xs text-gray-500 dark:text-gray-400">${timestamp}</span>
          <span class="truncate flex-1">${escapeHTML(message)}</span>
          ${contextHint ? `<span class="text-xs italic text-gray-400 dark:text-gray-500 ml-2">${escapeHTML(contextHint)}</span>` : ""}
        </div>
      </a>
    `;
  });
}






export async function initHome() {
  // Load core radial cards
  await loadAgentCard();
  await loadAlertRadial();
  await loadContainerCard();
  await loadSystemLoadRadial();
  await loadTopContainersByCpu();
  await loadTopContainersByMemory();
  /*
    // Load trend charts
    await loadTrend("/api/v1/query?metric=system.cpu.usage_percent&scope=total", "trend-cpu", "CPU %");
    await loadTrend("/api/v1/query?metric=system.memory.used_percent", "trend-mem", "Mem %");
    await loadTrend("/api/v1/query?metric=system.memory.used_percent", "trend-disk", "Disk %");
    await loadTrend("/api/v1/query?metric=system.network.bytes_recv&scope=total", "trend-net", "Net In");

    // Load top lists
    await loadTop("/api/v1/query?sort=desc&metric=system.cpu.usage_percent&scope=total&limit=5", "top-cpu-hosts", "Top CPU Hosts", "%");
    await loadTop("/api/v1/query?sort=desc&metric=container.mem.rss&limit=5", "top-mem-containers", "Top Memory Containers", "MB");
    */
  // Load logs/events
  const events = await fetchJson("/api/v1/events?limit=12");
  renderEventLog(events || []);

  // Summary update (alerts count)
  await updateAlertSummary();



}


document.addEventListener("DOMContentLoaded", initHome);
