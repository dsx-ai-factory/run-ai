#!/usr/bin/env python3
# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0
"""API-faithful synthetic Run:ai MCP fixture for workload health evals.

This is deterministic and network-free. It mirrors the current MCP schemas but
is not the live-recorded seeded-cluster fixture required for the ticket's v2.
"""

import json
import sys


CLUSTER_ID = "20000000-0000-4000-8000-000000000001"
PROJECT_ID = "4500002"
PROJECT_NAME = "ml-team"
DEPARTMENT_NAME = "Research"
HEALTHY_ID = "00000000-0000-4000-8000-000000000042"
OLDER_HEALTHY_ID = "00000000-0000-4000-8000-000000000041"
DEGRADED_ID = "00000000-0000-4000-8000-000000000043"
PENDING_ID = "00000000-0000-4000-8000-000000000044"
NO_POLICY_ID = "00000000-0000-4000-8000-000000000045"
NO_METRICS_ID = "00000000-0000-4000-8000-000000000046"
COMPLETED_ID = "00000000-0000-4000-8000-000000000047"
INFERENCE_ID = "00000000-0000-4000-8000-000000000102"
FAILED_IDS = [
    "00000000-0000-4000-8000-000000000101",
    INFERENCE_ID,
    "00000000-0000-4000-8000-000000000103",
    "00000000-0000-4000-8000-000000000104",
]


def node_pool_for(node_name):
    """Map a fixture node name to its matching synthetic node pool."""
    if node_name.startswith("node-b-"):
        return "node-pool-b"
    if node_name.startswith("node-c-"):
        return "node-pool-c"
    return "node-pool-a"


def row(
    workload_id,
    name,
    phase,
    kind,
    project,
    department,
    node,
    changed_at,
    age,
    *,
    priority="medium",
    preemptible=True,
    gpu_idle_minutes=0,
):
    """Build one upstream-like workload record plus private fixture metadata."""
    node_pool = node_pool_for(node)
    return {
        "id": workload_id,
        "name": name,
        "phase": phase,
        "type": "Training" if kind == "Train" else kind,
        "projectName": project,
        "projectId": {
            "ml-team": PROJECT_ID,
            "no-policy-team": "4500099",
            "finance-prod": "4500003",
            "sandbox-eval": "4500004",
            "cv-team": "4500005",
        }[project],
        "clusterId": CLUSTER_ID,
        "createdAt": "2026-05-31T00:00:00Z",
        "phaseUpdatedAt": changed_at,
        "requestedNodePools": [node_pool],
        "currentNodePools": [node_pool] if node else [],
        "priority": {"low": 50, "medium": 100, "high": 150}[priority],
        "priorityClassName": priority,
        "preemptible": preemptible,
        "idleAllocatedGpus": 1 if gpu_idle_minutes else 0,
        "allocatedGpu": 4,
        "requestedGpu": 4,
        "_departmentName": department,
        "_nodeName": node,
        "_age": age,
        "_gpuIdleMinutes": gpu_idle_minutes,
    }


WORKLOAD_ROWS = [
    row(
        FAILED_IDS[0],
        "train-llama-9",
        "Failed",
        "Train",
        "ml-team",
        "Research",
        "node-a-04",
        "2026-05-31T10:47:18Z",
        "6m",
    ),
    row(
        INFERENCE_ID,
        "inference-prod-9",
        "Failed",
        "Inference",
        "finance-prod",
        "Production",
        "node-a-04",
        "2026-05-31T10:47:22Z",
        "6m",
        priority="high",
        preemptible=False,
    ),
    row(
        FAILED_IDS[2],
        "eval-batch-3",
        "Failed",
        "Train",
        "sandbox-eval",
        "Research",
        "node-a-04",
        "2026-05-31T10:47:25Z",
        "5m",
        priority="low",
    ),
    row(
        FAILED_IDS[3],
        "trainer-cv-2",
        "Failed",
        "Train",
        "cv-team",
        "Research",
        "node-b-12",
        "2026-05-31T10:51:09Z",
        "2m",
    ),
    row(
        HEALTHY_ID,
        "training-job-42",
        "Running",
        "Distributed",
        PROJECT_NAME,
        DEPARTMENT_NAME,
        "node-a-04",
        "2026-05-31T08:16:00Z",
        "3h12m",
    ),
    row(
        OLDER_HEALTHY_ID,
        "training-job-42",
        "Completed",
        "Distributed",
        PROJECT_NAME,
        DEPARTMENT_NAME,
        "node-a-05",
        "2026-05-30T18:00:00Z",
        "17h28m",
    ),
    row(
        DEGRADED_ID,
        "idle-eval-7",
        "Running",
        "Train",
        PROJECT_NAME,
        DEPARTMENT_NAME,
        "node-c-02",
        "2026-05-31T00:00:00Z",
        "14h",
        gpu_idle_minutes=480,
    ),
    row(
        PENDING_ID,
        "pending-startup",
        "Pending",
        "Train",
        PROJECT_NAME,
        DEPARTMENT_NAME,
        "",
        "2026-05-31T10:50:00Z",
        "3m",
    ),
    row(
        NO_POLICY_ID,
        "no-policy-snapshot",
        "Running",
        "Train",
        "no-policy-team",
        DEPARTMENT_NAME,
        "node-a-06",
        "2026-05-31T09:00:00Z",
        "2h28m",
    ),
    row(
        NO_METRICS_ID,
        "no-metrics-snapshot",
        "Running",
        "Train",
        PROJECT_NAME,
        DEPARTMENT_NAME,
        "node-a-07",
        "2026-05-31T09:15:00Z",
        "2h13m",
    ),
    row(
        COMPLETED_ID,
        "completed-success",
        "Completed",
        "Training",
        PROJECT_NAME,
        DEPARTMENT_NAME,
        "node-a-08",
        "2026-05-31T10:00:00Z",
        "1h28m",
    ),
]


