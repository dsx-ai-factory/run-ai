# OSMO to Run:ai Workflows: Gaps

[TOC]

## Overview

This document collects the gaps found in [OSMO to Run:ai Workflows](osmo-to-runai.md), the research on a native Run:ai `workflow` type that the composer translates into one Argo Workflow (v4.1).

A gap is an OSMO capability that the proposed workflow doesn't cover yet, or covers with a difference in behavior.

Each item links to the section of the main document that discusses it.

## Implementation Difficulties

The following OSMO capabilities have no direct Argo translation.

We have to get input from the OSMO team on how important these requirements are to OSMO customers, and based on that decide whether each of them should be dropped or investigated further as part of the design.

### Group Barrier

OSMO's `barrier`, on by default in groups with more than one task, makes all the tasks of a group start their command at the same moment. The PodGroup only places the pods together: each pod still pulls its image and downloads its inputs at its own pace, so without a barrier a task can start early and hold its GPUs while it waits for the others.

See [Group Barrier](osmo-to-runai.md#group-barrier).

### Success on a Non-Zero Exit Code

OSMO's `exitActions.COMPLETE` marks a task as completed on the listed exit codes, while Argo marks a task as Failed on any non-zero exit code.

See [Group SuccessExitCodes](osmo-to-runai.md#group-successexitcodes) and [Success on a Non-Zero Exit Code](osmo-to-runai.md#success-on-a-non-zero-exit-code).

### Replicas and Host Names

A task with `replicas` spawns several pods, and each of them needs its own hostname. But `RUNAI_HOST_<TASK>` holds a single name, so the design will need to define how the other tasks find each replica.

See [Replicas and Host Names](osmo-to-runai.md#replicas-and-host-names).

### Group Failure Policy

With `ignoreNonleadStatus: false`, OSMO fails a group as soon as one of its tasks fails, and stops the group's other tasks. Argo fails the group only after its running tasks end, and a group with a leader keeps running, with its GPUs, until the leader ends.

See [Group Failure Policy](osmo-to-runai.md#group-failure-policy).

### Preemption

This case hasn't been tested yet. According to OSMO's source, OSMO detects a preemption from the pod condition `PreemptionByScheduler`, which KAI sets before it evicts the pod, and marks the task as FAILED_PREEMPTED with exit code 3006. In Argo, we expect a preempted attempt to end in the Error phase with the message "pod deleted", the same as any pod deleted outside Argo, so Argo's status alone can't tell a preemption from other deletions. This is unverified: Argo's code gives this message only when the pod is gone before the controller sees its final state, so the attempt may also end as Failed with the container's exit code.

See [Preemption](osmo-to-runai.md#preemption).

### UDP Port Forwarding

OSMO users reach a running task on demand with `exec`, `port-forward` and `rsync`, and Run:ai can offer the same by targeting the task's pod. But Kubernetes port-forward supports only TCP, while OSMO users may need to forward UDP (for example, for the Isaac Sim livestream).

See [External Access to Running Tasks](osmo-to-runai.md#external-access-to-running-tasks).

## Not Covered Yet

The proposed design doesn't cover these OSMO capabilities yet. As with the difficulties above, their importance should be checked with the OSMO team.

### Templates and Apps

OSMO specs are Jinja templates, with variables and their defaults, loops and conditions, and users override the variables at submission with `--set`. OSMO apps store such templates in OSMO, with versions. Run:ai's templates cover storing a spec and overriding whole fields at submission, but not variables inside strings, versions, or Jinja's loops and conditions, except for loops that create identical tasks.

See [Templates and Apps](osmo-to-runai.md#templates-and-apps).

### Restart of a Failed Workflow

`osmo workflow restart` runs again the groups of a failed workflow that didn't complete, and reads the outputs of the completed groups from the original run. Run:ai has no restart action. A task's failed attempts can be retried while the workflow runs, through Argo's `retryStrategy` (see `tasks[].retry`), but Argo retries a Workflow that has already failed only from its CLI and server, not through a spec field: Run:ai would have to vendor Argo's retry code, or create a new workflow as OSMO does.

See [Restart of a Failed Workflow](osmo-to-runai.md#restart-of-a-failed-workflow).

### Cancel

`osmo workflow cancel` stops a workflow and keeps it, with its status and logs. Deleting a workflow in Run:ai removes its pods, and their logs with them. A stop action, translated to Argo's `shutdown: Terminate`, would match OSMO's cancel.

See [Cancel](osmo-to-runai.md#cancel).

### Logs After the Workflow Ends

OSMO uploads a workflow's logs to S3 when it ends, so they last longer than its pods. In Run:ai, a task's logs last only as long as its pod, and Argo can archive them only to an Argo artifact repository.

See [Logs After the Workflow Ends](osmo-to-runai.md#logs-after-the-workflow-ends).

## Behavioral Differences

For each of these differences, it should be a product decision whether Run:ai adopts Argo's behavior or provides a workaround that keeps OSMO's behavior.

- **Status names:** OSMO and Argo give the same task outcomes different status names, for example FAILED_UPSTREAM and Omitted. The suggested approach is to use Argo's status names. See [Workflow Status](osmo-to-runai.md#workflow-status).
- **Retry attempts:** OSMO marks the attempts that were retried as RESCHEDULED and only the last one as FAILED. Argo marks every attempt as Failed, under a Retry node with the message "No more retries left". Both report the same attempts and the same final result. See [Retry Attempts](osmo-to-runai.md#retry-attempts).
- **Skipped downstream tasks:** when a task fails, OSMO marks the tasks that depend on it as FAILED_UPSTREAM right away. Argo adds them to the status as Omitted only after their other dependencies finish too, and until then they don't appear in the status. See [Skipped Downstream Tasks](osmo-to-runai.md#skipped-downstream-tasks).
- **Groups with a leader:** when the leader ends, OSMO gives the tasks still running the leader's status. In Argo, the other tasks are daemons: a daemon that Argo stops is marked Succeeded whatever the leader's status, and a daemon that exits on its own is marked Failed, even with exit code 0. See [Groups With a Leader](osmo-to-runai.md#groups-with-a-leader).

## Limitations

These differences from OSMO are part of the proposed design:

- **Node pools:** KAI gives each PodGroup a single node pool, taken from its first pod, so `nodePools` is a group field, and tasks that need different node pools must be in different groups. See [General Fields](osmo-to-runai.md#general-fields).
- **Image pull secrets:** Argo has no per-template field for them, so the composer should merge the secrets of all the tasks into `spec.imagePullSecrets`, and every pod of the workflow gets all of them. See [General Fields](osmo-to-runai.md#general-fields).
- **Platforms:** an OSMO platform is a named set of node selectors and tolerations inside a pool. Run:ai has no such set inside a node pool, so each task sets `nodeAffinityRequired` and `tolerations`. These two fields could also be allowed in a compute profile. See [Compute and Placement](osmo-to-runai.md#compute-and-placement).
- **Directory separation:** OSMO gives each task its own folder in the data store. `directorySeparation` gives each workflow its own directory, which all its tasks share, so avoiding file name clashes between tasks is the researcher's responsibility. The S3 copy option, if it is adopted, would give each task its own folder, as OSMO does. See [Input/Output Chain](osmo-to-runai.md#inputoutput-chain) and [S3 by Copy](osmo-to-runai.md#s3-by-copy).
- **S3 through goofys:** a file is uploaded only when it's closed, and appending to a file or rewriting part of it isn't supported. Tasks that need these, such as writing checkpoints in place, should use NFS or a PVC. The S3 copy option, if it is adopted, supports them on the task's local disk, but uploads the outputs only after the task ends, so a checkpoint written during the run isn't in S3 until then, and can't be used to recover the run. See [S3 Through goofys](osmo-to-runai.md#s3-through-goofys) and [S3 by Copy](osmo-to-runai.md#s3-by-copy).
- **PVC access modes:** with `readWriteOnce`, a task that starts while a task on another node holds the volume waits until the volume is released, and zonal block storage keeps every task that mounts it in the volume's availability zone. See [PVC](osmo-to-runai.md#pvc).
- **Workflow size:** Argo keeps a workflow in one Kubernetes object of about 1 MB at most, so a workflow with hundreds of tasks, as OSMO's Jinja loops can create, may not fit, and Run:ai's status sync must read Argo's compressed status. OSMO has no such limit. See [Large Workflows](osmo-to-runai.md#large-workflows).

## Dropped OSMO Capabilities

We don't plan to support these OSMO capabilities:

| OSMO capability | Reason |
| --- | --- |
| Masking credential values of 8 characters or more in logs | Run:ai and Argo don't mask logs, so a credential that a task prints can be read by anyone who can view the workflow's logs (the Run:ai UI and CLI, or `pods/log` access to the project's namespace). It stays readable as long as the pod exists (Argo keeps pods until the Workflow is deleted, unless `podGC` is set), and in downloaded logs and the cluster's log aggregation, if any. OSMO masks only the logs its own services serve, so masking in Run:ai's log views would give the same protection. Open decision for product, with security review: drop masking and document that tasks must not print credentials, or add it to Run:ai's log views |
| `kpis` | It isn't documented, and OSMO only uploads the files without reading them. A task can write such files to its storage |
| `inputs[].regex`, `outputs[].regex` | Tasks mount the whole directory |
| `checkpoint[]` | Tasks write checkpoints directly to mounted storage |
| `downloadType`, `cacheSize` | Storage is mounted, not downloaded, unless the S3 copy option is adopted (see [S3 by Copy](osmo-to-runai.md#s3-by-copy)) |

## Run:ai Fields Not Supported in Workflows

The composer should reject these fields of the other Run:ai types in a workflow:

| Run:ai field | Reason |
| --- | --- |
| `backoffLimit` | Replaced by `retry`, which can be limited to exit codes |
| `restartPolicy` | Argo pods always use `Never`, and a retry runs in a new pod |
| `ports`, `exposedUrls` | Left out for now, since users can reach a running task on demand (see [UDP Port Forwarding](#udp-port-forwarding)) |
| `completions`, `parallelism` | Training only (hyperparameter optimization) |
| `nodeType` (deprecated) | Replaced by `nodeAffinityRequired` |
