Node pools assist in managing heterogeneous resources effectively.
A node pool is a set of nodes grouped into a bucket of resources using a predefined (for example, GPU-Type) or
administrator-defined label (for example, key & value).
For more information, see [Node Pools](/run-ai/self-hosted/documentation/platform-management/manage-ai-initiatives/resources/node-pools).

Use the Node Pools endpoints to create, list, retrieve, update, patch, delete, and count node pools.

## Minimal Node Pool Data

The minimal data endpoints return only the essential node pool data needed for common operations, such as workload submission.

## Node Pool Metrics

Retrieve metrics data for a node pool.

## Deprecated Endpoints

The endpoints under `/api/v1/clusters/{clusterId}/node-pools` are deprecated. Use `/api/v1/node-pools` instead.
