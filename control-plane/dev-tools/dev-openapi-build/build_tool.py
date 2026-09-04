#!/usr/bin/env python3
import argparse
import subprocess
import os
import sys

# Determine the absolute path to the script's directory
SCRIPT_DIR = os.path.dirname(os.path.abspath(__file__))

SERVICE_CONFIG = {
    "workloads": {
        "path_from_script_dir": os.path.normpath(os.path.join(SCRIPT_DIR, "../../workloads")),
        "name": "Workloads"
    },
    "policy-service": {
        "path_from_script_dir": os.path.normpath(os.path.join(SCRIPT_DIR, "../../policy-service")),
        "name": "Policy Service"
    },
    "assets-service": {
        "path_from_script_dir": os.path.normpath(os.path.join(SCRIPT_DIR, "../../assets-service")),
        "name": "Assets Service"
    },
    "workloads-manager": {
        "path_from_script_dir": os.path.normpath(os.path.join(SCRIPT_DIR, "../../workloads-manager")),
        "name": "Workloads Manager"
    },
}

STEP_DEFINITIONS = {
    "mod-tidy-vendor": {
        "description": "Run 'go mod tidy && go mod vendor'",
        "command": "go mod tidy && go mod vendor",
    },
    "openapi-build": {
        "description": "Run 'make openapi-build'",
        "command": "make openapi-build",
    },
    "code-build": {
        "description": "Run 'rm -fr vendor && go mod tidy && go build ./... && go test ./...'",
        "command": "rm -fr vendor && go mod tidy && go build ./... && go test ./...",
    },
}

DEFAULT_STEPS = ["openapi-build", "code-build"]

def run_command(command, working_dir_abs, service_name, step_name):
    """Runs a shell command in a specified absolute working directory."""
    print(f"\n[{service_name} :: {step_name}] Attempting to run in '{working_dir_abs}': {command}")

    if not os.path.isdir(working_dir_abs):
        print(f"ERROR: Working directory '{working_dir_abs}' does not exist or is not a directory.")
        return False, ""

    try:
        # shell=True is used because commands can contain '&&', etc.
        # Ensure commands are from a trusted source.
        process = subprocess.Popen(
            command,
            cwd=working_dir_abs,
            shell=True,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            text=True,
            bufsize=1
        )

        # Print stdout and stderr line by line as they arrive
        from threading import Thread
        import sys

        # Store stderr output for error detection
        stderr_output = []

        def stream_output(stream, prefix, store_output=None):
            for line in iter(stream.readline, ''):
                print(line, flush=True)
                if store_output is not None:
                    store_output.append(line)
            stream.close()

        threads = []
        threads.append(Thread(target=stream_output, args=(process.stdout, 'STDOUT')))
        threads.append(Thread(target=stream_output, args=(process.stderr, 'STDERR', stderr_output)))
        for t in threads:
            t.start()
        for t in threads:
            t.join()

        returncode = process.wait()
        stderr_text = ''.join(stderr_output)

        if returncode != 0:
            print(f"ERROR: Command failed with exit code {returncode} for [{service_name} :: {step_name}]")
            return False, stderr_text
        print(f"[{service_name} :: {step_name}] Successfully completed.")
        return True, stderr_text
    except FileNotFoundError:
        print(f"ERROR: Working directory '{working_dir_abs}' not found when trying to execute command for [{service_name} :: {step_name}].")
        return False, ""
    except Exception as e:
        print(f"EXCEPTION during command execution for [{service_name} :: {step_name}]: {e}")
        return False, ""

def needs_go_mod_vendor(stderr_output):
    """Check if the error output suggests we need to run go mod vendor or go mod tidy."""
    vendor_indicators = [
        "go mod vendor",
        "go mod tidy",
        "vendor/",
        "missing sum",
        "no required module provides package",
        "build constraints exclude all Go files",
        "cannot find module providing package"
    ]

    stderr_lower = stderr_output.lower()
    return any(indicator in stderr_lower for indicator in vendor_indicators)

