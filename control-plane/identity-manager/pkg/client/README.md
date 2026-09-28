# Identity Manager Go Client

A Go client for the Identity Manager HTTP API. It bundles a hand-written
ergonomic layer on top of the OpenAPI-generated client in
[`generated/`](generated) and ships an OAuth2 client-credentials helper
(`AuthnManager`) that can be plugged into any `net/http` client as a transport.

## Contents

- [Importing the client](#importing-the-client)
- [Creating a client](#creating-a-client)
- [`AuthnManager`: authenticated HTTP transport](#authnmanager-authenticated-http-transport)
- [Versioning policy](#versioning-policy)
  - [Development versions](#development-versions)
  - [Production releases](#production-releases)

## Importing the client

The client is published as a **versioned Go module**. Always pin to an explicit
released tag:

```go
import idm "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/client/v1.0.0"
```

Two top-level constructors are provided:

- `NewIdmClient(baseURL)` — full internal API (realms, internal users, internal
  clients, …). Use this from inside the control plane.
- `NewIdmExternalClient(baseURL)` — only the public-facing endpoints
  (login, token exchange, refresh).

Both accept a base URL and return an interface that can be mocked in tests via
the generated mocks under [`mocks/`](mocks).

## Creating a client

```go
idmClient := idm.NewIdmClient("http://identity-manager.runai-backend.svc.cluster.local")

resp, _, err := idmClient.CreateRealm(ctx, "my-realm", "tenant-123")
if err != nil {
    return err
}
```

## `AuthnManager`: authenticated HTTP transport

`AuthnManager` manages OAuth2 client-credentials tokens and implements
`http.RoundTripper`. When used as the transport of an `http.Client` it:

1. Performs the initial `client_credentials` login on construction.
2. Caches the access token and re-uses it until it is within `RefreshBuffer`
   of expiry (default: 30 seconds).
3. Transparently refreshes the token when needed.
4. On a `401 Unauthorized` response, refreshes the token **once** and replays
   the original request.

### Usage

```go
external := idm.NewIdmExternalClient("https://app.run.ai")

mgr, err := idm.NewAuthnManager(ctx, idm.AuthnManagerConfig{
    ClientId:      "my-client",
    ClientSecret:  "s3cr3t",
    IdmClient:     external,
    RefreshBuffer: 30 * time.Second, // optional
})
if err != nil {
    return err
}

httpClient := mgr.HttpClient() // == &http.Client{Transport: mgr}
resp, err := httpClient.Get("https://app.run.ai/api/v1/users")
```

> **Important:** When an HTTP client uses `AuthnManager` as its transport, the
> manager **always overwrites the `Authorization` header** on outgoing
> requests. Do not rely on passing a bearer token via context or pre-set
> headers — the manager's token takes precedence.
>
> Also note that on retry the original request body must be replayable
> (`http.Request.GetBody` must be set). The standard library sets it
> automatically for `bytes.Buffer`, `bytes.Reader`, and `strings.Reader`
> bodies.

## Versioning policy

The client is released independently of the service binary and follows
semantic versioning. Consumers must import an explicit version (see
[Importing the client](#importing-the-client)).

### Development versions

While iterating on a feature branch you can publish a pre-release tag from the
service's main directory:

```
make client-bump-version-dev
```

This creates a tag of the form:

```
control-plane/identity-manager/pkg/client/v1.0.0-dev.0
```

Development versions are intended **only** for use in feature branches.
**It is strictly prohibited to merge a pull request into `main` that
references a development version of the client.** CI will fail any such pull
request to enforce this policy.

### Production releases

To publish a production release of the client:

1. Merge all relevant client changes into `main`.
2. From the service's main directory run:

   ```
   make client-bump-version
   ```

This creates a release tag of the form:

```
control-plane/identity-manager/pkg/client/v1.0.0
```

Production versions are safe to reference from any branch or pull request, and
should be used in place of development versions as soon as the release is
cut.
