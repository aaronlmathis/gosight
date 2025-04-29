import { gosightFetch } from "./api.js";

function renderSummaryStats(endpoints, containers) {
  const containerCount = containers.length;
  const runningContainers = containers.filter(c => c.status === "running").length;
  const runtimes = [...new Set(containers.map(c => c.runtime || "unknown"))].join(", ");

  const totalHosts = endpoints.length;
  const onlineHosts = endpoints.filter(ep => ep.status === "Online").length;

  const summaryHTML = `
  <div class="grid grid-cols-2 md:grid-cols-3 gap-4 mb-6">
    <div class="bg-white dark:bg-gray-800 rounded-lg p-4 shadow-sm">
      <div class="text-sm text-gray-500 dark:text-gray-400">Hosts Online</div>
      <div class="text-xl font-bold text-gray-900 dark:text-white">${onlineHosts} / ${totalHosts}</div>
    </div>
    <div class="bg-white dark:bg-gray-800 rounded-lg p-4 shadow-sm">
      <div class="text-sm text-gray-500 dark:text-gray-400">Containers Running</div>
      <div class="text-xl font-bold text-gray-900 dark:text-white">${runningContainers} / ${containerCount}</div>
    </div>
    <div class="bg-white dark:bg-gray-800 rounded-lg p-4 shadow-sm">
      <div class="text-sm text-gray-500 dark:text-gray-400">Runtimes Observed</div>
      <div class="text-xl font-bold text-gray-900 dark:text-white">${runtimes}</div>
    </div>
  </div>`;

  const summaryDiv = document.getElementById("endpoint-summary");
  if (summaryDiv) summaryDiv.innerHTML = summaryHTML;
  document.getElementById("summary-hosts").textContent = `${onlineHosts} / ${totalHosts}`;
  document.getElementById("summary-containers").textContent = `${runningContainers} / ${containerCount}`;
  document.getElementById("summary-runtimes").textContent = runtimes;
}
export async function loadHostTable() {
  const agentsRes = await gosightFetch("/api/v1/endpoints/hosts");
  const agents = await agentsRes.json();

  const tbody = document.getElementById("host-table-body");
  tbody.innerHTML = "";

  for (const agent of agents) {
    const labels = agent.labels || {};
    const endpointID = agent.endpoint_id || "";
    const agentID = agent.agent_id || "";
    const hname = agent.hostname;

    if (endpointID.startsWith("ctr-")) continue;

    const rowID = `host-${agent.agent_id}`;
    const containerRowID = `containers-${agent.agent_id}`;
    const isOnline = agent.status === "Online";

    const hostnameCell = isOnline
      ? `<a href="/endpoints/${endpointID}" class="text-blue-800 dark:text-blue-400 hover:underline">${hname}</a>`
      : `<span class="text-gray-500 dark:text-gray-400">${hname}</span>`;

    const expandCell = isOnline
      ? `<button onclick="toggleContainerRow('${containerRowID}')" class="text-blue-500 hover:text-blue-700">
<svg id="expand-icon-${containerRowID}"
   class="w-4 h-4 inline-block transform -rotate-90 transition-transform origin-center"
   fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
<path stroke-linecap="round" stroke-linejoin="round" d="M6 9l6 6 6-6" />
</svg>

      </button>`
      : `<span class="text-gray-400">—</span>`;

    const platform = `${labels.platform || ""} ${labels.platform_version || ""}`.trim();
    const dataStatus = isOnline ? 'data-status="online"' : 'data-status="offline"';

    const hostRowHTML = `
<tr id="${rowID}" ${dataStatus} class="endpoint-row hover:bg-gray-50 dark:hover:bg-gray-800 host-row">
<td class="px-3 py-2">
  ${isOnline
        ? '<span class="text-green-500 font-medium">● Online</span>'
        : '<span class="text-red-500 font-medium">● Offline</span>'}
</td>
<td class="px-3 py-2 font-medium">${hostnameCell}</td>
<td class="px-3 py-2">${agent.ip}</td>
<td class="px-3 py-2">${agent.os}</td>
<td class="px-3 py-2">${platform}</td>
<td class="px-3 py-2">${agent.arch}</td>
<td class="px-3 py-2">${agent.agent_id}</td>
<td class="px-3 py-2">${agent.version}</td>
<td class="px-3 py-2" id="lastseen-${rowID}">—</td>
<td class="px-3 py-2" id="uptime-${rowID}">—</td>
<td class="px-3 py-2">${expandCell}</td>

</tr>`;

    const containerRowHTML = `
<tr id="${containerRowID}" class="container-subtable">
<td colspan="13" class="p-0">
  <div id="container-wrapper-${containerRowID}"
       class="collapsed overflow-hidden transition-all duration-700 ease-in-out"
       style="max-height: 0; opacity: 0;">
    <div class="p-4 text-sm text-gray-400">Loading containers…</div>
  </div>
</td>
</tr>`;

    tbody.insertAdjacentHTML("beforeend", hostRowHTML);
    tbody.insertAdjacentHTML("beforeend", containerRowHTML);


    //
    // Update last seen / uptime AFTER rows are added
    const lastSeenCell = document.getElementById(`lastseen-${rowID}`);
    const uptimeCell = document.getElementById(`uptime-${rowID}`);

    if (isOnline) {
      if (uptimeCell) uptimeCell.textContent = formatUptime(agent.uptime_seconds);
      if (lastSeenCell) lastSeenCell.textContent = "—";
    } else {
      if (lastSeenCell) lastSeenCell.textContent = formatLastSeen(agent.last_seen);
      if (uptimeCell) uptimeCell.textContent = "—";
    }
  }
  renderSummaryStats(agents, []);
}
function formatLastSeen(isoTime) {
  if (!isoTime) return "—";
  const last = new Date(isoTime).getTime();
  const now = Date.now();
  const diff = Math.floor((now - last) / 1000);


  if (diff < 60) return `${diff}s ago`;
  if (diff < 3600) return `${Math.floor(diff / 60)}m ago`;
  if (diff < 86400) return `${Math.floor(diff / 3600)}h ago`;
  return `${Math.floor(diff / 86400)}d ago`;
}
function formatUptime(seconds) {
  const s = Math.floor(seconds);
  const d = Math.floor(s / 86400);
  const h = Math.floor((s % 86400) / 3600);
  const m = Math.floor((s % 3600) / 60);
  return `${d > 0 ? d + 'd ' : ''}${h}h ${m}m`;
}
export async function fetchGlobalContainerMetrics() {
  const runtimes = ["podman", "docker"];
  const metricNames = ["cpu_percent", "uptime_seconds"];
  const metrics = runtimes.flatMap(rt => metricNames.map(name => `container.${rt}.${name}`));
  const query = metrics.map(m => `metric=${m}`).join("&");

  try {
    const res = await gosightFetch(`/api/v1/query?${query}`);
    console.log(query)
    const rows = await res.json();

    const containers = new Map();
    const runtimeSet = new Set();

    for (const row of rows) {
      const tags = row.tags || {};
      const id = tags.container_id;
      const metricName = tags["__name__"];

      if (!id || !metricName) continue;

      const parts = metricName.split(".");
      if (parts.length >= 2) runtimeSet.add(parts[1]);

      if (!containers.has(id)) {
        containers.set(id, { id, status: tags.status || "unknown" });
      }
    }

    const containerCount = containers.size;
    const runningCount = [...containers.values()].filter(c => c.status === "running").length;
    const runtimes = Array.from(runtimeSet).join(", ");

    document.getElementById("summary-containers").textContent = `${runningCount} / ${containerCount}`;
    document.getElementById("summary-runtimes").textContent = runtimes;

  } catch (err) {
    console.error("❌ Failed to load global container metrics:", err);
  }
}



