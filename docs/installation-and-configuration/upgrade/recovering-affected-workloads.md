# Recovering Affected Workloads After Upgrade

This guide explains how to recreate legacy NVIDIA Run:ai workloads through the Workloads V2 submission path, to restore the capabilities described in [Post-Upgrade Recovery for Affected Workloads](upgrade.md#post-upgrade-recovery-for-affected-workloads).

The procedure uses only NVIDIA Run:ai control-plane APIs. You do not create a `WorkloadRunner` directly, and you do not need `kubectl` or Kubernetes API access.

## How the Migration Works

For each legacy workload, the migration:

1. Lists the workloads in the selected NVIDIA Run:ai project or cluster.
2. Gets the workload from its typed `/api/v1/workloads/...` endpoint.
3. Removes fields that are returned by GET but are not accepted during creation.
4. Restores any secrets or credentials omitted from the GET response.
5. Posts the transformed request to the same typed collection endpoint.
6. Verifies the resulting workload through the Workloads V2 and typed APIs.

When the target cluster supports the required workload capabilities, the workloads service composes the typed request into a Workloads V2 native manifest. The cluster agent then wraps that manifest in a `WorkloadRunner`.

<Warning>

Do not submit a hand-written manifest to `/api/v2/workloads`. Doing so bypasses the server-side composer used by this migration.

</Warning>

## Supported Workload Types

The typed creation APIs support this procedure for the following workload types:

| Workload type | Typed API resource | Expected Workloads V2 manifest |
| --- | --- | --- |
| Workspace | `workspaces` | `batch/v1` `Job` |
| Training | `trainings` | `batch/v1` `Job` |
| Distributed Training | `distributed` | Framework-specific distributed workload |
| Inference | `inferences` | `serving.knative.dev/v1` `Service` |
| Distributed Inference | `distributed-inferences` | `leaderworkerset.x-k8s.io/v1` `LeaderWorkerSet` |

The target cluster must support the selected workload type and its native manifest. In particular, Inference requires Knative Serving, and Distributed Inference requires LeaderWorkerSet (LWS).

## Before You Begin

Confirm that:

* The target NVIDIA Run:ai version and cluster support Workloads V2 submissions.
* Knative Serving is available when migrating Inference workloads.
* LWS is available when migrating Distributed Inference workloads.
* Your account can read and create the relevant workload types in every project selected for migration.
* You have the original values for secrets and credentials that the GET API redacts or omits.
* The `runai` CLI, `curl`, and `jq` are installed.
* You understand whether starting a copy can safely run alongside the legacy workload.

<Warning>

Creating a replacement can immediately start application code. Running the old and new workloads concurrently can duplicate training, write to shared storage, expose conflicting endpoints, or consume additional resources.

</Warning>

## 1. Configure the Target

Set the control plane URL, authentication URL, and target cluster ID:

```bash
BASE_URL="https://<runai-control-plane>"
AUTH_URL="$BASE_URL"
CLUSTER_ID="<cluster-id>"
CONFIG_PATH="${HOME}/.runai-workload-migration"
```

Use a separate CLI configuration so the migration does not change your regular CLI profile:

```bash
runai \
  --config-path "$CONFIG_PATH" \
  config set \
  --cp-url "$BASE_URL" \
  --auth-url "$AUTH_URL"

runai \
  --config-path "$CONFIG_PATH" \
  login

TOKEN="$(
  runai \
    --config-path "$CONFIG_PATH" \
    auth get-token \
    -o plaintext
)"
```

<Warning>

Do not print the token or enable shell tracing while it is set.

</Warning>

## 2. Choose the Migration Scope

You can discover workloads by project or across the entire cluster.

Project-by-project migration is recommended, even when the goal is to recreate every legacy workload in the cluster. It limits the number of simultaneously running copies, makes credential review manageable, and provides a clear rollback boundary.

<Tabs>

<Tab title="One project at a time (recommended)">

Set the project name:

```bash
PROJECT_NAME="<project-name>"
```

List workloads in that project and cluster:

```bash
curl -fsS -G "$BASE_URL/api/v1/workloads" \
  -H "Authorization: Bearer $TOKEN" \
  --data-urlencode "filterBy=clusterId==$CLUSTER_ID" \
  --data-urlencode "filterBy=projectName==$PROJECT_NAME" \
  --data-urlencode "limit=100" |
jq '{
  workloads: [
    .workloads[] |
    {
      id,
      name,
      type,
      phase,
      projectId,
      projectName,
      clusterId,
      sourceApi
    }
  ]
}'
```

Complete and verify the workloads in this project before selecting the next project.

</Tab>

<Tab title="Cluster-wide inventory">

To inventory all visible workloads in the cluster, omit the project filter:

```bash
curl -fsS -G "$BASE_URL/api/v1/workloads" \
  -H "Authorization: Bearer $TOKEN" \
  --data-urlencode "filterBy=clusterId==$CLUSTER_ID" \
  --data-urlencode "limit=100" |
jq '{
  workloads: [
    .workloads[] |
    {
      id,
      name,
      type,
      phase,
      projectId,
      projectName,
      clusterId,
      sourceApi
    }
  ]
}'
```

The cluster filter affects discovery only. Each typed GET response contains the workload's project information, and the typed POST recreates the workload in that same project.

Before a cluster-wide migration, group the inventory by project:

```bash
curl -fsS -G "$BASE_URL/api/v1/workloads" \
  -H "Authorization: Bearer $TOKEN" \
  --data-urlencode "filterBy=clusterId==$CLUSTER_ID" \
  --data-urlencode "limit=100" |
jq '
  .workloads |
  sort_by(.projectName) |
  group_by(.projectName) |
  map({
    projectName: .[0].projectName,
    projectId: .[0].projectId,
    workloadCount: length,
    supportedLegacyWorkloadCount: (
      map(select(
        .sourceApi != "WorkloadsV2" and
        (
          .type == "Workspace" or
          .type == "Training" or
          .type == "Distributed" or
          .type == "Inference" or
          .type == "DistributedInference"
        )
      )) |
      length
    ),
    unsupportedWorkloadCount: (
      map(select(
        .type != "Workspace" and
        .type != "Training" and
        .type != "Distributed" and
        .type != "Inference" and
        .type != "DistributedInference"
      )) |
      length
    )
  })
'
```

Use this inventory to define the project order, then follow the project-by-project procedure. Cluster-wide discovery is not a recommendation to submit all replacements concurrently.

To recreate the supported legacy workloads cluster-wide, repeat [Select One Workload](#3-select-one-workload) through [Repeat for the Remaining Workloads](#10-repeat-for-the-remaining-workloads) for each workload in the inventory, completing one project before starting the next.

The API returns only workloads the caller is authorized to read. A cluster-wide migration therefore requires read and create permissions for every project in scope.

If the selected scope contains more than 100 workloads, retrieve all pages before continuing.

</Tab>

</Tabs>

Record each legacy workload's ID, name, and type. Use the workload type to select the typed API resource from the [supported workload types](#supported-workload-types) table.

Do not recreate a workload whose `sourceApi` already indicates `WorkloadsV2`.

## 3. Select One Workload

Migrate and verify one workload at a time. Set:

```bash
RESOURCE="<typed-api-resource>"
OLD_ID="<legacy-workload-id>"
REPLACEMENT_NAME="<temporary-replacement-name>"
```

Set `RESOURCE` to `workspaces`, `trainings`, `distributed`, `inferences`, or `distributed-inferences` according to the workload type.

Use a temporary replacement name so the legacy workload remains available during verification.

## 4. Get and Inspect the Typed Workload

Create a private temporary directory for the migration files:

```bash
umask 077
MIGRATION_DIR="$(mktemp -d)"
trap 'rm -rf "$MIGRATION_DIR"' EXIT

SOURCE_FILE="$MIGRATION_DIR/source.json"
PAYLOAD_FILE="$MIGRATION_DIR/payload.json"
RESPONSE_FILE="$MIGRATION_DIR/response.json"

curl -fsS \
  "$BASE_URL/api/v1/workloads/$RESOURCE/$OLD_ID" \
  -H "Authorization: Bearer $TOKEN" \
  > "$SOURCE_FILE"

jq . "$SOURCE_FILE"
```

Inspect the response before creating the replacement. Pay particular attention to:

* Environment variables
* Image-pull secrets
* Storage credentials
* User credentials
* Tokens, passwords, and private keys

<Warning>

A GET response can be sanitized. It must not be assumed to contain reusable secret values. Never submit a redaction marker as a real credential.

</Warning>

## 5. Build the Creation Request

The following response-only fields must be removed:

* `workloadId`
* `requestedName`
* `departmentId`
* `createdBy`
* `createdAt`
* `deletedAt`
* `desiredPhase`
* `actualPhase`

Transform the response and assign the temporary replacement name:

```bash
jq --arg name "$REPLACEMENT_NAME" '
  .name = $name |
  del(
    .workloadId,
    .requestedName,
    .departmentId,
    .createdBy,
    .createdAt,
    .deletedAt,
    .desiredPhase,
    .actualPhase
  )
' "$SOURCE_FILE" > "$PAYLOAD_FILE"
```

If the GET response omitted or redacted credentials, restore their original values in `"$PAYLOAD_FILE"` now. Keep the file private and delete it after the migration.

Review the complete payload before submitting it:

```bash
jq . "$PAYLOAD_FILE"
```

## 6. Create the Replacement

Post the transformed request to the same typed API:

```bash
curl --fail-with-body -sS \
  -X POST "$BASE_URL/api/v1/workloads/$RESOURCE" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  --data-binary "@$PAYLOAD_FILE" \
  > "$RESPONSE_FILE"

jq '{
  workloadId,
  name,
  desiredPhase,
  actualPhase,
  projectId,
  clusterId
}' "$RESPONSE_FILE"

NEW_ID="$(jq -er '.workloadId' "$RESPONSE_FILE")"
```

The API normally returns `202 Accepted`. Save `NEW_ID`; it is required for verification and lifecycle operations.

If the request fails, do not repeatedly resubmit it without first checking whether a replacement was created.

## 7. Verify the Workloads V2 Representation

```bash
curl -fsS \
  "$BASE_URL/api/v2/workloads/$NEW_ID" \
  -H "Authorization: Bearer $TOKEN" |
jq '{
  id: .workload.metadata.id,
  name: .workload.metadata.name,
  syncStatus: .syncInfo.status,
  gvk: .workload.metadata.gvk,
  manifestApiVersion: .workload.workloadManifest.apiVersion,
  manifestKind: .workload.workloadManifest.kind
}'
```

Verify that:

* The returned ID and name identify the replacement.
* `syncStatus` is `Synced`.
* Workspace and Training workloads contain a `batch/v1` `Job` manifest.
* Distributed Training workloads contain the expected distributed-workload manifest, such as a `kubeflow.org/v1` `PyTorchJob`.
* Inference workloads contain a `serving.knative.dev/v1` `Service` manifest.
* Distributed Inference workloads contain a `leaderworkerset.x-k8s.io/v1` `LeaderWorkerSet` manifest.

The cluster agent creates the `WorkloadRunner` internally. The migration procedure does not create that resource directly.

## 8. Compare the Typed Specifications

Normalize the source and replacement GET responses:

```bash
EXPECTED_FILE="$MIGRATION_DIR/expected.json"
ACTUAL_FILE="$MIGRATION_DIR/actual.json"

jq --arg name "$REPLACEMENT_NAME" '
  .name = $name |
  del(
    .workloadId,
    .requestedName,
    .departmentId,
    .createdBy,
    .createdAt,
    .deletedAt,
    .desiredPhase,
    .actualPhase
  )
' "$SOURCE_FILE" > "$EXPECTED_FILE"

curl -fsS \
  "$BASE_URL/api/v1/workloads/$RESOURCE/$NEW_ID" \
  -H "Authorization: Bearer $TOKEN" |
jq '
  del(
    .workloadId,
    .requestedName,
    .departmentId,
    .createdBy,
    .createdAt,
    .deletedAt,
    .desiredPhase,
    .actualPhase
  )
' > "$ACTUAL_FILE"

if jq -e -s '.[0] == .[1]' \
  "$EXPECTED_FILE" "$ACTUAL_FILE" >/dev/null
then
  echo "Typed specifications match"
else
  echo "Typed specifications differ; inspect before continuing"
fi
```

<Info>

**Note**

The replacement API can add explicit defaults. For example, it can return `spec.security.privileged: false` when that field was absent from the legacy response. This default does not grant privileges, but every difference must still be reviewed.

</Info>

Do not print a diff in shared logs if the normalized files contain restored credentials.

## 9. Verify Runtime State

Confirm that the replacement is listed as a Workloads V2 workload:

```bash
curl -fsS -G "$BASE_URL/api/v1/workloads" \
  -H "Authorization: Bearer $TOKEN" \
  --data-urlencode "filterBy=clusterId==$CLUSTER_ID" \
  --data-urlencode "limit=100" |
jq --arg id "$NEW_ID" '
  .workloads[] |
  select(.id == $id) |
  {
    id,
    name,
    type,
    phase,
    sourceApi
  }
'
```

Inspect the pods reported by the control plane:

```bash
curl -fsS \
  "$BASE_URL/api/v1/workloads/$NEW_ID/pods" \
  -H "Authorization: Bearer $TOKEN" |
jq '{
  pods: [
    .pods[] |
    {
      name,
      k8sPhase,
      nodeName
    }
  ]
}'
```

Verify pod creation, storage, network access, credentials, resource allocation, and application behavior. Rendered pods need not be byte-for-byte identical because the newer submission path uses newer controllers and generated identities.

An Inference workload configured to scale to zero can have no running pods while idle. For that workload, verify the Knative service state and send a test inference request to confirm that it scales up and serves traffic.

For Distributed Inference, verify the expected leader and worker topology, the serving endpoint, and communication between replicas.

## 10. Repeat for the Remaining Workloads

After one replacement passes verification:

1. Remove the private temporary directory or allow the shell trap to remove it.
2. Set `RESOURCE`, `OLD_ID`, and `REPLACEMENT_NAME` for the next workload.
3. Repeat the GET, inspection, transformation, creation, and verification steps.

For a cluster-wide recreation, finish and validate all selected workloads in one project before moving to the next project. Track the old and new workload IDs for every project so cutover and rollback remain unambiguous.

<Warning>

Do not turn this procedure into an unattended bulk migration unless credential restoration and workload-specific safety checks are handled explicitly.

</Warning>

## Production Cutover

The steps above leave the legacy workload untouched.

For production cutover:

1. Fully verify the temporary replacement.
2. Stop or delete the legacy workload through the NVIDIA Run:ai control-plane API.
3. If retaining the original name is required, delete the verified temporary replacement and repeat the creation request with the original name.
4. Verify the final workload again.

<Warning>

Stopping or deleting the legacy workload is destructive. Perform that step only after application owners approve the replacement and the cutover plan.

</Warning>

## Limitations

* Sanitized GET responses are not guaranteed to round-trip secrets.
* Platform defaults can create small typed-spec differences.
* Rendered pods can differ because the Workloads V2 path uses different controllers, generated identities, or component versions.
* Workload-specific scripts can depend on legacy environment variables or controller behavior. Validate those dependencies before cutover.
