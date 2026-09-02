/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package versions

import (
	"fmt"
	"strings"

	"github.com/hashicorp/go-version"
)

const DEV_VERSION_PREFIX = "0.0.0"

// IsVersionSupported checks if the given version is supported by the given minSupportedVersion;
// expected format of versions - "X.Y" or "X.Y.Z" or "0.0.0-XXXX" (for master versions, for example)
func IsVersionSupported(version string, minSupportedVersion string) (bool, error) {
	if strings.HasPrefix(version, DEV_VERSION_PREFIX) {
		return true, nil
	}

	semverVersion, err := stringToSemver(version)
	if err != nil {
		return false, err
	}

	minSupportedSemverVersion, err := stringToSemver(minSupportedVersion)
	if err != nil {
		return false, err
	}

	return !semverVersion.LessThan(minSupportedSemverVersion), nil
}

func AreAllVersionsSupported(versions []string, minSupportedVersion string) (bool, error) {
	minSupportedSemverVersion, err := stringToSemver(minSupportedVersion)
	if err != nil {
		return false, err
	}

	for _, v := range versions {
		supported, innerErr := isVersionSupported(v, minSupportedSemverVersion)
		if innerErr != nil {
			return false, innerErr
		}
		if !supported {
			return false, nil
		}
	}
	return true, nil
}

func isVersionSupported(version string, minSupportedSemverVersion *version.Version) (bool, error) {
	if strings.HasPrefix(version, DEV_VERSION_PREFIX) {
		return true, nil
	}

	semverVersion, err := stringToSemver(version)
	if err != nil {
		return false, err
	}

	return !semverVersion.LessThan(minSupportedSemverVersion), nil
}

func stringToSemver(v string) (*version.Version, error) {
	semverVersion, err := version.NewSemver(v)
	if err != nil {
		return nil, fmt.Errorf("failed to parse to semversion: %s with err: %v", v, err)
	}
	return semverVersion, nil
}
