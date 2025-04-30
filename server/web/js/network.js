import { registerTabInitializer } from "./tabs.js";

const seenInterfaces = new Set();
let selectedInterface = null;

let networkTrafficLineChart = null;
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
    if (!networkTrafficLineChart) return;
    networkTrafficLineChart.data.labels = [];
    networkTrafficLineChart.data.datasets.forEach(ds => ds.data = []);
    networkTrafficLineChart.update();
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
            //console.log(`[${iface}] bytes_sent deltaSec=${deltaSec.toFixed(3)}, deltaBytes=${deltaBytes}, Mbps=${mbps.toFixed(3)}`);
            lastBandwidth[iface].tx = Math.max(mbps, 0);
        }
        //console.log(`✅ BANDWIDTH :: ${iface} → TX=${lastBandwidth[iface].tx.toFixed(2)} Mbps | RX=${lastBandwidth[iface].rx.toFixed(2)} Mbps`);
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
            //console.log(`[${iface}] bytes_recv deltaSec=${deltaSec.toFixed(3)}, deltaBytes=${deltaBytes}, Mbps=${mbps.toFixed(3)}`);
            lastBandwidth[iface].rx = Math.max(mbps, 0);
        }
        //console.log(`✅ BANDWIDTH :: ${iface} → TX=${lastBandwidth[iface].tx.toFixed(2)} Mbps | RX=${lastBandwidth[iface].rx.toFixed(2)} Mbps`);
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
function updateNetworkTrafficLineChart(direction, valueMbps) {
    if (!networkTrafficLineChart) return;

    const label = new Date().toLocaleTimeString();
    const chart = networkTrafficLineChart;

    chart.data.labels.push(label);

    // Add to the appropriate dataset
    if (direction === "tx") {
        chart.data.datasets[0].data.push(valueMbps);
    } else {
        chart.data.datasets[1].data.push(valueMbps);
    }

    // Keep everything in sync
    const maxPoints = 60;
    if (chart.data.labels.length > maxPoints) {
        chart.data.labels.shift();
        chart.data.datasets[0].data.shift();
        chart.data.datasets[1].data.shift();
    }

    chart.update();
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
// Error Rate Chart Initialization
let errorRateChart = null;

function updateErrorRateChart(inErrors, outErrors, packetsIn, packetsOut) {
    if (!errorRateChart) return;
    const timestamp = new Date().toLocaleTimeString();
    const errorPercentIn = packetsIn > 0 ? (inErrors / (packetsIn + inErrors)) * 100 : 0;
    const errorPercentOut = packetsOut > 0 ? (outErrors / (packetsOut + outErrors)) * 100 : 0;

    errorRateChart.data.labels.push(timestamp);
    errorRateChart.data.datasets[0].data.push(errorPercentIn);
    errorRateChart.data.datasets[1].data.push(errorPercentOut);

    if (errorRateChart.data.labels.length > 60) {
        errorRateChart.data.labels.shift();
        errorRateChart.data.datasets[0].data.shift();
        errorRateChart.data.datasets[1].data.shift();
    }
    errorRateChart.update();
}
function createErrorRateChart() {
    const canvas = document.getElementById("errorRateChart");
    if (!canvas || typeof Chart === "undefined") return;
    errorRateChart = new Chart(canvas, {
        type: "line",
        data: {
            labels: [],
            datasets: [
                {
                    label: "Input Errors",
                    data: [],
                    borderColor: "#f87171",
                    backgroundColor: "rgba(248, 113, 113, 0.1)",
                    fill: true,
                    tension: 0.3,
                    pointRadius: 2,
                },
                {
                    label: "Output Errors",
                    data: [],
                    borderColor: "#facc15",
                    backgroundColor: "rgba(250, 204, 21, 0.1)",
                    fill: true,
                    tension: 0.3,
                    pointRadius: 2,
                }
            ]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            scales: {
                y: {
                    beginAtZero: true,
                    max: 100,
                    ticks: {
                        callback: (val) => `${val}%`,
                        color: "#9CA3AF"
                    }
                },
                x: {
                    ticks: { color: "#9CA3AF" }
                }
            },
            plugins: {
                legend: { labels: { color: "#4B5563" } }
            }
        }
    });
    //console.log("✅ errorRateChart initialized");
}

function createNetworkTrafficLineChart() {
    const canvas = document.getElementById("networkTrafficLineChart");
    if (!canvas || typeof Chart === "undefined") return;

    networkTrafficLineChart = new Chart(canvas, {
        type: "line",
        data: {
            labels: [], // populated dynamically
            datasets: [
                {
                    label: "Upload (Mbps)",
                    data: [],
                    borderColor: "#3b82f6",
                    backgroundColor: "rgba(59, 130, 246, 0.1)",
                    fill: true,
                    tension: 0.3,
                    pointRadius: 2,
                    spanGaps: true
                },
                {
                    label: "Download (Mbps)",
                    data: [],
                    borderColor: "#10b981",
                    backgroundColor: "rgba(16, 185, 129, 0.1)",
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
            animation: {
                duration: 0  // ⛔ disable animation so shift is immediate
            },
            scales: {
                x: {
                    ticks: {
                        autoSkip: true,
                        maxTicksLimit: 10,
                        color: "#9CA3AF"
                    }
                },
                y: {
                    beginAtZero: true,
                    ticks: {
                        callback: (val) => `${val} Mbps`,
                        color: "#9CA3AF"
                    }
                }
            },
            plugins: {
                legend: {
                    labels: {
                        color: "#4B5563"
                    }
                },
                tooltip: {
                    callbacks: {
                        label: (ctx) => `${ctx.dataset.label}: ${ctx.parsed.y.toFixed(1)} Mbps`
                    }
                }
            }
        }
    });

    console.log("✅ networkTrafficLineChart initialized");
}

function redrawNetworkTrafficLineChartFromCache() {
    const iface = selectedInterface;
    const bw = lastBandwidth[iface];
    if (!bw) return;

    // You could optionally store historical samples per interface
    // For now, we just seed with the latest point
    updateNetworkTrafficLineChart("tx", bw.tx ?? 0);
    updateNetworkTrafficLineChart("rx", bw.rx ?? 0);
}

// INIT NETWORK TAB
function initNetworkTab() {

    createNetworkTrafficLineChart();
    createErrorRateChart();

    const dropdown = document.getElementById("interface-select");
    if (dropdown && !dropdown._bound) {
        dropdown.addEventListener("change", (e) => {
            selectedInterface = e.target.value;
            resetNetworkTrafficChart();
            updateCurrentBandwidthDisplay();
            updateStaticStatCardsFromCache();
            redrawNetworkTrafficLineChartFromCache();
        });

        dropdown._bound = true;
        console.log("✅ Bound interface switcher");
    }
}


window.networkMetricHandler = function (metrics) {
    for (const metric of metrics) {
        if (metric.name.startsWith("err_")) {
            //console.log(`🧪 Error Metric: ${metric.name} (${metric.dimensions?.interface}) = ${metric.value}`);
        }

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
            //console.log(`📦 ${iface} :: ${metric.name} = ${metric.value}`);
            //console.log(`➡ TX: ${lastBandwidth[iface]?.tx?.toFixed(2)} Mbps`);
            //console.log(`⬅ RX: ${lastBandwidth[iface]?.rx?.toFixed(2)} Mbps`);

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
            if (!networkTrafficLineChart) return; // Prevent crash if metric arrives mid-tab-load

            if (iface === selectedInterface && queuedBandwidthSample.tx != null && queuedBandwidthSample.rx != null) {
                const label = new Date().toLocaleTimeString();
                networkTrafficLineChart.data.labels.push(label);
                networkTrafficLineChart.data.datasets[0].data.push(queuedBandwidthSample.tx);
                networkTrafficLineChart.data.datasets[1].data.push(queuedBandwidthSample.rx);

                // keep sync
                const max = 60;
                if (networkTrafficLineChart.data.labels.length > max) {
                    networkTrafficLineChart.data.labels.shift();
                    networkTrafficLineChart.data.datasets[0].data.shift();
                    networkTrafficLineChart.data.datasets[1].data.shift();
                }

                networkTrafficLineChart.update();

                // reset for next tick
                queuedBandwidthSample.tx = null;
                queuedBandwidthSample.rx = null;
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
        }

        updateInterfaceTable(metric);
    }
};
registerTabInitializer("network", initNetworkTab);

window.addEventListener("metrics", ({ detail: payload }) => {
    if (payload?.metrics && payload?.meta?.endpoint_id?.startsWith("host-")) {
        // Call your existing function directly:
        window.networkMetricHandler(payload.metrics);
    }
});