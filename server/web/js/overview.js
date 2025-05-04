import { formatBytes, formatUptime } from "./format.js";
import { registerTabInitializer } from "./tabs.js";
import { gosightFetch } from "./api.js";

const miniCharts = {
    cpu: null,
    memory: null,
    swap: null,
};

let latestCpuPercent = 0;
let latestSwapUsedPercent = 0;
let latestMemUsedPercent = 0;

//
// Mini Charts
//

async function renderMiniCharts() {
    const cpuChart = createMiniApexChart("miniCpuChart", "#3b82f6", "CPU Usage %");
    const memoryChart = createMiniApexChart("miniMemoryChart", "#10b981", "Memory Usage %");
    const swapChart = createMiniApexChart("miniSwapChart", "#f87171", "Swap Usage %");

    miniCharts.cpu = cpuChart;
    miniCharts.memory = memoryChart;
    miniCharts.swap = swapChart;

    await Promise.all([
        cpuChart.render(),
        memoryChart.render(),
        swapChart.render()
    ]);
}

function createMiniApexChart(elementId, color, label) {
    const options = {
        chartLabel: label,
        dataLabels: {
            enabled: false,
        },

        noData: {
            text: "",
        },
        chart: {
            type: "area",
            height: 280,
            zoom: { enabled: false },
            toolbar: { show: false },
            animations: { enabled: true, easing: "easeinout", speed: 500 },
        },
        series: [{ name: label, data: [] }],
        stroke: { show: true, curve: "smooth", width: 2 },
        fill: {
            type: "gradient",
            gradient: { shadeIntensity: 1, opacityFrom: 0.4, opacityTo: 0, stops: [0, 90, 100] }
        },
        colors: [color],
        xaxis: {
            type: "datetime",
            labels: { format: "HH:mm:ss" },
            axisBorder: { show: false },
            axisTicks: { show: false },
        },
        yaxis: {
            labels: { formatter: val => `${val.toFixed(1)}%` }
        },
        tooltip: {
            x: { format: "HH:mm:ss" },
            custom: generateProcessTooltip(label.toLowerCase().includes("memory")),
            cssClass: "custom-process-tooltip"     // key to override styles
        },
        grid: { borderColor: "#e0e0e0", strokeDashArray: 4 }
    };

    const chart = new ApexCharts(document.getElementById(elementId), options);
    return chart;
}

function generateProcessTooltip(isMem) {
    return function ({ series, seriesIndex, dataPointIndex, w }) {
        const labelKey = isMem ? "mem_percent" : "cpu_percent";
        const point = w.config.series[seriesIndex].data[dataPointIndex];
        if (!point || !point.x) return "";

        const hoverTime = point.x;
        const snapshot = findClosestSnapshot(hoverTime);
        if (!snapshot || !snapshot.processes) return "No process data";

        const rows = snapshot.processes
            .sort((a, b) => b[labelKey] - a[labelKey])
            .slice(0, 5)
            .map(p => {
                const full = p.cmdline || p.executable || "(?)";
                const short = truncateCmd(full, 50);
                const value = p[labelKey].toFixed(1);
                return `
            <tr style="border-bottom: 1px solid #e5e7eb;">
              <td style="padding:4px 6px; font-size:11px; color:#6b7280;">${p.pid}</td>
              <td title="${full}" style="max-width:180px; overflow:hidden; text-overflow:ellipsis; white-space:nowrap; padding:4px 6px; font-size:11px;">
                ${short}
              </td>
              <td style="text-align:right; padding:4px 6px; font-weight:500; font-size:11px; color:${isMem ? '#10b981' : '#3b82f6'};">${value}%</td>
            </tr>`;
            }).join("");

        return `
        <div>
          <div style="font-weight: 600; font-size: 12px; margin-bottom: 6px;">
            Top Processes (${isMem ? "Memory" : "CPU"})
          </div>
          <table style="width:100%; border-collapse:collapse;">
            <thead>
              <tr style="text-align:left; font-size: 11px; color:#9ca3af;">
                <th style="padding: 4px 6px;">PID</th>
                <th style="padding: 4px 6px;">Command</th>
                <th style="padding: 4px 6px; text-align:right;">Usage</th>
              </tr>
            </thead>
            <tbody>
              ${rows}
            </tbody>
          </table>
        </div>`;
    };
}
function truncateCmd(cmd, max = 40) {
    if (!cmd) return "(?)";
    return cmd.length > max ? cmd.slice(0, max - 1) + "…" : cmd;
}
function findClosestSnapshot(ts) {
    let closest = null;
    let minDiff = Infinity;
    for (const snap of processHistory) {
        const diff = Math.abs(snap.timestamp - ts);
        if (diff < minDiff) {
            closest = snap;
            minDiff = diff;
        }
    }
    return closest;
}

