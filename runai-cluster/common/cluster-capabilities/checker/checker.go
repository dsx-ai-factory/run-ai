/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package checker

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/clients/cluster"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/cache/v2"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/cache/v2/options"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"

	goVersion "github.com/hashicorp/go-version"
)

// ClusterCapabilitiesChecker this class is responsible to check if a given cluster is supporting capability X.
// Ideally, we prefer the cluster to report this,
// but since framework for this is not yet available we keep an internal map here which states
// for each cluster capability - its minimal supported version.
type ClusterCapabilitiesChecker struct {
	clusterId   string
	clusterInfo *clusterInfo
}

const (
	// MinVersionGitSyncOneParameter old cluster support both branch and revision,
	// but new cluster no longer support this and in this case the request for old clusters will be rejected.
	MinVersionGitSyncOneParameter = iota

	// MinVersionInferenceMetricsThroughputLatency cluster support for inference metrics throughput and latency
	MinVersionInferenceMetricsThroughputLatency

	MinVersionNIMMetrics

	// MinVersionVLLMMetrics is the minimum cluster version that records and exports vLLM metrics.
	MinVersionVLLMMetrics

	// MinVersionOfGpuDeviceRequest - older clusters require whole gpus to be sent as gpu, not as gpuDevices
	MinVersionOfGpuDeviceRequest

	// MinVersionOfConfigMapAsVolume is the min cluster version that support ConfigMap as volume
	MinVersionOfConfigMapAsVolume

	// MinVersionOfSecretAsVolume is the min cluster version that support Secret as volume
	MinVersionOfSecretAsVolume

	// MinVersionOfInferenceWorkloadNotification is the min cluster version required for notification about workload creation
	// from asset service to workload service for inference workloads
	MinVersionOfInferenceWorkloadNotification

	// MinVersionOfTrainingWorkloadNotification is the min cluster version required for notification about workload creation
	// from asset service to workload service for training workloads (including also distributed training)
	MinVersionOfTrainingWorkloadNotification

	// MinVersionOfDatavolume is the min cluster version for datavolumes
	MinVersionOfDatavolume

	// MinVersionOfAuthenticatedGroup is the min cluster version where authenticated group is supported
	MinVersionOfAuthenticatedGroup

	// MinVersionOfProjectQuotaStatuses is the min cluster version where project quota statuses exist
	MinVersionOfProjectQuotaStatuses

	// MinVersionOfClusterResourceStatus - older clusters do not report to k8s-objects-tracker about cluster resource status and issues
	MinVersionOfClusterReportingResourceStatus

	//MinVersionOfPVCSyncStatus - older clusters do not report to asset-service about PVCs sync status
	MinVersionOfPVCSyncStatus

	// MinVersionOfServingPort - the min cluster version that supports serving port as service type
	MinVersionOfServingPort

	// MinVersionOfServingPortAuthorizationFields - the min cluster version that supports serving port authorization fields
	MinVersionOfServingPortAuthorizationFields

	// MinVersionOfFlatWorkloadSubmission - to support flat submission , the agent needs to support fully SAAS
	MinVersionOfFlatWorkloadSubmission

	// MinVersionOfNfsHostPathForInference - the min cluster version that supports NFS and hostPath volumes for inference workloads
	MinVersionOfNfsHostPathForInference

	// MinVersionOfNfsSectionInWorkloadPayload - the min cluster version that supports NFS section in workload payload
	MinVersionOfNfsSectionInWorkloadPayload

	// MinVersionOfClusterConfigMapWideScope is the cluster minimal version, which supports Config Map cluster wide scope
	MinVersionOfClusterConfigMapWideScope

	MinVersionOfNewWorkloadOwnershipProtection

	// MinVersionOfDeprecatedSpecName is the cluster minimal version, which workload name in spec is deprecated.
	MinVersionOfDeprecatedSpecName

	// MinVersionOfWorkloadVolumesForInference - the min cluster version that supports workload volumes for inference workloads
	MinVersionOfWorkloadVolumesForInference

	// MinVersionOfAdvancedAutoscalingSettingsForInference - the min cluster version that supports advanced autoscaling fields for inference workloads
	MinVersionOfAdvancedAutoscalingSettingsForInference

	// MinVersionOfTerminationGracePeriodSeconds - the mini cluster version supporting this field
	MinVersionOfTerminationGracePeriodSeconds

	// MinVersionOfMultiFractionalGpuDevices - the min cluster version that supports multi-fractional gpu devices
	MinVersionOfMultiFractionalGpuDevices

	// MinVersionOfDistributedCleanPodPolicy - the min cluster version that supports clean pod policy for distributed workloads
	MinVersionOfDistributedCleanPodPolicy

	// MinVersionOfConfigMapEnvValue - the min cluster version supporting the CONFIGMAP:xxx,yyy value format
	MinVersionOfConfigMapEnvValue

	// MinVersionOfPodFieldRefEnvValue - the min cluster version supporting the PODFIELDREF:xxx,yyy value format
	MinVersionOfPodFieldRefEnvValue

	// MinVersionOfAssetsAgentSync - the min version of cluster supporting agent sync for assets
	MinVersionOfAssetsAgentSync

	// MinVersionUpdateInferenceSpec - the min cluster version that supports update inference spec api
	MinVersionUpdateInferenceSpec

	// MinVersionOfOnlyProjectScopeExistingPvcSupported - the min cluster version that supports only project-scoped existing pvcs
	MinVersionOfOnlyProjectScopeExistingPvcSupported

	// MinVersionOfConfigMapSubPathSupported - the min cluster version that supports sub paths of config maps
	MinVersionOfConfigMapSubPathSupported

	// MinVersionOfDistributedJaxJobsSupported - the min cluster version that supports JAX distributed workload
	MinVersionOfDistributedJaxJobsSupported

	// MinVersionTrainingPriorityClassPatchFix - the min cluster version that supports training priority class patch fix
	MinVersionTrainingPriorityClassPatchFix

	// MinVersionRevisions - the min cluster version that supports revisions sync
	MinVersionRevisions

	// MinVersionOfClusterAndDepartmentScopeDataVolumes - the min cluster version that supports cluster/department scope data volumes
	MinVersionOfClusterAndDepartmentScopeDataVolumes

	// MinVersionAccessRulesStatuses - the min cluster version that supports access rules statuses
	MinVersionAccessRulesStatuses

	MinVersionOfPVCDataSharing

	// MinVersionOfLauncherCreationPolicy - the min cluster version that supports LauncherCreationPolicy
	MinVersionOfLauncherCreationPolicy

	// MinVersionOfNodeAffinity - the min cluster version that supports Node Affinity
	MinVersionOfNodeAffinity

	// MinVersionOfInferenceServingConfiguration - the min cluster version that supports inference serving configuration
	MinVersionOfInferenceServingConfiguration

	// MinVersionOfPvcAddedAttributes - PVC added attributes support
	MinVersionOfPvcAddedAttributes

	// MinVersionOfUserCredentials - the min cluster version that supports user credentials
	MinVersionOfUserCredentials

	// MinVersionOfImagePullSecrets - the min cluster version that supports image pull secrets
	MinVersionOfImagePullSecrets

	// MinVersionOfFlexiblePriorityClasses - Flexible priority classes from control plane
	MinVersionOfFlexiblePriorityClasses

	// MinVersionOfEmptyDirAsVolume - EmptyDir as volume support
	MinVersionOfEmptyDirAsVolume

	// MinVersionCreateDistributedInference - Support distributed inference create
	MinVersionCreateDistributedInference

	// MinVersionUpdateDistributedInferenceSpec - Support distributed inference update
	MinVersionUpdateDistributedInferenceSpec

	// MinVersionOfGitSyncWithSSHKey - the min cluster version that supports git sync with SSH key
	MinVersionOfGitSyncWithSSHKey

	// MinVersionOfDefaultModeSupported - Support secret and config maps default mode
	MinVersionOfDefaultModeSupported

	// MinVersionOfClusterPolicyBypassForCPWorkloads - Cluster does not apply cluster policies on CP workloads
	MinVersionOfClusterPolicyBypassForCPWorkloads

	// MinVersionOfGenericWorkloadsSupport - support generic workloads based on GVK + manifest
	MinVersionOfGenericWorkloadsSupport

	// MinVersionOfNgcApiKeySupport - support NGC API key as credentials asset
	MinVersionOfNgcApiKeySupport
	// NOTE -> Make sure to add the min version to the map below

	//	MinVersionOfPVCEvents
	MinVersionOfPVCEvents

	//	MinVersionOfPreemptibility
	MinVersionOfPreemptibility

	MinVersionNvlinkMetrics

	// MinVersionOfServiceAccountSubjectType - support for service-account as SubjectType in access rules
	MinVersionOfServiceAccountSubjectType

	// MinVersionOfAllAuthenticatedUsersSubjectType - support for all-authenticated-users as SubjectType in access rules
	MinVersionOfAllAuthenticatedUsersSubjectType

	// MinVersionPermissionSetRoles - minimum version for agent to sync access-rules with new predefined roles
	MinVersionPermissionSetRoles

	// MinVersionOfServiceAccountInGetAccessRulesApi - min version for service-account in GetAccessRules API (v2.19 doesn't support it)
	MinVersionOfServiceAccountInGetAccessRulesApi

	// MinVersionTopologyAddress - minimum cluster version that supports nodes topology cluster
	MinVersionTopologyAddress

	// MinVersionNIMIngressViaWorkloadV2 - NIM exposure via workload v2 networking + anyworkload controller; below this version NIM CR carries Ingress spec
	MinVersionNIMIngressViaWorkloadV2

	// MinVersionOfApplicationCreation - the min cluster version that supports create and delete applications API
	MinVersionOfApplicationCreation

	// MinVersionOfApplicationPersonalCredentials - the min cluster version that supports personal credentials in AI Applications
	MinVersionOfApplicationPersonalCredentials

	// MinVersionWorkloadsStructureElements - minimum cluster version that supports workload structure elements
	MinVersionWorkloadsStructureElements

	// MinVersionOfAnyWorkloadSuspendResume - the min cluster version that supports suspend/resume for any workload types (V2 API)
	MinVersionOfAnyWorkloadSuspendResume

	// MinVersionWorkloadV2Pvcs - the min cluster version that supports metadata.configuration.pvcs via the anyworkload-controller storage plugin
	MinVersionWorkloadV2Pvcs

	// MinVersionWorkloadV2NetworkingEndpoints - the min cluster version that supports endpoint-array networking updates.
	MinVersionWorkloadV2NetworkingEndpoints

	// MinVersionWorkloadV2ServiceSpecLabels - the min cluster version whose agents and plugins preserve ServiceSpec labels.
	MinVersionWorkloadV2ServiceSpecLabels

	// MinVersionOfAnyWorkloadForV1Workloads - the min cluster version supporting any-workload submission of v1 workloads.
	MinVersionOfAnyWorkloadForV1Workloads

	// MinVersionOfSegmentTopology - the min cluster version supporting segmented distributed workloads.
	MinVersionOfSegmentTopology

	// MinVersionNoNativeSubmission - starting this version the cluster no longer support native submission
	MinVersionNoNativeSubmission

	// MinVersionGpuMemoryLimitAnnotations - the min cluster version whose scheduler (KAI) + toolkit consume the
	// per-container CUDA/KAI GPU memory-limit annotations (nvidia.com/container.<name>.gpu-memory.limit and
	// kai.scheduler/container.<name>.gpu-memory.portion.limit) instead of the legacy
	// RUNAI_GPU_MEMORY_LIMIT / RUNAI_GPU_PORTION_LIMIT container env vars.
	MinVersionGpuMemoryLimitAnnotations

	// MinVersionResourceClassTemplates - the min cluster version that their agent/cluster sync
	// supports resource class templates
	MinVersionResourceClaimTemplates

	// MinVersionNetworkingOwnershipEnforcement - the min cluster version enforcing networking Service and Pod ownership.
	MinVersionNetworkingOwnershipEnforcement
	// MinVersionOAuth2ProxyTemplates - the min cluster version that provisions the proxy error-page ConfigMap.
	MinVersionOAuth2ProxyTemplates

	// MinVersionGpuComputeMode - the first cluster release that installs GPU fractioning
	// with support for the per-container gpu-compute.mode annotation.
	MinVersionGpuComputeMode

	// NOTE -> Make sure to add the min version to the map below
)

