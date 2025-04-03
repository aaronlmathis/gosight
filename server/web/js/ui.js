/*
SPDX-License-Identifier: GPL-3.0-or-later

Copyright (C) 2025 Aaron Mathis aaron.mathis@gmail.com

This file is part of GoSight.

GoSight is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

GoSight is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with GoSight. If not, see https://www.gnu.org/licenses/.
*/
// server/web/js/ui.js

// Toggle sidebar visibility on mobile
const sidebarToggle = document.getElementById('sidebarToggle');
const sidebar = document.getElementById('sidebar');

if (sidebarToggle && sidebar) {
  sidebarToggle.addEventListener('click', () => {
    sidebar.classList.toggle('hidden');
  });
}

// Dark mode toggle (optional)
const darkToggle = document.getElementById('darkToggle');

if (darkToggle) {
  darkToggle.addEventListener('click', () => {
    document.documentElement.classList.toggle('dark');
    document.body.classList.toggle('bg-gray-900');
    document.body.classList.toggle('text-gray-100');
  });
}

// Navigation logic placeholder
const navLinks = document.querySelectorAll('aside nav a');
navLinks.forEach(link => {
  link.addEventListener('click', (e) => {
    e.preventDefault();
    navLinks.forEach(l => l.classList.remove('bg-gray-100'));
    link.classList.add('bg-gray-100');
    // TODO: Load content dynamically if using SPA or update view context
  });
});

export function toggleMenu(id) {
  const el = document.getElementById(id);
  el.classList.toggle('max-h-0');
  el.classList.toggle('max-h-[999px]');
}