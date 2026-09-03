# runai-custom-providers

Custom Keycloak SPI extensions for Run:ai. Compiled to a single JAR (`runai-custom-providers.jar`) and loaded by Keycloak at startup via the providers directory.

## Requirements

- Java 21
- Gradle 8.x (wrapper included)
- Keycloak 26.x

## Build

```bash
./gradlew build
```

The output JAR is written to `build/libs/runai-custom-providers.jar`. Copy it to the Keycloak `providers/` directory (or let the sidecar image handle it).

Run tests:

```bash
./gradlew test
```

## Extensions overview

| Provider ID | Type | Description |
|---|---|---|
| `runai-claim-validator` | Identity Provider Mapper | Blocks federated login unless the user's OIDC claim matches an allowed value |
| `runai-oidc-user-attribute-idp-mapper` | Identity Provider Mapper | Maps an OIDC claim (string, list, or map) to a Keycloak user attribute |
| `runai-oidc-custom-groups-mapper` | OIDC Protocol Mapper | Emits a token claim containing the user's local groups, with optional intersection against an external authorization service |
| `runai-counter-mapper` | OIDC Protocol Mapper | Emits a token claim containing the count of a user attribute list |
| `runai-user-app-attribute-mapper` | OIDC Protocol Mapper | For service-account tokens: resolves the real user via `runai.created.by` and emits one of their attributes (groups, uid, gid, supplementarygroups, ...) |
| `last-login-event` | Event Listener | Records `runai.login.time` on users and clients after every login |
| `password-audit-event` | Event Listener | Posts an audit log entry to the audit service whenever a user resets or sets their password |
| `runaienteties` | Admin REST Resource | Adds a paginated, filterable, sortable `/admin/realms/{realm}/runaienteties/users` endpoint |

---

## Identity Provider Mappers

These mappers fire during federated (brokered) identity flows — i.e., when a user signs in via an external OIDC identity provider.

### ClaimValidator (`runai-claim-validator`)

Validates that an incoming OIDC claim contains one of the expected values. If the claim is absent or does not match, the login is blocked with a `401 Unauthorized` response.

**Compatible providers:** `keycloak-oidc`, `oidc`

**Configuration**

| Property | Required | Description |
|---|---|---|
| `claim` | Yes | Name of the claim to inspect (dot notation for nested claims; escape literal `.` with `\.`) |
| `claim.value` | Yes | Whitespace-separated list of allowed values — the user must match one |

**Error logging:** logs a `WARN` when the claim value is invalid or missing.

---

### CustomClaimToUserAttribute (`runai-oidc-user-attribute-idp-mapper`)

Maps an OIDC claim from an external identity provider to a Keycloak user attribute. Extends Keycloak's built-in mapper to additionally handle claims in **map** format (not just plain strings and lists). Data from the UserInfo endpoint takes priority over the same claim in the ID token.

**Configuration**

| Property | Required | Description |
|---|---|---|
| `claim` | Yes | Claim name (dot notation supported) |
| `user.attribute` | Yes | Keycloak user attribute to write to |

**Error logging:** logs errors when the raw ID token cannot be parsed.

---

## OIDC Protocol Mappers

These mappers add claims to access tokens, ID tokens, and/or userinfo responses for clients in the realm.

### CustomGroupsMapper (`runai-oidc-custom-groups-mapper`)

Adds a configurable claim to tokens containing the user's local group names. When the user has more than 10 groups and intersection is enabled, the list is intersected against groups fetched from the authorization service — this keeps very large group lists manageable for downstream consumers.

**Configuration**

| Property | Required | Default | Description |
|---|---|---|---|
| `tenant.id` | Yes | — | Tenant ID sent as `X-Runai-Tenant-Id` when calling the authorization service |
| `groups.user.attribute` | Yes | — | User attribute holding the list of local group names |
| `groups.intersection.enabled` | No | `true` | When `true` and the user has > 10 groups, intersect with external groups |
| Token claim name | No | `custom_groups_mapper` | Name of the claim written to the token |

**Error logging:** logs `ERROR` when the authorization service returns a non-200 status or the HTTP call fails.

---

### CounterMapper (`runai-counter-mapper`)

Reads a multi-valued user attribute and emits its element count as a numeric token claim. Useful for exposing `groups_count` without leaking group names.

If the attribute is absent or empty, the claim is **not added** to the token — callers should handle a missing claim.

**Configuration**

| Property | Required | Default | Description |
|---|---|---|---|
| `list.user.attribute` | Yes | — | User attribute containing the list to count |
| Token claim name | No | `groups_count` | Name of the claim written to the token |

---

### UserAppsMapper (`runai-user-app-attribute-mapper`)

Designed for **service-account tokens**. A service account in Run:ai represents an application created on behalf of a real user. This mapper reads the `runai.created.by` attribute on the OIDC client to find the real user, then copies a configurable attribute of that user (e.g. groups, uid, gid, supplementarygroups) into the token — with the same optional intersection logic as `CustomGroupsMapper` when mapping groups.

One mapper instance per attribute/claim is configured on the client (e.g. one for `groups`, one for `uid`, one for `gid`, one for `supplementarygroups`).

Skips non-service-account sessions with a `WARN` log.

