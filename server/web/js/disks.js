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
// Apex Radial Chart
// This chart shows the disk usage as a radial chart
//
let diskRadialChart = null;



export function createDiskUsageRadialChart() {
    //console.log("Creating Disk Usage Radial Chart");

    const options = {
        chart: {
            type: "radialBar",
            height: 400,
            toolbar: { show: false }
        },
        plotOptions: {
            radialBar: {
                offsetY: 0,
                hollow: {
                    size: "40%",
                    background: "transparent"
                },
                track: {
                    background: getGridColor()
                },
                dataLabels: {
                    name: {
                        show: true,
                        fontSize: "14px",
                        color: getTextColor()
                    },
                    value: {
                        show: true,
                        fontSize: "20px",
                        fontWeight: 600,
                        color: getTextColor(),
                        formatter: (val) => `${val}%`
                    },
                    total: {
                        show: true,
                        label: "Avg Used",
                        fontSize: "16px",
                        fontWeight: 500,
                        color: getTextColor(),
                        formatter: () => "0%" // updated in render
                    }
                }
            }
        },
        stroke: {
            lineCap: "round"
        },
        labels: [],
        series: [],
        colors: ["#3b82f6", "#10b981", "#f59e0b", "#ef4444", "#8b5cf6", "#14b8a6"],
        legend: {
            show: true,
            position: "bottom",
            fontSize: "12px",
            labels: {
                colors: getTextColor()
            },
            markers: {
                width: 10,
                height: 10,
                radius: 4
            }
        },
        tooltip: {
            y: {
                formatter: val => `${val.toFixed(1)}% used`
            }
        },
        theme: {
            mode: document.documentElement.classList.contains("dark") ? "dark" : "light"
        }
    };

    const el = document.querySelector("#diskRadialChart");
    if (!el) return;
    diskRadialChart = new ApexCharts(el, options);
    diskRadialChart.render();
}


export function renderDiskUsageRadialChart(usageByMount) {
    //console.log("renderDiskUsageRadialChart", usageByMount);

    const labels = [];
    const series = [];

    for (const mount in usageByMount) {
        const { total, used } = usageByMount[mount];
        if (total > 0) {
            const percent = (used / total) * 100;
            labels.push(mount.length > 14 ? mount.slice(0, 12) + "…" : mount);
            series.push(Math.round(percent));
        }
    }

    const avg = series.length > 0
        ? (series.reduce((a, b) => a + b, 0) / series.length).toFixed(1)
        : "0.0";

    if (diskRadialChart) {
        diskRadialChart.updateOptions({
            labels,
            series,
            plotOptions: {
                radialBar: {
                    dataLabels: {
                        total: {
                            formatter: () => `${avg}%`
                        }
                    }
                }
            }
        });
    }
}


function getTextColor() {
    return document.documentElement.classList.contains("dark") ? "#d1d5db" : "#374151";
}

function getGridColor() {
    return document.documentElement.classList.contains("dark") ? "#374151" : "#e5e7eb";
}

//
// -------------------------------------------------
// END Apex Radialt Chart
// -------------------------------------------------

//
// ------------------------------------------------
// Most active mountpoint
// This chart shows the most active mountpoints


function formatIO(val) {
    if (val >= 1024 ** 3) return (val / 1024 ** 3).toFixed(2) + " GB/s";
    if (val >= 1024 ** 2) return (val / 1024 ** 2).toFixed(1) + " MB/s";
    if (val >= 1024) return (val / 1024).toFixed(1) + " KB/s";
    return val.toFixed(0) + " B/s";
}
let activeMountChart = null;

