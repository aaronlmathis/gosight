import { registerTabInitializer } from "./tabs.js";
import { createApexAreaChart, createApexDonutChart } from './apex_helpers.js';
import { formatGB } from "./format.js";

let cpuUsageChart = null;
let cpuLoadChart = null;
let cpuActivityChart = null;
let memoryUsageChart = null;
let swapUsageChart = null;

let latestSwapUsedPercent = 0;

let prevCpuTimes = null;
const cpuTimeKeys = [
    "user", "system", "idle", "iowait",
    "nice", "irq", "softirq", "steal", "guest", "guest_nice"
];

const cpuUsageSeries = [];
const cpuLoadSeries = { load1: [], load5: [], load15: [] };
const cpuTimeParts = { User: 0, System: 0, Idle: 0, Other: 0 };
const memoryUsageSeries = [];
const swapUsageSeries = [];
const perCoreData = {};


const pendingLoad = { load1: null, load5: null, load15: null };
// Aggregate load data
const pendingLoadAvg = {
    "load_avg_1": null,
    "load_avg_5": null,
    "load_avg_15": null
};

let swapUsedBytes = 0;
let swapTotalBytes = 0;


let memoryTotal = null;
let memoryAvailable = null;



function updateCpuActivityChart(metrics) {
    const current = {};

    for (const m of metrics) {
        if (!m.name.startsWith("time_")) continue;
        const key = m.name.replace("time_", "");
        if (!cpuTimeKeys.includes(key)) continue;
        current[key] = m.value;
    }

    if (!prevCpuTimes) {
        prevCpuTimes = current;
        return;
    }

    const deltas = {};
    let total = 0;

    for (const key of cpuTimeKeys) {
        const prev = prevCpuTimes[key] ?? 0;
        const curr = current[key] ?? 0;
        const delta = Math.max(curr - prev, 0);
        deltas[key] = delta;
        total += delta;
    }

    prevCpuTimes = current;

    if (total === 0 || !cpuActivityChart) return;

    const user = deltas["user"] ?? 0;
    const system = deltas["system"] ?? 0;
    const idle = deltas["idle"] ?? 0;
    const iowait = deltas["iowait"] ?? 0;
    const other = total - user - system - idle - iowait;

    const series = [
        (user / total) * 100,
        (system / total) * 100,
        (idle / total) * 100,
        (iowait / total) * 100,
        (other / total) * 100
    ];

    cpuActivityChart.updateSeries(series);
}


function renderPerCoreGrid() {
    const container = document.getElementById("per-core-grid");
    if (!container) return;

    container.innerHTML = "";

    const sorted = Object.keys(perCoreData).sort((a, b) => {
        return parseInt(a.replace("core", "")) - parseInt(b.replace("core", ""));
    });

    for (const core of sorted) {
        const usage = perCoreData[core].usage?.toFixed(1) ?? "--";
        const clock = perCoreData[core].clock?.toFixed(0) ?? "--";

        const div = document.createElement("div");
        div.className = "p-2 rounded border border-gray-200 dark:border-gray-700 text-center bg-gray-50 dark:bg-gray-800";
        div.innerHTML = `
        <p class="font-semibold text-blue-600 dark:text-blue-400">${core}</p>
        <p class="text-xs text-gray-600 dark:text-gray-400">${usage}% @ ${clock} MHz</p>
      `;
        container.appendChild(div);
    }
}


function updateCpuUsage(timestamp, value) {
    if (!cpuUsageChart || isNaN(value)) return;
    cpuUsageSeries.push({ x: timestamp, y: value });
    if (cpuUsageSeries.length > 60) cpuUsageSeries.shift();
    cpuUsageChart.updateSeries([{ name: "Usage", data: cpuUsageSeries }]);
    const latest = cpuUsageSeries.at(-1)?.y;
    document.getElementById("label-cpu-percent").textContent = latest?.toFixed(1) + "%" || "--";
}

