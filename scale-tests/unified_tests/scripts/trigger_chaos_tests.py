#!/usr/bin/env python3
"""
Script: Trigger Node Chaos Tests
=================================

This script triggers scale tests with node chaos enabled to test system resilience
when nodes fail during workload execution.

Node chaos simulates node failures by bringing down GPU nodes during the test,
allowing you to verify that workloads recover properly.

Prerequisites:
- GitHub CLI (gh) installed and authenticated
- Access to trigger workflows in the repository

Usage:
  # Run with default chaos settings (3 nodes down, 1m delay)
  python trigger_chaos_tests.py --branch my-branch

  # Customize chaos parameters
  python trigger_chaos_tests.py --branch my-branch --chaos-count 5 --chaos-start-delay 2m

  # Run specific test configuration
  python trigger_chaos_tests.py --subset medium-configs --branch my-branch

  # Dry run to see what would be triggered
  python trigger_chaos_tests.py --branch my-branch --dry-run
"""

import subprocess
import sys
import time
import argparse
from typing import Dict, List, Any

GITHUB_REPO = "dsx-ai-factory/run-ai"
WORKFLOW_FILE = "scale-test-manual-trigger.yml"

# Configuration tuple format:
# (name, total_nodes, workload_count, project_count, resource_tier, chaos_count)
#
# - name: Identifier for the test run
# - total_nodes: Number of GPU nodes in the cluster
# - workload_count: Number of workloads to create
# - project_count: Number of projects
# - resource_tier: Resource tier for cluster services (small/medium/large)
# - chaos_count: Number of nodes to bring down during the test
TEST_CONFIGURATIONS = {
    "small-chaos": [
        ("chaos-small-2nodes", 3, 20, 3, "small", 2),
        ("chaos-small-all", 3, 20, 3, "small", 3),
    ],
    "medium-chaos": [
        ("chaos-medium-4nodes", 6, 50, 5, "small", 4),
        ("chaos-medium-5nodes", 6, 50, 5, "small", 5),
        ("chaos-medium-all", 6, 50, 5, "small", 6),
    ],
    "large-chaos": [
        ("chaos-large-8nodes", 10, 100, 10, "medium", 8),
        ("chaos-large-9nodes", 10, 100, 10, "medium", 9),
        ("chaos-large-all", 10, 100, 10, "medium", 10),
    ],
    "stress-chaos": [
        ("chaos-stress-6nodes", 8, 80, 8, "small", 6),
        ("chaos-stress-7nodes", 8, 80, 8, "small", 7),
        ("chaos-stress-all", 8, 80, 8, "small", 8),
    ],
}


def trigger_chaos_test(
    config_name: str,
    total_nodes: int,
    workload_count: int,
    project_count: int,
    resource_tier: str,
    chaos_count: int,
    chaos_start_delay: str,
    chaos_delay_between: str,
    branch: str,
    run_group: str = None,
    cluster_name: str = None,
    dry_run: bool = False,
) -> bool:
    """
    Trigger a single chaos test workflow run.

    Args:
        config_name: Name identifier for this configuration
        total_nodes: Total number of nodes
        workload_count: Number of workloads
        project_count: Number of projects
        resource_tier: Resource tier (small/medium/large)
        chaos_count: Number of nodes to bring down
        chaos_start_delay: Delay before starting chaos (e.g., '1m', '30s')
        chaos_delay_between: Delay between node failures (e.g., '10s')
        branch: Git branch to run from
        run_group: Optional custom run group name
        cluster_name: Optional existing cluster name to run against
        dry_run: If True, don't actually trigger

    Returns:
        True if successful, False otherwise
    """
    scale_test_params = (
        f"RESOURCES={resource_tier};"
        f"PROJECT_COUNT={project_count};"
        f"WORKLOAD_COUNT={workload_count};"
        f"WORKLOAD_TYPE=Distributed;"
        f"DISTRIBUTED_FRAMEWORK=PyTorch;"
        f"NODEPOOL_COUNT=0;"
        f"WORKLOAD_CREATE_TIMEOUT=90m;"
        f"PROJECT_CREATE_TIMEOUT=60m;"
        f"KWOK_ENABLED=false;"
        f"GPU_PER_WORKLOAD=1;"
        f"NODE_CHAOS_ENABLED=true;"
        f"NODE_CHAOS_COUNT={chaos_count};"
        f"NODE_CHAOS_START_DELAY={chaos_start_delay};"
        f"NODE_CHAOS_DELAY_BETWEEN={chaos_delay_between}"
    )

    cluster_config = (
        '{"gpu-type":"fake",'
        '"fake-gpu-count":"8",'
        '"kwok-enabled":"false",'
        f'"node-count":"{total_nodes}",'
        '"cloud-provider":"aws",'
        '"prometheus-remote-write-enabled":false,'
        '"kwok-ratio":"20"}'
    )

    run_group_name = run_group if run_group else f"chaos-test-{config_name}"

    cmd = [
        "gh", "workflow", "run", WORKFLOW_FILE,
        "--repo", GITHUB_REPO,
        "--ref", branch,
        "-f", f"scale-test-params={scale_test_params}",
        "-f", f"cluster-config={cluster_config}",
        "-f", f"run-group={run_group_name}",
        "-f", f"tests-branch={branch}",
        "-f", "delete-env=true",
        "-f", "cluster-env=test",
        "-f", "cp-env=test",
    ]

    if cluster_name:
        cmd.extend(["-f", f"cluster-name={cluster_name}"])

    print(f"  Triggering: {config_name}")
    print(f"    {total_nodes} nodes, {workload_count} workloads, {project_count} projects")
    print(f"    Chaos: {chaos_count} nodes down after {chaos_start_delay}, "
          f"{chaos_delay_between} between failures")

    if dry_run:
        print(f"    [DRY RUN] Would execute: gh workflow run ...")
        return True

    try:
        result = subprocess.run(cmd, capture_output=True, text=True, check=True)
        print(f"    ✓ Triggered successfully")
        return True
    except subprocess.CalledProcessError as e:
        print(f"    ✗ Failed: {e.stderr}", file=sys.stderr)
        return False


