# frontend — Build Context

**Type:** container (JavaScript/TypeScript SPA)
**Team:** Viper Team (frontend guild)
**Package:** `runai-frontend` (Vue 3 + Vite + Quasar + Pinia + TypeScript)

## Make targets

Targets come from the service `Makefile` plus the shared include
`../../build/makefile/javascript.mk` (via `control-plane.mk` / `index.mk`).
The Makefile overrides `NODEJS_VERSION=22-bullseye`.

| Target | What it does |
|--------|-------------|
| `build` | `build-js` + `copy-openapi-spec` — build the app then copy the public OpenAPI spec into the build context |
| `build-js` | Runs `npm i --save-dev && npm run build` inside `node:22-bullseye` (this is what local-dev.sh runs). `npm run build` = `vue-tsc --noEmit && vite build` → `dist/` |
| `copy-openapi-spec` | Copies `../infra/openapi/specs/runai-public.yaml` → `dist/openapi/runai-api.yaml` |
| `test` | `test-js` — Vitest unit tests in the Node container |
| `lint` | `lint-js` + `js-checks` — ESLint auto-fix plus TS/ESLint threshold checks |
| `js-checks` | TS type check + ESLint a11y/security thresholds in one container |
| `typecheck-threshold` / `lint-thresholds` | Threshold gates for TS errors / a11y+Options-API counts |
| `validate` / `purify` | No-ops |

## Key build notes

- Generated code: TS API models in `src/swagger-models/generated` come from the
  Run:ai public OpenAPI spec — do not edit generated dirs.
- Runtime base image: multi-stage — `nginx:1.31.2` base + `ubi9/ubi-minimal`
  FIPS provider stage; final image is `FROM scratch` serving `dist/` via nginx.
  Runs as user `1001:0`, exposes port `8080`, CMD `nginx -g "daemon off;"`.
- Build args: `NGINX_IMAGE` (defaults to `nginx:1.31.2`); no `TARGETARCH` arg —
  the image build is `--platform linux/amd64`.
- Build command (what local-dev.sh actually runs): the dedicated
  `build_frontend` function (skipped by `--skip-frontend`), in two steps:
  1. Build the app in the Node container:
     `make build-js DOCKERHUB_PROXY=`
     (`DOCKERHUB_PROXY=` pulls the `node` builder image from Docker Hub. No
     `GITHUB_TOKEN` — pure JS build, no private Go modules.)
  2. Build + load the nginx image: `docker buildx build --platform linux/amd64
     --provenance=false --load -t runai-local/frontend:0.0.0 .`
     Image ref: `runai-local/frontend:0.0.0` (tag `0.0.0` — the CP chart
     appVersion; any other tag breaks the local deploy).
- Host dev: Node ≥ 22 / npm ≥ 10; `npm run dev` for the Vite dev server.

## local-dev.sh iteration

This is a control-plane service:
```bash
./local-dev.sh --skip-build-rc --skip-cluster
```
</content>
