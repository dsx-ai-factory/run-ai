# Networking Plugin

## Overview

The Networking Plugin provides authentication and routing capabilities for workloads managed by the Any Workload Controller. It creates:

- **OAuth2 Proxy Deployment**: Standalone authentication proxy
- **Auth Proxy Service**: ClusterIP service for the auth proxy
- **Workload Service** (optional): Service for the workload pods
- **Ingress/Route**: External routing with optional TLS

## Why Standalone Auth Proxy?

The Any Workload Controller manages various workload types through external operators (e.g., NIMService, LeaderWorkerSet). Some of these CRDs don't expose a pod template spec for modification, making it impossible to inject sidecar containers. Therefore, authentication is implemented as a standalone OAuth2 proxy Deployment that sits in front of the workload, rather than as a sidecar within the workload pods.

## Input shapes

The plugin accepts the workload's networking configuration via one of two mutually-exclusive surfaces. `classifyInput` resolves the shape from the WorkloadRunner at every entry point — admission (`WebhookValidate`), reconcile (`HandleAdditionalResources`), and status (`Status`):

| Shape | Source on the WorkloadRunner | Handler |
|---|---|---|
| `inputLegacy` | `spec.instructions.networking` (single-endpoint shape, see "Configuration" below) | `handleLegacyAdditionalResources` / `statusLegacy` — the original code paths, behaviour unchanged. |
| `inputAnnotation` | `run.ai/endpoints` annotation (multi-endpoint JSON payload — see [`cluster/common/endpoints`](../../../../common/endpoints/annotation.go)) | `handleEndpointsAdditionalResources` / `statusEndpoints` — per-Endpoint reconcile with folded conditions. |
| `inputNone` | neither set | `Status` reports `PluginPhaseDisabled`; reconcile is a no-op. |

Setting **both** at once is rejected at admission with `field.Forbidden` on the annotation key. If the conflicting state somehow reaches runtime (e.g. admission disabled, or a pre-existing CR), `Status` independently re-runs `classifyInput` and marks the plugin `PluginPhaseFailed` with an `InvalidNetworkingInput` condition. This lets the legacy single-endpoint path keep running unchanged on existing CRs while consumers migrate to the annotation-driven multi-endpoint payload incrementally.

### Annotation mode (multi-endpoint)

`run.ai/endpoints` carries a `Config{ schemaVersion, endpoints[] }`. Each `Endpoint` has a `services[]` slice and an optional `externalAccess` (Ingress / Route / HTTPRoute). Each `Service` is either `create` or `ref`, and may carry an optional `auth` block (same OAuth2-proxy semantics as legacy mode). The plugin reconciles each Endpoint independently — a misconfigured Endpoint becomes a non-critical plugin issue (`issue.Network`, scoped to `endpoint[<idx>]`) and its phase contributes to the folded `AuthProxyReady` / `ExternalAccessReady` conditions, while sibling Endpoints keep converging.

Validation rules largely mirror the legacy validator, applied per Endpoint, plus:
- `services` is optional when `externalAccess` is provided. If present, it is currently capped at 1 (`field.TooMany` if `len > 1`); the slice is kept in the schema for forward compatibility.
- K8s Service names must be unique across all Endpoints in the payload.
- Mode (legacy vs annotation) is enforced mutually exclusive (see above).

### Activation (legacy mode)

The plugin is triggered when `spec.instructions.networking` is present in the WorkloadRunner specification.

### Configuration

```yaml
instructions:
  networking:
    # === Services ===
    # Array of services (currently max 1 supported)
    services:
      - # === Workload Service (choose ONE: create OR ref) ===

        # Option A: Create new service
        create:                            # Full corev1.Service
          metadata:
            name: "my-service"
          spec:
            selector:
              app: my-app
            ports:
              - port: 8000
                targetPort: 8000

        # Option B: Reference existing service
        # NOTE: When using 'ref', you MUST specify 'auth'
        # If you don't need authentication, use a plain Ingress resource instead
        ref:
          name: "existing-service"
          port: 8000

        # === Authentication (optional) ===
        # When auth section is present, authentication is enabled
        # When auth section is absent, authentication is disabled
        auth:
          authorizationType: authorizedUsers   # authenticatedUsers | authorizedUsers | authorizedGroups
          authorizedUsers: "user@example.com,user2@example.com"    # required when authorizationType=authorizedUsers
          authorizedGroups: "ml-team,data-scientists"              # required when authorizationType=authorizedGroups

    # === External Access ===
    # ingress and route are mutually exclusive; httpRoute may accompany either for Gateway API migration.
    externalAccess:
      # For standard Kubernetes:
      ingress:                             # Full networkingv1.Ingress
        metadata:
          name: "my-ingress"
          annotations:
            nginx.ingress.kubernetes.io/ssl-redirect: "true"
        spec:
          ingressClassName: nginx
          rules:
            - host: "my-app.example.com"
              http:
                paths:
                  - path: /
                    pathType: Prefix
                    backend:
                      service:
                        name: "my-service"    # Must match services[0] (name and port)
                        port:
                          number: 8000
          tls:
            - hosts:
                - "my-app.example.com"
              secretName: "my-tls-secret"

      # For OpenShift:
      route:                               # Full openshiftv1.Route
        metadata:
          name: "my-route"
        spec:
          host: "my-app.example.com"
          tls:
            termination: edge
            insecureEdgeTerminationPolicy: Redirect

    # === OAuth2 Proxy Advanced Options (optional) ===
    # Configure OAuth2 proxy response format
    forceJsonResponse: false           # Force JSON response format instead of HTML (default: false)
```

