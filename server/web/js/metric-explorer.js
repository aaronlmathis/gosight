// /js/metric-explorer.js
import { gosightFetch } from "./api.js";

let allMetrics = []; // Full list of { label, namespace, subnamespace, name }
let tagSuggestions = {}; // key:tag, value:Set of values
const chartSlots = []; // Array of { id, metrics:[], filters:{}, chart:null }
let activeSlotId = null;

// DOM references
const metricPanelsEl = document.getElementById("metric-panels");
const metricInput = document.getElementById("metric-search");
const suggestionsEl = document.getElementById("metric-suggestions");
const selectedEl = document.getElementById("selected-metrics");
const tagInput = document.getElementById("from-filter-input");
const tagSuggestionsEl = document.getElementById("from-suggestions");
const selectedFiltersEl = document.getElementById("from-selected");

window.addEventListener("DOMContentLoaded", async () => {
    initSlots(3); // Initialize 3 slots
    await loadMetrics();
    await loadEndpointTagSuggestions();
    setupMetricSearch();
    setupFilterSearch();
    setupControls();
});

// Initialize N slots
function initSlots(count) {
    for (let i = 0; i < count; i++) {
        const slot = document.createElement("div");
        slot.id = `chart-slot-${i}`;
        slot.className = "rounded border-2 border-dashed border-gray-100 dark:border-gray-700 p-4 bg-white dark:bg-gray-900 text-gray-400 cursor-pointer flex items-center justify-center h-[250px]";
        slot.innerHTML = `<span class="text-sm">➕ New Chart</span>`;
        slot.addEventListener("click", () => setActiveSlot(slot.id));
        metricPanelsEl.appendChild(slot);

        chartSlots.push({
            id: slot.id,
            metrics: [],
            filters: {},
            availableDimensions: [],
            period: "5m",
            graphType: "area",
            groupBy: "",
            aggregate: "",
            chart: null,
            shouldReset: false
        });


        if (i === 0) setActiveSlot(slot.id); // Select first slot by default
    }
}

// Set active slot
function setActiveSlot(slotId) {
    activeSlotId = slotId;
    document.querySelectorAll("#metric-panels > div").forEach(div => {
        div.classList.remove("border-blue-400");
        div.classList.add("border-gray-200", "dark:border-gray-700");
    });
    const activeSlot = document.getElementById(slotId);
    if (activeSlot) {
        activeSlot.classList.remove("border-gray-200", "dark:border-gray-700");
        activeSlot.classList.add("border-blue-400");
    }
    renderSelectedMetrics();
    renderSelectedFilters();
    refreshControlsForSlot();
    refreshGroupByOptions();
}
// Refresh controls for the active slot
function refreshControlsForSlot() {
    const panel = chartSlots.find(s => s.id === activeSlotId);
    if (!panel) return;

    document.getElementById("period").value = panel.period || "5m";
    document.getElementById("graph-type").value = panel.graphType || "area";
    document.getElementById("group-by").value = panel.groupBy || "";
    document.getElementById("aggregate").value = panel.aggregate || "";
}

// Load all metrics
async function loadMetrics() {
    const namespaces = await (await gosightFetch("/api/v1/")).json();
    for (const ns of namespaces) {
        const subs = await (await gosightFetch(`/api/v1/${ns}`)).json();
        for (const sub of subs) {
            const metrics = await (await gosightFetch(`/api/v1/${ns}/${sub}`)).json();
            for (const m of metrics) {
                allMetrics.push({
                    label: m,
                    namespace: ns,
                    subnamespace: sub,
                    name: m.split(".").pop(),
                });
            }
        }
    }
}

// Metric search box
function setupMetricSearch() {
    metricInput.addEventListener("input", () => {
        const q = metricInput.value.trim().toLowerCase();
        suggestionsEl.innerHTML = "";
        if (q.length < 2) {
            suggestionsEl.classList.add("hidden");
            return;
        }

        const filtered = allMetrics.filter(m => m.label.includes(q)).slice(0, 10);
        if (filtered.length === 0) {
            suggestionsEl.classList.add("hidden");
            return;
        }

        let currentGroup = "";

        for (const metric of filtered) {
            const group = `${metric.namespace}.${metric.subnamespace}`;
            if (group !== currentGroup) {
                const groupHeader = document.createElement("div");
                groupHeader.className = "text-xs font-semibold mt-2 text-gray-500 dark:text-gray-400";
                groupHeader.textContent = group;
                suggestionsEl.appendChild(groupHeader);
                currentGroup = group;
            }

            const item = document.createElement("div");
            item.className = "cursor-pointer px-2 py-1 hover:bg-gray-100 dark:hover:bg-gray-700 rounded";
            item.textContent = metric.label;
            item.addEventListener("click", () => {
                if (metric.label === "system.memory.used_percent") {
                    showMemoryTypePicker(metric, e);
                } else {
                    addSelectedMetric(metric);
                    metricInput.value = "";
                    suggestionsEl.classList.add("hidden");
                }
            });

            suggestionsEl.appendChild(item);
        }
        suggestionsEl.classList.remove("hidden");
    });
}

