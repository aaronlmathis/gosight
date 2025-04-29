import { gosightFetch } from "./api.js";

let currentStep = 0;
const stepTitles = ["Type", "Scope", "Condition", "Actions", "Review"];

const alertData = {
    type: "metric",
    match: {
        endpoint_ids: [],
        labels: {}
    },
    scope: {},
    expression: {},
    level: "warning",
    actions: [],
    options: {
        cooldown: "30s",
        eval_interval: "10s",
        repeat_interval: "1m",
        notify_on_resolve: true
    }
};

export function initStepper() {
    renderStep();
    document.getElementById("next-btn").addEventListener("click", () => handleFormSubmission("next"));
    document.getElementById("prev-btn").addEventListener("click", () => handleFormSubmission("prev"));
}
export async function handleFormSubmission(direction) {
    let newStep = currentStep;

    if (direction === "next") {
        newStep = currentStep + 1;
    } else if (direction === "prev") {
        newStep = currentStep - 1;
    }

    // Only validate if trying to go FROM Step 1 (Scope) TO Step 2 (Condition)
    if (currentStep === 1 && newStep === 2 && alertData.type === "metric") {
        const ns = alertData.scope.namespace;
        const subns = alertData.scope.subnamespace;
        const metric = alertData.scope.metric;

        if (!ns || !subns || !metric) {
            alert("Please complete Namespace, SubNamespace, and Metric before proceeding.");
            return; // Stop, don't move forward
        }
    }

    // Now apply the move
    if (newStep >= 0 && newStep < stepTitles.length) {
        currentStep = newStep;
    }

    await renderStep();
    updateStepperVisual();
}


function updateStepperVisual() {
    const timeline = document.getElementById("timeline-stepper");
    timeline.innerHTML = "";

    const stepLabels = ["Type", "Scope", "Condition", "Actions", "Review"];

    for (let idx = 0; idx < stepLabels.length; idx++) {
        const isCompleted = idx < currentStep;
        const isCurrent = idx === currentStep;

        timeline.innerHTML += `
        <li class="mb-10 ms-6">
          <span class="absolute flex items-center justify-center w-8 h-8
            ${isCompleted ? "bg-green-200 dark:bg-green-900" : isCurrent ? "bg-blue-200 dark:bg-blue-800" : "bg-gray-100 dark:bg-gray-700"}
            rounded-full -start-4 ring-4 ring-white dark:ring-gray-900">
            ${isCompleted ? `
              <svg class="w-3.5 h-3.5 text-green-600 dark:text-green-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="3" d="M5 13l4 4L19 7" />
              </svg>
            ` : `
              <span class="text-sm font-bold text-gray-600 dark:text-gray-300">${idx + 1}</span>
            `}
          </span>
  
          <div class="flex flex-col ms-10">
            <h3 class="font-semibold leading-tight ${isCurrent ? "text-blue-600" : "text-gray-900 dark:text-white"}">${stepLabels[idx]}</h3>
            ${isCurrent ? `<p class="text-sm text-gray-500 dark:text-gray-400">Current step</p>` : ""}
          </div>
        </li>
      `;
    }

    const prevBtn = document.getElementById("prev-btn");
    const nextBtn = document.getElementById("next-btn");

    prevBtn.classList.toggle("hidden", currentStep === 0);
    nextBtn.textContent = currentStep === stepLabels.length - 1 ? "Submit" : "Next";

    if (currentStep === 1 && alertData.type === "metric") {
        const ns = alertData.scope.namespace;
        const subns = alertData.scope.subnamespace;
        const metric = alertData.scope.metric;
        if (!ns || !subns || !metric) {
            nextBtn.disabled = true;
            nextBtn.classList.add("opacity-50", "cursor-not-allowed");
        } else {
            nextBtn.disabled = false;
            nextBtn.classList.remove("opacity-50", "cursor-not-allowed");
        }
    } else {
        nextBtn.disabled = false;
        nextBtn.classList.remove("opacity-50", "cursor-not-allowed");
    }
}



