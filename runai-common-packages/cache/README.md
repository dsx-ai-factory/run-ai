# `cache` Package Documentation

The `cache` package supports multiple backend integrations (in-memory, Redis, KeyDB, Valkey, NATS) with a simple generic API for storing, retrieving, and deleting cached values. It also provides distributed locking, observability, and store-level isolation.

---

## Table of Contents

- [Features and Public API](#features-and-public-api)
  - [Initialization (`Init`)](#initialization-init)
  - [Async Initialization (`InitAsync`)](#async-initialization-initasync)
  - [SetLogger](#setlogger)
  - [IsInitialized](#isinitialized)
  - [Closing (`Close`)](#closing-close)
- [Cache Operations](#cache-operations)
  - [Get](#get)
  - [Set](#set)
  - [Del](#del)
- [Stores](#stores)
  - [RegisterStore](#registerstore)
  - [RegisterSharedStore](#registersharedstore)
  - [RegisterStoreAsync / RegisterSharedStoreAsync](#registerstoreasync--registersharedstoreasync)
  - [ClearStore](#clearstore)
  - [ClearCache](#clearcache)
  - [Store Options](#store-options)
- [Distributed Locks](#distributed-locks)
  - [Lock](#lock)
  - [SharedLock](#sharedlock)
- [Supported Backends](#supported-backends)
- [Configuration](#configuration)
- [Metrics](#metrics)
- [Example Usage](#example-usage)
- [Testing and Mocking](#testing-and-mocking)

---

## Features and Public API

### Initialization (`Init`)

Initializes the cache system and connects to the configured backend. This is a **blocking** operation: it will retry connecting according to `CACHE_CONNECT_RETRIES` and will not return until the connection is established or all retries fail.

You can pass a configuration struct to `Init` to programmatically set options instead of (or in addition to) using environment variables.

> **Note:** As a general practice, it is usually preferred to allow the package to load the configuration by itself (using environment variables) rather than providing a config struct to the `Init` function. This ensures a consistent configuration standard across all services using this package.

**Example:**

```go
cfg := &cache_config.Cache{
    Type:            cache_config.CTRedis,
    Enabled:         ptr.Bool(true),
    Url:             "redis://localhost:6379",
    ServiceName:     "my-service",
    TtlSeconds:      3600,
    MaxStoreSizeMB:  200,
    MaxRecordSizeKB: 2,
}
err := cache.Init(cfg)
if err != nil {
    panic(err)
}
```

Pass `nil` to load defaults from environment variables:

```go
err := cache.Init(nil)
if err != nil {
    panic(err)
}
```

- **Blocking**: Yes (retries connection on startup)
- **Retry:** Exponential backoff controlled by `CACHE_CONNECT_RETRIES` and `CACHE_INIT_RETRY_INTERVAL_DURATION`.


---

### Async Initialization (`InitAsync`)

Initializes the cache system asynchronously. Unlike `Init`, this method **does not block**; it immediately returns a channel of errors. The connection process happens in the background, and any errors encountered during initialization are sent to the returned error channel.

This is useful for services that want to start up quickly and handle cache connection errors reactively. Typically, you will use `InitAsync` in combination with `RegisterStoreAsync` / `RegisterSharedStoreAsync` to register stores only after the connection is ready.

**Example:**

```go
errCh, err := cache.InitAsync(nil)
if err != nil {
    // Configuration error — the cache will never connect
    panic(err)
}

// Register a store that waits for init to complete in the background
storeErrCh := cache.RegisterStoreAsync[MyStruct](options.Ttl(time.Hour))

// Listen for errors from async init and store registration
go func() {
    for err := range errCh {
        log.Errorf("Cache init error: %v", err)
    }
}()
go func() {
    for err := range storeErrCh {
        log.Errorf("Store registration error: %v", err)
    }
}()
```

- **Blocking**: No (returns immediately)
- **Error Channel:** Initialization errors are sent to the returned channel.
- **Typical Usage:** Combine with `RegisterStoreAsync` / `RegisterSharedStoreAsync` for fully asynchronous cache handling.

---

### SetLogger

Sets a custom logger for the cache system. Should be called **before** `Init`. If not set, a default no-op logger is used (no logs).

The logger must implement:

```go
type Logger interface {
    Debugf(format string, args ...any)
    Infof(format string, args ...any)
    Warnf(format string, args ...any)
    Errorf(format string, args ...any)
}
```

**Example:**

```go
cache.SetLogger(myLogger)
```

- **Blocking**: No

---

### IsInitialized

Returns whether the cache has been successfully initialized.

```go
if cache.IsInitialized() {
    // cache is ready
}
```

---

### Closing (`Close`)

Gracefully closes all cache connections, stops observability samplers, and cleans up resources. After calling `Close`, the cache can be re-initialized with a new `Init` call.

```go
err := cache.Close()
```

- **Blocking**: Yes

---

## Cache Operations

### Get

Retrieves a value from the cache by key. The value is deserialized into the generic type `T`. Returns `nil, nil` if the key is not found or if the cache is disabled.

```go
val, err := cache.Get[MyStruct](ctx, "my-key")
if err != nil {
    // handle error
}
if val != nil {
    // use *val
}
```

- **Blocking**: Yes

---

### Set

Stores a value in the cache under the given key. The value is serialized to JSON before storage.

```go
err := cache.Set[MyStruct](ctx, "my-key", myValue)
```

- **Blocking**: Yes

---

### Del

Deletes a value from the cache by key.

```go
err := cache.Del[MyStruct](ctx, "my-key")
```

- **Blocking**: Yes

---

## Stores

A **Store** is a logical unit that holds cached values of a certain struct type. Each store has its own configuration (TTL, max size, max record size). If you don't register a store for a type, values are stored in the **default store** (named after your `ServiceName`).

### RegisterStore

Registers a dedicated store for a specific type. Accepts functional options to override default configuration.

```go
err := cache.RegisterStore[MyStruct](
    options.Ttl(30 * time.Minute),
    options.MaxStoreSizeMB(200),
    options.MaxRecordSizeKB(50),
)
```

The store name is automatically derived: `{ServiceName}-{TypeName}`.

- **Blocking**: Yes (requires cache to be initialized)

---

### RegisterSharedStore

Registers a **shared store** that can be accessed across multiple services. Shared stores use the naming pattern `shared-{name}-{TypeName}`, which means different services using the same `name` and type will read/write to the same store.

```go
err := cache.RegisterSharedStore[MyStruct]("users",
    options.ReadOnly(),
    options.Ttl(time.Hour),
)
```

- **Blocking**: Yes (requires cache to be initialized)
- **ReadOnly**: Only applicable to shared stores — prevents `Set` and `Del` operations.

---

### RegisterStoreAsync / RegisterSharedStoreAsync

Async variants that wait for `InitAsync` to complete before registering the store. Returns a channel of errors.

```go
errCh := cache.RegisterStoreAsync[MyStruct](options.Ttl(time.Hour))
errCh := cache.RegisterSharedStoreAsync[MyStruct]("users", options.ReadOnly())
```

- **Blocking**: No (registration happens once init completes)

---

### ClearStore

Clears all entries from a specific store.

```go
err := cache.ClearStore[MyStruct]()
```

---

### ClearCache

Clears all entries from all stores.

```go
err := cache.ClearCache()
```

---

### Store Options

Options follow the functional options pattern:

| Option | Description | Limits |
|--------|-------------|--------|
| `options.Ttl(d)` | Time-to-live for entries in this store | Must be > 0 |
| `options.MaxStoreSizeMB(n)` | Maximum store size in MB | Min: 20, Max: 1000 |
| `options.MaxRecordSizeKB(n)` | Maximum single record size in KB | Max: 2000 |
| `options.ReadOnly()` | Mark store as read-only (shared stores only) | — |

If an option is not provided, defaults are taken from the global configuration.

---

## Distributed Locks

The cache package provides distributed locking built on top of the cache backend. Locks are useful for coordinating work across multiple service instances.

### Lock

Creates a **service-scoped** lock. The lock is stored in a dedicated store named `locks-{ServiceName}`.

```go
l := cache.Lock("my-resource")
err := l.Lock(ctx, 10 * time.Second)  // TTL is required and must be > 0
if err != nil {
    // handle error (context cancelled, lock already held, etc.)
}
defer l.Unlock(ctx)

// critical section
```

---

### SharedLock

Creates a **shared lock** accessible across all services. The lock is stored in a dedicated `locks-shared` store.

```go
l := cache.SharedLock("global-resource")
err := l.Lock(ctx, 30 * time.Second)
if err != nil {
    // handle error
}
defer l.Unlock(ctx)
```

---

## Supported Backends

| Backend | Type constant | Description |
|---------|--------------|-------------|
| **In-Memory** | `inmemory` | Uses [Otter](https://github.com/maypok86/otter) with weight-based eviction and per-entry TTL. No external dependencies. |
| **Redis** | `redis` | Redis hashes per store with `HExpire` for per-field TTL. Supports read/write URL split. |
| **KeyDB** | `keydb` | Redis-compatible; uses `EXPIREMEMBER` for per-field TTL. Supports read/write URL split. |
| **Valkey** | `valkey` | Redis-compatible; TTL managed via deadline embedded in the stored JSON value. Supports read/write URL split. |
| **NATS** | `nats` | JetStream KV buckets per store with memory storage and bucket-level TTL. |

---

## Configuration

All configuration options can be set via environment variables or via the `Init` function. Below is the full list of supported options:

| ENV Variable | Description | Default |
|---|---|---|
| `CACHE_TYPE` | Backend type: `redis`, `nats`, `inmemory`, `keydb`, `valkey` | `inmemory` |
| `CACHE_ENABLED` | Enable or disable the cache | `true` |
| `CACHE_METRICS_ENABLED` | Enable Prometheus metrics collection | `true` |
| `CACHE_URL` | Backend connection URL (not required for `inmemory`) | (required) |
| `CACHE_READ_URL` | Separate URL for read operations (Redis/KeyDB/Valkey) | — |
| `CACHE_USER` | Username for backend authentication | — |
| `CACHE_PASSWORD` | Password for backend authentication | — |
| `CACHE_TTL_SECONDS` | Default TTL for cache entries (seconds) | `3600` |
| `CACHE_MAX_STORE_SIZE_MB` | Default maximum store size in MB | `100` |
| `CACHE_MAX_RECORD_SIZE_KB` | Default maximum single record size in KB | `1` |
| `CACHE_CONNECT_RETRIES` | Number of times to retry connecting to backend at startup | `10` |
| `CACHE_INIT_RETRY_INTERVAL_DURATION` | Interval between connection retries | `2s` |
| `SERVICE_NAME` | Service name (used for store naming and lock ownership) | (required) |

---

## Metrics

When `CACHE_METRICS_ENABLED` is `true`, the cache package automatically tracks and exposes Prometheus metrics for cache operations. Metrics include operation type (get, set, del, etc.), store name, cache hits/misses, and latency. These are registered automatically and can be scraped by any Prometheus-compatible monitoring system.

---

## Example Usage

```go
import (
    "context"
    "time"

    cache "github.com/dsx-ai-factory/run-ai/runai-common-packages/cache/v2"
    cache_config "github.com/dsx-ai-factory/run-ai/runai-common-packages/cache/v2/config"
    "github.com/dsx-ai-factory/run-ai/runai-common-packages/cache/v2/options"
)

type User struct {
    ID   string
    Name string
}

func main() {
    // Optional: set a custom logger before Init
    cache.SetLogger(myLogger)

    // Initialize cache (blocking, will retry as configured)
    err := cache.Init(nil)
    if err != nil {
        panic(err)
    }
    defer cache.Close()

    ctx := context.Background()

    // Register a dedicated store with custom options
    err = cache.RegisterStore[User](
        options.Ttl(30 * time.Minute),
        options.MaxRecordSizeKB(10),
    )
    if err != nil {
        panic(err)
    }

    // Set a value
    err = cache.Set[User](ctx, "user-123", User{ID: "123", Name: "Alice"})

    // Get a value
    user, err := cache.Get[User](ctx, "user-123")
    if user != nil {
        fmt.Println(user.Name) // "Alice"
    }

    // Delete a value
    err = cache.Del[User](ctx, "user-123")

    // Use a distributed lock
    l := cache.Lock("process-user-123")
    if err := l.Lock(ctx, 5 * time.Second); err != nil {
        panic(err)
    }
    defer l.Unlock(ctx)
    // ... critical section ...
}
```

---

## Testing and Mocking

The `testing` sub-package provides helpers for unit testing with the cache.

### SetMock

Injects a mock `Cacher` for testing. Returns a `*MockCacher` (testify mock) that you can set expectations on.

```go
import cache_testing "github.com/dsx-ai-factory/run-ai/runai-common-packages/cache/v2/testing"

func TestMyFunction(t *testing.T) {
    cacheMock := cache_testing.SetMock(t)

    // Expect a Get call and return a value
    cacheMock.On("Get", mock.Anything, cache_testing.KeyFromType[MyStruct](), "my-key").
        Return(&MyStruct{Name: "test"}, nil)

    // Expect a Set call
    cacheMock.On("Set", mock.Anything, cache_testing.KeyFromType[MyStruct](), "my-key", mock.Anything).
        Return(nil)
}
```

### SetLocksMock

Injects a mock `Locker` for testing distributed locks.

```go
lockMock := cache_testing.SetLocksMock(t)
lockMock.On("Lock", mock.Anything, mock.Anything).Return(nil)
lockMock.On("Unlock", mock.Anything).Return(nil)
```

### EnableCache / DisableCache

Test helpers that initialize a real in-memory cache or disable it entirely. Useful for integration-style tests.

```go
func TestWithRealCache(t *testing.T) {
    cache_testing.EnableCache()
    defer cache_testing.DisableCache()
    // ... test code that uses real in-memory cache ...
}
```

---

For more details, see the code and integration tests under `cache/`.
