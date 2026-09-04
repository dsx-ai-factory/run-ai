# API Response Cache

A transparent caching layer for API requests that reduces backend load by caching GET responses and automatically invalidating the cache on mutating operations.

## Overview

This caching solution works at the Axios interceptor level, meaning **no changes are required in your service files**. All caching happens automatically behind the scenes.

### Key Features

- **Transparent caching** - No code changes needed in services
- **Explicit path mapping** - Manual control over which paths are cached
- **TTL-based expiration** - Configurable time-to-live per entity type
- **Automatic invalidation** - Cache is cleared on POST/PUT/DELETE/PATCH requests
- **Query parameter aware** - Different filters = different cache entries
- **Request deduplication** - Prevents identical in-flight requests
- **Debug logging** - Optional console logs for cache hits/misses

## How It Works

```
┌─────────────────────────────────────────────────────────────────┐
│                        GET Request                               │
├─────────────────────────────────────────────────────────────────┤
│  1. Check if URL is excluded from caching                       │
│  2. Match URL path against pathToEntity mapping                 │
│  3. If no match → skip caching                                  │
│  4. Generate cache key (path + sorted query params)             │
│  5. If cached & not expired → return cached response            │
│  6. If not cached → make request, cache response, return        │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                   POST/PUT/DELETE/PATCH Request                  │
├─────────────────────────────────────────────────────────────────┤
│  1. Make the request normally                                   │
│  2. Match URL path against pathToEntity mapping                 │
│  3. Invalidate ALL cache entries for that entity                │
└─────────────────────────────────────────────────────────────────┘
```

## Configuration

Edit `cache.config.ts` to configure caching behavior.

### Step 1: Add Path to Entity Mapping

Map API paths to entity names. The path is matched as a prefix.

```typescript
// cache.config.ts
export const pathToEntity: Record<string, string> = {
  // Settings
  "/v1/k8s/setting": "setting",

  // Add more mappings:
  "/api/v1/assets/environment": "environment",
  "/api/v1/assets/compute": "compute",
  "/v1/k8s/clusters": "cluster",
  "/api/v2/org/projects": "project",
};
```

**Path matching rules:**

- Exact match: `/v1/k8s/setting` matches `/v1/k8s/setting`
- Prefix match: `/v1/k8s/setting` also matches `/v1/k8s/setting/123`
- First match wins: put more specific paths first if needed

### Step 2: Set TTL for Entity

Configure how long responses should be cached for each entity.

```typescript
// cache.config.ts
export const entityTTL: Record<string, number> = {
  setting: 5 * MINUTES,
  environment: 2 * MINUTES,
  compute: 5 * MINUTES,
  cluster: 2 * MINUTES,
  project: 2 * MINUTES,
};
```

### TTL Helper Constants

```typescript
import { SECONDS, MINUTES, HOURS } from "./cache.config";

// Examples:
30 * SECONDS; // 30 seconds
2 * MINUTES; // 2 minutes
1 * HOURS; // 1 hour
```

### Grouped Invalidation

Multiple paths can map to the same entity. When any path is mutated, ALL cached entries for that entity are invalidated.

```typescript
export const pathToEntity: Record<string, string> = {
  "/v1/k8s/setting": "setting",
  "/v1/k8s/setting/batch": "setting", // Same entity - invalidates together
};
```

### Excluded URLs

Some URLs are always excluded from caching regardless of mapping:

```typescript
export const excludeFromCache: RegExp[] = [
  /\/health/,
  /\/ready/,
  /\/token/,
  /\/login/,
  /\/logout/,
  /\/auth\//,
  /\/metrics\//,
  /\/audit\//,
];
```

### Debug Logging

Enable console logging to see cache operations:

```typescript
export const CACHE_DEBUG = true; // Set to false in production
```

Console output example:

```
[Cache MISS] setting: /v1/k8s/setting
[Cache SET] setting: /v1/k8s/setting (entity: setting, ttl: 300000ms)
[Cache HIT] setting: /v1/k8s/setting (age: 5000ms, ttl: 300000ms)
[Cache INVALIDATE] setting: cleared 3 entries
```

