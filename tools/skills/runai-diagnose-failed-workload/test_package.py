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


REPO = Path(__file__).resolve().parents[3]
SKILL = REPO / "skills/runai-diagnose-failed-workload"
SERVER = SKILL / "evals/files/mock_mcp.py"
# Import fixtures without creating bytecode inside the publication package.
sys.dont_write_bytecode = True
spec = importlib.util.spec_from_file_location("failed_workload_mock", SERVER)
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
                self.assertRegex(case["id"], r"^[a-zA-Z0-9][a-zA-Z0-9._-]{0,127}$")
                self.assertTrue(case["prompt"])
                self.assertTrue(case["expected_output"])
                self.assertTrue(case["assertions"])
                self.assertNotIn("expected_behavior", case)

    def test_fixture_prompts_reference_existing_workloads(self):
        cases = json.loads((SKILL / "evals/evals.json").read_text())["evals"]
        for case in cases:
            if not case["id"].startswith(("v2-fixture-", "v2-recorded-")):
                continue
            for workload_id in re.findall(
                r"00000000-0000-4000-8000-\d{12}", case["prompt"]
            ):
                self.assertIn(workload_id, mock.FIXTURES.keys() | mock.RECORDED.keys())

    def test_mcp_command_is_executable_and_scoped_to_shared_input(self):
        config = tomllib.loads((SKILL / "evals/environment/mcp_servers.toml").read_text())
        servers = config["mcp_servers"]
        self.assertEqual(len(servers), 1)
        self.assertEqual(servers[0]["transport"], "stdio")
        self.assertEqual(servers[0]["command"], "/workspace/input/mock_mcp.py")
        self.assertTrue(SERVER.stat().st_mode & 0o111)


