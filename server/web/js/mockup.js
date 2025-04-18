const socket = new WebSocket("ws://" + location.host + "/ws/metrics?endpointID=" + encodeURIComponent(window.endpointID));

const chartAnimation = {
    tension: {
        duration: 1000,
        easing: "easeOutQuart",
        from: 0.4,
        to: 0,
        loop: false,
    },
};

const tooltipPlugin = {
    enabled: true,
    callbacks: {
        label: function (context) {
            return `${context.dataset.label || ""}: ${context.parsed.y}`;
        },
    },
};

const miniCharts = {
    cpu: null,
    memory: null,
    swap: null,  // new
};

let latestCpuPercent = 0;
let latestSwapUsedPercent = 0;
let latestMemUsedPercent = 0;

socket.onmessage = (event) => {
    try {
        const envelope = JSON.parse(event.data);
        console.log("📦 WebSocket message:", envelope);
        if (envelope.type === "logs") {
            console.log("📄 Logs:\n" + JSON.stringify(envelope.data.Logs, null, 2));
        }
        if (envelope.type === "metrics") {
            const payload = envelope.data;
            if (!payload?.metrics || !payload?.meta) return;

            if (payload.meta.endpoint_id?.startsWith("host-")) {
                updateMiniCharts(payload.metrics);
                const summary = extractHostSummary(payload.metrics, payload.meta);
                renderOverviewSummary(summary);
            }

            if (payload.meta.endpoint_id?.startsWith("ctr-")) {
                updateContainerTable(payload);
            }
        }

        if (envelope.type === "logs") {
            const logPayload = envelope.data;
            if (logPayload?.Logs?.length > 0) {
                for (const log of logPayload.Logs) {
                    appendLogLine(log);
                    appendActivityRow(log);    // Activity tab (table)
                }

            }
        }

    } catch (err) {
        console.error("❌ Failed to parse WebSocket JSON:", err);
    }
};
//
//
// LOG STREAMING SECTION
//
//
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
const activityLogs = [];
const logsPerPage = 50;
let currentPage = 1;


function appendActivityRow(log) {
    activityLogs.unshift(log); // 👈 newest first

    // Cap memory
    if (activityLogs.length > 500) {
        activityLogs.length = 500;
    }

    // Re-render current page
    renderActivityPage(currentPage);
}

function renderActivityPage(page) {
    const tbody = document.getElementById("activity-log-body");
    if (!tbody) return;

    tbody.innerHTML = "";

    const start = (page - 1) * logsPerPage;
    const end = start + logsPerPage;
    const pageLogs = activityLogs.slice(start, end);

    for (const log of pageLogs) {
        const row = document.createElement("tr");
        const level = log.level || "info";
        const timestamp = new Date(log.timestamp).toLocaleString();
        const message = log.message || "";

        const levelClass = {
            error: "bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-300",
            warn: "bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-300",
            info: "bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-300",
            notice: "bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-300",
            debug: "bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-300",
        }[level] || "bg-gray-200 text-gray-800 dark:bg-gray-800 dark:text-gray-300";

        row.innerHTML = `
            <td class="px-4 py-2 whitespace-nowrap">
                <span class="inline-block text-xs font-medium px-2 py-0.5 rounded ${levelClass}">${level}</span>
            </td>
            <td class="px-4 py-2 whitespace-nowrap">${timestamp}</td>
            <td class="px-4 py-2">${message}</td>
        `;
        tbody.appendChild(row);
    }

    updateActivityPagination();
}

function updateActivityPagination() {
    const pageIndicator = document.getElementById("activity-page-indicator");
    if (pageIndicator) {
        const maxPage = Math.max(1, Math.ceil(activityLogs.length / logsPerPage));
        pageIndicator.textContent = `Page ${currentPage} of ${maxPage}`;
    }
}


document.getElementById("activity-prev").addEventListener("click", () => {
    if (currentPage > 1) {
        currentPage--;
        renderActivityPage(currentPage);
    }
});

document.getElementById("activity-next").addEventListener("click", () => {
    const maxPage = Math.ceil(activityLogs.length / logsPerPage);
    if (currentPage < maxPage) {
        currentPage++;
        renderActivityPage(currentPage);
    }
});
////


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

function renderMiniCharts() {
    miniCharts.cpu = new Chart(document.getElementById("miniCpuChart"), {
        type: "line",
        data: {
            labels: [],
            datasets: [{
                data: [],
                borderColor: "#3b82f6",
                backgroundColor: "rgba(59, 130, 246, 0.1)",
                tension: 0.4,
                fill: true,
                pointRadius: 0,
            }],
        },
        options: {
            responsive: true,
            plugins: {
                legend: { display: false },
                tooltip: tooltipPlugin,
            },
            scales: { y: { display: true }, x: { display: false } },
            elements: { line: { borderWidth: 2 } },
            animations: chartAnimation,
        },
    });

    miniCharts.memory = new Chart(document.getElementById("miniMemoryChart"), {
        type: "line",
        data: {
            labels: [],
            datasets: [{
                data: [],
                borderColor: "#10b981",
                backgroundColor: "rgba(16, 185, 129, 0.1)",
                tension: 0.4,
                fill: true,
                pointRadius: 0,
            }],
        },
        options: {
            responsive: true,
            plugins: {
                legend: { display: false },
                tooltip: tooltipPlugin,
            },
            scales: { y: { display: true }, x: { display: false } },
            elements: { line: { borderWidth: 2 } },
            animations: chartAnimation,
        },
    });
    miniCharts.swap = new Chart(document.getElementById("miniSwapChart"), {
        type: "line",
        data: {
            labels: [],
            datasets: [{
                data: [],
                borderColor: "#f87171", // red-400
                backgroundColor: "rgba(248, 113, 113, 0.1)",
                tension: 0.4,
                fill: true,
                pointRadius: 0,
            }],
        },
        options: {
            responsive: true,
            plugins: {
                legend: { display: false },
                tooltip: tooltipPlugin,
            },
            scales: { y: { display: true }, x: { display: false } },
            elements: { line: { borderWidth: 2 } },
            animations: chartAnimation,
        },
    });
}

