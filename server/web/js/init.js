import { initWebSockets } from "./ws.js";

document.addEventListener('DOMContentLoaded', () => {
    initWebSockets(window.endpointID);
});