#!/usr/bin/env python3
# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0
"""Offline checks for the workload-health publication package and MCP protocol."""

import importlib.util
import json
from pathlib import Path
import re
import subprocess
import sys
import tomllib
import unittest


REPO = Path(__file__).resolve().parents[3]
SKILL = REPO / "skills/runai-get-workload-health-summary"
SERVER = SKILL / "evals/files/mock_mcp.py"
sys.dont_write_bytecode = True
spec = importlib.util.spec_from_file_location("workload_health_mock", SERVER)
mock = importlib.util.module_from_spec(spec)
spec.loader.exec_module(mock)


def exchange(requests):
    result = subprocess.run(
        [sys.executable, str(SERVER)],
        input="".join(json.dumps(request) + "\n" for request in requests),
        text=True,
        capture_output=True,
        timeout=10,
        check=True,
    )
    return [json.loads(line) for line in result.stdout.splitlines()]


def call(name, arguments, request_id=1):
    return {
        "jsonrpc": "2.0",
        "id": request_id,
        "method": "tools/call",
        "params": {"name": name, "arguments": arguments},
    }


def payload(response):
    return json.loads(response["result"]["content"][0]["text"])


class PackageTests(unittest.TestCase):
    def test_package_contains_no_generated_bytecode(self):
        self.assertEqual(list(SKILL.rglob("*.pyc")), [])
        self.assertEqual(list(SKILL.rglob("*.pyo")), [])
        self.assertEqual(list(SKILL.rglob("__pycache__")), [])

    def test_eval_cases_are_unique_and_use_nvcarps_schema(self):
        cases = json.loads((SKILL / "evals/evals.json").read_text())["evals"]
        extended = json.loads((Path(__file__).parent / "extended-evals.json").read_text())["evals"]
        by_id = {case["id"]: case for case in extended}
        self.assertEqual(len(cases), 30)
        self.assertEqual(len(extended), 39)
        self.assertEqual(len(by_id), len(extended))
        self.assertEqual(len({case["id"] for case in cases}), len(cases))
        for case in cases:
            with self.subTest(case=case["id"]):
                self.assertEqual(case, by_id[case["id"]])
                self.assertRegex(case["id"], r"^[a-zA-Z0-9][a-zA-Z0-9._-]{0,127}$")
                self.assertTrue(case["prompt"])
                self.assertTrue(case["expected_output"])
                self.assertTrue(case["assertions"])
                self.assertNotIn("expected_behavior", case)
                self.assertIn(case["expected_skill"], (None, "runai-get-workload-health-summary"))

    def test_all_prompt_workload_ids_resolve(self):
        cases = json.loads((Path(__file__).parent / "extended-evals.json").read_text())["evals"]
        known = {row["id"] for row in mock.WORKLOAD_ROWS} | mock.RECORDED.keys()
        for case in cases:
            for workload_id in re.findall(r"00000000-0000-4000-8000-\d{12}", case["prompt"]):
                with self.subTest(case=case["id"], workload=workload_id):
                    self.assertIn(workload_id, known)

    def test_mcp_command_is_executable_and_scoped_to_shared_input(self):
        config = tomllib.loads((SKILL / "evals/environment/mcp_servers.toml").read_text())
        servers = config["mcp_servers"]
        self.assertEqual(len(servers), 1)
        self.assertEqual(servers[0]["transport"], "stdio")
        self.assertEqual(servers[0]["command"], "/workspace/input/mock_mcp.py")
        self.assertTrue(SERVER.stat().st_mode & 0o111)


