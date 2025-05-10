document.addEventListener("DOMContentLoaded", () => {
    const filter = document.getElementById("activityFilter");
    const exportBtn = document.getElementById("export-csv");
  
    if (filter) {
      filter.addEventListener("input", () => {
        console.log("TODO: filter activity logs");
      });
    }
  
    if (exportBtn) {
      exportBtn.addEventListener("click", () => {
        console.log("TODO: export activity table to CSV");
      });
    }
  });
  