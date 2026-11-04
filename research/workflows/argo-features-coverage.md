# Argo Features Not Covered by the Run:ai Workflow API

[TOC]

## Overview

[OSMO to Run:ai Workflows](osmo-to-runai.md) proposes a native Run:ai `workflow` type that the composer translates into one Argo Workflow, and [OSMO to Run:ai Workflows: Gaps](osmo-to-runai-gaps.md) summarizes what that translation leaves open. Both documents start from OSMO: OSMO's capabilities are the requirements, and Argo is the engine that implements them.

This document starts from the other side. It goes over the features of Argo Workflows v4.1 and lists the ones that the proposed API doesn't give users. Argo already implements these features, so many of them can be exposed with a translation in the composer rather than new machinery. Users who know Argo, directly or through tools built on it such as Kubeflow Pipelines, may also expect them.

The review covered:

- every field a user can set in the [Argo v4.1 field reference](https://argo-workflows.readthedocs.io/en/release-4.1/fields/): the `WorkflowSpec`, `Template`, `DAGTask` and `WorkflowStep` types, and the types under them;
- the other Argo resources: `WorkflowTemplate`, `ClusterWorkflowTemplate`, `CronWorkflow` and `WorkflowEventBinding`;
- the operations on a workflow that are not spec fields: stop, terminate, suspend, resume, retry and resubmit.

The [Fields Translation](osmo-to-runai.md#fields-translation) tables of the main document were used as the list of what the API covers. Each Argo feature got one of these statuses:

| Status | Meaning |
| --- | --- |
| **Covered** | A Run:ai field translates to it. |
| **Composer only** | The composer uses it to implement Run:ai fields, but users can't use it directly. |
| **Not covered** | No Run:ai field gives it. These features are the subject of this document. |
| **Not applicable** | An installation setting of the Argo controller, or something Run:ai provides in its own way. |

Nothing in this document was tested on a cluster. The statements about Argo come from its v4.1 documentation and source code (the `release-4.1` branch), and the statements about KAI from the KAI v0.18.2 source. Items that need a test say so.

### Main Findings

The proposed API covers Argo's execution core: a DAG of container tasks with retries, timeouts, storage, scheduling and security settings. What it leaves out are mostly the features that make the graph dynamic, or that manage the workflow over time:

- **Passing values between tasks**, and using them to decide what runs next (output parameters, `when`, `withParam`). Tasks can share files today, but a file can't change the graph.
- **Reacting to failure**: groups that run when another group fails (`depends` on `.Failed`), and exit handlers that run at the end of the workflow, whatever its result.
- **Fan-out of independent tasks** (`withItems`, `withSequence`, `withParam`), with a limit on how many run at once (`parallelism`). The proposed `replicas` always gang-schedules its copies, and `groups[].copies` runs a fixed number of independent copies, but there is no fan-out over a list or an output, and no limit.
- **Workflow operations**: suspend steps for manual approval, and retrying a failed workflow from the group that failed. Argo's suspend also doesn't match Run:ai's suspend, which deletes the pods.
- **Reuse and scheduling**: workflow parameters, workflow templates and scheduled runs (CronWorkflow).

Most of the other Argo features that the API leaves out are left out by the other Run:ai workload types too (for example, a single user container per task, or readiness probes only), or depend on an Argo artifact repository, which the design replaces with mounted storage, or with S3 artifacts that name their own bucket.

The review also found points that affect the main document, listed in [Findings That Affect the Main Document](#findings-that-affect-the-main-document). Most of them are now discussed in the main document. The most important one is that a failed task in a gang group can get the rest of the group evicted by KAI, including the leader of a group with `failurePolicy: Leader`.

## Summary

The table lists the features that are not covered, by area, with a suggested priority. The priority is our estimate of how often AI/ML workflows need the feature, and is meant as a starting point for discussion.

| Area | Not covered | Suggested priority |
| --- | --- | --- |
| [Data Passing](#data-passing) | Output parameters, `outputs.result`, workflow outputs | High |
| | Artifacts other than git, raw and S3 | Low |
| [Graph Control](#graph-control) | `when`, `depends` on failure, fan-out over a list or an output | High |
| | Loops and recursion, `dag.target` | Low |
| [Exit Handlers and Lifecycle Hooks](#exit-handlers-and-lifecycle-hooks) | `onExit`, `hooks` | Medium |
| [Suspend Steps and Workflow Actions](#suspend-steps-and-workflow-actions) | Retry from the failed group, stop with exit handlers | High |
| | Suspend steps and intermediate parameters | Medium |
| [Reuse and Scheduling](#reuse-and-scheduling) | Workflow parameters, workflow templates, CronWorkflow | Medium |
| | `templateRef`, WorkflowEventBinding, memoization | Low |
| [Concurrency Control](#concurrency-control) | `parallelism` | Medium, High together with fan-out |
| | Mutexes and semaphores | Low |
| [Retry and Timeouts](#retry-and-timeouts) | Retry on another node, backoff, a deadline for the whole workflow | Medium |
| | A workflow-level retry default, retry by failure message, TTL by result | Low |
| [Task Shape](#task-shape) | Services shared across groups, per-group priority and preemptibility | Medium |
| | Sidecars, init containers, container sets, the HTTP template, more container and pod fields | Low |
| [Observability](#observability) | Progress, log archiving | Low |

## Data Passing

### Parameters and Results

In Argo, a template can declare output parameters, each read from a file that the task writes (`valueFrom.path`). Argo also keeps the standard output of a task's main container as `outputs.result`. Later tasks use these values in their arguments, in `when` conditions, as the list of a fan-out (`withParam`), or in their command, through tags such as `{{tasks.evaluate.outputs.parameters.accuracy}}`. An output parameter with a `globalName` also becomes an output of the whole workflow, in the Workflow's `status.outputs`.

```yaml
spec:
  entrypoint: main
  templates:
  - name: main
    dag:
      tasks:
      - name: evaluate
        template: evaluate
      - name: deploy
        template: deploy
        dependencies: [evaluate]
        when: "{{tasks.evaluate.outputs.parameters.accuracy}} > 0.9"   # Skipped otherwise

  - name: evaluate
    container:
      image: my-eval:1.0
      command: [bash, -c]
      args: ["python eval.py --out /tmp/accuracy"]
    outputs:
      parameters:
      - name: accuracy
        valueFrom:
          path: /tmp/accuracy       # read by Argo when the task ends
        globalName: accuracy        # also in the Workflow's status.outputs

  - name: deploy
    container:
      # ...
```

In the proposed API, tasks share data only through files on shared storage (see [Input/Output Chain](osmo-to-runai.md#inputoutput-chain)). That passes data, but it can't change the workflow:

- the next group always runs, and has to read the value itself;
- a value can't decide whether a group runs, or how many copies of it run;
- the value isn't visible in Run:ai's status.

A possible direction is an `outputs` list on a task (a name and a path), with a reference syntax that can be used in `when`, in environment variable values and in a fan-out list. The main DAG sees groups and not tasks, so the composer would also have to expose each task's outputs as outputs of its group's DAG template (`valueFrom.parameter: "{{tasks.<task>.outputs.parameters.<name>}}"`). Workflow outputs could be shown in Run:ai's workflow status, for example the final accuracy of a training pipeline. OSMO's `kpis`, dropped in the main document, aim at a similar need.

Argo keeps parameter values in the Workflow's status, so they are meant for short values, not for data.

Argo reads the `{{...}}` syntax in every string of a template, which already affects user text in the proposed API. See [User Strings and Argo Templating](#user-strings-and-argo-templating).

### Artifacts

Argo can download input artifacts into a task before it starts, and upload output artifacts after it ends. It supports S3, GCS, Azure, HTTP, HDFS, OSS, Artifactory, git, and raw content given in the spec. The main document uses git and raw artifacts, and proposes S3 input and output artifacts as an option for S3 storage (see [S3 by Copy](osmo-to-runai.md#s3-by-copy)). It mounts all other storage (see [Run:ai Supported Storage](osmo-to-runai.md#runai-supported-storage)). Not covered:

- artifacts other than git, raw and S3 (GCS, Azure, HTTP, HDFS, OSS, Artifactory);
- the features that depend on an artifact repository: `archiveLocation`, `archiveLogs`, `artifactGC`, the `data` template, and passing an artifact from one task to another (`from: "{{tasks.a.outputs.artifacts.x}}"`) without naming its location. Run:ai doesn't plan to configure an artifact repository, so these stay not covered.

S3 artifacts don't need a repository, because an S3 artifact can name its bucket, endpoint and credentials Secret itself. That is what makes the main document's S3 by Copy option possible. Not tested.

## Graph Control

### Conditional Groups

A DAG task in Argo can have a `when` condition, which Argo evaluates just before the task would start. If it is false, the task is Skipped. Conditions can use workflow parameters, output parameters, and the status of earlier tasks (`{{tasks.train.status}}`). With `dependencies`, a Skipped task counts as done, so the tasks that depend on it still run.

Uses: deploy or register a model only if its evaluation passed; run an expensive validation only when a parameter asks for it.

The proposed API has no condition on a group. A `groups[].when` would translate directly to `when` on the group's task in the main DAG, but it needs values to test, so it depends on [Parameters and Results](#parameters-and-results) and [Workflow Parameters and Templates](#workflow-parameters-and-templates). Run:ai's status would also need a Skipped state for the tasks of a skipped group, next to the upstream-failure case in [Skipped Downstream Tasks](osmo-to-runai.md#skipped-downstream-tasks).

### Dependencies on Failure

Argo's `depends` field takes a boolean expression over the results of other tasks: `Succeeded`, `Failed`, `Errored`, `Skipped`, `Omitted` and `Daemoned`, plus `AnySucceeded` and `AllFailed` for a fan-out. A task name without a result means `Succeeded || Skipped || Daemoned`, the same as `dependencies`.

```yaml
- name: main
  dag:
    tasks:
    - name: train
      template: train
    - name: evaluate-last-checkpoint      # runs whether train succeeded or failed
      template: evaluate
      depends: "train.Succeeded || train.Failed"
    - name: cleanup-on-failure            # runs only if train failed
      template: cleanup
      depends: "train.Failed || train.Errored"
```

In the proposed API, `dependsOn` starts a group only when all the listed groups succeed. It can't express:

- evaluating the last checkpoint of a training that failed or hit its timeout;
- cleaning up or alerting after a failure;
- continuing the workflow when an optional group fails.

A possible translation is a condition per dependency, for example `dependsOn: [{group: train, on: [Succeeded, Failed]}]`, which the composer turns into a `depends` expression. Two points come with it:

- Argo rejects a DAG template that uses both `depends` and `dependencies`, and rejects `continueOn` in a DAG that uses `depends` (`validate.go`). The composer uses `continueOn` only in the group DAGs, so the main DAG can switch to `depends` as a whole (`dependencies: [a, b]` is the same as `depends: "a && b"`).
- In Argo, a failure that a later task handles doesn't fail the DAG. The DAG's result comes from its leaf tasks, so if `cleanup-on-failure` succeeds, the workflow ends as Succeeded although the training failed (from `assessDAGPhase` in `dag.go`; to be tested). Run:ai should decide which status the workflow gets in this case.

### Fan-Out of Independent Tasks

`withItems` (a static list), `withSequence` (a range of numbers) and `withParam` (a JSON list, usually an output of an earlier task) expand one DAG task into one task per item. Each copy runs in its own pod with its own PodGroup, so KAI schedules each copy on its own, as resources allow. The item is available as `{{item}}`, the template's `parallelism` limits how many copies run at once, and the tasks that depend on the fan-out wait for all the copies (or use `.AnySucceeded` and `.AllFailed`). The output parameters of the copies are collected into a JSON list.

```yaml
spec:
  entrypoint: main
  templates:
  - name: main
    dag:
      tasks:
      - name: sweep
        template: sweep
      - name: pick-best
        template: pick-best
        dependencies: [sweep]

  - name: sweep
    parallelism: 2                  # at most 2 trials at a time
    dag:
      tasks:
      - name: trial
        template: train
        arguments:
          parameters:
          - name: lr
            value: "{{item}}"
        withItems: ["0.1", "0.01", "0.001", "0.0001"]

  - name: train
    inputs:
      parameters:
      - name: lr
    container:
      image: my-train:1.0
      args: ["--lr", "{{inputs.parameters.lr}}"]

  - name: pick-best
    container:
      # ...
```

Uses: hyperparameter sweeps; batch inference or preprocessing over the shards of a dataset; evaluation over a list of benchmarks or checkpoints.

In the proposed API, `tasks[].replicas` creates copies of a task in its group's PodGroup, so all the copies must be scheduled together. This is right for distributed training, but a sweep of 50 trials would wait until all 50 fit at once. The main document also proposes `groups[].copies`, a fixed number of independent copies of a group, which the composer translates to `withSequence` on the group's task in the main DAG (see [General Fields](osmo-to-runai.md#general-fields)). It covers a fan-out over a count, but not over a list of values or an earlier task's output, and it has no limit on how many copies run at once.

A possible direction is to extend `copies` into a fan-out, for example with `items` or a reference to an output, and `parallelism`. The composer would put `withItems` or `withParam` on the group's task, as it puts `withSequence` for `copies`, and pass the item to the tasks as an environment variable. Run:ai's status would show one entry per copy (Argo names them `sweep.trial(0:0.1)`, `sweep.trial(1:0.01)`, and so on).

A fan-out also needs [Parallelism](#parallelism).

### Loops and Recursion

An Argo template can call itself, with a `when` condition that ends the loop. This runs a group until something happens, for example active-learning rounds that repeat until a metric reaches its target. `dag.target` runs only some tasks of a DAG and the tasks they depend on. Neither is covered. Recursion needs output parameters and `when` first, and the Workflow's status grows with each round (see [Large Workflow Status](#large-workflow-status)). Low priority.

## Exit Handlers and Lifecycle Hooks

Argo runs `spec.onExit` (or the `exit` hook in `spec.hooks`) at the end of the workflow, whether it succeeded, failed or was stopped, but not when it is terminated. The exit template can read `{{workflow.status}}` and `{{workflow.failures}}` (a JSON list of the failed tasks), and can be a DAG that branches on them. A DAG task can have its own exit hook, which runs after the task ends. Other hooks have an `expression`, and run once when it becomes true, for example `workflow.status == "Running"` to report that the workflow has started.

```yaml
spec:
  entrypoint: main
  hooks:
    exit:                             # the same as spec.onExit: exit-handler
      template: exit-handler
    running:
      expression: workflow.status == "Running"
      template: notify-started
  templates:
  - name: exit-handler
    container:
      image: my-tools:1.0
      command: [bash, -c]
      args: ["report.sh --status {{workflow.status}}"]
  # ...
```

Uses: release external resources that the workflow created (a temporary bucket, an endpoint registration); write a final report; notify an external system, such as a model registry or a CI job.

The proposed API has no exit handler or hook. Run:ai's notifications (email and Slack) may cover a simple notification, but not cleanup. A possible direction is a `spec.onExit` group: a normal group that the composer places in the exit hook, with the workflow's status passed to its tasks as an environment variable. An exit group gets the same quota and scheduling as the other groups, so it can wait in the queue after the rest of the workflow has ended, and its pending timeout matters. The exit handler also affects the stop action, see [Workflow Actions](#workflow-actions).

## Suspend Steps and Workflow Actions

### Suspend Steps

An Argo `suspend` template pauses its branch of the workflow until someone resumes it, or for a fixed `duration`. It creates no pod and holds no resources. With intermediate parameters, the person who resumes it supplies values, for example a choice from an `enum`, that later tasks read as its outputs.

```yaml
- name: approve
  suspend: {}                         # or: suspend: {duration: "30m"}
  inputs:
    parameters:
    - name: decision
      enum: [deploy, skip]
      default: skip
  outputs:
    parameters:
    - name: decision
      valueFrom:
        supplied: {}                  # given when the step is resumed
```

Uses: manual approval before an expensive step (such as a full-scale run after a small trial) or before deploying; a planned delay.

Not covered. Besides a new kind of group in the spec, a suspend step needs a Run:ai action that resumes a specific step and passes its values, and UI for it. Argo resumes a suspend step by updating the Workflow's node status: the CLI's `argo resume --node-field-selector` and `argo node set` do it through functions in Argo's `workflow/util` package, so Run:ai's cluster component would have to do the same.

### Workflow Actions

Run:ai has `suspend`, `resume` and `delete` actions on workloads. In Argo, some operations are spec fields, and others are code in Argo's CLI and server, which Run:ai doesn't install:

| Need | Argo | Notes |
| --- | --- | --- |
| Stop now | `spec.shutdown: Terminate` or `Stop` | Both stop the running pods and end the workflow as Failed. `Stop` still runs the exit handler; `Terminate` doesn't. |
| Suspend, as Run:ai defines it | No direct equivalent | `spec.suspend: true` only stops Argo from starting new tasks: running pods keep running and keep their GPUs. Run:ai's suspend deletes the pods. A closer match is to terminate the workflow, and retry it on resume (below). |
| Resume | `spec.suspend: false`, or a retry after a terminate | |
| Retry a failed workflow from where it failed | Not a spec field | Argo's retry (`FormulateRetryWorkflow` in `workflow/util`) rewrites the Workflow's status: it resets the failed tasks and the tasks after them, keeps the tasks that succeeded, and returns the pods to delete. It works on Failed and Error workflows; a Succeeded workflow needs `restartSuccessful` and a node selector. |
| Run again from the start | `FormulateResubmitWorkflow` | Creates a new Workflow from the same spec. In Run:ai, submitting the spec again does the same. |

Retrying from the failed group matters more for workflows than for single workloads: a long pipeline that fails in its last group shouldn't run its training again. No Run:ai component runs the Argo server, so Run:ai's cluster component would have to vendor Argo's retry logic. The main document discusses this for OSMO's restart and cancel, in [Restart of a Failed Workflow](osmo-to-runai.md#restart-of-a-failed-workflow) and [Cancel](osmo-to-runai.md#cancel).

Gang groups need care in a retry. Argo resets only the tasks that failed, and keeps the other tasks of the group as they are, including the daemons that Argo stopped and marked as Succeeded. The retried task then joins a PodGroup that expects all the group's pods, and can't be scheduled. A retry has to reset the whole group, which Argo's retry supports with `restartSuccessful` and a node selector that matches the group's tasks. The same applies to the terminate-and-retry approach to suspend, and to the open question in [Retry of a Multi-Task Group](osmo-to-runai.md#retry-of-a-multi-task-group). From the v4.1 source (`planReset` in `workflow/util/util.go`); not tested.

## Reuse and Scheduling

### Workflow Parameters and Templates

`spec.arguments.parameters` declares the inputs of a workflow, with defaults, an `enum` and a description, which the spec uses as `{{workflow.parameters.<name>}}`. A WorkflowTemplate (or a cluster-wide ClusterWorkflowTemplate) stores a workflow in the cluster. Users submit it with `workflowTemplateRef` and their own argument values, and other workflows reuse single templates from it with `templateRef`.

Uses: the same pipeline on different datasets, model sizes or learning rates, without editing the spec; a team's standard pipelines; shared building blocks such as an evaluation group.

The proposed API carries the whole spec in each submission. Run:ai already has templates for its other workload types (marked Experimental in the API), so workflow templates fit better there, in the control plane, than as Argo WorkflowTemplates, since the control plane composes the Workflow anyway. Parameters inside the spec, such as a value used in the commands of several tasks, are a separate question that comes with [User Strings and Argo Templating](#user-strings-and-argo-templating).

### Scheduled Workflows

A CronWorkflow creates a Workflow on a schedule. It has:

- one or more cron `schedules`, with a `timezone`;
- a `concurrencyPolicy` that allows, forbids or replaces a run when the previous one is still running;
- `startingDeadlineSeconds` for runs that were missed;
- limits on how many finished runs to keep;
- a `when` condition, and a `stopStrategy` that stops scheduling after a condition, such as a number of failed runs.

Uses: nightly retraining on new data; periodic evaluation against a benchmark; a scheduled data refresh.

No Run:ai workload type has scheduled runs. There are two ways to add them:

1. The control plane submits the workflow on schedule, so each run goes through the usual submission: policies, quota checks and a new workload.
2. The composer creates an Argo CronWorkflow, and Run:ai tracks each Workflow it creates as a workload. Run:ai would then have to adopt workloads that it didn't submit.

Either way, a schedule is a new kind of Run:ai object, separate from the workloads it creates.

### Event-Triggered Workflows

A WorkflowEventBinding submits a WorkflowTemplate when an event reaches the Argo server's event API. It needs the Argo server, so it isn't applicable as it is. A Run:ai equivalent would be an API for triggers. Low priority.

### Memoization

`memoize` on a template skips the task when the cache has an entry with the same `key` (usually built from the task's inputs), and reuses the outputs of that earlier run. The cache is a ConfigMap in the workflow's namespace, with `maxAge` for expiry. Each ConfigMap is limited to 1 MB, and the Argo controller needs permission to create and update ConfigMaps.

Uses: running a pipeline again when its preprocessing hasn't changed.

Not covered. For the same workflow, retrying from the failed group covers much of it (see [Workflow Actions](#workflow-actions)). Across workflows, tasks can skip work today by checking for their outputs on shared storage, as Argo's own [work avoidance](https://argo-workflows.readthedocs.io/en/release-4.1/work-avoidance/) page suggests. Low priority.

## Concurrency Control

### Parallelism

`spec.parallelism` limits how many pods of a workflow run at once, and a template's `parallelism` does the same within one run of the template, for example within one fan-out.

Not covered. Without a fan-out, the shape of the graph bounds how many tasks run at once, so a limit matters less. With a fan-out it is needed, so that a sweep of 200 trials doesn't queue 200 pods at once. Run:ai's training spec already has a `parallelism` field with the same meaning, which the main document rejects for workflows as training-only. It could be kept instead, on the workflow and on a fan-out group.

Note that Run:ai's resource interface for the Argo Workflow kind reads `.spec.parallelism` and `spec.templates[].parallelism` as the workload's replicas (`argo_workflow.go` in the workloads service's migrations), which will need a change for the native type.

### Mutexes and Semaphores

`synchronization` on a workflow or a template takes a mutex (one holder at a time) or a semaphore (a limit kept in a ConfigMap, or in a database when there are several controllers). Workflows that wait for a lock are queued by `priority` and creation time.

Uses: only one workflow at a time writes to a shared dataset or updates a model registry entry; a limit on the load on a shared external service.

Not covered. A lock per project would map naturally to the project's namespace. Low priority.

## Retry and Timeouts

The `retry` section of the proposed API covers the limit, the exit codes and preemption (see [Failure and Retry](osmo-to-runai.md#failure-and-retry) and [Preemption](osmo-to-runai.md#preemption)). These retry and timeout features are left out:

| Argo feature | What it does | Notes |
| --- | --- | --- |
| `retryStrategy.affinity.nodeAntiAffinity: {}` | Runs each retry on a node where the task's earlier attempts didn't run | Useful for GPU failures tied to a node, such as an Xid error. A direct translation, for example a `retry.avoidSameNode` field. How it combines with the node-pool affinity and with KAI needs a test. Medium priority. |
| `retryStrategy.backoff` (`duration`, `factor`, `cap`, `maxDuration`) | Waits between attempts, with an exponential factor and a total limit | For failures of an external service, such as a dataset server or an API. A direct translation. Medium priority. |
| `retryStrategy.expression` on `lastRetry.message` or `lastRetry.duration` | Retries by the failure message, or by how long the attempt ran | `retry.onExitCodes` covers exit codes only. Low priority. |
| `spec.retryStrategy` | A default retry for every task | Could be a `spec.retry` default, like `runTimeoutSeconds`. Low priority. |
| `spec.activeDeadlineSeconds` | A deadline for the whole workflow | `runTimeoutSeconds` limits the run time of each task, so a workflow with many groups has no total limit. Medium priority. |
| Template `timeout` | Limits a task's total time, pending time included | `runTimeoutSeconds` and `pendingTimeoutSeconds` cover it. Not needed. |
| `ttlStrategy.secondsAfterSuccess`, `secondsAfterFailure` | Keeps failed workflows longer than succeeded ones | `autoDeletionTimeAfterCompletionSeconds` translates to `secondsAfterCompletion` only. Low priority. |
| Automatic pod restarts | Restarts a pod that failed before its main container started (Evicted, NodeShutdown, NodeAffinity, UnexpectedAdmissionError) | Not applicable: a setting in the controller's ConfigMap (`failedPodRestart`, off by default, up to 3 restarts), so Run:ai decides it when it installs Argo. It doesn't cover KAI preemption, where the pod is deleted (see [Preemption](osmo-to-runai.md#preemption)). |

## Task Shape

### Multiple Containers in a Task

An Argo template can have `sidecars`, which run next to the main container and are stopped when it ends, and `initContainers`. A `containerSet` template runs several containers in one pod, ordered by `dependencies` and sharing volumes, with its own retry, which runs a failed container's process again in place. The pod requests the sum of its containers' resources.

Uses: a local model server next to an evaluation script; a step that prefetches data before training; a log or metrics agent.

Not covered: a task has one user container, as in the other Run:ai types (the composer adds its own containers, such as the goofys sidecar). A two-task group, with the gang and the headless Service of [Multiple Tasks In a Group](osmo-to-runai.md#multiple-tasks-in-a-group), can stand in for a sidecar, at the cost of a second pod that may land on another node unless `podAffinity` keeps them together. Low priority.

### Services Shared Across Groups

An Argo task with `daemon: true` keeps running while the DAG goes on, and Argo stops it when the DAG that started it ends. The tasks that depend on it start once it is ready, by its readiness probe, and reach it through `{{tasks.<name>.ip}}`. A daemon in the main DAG runs until the workflow ends.

Uses: a model server that several evaluation groups query; a database, cache or replay buffer that later groups use.

The proposed API uses daemons only inside a group, for the leader (see [Multiple Tasks In a Group](osmo-to-runai.md#multiple-tasks-in-a-group)), and `RUNAI_HOST_<TASK>` names only tasks of the same group. A group can't serve the groups that come after it. A possible direction is a service flag on a single-task group. The composer would put that task directly in the main DAG, and not in a group DAG, because Argo stops a daemon when the DAG that started it ends. It would also pass the task's address to the groups that depend on it. The task should have a readiness probe, so that those groups start only when it is ready (`probes.readiness` is covered). Medium priority.

### Per-Group Priority and Preemptibility

Argo sets `priorityClassName` per template, and each gang group has its own PodGroup, with its own priority and preemptibility. In the proposed API, `priorityClass` and `preemptibility` apply to the whole workflow. Per group, a workflow could run its preprocessing and evaluation as preemptible, and only its main training as non-preemptible. Whether this fits Run:ai's quota rules, under which non-preemptible workloads must fit in the project's quota, is a product question. Medium priority.

### Other Template Types

| Argo template | Status | Notes |
| --- | --- | --- |
| `container` | Covered | Every task |
| `dag` | Composer only | The main DAG and the group DAGs |
| `steps` | Not needed | A DAG expresses any sequence of parallel steps |
| `script` | Covered in practice | A script in `args` with `bash -c` or `python -c`, or in `files`. Its result is part of [Parameters and Results](#parameters-and-results) |
| `resource` | Composer only | Creates the PodGroups and Services (see [PodGroup Management](osmo-to-runai.md#podgroup-management)). It shouldn't be open to users: it creates any Kubernetes object with the workflow's service account, outside Run:ai's authorization and policies. This also rules out a task that submits a child workflow (Argo's "workflow of workflows") |
| `suspend` | Not covered | See [Suspend Steps](#suspend-steps) |
| `containerSet` | Not covered | See [Multiple Containers in a Task](#multiple-containers-in-a-task) |
| `http` | Not covered | Sends an HTTP request from Argo's agent pod, which needs its own RBAC for `WorkflowTaskSet` objects. A container task with `curl` does the same. Low priority |
| `plugin` | Not applicable | Needs executor plugins that an admin installs |
| `data` | Not applicable | Reads from an artifact repository |

### Container and Pod Fields

Argo accepts most Kubernetes container and pod fields on a template. These have no Run:ai field in any workload type, so they aren't specific to workflows, and are listed for completeness:

- Container: `livenessProbe` and `startupProbe` (Run:ai supports a readiness probe only), `lifecycle` (`postStart` and `preStop`), `envFrom`, `volumeDevices`, `resizePolicy`, and `terminationMessagePath` and `terminationMessagePolicy`.
- Pod: `dnsConfig`, `hostAliases`, `automountServiceAccountToken`, and `podResources` and DRA `resourceClaims`, both new in v4.1.
- Workflow: `podDisruptionBudget`.

## Observability

| Argo feature | Status | Notes |
| --- | --- | --- |
| Progress (`status.progress`, and `ARGO_PROGRESS_FILE` for a task's own progress) | Not covered | Argo counts finished tasks out of all tasks, and a task can report its own progress by writing to the file in `ARGO_PROGRESS_FILE`. [Workflow Status](osmo-to-runai.md#workflow-status) doesn't show it yet, but could. Low priority |
| Debug pause (`ARGO_DEBUG_PAUSE_BEFORE`, `ARGO_DEBUG_PAUSE_AFTER`) | Covered | Environment variables that pause a task before or after its command, until a marker file is created, for debugging through exec. They should work through `environmentVariables` (not tested) |
| Log archiving (`archiveLogs`) | Not covered | Needs an artifact repository. Without it, a task's logs live as long as its pod, which Argo keeps until the Workflow is deleted, unless `podGC` is set. Low priority, but OSMO keeps logs after the workflow ends, see [Logs After the Workflow Ends](osmo-to-runai.md#logs-after-the-workflow-ends) |
| Custom Prometheus metrics (`metrics` on the workflow or a template) | Not applicable | Exported by the Argo controller; Run:ai has its own metrics |
| `estimatedDuration` | Not applicable | Computed from earlier runs of the same WorkflowTemplate or CronWorkflow |
| Title and description annotations, links in the Argo UI | Not applicable | Run:ai has its own UI |

## Covered by the Composer Only

The composer uses these Argo features to implement Run:ai fields (see [Fields Translation](osmo-to-runai.md#fields-translation)). Users can't set them directly:

| Argo feature | Used for |
| --- | --- |
| `entrypoint`, `templates`, `dag`, `dependencies` | Groups, tasks and `dependsOn` |
| `daemon`, `continueOn` | `leader` and `failurePolicy` |
| `resource` templates | PodGroups and headless Services |
| `podSpecPatch` | Hostname and subdomain, `hostNetwork`, `hostIpc`, `terminationGracePeriodSeconds` |
| `withSequence`, `withItems` | `replicas` and segments |
| `initContainers` | The goofys sidecar |
| `volumes`, `volumeClaimTemplates` | Storage |
| `inputs.artifacts` with `git` and `raw` | Git storage and `files` |
| `retryStrategy` | `retry` and `terminateAfterPreemption` |
| `activeDeadlineSeconds`, `pendingTimeout` | `runTimeoutSeconds`, `pendingTimeoutSeconds` |
| `ttlStrategy.secondsAfterCompletion` | `autoDeletionTimeAfterCompletionSeconds` |
| `podMetadata`, `podPriorityClassName`, `imagePullSecrets`, template `metadata`, `affinity`, `tolerations`, `securityContext` | General, placement and security fields |
| `{{workflow.name}}` | Per-workflow directories, and the names of PodGroups and Services |

The main document doesn't decide yet on these settings, which the composer should set explicitly rather than leave to Argo's defaults:

- `failFast`, and `dnsPolicy` with `hostNetwork`, discussed in the next chapter;
- `podGC`: Argo keeps pods after they end by default, so their logs stay available until the Workflow is deleted, at the cost of finished pods in the project's namespace;
- `serviceAccountName`, which needs permission to create PodGroups and Services (see [PodGroup Management](osmo-to-runai.md#podgroup-management));
- `templateDefaults`, which may hold the parts that all tasks share once, to keep the Workflow small (see [Large Workflow Status](#large-workflow-status); to be tested).

## Not Applicable

| Argo feature | Why |
| --- | --- |
| Executor plugins, `plugin` templates | Installed by an admin, not part of a workflow |
| `artifactRepositoryRef`, `archiveLocation`, `artifactGC` | No artifact repository in the design |
| `priority` | Orders only the workflows that wait for the Argo controller's own limits or for a lock; KAI decides what runs |
| The Argo server, UI, SSO and workflow archive | The Run:ai control plane provides the API, UI and history |
| `workflowRestrictions` (`templateReferencing`) | Restricts the Workflows that users submit directly; with the composer, users don't |
| Controller settings: `parallelism`, `namespaceParallelism`, `workflowDefaults`, `failedPodRestart`, node status offloading, `initlessPod` | Set when Run:ai installs Argo. Some of them affect the composer, see [Large Workflow Status](#large-workflow-status) |
| Custom metrics, `estimatedDuration` | See [Observability](#observability) |

## Findings That Affect the Main Document

These points came up during the review. They don't change what the API covers, but they affect the design in the main document, and most of them are now discussed there. Unless noted, none of them was tested.

### Gang Eviction After a Task Fails

KAI evicts all the running pods of a PodGroup that has running pods, fewer of them than its `minMember`, and no pod that has succeeded, after a grace period (60 seconds by default). In a workflow group, this happens when a non-leader task fails in a group with `failurePolicy: Leader`, and when a retried task's new pod isn't running within the grace period. The main document has the composer turn the eviction off with a negative `stalenessGracePeriod` on the PodGroups it creates, see [PodGroup Management](osmo-to-runai.md#podgroup-management). For `failurePolicy: AnyTask`, where the eviction would be the only thing that stops the group, see [Group Failure Policy](osmo-to-runai.md#group-failure-policy). From `IsStale` in KAI's `podgroup_info/job_info.go` and its `stalegangeviction` action, and from Run:ai's [Stale Gang Eviction](../../docs/platform-management/runai-scheduler/scheduling/concepts-and-principles.md#stale-gang-eviction) docs.

### User Strings and Argo Templating

Argo substitutes `{{...}}` tags in every string of a template before it runs, including the user's command and arguments, and, as far as we can tell from the code, environment variable values and the content of `files`. From the v4.1 source:

- a tag that Argo knows, such as `{{workflow.name}}` or `{{pod.name}}`, is replaced with its value;
- a tag that starts with one of Argo's prefixes (`workflow.`, `inputs.`, `outputs.`, `tasks.`, `steps.`, `item.`, `pod.`) but has no value fails the workflow's validation, with "failed to resolve {{...}}";
- `{{=...}}` is evaluated as an expression, unless the controller runs with `EXPRESSION_TEMPLATES=false`;
- other tags, such as Jinja or Helm syntax in a script, are left as they are.

So a user's script can fail the workflow by accident, or read values that the API doesn't document. The design should decide whether some Argo variables become part of the Run:ai API (for example the workflow name, or the retry attempt as `{{retries}}`), and how the composer handles other `{{` in user strings. We didn't find a documented escape in Argo. The main document covers only the case of Run:ai's own template variables, which must be rendered before the composer, in [Templates and Apps](osmo-to-runai.md#templates-and-apps).

A related naming rule: Argo rejects a DAG task whose name begins with a digit (`validate.go`), so the composer has to reject or rename groups and tasks with such names.

### failFast

Argo's documentation says that with `failFast` (the default), a failed DAG task stops Argo from starting new tasks. In the v4.1 code (`executeDAG` and `assessDAGPhase` in `dag.go`), Argo starts every task whose dependencies are met, and doesn't end the DAG while any of its tasks is running. `failFast` only decides whether a failed branch fails the DAG before the tasks that haven't started yet, which in practice means tasks held back by `parallelism`. The main document now covers both effects: independent groups keep running after a failure, as its Argo run shows (see [Failure and Retry](osmo-to-runai.md#failure-and-retry)), and a failed task doesn't stop a group with `failurePolicy: AnyTask` (see [Group Failure Policy](osmo-to-runai.md#group-failure-policy)).

If `parallelism` is added, the composer should set `failFast` explicitly. To be tested.

### Large Workflow Status

When a Workflow grows close to etcd's object size limit (about 1 MB), Argo compresses `status.nodes` into `status.compressedNodes`, with gzip by default (v4.1 adds zstd and brotli, through `WORKFLOW_COMPRESSION_ALGORITHM`), and with node status offloading it moves the nodes to a database. The spec counts toward the same limit. Now in [Large Workflows](osmo-to-runai.md#large-workflows).

### hostNetwork and DNS

The main document sets `hostNetwork` through `podSpecPatch` (see [Security](osmo-to-runai.md#security)). With `hostNetwork` and the default `dnsPolicy`, a pod uses the node's DNS, and can't resolve cluster names such as the names in `RUNAI_HOST_<TASK>`. Run:ai's composer for the other types sets `dnsPolicy: ClusterFirstWithHostNet` whenever `hostNetwork` is true (`ComposeDnsPolicy` in `control-plane/workloads-and-policies/composer/fields_composer_security.go`), so the workflow's `podSpecPatch` should include it too.

### Jinja Loops and Replicas

The main document maps a Jinja loop inside a group to `tasks[].replicas`, and a loop in OSMO's flat `tasks` list, which creates independent tasks, to `groups[].copies` (see [General Fields](osmo-to-runai.md#general-fields)). A fan-out beyond a fixed count is discussed in [Fan-Out of Independent Tasks](#fan-out-of-independent-tasks).
