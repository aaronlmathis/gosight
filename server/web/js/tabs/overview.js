document.addEventListener("DOMContentLoaded", () => {
    renderMiniChart("miniCpuChart");
    renderMiniChart("miniMemoryChart");
    renderMiniChart("miniDiskChart");
  });
  
  function renderMiniChart(id) {
    const ctx = document.getElementById(id);
    if (!ctx) return;
  
    new Chart(ctx, {
      type: "line",
      data: {
        labels: [], // to be filled with timestamps
        datasets: [{
          label: id,
          data: [],   // to be filled with metric values
          borderWidth: 1,
          borderColor: "rgba(59, 130, 246, 1)", // Tailwind blue-500
          tension: 0.3,
          pointRadius: 0,
        }],
      },
      options: {
        responsive: true,
        plugins: { legend: { display: false } },
        scales: {
          x: { display: false },
          y: { display: false }
        },
      },
    });
  }
  