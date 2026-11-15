# Dashboard fixtures

Real nights, copied unchanged from the report artifacts of the FlowMark
nightly in run-ai/runai-e2e-internal. They are the same files the nightly
pushes to the results store, so the page can be developed and checked against
them without that store.

| file | night | verdict |
|---|---|---|
| results-20261014T015011Z | first nightly, 2026-10-14 | passed, four scenarios including the since-removed deployment-burst |
| results-20261015T015138Z | 2026-10-15 | regression: workspace-ttj pod_created p90 3040 ms over the 3000 ms SLO |

Serve the dashboard folder with `python3 -m http.server` and open it with
`?data=fixtures` to read these instead of the published results. Add a night by
dropping its results file here and listing it in `manifest.json`.
