import { registerTabInitializer } from "./tabs.js";

registerTabInitializer("activity", initActivityTab);

function initActivityTab() {
    fetchAndRenderActivity();


}

async function fetchAndRenderActivity() {
    try {
        const res = await fetch("/api/v1/events?limit=100");
        const events = await res.json();
        const tbody = document.getElementById("activity-log-body");
        tbody.innerHTML = "";
        for (const e of events.reverse()) {
            appendActivityRow(e);
        }
    } catch (err) {
        console.error("❌ Failed to load initial events:", err);
    }
}

function appendActivityRow(e) {
    const tbody = document.getElementById("activity-log-body");
    if (!tbody) return;

    const rowId = `event-${e.id}`;
    const detailRowId = `${rowId}-details`;

    // Main clickable row
    const tr = document.createElement("tr");
    tr.className = "group hover:bg-gray-50 dark:hover:bg-gray-700 cursor-pointer";
    tr.dataset.detailRow = detailRowId;

    tr.onclick = () => {
        const isAlreadyOpen = tr.classList.contains("expanded");

        // Collapse all others
        const expanded = tbody.querySelectorAll("tr.expanded");
        for (const openRow of expanded) {
            openRow.classList.remove("expanded", "bg-gray-100");
            const wrapper = document.getElementById(`${openRow.dataset.detailRow}-wrapper`);
            if (wrapper) {
                wrapper.style.maxHeight = "0px";
                wrapper.style.opacity = "0";
            }
        }

        if (isAlreadyOpen) return;

        // Expand this one
        tr.classList.add("expanded", "bg-gray-100");
        const wrapper = document.getElementById(`${detailRowId}-wrapper`);
        if (wrapper) {
            wrapper.style.maxHeight = wrapper.scrollHeight + "px";
            wrapper.style.opacity = "1";
        }
    };


    tr.innerHTML = `
		<td class="px-4 py-2 font-medium text-${colorClass(e.level)} capitalize flex items-center gap-2">
			<svg class="w-4 h-4 transition-transform transform group-[.expanded]:rotate-90" viewBox="0 0 20 20" fill="currentColor">
				<path fill-rule="evenodd" d="M6 6L14 10L6 14V6Z" clip-rule="evenodd" />
			</svg>
			${iconForCategory(e.category)} ${e.category}
		</td>
		<td class="px-4 py-2 whitespace-nowrap text-gray-500 dark:text-gray-400">
			${new Date(e.timestamp).toLocaleString()}
		</td>
		<td class="px-4 py-2 text-xs text-gray-600 dark:text-gray-400">${e.scope || "-"}</td>
		<td class="px-4 py-2">${e.message}</td>
	`;

    const detailTr = document.createElement("tr");
    detailTr.id = detailRowId;

    detailTr.innerHTML = `
    <td colspan="4">
      <div id="${detailRowId}-wrapper"
        class="overflow-hidden transition-all duration-300 ease-in-out max-h-0 opacity-0"
        style="will-change: max-height, opacity;">
  
        <div class="px-5 py-4 space-y-4 text-sm bg-white dark:bg-gray-900 border-t border-gray-200 dark:border-gray-700 shadow-md rounded-b-md">
  
          <div class="grid grid-cols-2 md:grid-cols-4 gap-y-2">
            <div>
              <span class="text-gray-500 dark:text-gray-400 font-medium">Level</span>
              <div><span class="px-2 py-0.5 rounded bg-yellow-100 text-yellow-800 text-xs font-semibold">${e.level}</span></div>
            </div>
            <div>
              <span class="text-gray-500 dark:text-gray-400 font-medium">Scope</span>
              <div>${e.scope || "-"}</div>
            </div>
            <div>
              <span class="text-gray-500 dark:text-gray-400 font-medium">Target</span>
              <div class="font-mono text-blue-600 dark:text-blue-400">${e.target || "-"}</div>
            </div>
            <div>
              <span class="text-gray-500 dark:text-gray-400 font-medium">Source</span>
              <div>${e.source || "-"}</div>
            </div>
          </div>
  
          <div>
            <span class="text-gray-500 dark:text-gray-400 font-medium block mb-1">Message</span>
            <div class="italic text-base text-gray-800 dark:text-gray-200">${e.message || "-"}</div>
          </div>
  
          ${Object.keys(e.meta || {}).length > 0 ? `
            <div>
              <span class="text-gray-500 dark:text-gray-400 font-medium block mb-2">Metadata</span>
              <div class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-x-6 gap-y-1 text-sm text-gray-700 dark:text-gray-400">
                ${Object.entries(e.meta).map(([k, v]) => `
                  <div><span class="text-gray-500 dark:text-gray-400 font-medium">${k}:</span> ${v}</div>
                `).join("")}
              </div>
            </div>
          ` : ""}
  
        </div>
      </div>
    </td>
  `;
  
  

    tbody.appendChild(tr);
    tbody.appendChild(detailTr);

    const wrapper = document.getElementById(`${detailRowId}-wrapper`);
    if (wrapper) {
        wrapper.style.maxHeight = "0px";
        wrapper.style.opacity = "0";
    }


    // Trim off oldest entries
    while (tbody.children.length > 400) {
        tbody.removeChild(tbody.firstChild);   // main row
        tbody.removeChild(tbody.firstChild);   // detail row
    }
}


