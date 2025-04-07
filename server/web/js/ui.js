console.log("🛠️ ui.js loaded");
// server/web/js/ui.js
document.addEventListener('DOMContentLoaded', () => {
  console.log("🛠️ ui.js DOM ready");

  const darkToggle = document.getElementById('darkToggle');
  if (darkToggle) {
    darkToggle.addEventListener('click', () => {
      const html = document.documentElement;
      const isDark = html.classList.toggle('dark');
      localStorage.setItem('theme', isDark ? 'dark' : 'light');
    });

    // Apply saved theme
    let saved = localStorage.getItem('theme');
    if (!saved) {
      saved = window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
    }
    if (saved === 'dark') {
      document.documentElement.classList.add('dark');
    } else {
      document.documentElement.classList.remove('dark');
    }
  }
});