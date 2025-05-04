// /js/alert_history.js
import { gosightFetch } from "./api.js";

let alerts = [];
let currentPage = 1;
const pageSize = 20;
let sortField = "first_fired";
let sortAsc = false;
let totalCount = 0;

const tbody = document.getElementById("alert-history-body");
const searchInput = document.getElementById("alert-history-search");
const paginationInfo = document.getElementById("pagination-info");
const stateFilter = document.getElementById("filter-state");
const levelFilter = document.getElementById("filter-level");
const activeTagFilters = document.getElementById("active-tag-filters");

async function loadAlerts() {
    const params = new URLSearchParams(window.location.search);
    const ruleFilter = params.get("rule_id") || "";
    const state = params.get("state") || "";
    const level = params.get("level") || "";
    const tag = params.getAll("tag");

    sortField = params.get("sort") || sortField;
    sortAsc = params.get("order") === "asc";
    currentPage = parseInt(params.get("page")) || 1;

    stateFilter.value = state;
    levelFilter.value = level;
    searchInput.value = ruleFilter;

    const sort = sortField;
    const order = sortAsc ? "asc" : "desc";
    let url = `/api/v1/alerts?limit=${pageSize}&page=${currentPage}&sort=${sort}&order=${order}`;
    if (ruleFilter) url += `&rule_id=${encodeURIComponent(ruleFilter)}`;
    if (state) url += `&state=${encodeURIComponent(state)}`;
    if (level) url += `&level=${encodeURIComponent(level)}`;
    tag.forEach(t => url += `&tag=${encodeURIComponent(t)}`);

    const res = await fetch(url);
    totalCount = parseInt(res.headers.get("X-Total-Count"), 10) || 0;
    alerts = await res.json();
    renderTable();
    renderTagChips();
    updateFilterVisuals();
}

function renderTagChips() {
    const params = new URLSearchParams(window.location.search);
    const tags = params.getAll("tag");
    activeTagFilters.innerHTML = "";

    tags.forEach(tagKV => {
        const chip = document.createElement("span");
        chip.className = `
  tag-filter-btn inline-block
  bg-blue-100 dark:bg-gray-700
  text-blue-800 dark:text-gray-200
  px-2 py-2 mr-1 mb-1 rounded

  cursor-pointer text-sm
  transition transform hover:scale-105
`.trim();

        chip.textContent = tagKV;

        const closeBtn = document.createElement("button");
        closeBtn.className = "ml-2 text-gray-500 hover:text-gray-800 dark:hover:text-white text-lg leading-none";
        closeBtn.textContent = "×";
        closeBtn.addEventListener("click", (e) => {
            e.stopPropagation();
            removeTagFilter(tagKV);
        });

        chip.appendChild(closeBtn);
        activeTagFilters.appendChild(chip);
    });
}

function renderTable() {
    tbody.innerHTML = "";

    for (const a of alerts) {
        const tr = document.createElement("tr");
        tr.className = "alert-row transition-colors hover:bg-gray-50 dark:hover:bg-gray-800 border-t border-gray-200 dark:border-gray-700 cursor-pointer";

        tr.innerHTML = `
<td class="px-4 py-3 font-medium text-blue-700 dark:text-blue-400">
<a href="#" class="rule-link hover:underline" data-rule="${a.rule_id}">${a.rule_id}</a>
</td>
<td class="px-4 py-3">${badge(a.state)}</td>
<td class="px-4 py-3">${a.level}</td>
<td class="px-4 py-3">${a.target || "-"}</td>
<td class="px-4 py-3">${a.scope || "-"}</td>
<td class="px-4 py-3">${formatDate(a.first_fired)}</td>
<td class="px-4 py-3">${formatDate(a.last_ok)}</td>
<td class="px-4 py-3 text-sm text-blue-500">[+]</td>`;

        tr.addEventListener("click", () => toggleExpandRow(tr, a));
        tbody.appendChild(tr);
    }

    const totalPages = Math.ceil(totalCount / pageSize);
    paginationInfo.textContent = `Page ${currentPage} of ${totalPages}`;
    document.querySelectorAll(".rule-link").forEach(link => {
        link.addEventListener("click", (e) => {
            e.preventDefault();
            const ruleID = link.dataset.rule;
            updateURLParam("rule_id", ruleID);
            updateURLParam("page", 1);
        });
    });
}

