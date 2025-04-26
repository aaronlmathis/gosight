// metric-explorer.js
import { gosightFetch } from "./api.js";

let allMetrics = []; // Full list of { label, namespace, subnamespace, name }
let tagSuggestions = {}; // key:tag, value:Set of values
const selectedFilters = {}; // key => value
const chartMap = {}; // key: metric.label → ApexCharts instance

window.addEventListener("DOMContentLoaded", async () => {
    const metricInput = document.getElementById("metric-search");
    const suggestionsEl = document.getElementById("metric-suggestions");
    const selectedEl = document.getElementById("selected-metrics");
    const tagInput = document.getElementById("tag-search");
    const tagSuggestionsEl = document.getElementById("tag-suggestions");
    document.getElementById("group-by").addEventListener("change", loadData);
    const timeRange = document.getElementById("time-series")?.value || "1h";
    document.getElementById("aggregate").addEventListener("change", loadData);
    document.getElementById("group-by").addEventListener("change", loadData);
    document.getElementById("refresh-btn")?.addEventListener("click", loadData);

    document.getElementById("layout-cols")?.addEventListener("input", () => {
        const cols = parseInt(document.getElementById("layout-cols").value) || 2;
        const panels = document.getElementById("metric-panels");
        panels.className = `grid grid-cols-${cols} gap-4`;
    });

    // Step 1: Load metric names
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

    // Step 2: Search box listener
    metricInput.addEventListener("input", () => {
        const q = metricInput.value.trim().toLowerCase();
        suggestionsEl.innerHTML = "";
        if (q.length < 2) {
            suggestionsEl.innerHTML = ""; // Clear box on short query
            return;
        }

        const filtered = allMetrics.filter(m => m.label.includes(q)).slice(0, 10);
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
                addSelectedMetric(metric);
                metricInput.value = "";
                suggestionsEl.innerHTML = "";
            });
            suggestionsEl.appendChild(item);
        }
    });

    function addSelectedMetric(metric) {
        const pill = document.createElement("span");
        pill.className = "inline-flex items-center bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200 px-2 py-1 rounded-full text-xs mr-1 mb-1";
        pill.textContent = metric.label;

        const remove = document.createElement("button");
        remove.className = "ml-2 text-blue-500 hover:text-red-500 text-sm";
        remove.innerHTML = "&times;";
        remove.addEventListener("click", () => {
            pill.remove();
            const chartId = `chart-${metric.label}`;
            const panel = document.getElementById(chartId);
            if (panel) panel.remove();
            delete chartMap[metric.label];
        });

        pill.appendChild(remove);
        selectedEl.appendChild(pill);
        renderChartPanel(metric);
        loadData();
    }


    await loadEndpointTagSuggestions();

    document.getElementById("from-filter-input").addEventListener("input", (e) => {
        const query = e.target.value.toLowerCase();
        const box = document.getElementById("from-suggestions");
        box.innerHTML = "";

        if (!query || query.length < 2) {
            box.classList.add("hidden");
            return;
        }

        const matches = [];

        for (const [key, values] of Object.entries(tagSuggestions)) {
            for (const val of values) {
                const entry = `${key}:${val}`;
                if (entry.toLowerCase().includes(query)) {
                    matches.push({ key, val });
                }
            }
        }

        if (matches.length === 0) {
            box.classList.add("hidden");
            return;
        }

        box.classList.remove("hidden");
        box.style.maxHeight = "240px";
        box.style.overflowY = "auto";

        matches.slice(0, 10).forEach(({ key, val }) => {
            const item = document.createElement("div");
            item.className = "px-3 py-2 hover:bg-gray-100 dark:hover:bg-gray-700 cursor-pointer text-sm";
            item.textContent = `${key}: ${val}`;
            item.onclick = () => {
                addSelectedFilter(key, val);
                box.classList.add("hidden");
                document.getElementById("from-filter-input").value = "";
            };
            box.appendChild(item);
        });
    });


    function addSelectedFilter(key, val) {
        const id = `${key}:${val}`;
        if (selectedFilters[id]) return;
        selectedFilters[id] = true;

        const container = document.getElementById("from-selected");
        const chip = document.createElement("span");
        chip.className = "inline-flex items-center bg-gray-200 dark:bg-gray-800 text-xs rounded px-2 py-1";
        chip.innerHTML = `${key}:${val} <button class="ml-1 text-red-500 hover:text-red-700">&times;</button>`;

        chip.querySelector("button").onclick = () => {
            delete selectedFilters[id];
            chip.remove();
        };

        container.appendChild(chip);
        loadData();

    }

    async function loadEndpointTagSuggestions() {
        const hosts = await gosightFetch("/api/v1/endpoints/hosts").then(res => res.json());
        const containers = await gosightFetch("/api/v1/endpoints/containers").then(res => res.json());
        const endpoints = [...hosts, ...containers];

        tagSuggestions = {};
        const blacklist = ["agent_start_time", "_cmdline", "_uid", "_exe"];

        for (const ep of endpoints) {
            const tags = { ...ep.labels };

            // Normalize common tag fields from both hosts and containers
            if (ep.Hostname) tags["hostname"] = ep.Hostname;
            if (ep.Name) tags["container_name"] = ep.Name;
            if (ep.container_name) tags["container_name"] = ep.container_name;
            if (ep.EndpointID) tags["endpoint_id"] = ep.EndpointID;
            if (ep.ImageName) tags["image_name"] = ep.ImageName;

            for (const [rawKey, val] of Object.entries(tags)) {
                const k = rawKey.toLowerCase();
                if (!val || blacklist.includes(k)) continue;
                if (!tagSuggestions[k]) tagSuggestions[k] = new Set();
                tagSuggestions[k].add(val);
            }
        }
    }

    function buildSeries(dataArray, groupKey, agg) {
        // No aggregation? → group by key
        if (!agg) {
            const groups = {};
            dataArray.forEach((d) => {
                const label =
                    groupKey && d.tags?.[groupKey]
                        ? d.tags[groupKey]
                        : d.tags?.instance || d.tags?.hostname || d.tags?.container_name || "unknown";

                if (!groups[label]) groups[label] = [];
                groups[label].push([d.timestamp, d.value]);
            });
            return Object.entries(groups).map(([name, data]) => ({ name, data }));
        }

        // Aggregated view (1 line)
        const buckets = {};
        dataArray.forEach((d) => {
            const ts = d.timestamp;
            if (!buckets[ts]) buckets[ts] = [];
            buckets[ts].push(d.value);
        });

        const result = Object.entries(buckets).map(([ts, values]) => {
            const t = parseInt(ts);
            let v = 0;
            switch (agg) {
                case "sum": v = values.reduce((a, b) => a + b, 0); break;
                case "avg": v = values.reduce((a, b) => a + b, 0) / values.length; break;
                case "min": v = Math.min(...values); break;
                case "max": v = Math.max(...values); break;
                case "stddev":
                    const mean = values.reduce((a, b) => a + b, 0) / values.length;
                    const variance = values.reduce((a, b) => a + Math.pow(b - mean, 2), 0) / values.length;
                    v = Math.sqrt(variance);
                    break;
            }
            return [t, v];
        });

        return [{ name: agg.toUpperCase(), data: result }];
    }

    async function loadData() {
        console.log("loadData() triggered");
    
        const metricLabels = Array.from(document.querySelectorAll("#selected-metrics span"))
            .map(el => el.textContent.trim().replace("×", "").trim());
    
        console.log("Selected metrics:", metricLabels);
    
        if (metricLabels.length === 0) {
            console.log("No metrics selected; exiting loadData.");
            return;
        }
    
        const groupKey = document.getElementById("group-by")?.value || "";
        const agg = document.getElementById("aggregate")?.value || "";
        const timeRange = document.getElementById("time-range")?.value || "1h";
    
        const now = new Date();
        let startDate = new Date(now);
    
        if (timeRange.endsWith("m")) startDate.setMinutes(now.getMinutes() - parseInt(timeRange));
        else if (timeRange.endsWith("h")) startDate.setHours(now.getHours() - parseInt(timeRange));
        else if (timeRange.endsWith("d")) startDate.setDate(now.getDate() - parseInt(timeRange));
        else startDate.setHours(now.getHours() - 1);
    
        const start = startDate.toISOString();
        const end = now.toISOString();
    
        const tagFilter = Object.keys(selectedFilters)
            .map(f => f.replace(":", "="))
            .join(",");
    
        console.log("Building API queries with time range", start, "→", end, "and tags:", tagFilter);
    
        for (const metric of metricLabels) {
            const url = `/api/v1/query?metric=${encodeURIComponent(metric)}&start=${encodeURIComponent(start)}&end=${encodeURIComponent(end)}` +
                (tagFilter ? `&tags=${encodeURIComponent(tagFilter)}` : "");
            
            console.log("Fetching URL:", url);
    
            try {
                const res = await gosightFetch(url);
                const data = await res.json();
                console.log("Received data for", metric, data);
    
                const series = buildSeries(data, groupKey, agg);
                const chart = chartMap[metric];
                if (chart) {
                    console.log("💡 Calling updateSeries for", metric, series);
                    chart.updateSeries(series);
                }
                
                if (chart) chart.updateSeries(series);
            } catch (err) {
                console.error("Failed to load metric", metric, err);
            }
        }
    }
    
    function showChartError(metric, message) {
        const panel = document.getElementById(`chart-${metric}`);
        if (!panel || !panel.chart) return;

        panel.chart.updateSeries([]);
        panel.chart.updateOptions({
            noData: {
                text: message || "Error loading data",
                style: {
                    color: "#f87171", // Tailwind red-400
                    fontSize: '14px',
                }
            }
        });
    }
    function renderChartPanel(metric) {
        const panel = document.createElement("div");
        panel.id = `chart-${metric.label}`;
        panel.className = "rounded border p-2 bg-white dark:bg-gray-900";

        document.getElementById("metric-panels").appendChild(panel);

        const graphType = document.getElementById("graph-type")?.value || "line";
        const legendPos = document.getElementById("legend")?.value?.toLowerCase() || "bottom";

        const chart = new ApexCharts(panel, {
            chart: {
                type: graphType,
                height: 250,
                animations: { enabled: false },
                toolbar: { show: false },
            },
            stroke: { curve: "smooth" },
            theme: { mode: document.documentElement.classList.contains("dark") ? "dark" : "light" },
            legend: { position: legendPos },
            xaxis: { type: "datetime" },
            series: [],
            noData: { text: "Loading..." },
        });

        panel.chart = chart;
        chartMap[metric.label] = chart;
        chart.render();
    }
});