**Configuration**

| Property | Required | Default | Description |
|---|---|---|---|
| `tenant.id` | No | — | Tenant ID for the authorization service call |
| `groups.user.attribute` | Yes | — | User attribute to copy from the real user (e.g. `GROUPS`, `UID`, `GID`, `SUPPLEMENTARYGROUPS`) |
| `multivalued` | No | `true` | If `true`, emits the claim as a list; if `false`, emits only the first value as a scalar string |
| `groups.intersection.enabled` | No | `true` | Intersect with external groups when count > 10 (only meaningful when mapping groups) |
| Token claim name | No | `user_app_groups` | Name of the claim written to the token |

**Error handling:** when fetching external groups fails, the exception is caught and an empty list/empty scalar is emitted instead of aborting token issuance (contrast with `CustomGroupsMapper`, which propagates a 500).

**Error logging:** logs `WARN` when client/user is missing from the session or `runai.created.by` is absent; logs `ERROR` on unexpected exceptions and when fetching external groups fails.

---

## Event Listeners

Register these in the Keycloak Admin Console under **Realm Settings → Events → Event listeners**.

### LastLoginEventListenerProvider (`last-login-event`)

Records the Unix epoch (seconds) of the last successful login:

- On `LOGIN` events — sets `runai.login.time` on the **user**.
- On `CLIENT_LOGIN` events — sets `runai.login.time` on the **client**.

No external calls; no configuration required.

**Error logging:** logs errors when user or client lookup fails unexpectedly.

---

### PasswordAuditEventListenerProvider (`password-audit-event`)

Listens for `UPDATE_CREDENTIAL` events where `credential_type=password` and asynchronously POSTs an audit log entry to the Run:ai audit service.

Determines the action label from the `redirect_uri` query parameter:

- `?runai_flow=invite` → `"Password set (Invite link)"`
- anything else → `"Password reset (Forgot password)"`

**Environment variables**

| Variable | Default | Description |
|---|---|---|
| `KC_RUNAI_AUDIT_SERVICE_URL` | `http://runai-backend-audit-service:8080` | Base URL of the audit service |

**Audit payload**

```json
{
  "tenant_id": "<realm attribute runai.tenant.id>",
  "subject": "<user email or username>",
  "subject_type": "user",
  "action": "Password reset (Forgot password)",
  "result": "Succeeded"
}
```

**Error logging:** logs `WARN` when the tenant ID cannot be resolved; logs `ERROR` when the audit HTTP POST fails or returns a non-2xx status.

---

## Custom Admin REST Resource (`runaienteties`)

Mounts an additional endpoint under the Keycloak admin REST API:

```
GET /admin/realms/{realm}/runaienteties/users
```

Requires the caller to hold the `view-users` permission in realm management.

### Query parameters

| Parameter | Type | Default | Description |
|---|---|---|---|
| `max` | integer | 1000 | Maximum users to return |
| `offset` | integer | 0 | Pagination offset |
| `sortBy` | string | `username` | Field to sort by: `username`, `createdBy`, `creationTime`, `lastLogin`, `lastUpdated`, `type`, `fullName` |
| `sortOrder` | string | `asc` | `asc` or `desc` |
| `search` | string | — | Full-text search across username, email, full name, createdBy |
| `minimalSearch` | boolean | `false` | When `true`, search only matches username and full name |
| `username` | string | — | Filter by email or username substring |
| `fullName` | string | — | Filter by first + last name substring |
| `createdBy` | string | — | Filter by `runai.created.by` attribute substring |
| `q` | string | — | Comma-separated `key:value` attribute filters passed to Keycloak's user search |
| `creationFilter` | string | — | Date filter on `runai.creation.time` — prefix with `>` (on or after), `<` (on or before), or `=` (same calendar day) followed by epoch seconds |
| `loginFilter` | string | — | Date filter on `runai.login.time` — same operators as `creationFilter` |
| `updateFilter` | string | — | Date filter on `runai.update.time` — same operators as `creationFilter` |

### Response

```json
{
  "users": [
    {
      "id": "...",
      "username": "...",
      "email": "...",
      "firstName": "...",
      "lastName": "...",
      "createdAt": "1700000000",
      "isLocal": "true",
      "lastLogin": "1700100000",
      "updatedAt": "1700050000",
      "createdBy": "admin@example.com",
      "groups": "group-a,group-b"
    }
  ],
  "count": 42
}
```

Service-account users and users without an email address are excluded from results.

**Error logging:** logs `WARN` on malformed query parameters; logs `ERROR` when date attribute parsing fails.

---

## External service integration

### Authorization service

`CustomGroupsMapper` and `UserAppsMapper` call the Run:ai authorization service to fetch the list of known groups when performing intersection.

| Variable | Default | Description |
|---|---|---|
| `KC_RUNAI_AUTHORIZATION_SERVICE_URL` | `http://runai-backend-authorization:8080` | Base URL |

Endpoint called: `GET {url}/internal/authorization/groups` with optional header `X-Runai-Tenant-Id`.

Responses are expected in the form `{"names": ["group-a", "group-b"]}`.

Requests time out after **2 seconds**. Errors are logged and propagated as `500 Internal Server Error` to abort the token issuance.