function labelToKey(label) {
    return label.toLowerCase().includes("memory") ? "mem_percent" : "cpu_percent";
}
function appendMiniPoint(chart, timestamp, val) {
    if (!chart || !chart.w || !chart.w.config) return;

    const series = chart.w.config.series[0].data;
    series.push({ x: timestamp, y: val });

    // 🧹 Keep sliding window of 30 seconds
    const now = Date.now();
    const cutoff = now - 30000; // 30 seconds ago
    while (series.length > 0 && series[0].x < cutoff) {
        series.shift();
    }

    chart.updateSeries([{ data: series }], false); // false = NO full redraw
}



window.addEventListener("metrics", ({ detail: payload }) => {
    if (!payload?.metrics || !payload?.meta) return;

    if (payload.meta.endpoint_id?.startsWith("host-")) {
        updateMiniCharts(payload.metrics);
        const summary = extractHostSummary(payload.metrics, payload.meta);
        renderOverviewSummary(summary);
    }

    if (payload.meta.endpoint_id?.startsWith("ctr-")) {
        updateContainerTable(payload);
    }
});

window.addEventListener("logs", ({ detail: logPayload }) => {
    if (logPayload?.Logs?.length > 0) {
        for (const log of logPayload.Logs) {
            appendLogLine(log);
        }
    }
});

//
// Metric Updaters
//

function updateMiniCharts(metrics) {
    let cpuVal = null, memVal = null;
    let swapUsed = null, swapTotal = null;
    let metricTimestamp = null;

    for (const m of metrics) {
        if (m.namespace !== "System") continue;

        // Capture timestamp
        if (!metricTimestamp && m.timestamp) {
            metricTimestamp = m.timestamp * 1000;
        }

        if (m.subnamespace === "CPU" && m.name === "usage_percent" && m.dimensions?.scope === "total") {
            cpuVal = m.value;
        }

        if (m.subnamespace === "Memory") {
            if (m.name === "used_percent") {
                memVal = m.value;
            }
            if (m.name === "swap_used") {
                swapUsed = m.value;
            }
            if (m.name === "swap_total") {
                swapTotal = m.value;
            }
        }
    }

    const timestamp = metricTimestamp || Date.now();

    if (miniCharts.cpu) {
        const value = cpuVal !== null ? cpuVal : 0;
        appendMiniPoint(miniCharts.cpu, timestamp, value);
        latestCpuPercent = value;
        if (typeof cpuVal === "number" && !isNaN(cpuVal)) {
            document.getElementById("cpu-percent-label").textContent = cpuVal.toFixed(1) + "%";
        } else {
            document.getElementById("cpu-percent-label").textContent = "N/A";
        }
    }

    if (miniCharts.memory && memVal !== null) {
        appendMiniPoint(miniCharts.memory, timestamp, memVal);
        latestMemUsedPercent = memVal;
        document.getElementById("mem-percent-label").textContent = `${memVal.toFixed(1)}%`;
    }

    if (miniCharts.swap) {
        let swapPercent = 0;
        if (typeof swapUsed === "number" && typeof swapTotal === "number" && swapTotal > 0) {
            swapPercent = (swapUsed / swapTotal) * 100;
        }

        appendMiniPoint(miniCharts.swap, timestamp, swapPercent);
        latestSwapUsedPercent = swapPercent;
        document.getElementById("swap-percent-label").textContent = `${swapPercent.toFixed(1)}%`;
    }
}

