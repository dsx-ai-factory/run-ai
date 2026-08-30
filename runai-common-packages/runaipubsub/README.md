# `runaipubsub` Package Documentation

The `runaipubsub` package provides a unified, extensible, and production-grade publish/subscribe (Pub/Sub) abstraction for Run.ai services. It enables event-driven communication between microservices, supporting multiple backend integrations such as Redis, NATS and Kafka.

---

## Features & Public API


### Initialization (`Init`)

Initializes the pubsub system and connects to the configured backend(s). This is a **blocking** operation: the function will retry connecting according to `PUBSUB_CONNECT_RETRIES` and will not return until the connection is established or all retries fail.

You can pass a configuration struct to `Init` to programmatically set options instead of (or in addition to) using environment variables. This allows for dynamic configuration at runtime.

> **Note:** As a general practice, it is usually preferred to allow the package to load the configuration by itself (using environment variables) rather than providing a config struct to the `Init` function. This ensures a consistent configuration standard across all services using this package.

**Example:**

```go
cfg := runaipubsub.Config{
    Type: "redis",
    URL:  "redis://localhost:6379",
    // ...other fields...
}
err := runaipubsub.Init(&cfg)
if err != nil {
    panic(err)
}
```

See the [Configuration](#configuration) section below for the full list of available settings.

Pass nil to load defaults from env vars.
```go
err := runaipubsub.Init(nil)
if err != nil {
    panic(err)
}
```

- **Blocking**: Yes (retries connection on startup)
- **Retry:** Uses exponential backoff, starting from 1 second, controlled by config.

---

### Async Initialization (`InitAsync`)

Initializes the pubsub system asynchronously. Unlike `Init`, this method **does not block**; it immediately returns a channel of errors. The connection process happens in the background, and any errors encountered during initialization are sent to the returned error channel.
Errors will also be printed by the provided logger `SetLogger()` if provided.

This is useful for services that want to start up quickly and handle pubsub connection errors reactively, rather than blocking startup. Typically, you will use `InitAsync` in combination with the `SubscribeAsync` method to fully handle pubsub events and errors in a non-blocking fashion.


**Example:**

```go
// The error channel returned from InitAsync and SubscribeAsync is the same shared channel.
errChan, err := runaipubsub.InitAsync(nil)
if err != nil {
    // if this error happens, the pubsub will never connect - decide what to do in this situation
}
runaipubsub.SubscribeAsync[MyEvent](func(ctx context.Context, event *MyEvent) error {
    // handle event
    return nil
})

// Listen for errors from both async init and async subscribe in a single goroutine
go func() {
    for err := range errChan {
        if err != nil {
            // print err or take any action if needed
        }
    }
}()
```

> **Note:** The error channel is shared for both async initialization and async subscription. You only need to listen on it once to receive all errors.

- **Blocking**: No (returns immediately)
- **Error Channel:** All errors are sent to the returned channel. You can choose to monitor or ignore it.
- **Typical Usage:** Combine with `SubscribeAsync` for fully asynchronous pubsub handling.

---

### SetLogger

Sets a custom logger for the pubsub system. Should be called before `Init`. If not set, a default dummy logger is used (no logs).

```go
runaipubsub.SetLogger(myLogger)
```

- **Blocking**: No

---

### Publishing Events (`Publish`)

Publishes an event to the configured backend. Retries are handled in the background according to `PUBSUB_PUBLISH_RETRIES` and `PUBSUB_PUBLISH_RETRY_INTERVAL_SEC`.

```go
err := runaipubsub.Publish(ctx, &event)
```

- **Blocking**: Yes (returns after publish attempt, retries in background if configured)
- **Retry:** Retries use a fixed interval, controlled by config.

---

### Subscribing to Events (`Subscribe`)

Subscribes to a topic and registers a callback for incoming events. This is a **blocking** operation: the function will retry subscribing according to `PUBSUB_SUBSCRIPTION_INIT_RETRIES` and will not return until the subscription is established or all retries fail. Also automatically subscribes to an additional pubsuber if configured.

```go
err := runaipubsub.Subscribe[MyEvent](ctx, func(ctx context.Context, event *MyEvent) error {
    // handle event
    return nil
})
```

- **Blocking**: Yes (retries subscription on startup)
- **Retry:** Uses exponential backoff, starting from 1 second, controlled by config.

---

### Subscribing to Events Asynchronously (`SubscribeAsync`)

Initiates an **asynchronous** subscription to a topic and returns a channel that receives subscription errors (if any). The subscription is established in a background goroutine, and the caller is responsible for monitoring the returned error channel. Also automatically subscribes to an additional pubsuber if configured.

```go
errChan := runaipubsub.SubscribeAsync[MyEvent](ctx, func(ctx context.Context, event *MyEvent) error {
    // handle event
    return nil
})

// Optional: observe subscription errors (already logged internally)
for err := range errChan {
    log.Errorf("Subscription failed: %v", err)
}
```

- **Blocking**: No (subscription happens in a background goroutine).
- **Retry:** Retries immediately on failure (infinite loop), unless ctx is canceled.
- **Cancellation:** Respects `ctx.Done()`.
- **Error Channel:**  Buffered (size 1). Can be ignored — errors are logged.

---

### Closing (`Close`)

Gracefully closes all pubsub connections and cleans up resources. This is a blocking operation and should be called on service shutdown.

```go
err := runaipubsub.Close()
```

- **Blocking**: Yes

---

### Mocking (`SetMock`)

Allows injecting a mock PubSuber for testing purposes. This is useful for unit tests and local development.

```go
runaipubsub.SetMock(myMockPubSuber)
```

---

### GetPubSuber

Returns the current primary PubSuber instance. Useful for advanced scenarios or testing.

```go
ps := runaipubsub.GetPubSuber()
```

---

### GetAdditionalPubSuber

Returns the current additional PubSuber instance (if enabled/configured). Useful for advanced scenarios or testing.

```go
ps := runaipubsub.GetAdditionalPubSuber()
```

---

## Supported Integrations

The following integrations are implemented and can be found as subfolders under this package:

- **redis**: High-performance in-memory pub/sub and stream support
- **nats**: Lightweight, high-throughput messaging system
- **kafka**: Distributed event streaming platform
- **mocks**: For testing and local development
- (See the `runaipubsub/` directory for additional or custom integrations)

---

## Configuration

All configuration options can be set via environment variables or via the init function. Below is a list of all supported options, their ENV var names, and a short description including whether the option is blocking (affects startup/init) or background (affects runtime behavior).

| ENV Variable                          | Description                                                                 | Default         |
|----------------------------------------|-----------------------------------------------------------------------------|-----------------|
| PUBSUB_TYPE                           | Backend type: `redis`, `nats`, `kafka`                                      | (required)      |
| PUBSUB_URL                            | Backend connection URL                                                      | (required)      |
| PUBSUB_USER                           | Username for backend (if needed)                                            |                 |
| PUBSUB_PASSWORD                       | Password for backend (if needed)                                            |                 |
| PUBSUB_ENABLE_TLS                     | Enable TLS for backend connection                                          | true            |
| PUBSUB_SUBSCRIBE_RETRIES              | Number of times to retry a failed subscribe                                 | 1               |
| PUBSUB_PUBLISH_RETRIES                | Number of times to retry a failed publish                                   | 3               |
| PUBSUB_CONNECT_RETRIES                | Number of times to retry connecting to backend (at startup)                 | 3               |
| PUBSUB_SUBSCRIPTION_INIT_RETRIES      | Retries for initializing a subscription                                     | 5               |
| PUBSUB_ENABLE_DLQ                     | Enable dead-letter queue                                                    | true            |
| PUBSUB_WAIT_FOR_ACK_SEC               | Seconds to wait for message ack                                             | 180             |
| PUBSUB_PUBLISH_RETRY_INTERVAL_SEC     | Interval (sec) between publish retries                                      | 1               |
| PUBSUB_RETRY_ON_REJECT                | Retry publish if message is rejected                                        | false           |
| PUBSUB_CONNECT_TIMEOUT_SECONDS        | Timeout (sec) for connecting to backend                                     | 2               |
| PUBSUB_CLEANER_INTERVAL_SECONDS       | Interval (sec) for cleaning up old messages                                 | 60              |
| PUBSUB_CONN_MAX_IDLE_TIME_SEC         | Max idle time (sec) for backend connection                                  | 300             |
| PUBSUB_CONSUMER_TIMEOUT_SEC           | Idle timeout (sec) for consumers                                            | 15552000        |
| PUBSUB_MAX_ENTRIES                    | Max number of entries in a topic/stream                                     | 500000          |
| PUBSUB_MAX_MESSAGE_AGE_MINUTES        | Max message age (minutes)                                                   | 10080           |
| PUBSUB_REPLICAS                       | Number of replicas for backend topic/stream                                 | 3               |
| SERVICE_NAME                          | Service name for connection ownership                                       | (required)      |
| PUBSUB_ENABLE_ADDITIONAL_SUBSCRIBER   | Enable a secondary/additional subscriber                                    | false           |
| PUBSUB_ADDITIONAL_SUBSCRIBER_TYPE     | Type for additional subscriber                                              |                 |
| PUBSUB_ADDITIONAL_SUBSCRIBER_URL      | URL for additional subscriber                                               |                 |
| PUBSUB_ADDITIONAL_SUBSCRIBER_USER     | User for additional subscriber                                              |                 |
| PUBSUB_ADDITIONAL_SUBSCRIBER_PASSWORD | Password for additional subscriber                                          |                 |

---

## Metrics

The `runaipubsub` package automatically tracks and exposes metrics for pubsub event processing. These metrics are useful for monitoring system health, debugging, and alerting. The following metric is created:

- **pubsub_event_count** (CounterVec):
  - **Description:** Count of pubsub events processed by subscriber.
  - **Labels:**
    - `event_type`: The topic or event type processed.
    - `error`: Indicates if the event processing resulted in an error (`true` or `false`).

The metric is incremented for every event processed, with the appropriate labels. This allows you to monitor the number of events handled and track error rates per event type.

Metrics are registered automatically and can be scraped by any Prometheus-compatible monitoring system if the metrics package is configured accordingly.

---

## Example Usage

```go
import "github.com/run-ai/runai/runai-common-packages/runaipubsub"

func main() {
    // Initialize pubsub (blocking, will retry connection as configured)
    err := runaipubsub.Init(nil)
    if err != nil {
        panic(err)
    }

    // Publish an event
    event := MyEvent{...}
    err = runaipubsub.Publish(context.Background(), &event)
    if err != nil {
        // handle error
    }

    // Subscribe to events
    runaipubsub.Subscribe[MyEvent](context.Background(), func(ctx context.Context, event *MyEvent) error {
        // handle event
        return nil
    })
}
```

---

## Testing & Mocks

- Use the `mocks` integration for unit testing.
- You can inject a mock PubSuber using `SetMock` for full control in tests.

---

For more details, see the code and integration folders under `runaipubsub/`.
