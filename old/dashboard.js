import {
  renderHostSection,
  renderContainerSections
} from './ui.js';

const ws = new WebSocket("ws://" + location.host + "/ws");
const lastUpdate = document.getElementById("lastUpdate");
const container = document.getElementById("metricsTables");

function groupMetrics(metrics, meta) {
  const groupedMeta = {};
  for (const key in meta) {
    if (!key.startsWith("container_podman_")) continue;
    const match = key.match(/^container_podman_(.+)_([^_]+)$/);
    if (!match) continue;
    const [, type, name] = match;
    groupedMeta[name] = groupedMeta[name] || [];
    groupedMeta[name].push({ name: type, value: meta[key] });
  }

  const knownContainers = new Set(Object.keys(groupedMeta));

  const groupedMetrics = {
    HOST: [],
    INTERFACES: {},
    TOTALS: [],
    CONTAINERS: {},
  };

  for (const key in metrics) {
    const value = metrics[key];

    if (key.startsWith("container_podman_")) {
      const lastUnderscore = key.lastIndexOf('_');
      const metricName = key.slice("container_podman_".length, lastUnderscore);
      const containerName = key.slice(lastUnderscore + 1);

      groupedMetrics.CONTAINERS[containerName] ||= [];
      groupedMetrics.CONTAINERS[containerName].push({ name: metricName, full: key, value });

    } else if (key.startsWith("net_rx_bytes_") || key.startsWith("net_tx_bytes_")) {
      const iface = key.split("_").slice(3).join("_");
      if (knownContainers.has(iface)) {
        groupedMetrics.CONTAINERS[iface] ||= [];
        groupedMetrics.CONTAINERS[iface].push({ name: key.split("_").slice(0, 3).join("_"), full: key, value });
      } else {
        groupedMetrics.INTERFACES[iface] ||= [];
        groupedMetrics.INTERFACES[iface].push({ name: key.split("_").slice(0, 3).join("_"), full: key, value });
      }

    } else if (key === "net_rx_bytes_total" || key === "net_tx_bytes_total") {
      groupedMetrics.TOTALS.push({ name: key, full: key, value });

    } else {
      groupedMetrics.HOST.push({ name: key, full: key, value });
    }
  }

  return { groupedMetrics, groupedMeta };
}

ws.onmessage = (event) => {
  try {
    const { metrics, thresholds, meta } = JSON.parse(event.data);
    const { groupedMetrics, groupedMeta } = groupMetrics(metrics, meta);

    container.innerHTML = "";

    const host = renderHostSection(
      groupedMetrics.HOST,
      groupedMetrics.INTERFACES,
      groupedMetrics.TOTALS,
      thresholds
    );
    container.appendChild(host);

    const containerHTML = renderContainerSections(groupedMetrics.CONTAINERS, groupedMeta, thresholds);
    container.insertAdjacentHTML("beforeend", containerHTML);

    lastUpdate.textContent = "Last updated: " + new Date().toLocaleTimeString();
  } catch (err) {
    console.error("Render error:", err);
  }
};

ws.onerror = (e) => console.error("WebSocket error:", e);
