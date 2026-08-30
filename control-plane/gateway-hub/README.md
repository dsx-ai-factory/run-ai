# gateway-hub

`gateway-hub` is the authentication and middleware sidecar for the Run:AI control plane. It runs alongside Traefik and handles the "smart" parts of every inbound request through Traefik's [forwardAuth](https://doc.traefik.io/traefik/middlewares/http/forwardauth/) mechanism.

Traefik is a thin edge proxy — it routes requests and runs lightweight built-in middlewares, but delegates all business logic (JWT validation, tenant resolution, redirect validation) to gateway-hub over a plain HTTP call. If gateway-hub returns `2xx`, Traefik forwards the request to the upstream service with any response headers copied over. If gateway-hub returns a non-`2xx`, Traefik stops the request and sends gateway-hub's response body directly to the client.

```
Client ──► Traefik ──────────────────────────────────────────────────► Upstream service
               │    forwardAuth sub-request (per middleware)                   ▲
               └──► gateway-hub                                                │
                       │  200 + X-Runai-* headers ──────────────────────────► │
                       │  401/400 ──────────────────────────────────────────── ✕
```

Owned by the **Sphinx** team.

## Table of contents

- [ForwardAuth plugins](#forwardauth-plugins)
  - [authentication-handler](#1-authentication-handler)
  - [extract-tenant-id](#2-extract-tenant-id)
  - [workload-redirect-validation](#3-workload-redirect-validation)
- [How Traefik wires the plugins](#how-traefik-wires-the-plugins)
- [Downstream header consumption](#downstream-header-consumption)
- [Repository layout](#repository-layout)
- [Configuration](#configuration)
- [Building and running](#building-and-running)
- [Testing](#testing)
- [OpenAPI and code generation](#openapi-and-code-generation)
- [Go client](#go-client)
- [Adding a new forwardAuth plugin](#adding-a-new-forwardauth-plugin)

---

## ForwardAuth plugins

Each plugin is a plain Go HTTP handler mounted at `/internal/plugins/<name>`. Traefik calls it as a forwardAuth sub-request on every matching route.

### 1. `authentication-handler`

**Endpoint:** `GET /internal/plugins/authentication-handler`

Validates the Bearer JWT carried by the client request, then returns the identity context as response headers so Traefik can copy them onto the upstream request.

**Flow:**

1. Extract the `Authorization: Bearer <token>` header (max 16 KiB).
2. Decode the JWT payload without verification to read the `iss` (issuer) claim.
3. Derive the Keycloak realm from the issuer URL (`…/realms/<realm>`).
4. Fetch the realm's RSA public key from Keycloak (cached for 24 h; failures cached for 1 h).
5. Verify the JWT signature (RS256 / RS384 / RS512).
6. Validate standard time claims (`exp`, `nbf`, `iat`).
7. Require non-empty `sub` and `tenant_id` claims.
8. Write identity headers and return `200 OK`.

**Response headers on success:**

| Header | JWT claim / source | Notes |
|---|---|---|
| `X-Runai-Tenant-Id` | `tenant_id` | Numeric, supports `"123"` and `123` in the token |
| `X-Runai-Subject-Id` | `sub` | User or service-account identifier |
| `X-Runai-Subject-Type` | derived | `User` when `client_id` is empty, `ServiceAccount` otherwise |
| `X-Runai-Groups` | `groups` | Comma-joined list |
| `X-Runai-Supplemental-Groups` | `supplementarygroups` | Comma-joined integer list |
| `X-Runai-GID` | `gid` | |
| `X-Runai-UID` | `uid` | |
| `X-Runai-Session-Id` | `sid` | |
| `X-Runai-Is-Management-Tenant` | `MANAGEMENT_REALM` env | `"true"` when realm equals the management realm |
| `X-Runai-Client-Id` | `client_id` | Set only for `personal-application` service accounts |

Returns `401` on any validation failure.

---

### 2. `extract-tenant-id`

**Endpoint:** `GET /internal/plugins/extract-tenant-id`

Resolves the tenant ID from the subdomain of the original request hostname. Used on routes that identify the tenant by hostname (e.g. `<tenant>.example.com`) rather than by JWT.

**Flow:**

1. Read the original hostname from `X-Forwarded-Host` (Traefik sets this in forwardAuth sub-requests).
2. Extract the first subdomain label (everything before the first `.`).
3. Look up the tenant ID from `tenants-manager` by name or alias.
4. Return `X-Runai-Tenant-Id` and `200 OK`.

When the tenant cannot be resolved the handler still returns `200` (with no header), so the request passes through. This matches the permissive behaviour needed for unauthenticated login flows.

**Response headers on success:**

| Header | Source |
|---|---|
| `X-Runai-Tenant-Id` | `tenants-manager` lookup |

---

### 3. `workload-redirect-validation`

**Endpoint:** `GET /internal/plugins/workload-redirect-validation`

Guards the Keycloak OAuth authorization endpoint for workload auth flows. Prevents open-redirect attacks by checking that `redirect_uri` belongs to the cluster that issued the `client_id`.

**Flow:**

1. Read `redirect_uri` and `client_id` from the query string. Traefik forwardAuth strips query params from the forwarded URL, so the handler reads them from `X-Forwarded-Uri` when present.
2. Parse the cluster UUID from `client_id` — must match `runai-internal.<uuid>.workloads_auth`.
3. Fetch the cluster's registered external URL from `cluster-service`.
4. Validate that `redirect_uri` shares the same scheme, host (or subdomain), and port as the cluster URL, and contains no path traversal (`..`).
5. Return `200 OK` or `400 Bad Request` + an HTML error page.

Because Traefik passes non-`2xx` bodies directly to the client, the `400` response renders the error page in the browser without any additional error-page middleware.

---

## How Traefik wires the plugins

Traefik middlewares are defined in `control-plane/control-plane-chart/templates/traefik/configmap.yaml` (`middlewares.yaml` key). Each plugin maps to a `forwardAuth` middleware:

```yaml
# middlewares.yaml (rendered from Helm)
http:
  middlewares:
    runai-auth:
      forwardAuth:
        address: "http://<gateway-hub-host>:8181/internal/plugins/authentication-handler"
        authResponseHeaders:
          - X-Runai-Tenant-Id
          - X-Runai-Subject-Id
          - X-Runai-Subject-Type
          - X-Runai-Groups
          - X-Runai-Supplemental-Groups
          - X-Runai-GID
          - X-Runai-UID
          - X-Runai-Is-Management-Tenant
          - X-Runai-Client-Id
          - X-Runai-Session-Id

    runai-extract-tenant-id:
      forwardAuth:
        address: "http://<gateway-hub-host>:8181/internal/plugins/extract-tenant-id"
        authResponseHeaders:
          - X-Runai-Tenant-Id

    workload-redirect-validation:
      forwardAuth:
        address: "http://<gateway-hub-host>:8181/internal/plugins/workload-redirect-validation"
```

Routers reference these middlewares by name:

```yaml
# routes.yaml — example: protected API route
identity-manager-router:
  middlewares:
    - runai-auth
    - no-cache

# routes.yaml — example: unauthenticated hostname-scoped route (multi-tenancy)
identity-manager-unauthenticated-hostname-router:
  middlewares:
    - runai-extract-tenant-id
    - no-cache

# routes.yaml — example: Keycloak workload auth
keycloak-workload-router:
  middlewares:
    - keycloak-custom-headers
    - workload-redirect-validation
```

The `authResponseHeaders` list controls which headers Traefik copies from gateway-hub's response onto the upstream request. Headers not in the list are discarded.

---

## Downstream header consumption

Upstream services read the forwarded headers through `runai-common-packages/runaictx`. The package maps each `X-Runai-*` header to a typed context key:

| Header | `runaictx` key |
|---|---|
| `X-Runai-Tenant-Id` | `runaictx.CfTenantId` |
| `X-Runai-Subject-Id` | `runaictx.CfSubjectID` |
| `X-Runai-Subject-Type` | `runaictx.CfSubjectType` |
| `X-Runai-Groups` | `runaictx.CfGroups` |
| `X-Runai-Supplemental-Groups` | `runaictx.CfSupplementalGroups` |
| `X-Runai-GID` | `runaictx.CfGId` |
| `X-Runai-UID` | `runaictx.CfUId` |
| `X-Runai-Session-Id` | `runaictx.CfSId` |
| `X-Runai-Is-Management-Tenant` | `runaictx.CfIsManagementTenant` |
| `X-Runai-Client-Id` | `runaictx.CfClientId` |

Services use these keys for authorization decisions, tenant scoping, audit logging, and structured log enrichment.

---

## Repository layout

```text
.
├── main.go                     # Service entry point (root package; BUILD_IN_PATH=.)
├── internal/
│   ├── auth/                   # authentication-handler plugin: JWT validation, Keycloak key fetch, identity headers
│   ├── tenant/                 # extract-tenant-id plugin: subdomain → tenant-id via tenants-manager
│   ├── redirect/               # workload-redirect-validation plugin: redirect_uri guard via cluster-service
│   │   └── assets/             # HTML error page served on a blocked redirect
│   └── generated/              # OpenAPI-generated server stubs (do not edit)
├── pkg/
│   ├── config/                 # Config struct + defaults (registered with the shared config loader)
│   └── client/                 # Versioned Go client (module of its own; see pkg/client/generated)
│       └── generated/          # OpenAPI-generated client (do not edit)
├── swagger/                    # OpenAPI source and built bundle
│   ├── source/                 # openapi.yaml + resources/, schemas/, parameters/
│   └── _build/                 # Generated bundle (swagger.yaml/json/html)
├── Dockerfile
├── Makefile
└── go.mod
```

All three plugins are plain Go HTTP handlers wired together in [`main.go`](main.go) — there is no `cmd/` directory; the Makefile sets `BUILD_IN_PATH=.` so the module root is the build entry point.

## Configuration

gateway-hub loads its settings from a YAML config file resolved at startup, with **defaults defined in code** ([`pkg/config/types.go`](pkg/config/types.go)) and per-key environment-variable overrides. The Helm chart mounts the config file and injects overrides through the `<release>-gateway-hub` ConfigMap. The service-specific keys are:

```yaml
# gateway-hub config (keys from pkg/config/types.go; defaults shown)
keycloak_url: http://keycloak-http                       # Keycloak base URL
management_realm: ""                                      # Keycloak realm for the management tenant (multi-tenancy only)
tenants_manager_url: http://tenants-manager:8080         # Internal URL of the tenants-manager service
cluster_service_url: http://runai-backend-cluster-service:8080  # Internal URL of the cluster-service

# HTTP server keys (shared runaihttp config; default service_port is 8080, set to 8181 in deploy)
service_port: 8181                                       # Port to listen on
```

The shared logging, metrics and tracing components register their own config sections into the same file. **Any config key can be overridden by its uppercase environment-variable equivalent** (nested keys join with `_`), which is how the chart supplies most values:

| Variable | Config key | Description |
|---|---|---|
| `SERVICE_PORT` | `service_port` | Port to listen on (deployed as `8181`, matching the Dockerfile `EXPOSE`) |
| `KEYCLOAK_URL` | `keycloak_url` | Keycloak base URL (e.g. `https://host/auth`) |
| `MANAGEMENT_REALM` | `management_realm` | Keycloak realm for the management tenant; when set, tokens from this realm get `X-Runai-Is-Management-Tenant: true`. Only set when multi-tenancy is enabled. |
| `TENANTS_MANAGER_URL` | `tenants_manager_url` | Internal URL of the `tenants-manager` service |
| `CLUSTER_SERVICE_URL` | `cluster_service_url` | Internal URL of the `cluster-service` |

Bootstrap environment variables select *which* file to load rather than individual values:

| Variable | Description |
|---|---|
| `CONFIG_FILE_PATH` | Path to the YAML config file to load |
| `CONFIG_SECRETS_FILE_PATH` | Path to a separate file for secret-tagged fields |
| `CONFIG_KEY_DELIMITER` | Override the nesting delimiter used for env-var key mapping |

---

## Building and running

The project is built with Go (see [`go.mod`](go.mod) for the required version) and orchestrated through the [`Makefile`](Makefile), which includes the shared `build/makefile/control-plane.mk` and `index.mk`. The build entry point is the module root (`BUILD_IN_PATH=.`) and the binary is emitted to `bin/gateway-hub-<arch>`.

| Target | Description |
|---|---|
| `make build` | `openapi-build` then `build-go` — build the OpenAPI bundle + generate code, then compile both arch binaries |
| `make build-go` | `build-go-amd` + `build-go-arm` — both architecture binaries |
| `make build-go-amd` | Compile `.` for `GOOS=linux GOARCH=amd64` (FIPS `GOFIPS140=v1.0.0`) in the Go builder container → `bin/gateway-hub-amd64` |
| `make build-go-arm` | Same for `arm64` → `bin/gateway-hub-arm64` |
| `make lint` | `lint-go` — run golangci-lint in the builder container |
| `make purify` | `purify-go` — `go fmt` + `go vet` + `go mod tidy` |
| `make validate` | `openapi-validate` — validate the OpenAPI spec and generated code are up to date |
| `make openapi-build` | Bundle the OpenAPI source and regenerate the server stubs + Go client |
| `make openapi-lint` | Lint the built OpenAPI bundle |
| `make docker-build-generic` / `make docker-push-generic` | Build / push the container image (used by CI) |
| `make client-bump-version` / `make client-bump-version-dev` | Tag a new (or dev) version of the `pkg/client` Go client |

> gateway-hub has no frontend, so the shared JS/TS client-generation targets are stubbed out to no-ops in the Makefile.

### Local build (local-dev.sh)

gateway-hub is a control-plane Go service (listed in `CP_GO_SERVICES`), so `local-dev.sh` builds it in two steps:

1. Compile the amd64 binary in the builder container:
   ```bash
   make build-go-amd OS=linux ARCH=amd64 DOCKERHUB_PROXY= GITHUB_TOKEN=
   ```
   (`DOCKERHUB_PROXY=` pulls the Go builder image straight from Docker Hub; `GITHUB_TOKEN=` is emptied in the make call — the private-module token is applied elsewhere.)
2. Build + load the image:
   ```bash
   docker buildx build --platform linux/amd64 --provenance=false --load \
     --build-arg TARGETARCH=amd64 -t runai-local/gateway-hub:0.0.0 .
   ```

The image tag **must** be `0.0.0` (the control-plane chart `appVersion`) — any other tag breaks the local deploy. To rebuild only the control plane and redeploy against an existing cluster:

```bash
./local-dev.sh --skip-build-rc --skip-cluster
```

### Dockerfile

The Dockerfile is a two-stage `scratch` image:

1. **`base`** — pulls `registry.access.redhat.com/ubi9/ubi-minimal` to supply `/etc/passwd`, `/etc/group`, OS certificates (`/etc/ssl`, `/etc/pki`), `nsswitch.conf`, and a writable `/tmp` owned by `65532:0`.
2. **`scratch`** — copies the pre-built binary from `bin/gateway-hub-$TARGETARCH` (build arg `TARGETARCH`) plus the OS files, runs as user `65532:65532`, and exposes port `8181`.

```bash
docker build --build-arg TARGETARCH=amd64 -t gateway-hub:dev .
```

The image is distributed through the Helm chart only — it is not published to a plugin catalog and is not installed into Traefik as a plugin binary.

---

## Testing

| Target | Description |
|---|---|
| `make test` | `test-go` — build and run the Go unit tests in the builder container |

Each plugin package ships table-driven unit tests (`internal/auth`, `internal/tenant`, `internal/redirect`) built with `net/http/httptest`; there is no separate docker-compose integration suite for this service.

---

## OpenAPI and code generation

The HTTP API is defined in [`swagger/source/openapi.yaml`](swagger/source/openapi.yaml), split into per-resource files under `swagger/source/resources`, shared `swagger/source/schemas`, and `swagger/source/parameters`. From this source `make openapi-build` produces:

- The **server stubs** in [`internal/generated`](internal/generated) — consumed by the chi router.
- The **Go client** under [`pkg/client/generated`](pkg/client/generated).
- The bundled spec in `swagger/_build/` (`swagger.yaml`, `swagger.json`, `swagger.html`).

Generation uses [OpenAPI Generator](https://openapi-generator.tech) `v7.12.0` with `VERSIONED_CLIENT=true` (pinned in the `Makefile`). Both `generated/` directories **must not be edited by hand** — changes are overwritten on the next generation run. `make validate` checks the committed bundle and generated code match the source.

---

## Go client

gateway-hub ships a versioned Go client at [`pkg/client`](pkg/client) (its own Go module) that other services can import:

```go
import openapi "github.com/run-ai/runai/control-plane/gateway-hub/pkg/client"
```

The client is regenerated from the OpenAPI spec (see above) and versioned via `make client-bump-version` (release) / `make client-bump-version-dev` (development tag). See [`pkg/client/generated/README.md`](pkg/client/generated/README.md) for the generated reference.

---

## Adding a new forwardAuth plugin

Follow these steps to add a fourth plugin handler to gateway-hub and wire it into Traefik.

### 1. Implement the handler

Create a new package under `internal/`:

```
internal/
  myplugin/
    handler.go
    handler_test.go
```

`handler.go` must expose an HTTP handler that:

- Returns `200 OK` (with optional response headers) when the request should proceed.
- Returns a non-`2xx` status to block the request. Traefik forwards the body directly to the client, so include a meaningful response body when blocking.

```go
package myplugin

import "net/http"

type Handler struct{ /* dependencies */ }

func NewHandler( /* deps */ ) *Handler { return &Handler{} }

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    // Read X-Forwarded-* headers set by Traefik forwardAuth:
    //   X-Forwarded-Host   — original hostname
    //   X-Forwarded-Uri    — original URI (path + query)
    //   X-Forwarded-Proto  — original scheme
    //   X-Forwarded-For    — original client IP
    //   Authorization      — Bearer token (passed through by Traefik)

    // Set response headers that Traefik should copy onto the upstream request:
    w.Header().Set("X-Runai-My-Header", "value")
    w.WriteHeader(http.StatusOK)
}
```

### 2. Register the route in `main.go`

```go
myHandler := myplugin.NewHandler( /* deps from cfg */ )

r.Get("/internal/plugins/my-plugin", myHandler.ServeHTTP)
```

### 3. Add OpenAPI documentation

Create `swagger/source/resources/my_plugin.yaml` following the pattern of the existing resource files, then reference it from `swagger/source/openapi.yaml`:

```yaml
/internal/plugins/my-plugin:
  x-internal: true
  $ref: './resources/my_plugin.yaml'
```

Regenerate the OpenAPI bundle:

```bash
make openapi-build
```

### 4. Add the forwardAuth middleware to Traefik

In `control-plane/control-plane-chart/templates/traefik/configmap.yaml`, add a new entry under `http.middlewares` in the `middlewares.yaml` key:

```yaml
my-plugin:
  forwardAuth:
    address: {{ printf "http://%s:%v/internal/plugins/my-plugin" $top.Values.traefik.gatewayHub.host $top.Values.traefik.gatewayHub.port | quote }}
    authResponseHeaders:
      - X-Runai-My-Header   # only headers listed here are copied to the upstream
```

### 5. Apply the middleware to a router

Reference the middleware name in the relevant router definition in the same configmap:

```yaml
my-service-router:
  rule: "PathPrefix(`/api/v1/my-service`)"
  entryPoints:
  - web
  service: my-service
  middlewares:
  - my-plugin
  - no-cache
  priority: 3
```

### 6. Add unit tests

Write tests for the handler in `internal/myplugin/handler_test.go`, using `net/http/httptest`. See `internal/auth/handler_test.go` and `internal/tenant/handler_test.go` for examples.

```bash
make test
```
