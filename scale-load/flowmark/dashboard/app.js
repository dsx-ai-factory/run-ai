/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

'use strict';

const DATA_BASE = (() => {
  const fromQuery = new URLSearchParams(location.search).get('data');
  const placeholder = '__DATA_BASE_URL__';
  const raw = fromQuery || (placeholder.startsWith('__') ? '.' : placeholder);
  return raw.replace(/\/$/, '');
})();

const CATEGORY_ORDER = ['scale', 'load'];
const LINE_COLOR = '#2f6fdb';
const FLAG_COLOR = '#d0342c';
const SLO_COLOR = '#8a8f98';

const state = { runs: [], unreadable: [], clusterType: '', tab: 'runs', category: '', filter: '', expanded: new Set() };
const charts = [];

function esc(value) {
  return String(value).replace(/[&<>"']/g, (c) => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[c]));
}

function fmtDate(iso) {
  const d = new Date(iso);
  const month = d.toLocaleString('en-US', { month: 'short', timeZone: 'UTC' });
  const pad = (n) => String(n).padStart(2, '0');
  return `${month} ${d.getUTCDate()}, ${d.getUTCFullYear()} ${pad(d.getUTCHours())}:${pad(d.getUTCMinutes())}`;
}

function shortDate(iso) {
  const d = new Date(iso);
  return `${d.getUTCDate()}/${d.getUTCMonth() + 1}`;
}

function baseName(metric) {
  return metric.split('@')[0];
}

function unitOf(metric) {
  const base = baseName(metric);
  if (base.endsWith('_ms')) return 's';
  if (base.endsWith('_per_sec')) return '/s';
  return '';
}

function toUnit(metric, value) {
  if (value === null || value === undefined) return null;
  return unitOf(metric) === 's' ? value / 1000 : value;
}

function fmtMetric(metric, value) {
  if (value === null || value === undefined) return '–';
  const unit = unitOf(metric);
  if (unit === 's') return value >= 1000 ? `${(value / 1000).toFixed(1)} s` : `${Math.round(value)} ms`;
  if (unit === '/s') return `${Number(value).toFixed(1)}/s`;
  return String(value);
}

function sloFor(section, metric) {
  const slo = section.slo || {};
  if (slo[metric] !== undefined) return slo[metric];
  if (slo[baseName(metric)] !== undefined) return slo[baseName(metric)];
  return null;
}

function runStatus(doc) {
  const failed = Object.entries(doc.sections || {}).filter(([, s]) => s.failed).map(([name]) => name);
  if (failed.length) {
    return { kind: 'failed', label: '✕ failed', text: `${failed.length} scenario${failed.length > 1 ? 's' : ''} failed: ${failed.join(', ')}` };
  }
  const regressions = doc.regressions || [];
  if (doc.regression_detected && regressions.length) {
    const more = regressions.length > 1 ? ` +${regressions.length - 1} more` : '';
    return { kind: 'flagged', label: '! flagged', text: regressionShort(regressions[0]) + more };
  }
  return { kind: 'passed', label: '✓ passed', text: 'all in range' };
}

function regressionShort(regression) {
  const match = regression.message.match(/(\+\d+% over the median|over the SLO|under the SLO)/);
  const how = match ? match[1].replace('over the median', 'vs median') : 'flagged';
  return `${regression.metric} ${how}`;
}

function sectionSummary(section) {
  const m = section.metrics || {};
  const parts = [];
  const rounds = Object.keys(m).filter((k) => k.startsWith('p90_ms@') && !k.endsWith('_regression')).map((k) => Number(k.split('@')[1])).sort((a, b) => a - b);
  if (rounds.length) {
    parts.push('p90 ' + rounds.map((r) => `${fmtMetric('p90_ms', m[`p90_ms@${r}`])} @${r}`).join(' · '));
    const throughput = rounds.filter((r) => m[`throughput_per_sec@${r}`] !== undefined && m[`throughput_per_sec@${r}`] !== null);
    if (throughput.length) parts.push('throughput ' + throughput.map((r) => `${fmtMetric('throughput_per_sec', m[`throughput_per_sec@${r}`])} @${r}`).join(' · '));
  }
  for (const span of ['ttj', 'api_accepted', 'cluster_synced', 'pod_created']) {
    const value = m[`${span}_p90_ms`];
    if (value !== undefined && value !== null) parts.push(`${span.replace('_', ' ')} p90 ${fmtMetric('p90_ms', value)}`);
  }
  if (m.failed_count !== undefined && m.failed_count !== null) parts.push(`${m.failed_count} rejected`);
  if (!parts.length) {
    const numbers = Object.entries(m).filter(([k, v]) => typeof v === 'number' && !k.endsWith('_regression')).slice(0, 4);
    parts.push(numbers.map(([k, v]) => `${k} ${fmtMetric(k, v)}`).join(' · ') || 'no metrics');
  }
  return parts.join(' · ');
}

function sectionVerdict(name, doc) {
  const section = doc.sections[name];
  if (section.failed) return { kind: 'failed', text: section.error || 'failed' };
  const regressions = (doc.regressions || []).filter((r) => r.section === name);
  if (regressions.length) return { kind: 'flagged', text: regressions.map((r) => r.message.replace(`${name} `, '')).join('; ') };
  return { kind: 'passed', text: 'in range' };
}

function selectedRuns() {
  return state.runs.filter((run) => run.doc.metadata.cluster_type === state.clusterType);
}

function clusterTypes() {
  return [...new Set(state.runs.map((run) => run.doc.metadata.cluster_type))].sort();
}

function categories(runs) {
  const found = new Set();
  for (const { doc } of runs) for (const section of Object.values(doc.sections || {})) found.add(section.category || 'other');
  return [...found].sort((a, b) => {
    const ia = CATEGORY_ORDER.indexOf(a); const ib = CATEGORY_ORDER.indexOf(b);
    return (ia < 0 ? 99 : ia) - (ib < 0 ? 99 : ib) || a.localeCompare(b);
  });
}

function metricOrder(metric) {
  const base = baseName(metric);
  if (base.includes('p90')) return 0;
  if (base.endsWith('_per_sec')) return 1;
  if (base === 'failed_count') return 2;
  return 3;
}

function latestSections(runs) {
  const latest = new Map();
  for (const { doc } of runs) for (const [name, section] of Object.entries(doc.sections || {})) latest.set(name, section);
  return latest;
}

function chartSpecs(runs) {
  const specs = new Map();
  const latest = latestSections(runs);
  for (const { doc } of runs) {
    for (const [name, section] of Object.entries(doc.sections || {})) {
      for (const [metric, value] of Object.entries(section.metrics || {})) {
        if (typeof value !== 'number' || metric.endsWith('_regression')) continue;
        if (metric.includes('p90') || sloFor(latest.get(name), metric) !== null) {
          specs.set(`${name}|${metric}`, { section: name, metric, category: section.category || 'other' });
        }
      }
    }
  }
  return [...specs.values()].sort((a, b) =>
    a.section.localeCompare(b.section) || metricOrder(a.metric) - metricOrder(b.metric) ||
    (Number(a.metric.split('@')[1] || 0) - Number(b.metric.split('@')[1] || 0)) || a.metric.localeCompare(b.metric));
}

function chartTitle(spec) {
  const [base, round] = spec.metric.split('@');
  let label = base.replace(/_ms$/, '').replace(/_/g, ' ');
  if (base.endsWith('_per_sec')) label = 'throughput';
  const unit = unitOf(spec.metric);
  return `${spec.section} — ${label}${round ? ` @${round}` : ''}${unit ? ` (${unit})` : ''}`;
}

function buildChart(canvas, spec, runs) {
  const labels = runs.map((run) => shortDate(run.doc.metadata.date));
  const values = runs.map((run) => toUnit(spec.metric, run.doc.sections[spec.section]?.metrics?.[spec.metric] ?? null));
  const flagged = runs.map((run) => run.doc.sections[spec.section]?.metrics?.[`${spec.metric}_regression`] === true);
  const slos = runs.map((run) => {
    const section = run.doc.sections[spec.section];
    return section ? toUnit(spec.metric, sloFor(section, spec.metric)) : null;
  });
  const versions = runs.map((run) => run.doc.metadata.control_plane_version || 'unknown');
  const hasSLO = slos.some((v) => v !== null);
  const maxValue = Math.max(0, ...values.filter((v) => v !== null));
  const maxSLO = Math.max(0, ...slos.filter((v) => v !== null));
  const sloOffScale = hasSLO && maxValue > 0 && maxSLO > 3 * maxValue;
  const latestRawSLO = runs.map((run) => run.doc.sections[spec.section]).filter(Boolean)
    .map((section) => sloFor(section, spec.metric)).filter((v) => v !== null).pop();
  const datasets = [{
    label: chartTitle(spec).split(' — ')[1],
    data: values,
    borderColor: LINE_COLOR,
    backgroundColor: LINE_COLOR,
    pointBackgroundColor: flagged.map((f) => (f ? FLAG_COLOR : LINE_COLOR)),
    pointBorderColor: flagged.map((f) => (f ? FLAG_COLOR : LINE_COLOR)),
    pointRadius: flagged.map((f) => (f ? 7 : 3.5)),
    pointHoverRadius: 8,
    tension: 0.15,
    spanGaps: false,
  }];
  if (hasSLO) {
    datasets.push({
      label: `SLO${unitOf(spec.metric) === '/s' ? ' (floor)' : ''}${sloOffScale ? ` ${fmtMetric(spec.metric, latestRawSLO)}, off scale` : ''}`,
      data: sloOffScale ? slos.map(() => null) : slos,
      borderColor: SLO_COLOR,
      borderDash: [6, 5],
      borderWidth: 1.5,
      pointRadius: runs.length === 1 ? 3 : 0,
      pointHoverRadius: 0,
      fill: false,
    });
  }
  const chart = new Chart(canvas, {
    type: 'line',
    data: { labels, datasets },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      animation: false,
      interaction: { mode: 'nearest', intersect: false },
      plugins: {
        legend: { display: hasSLO, labels: { boxWidth: 18, font: { size: 11 } } },
        tooltip: {
          callbacks: {
            title: (items) => fmtDate(runs[items[0].dataIndex].doc.metadata.date),
            label: (item) => {
              const run = runs[item.dataIndex];
              const raw = run.doc.sections[spec.section]?.metrics?.[spec.metric];
              if (item.datasetIndex === 1) return `SLO ${fmtMetric(spec.metric, sloFor(run.doc.sections[spec.section], spec.metric))}`;
              return `${fmtMetric(spec.metric, raw)}${flagged[item.dataIndex] ? ' — flagged' : ''}`;
            },
            afterBody: (items) => [`cp ${versions[items[0].dataIndex]}`],
          },
        },
      },
      scales: {
        y: { beginAtZero: true, title: { display: Boolean(unitOf(spec.metric)), text: unitOf(spec.metric) } },
        x: { grid: { display: false } },
      },
    },
  });
  charts.push(chart);
}

function renderHeader() {
  const runs = selectedRuns();
  const latest = runs[runs.length - 1];
  const subtitle = document.getElementById('subtitle');
  if (!latest) {
    subtitle.textContent = state.runs.length ? 'no runs for this cluster type' : 'no runs found';
  } else {
    subtitle.innerHTML = `latest run ${esc(fmtDate(latest.doc.metadata.date))} UTC · cp ${esc(latest.doc.metadata.control_plane_version || 'unknown')} · ${runs.length} run${runs.length === 1 ? '' : 's'} on record`;
  }
  const select = document.getElementById('cluster-type');
  select.innerHTML = clusterTypes().map((type) => `<option value="${esc(type)}"${type === state.clusterType ? ' selected' : ''}>${esc(type)}</option>`).join('');
}

function renderRuns() {
  const container = document.getElementById('runs');
  const runs = selectedRuns().slice().reverse();
  if (!runs.length) {
    container.innerHTML = '<p class="empty">No runs to show.</p>';
    return;
  }
  const rows = runs.map((run) => {
    const { doc, file } = run;
    const status = runStatus(doc);
    const sections = Object.keys(doc.sections || {}).sort();
    const open = state.expanded.has(file);
    const detail = sections.map((name) => {
      const verdict = sectionVerdict(name, doc);
      return `<div class="section-line"><span class="name">${esc(name)}</span><span>${esc(sectionSummary(doc.sections[name]))}</span><span class="verdict ${verdict.kind}">${esc(verdict.text)}</span></div>`;
    }).join('');
    return `
      <tr class="run" data-file="${esc(file)}">
        <td><span class="caret">${open ? '▼' : '▶'}</span> <span class="date">${esc(fmtDate(doc.metadata.date))}</span></td>
        <td><span class="version" title="cluster ${esc(doc.metadata.cluster_version || 'unknown')}">cp ${esc(doc.metadata.control_plane_version || 'unknown')}</span></td>
        <td><a class="results-link" href="${esc(`${DATA_BASE}/${file}`)}" download onclick="event.stopPropagation()">results</a></td>
        <td>${sections.length} scenario${sections.length === 1 ? '' : 's'}</td>
        <td class="check ${status.kind}">${esc(status.text)}</td>
        <td><span class="badge ${status.kind}">${esc(status.label)}</span></td>
      </tr>
      ${open ? `<tr class="detail"><td colspan="6">${detail}</td></tr>` : ''}`;
  }).join('');
  const unreadable = state.unreadable.length ? `<p class="unreadable">${state.unreadable.length} listed run${state.unreadable.length === 1 ? '' : 's'} could not be read: ${esc(state.unreadable.join(', '))}</p>` : '';
  container.innerHTML = `
    <table class="runs">
      <thead><tr><th>Date (UTC)</th><th>Version (control plane build)</th><th>Results</th><th>Scenarios</th><th>Regression check</th><th>Status</th></tr></thead>
      <tbody>${rows}</tbody>
    </table>${unreadable}`;
  for (const row of container.querySelectorAll('tr.run')) {
    row.addEventListener('click', () => {
      const file = row.dataset.file;
      if (state.expanded.has(file)) state.expanded.delete(file); else state.expanded.add(file);
      renderRuns();
    });
  }
}

function renderMetrics() {
  const runs = selectedRuns();
  const cats = categories(runs);
  if (!cats.includes(state.category)) state.category = cats[0] || '';
  document.getElementById('subtabs').innerHTML = cats.map((cat) =>
    `<button class="subtab${cat === state.category ? ' active' : ''}" data-category="${esc(cat)}">${esc(cat[0].toUpperCase() + cat.slice(1))}</button>`).join('');
  for (const button of document.querySelectorAll('.subtab')) {
    button.addEventListener('click', () => { state.category = button.dataset.category; renderMetrics(); });
  }
  while (charts.length) charts.pop().destroy();
  const container = document.getElementById('charts');
  const filter = state.filter.trim().toLowerCase();
  const specs = chartSpecs(runs).filter((spec) => spec.category === state.category)
    .filter((spec) => !filter || chartTitle(spec).toLowerCase().includes(filter) || spec.metric.toLowerCase().includes(filter));
  if (!specs.length) {
    container.innerHTML = `<p class="empty">${runs.length ? 'No charts match.' : 'No runs to show.'}</p>`;
    return;
  }
  history.replaceState(null, '', `#metrics/${state.category}`);
  container.innerHTML = specs.map((spec, i) => `<div class="chart-card"><h3>${esc(chartTitle(spec))}</h3><canvas id="chart-${i}"></canvas></div>`).join('');
  specs.forEach((spec, i) => buildChart(document.getElementById(`chart-${i}`), spec, runs));
}

function renderAll() {
  renderHeader();
  renderRuns();
  if (state.tab === 'metrics') renderMetrics();
}

function showTab(tab) {
  state.tab = tab;
  if (tab === 'runs') history.replaceState(null, '', location.pathname + location.search);
  for (const button of document.querySelectorAll('.tab')) button.classList.toggle('active', button.dataset.tab === tab);
  for (const content of document.querySelectorAll('.tab-content')) content.classList.toggle('hidden', content.dataset.content !== tab);
  if (tab === 'metrics') renderMetrics();
}

async function fetchJSON(url) {
  const response = await fetch(url, { cache: 'no-cache' });
  if (!response.ok) throw new Error(`${url}: HTTP ${response.status}`);
  return response.json();
}

async function load() {
  const manifest = await fetchJSON(`${DATA_BASE}/manifest.json`);
  const files = Array.isArray(manifest.runs) ? manifest.runs : [];
  const loaded = await Promise.all(files.map((file) =>
    fetchJSON(`${DATA_BASE}/${file}`).then((doc) => ({ file, doc })).catch(() => ({ file, error: true }))));
  state.runs = loaded.filter((run) => run.doc && run.doc.metadata && run.doc.sections)
    .sort((a, b) => new Date(a.doc.metadata.date) - new Date(b.doc.metadata.date));
  state.unreadable = loaded.filter((run) => run.error).map((run) => run.file);
  const params = new URLSearchParams(location.search);
  const requested = params.get('cluster');
  if (params.get('run')) state.expanded.add(params.get('run'));
  const types = clusterTypes();
  const latest = state.runs[state.runs.length - 1];
  state.clusterType = types.includes(requested) ? requested : (latest ? latest.doc.metadata.cluster_type : '');
}

async function init() {
  for (const button of document.querySelectorAll('.tab')) button.addEventListener('click', () => showTab(button.dataset.tab));
  document.getElementById('cluster-type').addEventListener('change', (event) => { state.clusterType = event.target.value; state.expanded.clear(); renderAll(); });
  document.getElementById('filter').addEventListener('input', (event) => { state.filter = event.target.value; renderMetrics(); });
  try {
    await load();
    renderAll();
    const [, hashTab, hashCategory] = location.hash.match(/^#(metrics)(?:\/([\w-]+))?$/) || [];
    if (hashTab) {
      if (hashCategory) state.category = hashCategory;
      showTab('metrics');
    }
  } catch (error) {
    const notice = document.getElementById('notice');
    notice.textContent = `Could not load the results: ${error.message}`;
    notice.classList.remove('hidden');
    document.getElementById('subtitle').textContent = 'results unavailable';
  }
}

init();
