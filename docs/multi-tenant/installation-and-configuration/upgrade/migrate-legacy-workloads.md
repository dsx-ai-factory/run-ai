# Migrate Legacy Workloads

The controllers and CRDs behind native workload types (workspace, training, distributed training, standard inference, distributed inference) are removed in v3.0 and replaced by a single, unified controller and CRD. Running and pending native workloads that were submitted through a legacy workload type before upgrading are affected by the upgrade and lose some capabilities until resubmitted.

Resubmit each affected workload to restore full functionality. Workloads can be resubmitted independently, so you can resubmit them in stages after upgrading. Resubmit an affected workload either [manually](#submit-workload-via-ui-api-cli) through the NVIDIA Run:ai UI, API, or CLI, or, to reuse each workload's existing configuration instead of re-entering it by hand, [using the migration script](#submit-workload-via-migration-script) against the control plane API.

## Affected Capabilities

Resubmit affected workloads to restore the capabilities listed in the table below.

| Capability | Status after upgrade |
| --- | --- |
| Suspend / Resume | Not functional |
| Error handling / auto-restart | Not functional. A workload backed by a third-party controller, such as Knative, keeps running, but NVIDIA Run:ai no longer tracks its status |
| Fractional GPU support | Not functional. Workloads using GPU fractions should also be migrated, following [Migrate GPU Fractions](migrate-gpu-fractions.md), since fractions are now CUDA-based |
| Events / status visibility | Partially functional. Pods are still detected and their status is still updated |
| Update (modify spec in place) | Not functional. Applies to standard inference and distributed inference workloads |
| Make a Copy, metrics and monitoring, log streaming, port forwarding, CLI operations (`runai get`/`describe`/`logs`), delete | Unaffected |

## Submit Workload via UI/API/CLI

Resubmitting an affected workload through the NVIDIA Run:ai UI, API, or CLI binds it to the unified controller, resolving the affected capabilities listed in the table above.

1. Identify the affected workloads using the [Affected Capabilities](#affected-capabilities) table above.
2. Make a copy of each workload - using **Make a Copy** in the NVIDIA Run:ai UI, or by resubmitting the same configuration through the API or CLI.
3. Give the copy a different name so it can coexist with the legacy workload during verification. Submit it as-is, with the same fields and configuration - no changes are required.
4. Verify the replacement works as expected.
5. Stop or delete the legacy workload once the replacement is verified.

<Note>

**Note**

If the workload must keep its original name, delete the temporary replacement and resubmit it with the original name, then verify it again.

</Note>

## Submit Workload via Migration Script

<Warning>

**Warning**

Do not turn this procedure into an unattended bulk migration unless credential restoration and workload-specific safety checks are handled explicitly.

</Warning>

This script recreates a legacy workload under a new name through the NVIDIA Run:ai control plane API, reusing its existing configuration instead of requiring you to re-enter it by hand. It uses only control plane APIs - no `kubectl` or Kubernetes API access required.

Set `RESOURCE`, `OLD_ID`, and `REPLACEMENT_NAME` for the workload you want to migrate, then run the steps below. Repeat the steps for each additional workload. The `runai` CLI, `curl`, and `jq` must be installed.

### Configure the Target

Configure the target once per session - it does not need to be repeated per workload.

1. Set the control plane URL, authentication URL, and target cluster ID:

   ```bash
   BASE_URL="https://<runai-control-plane>"
   AUTH_URL="$BASE_URL"
   CLUSTER_ID="<cluster-id>"
   CONFIG_PATH="${HOME}/.runai-workload-migration"
   ```

2. Authenticate using a separate CLI configuration, so the migration does not change your regular CLI profile:

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

   **Warning**

   Do not print the token or enable shell tracing while it is set.

   </Warning>

### Identify the Legacy Workloads

1. Identify the legacy workloads to migrate, either by project or across the entire cluster. Project-by-project migration is recommended, even when the goal is to recreate every legacy workload in the cluster. It limits the number of simultaneously running copies, makes credential review manageable, and provides a clear rollback boundary.

   <Tabs>

   <Tab title="One project at a time (recommended)">

   1. Set the project name:

      ```bash
      PROJECT_NAME="<project-name>"
      ```

   2. List workloads in that project and cluster:

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

   3. Complete and verify the workloads in this project before selecting the next project.

   </Tab>

   <Tab title="Cluster-wide inventory">

   1. Inventory all visible workloads in the cluster by omitting the project filter:

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

   2. Group the inventory by project:

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

   To recreate the supported legacy workloads cluster-wide, repeat [Identify the Legacy Workloads](#identify-the-legacy-workloads) through [Delete the Legacy Workload](#delete-the-legacy-workload) for each workload in the inventory, completing one project before starting the next.

   The API returns only workloads the caller is authorized to read. A cluster-wide migration therefore requires read and create permissions for every project in scope.

   If the selected scope contains more than 100 workloads, retrieve all pages before continuing.

   </Tab>

   </Tabs>

2. Record each legacy workload's ID, name, and type.

3. Set the resource variables for one workload:

   ```bash
   RESOURCE="<typed-api-resource>"
   OLD_ID="<legacy-workload-id>"
   REPLACEMENT_NAME="<temporary-replacement-name>"
   ```

   1. Set `RESOURCE` to the typed API resource matching the workload's type:

      | Workload type | Typed API resource |
      | --- | --- |
      | Workspace | `workspaces` |
      | Training | `trainings` |
      | Distributed training | `distributed` |
      | Standard inference | `inferences` |
      | Distributed inference | `distributed-inferences` |

   2. Set `OLD_ID` to the legacy workload's ID, recorded in the previous step.

   3. Set `REPLACEMENT_NAME` to a temporary name, so the legacy workload remains available during verification.

4. Create a private temporary directory for the migration files, then get the workload:

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

5. Inspect the response before creating the replacement. Pay particular attention to:

   * Environment variables
   * Image-pull secrets
   * Storage credentials
   * User credentials
   * Tokens, passwords, and private keys

   <Warning>

   **Warning**

   A GET response can be sanitized. It must not be assumed to contain reusable secret values. Never submit a redaction marker as a real credential.

   </Warning>

### Submit the Replacement

1. Transform the response and assign the temporary replacement name. The following response-only fields must be removed:

   * `workloadId`
   * `requestedName`
   * `departmentId`
   * `createdBy`
   * `createdAt`
   * `deletedAt`
   * `desiredPhase`
   * `actualPhase`

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

2. Review the complete payload before submitting it:

   ```bash
   jq . "$PAYLOAD_FILE"
   ```

3. Post the transformed request to the same typed API. The target cluster must support the required workload capabilities for the replacement to be created successfully.

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

### Verify the Replacement

1. Verify the Workloads V2 representation:

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
   * Distributed training workloads contain the expected distributed-workload manifest, such as a `kubeflow.org/v1` `PyTorchJob`.
   * Standard inference workloads contain a `serving.knative.dev/v1` `Service` manifest.
   * Distributed inference workloads contain a `leaderworkerset.x-k8s.io/v1` `LeaderWorkerSet` manifest.

2. Compare the typed specifications. Normalize the source and replacement GET responses:

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

   <Note>

   **Note**

   The replacement API can add explicit defaults. For example, it can return `spec.security.privileged: false` when that field was absent from the legacy response. This default does not grant privileges, but every difference must still be reviewed.

   </Note>

   Do not print a diff in shared logs if the normalized files contain restored credentials.

3. Confirm that the replacement is listed as a Workloads V2 workload:

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

4. Inspect the pods reported by the control plane:

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

   Verify that:

   * Pod creation, storage, network access, credentials, resource allocation, and application behavior are as expected. Rendered pods need not be byte-for-byte identical, since the newer submission path uses newer controllers and generated identities.
   * A standard inference workload configured to scale to zero can have no running pods while idle - verify the Knative service state and send a test inference request to confirm that it scales up and serves traffic.
   * For distributed inference, verify the expected leader and worker topology, the serving endpoint, and communication between replicas.

### Delete the Legacy Workload

The steps above leave the legacy workload untouched.

1. Fully verify the temporary replacement.
2. Delete the legacy workload through the NVIDIA Run:ai control plane API.
3. If retaining the original name is required, delete the verified temporary replacement and repeat the creation request with the original name.
4. Verify the final workload again.

### Repeat for the Remaining Workloads

You can migrate workloads one by one, or across all workloads in a project or the entire cluster. Either way, after one replacement passes verification:

1. Remove the private temporary directory or allow the shell trap to remove it.
2. Set `RESOURCE`, `OLD_ID`, and `REPLACEMENT_NAME` for the next workload.
3. Repeat the steps above for the next workload.

For a cluster-wide recreation, finish and validate all selected workloads in one project before moving to the next project. Track the old and new workload IDs for every project so cutover and rollback remain unambiguous.
