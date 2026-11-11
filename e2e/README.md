# End-to-end tests

Playwright test suite for NVIDIA Run:ai. It exercises the control-plane APIs, the `runai` CLI, and the web UI against a live Kubernetes cluster with Run:ai installed.

This directory is the e2e package in the [dsx-ai-factory/run-ai](https://github.com/dsx-ai-factory/run-ai) monorepo. License, contributing, and community guidelines live at the repository root:

- [LICENSE](../LICENSE)
- [CONTRIBUTING.md](../CONTRIBUTING.md)
- [CODE_OF_CONDUCT.md](../CODE_OF_CONDUCT.md)
- [SECURITY.md](../SECURITY.md)

## What you can run

| Audience | What works |
| --- | --- |
| Anyone (no cluster) | Install dependencies, lint, type-check, and the offline unit tests. |
| Anyone with their own Run:ai | A subset of Playwright tests, if you point the suite at your control plane and kubeconfig. Tests that need private images, NVIDIA-only hosts, or SaaS-only features will skip or fail. |
| NVIDIA CI / nightly | The full parallel and serial suites, including tests tagged `@parallel-ci` / `@serial-ci`. Fork PRs do not get this automatically - see [CONTRIBUTING.md](../CONTRIBUTING.md). |

Private nightly pipelines and external report publishing depend on their own infrastructure and credentials. Contributors can generate local Allure reports and run the shared result aggregation.

## Prerequisites

- Node.js 22
- npm
- [Playwright browsers](https://playwright.dev/docs/browsers) (`npx playwright install`)
- For live tests: `kubectl` access to a cluster with Run:ai, plus control-plane admin credentials

## Setup

```bash
cd e2e
npm install --save-dev
npx playwright install
cp .envExample .env
```

Edit `.env` with **your** environment. Do not commit it. Values in `.envExample` are placeholders, not working credentials.

## Environment variables

Required for any live Playwright run:

| Variable | Purpose |
| --- | --- |
| `CONTROL_PLANE_URL` | Control-plane / tenant URL (for example `https://runai.example.com`) |
| `CONTROL_PLANE_ADMIN_USERNAME` | Admin user used by the suite |
| `CONTROL_PLANE_ADMIN_PASSWORD` | Admin password. No default - must be set. |
| `KUBECONFIG_PATH` or `KUBECONFIG` | Kubeconfig for the target cluster |
| `CLUSTER_NAME_PREFIX` | Cluster name as shown in the Run:ai UI |

Common optional variables:

| Variable | Purpose |
| --- | --- |
| `APP_URL` | UI URL if it differs from `CONTROL_PLANE_URL` |
| `CLUSTER_URL` | Cluster endpoint if it cannot be derived from `CLUSTER_NAME_PREFIX` |
| `KUBECONFIG_CLUSTER_NAME` | Context name when the kubeconfig has more than one cluster |
| `IS_SELF_HOSTED` | `true` for self-hosted, `false` for SaaS |
| `KUBERNETES_PLATFORM` | `k8s` (default) or `ocp` |
| `GPU_TYPE` | `fake` or `real` |
| `LOCAL_RUN` | `true` locally (single worker, less reporting) |
| `HEADLESS_BROWSER` | `true` to hide the browser |
| `LOG_LEVEL` | `info` or `debug` |
| `ENABLE_CONSOLE_LOGS` | `true` to print logs to the console |
| `RETRIES` | Playwright retries (CI default is 2; use `0` locally) |
| `ENABLE_SSO` | Only if the environment uses SSO |
| `KEYCLOAK_ADMIN_PASSWORD` | Self-hosted Keycloak admin password, when a test needs it |

## How to run

From `e2e/`:

```bash
# Offline - no cluster required
npm run test:unit
npm run typecheck
npm run linter
npm run format

# Or via Make (used in CI): lint + type-check
make lint
make test
```

Live Playwright tests (need a configured `.env` and a reachable cluster):

```bash
# Full suite
npx playwright test

# Tests tagged for PR CI
npm run playwright:parallel
npm run playwright:serial

# Filter by title
npx playwright test -g "department"
```

The runner uses the current kubeconfig context unless `KUBECONFIG_PATH` / `KUBECONFIG_CLUSTER_NAME` override it.

## Layout

```
e2e/
├── src/test/suites/playwright/parallel/   # Safe to run concurrently
├── src/test/suites/playwright/serial/     # Must run one after another
├── src/test/unit/                         # Offline unit tests
├── src/modules/                           # API, CLI, UI, and helper libraries
├── playwright.config.ts
└── .envExample
```

## License

Apache License 2.0. See [LICENSE](../LICENSE) at the repository root.

## Source license headers

First-party source files selected for OSS publication carry NVIDIA copyright and Apache-2.0 SPDX headers. The header requirement excludes generated clients and vendored code; existing third-party notices must be preserved.

Only the approved application test framework, suites, selected guides, and kits are published. Private infrastructure, pipeline launchers, and external-system adapters remain outside the public copy.

## API-client development

The committed TypeScript clients are used by the tests. Regeneration is only needed when changing API definitions.

Use the service specifications in the public monorepo, with Docker running:

```bash
# From e2e/ inside the monorepo:
export RUNAI_REPO_PATH=..
npm run api-codegen-assetsApi

# From a separate checkout, use its local path instead:
# export RUNAI_REPO_PATH=/path/to/run-ai
```

The existing generator version and templates are retained. Documentation links must be corrected in the service API source definitions before regeneration. Review generated API/type changes and run type-checks before submitting changes.

Copied service Swagger files are not published. The cleaned, hand-maintained `backendUndocumentedApi/swagger.yaml` is an exception: it defines APIs still used by the test framework and has no corresponding monorepo source.

## Runtime settings and reporting

Set `CLUSTER_URL` to override the cluster endpoint. Otherwise, the URL is built from `CLUSTER_NAME_PREFIX` and the supplied `E2E_BASE_DOMAIN`, using the OpenShift route format when `KUBERNETES_PLATFORM=ocp`. OpenShift SSO uses `OPENSHIFT_API_URL` when set, otherwise it builds the API endpoint from the same cluster name and base domain. The callback URL is based on the configured tenant URL. No private domain is stored in the test code.

All runtime settings come from environment variables. Private CI supplies them from its own Actions variables and secrets; no private defaults file is needed. Shared-environment user cleanup requires `E2E_PRESERVED_USERS` to protect existing accounts.

Air-gapped runs retain the existing Harbor registry calculation using the supplied cluster name, platform, and `E2E_BASE_DOMAIN`. No private domain is stored in that calculation. The configured public demo images are retained. NGC/NIM tests may require separately supplied registry access.

Use `E2E_PRESERVED_USERS` for any additional accounts that shared-environment cleanup must retain. Test-created users use reserved example addresses; the product's default console administrator name is preserved for compatibility.

Local Allure reports and the shared result aggregation do not require Slack credentials. The Slack template has an empty destination; private CI supplies its webhook at runtime. No webhook, CQA credential, or Mailtrap credential belongs in committed configuration. Mailtrap-backed notification tests require environment credentials.

For investigation guidance, see [Debug information collection](docs/DEBUG_INFO_COLLECTION_GUIDE.md).
