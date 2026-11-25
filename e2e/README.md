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

## Contributing E2E changes

### Migration handover

The application tests are in this repository under `e2e/`. During migration, NVIDIA contributors should keep using their team's current contribution process until the developer handover is announced. The PR selection rules below become active at that handover. Environment setup, credentials, and internal execution profiles stay in private CI.

After handover, this repository is the source of truth for migrated tests on `main` and `release-3.0`. Submit matching product and E2E changes in the same PR, against the same target branch.

### Product and test changes in one PR

For example, a workspace API change can include both:

- The product change under `control-plane/workloads/`.
- Its test update in `e2e/src/test/suites/playwright/parallel/viper/submitApi/workspaceApi.test.ts`.

Run the changed service's checks, then run the public E2E checks from `e2e/`:

```bash
npm ci --ignore-scripts
./node_modules/.bin/eslint . --ignore-path .eslintignore --ext .ts --quiet
npm run typecheck
npm run test:unit
```

These checks run without a cluster. Live tests also need a configured environment, as described in [How to run](#how-to-run). Keep credentials out of the PR.

If an API definition changes, use the target branch's service Swagger files and the [API-client development](#api-client-development) commands below. Generate into an empty preview directory and review the complete output, including removed files, before updating committed clients. Do not copy client output from a different suite or generator. CI uses the committed clients; it does not regenerate them for you.

### PR test selection after activation

Private CI builds the selected product commit and runs tests from that same public checkout, including E2E changes in the PR. Test selection follows the changed files:

| Change | Live E2E selection |
| --- | --- |
| Added or edited test files only | Changed parallel and serial tests, within supported environment and profile boundaries. |
| Shared helpers, clients, dependencies, configuration, or deleted tests | The broader parallel and serial PR suites. |
| Product code only | The existing product component selection. |
| Product code and E2E together | Combined product and E2E selection from the same selected commit. |
| A recognized skip-only change | No live environment is needed. Public checks still apply. |
| Removing a skip or re-enabling a test | Live validation is required. |
| Documentation only | No live environment is needed. |

Existing tags, version restrictions, and profile restrictions still apply. A selected file does not guarantee every test inside it runs.

Same-repository PRs trigger CI automatically. Fork PRs can run public checks without private credentials. Live CI requires a reviewer with write access to approve the reviewed commit with `/ok-to-test <commit-sha>`. A new commit needs fresh approval. Follow the existing [CI contribution process](../CONTRIBUTING.md#ci-for-contributors--what-runs-and-when).

A green offline check does not mean live tests ran. Check the live CI result and its executed commit. When live E2E is required, skipped setup or skipped tests do not count as successful validation. Use the report link from the run; CI logs and report access follow the existing access policy.

### Broken tests, skips, and release fixes

For migrated branches, keep the product fix and matching test change in one PR instead of disabling a test, merging the product separately, and then re-enabling the test in another repository.

Preserve intentional `test.skip`, `test.fixme`, and conditional skips for unsupported versions, platforms, or profiles. If a broken test needs a temporary skip, explain the failure, name its owner, and record the next action. Use a public issue when available; do not publish private ticket URLs or internal details. Re-enabling the test needs live validation.

Older release branches that have not migrated still use their private release suite. Coordinate the matching test fix there as well as any public backport. Do not assume a public change updates an older private release.

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
# Check every client mapping without changing files:
npm run api-codegen-check

# Generate a client into a clean preview directory:
npm run api-codegen-assetsApi -- --output-dir /tmp/e2e-clients

# Use a separate matching monorepo checkout if needed:
RUNAI_REPO_PATH=/path/to/run-ai npm run api-codegen-assetsApi -- --output-dir /tmp/e2e-clients-new
```

Generator version 6.2.1 and the existing templates are retained. Specs are read directly from the monorepo and are not copied into `e2e/`. The default runner uses Docker; local tooling can instead set `OPENAPI_GENERATOR_JAR` to the same generator version, with `JAVA_HOME` pointing to its Java runtime.

Generation requires an empty output directory to avoid retaining stale generated files. Review the preview, removed files and compatibility changes before replacing committed clients. `api-codegen-all` validates every generation input before writing output.

Two legacy clients are retained unchanged as temporary regeneration exceptions:

- `backendApi`: existing tests still import its models and enums.
- `backendClusterApi`: existing tests still exercise its legacy cluster, node-pool, job and pod APIs.

Their original generation inputs have no verified replacement in the public monorepo. The all-client commands explicitly report these exceptions and process the other 19 clients. Requesting generation of either retained client directly fails with an explanation. Replacing the legacy dependencies and deciding which legacy endpoints remain covered are follow-up work; retaining these clients does not mean they were regenerated.

Documentation links must be corrected in the service API source definitions before regeneration. Review generated API/type changes and run type-checks before submitting changes.

Copied service Swagger files are not published. The cleaned, hand-maintained `backendUndocumentedApi/swagger.yaml` is an exception: it defines APIs still used by the test framework and has no corresponding monorepo source.

## Runtime settings and reporting

Set `CLUSTER_URL` to override the cluster endpoint. Otherwise, the URL is built from `CLUSTER_NAME_PREFIX` and the supplied `E2E_BASE_DOMAIN`, using the OpenShift route format when `KUBERNETES_PLATFORM=ocp`. OpenShift SSO uses `OPENSHIFT_API_URL` when set, otherwise it builds the API endpoint from the same cluster name and base domain. The callback URL is based on the configured tenant URL. No private domain is stored in the test code.

All runtime settings come from environment variables. Private CI supplies them from its own Actions variables and secrets; no private defaults file is needed. Shared-environment user cleanup requires `E2E_PRESERVED_USERS` to protect existing accounts.

Air-gapped runs retain the existing Harbor registry calculation using the supplied cluster name, platform, and `E2E_BASE_DOMAIN`. No private domain is stored in that calculation. The configured public demo images are retained. NGC/NIM tests may require separately supplied registry access.

Use `E2E_PRESERVED_USERS` for any additional accounts that shared-environment cleanup must retain. Test-created users use reserved example addresses; the product's default console administrator name is preserved for compatibility.

Local Allure reports and the shared result aggregation do not require Slack credentials. The Slack template has an empty destination; private CI supplies its webhook at runtime. No webhook, CQA credential, or Mailtrap credential belongs in committed configuration. Mailtrap-backed notification tests require environment credentials.

For investigation guidance, see [Debug information collection](docs/DEBUG_INFO_COLLECTION_GUIDE.md).