var clusterVersionToCapabilityMap = map[int]string{
	MinVersionGitSyncOneParameter:                       "2.15.0",
	MinVersionOfGpuDeviceRequest:                        "2.16.0",
	MinVersionInferenceMetricsThroughputLatency:         "2.18.0",
	MinVersionNIMMetrics:                                "2.23.0",
	MinVersionVLLMMetrics:                               "3.0.0",
	MinVersionOfConfigMapAsVolume:                       "2.17.0",
	MinVersionOfInferenceWorkloadNotification:           "2.16.0",
	MinVersionOfTrainingWorkloadNotification:            "2.13.0",
	MinVersionOfSecretAsVolume:                          "2.18.0",
	MinVersionOfDatavolume:                              "2.18.0",
	MinVersionOfAuthenticatedGroup:                      "2.18.0",
	MinVersionOfProjectQuotaStatuses:                    "2.15.0",
	MinVersionOfClusterReportingResourceStatus:          "2.17.0",
	MinVersionOfPVCSyncStatus:                           "2.17.0",
	MinVersionOfServingPort:                             "2.17.0",
	MinVersionOfServingPortAuthorizationFields:          "2.19.0",
	MinVersionOfFlatWorkloadSubmission:                  "2.18.0",
	MinVersionOfNfsHostPathForInference:                 "2.19.0",
	MinVersionOfNfsSectionInWorkloadPayload:             "2.16.0",
	MinVersionOfClusterConfigMapWideScope:               "2.19.0",
	MinVersionOfNewWorkloadOwnershipProtection:          "2.20.0",
	MinVersionOfDeprecatedSpecName:                      "2.20.0",
	MinVersionOfWorkloadVolumesForInference:             "2.20.0",
	MinVersionOfAdvancedAutoscalingSettingsForInference: "2.20.0",
	MinVersionOfTerminationGracePeriodSeconds:           "2.20.0",
	MinVersionOfMultiFractionalGpuDevices:               "2.20.0",
	MinVersionOfDistributedCleanPodPolicy:               "2.20.0",
	MinVersionOfServiceAccountInGetAccessRulesApi:       "2.20.0",
	MinVersionOfConfigMapEnvValue:                       "2.21.0",
	MinVersionOfPodFieldRefEnvValue:                     "2.19.0",
	MinVersionOfAssetsAgentSync:                         "2.21.0",
	MinVersionUpdateInferenceSpec:                       "2.19.0",
	MinVersionOfOnlyProjectScopeExistingPvcSupported:    "2.21.0",
	MinVersionOfConfigMapSubPathSupported:               "2.21.0",
	MinVersionOfDistributedJaxJobsSupported:             "2.21.0",
	MinVersionTrainingPriorityClassPatchFix:             "2.18.0",
	MinVersionRevisions:                                 "2.21.0",
	MinVersionOfClusterAndDepartmentScopeDataVolumes:    "2.22.0",
	MinVersionAccessRulesStatuses:                       "2.22.0",
	MinVersionOfPVCDataSharing:                          "2.22.0",
	MinVersionOfLauncherCreationPolicy:                  "2.22.0",
	MinVersionOfNodeAffinity:                            "2.22.0",
	MinVersionOfInferenceServingConfiguration:           "2.22.0",
	MinVersionOfPvcAddedAttributes:                      "2.22.0",
	MinVersionOfUserCredentials:                         "2.22.0",
	MinVersionOfImagePullSecrets:                        "2.22.0",
	MinVersionOfFlexiblePriorityClasses:                 "2.22.0",
	MinVersionOfEmptyDirAsVolume:                        "2.22.0",
	MinVersionCreateDistributedInference:                "2.22.0",
	MinVersionUpdateDistributedInferenceSpec:            "2.23.0",
	MinVersionOfGitSyncWithSSHKey:                       "2.23.0",
	MinVersionOfDefaultModeSupported:                    "2.23.0",
	MinVersionOfClusterPolicyBypassForCPWorkloads:       "2.21.0",
	MinVersionOfGenericWorkloadsSupport:                 "2.23.0",
	MinVersionOfNgcApiKeySupport:                        "2.23.0",
	MinVersionOfPVCEvents:                               "2.23.0",
	MinVersionNvlinkMetrics:                             "2.23.0",
	MinVersionOfPreemptibility:                          "2.24.0",
	MinVersionOfServiceAccountSubjectType:               "2.24.0",
	MinVersionPermissionSetRoles:                        "2.24.0",
	MinVersionTopologyAddress:                           "2.25.0",
	MinVersionNIMIngressViaWorkloadV2:                   "2.25.0",
	MinVersionOfApplicationCreation:                     "2.25.0",
	MinVersionOfApplicationPersonalCredentials:          "3.1.0",
	MinVersionWorkloadsStructureElements:                "2.25.0",
	MinVersionOfAnyWorkloadSuspendResume:                "2.26.0",
	MinVersionWorkloadV2Pvcs:                            "2.26.0",
	MinVersionWorkloadV2NetworkingEndpoints:             "2.26.0",
	MinVersionWorkloadV2ServiceSpecLabels:               "3.0.0",
	MinVersionOfAnyWorkloadForV1Workloads:               "2.26.0",
	MinVersionOfAllAuthenticatedUsersSubjectType:        "3.0.0",
	MinVersionOfSegmentTopology:                         "2.27.0",
	MinVersionNoNativeSubmission:                        "3.0.0",
	MinVersionGpuMemoryLimitAnnotations:                 "3.0.0",
	MinVersionGpuComputeMode:                            "3.0.14",
	MinVersionResourceClaimTemplates:                    "3.1.0",
	MinVersionNetworkingOwnershipEnforcement:            "3.0.0",
	MinVersionOAuth2ProxyTemplates:                      "3.0.18",
}

