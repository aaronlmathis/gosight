import { formatBytes, formatUptime } from "./format.js";
import { registerTabInitializer } from "./tabs.js";

// disk.js
let diskMetricsBuffer = [];
let diskTabInitialized = false;
let selectedDevice = null;

const deviceMetricCache = {};  // { [device]: { timestamps: [], readCount: [], writeCount: [], readBytes: [], writeBytes: [] } }
const MAX_POINTS = 30;         // Keep the charts to last 30 intervals

//
// Disk Summary Donut Chart
// ------------------------------------------------
// This chart shows the disk usage as a donut chart
// with two segments: used and free space.

let diskUsageDonutChart = null;
function createDiskUsageDonut() {
    const ctx = document.getElementById("diskUsageDonutChart").getContext("2d");
    diskUsageDonutChart = new Chart(ctx, {
        type: "doughnut",
        data: {
            labels: ["Used", "Free"],
            datasets: [{
                label: "Disk Usage",
                data: [0, 0],
                backgroundColor: ["#3b82f6", "#10b981"],
                borderWidth: 1,
            }],
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            cutout: "65%",
            plugins: {
                legend: {
                    position: "bottom",
                    labels: {
                        color: "#6B7280", // gray-500
                        boxWidth: 14,
                        font: { size: 12, weight: "500" },
                    },
                },
                tooltip: {
                    callbacks: {
                        label: (ctx) => {
                            const value = ctx.parsed ?? 0;
                            const pct = ctx.percentage ?? 0;
                            return `${ctx.label}: ${formatBytes(value)} (${pct.toFixed(1)}%)`;
                        },
                    },
                },
            },
        },
    });
}

function renderDiskDonut(usageByMount) {
    let total = 0;
    let used = 0;

    for (const mp in usageByMount) {
        const m = usageByMount[mp];
        total += m.total || 0;
        used += m.used || 0;
    }

    const free = total - used;
    const percent = total > 0 ? (used / total) * 100 : 0;

    // Update stat cards
    document.getElementById("disk-total").textContent = formatBytes(total);
    document.getElementById("disk-used").textContent = formatBytes(used);
    document.getElementById("disk-percent").textContent = percent.toFixed(1);
    document.getElementById("disk-free").textContent = formatBytes(free);

    // Update donut chart
    if (diskUsageDonutChart) {
        diskUsageDonutChart.data.datasets[0].data = [used, free];
        diskUsageDonutChart.update();
    }
}

//
// -------------------------------------------------
// END Disk Summary Donut Chart
// -------------------------------------------------

//
// ------------------------------------------------
// Disk Inode Bar Chart
// This chart shows the inode usage as a bar chart
// -------------------------------------------------
let inodeUsageChart = null;

function createInodeUsageBarChart() {
    const ctx = document.getElementById("inodeUsageBarChart").getContext("2d");

    inodeUsageChart = new Chart(ctx, {
        type: "bar",
        data: {
            labels: [],
            datasets: [{
                label: "Inodes Used %",
                data: [],
                backgroundColor: "#6366f1", // indigo
                borderRadius: 4,
            }]
        },
        options: {
            indexAxis: "y",
            responsive: true,
            maintainAspectRatio: false,
            scales: {
                x: {
                    beginAtZero: true,
                    max: 100,
                    ticks: {
                        callback: (v) => `${v}%`,
                        color: "#6B7280", // gray-500
                    }
                },
                y: {
                    ticks: {
                        color: "#6B7280",
                    }
                }
            },
            plugins: {
                legend: { display: false },
                tooltip: {
                    callbacks: {
                        label: (ctx) => `${ctx.label}: ${ctx.parsed.x.toFixed(1)}%`
                    }
                }
            }
        }
    });
}

// Render Inode Bar Chart
function renderInodeBar(usageByMount) {
    const labels = [];
    const values = [];

    for (const mount in usageByMount) {
        const usage = usageByMount[mount];
        if (usage.inodes_used_percent !== undefined) {
            labels.push(mount);
            values.push(usage.inodes_used_percent);
        }
    }

    if (inodeUsageChart) {
        inodeUsageChart.data.labels = labels;
        inodeUsageChart.data.datasets[0].data = values;
        inodeUsageChart.update();
    }
}

//
// -------------------------------------------------
// END Disk Inode Bar Chart
// -------------------------------------------------
//