async function renderStep() {
    const container = document.getElementById("step-content");
    container.innerHTML = "";

    switch (currentStep) {
        case 0: renderTypeStep(container); break;
        case 1: await renderScopeStep(container); break;
        case 2: renderConditionStep(container); break;
        case 3: renderActionsStep(container); break;
        case 4: renderReviewStep(container); break;
    }
}
function renderTypeStep(container) {
    container.innerHTML = `
      <h2 class="text-lg font-semibold mb-4">Alert Type, Name, and Description</h2>
  
      <div class="space-y-4">
        <label class="font-semibold">Alert Type</label>
        <div class="flex flex-col space-y-2">
          <label><input type="radio" name="alert-type" value="metric" checked> Metric Alert</label>
          <label><input type="radio" name="alert-type" value="log"> Log Alert</label>
          <label><input type="radio" name="alert-type" value="event"> Event Alert</label>
        </div>
  
        <label class="font-semibold mt-4">Alert Name</label>
        <input id="alert-name-input" type="text" class="border rounded p-2 w-full" placeholder="e.g. High CPU Usage">
  
        <label class="font-semibold mt-4">Alert Description</label>
        <textarea id="alert-description-input" class="border rounded p-2 w-full" placeholder="Short description of this alert..."></textarea>
      </div>
    `;

    // ✨ Wire up Type selection
    container.querySelectorAll('input[name="alert-type"]').forEach(radio => {
        radio.addEventListener("change", (e) => {
            alertData.type = e.target.value;
        });
    });

    // ✨ Wire up Name and Description inputs
    document.getElementById("alert-name-input").addEventListener("input", (e) => {
        alertData.name = e.target.value.trim();
    });

    document.getElementById("alert-description-input").addEventListener("input", (e) => {
        alertData.description = e.target.value.trim();
    });
}
async function renderScopeStep(container) {
    container.innerHTML = `<h2 class="text-lg font-semibold mb-4">Define Scope</h2><div id="scope-fields" class="space-y-4"></div>`;

    if (alertData.type === "metric") {
        await renderMetricScopeFields();
    } else {
        renderLogEventScopeFields();
    }
}

async function renderMetricScopeFields() {
    const fields = document.getElementById("scope-fields");

    fields.innerHTML = `
      <label>Namespace</label>
      <select id="namespace-select" class="border rounded p-2 w-full"></select>
  
      <label>SubNamespace</label>
      <select id="subnamespace-select" class="border rounded p-2 w-full" disabled></select>
  
      <label>Metric</label>
      <select id="metric-select" class="border rounded p-2 w-full" disabled></select>
  
      <div id="filters-area" class="mt-6">
        <h3 class="font-semibold mb-2">Filters (Optional)</h3>
        <div id="filters-container" class="space-y-2"></div>
      </div>
    `;
    setupScopeListeners();
    await loadNamespaces();

}

function renderLogEventScopeFields() {
    const fields = document.getElementById("scope-fields");

    fields.innerHTML = `
      <label>Category</label><input id="category-input" class="border rounded p-2 w-full">
      <label>Source</label><input id="source-input" class="border rounded p-2 w-full">
      <label>Scope</label><input id="scope-input" class="border rounded p-2 w-full">
    `;

    document.getElementById("category-input").addEventListener("input", (e) => {
        alertData.match.category = e.target.value.trim();
    });
    document.getElementById("source-input").addEventListener("input", (e) => {
        alertData.match.source = e.target.value.trim();
    });
    document.getElementById("scope-input").addEventListener("input", (e) => {
        alertData.match.scope = e.target.value.trim();
    });
}

async function loadNamespaces() {
    const nsSelect = document.getElementById("namespace-select");
    nsSelect.innerHTML = `<option>Loading...</option>`;

    try {
        const res = await gosightFetch("/api/v1");
        const namespaces = await res.json();
        nsSelect.innerHTML = `<option value="">Select Namespace</option>` +
            namespaces.map(ns => `<option>${ns}</option>`).join("");
    } catch (err) {
        console.error("Failed to load namespaces:", err);
        nsSelect.innerHTML = `<option value="">(Error loading)</option>`;
    }
}

