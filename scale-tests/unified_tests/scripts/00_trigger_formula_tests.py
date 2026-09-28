#!/usr/bin/env python3

# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

"""
Script 0: Trigger Formula Derivation Tests
===========================================

This script triggers multiple runs of the scale-test-manual-trigger.yml
workflow with different configurations to collect data for formula derivation.

Uses the existing manual trigger workflow so it can be run from any branch.

Prerequisites:
- GitHub CLI (gh) installed and authenticated
- Access to trigger workflows in the repository

Usage:
  python 00_trigger_formula_tests.py --subset all
  python 00_trigger_formula_tests.py --subset small-configs
  python 00_trigger_formula_tests.py --branch your-branch-name
"""

import subprocess
import sys
import time
import argparse
from typing import Dict, List, Any

# Configuration
GITHUB_REPO = "dsx-ai-factory/run-ai"
WORKFLOW_FILE = "scale-test-manual-trigger.yml"

# Test configurations to run
# Format: name, nodes, workloads, projects, resource_tier
TEST_CONFIGURATIONS = {
    "small-configs": [
        ("formula-small-low", 30, 480, 10, "small")
    ],
    "medium-configs": [
        ("formula-medium-med", 100, 1600, 50, "medium"),
    ],
    "large-configs": [
        ("formula-large-high", 500, 9000, 200, "large"),
    ],
    "edge-cases": [
        # Sparse: many nodes, few workloads
        ("formula-sparse-small", 50, 100, 5, "small"),
        ("formula-sparse-large", 200, 500, 20, "medium"),
        # Dense: few nodes, many workloads
        ("formula-dense-small", 20, 1500, 30, "small"),
        ("formula-dense-medium", 60, 4000, 80, "medium"),
        # Many projects
        ("formula-many-projects-medium", 120, 2000, 200, "medium"),
        ("formula-many-projects-large", 300, 5000, 500, "large"),
    ],
}