//
// ------------------------------------------------
// Mountpoint Table
// This table shows the disk usage by mountpoint
// -------------------------------------------------

function renderMountpointTable(usageByMount) {
    const tbody = document.getElementById("mountpoint-table-body");
    if (!tbody) return;

    tbody.innerHTML = ""; // Clear old rows

    for (const mp in usageByMount) {
        const usage = usageByMount[mp];

        const row = document.createElement("tr");
        row.innerHTML = `
          <td class="px-4 py-2 font-medium text-blue-500">${mp}</td>
          <td class="px-4 py-2">${usage.fstype || "—"}</td>
          <td class="px-4 py-2">${formatBytes(usage.total)}</td>
          <td class="px-4 py-2">${formatBytes(usage.used)}</td>
          <td class="px-4 py-2">${formatBytes(usage.free)}</td>
          <td class="px-4 py-2">${(usage.used_percent || 0).toFixed(1)}%</td>
          <td class="px-4 py-2">${usage.device || "—"}</td>
        `;
        tbody.appendChild(row);
    }
}

//
// -------------------------------------------------
// END Mountpoint Table
// -------------------------------------------------



//
// -------------------------------------------------
// Disk IO Charts
// -------------------------------------------------
let diskIopsChart = null;
let diskThroughputChart = null;


function createDiskIOCharts() {
    // IOPS Chart
    const iopsCtx = document.getElementById("diskIopsChart").getContext("2d");
    diskIopsChart = new Chart(iopsCtx, {
        type: "line",
        data: {
            labels: [], // Time buckets (e.g. "1m", "2m")
            datasets: [
                {
                    label: "Read Count",
                    data: [],
                    borderColor: "#3b82f6",
                    backgroundColor: "rgba(59, 130, 246, 0.1)",
                    tension: 0.3,
                    fill: true,
                    pointRadius: 0,
                },
                {
                    label: "Write Count",
                    data: [],
                    borderColor: "#10b981",
                    backgroundColor: "rgba(16, 185, 129, 0.1)",
                    tension: 0.3,
                    fill: true,
                    pointRadius: 0,
                }
            ]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            scales: {
                y: {
                    beginAtZero: true,
                    title: { display: true, text: "Ops/sec" },
                }
            },
            plugins: {
                legend: {
                    labels: { color: "#6B7280" }
                }
            }
        }
    });

    // Throughput Chart
    const tpCtx = document.getElementById("diskThroughputChart").getContext("2d");
    diskThroughputChart = new Chart(tpCtx, {
        type: "line",
        data: {
            labels: [],
            datasets: [
                {
                    label: "Read Bytes",
                    data: [],
                    borderColor: "#f59e0b",
                    backgroundColor: "rgba(245, 158, 11, 0.1)",
                    tension: 0.3,
                    fill: true,
                    pointRadius: 0,
                },
                {
                    label: "Write Bytes",
                    data: [],
                    borderColor: "#ef4444",
                    backgroundColor: "rgba(239, 68, 68, 0.1)",
                    tension: 0.3,
                    fill: true,
                    pointRadius: 0,
                }
            ]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            scales: {
                y: {
                    beginAtZero: true,
                    title: { display: true, text: "Bytes/sec" },
                    ticks: {
                        callback: (v) => formatBytes(v)
                    }
                }
            },
            plugins: {
                legend: {
                    labels: { color: "#6B7280" }
                }
            }
        }
    });
}

//
// ------------------------------------------------
// END Disk IO Charts
// -------------------------------------------------

//
// ------------------------------------------------
// Render IODropdown
// -------------------------------------------------
function renderIODropdown(ioByDevice, usageByMount) {
    const dropdown = document.getElementById("disk-device-select");
    if (!dropdown) return;

    const current = selectedDevice;
    const keys = Object.keys(ioByDevice).sort();
    console.log("🔧 Matching device-to-mount:", usageByMount);
    console.log("🔧 Devices:", keys);
    dropdown.innerHTML = keys.map(dev => {
        // 🔍 Match mountpoint for this device
        const matchEntry = Object.entries(usageByMount).find(([_, usage]) => usage.device === dev);
        const label = matchEntry
            ? `${dev} (${matchEntry[0]})`
            : dev;

        return `<option value="${dev}" ${dev === current ? "selected" : ""}>${label}</option>`;
    }).join("");

    // Select first if not already
    if (!selectedDevice && keys.length > 0) {
        selectedDevice = keys[0];
    }

    // Store the metrics
    for (const dev of keys) {
        updateDeviceMetricCache(dev, ioByDevice[dev]);
    }

    updateDiskIOChartsFromCache();
}
//
// ------------------------------------------------
// END Render IODropdown
// -------------------------------------------------
//

