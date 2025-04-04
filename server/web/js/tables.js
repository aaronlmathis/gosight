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
// server/web/js/tables.js

export function renderAgentTable(agents) {
  const tbody = document.querySelector('#agentTableBody');
  tbody.innerHTML = '';

  for (const a of agents) {
    let statusColor = 'bg-red-100 text-red-800';
    if (a.status === 'Online') {
      statusColor = 'bg-green-100 text-green-800';
    } else if (a.status === 'Idle') {
      statusColor = 'bg-yellow-100 text-yellow-800';
    }

    tbody.insertAdjacentHTML('beforeend', `
      <tr class="hover:bg-gray-50">
        <td class="px-6 py-4 font-medium text-gray-900">${a.hostname}</td>
        <td class="px-6 py-4">${a.ip}</td>
        <td class="px-6 py-4">${a.os || '-'}</td>
        <td class="px-6 py-4">${a.since}</td>
        <td class="px-6 py-4">
          <span class="inline-block px-2 py-1 text-xs font-semibold rounded-full ${statusColor}">
            ${a.status}
          </span>
        </td>
      </tr>
    `);
  }
}
