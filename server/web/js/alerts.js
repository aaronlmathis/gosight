import { gosightFetch } from "./api.js";

document.addEventListener("DOMContentLoaded", async () => {
    const alertsTableBody = document.getElementById("alerts-table-body");
    const searchInput = document.getElementById("alert-search");
    const selectAllCheckbox = document.getElementById("select-all");
    const bulkActionsBtn = document.getElementById("bulk-actions-btn");
    const bulkActionsMenu = document.getElementById("bulk-actions-menu");

    let rules = [];
    let summaries = [];
    let tableData = [];
    let sortField = "name";
    let sortAsc = true;
    let bulkActionsMenuOpen = false;

    async function loadAlerts() {
        rules = await gosightFetch('/api/v1/alerts/rules').then(res => res.json());
        summaries = await gosightFetch('/api/v1/alerts/summary').then(res => res.json());
        buildTableData();
        renderAlerts(tableData);
    }

    function buildTableData() {
        tableData = rules.map(rule => {
            const summary = summaries.find(s => s.rule_id === rule.id);

            let state = "Insufficient Data";
            let lastStateChange = "-";

            if (summary) {
                if (summary.state === "firing") {
                    state = "Alarm";
                } else if (summary.state === "resolved") {
                    state = "OK";
                }
                lastStateChange = summary.last_change;
            }

            return {
                id: rule.id,
                name: rule.name,
                state: state,
                last_state_change: lastStateChange,
                conditions_summary: `${rule.expression} ${formatMatchCriteria(rule.match)}`,
                actions: rule.actions || [],
            };
        });
    }

    function formatMatchCriteria(match) {
        if (!match || Object.keys(match).length === 0) {
            return "";
        }
        return "(" + Object.entries(match)
            .map(([k, v]) => `${k}=${v}`)
            .join(", ") + ")";
    }

    function renderAlerts(data) {
        alertsTableBody.innerHTML = '';
        data.forEach(alert => {
            let stateBadgeClass = "bg-gray-300 text-gray-800"; // Default
            if (alert.state === "Alarm") {
                stateBadgeClass = "bg-red-100 text-red-800";
            } else if (alert.state === "OK") {
                stateBadgeClass = "bg-green-100 text-green-800";
            }

            alertsTableBody.innerHTML += `
<tr class="hover:bg-gray-50 dark:hover:bg-gray-800 group border-t border-gray-200 dark:border-gray-700 cursor-pointer" data-id="${alert.id}">
    <td class="px-4 py-3">
        <input type="checkbox" class="row-checkbox accent-blue-500 rounded" data-id="${alert.id}">
    </td>
    <td class="px-4 py-3 font-medium text-gray-900 dark:text-white">
        ${alert.name}
    </td>
    <td class="px-4 py-3">
        <span class="px-2 py-1 rounded ${stateBadgeClass}">${alert.state}</span>
    </td>
    <td class="px-4 py-3 text-gray-700 dark:text-gray-300 text-xs">
        ${alert.last_state_change !== "-" ? new Date(alert.last_state_change).toLocaleString() : "-"}
    </td>
    <td class="px-4 py-3 text-gray-700 dark:text-gray-300 text-xs">
        ${alert.conditions_summary}
    </td>
    <td class="px-4 py-3 flex items-center gap-2 text-gray-700 dark:text-gray-300 text-xs">
        ${alert.actions.length > 0 ? alert.actions.join(", ") : "-"}
        <button class="text-blue-600 hover:text-blue-800 dark:text-blue-400 dark:hover:text-blue-300 ml-auto" title="Edit Alert" onclick="editAlert('${alert.id}')">
            <i class="fas fa-edit"></i>
        </button>
    </td>
</tr>`;
        });
        attachRowHandlers();
        updateBulkActionsState();
    }

    function editAlert(alertId) {
        console.log("Editing alert:", alertId);
        window.location.href = `/alerts/edit/${alertId}`;
    }

    searchInput.addEventListener("input", () => {
        const term = searchInput.value.toLowerCase();
        const filtered = tableData.filter(a =>
            a.name.toLowerCase().includes(term) ||
            a.conditions_summary.toLowerCase().includes(term));
        renderAlerts(filtered);
    });

    selectAllCheckbox.addEventListener("change", () => {
        const checkboxes = document.querySelectorAll(".row-checkbox");
        checkboxes.forEach(cb => cb.checked = selectAllCheckbox.checked);
        updateBulkActionsState();
    });

    document.addEventListener("change", (e) => {
        if (e.target.classList.contains("row-checkbox")) {
            updateBulkActionsState();
        }
    });

    function updateBulkActionsState() {
        const checked = document.querySelectorAll(".row-checkbox:checked").length;
        if (checked > 0) {
            bulkActionsBtn.disabled = false;
            bulkActionsBtn.classList.remove("bg-gray-300", "text-gray-600", "cursor-not-allowed");
            bulkActionsBtn.classList.add("bg-blue-600", "text-white", "hover:bg-blue-700", "cursor-pointer");
        } else {
            bulkActionsBtn.disabled = true;
            bulkActionsBtn.classList.remove("bg-blue-600", "text-white", "hover:bg-blue-700", "cursor-pointer");
            bulkActionsBtn.classList.add("bg-gray-300", "text-gray-600", "cursor-not-allowed");
        }
    }

    bulkActionsBtn.addEventListener("click", (e) => {
        if (bulkActionsBtn.disabled) return;
        e.stopPropagation();
        bulkActionsMenu.classList.toggle("hidden");
        bulkActionsMenuOpen = !bulkActionsMenuOpen;
    });

    document.addEventListener("click", () => {
        if (bulkActionsMenuOpen) {
            bulkActionsMenu.classList.add("hidden");
            bulkActionsMenuOpen = false;
        }
    });

    document.getElementById("disable-selected").addEventListener("click", () => {
        const selected = getSelectedIds();
        console.log("Disabling selected:", selected);
        bulkActionsMenu.classList.add("hidden");
        // TODO: Send API request
    });

    document.getElementById("delete-selected").addEventListener("click", () => {
        const selected = getSelectedIds();
        console.log("Deleting selected:", selected);
        bulkActionsMenu.classList.add("hidden");
        // TODO: Send API request
    });

    function getSelectedIds() {
        return Array.from(document.querySelectorAll(".row-checkbox:checked"))
            .map(cb => cb.dataset.id);
    }

    function attachRowHandlers() {
        document.querySelectorAll("#alerts-table-body tr").forEach(tr => {
            tr.addEventListener("click", (e) => {
                if (e.target.tagName === "INPUT" || e.target.tagName === "BUTTON" || e.target.closest("button")) {
                    return;
                }
                const checkbox = tr.querySelector(".row-checkbox");
                if (checkbox) {
                    checkbox.checked = !checkbox.checked;
                    updateBulkActionsState();
                }
            });
        });
    }

    // Sortable Columns
    document.querySelectorAll("th[data-sort]").forEach(th => {
        th.addEventListener("click", () => {
            const field = th.dataset.sort;
            if (sortField === field) {
                sortAsc = !sortAsc;
            } else {
                sortField = field;
                sortAsc = true;
            }
            sortTable();
        });
    });

    function sortTable() {
        tableData.sort((a, b) => {
            let valA = a[sortField] || "";
            let valB = b[sortField] || "";
            if (typeof valA === "string") valA = valA.toLowerCase();
            if (typeof valB === "string") valB = valB.toLowerCase();
            if (valA < valB) return sortAsc ? -1 : 1;
            if (valA > valB) return sortAsc ? 1 : -1;
            return 0;
        });
        renderAlerts(tableData);
    }

    loadAlerts();
});