function setupScopeListeners() {
    document.getElementById("namespace-select").addEventListener("change", async (e) => {
        const ns = e.target.value;
        alertData.scope.namespace = ns;
        const subnsSelect = document.getElementById("subnamespace-select");
        subnsSelect.disabled = false;

        try {
            const res = await gosightFetch(`/api/v1/${ns}`);
            const subs = await res.json();
            subnsSelect.innerHTML = `<option value="">Select SubNamespace</option>` +
                subs.map(sb => `<option>${sb}</option>`).join("");
        } catch (err) {
            console.error("Failed to load subnamespaces:", err);
            subnsSelect.innerHTML = `<option value="">(Error loading)</option>`;
        }
    });

    document.getElementById("subnamespace-select").addEventListener("change", async (e) => {
        const ns = document.getElementById("namespace-select").value;
        const subns = e.target.value;
        alertData.scope.subnamespace = subns;
        const metricSelect = document.getElementById("metric-select");
        metricSelect.disabled = false;

        try {
            const res = await gosightFetch(`/api/v1/${ns}/${subns}`);
            const metrics = await res.json();
            metricSelect.innerHTML = `<option value="">Select Metric</option>` +
                metrics.map(m => `<option>${m}</option>`).join("");
        } catch (err) {
            console.error("Failed to load metrics:", err);
            metricSelect.innerHTML = `<option value="">(Error loading)</option>`;
        }
        updateStepperVisual();
    });

    document.getElementById("metric-select").addEventListener("change", async (e) => {
        const selectedMetric = e.target.value;
        alertData.scope.metric = selectedMetric;

        const ns = alertData.scope.namespace;
        const subns = alertData.scope.subnamespace;

        // 🛠 Strip the namespace.subnamespace. prefix off metric before querying dimensions
        let shortMetric = selectedMetric;
        const expectedPrefix = `${ns}.${subns}.`;
        if (selectedMetric.startsWith(expectedPrefix)) {
            shortMetric = selectedMetric.slice(expectedPrefix.length);
        }

        await loadDimensions(ns, subns, shortMetric);
        updateStepperVisual();
    });
}

function renderActionsStep(container) {
    container.innerHTML = `
      <h2 class="text-lg font-semibold mb-4">Set Actions and Options</h2>
      <div class="space-y-4">
        <label>Action IDs (comma separated)</label>
        <input id="actions-input" class="border rounded p-2 w-full" placeholder="notify-email, notify-webhook">
  
        <label>Cooldown</label>
        <input id="cooldown-input" class="border rounded p-2 w-full" value="30s">
  
        <label>Eval Interval</label>
        <input id="eval-interval-input" class="border rounded p-2 w-full" value="10s">
  
        <label>Repeat Interval</label>
        <input id="repeat-interval-input" class="border rounded p-2 w-full" value="1m">
  
        <div class="flex items-center space-x-2 mt-2">
          <input id="notify-resolve-checkbox" type="checkbox" checked>
          <label for="notify-resolve-checkbox">Notify on Resolve</label>
        </div>
      </div>
    `;

    // ✨ Wire up actions input
    document.getElementById("actions-input").addEventListener("input", (e) => {
        alertData.actions = e.target.value.split(",").map(a => a.trim()).filter(a => a.length > 0);
    });

    // ✨ Wire up options inputs
    document.getElementById("cooldown-input").addEventListener("input", (e) => {
        alertData.options.cooldown = e.target.value.trim();
    });

    document.getElementById("eval-interval-input").addEventListener("input", (e) => {
        alertData.options.eval_interval = e.target.value.trim();
    });

    document.getElementById("repeat-interval-input").addEventListener("input", (e) => {
        alertData.options.repeat_interval = e.target.value.trim();
    });

    document.getElementById("notify-resolve-checkbox").addEventListener("change", (e) => {
        alertData.options.notify_on_resolve = e.target.checked;
    });
}

function buildFinalPayload() {
    return {
        id: `alert-${Date.now()}`,
        name: alertData.name || `Alert ${Date.now()}`,
        description: alertData.description || "",
        enabled: true,
        type: alertData.type,
        match: alertData.match,
        scope: alertData.scope,
        expression: buildExpressionString(),
        level: alertData.level,
        actions: alertData.actions,
        options: alertData.options
    };
}


function renderReviewStep(container) {
    container.innerHTML = `
      <h2 class="text-lg font-semibold mb-4">Review & Submit</h2>
  
      <div class="bg-gray-100 dark:bg-gray-800 p-4 rounded overflow-x-auto text-sm">
        <pre id="alert-preview" class="whitespace-pre-wrap text-xs"></pre>
      </div>
  
      <div class="flex justify-end mt-6">
        <button id="submit-alert-btn" class="bg-green-600 hover:bg-green-700 text-white font-bold py-2 px-6 rounded">
          Submit Alert
        </button>
      </div>
    `;

    const previewEl = document.getElementById("alert-preview");
    previewEl.textContent = JSON.stringify(buildFinalPayload(), null, 2);

    document.getElementById("submit-alert-btn").addEventListener("click", submitAlert);
}

