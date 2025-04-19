///// START LOGS TAB
//////
(() => {
    const logsTableBody = document.getElementById("logs-table-body");
    const logsFilterLevel = document.getElementById("logs-filter-level");
    const logsFilterSearch = document.getElementById("logs-filter-search");
    const logsFilterStart = document.getElementById("logs-filter-start");
    const logsFilterEnd = document.getElementById("logs-filter-end");
    const logsRefreshButton = document.getElementById("logs-refresh-button");
    const logsPageIndicator = document.getElementById("logs-page-indicator");
    const logsPrevButton = document.getElementById("logs-prev");
    const logsNextButton = document.getElementById("logs-next");
    
    let allLogs = [];
    let logsCurrentPage = 1;
    const logsPerPage = 50;
    
    function fetchLogs(params = null) {
        if (!params) {
            const level = logsFilterLevel.value;
            const contains = logsFilterSearch.value;
            const start = logsFilterStart.value;
            const end = logsFilterEnd.value;
    
            params = new URLSearchParams({
                endpointID: window.endpointID,
                level,
                contains,
                start,
                end,
                limit: 1000,
            });
        }
    
        fetch("/api/v1/logs?" + params.toString())
            .then((res) => res.json())
            .then((logs) => {
                if (!Array.isArray(logs)) {
                    console.warn("Invalid logs response:", logs);
                    allLogs = [];
                } else {
                    allLogs = logs;
                }
                logsCurrentPage = 1;
                renderLogsPage(logsCurrentPage);
            })
            .catch((err) => {
                console.error("Failed to fetch logs:", err);
                allLogs = [];
                renderLogsPage(1);
            });
    }
    
    function renderLogsPage(page) {
        logsTableBody.innerHTML = "";
        const start = (page - 1) * logsPerPage;
        const pageLogs = allLogs.slice(start, start + logsPerPage);
        if (pageLogs.length === 0) {
            const row = document.createElement("tr");
            row.innerHTML = `<td colspan="5" class="px-4 py-4 text-center text-sm text-gray-500 dark:text-gray-400">No logs found for this filter</td>`;
            logsTableBody.appendChild(row);
        }
        for (const log of pageLogs) {
            const tr = document.createElement("tr");
    
            const levelClass = logLevelColorClass(log.level);
            const ts = new Date(log.timestamp).toLocaleString();
    
            tr.innerHTML = `
                <td class="px-4 py-2"><span class="inline-block text-xs font-medium px-2 py-0.5 rounded ${levelClass}">${log.level}</span></td>
                <td class="px-4 py-2">${log.source}</td>
                <td class="px-4 py-2">${log.category}</td>
                <td class="px-4 py-2">${ts}</td>
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
    
    function exportLogsCSV() {
        let csv = "Level,Source,Category,Timestamp,Message\n";
        allLogs.forEach(log => {
            const ts = new Date(log.timestamp).toLocaleString();
            const message = (log.message || "").replace(/"/g, '""');
            csv += `"${log.level}","${log.source}","${log.category}","${ts}","${message}"\n`;
        });
    
        const blob = new Blob([csv], { type: "text/csv" });
        const url = URL.createObjectURL(blob);
        const a = document.createElement("a");
        a.href = url;
        a.download = "logs.csv";
        a.click();
        URL.revokeObjectURL(url);
    }
    
    document.getElementById("logs-export-csv")?.addEventListener("click", exportLogsCSV);
    
    function sortLogsBy(field) {
        allLogs.sort((a, b) => (a[field] > b[field] ? 1 : -1));
        logsCurrentPage = 1;
        renderLogsPage(logsCurrentPage);
    }
    
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
    
    document.addEventListener("DOMContentLoaded", () => {
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
    
        logsFilterLevel?.addEventListener("change", fetchLogs);
        logsFilterSearch?.addEventListener("input", fetchLogs);
        logsFilterStart?.addEventListener("change", fetchLogs);
        logsFilterEnd?.addEventListener("change", fetchLogs);
    
        const logsTabPanel = document.getElementById("logs");
        if (logsTabPanel) {
            const observer = new MutationObserver(() => {
                const isVisible = !logsTabPanel.classList.contains("hidden");
                if (isVisible) {
                    console.log("📥 Logs tab became visible — fetching logs");
                    fetchLogs();
                }
            });
            observer.observe(logsTabPanel, { attributes: true, attributeFilter: ["class"] });
        }
    
        // 🔍 Advanced search logic
        document.getElementById("log-search-button")?.addEventListener("click", () => {
            const advKeyword = document.getElementById("adv-keyword")?.value.trim();
            const advLevel = document.getElementById("adv-level")?.value.trim();
            const advSource = document.getElementById("adv-source")?.value.trim();
            function padTime(val) {
                return val && val.length === 16 ? val + ":00" : val;
            }
            const advStart = padTime(document.getElementById("adv-start")?.value.trim());
            const advEnd = padTime(document.getElementById("adv-end")?.value.trim());

    
            const params = new URLSearchParams();
            params.set("endpointID", window.endpointID);
            params.set("limit", "1000");
    
            if (advKeyword) params.set("keyword", advKeyword);
            if (advLevel) params.set("level", advLevel);
            if (advSource) params.set("source", advSource);
            if (advStart) params.set("start", advStart);
            if (advEnd) params.set("end", advEnd);
    
            fetchLogs(params);
        });
    
        document.getElementById("log-reset-button")?.addEventListener("click", () => {
            document.getElementById("adv-keyword").value = "";
            document.getElementById("adv-level").value = "";
            document.getElementById("adv-source").value = "";
            document.getElementById("adv-start").value = "";
            document.getElementById("adv-end").value = "";
        });
    });
    /////// END LOGS TAB
    })();
    