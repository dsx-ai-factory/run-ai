# Interacting with Workloads Using the Cluster API

Use the Cluster API for live, cluster-scoped operations on running workloads—checking cluster/workload status, streaming and downloading logs, executing commands, attaching to containers, and forwarding ports. For creating or managing workloads across a tenant, use the [Control plane API](../about-the-rest-api.md).

The API is exposed per cluster at `https://<cluster-fqdn>/cluster-api/`. The interactive endpoints (`exec`, `attach`, `port-forward`, and both `logs` modes) use WebSocket; `status`, `logs/download`, and the cluster `status` endpoint use standard HTTPS. All endpoints require a bearer token (JWT)—refer to [How to authenticate to the API](../how-to-authenticate-to-the-api.md).

## Before You Start

The Cluster API runs on each NVIDIA Run:ai cluster, not on the control plane. Before making a call, you need:

* **The cluster FQDN** - Each cluster exposes the API at `https://<cluster-fqdn>/cluster-api/`. Find the FQDN in the NVIDIA Run:ai UI under _Clusters → your cluster → Connection details_, or from your platform admin.
* **Network reachability to the cluster** - The Cluster API endpoint must be reachable from the machine making the call. If your cluster is behind a VPN or private network, you must be on that network.
* **A bearer token (JWT) with the right permissions** - All endpoints require a JWT. The token must be associated with a user or application that has permission to access the project and workload in question. Refer to [How to authenticate to the API](../how-to-authenticate-to-the-api.md) for how to obtain a token.
* **Valid project and workload identifiers** - Most endpoints require the project name, workload type, workload framework, and workload name in the path. The identifiers must match an existing, running workload in that project. Refer to [Common path parameters](interacting-with-workloads-using-the-cluster-api.md#common-path-parameters) below.

## Common Path Parameters

Most endpoints take the same four path parameters. They are documented once here and referenced by each endpoint below.

| Parameter           | Required | Description                                                                                                                                                                                                                                                                                                                                  | Example    |
| ------------------- | -------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------- |
| `projectName`       | Required | The NVIDIA Run:ai project name that owns the workload.                                                                                                                                                                                                                                                                                       | `project1` |
| `workloadType`      | Required | The workload type. One of: `workspace`, `training`, `distributed`, `inference`, `external`.                                                                                                                                                                                                                                                  | `training` |
| `workloadFramework` | Required | The workload framework. One of: `runai`, `mpi`, `pytorch`, `tf`, `xgboost`, `any`, `external`. Use `external` together with `workloadType=external` for workloads created from third-party CRDs (refer to the CRD-mapping table below). The `any` value does **not** route to anyworkload-managed third-party CRDs—use `external` for those. | `pytorch`  |
| `workloadName`      | Required | The name of the workload, as it appears in the NVIDIA Run:ai UI or CLI. For workloads created from third-party CRDs, refer to the CRD-mapping table below.                                                                                                                                                                                   | `my-jp`    |

### CRD to Path Parameter Mapping

The cluster-api resolves the path parameters to an NVIDIA Run:ai V2 CRD inside the project namespace (`runai-{project}`). Use the table below to pick the right `workloadType`, `workloadFramework`, and `workloadName` for your workload.

| CRD                                                 | Workload type | Workload framework | Workload name                           |
| --------------------------------------------------- | ------------- | ------------------ | --------------------------------------- |
| `InteractiveWorkload` (NVIDIA Run:ai workspace)     | `workspace`   | `runai`            | The workload's name as shown in UI/CLI. |
| `TrainingWorkload` (NVIDIA Run:ai training, single) | `training`    | `runai`            | The workload's name as shown in UI/CLI. |
| `DistributedWorkload` (PyTorch)                     | `distributed` | `pytorch`          | The workload's name as shown in UI/CLI. |
| `DistributedWorkload` (TensorFlow)                  | `distributed` | `tf`               | The workload's name as shown in UI/CLI. |
| `DistributedWorkload` (MPI)                         | `distributed` | `mpi`              | The workload's name as shown in UI/CLI. |
| `DistributedWorkload` (XGBoost)                     | `distributed` | `xgboost`          | The workload's name as shown in UI/CLI. |
| `InferenceWorkload`                                 | `inference`   | `runai`            | The workload's name as shown in UI/CLI. |
| `ExternalWorkload` (third-party `PyTorchJob`)       | `external`    | `external`         | `ew-pytorchjob-{originalName}`          |
| `ExternalWorkload` (third-party `TFJob`)            | `external`    | `external`         | `ew-tfjob-{originalName}`               |
| `ExternalWorkload` (third-party `MPIJob`)           | `external`    | `external`         | `ew-mpijob-{originalName}`              |
| `ExternalWorkload` (third-party `LeaderWorkerSet`)  | `external`    | `external`         | `ew-leaderworkerset-{originalName}`     |
| `ExternalWorkload` (any other third-party CRD)      | `external`    | `external`         | `ew-{kind-lowercase}-{originalName}`    |

<Info>

**Tip**

For finding the right name:

* For NVIDIA Run:ai-native workloads (workspace / training / distributed / inference): the name in the UI matches the `workloadName` directly.
* For third-party CRDs: run `kubectl get externalworkloads -n runai-{project}` to list the auto-generated `ew-…` names, or look at the workload entry in the NVIDIA Run:ai UI.

</Info>

<Info>

**Note**

`GET /cluster-api/status` is cluster-scoped and does not take these parameters. All other endpoints are project- and workload-scoped.

</Info>

## Request and Response Conventions

### Base URL

Replace `<cluster-fqdn>` with your cluster's FQDN in every example:

```
https://<cluster-fqdn>/cluster-api/
```

Throughout this guide, examples use `$CLUSTER` for the FQDN and `$TOKEN` for the bearer token:

```bash
CLUSTER="mycluster.run.ai"
TOKEN="eyJhbGciOi..."
```

### Authentication

Every request must include the bearer token in the `Authorization` header:

```
Authorization: Bearer $TOKEN
```

For WebSocket endpoints (`exec`, `attach`, `port-forward`, `logs`), the same header is sent during the WebSocket upgrade handshake.

### Content Types

* JSON endpoints return `application/json`.
* The `logs/download` endpoint returns `application/octet-stream`.
* WebSocket log frames carry log-line payloads as binary messages.

## Connection Behavior

The following endpoints use WebSocket connections and have different semantics from standard REST endpoints:

* `logs` (both one-shot and `follow=true` modes)
* `exec`
* `attach`
* `port-forward`

WebSocket examples in this guide use [`websocat`](https://github.com/vi/websocat). Install with `cargo install websocat` or `brew install websocat`. Equivalent calls work from any WebSocket-capable client (Python `websockets`, Node `ws`, Go `gorilla/websocket`).

### Upgrade Flow

Clients issue a `GET` request with the standard WebSocket upgrade headers (`Connection: Upgrade`, `Upgrade: websocket`, `Sec-WebSocket-Key`, `Sec-WebSocket-Version: 13`). The bearer token is passed in `Authorization` on the handshake request.

On success, the server responds with `101 Switching Protocols`. On authentication failure, the server responds with `401` before upgrading.

### Keep-Alive and Timeouts

* The server sends periodic ping frames every 30 seconds. Clients must respond with pong frames to keep the connection alive.
* Idle connections (no traffic for 5 minutes) are closed by the server.
* Long-running sessions (`exec`, `attach`) have no maximum duration as long as pings are acknowledged.

### Termination

The connection closes when:

* The client closes the WebSocket.
* The target pod or container terminates.
* The bearer token expires (server closes with code `1008`).
* Idle timeout is reached.

Clients should handle abrupt disconnection and reconnect if needed. No replay is provided—on reconnect, only new log lines or output are delivered.

## Errors

All endpoints return standard HTTP status codes. Error responses have the following JSON shape:

```json
{
  "code": 403,
  "reason": "Forbidden",
  "message": "user does not have permission to access project 'project1'"
}
```

### Common Error Codes

| Code  | Meaning               | Typical cause                                                                                                                                                                                                                                                                                                                                                                                                                                  |
| ----- | --------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `400` | Bad Request           | Invalid query parameter (for example, `sinceSeconds=-1`), malformed path parameter, or missing required query parameter.                                                                                                                                                                                                                                                                                                                       |
| `401` | Unauthorized          | Bearer token missing, malformed, or expired.                                                                                                                                                                                                                                                                                                                                                                                                   |
| `403` | Forbidden             | Token is valid but the associated user lacks permission to access the project or workload.                                                                                                                                                                                                                                                                                                                                                     |
| `404` | Not Found             | Project, pod, or container does not exist. Also returned when a workload exists but is not yet scheduled (no pods).                                                                                                                                                                                                                                                                                                                            |
| `409` | Conflict              | Operation cannot be performed in the workload's current state (for example, `exec` on a pod that is terminating).                                                                                                                                                                                                                                                                                                                              |
| `500` | Internal Server Error | Unexpected cluster-side failure. Also currently returned (with a plain-text body) when the requested workload cannot be located in the cluster—verify that `projectName`, `workloadType`, `workloadFramework`, and `workloadName` match an existing workload (refer to [CRD to Path Parameter Mapping](interacting-with-workloads-using-the-cluster-api.md#crd-to-path-parameter-mapping)). Retry with backoff. |
| `503` | Service Unavailable   | Cluster is temporarily unreachable, the cluster-api service is restarting, or the requested operation cannot complete (for example, `attach` to a container whose primary process already holds stdio).                                                                                                                                                                                                                                        |

For WebSocket endpoints, authentication and permission errors are returned on the HTTP handshake (before upgrade). Errors during an active stream are reported through WebSocket close codes:

| Close code | Meaning                                                  |
| ---------- | -------------------------------------------------------- |
| `1000`     | Normal closure (target terminated cleanly)               |
| `1001`     | Going away (server shutdown or pod terminated)           |
| `1002`     | Protocol error (malformed frames)                        |
| `1006`     | Abnormal closure (network drop, no close frame received) |
| `1008`     | Policy violation (token expired mid-stream)              |
| `1011`     | Server error (unexpected cluster-side failure)           |

## Endpoints

### Cluster Status

| Method | Connection | Auth         | Idempotent | Side effects | Response           |
| ------ | ---------- | ------------ | ---------- | ------------ | ------------------ |
| `GET`  | HTTPS      | Bearer (JWT) | Yes        | None         | `application/json` |

Returns the cluster status as a JSON object, including the version and the state of features within the cluster.

```
GET https://<cluster-fqdn>/cluster-api/status
```

**Example**

```bash
curl -H "Authorization: Bearer $TOKEN" \
  "https://$CLUSTER/cluster-api/status"
```

Response:

```json
{
  "version": "1.0",
  "features": {
    "message_enabled": true,
    "message_resize_enabled": true
  }
}
```

### Workload Status

| Method | Connection | Auth         | Idempotent | Side effects | Response           |
| ------ | ---------- | ------------ | ---------- | ------------ | ------------------ |
| `GET`  | HTTPS      | Bearer (JWT) | Yes        | None         | `application/json` |

Returns the workload status as a JSON object with a `status` field carrying the workload phase.

```
GET https://<cluster-fqdn>/cluster-api/api/v1/{projectName}/workloads/{workloadType}/{workloadFramework}/{workloadName}/status
```

**Path Parameters**: `projectName`, `workloadType`, `workloadFramework`, `workloadName`—refer to [Common path parameters](interacting-with-workloads-using-the-cluster-api.md#common-path-parameters) for values and the CRD-mapping table.

**Example**

```bash
curl -H "Authorization: Bearer $TOKEN" \
  "https://$CLUSTER/cluster-api/api/v1/project1/workloads/training/pytorch/my-jp/status"
```

Response:

```json
{
  "status": "Running"
}
```

### Logs

| Method | Connection | Auth         | Idempotent | Side effects | Response                |
| ------ | ---------- | ------------ | ---------- | ------------ | ----------------------- |
| `GET`  | WebSocket  | Bearer (JWT) | Yes        | None         | WebSocket binary frames |

Read the logs of a container in a pod. If the pod has multiple containers, the container is auto-selected if the container name is not provided.

```
GET wss://<cluster-fqdn>/cluster-api/api/v1/{projectName}/workloads/{workloadType}/{workloadFramework}/{workloadName}/logs
```

**The `logs` endpoint is WebSocket-only**, in both one-shot and follow modes. Plain HTTPS GET returns `HTTP 426 Upgrade Required`. Each log line is delivered as a binary WebSocket message. In one-shot mode (`follow=false` or omitted), the server closes the connection with code `1000` after delivering the requested range; in follow mode, the connection stays open and new lines are pushed as they arrive.

**Path Parameters**: `projectName`, `workloadType`, `workloadFramework`, `workloadName`—refer to [Common path parameters](interacting-with-workloads-using-the-cluster-api.md#common-path-parameters).

**Query Parameters**

| Field          | Required | Description                                                                                              | Example                |
| -------------- | -------- | -------------------------------------------------------------------------------------------------------- | ---------------------- |
| `pod`          | Optional | The name of the pod                                                                                      | `my-pod`               |
| `container`    | Optional | The container to use. Defaults to the only container if there is one.                                    | `my-c`                 |
| `follow`       | Optional | Keep the connection open and stream new log lines. Defaults to `false`.                                  | `false`                |
| `limitBytes`   | Optional | Number of bytes to read from the server before terminating log output.                                   | `200`                  |
| `previous`     | Optional | Return previous terminated container logs. Defaults to `false`.                                          | `false`                |
| `sinceSeconds` | Optional | Return logs newer than a relative duration (for example, `5s`, `2m`, `3h`). Defaults to `0s` (all logs). | `5s`                   |
| `timestamps`   | Optional | Include timestamps on each line in the log output. Defaults to `false`.                                  | `false`                |
| `sinceTime`    | Optional | Return logs after a specific date (RFC3339).                                                             | `2024-01-01T00:00:00Z` |
| `tailLines`    | Optional | Number of recent log lines to display. Defaults to `-1` (all lines).                                     | `2`                    |

**Example (one-shot)**

```bash
websocat -H "Authorization: Bearer $TOKEN" \
  "wss://$CLUSTER/cluster-api/api/v1/project1/workloads/training/pytorch/my-jp/logs?tailLines=200&timestamps=true"
```

**Example (follow / stream)**

```bash
websocat -H "Authorization: Bearer $TOKEN" \
  "wss://$CLUSTER/cluster-api/api/v1/project1/workloads/training/pytorch/my-jp/logs?follow=true&sinceSeconds=30"
```

### Logs Download

| Method | Connection | Auth         | Idempotent | Side effects | Response                   |
| ------ | ---------- | ------------ | ---------- | ------------ | -------------------------- |
| `GET`  | HTTPS      | Bearer (JWT) | Yes        | None         | `application/octet-stream` |

Download the logs for a container in a pod as a file. If the pod has multiple containers, the container is auto-selected if the container name is not provided.

```
GET https://<cluster-fqdn>/cluster-api/api/v1/{projectName}/workloads/{workloadType}/{workloadFramework}/{workloadName}/logs/download
```

**Path Parameters**: `projectName`, `workloadType`, `workloadFramework`, `workloadName`—refer to [Common path parameters](interacting-with-workloads-using-the-cluster-api.md#common-path-parameters).

**Query Parameters**

| Field          | Required | Description                                                                                              | Example                |
| -------------- | -------- | -------------------------------------------------------------------------------------------------------- | ---------------------- |
| `pod`          | Optional | The name of the pod                                                                                      | `my-pod`               |
| `container`    | Optional | The container to use. Defaults to the only container if there is one.                                    | `my-c`                 |
| `limitBytes`   | Optional | Number of bytes to read from the server before terminating log output.                                   | `200`                  |
| `previous`     | Optional | Return previous terminated container logs. Defaults to `false`.                                          | `false`                |
| `sinceSeconds` | Optional | Return logs newer than a relative duration (for example, `5s`, `2m`, `3h`). Defaults to `0s` (all logs). | `5s`                   |
| `timestamps`   | Optional | Include timestamps on each line in the log output. Defaults to `false`.                                  | `false`                |
| `sinceTime`    | Optional | Return logs after a specific date (RFC3339).                                                             | `2024-01-01T00:00:00Z` |
| `tailLines`    | Optional | Number of recent log lines to display. Defaults to `-1` (all lines).                                     | `2`                    |

**Example**

```bash
curl -H "Authorization: Bearer $TOKEN" \
  -o my-jp.log \
  "https://$CLUSTER/cluster-api/api/v1/project1/workloads/training/pytorch/my-jp/logs/download?timestamps=true"
```

### Attach

| Method | Connection | Auth         | Idempotent | Side effects                                        | Response                |
| ------ | ---------- | ------------ | ---------- | --------------------------------------------------- | ----------------------- |
| `GET`  | WebSocket  | Bearer (JWT) | No         | Opens a stdio session against the running container | WebSocket binary frames |

Attach to a container in a pod.

```
GET wss://<cluster-fqdn>/cluster-api/api/v1/{projectName}/workloads/{workloadType}/{workloadFramework}/{workloadName}/attach
```

<Info>

**Tip**

Attaching to a container whose primary process is already consuming stdio (for example, an actively-running training job) can fail with `503`. Use `exec` to start a new shell in the same container instead.

</Info>

**Path Parameters**: `projectName`, `workloadType`, `workloadFramework`, `workloadName`—refer to [Common path parameters](interacting-with-workloads-using-the-cluster-api.md#common-path-parameters).

**Query Parameters**

| Field       | Required | Description                                                           | Example  |
| ----------- | -------- | --------------------------------------------------------------------- | -------- |
| `pod`       | Optional | The name of the pod                                                   | `my-pod` |
| `container` | Optional | The container to use. Defaults to the only container if there is one. | `my-c`   |
| `stderr`    | Optional | Redirect stderr for the attach call. Defaults to `true`.              | `true`   |
| `stdin`     | Optional | Redirect standard input stream of the pod. Defaults to `false`.       | `true`   |
| `stdout`    | Optional | Redirect stdout for the attach call. Defaults to `true`.              | `true`   |
| `tty`       | Optional | Allocate a TTY for the attach call. Defaults to `false`.              | `false`  |

**Example**

```bash
websocat -H "Authorization: Bearer $TOKEN" \
  "wss://$CLUSTER/cluster-api/api/v1/project1/workloads/training/pytorch/my-jp/attach?tty=true&stdin=true&stdout=true&stderr=true"
```

### Port Forward

| Method | Connection | Auth         | Idempotent | Side effects                              | Response                |
| ------ | ---------- | ------------ | ---------- | ----------------------------------------- | ----------------------- |
| `GET`  | WebSocket  | Bearer (JWT) | No         | Opens a TCP forwarding session to the pod | WebSocket binary frames |

Port forward to the specified pod in the given workload.

```
GET wss://<cluster-fqdn>/cluster-api/api/v1/{projectName}/workloads/{workloadType}/{workloadFramework}/{workloadName}/port-forward
```

To forward more than one port, repeat the `port` query key (for example `?port=8090&port=8080`). Comma-separated lists are not accepted.

**Path Parameters**: `projectName`, `workloadType`, `workloadFramework`, `workloadName`—refer to [Common path parameters](interacting-with-workloads-using-the-cluster-api.md#common-path-parameters).

**Query Parameters**

| Field       | Required | Description                                                                 | Example  |
| ----------- | -------- | --------------------------------------------------------------------------- | -------- |
| `pod`       | Optional | The name of the pod                                                         | `my-pod` |
| `container` | Optional | The container to use. Defaults to the only container if there is one.       | `my-c`   |
| `port`      | Required | Port to forward to the pod. Repeat the parameter to forward multiple ports. | `8090`   |

**Example (single port)**

```bash
websocat -H "Authorization: Bearer $TOKEN" \
  "wss://$CLUSTER/cluster-api/api/v1/project1/workloads/training/pytorch/my-jp/port-forward?port=8090"
```

**Example (multiple ports)**

```bash
websocat -H "Authorization: Bearer $TOKEN" \
  "wss://$CLUSTER/cluster-api/api/v1/project1/workloads/training/pytorch/my-jp/port-forward?port=8090&port=8080"
```

### Exec

| Method | Connection | Auth         | Idempotent | Side effects                              | Response                |
| ------ | ---------- | ------------ | ---------- | ----------------------------------------- | ----------------------- |
| `GET`  | WebSocket  | Bearer (JWT) | No         | Spawns a new process inside the container | WebSocket binary frames |

Execute a command in a container.

```
GET wss://<cluster-fqdn>/cluster-api/api/v1/{projectName}/workloads/{workloadType}/{workloadFramework}/{workloadName}/exec
```

The command can be passed in two equivalent forms:

* as the `command` query parameter (URL-encoded base64 of a JSON argv array), or
* as the `Command` HTTP header on the WebSocket upgrade (raw base64 of a JSON argv array, no URL-encoding needed).

For example, `["/bin/bash"]` base64-encodes to `WyIvYmluL2Jhc2giXQ==`.

**Path Parameters**: `projectName`, `workloadType`, `workloadFramework`, `workloadName`—refer to [Common path parameters](interacting-with-workloads-using-the-cluster-api.md#common-path-parameters).

**Query Parameters**

| Field       | Required | Description                                                                             | Example                    |
| ----------- | -------- | --------------------------------------------------------------------------------------- | -------------------------- |
| `pod`       | Optional | The name of the pod                                                                     | `my-pod`                   |
| `container` | Optional | The container to use. Defaults to the only container if there is one.                   | `my-c`                     |
| `stderr`    | Optional | Redirect stderr for the exec call. Defaults to `true`.                                  | `true`                     |
| `stdin`     | Optional | Redirect standard input stream of the pod. Defaults to `false`.                         | `true`                     |
| `stdout`    | Optional | Redirect stdout for the exec call. Defaults to `true`.                                  | `true`                     |
| `tty`       | Optional | Allocate a TTY for the exec call. Defaults to `false`.                                  | `false`                    |
| `command`   | Required | Command to execute, either as a query string or base64-encoded in the `Command` header. | `WyIvYmluL2Jhc2giXQ%3D%3D` |

**Example (`Command` header)**

```bash
websocat \
  -H "Authorization: Bearer $TOKEN" \
  -H "Command: WyIvYmluL2Jhc2giXQ==" \
  "wss://$CLUSTER/cluster-api/api/v1/project1/workloads/training/pytorch/my-jp/exec?tty=true&stdin=true&stdout=true&stderr=true"
```

**Example (query parameter)**

```bash
websocat -H "Authorization: Bearer $TOKEN" \
  "wss://$CLUSTER/cluster-api/api/v1/project1/workloads/training/pytorch/my-jp/exec?command=WyIvYmluL2Jhc2giXQ%3D%3D&tty=true&stdin=true"
```

### Fetch Workload Object

<Info>

**Note**

The following endpoints require **cluster v2.26 or later**.

</Info>

These endpoints return a point-in-time snapshot of the Kubernetes object for a workload, one of its child elements, or one of its pods, as read directly from the Kubernetes API server. The response includes the full `spec`, `status`, and metadata; `metadata.managedFields` and the `last-applied-configuration` annotation are stripped. Use these endpoints for inspection and debugging -- for example, to see why a component is degraded or to confirm the exact spec your workload produced after reconciliation.

These endpoints use the **v2 path structure**, which identifies workloads by UUID rather than by project name and workload type. Obtain the `workloadId`, `elementId`, and `podId` values from the [Control Plane API](../about-the-rest-api.md).

#### Get Workload Object

| Method | Connection | Auth         | Idempotent | Side effects | Response           |
| ------ | ---------- | ------------ | ---------- | ------------ | ------------------ |
| `GET`  | HTTPS      | Bearer (JWT) | Yes        | None         | `application/json` |

Returns the Kubernetes custom resource of the top-level workload element.

```
GET https://<cluster-fqdn>/cluster-api/api/v2/workloads/{workloadId}/object
```

**Path Parameters**

| Parameter    | Required | Description                                        | Example                                |
| ------------ | -------- | -------------------------------------------------- | -------------------------------------- |
| `workloadId` | Required | The unique identifier (UUID) of the workload.      | `8f3c4a2e-1b5d-4f7a-9c3e-2a8b1c4d5e6f` |

**Example**

```bash
curl -H "Authorization: Bearer $TOKEN" \
  "https://$CLUSTER/cluster-api/api/v2/workloads/8f3c4a2e-1b5d-4f7a-9c3e-2a8b1c4d5e6f/object"
```

Response:

```json
{
  "object": {
    "apiVersion": "nvidia.com/v1alpha1",
    "kind": "DynamoGraphDeployment",
    "metadata": {
      "name": "rag-pipeline-prod",
      "namespace": "runai-research-team",
      "uid": "8f3c4a2e-1b5d-4f7a-9c3e-2a8b1c4d5e6f"
    },
    "spec": { ... },
    "status": { ... }
  },
  "retrievedAt": "2026-06-07T10:22:14Z"
}
```

#### Get Element Object

| Method | Connection | Auth         | Idempotent | Side effects | Response           |
| ------ | ---------- | ------------ | ---------- | ------------ | ------------------ |
| `GET`  | HTTPS      | Bearer (JWT) | Yes        | None         | `application/json` |

Returns the Kubernetes object for a specific child element within a workload.

```
GET https://<cluster-fqdn>/cluster-api/api/v2/workloads/{workloadId}/elements/{elementId}/object
```

**Path Parameters**

| Parameter    | Required | Description                                                 | Example                                |
| ------------ | -------- | ----------------------------------------------------------- | -------------------------------------- |
| `workloadId` | Required | The unique identifier (UUID) of the workload.               | `8f3c4a2e-1b5d-4f7a-9c3e-2a8b1c4d5e6f` |
| `elementId`  | Required | The unique identifier (UUID) of the element within the workload. | `a1b2c3d4-5e6f-7890-abcd-ef1234567890` |

**Example**

```bash
curl -H "Authorization: Bearer $TOKEN" \
  "https://$CLUSTER/cluster-api/api/v2/workloads/8f3c4a2e-1b5d-4f7a-9c3e-2a8b1c4d5e6f/elements/a1b2c3d4-5e6f-7890-abcd-ef1234567890/object"
```

Response:

```json
{
  "object": {
    "apiVersion": "apps/v1",
    "kind": "Deployment",
    "metadata": {
      "name": "rag-pipeline-prod-backend",
      "namespace": "runai-research-team"
    },
    "spec": { ... },
    "status": { ... }
  },
  "retrievedAt": "2026-06-07T10:22:31Z"
}
```

#### Get Pod Object

| Method | Connection | Auth         | Idempotent | Side effects | Response           |
| ------ | ---------- | ------------ | ---------- | ------------ | ------------------ |
| `GET`  | HTTPS      | Bearer (JWT) | Yes        | None         | `application/json` |

Returns the Kubernetes Pod object for a specific pod within a workload.

```
GET https://<cluster-fqdn>/cluster-api/api/v2/workloads/{workloadId}/pods/{podId}/object
```

**Path Parameters**

| Parameter    | Required | Description                                              | Example                                |
| ------------ | -------- | -------------------------------------------------------- | -------------------------------------- |
| `workloadId` | Required | The unique identifier (UUID) of the workload.            | `8f3c4a2e-1b5d-4f7a-9c3e-2a8b1c4d5e6f` |
| `podId`      | Required | The unique identifier (UUID) of the pod within the workload. | `b2c3d4e5-6f70-8901-bcde-f12345678901` |

**Example**

```bash
curl -H "Authorization: Bearer $TOKEN" \
  "https://$CLUSTER/cluster-api/api/v2/workloads/8f3c4a2e-1b5d-4f7a-9c3e-2a8b1c4d5e6f/pods/b2c3d4e5-6f70-8901-bcde-f12345678901/object"
```

Response:

```json
{
  "object": {
    "apiVersion": "v1",
    "kind": "Pod",
    "metadata": {
      "name": "rag-pipeline-prod-backend-7c9f4b-jx2hp",
      "namespace": "runai-research-team"
    },
    "spec": { ... },
    "status": { ... }
  },
  "retrievedAt": "2026-06-07T10:22:45Z"
}
```