def main():
    """Main execution function."""
    parser = argparse.ArgumentParser(
        description="Trigger scale tests with node chaos enabled"
    )
    parser.add_argument(
        "--subset",
        choices=["all", "small-chaos", "medium-chaos", "large-chaos", "stress-chaos"],
        default="small-chaos",
        help="Which subset of chaos tests to run (default: small-chaos)",
    )
    parser.add_argument(
        "--config-name",
        help="Run a specific configuration by name (e.g., chaos-medium-2nodes)",
    )
    parser.add_argument(
        "--branch",
        default="master",
        help="Git branch to run tests from",
    )
    parser.add_argument(
        "--cluster-name",
        help="Existing cluster name to run against (skips cluster creation)",
    )
    parser.add_argument(
        "--chaos-count",
        type=int,
        help="Override the number of nodes to bring down",
    )
    parser.add_argument(
        "--chaos-start-delay",
        default="1m",
        help="Delay before starting chaos (default: 1m)",
    )
    parser.add_argument(
        "--chaos-delay-between",
        default="10s",
        help="Delay between node failures (default: 10s)",
    )
    parser.add_argument(
        "--delay",
        type=int,
        default=5,
        help="Delay in seconds between triggering tests (default: 5)",
    )
    parser.add_argument(
        "--run-group",
        help="Custom run group name",
    )
    parser.add_argument(
        "--dry-run",
        action="store_true",
        help="Print configurations without triggering workflows",
    )

    args = parser.parse_args()

    print("=" * 80)
    print("Node Chaos Scale Tests")
    print("=" * 80)
    print()

    if not args.dry_run:
        try:
            subprocess.run(["gh", "--version"], capture_output=True, check=True)
        except (subprocess.CalledProcessError, FileNotFoundError):
            print("ERROR: GitHub CLI (gh) is not installed or not in PATH.", file=sys.stderr)
            print("Install it from: https://cli.github.com/", file=sys.stderr)
            sys.exit(1)

    if args.config_name:
        configs_to_run = []
        for subset in TEST_CONFIGURATIONS.values():
            for config in subset:
                if config[0] == args.config_name:
                    configs_to_run.append(config)
                    break
        if not configs_to_run:
            print(f"ERROR: Configuration '{args.config_name}' not found.", file=sys.stderr)
            print("Available configurations:")
            for subset_name, configs in TEST_CONFIGURATIONS.items():
                for config in configs:
                    print(f"  {config[0]} ({subset_name})")
            sys.exit(1)
    elif args.subset == "all":
        configs_to_run = []
        for subset in TEST_CONFIGURATIONS.values():
            configs_to_run.extend(subset)
    else:
        configs_to_run = TEST_CONFIGURATIONS[args.subset]

    print(f"Subset: {args.subset}")
    print(f"Branch: {args.branch}")
    if args.cluster_name:
        print(f"Cluster: {args.cluster_name} (existing)")
    else:
        print(f"Cluster: will create new")
    print(f"Chaos Start Delay: {args.chaos_start_delay}")
    print(f"Chaos Delay Between: {args.chaos_delay_between}")
    print(f"Total configurations: {len(configs_to_run)}")
    print()

    if args.dry_run:
        print("DRY RUN - No workflows will be triggered")
        print()
        for name, nodes, workloads, projects, tier, default_chaos_count in configs_to_run:
            chaos_count = args.chaos_count if args.chaos_count else default_chaos_count
            print(f"  {name}: {nodes} nodes, {workloads} workloads, "
                  f"{projects} projects ({tier}), {chaos_count} nodes chaos")
        print()
        return

    print("WARNING: This will trigger chaos test workflow runs.")
    print("Nodes will be brought down during the test to simulate failures.")
    print()
    response = input("Continue? (yes/no): ")
    if response.lower() not in ["yes", "y"]:
        print("Aborted.")
        sys.exit(0)

    print()
    print("Triggering chaos test workflow runs...")
    print()

    successful = 0
    failed = 0

    for name, nodes, workloads, projects, tier, default_chaos_count in configs_to_run:
        chaos_count = args.chaos_count if args.chaos_count else default_chaos_count

        success = trigger_chaos_test(
            config_name=name,
            total_nodes=nodes,
            workload_count=workloads,
            project_count=projects,
            resource_tier=tier,
            chaos_count=chaos_count,
            chaos_start_delay=args.chaos_start_delay,
            chaos_delay_between=args.chaos_delay_between,
            branch=args.branch,
            run_group=args.run_group,
            cluster_name=args.cluster_name,
            dry_run=args.dry_run,
        )

        if success:
            successful += 1
        else:
            failed += 1

        if args.delay > 0:
            time.sleep(args.delay)

        print()

    print("=" * 80)
    print("Summary")
    print("=" * 80)
    print(f"Triggered: {successful}")
    print(f"Failed: {failed}")
    print()

    if successful > 0:
        print("✓ Chaos test workflow runs triggered successfully!")
        print()
        print("Monitor progress:")
        print(f"  gh run list --workflow={WORKFLOW_FILE} --repo {GITHUB_REPO}")
        print()
        print("View specific run:")
        print(f"  gh run watch <run-id> --repo {GITHUB_REPO}")
    else:
        print("✗ No workflows were triggered successfully.")
        print("Check GitHub CLI authentication and repository access.")


if __name__ == "__main__":
    main()
