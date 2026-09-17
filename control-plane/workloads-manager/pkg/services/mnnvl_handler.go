/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"fmt"
	"net/http"

	clustersApi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/internal/generated/pkg/api"
	clustersClient "github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/pkg/clients/clusters"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	orgUnitApi "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client/generated"
	workloadsApi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
	"github.com/go-viper/mapstructure/v2"
	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
)

// DetermineMNNVLConfiguration determines MNNVL configuration including mode and label keys
func DetermineMNNVLConfiguration(ctx context.Context, mnnvlMode *api.MnnvlMode, project *orgUnitApi.Project, clusterId string,
	workloadType *workloadsApi.WorkloadTypeConfig, resourceInfo *ResourceInfo) (*models.WorkloadConfiguration, *api.SubmissionErrorV2) {
	if workloadType == nil || resourceInfo == nil {
		return nil, NewSubmissionError(http.StatusInternalServerError,
			"workload type or manifest resource is nil", nil)
	}
	if mnnvlMode == nil || *mnnvlMode == api.MNNVLMODE_NONE {
		return &models.WorkloadConfiguration{
			Mnnvl: &models.MnnvlConfiguration{
				Mode: api.MNNVLMODE_NONE,
			},
		}, nil
	}

	nodePoolSupport, mnnvlLabelKeys, subError := checkProjectNodePoolsSupportMNNVL(ctx, project, clusterId)
	if subError != nil {
		return nil, subError
	}

	kartaSupportResult, subError := validateKartaSupportsMNNVL(workloadType, resourceInfo)
	if subError != nil {
		return nil, subError
	}

	switch *mnnvlMode {
	case api.MNNVLMODE_REQUIRED:
		if !nodePoolSupport {
			return nil, NewConfigurationError(
				"MNNVL required but project node pools does not support GPU network acceleration")
		}
		if !kartaSupportResult.supports {
			return nil, NewConfigurationError(
				"MNNVL required but ResourceInterface does not support MNNVL: " + convutils.Deref(kartaSupportResult.reason, ""))
		}

		return &models.WorkloadConfiguration{
			Mnnvl: &models.MnnvlConfiguration{
				Mode:                            api.MNNVLMODE_REQUIRED,
				GpuNetworkAccelerationLabelKeys: mnnvlLabelKeys,
			},
		}, nil

	default:
		return nil, NewConfigurationError(
			fmt.Sprintf("Invalid MNNVL configuration value: %s. Valid values are: %v", *mnnvlMode, api.AllowedMnnvlModeEnumValues))
	}
}

// checkProjectNodePoolsSupportMNNVL checks if project's node pools support MNNVL
// project supports MNNVL if one of the node pools contains MNNVL configuration.
func checkProjectNodePoolsSupportMNNVL(ctx context.Context, project *orgUnitApi.Project, clusterId string) (bool, []string, *api.SubmissionErrorV2) {
	projectNodePools := getProjectNodePools(project)
	if len(projectNodePools) == 0 {
		return false, nil, NewSubmissionError(http.StatusInternalServerError,
			fmt.Sprintf("no node pools found for project %s", project.Name), nil)
	}

	clusterNodePools, err := fetchClusterNodePools(ctx, clusterId)
	if err != nil {
		return false, nil, NewSubmissionError(http.StatusInternalServerError,
			"failed to fetch cluster node pool", err)
	}

	type mnnvlInfo struct {
		detected bool
		labelKey string
	}
	nodePoolMnnvlMap := make(map[string]mnnvlInfo)

	for _, np := range clusterNodePools {
		detected := false
		labelKey := ""

		if np.Status != nil && np.Status.GpuNetworkAccelerationDetected != nil {
			detected = *np.Status.GpuNetworkAccelerationDetected
		}

		if np.GpuNetworkAccelerationLabelKey != nil {
			labelKey = *np.GpuNetworkAccelerationLabelKey
		}

		nodePoolMnnvlMap[np.GetName()] = mnnvlInfo{
			detected: detected,
			labelKey: labelKey,
		}
	}

	mnnvlLabelKeys := make([]string, 0)
	// Check if one of the project node pools support GPU acceleration
	for _, nodePoolName := range projectNodePools {
		npInfo, exists := nodePoolMnnvlMap[nodePoolName]
		if !exists {
			return false, nil, NewSubmissionError(http.StatusInternalServerError,
				fmt.Sprintf("node pool %s not found in cluster %s", nodePoolName, clusterId), nil)
		}

		if !npInfo.detected || npInfo.labelKey == "" {
			continue
		}

		mnnvlLabelKeys = append(mnnvlLabelKeys, npInfo.labelKey)
	}

	return len(mnnvlLabelKeys) != 0, mnnvlLabelKeys, nil
}

type kartaSupportResult struct {
	supports bool
	reason   *string
}