### Important Notes

1. **Service Definition is Optional**: You can omit the `services` array entirely if you only want to create an Ingress/Route to benefit from cluster-wide defaults (body size limits, annotations, TLS configuration). In this case, the plugin won't modify your backends.

2. **Backend Validation**: When a service is provided, all backend services specified in Ingress paths MUST match the service (both name and port). The plugin validates this and will reject configurations that don't match.

3. **Backend Override**: When a service is provided, the plugin automatically overrides all Ingress/Route backends:
   - **With auth**: Routes to OAuth2 proxy service (which forwards to your service)
   - **Without auth**: Routes directly to your service

4. **Service Reference Requires Auth**: When using `ref` (referencing an existing service), you MUST specify `auth`. If you don't need authentication, create a plain Kubernetes Ingress resource instead of using this plugin.

### OAuth2 Proxy Advanced Configuration

#### `forceJsonResponse`
- **Type**: boolean
- **Default**: false
- **Purpose**: Configures OAuth2 proxy to return JSON-formatted error responses instead of HTML

When `true`:
- OAuth2 proxy returns JSON error responses (better for API clients and programmatic access)
- Use when the workload is primarily an API endpoint that clients will access programmatically

When `false` (default):
- OAuth2 proxy returns HTML error pages (better for human-facing web applications)

#### Usage Examples

**Example 1: Standard Web Application (Default)**
```yaml
instructions:
  networking:
    services:
      - ref:
          name: "my-service"
          port: 8000
        auth:
          authorizationType: authenticatedUsers
    # forceJsonResponse not set (defaults to false, HTML error pages)
    externalAccess:
      ingress:
        spec:
          rules:
            - host: "my-app.example.com"
              http:
                paths:
                  - path: /
                    pathType: Prefix
                    backend:
                      service:
                        name: "my-service"
                        port:
                          number: 8000
```

**Example 2: API Endpoint with JSON Responses**
```yaml
instructions:
  networking:
    services:
      - ref:
          name: "api-service"
          port: 8080
        auth:
          authorizationType: authorizedUsers
          authorizedUsers: "data-scientist@example.com"
    forceJsonResponse: true           # Return JSON errors for API clients
    externalAccess:
      ingress:
        spec:
          rules:
            - host: "api.example.com"
              http:
                paths:
                  - path: /
                    pathType: Prefix
                    backend:
                      service:
                        name: "api-service"
                        port:
                          number: 8080
```

#### When to Use This Flag

| Scenario | forceJsonResponse | Use Case |
|----------|-------------------|----------|
| Web applications | `false` | Human-facing apps with HTML error pages |
| REST APIs | `true` | Programmatic access requiring JSON responses |
| ML model inference endpoints | `true` | API clients expecting JSON error format |

**Example 3: Ingress Only (No Service Management)**
```yaml
instructions:
  networking:
    # No services array - just creating Ingress for cluster defaults
    externalAccess:
      ingress:
        metadata:
          name: my-ingress
        spec:
          rules:
            - host: "external-service.example.com"
              http:
                paths:
                  - path: /
                    pathType: Prefix
                    backend:
                      service:
                        name: "some-external-service"  # Service not managed by plugin
                        port:
                          number: 8080
```

**Use Case**: You want to leverage cluster-wide Ingress defaults (body size limits, annotations, TLS config) without using the plugin's service management or auth features.

### Resources Created

| Resource | Created When |
|----------|--------------|
| Auth Proxy Deployment | `auth` section is present |
| Auth Proxy Service | `auth` section is present |
| Workload Service | `create` is specified |
| Ingress | `externalAccess.ingress` is present (backends overridden if service defined, preserved if no service) |
| Route | `externalAccess.route` is present (backend overridden if service defined, preserved if no service) |
| HTTPRoute | `externalAccess.httpRoute` is present (backends overridden if service defined, preserved if no service) |

### Request Flow

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              EXTERNAL REQUEST                               │
└─────────────────────────────────────────────────────────────────────────────┘
                                      │
                                      ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                    INGRESS / ROUTE (optional)                               │
