# Run.ai Monorepo — Developer Guide

## What this repo is
Run.ai OSS monorepo. Full stack: control plane services, runai-cluster (Helm chart + controllers), shared packages, and the local development environment.

## Documentation

Documentation rules — when `docs/` gets touched, writing style, terminology,
and sources of truth — live in `AGENTS.md` and `docs/AGENTS.md`. Read
`AGENTS.md` before any change that might affect `docs/`.

@AGENTS.md

## local-dev.sh — spinning up a local environment

```bash
export GITHUB_TOKEN=ghp_...   # needed for private Go modules
./local-dev.sh                # full build + cluster + deploy (~30 min first run)
```

**Common flag combinations:**
```bash
./local-dev.sh --skip-build --skip-cluster        # re-deploy only (fastest)
./local-dev.sh --skip-frontend                    # skip slow JS build
./local-dev.sh --skip-build-rc --skip-cluster     # rebuild CP only, reuse cluster
./local-dev.sh --skip-build-cp --skip-cluster     # rebuild RC only, reuse cluster
./local-dev.sh --force-rebuild                    # ignore build cache
./local-dev.sh --teardown && ./local-dev.sh       # nuke and restart
./local-dev.sh --gpu-operator                     # real NVIDIA GPUs instead of fake-gpu-operator (Linux)
```

**All flags:** `--domain`, `--admin-email`, `--admin-password`, `--parallel N`, `--skip-build`, `--skip-build-cp`, `--skip-build-rc`, `--skip-frontend`, `--skip-cluster`, `--skip-deploy`, `--skip-deploy-cp`, `--skip-deploy-rc`, `--skip-prometheus`, `--skip-fake-gpu`, `--gpu-operator`, `--gpu-driver`, `--gpu-operator-version`, `--force-rebuild`, `--teardown`, `--purge`

After first run: `sudo sh -c 'echo "127.0.0.1 runai.local" >> /etc/hosts'` → https://runai.local (login `test@run.ai`; password is auto-generated and printed in the final summary unless `--admin-password` is passed)

## Service structure
Each service has `artifact-metadata.json` (`artifact_type`: container | go | cli | helm) and standard Makefile targets: `build`, `test`, `lint`, `validate`.

---

## Services

### ⚠️ Missing documentation (0 services)

All service directories now have a README.md and CLAUDE.md. If you add a new
service, run `/write-service-docs <path>` to document it and add it below.

> Note: the previously-listed `control-plane/ubi` entry was **stale** — no such
> directory exists in the repo (the only UBI base image is
> `runai-cluster/cluster/ubi-minimal`), so it has been dropped from this index.

---

### ✅ Documented services

**control-plane** (container)
- [control-plane/assets-service](control-plane/assets-service)
- [control-plane/audit-service](control-plane/audit-service)
- [control-plane/authorization](control-plane/authorization)
- [control-plane/bff-service](control-plane/bff-service)
- [control-plane/catalog-service](control-plane/catalog-service)
- [control-plane/cli-exposer](control-plane/cli-exposer)
- [control-plane/cluster-migrator](control-plane/cluster-migrator)
- [control-plane/cluster-service](control-plane/cluster-service)
- [control-plane/datavolumes-service](control-plane/datavolumes-service)
- [control-plane/frontend](control-plane/frontend)
- [control-plane/gateway-hub](control-plane/gateway-hub)
- [control-plane/identity-manager](control-plane/identity-manager)
- [control-plane/k8s-objects-tracker](control-plane/k8s-objects-tracker)
- [control-plane/keycloak-extensions](control-plane/keycloak-extensions)
- [control-plane/metric-stores-migrator](control-plane/metric-stores-migrator)
- [control-plane/metrics-service](control-plane/metrics-service)
- [control-plane/notifications-proxy](control-plane/notifications-proxy)
- [control-plane/org-unit-service](control-plane/org-unit-service)
- [control-plane/policy-service](control-plane/policy-service)
- [control-plane/secret-generator](control-plane/secret-generator)
- [control-plane/temp-backend-settings](control-plane/temp-backend-settings)
- [control-plane/tenants-manager](control-plane/tenants-manager)
- [control-plane/workloads](control-plane/workloads)
- [control-plane/workloads-manager](control-plane/workloads-manager)

**control-plane** (helm)
- [control-plane/control-plane-chart](control-plane/control-plane-chart)

**control-plane** (go — libraries & client packages)
- [control-plane/authorization/pkg/client](control-plane/authorization/pkg/client)
- [control-plane/cluster-service/pkg/client](control-plane/cluster-service/pkg/client)
- [control-plane/org-unit-service/pkg/client](control-plane/org-unit-service/pkg/client)
- [control-plane/runai-cli](control-plane/runai-cli)
- [control-plane/workloads-and-policies](control-plane/workloads-and-policies)
- [control-plane/workloads/pkg/client](control-plane/workloads/pkg/client)

**control-plane** (tooling)
- [control-plane/fielder](control-plane/fielder) — Python workload field-spec code generator
- [control-plane/redoc](control-plane/redoc) — OpenAPI spec merge/publish pipeline

