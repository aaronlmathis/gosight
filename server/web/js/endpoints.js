import { formatBytes, formatUptime } from './format.js';
console.log("✅ endpoints.js is running");

// 🌐 Filter DOM references
const containerStatusFilter = document.getElementById('filter-container-status');
const runtimeFilter = document.getElementById('filter-runtime');
const hostFilter = document.getElementById('filter-host');
const containerTableBody = document.getElementById('container-table-body');

const hostStatusFilter = document.getElementById('filter-host-status');
const lastUpdated = document.getElementById('last-updated');

// 📦 Data storage
let allContainers = [];
let allEndpoints = [];

// 🧠 Timestamp updater
function updateTimestamp() {
  if (lastUpdated) {
    lastUpdated.textContent = new Date().toLocaleTimeString();
  }
}

// 🔁 Fetch containers
async function fetchContainers() {
  try {
    const res = await fetch('/api/endpoints/containers');
    const data = await res.json();

    if (!Array.isArray(data)) {
      console.warn("⚠️ /api/endpoints/containers returned unexpected format:", data);
      return;
    }

    allContainers = data;
    console.log("📦 Container API data:", data);
    updateContainerTable();
    updateTimestamp();
  } catch (err) {
    console.error('❌ Failed to load container data:', err);
  }
}

// 🔁 Fetch host endpoints from /api/endpoints/hosts
async function fetchHosts() {
  try {
    const res = await fetch('/api/endpoints/hosts');
    const data = await res.json();

    if (!Array.isArray(data)) {
      console.warn("⚠️ /api/endpoints/hosts returned unexpected format:", data);
      return;
    }

    allEndpoints = data;
    console.log("🖥️ Host endpoint API data:", data);
    filterAndRenderEndpoints();
  } catch (err) {
    console.error("❌ Failed to load host endpoint data:", err);
  }
}

// 🖨 Render container table
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

// 🖨 Render host endpoints
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

// 🧠 Filter and update endpoints
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

// 🚀 Init
async function initializePage() {
  try {
    await fetchHosts();
    fetchContainers();
    setInterval(fetchContainers, 10000);
    setInterval(fetchHosts, 15000);
  } catch (err) {
    console.error('❌ Failed to initialize page:', err);
  }
}

document.addEventListener('DOMContentLoaded', () => {
  // 🔗 Container filter bindings
  [containerStatusFilter, runtimeFilter, hostFilter].forEach(el => {
    if (el) el.addEventListener('input', updateContainerTable);
  });

  // 🔗 Host filter bindings
  ['filter-hostname', 'filter-ip', 'filter-os', 'filter-arch', 'filter-host-status'].forEach(id => {
    const el = document.getElementById(id);
    if (el) el.addEventListener('input', filterAndRenderEndpoints);
  });

  initializePage();
});