export function createActiveMountChart() {
    const el = document.querySelector("#activeMountChart");
    if (!el) return;

    activeMountChart = new ApexCharts(el, {
        chart: {
            type: "bar",
            height: 260,
            stacked: true,
            animations: { enabled: true, easing: "easeinout" },
            toolbar: { show: false }
        },
        plotOptions: {
            bar: {
                horizontal: true,
                borderRadius: 4,
                barHeight: "70%",
            },
        },
        dataLabels: {
            enabled: false
        },
        xaxis: {
            categories: [],
            labels: {
                style: { colors: getTextColor() },
                formatter: val => formatIO(parseFloat(val))
            },
            title: {
                text: "I/O Rate",
                style: {
                    color: getTextColor(),
                    fontSize: "12px"
                }
            }
        },
        tooltip: {
            shared: true,
            intersect: false,
            y: {
                formatter: val => formatIO(val)
            }
        },
        colors: ["#3b82f6", "#f59e0b"], // blue: read, amber: write
        grid: {
            borderColor: getGridColor()
        },
        theme: {
            mode: document.documentElement.classList.contains("dark") ? "dark" : "light"
        },
        legend: {
            position: "bottom",
            labels: {
                colors: getTextColor()
            }
        },
        series: [
            {
                name: "Read",
                data: []
            },
            {
                name: "Write",
                data: []
            }
        ],
    });

    activeMountChart.render();
}
export function renderActiveMountChart(ioMetricsByMount) {
    const top = Object.entries(ioMetricsByMount)
        .map(([mount, { read_bytes = 0, write_bytes = 0 }]) => ({
            mount,
            read: read_bytes,
            write: write_bytes
        }))
        .sort((a, b) => (b.read + b.write) - (a.read + a.write))
        .slice(0, 6);

    const categories = top.map(m => m.mount);
    const readSeries = top.map(m => m.read);
    const writeSeries = top.map(m => m.write);

    activeMountChart?.updateOptions({
        xaxis: { categories }
    });

    activeMountChart?.updateSeries([
        { name: "Read", data: readSeries },
        { name: "Write", data: writeSeries }
    ]);
}
//
// ------------------------------------------------
// Disk Inode Bar Chart
// This chart shows the inode usage as a bar chart
// -------------------------------------------------
let inodeUsageChart = null;

function createInodeUsageBarChart() {
    const el = document.querySelector("#inodeUsageBarChart");
    if (!el) return;

    inodeUsageChart = new ApexCharts(el, {
        chart: {
            type: "bar",
            height: 250,
            toolbar: { show: false }
        },
        plotOptions: {
            bar: {
                horizontal: true,
                borderRadius: 4
            }
        },
        dataLabels: {
            enabled: false
        },

        xaxis: {
            categories: [],
            labels: {
                formatter: (val) => `${parseFloat(val).toFixed(2)}%`,
                style: { colors: getTextColor() }
            },
            title: {
                text: "Inodes Used",
                style: {
                    color: getTextColor(),
                    fontSize: "12px"
                }
            },
            max: 100
        },
        yaxis: {
            labels: {
                style: { colors: getTextColor() }
            }
        },
        series: [{
            name: "Inodes Used %",
            data: []
        }],
        colors: ["#6366f1"],
        theme: {
            mode: document.documentElement.classList.contains("dark") ? "dark" : "light"
        },
        grid: {
            borderColor: getGridColor()
        }
    });

    inodeUsageChart.render();
}

