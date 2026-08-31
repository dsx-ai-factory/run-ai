/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package checker

import (
	"context"
	"errors"
	"fmt"

	goVersion "github.com/hashicorp/go-version"
	clusterService "github.com/run-ai/runai/control-plane/cluster-service/pkg/client/generated"
	"github.com/run-ai/runai/runai-application-packages/convutils"
	workloadsService "github.com/run-ai/runai/runai-clients/workloads-client/generated"
	"github.com/run-ai/runai/runai-common-packages/cache/v2"
	"github.com/run-ai/runai/runai-common-packages/logging"
)

var ErrClusterVersionEmpty = errors.New("cluster version is empty")

const (
	KnativeDependency     = "knative"
	ScaleToZeroDependency = "enable-scale-to-zero"
	NodeTypeDependency    = "kubernetes.podspec-nodeselector"
	LWSDependency         = "lws"
)

// clusterInfo holds the cluster information gathered from the cluster service
type clusterInfo struct {
	Version       *goVersion.Version
	Configuration *ClusterConfiguration
}

// getCacheClusterInfo retrieves cluster info from the cache
func getCacheClusterInfo(clusterId string) (*clusterInfo, bool) {
	cachedInfo, err := cache.Get[clusterInfo](context.Background(), clusterId)
	if err != nil {
		logging.Warnf("Failed to get cluster info from cache for cluster %s: %v", clusterId, err)
	}
	if cachedInfo == nil {
		return nil, false
	}

	return cachedInfo, true
}

// setCacheClusterInfo stores cluster info in the cache
func setCacheClusterInfo(clusterId string, displayedCluster *clusterService.DisplayedCluster) (*clusterInfo, error) {
	version := normalizeVersion(convutils.Deref(displayedCluster.Version.Get(), ""))

	if version == "" {
		return nil, fmt.Errorf("%w for cluster %s", ErrClusterVersionEmpty, clusterId)
	}

	clusterSemverVersion, err := goVersion.NewSemver(version)
	if err != nil {
		return nil, fmt.Errorf("malformed cluster version: %s, err: %w", version, err)
	}

	knativeAvailable, knativeScaleToZeroAvailable, knativeNodeTypeAvailable := isKnativeAvailable(displayedCluster)
	lwsAvailable := isLWSAvailable(displayedCluster)
	info := &clusterInfo{
		Version: clusterSemverVersion,
		Configuration: &ClusterConfiguration{
			WorkloadOwnershipProtection: workloadOwnershipProtection(displayedCluster),
			KnativeAvailable:            knativeAvailable,
			KnativeScaleToZeroAvailable: knativeScaleToZeroAvailable,
			KnativeNodeTypeAvailable:    knativeNodeTypeAvailable,
			DistributedFrameworksStatus: constructDistributedFrameworksStatus(displayedCluster),
			SubdomainRouting:            subdomainRouting(displayedCluster),
			LWSAvailable:                lwsAvailable,
			ClusterDomain:               displayedCluster.GetDomain(),
		},
	}
	if err := cache.Set[clusterInfo](context.Background(), clusterId, *info); err != nil {
		logging.Warnf("Failed to set cluster info in cache for cluster %s: %v", clusterId, err)
	}
	return info, nil
}

// workloadOwnershipProtection extract the value of the workloadOwnershipProtection from the displayedCluster in safe manner
func workloadOwnershipProtection(displayedCluster *clusterService.DisplayedCluster) bool {
	status := displayedCluster.Status.Get()
	if status == nil {
		return false
	}
	config := status.Config.Get()
	if config == nil {
		return false
	}
	protection := config.WorkloadOwnershipProtection.Get()
	if protection == nil {
		return false
	}
	return *protection
}

// subdomainRouting extract the value of the subdomain enable value from the displayedCluster in safe manner
func subdomainRouting(displayedCluster *clusterService.DisplayedCluster) bool {
	status := displayedCluster.Status.Get()
	if status == nil {
		return false
	}
	config := status.Config.Get()
	if config == nil {
		return false
	}
	protection := config.SubdomainEnabled.Get()
	if protection == nil {
		return false
	}
	return *protection
}

