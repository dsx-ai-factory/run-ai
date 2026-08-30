#!/usr/bin/env python3
"""
Parameter Sweep: Multi-Service Group Resources
===============================================

This script performs a parameter sweep to find optimal resource allocation
for Run:AI services by running scale tests with different CPU and memory combinations.

Supports sweeping multiple service groups:
- scheduling: scheduler, admission-controller
- sync: cluster-sync, asset-service
- workload: workload-controller, workload-exporter

You can choose which groups to sweep (dynamic) and which to keep static.

Creates a cartesian product of CPU and Memory values, testing all combinations.
Each test is assigned an indicative run group name based on the groups being swept.

Examples:
  # Sweep scheduling only (other groups static)
  python parameter_sweep_scheduling.py \
    --sweep-groups scheduling \
    --cpu-start 0.5 --cpu-end 2.0 --cpu-step 0.5 \
    --memory-start 1.0 --memory-end 4.0 --memory-step 1.0 \
    --branch my-branch

  # Sweep scheduling and workload together
  python parameter_sweep_scheduling.py \
    --sweep-groups scheduling workload \
    --cpu-start 1.0 --cpu-end 2.0 --cpu-step 0.5 \
    --memory-start 2.0 --memory-end 4.0 --memory-step 1.0 \
    --branch my-branch

  # Static values for non-swept groups
  python parameter_sweep_scheduling.py \
    --sweep-groups scheduling \
    --sync-cpu-request 1.0 \
    --sync-memory-request 2Gi \
    --workload-cpu-request 1.5 \
    --workload-memory-request 3Gi \
    --branch my-branch

Prerequisites:
- GitHub CLI (gh) installed and authenticated
- Access to trigger workflows in the repository
"""

import subprocess
import sys
import time
import argparse
import itertools
import os
from typing import List, Tuple


def run_sweep_test(
    cpu_value: float,
    memory_value: float,
    subset: str,
    branch: str,
    sweep_groups: List[str],
    static_resources: dict,
    scale_group: str,
    run_group_prefix: str = "sweep",
    dry_run: bool = False,
) -> bool:
    """
    Run a single test with specific CPU and memory values for selected service groups.

    Args:
        cpu_value: CPU request/limit value (e.g., 0.5, 1.0, 1.5)
        memory_value: Memory request/limit value in Gi (e.g., 0.5, 1.0, 1.5)
        subset: Test subset to run (small-configs, medium-configs, etc.)
        branch: Git branch to run from
        sweep_groups: List of service groups to sweep (e.g., ['scheduling', 'workload'])
        static_resources: Dict of static resource values for non-swept groups
        scale_group: Scale group identifier (small/medium/large) based on subset
        run_group_prefix: Prefix for the run group name (default: "sweep")
        dry_run: If True, don't actually run, just print

    Returns:
        True if successful, False otherwise
    """
    # Format values for Kubernetes
    cpu_str = str(cpu_value)
    memory_str = f"{memory_value}Gi"

    # Generate indicative run group name based on swept groups
    # Format: prefix-{groups}-cpu{value}-mem{value}
    # Example: "sweep-sched-wkld-cpu1.5-mem2.0"
    group_abbr = {
        "scheduling": "sched",
        "sync": "sync",
        "workload": "wkld",
    }
    groups_str = "-".join([group_abbr[g] for g in sorted(sweep_groups)])
    run_group = f"{run_group_prefix}-{groups_str}-cpu{cpu_value}-mem{memory_value}"

    print(f"\n{'=' * 80}")
    print(f"Testing: CPU={cpu_str}, Memory={memory_str}")
    print(f"Sweep Groups: {', '.join(sweep_groups)}")
    print(f"Run Group: {run_group}")
    print(f"{'=' * 80}")

    # Build command - use absolute path to trigger script
    script_dir = os.path.dirname(os.path.abspath(__file__))
    trigger_script = os.path.join(script_dir, "00_trigger_formula_tests.py")

    cmd = [
        "python3",
        trigger_script,
        "--subset", subset,
        "--branch", branch,
        "--custom-resources",
        "--run-group", run_group,
        "--scale-group", scale_group,
        "--delay", "3",  # Shorter delay within batch
    ]

    # Add swept group resources (dynamic)
    for group in sweep_groups:
        cmd.extend([
            f"--{group}-cpu-request", cpu_str,
            f"--{group}-cpu-limit", cpu_str,
            f"--{group}-memory-request", memory_str,
            f"--{group}-memory-limit", memory_str,
        ])

    # Add static group resources
    for key, value in static_resources.items():
        if value is not None:
            cmd.extend([f"--{key}", str(value)])

    if dry_run:
        cmd.append("--dry-run")

    print(f"Command: {' '.join(cmd)}")
    print()

    if dry_run:
        print("  [DRY RUN] Would execute command above")
        return True

    try:
        result = subprocess.run(
            cmd,
            capture_output=False,  # Let output show directly
            text=True,
            check=True,
            input="yes\n",  # Auto-confirm the prompt
        )
        print(f"✓ Successfully triggered tests for CPU={cpu_str}, Memory={memory_str}")
        return True
    except subprocess.CalledProcessError as e:
        print(f"✗ Failed to trigger tests: {e}", file=sys.stderr)
        return False