// ClustersCacheConfig is a struct that holds the configuration for the clusters cache.
// Holds the expiration time for the cache and the cleanup frequency.
type ClustersCacheConfig struct {
	ExpirationSec int `mapstructure:"cluster_cache_expiration_sec"`
	CleanupSec    int `mapstructure:"cluster_cache_cleanup_sec"`
}

func (c *ClustersCacheConfig) SetDefaults() {
	c.ExpirationSec = 600
	c.CleanupSec = 1200
}

func MinVersionForSupporting(capability int) string {
	return clusterVersionToCapabilityMap[capability]
}

// InitClusterCapabilitiesChecker initializes the cluster capabilities checker with the cluster client connection.
// and the clusters cache configuration.
// note that if you initialize the checker with specific version, this initialization is not really needed.
func InitClusterCapabilitiesChecker(clusterUrl string, clustersCache *ClustersCacheConfig, logLevel string) {
	if clustersCache == nil {
		var err error
		clustersCache, err = config.GetR[*ClustersCacheConfig]()
		if err != nil {
			logging.Warnf("clusterCacheConfig: configuration input could not be fully interpreted; using defaults: %+v. err: %v", clustersCache, err)
		}
	}

	cluster.InitConnection(clusterUrl, logLevel)
	InitClustersCache(clustersCache)
}

