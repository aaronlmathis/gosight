import { registerTabInitializer } from "./tabs.js";
import { gosightFetch } from "./api.js";
console.log("logs.js loaded");
// logs.js
export function initLogsTab() {
    try {
        renderLogsTableShell();

        const logsTableBody = document.getElementById("logs-table-body");
        const logsFilterLevel = document.getElementById("logs-filter-level");
        const logsFilterSearch = document.getElementById("logs-filter-search");
        const logsFilterStart = document.getElementById("logs-filter-start");
        const logsFilterEnd = document.getElementById("logs-filter-end");
        const logsRefreshButton = document.getElementById("logs-refresh-button");
        const logsPageIndicator = document.getElementById("logs-page-indicator");
        const logsPrevButton = document.getElementById("logs-prev");
        const logsNextButton = document.getElementById("logs-next");
        const exportBtn = document.getElementById("logs-export-csv");

        let allLogs = [];
        let logsCurrentPage = 1;
        const logsPerPage = 50;

        function logLevelColorClass(level) {
            switch ((level || "").toLowerCase()) {
                case "error": return "bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-300";
                case "warn": return "bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-300";
                case "info": return "bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-300";
                case "debug": return "bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-300";
                case "notice": return "bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-300";
                default: return "bg-gray-200 text-gray-800 dark:bg-gray-800 dark:text-gray-300";
            }
        }

        function renderLogsPage(page) {
            if (!logsTableBody) {
                console.error("❌ logsTableBody not found");
                return;
            }
            logsTableBody.innerHTML = "";
            const start = (page - 1) * logsPerPage;
            const pageLogs = allLogs.slice(start, start + logsPerPage);

            if (pageLogs.length === 0) {
                logsTableBody.innerHTML = `
          <tr><td colspan="5" class="px-4 py-4 text-center text-sm text-gray-500 dark:text-gray-400">
            No logs found for this filter
          </td></tr>`;
                return;
            }

            for (const log of pageLogs) {
                const levelClass = logLevelColorClass(log.level);
                const ts = new Date(log.timestamp).toLocaleString();
                const tr = document.createElement("tr");
                tr.innerHTML = `
                          <td class="px-4 py-2">${ts}</td>
          <td class="px-4 py-2">
            <span class="inline-block text-xs font-medium px-2 py-0.5 rounded ${levelClass}">${log.level}</span>
          </td>
          <td class="px-4 py-2">${log.source}</td>
          <td class="px-4 py-2">${log.category}</td>

          <td class="px-4 py-2">${log.message}</td>
        `;
                logsTableBody.appendChild(tr);
            }

            updateLogsPagination();
        }

        function updateLogsPagination() {
            const totalPages = Math.max(1, Math.ceil(allLogs.length / logsPerPage));
            logsPageIndicator.textContent = `Page ${logsCurrentPage} of ${totalPages}`;
            logsPrevButton.disabled = logsCurrentPage === 1;
            logsNextButton.disabled = logsCurrentPage === totalPages;
        }

        function fetchLogs(params = null) {
            if (!params) {
                const level = logsFilterLevel.value;
                const contains = logsFilterSearch.value;
                const start = logsFilterStart.value;
                const end = logsFilterEnd.value;
                params = new URLSearchParams({ endpointID: window.endpointID, level, contains, start, end, limit: 1000 });
            }

            gosightFetch("/api/v1/logs?" + params.toString())
                .then(res => res.json())
                .then(logs => {
                    allLogs = Array.isArray(logs) ? logs : [];
                    logsCurrentPage = 1;
                    renderLogsPage(logsCurrentPage);
                })
                .catch(err => {
                    console.error("Failed to fetch logs:", err);
                    allLogs = [];
                    renderLogsPage(1);
                });
        }

        function exportLogsCSV() {
            let csv = "Level,Source,Category,Timestamp,Message\n";
            allLogs.forEach(log => {
                const ts = new Date(log.timestamp).toLocaleString();
                const msg = (log.message || "").replace(/"/g, '""');
                csv += `"${log.level}","${log.source}","${log.category}","${ts}","${msg}"\n`;
            });
            const blob = new Blob([csv], { type: "text/csv" });
            const url = URL.createObjectURL(blob);
            const a = document.createElement("a");
            a.href = url;
            a.download = "logs.csv";
            a.click();
            URL.revokeObjectURL(url);
        }

        function sortLogsBy(field) {
            allLogs.sort((a, b) => (a[field] > b[field] ? 1 : -1));
            logsCurrentPage = 1;
            renderLogsPage(logsCurrentPage);
        }

        function renderLogsTableShell(containerId = "logs-table-container") {
            const container = document.getElementById(containerId);
            if (!container) {
                console.error(`❌ Container #${containerId} not found`);
                return;
            }

            container.innerHTML = `
            <div class="overflow-x-auto border border-gray-200 dark:border-gray-700 rounded-lg">
              <table class="min-w-full divide-y divide-gray-200 dark:divide-gray-700 text-sm text-left">
                <thead class="bg-gray-50 dark:bg-gray-800 text-gray-600 dark:text-gray-300 font-semibold text-xs uppercase">
                  <tr>
                    <th scope="col" class="px-4 py-2">Time</th>
                    <th scope="col" class="px-4 py-2">Level</th>
                    <th scope="col" class="px-4 py-2">Source</th>
                    <th scope="col" class="px-4 py-2">Category</th>
                    <th scope="col" class="px-4 py-2">Message</th>
                  </tr>
                </thead>
                <tbody id="logs-table-body" class="bg-white dark:bg-gray-900 divide-y divide-gray-100 dark:divide-gray-800 text-gray-900 dark:text-gray-200">
                  <!-- Log rows will be inserted here -->
                </tbody>
              </table>
            </div>
            `;

            container.querySelectorAll("[data-sort]").forEach(th => {
                th.addEventListener("click", () => {
                    const field = th.getAttribute("data-sort");
                    if (field) sortLogsBy(field);
                });
            });
        }

        function bindEventListeners() {
            logsFilterLevel?.addEventListener("change", fetchLogs);
            logsFilterSearch?.addEventListener("input", fetchLogs);
            logsFilterStart?.addEventListener("change", fetchLogs);
            logsFilterEnd?.addEventListener("change", fetchLogs);
            logsRefreshButton?.addEventListener("click", fetchLogs);
            logsPrevButton?.addEventListener("click", () => {
                if (logsCurrentPage > 1) {
                    logsCurrentPage--;
                    renderLogsPage(logsCurrentPage);
                }
            });
            logsNextButton?.addEventListener("click", () => {
                const totalPages = Math.ceil(allLogs.length / logsPerPage);
                if (logsCurrentPage < totalPages) {
                    logsCurrentPage++;
                    renderLogsPage(logsCurrentPage);
                }
            });
            exportBtn?.addEventListener("click", exportLogsCSV);

            document.getElementById("log-search-button")?.addEventListener("click", () => {
                const pad = val => val && val.length === 16 ? val + ":00" : val;
                const params = new URLSearchParams({
                    endpointID: window.endpointID,
                    keyword: document.getElementById("adv-keyword")?.value.trim() || "",
                    level: document.getElementById("adv-level")?.value.trim() || "",
                    source: document.getElementById("adv-source")?.value.trim() || "",
                    start: pad(document.getElementById("adv-start")?.value.trim() || ""),
                    end: pad(document.getElementById("adv-end")?.value.trim() || ""),
                    limit: 1000
                });
                fetchLogs(params);
            });

            document.getElementById("log-reset-button")?.addEventListener("click", () => {
                ["adv-keyword", "adv-level", "adv-source", "adv-start", "adv-end"].forEach(id => {
                    const el = document.getElementById(id);
                    if (el) el.value = "";
                });
            });

            const tab = document.getElementById("logs");
            if (tab) {
                const observer = new MutationObserver(() => {
                    const isVisible = !tab.classList.contains("hidden");
                    if (isVisible) {
                        console.log("Logs tab became visible — fetching logs");
                        fetchLogs();
                    }
                });
                observer.observe(tab, { attributes: true, attributeFilter: ["class"] });
            }
        }


        bindEventListeners();
        fetchLogs(); // preload logs immediately when the tab loads

    } catch (err) {
        console.error("logs.js init failed:", err);
    }
}

// Register with global tab system
registerTabInitializer("logs", initLogsTab);