#!/usr/bin/env python3
# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0
"""Offline checks for the published package and its actual MCP wire protocol."""

import importlib.util
import json
from pathlib import Path
import re
import subprocess
import sys
import tomllib
import unittest


NAME = "runai-get-utilization-overview"
REPO = Path(__file__).resolve().parents[3]
SKILL = REPO / "skills" / NAME
SERVER = SKILL / "evals/files/mock_mcp.py"
# Import fixtures without creating bytecode inside the publication package.
sys.dont_write_bytecode = True
spec = importlib.util.spec_from_file_location("utilization_mock", SERVER)
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


def mean(points):
    return sum(point["v"] for point in points) / len(points)


class PackageTests(unittest.TestCase):
    def test_package_contains_no_generated_bytecode(self):
        self.assertEqual(list(SKILL.rglob("*.pyc")), [])
        self.assertEqual(list(SKILL.rglob("*.pyo")), [])
        self.assertEqual(list(SKILL.rglob("__pycache__")), [])

    def test_smoke_is_subset_of_extended_with_unique_valid_cases(self):
        smoke = json.loads((SKILL / "evals/evals.json").read_text())["evals"]
        extended = json.loads(
            (Path(__file__).parent / "extended-evals.json").read_text()
        )["evals"]
        by_id = {case["id"]: case for case in extended}
        self.assertEqual(len(by_id), len(extended))
        self.assertEqual(len({case["id"] for case in smoke}), len(smoke))
        for case in smoke:
            with self.subTest(case=case["id"]):
                self.assertEqual(case, by_id[case["id"]])
        for case in extended:
            with self.subTest(case=case["id"]):
                self.assertRegex(case["id"], r"^[a-zA-Z0-9][a-zA-Z0-9._-]{0,127}$")
                self.assertTrue(case["prompt"])
                self.assertTrue(case["expected_output"])
                self.assertTrue(case["assertions"])
                self.assertNotIn("expected_behavior", case)
                # Only this skill is installed in the isolated workspace.
                self.assertIn(case["expected_skill"], (NAME, None))

    def test_prompts_reference_existing_clusters(self):
        for path in (SKILL / "evals/evals.json", Path(__file__).parent / "extended-evals.json"):
            for case in json.loads(path.read_text())["evals"]:
                for cluster_id in re.findall(
                    r"\b[0-9a-f]{8}-0000-4000-8000-[0-9a-f]{12}\b", case["prompt"]
                ):
                    with self.subTest(case=case["id"], cluster=cluster_id):
                        self.assertIn(cluster_id, mock.NOW)

    def test_mcp_command_is_executable_and_scoped_to_shared_input(self):
        config = tomllib.loads((SKILL / "evals/environment/mcp_servers.toml").read_text())
        servers = config["mcp_servers"]
        self.assertEqual(len(servers), 1)
        self.assertEqual(servers[0]["transport"], "stdio")
        self.assertEqual(servers[0]["command"], "/workspace/input/mock_mcp.py")
        self.assertTrue(SERVER.stat().st_mode & 0o111)


