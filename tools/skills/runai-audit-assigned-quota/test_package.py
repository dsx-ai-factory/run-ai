#!/usr/bin/env python3
# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0
"""Offline checks for the published package and its actual MCP wire protocol."""

import importlib.util
import json
from pathlib import Path
import subprocess
import sys
import tomllib
import unittest


NAME = "runai-audit-assigned-quota"
REPO = Path(__file__).resolve().parents[3]
SKILL = REPO / "skills" / NAME
SERVER = SKILL / "evals/files/mock_mcp.py"
# Import fixtures without creating bytecode inside the publication package.
sys.dont_write_bytecode = True
spec = importlib.util.spec_from_file_location("quota_mock", SERVER)
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
        for name in ("update_project_quota", "rebalance_quota", "delete_workload"):
            with self.subTest(tool=name):
                response = exchange([call(name, {})])[0]
                self.assertEqual(response["error"]["code"], -32601)

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
    def test_node_pool_2_is_over_subscribed_and_full(self):
        cells = payload(exchange([call("list_project_resources", {"nodePoolName": "node-pool-2"})])[0])
        deserved = sum(cell["gpu"]["deserved"] for cell in cells["resources"])
        inventory = payload(exchange([call("get_cluster_physical_inventory", {"nodePoolName": "node-pool-2"})])[0])
        pool = inventory["nodePools"][0]["gpu"]
        self.assertGreater(deserved, pool["total"])
        self.assertEqual(pool["free"], 0)

    def test_node_pool_filter_drops_unscheduled_pending(self):
        scoped = payload(exchange([call("get_workloads_summary", {
            "orgType": "project", "orgName": "hpc-fill", "nodePoolName": "node-pool-2",
        })])[0])
        unscoped = payload(exchange([call("get_workloads_summary", {
            "orgType": "project", "orgName": "hpc-fill",
        })])[0])
        self.assertEqual(scoped["counts"]["pending"], 0)
        self.assertEqual(unscoped["counts"]["pending"], 2)

    def pending_ids(self, project):
        listed = payload(exchange([call("list_workloads", {
            "phase": "Pending", "projectName": {"operator": "==", "value": project},
            "includeDetails": True,
        })])[0])
        return listed["workloads"]

    def test_pending_ids_match_summary_count(self):
        for project, (count, _, _) in mock.PENDING.items():
            with self.subTest(project=project):
                self.assertEqual(len(self.pending_ids(project)), count)

    def test_self_inflicted_pending_reason_is_not_capacity_wait(self):
        rows = self.pending_ids("quota-hog")
        self.assertFalse(rows[0]["preemptible"])
        status = payload(exchange([call("get_workload_status", {"workloadId": rows[0]["id"]})])[0])
        self.assertEqual(status["phase"], "Pending")
        self.assertIn("quota exceeded for non-preemptible", status["phaseMessage"])

    def test_status_requires_a_known_workload_id(self):
        responses = exchange([
            call("get_workload_status", {}, 1),
            call("get_workload_status", {"workloadId": "missing"}, 2),
            call("get_workload_status", {"orgName": "quota-hog"}, 3),
        ])
        for response in responses:
            self.assertTrue(response["result"]["isError"])

    def test_department_rollups_reconcile_with_project_cells(self):
        for (department, pool), (deserved, allocated) in mock.DEPTS.items():
            cells = [c for c in mock.CELLS if c[0] == pool and c[2] == department]
            with self.subTest(department=department, pool=pool):
                self.assertEqual(sum(c[3] for c in cells), deserved)
                self.assertEqual(sum(c[4] for c in cells), allocated)


if __name__ == "__main__":
    unittest.main()