STATUS = {
    HEALTHY_ID: {
        "id": HEALTHY_ID,
        "clusterId": CLUSTER_ID,
        "name": "training-job-42",
        "phase": "Running",
        "projectName": PROJECT_NAME,
        "departmentName": DEPARTMENT_NAME,
        "kind": "Distributed",
        "submittedAt": "2026-05-31T08:14:00Z",
        "startedAt": "2026-05-31T08:16:00Z",
        "lastTransitionTime": "2026-05-31T08:16:00Z",
        "events": [
            {
                "createdAt": "2026-05-31T10:50:00Z",
                "reason": "ScalingReplicaSet",
                "message": "ReplicaSet is stable.",
            }
        ],
    },
    DEGRADED_ID: {
        "id": DEGRADED_ID,
        "clusterId": CLUSTER_ID,
        "name": "idle-eval-7",
        "phase": "Running",
        "projectName": PROJECT_NAME,
        "departmentName": DEPARTMENT_NAME,
        "kind": "Train",
        "submittedAt": "2026-05-31T00:00:00Z",
        "startedAt": "2026-05-31T00:04:00Z",
        "lastTransitionTime": "2026-05-31T00:04:00Z",
        "events": [
            {
                "createdAt": "2026-05-31T08:00:00Z",
                "reason": "LowGpuUtilization",
                "message": "GPU utilization below 5% for 8h.",
            }
        ],
    },
    INFERENCE_ID: {
        "id": INFERENCE_ID,
        "clusterId": CLUSTER_ID,
        "name": "inference-prod-9",
        "phase": "Failed",
        "projectName": "finance-prod",
        "departmentName": "Production",
        "kind": "Inference",
        "submittedAt": "2026-05-31T10:30:00Z",
        "startedAt": "2026-05-31T10:31:00Z",
        "lastTransitionTime": "2026-05-31T10:47:22Z",
        "events": [
            {
                "createdAt": "2026-05-31T10:47:22Z",
                "reason": "Failed",
                "message": "Inference workload failed; use diagnosis for root cause.",
            }
        ],
    },
    PENDING_ID: {
        "id": PENDING_ID,
        "clusterId": CLUSTER_ID,
        "name": "pending-startup",
        "phase": "Pending",
        "projectName": PROJECT_NAME,
        "departmentName": DEPARTMENT_NAME,
        "kind": "Train",
        "submittedAt": "2026-05-31T10:50:00Z",
        "lastTransitionTime": "2026-05-31T10:50:00Z",
        "events": [
            {
                "createdAt": "2026-05-31T11:21:00Z",
                "reason": "Pending",
                "message": "Waiting for initialization; no direct startup failure reported.",
            }
        ],
    },
    NO_POLICY_ID: {
        "id": NO_POLICY_ID,
        "clusterId": CLUSTER_ID,
        "name": "no-policy-snapshot",
        "phase": "Running",
        "projectName": "no-policy-team",
        "departmentName": DEPARTMENT_NAME,
        "kind": "Train",
        "submittedAt": "2026-05-31T09:00:00Z",
        "startedAt": "2026-05-31T09:02:00Z",
        "lastTransitionTime": "2026-05-31T09:02:00Z",
        "events": [],
    },
    NO_METRICS_ID: {
        "id": NO_METRICS_ID,
        "clusterId": CLUSTER_ID,
        "name": "no-metrics-snapshot",
        "phase": "Running",
        "projectName": PROJECT_NAME,
        "departmentName": DEPARTMENT_NAME,
        "kind": "Train",
        "submittedAt": "2026-05-31T09:15:00Z",
        "startedAt": "2026-05-31T09:17:00Z",
        "lastTransitionTime": "2026-05-31T09:17:00Z",
        "events": [],
    },
    COMPLETED_ID: {
        "id": COMPLETED_ID,
        "clusterId": CLUSTER_ID,
        "name": "completed-success",
        "phase": "Succeeded",
        "projectName": PROJECT_NAME,
        "departmentName": DEPARTMENT_NAME,
        "kind": "Train",
        "submittedAt": "2026-05-31T09:40:00Z",
        "startedAt": "2026-05-31T09:42:00Z",
        "lastTransitionTime": "2026-05-31T10:00:00Z",
        "events": [
            {
                "createdAt": "2026-05-31T10:00:00Z",
                "reason": "Succeeded",
                "message": "Workload completed successfully.",
            }
        ],
    },
}

