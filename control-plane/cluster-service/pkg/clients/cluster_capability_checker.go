/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package clients

import (
	"context"
	"fmt"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/utils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/versions"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/checker"
	"github.com/google/uuid"
	"github.com/maypok86/otter/v2"
)

const (
	LocalMinVersionOfSwapNodepool = 10000 + iota
	LocalMinVersionOfMinGuaranteedRuntime
	LocalMinVersionOfTimeBasedFairShare
	LocalMinVersionOfGPUNetworkAccelerationDetection
	LocalMinVersionOfGroupByWorkloadCategory
	LocalMinVersionOfGpuDevicePlacementStrategy
	LocalMinVersionOfNumaAwareScheduling
)

var localCapabilityVersionMap = map[int]string{
	LocalMinVersionOfSwapNodepool:                    "2.22.0",
	LocalMinVersionOfMinGuaranteedRuntime:            "2.23.0",
	LocalMinVersionOfTimeBasedFairShare:              "2.24.0",
	LocalMinVersionOfGPUNetworkAccelerationDetection: "2.21.0",
	LocalMinVersionOfGroupByWorkloadCategory:         "2.22.0",
	LocalMinVersionOfGpuDevicePlacementStrategy:      "2.26.0",
	LocalMinVersionOfNumaAwareScheduling:             "2.26.0",
}

// LocalMinVersionForSupporting returns the minimum version string for a given
// local capability constant, for use in error messages.
func LocalMinVersionForSupporting(capability int) string {
	return localCapabilityVersionMap[capability]
}

// ClusterCapabilityChecker interface for checking cluster capabilities
type ClusterCapabilityChecker interface {
	ValidateClusterCapability(ctx context.Context, capabilityToCheck int) (bool, error)
}

// ClusterCapabilityCheckerFactory creates a ClusterCapabilityChecker for a given cluster
type ClusterCapabilityCheckerFactory func(clusterUuid string) ClusterCapabilityChecker

// DefaultClusterCapabilityCheckerFactory returns the real checker
func DefaultClusterCapabilityCheckerFactory(clusterUuid string) ClusterCapabilityChecker {
	return checker.NewClusterCapabilitiesChecker(clusterUuid)
}

const clusterVersionCacheTTL = 10 * time.Minute

var clusterVersionCache = otter.Must(&otter.Options[string, string]{
	MaximumSize:      1000,
	ExpiryCalculator: otter.ExpiryWriting[string, string](clusterVersionCacheTTL),
})

// localClusterCapabilityChecker checks cluster capabilities by reading the
// cluster version directly from the database, avoiding the HTTP round-trip
// that the default checker uses. Cluster versions are cached in memory to
// avoid hitting the database on every capability check.
type localClusterCapabilityChecker struct {
	clusterUuid string
}

func (c *localClusterCapabilityChecker) ValidateClusterCapability(ctx context.Context, capabilityToCheck int) (bool, error) {
	tenantId, err := utils.ParseTenantId(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to parse tenant id: %w", err)
	}

	clusterID, err := uuid.Parse(c.clusterUuid)
	if err != nil {
		return false, fmt.Errorf("failed to parse cluster uuid: %w", err)
	}

	version, err := c.getClusterVersionCached(ctx, tenantId, clusterID)
	if err != nil {
		return false, fmt.Errorf("failed to get cluster version: %w", err)
	}

	if version == "" {
		return false, nil
	}

	// Check local capability map first
	if minVersion, ok := localCapabilityVersionMap[capabilityToCheck]; ok {
		return versions.IsVersionSupported(version, minVersion)
	}

	// Otherwise, delegate to the vendored checker
	versionChecker, err := checker.NewClusterCapabilitiesCheckerForVersion(version)
	if err != nil {
		return false, fmt.Errorf("failed to create version checker: %w", err)
	}

	return versionChecker.ValidateClusterCapability(ctx, capabilityToCheck)
}

func (c *localClusterCapabilityChecker) getClusterVersionCached(ctx context.Context, tenantId int32, clusterID uuid.UUID) (string, error) {
	cacheKey := fmt.Sprintf("%d_%s", tenantId, clusterID.String())

	if version, ok := clusterVersionCache.GetIfPresent(cacheKey); ok {
		return version, nil
	}

	version, err := dao.GetClusterVersion(ctx, tenantId, clusterID)
	if err != nil {
		return "", err
	}

	if version != "" {
		clusterVersionCache.Set(cacheKey, version)
	}

	return version, nil
}

// LocalClusterCapabilityCheckerFactory returns a checker that reads the cluster
// version from the local database instead of making an HTTP call.
func LocalClusterCapabilityCheckerFactory(clusterUuid string) ClusterCapabilityChecker {
	return &localClusterCapabilityChecker{clusterUuid: clusterUuid}
}