// InitClustersCache initializes the clusters cache with the given configuration.
func InitClustersCache(clustersCacheConfig *ClustersCacheConfig) {
	cache.ClearStore[clusterInfo]()
	err := cache.RegisterStore[clusterInfo](options.Ttl(time.Duration(clustersCacheConfig.ExpirationSec) * time.Second))
	if err != nil {
		logging.Warnf("Failed to register cache store for type clusterInfo: %v", err)
	}
}

// NewClusterCapabilitiesChecker creates a new instance of ClusterCapabilitiesChecker
// with clusterId. cluster version will be loaded from the cluster service when needed.
// This is one option to initialize the checker, another option is to initialize it with
// cluster version (see below)
func NewClusterCapabilitiesChecker(clusterId string) *ClusterCapabilitiesChecker {
	return &ClusterCapabilitiesChecker{clusterId: clusterId}
}

func NewClusterCapabilitiesCheckerForVersion(clusterVersion string) (*ClusterCapabilitiesChecker, error) {
	clusterSemverVersion, err := goVersion.NewSemver(normalizeVersion(clusterVersion))
	if err != nil {
		return nil, fmt.Errorf("failed to parse cluster version: %s, err: %w", clusterVersion, err)
	}
	return &ClusterCapabilitiesChecker{clusterInfo: &clusterInfo{Version: clusterSemverVersion}}, nil
}

