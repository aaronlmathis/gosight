import { formatBytes, formatUptime } from "./format.js";
import { registerTabInitializer } from "./tabs.js";

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
            y: { formatter: val => `${val.toFixed(1)}%` }
        },
        grid: { borderColor: "#e0e0e0", strokeDashArray: 4 }
    };

    const chart = new ApexCharts(document.getElementById(elementId), options);
    return chart;
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

//
// WebSocket Connection
//

function connectWebSocket() {
    const socket = new WebSocket(`wss://${location.host}/ws/metrics?endpointID=${encodeURIComponent(window.endpointID)}`);
    console.log("WebSocket connecting for:", window.endpointID);
    const statusBadge = document.getElementById("ws-status-badge");

    socket.addEventListener("open", () => {
        console.log("✅ WebSocket connected");

        if (statusBadge) {
            statusBadge.classList.remove("hidden", "bg-red-500", "text-red-100");
            statusBadge.classList.add("bg-green-500", "text-green-100");
            statusBadge.textContent = "Connected";
        }
    });

    socket.addEventListener("close", () => {
        console.log("❌ WebSocket disconnected");

        if (statusBadge) {
            statusBadge.classList.remove("hidden", "bg-green-500", "text-green-100");
            statusBadge.classList.add("bg-red-500", "text-red-100");
            statusBadge.textContent = "Disconnected";
        }
    });

    socket.addEventListener("error", (e) => {
        console.error("WebSocket error:", e);
        // Optional: show error visually too
    });
    socket.onmessage = (event) => {
        try {
            const envelope = JSON.parse(event.data);
            if (!envelope?.type) return;

            switch (envelope.type) {
                case "metrics":
                    handleMetricsPayload(envelope.data);
                    break;
                case "logs":
                    handleLogsPayload(envelope.data);
                    break;
                case "event":
                    if (window.eventHandler) window.eventHandler([envelope.data]);
                    break;
            }
        } catch (err) {
            console.error("Failed to parse WebSocket JSON:", err);
        }
    };
}

function handleMetricsPayload(payload) {
    if (!payload?.metrics || !payload?.meta) return;

    if (payload.meta.endpoint_id?.startsWith("host-")) {
        updateMiniCharts(payload.metrics);
        const summary = extractHostSummary(payload.metrics, payload.meta);
        renderOverviewSummary(summary);
        if (window.networkMetricHandler) window.networkMetricHandler(payload.metrics);
        if (window.cpuMetricHandler) window.cpuMetricHandler(payload.metrics);
        if (window.diskMetricHandler) window.diskMetricHandler(payload.metrics);
    }

    if (payload.meta.endpoint_id?.startsWith("ctr-")) {
        updateContainerTable(payload);
    }
}

function handleLogsPayload(logPayload) {
    if (logPayload?.Logs?.length > 0) {
        for (const log of logPayload.Logs) {
            appendLogLine(log);
            appendActivityRow(log);
        }
    }
}

//
// Metric Updaters
//

function updateMiniCharts(metrics) {
    let cpuVal = null, memVal = null, swapVal = null;
    let metricTimestamp = null;

    for (const m of metrics) {
        if (m.namespace !== "System") continue;

        // Capture timestamp of the metric
        if (!metricTimestamp && m.timestamp) {
            metricTimestamp = m.timestamp * 1000;
        }

        if (m.subnamespace === "CPU" && m.name === "usage_percent" && m.dimensions?.scope === "total") {
            cpuVal = m.value;
        }
        if (m.subnamespace === "Memory" && m.name === "used_percent") {
            memVal = m.value;
        }
        if (m.subnamespace === "Memory" && m.name === "swap_used_percent") {
            swapVal = m.value;
        }
    }

    // Fallback if metricTimestamp wasn't found
    const timestamp = metricTimestamp || Date.now();

    if (miniCharts.cpu) {
        const value = cpuVal !== null ? cpuVal : 0;
        appendMiniPoint(miniCharts.cpu, timestamp, value);
        latestCpuPercent = value;
        document.getElementById("cpu-percent-label").textContent = `${value.toFixed(1)}%`;
    }

    if (miniCharts.memory && memVal !== null) {
        appendMiniPoint(miniCharts.memory, timestamp, memVal);
        latestMemUsedPercent = memVal;
        document.getElementById("mem-percent-label").textContent = `${memVal.toFixed(1)}%`;
    }

    if (miniCharts.swap && typeof swapVal === "number" && !isNaN(swapVal)) {
        appendMiniPoint(miniCharts.swap, timestamp, swapVal);
        latestSwapUsedPercent = swapVal;
        document.getElementById("swap-percent-label").textContent = `${swapVal.toFixed(1)}%`;
    }
}





//
//
// LOG STREAMING SECTION
//
//
function appendActivityRow(log) {
    const tbody = document.getElementById("activity-log-body");
    if (!tbody || !log) return;

    const row = document.createElement("tr");

    // Determine badge color based on level
    const level = (log.level || "").toLowerCase();
    const badgeClass = {
        error: "bg-red-100 text-red-800 dark:bg-red-800 dark:text-red-200",
        warn: "bg-yellow-100 text-yellow-800 dark:bg-yellow-800 dark:text-yellow-100",
        info: "bg-blue-100 text-blue-800 dark:bg-blue-800 dark:text-blue-100",
        debug: "bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-300",
        notice: "bg-green-100 text-green-800 dark:bg-green-800 dark:text-green-300",
    }[level] || "bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-300";

    const category = log.category || log.meta?.unit || log.meta?.service || log.source || "unknown";
    const message = log.message || "";
    const timestamp = new Date(log.timestamp).toLocaleString();

    row.innerHTML = `
        <td class="px-4 py-2">
            <span class="text-xs font-semibold px-2 py-0.5 rounded ${badgeClass}">
                ${level.toUpperCase()}
            </span>
        </td>
        <td class="px-4 py-2 whitespace-nowrap text-xs text-gray-500 dark:text-gray-400">${timestamp}</td>
        <td class="px-4 py-2 text-sm text-gray-700 dark:text-gray-300">${message}</td>
    `;

    tbody.prepend(row);

    // Limit to 100 rows max
    while (tbody.children.length > 100) {
        tbody.removeChild(tbody.lastChild);
    }
}
const maxLogLines = 10;
const logContainer = document.getElementById("log-stream");

function renderLogLine(log) {
    const ts = new Date(log.timestamp).toLocaleTimeString();
    const level = log.level?.toUpperCase() || "INFO";
    const source = log.source || log.meta?.service || "unknown";
    const message = log.message || "";

    return `[${ts}] [${level}] ${source}: ${message}`;
}
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
    document.getElementById("hostname").textContent = summary.hostname;
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
        image: meta.image_id || "—",
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


//
// Initialization
//

async function initOverviewTab() {
    console.log("✅ initOverviewTab running");

    document.getElementById("overview-content")?.classList.remove("hidden");
    console.log("✅ overview-content unhidden");

    await renderMiniCharts();
    console.log("✅ miniCharts rendered");

    setupContainerFilters();
    connectWebSocket();
    console.log("✅ WebSocket connected");

    document.getElementById("overview-skeleton")?.classList.add("hidden");
    console.log("✅ Skeleton hidden, dashboard fully live");
}


registerTabInitializer("overview", async () => {
    await initOverviewTab();
});