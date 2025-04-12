// /js/user-menu.js
document.addEventListener("DOMContentLoaded", () => {
    const toggle = document.getElementById("user-menu-button");
    const dropdown = document.getElementById("user-dropdown");
  
    if (toggle && dropdown) {
      toggle.addEventListener("click", () => {
        dropdown.classList.toggle("hidden");
      });
  
      document.addEventListener("click", (e) => {
        if (!dropdown.contains(e.target) && !toggle.contains(e.target)) {
          dropdown.classList.add("hidden");
        }
      });
    }
  });
  