// Add selected metric to active slot
async function addSelectedMetric(metric) {
    const panel = chartSlots.find(s => s.id === activeSlotId);
    if (!panel) return;

    panel.metrics.push(metric);
    const [namespace, subnamespace, ...metricParts] = metric.label.split(".");
    const shortMetric = metricParts.join(".");

    try {
        const res = await gosightFetch(`/api/v1/${namespace}/${subnamespace}/${shortMetric}/dimensions`);
        const dims = await res.json();
        if (Array.isArray(dims)) {
            panel.availableDimensions = dims;
        } else {
            console.warn("Unexpected dimensions response", dims);
        }
    } catch (err) {
        console.error("Failed to fetch dimensions for", metric.label, err);
    }
    renderSelectedMetrics();
    refreshGroupByOptions();
    loadData();
}

function refreshGroupByOptions() {
    const panel = chartSlots.find(s => s.id === activeSlotId);
    if (!panel) return;

    const dropdown = document.getElementById("group-by");
    if (!dropdown) return;

    dropdown.innerHTML = `<option value="">(none)</option>`;

    for (const dim of panel.availableDimensions) {
        const opt = document.createElement("option");
        opt.value = dim;
        opt.textContent = dim;
        dropdown.appendChild(opt);
    }

    // Set selected to current saved value
    dropdown.value = panel.groupBy || "";
}


// Render selected metrics
function renderSelectedMetrics() {
    selectedEl.innerHTML = "";
    const panel = chartSlots.find(s => s.id === activeSlotId);
    if (!panel) return;
    if (panel.metrics.length === 0) {
        resetSlot(panel);

        renderSelectedFilters();
        return;
    }
    for (const metric of panel.metrics) {
        const pill = document.createElement("span");
        pill.className = "inline-flex items-center bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200 px-2 py-1 rounded-full text-xs mr-1 mb-1";
        pill.textContent = metric.label;

        const remove = document.createElement("button");
        remove.className = "ml-2 text-blue-500 hover:text-red-500 text-sm";
        remove.innerHTML = "&times;";
        remove.addEventListener("click", () => {
            panel.metrics = panel.metrics.filter(m => m.label !== metric.label);

            renderSelectedMetrics(); // <--- always re-render pills FIRST
            renderSelectedFilters(); // optional: if filters changed

            if (panel.metrics.length === 0) {
                resetSlot(panel);
                return; // no need to loadData
            }

            loadData();
        });
        pill.appendChild(remove);
        selectedEl.appendChild(pill);
    }
}

// Setup "From" tag filter search
function setupFilterSearch() {
    tagInput.addEventListener("input", () => {
        const q = tagInput.value.trim().toLowerCase();
        tagSuggestionsEl.innerHTML = "";

        if (q.length < 2) {
            tagSuggestionsEl.classList.add("hidden");
            return;
        }

        const allEntries = [];

        for (const [key, values] of Object.entries(tagSuggestions)) {
            for (const val of values) {
                if (!val) continue;
                allEntries.push(`${key}:${val}`);
            }
        }

        const matches = allEntries.filter(entry => entry.toLowerCase().includes(q));

        if (matches.length === 0) {
            tagSuggestionsEl.classList.add("hidden");
            return;
        }

        tagSuggestionsEl.classList.remove("hidden");
        tagSuggestionsEl.style.maxHeight = "240px";
        tagSuggestionsEl.style.overflowY = "auto";

        matches.slice(0, 10).forEach(full => {
            const item = document.createElement("div");
            item.className = "px-3 py-2 hover:bg-gray-100 dark:hover:bg-gray-700 cursor-pointer text-sm whitespace-nowrap";
            item.textContent = full;
            item.onclick = () => {
                const [key, value] = full.split(":");
                addSelectedFilter(key, value);
                tagInput.value = "";
                tagSuggestionsEl.classList.add("hidden");
            };
            tagSuggestionsEl.appendChild(item);
        });
    });
}



