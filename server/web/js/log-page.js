import { gosightFetch } from "./api.js";
document.addEventListener("DOMContentLoaded", () => {
  const searchForm = document.getElementById("log-search-form");
  const resultsTable = document.getElementById("log-results");
  const cursorDisplay = document.getElementById("cursor-pos");
  const expandedLogKeys = new Set();
  let currentCursor = null;
  let previousCursors = [];

  async function fetchLogs() {
    const params = new URLSearchParams();

    const keyword = document.getElementById("filter-keyword").value.trim();
    const levels = Array.from(document.querySelectorAll(".filter-level-option:checked")).map(el => el.value);
    levels.forEach(v => params.append("level", v));

    const categories = Array.from(document.querySelectorAll(".filter-category-option:checked")).map(el => el.value);
    categories.forEach(v => params.append("category", v));

    const source = document.getElementById("filter-source").value;
    const container = document.getElementById("container-name")?.value || "";
    const endpoint = document.getElementById("endpoint-name")?.value || "";
    const app = document.getElementById("app-name")?.value || "";
    const start = document.getElementById("start-time").value;
    const end = document.getElementById("end-time").value;

    if (keyword) params.append("contains", keyword);
    levels.forEach(level => params.append("level", level));
    categories.forEach(cat => params.append("category", cat));
    if (source) params.append("source", source);
    if (container) params.append("container", container);
    if (endpoint) params.append("endpoint", endpoint);
    if (app) params.append("app", app);
    if (start) params.append("start", new Date(start).toISOString());
    if (end) params.append("end", new Date(end).toISOString());
    if (currentCursor) params.append("cursor", currentCursor);

    params.append("limit", 50);
    params.append("order", "desc");

    try {
      console.log(`/api/v1/logs?${params.toString()}`);
      const res = await gosightFetch(`/api/v1/logs?${params.toString()}`);
      const data = await res.json();
      renderLogs(data.logs);
    } catch (err) {
      console.error("Failed to fetch logs", err);
    }
  }
  function getLogKey(log) {
    return `${log.timestamp}_${log.message}`;
  }
  function renderLogs(logs) {
    resultsTable.innerHTML = "";

    if (logs.length === 0) {
      resultsTable.innerHTML = `<tr><td colspan="7" class="text-center py-4 text-gray-500">No logs found</td></tr>`;

      return;
    }
    document.getElementById("log-count").textContent = logs.length;
    logs.forEach((log, i) => {
      const timestamp = new Date(log.timestamp).toLocaleString();

      const row = document.createElement("tr");
      row.className = "border-b border-gray-200 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-800";

      row.innerHTML = `
          <td class="px-4 py-2 text-xs text-gray-500 whitespace-nowrap">${timestamp}</td>
          <td class="px-4 py-2 text-xs font-medium text-${getColor(log.level)}-600">${log.level}</td>
          <td class="px-4 py-2 text-sm">${log.source || ""}</td>
          <td class="px-4 py-2 text-sm">
  ${log.tags?.container_name || log.tags?.hostname || ""}
</td>
          <td class="px-4 py-2 text-xs text-gray-700 dark:text-gray-200 font-mono truncate max-w-[400px]">
  ${sanitize(log.message)}
</td>
          <td class="px-4 py-2 text-sm">${log.meta?.user || ""}</td>
          <td class="px-4 py-2 text-sm">
            <button class="text-blue-600 hover:underline text-xs expand-log" data-log-index="${i}">Details</button>
          </td>
        `;

      const detailsRow = document.createElement("tr");
      detailsRow.className = "hidden bg-gray-50 dark:bg-gray-800";
      detailsRow.innerHTML = `
      <td colspan="7" class="px-4 py-2 text-sm text-gray-600 dark:text-gray-300">
        <div class="mb-2">
<div class="mb-4">
  <div class="text-xs uppercase font-semibold text-gray-500 dark:text-gray-400 mb-1">Full Message</div>
  <pre class="text-sm text-gray-800 dark:text-gray-100 font-mono whitespace-pre-wrap break-words bg-gray-50 dark:bg-gray-900 border border-gray-200 dark:border-gray-700 p-4 rounded-lg">
    ${sanitize(log.message)}
  </pre>
</div>
        <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
          ${renderSection("Tags", log.tags)}
          ${renderSection("Meta", filterOutExtra(log.meta))}
          ${renderSection("Fields", log.fields)}
        </div>
      </td>
    `;
      //${renderSection("Extra", log.meta?.extra)}
      resultsTable.appendChild(row);
      resultsTable.appendChild(detailsRow);
      if (expandedLogKeys.has(getLogKey(log))) {
        detailsRow.classList.remove("hidden");
      }

      row.querySelector(".expand-log").addEventListener("click", (e) => {
        e.preventDefault();
        e.stopPropagation();

        const logKey = getLogKey(log);
        const isOpen = expandedLogKeys.has(logKey);

        // Close all expanded rows first
        document.querySelectorAll("tr.details-expanded").forEach(row => row.classList.add("hidden"));
        expandedLogKeys.clear();

        if (!isOpen) {
          // Reopen the clicked on
          detailsRow.classList.remove("hidden");
          detailsRow.classList.add("details-expanded");
          expandedLogKeys.add(logKey);
        }
      });
    });

    currentCursor = logs[logs.length - 1].timestamp;
    cursorDisplay.textContent = currentCursor;
  }
  function filterOutExtra(meta) {
    if (!meta || typeof meta !== "object") return {};
    const copy = { ...meta };
    delete copy.extra;
    return copy;
  }
  function formatTimestampForDisplay(isoString) {
    const date = new Date(isoString);
    return date.toLocaleString(undefined, {
      dateStyle: "medium",
      timeStyle: "short"
    });
  }
  function updateURLFromForm() {
    const params = new URLSearchParams();

    const keyword = document.getElementById("filter-keyword").value.trim();
    if (keyword) params.set("contains", keyword);

    document.querySelectorAll(".filter-level-option:checked").forEach(opt => {
      params.append("level", opt.value);
    });

    document.querySelectorAll(".filter-category-option:checked").forEach(opt => {
      params.append("category", opt.value);
    });

    const mappings = {
      source: "filter-source",
      container: "container-name",
      endpoint: "endpoint-name",
      app: "app-name",
      start: "start-time",
      end: "end-time"
    };

    for (const [key, id] of Object.entries(mappings)) {
      const val = document.getElementById(id)?.value;
      if (val) {
        if (key === "start" || key === "end") {
          const iso = new Date(val).toISOString(); // e.g. 2025-05-06T14:51:00.000Z
          params.set(key, iso);
        } else {
          params.set(key, val);
        }
      }
    }

    const builtInKeys = new Set(["level", "category", "source", "endpoint", "app", "start", "end", "contains"]);

    document.querySelectorAll("#tag-filters span").forEach(span => {
      const raw = span.textContent.replace("×", "").trim();
      const [k, v] = raw.split(":");
      if (k && v && !builtInKeys.has(k)) {
        params.append(`tag_${k}`, v);
      }
    });

    if (currentCursor) {
      params.set("cursor", currentCursor);
    }

    const newURL = `${window.location.pathname}?${params.toString()}`;
    history.replaceState(null, "", newURL);
  }

  function renderSection(title, obj) {
    if (!obj || typeof obj !== "object" || Object.keys(obj).length === 0) return "";

    const entries = Object.entries(obj);
    const rows = entries.map(([key, val], i) => {
      const stringVal = typeof val === "object" ? JSON.stringify(val, null, 2) : String(val);
      const safeKey = sanitize(key);
      const isTruncated = stringVal.length > 60;
      const displayShort = isTruncated ? sanitize(stringVal.slice(0, 60) + "…") : sanitize(stringVal);
      const fullValue = stringVal; // raw for tooltip and copy
      const bg = i % 2 === 0 ? "bg-white dark:bg-gray-900" : "bg-gray-50 dark:bg-gray-800";

      return `
            <tr class="group ${bg} border-b border-gray-100 dark:border-gray-700 hover:bg-gray-100 dark:hover:bg-gray-700 align-top">
              <td class="px-2 py-1 font-medium text-xs text-gray-500 whitespace-nowrap align-top border-l-4 border-blue-200 dark:border-blue-800">${safeKey}</td>
              <td class="px-2 py-1 text-xs break-all text-gray-800 dark:text-gray-100 font-mono align-top">
                <span title="${fullValue}" class="inline-block">${displayShort}</span>
                ${isTruncated ? `
                  <button class="ml-2 text-gray-400 hover:text-gray-700 dark:hover:text-white text-xs copy-btn" data-copy="${fullValue}" title="Copy full value">📋</button>
                ` : ""}
                <button
                  class="ml-2 inline-block text-blue-600 hover:underline text-xs font-medium"
                  data-filter-key="${safeKey}" data-filter-value="${fullValue}">
                  + Add Filter
                </button>
              </td>
            </tr>
          `;
    }).join("");

    return `
          <div class="mb-4 border rounded border-gray-200 dark:border-gray-700">
            <div class="px-3 py-2 border-b border-gray-100 dark:border-gray-700 font-semibold text-xs uppercase text-gray-500 dark:text-gray-400 bg-gray-50 dark:bg-gray-800">
              ${title}
            </div>
            <table class="w-full text-left text-xs table-fixed">${rows}</table>
          </div>
        `;
  }

  function loadFormFromURL() {
    const params = new URLSearchParams(window.location.search);

    const setInput = (id, value) => {
      const el = document.getElementById(id);
      if (el) el.value = value;
    };

    const setFlowbiteMultiSelect = (checkboxClass, values) => {
      document.querySelectorAll(`.${checkboxClass}`).forEach(checkbox => {
        checkbox.checked = values.includes(checkbox.value);
      });
    };

    // Single fields
    setInput("filter-keyword", params.get("contains") || "");
    setInput("filter-source", params.get("source") || "");
    setInput("container-name", params.get("container") || "");
    setInput("endpoint-name", params.get("endpoint") || "");
    setInput("app-name", params.get("app") || "");
    setInput("start-time", params.get("start") ? new Date(params.get("start")).toISOString().slice(0, 16) : "");
    setInput("end-time", params.get("end") ? new Date(params.get("end")).toISOString().slice(0, 16) : "");

    // Multi-select
    setFlowbiteMultiSelect("filter-level-option", params.getAll("level"));
    setFlowbiteMultiSelect("filter-category-option", params.getAll("category"));

    // Tags
    const tagContainer = document.getElementById("tag-filters");
    tagContainer.innerHTML = ""; // clear existing
    for (const [key, value] of params.entries()) {
      if (key.startsWith("tag_")) {
        const tagKey = key.slice(4);
        const span = document.createElement("span");
        span.className = "inline-flex items-center px-3 py-1 rounded-sm text-sm font-medium bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-100";
        span.innerHTML = `${tagKey}:${value} <button class="ml-1 text-xs remove-tag" type="button">&times;</button>`;
        tagContainer.appendChild(span);
      }
    }

    // Cursor
    currentCursor = params.get("cursor") || null;
  }
  function renderActiveFiltersFromForm() {
    const tagContainer = document.getElementById("tag-filters");
    tagContainer.innerHTML = ""; // clear all first

    const mappings = {
      contains: "filter-keyword",
      source: "filter-source",
      container: "container-name",
      endpoint: "endpoint-name",
      app: "app-name",
      start: "start-time",
      end: "end-time"
    };

    for (const [key, id] of Object.entries(mappings)) {
      const val = document.getElementById(id)?.value?.trim();
      if (val) {
        addTagPill(key, val);
      }
    }

    const getChecked = cls =>
      Array.from(document.querySelectorAll(`.${cls}:checked`)).map(el => el.value);

    getChecked("filter-level-option").forEach(v => addTagPill("level", v));
    getChecked("filter-category-option").forEach(v => addTagPill("category", v));

    // Add custom tags already in tag-filters
    const existingSpans = document.querySelectorAll("#tag-filters span[data-custom-tag]");
    existingSpans.forEach(span => tagContainer.appendChild(span));
  }
  function addTagPill(key, value) {
    const span = document.createElement("span");
    span.className =
      "inline-flex items-center text-sm font-medium me-2 px-3 py-1 rounded-sm bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-100 hover:bg-blue-200 dark:hover:bg-blue-800";

    const display = (key === "start" || key === "end")
      ? formatTimestampForDisplay(value)
      : `${key}:${value}`;

    const label = document.createElement("span");
    label.textContent = (key === "start" || key === "end") ? `${key}: ${display}` : `${key}:${value}`;
    span.appendChild(label);

    const button = document.createElement("button");
    button.type = "button";
    button.className = "ms-2 inline-flex items-center justify-center w-4 h-4 text-xs text-blue-800 hover:text-blue-900 dark:text-blue-200 dark:hover:text-white focus:outline-none remove-tag";
    button.title = "Remove filter";
    button.innerHTML = "&times;";
    span.appendChild(button);

    span.dataset.customTag = "true";
    document.getElementById("tag-filters").appendChild(span);
  }


  async function populateEndpointDropdown() {
    const dropdown = document.getElementById("endpoint-dropdown");
    const input = document.getElementById("endpoint-name");

    let allItems = [];

    try {
      const [hosts, containers] = await Promise.all([
        fetch("/api/v1/endpoints/hosts").then(r => r.json()),
        fetch("/api/v1/endpoints/containers").then(r => r.json()),
      ]);

      allItems = [
        { label: "Hosts", items: hosts.map(h => h.hostname).filter(Boolean) },
        { label: "Containers", items: containers.map(c => c.Name ?? "").filter(Boolean) },
      ];
    } catch (err) {
      console.error("Failed to load endpoints", err);
    }

    // Filter logic
    input.addEventListener("input", () => {
      const val = input.value.toLowerCase();
      dropdown.innerHTML = "";
      dropdown.classList.remove("hidden");

      allItems.forEach(group => {
        const matched = group.items.filter(item => item.toLowerCase().includes(val));
        if (matched.length === 0) return;

        const groupLabel = document.createElement("div");
        groupLabel.className = "px-3 py-1 text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase";
        groupLabel.textContent = group.label;
        dropdown.appendChild(groupLabel);

        matched.forEach(item => {
          const opt = document.createElement("div");
          opt.className = "cursor-pointer px-4 py-2 hover:bg-gray-100 dark:hover:bg-gray-700 text-gray-800 dark:text-white";
          opt.textContent = item;
          opt.addEventListener("click", () => {
            setTimeout(() => {
              input.value = item;
              dropdown.classList.add("hidden");
              renderActiveFiltersFromForm(); // update pill after value is set
              updateURLFromForm();
              fetchLogs();
            }, 10); // short delay to ensure value is actually set
          });
          dropdown.appendChild(opt);
        });
      });

      if (dropdown.innerHTML === "") {
        dropdown.classList.add("hidden");
      }
    });

    // Hide dropdown on blur
    input.addEventListener("blur", () => setTimeout(() => dropdown.classList.add("hidden"), 150));


  }

  function sanitize(str) {
    const div = document.createElement("div");
    div.innerText = str;
    return div.innerHTML;
  }

  function getColor(level) {
    switch ((level || "").toLowerCase()) {
      case "error": return "red";
      case "warning": return "yellow";
      case "info": return "blue";
      case "debug": return "gray";
      default: return "gray";
    }
  }

  document.getElementById("next-page").addEventListener("click", (e) => {
    e.preventDefault();
    if (currentCursor) {
      previousCursors.push(currentCursor);
      fetchLogs();
      updateURLFromForm();
    }
  });

  document.getElementById("prev-page").addEventListener("click", (e) => {
    e.preventDefault();
    if (previousCursors.length > 0) {
      currentCursor = previousCursors.pop();
      fetchLogs();
      updateURLFromForm();
    }
  });

  document.getElementById("reset-search").addEventListener("click", (e) => {
    e.preventDefault();
    searchForm.reset();
    currentCursor = null;
    previousCursors = [];
    fetchLogs();
    updateURLFromForm();
  });

  searchForm.addEventListener("submit", (e) => {
    e.preventDefault();
    currentCursor = null;
    previousCursors = [];
    fetchLogs();
    renderActiveFiltersFromForm();
    updateURLFromForm();
  });

  document.getElementById("tag-filters").addEventListener("click", (e) => {
    if (!e.target.classList.contains("remove-tag")) return;

    const pill = e.target.closest("span");
    if (!pill) return;

    // Parse the key:value
    const raw = pill.textContent.replace("×", "").trim();
    const [key, value] = raw.split(":");

    // Remove matching form values
    if (key === "level") {
      document.querySelectorAll(".filter-level-option:checked").forEach(cb => {
        if (cb.value === value) cb.checked = false;
      });
    } else if (key === "category") {
      document.querySelectorAll(".filter-category-option:checked").forEach(cb => {
        if (cb.value === value) cb.checked = false;
      });
    } else {
      // Handle individual fields like source, endpoint, app, etc.
      const id = {
        contains: "filter-keyword",
        source: "filter-source",
        container: "container-name",
        endpoint: "endpoint-name",
        app: "app-name",
        start: "start-time",
        end: "end-time"
      }[key];
      if (id) {
        const input = document.getElementById(id);
        if (input && input.value === value) input.value = "";
      }
    }

    // Remove the pill visually
    pill.remove();

    // Trigger new search + update URL
    currentCursor = null;
    previousCursors = [];
    fetchLogs();
    updateURLFromForm();
    renderActiveFiltersFromForm(); // rebuild pills to reflect any additional changes
  });

  resultsTable.addEventListener("click", (e) => {
    if (e.target.classList.contains("copy-btn")) {
      const value = e.target.dataset.copy;
      if (value) {
        navigator.clipboard.writeText(value).then(() => {
          e.target.textContent = "✅";
          setTimeout(() => e.target.textContent = "📋", 1000);
        });
      }
    }
  });
  resultsTable.addEventListener("click", (e) => {
    const tag = e.target?.dataset?.tag;
    if (!tag) return;

    const [key, value] = tag.split(":");
    if (!key || !value) return;

    const tagContainer = document.getElementById("tag-filters");
    if ([...tagContainer.querySelectorAll("span")].some(span => span.textContent.includes(`${key}:${value}`))) return;

    const span = document.createElement("span");
    span.className = "inline-flex items-center px-3 py-1 rounded-sm text-sm font-medium bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-100";
    span.innerHTML = `${key}:${value} <button class="ml-1 text-xs remove-tag" type="button">&times;</button>`;
    tagContainer.appendChild(span);
    updateURLFromForm();
  });
  resultsTable.addEventListener("click", (e) => {
    if (e.target.matches("[data-filter-key]")) {
      e.preventDefault();

      const key = e.target.dataset.filterKey;
      const value = e.target.dataset.filterValue;
      if (!key || !value) return;

      const tagContainer = document.getElementById("tag-filters");
      if ([...tagContainer.querySelectorAll("span")].some(span => span.textContent.includes(`${key}:${value}`))) return;

      const span = document.createElement("span");
      span.className = "inline-flex items-center px-3 py-1 rounded-sm text-sm font-medium bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-100";
      span.innerHTML = `${key}:${value} <button class="ml-1 text-xs remove-tag" type="button">&times;</button>`;
      tagContainer.appendChild(span);
      currentCursor = null;
      previousCursors = [];
      fetchLogs();
      updateURLFromForm();
    }
  });
  function bindFormAutoTags() {
    const inputs = [
      "filter-keyword", "filter-source", "container-name", "endpoint-name",
      "app-name", "start-time", "end-time"
    ];

    inputs.forEach(id => {
      const el = document.getElementById(id);
      if (el) {
        el.addEventListener("change", () => {
          renderActiveFiltersFromForm();
          updateURLFromForm();
          fetchLogs();
        });
      }
    });

    // Multi-select checkboxes (levels/categories)
    document.querySelectorAll(".filter-level-option, .filter-category-option").forEach(cb => {
      cb.addEventListener("change", () => {
        renderActiveFiltersFromForm();
        updateURLFromForm();
        fetchLogs();
      });
    });

    // Enter key on any input triggers submit
    document.querySelectorAll("#log-search-form input").forEach(input => {
      input.addEventListener("keypress", (e) => {
        if (e.key === "Enter") {
          e.preventDefault();
          currentCursor = null;
          previousCursors = [];
          fetchLogs();
          renderActiveFiltersFromForm();
          updateURLFromForm();
        }
      });
    });
  }
  loadFormFromURL();
  populateEndpointDropdown();
  bindFormAutoTags();
  fetchLogs();
});