// isKnativeAvailable extract the value of the isKnativeAvailable, knativeScaleToZeroAvailable and knativeNodeTypeAvailable from the displayedCluster in safe manner
func isKnativeAvailable(displayedCluster *clusterService.DisplayedCluster) (bool, bool, bool) {
	status := displayedCluster.Status.Get()
	if status == nil {
		return false, false, false
	}
	optionalDeps := status.GetDependencies().Optional
	if optionalDeps == nil {
		return false, false, false
	}
	knativeStatus, ok := optionalDeps[KnativeDependency]
	if !ok {
		return false, false, false
	}

	if !knativeStatus.Available {
		return false, false, false
	}

	knativeComponents := knativeStatus.Components
	if knativeComponents == nil {
		return true, false, false
	}

	knativeSubStatus, ok := (*knativeComponents)[KnativeDependency]
	if !ok {
		return true, false, false
	}

	knativeSubComponents := knativeSubStatus.Components
	if knativeSubComponents == nil {
		// if no subcomponents are available (e.g. for older clusters), we assume that they are all supported
		return true, true, true
	}

	scaleToZero, ok := (*knativeSubComponents)[ScaleToZeroDependency]
	var scaleToZeroAvailable bool
	if !ok {
		// if the scale-to-zero component is not available (e.g. for older clusters), we assume that the cluster is scale-to-zero capable
		scaleToZeroAvailable = true
	} else {
		scaleToZeroAvailable = scaleToZero.Available
	}

	nodeType, ok := (*knativeSubComponents)[NodeTypeDependency]
	var nodeTypeAvailable bool
	if !ok {
		// if the nodeType component is not available (e.g. for older clusters), we assume that the cluster is nodeType capable
		nodeTypeAvailable = true
	} else {
		nodeTypeAvailable = nodeType.Available
	}

	return true, scaleToZeroAvailable, nodeTypeAvailable
}

// isLWSAvailable extract the value of the LWS availability from the displayedCluster in safe manner
func isLWSAvailable(displayedCluster *clusterService.DisplayedCluster) bool {
	status := displayedCluster.Status.Get()
	if status == nil {
		return false
	}
	optionalDeps := status.GetDependencies().Optional
	if optionalDeps == nil {
		return false
	}
	lwsStatus, ok := optionalDeps[LWSDependency]
	if !ok {
		return false
	}

	return lwsStatus.Available
}

// constructDistributedFrameworksStatus builds the distributedFrameworksStatus map from the displayedCluster in safe manner
func constructDistributedFrameworksStatus(displayedCluster *clusterService.DisplayedCluster) map[workloadsService.DistributedFramework]DistributedFrameworkStatus {
	frameworksStatus := make(map[workloadsService.DistributedFramework]DistributedFrameworkStatus)

	clusterStatus := displayedCluster.Status.Get()
	if clusterStatus == nil {
		return frameworksStatus
	}
	optionalDeps := clusterStatus.GetDependencies().Optional
	if optionalDeps == nil {
		return frameworksStatus
	}

	for _, framework := range DistributedFrameworksDependencies {
		frameworkStatus := DistributedFrameworkStatus{
			Available:   false,
			Suspendable: false,
		}

		depStatus, hasFramework := optionalDeps[framework]
		if hasFramework {
			frameworkStatus.Available = depStatus.Available

			if depStatus.Components != nil {
				if suspendStatus, hasSuspend := (*depStatus.Components)[DistributedSuspendComponent]; hasSuspend {
					frameworkStatus.Suspendable = suspendStatus.Available
				}
			}
		}

		key, err := getDistributedFrameworkKey(framework)
		if err != nil {
			logging.Errorf("%s", err.Error())
			continue
		}
		frameworksStatus[key] = frameworkStatus
	}

	return frameworksStatus
}
