async function fetchContainers() {
    try {
      const res = await fetch("/api/containers");
      if (!res.ok) throw new Error("Failed to fetch container data");
      return await res.json();
    } catch (err) {
      console.error("[Container Fetch Error]", err);
      return [];
    }
  }
  
  function formatBytes(bytes) {
    if (bytes === 0 || bytes == null) return "—";
    const sizes = ["B", "KB", "MB", "GB", "TB"];
    const i = Math.floor(Math.log(bytes) / Math.log(1024));
    return (bytes / Math.pow(1024, i)).toFixed(1) + " " + sizes[i];
  }
  
  function formatUptime(seconds) {
    if (!seconds || seconds >= 9e9) return "—"; // catch absurd values
    const h = Math.floor(seconds / 3600);
    const m = Math.floor((seconds % 3600) / 60);
    return `${h}h ${m}m`;
  }
  
  function renderContainers(containers) {
    const tbody = document.getElementById("container-table");
    tbody.innerHTML = "";
  
    if (!containers.length) {
      tbody.innerHTML = `<tr><td colspan="9" class="text-center py-6 text-gray-500">No container data available</td></tr>`;
      return;
    }
  
    for (const ctr of containers) {
      const row = document.createElement("tr");
  
      row.innerHTML = `
        <td class="px-4 py-2">${ctr.host || "—"}</td>
        <td class="px-4 py-2 font-medium">${ctr.name}</td>
        <td class="px-4 py-2">${ctr.image}</td>
        <td class="px-4 py-2">${ctr.status}</td>
        <td class="px-4 py-2">${ctr.cpu != null ? ctr.cpu.toFixed(1) + "%" : "—"}</td>
        <td class="px-4 py-2">${formatBytes(ctr.mem)}</td>
        <td class="px-4 py-2">${formatBytes(ctr.rx)}</td>
        <td class="px-4 py-2">${formatBytes(ctr.tx)}</td>
        <td class="px-4 py-2">${formatUptime(ctr.uptime)}</td>
      `;
  
      tbody.appendChild(row);
    }
  }
  
  async function updateContainersPage() {
    const data = await fetchContainers();
    renderContainers(data);
  
    const lastUpdate = document.getElementById("lastUpdate");
    if (lastUpdate) {
      lastUpdate.textContent = `Last updated: ${new Date().toLocaleTimeString()}`;
    }
  }
  
  document.addEventListener("DOMContentLoaded", () => {
    updateContainersPage();
    setInterval(updateContainersPage, 10000);
  });
  