class ReplayTests(unittest.TestCase):
    def test_ping_heartbeat_and_notification_preserve_the_session(self):
        responses = exchange([
            {"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {}},
            {"jsonrpc": "2.0", "method": "notifications/initialized"},
            {"jsonrpc": "2.0", "id": 2, "method": "ping"},
            {"jsonrpc": "2.0", "method": "ping"},
            call("get_workload_status", {"workloadId": mock.HEALTHY_ID}, 3),
        ])
        self.assertEqual([response["id"] for response in responses], [1, 2, 3])
        self.assertEqual(responses[1]["result"], {})
        self.assertEqual(payload(responses[2])["id"], mock.HEALTHY_ID)

    def test_oversized_pages_and_overflowing_offsets_return_tool_errors(self):
        requests = [
            call("list_projects", {"limit": 201}),
            call("get_node_pods", {"nodeName": "node-a-04", "limit": 201}),
            call("get_workload_pods", {"workloadId": mock.HEALTHY_ID, "limit": 501}),
            call("list_workloads", {"offset": 2**31}),
            call("get_node_pods", {"nodeName": "node-a-04", "limit": 200, "offset": 2**31 - 1}),
            call("get_node_pods", {"nodeName": mock.RECORDED_SCOPE["get_node_pods"]["name"], "offset": 2**31 - 1}),
        ]
        for response in exchange(requests):
            self.assertTrue(response["result"]["isError"])

    def test_numeric_filter_uses_real_api_string_value(self):
        result = payload(exchange([call("list_workloads", {
            "clusterId": mock.CLUSTER_ID,
            "filterBy": [{"field": "priority", "operator": ">=", "value": "150"}],
        })])[0])
        self.assertEqual([row["id"] for row in result["workloads"]], [mock.INFERENCE_ID])

    def test_phase_window_compares_instants_with_timezone_offsets(self):
        result = payload(exchange([call("list_workloads", {
            "clusterId": mock.CLUSTER_ID, "phase": "Failed",
            "phaseUpdatedAfter": "2026-05-31T12:47:00+02:00",
            "filterBy": [{"field": "phaseUpdatedAt", "operator": "<=", "value": "2026-05-31T12:48:00+02:00"}],
        })])[0])
        self.assertEqual({row["id"] for row in result["workloads"]}, set(mock.FAILED_IDS[:3]))

    def test_metrics_do_not_return_samples_outside_requested_window(self):
        requests = [
            call("get_workload_metrics", {"workloadId": mock.DEGRADED_ID, "start": "2026-05-31T12:30:00+02:00", "end": "2026-05-31T13:30:00+02:00"}, 1),
            call("get_workload_metrics", {"workloadId": mock.DEGRADED_ID, "start": "2026-06-01T10:30:00Z", "end": "2026-06-01T11:30:00Z"}, 2),
        ]
        inside, outside = [payload(response) for response in exchange(requests)]
        self.assertTrue(all(series["values"] for series in inside["measurements"]))
        self.assertTrue(all(series["values"] == [] for series in outside["measurements"]))

    def test_unknown_nodes_projects_and_pools_do_not_fabricate_health_or_policy(self):
        requests = [
            call("get_node_pods", {"nodeName": "missing-node", "clusterUuid": mock.CLUSTER_ID}),
            call("get_workload_effective_policy", {"projectId": "9999999", "kind": "training"}),
            call("get_workload_effective_policy", {"projectId": mock.PROJECT_ID, "kind": "invented"}),
            call("get_cluster_infrastructure_health", {"clusterId": mock.CLUSTER_ID, "nodePoolName": "missing-pool"}),
            call("get_workloads_summary", {"orgType": "project"}),
            call("get_workloads_summary", {"clusterId": mock.CLUSTER_ID, "orgType": "project", "orgName": "missing-project"}),
        ]
        for response in exchange(requests):
            self.assertTrue(response["result"]["isError"])

    def test_synthetic_spec_uses_current_mcp_field_paths_and_priority(self):
        result = payload(exchange([call("get_workload_spec", {"workloadId": mock.INFERENCE_ID})])[0])
        self.assertEqual(result["spec"]["image"], "nvcr.io/nvidia/pytorch:24.05-py3")
        self.assertEqual(result["spec"]["compute"]["gpuDevicesRequest"], 4)
        self.assertEqual(result["spec"]["priorityClass"], "high")
        self.assertEqual(result["spec"]["preemptibility"], "non-preemptible")
        self.assertNotIn("runtime", result["spec"])
        self.assertNotIn("scheduling", result["spec"])
        self.assertEqual(result["allocatedResources"], {})

    def test_inference_scalar_latency_does_not_invent_a_percentile_label(self):
        result = payload(exchange([call("get_workload_metrics", {"workloadId": mock.INFERENCE_ID})])[0])
        latency = next(series for series in result["measurements"] if series["type"] == "LATENCY")
        self.assertNotIn("labels", latency)
        self.assertEqual(latency["values"][0]["value"], "0.42")

    def test_budget_probe_detects_overflow_without_affecting_other_clusters(self):
        probe = payload(exchange([call("list_workloads", {
            "clusterId": mock.BUDGET_CLUSTER,
            "projectName": {"operator": "==", "value": "budget-team"},
            "phase": "Failed", "includeDetails": True, "limit": 51,
        })])[0])
        self.assertEqual(len(probe["workloads"]), 51)
        self.assertTrue(all(row["clusterId"] == mock.BUDGET_CLUSTER for row in probe["workloads"]))
        wrong = payload(exchange([call("list_workloads", {
            "clusterId": mock.CLUSTER_ID,
            "projectName": {"operator": "==", "value": "budget-team"},
        })])[0])
        self.assertEqual(wrong["workloads"], [])

    def test_malformed_tool_name_returns_protocol_error_and_recovers(self):
        responses = exchange([call([], {}, 1), call("get_workload_status", {"workloadId": mock.HEALTHY_ID}, 2)])
        self.assertEqual(responses[0]["error"]["code"], -32602)
        self.assertEqual(payload(responses[1])["id"], mock.HEALTHY_ID)

    def test_every_recorded_call_replays_exact_response_over_stdio(self):
        self.assertEqual(len(mock.RECORDED), 3)
        for recording in mock.LIVE_RECORDINGS["recordings"]:
            for recorded_call in recording["calls"]:
                with self.subTest(scenario=recording["scenario"], tool=recorded_call["tool"]):
                    actual = payload(exchange([call(recorded_call["tool"], recorded_call["arguments"])])[0])
                    self.assertEqual(actual, recorded_call["response"])

    def test_recorded_workloads_resolve_only_in_recorded_cluster_and_project(self):
        for workload_id, recording in mock.RECORDED.items():
            status = recording["get_workload_status"]
            arguments = {
                "clusterId": status["clusterId"], "name": status["name"],
                "projectName": {"operator": "==", "value": status["projectName"]},
            }
            result = payload(exchange([call("list_workloads", arguments)])[0])
            self.assertEqual([row["id"] for row in result["workloads"]], [workload_id])
            arguments["clusterId"] = mock.CLUSTER_ID
            self.assertEqual(payload(exchange([call("list_workloads", arguments)])[0])["workloads"], [])
            arguments["clusterId"] = status["clusterId"]
            arguments["projectName"]["value"] = mock.PROJECT_NAME
            self.assertEqual(payload(exchange([call("list_workloads", arguments)])[0])["workloads"], [])

    def test_recorded_overview_and_node_pods_paginate_without_changing_evidence(self):
        for tool, collection, arguments in [
            ("list_workloads", "workloads", {"clusterId": mock.RECORDED_CLUSTER, "includeDetails": True}),
            ("get_node_pods", "pods", {"nodeName": mock.RECORDED_SCOPE["get_node_pods"]["name"], "clusterUuid": mock.RECORDED_CLUSTER}),
        ]:
            with self.subTest(tool=tool):
                arguments.update(limit=1, offset=0)
                items = []
                while True:
                    result = payload(exchange([call(tool, arguments)])[0])
                    items.extend(result[collection])
                    if not result["hasMore"]:
                        break
                    self.assertGreater(result["nextOffset"], arguments["offset"])
                    arguments["offset"] = result["nextOffset"]
                self.assertEqual(items, mock.RECORDED_SCOPE[tool][collection])

    def test_recorded_null_gpu_and_empty_metric_series_are_preserved(self):
        self.assertTrue(all(row["allocatedGpu"] is None for row in mock.RECORDED_ROWS))
        for recording in mock.RECORDED.values():
            metrics = recording["get_workload_metrics"]["measurements"]
            self.assertTrue(metrics)
            self.assertTrue(all(series["values"] == [] for series in metrics))
            self.assertNotIn("idleTimeout", json.dumps(mock.RECORDED_SCOPE["get_workload_effective_policy"]))

    def test_node_and_infrastructure_reject_wrong_or_unrecorded_clusters(self):
        for tool, arguments in [
            ("get_node_pods", {"nodeName": mock.RECORDED_SCOPE["get_node_pods"]["name"], "clusterUuid": mock.CLUSTER_ID}),
            ("get_cluster_infrastructure_health", {"clusterId": mock.RECORDED_CLUSTER}),
        ]:
            with self.subTest(tool=tool):
                self.assertTrue(exchange([call(tool, arguments)])[0]["result"]["isError"])

    def test_untrusted_instructions_are_returned_only_as_event_data(self):
        response = payload(exchange([call("get_workload_status", {"workloadId": mock.UNTRUSTED_ID})])[0])
        self.assertEqual(response["phase"], "Running")
        self.assertIn("kubectl delete", response["events"][0]["message"])
        self.assertIn("Secret", response["events"][0]["message"])

    def test_initialize_and_read_only_discovery_over_stdio(self):
        responses = exchange([
            {"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {}},
            {"jsonrpc": "2.0", "method": "notifications/initialized"},
            {"jsonrpc": "2.0", "id": 2, "method": "tools/list"},
        ])
        self.assertEqual([response["id"] for response in responses], [1, 2])
        self.assertIn("protocolVersion", responses[0]["result"])
        tools = responses[1]["result"]["tools"]
        self.assertEqual({tool["name"] for tool in tools}, set(mock.TOOLS))
        for tool in tools:
            self.assertTrue(tool["annotations"]["readOnlyHint"])
            self.assertFalse(tool["annotations"]["destructiveHint"])

    def test_unknown_and_mutating_tools_are_rejected(self):
        for name in ("delete_workload", "restart_workload", "get_secrets"):
            with self.subTest(tool=name):
                response = exchange([call(name, {})])[0]
                self.assertEqual(response["error"]["code"], -32601)

    def test_invalid_arguments_return_tool_errors(self):
        requests = [
            call("get_workload_status", {}),
            call("get_workload_status", {"workloadId": "missing"}),
            call("list_workloads", {"limit": 201}),
            call("get_node_pods", {"nodeName": "node-a-04", "offset": -1}),
            call("get_workload_status", {"workloadId": mock.HEALTHY_ID, "clusterId": mock.CLUSTER_ID}),
            call("get_workload_status", {"workloadId": []}),
            call("list_workloads", {"limit": True}),
            call("list_workloads", {"projectName": {"value": "ml-team"}}),
            call("list_workloads", {"filterBy": [{"field": "priority", "operator": ">=", "value": 150}]}),
            call("list_workloads", {"filterBy": [{"field": "invented", "operator": "==", "value": "x"}]}),
            call("get_workload_metrics", {"workloadId": mock.HEALTHY_ID, "start": "tomorrow"}),
            call("get_workload_metrics", {"workloadId": mock.HEALTHY_ID, "start": "2026-05-31T11:30:00Z", "end": "2026-05-31T10:30:00Z"}),
            call("get_workload_metrics", {"workloadId": mock.HEALTHY_ID, "numberOfSamples": 1001}),
        ]
        for response in exchange(requests):
            self.assertTrue(response["result"]["isError"])

    def test_list_filters_sorts_and_paginates_without_duplicates(self):
        arguments = {
            "clusterId": mock.CLUSTER_ID,
            "phase": "Failed",
            "includeDetails": True,
            "sortBy": "phaseUpdatedAt",
            "sortOrder": "desc",
            "limit": 2,
            "offset": 0,
        }
        rows = []
        while True:
            page = payload(exchange([call("list_workloads", arguments)])[0])
            rows.extend(page["workloads"])
            if not page["hasMore"]:
                break
            arguments["offset"] = page["nextOffset"]
        self.assertEqual({row["id"] for row in rows}, set(mock.FAILED_IDS))
        self.assertEqual(len(rows), len({row["id"] for row in rows}))
        self.assertEqual(
            [row["phaseUpdatedAt"] for row in rows],
            sorted((row["phaseUpdatedAt"] for row in rows), reverse=True),
        )

    def test_node_pods_paginate_and_preserve_workload_identity(self):
        arguments = {"nodeName": "node-a-04", "clusterUuid": mock.CLUSTER_ID, "limit": 1, "offset": 0}
        pods = []
        while True:
            page = payload(exchange([call("get_node_pods", arguments)])[0])
            pods.extend(page["pods"])
            if not page["hasMore"]:
                break
            arguments["offset"] = page["nextOffset"]
        self.assertTrue(pods)
        self.assertEqual(len(pods), len({pod["id"] for pod in pods}))
        self.assertTrue(all(pod["workloadId"] for pod in pods))

    def test_name_project_cluster_resolution_is_exact(self):
        arguments = {
            "name": "training-job-42",
            "clusterId": mock.CLUSTER_ID,
            "projectName": {"operator": "==", "value": mock.PROJECT_NAME},
            "sortBy": "phaseUpdatedAt",
            "sortOrder": "desc",
        }
        result = payload(exchange([call("list_workloads", arguments)])[0])
        self.assertEqual(result["workloads"][0]["id"], mock.HEALTHY_ID)
        self.assertEqual(len(result["workloads"]), 2)
        arguments["clusterId"] = "20000000-0000-4000-8000-000000000099"
        result = payload(exchange([call("list_workloads", arguments)])[0])
        self.assertEqual(result["workloads"], [])

    def test_missing_policy_and_metrics_are_explicit(self):
        policy = exchange([call("get_workload_effective_policy", {
            "projectId": "4500099", "kind": "training",
        })])[0]
        self.assertTrue(policy["result"]["isError"])
        metrics = payload(exchange([call("get_workload_metrics", {
            "workloadId": mock.NO_METRICS_ID,
        })])[0])
        self.assertEqual(metrics["measurements"], [])

    def test_node_correlation_has_independent_infrastructure_evidence(self):
        result = payload(exchange([call("get_cluster_infrastructure_health", {
            "clusterId": mock.CLUSTER_ID, "nodePoolName": "node-pool-a",
        })])[0])
        self.assertEqual(result["unhealthyNodes"][0]["name"], "node-a-04")
        self.assertEqual(result["unhealthyNodes"][0]["status"], "NotReady")

    def test_protocol_reports_malformed_json_and_recovers(self):
        result = subprocess.run(
            [sys.executable, str(SERVER)],
            input='invalid json\n{"jsonrpc":"2.0","id":2,"method":"tools/list"}\n',
            text=True,
            capture_output=True,
            timeout=10,
            check=True,
        )
        responses = [json.loads(line) for line in result.stdout.splitlines()]
        self.assertEqual(responses[0]["error"]["code"], -32700)
        self.assertEqual(responses[1]["id"], 2)


if __name__ == "__main__":
    unittest.main()