function toggleContainerRow(rowID) {
  const wrapper = document.getElementById(`container-wrapper-${rowID}`);
  const expandIcon = document.getElementById(`expand-icon-${rowID}`);
  const agentID = rowID.replace(/^containers-/, "");
  const hostRow = document.getElementById(`host-${agentID}`);
  if (!wrapper || !hostRow) return;

  const hostnameCell = hostRow.querySelector("td:nth-child(2)");
  const hostname = hostnameCell?.textContent?.trim() || "";

  const isExpanded = wrapper.classList.contains("expanded");

  if (isExpanded) {
    wrapper.classList.remove("expanded");
    wrapper.classList.add("collapsed");
    wrapper.style.maxHeight = "0px";
    wrapper.style.opacity = "0";
    if (expandIcon) expandIcon.classList.remove("rotate-90");
    return;
  }

  wrapper.classList.remove("collapsed");
  wrapper.classList.add("expanded");
  wrapper.style.opacity = "1";
  wrapper.style.maxHeight = "1000px"; // enough for a full table
  if (expandIcon) expandIcon.classList.add("rotate-90");

  if (!wrapper.dataset.loaded) {
    wrapper.innerHTML = `<div class="p-4 text-sm text-gray-400">Loading...</div>`;
    const runtimes = ["podman", "docker"];
    const metricNames = [
      "cpu_percent",
      "mem_usage_bytes",
      "net_rx_bytes",
      "net_tx_bytes",
      "uptime_seconds"
    ];
    const metrics = runtimes.flatMap(rt =>
      metricNames.map(name => `container.${rt}.${name}`)
    );
    const query = metrics.map(m => `metric=${m}`).join("&") + `&hostname=${encodeURIComponent(hostname)}`;
    console.log("Container query:", query);
    gosightFetch(`/api/v1/query?${query}`)
      .then(res => res.json())
      .then(json => {
        const rows = Array.isArray(json) ? json : [];
        console.log("📊 Raw container rows from API:", rows);
        const grouped = groupContainers(rows);
        wrapper.innerHTML = `<div class="p-4">${buildContainerTable(grouped)}</div>`;
        wrapper.dataset.loaded = "true";
      })
      .catch(err => {
        console.error("Failed to fetch containers:", err);
        wrapper.innerHTML = `<div class="p-4 text-sm text-red-500">Container fetch failed: ${err}</div>`;
      });
  }
}
window.toggleContainerRow = toggleContainerRow;
function heartbeatBadge(heartbeat) {
  if (heartbeat === "Online")
    return `<span class="px-2 py-0.5 rounded-sm text-xs font-bold bg-green-100 text-green-800 dark:bg-green-800 dark:text-green-100">${heartbeat}</span>`;
  if (heartbeat === "Idle")
    return `<span class="px-2 py-0.5 rounded-sm text-xs font-bold bg-yellow-100 text-yellow-800 dark:bg-yellow-800 dark:text-yellow-100">${heartbeat}</span>`;
  return `<span class="px-2 py-0.5 rounded-sm text-xs font-bold bg-red-100 text-red-800 dark:bg-red-800 dark:text-red-100">${heartbeat}</span>`;
}

