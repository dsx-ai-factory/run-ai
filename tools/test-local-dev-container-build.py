#!/usr/bin/env python3
# SPDX-FileCopyrightText: Copyright (c) 2026 Alen Peric
# SPDX-License-Identifier: Apache-2.0

"""Exercise local-dev build error handling without a container daemon."""

import re
import subprocess
import unittest
from pathlib import Path


SOURCE = (Path(__file__).resolve().parents[1] / "local-dev.sh").read_text()


def shell_function(name):
    # Extract definitions only: sourcing local-dev.sh would deploy the stack.
    match = re.search(rf"^{name}\(\) \{{\n.*?^\}}", SOURCE, re.MULTILINE | re.DOTALL)
    if match is None:
        raise AssertionError(f"Could not find {name} in local-dev.sh")
    return match.group(0)


FUNCTIONS = "\n".join(shell_function(name) for name in ("container_build", "retry_build"))
MOCKS = r"""
build_calls=0
podman() {
  case "$1" in
    build)
      build_calls=$((build_calls + 1))
      printf 'build:%s\n' "$build_calls"
      if (( build_calls <= FAIL_BUILDS )); then return 42; fi
      return 0
      ;;
    tag)
      printf 'tag\n'
      return "$TAG_STATUS"
      ;;
    *) return 99 ;;
  esac
}
docker() { printf 'docker-build\n'; return "$DOCKER_STATUS"; }
warn() { :; }
sleep() { :; }
"""


class ContainerBuildTests(unittest.TestCase):
    def run_build(self, *, failures=0, tag_status=0, retry=False, docker_status=None):
        tool = "podman" if docker_status is None else "docker"
        invocation = "retry_build container_build" if retry else "container_build"
        script = (
            "set -eo pipefail\n"
            f"CONTAINER_TOOL={tool}\nFAIL_BUILDS={failures}\n"
            f"TAG_STATUS={tag_status}\nDOCKER_STATUS={docker_status or 0}\n"
            + FUNCTIONS
            + MOCKS
            + f'\nif {invocation} runai-local/test:dev .; then status=0; '
            'else status=$?; fi\nprintf "status:%s\\n" "$status"\n'
        )
        result = subprocess.run(
            ["bash", "-c", script], capture_output=True, text=True, timeout=5
        )
        self.assertEqual(result.returncode, 0, result.stderr)
        return result.stdout.splitlines()

    def test_failed_podman_build_preserves_status_and_skips_tagging(self):
        for tag_status in (0, 1):
            with self.subTest(tag_status=tag_status):
                self.assertEqual(
                    self.run_build(failures=1, tag_status=tag_status),
                    ["build:1", "status:42"],
                )

    def test_successful_podman_build_is_tagged(self):
        self.assertEqual(self.run_build(), ["build:1", "tag", "status:0"])

    def test_optional_tag_failure_remains_nonfatal(self):
        self.assertEqual(self.run_build(tag_status=1), ["build:1", "tag", "status:0"])

    def test_retry_recovers_after_a_failed_build(self):
        self.assertEqual(
            self.run_build(failures=1, retry=True),
            ["build:1", "build:2", "tag", "status:0"],
        )

    def test_retry_exhaustion_reports_failure_without_tagging(self):
        self.assertEqual(
            self.run_build(failures=3, retry=True),
            ["build:1", "build:2", "build:3", "status:1"],
        )

    def test_successful_docker_build_is_unchanged(self):
        self.assertEqual(self.run_build(docker_status=0), ["docker-build", "status:0"])

    def test_failed_docker_build_preserves_status(self):
        self.assertEqual(self.run_build(docker_status=42), ["docker-build", "status:42"])


if __name__ == "__main__":
    unittest.main(verbosity=2)