def run_openapi_build_with_retry(working_dir_abs, service_name):
    """Run openapi-build with automatic retry using mod-tidy-vendor if dependency issues are detected."""

    # First attempt at openapi-build
    success, stderr_output = run_command(
        STEP_DEFINITIONS["openapi-build"]["command"],
        working_dir_abs,
        service_name,
        "openapi-build"
    )

    if success:
        return True

    # Check if the failure is due to missing dependencies
    if needs_go_mod_vendor(stderr_output):
        print(f"\n[{service_name} :: openapi-build] Detected dependency issues. Running mod-tidy-vendor first...")

        # Run mod-tidy-vendor
        vendor_success, _ = run_command(
            STEP_DEFINITIONS["mod-tidy-vendor"]["command"],
            working_dir_abs,
            service_name,
            "mod-tidy-vendor"
        )

        if not vendor_success:
            print(f"[{service_name} :: openapi-build] mod-tidy-vendor failed. Cannot retry openapi-build.")
            return False

        # Retry openapi-build
        print(f"\n[{service_name} :: openapi-build] Retrying openapi-build after running mod-tidy-vendor...")
        retry_success, _ = run_command(
            STEP_DEFINITIONS["openapi-build"]["command"],
            working_dir_abs,
            service_name,
            "openapi-build (retry)"
        )

        if retry_success:
            print(f"[{service_name} :: openapi-build] Successfully completed after retry with mod-tidy-vendor.")
        else:
            print(f"[{service_name} :: openapi-build] Failed even after running mod-tidy-vendor.")

        return retry_success
    else:
        # Different type of error, don't retry
        print(f"[{service_name} :: openapi-build] Failed with non-dependency error. Not retrying.")
        return False

def main():
    parser = argparse.ArgumentParser(
        description="Build tool for specified services and steps. Paths to services are relative to the script's location.",
        formatter_class=argparse.RawTextHelpFormatter
    )
    parser.add_argument(
        "--services",
        nargs="+",
        choices=list(SERVICE_CONFIG.keys()),
        default=list(SERVICE_CONFIG.keys()),
        help=f"Services to target. Choices: {', '.join(SERVICE_CONFIG.keys())}.\nDefault: all services ({', '.join(list(SERVICE_CONFIG.keys()))})."
    )

    steps_help_text = f"Specific steps to run in order. Choices: {', '.join(STEP_DEFINITIONS.keys())}.\n"
    steps_help_text += f"If this option is not provided, default steps are run: {', '.join(DEFAULT_STEPS)}.\n"
    steps_help_text += "If --steps is provided with no arguments, no steps will be run."

    parser.add_argument(
        "--steps",
        nargs="*",
        choices=list(STEP_DEFINITIONS.keys()),
        default=None, # Will be handled to mean DEFAULT_STEPS if not provided
        help=steps_help_text
    )

    args = parser.parse_args()

    selected_services = args.services

    steps_to_run = args.steps
    if steps_to_run is None:  # --steps flag was not used
        steps_to_run = DEFAULT_STEPS
    elif not steps_to_run:  # --steps flag was used but no specific steps listed
        print("No steps specified (e.g., --steps was used with no arguments). Nothing to do.")
        parser.print_help()
        return

    if not steps_to_run: # Still no steps (e.g. DEFAULT_STEPS was empty, though it's not here)
        print("No steps defined to run. Exiting.")
        return

    service_names_display = [SERVICE_CONFIG[s]["name"] for s in selected_services]
    print(f"Target services: {', '.join(service_names_display)}")
    print(f"Steps to run: {', '.join(steps_to_run)}")
    print("-" * 30)

    overall_success = True

    for step_name in steps_to_run:
        if step_name not in STEP_DEFINITIONS:
            print(f"WARNING: Step '{step_name}' is not defined. Skipping.")
            continue

        step_info = STEP_DEFINITIONS[step_name]
        print(f"\nExecuting Step: {step_name} ({step_info['description']})")
        print("=" * (18 + len(step_name) + len(step_info['description'])))

        for service_key in selected_services:
            if service_key not in SERVICE_CONFIG:
                print(f"WARNING: Service '{service_key}' is not defined. Skipping for step '{step_name}'.")
                continue

            service_info = SERVICE_CONFIG[service_key]
            service_display_name = service_info["name"]
            absolute_service_path = service_info["path_from_script_dir"]
            command_to_run = step_info["command"]

            # Special handling for openapi-build with retry logic
            if step_name == "openapi-build":
                step_success = run_openapi_build_with_retry(absolute_service_path, service_display_name)
            else:
                step_success, _ = run_command(command_to_run, absolute_service_path, service_display_name, step_name)

            if not step_success:
                overall_success = False
                print(f"!!! Step '{step_name}' failed for service '{service_display_name}'. !!!")
                # Continue with other services/steps, but report overall failure at the end.

    print("-" * 30)
    if overall_success:
        print("All selected steps completed successfully for all selected services.")
    else:
        print("One or more steps failed. Please check the logs above.")
        sys.exit(1)

if __name__ == "__main__":
    main() 