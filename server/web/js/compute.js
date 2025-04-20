let cpuLineChart = null;
let cpuLoadChart = null;
let memoryLineChart = null;
let cpuDonutChart = null;
let memoryDonutChart = null;

const cpuTimeMetrics = { User: 0, System: 0, Idle: 0, Nice: 0 };
const memoryMetrics = { used: 0, free: 0, buffers: 0, cache: 0 };

// Buffer for chart data before render
const cpuUsageBuffer = [];
const memoryUsageBuffer = [];
const cpuDonutBuffer = {};
const memoryDonutBuffer = {};
const cpuLoadBuffer = []; // ✅ array of full { time, load1, load5, load15 }

// Aggregate load data
const pendingLoadAvg = {
    "load_avg_1": null,
    "load_avg_5": null,
    "load_avg_15": null
};

function createCpuCharts() {
    if (typeof Chart === "undefined") return;
    const chartIds = [
        "cpuUsageChart",
        "cpuLoadChart",
        "memoryUsageChart",
        "cpuDonutChart",
        "memoryDonutChart"
    ];

    for (const id of chartIds) {
        const el = document.getElementById(id);
        if (!el || !el.offsetParent) {
            console.warn(`⏳ ${id} is not visible yet. Retrying...`);
            setTimeout(createCpuCharts, 100); // try again shortly
            return;
        }
    }

    // ✅ Safe to create charts now
    console.log("🎨 All canvases visible. Creating charts...");
    const ctx = document.getElementById("cpuLoadChart").getContext("2d");
    cpuLineChart = new Chart(document.getElementById("cpuUsageChart"), {
        type: "line",
        data: {
            labels: [],
            datasets: [{
                label: "CPU Usage %",
                data: [],
                borderColor: "#3b82f6",
                backgroundColor: "rgba(59, 130, 246, 0.1)",
                fill: true,
                tension: 0.3,
                pointRadius: 2,
                spanGaps: true
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            animation: { duration: 0 },
            scales: {
                x: {
                    ticks: { color: "#9CA3AF", maxTicksLimit: 10, autoSkip: true }
                },
                y: {
                    beginAtZero: true,
                    ticks: {
                        color: "#9CA3AF",
                        callback: val => `${val}%`
                    }
                }
            },
            plugins: {
                legend: { labels: { color: "#4B5563" } },
                tooltip: {
                    callbacks: {
                        label: ctx => `CPU: ${ctx.parsed.y.toFixed(1)}%`
                    }
                }
            }
        }
    });


    cpuLoadChart = new Chart(document.getElementById("cpuLoadChart"), {
        type: "line",
        data: {
            labels: [],
            datasets: [
                {
                    label: "Load Avg (1m)",
                    data: [],
                    borderColor: "#3b82f6",
                    backgroundColor: "rgba(59, 130, 246, 0.1)",
                    fill: true,
                    tension: 0.3,
                    pointRadius: 2,
                    spanGaps: true
                },
                {
                    label: "Load Avg (5m)",
                    data: [],
                    borderColor: "#10b981",
                    backgroundColor: "rgba(16, 185, 129, 0.1)",
                    fill: true,
                    tension: 0.3,
                    pointRadius: 2,
                    spanGaps: true
                },
                {
                    label: "Load Avg (15m)",
                    data: [],
                    borderColor: "#f59e0b",
                    backgroundColor: "rgba(245, 158, 11, 0.1)",
                    fill: true,
                    tension: 0.3,
                    pointRadius: 2,
                    spanGaps: true
                }
            ]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            animation: { duration: 0 },
            scales: {
                x: {
                    ticks: { color: "#9CA3AF", maxTicksLimit: 10, autoSkip: true }
                },
                y: {
                    beginAtZero: true,
                    ticks: {
                        color: "#9CA3AF"
                    }
                }
            },
            plugins: {
                legend: { labels: { color: "#4B5563" } },
                tooltip: {
                    callbacks: {
                        label: ctx => `${ctx.dataset.label}: ${ctx.parsed.y.toFixed(2)}`
                    }
                }
            }
        }
    });

    memoryLineChart = new Chart(document.getElementById("memoryUsageChart"), {
        type: "line",
        data: {
            labels: [],
            datasets: [{
                label: "Memory Usage %",
                data: [],
                borderColor: "#f59e0b", // amber
                backgroundColor: "rgba(245, 158, 11, 0.1)",
                fill: true,
                tension: 0.3,
                pointRadius: 2,
                spanGaps: true
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            animation: { duration: 0 },
            scales: {
                x: {
                    ticks: {
                        color: "#9CA3AF",
                        maxTicksLimit: 10,
                        autoSkip: true
                    }
                },
                y: {
                    beginAtZero: true,
                    ticks: {
                        color: "#9CA3AF",
                        callback: val => `${val}%`
                    }
                }
            },
            plugins: {
                legend: { labels: { color: "#4B5563" } },
                tooltip: {
                    callbacks: {
                        label: ctx => `Memory: ${ctx.parsed.y.toFixed(1)}%`
                    }
                }
            }
        }
    });

    cpuDonutChart = new Chart(document.getElementById("cpuDonutChart"), {
        type: "doughnut",
        data: { labels: Object.keys(cpuTimeMetrics), datasets: [{ data: Object.values(cpuTimeMetrics), backgroundColor: ["#3b82f6", "#10b981", "#9ca3af", "#f59e0b"] }] },
        options: { responsive: true }
    });

    memoryDonutChart = new Chart(document.getElementById("memoryDonutChart"), {
        type: "doughnut",
        data: { labels: Object.keys(memoryMetrics), datasets: [{ data: Object.values(memoryMetrics), backgroundColor: ["#10b981", "#3b82f6", "#f59e0b", "#f87171"] }] },
        options: { responsive: true }
    });
}

function appendCpuLoadPoint(point) {
    cpuLoadChart.data.labels.push(point.time);
    cpuLoadChart.data.datasets[0].data.push(point.load1);
    cpuLoadChart.data.datasets[1].data.push(point.load5);
    cpuLoadChart.data.datasets[2].data.push(point.load15);

    if (cpuLoadChart.data.labels.length > 60) {
        cpuLoadChart.data.labels.shift();
        for (const ds of cpuLoadChart.data.datasets) ds.data.shift();
    }

    cpuLoadChart.update();
}

function updateComputeLineChart(chart, value, buffer) {
    const timeLabel = new Date().toLocaleTimeString();
    if (!chart) {
        if (buffer.length > 100) buffer.shift(); // avoid unbounded memory growth
        buffer.push({ time: timeLabel, value });
        return;
    }

    // If chart exists, flush any buffered values first
    if (buffer.length) {
        console.log("⏩ Flushing buffer with", buffer.length, "points");
        for (const entry of buffer) {
            chart.data.labels.push(entry.time);
            chart.data.datasets[0].data.push(entry.value);
        }
        buffer.length = 0; // clear the buffer
    }
    chart.data.labels.push(timeLabel);
    chart.data.datasets[0].data.push(value);

    if (chart.data.labels.length > 60) {
        chart.data.labels.shift();
        chart.data.datasets[0].data.shift();
    }

    chart.update();
}

function updateCpuDonutChart() {
    if (!cpuDonutChart) return;

    for (const key of Object.keys(cpuTimeMetrics)) {
        if (cpuDonutBuffer[key] != null) {
            cpuTimeMetrics[key] = cpuDonutBuffer[key];
        }
    }

    cpuDonutChart.data.datasets[0].data = Object.values(cpuTimeMetrics);
    cpuDonutChart.update();
}

function updateMemoryDonutChart() {
    if (!memoryDonutChart) return;

    const { total, available } = memoryDonutBuffer;
    if (total == null || available == null) return;

    const used = total - available;
    memoryMetrics.used = used;
    memoryMetrics.free = available;
    memoryDonutChart.data.datasets[0].data = Object.values(memoryMetrics);
    memoryDonutChart.update();
}

window.cpuMetricHandler = function (metrics) {
    console.log("📡 Received metrics:", metrics);

    for (const metric of metrics) {
        if (metric.namespace !== "System") continue;

        const key = `${metric.namespace}.${metric.subnamespace}.${metric.name}`.toLowerCase();

        switch (key) {
            case "system.cpu.usage_percent":
                if (metric.dimensions.scope === "total") {
                    updateComputeLineChart(cpuLineChart, metric.value, cpuUsageBuffer);
                }
                break;
            case "system.cpu.load_avg_1":
            case "system.cpu.load_avg_5":
            case "system.cpu.load_avg_15":
                const keyName = metric.name;
                pendingLoadAvg[keyName] = metric.value;

                if (
                    pendingLoadAvg["load_avg_1"] != null &&
                    pendingLoadAvg["load_avg_5"] != null &&
                    pendingLoadAvg["load_avg_15"] != null
                ) {
                    const label = new Date().toLocaleTimeString();
                    const point = {
                        time: label,
                        load1: pendingLoadAvg["load_avg_1"],
                        load5: pendingLoadAvg["load_avg_5"],
                        load15: pendingLoadAvg["load_avg_15"]
                    };

                    if (!cpuLoadChart) {
                        if (cpuLoadBuffer.length > 100) cpuLoadBuffer.shift();
                        cpuLoadBuffer.push(point);
                    } else {
                        appendCpuLoadPoint(point);
                    }

                    // Clear buffer
                    pendingLoadAvg["load_avg_1"] = null;
                    pendingLoadAvg["load_avg_5"] = null;
                    pendingLoadAvg["load_avg_15"] = null;
                }
                break;

            case "system.cpu.time_user":
                cpuDonutBuffer.User = metric.value;
                if (cpuDonutChart) updateCpuDonutChart();
                break;

            case "system.cpu.time_system":
                cpuDonutBuffer.System = metric.value;
                if (cpuDonutChart) updateCpuDonutChart();
                break;

            case "system.cpu.time_idle":
                cpuDonutBuffer.Idle = metric.value;
                if (cpuDonutChart) updateCpuDonutChart();
                break;

            case "system.cpu.time_nice":
                cpuDonutBuffer.Nice = metric.value;
                if (cpuDonutChart) updateCpuDonutChart();
                break;

            case "system.memory.used_percent":
                updateComputeLineChart(memoryLineChart, metric.value, memoryUsageBuffer);
                break;

            case "system.memory.total":
                memoryDonutBuffer.total = metric.value;
                if (memoryDonutChart) updateMemoryDonutChart();
                break;

            case "system.memory.available":
                memoryDonutBuffer.available = metric.value;
                if (memoryDonutChart) updateMemoryDonutChart();
                break;

            case "system.memory.memory_details":
                memoryMetrics.used = metric.dimensions.used;
                memoryMetrics.free = metric.dimensions.free;
                memoryMetrics.buffers = metric.dimensions.buffers;
                memoryMetrics.cache = metric.dimensions.cache;
                updateMemoryDonutChart();
                break;

            default:
                break;
        }
    }
};

function initComputeTab() {
    createCpuCharts();
}

registerTabInitializer("compute", initComputeTab);