//
//
// LOG STREAMING SECTION
//
///*

const maxLogLines = 10;

function appendLogLine(log) {
    const container = document.getElementById("log-stream");

    const div = document.createElement("div");
    div.className =
        "flex items-start space-x-2 mb-1 p-2 rounded-md shadow-sm border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 transition";

    const ts = new Date(log.timestamp).toLocaleTimeString();
    const level = log.level?.toUpperCase() || "INFO";
    const source = log.source || log.meta?.service || "unknown";
    const message = log.message || "";

    const levelColors = {
        ERROR: "bg-red-200 text-red-900 dark:bg-red-800 dark:text-red-200",
        WARN: "bg-yellow-200 text-yellow-900 dark:bg-yellow-800 dark:text-yellow-100",
        INFO: "bg-blue-200 text-blue-900 dark:bg-blue-800 dark:text-blue-100",
        DEBUG: "bg-gray-300 text-gray-900 dark:bg-gray-700 dark:text-gray-200",
    };

    const badge = document.createElement("span");
    badge.className = `text-[10px] font-semibold px-2 py-0.5 rounded ${levelColors[level] || "bg-gray-100 text-gray-600"}`;
    badge.textContent = level;

    const text = document.createElement("div");
    text.className = "flex-1 text-xs font-mono whitespace-pre-wrap break-words text-gray-800 dark:text-gray-200";
    text.textContent = `[${ts}] ${source}: ${message}`;

    div.appendChild(badge);
    div.appendChild(text);

    container.appendChild(div);

    while (container.children.length > maxLogLines) {
        container.removeChild(container.firstChild);
    }

    container.scrollTop = container.scrollHeight;
}

function logLevelColorClass(level) {
    switch (level.toLowerCase()) {
        case "error": return "text-red-600 dark:text-red-400";
        case "warn": return "text-yellow-600 dark:text-yellow-300";
        case "info": return "text-blue-600 dark:text-blue-400";
        case "debug": return "text-gray-600 dark:text-gray-400";
        default: return "text-gray-700 dark:text-gray-300";
    }
}

async function fetchRecentLogs() {
    try {
        const url = `/api/v1/logs?endpointID=${encodeURIComponent(window.endpointID)}&limit=${maxLogLines}`;
        const res = await gosightFetch(`/api/v1/logs?endpointID=${encodeURIComponent(window.endpointID)}&limit=10`);
        console.log("Fetching recent logs from:", url);
        if (!res.ok) throw new Error("Failed to fetch recent logs");
        const logs = await res.json();
        console.log("[Logs] Loaded:", logs);
        for (const log of logs) {
            appendLogLine(log);
        }
    } catch (err) {
        console.error("Failed to preload logs:", err);
    }
}
// END LOG STREAMING
///
/// ACTIVITY TAB SECTION
///



//// END ACTIVITY TAB
/////