// ValidateClusterCapability checks if the cluster supports the given capability.
func (c *ClusterCapabilitiesChecker) ValidateClusterCapability(ctx context.Context, capabilityToCheck int) (bool, error) {
	clusterInfo, err := c.getClusterInfo(ctx)
	if err != nil {
		return false, err
	}

	capabilityMinVersion, ok := clusterVersionToCapabilityMap[capabilityToCheck]
	if !ok {
		return false, errors.New(fmt.Sprintf("capability %d is not supported by the checker", capabilityToCheck))
	}

	capabilityRequiredVersion, err := goVersion.NewVersion(capabilityMinVersion)
	if err != nil {
		return false, errors.New(fmt.Sprintf("failed to parse capability version: %s, err: %v", capabilityMinVersion, err))
	}

	if clusterInfo.Version.LessThan(capabilityRequiredVersion) {
		return false, nil
	}

	return true, nil
}

// ClusterConfiguration returns the cluster configuration
func (c *ClusterCapabilitiesChecker) ClusterConfiguration(ctx context.Context) (*ClusterConfiguration, error) {
	clusterInfo, err := c.getClusterInfo(ctx)
	if err != nil {
		return nil, err
	}
	// If ClusterCapabilitiesChecker was init without fetching the cluster configuration(NewClusterCapabilitiesCheckerForVersion),
	// force fetch cluster info
	if clusterInfo.Configuration == nil {
		clusterInfo, err = c.fetchClusterInfo(ctx)
		if err != nil {
			return nil, err
		}
	}
	return clusterInfo.Configuration, nil
}

