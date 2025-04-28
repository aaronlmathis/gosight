import { initWebSockets } from "./ws.js";
import { loadEndpointHeader, setupTagButton, setupTagsModal } from "./singleEndpoint.js";

document.addEventListener('DOMContentLoaded', async () => {
    await loadEndpointHeader();

    await setupTagsModal();
    initWebSockets(window.endpointID);
});