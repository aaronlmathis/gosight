import { registerTabInitializer } from "./tabs.js";
import { createApexAreaChart } from "./apex_helpers.js";

const seenInterfaces = new Set();
let selectedInterface = null;

let networkTrafficChart = null;
let errorRateChart = null;
const lastMetrics = {}; // key: interface, value: { metricName: value }
const lastBandwidth = {}; // key: iface, value: { tx: Mbps, rx: Mbps }
const lastBytesAndTime = {}; // key: iface, value: { sent, recv, timestamp }
const queuedBandwidthSample = {
    timestamp: null,
    tx: null,
    rx: null,
};
const lastBandwidthHistory = {}; // key: iface, value: { tx: [], rx: [] }

function resetNetworkTrafficChart() {
    if (!networkTrafficChart) return;
    // Clear data for ApexCharts
    networkTrafficChart.updateSeries([
        { name: "Upload (Mbps)", data: [] },
        { name: "Download (Mbps)", data: [] }
    ], true);
}

function addToBandwidthHistory(iface, direction, value) {
    if (!lastBandwidthHistory[iface]) {
        lastBandwidthHistory[iface] = { tx: [], rx: [] };
    }
    const history = lastBandwidthHistory[iface][direction];
    const now = Date.now();
    history.push({ time: now, value });

    // Keep only last 10 minutes
    const cutoff = now - 10 * 60 * 1000;
    while (history.length && history[0].time < cutoff) {
        history.shift();
    }
}

function updateCurrentBandwidthDisplay() {
    const bw = lastBandwidth[selectedInterface];
    const tx = bw?.tx;
    const rx = bw?.rx;

    document.getElementById("current-tx").textContent =
        tx != null ? formatMbps(tx) + " Mbps" : "--";

    document.getElementById("current-rx").textContent =
        rx != null ? formatMbps(rx) + " Mbps" : "--";

    document.getElementById("current-interface-label").textContent = selectedInterface;
    const peak = getPeakBandwidthMbps(selectedInterface);
    document.getElementById("peak-bandwidth").textContent = `↑ ${formatMbps(peak.tx)} / ↓ ${formatMbps(peak.rx)}`;
}

function formatMbps(val) {
    if (val >= 1) return val.toFixed(1);
    if (val > 0.01) return val.toFixed(2);
    if (val > 0) return "< 0.01";
    return "0.00";
}

function getPeakBandwidthMbps(iface) {
    const hist = lastBandwidthHistory[iface];
    if (!hist) return { tx: 0, rx: 0 };

    const peakTx = Math.max(...hist.tx.map(p => p.value), 0);
    const peakRx = Math.max(...hist.rx.map(p => p.value), 0);
    return { tx: peakTx, rx: peakRx };
}

function calculateAndUpdateBandwidth(iface, metricName, newValue) {
    const now = Date.now();

    if (!lastBytesAndTime[iface]) {
        lastBytesAndTime[iface] = {
            sent: 0,
            recv: 0,
            timestamp_sent: now,
            timestamp_recv: now
        };
    }

    if (!lastBandwidth[iface]) {
        lastBandwidth[iface] = { tx: 0, rx: 0 };
    }

    const state = lastBytesAndTime[iface];

    if (metricName === "bytes_sent") {
        if (state.sent === 0) {
            state.sent = newValue;
            state.timestamp_sent = now;
            return;
        }
        const deltaBytes = newValue - state.sent;
        const deltaSec = (now - state.timestamp_sent) / 1000;
        if (deltaSec > 0) {
            const mbps = (deltaBytes * 8) / 1_000_000 / deltaSec;
            lastBandwidth[iface].tx = Math.max(mbps, 0);
        }

        state.sent = newValue;
        state.timestamp_sent = now;
    }

    if (metricName === "bytes_recv") {
        if (state.recv === 0) {
            state.recv = newValue;
            state.timestamp_recv = now;
            return;
        }
        const deltaBytes = newValue - state.recv;
        const deltaSec = (now - state.timestamp_recv) / 1000;
        if (deltaSec > 0) {
            const mbps = (deltaBytes * 8) / 1_000_000 / deltaSec;
            lastBandwidth[iface].rx = Math.max(mbps, 0);
        }

        state.recv = newValue;
        state.timestamp_recv = now;
    }
    addToBandwidthHistory(iface, "tx", lastBandwidth[iface].tx);
    addToBandwidthHistory(iface, "rx", lastBandwidth[iface].rx);
}