def generate_sweep_values(start: float, end: float, step: float) -> List[float]:
    """
    Generate list of values for the sweep.

    Args:
        start: Starting value (inclusive)
        end: Ending value (inclusive)
        step: Increment step

    Returns:
        List of float values
    """
    values = []
    current = start
    # Use a small epsilon for float comparison
    while current <= end + 0.001:
        values.append(round(current, 1))
        current += step
    return values


def main():
    """Main execution function."""
    parser = argparse.ArgumentParser(
        description="Run parameter sweep for service group resources",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Examples:
  # Sweep scheduling only
  python parameter_sweep_scheduling.py --sweep-groups scheduling --branch my-branch

  # Sweep scheduling and workload with custom ranges
  python parameter_sweep_scheduling.py --sweep-groups scheduling workload \\
    --cpu-start 1.0 --cpu-end 2.0 --cpu-step 0.5 --branch my-branch

  # Sweep with static values for non-swept groups
  python parameter_sweep_scheduling.py --sweep-groups scheduling \\
    --sync-cpu-request 1.0 --sync-memory-request 2Gi \\
    --workload-cpu-request 1.5 --workload-memory-request 3Gi \\
    --branch my-branch
        """
    )

    # Main configuration
    parser.add_argument(
        "--subset",
        choices=["all", "small-configs", "medium-configs", "large-configs", "edge-cases"],
        default="small-configs",
        help="Which subset of tests to run (default: small-configs)",
    )
    parser.add_argument(
        "--branch",
        default="master",
        help="Git branch to run tests from",
    )
    parser.add_argument(
        "--sweep-groups",
        nargs="+",
        choices=["scheduling", "sync", "workload"],
        default=["scheduling"],
        help="Service groups to sweep (default: scheduling). Groups not specified will use static values.",
    )

    # CPU parameters for sweep
    parser.add_argument(
        "--cpu-start",
        type=float,
        default=0.5,
        help="Starting value for CPU (default: 0.5)",
    )
    parser.add_argument(
        "--cpu-end",
        type=float,
        default=2.0,
        help="Ending value for CPU (default: 2.0)",
    )
    parser.add_argument(
        "--cpu-step",
        type=float,
        default=0.1,
        help="CPU increment step (default: 0.1)",
    )

    # Memory parameters for sweep
    parser.add_argument(
        "--memory-start",
        type=float,
        default=0.5,
        help="Starting value for Memory in Gi (default: 0.5)",
    )
    parser.add_argument(
        "--memory-end",
        type=float,
        default=2.0,
        help="Ending value for Memory in Gi (default: 2.0)",
    )
    parser.add_argument(
        "--memory-step",
        type=float,
        default=0.1,
        help="Memory increment step (default: 0.1)",
    )

    # Static resource values for scheduling services
    parser.add_argument("--scheduling-cpu-request", type=float, help="Static CPU request for scheduling services (if not swept)")
    parser.add_argument("--scheduling-cpu-limit", type=float, help="Static CPU limit for scheduling services (if not swept)")
    parser.add_argument("--scheduling-memory-request", help="Static memory request for scheduling services (if not swept)")
    parser.add_argument("--scheduling-memory-limit", help="Static memory limit for scheduling services (if not swept)")

    # Static resource values for sync services
    parser.add_argument("--sync-cpu-request", type=float, help="Static CPU request for sync services (if not swept)")
    parser.add_argument("--sync-cpu-limit", type=float, help="Static CPU limit for sync services (if not swept)")
    parser.add_argument("--sync-memory-request", help="Static memory request for sync services (if not swept)")
    parser.add_argument("--sync-memory-limit", help="Static memory limit for sync services (if not swept)")

    # Static resource values for workload services
    parser.add_argument("--workload-cpu-request", type=float, help="Static CPU request for workload services (if not swept)")
    parser.add_argument("--workload-cpu-limit", type=float, help="Static CPU limit for workload services (if not swept)")
    parser.add_argument("--workload-memory-request", help="Static memory request for workload services (if not swept)")
    parser.add_argument("--workload-memory-limit", help="Static memory limit for workload services (if not swept)")

    # Other parameters
    parser.add_argument(
        "--sleep-between",
        type=int,
        default=60,
        help="Seconds to sleep between each run (default: 60)",
    )
    parser.add_argument(
        "--run-group-prefix",
        type=str,
        default="sweep",
        help="Prefix for run group names (default: 'sweep')",
    )
    parser.add_argument(
        "--dry-run",
        action="store_true",
        help="Print what would be run without actually triggering workflows",
    )
    args = parser.parse_args()

    print("=" * 80)
    print("Multi-Service Group Resource Parameter Sweep")
    print("=" * 80)
    print()
    print(f"Configuration:")
    print(f"  Subset: {args.subset}")
    print(f"  Branch: {args.branch}")
    print(f"  Sweep Groups: {', '.join(args.sweep_groups)}")
    print(f"  CPU Range: {args.cpu_start} to {args.cpu_end} (step: {args.cpu_step})")
    print(f"  Memory Range: {args.memory_start}Gi to {args.memory_end}Gi (step: {args.memory_step})")
    print(f"  Sleep between runs: {args.sleep_between}s")
    print()

    # Build static resources dict for non-swept groups
    all_groups = {"scheduling", "sync", "workload"}
    static_groups = all_groups - set(args.sweep_groups)

    static_resources = {}
    for group in static_groups:
        for resource_type in ["cpu-request", "cpu-limit", "memory-request", "memory-limit"]:
            key = f"{group}-{resource_type}"
            value = getattr(args, key.replace("-", "_"), None)
            if value is not None:
                static_resources[key] = value

    if static_resources:
        print("Static Resources (non-swept groups):")
        for key, value in static_resources.items():
            print(f"  {key}: {value}")
        print()

    # Generate sweep values for CPU and Memory separately
    cpu_values = generate_sweep_values(args.cpu_start, args.cpu_end, args.cpu_step)
    memory_values = generate_sweep_values(args.memory_start, args.memory_end, args.memory_step)

    # Create cartesian product of CPU and Memory values
    sweep_combinations = list(itertools.product(cpu_values, memory_values))
    total_runs = len(sweep_combinations)

    # Determine scale_group from subset
    # Map subset to scale group (small/medium/large)
    subset_to_scale_group = {
        "small-configs": "small",
        "medium-configs": "medium",
        "large-configs": "large",
        "edge-cases": "edge",
        "all": "all",
    }
    scale_group = subset_to_scale_group.get(args.subset, "unknown")

    print(f"Total test configurations to run: {total_runs}")
    print(f"CPU values ({len(cpu_values)}): {cpu_values}")
    print(f"Memory values ({len(memory_values)}): {memory_values}")
    print(f"Scale Group: {scale_group}")
    print()

    if not args.dry_run:
        print("WARNING: This will trigger multiple workflow runs.")
        print(f"Each sweep point will trigger workflows for the '{args.subset}' subset.")
        print(f"Swept groups will use dynamic resources: {', '.join(args.sweep_groups)}")
        print(f"Static groups will use fixed resources: {', '.join(static_groups) if static_groups else 'none'}")
        print(f"Total time estimate: ~{total_runs * (args.sleep_between // 60)} minutes (excluding test execution)")
        print()
        response = input("Continue? (yes/no): ")
        if response.lower() not in ["yes", "y"]:
            print("Aborted.")
            sys.exit(0)
        print()

    # Run sweep
    for i, (cpu_value, memory_value) in enumerate(sweep_combinations, 1):
        print(f"\n{'#' * 80}")
        print(f"# Sweep {i}/{total_runs}: CPU={cpu_value}, Memory={memory_value}Gi")
        print(f"# Groups: {', '.join(args.sweep_groups)}")
        print(f"{'#' * 80}")

        run_sweep_test(
            cpu_value=cpu_value,
            memory_value=memory_value,
            subset=args.subset,
            branch=args.branch,
            sweep_groups=args.sweep_groups,
            static_resources=static_resources,
            scale_group=scale_group,
            run_group_prefix=args.run_group_prefix,
            dry_run=args.dry_run,
        )

        # Sleep between runs (except after the last one)
        if i < total_runs and args.sleep_between > 0 and not args.dry_run:
            print(f"\nSleeping for {args.sleep_between} seconds before next run...")
            time.sleep(args.sleep_between)

    print("\n✓ Parameter sweep complete!")
    print()
    print("Monitor progress:")
    print("  gh run list --workflow=scale-test-manual-trigger.yml --repo run-ai/runai")
    print()
    print("View results in Grafana after tests complete.")


if __name__ == "__main__":
    main()
