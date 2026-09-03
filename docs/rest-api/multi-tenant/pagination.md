# Pagination

Several List endpoints in the NVIDIA Run:ai control plane API return large collections. To retrieve results in smaller chunks, use offset/limit pagination.

## How Pagination Works

### Request Parameters

When an endpoint supports pagination, it typically exposes:

* `limit` - Maximum number of items to return (default 50, range 1..500)
* `offset` - The offset of the first item returned in the collection

Example: `GET /api/v1/authorization/access-rules?offset=0&limit=50`

### Pagination Response

List responses may include a `next` field, alongside the collection itself.

Example response shape:

```json
{
  "next": 1,
  "roles": [ ... ]
}
```

Use `next` as the pointer for retrieving the next page of results. If `next` is not returned (or is empty), you've reached the end of the collection.

### Example: Paginate Through Results

#### Fetch the first page

```bash
curl -sS \
  -H "Authorization: Bearer <TOKEN>" \
  "https://console.<DOMAIN>/api/v2/authorization/roles?offset=0&limit=50"
```

#### Fetch the next page

Take the `next` value from the response and use it as the offset in the next request.

### Combine Pagination with Sorting and Filtering

List endpoints also support:

* **Sorting** - `sortBy` and `sortOrder`
* **Structured filtering** - `filterBy` with operators like `==`, `!=`, `>=`, `=@` (contains), etc.
* **Free-text search** - `search`

Example:

```bash
curl -sS \
  -H "Authorization: Bearer <TOKEN>" \
  "https://console.<DOMAIN>/api/v2/authorization/roles?limit=100&offset=0&sortBy=createdAt&sortOrder=desc&filterBy=allocatedGPU>=2"
```
