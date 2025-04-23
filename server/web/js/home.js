document.addEventListener("DOMContentLoaded", async function () {
    await loadRadial("/api/v1/alerts", "radial-alerts", "Alerts", "#ef4444");
    await loadRadial("/api/v1/agents", "radial-agents", "Agents", "#10b981");
    await loadRadial("/api/v1/query?metric=container.podman.running&value=1", "radial-containers", "Containers", "#3b82f6");
    await loadRadial("/api/v1/query?metric=system.cpu.usage_percent&scope=total", "radial-load", "System Load", "#f59e0b");

    await loadTrend("/api/v1/query?metric=system.cpu.usage_percent&scope=total", "trend-cpu", "CPU %");
    await loadTrend("/api/v1/query?metric=system.memory.used_percent", "trend-mem", "Mem %");
    await loadTrend("/api/v1/query?metric=system.disk.used_percent", "trend-disk", "Disk %");
    await loadTrend("/api/v1/query?metric=system.network.bytes_recv&scope=total", "trend-net", "Net In");

    await loadTop("/api/v1/query?metric=system.cpu.usage_percent&scope=total&limit=5", "top-cpu-hosts", "Top CPU Hosts", "%");
    await loadTop("/api/v1/query?sort=desc&metric=container.mem.rss&limit=5", "top-mem-containers", "Top Memory Containers", "MB");

    const events = await fetchJson("/api/v1/events?limit=10");
    renderEventLog(events || []);
});

async function fetchJson(url) {
    try {
        const res = await fetch(url);
        if (!res.ok) throw new Error("Failed to fetch");
        return await res.json();
    } catch (e) {
        console.error("Fetch error", e);
        return null;
    }
}

async function loadRadial(url, elId, label, color, total = 100) {
    const data = await fetchJson(url);
    const value = Array.isArray(data) ? (data.length / total) * 100 : (data?.percentage || 0);
    new ApexCharts(document.getElementById(elId), {
        chart: { type: "radialBar", height: 140 },
        series: [value],
        labels: [label],
        colors: [color],
        plotOptions: {
            radialBar: {
                hollow: { size: "60%" },
                dataLabels: {
                    name: { fontSize: "12px" },
                    value: { fontSize: "16px", fontWeight: 600 },
                },
            },
        },
    }).render();
}


async function updateAgentSummary() {
    try {
        const res = await fetch("/api/v1/agents");
        const agents = await res.json();

        const online = agents.filter(a => a.status === "Online").length;
        const offline = agents.length - online;

        const statusText = `${online} Online / ${offline} Offline`;

        document.getElementById("agent-health-status").textContent = statusText;
    } catch (err) {
        console.error("Failed to fetch agents:", err);
        document.getElementById("agent-health-status").textContent = "Unavailable";
    }
}
async function loadTrend(url, elId, label) {
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
export async function updateAlertSummary() {
    try {
        const res = await fetch("/api/v1/alerts");
        const alerts = await res.json();

        if (!Array.isArray(alerts)) throw new Error("invalid alert format");

        const counts = alerts.reduce((acc, alert) => {
            const level = alert.level || "unknown";
            acc[level] = (acc[level] || 0) + 1;
            return acc;
        }, {});

        let summary = Object.entries(counts)
            .map(([level, count]) => `${count} ${capitalize(level)}`)
            .join(", ");

        if (!summary) summary = "None";

        document.getElementById("alert-status").textContent = summary;
    } catch (err) {
        console.error("Failed to fetch alerts:", err);
        document.getElementById("alert-status").textContent = "Unavailable";
    }
}

function capitalize(str) {
    return str.charAt(0).toUpperCase() + str.slice(1);
}

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
}

function renderEventLog(events) {
    const el = document.getElementById("event-log");
    el.innerHTML = "";
    events.forEach(e => {
        const level = e.level || "Info";
        const icon = {
            Info: "text-blue-500",
            Warning: "text-yellow-500",
            Error: "text-red-500"
        }[level] || "text-gray-500";

        el.innerHTML += `
        <div class="flex items-start space-x-2">
          <span class="font-bold ${icon}">${level}</span>
          <span class="text-xs text-gray-400">${new Date(e.timestamp).toLocaleTimeString()}</span>
          <span class="flex-1">${e.message || e.summary || "-"}</span>
        </div>
      `;
    });
}
