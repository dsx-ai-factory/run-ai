# API Access Setup

To interact with the NVIDIA Run:ai platform programmatically, configure API access using access keys. This section outlines the steps to create a service account, generate tokens using the appropriate OAuth2 grant types, and authenticate securely with the platform APIs.

## Authenticate with Admin Credentials

Before registering a service account, authenticate using administrator credentials. Use the credentials configured via `global.management.user` and `global.management.password` as detailed in [Install the Control Plane](../../../docs/multi-tenant/installation-and-configuration/installation/install-using-helm/install-control-plane.md).

Send a `POST` request to the `/api/v1/token` endpoint with `grantType: password` to generate a token.

**Example request:**

```bash
curl -L 'https://console.<DOMAIN>/api/v1/token' \
-H 'Content-Type: application/json' \
-d '{
  "grantType": "password",
  "username": "user@example.com",
  "password": "yourPassword"
}'
```

**Example response:**

```json
{
  "accessToken": "<TOKEN>"
}
```

Use the returned `accessToken` as a Bearer token in the `Authorization` header for subsequent API requests.

## Create a Service Account

After generating a token using the admin password, use it to register a service account. Service accounts contain a `clientId` and `clientSecret` that are used to authenticate future API requests via the `client_credentials` grant type.

Send a `POST` request to the `/api/v1/service-accounts` endpoint.

**Example request:**

```bash
curl -L 'https://console.<DOMAIN>/api/v1/service-accounts' \
-H 'Content-Type: application/json' \
-H 'Authorization: Bearer <TOKEN>' \
-d '{
  "name": "serviceaccount-123"
}'
```

**Example response:**

```json
{
  "id": "0eeaf222-e503-4f35-9d9c-c419816272e3",
  "name": "app123",
  "secret": "asdasidjn9d",
  "clientId": "app123"
}
```

## Generate a Token with Client Credentials

Use the `clientId` and `clientSecret` from your service account to generate an access token for authenticated API calls using the `client_credentials` grant type.

Send a `POST` request to the `/api/v1/token` endpoint with `grantType: client_credentials`.

**Example request:**

```bash
curl -L 'https://console.<DOMAIN>/api/v1/token' \
-H 'Content-Type: application/json' \
-d '{
  "grantType": "client_credentials",
  "clientId": "clientID",
  "clientSecret": "clientSecret"
}'
```

**Example response:**

```json
{
  "accessToken": "string",
  "idToken": "string",
  "refreshToken": "string",
  "externalToken": "string"
}
```

Use the `accessToken` as a Bearer token in the `Authorization` header for all subsequent API requests.

## Create a Local User

Send a `POST` request to the `/api/v1/users` endpoint to create a local user.

**Example request:**

```bash
curl -L 'https://console.<DOMAIN>/api/v1/users' \
-H 'Content-Type: application/json' \
-H 'Authorization: Bearer <TOKEN>' \
-d '{
  "email": "admin@hostorg.com",
  "resetPassword": false
}'
```

**Example response:**

```json
{
  "id": "1a2b3c4d",
  "email": "admin@hostorg.com",
  "createdAt": "2024-06-01T00:00:00Z"
}
```
