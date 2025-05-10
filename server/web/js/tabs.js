// tabs.js
const tabInitRegistry = {};

const observer = new MutationObserver(() => {
    for (const [tabId, initFn] of Object.entries(tabInitRegistry)) {
        const panel = document.getElementById(tabId);
        if (panel && !panel.classList.contains("hidden") && !panel._initialized) {
            panel._initialized = true;
            initFn();
        }
    }
});

observer.observe(document.body, { childList: true, subtree: true });

export function registerTabInitializer(tabId, initFn) {
    tabInitRegistry[tabId] = initFn;
}