func (c *ClusterCapabilitiesChecker) getClusterInfo(ctx context.Context) (*clusterInfo, error) {
	//   if cluster version is not known to the capabilities checker, it will need to get it from the
	//   cluster service
	if c.clusterInfo != nil {
		return c.clusterInfo, nil
	}
	return c.fetchClusterInfo(ctx)
}

func (c *ClusterCapabilitiesChecker) fetchClusterInfo(ctx context.Context) (*clusterInfo, error) {
	cachedInfo, clusterFound := getCacheClusterInfo(c.clusterId)
	if clusterFound {
		return cachedInfo, nil
	}

	clusterDetails, apiErr := cluster.Get().Client().GetClusterInternal(ctx, c.clusterId)
	if apiErr != nil || clusterDetails == nil {
		return nil, errors.New(fmt.Sprintf("failed to check cluster capabilities, err: %v", apiErr))
	}
	clusterInfo, err := setCacheClusterInfo(c.clusterId, clusterDetails)
	if err != nil {
		return nil, err
	}

	return clusterInfo, nil
}

// FetchClusterConfigurationSyncAndStore fetches cluster configuration directly and caches the result
func FetchClusterConfigurationSyncAndStore(ctx context.Context, clusterId string) (*ClusterConfiguration, error) {
	clusterDetails, apiErr := cluster.Get().Client().GetClusterByUuid(ctx, clusterId)
	if apiErr != nil {
		return nil, fmt.Errorf("failed to check cluster capabilities, err: %v", apiErr)
	}

	// Cache the fresh cluster info
	clusterInfo, err := setCacheClusterInfo(clusterId, clusterDetails)
	if err != nil {
		return nil, fmt.Errorf("failed to cache cluster info for cluster %s: %w", clusterId, err)
	}

	return clusterInfo.Configuration, nil
}

func normalizeVersion(version string) string {
	if strings.HasPrefix(version, "0.0.0") || strings.Contains(version, "master") {
		// That indicates that this is a Dev version, therefore we consider it as Max value
		return "9.9.9"
	}

	index := strings.Index(version, "-")
	if index != -1 {
		// Extract substring before the dash
		return version[:index]
	}

	return version
}