FAILED_STATUS_SIGNALS = {
    FAILED_IDS[0]: {
        "phaseReason": "NodeLost",
        "phaseMessage": "Pod terminated after node-a-04 became NotReady.",
    },
    INFERENCE_ID: {
        "phaseReason": "NodeLost",
        "phaseMessage": "Inference pod terminated after node-a-04 became NotReady.",
    },
    FAILED_IDS[2]: {
        "phaseReason": "NodeLost",
        "phaseMessage": "Pod terminated after node-a-04 became NotReady.",
    },
    FAILED_IDS[3]: {
        "phaseReason": "BackoffLimitExceeded",
        "phaseMessage": "Container exited with exit code 1 until retries were exhausted.",
    },
}


def base_spec(workload_id, name):
    """Return a compact spec whose placement matches the fixture's pod data."""
    workload = next(item for item in WORKLOAD_ROWS if item["id"] == workload_id)
    placements = sorted(
        {
            item["nodeName"]
            for item in PODS.get(workload_id, {"pods": []})["pods"]
            if item.get("nodeName")
        }
    )
    return {
        "id": workload_id,
        "name": name,
        "clusterId": CLUSTER_ID,
        "compute": {"gpu": 4, "gpuType": "h100", "cpu": 16, "memory": "256Gi"},
        "scheduling": {
            "nodePools": workload["requestedNodePools"],
            "priority": "medium",
            "preemptible": True,
        },
        "runtime": {
            "image": "nvcr.io/nvidia/pytorch:24.05-py3",
            "allocated_resources": {
                "gpu": 4,
                "gpuMemory": "320Gi",
                "cpu": 16,
                "cpuMemory": "256Gi",
            },
            "actual_placement": placements,
        },
        "environment": {"variables": []},
        "dataSources": [{"name": "training-data", "mountPath": "/data"}],
    }


def pod(name, node, phase):
    """Build a pod using only fields exposed by the current MCP pod tools."""
    return {
        "name": name,
        "nodeName": node,
        "k8sPhase": phase,
    }


PODS = {
    FAILED_IDS[0]: {
        "pods": [pod("train-llama-9-0", "node-a-04", "Failed")],
    },
    FAILED_IDS[2]: {
        "pods": [pod("eval-batch-3-0", "node-a-04", "Failed")],
    },
    FAILED_IDS[3]: {
        "pods": [pod("trainer-cv-2-0", "node-b-12", "Failed")],
    },
    HEALTHY_ID: {
        "pods": [
            pod("training-job-42-0", "node-a-04", "Running"),
            pod("training-job-42-1", "node-a-04", "Running"),
            pod("training-job-42-2", "node-a-05", "Running"),
            pod("training-job-42-3", "node-a-05", "Running"),
        ],
    },
    DEGRADED_ID: {
        "pods": [pod("idle-eval-7-0", "node-c-02", "Running")],
    },
    INFERENCE_ID: {
        "pods": [
            pod("inference-prod-9-0", "node-a-04", "Failed"),
            pod("inference-prod-9-1", "node-a-04", "Failed"),
        ],
    },
    PENDING_ID: {
        "pods": [pod("pending-startup-0", "", "Pending")],
    },
    NO_POLICY_ID: {
        "pods": [pod("no-policy-snapshot-0", "node-a-06", "Running")],
    },
    NO_METRICS_ID: {
        "pods": [pod("no-metrics-snapshot-0", "node-a-07", "Running")],
    },
    COMPLETED_ID: {
        "pods": [pod("completed-success-0", "node-a-08", "Succeeded")],
    },
}


METRICS = {
    HEALTHY_ID: {
        "window": "1h",
        "gpu_utilization": {"avg": 87, "p95": 94},
        "gpu_memory": {"usedGiB": 78, "limitGiB": 80},
        "network_egress_gbps": 4.2,
        "gpu_idle_minutes": 0,
    },
    DEGRADED_ID: {
        "window": "1h",
        "gpu_utilization": {"avg": 2, "p95": 5},
        "gpu_memory": {"usedGiB": 4, "limitGiB": 80},
        "network_egress_gbps": 0.1,
        "gpu_idle_minutes": 480,
    },
    INFERENCE_ID: {
        "window": "1h",
        "gpu_utilization": {"avg": 22, "p95": 41},
        "gpu_memory": {"usedGiB": 12, "limitGiB": 80},
        "network_egress_gbps": 1.1,
        "request_rate": 320,
        "latency_ms": {"p50": 85, "p95": 420, "p99": 780, "target": 200},
        "error_rate": 0.08,
    },
    NO_POLICY_ID: {
        "window": "1h",
        "gpu_utilization": {"avg": 64, "p95": 77},
        "gpu_memory": {"usedGiB": 20, "limitGiB": 80},
        "network_egress_gbps": 0.6,
    },
    COMPLETED_ID: {
        "window": "1h",
        "measurements": [],
    },
}