class ProtocolTests(unittest.TestCase):
    def test_initialize_and_read_only_discovery_over_stdio(self):
        responses = exchange([
            {"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {}},
            {"jsonrpc": "2.0", "method": "notifications/initialized"},
            {"jsonrpc": "2.0", "id": 2, "method": "tools/list"},
        ])
        self.assertEqual([r["id"] for r in responses], [1, 2])
        self.assertIn("protocolVersion", responses[0]["result"])
        tools = responses[1]["result"]["tools"]
        self.assertEqual({t["name"] for t in tools}, set(mock.TOOLS))
        for tool in tools:
            self.assertTrue(tool["annotations"]["readOnlyHint"])
            self.assertFalse(tool["annotations"]["destructiveHint"])

    def test_unknown_and_mutating_tools_are_rejected(self):
        for name in ("delete_workload", "suspend_workload", "update_project_quota"):
            with self.subTest(tool=name):
                response = exchange([call(name, {})])[0]
                self.assertEqual(response["error"]["code"], -32601)

    def test_invalid_identity_and_arguments_return_tool_errors(self):
        requests = [
            call("get_cluster_metrics", {}),
            call("get_cluster_metrics", {"clusterId": "20000000-0000-4000-8000-000000000099"}),
            call("get_org_unit_metrics", {"orgType": "project"}),
            call("get_org_unit_metrics", {"orgType": "team", "orgName": "vision-train"}),
            call("list_project_resources", {"clusterId": mock.LAB_CLUSTER, "unknown": 1}),
        ]
        for response in exchange(requests):
            self.assertTrue(response["result"]["isError"])

    def test_protocol_reports_malformed_json_and_recovers(self):
        result = subprocess.run(
            [sys.executable, str(SERVER)],
            input='invalid json\n{"jsonrpc":"2.0","id":2,"method":"tools/list"}\n',
            text=True, capture_output=True, timeout=10, check=True,
        )
        responses = [json.loads(line) for line in result.stdout.splitlines()]
        self.assertEqual(responses[0]["error"]["code"], -32700)
        self.assertEqual(responses[1]["id"], 2)


class FixtureTests(unittest.TestCase):
    def org_metrics(self, org_name, cluster_id, **extra):
        arguments = {
            "orgType": "project", "orgName": org_name, "clusterId": cluster_id,
            "metricTypes": ["gpuAllocation", "gpuUtilization", "gpuMemoryUtilization"],
            "window": "last_7d", "numberOfSamples": 28, **extra,
        }
        return exchange([call("get_org_unit_metrics", arguments)])[0]

    def test_idle_holder_average_is_exact_over_full_coverage(self):
        result = payload(self.org_metrics("vision-train", mock.LAB_CLUSTER))
        utilization = result["series"]["gpuUtilization"]
        self.assertEqual(len(utilization), 28)
        self.assertEqual(mean(utilization), 3.0)
        self.assertEqual({point["v"] for point in result["series"]["gpuAllocation"]}, {16})

    def test_unmeasured_holder_has_no_utilization_series(self):
        result = payload(self.org_metrics("genomics", mock.LAB_CLUSTER))
        self.assertTrue(result["partialData"])
        self.assertNotIn("gpuUtilization", result["series"])
        self.assertIn("gpuAllocation", result["series"])

    def test_old_cluster_rejects_memory_metric_and_serves_without_it(self):
        rejected = self.org_metrics("legacy-cv", mock.OLD_CLUSTER)
        self.assertTrue(rejected["result"]["isError"])
        self.assertIn("2.20", rejected["result"]["content"][0]["text"])
        accepted = payload(self.org_metrics(
            "legacy-cv", mock.OLD_CLUSTER, metricTypes=["gpuAllocation", "gpuUtilization"],
        ))
        self.assertIn("gpuUtilization", accepted["series"])

    def test_sparse_holder_reports_only_recorded_samples(self):
        result = payload(self.org_metrics(
            "legacy-asr", mock.OLD_CLUSTER, metricTypes=["gpuAllocation", "gpuUtilization"],
        ))
        self.assertEqual(len(result["series"]["gpuUtilization"]), 7)

    def test_cluster_metrics_ignore_node_pool_filter(self):
        arguments = {
            "clusterId": mock.LAB_CLUSTER, "window": "last_7d", "numberOfSamples": 28,
            "metricTypes": ["allocatedGpu", "totalGpu", "gpuUtilization"],
        }
        responses = exchange([
            call("get_cluster_metrics", arguments, 1),
            call("get_cluster_metrics", {**arguments, "nodePoolName": "l40s-infer"}, 2),
        ])
        self.assertEqual(payload(responses[0])["series"], payload(responses[1])["series"])

    def test_idle_workload_snapshot_lists_only_idle_gpus(self):
        result = payload(exchange([call("list_workloads", {
            "clusterId": mock.LAB_CLUSTER, "gpuIdleOnly": True, "includeDetails": True,
            "sortBy": "idleAllocatedGpus", "sortOrder": "desc",
        })])[0])
        self.assertTrue(result["workloads"])
        for workload in result["workloads"]:
            with self.subTest(workload=workload["name"]):
                self.assertGreater(workload["idleAllocatedGpus"], 0)


if __name__ == "__main__":
    unittest.main()