class ReplayTests(unittest.TestCase):
    def test_live_recordings_replay_exact_responses_over_stdio(self):
        for workload_id, recording in mock.RECORDED.items():
            for tool, expected in recording.items():
                with self.subTest(workload=workload_id, tool=tool):
                    arguments = {"workloadId": workload_id}
                    if tool == "get_workload_history":
                        arguments.update(limit=50, offset=0)
                    actual = payload(exchange([call(tool, arguments)])[0])
                    self.assertEqual(actual, expected)

    def test_recorded_history_paginates_without_losing_or_adding_evidence(self):
        for workload_id, recording in mock.RECORDED.items():
            records = []
            offset = 0
            while True:
                page = payload(exchange([call("get_workload_history", {
                    "workloadId": workload_id, "limit": 2, "offset": offset,
                })])[0])
                records.extend(page["records"])
                if page["next"] is None:
                    break
                self.assertGreater(page["next"], offset)
                offset = page["next"]
            self.assertEqual(records, recording["get_workload_history"]["records"])

    def test_recorded_workloads_can_be_resolved_by_name_project_and_cluster(self):
        for workload_id, recording in mock.RECORDED.items():
            status = recording["get_workload_status"]
            arguments = {
                "name": status["name"], "clusterId": status["clusterId"],
                "projectName": {"operator": "==", "value": status["projectName"]},
            }
            result = payload(exchange([call("list_workloads", arguments)])[0])
            self.assertEqual([w["id"] for w in result["workloads"]], [workload_id])
            arguments["clusterId"] = "20000000-0000-4000-8000-000000000099"
            result = payload(exchange([call("list_workloads", arguments)])[0])
            self.assertEqual(result["workloads"], [])

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
        for name in ("delete_workload", "restart_workload", "get_secrets"):
            with self.subTest(tool=name):
                response = exchange([call(name, {})])[0]
                self.assertEqual(response["error"]["code"], -32601)

    def test_invalid_identity_and_arguments_return_tool_errors(self):
        requests = [
            call("get_workload_status", {}),
            call("get_workload_status", {"workloadId": "missing"}),
            call("get_workload_status", {"workloadId": []}),
            call("get_workload_status", {"workloadId": mock.OOM_ID, "clusterId": mock.CLUSTER_ID}),
            call("get_workload_history", {"workloadId": mock.OOM_ID, "offset": -1}),
        ]
        for response in exchange(requests):
            self.assertTrue(response["result"]["isError"])

    def test_history_pagination_preserves_order_and_every_record(self):
        records = []
        offset = 0
        while True:
            page = payload(exchange([call("get_workload_history", {
                "workloadId": mock.RECURRING_OOM_ID, "limit": 1, "offset": offset,
            })])[0])
            records.extend(page["records"])
            if page["next"] is None:
                break
            self.assertGreater(page["next"], offset)
            offset = page["next"]
        self.assertEqual(records, mock.FIXTURES[mock.RECURRING_OOM_ID]["records"])
        oom_events = [r for r in records if r["meta"]["type"] == "Event"]
        self.assertEqual(len(oom_events), 3)

    def test_exact_name_project_cluster_filter_and_pagination(self):
        arguments = {
            "name": "eval-resolved-failure", "clusterId": mock.CLUSTER_ID,
            "projectName": {"operator": "==", "value": mock.PROJECT_NAME},
            "sortBy": "phaseUpdatedAt", "sortOrder": "desc", "limit": 1,
        }
        first = payload(exchange([call("list_workloads", arguments)])[0])
        self.assertEqual(first["workloads"][0]["id"], mock.RESOLUTION_RUNNING_ID)
        arguments["offset"] = first["nextOffset"]
        second = payload(exchange([call("list_workloads", arguments)])[0])
        self.assertEqual(second["workloads"][0]["id"], mock.RESOLVED_FAILURE_ID)
        arguments["clusterId"] = "20000000-0000-4000-8000-000000000099"
        self.assertEqual(payload(exchange([call("list_workloads", arguments)])[0])["workloads"], [])

    def test_oom_limit_is_distinct_from_request_and_no_peak_is_fabricated(self):
        responses = exchange([
            call("get_workload_pods", {"workloadId": mock.OOM_ID}, 1),
            call("get_workload_metrics", {"workloadId": mock.OOM_ID}, 2),
        ])
        pod = payload(responses[0])["pods"][0]
        self.assertEqual(pod["k8sPhase"], "OOMKilled")
        self.assertEqual(pod["requestedResources"]["cpuMemory"], {"request": "24Gi", "limit": "32Gi"})
        self.assertEqual(payload(responses[1])["measurements"], [])

    def test_sigkill_fixture_does_not_supply_an_oom_reason(self):
        status = payload(exchange([call("get_workload_status", {
            "workloadId": mock.UNEXPLAINED_SIGKILL_ID,
        })])[0])
        self.assertEqual(status["phaseReason"], "ContainerExit")
        self.assertIn("137", status["phaseMessage"])
        self.assertNotEqual(status["events"][0]["reason"], "OOMKilled")

    def test_injection_is_returned_as_evidence_not_executed(self):
        result = payload(exchange([call("get_workload_history", {
            "workloadId": mock.UNTRUSTED_EVENT_ID,
        })])[0])
        events = [r["spec"]["event"] for r in result["records"] if r["meta"]["type"] == "Event"]
        self.assertTrue(any("kubectl delete" in e["message"] for e in events))
        self.assertTrue(any("ValueError" in e["message"] for e in events))

    def test_protocol_reports_malformed_json_and_recovers(self):
        result = subprocess.run(
            [sys.executable, str(SERVER)],
            input='invalid json\n{"jsonrpc":"2.0","id":2,"method":"tools/list"}\n',
            text=True, capture_output=True, timeout=10, check=True,
        )
        responses = [json.loads(line) for line in result.stdout.splitlines()]
        self.assertEqual(responses[0]["error"]["code"], -32700)
        self.assertEqual(responses[1]["id"], 2)


if __name__ == "__main__":
    unittest.main()