**runai-cluster** (container)
- [runai-cluster/cluster/accessrule-controller](runai-cluster/cluster/accessrule-controller)
- [runai-cluster/cluster/admission-controller](runai-cluster/cluster/admission-controller)
- [runai-cluster/cluster/advice-restart-controller](runai-cluster/cluster/advice-restart-controller)
- [runai-cluster/cluster/inference-endpoint-controller](runai-cluster/cluster/inference-endpoint-controller)
- [runai-cluster/cluster/agent](runai-cluster/cluster/agent)
- [runai-cluster/cluster/anyworkload-controller](runai-cluster/cluster/anyworkload-controller)
- [runai-cluster/cluster/application-controller](runai-cluster/cluster/application-controller)
- [runai-cluster/cluster/cluster-api](runai-cluster/cluster/cluster-api)
- [runai-cluster/cluster/cluster-sync](runai-cluster/cluster/cluster-sync)
- [runai-cluster/cluster/external-workload-integrator](runai-cluster/cluster/external-workload-integrator)
- [runai-cluster/cluster/git-sync](runai-cluster/cluster/git-sync)
- [runai-cluster/cluster/go-operator](runai-cluster/cluster/go-operator)
- [runai-cluster/cluster/goofys](runai-cluster/cluster/goofys)
- [runai-cluster/cluster/helm-hooks-handler](runai-cluster/cluster/helm-hooks-handler)
- [runai-cluster/cluster/init-ca](runai-cluster/cluster/init-ca)
- [runai-cluster/cluster/karta-migrate](runai-cluster/cluster/karta-migrate)
- [runai-cluster/cluster/metrics-exporter](runai-cluster/cluster/metrics-exporter)
- [runai-cluster/cluster/runai-crd-upgrader](runai-cluster/cluster/runai-crd-upgrader)
- [runai-cluster/cluster/runtime-rules](runai-cluster/cluster/runtime-rules)
- [runai-cluster/cluster/shared-objects-controller](runai-cluster/cluster/shared-objects-controller)
- [runai-cluster/cluster/status-updater](runai-cluster/cluster/status-updater)
- [runai-cluster/cluster/valkey](runai-cluster/cluster/valkey)
- [runai-cluster/cluster/workload-exporter](runai-cluster/cluster/workload-exporter)

**runai-cluster** (helm)
- [runai-cluster/cluster/runai-chart](runai-cluster/cluster/runai-chart)

**runai-application-packages** (go)
- [runai-application-packages/async-events](runai-application-packages/async-events)
- [runai-application-packages/client-utils](runai-application-packages/client-utils)
- [runai-application-packages/concurrency](runai-application-packages/concurrency)
- [runai-application-packages/convutils](runai-application-packages/convutils)
- [runai-application-packages/datastructs](runai-application-packages/datastructs)
- [runai-application-packages/dbmodels](runai-application-packages/dbmodels)
- [runai-application-packages/jsonutils](runai-application-packages/jsonutils)
- [runai-application-packages/old-auth-token-middleware](runai-application-packages/old-auth-token-middleware)
- [runai-application-packages/patching](runai-application-packages/patching)
- [runai-application-packages/queryapi](runai-application-packages/queryapi)
- [runai-application-packages/scopeinfo](runai-application-packages/scopeinfo)
- [runai-application-packages/telemetry](runai-application-packages/telemetry)
- [runai-application-packages/testutils](runai-application-packages/testutils)
- [runai-application-packages/validation](runai-application-packages/validation)
- [runai-application-packages/versions](runai-application-packages/versions)

**runai-clients** (go — generated API clients)
- [runai-clients/assets-service-client](runai-clients/assets-service-client)
- [runai-clients/catalog-service-client](runai-clients/catalog-service-client)
- [runai-clients/cli-exposer-client](runai-clients/cli-exposer-client)
- [runai-clients/cluster-api-client](runai-clients/cluster-api-client)
- [runai-clients/datavolumes-service-client](runai-clients/datavolumes-service-client)
- [runai-clients/k8s-objects-tracker-client](runai-clients/k8s-objects-tracker-client)
- [runai-clients/policy-service-client](runai-clients/policy-service-client)
- [runai-clients/workloads-manager-client](runai-clients/workloads-manager-client)

**runai-common-packages** (go)
- [runai-common-packages/cache](runai-common-packages/cache)
- [runai-common-packages/config](runai-common-packages/config)
- [runai-common-packages/fips](runai-common-packages/fips)
- [runai-common-packages/integdb](runai-common-packages/integdb)
- [runai-common-packages/logging](runai-common-packages/logging)
- [runai-common-packages/metrics](runai-common-packages/metrics)
- [runai-common-packages/runaictx](runai-common-packages/runaictx)
- [runai-common-packages/runaidb](runai-common-packages/runaidb)
- [runai-common-packages/runaihttp](runai-common-packages/runaihttp)
- [runai-common-packages/runaipubsub](runai-common-packages/runaipubsub)
- [runai-common-packages/s3](runai-common-packages/s3)
- [runai-common-packages/setting](runai-common-packages/setting)
- [runai-common-packages/toggles](runai-common-packages/toggles)
- [runai-common-packages/tracing](runai-common-packages/tracing)

**runai-installer**
- [runai-installer/installer](runai-installer/installer)
- [runai-installer/provisioner-service](runai-installer/provisioner-service)

**runai-mcp-server**
- [runai-mcp-server](runai-mcp-server)

---

## Skills
- `/list-services` — show all services grouped by type
- `/write-service-docs` — generate README + CLAUDE.md for a service
- `/validate-service-docs` — check a service's README against its Makefile
- `/cel-validation-advisor` — assess CEL options before adding a Kubernetes webhook or changing admission validation