function updateCpuLoad(timestamp, load1, load5, load15) {
    if (!cpuLoadChart) return;
    cpuLoadSeries.load1.push({ x: timestamp, y: load1 });
    cpuLoadSeries.load5.push({ x: timestamp, y: load5 });
    cpuLoadSeries.load15.push({ x: timestamp, y: load15 });

    for (const key of ["load1", "load5", "load15"]) {
        if (cpuLoadSeries[key].length > 60) cpuLoadSeries[key].shift();
    }

    cpuLoadChart.updateSeries([
        { name: "1m", data: cpuLoadSeries.load1 },
        { name: "5m", data: cpuLoadSeries.load5 },
        { name: "15m", data: cpuLoadSeries.load15 }
    ]);
}


function updateMemoryUsage(timestamp, value) {
    if (!memoryUsageChart || isNaN(value)) return;
    memoryUsageSeries.push({ x: timestamp, y: value });
    if (memoryUsageSeries.length > 60) memoryUsageSeries.shift();
    memoryUsageChart.updateSeries([{ name: "Used", data: memoryUsageSeries }]);
}

function updateSwapUsage(timestamp, value) {
    if (!swapUsageChart || isNaN(value)) return;
    swapUsageSeries.push({ x: timestamp, y: value });
    if (swapUsageSeries.length > 60) swapUsageSeries.shift();
    swapUsageChart.updateSeries([{ name: "Used", data: swapUsageSeries }]);

    const el = document.getElementById("label-swap-percent");
    if (el) {
        el.textContent = `${value.toFixed(1)}%`;
    }
}



function createCpuActivityChart(id) {
    const options = {
        chart: {
            type: "donut",
            height: 300,
            animations: {
                enabled: true,
                easing: "easeinout",
                speed: 300
            }
        },
        plotOptions: {
            pie: {
                donut: {
                    size: "75%",  // ⬅️ Bigger hole = thinner ring
                    labels: {
                        show: true,
                        total: {
                            show: true,
                            label: "CPU",
                            fontSize: "14px",
                            fontWeight: 600
                        }
                    }
                }
            }
        },
        labels: ["User", "System", "Idle", "I/O Wait", "Other"],
        series: [0, 0, 0, 0, 0],
        colors: [
            "#3b82f6", // blue
            "#10b981", // green
            "#9ca3af", // gray for idle
            "#f59e0b", // amber
            "#f87171"  // red
        ],
        tooltip: {
            y: {
                formatter: val => `${val.toFixed(1)}%`
            }
        },
        legend: {
            position: "bottom",
            horizontalAlign: "center",
            fontSize: "12px",
            labels: {
                colors: ["#374151"]
            }
        },
        theme: {
            mode: document.documentElement.classList.contains("dark") ? "dark" : "light"
        }
    };

    const chart = new ApexCharts(document.getElementById(id), options);
    chart.render();
    return chart;
}