function extractHostSummary(metrics, meta) {
    const summary = {
        hostname: meta.hostname,
        os: `${meta.platform} ${meta.platform_version} (${meta.architecture})`,
        uptime: 0,
        users: 0,
        procs: 0,
        cpu: {
            clock_mhz: 0,
            physical: 0,
            logical: 0,
            model: ""
        },
        memory: {
            total: 0,
            used: 0,
            used_percent: 0
        },
        disk: {
            total: 0,
            used: 0,
            used_percent: 0
        }
    };

    for (const m of metrics) {
        const { namespace, subnamespace, name, value, dimensions } = m;
        if (namespace !== "System") continue;

        if (subnamespace === "Host") {
            if (name === "uptime") summary.uptime = value;
            if (name === "procs") summary.procs = value;
            if (name === "users_loggedin") summary.users = value;
        }

        if (subnamespace === "CPU") {
            if (name === "count_physical") summary.cpu.physical = value;
            if (name === "count_logical") summary.cpu.logical = value;
            if (name === "clock_mhz") {
                summary.cpu.clock_mhz = value;
                if (!summary.cpu.model && dimensions?.model) {
                    summary.cpu.model = dimensions.model;
                }
            }
        }

        if (subnamespace === "Memory" && dimensions?.source === "physical") {
            if (name === "total") summary.memory.total = value;
            if (name === "used") summary.memory.used = value;
            if (name === "used_percent") summary.memory.used_percent = value;
        }

        if (subnamespace === "Disk" && dimensions?.mountpoint === "/") {
            if (name === "total") summary.disk.total = value;
            if (name === "used") summary.disk.used = value;
            if (name === "used_percent") summary.disk.used_percent = value;
        }
    }

    return summary;
}




function setupContainerFilters() {
    const statusFilter = document.getElementById("filter-container-status");
    const runtimeFilter = document.getElementById("filter-runtime");
    const hostFilter = document.getElementById("filter-container-name");

    function applyContainerFilters() {
        const statusVal = statusFilter.value.toLowerCase();
        const runtimeVal = runtimeFilter.value.toLowerCase();
        const hostVal = hostFilter.value.toLowerCase();

        const rows = document.querySelectorAll("#container-table-body tr");

        rows.forEach((row) => {
            const status = row.getAttribute("data-status")?.toLowerCase() || "";
            const runtime = row.getAttribute("data-runtime")?.toLowerCase() || "";
            const host = row.getAttribute("data-container-name")?.toLowerCase() || "";

            const matchStatus = !statusVal || status === statusVal;
            const matchRuntime = !runtimeVal || runtime === runtimeVal;
            const matchHost = !hostVal || host.includes(hostVal);

            row.style.display = matchStatus && matchRuntime && matchHost ? "" : "none";
        });
    }

    statusFilter.addEventListener("change", applyContainerFilters);
    runtimeFilter.addEventListener("change", applyContainerFilters);
    hostFilter.addEventListener("input", applyContainerFilters);
}

function renderOverviewSummary(summary) {

    document.getElementById("uptime").textContent = formatUptime(summary.uptime);
    document.getElementById("users").textContent = summary.users;
    document.getElementById("procs").textContent = summary.procs;
    document.getElementById("osinfo").textContent = summary.os;

    document.getElementById("cpu-info").textContent =
        `${summary.cpu.model} (${summary.cpu.physical} physical / ${summary.cpu.logical} logical @ ${summary.cpu.clock_mhz} MHz)`;



    document.getElementById("disk-used").textContent = formatBytes(summary.disk.used);
    document.getElementById("disk-total").textContent = formatBytes(summary.disk.total);
    document.getElementById("disk-percent").textContent = `${summary.disk.used_percent.toFixed(1)}%`;
}

