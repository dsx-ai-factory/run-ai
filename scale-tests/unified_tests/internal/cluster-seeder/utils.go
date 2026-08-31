/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cluster_seeder

import (
	"fmt"
	"regexp"
	"strings"
)

func extractBranchVersion(version string) (string, error) {
	if strings.HasPrefix(version, "0.0.0") {
		return "master", nil
	}

	if strings.HasPrefix(version, "v") {
		return extractReleaseBranchVersion(version)
	}

	return "dev", nil
}

func extractReleaseBranchVersion(version string) (string, error) {
	re := regexp.MustCompile(`^(\d+)\.(\d+)`)
	matches := re.FindStringSubmatch(version)

	if len(matches) >= 3 {
		return fmt.Sprintf("%s.%s", matches[1], matches[2]), nil
	}

	return "", fmt.Errorf("failed to extract branch version from %s", version)
}
