let cpuLineChart = null;
let cpuLoadChart = null;
let memoryLineChart = null;
let cpuDonutChart = null;
let memoryDonutChart = null;
let swapLineChart = null;

const cpuTimeMetrics = { User: 0, System: 0, Idle: 0, Nice: 0 };
const memoryMetrics = { used: 0, free: 0, buffers: 0, cache: 0 };

// Buffer for chart data before render
const cpuUsageBuffer = [];
const memoryUsageBuffer = [];
const cpuDonutBuffer = {};
const memoryDonutBuffer = {};
const cpuLoadBuffer = []; // ✅ array of full { time, load1, load5, load15 }
const swapUsageBuffer = [];
const swapPercentMeta = { total: null, free: null };

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
        "memoryDonutChart",
        "swapUsageChart"
    ];

    for (const id of chartIds) {
        const el = document.getElementById(id);
        if (!el || !el.offsetParent) {
            console.warn(`⏳ ${id} is not visible yet. Retrying...`);
            setTimeout(createCpuCharts, 100); // try again shortly
            return;
        }
    }

    // Safe to create charts now
    console.log(" All canvases visible. Creating charts...");
    const canvas = document.getElementById("cpuLoadChart");
    const ctx = canvas.getContext("2d");

    //  Define gradients (vertical fade)
    const gradient1 = ctx.createLinearGradient(0, 0, 0, canvas.height);
    gradient1.addColorStop(0, "rgba(59, 130, 246, 0.2)");
    gradient1.addColorStop(1, "rgba(59, 130, 246, 0)");

    const gradient5 = ctx.createLinearGradient(0, 0, 0, canvas.height);
    gradient5.addColorStop(0, "rgba(16, 185, 129, 0.2)");
    gradient5.addColorStop(1, "rgba(16, 185, 129, 0)");

    const gradient15 = ctx.createLinearGradient(0, 0, 0, canvas.height);
    gradient15.addColorStop(0, "rgba(245, 158, 11, 0.2)");
    gradient15.addColorStop(1, "rgba(245, 158, 11, 0)");

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
                legend: { position: "bottom", labels: { color: "#4B5563" } },
                tooltip: {
                    callbacks: {
                        label: ctx => `CPU: ${ctx.parsed.y.toFixed(1)}%`
                    }
                }
            }
        }
    });


    cpuLoadChart = new Chart(canvas, {
        type: "line",
        data: {
            labels: [],
            datasets: [
                {
                    label: "Load Avg (1m)",
                    data: [],
                    borderColor: "#3b82f6",
                    backgroundColor: gradient1,
                    fill: true,
                    tension: 0.3,
                    pointRadius: 2,
                    spanGaps: true
                },
                {
                    label: "Load Avg (5m)",
                    data: [],
                    borderColor: "#10b981",
                    backgroundColor: gradient5,
                    fill: true,
                    tension: 0.3,
                    pointRadius: 2,
                    spanGaps: true
                },
                {
                    label: "Load Avg (15m)",
                    data: [],
                    borderColor: "#f59e0b",
                    backgroundColor: gradient15,
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
                    max: 2.0,
                    ticks: {
                        color: "#9CA3AF"
                    }
                }
            },
            plugins: {
                legend: { position: "bottom", labels: { color: "#4B5563" } },
                tooltip: {
                    callbacks: {
                        label: ctx => {
                            const val = ctx.parsed.y;
                            const label = ctx.dataset.label;

                            if (val >= 1.5) return `${label}: ${val.toFixed(2)} 🚨 Critical`;
                            if (val >= 1.0) return `${label}: ${val.toFixed(2)} ⚠ Warning`;
                            return `${label}: ${val.toFixed(2)}`;
                        }
                    }
                },
                annotation: {
                    annotations: {
                        warningThreshold: {
                            type: 'line',
                            yMin: 1.0,
                            yMax: 1.0,
                            borderColor: '#facc15',
                            borderWidth: 1,
                            borderDash: [4, 2],
                            label: {
                                enabled: true,
                                content: 'Warn ≥ 1.0',
                                position: 'end',
                                backgroundColor: 'rgba(250, 204, 21, 0.1)',
                                color: '#facc15'
                            }
                        },
                        criticalThreshold: {
                            type: 'line',
                            yMin: 1.5,
                            yMax: 1.5,
                            borderColor: 'red',
                            borderWidth: 1,
                            borderDash: [6, 4],
                            label: {
                                enabled: true,
                                content: 'Critical ≥ 1.5',
                                position: 'end',
                                backgroundColor: 'rgba(255, 0, 0, 0.1)',
                                color: 'red',
                                font: {
                                    weight: 'bold'
                                }
                            }
                        }
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
                legend: { position: "bottom", labels: { color: "#4B5563" } },
                tooltip: {
                    callbacks: {
                        label: ctx => `Memory: ${ctx.parsed.y.toFixed(1)}%`
                    }
                }
            }
        }
    });
    swapLineChart = new Chart(document.getElementById("swapUsageChart"), {
        type: "line",
        data: {
            labels: [],
            datasets: [{
                label: "Swap Usage %",
                data: [],
                borderColor: "#f87171", // red
                backgroundColor: "rgba(248, 113, 113, 0.1)",
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
                legend: { position: "bottom", labels: { color: "#4B5563" } },
                tooltip: {
                    callbacks: {
                        label: ctx => `Swap: ${ctx.parsed.y.toFixed(1)}%`
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


    //  Update metric labels
    document.getElementById("label-load-1").textContent = point.load1.toFixed(2);
    document.getElementById("label-load-5").textContent = point.load5.toFixed(2);
    document.getElementById("label-load-15").textContent = point.load15.toFixed(2);
}

function updateComputeLineChart(chart, value, buffer, labelId) {
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
    if (labelId) {
        const el = document.getElementById(labelId);
        if (el) {
          if (typeof value === "number" && !isNaN(value)) {
            el.textContent = `${value.toFixed(1)}%`;
          } else {
            el.textContent = "--";
          }
        }
      }
      
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

function tryUpdateSwapPercent() {
    const { total, free } = swapPercentMeta;
    if (total == null || free == null || total === 0) return;

    const used = total - free;
    const usedPercent = (used / total) * 100;

    updateComputeLineChart(swapLineChart, usedPercent, swapUsageBuffer, "label-swap-percent");
}
function formatGB(bytes) {
    if (bytes == null || isNaN(bytes)) return "--";
    return (bytes / 1024 / 1024 / 1024).toFixed(1);
}

function updateMemoryAndSwapStats() {

    const totalMem = memoryDonutBuffer.total;
    const freeMem = memoryDonutBuffer.available;
    const usedMem = totalMem && freeMem ? totalMem - freeMem : null;

    const totalSwap = swapPercentMeta.total;
    const freeSwap = swapPercentMeta.free;
    const usedSwap = (totalSwap != null && freeSwap != null) ? totalSwap - freeSwap : null;

    document.getElementById("mem-total").textContent = formatGB(totalMem);
    document.getElementById("mem-used").textContent = formatGB(usedMem);
    document.getElementById("mem-free").textContent = formatGB(freeMem);

    document.getElementById("swap-total").textContent = formatGB(totalSwap);
    document.getElementById("swap-used").textContent = formatGB(usedSwap);
    document.getElementById("swap-free").textContent = formatGB(freeSwap);
}

window.cpuMetricHandler = function (metrics) {
    console.log("📡 Received metrics:", metrics);

    for (const metric of metrics) {
        if (metric.namespace !== "System") continue;

        const key = `${metric.namespace}.${metric.subnamespace}.${metric.name}`.toLowerCase();

        switch (key) {
            case "system.cpu.usage_percent":
                if (metric.dimensions.scope === "total") {
                    updateComputeLineChart(cpuLineChart, metric.value, cpuUsageBuffer, "label-cpu-percent");
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
                if (metric.dimensions?.source === "physical") {
                    updateComputeLineChart(memoryLineChart, metric.value, memoryUsageBuffer, "label-memory-percent");
                    updateMemoryAndSwapStats();
                }
                break;

            case "system.memory.total":
                if (metric.dimensions?.source === "swap") {
                    swapPercentMeta.total = metric.value;
                    console.log("🧪 available metric:", metric);
                    tryUpdateSwapPercent();
                    updateMemoryAndSwapStats();
                } else if (metric.dimensions?.source === "physical") {
                    memoryDonutBuffer.total = metric.value;
                    updateMemoryDonutChart();
                    updateMemoryAndSwapStats();
                }
                break;

            case "system.memory.available":
                if (metric.dimensions?.source === "swap") {
                    swapPercentMeta.free = metric.value;
                    console.log("🧪 available metric:", metric);
                    tryUpdateSwapPercent();
                    updateMemoryAndSwapStats();
                } else if (metric.dimensions?.source === "physical") {
                    memoryDonutBuffer.available = metric.value;
                    updateMemoryDonutChart();
                    updateMemoryAndSwapStats();
                }
                break;

            case "system.memory.used":
                if (metric.dimensions?.source === "swap") {
                    swapPercentMeta.used = metric.value;
                    //tryUpdateSwapPercent(); // optional if you want to use this instead of deriving from total/free
                } else if (metric.dimensions?.source === "physical") {
                    memoryMetrics.used = metric.value;
                    updateMemoryDonutChart();
                    updateMemoryAndSwapStats();
                }
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