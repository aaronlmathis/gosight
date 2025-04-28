import { gosightFetch } from "./api.js";

export async function loadEndpointHeader() {
    try {
        const endpointArray = await gosightFetch(`/api/v1/endpoints?endpointID=${encodeURIComponent(window.endpointID)}`).then(res => res.json());
        const endpoint = endpointArray[0];
        console.log("Loaded endpoint data:", endpoint);
        if (!endpoint) return;

        // Populate Hostname

        const hostnameEl = document.getElementById("hostname");
        if (!hostnameEl) {
            console.error("Hostname element not found!");
        } else {
            console.log("Setting hostname to:", endpoint.hostname);
            hostnameEl.textContent = endpoint.hostname || "Unknown Host";
        }

        // Populate Online/Offline Status
        const statusEls = document.querySelectorAll("span.font-medium.text-green-500, span.font-medium.text-red-500");
        const lastReportEl = document.querySelector("span.text-gray-500.dark\\:text-gray-400");

        if (endpoint.status === "Online") {
            if (statusEls[0]) {
                statusEls[0].textContent = "● Online";
                statusEls[0].classList.remove("text-red-500");
                statusEls[0].classList.add("text-green-500");
            }
        } else {
            if (statusEls[0]) {
                statusEls[0].textContent = "● Offline";
                statusEls[0].classList.remove("text-green-500");
                statusEls[0].classList.add("text-red-500");
            }
        }

        // Populate Last Report
        if (lastReportEl && endpoint.last_seen) {
            const lastSeenDate = new Date(endpoint.last_seen);
            const minutesAgo = Math.floor((Date.now() - lastSeenDate.getTime()) / 60000);
            lastReportEl.textContent = `Last report: ${minutesAgo}m ago`;
        }

        const badgeContainer = document.getElementById("tag-badges");

        // Now fetch tags
        loadEndpointTags(badgeContainer);

        // show Add Tag button if permission
        if (window.permissions?.includes("gosight:api:tags:patch")) {
            document.getElementById("add-tag-button")?.classList.remove("hidden");
        }

    } catch (err) {
        console.error("Failed to load endpoint header info:", err);
    }
}

export async function setupTagButton() {
    const addTagButton = document.getElementById("add-tag-button");
    const modal = document.getElementById("add-tag-modal");
    const closeModalBtn = document.getElementById("close-add-tag");
    const cancelModalBtn = document.getElementById("cancel-add-tag");
    const confirmModalBtn = document.getElementById("confirm-add-tag");

    if (!addTagButton || !modal) return;

    addTagButton.addEventListener("click", () => {
        modal.classList.remove("hidden");
    });

    closeModalBtn?.addEventListener("click", () => {
        modal.classList.add("hidden");
    });

    cancelModalBtn?.addEventListener("click", () => {
        modal.classList.add("hidden");
    });

    confirmModalBtn?.addEventListener("click", async () => {
        const key = document.getElementById("tag-key").value.trim();
        const value = document.getElementById("tag-value").value.trim();

        if (!key || !value) {
            alert("Please enter both key and value.");
            return;
        }

        try {
            await gosightFetch(`/api/v1/tags/${encodeURIComponent(window.endpointID)}`, {
                method: "PATCH",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ [key]: value })
            });

            const badgeContainer = document.getElementById("tag-badges");
            if (badgeContainer) {
                badgeContainer.innerHTML = "";
                await loadEndpointTags(badgeContainer);
            }

            modal.classList.add("hidden");
            document.getElementById("tag-key").value = "";
            document.getElementById("tag-value").value = "";

        } catch (err) {
            console.error("Failed to add tag:", err);
        }
    });
}

function openAddTagsModal() {
    const modal = document.getElementById("add-tags-modal");
    modal.classList.remove("hidden");
    setTimeout(() => {
        modal.classList.remove("opacity-0");
        const panel = modal.querySelector(".bg-white, .dark\\:bg-gray-800");
        panel?.classList.remove("scale-95");
        panel?.classList.add("scale-100");
    }, 10);
}

function closeAddTagsModal() {
    const modal = document.getElementById("add-tags-modal");
    modal.classList.add("opacity-0");
    const panel = modal.querySelector(".bg-white, .dark\\:bg-gray-800");
    panel?.classList.remove("scale-100");
    panel?.classList.add("scale-95");
    setTimeout(() => {
        modal.classList.add("hidden");
    }, 300);
}

async function fetchTagKeys() {
    try {
        const res = await fetch("/api/v1/tags/keys");
        return await res.json();
    } catch {
        return [];
    }
}

async function fetchTagValues(key) {
    try {
        const res = await fetch(`/api/v1/tags/values?key=${encodeURIComponent(key)}`);
        return await res.json();
    } catch {
        return [];
    }
}