function groupContainers(rows) {
  if (!Array.isArray(rows)) return [];
  const map = {};

  rows.forEach(row => {
    const tags = row.tags || {};
    const id = tags.container_id;
    const metricName = tags["__name__"] || "";
    const parts = metricName.split(".");
    const runtime = parts.length >= 2 ? parts[1] : "unknown";

    if (!map[id]) {
      map[id] = {
        name: tags.container_name || "—",
        image: tags.image || "—",
        status: tags.status || "unknown",
        heartbeat: "Online",
        cpu: "—",
        mem: "—",
        rx: "—",
        tx: "—",
        uptime: "—",
        runtime: runtime
      };
    }

    console.log("Checking row with container_id:", id, "name:", tags["__name__"]);
    console.log("🧪 Row received:", row);
    console.log("📛 container_id:", row.tags?.container_id);
    console.log("🧩 __name__:", row.tags?.__name__);

    if (!id) {
      console.warn("Skipping row: missing container_id", row);
      return;
    }

    if (!map[id]) {
      map[id] = {
        name: tags.container_name || "—",
        image: tags.image || "—",
        status: tags.status || "unknown",
        heartbeat: "Online",
        cpu: "—",
        mem: "—",
        rx: "—",
        tx: "—",
        uptime: "—",
      };
    }


    const value = row.value;

    const shortName = metricName.split(".").pop();

    switch (shortName) {
      case "cpu_percent":
        map[id].cpu = `${value.toFixed(1)}%`;
        break;
      case "mem_usage_bytes":
        map[id].mem = formatBytes(value);
        break;
      case "net_rx_bytes":
        map[id].rx = formatBytes(value);
        break;
      case "net_tx_bytes":
        map[id].tx = formatBytes(value);
        break;
      case "uptime_seconds":
        map[id].uptime = formatUptime(value);
        break;
    }


  });

  return Object.values(map);
}