function updateCpuTimeCounters(metric) {
    if (!metric || typeof metric.value !== "number") return;

    const name = metric.name.replace("time_", "");
    const el = document.getElementById(`cpu-time-${name}`);
    if (el) {
        el.textContent = `${metric.value.toFixed(1)}s`;
    }
}
function updateMemoryAndSwapStats() {
    // Memory
    const totalMem = memoryTotal;
    const freeMem = memoryAvailable;
    const usedMem = (totalMem != null && freeMem != null) ? totalMem - freeMem : null;

    document.getElementById("mem-total").textContent = formatGB(totalMem);
    document.getElementById("mem-used").textContent = formatGB(usedMem);
    document.getElementById("mem-free").textContent = formatGB(freeMem);

    // Swap: only update if both values are present
    if (typeof swapUsedBytes === "number" && typeof swapTotalBytes === "number" && swapTotalBytes > 0) {
        const usedSwap = swapUsedBytes;
        const totalSwap = swapTotalBytes;
        const freeSwap = totalSwap - usedSwap;

        document.getElementById("swap-total").textContent = formatGB(totalSwap);
        document.getElementById("swap-used").textContent = formatGB(usedSwap);
        document.getElementById("swap-free").textContent = formatGB(freeSwap);
    }
}
window.cpuMetricHandler = function (metrics) {
    for (const metric of metrics) {

        const now = Date.now();
        const key = `${metric.namespace}.${metric.subnamespace}.${metric.name}`.toLowerCase();
        switch (key) {
            case "system.cpu.usage_percent":
                if (metric.dimensions.scope === "total") {
                    updateCpuUsage(now, metric.value);
                } else if (metric.dimensions.scope === "per_core") {
                    updatePerCoreMetrics(metric);
                }
                break;

            case "system.cpu.load_avg_1":
                pendingLoad.load1 = metric.value;
                break;
            case "system.cpu.load_avg_5":
                pendingLoad.load5 = metric.value;
                break;
            case "system.cpu.load_avg_15":
                pendingLoad.load15 = metric.value;
                break;

            case "system.cpu.time_user":
            case "system.cpu.time_system":
            case "system.cpu.time_idle":
            case "system.cpu.time_nice":
            case "system.cpu.time_iowait":
            case "system.cpu.time_irq":
            case "system.cpu.time_softirq":
            case "system.cpu.time_steal":
            case "system.cpu.time_guest":
            case "system.cpu.time_guest_nice":
                updateCpuActivityChart(metrics);
                updateCpuTimeCounters(metric)
                break;

            case "system.memory.used_percent":
                updateMemoryUsage(now, metric.value);
                document.getElementById("label-memory-percent").textContent =
                    typeof metric.value === "number" ? `${metric.value.toFixed(1)}%` : "--";
                break;

            case "system.memory.total":
                memoryTotal = metric.value;
                updateMemoryAndSwapStats();
                break;

            case "system.memory.available":
                memoryAvailable = metric.value;
                updateMemoryAndSwapStats();
                break;


            case "system.memory.swap_total":
                swapTotalBytes = metric.value;
                tryUpdateSwapPercentFromUsed();
                console.log("[Swap Debug]", {
                    used: swapUsedBytes,
                    total: swapTotalBytes
                });
                updateMemoryAndSwapStats();
                break;

            case "system.memory.swap_used":
                console.log("🔥 Received swap_used metric", metric.value);
                swapUsedBytes = metric.value;
                tryUpdateSwapPercentFromUsed();
                console.log("[Swap Debug]", {
                    used: swapUsedBytes,
                    total: swapTotalBytes
                });
                updateMemoryAndSwapStats();
                break;

            case "system.cpu.count_logical":
                document.getElementById("cpu-threads").textContent = metric.value;
                break;

            case "system.cpu.count_physical":
                document.getElementById("cpu-cores").textContent = metric.value;
                break;

            case "system.cpu.clock_mhz":
                if (metric.dimensions?.core === "core0") {
                    document.getElementById("cpu-base-clock").textContent = `${metric.value.toFixed(0)} MHz`;
                    document.getElementById("cpu-vendor").textContent = metric.dimensions.vendor || "--";
                    document.getElementById("cpu-model").textContent = metric.dimensions.model || "--";
                    document.getElementById("cpu-family").textContent = metric.dimensions.family || "--";
                    document.getElementById("cpu-stepping").textContent = metric.dimensions.stepping || "--";
                    document.getElementById("cpu-cache").textContent = metric.dimensions.cache ? `${parseInt(metric.dimensions.cache).toLocaleString()} KB` : "--";
                    document.getElementById("cpu-physical").textContent = metric.dimensions.physical === "true" ? "Yes" : "No";
                }
                break;
        }
    }

    // If all load_avg values ready, flush
    if (pendingLoad.load1 != null && pendingLoad.load5 != null && pendingLoad.load15 != null) {
        const now = Date.now();
        updateCpuLoad(now, pendingLoad.load1, pendingLoad.load5, pendingLoad.load15);
        pendingLoad.load1 = pendingLoad.load5 = pendingLoad.load15 = null;
    }
};

function updatePerCoreMetrics(metric) {
    const core = metric.dimensions?.core;
    if (!core) return;

    if (!perCoreData[core]) {
        perCoreData[core] = {};
    }

    if (metric.name === "usage_percent" && metric.dimensions.scope === "per_core") {
        perCoreData[core].usage = metric.value;
    } else if (metric.name === "clock_mhz") {
        perCoreData[core].clock = metric.value;
    }
    renderPerCoreGrid();
    //renderPerCoreTable();
}

