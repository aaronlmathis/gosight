document.addEventListener("DOMContentLoaded", () => {
    const inputEl = document.getElementById("console-input");
    if (!inputEl) return;
  
    const fakeCommand = "uptime";
    let idx = 0;
  
    function typeCommand() {
      if (idx < fakeCommand.length) {
        inputEl.classList.remove("animate-pulse");
        inputEl.innerText = fakeCommand.slice(0, idx + 1);
        idx++;
        setTimeout(typeCommand, 100);
      } else {
        inputEl.classList.add("animate-pulse");
      }
    }
  
    setTimeout(typeCommand, 800);
  });
  