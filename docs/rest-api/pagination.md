# Pagination

Several List endpoints in the NVIDIA Run:ai control plane API return large collections. To retrieve results in smaller chunks, use offset/limit pagination.

## How Pagination Works

### Request Parameters

When an endpoint supports pagination, it typically exposes:

* `limit` - Maximum number of items to return (default 50, range 1..500)
* `offset` - The offset of the first item returned in the collection

Example (workloads): `GET /api/v1/workloads?offset=0&limit=50`

### Pagination Response

List responses may include a `next` field, alongside the collection itself (for example, `workloads`).

Example response shape (workloads):

```json
{
  "next": 1,
  "workloads": [ ... ]
}
```

Use `next` as the pointer for retrieving the next "page" of results. If `next` is not returned (or is empty), you've reached the end of the collection.

### Example: Paginate Through Workloads

#### Fetch the first page

```bash
curl -sS \
  -H "Authorization: Bearer <TOKEN>" \
  "https://<COMPANY-URL>/api/v1/workloads?offset=0&limit=50"
```

The endpoint supports `offset` and `limit`.

#### Fetch the next page

Take the `next` value from the response and use it in the next request (as the next offset/page pointer).

### Combine Pagination with Sorting and Filtering

The workloads list endpoint also supports:

* **Sorting** - `sortBy` and `sortOrder`
* **Structured filtering** - `filterBy` with operators like `==`, `!=`, `>=`, `=@` (contains), etc.
* **Free-text search** - `search`

Example:

```bash
curl -sS \
  -H "Authorization: Bearer <TOKEN>" \
  "https://<COMPANY-URL>/api/v1/workloads?limit=100&offset=0&sortBy=createdAt&sortOrder=desc&filterBy=allocatedGPU>=2"
```