function updateStaticStatCardsFromCache() {
    const cached = lastMetrics[selectedInterface];
    if (!cached) return;

    if (cached["packets_sent"] !== undefined) {
        document.getElementById("stat-packets-sent").textContent = cached["packets_sent"] != null ? cached["packets_sent"].toLocaleString() : "--";
    }
    if (cached["packets_recv"] !== undefined) {
        document.getElementById("stat-packets-recv").textContent = cached["packets_recv"] != null ? cached["packets_recv"].toLocaleString() : "--";
    }
    if (cached["err_in"] !== undefined) {
        const valIn = cached["err_in"];
        document.getElementById("stat-errors-in").textContent =
            typeof valIn === "number" ? valIn.toLocaleString() : "--";
    }
    if (cached["err_out"] !== undefined) {
        const valOut = cached["err_out"];
        document.getElementById("stat-errors-out").textContent =
            typeof valOut === "number" ? valOut.toLocaleString() : "--";
    }

    if (cached["packets_recv"] && cached["err_in"]) {
        if (typeof cached["packets_recv"] === "number" && typeof cached["err_in"] === "number") {
            const errPercentIn = (cached["err_in"] / (cached["packets_recv"] + cached["err_in"])) * 100;
            document.getElementById("stat-error-percent-in").textContent = errPercentIn.toFixed(2) + "%";
        }
    }
    if (cached["packets_sent"] && cached["err_out"]) {
        if (typeof cached["packets_sent"] === "number" && typeof cached["err_out"] === "number") {
            const errPercentOut = (cached["err_out"] / (cached["packets_sent"] + cached["err_out"])) * 100;
            document.getElementById("stat-error-percent-out").textContent = errPercentOut.toFixed(2) + "%";
        }
    }
}

function updateNetworkTrafficChart(direction, valueMbps) {
    if (!networkTrafficChart) return;

    const now = new Date().getTime();
    const seriesIndex = direction === "tx" ? 0 : 1;

    // Get current series data
    let series = networkTrafficChart.w.config.series.map(s => ({
        name: s.name,
        data: s.data ? [...s.data] : []
    }));

    // Ensure value is non-negative
    const safeValue = Math.max(0, valueMbps);

    // Add new data point
    series[seriesIndex].data.push([now, safeValue]);

    // Keep only last 60 data points
    const maxPoints = 60;
    if (series[seriesIndex].data.length > maxPoints) {
        series[seriesIndex].data.splice(0, series[seriesIndex].data.length - maxPoints);
    }

    // Update chart
    networkTrafficChart.updateSeries(series, false);
}

function updateInterfaceTable(metric) {
    const iface = metric.dimensions?.interface;
    const value = metric.value;
    if (!iface) return;

    if (!seenInterfaces.has(iface)) {
        seenInterfaces.add(iface);
        addInterfaceOption(iface);
        addInterfaceRow(iface);
    }

    const idMap = {
        "bytes_sent": `${iface}-tx-mbps`,
        "bytes_recv": `${iface}-rx-mbps`,
        "packets_sent": `${iface}-packets_sent`,
        "packets_recv": `${iface}-packets_recv`,
        "err_in": `${iface}-err_in`,
        "err_out": `${iface}-err_out`,
    };

    const cellId = idMap[metric.name];
    if (!cellId) return;

    const cell = document.getElementById(cellId);
    if (!cell) return;

    if (metric.name.startsWith("bytes_")) {
        const bw = lastBandwidth[iface]?.[metric.name === "bytes_sent" ? "tx" : "rx"] ?? 0;
        cell.textContent = `${bw.toFixed(1)} Mbps`;
    } else {
        cell.textContent = value.toLocaleString();
    }
}

function addInterfaceOption(iface) {
    const select = document.getElementById("interface-select");
    const option = document.createElement("option");
    option.value = option.textContent = iface;
    select.appendChild(option);
}

function addInterfaceRow(iface) {
    const tbody = document.getElementById("interface-table-body");
    const tr = document.createElement("tr");
    tr.innerHTML = `
        <td class="px-4 py-2 font-medium text-blue-500">${iface}</td>
        <td class="px-4 py-2" id="${iface}-tx-mbps">--</td>
        <td class="px-4 py-2" id="${iface}-rx-mbps">--</td>
        <td class="px-4 py-2" id="${iface}-packets_sent">--</td>
        <td class="px-4 py-2" id="${iface}-packets_recv">--</td>
        <td class="px-4 py-2" id="${iface}-err_in">--</td>
        <td class="px-4 py-2" id="${iface}-err_out">--</td>
    `;
    tbody.appendChild(tr);
}