function updateContainerTable(payload) {
    const tbody = document.getElementById("container-table-body");
    if (!tbody || !payload?.metrics || !payload?.meta) return;

    const meta = payload.meta;
    const metrics = payload.metrics;
    const id = meta.container_id;
    if (!id) return;
    //console.log(" Incoming container metrics for:", meta.container_name);
    metrics.forEach(m => {
        if (["cpu_percent", "mem_usage_bytes", "net_rx_bytes", "net_tx_bytes"].includes(m.name)) {
            //console.log(`🔧 ${m.name}:`, m.value);
        }
    });
    const container = {
        id,
        name: meta.container_name || "—",
        host: meta.hostname || "—",
        image: meta.image_name || "—",
        status: meta.tags?.status || "unknown",
        cpu: null,
        mem: null,
        rx: null,
        tx: null,
        uptime: null,
    };

    for (const m of metrics) {
        switch (m.name) {
            case "cpu_percent":
                container.cpu = typeof m.value === "number" ? m.value : null;
                break;
            case "mem_usage_bytes":
                container.mem = formatBytes(m.value);
                break;
            case "net_rx_bytes":
                container.rx = formatBytes(m.value);
                break;
            case "net_tx_bytes":
                container.tx = formatBytes(m.value);
                break;
            case "uptime_seconds":
                container.uptime = formatUptime(m.value);
                break;
        }
    }

    const isRunning = container.status === "running";
    const statusClass = isRunning
        ? "bg-green-100 text-green-800 dark:bg-green-800 dark:text-green-100"
        : "bg-red-100 text-red-800 dark:bg-red-800 dark:text-red-100";

    let row = tbody.querySelector(`tr[data-id="container-${id}"]`);
    const html = `
        <td class="px-4 py-2">${container.name}</td>
        <td class="px-4 py-2">${container.host}</td>
        <td class="px-4 py-2">${container.image}</td>
        <td class="px-4 py-2">
            <span class="inline-block px-3 py-1 text-xs font-bold rounded-full ${statusClass}">
                ${container.status}
            </span>
        </td>
<td class="px-4 py-2">${typeof container.cpu === "number" ? container.cpu.toFixed(1) + "%" : "0.0%"}</td>
        <td class="px-4 py-2">${container.mem || "—"}</td>
        <td class="px-4 py-2">${container.rx || "—"}</td>
        <td class="px-4 py-2">${container.tx || "—"}</td>
        <td class="px-4 py-2" title="">${container.uptime || "—"}</td>
    `;

    if (row) {
        row.innerHTML = html;
    } else {
        row = document.createElement("tr");
        row.setAttribute("data-id", `container-${id}`);
        row.setAttribute("data-status", container.status);      // "running" or "stopped"
        row.setAttribute("data-runtime", meta.subnamespace || ""); // "podman" or "docker"
        row.setAttribute("data-host", container.host);           // e.g. "DeepThought"
        row.setAttribute("data-container-name", container.name);
        row.innerHTML = html;
        tbody.appendChild(row);
    }
}
const processHistory = []; // stores { timestamp: ms, processes: [] }
window.addEventListener("process", (e) => {

    const ts = new Date(e.detail.timestamp).getTime();
    processHistory.push({ timestamp: ts, processes: e.detail.processes });

    // keep 30 minutes max
    const cutoff = Date.now() - 30 * 60 * 1000;
    while (processHistory.length > 0 && processHistory[0].timestamp < cutoff) {
        processHistory.shift();
    }

    const processes = e.detail.processes || [];

    // Top 5 by CPU %
    const topCPU = [...processes]
        .sort((a, b) => b.cpu_percent - a.cpu_percent)
        .slice(0, 5);

    // Top 5 by Memory %
    const topMem = [...processes]
        .sort((a, b) => b.mem_percent - a.mem_percent)
        .slice(0, 5);

    renderProcessTable("cpu-table", topCPU, "cpu_percent");
    renderProcessTable("mem-table", topMem, "mem_percent");
});

function renderProcessTable(tableId, processes, key) {
    const tbody = document.querySelector(`#${tableId} tbody`);
    if (!tbody) return;
    tbody.innerHTML = "";

    for (const proc of processes) {
        const row = document.createElement("tr");
        row.className = "hover:bg-gray-50 dark:hover:bg-gray-700";

        const value = proc[key] != null ? proc[key].toFixed(1) : "--";

        row.innerHTML = `
        <td class="px-4 py-2">${proc.pid}</td>
        <td class="px-4 py-2">${proc.user}</td>
        <td class="px-4 py-2">${value}</td>
        <td class="px-4 py-2 truncate max-w-xs" title="${proc.cmdline}">${proc.cmdline}</td>
      `;
        tbody.appendChild(row);
    }
}

//
// Initialization
//

async function initOverviewTab() {
    await renderMiniCharts();
    await fetchRecentLogs();
    setupContainerFilters();

}


registerTabInitializer("overview", async () => {
    await initOverviewTab();
});