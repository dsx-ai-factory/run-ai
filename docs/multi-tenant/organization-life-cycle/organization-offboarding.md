# Organization Offboarding

## Delete Tenants

Use this endpoint to remove a tenant from your NVIDIA Run:ai environment. You can choose between a soft delete (retains tenant data for future recovery) or a hard delete (permanently removes the tenant and all associated data). See [Tenant](https://app.gitbook.com/s/LYHZWRVgpD7c9mHAx1Vk/organizations/tenant) API endpoint for more details.

Send a `DELETE` request to the `/api/v1/tenants/{tenantId}` endpoint:

* **Tenant ID** (required, _path parameter_) - The numeric identifier of the tenant you want to delete.
* **isHardDelete** (optional, query parameter):
  * Soft delete - `false` (default) - The tenant is marked as deleted but data remains in the system.
  * Hard delete - `true` - All tenant data is permanently removed.

**Example request:**

```bash
curl -X DELETE 'https://console.<DOMAIN>/api/v1/tenants/123?isHardDelete=true' \
  -H 'Authorization: Bearer <TOKEN>'
```

**Example response:**

```bash
{
  "uid": "string"
}
```