function updateErrorRateChart(inErrors, outErrors, packetsIn, packetsOut) {
    if (!errorRateChart) return;

    const now = new Date().getTime();
    const errorPercentIn = packetsIn > 0 ? (inErrors / (packetsIn + inErrors)) * 100 : 0;
    const errorPercentOut = packetsOut > 0 ? (outErrors / (packetsOut + outErrors)) * 100 : 0;

    // Get current series data
    let series = errorRateChart.w.config.series.map(s => ({
        name: s.name,
        data: s.data ? [...s.data] : []
    }));

    // Add new data points (ensuring non-negative values)
    series[0].data.push([now, Math.max(0, errorPercentIn)]);
    series[1].data.push([now, Math.max(0, errorPercentOut)]);

    // Keep only last 60 data points
    const maxPoints = 60;
    if (series[0].data.length > maxPoints) {
        series[0].data.splice(0, series[0].data.length - maxPoints);
    }
    if (series[1].data.length > maxPoints) {
        series[1].data.splice(0, series[1].data.length - maxPoints);
    }

    // Update chart
    errorRateChart.updateSeries(series, false);
}

function createErrorRateChart() {
    const container = document.getElementById("errorRateChart");
    if (!container) return;

    errorRateChart = createApexAreaChart(
        "errorRateChart",
        "Error Rate",
        ["Input Errors", "Output Errors"],
        false
    );

    // Customize Y axis to show percentages
    errorRateChart.updateOptions({
        yaxis: {
            min: 0,
            max: 100,
            forceNiceScale: true,
            labels: {
                formatter: (val) => `${val.toFixed(1)}%`
            },
            title: {
                text: "Error Rate (%)"
            }
        },
        stroke: {
            curve: "smooth",
            width: 2
        },
        title: {
            text: undefined,
            show: false
        },
        colors: ["#f87171", "#facc15"]
    });
}

function createNetworkTrafficChart() {
    const container = document.getElementById("networkTrafficLineChart");
    if (!container) return;

    networkTrafficChart = createApexAreaChart(
        "networkTrafficLineChart",
        "Network Traffic",
        ["Upload (Mbps)", "Download (Mbps)"],
        false
    );


    networkTrafficChart.updateOptions({
        yaxis: {
            min: 0,
            forceNiceScale: true,
            labels: {
                formatter: (val) => `${val.toFixed(1)} Mbps`
            },
            title: {
                text: "Bandwidth (Mbps)"
            }
        },
        colors: ["#3b82f6", "#10b981"],
        stroke: {
            curve: "smooth",
            width: 2
        },
        title: {
            text: undefined,
            show: false
        },
        tooltip: {
            enabled: true,
            x: { format: 'HH:mm:ss' },
            y: { formatter: (val) => `${val.toFixed(2)} Mbps` }
        },
    });

}

function redrawNetworkTrafficChartFromCache() {
    const iface = selectedInterface;
    const bw = lastBandwidth[iface];
    if (!bw) return;

    // For ApexCharts, we need to clear and add a single point
    const now = new Date().getTime();

    networkTrafficChart.updateSeries([
        { name: "Upload (Mbps)", data: [[now, Math.max(0, bw.tx ?? 0)]] },
        { name: "Download (Mbps)", data: [[now, Math.max(0, bw.rx ?? 0)]] }
    ], true);
}

// INIT NETWORK TAB
function initNetworkTab() {
    createNetworkTrafficChart();
    createErrorRateChart();

    const dropdown = document.getElementById("interface-select");
    if (dropdown && !dropdown._bound) {
        dropdown.addEventListener("change", (e) => {
            selectedInterface = e.target.value;
            resetNetworkTrafficChart();
            updateCurrentBandwidthDisplay();
            updateStaticStatCardsFromCache();
            redrawNetworkTrafficChartFromCache();
        });

        dropdown._bound = true;
    }
}