//
// ------------------------------------------------
// Mini Mount Tables
// -------------------------------------------------

function renderTopMountUsage(usageByMount) {
    const entries = Object.entries(usageByMount)
        .filter(([_, u]) => typeof u.used_percent === "number")
        .sort((a, b) => b[1].used_percent - a[1].used_percent)
        .slice(0, 3);

    const html = entries.map(([mp, u]) =>
        `<li><span class="font-semibold text-blue-500">${mp}</span> → ${u.used_percent.toFixed(1)}%</li>`
    ).join("");

    const el = document.getElementById("top-mount-usage");
    if (el) el.innerHTML = html;
}

function renderMiniMountTable(usageByMount) {
    const sorted = Object.entries(usageByMount)
        .filter(([_, u]) => typeof u.total === "number" && typeof u.used_percent === "number")
        .sort((a, b) => b[1].used_percent - a[1].used_percent);

    const html = sorted.map(([mp, u]) =>
        `<div><span class="text-blue-600 font-semibold">${mp}</span></div>
       <div>${formatBytes(u.used)} / ${formatBytes(u.total)} (${u.used_percent.toFixed(1)}%)</div>`
    ).join("");

    const el = document.getElementById("disk-mini-mounts");
    if (el) el.innerHTML = html;
}
//
// ------------------------------------------------
// Update Device Metric Cache
// -------------------------------------------------
function updateDeviceMetricCache(device, metrics) {
    if (!deviceMetricCache[device]) {
        deviceMetricCache[device] = {
            timestamps: [],
            readCount: [],
            writeCount: [],
            readBytes: [],
            writeBytes: [],
        };
    }

    const cache = deviceMetricCache[device];
    const ts = new Date().toLocaleTimeString([], { hour12: false, hour: "2-digit", minute: "2-digit", second: "2-digit" });

    cache.timestamps.push(ts);
    cache.readCount.push(metrics.read_count || 0);
    cache.writeCount.push(metrics.write_count || 0);
    cache.readBytes.push(metrics.read_bytes || 0);
    cache.writeBytes.push(metrics.write_bytes || 0);

    // Trim to MAX_POINTS
    if (cache.timestamps.length > MAX_POINTS) {
        Object.keys(cache).forEach(k => cache[k].shift());
    }
}

function updateDiskIOChartsFromCache() {
    if (!selectedDevice || !deviceMetricCache[selectedDevice]) return;

    const cache = deviceMetricCache[selectedDevice];

    // IOPS
    if (diskIopsChart) {
        diskIopsChart.data.labels = cache.timestamps;
        diskIopsChart.data.datasets[0].data = cache.readCount;
        diskIopsChart.data.datasets[1].data = cache.writeCount;
        diskIopsChart.update();
    }

    // Throughput
    if (diskThroughputChart) {
        diskThroughputChart.data.labels = cache.timestamps;
        diskThroughputChart.data.datasets[0].data = cache.readBytes;
        diskThroughputChart.data.datasets[1].data = cache.writeBytes;
        diskThroughputChart.update();
    }
}

function resetDiskIOCharts() {
    if (diskIopsChart) {
        diskIopsChart.data.labels = [];
        diskIopsChart.data.datasets.forEach(ds => ds.data = []);
        diskIopsChart.update();
    }
    if (diskThroughputChart) {
        diskThroughputChart.data.labels = [];
        diskThroughputChart.data.datasets.forEach(ds => ds.data = []);
        diskThroughputChart.update();
    }
}