function createTagRow(keysList = []) {
    const row = document.createElement("div");
    row.className = "flex space-x-2";

    const keyInput = document.createElement("input");
    keyInput.className = "flex-1 p-2 border border-gray-300 rounded-md text-sm bg-white dark:bg-gray-700 dark:border-gray-600 dark:text-white focus:ring-1 focus:ring-blue-500 focus:border-blue-500";
    keyInput.placeholder = "Key";

    const valueInput = document.createElement("input");
    valueInput.className = "flex-1 p-2 border border-gray-300 rounded-md text-sm bg-white dark:bg-gray-700 dark:border-gray-600 dark:text-white focus:ring-1 focus:ring-blue-500 focus:border-blue-500";
    valueInput.placeholder = "Value";

    const removeBtn = document.createElement("button");
    removeBtn.className = "text-red-500 hover:text-red-700 text-xs";
    removeBtn.textContent = "✖";
    removeBtn.onclick = () => row.remove();

    // Autosuggest for keys
    keyInput.addEventListener("input", async () => {
        const key = keyInput.value.trim();
        if (!key) return;
        const values = await fetchTagValues(key);
        if (values.length > 0) {
            valueInput.setAttribute("list", `values-${key}`);
            if (!document.getElementById(`values-${key}`)) {
                const datalist = document.createElement("datalist");
                datalist.id = `values-${key}`;
                values.forEach(v => {
                    const opt = document.createElement("option");
                    opt.value = v;
                    datalist.appendChild(opt);
                });
                document.body.appendChild(datalist);
            }
        }
    });

    if (keysList.length > 0) {
        const datalist = document.createElement("datalist");
        datalist.id = "available-keys";
        keysList.forEach(k => {
            const opt = document.createElement("option");
            opt.value = k;
            datalist.appendChild(opt);
        });
        document.body.appendChild(datalist);
        keyInput.setAttribute("list", "available-keys");
    }

    row.appendChild(keyInput);
    row.appendChild(valueInput);
    row.appendChild(removeBtn);
    return row;
}

export async function setupTagsModal() {
    const closeButton = document.getElementById("close-add-tags");
    const cancelButton = document.getElementById("cancel-tags");
    const addTagRowButton = document.getElementById("add-tag-row");
    const tagsForm = document.getElementById("tags-form");
    const tagsRowsContainer = document.getElementById("tags-rows");

    if (!closeButton || !cancelButton || !addTagRowButton || !tagsForm || !tagsRowsContainer) {
        console.error("Missing modal elements!");
        return;
    }

    let availableKeys = [];

    closeButton.addEventListener("click", closeAddTagsModal);
    cancelButton.addEventListener("click", closeAddTagsModal);

    addTagRowButton.addEventListener("click", async () => {
        if (availableKeys.length === 0) {
            availableKeys = await fetchTagKeys();
        }
        tagsRowsContainer.appendChild(createTagRow(availableKeys));
    });

    tagsForm.addEventListener("submit", async (e) => {
        e.preventDefault();

        const payload = {};
        tagsRowsContainer.querySelectorAll("div.flex").forEach(row => {
            const inputs = row.querySelectorAll("input");
            const key = inputs[0]?.value.trim();
            const value = inputs[1]?.value.trim();
            if (key && value) {
                payload[key] = value;
            }
        });

        if (Object.keys(payload).length === 0) {
            alert("Please add at least one valid tag.");
            return;
        }

        try {
            await gosightFetch(`/api/v1/tags/${encodeURIComponent(window.endpointID)}`, {
                method: "PATCH",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify(payload)
            });

            const badgeContainer = document.getElementById("tag-badges");
            if (badgeContainer) {
                badgeContainer.innerHTML = "";
                await loadEndpointTags();
            }
            closeAddTagsModal();
        } catch (err) {
            console.error("Failed to save tags:", err);
        }
    });
}


async function loadEndpointTags(container) {
    try {
        const container = document.getElementById("tag-badges");
        if (!container) {
            console.error("No #tag-badges found!");
            return;
        }

        const tags = await gosightFetch(`/api/v1/tags/${encodeURIComponent(window.endpointID)}`).then(res => res.json());

        container.innerHTML = "";

        if (!tags || Object.keys(tags).length === 0) return;

        if (tags && Object.keys(tags).length > 0) {
            const label = document.createElement("span");
            label.className = "text-xs font-semibold text-gray-500 dark:text-gray-400 mr-2";
            label.textContent = "Tags:";
            container.appendChild(label);
        }

        for (const [key, value] of Object.entries(tags)) {
            const badge = document.createElement("span");
            badge.className = "bg-blue-100 text-blue-800 text-xs font-semibold px-2.5 py-0.5 rounded dark:bg-blue-900 dark:text-blue-300";
            badge.textContent = `${key}: ${value}`;
            container.appendChild(badge);
        }

        if (window.permissions?.includes("gosight:api:tags:patch")) {
            const addBadge = document.createElement("span");
            addBadge.id = "add-tag-button";
            addBadge.className = "bg-gray-100 text-gray-800 text-xs font-semibold px-2.5 py-0.5 rounded cursor-pointer hover:bg-blue-200 dark:bg-blue-900 dark:text-blue-300 dark:hover:bg-blue-800";
            addBadge.textContent = "+ Add Tag";
            addBadge.addEventListener("click", async () => {
                const tagsRowsContainer = document.getElementById("tags-rows");
                if (tagsRowsContainer) {
                    tagsRowsContainer.innerHTML = "";
                    const availableKeys = await fetchTagKeys();
                    tagsRowsContainer.appendChild(createTagRow(availableKeys));
                }
                openAddTagsModal();
            });
            container.appendChild(addBadge);
        }
    } catch (err) {
        console.error("Failed to load tags:", err);
    }
}
