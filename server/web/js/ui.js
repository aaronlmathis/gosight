import { formatMetricValue, formatUptime, formatTimestamp } from './utils.js';

export function renderHostSection(hostMetrics, interfaces, totals, thresholds) {
  const section = document.createElement('section');
  section.className = 'mb-8';

  section.innerHTML = `
    <h2 class="text-xl font-semibold mb-2 text-white border-b border-gray-700 pb-1">Host Metrics</h2>
    ${renderTable(hostMetrics, thresholds)}
    ${Object.keys(interfaces).length > 0 ? `<h3 class="mt-6 text-white font-semibold">Host Network Interfaces</h3>` : ''}
    ${renderInterfaces(interfaces)}
    ${totals.length ? `
      <h3 class="mt-6 text-white font-semibold">Network Totals</h3>
      ${renderTable(totals, thresholds)}
    ` : ''}
  `;

  return section;
}

export function renderContainerSections(containerGroups, meta, thresholds) {
  return Object.entries(containerGroups).map(([name, metrics]) => {
    const metaInfo = meta[name] || [];
    return `
      <section class="mb-8">
        <div class="flex justify-between items-center">
          <h2 class="text-xl font-semibold text-white mb-1 border-b border-gray-700 pb-1">${name}</h2>
          <span class="text-xs text-gray-400">${getContainerState(metaInfo)}</span>
        </div>
        ${renderMetaInfo(metaInfo)}
        ${renderTable(metrics, thresholds)}
      </section>
    `;
  }).join('');
}

function renderInterfaces(interfaces) {
  return Object.entries(interfaces).map(([iface, metrics]) => {
    return `
      <h4 class="mt-4 text-gray-300 font-semibold">${iface}</h4>
      ${renderTable(metrics)}
    `;
  }).join('');
}

function renderTable(metrics, thresholds = {}) {
  return `
    <div class="overflow-x-auto rounded-lg shadow border border-gray-700">
      <table class="w-full text-sm text-left border-collapse text-gray-300">
        <thead class="bg-gray-800 text-gray-400">
          <tr>
            <th class="px-4 py-2">Metric</th>
            <th class="px-4 py-2">Value</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-gray-800">
          ${metrics.map(m => renderRow(m, thresholds)).join('')}
        </tbody>
      </table>
    </div>
  `;
}

function renderRow(metric, thresholds = {}) {
  const t = thresholds[metric.full];
  let cls = '';
  if (t) {
    if (metric.value > t.high) cls = 'text-red-400';
    else if (metric.value < t.low) cls = 'text-green-400';
  }

  return `
    <tr>
      <td class="px-4 py-2 text-left font-mono">${metric.name}</td>
      <td class="px-4 py-2 font-mono text-left ${cls}">${formatMetricValue(metric.value, metric.name)}</td>
    </tr>
  `;
}

function renderMetaInfo(metaList) {
  const fieldsToShow = ['image', 'created_at', 'uptime_seconds'];
  return `
    <div class="text-xs text-gray-400 mb-2 space-x-4">
      ${metaList.filter(m => fieldsToShow.includes(m.name))
        .map(m => `<span><strong>${m.name.replace('_', ' ')}:</strong> ${formatMetaValue(m)}</span>`).join('')}
    </div>`;
}

function getContainerState(metaList) {
  const state = metaList.find(m => m.name === 'state');
  return state ? state.value : '';
}

function formatMetaValue(m) {
  if (m.name === 'created_at') return formatTimestamp(m.value);
  if (m.name === 'uptime_seconds') return formatUptime(m.value);
  return m.value;
}