function buildExpressionString() {
    if (!alertData.expression) return "";

    const { operator, value, datatype } = alertData.expression;

    if (alertData.type === "metric") {
        // Example: "> 80" or "< 30" etc
        return `${operator} ${value}`;
    } else {
        // log or event
        return `${operator}:${value}`;
    }
}

async function submitAlert() {
    const payload = buildFinalPayload();

    try {
        const res = await gosightFetch('/api/v1/alerts', {
            method: "POST",
            headers: {
                "Content-Type": "application/json"
            },
            body: JSON.stringify(payload)
        });

        if (res.ok) {
            alert("✅ Alert created!");
            window.location.href = "/alerts";
        } else {
            const errorText = await res.text();
            console.error("Submit failed:", errorText);
            alert("❌ Failed to create alert.");
        }
    } catch (err) {
        console.error("Submit error:", err);
        alert("❌ Network error submitting alert.");
    }
}

async function loadDimensions(namespace, subnamespace, metric) {
    const container = document.getElementById("filters-container");
    container.innerHTML = `Loading filters...`;

    try {

        const res = await gosightFetch(`/api/v1/${namespace}/${subnamespace}/${metric}/dimensions`);
        const dims = await res.json();

        container.innerHTML = dims.map(dim => `
        <div class="flex items-center space-x-2">
          <label class="text-sm w-40">${dim}</label>
          <input data-dim="${dim}" class="border p-1 rounded text-sm flex-1" placeholder="Value for ${dim}">
        </div>
      `).join("");

        document.querySelectorAll("#filters-container input[data-dim]").forEach(input => {
            input.addEventListener("input", (e) => {
                const key = e.target.dataset.dim;
                const value = e.target.value.trim();
                if (value) {
                    alertData.match.labels[key] = value;
                } else {
                    delete alertData.match.labels[key];
                }
            });
        });
    } catch (err) {
        console.error("Failed to load dimensions:", err);
        container.innerHTML = `<div class="text-red-500 text-sm">(Error loading filters)</div>`;
    }
}


function renderConditionStep(container) {
    container.innerHTML = `
      <h2 class="text-lg font-semibold mb-4">Define Condition</h2>
      <div class="space-y-4">
        <label>Operator</label>
        <select id="operator-select" class="border rounded p-2 w-full">
          ${alertData.type === "metric" ? `
            <option value=">">Greater Than</option>
            <option value="<">Less Than</option>
            <option value="=">Equal To</option>
          ` : `
            <option value="contains">Contains</option>
            <option value="regex">Regex</option>
          `}
        </select>
  
        <label>Value</label>
        <input id="expr-value" class="border rounded p-2 w-full" placeholder="Enter threshold or pattern">
  
        ${alertData.type === "metric" ? `
          <label>Data Type</label>
          <select id="datatype-select" class="border rounded p-2 w-full">
            <option value="percent">Percent</option>
            <option value="numeric">Numeric</option>
            <option value="status">Status</option>
          </select>
        ` : ""}
      </div>
    `;

    // ✨ Now wire up properly
    const operatorSelect = document.getElementById("operator-select");
    const exprValueInput = document.getElementById("expr-value");
    const datatypeSelect = document.getElementById("datatype-select");

    operatorSelect.addEventListener("change", (e) => {
        alertData.expression.operator = e.target.value;
    });

    exprValueInput.addEventListener("input", (e) => {
        alertData.expression.value = e.target.value;
    });

    if (alertData.type === "metric") {
        datatypeSelect.addEventListener("change", (e) => {
            alertData.expression.datatype = e.target.value;
        });
    }

    alertData.expression = {
        operator: operatorSelect.value,
        value: "",
        ...(alertData.type === "metric" ? { datatype: datatypeSelect?.value || "numeric" } : {})
    };
}

document.addEventListener("DOMContentLoaded", () => {
    initStepper();
    updateStepperVisual();

});