def matches(row_data, args):
    """Apply the subset of current list_workloads filters used by the evals."""
    if args.get("cluster_id") and args["cluster_id"] != CLUSTER_ID:
        return False
    if args.get("clusterId") and args["clusterId"] != CLUSTER_ID:
        return False
    if args.get("phase") and row_data["phase"].lower() != args["phase"].lower():
        return False
    if args.get("projectId") and row_data["projectId"] != args["projectId"]:
        return False
    project_filter = args.get("projectName")
    if project_filter and row_data["projectName"] != comparison_value(project_filter, "projectName"):
        return False
    department_filter = args.get("departmentName")
    if department_filter and row_data["_departmentName"] != comparison_value(
        department_filter, "departmentName"
    ):
        return False
    if args.get("name") and row_data["name"] != args["name"]:
        return False
    if args.get("type") and row_data["type"].lower() != args["type"].lower():
        return False
    if args.get("phaseUpdatedAfter") and row_data["phaseUpdatedAt"] < args["phaseUpdatedAfter"]:
        return False
    for item in args.get("filterBy", []):
        if not matches_typed_filter(row_data, item):
            return False
    if args.get("gpuIdleOnly") and (
        row_data["phase"] != "Running" or not row_data["idleAllocatedGpus"]
    ):
        return False
    return True


def workload_field(row_data, field):
    """Return a sortable/filterable workload field, including API aliases."""
    aliases = {
        "nodepool": "currentNodePools",
        "nodePoolName": "currentNodePools",
        "allocatedGPU": "allocatedGpu",
        "preemptibility": "preemptible",
        "departmentName": "_departmentName",
    }
    value = row_data.get(aliases.get(field, field))
    if isinstance(value, list):
        return value[0] if value else None
    return value


def matches_typed_filter(row_data, item):
    """Apply one current-MCP-shaped typed list_workloads filter."""
    if not isinstance(item, dict):
        raise TypeError("filterBy entries must be objects")
    field = item.get("field")
    operator = item.get("operator")
    if not field or operator not in (">", ">=", "<", "<=", "==", "!="):
        raise ValueError("filterBy entries require a field and valid operator")
    if "value" not in item:
        raise KeyError("filterBy.value")
    actual = workload_field(row_data, field)
    expected = item["value"]
    if actual is None:
        return False
    comparisons = {
        ">": lambda: actual > expected,
        ">=": lambda: actual >= expected,
        "<": lambda: actual < expected,
        "<=": lambda: actual <= expected,
        "==": lambda: actual == expected,
        "!=": lambda: actual != expected,
    }
    try:
        return comparisons[operator]()
    except TypeError as error:
        raise TypeError(f"filterBy value type does not match {field}") from error


def comparison_value(value, field):
    """Validate and return a structured comparison value."""
    if not isinstance(value, dict):
        raise TypeError(f"{field} must be an object")
    if value.get("operator") not in (None, "=="):
        raise ValueError(f"{field}.operator must be ==")
    if "value" not in value:
        raise KeyError(f"{field}.value")
    return value["value"]


DETAIL_FIELDS = {
    "createdAt",
    "requestedNodePools",
    "currentNodePools",
    "priority",
    "priorityClassName",
    "preemptible",
    "idleAllocatedGpus",
}

SORT_FIELDS = {
    "type",
    "name",
    "clusterId",
    "projectId",
    "projectName",
    "departmentId",
    "departmentName",
    "createdAt",
    "deletedAt",
    "submittedBy",
    "phase",
    "completedAt",
    "nodepool",
    "nodePoolName",
    "distributedFramework",
    "allocatedGPU",
    "allocatedGpu",
    "idleAllocatedGpus",
    "phaseUpdatedAt",
    "category",
    "priority",
    "totalPendingTimeSeconds",
    "totalRunningTimeSeconds",
    "priorityClassName",
    "preemptibility",
    "guaranteedRuntimeEndsAt",
    "aiApplicationId",
    "aiApplicationName",
}


def public_row(item, include_details=False):
    """Map one record to the compact or detailed current-MCP row shape."""
    return {
        key: value
        for key, value in item.items()
        if not key.startswith("_") and (include_details or key not in DETAIL_FIELDS)
    }


def limited(rows, args):
    """Apply list_workloads offset/limit pagination."""
    limit = int(args.get("limit", 50))
    offset = int(args.get("offset", 0))
    if limit < 0 or limit > 200 or offset < 0:
        raise ValueError("limit must be between 0 and 200 and offset must be non-negative")
    if limit == 0:
        limit = 50
    page = rows[offset : offset + limit]
    has_more = offset + limit < len(rows)
    result = {
        "workloads": [
            public_row(item, bool(args.get("includeDetails"))) for item in page
        ],
        "hasMore": has_more,
    }
    if has_more:
        result["nextOffset"] = offset + limit
    return result


