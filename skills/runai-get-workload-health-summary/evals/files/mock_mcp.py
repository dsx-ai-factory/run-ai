#!/usr/bin/env python3
# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0
"""Deterministic, network-free Run:ai MCP fixture for workload health evals."""

import json
import sys


CLUSTER_ID = "20000000-0000-4000-8000-000000000001"
PROJECT_ID = "4500002"
PROJECT_NAME = "ml-team"
DEPARTMENT_NAME = "Research"
HEALTHY_ID = "00000000-0000-4000-8000-000000000042"
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
        "requestedNodePools": ["node-pool-a"],
        "currentNodePools": ["node-pool-a"] if node else [],
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


def base_spec(workload_id, name):
    return {
        "id": workload_id,
        "name": name,
        "clusterId": CLUSTER_ID,
        "compute": {"gpu": 4, "gpuType": "h100", "cpu": 16, "memory": "256Gi"},
        "scheduling": {
            "nodePools": ["node-pool-a"],
            "priority": "medium",
            "preemptible": True,
        },
        "runtime": {
            "image": "nvcr.io/nvidia/pytorch:24.05-py3",
            "allocated_resources": {"gpu": 4, "cpu": 16, "memory": "256Gi"},
            "actual_placement": ["node-a-04", "node-a-05"],
        },
        "environment": {"variables": []},
        "dataSources": [{"name": "training-data", "mountPath": "/data"}],
    }


def pod(name, node, ready=True, restarts=0):
    return {
        "name": name,
        "nodeName": node,
        "ready": ready,
        "restarts": restarts,
        "age": "3h12m",
        "k8sPhase": "Running" if ready else "Pending",
    }


