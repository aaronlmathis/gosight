// /js/theme-toggle.js
document.addEventListener("DOMContentLoaded", () => {
    const toggleBtn = document.getElementById("theme-toggle");
    const darkIcon = document.getElementById("theme-toggle-dark-icon");
    const lightIcon = document.getElementById("theme-toggle-light-icon");
  
    // Load initial state
    if (
      localStorage.getItem("color-theme") === "dark" ||
      (!localStorage.getItem("color-theme") &&
        window.matchMedia("(prefers-color-scheme: dark)").matches)
    ) {
      document.documentElement.classList.add("dark");
      darkIcon?.classList.remove("hidden");
    } else {
      lightIcon?.classList.remove("hidden");
    }
  
    toggleBtn?.addEventListener("click", () => {
      darkIcon?.classList.toggle("hidden");
      lightIcon?.classList.toggle("hidden");
  
      const currentTheme = document.documentElement.classList.contains("dark")
        ? "dark"
        : "light";
  
      const newTheme = currentTheme === "dark" ? "light" : "dark";
      document.documentElement.classList.toggle("dark");
      localStorage.setItem("color-theme", newTheme);
    });
  });