def list_workloads(args):
    """Return filtered, sorted, paginated workload rows."""
    rows = [item for item in WORKLOAD_ROWS if matches(item, args)]
    sort_by = args.get("sortBy", "phaseUpdatedAt")
    if sort_by not in SORT_FIELDS:
        raise ValueError(f"unsupported fixture sortBy {sort_by}")
    sort_order = args.get("sortOrder", "desc")
    if sort_order not in ("asc", "desc"):
        raise ValueError("sortOrder must be asc or desc")
    present = [item for item in rows if workload_field(item, sort_by) is not None]
    missing = [item for item in rows if workload_field(item, sort_by) is None]
    present.sort(
        key=lambda item: workload_field(item, sort_by),
        reverse=sort_order == "desc",
    )
    rows = present + missing
    return limited(rows, args)


def get_workloads_summary(args):
    """Return deterministic phase counts for a scoped workload set."""
    rows = [item for item in WORKLOAD_ROWS if matches(item, args)]
    if args.get("orgType") == "project":
        rows = [item for item in rows if item["projectName"] == args.get("orgName")]
    if args.get("orgType") == "department":
        rows = [item for item in rows if item["_departmentName"] == args.get("orgName")]
    buckets = {"running": 0, "pending": 0, "failed": 0, "completed": 0, "stopped": 0, "other": 0}
    by_phase = {}
    for item in rows:
        phase = item["phase"]
        by_phase[phase] = by_phase.get(phase, 0) + 1
        bucket = {"Running": "running", "Pending": "pending", "Failed": "failed", "Degraded": "failed", "Completed": "completed", "Stopped": "stopped"}.get(phase, "other")
        buckets[bucket] += 1
    return {"scope": {"clusterId": args.get("clusterId", CLUSTER_ID)}, "asOf": "2026-05-31T11:30:00Z", "total": len(rows), "counts": buckets, "byPhase": by_phase}


def get_node_pods(args):
    """Return one paginated current-MCP-shaped page of pods on a node."""
    node_name = args["nodeName"]
    limit = int(args.get("limit", 10)) or 10
    offset = int(args.get("offset", 0))
    if limit < 0 or limit > 200 or offset < 0:
        raise ValueError("limit must be between 0 and 200 and offset must be non-negative")
    rows_by_id = {item["id"]: item for item in WORKLOAD_ROWS}
    all_pods = []
    for workload_id, workload_pods in PODS.items():
        workload = rows_by_id[workload_id]
        pod_count = max(len(workload_pods["pods"]), 1)
        pod_gpu = 4 / pod_count
        pod_cpu = 16 / pod_count
        pod_gpu_memory_gib = 320 / pod_count
        pod_cpu_memory_gib = 256 / pod_count
        for index, item in enumerate(workload_pods["pods"]):
            if item.get("nodeName") != node_name:
                continue
            all_pods.append(
                {
                    "name": item["name"],
                    "id": f"pod-{workload_id}-{index}",
                    "nodeName": node_name,
                    "currentNodePool": workload["currentNodePools"][0],
                    "k8sPhase": item["k8sPhase"],
                    "createdAt": "2026-05-31T10:30:00Z",
                    "clusterId": CLUSTER_ID,
                    "requestedResources": {
                        "gpu": {"request": pod_gpu, "limit": pod_gpu},
                        "gpuMemory": {
                            "request": f"{pod_gpu_memory_gib:g}Gi",
                            "limit": f"{pod_gpu_memory_gib:g}Gi",
                        },
                        "cpu": {"request": pod_cpu, "limit": pod_cpu},
                        "cpuMemory": {
                            "request": f"{pod_cpu_memory_gib:g}Gi",
                            "limit": f"{pod_cpu_memory_gib:g}Gi",
                        },
                    },
                    "allocatedResources": {
                        "gpu": pod_gpu,
                        "gpuMemory": f"{pod_gpu_memory_gib:g}Gi",
                        "cpu": pod_cpu,
                        "cpuMemory": f"{pod_cpu_memory_gib:g}Gi",
                    },
                    "workloadId": workload_id,
                    "workloadName": workload["name"],
                    "workloadType": workload["type"],
                }
            )
    all_pods.sort(key=lambda item: item["name"])
    page = all_pods[offset : offset + limit]
    has_more = offset + limit < len(all_pods)
    node_pool = node_pool_for(node_name)
    return {
        "name": node_name,
        "nodePool": node_pool,
        "status": "NotReady" if node_name == "node-a-04" else "Ready",
        "gpuInfo": {"gpuType": "NVIDIA H100 80GB HBM3", "gpuCount": 8},
        "pods": page,
        "hasMore": has_more,
        **({"nextOffset": offset + limit} if has_more else {}),
    }


def list_projects(args):
    """Return project IDs used by policy lookup cases."""
    projects = []
    seen = set()
    for item in WORKLOAD_ROWS:
        if item["projectId"] in seen:
            continue
        seen.add(item["projectId"])
        projects.append({"id": item["projectId"], "name": item["projectName"], "clusterId": CLUSTER_ID, "departmentId": "5500001", "phase": "Ready"})
    return {"projects": projects, "totalCount": len(projects), "hasMore": False}


