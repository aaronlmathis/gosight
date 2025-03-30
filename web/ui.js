import { formatMetricValue, formatUptime, formatTimestamp } from './utils.js';

export function renderHostSection(hostMetrics, interfaces, totals, thresholds) {
  const section = document.createElement('section');
  section.className = 'bg-gray-900 rounded-xl p-4 shadow space-y-4';

  section.innerHTML = `
    <h2 class="text-xl font-semibold text-white">Host Metrics</h2>
    <div class="grid grid-cols-2 md:grid-cols-3 gap-4">
      ${hostMetrics.map(m => renderMetric(m, thresholds)).join('')}
    </div>
    ${renderInterfaceSections(interfaces)}
    ${totals.length > 0 ? `
      <h3 class="text-lg font-semibold mt-4">Network Totals</h3>
      <div class="grid grid-cols-2 md:grid-cols-3 gap-4">
        ${totals.map(m => renderMetric(m, thresholds)).join('')}
      </div>` : ''}
  `;

  return section;
}

export function renderContainerSections(containerGroups, meta, thresholds) {
  return Object.entries(containerGroups).map(([name, metrics]) => {
    const metaInfo = meta[name] || [];

    return `
      <section class="bg-gray-900 rounded-xl p-4 shadow space-y-2">
        <div class="flex justify-between items-center">
          <h2 class="text-xl font-semibold text-white">${name}</h2>
          <span class="text-xs text-gray-400">${getContainerState(metaInfo)}</span>
        </div>
        ${renderMetaInfo(metaInfo)}
        <div class="grid grid-cols-2 md:grid-cols-3 gap-4">
          ${metrics.map(m => renderMetric(m, thresholds)).join('')}
        </div>
      </section>`;
  }).join('');
}

function renderInterfaceSections(interfaces) {
  return Object.entries(interfaces).map(([iface, metrics]) => {
    return `
      <div>
        <h3 class="text-lg font-semibold mt-4 text-white">Interface: ${iface}</h3>
        <div class="grid grid-cols-2 md:grid-cols-3 gap-4">
          ${metrics.map(m => renderMetric(m)).join('')}
        </div>
      </div>`;
  }).join('');
}

function renderMetric(metric, thresholds = {}) {
  const t = thresholds[metric.full];
  let cls = 'text-sm text-white';
  if (t) {
    if (metric.value > t.high) cls = 'text-red-400';
    else if (metric.value < t.low) cls = 'text-green-400';
  }

  return `<div class="bg-gray-800 p-3 rounded-md">
    <div class="text-gray-300">${metric.name}</div>
    <div class="${cls}">${formatMetricValue(metric.value)}</div>
  </div>`;
}

function getContainerState(metaList) {
  const state = metaList.find(m => m.name === 'state');
  return state ? state.value : '';
}

function renderMetaInfo(metaList) {
  const fieldsToShow = ['image', 'created_at', 'uptime_seconds'];
  return `
    <div class="text-xs text-gray-400 space-x-4">
      ${metaList.filter(m => fieldsToShow.includes(m.name))
        .map(m => `<span><strong>${m.name.replace('_', ' ')}:</strong> ${formatMetaValue(m)}</span>`).join('')}
    </div>`;
}

function formatMetaValue(m) {
  if (m.name === 'created_at') return formatTimestamp(m.value);
  if (m.name === 'uptime_seconds') return formatUptime(m.value);
  return m.value;
}
