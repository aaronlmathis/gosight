document.addEventListener("DOMContentLoaded", () => {
    renderDiskUsageDonut("diskUsageDonutChart");
    renderBarChart("inodeUsageBarChart", "Inodes");
  });
  
  function renderDiskUsageDonut(id) {
    const ctx = document.getElementById(id);
    if (!ctx) return;
  
    new Chart(ctx, {
      type: "doughnut",
      data: {
        labels: ["Used", "Free"],
        datasets: [{
          data: [],
          backgroundColor: ["#f87171", "#34d399"],
          borderWidth: 1,
        }],
      },
      options: {
        responsive: true,
        cutout: "70%",
        plugins: {
          legend: { position: "right" }
        },
      },
    });
  }
  
  function renderBarChart(id, label) {
    const ctx = document.getElementById(id);
    if (!ctx) return;
  
    new Chart(ctx, {
      type: "bar",
      data: {
        labels: [],
        datasets: [{
          label: label,
          data: [],
          backgroundColor: "#60a5fa",
        }],
      },
      options: {
        responsive: true,
        scales: {
          x: { display: true },
          y: { beginAtZero: true },
        },
      },
    });
  }
  