function buildContainerTable(containers) {
  if (!containers.length)
    return `<p class="text-sm italic text-gray-400">No containers found.</p>`;

  const rows = containers.map((c, i) => `
<tr class="${i % 2 === 0 ? 'bg-white dark:bg-gray-800' : 'bg-gray-50 dark:bg-gray-700'} hover:bg-gray-50 dark:hover:bg-gray-600 transition-colors">
<td class="px-4 py-2">${statusBadge(c.status)}</td>
<td class="px-4 py-2">${heartbeatBadge(c.heartbeat)}</td>
<td class="px-4 py-2 text-left">
  <span class="inline-block px-2 py-0.5 rounded-sm text-xs bg-gray-200 text-gray-800 dark:bg-gray-700 dark:text-gray-300">
    ${c.runtime}
  </span>
</td>
<td class="px-4 py-2">${c.name}</td>

<td class="px-4 py-2 text-gray-700 dark:text-gray-300">${c.image}</td>
<td class="px-4 py-2 text-right">${c.cpu}</td>
<td class="px-4 py-2 text-right">${c.mem}</td>
<td class="px-4 py-2 text-right">${c.rx}</td>
<td class="px-4 py-2 text-right">${c.tx}</td>
<td class="px-4 py-2 text-right">${c.uptime}</td>
</tr>
`).join("");

  return `
<div class="overflow-x-auto border border-gray-100 dark:border-gray-700 rounded-lg shadow-sm">
<table class="min-w-full text-sm text-left">
  <thead class="text-xs text-gray-400 bg-gray-100 dark:bg-gray-700 dark:text-gray-300">
    <tr>
      <th class="px-4 py-2">Status</th>
      <th class="px-4 py-2">Heartbeat</th> 
            <th class="px-4 py-2 text-left">Runtime</th>
      <th class="px-4 py-2">Name</th>

      <th class="px-4 py-2">Image</th>
      <th class="px-4 py-2 text-right">CPU %</th>
      <th class="px-4 py-2 text-right">Mem</th>
      <th class="px-4 py-2 text-right">RX</th>
      <th class="px-4 py-2 text-right">TX</th>
      <th class="px-4 py-2 text-right">Uptime</th>
    </tr>
  </thead>
  <tbody>${rows}</tbody>
</table>
</div>
`;
}

function formatBytes(bytes) {
  if (!bytes || isNaN(bytes)) return "—";
  const units = ["B", "KB", "MB", "GB"];
  let i = 0;
  while (bytes >= 1024 && i < units.length - 1) {
    bytes /= 1024;
    i++;
  }
  return `${bytes.toFixed(1)} ${units[i]}`;
}

function statusBadge(status) {
  if (status === "running")
    return `<span class="px-2 py-0.5 rounded-sm text-xs font-bold bg-green-100 text-green-800 dark:bg-green-800 dark:text-green-100">running</span>`;
  return `<span class="px-2 py-0.5 rounded-sm text-xs font-bold bg-red-100 text-red-800 dark:bg-red-800 dark:text-red-100">stopped</span>`;
}


