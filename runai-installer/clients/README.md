# Run:ai Installer Client

This module provides the shared authenticated control-plane client used by the installer. It handles token rotation, TLS configuration, and access to the identity-manager, authorization, cluster-service, and org-unit-service APIs.

The service API clients are maintained in their canonical modules:

- `control-plane/authorization/pkg/client/generated`
- `control-plane/cluster-service/pkg/client/generated`
- `control-plane/org-unit-service/pkg/client/generated`
- `control-plane/identity-manager/pkg/client/generated`

Local `replace` directives keep monorepo builds on the canonical source. Do not add generated API clients to this module.

## Usage

```go
client, err := clients.NewClient(fqdn, clients.Credentials{
	Username: username,
	Password: password,
})
if err != nil {
	return err
}
```

The returned client shares one authenticated HTTP transport across the service clients and refreshes its access token when needed.

## Validation

```bash
make build test lint validate
```