window.networkMetricHandler = function (metrics) {
    for (const metric of metrics) {
        if (
            metric.namespace !== "System" ||
            metric.subnamespace !== "Network" ||
            typeof metric.value !== "number" ||
            !metric.dimensions?.interface
        ) {
            continue;
        }

        const iface = metric.dimensions.interface;
        if (!lastMetrics[iface]) lastMetrics[iface] = {};
        lastMetrics[iface][metric.name] = metric.value;

        if (!lastBandwidth[iface]) lastBandwidth[iface] = { tx: 0, rx: 0 };
        if (!seenInterfaces.has(iface)) {
            seenInterfaces.add(iface);
            addInterfaceOption(iface);
            addInterfaceRow(iface);
        }

        if (!selectedInterface && iface) {
            selectedInterface = iface;
            document.getElementById("interface-select").value = iface;
            updateCurrentBandwidthDisplay();
            updateStaticStatCardsFromCache();
            console.log("Auto-selected first interface:", iface);
        }

        if (iface === selectedInterface) {
            if (metric.name === "bytes_sent") {
                calculateAndUpdateBandwidth(iface, metric.name, metric.value);
                queuedBandwidthSample.tx = lastBandwidth[iface].tx;
            } else if (metric.name === "bytes_recv") {
                calculateAndUpdateBandwidth(iface, metric.name, metric.value);
                queuedBandwidthSample.rx = lastBandwidth[iface].rx;
            } else if (metric.name === "packets_sent") {
                document.getElementById("stat-packets-sent").textContent = metric.value.toLocaleString();
            } else if (metric.name === "packets_recv") {
                document.getElementById("stat-packets-recv").textContent = metric.value.toLocaleString();
            } else if (metric.name === "err_in") {
                document.getElementById("stat-errors-in").textContent = metric.value.toLocaleString();
            } else if (metric.name === "err_out") {
                document.getElementById("stat-errors-out").textContent = metric.value.toLocaleString();
            }

            if (!networkTrafficChart) return; // Prevent crash if metric arrives mid-tab-load

            if (iface === selectedInterface && queuedBandwidthSample.tx != null && queuedBandwidthSample.rx != null) {


                if (
                    queuedBandwidthSample.tx != null &&
                    queuedBandwidthSample.rx != null
                ) {
                    const now = Date.now();
                    const txPoint = [now, Math.max(0, queuedBandwidthSample.tx)];
                    const rxPoint = [now, Math.max(0, queuedBandwidthSample.rx)];

                    let series = networkTrafficChart.w.config.series.map(s => ({
                        name: s.name,
                        data: s.data ? [...s.data] : []
                    }));

                    series[0].data.push(txPoint);
                    series[1].data.push(rxPoint);

                    const maxPoints = 60;
                    if (series[0].data.length > maxPoints) {
                        series[0].data.splice(0, series[0].data.length - maxPoints);
                    }
                    if (series[1].data.length > maxPoints) {
                        series[1].data.splice(0, series[1].data.length - maxPoints);
                    }

                    networkTrafficChart.updateSeries(series, false);

                    // Clear queue after successful chart update
                    queuedBandwidthSample.tx = null;
                    queuedBandwidthSample.rx = null;
                }

            }

            updateCurrentBandwidthDisplay();

            const pktIn = lastMetrics[iface]["packets_recv"];
            const pktOut = lastMetrics[iface]["packets_sent"];
            const errIn = lastMetrics[iface]["err_in"];
            const errOut = lastMetrics[iface]["err_out"];

            const inPct = (errIn != null && pktIn != null)
                ? (errIn / (pktIn + errIn)) * 100
                : null;
            const outPct = (errOut != null && pktOut != null)
                ? (errOut / (pktOut + errOut)) * 100
                : null;

            const errInElem = document.getElementById("stat-errors-in");
            const errOutElem = document.getElementById("stat-errors-out");
            const errPctInElem = document.getElementById("stat-error-percent-in");
            const errPctOutElem = document.getElementById("stat-error-percent-out");

            if (errInElem) {
                errInElem.textContent = errIn != null ? errIn.toLocaleString() : "--";
            }
            if (errOutElem) {
                errOutElem.textContent = errOut != null ? errOut.toLocaleString() : "--";
            }
            if (errPctInElem) {
                errPctInElem.textContent = inPct != null ? inPct.toFixed(2) + "%" : "--";
            }
            if (errPctOutElem) {
                errPctOutElem.textContent = outPct != null ? outPct.toFixed(2) + "%" : "--";
            }

            // Update error rate chart if we have all needed data
            if (errIn != null && errOut != null && pktIn != null && pktOut != null) {
                updateErrorRateChart(errIn, errOut, pktIn, pktOut);
            }
        }

        updateInterfaceTable(metric);
    }
};

registerTabInitializer("network", initNetworkTab);

window.addEventListener("metrics", ({ detail: payload }) => {
    if (payload?.metrics && payload?.meta?.endpoint_id?.startsWith("host-")) {
        // Call your existing function directly:
        window.networkMetricHandler(payload.metrics);
        console.log("Network metrics received:", payload.metrics);
    }
});