export let alertsSocket = null;

function updateWsStatus(badgeEl, online) {
    if (!badgeEl) return;
    badgeEl.classList.toggle("bg-red-600", !online);
    badgeEl.classList.toggle("bg-green-500", online);
    badgeEl.title = online ? "Connected to alert stream" : "Disconnected from alert stream";
}

function ensureAlertBanner() {
    if (document.getElementById("global-alert-banner")) return;

    const banner = document.createElement("div");
    banner.id = "global-alert-banner";
    banner.className = "hidden bg-red-100 text-red-800 border border-red-300 px-4 py-3 text-sm text-center fixed top-0 left-0 right-0 z-50";

    banner.innerHTML = `
        <span id="global-alert-message">Alert fired</span>
        <a href="/alerts" class="ml-2 underline text-sm font-semibold text-red-700 hover:text-red-900">
            View Alerts
        </a>
    `;
    document.body.prepend(banner);
}

function showAlertBanner(message) {
    ensureAlertBanner();
    const banner = document.getElementById("global-alert-banner");
    const msg = document.getElementById("global-alert-message");
    const spacer = document.getElementById("alert-spacer");

    msg.textContent = message;
    banner.classList.remove("hidden");
    spacer.style.height = "48px"; // adjust to match banner height

    setTimeout(() => {
        banner.classList.add("hidden");
        spacer.style.height = "0px";
    }, 12000);
}


function setBellIndicator(count) {
    const dot = document.getElementById("alert-indicator");
    if (!dot) return;

    if (count > 0) {
        dot.classList.remove("hidden");
    } else {
        dot.classList.add("hidden");
    }
}



async function fetchActiveAlerts() {
    try {
        const res = await fetch("/api/v1/alerts/active");
        if (!res.ok) return;

        const alerts = await res.json();
        if (Array.isArray(alerts) && alerts.length > 0) {
            setBellIndicator(alerts.length);
            const first = alerts[0];
            showAlertBanner(`${first.message || first.name}`);
        }
    } catch (err) {
        console.error("Failed to fetch active alerts:", err);
    }
}

export function connectAlertsSocket(alertsBadge = null) {
    if (alertsSocket && alertsSocket.readyState === WebSocket.OPEN) return;

    alertsSocket = new WebSocket(`wss://${location.host}/ws/alerts`);

    alertsSocket.addEventListener("open", () => updateWsStatus(alertsBadge, true));
    alertsSocket.addEventListener("close", () => updateWsStatus(alertsBadge, false));
    alertsSocket.addEventListener("error", (e) => console.error("Alerts WebSocket error:", e));

    alertsSocket.addEventListener("message", (event) => {
        if (!event.data || event.data === "ping") return;

        try {
            const payload = JSON.parse(event.data);
            window.dispatchEvent(new CustomEvent("alerts", { detail: payload }));
        } catch (err) {
            console.error("Failed to parse alerts WebSocket message:", err);
        }
    });
}



window.addEventListener("alerts", (e) => {
    const alert = e.detail;
    console.log("Received alert:", alert);
    if (alert?.state == "firing" && alert.message) {
        showAlertBanner(`${alert.message}`);
        fetchActiveAlerts(); // refresh count

    }
});

document.addEventListener("DOMContentLoaded", () => {
    connectAlertsSocket();
    fetchActiveAlerts();
});