│        Created when: externalAccess.ingress or externalAccess.route         │
└─────────────────────────────────────────────────────────────────────────────┘
                                      │
                    ┌─────────────────┴─────────────────┐
                    │                                   │
                    ▼                                   ▼
         (auth present)                         (auth absent)
                    │                                   │
                    ▼                                   │
┌───────────────────────────────────┐                   │
│       AUTH PROXY SERVICE          │                   │
│  Created when: auth is present    │                   │
│                                   │                   │
└───────────────────────────────────┘                   │
                    │                                   │
                    ▼                                   │
                    └─────────────────┬─────────────────┘
                                      │
                                      ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                         WORKLOAD SERVICE                                    │
│   Option A: ref (existing service from workload)                            │
│   Option B: create (plugin creates service)                                 │
└─────────────────────────────────────────────────────────────────────────────┘
                                      │
                                      ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                           WORKLOAD PODS                                     │
│                     (NIMService, LeaderWorkerSet, etc.)                     │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Flow Summary

| Auth | External Access | Flow |
|------|-----------------|------|
| Present | Ingress/Route | `Ingress/Route → Auth Proxy → Workload Service → Pods` |
| Absent | Ingress/Route | `Ingress/Route → Workload Service → Pods` |
| Present | None | `Auth Proxy Service → Workload Service → Pods` (ClusterIP access) |
| Absent | None | `Workload Service → Pods` (ClusterIP access) |

## Auth-Enabled Ingress Behavior

When auth is enabled, the plugin configures oauth2-proxy for standalone reverse proxy mode:

- `--reverse-proxy=true` — trust `X-Forwarded-*` headers from the ingress controller
- `--proxy-prefix=/oauth2` — fixed callback path prefix

The plugin also strips regex capture group suffixes from the ingress path when computing the oauth2-proxy upstream and cookie path (e.g., `/my-app(/|$)(.*)` → `/my-app`).

### Path rewriting and forwarded URI headers

When auth is combined with path-based routing (non-root path), the ingress controller needs to forward the original request URI to oauth2-proxy so it can build the correct redirect URL after path rewriting. These annotations are set by the **control plane** when creating the WorkloadRunner — not by this plugin:

| Ingress controller | Annotation set by control plane |
|---|---|
| nginx | `nginx.ingress.kubernetes.io/configuration-snippet: proxy_set_header X-Forwarded-Uri $request_uri;` |
| HAProxy | `haproxy.ingress.kubernetes.io/request-set-header: X-Forwarded-Uri %[req.uri]` |

For nginx, `configuration-snippet` annotations require cluster admin enablement — see the NIM networking documentation for the required `kubectl patch` command.

## Cluster-Wide Defaults

The plugin enriches Ingress/Route specs with defaults. **User values always win** - defaults only apply to unspecified fields.

### Defaults Applied

**Ingress:**
- IngressClassName from `INGRESS_CLASS` (cluster-configurable)
- Annotations from `INGRESS_ANNOTATIONS` (cluster-configurable, merged with user annotations)
- Body size limit `8GB` (hardcoded, overridable via annotation)

**OpenShift Route:**
- TLS termination with redirect (defaults to `edge`; switch cluster-wide to `passthrough` via
  `spec.anyworkload-controller.openshiftRouteConfiguration.terminationPolicy` in `RunaiConfig`,
  overridable per-route via `spec.tls`)
- Annotations from `INGRESS_ANNOTATIONS` (cluster-configurable, merged with user annotations)

**Cluster-configurable**: Set via controller env vars, changeable by admins.
**Hardcoded**: Built into plugin code, overridable by users in YAML.

## Validation

The plugin validates:
- **Service definition**: Optional. If provided, must specify either `create` OR `ref` (mutually exclusive)
- **Service selector**: When using `create`, service must have a `selector`
- **Service reference with auth**: When using `ref`, `auth` is REQUIRED (if no auth needed, use plain Ingress instead)
- **Auth requires external access**: When `auth` is present, `externalAccess` (Ingress or Route) is required
- **Authorization type**: `authorizationType` is required when auth section is present
- **Authorized users**: `authorizedUsers` required when `authorizationType=authorizedUsers`
- **Authorized groups**: `authorizedGroups` required when `authorizationType=authorizedGroups`
- **External access**: `externalAccess.ingress` and `externalAccess.route` are mutually exclusive; `externalAccess.httpRoute` may be supplied with either one for Gateway API migration.
- **Ingress validation**: Must have at least one rule with HTTP paths
- **Ingress backends**: When service is defined, all backend services in all paths must match the service name and port. When no service is defined, backends are not validated.
- **Route validation**: Must have a host specified

## Owner References

All created resources have owner references to the WorkloadRunner, ensuring automatic cleanup when the WorkloadRunner is deleted.