def get_workload_status(args):
    """Return current-MCP-shaped status with diagnostic phase fields."""
    workload_id = args["workloadId"]
    if workload_id in STATUS:
        raw = STATUS[workload_id]
        result = {
            "id": raw["id"],
            "name": raw["name"],
            "clusterId": raw["clusterId"],
            "type": raw.get("type", raw.get("kind", "Training")).replace("Train", "Training"),
            "phase": "Completed" if raw["phase"] == "Succeeded" else raw["phase"],
            "projectName": raw["projectName"],
            "departmentName": raw.get("departmentName", ""),
            "createdAt": raw.get("createdAt", raw.get("submittedAt", "")),
            "phaseUpdatedAt": raw.get("phaseUpdatedAt", raw.get("lastTransitionTime", "")),
            "conditions": raw.get("conditions", []),
            "events": raw.get("events", []),
        }
        result.update(FAILED_STATUS_SIGNALS.get(workload_id, {}))
        return result
    for item in WORKLOAD_ROWS:
        if item["id"] == workload_id:
            result = {
                "id": workload_id,
                "clusterId": CLUSTER_ID,
                "name": item["name"],
                "phase": item["phase"],
                "projectName": item["projectName"],
                "type": item["type"],
                "departmentName": item["_departmentName"],
                "createdAt": item["createdAt"],
                "lastTransitionTime": item["phaseUpdatedAt"],
                "phaseUpdatedAt": item["phaseUpdatedAt"],
                "conditions": [],
                "events": [],
            }
            result.update(FAILED_STATUS_SIGNALS.get(workload_id, {}))
            return result
    raise ValueError(f"unknown workloadId {workload_id}")


def get_workload_spec(args):
    """Return a compact spec with consistent placement and memory allocation."""
    status = get_workload_status(args)
    raw = base_spec(args["workloadId"], status["name"])
    return {
        "workloadId": args["workloadId"],
        "name": status["name"],
        "type": status["type"],
        "projectName": status["projectName"],
        "departmentName": status.get("departmentName", ""),
        "spec": {"compute": raw["compute"], "scheduling": raw["scheduling"], "runtime": raw["runtime"], "dataSources": raw["dataSources"]},
        "allocatedResources": raw["runtime"]["allocated_resources"],
        "topology": (
            {
                "actualPlacement": {
                    "level": "node" if len(raw["runtime"]["actual_placement"]) == 1 else "nodePool",
                    "value": (
                        raw["runtime"]["actual_placement"][0]
                        if len(raw["runtime"]["actual_placement"]) == 1
                        else raw["scheduling"]["nodePools"][0]
                    ),
                }
            }
            if raw["runtime"]["actual_placement"]
            else {}
        ),
        "pendingSchedulingMessages": [],
    }


def get_workload_pods(args):
    """Return a paginated pod page without Kubernetes-only readiness fields."""
    workload_id = args["workloadId"]
    status = get_workload_status(args)
    limit = int(args.get("limit", 20)) or 20
    offset = int(args.get("offset", 0))
    if limit < 0 or limit > 500 or offset < 0:
        raise ValueError("limit must be between 0 and 500 and offset must be non-negative")
    if workload_id not in PODS:
        return {"workloadId": workload_id, "type": status["type"], "pods": [], "hasMore": False}
    pods = []
    workload = next(item for item in WORKLOAD_ROWS if item["id"] == workload_id)
    pod_count = max(len(PODS[workload_id]["pods"]), 1)
    pod_gpu = 4 / pod_count
    pod_cpu = 16 / pod_count
    pod_cpu_memory_gib = 256 / pod_count
    for index, item in enumerate(PODS[workload_id]["pods"]):
        pods.append({
            "name": item["name"],
            "id": f"pod-{workload_id}-{index}",
            "nodeName": item.get("nodeName", ""),
            "currentNodePool": (
                workload["currentNodePools"][0]
                if item.get("nodeName") and workload["currentNodePools"]
                else ""
            ),
            "k8sPhase": item["k8sPhase"],
            "createdAt": "2026-05-31T08:16:00Z",
            "clusterId": CLUSTER_ID,
            "requestedResources": {
                "gpu": {"request": pod_gpu, "limit": pod_gpu},
                "cpu": {"request": pod_cpu, "limit": pod_cpu},
                "cpuMemory": {
                    "request": f"{pod_cpu_memory_gib:g}Gi",
                    "limit": f"{pod_cpu_memory_gib:g}Gi",
                },
            },
            "allocatedResources": {
                "gpu": pod_gpu,
                "cpu": pod_cpu,
                "cpuMemory": f"{pod_cpu_memory_gib:g}Gi",
            },
        })
    page = pods[offset : offset + limit]
    has_more = offset + limit < len(pods)
    return {
        "workloadId": workload_id,
        "type": status["type"],
        "pods": page,
        "hasMore": has_more,
        **({"nextOffset": offset + limit} if has_more else {}),
    }


