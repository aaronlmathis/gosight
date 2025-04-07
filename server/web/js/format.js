export function formatBytes(bytes) {
    if (bytes === 0 || bytes == null) return "—";
    const sizes = ["B", "KB", "MB", "GB", "TB"];
    const i = Math.floor(Math.log(bytes) / Math.log(1024));
    return (bytes / Math.pow(1024, i)).toFixed(1) + " " + sizes[i];
  }
  
export  function formatUptime(seconds) {
    if (!seconds || seconds >= 9e9) return "—"; // catch absurd values
    const h = Math.floor(seconds / 3600);
    const m = Math.floor((seconds % 3600) / 60);
    return `${h}h ${m}m`;
  }