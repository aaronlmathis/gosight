import { registerTabInitializer } from "./tabs.js";
import { gosightFetch } from "./api.js";

registerTabInitializer("console", () => {
    console.log("Endpoint metadata: ", window.endpointMetadata);
    const input = document.getElementById("console-command");
    const output = document.getElementById("console-output");
    const responsesEl = document.getElementById("console-responses");

    const history = [];
    let historyIndex = -1;

    if (!input || !output || input._bound) return;
    input._bound = true;

    input.addEventListener("keydown", async (e) => {
        if (e.key === "ArrowUp") {
            if (history.length === 0) return;
            historyIndex = Math.max(0, historyIndex - 1);
            input.value = history[historyIndex];
            return;
        }

        if (e.key === "ArrowDown") {
            if (history.length === 0) return;
            historyIndex = Math.min(history.length - 1, historyIndex + 1);
            input.value = history[historyIndex] || "";
            return;
        }

        if (e.key !== "Enter") return;

        const cmd = input.value.trim();
        if (!cmd) return;

        history.push(cmd);
        historyIndex = history.length;

        const inputText = input.value.trim();

        // Split input by whitespace
        const parts = inputText.split(/\s+/);

        // Extract command and arguments
        const command_data = parts[0] || "";
        const args = parts.slice(1);

        // Example use
        const payload = {
            agent_id: window.endpointMetadata.agent_id,
            command_type: "shell",
            command_data,
            args
        };

        console.log("Command payload:", payload);

        // Echo command
        const echo = document.createElement("div");
        echo.innerHTML = `<span class="text-blue-400">user</span>@<span class="text-purple-400">host</span>:<span class="text-red-400">~</span>$ <span class="text-green-400">${cmd}</span>`;
        responsesEl.appendChild(echo);

        // Show placeholder while waiting
        const pending = document.createElement("div");
        pending.className = "text-gray-500 whitespace-pre-wrap";
        pending.textContent = "[executing...]";
        responsesEl.appendChild(pending);

        input.value = "";
        output.scrollTop = output.scrollHeight;
        input.focus();

        try {
            console.log("Sending command to agent: ", window.endpointMetadata.agent_id);
            await gosightFetch("/api/v1/command", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify(payload)
            })
                .then(res => res.json())
                .then(data => {
                    console.log("Command response:", data); // ✅ This
                })
                .catch(err => console.error("Command request failed:", err));
            // response will arrive via websocket
        } catch (err) {
            pending.className = "text-red-400";
            pending.textContent = "❌ " + err.message;
        }
    });

    // Listen for command_result events from websocket
    window.addEventListener("command", (e) => {
        const result = e.detail;
        console.log("Received command result:", result);
        // Confirm it's for this endpoint
        if (result.endpoint_id !== window.endpointID) return;

        if (window._lastPendingConsoleLine) {
            window._lastPendingConsoleLine.remove();
            window._lastPendingConsoleLine = null;
        }
        const div = document.createElement("div");
        div.className = result.success ? "text-gray-400 whitespace-pre-wrap" : "text-red-400 whitespace-pre-wrap";
        div.textContent = result.output || result.error_message || "[no output]";

        responsesEl.appendChild(div);
        responsesEl.lastElementChild?.scrollIntoView({ behavior: "smooth", block: "end" });

    });
});
