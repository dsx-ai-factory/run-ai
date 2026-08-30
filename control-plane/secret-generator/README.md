# Secret Generator

Secret Generator is a control-plane job (owned by the **Griffin** team) that generates and
manages the Kubernetes Secrets used by the Run:AI control-plane. It runs as a Helm
pre-install/pre-upgrade hook: instead of shipping hardcoded default credentials in the
control-plane Helm chart, it creates strong random secrets at deploy time, migrates any
existing Helm-managed credentials into generator-managed Secrets, and honours externally
managed (`existingSecret`) values. It is a short-lived batch job, not a long-running server —
it reads a JSON list of secret specifications, reconciles the corresponding Secrets in the
release namespace, and exits.

## Contents

- [Overview](#overview)
- [Repository layout](#repository-layout)
- [Architecture](#architecture)
- [Key types](#key-types)
- [Secret naming convention](#secret-naming-convention)
- [Migration strategy](#migration-strategy)
- [Building and running](#building-and-running)
- [Testing](#testing)
- [Configuration](#configuration)
- [Security considerations](#security-considerations)
- [Troubleshooting](#troubleshooting)

## Overview

The Secret Generator solves the problem of hardcoded default credentials in the control-plane
Helm chart by:

1. **Eliminating hardcoded defaults** — no more `password` or `admin` as default values.
2. **Preserving existing credentials** — migrates data from Helm-managed secrets to
   generator-managed secrets.
3. **Supporting existingSecret patterns** — respects externally managed secrets.
4. **Using a Kubernetes-native lifecycle** — generated secrets are standard Kubernetes Secrets,
   labeled for identification (`app.kubernetes.io/managed-by: secret-generator`,
   `app.kubernetes.io/part-of: runai-backend`).

The job connects to the cluster with the in-cluster ServiceAccount, loads its list of secret
specifications from a JSON file (mounted from a ConfigMap), and processes each spec with a
5-minute overall timeout, retrying individual API operations up to 3 times.

## Repository layout

```text
.
├── cmd/                     # Service entry point (main.go) — wires config + k8s client, runs the job
├── config/                  # Configuration struct, defaults, and loading (config.go)
├── pkg/
│   └── generator/           # Core logic
│       ├── generator.go     #   Secret reconciliation loop (ProcessSecretSpec, create/reconcile)
│       ├── key_resolver.go  #   Resolver chain (static / copyFrom / generate)
│       ├── random.go        #   Cryptographically-strong value generation per key type
│       ├── nats.go          #   NATS-specific migration handler
│       ├── migration.go     #   Default legacy-secret migration handler
│       └── types.go         #   SecretSpec / SecretKeySpec / GeneratorConfig models
├── bin/                     # Compiled binary output (secret-generator-amd64)
├── Dockerfile
├── Makefile
└── go.mod
```

There is no `swagger/`, `internal/`, or `pkg/client/` — this service exposes no HTTP API and
publishes no Go client.

## Architecture

### Components

1. **Generator Configuration ConfigMap** (ephemeral)
   - Created as a Helm pre-install/pre-upgrade hook (weight -10).
   - Passes secret specifications from Helm values to the generator job as JSON.
   - Deleted after successful job completion
     (`hook-delete-policy: before-hook-creation,hook-succeeded`).

2. **Secret Generator Job**
   - Runs as a pre-install/pre-upgrade hook (weight 0).
   - Reads configuration and creates/migrates secrets.
   - Labels created secrets for identification (see above).

3. **RBAC** (ServiceAccount, Role, RoleBinding)
   - Created as a Helm pre-install/pre-upgrade hook (weight -5).
   - Grants the job permission to read ConfigMaps and create/read Secrets in the release
     namespace.

### Flow

```
helm install/upgrade
        │
        ▼
┌───────────────────────────────┐
│ Weight -10: Create config CM  │
│ (secret specs as JSON)        │
└───────────────────────────────┘
        │
        ▼
┌───────────────────────────────┐
│ Weight -5: Create RBAC         │
│ (SA, Role, RoleBinding)        │
└───────────────────────────────┘
        │
        ▼
┌───────────────────────────────────────────────────┐
│ Weight 0: Run Secret Generator Job                 │
│                                                    │
│  For each secret spec:                             │
│  1. If existingSecret → validate and skip          │
│  2. Extract legacy data (if legacy secret exists)  │
│  3. If generator-managed exists → reconcile        │
│     (seed legacy data for missing keys)            │
│  4. Otherwise → create new (seed legacy data)      │
└───────────────────────────────────────────────────┘
        │
        ▼
   Regular chart deployment
```

## Key types

Each key in a secret spec has a `type` that determines how its value is generated:

| Type           | Character Set                    | Use Case |
|----------------|----------------------------------|----------|
| `password`     | `a-z A-Z 0-9` + `!@#%^*()-_=+`   | Database passwords, admin credentials — anywhere the consumer handles special characters safely (env vars, secret mounts). |
| `alphanumeric` | `a-z A-Z 0-9`                    | Tokens and credentials used where special characters cause problems: NATS auth tokens (embedded in config files), Redis AUTH passwords, values used in URLs/connection strings. |
| `hex`          | `0-9 a-f`                        | Hex-encoded random bytes. |
| `uuid`         | UUID v4 format                   | Client IDs or unique identifiers. |

If no type is specified, `password` is used as the default.

In addition to generated types, keys support:
- **`static`** — a fixed value (e.g., a username that never changes).
- **`copyFrom`** — copies the value of another key in the same secret (e.g., `pubsub-password`
  copies from `AUTH_TOKEN`).
- **`chartValue`** — an explicit value passed from Helm chart values.

### Key resolution

Key resolution happens in two phases:

1. **Legacy migration** — migration handlers (`DefaultMigrationHandler`, `NatsMigrationHandler`)
   extract data from legacy secrets and pre-load it into the data map. Keys with legacy values
   are never passed to the resolver chain — legacy always wins.

2. **Resolver chain** — for keys not found in legacy data, a chain of resolvers is tried in
   order:

| Resolver Type | Chain | Description |
|---|---|---|
| `static` | `ChartValueResolver{AllowEmpty: true}` | Fixed values (e.g., usernames). Accepts empty strings as valid. |
| `copyFrom` | `ChartValueResolver` → `CopyFromOnlyResolver` | ChartValue override checked first, then copies from source key. |
| `generate` | `ChartValueResolver` → `generateOnlyResolver` | ChartValue override checked first, then generates a random value. Blocked for `externalOnly` keys. |

`ChartValueResolver` uses an `AllowEmpty` flag: when false (default), empty strings are skipped
so downstream resolvers get a chance to run; when true (used for `static` type), empty strings
are accepted as valid values.

## Secret naming convention

Generator-managed secrets follow this naming pattern:
- `runai-backend-gen-{component}`
- Example: `runai-backend-gen-postgresql`

This naming convention:
- Distinguishes generator-managed secrets from legacy Helm-managed ones.
- Enables smooth migration without conflicts.
- Allows easy identification of the secret management method.

## Migration strategy

### Fresh install

1. Generator creates new secrets with random values.
2. Secrets are created as standard Kubernetes Secret resources with labels.

### Upgrade with existing Helm-managed secrets

1. Generator detects existing secrets (e.g., `runai-backend-postgresql`).
2. Extracts legacy data via the appropriate migration handler.
3. Creates or reconciles generator-managed secrets (e.g., `runai-backend-gen-postgresql`)
   seeded with legacy values.
4. Helm removes old secrets (no longer in templates).

### Multiple specs targeting the same secret

Several specs can target the same generated secret (e.g., `runai-backend-gen-cache-pubsub` is
shared by NATS, pubsub-auth, cache, and additional-subscriber specs). Each spec may reference a
different legacy secret. The first spec creates the generated secret; subsequent specs reconcile
by adding their missing keys. Legacy data is extracted before reconciliation so migrated values
are preserved rather than regenerated.

### Preserving credentials

The migration process:
- **Never changes passwords/tokens** — legacy data is copied 1:1.
- **Never overwrites existing generator-managed keys** — only missing keys are added.
- **Legacy data takes highest priority** — it is pre-loaded before the resolver chain runs, so
  keys with legacy values bypass ChartValue and generation entirely.
- **Respects existingSecret configurations** — specs with `existingSecretName` are skipped
  independently.

## Building and running

The project is built with Go (see [`go.mod`](go.mod)) and orchestrated through the
[`Makefile`](Makefile), which includes the shared control-plane build infrastructure
(`../../build/makefile/control-plane.mk` and `index.mk`, which pull in `golang.mk` and
`base.mk`). The service compiles inside a Go builder container, and the runtime image is built
`FROM scratch`.

| Target | Description |
|--------|-------------|
| `make build` | `build-go` — compile the binary for amd64 **and** arm64 |
| `make build-go` | `build-go-amd` + `build-go-arm` |
| `make build-go-amd` | Compile `./cmd/main.go` for `GOOS=linux GOARCH=amd64` (FIPS `GOFIPS140=v1.0.0`) inside the Go builder container → `bin/secret-generator-amd64` |
| `make build-go-arm` | Same for `arm64` → `bin/secret-generator-arm64` |
| `make lint` | `lint-go` — run golangci-lint in the linter container |
| `make purify` | `purify-go` — `go fmt` + `go vet` + `go mod tidy` |
| `make validate` | No-op for this service (kept for a uniform target set across control-plane services) |
| `make docker-build-generic` | Build the container image via `docker buildx` |
| `make docker-push-generic` | Push the container image to the registry |
| `make local-deploy` | `build-go` (amd64) → `docker-build-generic` → `docker-push-generic` |
| `make help` | Print the annotated target list |

The [`Dockerfile`](Dockerfile) is a two-stage build:

- **`base`** — `registry.access.redhat.com/ubi9/ubi-minimal`; supplies `/etc/passwd`,
  `/etc/group`, TLS certs (`/etc/ssl`, `/etc/pki`), `/etc/nsswitch.conf`, and prepares a
  writable `/tmp` owned by `65532:0`.
- **final** — `FROM scratch`; copies the identity/cert files and the compiled
  `bin/secret-generator-$TARGETARCH` binary plus the `NOTICE` file, runs as user
  `65532:65532`, and sets `ENTRYPOINT ["./secret-generator"]`.

The build arg `TARGETARCH` selects which architecture binary is copied in:

```bash
docker build --build-arg TARGETARCH=amd64 -t secret-generator:dev .
```

### Local build (local-dev.sh)

`secret-generator` is a control-plane Go service (listed in `CP_GO_SERVICES` in
[`local-dev.sh`](../../local-dev.sh)), so `local-dev.sh` builds it in two steps:

1. Compile the amd64 binary in the Go builder container:
   ```bash
   make build-go-amd OS=linux ARCH=amd64 DOCKERHUB_PROXY= GITHUB_TOKEN=
   ```
   (`DOCKERHUB_PROXY=` pulls the Go builder image straight from Docker Hub; `GITHUB_TOKEN=` is
   emptied in the make call — the private-module token is applied elsewhere.)
2. Build + load the image:
   ```bash
   docker buildx build --platform linux/amd64 --provenance=false --load \
     --build-arg TARGETARCH=amd64 -t runai-local/secret-generator:0.0.0 .
   ```

The image ref is `runai-local/secret-generator:0.0.0`. The tag **`0.0.0`** is the control-plane
chart appVersion — any other tag breaks the local deploy.

## Testing

| Target | Description |
|--------|-------------|
| `make test` | `test-ginkgo` — run the Ginkgo unit-test suites |
| `make test-go` | Run the plain `go test ./...` suite in the builder container |

The tests live alongside the source under `pkg/generator/` (`generator_test.go`, `nats_test.go`,
`random_test.go`, `key_resolver_test.go`, `types_test.go`) and are pure unit tests — they use a
fake Kubernetes client and need no backing services or docker-compose stack.

## Configuration

The job is configured through a small set of keys with sensible built-in defaults (see
[`config/config.go`](config/config.go)). Any key can be overridden by its uppercase environment
variable equivalent, which is how the Helm chart injects values.

| Variable | Config key | Description | Default |
|----------|------------|-------------|---------|
| `NAMESPACE` | `namespace` | Release namespace the job operates in | `runai-backend` |
| `CONFIG_FILE_PATH` | `config_file_path` | Path to the mounted JSON file listing secret specs | `/etc/secret-generator/secrets.json` |
| `FULLNAME_OVERRIDE` | `fullname_override` | Release fullname prefix used for secret names | `runai-backend` |
| `LOG_LEVEL` | `log_level` | Log verbosity (`Debug`, `Info`, …) | `Info` |
| `DRY_RUN` | `dry_run` | When `true`, log intended changes without applying them | `false` |

### Secret specifications file

`CONFIG_FILE_PATH` points at a JSON document listing the secrets to reconcile. Its structure
mirrors the Go models in [`pkg/generator/types.go`](pkg/generator/types.go):

```json
{
  "secrets": [
    {
      "name": "postgresql",
      "legacySecretName": "runai-backend-postgresql",
      "keys": [
        { "name": "postgres-password", "type": "password", "length": 24, "legacyKeyName": "postgres-password" },
        { "name": "username", "type": "static", "chartValue": "runai" },
        { "name": "pubsub-password", "copyFrom": "AUTH_TOKEN" }
      ]
    }
  ]
}
```

- `existingSecretName` on a spec makes the generator validate-and-skip (external management).
- `externalOnly` on a key forbids random generation (it must come from chart values / legacy).

### Helm configuration

Under `secretGenerator` in the chart's `values.yaml`:

```yaml
secretGenerator:
  name: secret-generator
  image:
    repository: secret-generator
    pullPolicy: Always
  config:
    logLevel: Info
    configFilePath: /etc/secret-generator/secrets.json
    dryRun: false  # Set to true to test without making changes
```

To use an externally managed secret instead of auto-generation:

```yaml
global:
  postgresql:
    auth:
      existingSecret: "my-external-postgresql-secret"
      secretKeys:
        userPasswordKey: "password"
        adminPasswordKey: "postgres-password"
```

## Security considerations

1. **Cryptographically strong randomness** — uses `crypto/rand` for all secret generation.
2. **No secret logging** — secret values are never logged or exposed.
3. **Least-privilege RBAC** — only namespace-scoped permissions:
   - ConfigMaps: `get`, `list`
   - Secrets: `get`, `list`, `create`
4. **Non-root container** — runs as user `65532:65532` (see [`Dockerfile`](Dockerfile)).
5. **FIPS build** — compiled with `GOFIPS140=v1.0.0` and imports the shared FIPS package.

## Troubleshooting

### Job failed

Check the job logs:
```bash
kubectl logs job/runai-backend-secret-generator-<revision> -n runai-backend
```

### Secrets not created

1. Check the generator config:
   ```bash
   kubectl get configmap runai-backend-secret-generator-config -n runai-backend -o yaml
   ```

2. Verify RBAC (adjust the SA name if your release name differs):
   ```bash
   kubectl auth can-i create secrets \
     --as=system:serviceaccount:runai-backend:runai-backend-secret-generator -n runai-backend
   ```

### DryRun mode

Enable dry-run to test without making changes:
```yaml
secretGenerator:
  config:
    dryRun: true
```

Check logs for `[DRY-RUN]` messages indicating what would be created.
</content>
</invoke>