document.addEventListener("DOMContentLoaded", async () => {
  await loadHostTable();
  await fetchGlobalContainerMetrics();
  const searchInput = document.getElementById("filter-by");
  const statusSelect = document.getElementById("filter-status");

  function normalize(text) {
    return text.toLowerCase().trim();
  }

  function filterTable() {
    const query = normalize(searchInput.value);
    const status = normalize(statusSelect.value);

    // ⬇Needs to be here to reflect dynamically inserted rows
    const rows = document.querySelectorAll(".endpoint-row");

    rows.forEach(row => {
      const tds = row.querySelectorAll("td");
      const statusText = row.dataset.status?.toLowerCase();

      let searchableText = "";
      for (let i = 0; i < tds.length - 2; i++) {
        searchableText += " " + tds[i].textContent.toLowerCase();
      }

      const matchesText = searchableText.includes(query);
      const matchesStatus = !status || statusText === status;
      const shouldShow = matchesText && matchesStatus;

      row.style.display = shouldShow ? "" : "none";

      // ⬇Also collapse container row if hiding parent
      const containersRow = document.getElementById(`containers-${row.id.replace(/^host-/, "")}`);
      console.log(`containers-${row.id.replace(/^host-/, "")}`);
      if (containersRow) {
        containersRow.style.display = shouldShow ? "" : "none";
      }
    });
  }

  searchInput.addEventListener("input", filterTable);
  statusSelect.addEventListener("change", filterTable);

  // Optional: run once initially after loadHostTable fills content
  setTimeout(filterTable, 500); // or call it directly after loadHostTable() if desired
});
/*
import { formatBytes, formatUptime } from './format.js';
console.log("endpoints.js is running");

// Filter DOM references
const containerStatusFilter = document.getElementById('filter-container-status');
const runtimeFilter = document.getElementById('filter-runtime');
const hostFilter = document.getElementById('filter-host');
const containerTableBody = document.getElementById('container-table-body');

const hostStatusFilter = document.getElementById('filter-host-status');
const lastUpdated = document.getElementById('last-updated');

//  Data storage
let allContainers = [];
let allEndpoints = [];

// Timestamp updater
function updateTimestamp() {
  if (lastUpdated) {
    lastUpdated.textContent = new Date().toLocaleTimeString();
  }
}

// Fetch containers
async function fetchContainers() {
  try {
    const res = await fetch('/api/endpoints/containers');
    const data = await res.json();

    if (!Array.isArray(data)) {
      console.warn(" /api/endpoints/containers returned unexpected format:", data);
      return;
    }

    allContainers = data;
    console.log(" Container API data:", data);
    updateContainerTable();
    updateTimestamp();
  } catch (err) {
    console.error(' Failed to load container data:', err);
  }
}

//  Fetch host endpoints from /api/endpoints/hosts
async function fetchHosts() {
  try {
    const res = await fetch('/api/endpoints/hosts');
    const data = await res.json();

    if (!Array.isArray(data)) {
      console.warn(" /api/endpoints/hosts returned unexpected format:", data);
      return;
    }

    allEndpoints = data;
    console.log(" Host endpoint API data:", data);
    filterAndRenderEndpoints();
  } catch (err) {
    console.error(" Failed to load host endpoint data:", err);
  }
}

//  Render container table
function updateContainerTable() {
  if (!containerTableBody) return;

  const statusVal = containerStatusFilter?.value;
  const runtimeVal = runtimeFilter?.value;
  const hostVal = hostFilter?.value?.trim().toLowerCase() ?? '';

  const filtered = allContainers.filter(c =>
    (!statusVal || c.status === statusVal) &&
    (!runtimeVal || c.subnamespace === runtimeVal) &&
    (!hostVal || c.host?.toLowerCase().includes(hostVal))
  );

  containerTableBody.innerHTML = filtered.map(container => {
    const cpu = container.cpu?.toFixed(1) + '%' || '0.0%';
    const rx = container.rx ? formatBytes(container.rx) : '—';
    const tx = container.tx ? formatBytes(container.tx) : '—';
    const mem = container.mem ? formatBytes(container.mem) : '—';
    const uptime = formatUptime(container.uptime);
    const tooltip = container.started_at ? new Date(container.started_at).toLocaleString() : '';

    return `
      <tr>
              <td class="px-4 py-2">${container.name ?? '—'}</td>
        <td class="px-4 py-2">${container.host ?? '—'}</td>

        <td class="px-4 py-2">${container.image ?? '—'}</td>
        <td class="px-4 py-2">
          <span class="inline-block px-3 py-1 text-xs font-bold rounded-full 
            ${container.status === 'running'
              ? 'bg-green-100 text-green-800 dark:bg-green-800 dark:text-green-100'
              : 'bg-red-100 text-red-800 dark:bg-red-800 dark:text-red-100'}">
            ${container.status ?? '—'}
          </span>
        </td>
        <td class="px-4 py-2">${cpu}</td>
        <td class="px-4 py-2">${mem}</td>
        <td class="px-4 py-2">${rx}</td>
        <td class="px-4 py-2">${tx}</td>
        <td class="px-4 py-2" title="${tooltip}">${uptime}</td>
      </tr>`;
  }).join('');
}

//  Render host endpoints
function renderHostEndpoints(endpoints) {
  const tbody = document.getElementById('endpoint-table-body');
  if (!tbody) return;

  tbody.innerHTML = '';
  endpoints.forEach(ep => {
    const row = document.createElement('tr');
    row.innerHTML = `
      <td class="px-4 py-2">${ep.hostname ?? '—'}</td>
      <td class="px-4 py-2">${ep.ip ?? '—'}</td>
      <td class="px-4 py-2">${ep.os ?? '—'}</td>
      <td class="px-4 py-2">${ep.arch ?? '—'}</td>
      <td class="px-4 py-2">
        <span class="inline-block px-3 py-1 text-xs font-bold rounded-full 
          ${ep.status === 'online'
            ? 'bg-green-100 text-green-800 dark:bg-green-800 dark:text-green-100'
            : 'bg-red-100 text-red-800 dark:bg-red-800 dark:text-red-100'}">
          ${ep.status ?? 'unknown'}
        </span>
      </td>`;
    tbody.appendChild(row);
  });

  const countLabel = document.getElementById('endpoint-count');
  if (countLabel) {
    countLabel.textContent = `${endpoints.length} total`;
  }
}

//  Filter and update endpoints
function filterAndRenderEndpoints() {
  const hostname = document.getElementById('filter-hostname')?.value.toLowerCase() ?? '';
  const ip = document.getElementById('filter-ip')?.value.toLowerCase() ?? '';
  const os = document.getElementById('filter-os')?.value.toLowerCase() ?? '';
  const arch = document.getElementById('filter-arch')?.value.toLowerCase() ?? '';
  const status = hostStatusFilter?.value ?? '';

  const filtered = allEndpoints.filter(ep =>
    ep &&
    (ep.hostname?.toLowerCase().includes(hostname)) &&
    (ep.ip?.toLowerCase().includes(ip)) &&
    (ep.os?.toLowerCase().includes(os)) &&
    (ep.arch?.toLowerCase().includes(arch)) &&
    (status === '' || ep.status === status)
  );

  renderHostEndpoints(filtered);
}

//  Init
async function initializePage() {
  try {
    await fetchHosts();
    fetchContainers();
    setInterval(fetchContainers, 10000);
    setInterval(fetchHosts, 15000);
  } catch (err) {
    console.error(' Failed to initialize page:', err);
  }
}

document.addEventListener('DOMContentLoaded', () => {
  //  Container filter bindings
  [containerStatusFilter, runtimeFilter, hostFilter].forEach(el => {
    if (el) el.addEventListener('input', updateContainerTable);
  });

  //  Host filter bindings
  ['filter-hostname', 'filter-ip', 'filter-os', 'filter-arch', 'filter-host-status'].forEach(id => {
    const el = document.getElementById(id);
    if (el) el.addEventListener('input', filterAndRenderEndpoints);
  });

  initializePage();
});

*/