// validateKartaSupportsMNNVL validates if the karta (formerly ResourceInterface) supports MNNVL.
// It checks if the karta includes paths to containers, resource claims, and pod affinity, which is required for MNNVL plugin.
// Returns true if the karta supports MNNVL, false otherwise, and a reason string.
// TODO: this check should be reported as capability per workload-type from the controller in the future.
func validateKartaSupportsMNNVL(workloadType *workloadsApi.WorkloadTypeConfig, resourceInfo *ResourceInfo) (kartaSupportResult, *api.SubmissionErrorV2) {
	if len(workloadType.Kartas) == 0 {
		// This could happen for workload types that were added as part of Run:ai version that external-workload-integrator supports.
		// These type should have a "Dummy" karta but the interfaces supports empty list too.
		return kartaSupportResult{
			supports: false,
			reason:   convutils.Ref("workload type karta is empty"),
		}, nil
	}

	karta, err := decodeKartaForVersion(workloadType, resourceInfo.Version)
	if err != nil {
		return kartaSupportResult{
			supports: false,
			reason:   nil,
		}, err
	}

	if isDummyKarta(karta) {
		// Empty/Dummy ResourceInterfaces were added to allow submission of workloads supported by the external-workload-integrator.
		// TODO: They are not supported by the MNNVL plugin- they will be replaced
		return kartaSupportResult{
			supports: false,
			reason:   convutils.Ref("workload type ResourceInterface is empty"),
		}, nil
	}

	specDefExist := false
	structureDefinition := karta.Spec.StructureDefinition

	if structureDefinition.RootComponent.SpecDefinition != nil {
		specDefExist = true
		kartaSupportResult := validateSpecDefinition(structureDefinition.RootComponent.SpecDefinition)
		if !kartaSupportResult.supports {
			return kartaSupportResult, nil
		}
	}

	if structureDefinition.ChildComponents != nil {
		for _, componentDef := range structureDefinition.ChildComponents {
			if componentDef.SpecDefinition == nil {
				continue
			}
			specDefExist = true
			kartaSupportResult := validateSpecDefinition(componentDef.SpecDefinition)
			if !kartaSupportResult.supports {
				return kartaSupportResult, nil
			}
		}
	}
	if !specDefExist {
		return kartaSupportResult{
			supports: false,
			reason:   convutils.Ref("workload type ResourceInterface has no Spec definition"),
		}, nil
	}

	return kartaSupportResult{
		supports: true,
		reason:   nil,
	}, nil
}

// validateSpecDefinition validates if a Spec definition supports MNNVL
// It checks if the Spec definition has a path to containers, resource claims, and pod affinity, which is required for MNNVL plugin.
func validateSpecDefinition(specDef *kartav1alpha1.SpecDefinition) kartaSupportResult {
	if specDef == nil {
		return kartaSupportResult{
			supports: false,
			reason:   convutils.Ref("workload type ResourceInterface has empty Spec Definition"),
		}
	}

	if specDef.PodTemplateSpecPath != nil || specDef.PodSpecPath != nil {
		// PodTemplateSpec or PodSpec is used, so they have all the necessary fields
		return kartaSupportResult{
			supports: true,
			reason:   nil,
		}
	}

	fragDef := specDef.FragmentedPodSpecDefinition
	if fragDef == nil {
		return kartaSupportResult{
			supports: false,
			reason:   convutils.Ref("workload type ResourceInterface must have either PodSTemplateSpec PodSpec or FragmentedPodSpec"),
		}
	}

	if convutils.Deref(fragDef.ContainersPath, "") == "" && convutils.Deref(fragDef.ContainerPath, "") == "" {
		return kartaSupportResult{
			supports: false,
			reason:   convutils.Ref("workload type ResourceInterface must have path to containers"),
		}
	}
	if convutils.Deref(fragDef.ResourceClaimsPath, "") == "" {
		return kartaSupportResult{
			supports: false,
			reason:   convutils.Ref("workload type ResourceInterface must have path to resource claims"),
		}
	}

	if convutils.Deref(fragDef.PodAffinityPath, "") == "" {
		return kartaSupportResult{
			supports: false,
			reason:   convutils.Ref("workload type ResourceInterface must have path to pod affinity"),
		}
	}
	if convutils.Deref(fragDef.NodeAffinityPath, "") == "" {
		return kartaSupportResult{
			supports: false,
			reason:   convutils.Ref("workload type ResourceInterface must have path to node affinity"),
		}
	}

	return kartaSupportResult{
		supports: true,
		reason:   nil,
	}
}

func fetchClusterNodePools(ctx context.Context, clusterId string) ([]clustersApi.Nodepool, error) {
	nodePoolsResponse, err := clustersClient.Get().GetNodePoolsInternalByCluster(ctx, clusterId)
	if err != nil {
		return nil, err
	}

	if nodePoolsResponse == nil || len(nodePoolsResponse.GetNodepools()) == 0 {
		return nil, fmt.Errorf("no node pools found for cluster %s", clusterId)
	}

	return nodePoolsResponse.GetNodepools(), nil
}

func isDummyKarta(karta *kartav1alpha1.Karta) bool {
	return karta.Spec.StructureDefinition.RootComponent.Name == ""
}

func getProjectNodePools(project *orgUnitApi.Project) []string {
	return project.Effective.DefaultNodePools
}

func decodeKartaForVersion(workloadType *workloadsApi.WorkloadTypeConfig, version string) (*kartav1alpha1.Karta, *api.SubmissionErrorV2) {
	if workloadType == nil {
		return nil, NewSubmissionError(http.StatusInternalServerError,
			"workload type is nil", nil)
	}

	// The submission flow has already fetched workloadType. This function does
	// not make another API call; it selects and decodes an embedded Karta.
	for _, kartaInstance := range workloadType.Kartas {
		if convutils.Deref(kartaInstance.Version, "") != version {
			continue
		}
		var karta kartav1alpha1.Karta
		err := mapstructure.Decode(kartaInstance.Karta, &karta)
		if err != nil {
			return nil, NewSubmissionError(http.StatusInternalServerError,
				"failed to decode karta", err)
		}
		return &karta, nil
	}

	return nil, NewSubmissionError(http.StatusNotFound,
		fmt.Sprintf("no karta found for version %s", version), nil)
}
