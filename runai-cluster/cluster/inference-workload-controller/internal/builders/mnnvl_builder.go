/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package builders

import (
	"context"
	"strconv"
	"strings"

	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	mnnvltypes "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/mnnvl"
	mnnvlbuilder "github.com/run-ai/runai/runai-cluster/cluster/workload-common/builders/mnnvl"
	nodepoolbuilders "github.com/run-ai/runai/runai-cluster/cluster/workload-common/builders/nodepool"
	gpu_validator "github.com/run-ai/runai/runai-cluster/cluster/workload-common/gpu-validator"
	projectutils "github.com/run-ai/runai/runai-cluster/cluster/workload-common/project"
	dinfw "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedinferenceworkload/v1alpha1"
	commonconstants "github.com/run-ai/runai/runai-cluster/common/constants"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type MnnvlBuilder struct {
	k8sClient      client.Client
	elasticEnabled bool
}

func NewMnnvlBuilder(k8sClient client.Client, elasticEnabled bool) *MnnvlBuilder {
	return &MnnvlBuilder{
		k8sClient:      k8sClient,
		elasticEnabled: elasticEnabled,
	}
}

func (b *MnnvlBuilder) GetPreview(workload *dinfw.DistributedInferenceWorkload) ([]client.Object, error) {
	if !b.IsMnnvlRequired(workload) {
		return nil, nil
	}

	resource := mnnvlbuilder.PreviewComputeDomain(DistributedInferenceWorkloadInitial, workload.Name, workload.Namespace)
	if resource == nil {
		return nil, nil
	}

	return []client.Object{resource}, nil
}

func (b *MnnvlBuilder) Build(ctx context.Context, workload *dinfw.DistributedInferenceWorkload) ([]client.Object, error) {
	if !b.IsMnnvlRequired(workload) {
		return nil, nil
	}

	// Prepare ComputeDomainWorkloadParams
	params := mnnvlbuilder.ComputeDomainWorkloadParams{
		WorkloadKind:      workload.Kind,
		JobType:           constants.LwsKind,
		WorkloadInitials:  DistributedInferenceWorkloadInitial,
		WorkloadName:      workload.Name,
		WorkloadNamespace: workload.Namespace,
		NumNodes:          b.GetNumGpuNodes(workload),
		ElasticEnabled:    b.elasticEnabled,
	}

	// Call BuildComputeDomain
	computeDomain, err := mnnvlbuilder.BuildComputeDomain(ctx, params, b.k8sClient.Scheme(), workload)
	if err != nil {
		return nil, err
	}

	if computeDomain == nil {
		return nil, nil
	}

	return []client.Object{computeDomain}, nil
}

// GetMnnvlMode returns the MNNVL mode for the workload, from the label.
// if the mode is not set, it returns an empty string
// if the mode is invalid, it returns an error
func (b *MnnvlBuilder) GetMnnvlMode(workload *dinfw.DistributedInferenceWorkload) (mnnvltypes.Mode, error) {
	if workload.Labels == nil {
		return "", nil
	}

	mnnvlMode, exists := workload.Labels[mnnvltypes.FeatureLabelKey]
	if !exists {
		return "", nil
	}

	return mnnvltypes.GetMode(mnnvlMode)
}

// IsMnnvlRequired checks if the workload has the MNNVL mode set to Required
func (b *MnnvlBuilder) IsMnnvlRequired(workload *dinfw.DistributedInferenceWorkload) bool {
	mnnvlMode, err := b.GetMnnvlMode(workload)
	if err != nil {
		return false
	}

	return mnnvlMode == mnnvltypes.Required
}

// GetNumGpuNodes extracts the number of nodes from the workload specification.
func (b *MnnvlBuilder) GetNumGpuNodes(workload *dinfw.DistributedInferenceWorkload) int {
	numWorkers := int(workload.Spec.Workers.Deref32default(0))
	numReplicas := int(workload.Spec.Replicas.Deref32default(1))

	leaderGpuNodes := 0
	if gpu_validator.HasRequestedGpus(&workload.LeaderSpec.ResourceAllocationFields) {
		leaderGpuNodes = 1
	}

	workerGpuNodes := 0
	if workload.WorkerSpec != nil && gpu_validator.HasRequestedGpus(&workload.WorkerSpec.ResourceAllocationFields) {
		workerGpuNodes = numWorkers
	}

	return numReplicas * (leaderGpuNodes + workerGpuNodes)
}

// IsNodePoolsSupportMnnvl checks if all the node pools of the workload support MNNVL
func (b *MnnvlBuilder) IsNodePoolsSupportMnnvl(ctx context.Context, workload *dinfw.DistributedInferenceWorkload) (bool, error) {
	nodepoolNames, err := b.getNodePoolsNames(ctx, workload)
	if err != nil {
		return false, err
	}

	nodepools, err := nodepoolbuilders.GetNodePools(ctx, b.k8sClient, nodepoolNames, "")
	if err != nil {
		return false, err
	}
	return nodePoolsSupportMnnvl(nodepools), nil
}

// nodePoolsSupportMnnvl reports whether every node pool supports MNNVL. On the KAI CRD the
// detection result is an annotation rather than a status field; an absent or unparseable
// value counts as not detected, so unsupported pools stay unsupported.
func nodePoolsSupportMnnvl(nodepools []kaiv1alpha1.NodePool) bool {
	for _, nodepool := range nodepools {
		detected, err := strconv.ParseBool(nodepool.Annotations[kaiv1alpha1.AnnotationGPUNetworkAccelerationDetected])
		if err != nil || !detected {
			return false
		}
	}
	return true
}

// getNodePools gets the nodepools from the workload spec, project, or default nodepool
func (b *MnnvlBuilder) getNodePoolsNames(ctx context.Context, workload *dinfw.DistributedInferenceWorkload) (string, error) {
	// get nodepools from the workload spec
	nodepoolNames := workload.Spec.NodePools.Deref()
	if nodepoolNames != "" {
		return nodepoolNames, nil
	}

	// if no nodepools are found in spec, get the default nodepools from the project
	project, err := projectutils.ProjectOfNamespace(ctx, b.k8sClient, workload.Namespace)
	if err != nil {
		return "", err
	}
	nodepoolNames = strings.Join(project.Spec.DefaultNodePools, " ")
	if nodepoolNames != "" {
		return nodepoolNames, nil
	}

	// if no nodepools are found in the project, use the default nodepool
	return commonconstants.DefaultNodePoolName, nil
}
