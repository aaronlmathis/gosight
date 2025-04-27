// ws.js
//
// Handles WebSocket connections to metrics, logs, events, and alerts streams
// for the GoSight dashboard.

// === Global WebSocket references ===
export let metricsSocket = null;
export let logsSocket = null;
export let eventsSocket = null;
export let alertsSocket = null;

// === Status badge tracking ===
let metricsBadge = document.getElementById("metrics-ws-status");
let logsBadge = document.getElementById("logs-ws-status");
let eventsBadge = document.getElementById("events-ws-status");
let alertsBadge = document.getElementById("alerts-ws-status");

// === Connection logic per socket ===
function connectMetricsSocket() {
    metricsSocket = new WebSocket(`wss://${location.host}/ws/metrics?endpointID=${encodeURIComponent(window.endpointID)}`);

    metricsSocket.addEventListener("open", () => updateWsStatus(metricsBadge, true));
    metricsSocket.addEventListener("close", () => updateWsStatus(metricsBadge, false));
    metricsSocket.addEventListener("error", (e) => console.error("Metrics WebSocket error:", e));

    metricsSocket.addEventListener("message", (event) => {
        try {
            const payload = JSON.parse(event.data);
            window.dispatchEvent(new CustomEvent("metrics", { detail: payload }));
        } catch (err) {
            console.error("Failed to parse metrics WS JSON:", err);
        }
    });
}

function connectLogsSocket() {
    logsSocket = new WebSocket(`wss://${location.host}/ws/logs?endpointID=${encodeURIComponent(window.endpointID)}`);

    logsSocket.addEventListener("open", () => updateWsStatus(logsBadge, true));
    logsSocket.addEventListener("close", () => updateWsStatus(logsBadge, false));
    logsSocket.addEventListener("error", (e) => console.error("Logs WebSocket error:", e));

    logsSocket.addEventListener("message", (event) => {
        try {
            const payload = JSON.parse(event.data);
            window.dispatchEvent(new CustomEvent("logs", { detail: payload }));
        } catch (err) {
            console.error("Failed to parse logs WS JSON:", err);
        }
    });
}

function connectEventsSocket() {
    eventsSocket = new WebSocket(`wss://${location.host}/ws/events?endpointID=${encodeURIComponent(window.endpointID)}`);

    eventsSocket.addEventListener("open", () => {
        console.log("Events WebSocket connected!");
        updateWsStatus(eventsBadge, true);
    });
    eventsSocket.addEventListener("close", () => updateWsStatus(eventsBadge, false));
    eventsSocket.addEventListener("error", (e) => console.error("Events WebSocket error:", e));

    eventsSocket.addEventListener("message", (event) => {
        try {
            const payload = JSON.parse(event.data);
            window.dispatchEvent(new CustomEvent("events", { detail: payload }));
        } catch (err) {
            console.error("Failed to parse events WS JSON:", err);
        }
    });
}

function connectAlertsSocket() {
    eventsSocket = new WebSocket(`wss://${location.host}/ws/alerts?endpointID=${encodeURIComponent(window.endpointID)}`);

    eventsSocket.addEventListener("open", () => updateWsStatus(alertsBadge, true));
    eventsSocket.addEventListener("close", () => updateWsalerts(alertsBadge, false));
    eventsSocket.addEventListener("error", (e) => console.error("Events WebSocket error:", e));

    eventsSocket.addEventListener("message", (event) => {
        try {
            const payload = JSON.parse(event.data);
            window.dispatchEvent(new CustomEvent("alerts", { detail: payload }));
        } catch (err) {
            console.error("Failed to parse events WS JSON:", err);
        }
    });
}

// === Badge updater helper ===
function updateWsStatus(badge, connected) {
    if (!badge) return;
    badge.classList.remove("text-green-500", "text-red-500");

    badge.textContent = connected ? "Connected" : "Disconnected";
    badge.classList.add(connected ? "text-green-500" : "text-red-500");
}


export function initWebSockets(endpointID) {
    window.endpointID = endpointID;
    connectMetricsSocket();
    connectLogsSocket();
    connectEventsSocket();
    connectAlertsSocket();
}