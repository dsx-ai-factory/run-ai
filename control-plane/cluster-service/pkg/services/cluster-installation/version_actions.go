/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cluster_installation

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/versions"

	"github.com/Masterminds/semver/v3"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
)

var (
	absoluteMaxVersion, _             = semver.NewVersion("999.999.999")
	absoluteMinVersion, _             = semver.NewVersion("0.0.0")
	devVersionStringDisplayIdentifier = "999.999.dev"
)

func adjustMinMaxVersion(ctx context.Context, cfgMinVersion, cfgMaxVersion, cfgCurVersion string) (minVer, maxVer *semver.Version) {
	minVer, err := semver.NewVersion(cfgMinVersion)
	if err != nil {
		minVer = absoluteMinVersion
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Cluster installation Min Version is not valid: %s. Allowing versions from %s.", err.Error(), minVer.Original())
	}

	maxVer, err = semver.NewVersion(cfgMaxVersion)
	if err != nil {
		maxVer = absoluteMaxVersion
		logging.Warnf("Cluster installation Max Version is not valid: %s. Allowing versions up to %s.", err.Error(), maxVer.Original())
	}

	curVer, err := semver.NewVersion(cfgCurVersion)
	if err == nil {
		// if current version is lower than the max allowed version, limit the max to the current version
		if curVer.LessThan(maxVer) {
			maxVer = curVer
			logging.WithCtx(runaictx.LoggerHook(ctx)).Infof(
				"Cluster installation Max Version (%s) is higher than current version (%s). Using current version as max allowed version.",
				cfgMaxVersion, cfgCurVersion,
			)
		}
	}

	if maxVer.LessThan(minVer) {
		minVer = absoluteMinVersion
		maxVer = absoluteMaxVersion
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf(
			"Cluster installation Min Version (%s) and Max Version (%s) must form a valid semver range, but got min: %s, max: %s. Allowing all versions.",
			cfgMinVersion, cfgMaxVersion, minVer.Original(), maxVer.Original(),
		)
	}

	return minVer, maxVer
}

func getLatestMinorVersions(versionsList VersionsList, minVer *semver.Version, maxVer *semver.Version) map[string]string {
	ret := map[string]string{}
	for _, verStr := range versionsList {
		ver, err := semver.NewVersion(verStr.Version)
		if err != nil {
			continue
		}
		curVersionMinorStr := fmt.Sprintf("%d.%d", ver.Major(), ver.Minor())
		curVerMinor, err := semver.NewVersion(curVersionMinorStr)
		if err != nil {
			continue
		}
		if minVer.GreaterThan(ver) {
			continue
		}
		// if maxVer is a patch version we need to compare the actual version numbers
		if len(strings.Split(maxVer.Original(), ".")) > 2 {
			if maxVer.LessThan(ver) {
				continue
			}
		} else { // otherwise compare only the minor version
			if maxVer.LessThan(curVerMinor) {
				continue
			}
		}

		verInMapStr, ok := ret[curVersionMinorStr]
		if ok {
			inMapVer, _ := semver.NewVersion(verInMapStr)
			if inMapVer.LessThan(ver) {
				ret[curVersionMinorStr] = verStr.Version
			}
		} else {
			ret[curVersionMinorStr] = verStr.Version
		}
	}
	return ret
}

func excludeVersions(versions map[string]string, excludedVers []string) map[string]string {
	re := regexp.MustCompile(`[\s\-""]+`)
	if excludedVers == nil {
		return versions
	}
	for _, key := range excludedVers {
		strippedKey := re.ReplaceAllString(key, "")
		delete(versions, strippedKey)
	}
	return versions
}

func getCurrentVersionAsVersionList(curVersion string, minVersion, maxVersion *semver.Version) map[string]string {
	if strings.HasPrefix(curVersion, versions.DEV_VERSION_PREFIX) {
		return map[string]string{devVersionStringDisplayIdentifier: curVersion}
	}
	defaultVersionsList := VersionsList{{curVersion}}
	return getLatestMinorVersions(defaultVersionsList, minVersion, maxVersion)
}
