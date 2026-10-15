# FlowMark dashboard

A static page that shows the nightly FlowMark results: one row per run and a
trend chart per scenario and KPI. Plain HTML, CSS and JavaScript with Chart.js
from a CDN, no build step. The browser fetches `manifest.json` and every
`results-*.json` it lists and draws everything client-side.

Published at https://dsx-ai-factory.github.io/run-ai/flowmark/ from the
`gh-pages` branch, which holds these three files next to the results
the nightly workflow pushes there.

## Run it locally

```
cd scale-load/flowmark/dashboard
python3 -m http.server 8000
```

then open http://localhost:8000/?data=fixtures to read the real nights kept
under `fixtures/`, or `?data=https://dsx-ai-factory.github.io/run-ai/flowmark` to read
the published results. Without `?data=` the page reads the folder it is served
from, which is how it runs on Pages.

## What the page shows

- **Test Runs**: date, control plane build, a download link for the night's
  results file, scenario count, the regression check, and a badge: green
  passed, red flagged when a regression was detected, red failed when a
  scenario did not complete. A click on a row lists each scenario's key numbers
  and its verdict or error text.
- **Metrics**: one sub-tab per category found in the results, one chart per
  scenario and KPI. A KPI is every p90 plus every metric that has an SLO. The
  dashed line is the SLO from the latest run, left out and named in the legend
  when it is more than three times above the data. A red dot is a metric
  FlowMark flagged, and every point's tooltip carries the control plane build
  it measured.
  The filter box narrows the charts by name.
- The cluster type selector keeps trends of one cluster type apart from
  another; the page opens on the type of the latest run.

Links: `#metrics`, `#metrics/load` and `?run=<results file>` open a tab, a
sub-tab or an expanded row directly; `?cluster=<type>` picks the cluster type.
