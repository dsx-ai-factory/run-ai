# Audit Logs API

This section provides details about NVIDIA Run:ai’s Audit log.

The NVIDIA Run:ai control plane provides the audit log API to reflect the information regarding changes to business objects: clusters, projects and assets etc. Go to the [Audit Logs](api:multi-tenant-api:GET/api/v1/audit/log) API reference to view the available actions. Since the amount of data is not trivial, the API is based on paging. It retrieves a specified number of items for each API call. You can get more data by using subsequent calls.

<Note>

**Note**

Only system administrator users with tenant-wide permissions can access Audit log.

</Note>

## Retrieve Audit Logs

You can retrieve audit logs programmatically to track tenant-related operations, user actions, and system changes across their managed environment.

### Define Query Parameters

Use the following optional parameters to filter results:

* `startTime` / `endTime` - ISO 8601 format (e.g., `2024-06-01T00:00:00Z`)
* `eventTypes` - Filter specific event types (e.g., `CREATE`, `DELETE`, `LOGIN`)
* `resourceTypes` - Filter by resource type (e.g., `TENANT`, `CLUSTER`, `USER`)
* `userIds` - Filter by one or more user IDs
* `page` / `pageSize` - Paginate results
* `sort` – `asc` or `desc`

### Send the API Request

```bash
curl -L 'https://console.<DOMAIN>/api/v1/audit-logs' \
  -H 'Authorization: Bearer <ACCESS_TOKEN>' \
  -H 'Content-Type: application/json' \
  -G \
  --data-urlencode "startTime=2024-06-01T00:00:00Z" \
  --data-urlencode "endTime=2024-06-08T00:00:00Z" \
  --data-urlencode "resourceTypes=TENANT,CLUSTER" \
  --data-urlencode "eventTypes=CREATE,DELETE" \
  --data-urlencode "sort=desc"
```

Each entry contains metadata about the action:

* `timestamp`
* `eventType`
* `resourceType` and `resourceId`
* `userId` who initiated the action
* `details` describing the change

**Example Response**

```json
{
  "items": [
    {
      "timestamp": "2024-06-08T10:30:00Z",
      "eventType": "CREATE",
      "resourceType": "TENANT",
      "resourceId": "tenant-a",
      "userId": "admin@hostorg.com",
      "details": "Tenant created via Management Interface API"
    },
    {
      "timestamp": "2024-06-07T12:15:00Z",
      "eventType": "DELETE",
      "resourceType": "CLUSTER",
      "resourceId": "cluster-b",
      "userId": "ops@hostorg.com",
      "details": "Cluster deleted via host org automation"
    }
  ],
  "page": 1,
  "pageSize": 50,
  "total": 2
}
```