// Add selected filter to active slot
function addSelectedFilter(key, val) {
    const panel = chartSlots.find(s => s.id === activeSlotId);
    if (!panel) return;

    panel.filters[`${key}:${val}`] = true;
    renderSelectedFilters();
    loadData();
}

// Render selected filters
function renderSelectedFilters() {
    selectedFiltersEl.innerHTML = "";
    const panel = chartSlots.find(s => s.id === activeSlotId);
    if (!panel) return;

    for (const id of Object.keys(panel.filters)) {
        const chip = document.createElement("span");
        chip.className = "inline-flex items-center bg-gray-200 dark:bg-gray-800 text-xs rounded px-2 py-1 mr-1 mb-1";
        chip.innerHTML = `${id} <button class="ml-1 text-red-500 hover:text-red-700">&times;</button>`;

        chip.querySelector("button").onclick = () => {
            delete panel.filters[id];
            renderSelectedFilters();
            loadData();
        };

        selectedFiltersEl.appendChild(chip);
    }
}

// Load metrics and render charts
async function loadData() {
    for (const panel of chartSlots) {

        const slotEl = document.getElementById(panel.id);
        if (!slotEl) continue;
        if (panel.shouldReset || panel.metrics.length === 0) {
            resetSlot(panel);
            continue;
        }


        // Destroy old chart before reload
        if (panel.chart) {
            panel.chart.destroy();
            panel.chart = null;
        }

        const now = new Date();
        const startDate = new Date(now);
        const timeRange = panel.period || "5m";

        if (timeRange.endsWith("m")) startDate.setMinutes(now.getMinutes() - parseInt(timeRange));
        else if (timeRange.endsWith("h")) startDate.setHours(now.getHours() - parseInt(timeRange));
        else if (timeRange.endsWith("d")) startDate.setDate(now.getDate() - parseInt(timeRange));

        const start = startDate.toISOString();
        const end = now.toISOString();

        let step = "15s"; // default step

        if (timeRange.endsWith("m")) {
            const minutes = parseInt(timeRange);
            if (minutes <= 15) step = "5s";
            else if (minutes <= 30) step = "15s";
            else if (minutes <= 60) step = "30s";
            else step = "60s";
        } else if (timeRange.endsWith("h")) {
            const hours = parseInt(timeRange);
            if (hours <= 6) step = "2m";
            else if (hours <= 12) step = "5m";
            else step = "10m";
        } else if (timeRange.endsWith("d")) {
            const days = parseInt(timeRange);
            if (days <= 1) step = "10m";
            else step = "30m";
        }

        const allSeries = [];

        for (const metric of panel.metrics) {
            const tagFilter = Object.keys(panel.filters)
                .map(f => f.replace(":", "="))
                .join(",");

            const url = `/api/v1/query?metric=${encodeURIComponent(metric.label)}&start=${encodeURIComponent(start)}&end=${encodeURIComponent(end)}&step=${encodeURIComponent(step)}` +
                (tagFilter ? `&tags=${encodeURIComponent(tagFilter)}` : "");
            console.log(`[DEBUG] Querying metric: ${metric.label} with step: ${step} using url: ${url}`);
            const res = await gosightFetch(url);
            const data = await res.json();
            if (!data || !Array.isArray(data)) continue;

            // Merge tagSuggestions here if needed
            for (const point of data) {
                if (!point.tags) continue;
                for (const [key, value] of Object.entries(point.tags)) {
                    const k = key.toLowerCase();
                    if (!value) continue;
                    if (!tagSuggestions[k]) tagSuggestions[k] = new Set();
                    tagSuggestions[k].add(value);
                }
            }

            // Build series for each metric
            const metricSeries = buildSeries(data, panel);
            allSeries.push(...metricSeries);
        }

        // Now render chart ONCE with combined allSeries
        if (allSeries.length > 0) {
            renderChartPanel(slotEl, allSeries, panel.metrics.map(m => m.name).join(", "), panel.graphType);
        }
    }
}