function renderMiniMountStats(usageByMount) {
    const container = document.getElementById("disk-mini-mounts");
    if (!container) return;

    container.innerHTML = "";

    for (const [mount, usage] of Object.entries(usageByMount)) {
        const total = formatBytes(usage.total);
        const used = formatBytes(usage.used);
        const percent = usage.used_percent?.toFixed(1) ?? "—";

        const row = document.createElement("div");
        row.className = "grid grid-cols-2 gap-x-4 px-2 py-2 bg-gray-50 dark:bg-gray-800";

        row.innerHTML = `
        <dt class="font-medium text-gray-500 dark:text-gray-400">${mount}</dt>
        <dd class="text-right text-sm text-gray-700 dark:text-gray-300">
          <div class="text-xs mb-1">${used} / ${total}</div>
          <div class="w-full h-2 bg-gray-300 dark:bg-gray-700 rounded">
            <div class="h-2 bg-emerald-500 rounded" style="width: ${percent}%;"></div>
          </div>
        </dd>
      `;

        container.appendChild(row);
    }
}
function renderTopMountsBar(usageByMount) {
    const topMounts = Object.entries(usageByMount)
        .filter(([_, usage]) => typeof usage.used_percent === "number")
        .sort((a, b) => b[1].used_percent - a[1].used_percent)
        .slice(0, 3);

    const container = document.getElementById("top-mount-usage");
    if (!container) return;

    container.innerHTML = "";

    for (const [mount, usage] of topMounts) {
        const percent = usage.used_percent.toFixed(1);

        const row = document.createElement("div");
        row.className = "grid grid-cols-2 gap-x-4 px-2 py-2";

        row.innerHTML = `
        <dt class="font-medium text-blue-600 dark:text-blue-400">${mount}</dt>
        <dd>
          <div class="flex justify-between text-xs mb-1">
            <span>${percent}%</span>
          </div>
          <div class="w-full h-2 bg-gray-200 dark:bg-gray-700 rounded">
            <div class="h-2 bg-blue-500 rounded" style="width: ${percent}%"></div>
          </div>
        </dd>
      `;

        container.appendChild(row);
    }
}
// ------------------------------------------------
// INIT EVERYTHING
// -------------------------------------------------

function processDiskMetrics(metrics) {
    const usageByMount = {};
    const ioByDevice = {};

    for (const m of metrics) {
        const dims = m.dimensions || {};
        const mp = dims.mountpoint;
        const dev = dims.device;

        if (m.subnamespace === "Disk") {
            if (!mp) continue;
            if (!usageByMount[mp]) usageByMount[mp] = {};
            usageByMount[mp][m.name] = m.value;
            if (dims.device) {
                usageByMount[mp].device = dims.device.replace("/dev/", "");
            }

            if (dims.fstype) usageByMount[mp].fstype = dims.fstype;
        }

        if (m.subnamespace === "DiskIO") {
            if (!dev) continue;
            if (!ioByDevice[dev]) ioByDevice[dev] = {};
            ioByDevice[dev][m.name] = m.value;
        }
    }
    console.log("📦 Mountpoints received:");
    for (const mp in usageByMount) {
        console.log(`  • ${mp}:`, usageByMount[mp]);
    }
    renderMountpointTable(usageByMount);
    renderInodeBar(usageByMount);
    renderDiskDonut(usageByMount);
    renderIODropdown(ioByDevice, usageByMount);
    renderTopMountUsage(usageByMount);
    renderMiniMountTable(usageByMount);
    renderTopMountsBar(usageByMount);
    renderMiniMountStats(usageByMount);
}

// INIT Disk TAB
function initDiskTab() {
    createDiskUsageDonut();
    createInodeUsageBarChart();
    createDiskIOCharts();

    const dropdown = document.getElementById("disk-device-select");
    if (dropdown && !dropdown._bound) {
        dropdown.addEventListener("change", (e) => {
            selectedDevice = e.target.value;
            resetDiskIOCharts();
            updateDiskIOChartsFromCache();
        });
        dropdown._bound = true;
        console.log("✅ Bound disk device switcher");
    }

    processDiskMetrics(diskMetricsBuffer);
    diskMetricsBuffer = [];
    diskTabInitialized = true;
}

window.diskMetricHandler = function (metrics) {
    for (const metric of metrics) {
        if (!metric || metric.namespace?.toLowerCase() !== "system") continue;

        const sub = metric.subnamespace?.toLowerCase();
        if (!["disk", "diskio"].includes(sub)) continue;

        diskMetricsBuffer.push(metric);
    }

    if (diskTabInitialized && diskMetricsBuffer.length > 0) {
        processDiskMetrics(diskMetricsBuffer);
        diskMetricsBuffer = [];
    }
};

registerTabInitializer("disk", initDiskTab);