PODS = {
    FAILED_IDS[0]: {
        "desired": 1,
        "ready": 0,
        "pods": [pod("train-llama-9-0", "node-a-04", ready=False)],
    },
    FAILED_IDS[2]: {
        "desired": 1,
        "ready": 0,
        "pods": [pod("eval-batch-3-0", "node-a-04", ready=False)],
    },
    FAILED_IDS[3]: {
        "desired": 1,
        "ready": 0,
        "pods": [pod("trainer-cv-2-0", "node-b-12", ready=False)],
    },
    HEALTHY_ID: {
        "desired": 4,
        "ready": 4,
        "pods": [
            pod("training-job-42-0", "node-a-04"),
            pod("training-job-42-1", "node-a-04"),
            pod("training-job-42-2", "node-a-05"),
            pod("training-job-42-3", "node-a-05"),
        ],
    },
    DEGRADED_ID: {
        "desired": 1,
        "ready": 1,
        "pods": [pod("idle-eval-7-0", "node-c-02", restarts=1)],
    },
    INFERENCE_ID: {
        "desired": 2,
        "ready": 0,
        "pods": [
            pod("inference-prod-9-0", "node-a-04", ready=False, restarts=2),
            pod("inference-prod-9-1", "node-a-04", ready=False, restarts=2),
        ],
    },
    PENDING_ID: {
        "desired": 1,
        "ready": 0,
        "pods": [pod("pending-startup-0", "", ready=False)],
    },
    NO_POLICY_ID: {
        "desired": 1,
        "ready": 1,
        "pods": [pod("no-policy-snapshot-0", "node-a-06")],
    },
    NO_METRICS_ID: {
        "desired": 1,
        "ready": 1,
        "pods": [pod("no-metrics-snapshot-0", "node-a-07")],
    },
    COMPLETED_ID: {
        "desired": 1,
        "ready": 0,
        "pods": [pod("completed-success-0", "node-a-08")],
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
    if args.get("cluster_id") and args["cluster_id"] != CLUSTER_ID:
        return False
    if args.get("clusterId") and args["clusterId"] != CLUSTER_ID:
        return False
    if args.get("phase") and row_data["phase"].lower() != args["phase"].lower():
        return False
    if args.get("projectId") and row_data["projectId"] != args["projectId"]:
        return False
    project_filter = args.get("projectName")
    if project_filter and row_data["projectName"] != project_filter.get("value"):
        return False
    department_filter = args.get("departmentName")
    if department_filter and row_data["_departmentName"] != department_filter.get("value"):
        return False
    if args.get("name") and row_data["name"] != args["name"]:
        return False
    if args.get("type") and row_data["type"].lower() != args["type"].lower():
        return False
    if args.get("phaseUpdatedAfter") and row_data["phaseUpdatedAt"] < args["phaseUpdatedAfter"]:
        return False
    if args.get("gpuIdleOnly") and not row_data["idleAllocatedGpus"]:
        return False
    return True


def public_row(item):
    return {key: value for key, value in item.items() if not key.startswith("_")}


def limited(rows, args):
    limit = int(args.get("limit", 50))
    offset = int(args.get("offset", 0))
    page = rows[offset : offset + limit]
    has_more = offset + limit < len(rows)
    result = {"workloads": [public_row(item) for item in page], "hasMore": has_more}
    if has_more:
        result["nextOffset"] = offset + limit
    return result


def list_workloads(args):
    return limited([item for item in WORKLOAD_ROWS if matches(item, args)], args)


def get_workloads_summary(args):
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
    node_name = args["nodeName"]
    rows = [item for item in WORKLOAD_ROWS if item["_nodeName"] == node_name]
    return {
        "name": node_name,
        "nodePool": "node-pool-a",
        "status": "Ready",
        "gpuInfo": {"gpuType": "NVIDIA H100 80GB HBM3", "gpuCount": 8},
        "pods": [
            {
                "name": f"{item['name']}-0",
                "id": f"pod-{item['id']}",
                "nodeName": node_name,
                "currentNodePool": "node-pool-a",
                "k8sPhase": "Failed" if item["phase"] == "Failed" else "Running",
                "createdAt": "2026-05-31T10:30:00Z",
                "clusterId": CLUSTER_ID,
                "workloadId": item["id"],
                "workloadName": item["name"],
                "workloadType": item["type"],
            }
            for item in rows[:10]
        ],
    }


def list_projects(args):
    projects = []
    seen = set()
    for item in WORKLOAD_ROWS:
        if item["projectId"] in seen:
            continue
        seen.add(item["projectId"])
        projects.append({"id": item["projectId"], "name": item["projectName"], "clusterId": CLUSTER_ID, "departmentId": "5500001", "phase": "Ready"})
    return {"projects": projects, "totalCount": len(projects), "hasMore": False}


def get_workload_status(args):
    workload_id = args["workloadId"]
    if workload_id in STATUS:
        raw = STATUS[workload_id]
        return {
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
    for item in WORKLOAD_ROWS:
        if item["id"] == workload_id:
            return {
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
    raise ValueError(f"unknown workloadId {workload_id}")


def get_workload_spec(args):
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
        "topology": {"actualPlacement": {"level": "node", "value": "node-a-04"}},
        "pendingSchedulingMessages": [],
    }


def get_workload_pods(args):
    workload_id = args["workloadId"]
    status = get_workload_status(args)
    if workload_id not in PODS:
        return {"workloadId": workload_id, "type": status["type"], "pods": []}
    pods = []
    for index, item in enumerate(PODS[workload_id]["pods"]):
        pods.append({
            "name": item["name"],
            "id": f"pod-{workload_id}-{index}",
            "nodeName": item.get("nodeName", ""),
            "currentNodePool": "node-pool-a" if item.get("nodeName") else "",
            "k8sPhase": item["k8sPhase"],
            "createdAt": "2026-05-31T08:16:00Z",
            "clusterId": CLUSTER_ID,
            "requestedResources": {"gpu": {"request": 4, "limit": 4}, "cpu": {"request": 16, "limit": 16}, "cpuMemory": {"request": "256Gi", "limit": "256Gi"}},
            "allocatedResources": {"gpu": 4, "cpu": 16, "cpuMemory": "256Gi"},
        })
    return {"workloadId": workload_id, "type": status["type"], "pods": pods}


def get_workload_metrics(args):
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
    if args["projectId"] == "4500099":
        raise ValueError("effective policy unavailable for fixture")
    return {
        "projectId": args["projectId"],
        "kind": args["kind"].lower(),
        "effective": {"defaults": {"idleTimeout": "4h"}, "rules": {}},
    }


def schema(properties, required=None):
    return {
        "type": "object",
        "properties": properties,
        "required": required or [],
        "additionalProperties": False,
    }


TOOLS = {
    "list_workloads": {
        "description": "List workloads with compact identity, state, project, time, node-pool, priority, and GPU allocation fields. Read-only.",
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
                "gpuIdleOnly": {"type": "boolean"},
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
        "description": "Return up to 10 pods and their workload IDs on one node. Read-only.",
        "inputSchema": schema(
            {
                "nodeName": {"type": "string"},
                "clusterUuid": {"type": "string"},
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
        "description": "Return synthetic workload pods. Read-only.",
        "inputSchema": schema({"workloadId": {"type": "string"}}, ["workloadId"]),
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
}


def send(payload):
    sys.stdout.write(json.dumps(payload, sort_keys=True, separators=(",", ":")) + "\n")
    sys.stdout.flush()


def send_error(request_id, code, message):
    send({"jsonrpc": "2.0", "id": request_id, "error": {"code": code, "message": message}})


def tool_schema():
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
    payload = {"content": [{"type": "text", "text": json.dumps(result, sort_keys=True)}]}
    if is_error:
        payload["isError"] = True
    send({"jsonrpc": "2.0", "id": request_id, "result": payload})


def main():
    for raw_line in sys.stdin:
        line = raw_line.strip()
        if not line:
            continue
        try:
            request = json.loads(line)
        except json.JSONDecodeError:
            send_error(None, -32700, "parse error")
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
            definition = TOOLS.get(params.get("name"))
            if definition is None:
                send_error(request_id, -32601, f"unknown tool {params.get('name')}")
                continue
            try:
                tool_result(request_id, definition["handler"](params.get("arguments") or {}))
            except ValueError as error:
                tool_result(request_id, {"error": str(error)}, is_error=True)
        elif request_id is not None:
            send_error(request_id, -32601, f"unknown method {method}")


if __name__ == "__main__":
    main()