// Fetch data and render chart
async function fetchAndRenderSingleChart(url, panelEl, metric, panel) {


    try {

        const res = await gosightFetch(url);
        const data = await res.json();
        if (!data || !Array.isArray(data)) return;



        const series = buildSeries(data, panel);

        if (!panelEl.chart) {
            renderChartPanel(panelEl, series, metric.label, panel.graphType);
        } else {
            panelEl.chart.updateSeries(series);
        }
    } catch (err) {
        console.error("Failed to load metric", metric.label, err);
    }
}

// Build chart series
function buildSeries(dataArray, panel) {
    if (!panel) return [];

    const groupKey = panel.groupBy || "";

    const groups = {};

    for (const d of dataArray) {
        let id = "unknown";

        if (groupKey && d.tags?.[groupKey]) {
            id = d.tags[groupKey];
        } else {
            id = d.tags?.endpoint_id || d.tags?.instance || d.tags?.hostname || "unknown";
        }

        if (!groups[id]) groups[id] = [];
        groups[id].push([d.timestamp, d.value]);
    }
    let totalPoints = 0;
    for (const points of Object.values(groups)) {
        totalPoints += points.length;
    }
    console.log(`[DEBUG] Total points for this panel: ${totalPoints}`);
    console.log(`Grouping by: ${groupKey || "(default endpoint_id)"}`);
    for (const [key, points] of Object.entries(groups)) {
        console.log(`Group: ${key} → ${points.length} points`);
    }

    return Object.entries(groups).map(([name, data]) => ({ name, data }));
}

// Render chart
function renderChartPanel(panelEl, series, title, graphType) {
    panelEl.innerHTML = "";
    panelEl.className = "relative rounded border p-2 bg-white dark:bg-gray-900"; // <-- add relative positioning for badge

    const manySeries = series.length > 5;
    const manyPoints = series.reduce((sum, s) => sum + s.data.length, 0) > 500;
    const optimized = manySeries || manyPoints; // 

    // If in optimized mode, add a badge
    if (optimized) {
        const badge = document.createElement("div");
        badge.className = "absolute top-2 right-[180px] text-[10px] px-2 py-1 rounded bg-yellow-400 text-black font-semibold shadow";
        badge.textContent = "⚡ Optimized View";
        panelEl.appendChild(badge);
    }

    //  Handle stacked-area correctly
    let chartType = graphType;
    let stacked = false;

    if (graphType === "stacked-area") {
        chartType = "area";
        stacked = true;
    }

    const baseChartOptions = {
        chart: {
            type: chartType || "area",
            height: 250,
            zoom: {
                type: "x",
                enabled: true,
                autoScaleYaxis: true
            },
            toolbar: {
                autoSelected: "zoom"
            },
            stacked: stacked,
            animations: {
                enabled: !optimized
            }
        },
        stroke: {
            curve: manySeries ? "straight" : "smooth",
            width: 2
        },
        fill: {
            type: (chartType === "area") ? "gradient" : "solid",
            gradient: (chartType === "area") ? {
                shadeIntensity: 1,
                opacityFrom: 0.4,
                opacityTo: 0,
                stops: [0, 90, 100]
            } : undefined
        },
        dataLabels: {
            enabled: false
        },
        markers: {
            size: 0
        },
        title: {
            text: title,
            align: "left",
            style: {
                fontSize: "14px",
                fontWeight: 600,
                color: "#263238"
            }
        },
        xaxis: {
            type: "datetime",
            labels: {
                datetimeFormatter: {
                    month: "MMM 'yy",
                    day: "dd MMM",
                    hour: "HH:mm",
                    minute: "HH:mm"
                }
            }
        },
        yaxis: {
            labels: {
                formatter: val => val.toFixed(2)
            },
            title: {
                text: "Value"
            }
        },
        tooltip: {
            shared: true,
            intersect: false,
            x: { format: "MMM dd HH:mm" },
            y: { formatter: val => val.toFixed(2) }
        },
        series: series
    };

    if (graphType === "bar") {
        baseChartOptions.chart.type = "bar";
        baseChartOptions.stroke = { width: 0 };
        baseChartOptions.fill = { type: "solid" };
        baseChartOptions.plotOptions = {
            bar: {
                horizontal: false,
                columnWidth: "50%",
                endingShape: "rounded"
            }
        };
    }

    const chart = new ApexCharts(panelEl, baseChartOptions);
    chart.render();
    panelEl.chart = chart;
}



