// File: public/js/activity-stream.js

import { escapeHTML } from "./format.js";
import { gosightFetch } from "./api.js";

const el = document.getElementById("activity-feed");
const typeFilter = document.getElementById("filter-type");
const scopeFilter = document.getElementById("filter-scope");
const searchInput = document.getElementById("search-input");

let allActivity = [];
const MAX_ROWS = 100;
const TIME_WINDOW_MS = 30 * 1000; // 30 seconds

export async function loadActivity() {
  try {
    const [eventsRes, logsRes, alertsRes] = await Promise.all([
      gosightFetch("/api/v1/events").then(res => res.ok ? res.json() : []),
      gosightFetch("/api/v1/logs").then(res => res.ok ? res.json() : []),
      gosightFetch("/api/v1/alerts").then(res => res.ok ? res.json() : []),
    ]);

    const normalize = (arr, type) =>
      Array.isArray(arr) ? arr.map((e) => ({
        id: e.id || e.entry_id || e.alert_id || e.timestamp,
        type,
        scope: e.scope || e.meta?.scope || "unknown",
        level: e.level || "info",
        message: e.message || e.summary || e.log || e.log_message || "—",
        timestamp: new Date(e.timestamp || e.time || e.ts || new Date()),
        meta: e.meta || e.tags || {},
      })) : [];

    allActivity = [
      ...normalize(eventsRes, "event"),
      ...normalize(logsRes, "log"),
      ...normalize(alertsRes, "alert")
    ].sort((a, b) => b.timestamp - a.timestamp);

    renderActivityNodes();
  } catch (err) {
    console.error("Failed to load activity:", err);
    el.innerHTML = `<div class="p-4 text-red-500">Failed to load activity feed.</div>`;
  }
}

function clusterActivity(entries) {
  const nodes = [];
  entries.forEach((entry) => {
    const match = nodes.find(node => {
      const delta = Math.abs(node.timestamp - entry.timestamp);
      return delta <= TIME_WINDOW_MS && node.context === extractContextKey(entry);
    });
    if (match) {
      match.entries.push(entry);
      match.timestamp = new Date(Math.max(match.timestamp, entry.timestamp));
    } else {
      nodes.push({
        timestamp: entry.timestamp,
        context: extractContextKey(entry),
        entries: [entry],
      });
    }
  });
  return nodes.sort((a, b) => b.timestamp - a.timestamp);
}

function extractContextKey(entry) {
  return entry.meta?.container_id || entry.meta?.host_id || entry.meta?.hostname || entry.meta?.service || "unknown";
}

function renderActivityNodes() {
  el.innerHTML = "";

  const filtered = allActivity.filter((e) => {
    const typeMatch = !typeFilter.value || e.type === typeFilter.value;
    const scopeMatch = !scopeFilter.value || e.scope === scopeFilter.value;
    const searchMatch = !searchInput.value || (e.message || "").toLowerCase().includes(searchInput.value.toLowerCase());
    return typeMatch && scopeMatch && searchMatch;
  });

  if (filtered.length === 0) {
    el.innerHTML = `<div class="p-4 text-sm text-gray-400">No matching activity found.</div>`;
    return;
  }

  const clustered = clusterActivity(filtered).slice(0, MAX_ROWS);

  clustered.forEach((node, idx) => {
    const highest = node.entries.find(e => e.level === "error") || node.entries.find(e => e.level === "warning") || node.entries[0];
    const timeStr = node.timestamp.toLocaleTimeString();
    const rowShade = idx % 2 === 0 ? "bg-gray-50 dark:bg-gray-800" : "bg-white dark:bg-gray-900";
    const icon = { alert: "🚨", event: "🧠", log: "📄" }[highest.type] || "📌";

    el.innerHTML += `
      <div class="px-4 py-3 space-y-2 ${rowShade} border-b border-gray-200 dark:border-gray-700">
        <div class="flex items-center gap-3 text-sm text-gray-800 dark:text-gray-100">
          <span class="text-lg">${icon}</span>
          <span class="font-semibold">${escapeHTML(highest.message)}</span>
          <span class="text-xs text-gray-500 ml-auto">${escapeHTML(timeStr)}</span>
        </div>
        <div class="ml-6 text-xs text-gray-400 dark:text-gray-500">
          ${escapeHTML(extractContextKey(highest))}
        </div>
        <div class="ml-6 space-y-1">
          ${node.entries.map(e => `
            <div class="text-sm text-gray-700 dark:text-gray-300">
              <span class="font-medium">[${e.level.toUpperCase()}]</span>
              ${escapeHTML(e.message)}
            </div>`).join("")}
        </div>
      </div>
    `;
  });
}

[typeFilter, scopeFilter, searchInput].forEach(el => {
  el.addEventListener("input", renderActivityNodes);
});

loadActivity();
