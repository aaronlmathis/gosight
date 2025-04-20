const seenInterfaces = new Set();
let selectedInterface = null;

let networkTrafficLineChart = null;
const lastMetrics = {}; // key: interface, value: { metricName: value }
const lastBandwidth = {}; // key: iface, value: { tx: Mbps, rx: Mbps }
const lastBytesAndTime = {}; // key: iface, value: { sent, recv, timestamp }

function resetNetworkTrafficChart() {
    if (!networkTrafficLineChart) return;
    networkTrafficLineChart.data.labels = [];
    networkTrafficLineChart.data.datasets.forEach(ds => ds.data = []);
    networkTrafficLineChart.update();
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
}
function formatMbps(val) {
    if (val >= 1) return val.toFixed(1);
    if (val > 0.01) return val.toFixed(2);
    if (val > 0) return "< 0.01";
    return "0.00";
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
    const dsIndex = direction === "tx" ? 0 : 1;
    const ds = networkTrafficLineChart.data.datasets[dsIndex];
    const label = new Date().toLocaleTimeString();
    ds.data.push(valueMbps);
    networkTrafficLineChart.data.labels.push(label);
    if (ds.data.length > 60) {
        ds.data.shift();
        networkTrafficLineChart.data.labels.shift();
    }
    networkTrafficLineChart.update();
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
            labels: [],
            datasets: [
                {
                    label: "Upload (Mbps)",
                    data: [],
                    borderColor: "#3b82f6",
                    backgroundColor: "rgba(59, 130, 246, 0.1)",
                    fill: true,
                    tension: 0.3,
                    pointRadius: 2,
                },
                {
                    label: "Download (Mbps)",
                    data: [],
                    borderColor: "#10b981",
                    backgroundColor: "rgba(16, 185, 129, 0.1)",
                    fill: true,
                    tension: 0.3,
                    pointRadius: 2,
                },
            ],
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            scales: {
                y: {
                    beginAtZero: true,
                    ticks: {
                        callback: (val) => `${val} Mbps`,
                    },
                },
            },
            plugins: {
                legend: { labels: { color: "#4B5563" } },
                tooltip: {
                    callbacks: {
                        label: (ctx) => `${ctx.dataset.label}: ${ctx.parsed.y.toFixed(1)} Mbps`,
                    },
                },
            },
        },
    });
    //console.log("✅ networkChart initialized");

}
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
        //console.log("✅ Bound interface switcher");
    }
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
const observer = new MutationObserver(() => {
    const panel = document.getElementById("network");
    if (panel && !panel.classList.contains("hidden") && !panel._initialized) {
        panel._initialized = true;
        initNetworkTab();
    }
});
observer.observe(document.body, { childList: true, subtree: true });

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
            console.log("🎯 Auto-selected first interface:", iface);
        }

        if (iface === selectedInterface) {
            //console.log(`📦 ${iface} :: ${metric.name} = ${metric.value}`);
            //console.log(`➡ TX: ${lastBandwidth[iface]?.tx?.toFixed(2)} Mbps`);
            //console.log(`⬅ RX: ${lastBandwidth[iface]?.rx?.toFixed(2)} Mbps`);

            if (metric.name === "bytes_sent") {
                calculateAndUpdateBandwidth(iface, metric.name, metric.value);
                updateNetworkTrafficLineChart("tx", lastBandwidth[iface].tx);
            } else if (metric.name === "bytes_recv") {
                calculateAndUpdateBandwidth(iface, metric.name, metric.value);
                updateNetworkTrafficLineChart("rx", lastBandwidth[iface].rx);
            } else if (metric.name === "packets_sent") {
                document.getElementById("stat-packets-sent").textContent = metric.value.toLocaleString();
            } else if (metric.name === "packets_recv") {
                document.getElementById("stat-packets-recv").textContent = metric.value.toLocaleString();
            } else if (metric.name === "err_in") {
                document.getElementById("stat-errors-in").textContent = metric.value.toLocaleString();
            } else if (metric.name === "err_out") {
                document.getElementById("stat-errors-out").textContent = metric.value.toLocaleString();
            }

            updateCurrentBandwidthDisplay();

            const pktIn = lastMetrics[iface]["packets_recv"];
            const pktOut = lastMetrics[iface]["packets_sent"];
            const errIn = lastMetrics[iface]["err_in"];
            const errOut = lastMetrics[iface]["err_out"];

            const inPct = (errIn != null && packetsRecv != null)
                ? (errIn / (packetsRecv + errIn)) * 100
                : null;
            const outPct = (errOut != null && packetsSent != null)
                ? (errOut / (packetsSent + errOut)) * 100
                : null;

            const errInElem = document.getElementById("stat-errors-in");
            const errOutElem = document.getElementById("stat-errors-out");
            const errPctInElem = document.getElementById("stat-error-percent-in");
            const errPctOutElem = document.getElementById("stat-error-percent-out");

            console.log("🎯 stat-errors-in exists?", !!errInElem);
            console.log("🎯 stat-errors-out exists?", !!errOutElem);

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
