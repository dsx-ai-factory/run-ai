/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimcp

import (
	"fmt"

	"github.com/Masterminds/semver/v3"
)

// isDevBuildVersion reports whether v parses as a 0.0.0 semver — the
// shape every Run:ai control-plane dev/master build reports through
// /_version (e.g. "0.0.0-20002916.master-e494a34"). Treat it the same
// as an unknown runtime so local testing against master clusters isn't
// blocked by MinAPIVersion gates.
func isDevBuildVersion(v string) bool {
	parsed, err := semver.NewVersion(v)
	if err != nil {
		return false
	}
	return parsed.Major() == 0 && parsed.Minor() == 0 && parsed.Patch() == 0
}

func compareVersions(a, b string) (int, error) {
	pa, err := semver.NewVersion(a)
	if err != nil {
		return 0, fmt.Errorf("parse %q: %w", a, err)
	}
	pb, err := semver.NewVersion(b)
	if err != nil {
		return 0, fmt.Errorf("parse %q: %w", b, err)
	}
	switch {
	case pa.Major() != pb.Major():
		if pa.Major() < pb.Major() {
			return -1, nil
		}
		return 1, nil
	case pa.Minor() != pb.Minor():
		if pa.Minor() < pb.Minor() {
			return -1, nil
		}
		return 1, nil
	case pa.Patch() != pb.Patch():
		if pa.Patch() < pb.Patch() {
			return -1, nil
		}
		return 1, nil
	}
	return 0, nil
}

func versionInRange(runtime, min, max, toolName string) bool {
	if runtime == "" || isDevBuildVersion(runtime) {
		return true
	}
	if min != "" {
		cmp, err := compareVersions(min, runtime)
		if err != nil {
			panic(fmt.Sprintf("runaimcp.Register: tool %q has invalid MinAPIVersion %q: %v", toolName, min, err))
		}
		if cmp > 0 {
			return false
		}
	}
	if max != "" {
		cmp, err := compareVersions(max, runtime)
		if err != nil {
			panic(fmt.Sprintf("runaimcp.Register: tool %q has invalid MaxAPIVersion %q: %v", toolName, max, err))
		}
		if cmp < 0 {
			return false
		}
	}
	return true
}