function toggleExpandRow(row, alert) {
    if (row.nextSibling && row.nextSibling.classList.contains("alert-detail")) {
        row.nextSibling.remove();
        return;
    }
    const detailRow = document.createElement("tr");
    detailRow.className = "alert-detail bg-gray-50 dark:bg-gray-800 text-xs";
    const labelHTML = Object.entries(alert.labels || {})
        .map(([k, v]) =>
            `<button
          class="tag-filter-btn inline-block bg-blue-100 dark:bg-gray-700 !text-blue-300 dark:text-gray-200  px-2 py-1 mr-1 mb-1 rounded
                 hover:bg-blue-300 dark:hover:bg-gray-600 cursor-pointer transition transform !hover:scale-105"
          data-tag="${k}:${v}">
          ${k}: ${v}
        </button>`
        )
    detailRow.innerHTML = `
      <td colspan="8" class="px-4 py-3 space-y-4">
        <div>
          <div class="font-semibold mb-1 text-gray-700 dark:text-gray-200">Message:</div>
          <div class="text-gray-800 dark:text-gray-100">${alert.message}</div>
        </div>
        <div>
          <div class="font-semibold mb-1 text-gray-700 dark:text-gray-200">Tags:</div>
          <div class="flex flex-wrap gap-2">${labelHTML || "(none)"}</div>
        </div>
      </td>`;
    row.after(detailRow);
    detailRow.querySelectorAll(".tag-filter-btn").forEach(btn => {
        btn.addEventListener("click", (e) => {
            e.stopPropagation();
            const tag = btn.dataset.tag;
            updateTagFilter(tag);
        });
    });
}

function updateFilterVisuals() {
    const params = new URLSearchParams(window.location.search);
    stateFilter.classList.toggle("ring-2", !!params.get("state"));
    levelFilter.classList.toggle("ring-2", !!params.get("level"));
    searchInput.classList.toggle("ring-2", !!params.get("rule_id"));
}

function formatDate(dt) {
    return dt ? new Date(dt).toLocaleString() : "-";
}

function badge(state) {
    const map = {
        firing: "bg-red-100 text-red-800",
        resolved: "bg-green-100 text-green-800"
    };
    return `<span class="px-2 py-1 rounded ${map[state] || "bg-gray-200 text-gray-800"}">${state}</span>`;
}

document.getElementById("prev-page").onclick = () => {
    if (currentPage > 1) {
        updateURLParam("page", currentPage - 1);
    }
};
document.getElementById("next-page").onclick = () => {
    const totalPages = Math.ceil(totalCount / pageSize);
    if (currentPage < totalPages) {
        updateURLParam("page", currentPage + 1);
    }
};

stateFilter.addEventListener("change", () => {
    updateURLParam("state", stateFilter.value);
    updateURLParam("page", 1);
});
levelFilter.addEventListener("change", () => {
    updateURLParam("level", levelFilter.value);
    updateURLParam("page", 1);
});

searchInput.addEventListener("input", () => {
    updateURLParam("rule_id", searchInput.value);
    updateURLParam("page", 1);
});

document.getElementById("clear-filters").addEventListener("click", () => {
    const params = new URLSearchParams(window.location.search);
    ["state", "level", "rule_id", "page", "sort", "order", "tag"].forEach(k => params.delete(k));
    window.history.replaceState({}, "", `${window.location.pathname}?${params.toString()}`);
    loadAlerts();
});

function updateURLParam(key, value) {
    const params = new URLSearchParams(window.location.search);
    if (value) {
        params.set(key, value);
    } else {
        params.delete(key);
    }
    window.history.replaceState({}, "", `${window.location.pathname}?${params.toString()}`);
    loadAlerts();
}

function exportData(format) {
    const params = new URLSearchParams(window.location.search);
    params.set("limit", 10000); // export all matching rows
    const url = `/api/v1/alerts?${params.toString()}`;

    fetch(url).then(res => res.json()).then(data => {
        const blob = new Blob([
            format === "json"
                ? JSON.stringify(data, null, 2)
                : format === "csv"
                    ? toCSV(data)
                    : toYAML(data)
        ], { type: "text/plain" });

        const a = document.createElement("a");
        a.href = URL.createObjectURL(blob);
        a.download = `alerts_export.${format}`;
        a.click();
    });
}

document.getElementById("export-json").onclick = () => exportData("json");
document.getElementById("export-yaml").onclick = () => exportData("yaml");
document.getElementById("export-csv").onclick = () => exportData("csv");

function toCSV(data) {
    if (!data.length) return "";
    const keys = Object.keys(data[0]);
    const lines = [keys.join(",")];
    for (const row of data) {
        lines.push(keys.map(k => JSON.stringify(row[k] ?? "")).join(","));
    }
    return lines.join("\n");
}

function toYAML(data) {
    return data.map(obj => {
        return Object.entries(obj).map(([k, v]) => `${k}: ${JSON.stringify(v)}`).join("\n");
    }).join("\n---\n");
}

// Add tag filter support
function updateTagFilter(tagKV) {
    const params = new URLSearchParams(window.location.search);
    params.append("tag", tagKV);
    params.set("page", 1);
    window.history.replaceState({}, "", `${window.location.pathname}?${params.toString()}`);
    loadAlerts();
}

function removeTagFilter(tagKV) {
    const params = new URLSearchParams(window.location.search);
    const remaining = params.getAll("tag").filter(t => t !== tagKV);
    params.delete("tag");
    remaining.forEach(t => params.append("tag", t));
    params.set("page", 1);
    window.history.replaceState({}, "", `${window.location.pathname}?${params.toString()}`);
    loadAlerts();
}


loadAlerts();
