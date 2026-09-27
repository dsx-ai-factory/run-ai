# Create a Tenant

Creating a tenant requires an authenticated Bearer token - see [API Access Setup](../guided-installation/api-access-setup.md) if you haven't generated one yet.

## Prepare Tenant Information

Prepare the required details for the tenant you want to create:

* **Tenant name** - A unique identifier for the tenant.
* **Administrator email** - The primary contact who will receive access credentials and manage the tenant’s environment.

Refer to the [Tenant](api:multi-tenant-api:POST/api/v1/tenants) API endpoint for more details.

## Create the Tenant

To register a new tenant in the system, send a `POST` request to the `/api/v1/tenants` endpoint with the tenant information. This creates the tenant and assigns it a unique domain under your managed wildcard DNS (e.g., `tenant-name.runai.hostorg-domain.com`) and exposes tenant-facing services.

**Example request:**

```bash
curl -L 'https://console.<DOMAIN>/api/v1/tenants' \ 
-H 'Content-Type: application/json' \
-H 'Authorization: Bearer <TOKEN>' \ 
{
  "name": "mytenant",
  "email": "administrator@mytenant.com"
}
```

After the tenant is successfully created, the API response includes a `domain` field. This is the fully qualified domain name (FQDN) that the tenant will use to access NVIDIA Run:ai services, either through a web browser or programmatic integration.

You will [share](share-tenant-access-details.md) this domain with the tenant’s administrators once cluster provisioning and connectivity steps are complete.

**Example response:**

```bash
tenant": {
    "id": 1,
    "name": "example",
    "displayName": "Example",
    "status": "Ready",
    "tenantId": 123,
    "eulaSigningUser": {
      "email": "user@tenant",
      "date": "2020-01-01T00:00:00Z",
      "ip": "192.168.10.210"
    },
    "created_at": "2020-01-01T00:00:00Z",
    "updated_at": "2020-01-02T00:00:00Z",
    "deleted_at": null
  },
  "additionalData": {
    "tenantDomain": "mytenant.<globaldomain>",
    "testUser": {
      "tenant-test@run.ai": "Abc!23"
    },
    "customerUser": {
      "administrator@mytenant.com": "<password>"
    }
  }
}
```

## View Existing Tenants

To retrieve a list of tenants registered in the system, send a `GET` request to the `/api/v1/tenants` endpoint. This returns details for each tenant, including the name, status, and domain information. You can use this endpoint to verify whether a tenant already exists or to audit tenant states across environments.

**Example request:**

```bash
curl -L 'https://console.<DOMAIN>/api/v1/tenants' \
-H 'Authorization: Bearer <TOKEN>' \
-H 'Accept: application/json'
```

**Example response:**

```bash
[
  {
    "id": 1,
    "name": "example",
    "displayName": "Example",
    "status": "Ready",
    "tenantId": 123,
    "createdAt": "2020-01-01T00:00:00Z",
    "updatedAt": "2020-01-02T00:00:00Z",
    "deletedAt": null
  }
]
```

**Tenant status values:**

The `status` field indicates the current lifecycle state of the tenant. Possible values include:

* `Creating` - The tenant is being provisioned.
* `Updating` - The tenant is being reconfigured.
* `Ready` - The tenant is active and fully operational.
* `Deleting` - The tenant is being decommissioned.
* `Deleted` - The tenant has been removed.
* `Failed` - Tenant creation or update has failed.

Use these values to track the readiness of tenant environments or detect errors during provisioning.
