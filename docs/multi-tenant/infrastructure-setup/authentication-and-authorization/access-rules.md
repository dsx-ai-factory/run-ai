# Access Rules

Access rules provide users, groups, or service accounts privileges to system entities. An access rule is the assignment of a role to a subject in a scope: `<Subject>` is a `<Role>` in a `<Scope>`.&#x20;

## Scope Behavior

When creating an access rule, the `scopeType` and associated scope identifier, `scopeId`, determine where the access rule is available and who can use it in a multi-tenant environment:

* `scopeType = system` - Creates a global access rule that applies that is available across all customer tenants.
* `scopeType = tenant` + customer tenant ID - Creates an access rule that applies only within that customer tenant using `scopeId`.
* `scopeType = tenant` + Console/system tenant ID - Creates an access rule scoped to the platform owner (host organization) tenant using `scopeId`.

## Privileged Escalation Enforcement

When creating access rules across tenants, apply the principle of no privilege escalation:

* A user can assign roles only within the limits of their own permissions.
* Assigning a role that grants broader permissions than the caller is blocked or restricted by the platform.

## Create an Access Rule

To create an access rule:

* Send a `POST` request to the `/api/v1/authorization/access-rules` endpoint. See [Access rules](https://app.gitbook.com/s/LYHZWRVgpD7c9mHAx1Vk/authentication-and-authorization/access-rules) API for more details.
* Provide the following:
  * The **subject** you want to assign a role to:
    * `subjectType`: `user`, `group`, or `service-account`
    * `subjectId`: the unique identifier for that subject (for example, a user email or group name)
  * The **role ID** you want to assign (`roleId`). See Roles API to retrieve a list of roles.
  * The **scope** where the access rule applies:
    * `scopeType`: `system` or `tenant`
    * `scopeId`: the identifier of the system tenant or customer tenant

Example: Assign a role at **system scope** (global context):

```bash
curl -L 'https://<COMPANY-URL>/api/v1/authorization/access-rules' \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer <TOKEN>' \
  -d '{
    "subjectId": "platform-admin@company.com",
    "subjectType": "user",
    "roleId": 56,
    "scopeType": "system",
    "scopeId": "<SYSTEM_ID>"
  }'
```

Example: Assign a role within a specific **customer tenant**:

```bash
curl -L 'https://<COMPANY-URL>/api/v1/authorization/access-rules' \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer <TOKEN>' \
  -d '{
    "subjectId": "user@company.com",
    "subjectType": "user",
    "roleId": 12,
    "scopeType": "tenant",
    "scopeId": "<CUSTOMER-TENANT-ID>"
  }'
```

Example: Assign a role within the **console/system tenant**:

```bash
curl -L 'https://<COMPANY-URL>/api/v1/authorization/access-rules' \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer <TOKEN>' \
  -d '{
    "subjectId": "operator@company.com",
    "subjectType": "user",
    "roleId": 34,
    "scopeType": "tenant",
    "scopeId": "<CONSOLE-SYSTEM-TENANT-ID>"
  }'
```