function renderInodeBar(usageByMount) {
    const labels = [];
    const values = [];

    for (const mount in usageByMount) {
        const usage = usageByMount[mount];
        if (usage.inodes_used_percent !== undefined) {
            labels.push(mount);
            values.push(usage.inodes_used_percent.toFixed(2));
        }
    }

    inodeUsageChart?.updateOptions({ xaxis: { categories: labels } });
    inodeUsageChart?.updateSeries([{ name: "Inodes Used %", data: values }]);
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




export function createDiskIOCharts() {
    const iopsEl = document.querySelector("#diskIopsChart");
    const tpEl = document.querySelector("#diskThroughputChart");

    if (iopsEl) {
        diskIopsChart = new ApexCharts(iopsEl, {
            chart: {
                type: "line",
                height: 250,
                animations: { enabled: true },
                toolbar: { show: false }
            },
            stroke: { curve: "smooth", width: 2 },
            dataLabels: { enabled: false },
            colors: ["#3b82f6", "#10b981"],
            series: [
                { name: "Read Count", data: [] },
                { name: "Write Count", data: [] }
            ],
            xaxis: {
                categories: [],
                labels: { style: { colors: getTextColor() } }
            },
            yaxis: {
                title: { text: "Ops/sec", style: { color: getTextColor() } },
                labels: { style: { colors: getTextColor() } }
            },
            grid: { borderColor: getGridColor() },
            theme: { mode: document.documentElement.classList.contains("dark") ? "dark" : "light" },
            legend: { labels: { colors: getTextColor() } }
        });
        diskIopsChart.render();
    }

    if (tpEl) {
        diskThroughputChart = new ApexCharts(tpEl, {
            chart: {
                type: "line",
                height: 250,
                animations: { enabled: true },
                toolbar: { show: false }
            },
            stroke: { curve: "smooth", width: 2 },
            dataLabels: { enabled: false },
            colors: ["#f59e0b", "#ef4444"],
            series: [
                { name: "Read Bytes", data: [] },
                { name: "Write Bytes", data: [] }
            ],
            xaxis: {
                categories: [],
                labels: { style: { colors: getTextColor() } }
            },
            yaxis: {
                title: { text: "Throughput", style: { color: getTextColor() } },
                labels: {
                    style: { colors: getTextColor() },
                    formatter: val => formatBytes(val)
                }
            },
            tooltip: {
                y: {
                    formatter: val => formatBytes(val)
                }
            },
            grid: { borderColor: getGridColor() },
            theme: { mode: document.documentElement.classList.contains("dark") ? "dark" : "light" },
            legend: { labels: { colors: getTextColor() } }
        });
        diskThroughputChart.render();
        setTimeout(() => {
            diskIopsChart?.render();
            diskThroughputChart?.render();
        }, 300);
    }
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

    if (diskIopsChart) {
        diskIopsChart.updateOptions({
            xaxis: { categories: cache.timestamps }
        });
        diskIopsChart.updateSeries([
            { name: "Read Count", data: cache.readCount },
            { name: "Write Count", data: cache.writeCount }
        ]);
    }

    if (diskThroughputChart) {
        diskThroughputChart.updateOptions({
            xaxis: { categories: cache.timestamps }
        });
        diskThroughputChart.updateSeries([
            { name: "Read Bytes", data: cache.readBytes },
            { name: "Write Bytes", data: cache.writeBytes }
        ]);
    }
    //console.log("Updating Apex IOPS chart with", cache.readCount, cache.writeCount);
    //console.log("Updating Apex Throughput chart with", cache.readBytes, cache.writeBytes);
}

function resetDiskIOCharts() {
    diskIopsChart?.updateOptions({ xaxis: { categories: [] } });
    diskIopsChart?.updateSeries([
        { name: "Read Count", data: [] },
        { name: "Write Count", data: [] }
    ]);

    diskThroughputChart?.updateOptions({ xaxis: { categories: [] } });
    diskThroughputChart?.updateSeries([
        { name: "Read Bytes", data: [] },
        { name: "Write Bytes", data: [] }
    ]);
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
    console.log(usageByMount)

    renderMountpointTable(usageByMount);
    renderInodeBar(usageByMount);
    renderDiskDonut(usageByMount);
    renderDiskUsageRadialChart(usageByMount);
    renderIODropdown(ioByDevice, usageByMount);
    renderTopMountUsage(usageByMount);
    renderMiniMountTable(usageByMount);
    renderTopMountsBar(usageByMount);
    renderMiniMountStats(usageByMount);
    renderActiveMountChart(ioByDevice)
}

// INIT Disk TAB
function initDiskTab() {
    createDiskUsageDonut();
    createDiskUsageRadialChart();
    createInodeUsageBarChart();
    createDiskIOCharts();

    createActiveMountChart()

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