function tryUpdateSwapPercentFromUsed() {
    if (typeof swapUsedBytes !== "number" || typeof swapTotalBytes !== "number" || swapTotalBytes === 0) {
        return;
    }

    const percent = (swapUsedBytes / swapTotalBytes) * 100;
    const now = Date.now();
    updateSwapUsage(now, percent); // This updates chart and label
}

window.addEventListener("metrics", ({ detail: payload }) => {
    if (!payload?.metrics || !payload.meta) return;

    // Host metrics
    if (payload.meta.endpoint_id?.startsWith("host-") &&
        payload.meta.host_id === window.hostID) {
        window.cpuMetricHandler(payload.metrics);
    }

    // Container metrics (podman/docker) tied to this host
    if (payload.meta.endpoint_id?.startsWith("ctr-") &&
        payload.meta.host_id === window.hostID &&
        payload.metrics.some(m => m.name === "cpu_percent")) {

    }
});

function waitForElement(id, callback) {
    const el = document.getElementById(id);
    if (el && el.offsetParent !== null) {
        callback();
    } else {
        setTimeout(() => waitForElement(id, callback), 100);
    }
}

function initComputeCharts() {
    waitForElement("cpuUsageChart", () => {
        cpuUsageChart = createApexAreaChart("cpuUsageChart", "CPU Usage %", ["Usage"]);
    });

    waitForElement("cpuLoadChart", () => {
        cpuLoadChart = createApexAreaChart("cpuLoadChart", "CPU Load Average", ["1m", "5m", "15m"], false);
        cpuLoadChart.updateOptions({

            chart: {
                type: "area",
                height: 280,
                toolbar: { show: false },
                animations: {
                    enabled: true,
                    easing: "easeinout",
                    speed: 400
                }
            },
            title: {
                text: undefined,
                show: false
            },
            stroke: {
                curve: "smooth",
                width: 3
            },
            fill: {

                type: "gradient",
                gradient: {
                    shadeIntensity: 1,
                    opacityFrom: 0.5,  // ↑ bump this up (default is 0.35–0.4)
                    opacityTo: 0.2,     // ↑ increase to make the bottom still visible
                    stops: [0, 90, 100] // optionally adjust gradient curve
                }

            },
            xaxis: {
                type: "datetime",
                labels: {
                    format: "HH:mm:ss"
                }
            },
            yaxis: {
                min: 0,
                max: 4, // or dynamic
                tickAmount: 4,
                labels: {
                    formatter: val => val.toFixed(2)
                },
                title: {
                    text: "Load Avg"
                }
            },
            tooltip: {
                x: { format: "HH:mm:ss" },
                y: {
                    formatter: val => val.toFixed(2)
                }
            },
            legend: {
                position: "bottom",
                fontSize: "12px"
            },
            colors: ["#3b82f6", "#10b981", "#f59e0b"], // blue, green, amber
            annotations: {
                yaxis: [
                    {
                        y: 1.0,
                        borderColor: "#facc15",
                        label: {
                            text: "Warn ≥ 1.0",
                            style: { background: "#facc15", color: "#000" }
                        }
                    },
                    {
                        y: 1.5,
                        borderColor: "#f87171",
                        label: {
                            text: "High ≥ 1.5",
                            style: { background: "#f87171", color: "#fff" }
                        }
                    }
                ]
            }
        })
    });


    waitForElement("cpuActivityDonutChart", () => {
        cpuActivityChart = createCpuActivityChart("cpuActivityDonutChart");
    });

    waitForElement("memoryUsageChart", () => {
        memoryUsageChart = createApexAreaChart("memoryUsageChart", "Memory Usage %", ["Used"]);
    });

    waitForElement("swapUsageChart", () => {
        swapUsageChart = createApexAreaChart("swapUsageChart", "Swap Usage %", ["Used"]);
    });
}

function initComputeTab() {
    initComputeCharts();
    // Initialize any other elements or event listeners needed for the compute tab
}

registerTabInitializer("compute", initComputeTab);