## Cache Key Generation

Cache keys are generated from the URL path and query parameters:

| Request                           | Cache Key                                           |
| --------------------------------- | --------------------------------------------------- |
| `GET /v1/k8s/setting`             | `/v1/k8s/setting`                                   |
| `GET /v1/k8s/setting?projectId=1` | `/v1/k8s/setting?{"projectId":1}`                   |
| `GET /v1/k8s/setting?projectId=2` | `/v1/k8s/setting?{"projectId":2}`                   |
| `GET /v1/k8s/setting?a=1&b=2`     | `/v1/k8s/setting?{"a":1,"b":2}`                     |
| `GET /v1/k8s/setting?b=2&a=1`     | `/v1/k8s/setting?{"a":1,"b":2}` ← Same key (sorted) |

**Important:** Different query parameters create separate cache entries, but ALL entries for an entity are invalidated on mutation.

## Manual Cache Control

For cases where you need to manually control the cache:

```typescript
import { cacheUtils } from "@/services/infra/cache";

// Invalidate all cache entries for an entity
cacheUtils.invalidateEntity("setting");

// Invalidate a specific URL (with optional params)
cacheUtils.invalidateUrl("/v1/k8s/setting", { projectId: 1 });

// Clear the entire cache
cacheUtils.clearAll();

// Get cache statistics (for debugging)
const stats = cacheUtils.getStats();
console.log(stats);
// { size: 5, entries: [{ key: "...", entity: "setting", age: 1234, ttl: 300000 }, ...] }
```

## Architecture

```
src/services/infra/cache/
├── index.ts              # Barrel exports
├── cache.config.ts       # Configuration (path mapping, TTLs, exclusions)
├── cache.service.ts      # Core cache logic (storage, TTL, invalidation)
├── cache.interceptor.ts  # Axios interceptors (request/response handling)
└── README.md             # This file
```

### Integration Points

1. **`auth-axios.service.ts`** - Cache interceptors are installed on the authenticated Axios instance
2. **`auth.store.ts`** - Cache is cleared on user logout

## Adding a New Cached Endpoint

1. **Find the API path** - Check the swagger-generated client or network tab

   ```typescript
   // Example: /v1/k8s/setting
   ```

2. **Add path mapping** in `cache.config.ts`:

   ```typescript
   export const pathToEntity: Record<string, string> = {
     "/v1/k8s/setting": "setting",
     "/api/v1/assets/environment": "environment", // ← Add this
   };
   ```

3. **Add TTL** (if new entity) in `cache.config.ts`:
   ```typescript
   export const entityTTL: Record<string, number> = {
     setting: 5 * MINUTES,
     environment: 2 * MINUTES, // ← Add this
   };
   ```

That's it! All GET requests to that path will now be cached.

## Best Practices

1. **Start with short TTLs** - Begin with 30s-2min and increase based on data staleness tolerance
2. **Critical data** - Use shorter TTLs or don't add mapping for frequently changing data
3. **Reference data** - Use longer TTLs (5-10min) for data that rarely changes
4. **Group related paths** - Map related paths to the same entity for proper invalidation

## Troubleshooting

### Cache not working?

1. Check `CACHE_DEBUG = true` and look at console logs
2. Verify the path is in `pathToEntity` mapping
3. Verify the entity has TTL > 0 in `entityTTL`
4. Check URL path matches exactly (check network tab)
5. Ensure it's a GET request (only GET is cached)

### Stale data issues?

1. Reduce TTL for the entity
2. Manually invalidate: `cacheUtils.invalidateEntity("entityName")`
3. Check that mutations go through the same Axios instance

### Cache not invalidating?

1. Ensure POST/PUT/DELETE path is mapped to same entity
2. Verify the entity name matches between GET and mutation paths
3. Check console logs for `[Cache INVALIDATE]` messages
