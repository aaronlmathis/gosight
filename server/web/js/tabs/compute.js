document.addEventListener("DOMContentLoaded", () => {
    renderLineChart("cpuUsageChart", "CPU Usage (%)");
    renderLineChart("cpuLoadChart", "CPU Load Average");
    renderDonutChart("cpuDonutChart", ["User", "System", "Idle"]);
  });
  
  function renderLineChart(id, label) {
    const ctx = document.getElementById(id);
    if (!ctx) return;
  
    new Chart(ctx, {
      type: "line",
      data: {
        labels: [],
        datasets: [{
          label: label,
          data: [],
          borderColor: "rgba(59, 130, 246, 1)",
          tension: 0.3,
          fill: false,
          pointRadius: 0,
        }],
      },
      options: {
        responsive: true,
        plugins: { legend: { display: false } },
        scales: {
          x: { display: true },
          y: { beginAtZero: true },
        },
      },
    });
  }
  
  function renderDonutChart(id, segments) {
    const ctx = document.getElementById(id);
    if (!ctx) return;
  
    new Chart(ctx, {
      type: "doughnut",
      data: {
        labels: segments,
        datasets: [{
          data: [],
          backgroundColor: ["#3b82f6", "#10b981", "#fbbf24"],
          borderWidth: 1,
        }],
      },
      options: {
        responsive: true,
        cutout: "70%",
        plugins: {
          legend: { position: "right" },
        },
      },
    });
  }
  