def get_workload_metrics(args):
    """Return deterministic workload time-series measurements."""
    workload_id = args["workloadId"]
    status = get_workload_status(args)
    if workload_id not in METRICS:
        return {"workloadId": workload_id, "type": status["type"], "measurements": []}
    source = METRICS[workload_id]
    measurements = []
    if "gpu_utilization" in source:
        measurements.append({"type": "GPU_UTILIZATION", "values": [{"value": str(source["gpu_utilization"]["avg"]), "timestamp": "2026-05-31T10:52:00Z"}]})
    if "gpu_memory" in source:
        measurements.append({"type": "GPU_MEMORY_USAGE_BYTES", "values": [{"value": str(source["gpu_memory"]["usedGiB"] * 1073741824), "timestamp": "2026-05-31T10:52:00Z"}]})
    if "latency_ms" in source:
        measurements.append({"type": "LATENCY", "labels": {"statistic": "p95"}, "values": [{"value": str(source["latency_ms"]["p95"] / 1000), "timestamp": "2026-05-31T10:52:00Z"}]})
    return {"workloadId": workload_id, "type": status["type"], "measurements": measurements}


def get_workload_effective_policy(args):
    """Return an effective policy or a deterministic fixture error."""
    if args["projectId"] == "4500099":
        raise ValueError("effective policy unavailable for fixture")
    return {
        "projectId": args["projectId"],
        "kind": args["kind"].lower(),
        "effective": {"defaults": {"idleTimeout": "4h"}, "rules": {}},
    }


def get_cluster_infrastructure_health(args):
    """Return node-pool cardinality and independent node-health context."""
    node_pool = args.get("nodePoolName")
    if node_pool and node_pool != "node-pool-a":
        return {
            "clusterId": CLUSTER_ID,
            "nodePool": node_pool,
            "asOf": "2026-05-31T10:52:00Z",
            "totals": {"nodesTotal": 1, "nodesUnhealthy": 0, "gpusUnschedulable": 0},
            "byNodePool": [],
            "unhealthyNodes": [],
        }
    return {
        "clusterId": CLUSTER_ID,
        "nodePool": node_pool,
        "asOf": "2026-05-31T10:52:00Z",
        "totals": {"nodesTotal": 2, "nodesUnhealthy": 1, "gpusUnschedulable": 8},
        "byNodePool": [
            {"nodePool": "node-pool-a", "nodesUnhealthy": 1, "gpusUnschedulable": 8}
        ],
        "unhealthyNodes": [
            {
                "name": "node-a-04",
                "nodePool": "node-pool-a",
                "status": "NotReady",
                "gpus": {"model": "NVIDIA H100 80GB HBM3", "count": 8},
                "conditions": [
                    {
                        "type": "Ready",
                        "reason": "KubeletNotReady",
                        "message": "Kubelet stopped posting node status.",
                    }
                ],
            }
        ],
    }


def schema(properties, required=None):
    """Build a strict JSON Schema object for one fixture tool."""
    return {
        "type": "object",
        "properties": properties,
        "required": required or [],
        "additionalProperties": False,
    }


TOOLS = {
    "list_workloads": {
        "description": "List workloads with compact identity, state, project, phase time, and GPU allocation fields. Set includeDetails for timestamps, node pools, priority, preemptibility, and idle GPUs. Read-only.",
        "inputSchema": schema(
            {
                "clusterId": {"type": "string"},
                "phase": {"type": "string"},
                "type": {"type": "string"},
                "name": {"type": "string"},
                "projectId": {"type": "string"},
                "projectName": {"type": "object"},
                "departmentName": {"type": "object"},
                "phaseUpdatedAfter": {"type": "string"},
                "filterBy": {
                    "type": "array",
                    "items": {
                        "type": "object",
                        "properties": {
                            "field": {"type": "string"},
                            "operator": {
                                "type": "string",
                                "enum": [">", ">=", "<", "<=", "==", "!="],
                            },
                            "value": {"type": ["string", "number", "boolean"]},
                        },
                        "required": ["field", "operator", "value"],
                        "additionalProperties": False,
                    },
                },
                "gpuIdleOnly": {"type": "boolean"},
                "includeDetails": {"type": "boolean"},
                "sortBy": {"type": "string", "enum": sorted(SORT_FIELDS)},
                "sortOrder": {"type": "string"},
                "limit": {"type": "integer"},
                "offset": {"type": "integer"},
            }
        ),
        "handler": list_workloads,
    },
    "get_workloads_summary": {
        "description": "Return count-only phase buckets for a scoped workload set. Read-only.",
        "inputSchema": schema(
            {
                "clusterId": {"type": "string"},
                "orgType": {"type": "string"},
                "orgName": {"type": "string"},
                "nodePoolName": {"type": "string"},
                "gpuIdleOnly": {"type": "boolean"},
            }
        ),
        "handler": get_workloads_summary,
    },
    "get_node_pods": {
        "description": "Return a paginated list of pods and workload IDs on one node; default 10 and maximum 200 per page. Read-only.",
        "inputSchema": schema(
            {
                "nodeName": {"type": "string"},
                "clusterUuid": {"type": "string"},
                "limit": {"type": "integer"},
                "offset": {"type": "integer"},
            },
            ["nodeName"],
        ),
        "handler": get_node_pods,
    },
    "list_projects": {
        "description": "List Run:ai projects and numeric project IDs. Read-only.",
        "inputSchema": schema({"clusterId": {"type": "string"}, "limit": {"type": "integer"}, "offset": {"type": "integer"}}),
        "handler": list_projects,
    },
    "get_workload_status": {
        "description": "Return synthetic workload status. Read-only.",
        "inputSchema": schema({"workloadId": {"type": "string"}}, ["workloadId"]),
        "handler": get_workload_status,
    },
    "get_workload_spec": {
        "description": "Return synthetic workload submitted spec. Read-only.",
        "inputSchema": schema({"workloadId": {"type": "string"}}, ["workloadId"]),
        "handler": get_workload_spec,
    },
    "get_workload_pods": {
        "description": "Return paginated workload pods without readiness or restart fields. Read-only.",
        "inputSchema": schema(
            {
                "workloadId": {"type": "string"},
                "limit": {"type": "integer"},
                "offset": {"type": "integer"},
                "includeMetrics": {"type": "boolean"},
            },
            ["workloadId"],
        ),
        "handler": get_workload_pods,
    },
    "get_workload_metrics": {
        "description": "Return synthetic workload metrics. Read-only.",
        "inputSchema": schema(
            {"workloadId": {"type": "string"}, "start": {"type": "string"}, "end": {"type": "string"}, "numberOfSamples": {"type": "integer"}},
            ["workloadId"],
        ),
        "handler": get_workload_metrics,
    },
    "get_workload_effective_policy": {
        "description": "Return synthetic effective policy. Read-only.",
        "inputSchema": schema({"projectId": {"type": "string"}, "kind": {"type": "string"}}, ["projectId", "kind"]),
        "handler": get_workload_effective_policy,
    },
    "get_cluster_infrastructure_health": {
        "description": "Return node totals and unhealthy node evidence for a cluster or node pool. Read-only.",
        "inputSchema": schema(
            {
                "clusterId": {"type": "string"},
                "nodePoolName": {"type": "string"},
            },
            ["clusterId"],
        ),
        "handler": get_cluster_infrastructure_health,
    },
}