function updateMiniCharts(metrics) {
    let cpuVal = null;
    let memVal = null;
    let swapVal = null;
    metrics.forEach((m) => {
        if (m.subnamespace === "Memory" && m.dimensions?.source === "swap") {
            //console.log("🟢 SWAP METRIC RECEIVED:", m.name, m.value);
        }
    });
    for (const m of metrics) {
        if (
            m.namespace === "System" &&
            m.subnamespace === "CPU" &&
            m.name === "usage_percent" &&
            m.dimensions?.scope === "total"
        ) {
            cpuVal = m.value;
        }

        if (
            m.namespace === "System" &&
            m.subnamespace === "Memory" &&
            m.name === "used_percent" &&
            m.dimensions?.source === "physical"
        ) {
            memVal = m.value;
        }
        if (
            m.namespace === "System" &&
            m.subnamespace === "Memory" &&
            m.name === "used_percent" &&
            m.dimensions?.source === "swap"
        ) {
            swapVal = m.value;
        }
    }

    const timestamp = new Date().toLocaleTimeString([], {
        hour: "2-digit",
        minute: "2-digit",
        second: "2-digit",
    });

    if (miniCharts.cpu && cpuVal !== null) {
        const d = miniCharts.cpu.data;
        const val = Math.abs(cpuVal - latestCpuPercent) > 0.1 ? cpuVal : latestCpuPercent;

        d.labels.push(timestamp);
        d.datasets[0].data.push(val);

        if (d.labels.length > 30) {
            d.labels.shift();
            d.datasets[0].data.shift();
        }

        miniCharts.cpu.update();
        latestCpuPercent = val;

        const label = document.getElementById("cpu-percent-label");
        if (label) label.textContent = `${val.toFixed(1)}%`;
    }

    if (miniCharts.memory) {
        const val = memVal !== null ? memVal : latestMemUsedPercent;
        const d = miniCharts.memory.data;

        d.labels.push(timestamp);
        d.datasets[0].data.push(val);

        if (d.labels.length > 30) {
            d.labels.shift();
            d.datasets[0].data.shift();
        }

        miniCharts.memory.update();

        if (memVal !== null) {
            latestMemUsedPercent = val;
            const label = document.getElementById("mem-percent-label");
            if (label) label.textContent = `${val.toFixed(1)}%`;
        }
    }

    if (miniCharts.swap) {
        const val = typeof swapVal === "number" && !isNaN(swapVal) ? swapVal : latestSwapUsedPercent;
        const d = miniCharts.swap.data;

        d.labels.push(timestamp);
        d.datasets[0].data.push(val);

        if (d.labels.length > 30) {
            d.labels.shift();
            d.datasets[0].data.shift();
        }

        miniCharts.swap.update();

        if (typeof swapVal === "number" && !isNaN(swapVal)) {
            latestSwapUsedPercent = swapVal;
            const label = document.getElementById("swap-percent-label");
            if (label) label.textContent = `${val.toFixed(1)}%`;
        }
    }


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

    document.getElementById("mem-used").textContent = formatBytes(summary.memory.used);
    document.getElementById("mem-total").textContent = formatBytes(summary.memory.total);
    document.getElementById("mem-percent").textContent = `${summary.memory.used_percent.toFixed(1)}%`;

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
    //console.log("📦 Incoming container metrics for:", meta.container_name);
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


function formatBytes(bytes) {
    if (bytes === undefined || bytes === null || isNaN(bytes)) return "—";

    const units = ["B", "KB", "MB", "GB", "TB"];
    let i = 0;
    while (bytes >= 1024 && i < units.length - 1) {
        bytes /= 1024;
        i++;
    }
    return `${bytes.toFixed(1)} ${units[i]}`;
}

function formatUptime(seconds) {
    if (typeof seconds !== "number" || isNaN(seconds) || seconds <= 0) return "—";

    const d = Math.floor(seconds / (3600 * 24));
    const h = Math.floor((seconds % (3600 * 24)) / 3600);
    const m = Math.floor((seconds % 3600) / 60);
    return `${d}d ${h}h ${m}m`;
}


document.addEventListener("DOMContentLoaded", () => {
    renderMiniCharts();
    setupContainerFilters(); // 👈 Add this line
});
