# frontend

The Run:ai control-plane **frontend** is the single-page web application that
users interact with — the administrative UI for the Run:ai control plane. It is
a Vue 3 + TypeScript app built with [Vite](https://vitejs.dev/) and the
[Quasar](https://quasar.dev/) component framework, and it talks to the backend
control-plane services (identity-manager, bff-service, tenants-manager, the
workloads/assets/policy services, …) through the public OpenAPI surface. The
production artifact is a static bundle served by nginx from a `scratch`-based
container image. Owned by the **Viper Team** (frontend guild).

## Contents

- [Overview](#overview)
- [Repository layout](#repository-layout)
- [Building and running](#building-and-running)
- [Testing](#testing)
- [Configuration](#configuration)
- [OpenAPI models](#openapi-models)

## Overview

The service is responsible for:

- **The control-plane admin UI** — projects/departments, workloads, assets
  (environments, compute, data sources), policies, credentials, users and
  access management, cluster and settings screens.
- **Client-side API integration** — typed TypeScript models generated from the
  Run:ai public OpenAPI spec (`src/swagger-models/generated`), consumed by the
  service layer under `src/services`.
- **Static delivery** — the built bundle (`dist/`) is served by nginx inside the
  runtime image; nginx is hardened to run as a non-root user (`1001:0`) from a
  `scratch` base with a FIPS-capable OpenSSL provider.

The app is built on Vue 3 (Composition API), Vite, Quasar, Pinia (state),
Vue Router, and a set of visualisation/editor libraries (ECharts, Monaco,
xterm, Vue Flow).

## Repository layout

```text
.
├── src/
│   ├── App.vue                 # Root component
│   ├── main.ts                 # App bootstrap (Vue, Pinia, router, Quasar)
│   ├── components/             # Reusable UI components
│   ├── views/                  # Route-level pages
│   ├── composables/            # Reusable composition functions
│   ├── stores/                 # Pinia stores
│   ├── services/               # API/service layer
│   ├── router/                 # Vue Router configuration
│   ├── swagger-models/         # OpenAPI-generated TS models (do not edit generated/)
│   ├── plugins/ directives/ mixins/ models/ constants/ utils/
│   └── assets/ common/ table-models/ tests/
├── public/                     # Static public assets
├── scripts/                    # Node scripts (typecheck & eslint thresholds, a11y audit)
├── unit-test/                  # Unit-test setup
├── dist/                       # Build output (bundle served by nginx)
├── nginx.conf                  # Runtime nginx config
├── vite.config.js              # Vite build/dev config
├── Dockerfile                  # nginx-on-scratch runtime image
├── Makefile
└── package.json
```

## Building and running

This is a JavaScript/TypeScript service. The build runs inside a Node Docker
container and is orchestrated through the [`Makefile`](Makefile), which includes
the shared `build/makefile/javascript.mk` and pins the Node image via
`NODEJS_VERSION=22-bullseye`.

| Target | Description |
|--------|-------------|
| `make build` | `build-js` + `copy-openapi-spec` — build the app, then copy the public OpenAPI spec into the Docker build context |
| `make build-js` | Runs `npm i --save-dev && npm run build` inside the `node:22-bullseye` container (this is what local-dev.sh invokes) |
| `make copy-openapi-spec` | Copies `../infra/openapi/specs/runai-public.yaml` into `dist/openapi/` |
| `make test` | `test-js` — Vitest unit tests in the Node container |
| `make lint` | `lint-js` + `js-checks` — ESLint (auto-fix) plus the TS/ESLint threshold checks |
| `make js-checks` | TS type check + ESLint a11y/security thresholds in a single container |
| `make typecheck-threshold` | Fail if TS error count exceeds the threshold |
| `make lint-thresholds` | Fail if ESLint a11y / Options-API counts exceed thresholds |
| `make validate` / `make purify` | No-ops for this service |

`npm run build` = `vue-tsc --noEmit && vite build` (type-check then bundle),
emitting the static site to `dist/`.

### npm scripts (host, without Docker)

Run these directly with Node ≥ 22 / npm ≥ 10 for local iteration:

```bash
npm install
npm run dev              # Vite dev server with hot reload
npm run build            # type-check + production bundle → dist/
npm run test:unit        # Vitest (watch)
npm run test:unit:prod   # Vitest (single run)
npm run lint             # ESLint (auto-fix) + prettier
```

### Runtime image (Dockerfile)

The multi-stage Dockerfile produces a hardened static-serving image:

- **`fips-builder`** (`ubi9/ubi-minimal`) supplies the NIST-validated OpenSSL
  `fips.so` provider.
- **`base`** (`nginx:1.31.2`) creates the `nonroot` (`1001:1001`) identity,
  pre-creates all writable nginx temp/log/pid directories, and stages the
  arch-specific shared libraries + dynamic linker under `/dynlib`.
- **`prep`** copies `nginx.conf` and the built `dist/` into the nginx html root.
- **`runner`** (`FROM scratch`) assembles the final image from the stages above,
  runs as `1001:0`, exposes port **8080**, and starts nginx (`nginx -g "daemon off;"`).

### Local build (local-dev.sh)

`local-dev.sh` builds the frontend through its dedicated `build_frontend`
function (skipped by `--skip-frontend`). It runs:

```bash
# Step 1 — build the app inside the Node container:
make build-js DOCKERHUB_PROXY=
# Step 2 — build + load the nginx image:
docker buildx build --platform linux/amd64 --provenance=false --load \
  -t runai-local/frontend:0.0.0 .
```

`DOCKERHUB_PROXY=` (empty) pulls the `node` builder image from Docker Hub. No
`GITHUB_TOKEN` is involved — this is a pure JS build with no private Go modules.
The image tag is **`0.0.0`** (the control-plane chart appVersion); any other tag
breaks the local deploy. Image ref: `runai-local/frontend:0.0.0`.

## Testing

| Target | Description |
|--------|-------------|
| `make test` (`test-js`) | Vitest unit tests (`npm run test`) in the Node container |
| `npm run test:unit` | Vitest in watch mode (jsdom) |
| `npm run test:unit:prod` | Vitest single run |

## Configuration

The frontend is a static bundle; there are no server-side environment variables
to configure at runtime. The local Vite server accepts `RUN_AI_BACKEND_URL`,
`RUN_AI_FRONTEND_PORT`, and `RUN_AI_FRONTEND_HOST` for backend proxying.
Deployment-resolved, browser-safe runtime settings are fetched by the app at
boot from the BFF service's unauthenticated `GET /api/v1/boot-config` endpoint
rather than baked into the bundle.

## OpenAPI models

TypeScript API models under `src/swagger-models/generated` are generated from
the Run:ai public OpenAPI spec (`OPENAPI_PUBLIC_SPEC_PATH`, i.e.
`infra/openapi/specs/runai-public.yaml`). The generated directory must not be
edited by hand — regenerate it from the spec instead.
</content>
</invoke>