def trigger_workflow_run(
    config_name: str,
    total_nodes: int,
    workload_count: int,
    project_count: int,
    resource_tier: str,
    branch: str,
    custom_args,
    run_group: str = None,
    scale_group: str = None,
) -> bool:
    """
    Trigger a single workflow run with the given configuration.

    Args:
        config_name: Name identifier for this configuration
        total_nodes: Total number of nodes
        workload_count: Number of workloads
        project_count: Number of projects
        resource_tier: Resource tier (small/medium/large/custom)
        branch: Git branch to run from
        custom_args: Parsed command-line arguments with custom resource specifications
        run_group: Optional custom run group name (overrides default naming)
        scale_group: Optional scale group identifier to override resource_tier

    Returns:
        True if successful, False otherwise
    """
    # Calculate nodepool configuration (distribute evenly across 2 pools)
    nodepool_count = 2 if total_nodes >= 40 else 1
    nodes_per_pool = total_nodes // nodepool_count

    # Determine actual resource tier
    actual_tier = "custom" if custom_args.custom_resources else resource_tier

    # Determine scale group (use custom scale_group if provided, otherwise use resource_tier)
    actual_scale_group = scale_group if scale_group else resource_tier

    # Build scale test params
    scale_test_params = (
        f"RESOURCES={actual_tier};"
        f"PROJECT_COUNT={project_count};"
        f"WORKLOAD_COUNT={workload_count};"
        f"WORKLOAD_TYPE=Training;"
        f"WORKLOAD_SLEEP_TIMES=10;"
        f"RUN_INFINITE_WORKLOADS=false;"
        f"NODEPOOL_COUNT={nodepool_count};"
        f"NODES_PER_POOL={nodes_per_pool};"
        f"WORKLOAD_CREATE_TIMEOUT=90m;"
        f"PROJECT_CREATE_TIMEOUT=60m;"
        f"KWOK_ENABLED={str(not custom_args.no_kwok).lower()};"
        f"SCALE_GROUP={actual_scale_group}"
    )

    # Add QPS if specified
    if custom_args.qps:
        scale_test_params += f";CONTROL_PLANE_QPS={custom_args.qps}"

    # Add custom resource params if specified
    if custom_args.custom_resources:
        if custom_args.scheduling_cpu_request:
            scale_test_params += f";SCHEDULING_SERVICES_CPU_REQUEST={custom_args.scheduling_cpu_request}"
        if custom_args.scheduling_cpu_limit:
            scale_test_params += f";SCHEDULING_SERVICES_CPU_LIMIT={custom_args.scheduling_cpu_limit}"
        if custom_args.scheduling_memory_request:
            scale_test_params += f";SCHEDULING_SERVICES_MEMORY_REQUEST={custom_args.scheduling_memory_request}"
        if custom_args.scheduling_memory_limit:
            scale_test_params += f";SCHEDULING_SERVICES_MEMORY_LIMIT={custom_args.scheduling_memory_limit}"

        if custom_args.workload_cpu_request:
            scale_test_params += f";WORKLOAD_SERVICES_CPU_REQUEST={custom_args.workload_cpu_request}"
        if custom_args.workload_cpu_limit:
            scale_test_params += f";WORKLOAD_SERVICES_CPU_LIMIT={custom_args.workload_cpu_limit}"
        if custom_args.workload_memory_request:
            scale_test_params += f";WORKLOAD_SERVICES_MEMORY_REQUEST={custom_args.workload_memory_request}"
        if custom_args.workload_memory_limit:
            scale_test_params += f";WORKLOAD_SERVICES_MEMORY_LIMIT={custom_args.workload_memory_limit}"

        if custom_args.sync_cpu_request:
            scale_test_params += f";SYNC_SERVICES_CPU_REQUEST={custom_args.sync_cpu_request}"
        if custom_args.sync_cpu_limit:
            scale_test_params += f";SYNC_SERVICES_CPU_LIMIT={custom_args.sync_cpu_limit}"
        if custom_args.sync_memory_request:
            scale_test_params += f";SYNC_SERVICES_MEMORY_REQUEST={custom_args.sync_memory_request}"
        if custom_args.sync_memory_limit:
            scale_test_params += f";SYNC_SERVICES_MEMORY_LIMIT={custom_args.sync_memory_limit}"

        if custom_args.training_cpu_request:
            scale_test_params += f";TRAINING_OPERATOR_CPU_REQUEST={custom_args.training_cpu_request}"
        if custom_args.training_cpu_limit:
            scale_test_params += f";TRAINING_OPERATOR_CPU_LIMIT={custom_args.training_cpu_limit}"
        if custom_args.training_memory_request:
            scale_test_params += f";TRAINING_OPERATOR_MEMORY_REQUEST={custom_args.training_memory_request}"
        if custom_args.training_memory_limit:
            scale_test_params += f";TRAINING_OPERATOR_MEMORY_LIMIT={custom_args.training_memory_limit}"

        if custom_args.gpu_operator_cpu_request:
            scale_test_params += f";FAKE_GPU_OPERATOR_CPU_REQUEST={custom_args.gpu_operator_cpu_request}"
        if custom_args.gpu_operator_cpu_limit:
            scale_test_params += f";FAKE_GPU_OPERATOR_CPU_LIMIT={custom_args.gpu_operator_cpu_limit}"
        if custom_args.gpu_operator_memory_request:
            scale_test_params += f";FAKE_GPU_OPERATOR_MEMORY_REQUEST={custom_args.gpu_operator_memory_request}"
        if custom_args.gpu_operator_memory_limit:
            scale_test_params += f";FAKE_GPU_OPERATOR_MEMORY_LIMIT={custom_args.gpu_operator_memory_limit}"

        if custom_args.postgres_cpu_request:
            scale_test_params += f";POSTGRES_CPU_REQUEST={custom_args.postgres_cpu_request}"
        if custom_args.postgres_cpu_limit:
            scale_test_params += f";POSTGRES_CPU_LIMIT={custom_args.postgres_cpu_limit}"
        if custom_args.postgres_memory_request:
            scale_test_params += f";POSTGRES_MEMORY_REQUEST={custom_args.postgres_memory_request}"
        if custom_args.postgres_memory_limit:
            scale_test_params += f";POSTGRES_MEMORY_LIMIT={custom_args.postgres_memory_limit}"

    # Determine GPU type based on whether real GPUs are specified
    gpu_type = "real" if custom_args.gpu_product else "fake"

    # Build cluster config (enable Prometheus remote write for Thanos)
    # Include real-node-count and real-gpu-product directly in cluster-config
    # to avoid bash parsing issues with the gpus JSON array input
    gpu_nodes = custom_args.gpu_nodes_count if custom_args.gpu_nodes_count else 1
    kwok_enabled = not custom_args.no_kwok
    # When KWOK is enabled, we need fewer real nodes (KWOK simulates the rest)
    # When KWOK is disabled, we need all real nodes
    real_node_count = min(total_nodes // 10, 50) if kwok_enabled else total_nodes
    cluster_config_parts = [
        f'"gpu-type":"{gpu_type}"',
        '"fake-gpu-count":"8"',  # 8 GPUs per node (DGX A100 standard)
        f'"kwok-enabled":{str(kwok_enabled).lower()}',  # Boolean, not string
        f'"node-count":"{real_node_count}"',  # Real nodes
        '"cloud-provider":"aws"',
        '"kwok-ratio":"20"',
        '"prometheus-remote-write-enabled":true',  # Boolean, not string
    ]
    # Add real GPU config if specified
    if custom_args.gpu_product:
        cluster_config_parts.append(f'"real-node-count":"{gpu_nodes}"')
        cluster_config_parts.append(f'"real-gpu-product":"{custom_args.gpu_product}"')

    cluster_config = "{" + ",".join(cluster_config_parts) + "}"

    # Determine run group name (use custom run_group if provided)
    run_group_name = run_group if run_group else f"formula-derivation-{config_name}"

    # Build gh workflow run command
    cmd = [
        "gh", "workflow", "run", WORKFLOW_FILE,
        "--repo", GITHUB_REPO,
        "--ref", branch,
        "-f", f"scale-test-params={scale_test_params}",
        "-f", f"cluster-config={cluster_config}",
        "-f", f"run-group={run_group_name}",
        "-f", f"tests-branch={branch}",  # Tell workflow which branch to checkout for tests
        "-f", "delete-env=true",
        "-f", "cluster-env=test",
        "-f", "cp-env=test",
    ]

    # Set GPU info for display (actual GPU config is in cluster-config)
    gpu_info = "fake"
    if custom_args.gpu_product:
        gpu_info = f"{custom_args.gpu_product} x {gpu_nodes} nodes"

    print(f"  Triggering: {config_name}")
    print(f"    {total_nodes} nodes ({nodepool_count}x{nodes_per_pool}), "
          f"{workload_count} workloads, {project_count} projects, {actual_tier}")
    print(f"    GPU: {gpu_info}")
    print(f"    KWOK: {'enabled' if kwok_enabled else 'disabled'} (real nodes: {real_node_count})")

    try:
        result = subprocess.run(cmd, capture_output=True, text=True, check=True)
        print(f"    ✓ Triggered successfully")
        # Get the workflow run URL
        try:
            list_cmd = [
                "gh", "run", "list",
                "--repo", GITHUB_REPO,
                "--workflow", WORKFLOW_FILE,
                "--limit", "1",
                "--json", "url",
                "--jq", ".[0].url"
            ]
            url_result = subprocess.run(list_cmd, capture_output=True, text=True, check=True)
            run_url = url_result.stdout.strip()
            if run_url:
                print(f"    🔗 {run_url}")
        except subprocess.CalledProcessError:
            print(f"    🔗 https://github.com/{GITHUB_REPO}/actions/workflows/{WORKFLOW_FILE}")
        return True
    except subprocess.CalledProcessError as e:
        print(f"    ✗ Failed: {e.stderr}", file=sys.stderr)
        return False


def main():
    """Main execution function."""
    parser = argparse.ArgumentParser(
        description="Trigger multiple scale test runs for formula derivation"
    )
    parser.add_argument(
        "--subset",
        choices=["all", "small-configs", "medium-configs", "large-configs", "edge-cases"],
        default="all",
        help="Which subset of tests to run",
    )
    parser.add_argument(
        "--branch",
        default="main",
        help="Git branch to run tests from",
    )
    parser.add_argument(
        "--delay",
        type=int,
        default=5,
        help="Delay in seconds between triggering tests (default: 5)",
    )
    parser.add_argument(
        "--dry-run",
        action="store_true",
        help="Print configurations without triggering workflows",
    )
    parser.add_argument(
        "--run-group",
        help="Custom run group name (overrides default 'formula-derivation-{config_name}' naming)",
    )
    parser.add_argument(
        "--scale-group",
        help="Custom scale group identifier (overrides resource tier: small/medium/large)",
    )
    parser.add_argument(
        "--qps",
        type=int,
        help="Control Plane QPS (queries per second) - controls concurrent API calls",
    )

    # Custom resource arguments
    parser.add_argument(
        "--custom-resources",
        action="store_true",
        help="Use custom resource values instead of tier presets",
    )

    # Scheduling Services
    parser.add_argument("--scheduling-cpu-request", help="Scheduling services CPU request (e.g., '2', '500m')")
    parser.add_argument("--scheduling-cpu-limit", help="Scheduling services CPU limit")
    parser.add_argument("--scheduling-memory-request", help="Scheduling services memory request (e.g., '4Gi', '512Mi')")
    parser.add_argument("--scheduling-memory-limit", help="Scheduling services memory limit")

    # Workload Services
    parser.add_argument("--workload-cpu-request", help="Workload services CPU request")
    parser.add_argument("--workload-cpu-limit", help="Workload services CPU limit")
    parser.add_argument("--workload-memory-request", help="Workload services memory request")
    parser.add_argument("--workload-memory-limit", help="Workload services memory limit")

    # Sync Services
    parser.add_argument("--sync-cpu-request", help="Sync services CPU request")
    parser.add_argument("--sync-cpu-limit", help="Sync services CPU limit")
    parser.add_argument("--sync-memory-request", help="Sync services memory request")
    parser.add_argument("--sync-memory-limit", help="Sync services memory limit")

    # Training Operator
    parser.add_argument("--training-cpu-request", help="Training operator CPU request")
    parser.add_argument("--training-cpu-limit", help="Training operator CPU limit")
    parser.add_argument("--training-memory-request", help="Training operator memory request")
    parser.add_argument("--training-memory-limit", help="Training operator memory limit")

    # Fake GPU Operator
    parser.add_argument("--gpu-operator-cpu-request", help="Fake GPU operator CPU request")
    parser.add_argument("--gpu-operator-cpu-limit", help="Fake GPU operator CPU limit")
    parser.add_argument("--gpu-operator-memory-request", help="Fake GPU operator memory request")
    parser.add_argument("--gpu-operator-memory-limit", help="Fake GPU operator memory limit")

    # PostgreSQL
    parser.add_argument("--postgres-cpu-request", help="PostgreSQL CPU request")
    parser.add_argument("--postgres-cpu-limit", help="PostgreSQL CPU limit")
    parser.add_argument("--postgres-memory-request", help="PostgreSQL memory request")
    parser.add_argument("--postgres-memory-limit", help="PostgreSQL memory limit")

    # GPU configuration
    parser.add_argument(
        "--gpu-product",
        help="GPU product name (e.g., 'Tesla T4-1', 'Tesla A100-40GB-8'). Leave empty for fake GPUs.",
    )
    parser.add_argument(
        "--gpu-nodes-count",
        type=int,
        help="Number of GPU nodes to provision (required if --gpu-product is specified).",
    )

    # KWOK configuration
    parser.add_argument(
        "--no-kwok",
        action="store_true",
        help="Disable KWOK (use real nodes only instead of simulated nodes).",
    )

    args = parser.parse_args()

    print("=" * 80)
    print("Formula Derivation Scale Tests")
    print("=" * 80)
    print()

    # Check gh CLI is installed
    if not args.dry_run:
        try:
            subprocess.run(["gh", "--version"], capture_output=True, check=True)
        except (subprocess.CalledProcessError, FileNotFoundError):
            print("ERROR: GitHub CLI (gh) is not installed or not in PATH.", file=sys.stderr)
            print("Install it from: https://cli.github.com/", file=sys.stderr)
            sys.exit(1)

    # Determine which tests to run
    if args.subset == "all":
        configs_to_run = []
        for subset in TEST_CONFIGURATIONS.values():
            configs_to_run.extend(subset)
    else:
        configs_to_run = TEST_CONFIGURATIONS[args.subset]

    print(f"Subset: {args.subset}")
    print(f"Branch: {args.branch}")
    print(f"Total configurations: {len(configs_to_run)}")

    if args.custom_resources:
        print(f"Custom Resources: ENABLED")
        custom_params = []
        if args.scheduling_cpu_request:
            custom_params.append(f"scheduling-cpu-request={args.scheduling_cpu_request}")
        if args.scheduling_cpu_limit:
            custom_params.append(f"scheduling-cpu-limit={args.scheduling_cpu_limit}")
        if args.scheduling_memory_request:
            custom_params.append(f"scheduling-memory-request={args.scheduling_memory_request}")
        if args.scheduling_memory_limit:
            custom_params.append(f"scheduling-memory-limit={args.scheduling_memory_limit}")
        if args.workload_cpu_request:
            custom_params.append(f"workload-cpu-request={args.workload_cpu_request}")
        if args.workload_cpu_limit:
            custom_params.append(f"workload-cpu-limit={args.workload_cpu_limit}")
        if args.workload_memory_request:
            custom_params.append(f"workload-memory-request={args.workload_memory_request}")
        if args.workload_memory_limit:
            custom_params.append(f"workload-memory-limit={args.workload_memory_limit}")
        if args.sync_cpu_request:
            custom_params.append(f"sync-cpu-request={args.sync_cpu_request}")
        if args.sync_cpu_limit:
            custom_params.append(f"sync-cpu-limit={args.sync_cpu_limit}")
        if args.sync_memory_request:
            custom_params.append(f"sync-memory-request={args.sync_memory_request}")
        if args.sync_memory_limit:
            custom_params.append(f"sync-memory-limit={args.sync_memory_limit}")
        if args.training_cpu_request:
            custom_params.append(f"training-cpu-request={args.training_cpu_request}")
        if args.training_cpu_limit:
            custom_params.append(f"training-cpu-limit={args.training_cpu_limit}")
        if args.training_memory_request:
            custom_params.append(f"training-memory-request={args.training_memory_request}")
        if args.training_memory_limit:
            custom_params.append(f"training-memory-limit={args.training_memory_limit}")
        if args.gpu_operator_cpu_request:
            custom_params.append(f"gpu-operator-cpu-request={args.gpu_operator_cpu_request}")
        if args.gpu_operator_cpu_limit:
            custom_params.append(f"gpu-operator-cpu-limit={args.gpu_operator_cpu_limit}")
        if args.gpu_operator_memory_request:
            custom_params.append(f"gpu-operator-memory-request={args.gpu_operator_memory_request}")
        if args.gpu_operator_memory_limit:
            custom_params.append(f"gpu-operator-memory-limit={args.gpu_operator_memory_limit}")

        if custom_params:
            print(f"  {', '.join(custom_params)}")

    print()

    if args.dry_run:
        print("DRY RUN - No workflows will be triggered")
        print()
        actual_tier = "custom" if args.custom_resources else None
        for name, nodes, workloads, projects, tier in configs_to_run:
            display_tier = actual_tier if actual_tier else tier
            print(f"  {name}: {nodes} nodes, {workloads} workloads, {projects} projects ({display_tier})")
        print()
        return

    # Confirm before triggering
    print("WARNING: This will trigger multiple workflow runs.")
    print("Each test may take 30-90 minutes to complete.")
    print("Tests will run in parallel.")
    print()
    response = input("Continue? (yes/no): ")
    if response.lower() not in ["yes", "y"]:
        print("Aborted.")
        sys.exit(0)

    print()
    print("Triggering workflow runs...")
    print()

    # Trigger each configuration
    successful = 0
    failed = 0

    for name, nodes, workloads, projects, tier in configs_to_run:
        success = trigger_workflow_run(
            name, nodes, workloads, projects, tier, args.branch, args,
            run_group=args.run_group,
            scale_group=args.scale_group
        )
        if success:
            successful += 1
        else:
            failed += 1

        # Small delay to avoid overwhelming the API
        if args.delay > 0:
            time.sleep(args.delay)

        print()

    # Summary
    print("=" * 80)
    print("Summary")
    print("=" * 80)
    print(f"Triggered: {successful}")
    print(f"Failed: {failed}")
    print()

    if successful > 0:
        print("✓ Workflow runs triggered successfully!")
        print()
        print("Monitor progress:")
        print(f"  gh run list --workflow={WORKFLOW_FILE} --repo {GITHUB_REPO}")
        print()
        print("View specific run:")
        print(f"  gh run watch <run-id> --repo {GITHUB_REPO}")
        print()
        print("After tests complete:")
        print("  1. Run: python 01_collect_test_metadata.py")
        print("  2. Run: python 02_collect_resource_metrics.py")
        print("  3. Continue with analysis scripts (03, 04, 05)")
    else:
        print("✗ No workflows were triggered successfully.")
        print("Check GitHub CLI authentication and repository access.")


if __name__ == "__main__":
    main()