function renderEventDetails(e) {
    const lines = [];

    if (e.level) lines.push(`Level: ${e.level}`);
    if (e.category) lines.push(`Category: ${e.category}`);
    if (e.type) lines.push(`Type: ${e.type}`);
    if (e.source) lines.push(`Source: ${e.source}`);
    if (e.scope) lines.push(`Scope: ${e.scope}`);
    if (e.target) lines.push(`Target: ${e.target}`);
    if (e.message) lines.push(`Message: ${e.message}`);

    if (e.meta && Object.keys(e.meta).length > 0) {
        lines.push("\nMeta:");
        for (const [key, value] of Object.entries(e.meta)) {
            lines.push(`  ${key}: ${value}`);
        }
    }

    return lines.join("\n");
}


function colorClass(level) {
    switch (level) {
        case "info": return "blue-500";
        case "warning": return "yellow-500";
        case "critical": return "red-500";
        default: return "gray-400";
    }
}

function iconForCategory(category) {
    switch (category) {
        case "alert":
            return `<svg class="inline w-4 h-4 text-yellow-500 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01M5.293 6.707a1 1 0 011.414 0L12 12l5.293-5.293a1 1 0 111.414 1.414L12 14.828l-6.707-6.707a1 1 0 010-1.414z" /></svg>`;
        case "system":
            return `<svg class="inline w-4 h-4 text-blue-500 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 17v-6a2 2 0 012-2h2a2 2 0 012 2v6m4 4H5a2 2 0 01-2-2V5a2 2 0 012-2h14a2 2 0 012 2v14a2 2 0 01-2 2z" /></svg>`;
        default:
            return `<svg class="inline w-4 h-4 text-gray-400 mr-1" fill="currentColor" viewBox="0 0 20 20"><path d="M9 12h2V8H9v4zm0 4h2v-2H9v2zm1-16C4.48 0 0 4.48 0 10s4.48 10 10 10 10-4.48 10-10S15.52 0 10 0z"/></svg>`;
    }
}

window.addEventListener("resize", () => {
    const expanded = document.querySelector(".expanded");
    if (expanded) {
        const wrapper = document.getElementById(`${expanded.dataset.detailRow}-wrapper`);
        if (wrapper) wrapper.style.maxHeight = wrapper.scrollHeight + "px";
    }
});

// Attach event handler
window.eventHandler = function (events) {
    if (!events || !Array.isArray(events)) return;

    for (const e of events) {
        if (
            !window.endpointID ||               // global/unfiltered
            e.scope === "global" ||             // global
            e.target === window.endpointID      // scoped
        ) {
            appendActivityRow(e);
        }
    }
};