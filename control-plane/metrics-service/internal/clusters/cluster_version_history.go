/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package clusters

import (
	"sort"
	"strings"
	"time"

	api "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/hashicorp/go-version"
)

var (
	max_time           = time.Date(9999, 12, 31, 23, 59, 59, int(time.Second-time.Nanosecond), time.UTC)
	DEV_VERSION_PREFIX = "0.0.0"
	DevVersion, _      = version.NewSemver("999.999.999")
)

type ClusterVersionHistory struct {
	Versions       []*VersionWithTimeRange
	CurrentVersion *version.Version
}

type VersionWithTimeRange struct {
	Version *version.Version
	Start   *time.Time
	End     *time.Time
}

func NewClusterVersionHistory(histories []api.ClusterHistory) *ClusterVersionHistory {
	versions, currentVersion := createVersions(histories)
	return &ClusterVersionHistory{
		Versions:       versions,
		CurrentVersion: currentVersion,
	}
}

func createVersions(histories []api.ClusterHistory) ([]*VersionWithTimeRange, *version.Version) {
	versions := []*VersionWithTimeRange{}
	for _, h := range histories {
		v := convertToVersion(h)
		if v != nil {
			versions = append(versions, v)
		}
	}
	sortByStartTime(versions)
	fillEndTimes(versions)
	return versions, getCurrentVersion(versions)
}

func convertToVersion(h api.ClusterHistory) *VersionWithTimeRange {
	updateTime := time.Unix(h.UpdateTimestamp, 0)
	if strings.HasPrefix(h.Version, DEV_VERSION_PREFIX) {
		return &VersionWithTimeRange{
			Version: DevVersion,
			Start:   &updateTime,
		}
	}

	v, err := version.NewSemver(h.Version)
	if err != nil {
		logging.Warnf("cluster %s version in wrong format %s\n%s", h.ClusterId, h.Version, err.Error())
		return nil
	}

	return &VersionWithTimeRange{
		Version: v,
		Start:   &updateTime,
	}
}

func sortByStartTime(versions []*VersionWithTimeRange) {
	sort.Slice(versions, func(i, j int) bool {
		return versions[i].Start.Before(*versions[j].Start)
	})
}

func fillEndTimes(versions []*VersionWithTimeRange) {
	if len(versions) == 0 {
		return
	}
	for i := 0; i < len(versions)-1; i++ {
		versions[i].End = versions[i+1].Start
	}
	versions[len(versions)-1].End = &max_time
}

func getCurrentVersion(versions []*VersionWithTimeRange) *version.Version {
	if len(versions) == 0 {
		return nil
	}
	return versions[len(versions)-1].Version
}
