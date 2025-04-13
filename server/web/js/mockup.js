const ws = new WebSocket("ws://localhost:8080/ws/metrics");

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
};

let latestCpuPercent = 0;
let latestMemUsedPercent = 0;

ws.onmessage = (event) => {
    try {
        const data = JSON.parse(event.data);
        console.log("🧩 WebSocket payload:", data);
        if (!data.metrics || !data.meta) return;

        if (data.meta.endpoint_id?.startsWith("host-")) {
            updateMiniCharts(data.metrics);
            const summary = extractHostSummary(data.metrics, data.meta);
            renderOverviewSummary(summary);
        }

        if (data.meta.endpoint_id?.startsWith("container-")) {
            updateContainerTable(data);
        }

    } catch (err) {
        console.error("❌ Failed to parse WebSocket JSON:", err);
    }
};

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
}

function updateMiniCharts(metrics) {
    let cpuVal = null;
    let memVal = null;

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
    console.log("📦 Incoming container metrics for:", meta.container_name);
    metrics.forEach(m => {
        if (["cpu_percent", "mem_usage_bytes", "net_rx_bytes", "net_tx_bytes"].includes(m.name)) {
            console.log(`🔧 ${m.name}:`, m.value);
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
                container.cpu = m.value?.toFixed(1);
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
        <td class="px-4 py-2">${container.cpu !== null ? container.cpu + "%" : "—"}</td>
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
    const d = Math.floor(seconds / (3600 * 24));
    const h = Math.floor((seconds % (3600 * 24)) / 3600);
    const m = Math.floor((seconds % 3600) / 60);
    return `${d}d ${h}h ${m}m`;
}

document.addEventListener("DOMContentLoaded", () => {
    renderMiniCharts();
});