def send(payload):
    """Write one compact JSON-RPC message to stdout."""
    sys.stdout.write(json.dumps(payload, sort_keys=True, separators=(",", ":")) + "\n")
    sys.stdout.flush()


def send_error(request_id, code, message):
    """Write one JSON-RPC protocol error."""
    send({"jsonrpc": "2.0", "id": request_id, "error": {"code": code, "message": message}})


def tool_schema():
    """Return MCP tool descriptors for the deterministic fixture."""
    return [
        {
            "name": name,
            "description": definition["description"],
            "inputSchema": definition["inputSchema"],
            "annotations": {"readOnlyHint": True, "destructiveHint": False},
        }
        for name, definition in TOOLS.items()
    ]


def tool_result(request_id, result, *, is_error=False):
    """Wrap one fixture value as an MCP tool result."""
    payload = {"content": [{"type": "text", "text": json.dumps(result, sort_keys=True)}]}
    if is_error:
        payload["isError"] = True
    send({"jsonrpc": "2.0", "id": request_id, "result": payload})


def main():
    """Serve newline-delimited MCP JSON-RPC requests until stdin closes."""
    for raw_line in sys.stdin:
        line = raw_line.strip()
        if not line:
            continue
        try:
            request = json.loads(line)
        except json.JSONDecodeError:
            send_error(None, -32700, "parse error")
            continue

        if not isinstance(request, dict):
            send_error(None, -32600, "invalid request")
            continue

        method = request.get("method")
        request_id = request.get("id")
        if method == "initialize":
            send(
                {
                    "jsonrpc": "2.0",
                    "id": request_id,
                    "result": {
                        "protocolVersion": "2024-11-05",
                        "capabilities": {"tools": {}},
                        "serverInfo": {
                            "name": "runai-workload-health-summary-eval",
                            "version": "1.0.0",
                        },
                    },
                }
            )
        elif method == "notifications/initialized":
            continue
        elif method == "tools/list":
            send({"jsonrpc": "2.0", "id": request_id, "result": {"tools": tool_schema()}})
        elif method == "tools/call":
            params = request.get("params", {})
            if not isinstance(params, dict):
                send_error(request_id, -32602, "params must be an object")
                continue
            definition = TOOLS.get(params.get("name"))
            if definition is None:
                send_error(request_id, -32601, f"unknown tool {params.get('name')}")
                continue
            try:
                tool_result(request_id, definition["handler"](params.get("arguments") or {}))
            except Exception as error:  # noqa: BLE001 - JSON-RPC boundary must survive handler failures.
                print(
                    f"fixture tool {params.get('name')} failed: {type(error).__name__}: {error}",
                    file=sys.stderr,
                )
                tool_result(
                    request_id,
                    {"error": f"{type(error).__name__}: {error}"},
                    is_error=True,
                )
        elif request_id is not None:
            send_error(request_id, -32601, f"unknown method {method}")


if __name__ == "__main__":
    main()