// Load endpoint tags
async function loadEndpointTagSuggestions() {
    const hosts = await gosightFetch("/api/v1/endpoints/hosts").then(res => res.json());
    const containers = await gosightFetch("/api/v1/endpoints/containers").then(res => res.json());
    const endpoints = [...hosts, ...containers];

    tagSuggestions = {};
    const blacklist = new Set(["agent_start_time", "_cmdline", "_uid", "_exe"]);

    for (const ep of endpoints) {
        const tags = { ...ep.labels };
        if (ep.Hostname) tags["hostname"] = ep.Hostname;
        if (ep.Name) tags["container_name"] = ep.Name;
        if (ep.container_name) tags["container_name"] = ep.container_name;
        if (ep.EndpointID) tags["endpoint_id"] = ep.EndpointID;
        if (ep.ImageName) tags["image_name"] = ep.ImageName;

        for (const [rawKey, val] of Object.entries(tags)) {
            const k = rawKey.toLowerCase();
            if (!val || blacklist.has(k)) continue;
            if (!tagSuggestions[k]) tagSuggestions[k] = new Set();
            tagSuggestions[k].add(val);
        }
    }
}

// Setup misc controls
function setupControls() {
    document.getElementById("refresh-btn")?.addEventListener("click", loadData);
    document.getElementById("period").addEventListener("change", (e) => {
        const panel = chartSlots.find(s => s.id === activeSlotId);
        if (panel) {
            panel.period = e.target.value;
            loadData();
        }
    });

    document.getElementById("graph-type").addEventListener("change", (e) => {
        const panel = chartSlots.find(s => s.id === activeSlotId);
        if (panel) {
            panel.graphType = e.target.value;
            loadData();
        }
    });

    document.getElementById("group-by").addEventListener("change", (e) => {
        const panel = chartSlots.find(s => s.id === activeSlotId);
        if (panel) {
            panel.groupBy = e.target.value;
            loadData();
        }
    });

    document.getElementById("aggregate").addEventListener("change", (e) => {
        const panel = chartSlots.find(s => s.id === activeSlotId);
        if (panel) {
            panel.aggregate = e.target.value;
            loadData();
        }
    });
}
function resetSlot(panel) {
    if (!panel) return;

    const slotEl = document.getElementById(panel.id);
    if (panel.chart) {
        try {
            panel.chart.destroy();
        } catch (err) {
            console.warn("Chart destroy failed", err);
        }
        panel.chart = null;
    }

    if (slotEl) {
        // Remove the entire slot element
        slotEl.remove();
    }

    // Now recreate the slot
    const newSlot = document.createElement("div");
    newSlot.id = panel.id;
    newSlot.className = "rounded border-2 border-dashed border-gray-100 dark:border-gray-700 p-4 bg-white dark:bg-gray-900 text-gray-400 cursor-pointer flex items-center justify-center h-[250px]";
    newSlot.innerHTML = `<span class="text-sm">➕ New Chart</span>`;
    newSlot.addEventListener("click", () => setActiveSlot(panel.id));

    // Reinsert into the grid
    metricPanelsEl.appendChild(newSlot);

    panel.metrics = [];
    panel.filters = {};
    panel.availableDimensions = [];
    panel.period = "5m";
    panel.graphType = "area";
    panel.groupBy = "";
    panel.aggregate = "";
    panel.shouldReset = false;
}

function showMemoryTypePicker(metric, event) {
    const picker = document.createElement("div");
    picker.className = "absolute z-50 bg-white dark:bg-gray-800 border rounded p-4 shadow-lg space-y-2";

    // 🔥 Position it based on click location
    picker.style.top = `${event.clientY + 10}px`;
    picker.style.left = `${event.clientX}px`;

    const types = [
        { label: "Physical Memory", tag: "scope:physical" },
        { label: "Swap Memory", tag: "scope:swap" },
        { label: "Container Memory (Cgroup)", tag: "scope:cgroup" }
    ];

    types.forEach(type => {
        const btn = document.createElement("button");
        btn.className = "block w-full text-left px-4 py-2 hover:bg-gray-100 dark:hover:bg-gray-700 text-sm";
        btn.textContent = type.label;
        btn.addEventListener("click", () => {
            addSelectedMetric(metric);
            const panel = chartSlots.find(s => s.id === activeSlotId);
            if (panel) {
                panel.filters[type.tag] = true;
                renderSelectedFilters();
                loadData();
            }
            document.body.removeChild(picker);
            metricInput.value = "";
            suggestionsEl.classList.add("hidden");
        });
        picker.appendChild(btn);
    });

    document.body.appendChild(picker);
}
