document.addEventListener("DOMContentLoaded", () => {
    renderNetworkChart();
  
    const unitSelect = document.getElementById("chart-unit");
    unitSelect?.addEventListener("change", updateChartUnits);
  
    const exportBtn = document.getElementById("export-csv");
    exportBtn?.addEventListener("click", exportTableToCSV);
  });
  
  function renderNetworkChart() {
    const ctx = document.getElementById("networkChart");
    if (!ctx) return;
  
    new Chart(ctx, {
      type: "line",
      data: {
        labels: [],
        datasets: [{
          label: "Upload",
          data: [],
          borderColor: "#10b981",
          fill: false,
          tension: 0.3,
          pointRadius: 0,
        }, {
          label: "Download",
          data: [],
          borderColor: "#3b82f6",
          fill: false,
          tension: 0.3,
          pointRadius: 0,
        }],
      },
      options: {
        responsive: true,
        plugins: { legend: { position: "top" } },
        scales: {
          x: { display: true },
          y: { beginAtZero: true },
        },
      },
    });
  }
  
  function updateChartUnits() {
    console.log("TODO: convert chart values to new unit");
  }
  
  function exportTableToCSV() {
    console.log("TODO: implement CSV export for interface table");
  }
  