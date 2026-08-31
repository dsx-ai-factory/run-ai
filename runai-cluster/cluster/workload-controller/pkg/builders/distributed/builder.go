/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package distributed

import (
	"context"
	"fmt"
	gpu_validator "github.com/run-ai/runai/runai-cluster/cluster/workload-common/gpu-validator"

	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/internal/mnnvlutils"

	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	mnnvlcommon "github.com/run-ai/runai/runai-cluster/cluster/workload-common/builders/mnnvl"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/common"
	dwv2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedworkload/v2alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	builder "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/builders/common"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/builders/distributed/jax"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/builders/distributed/mpi"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/builders/distributed/pytorch"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/builders/distributed/tfjob"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/builders/distributed/xgboost"
	oldFactory "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/builders/factory"
	workloadCommon "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/common/constants"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// this struct is facade for old interface  cluster/workload-controller/pkg/builders/factory/interfaces.go
type interfaceBuilder func(fromObject interface{}, toResources interface{},
	k8sClient client.Client, options oldFactory.BuilderOptions,
	existingPodsSpecs map[workloadCommon.ReplicaType]*v1.PodSpec,
	podLabels map[string]string, podAnnotations map[string]string) (oldFactory.Builder, error)

type ComputeBuilder struct {
	distributedWorkloadJob string
	internalBuilderFn      interfaceBuilder
	k8sClient              client.Client
	finalObjectFn          func() client.Object
	podLabels              map[string]string
	podAnnotations         map[string]string
	elasticEnabled         bool
}

func NewComputeBuilder(distributedWorkloadJob string,
	internalBuilderFn interfaceBuilder,
	k8sClient client.Client, finalObjectFn func() client.Object,
	podLabels map[string]string, podAnnotations map[string]string, elasticEnabled bool) *ComputeBuilder {
	return &ComputeBuilder{
		distributedWorkloadJob: distributedWorkloadJob,
		internalBuilderFn:      internalBuilderFn,
		k8sClient:              k8sClient,
		finalObjectFn:          finalObjectFn,
		podLabels:              podLabels,
		podAnnotations:         podAnnotations,
		elasticEnabled:         elasticEnabled,
	}

}

func (d ComputeBuilder) Build(ctx context.Context, workload *jobs.JobWorkload, options *builder.ComputeBuilderOptions) (*builder.ComputeResources, error) {
	job, err := d.buildDistributedJob(ctx, workload, options)
	if err != nil {
		return nil, err
	}

	computeResources := &builder.ComputeResources{
		Job: job,
	}

	mnnvlStatus := mnnvlutils.FromContext(ctx)
	if mnnvlStatus.IsNetworkAccelerated() {
		params := mnnvlcommon.ComputeDomainWorkloadParams{
			WorkloadKind:      workload.Kind,
			JobType:           workload.Spec.JobType,
			WorkloadInitials:  builder.KindDistributedWorkloadInitial,
			WorkloadName:      workload.Name,
			WorkloadNamespace: workload.Namespace,
			NumNodes:          calculateNumNodesForMNNVL(workload),
			ElasticEnabled:    d.elasticEnabled,
		}

		var owner client.Object
		if options != nil {
			owner = options.Owner
		}

		computeDomain, err := mnnvlcommon.BuildComputeDomain(ctx, params, d.k8sClient.Scheme(), owner)
		if err != nil {
			return nil, err
		}
		computeResources.ComputeDomain = computeDomain
	}

	return computeResources, nil
}

func (d ComputeBuilder) buildDistributedJob(ctx context.Context, workload *jobs.JobWorkload, options *builder.ComputeBuilderOptions) (client.Object, error) {
	if options == nil {
		options = &builder.ComputeBuilderOptions{}
	}
	logger := logf.FromContext(ctx).WithValues("workload", workload.GetChildName())
	distributedWorkload := dwv2.NewFromJobWorkload(workload)
	finalObject := d.finalObjectFn()
	distributedBuilder, err := d.internalBuilderFn(distributedWorkload, finalObject, d.k8sClient,
		oldFactory.BuilderOptions{}, options.ExitingPodsSpec, d.podLabels, d.podAnnotations)
	if err != nil {
		return nil, fmt.Errorf("failed to create builder for job %s: %w", d.distributedWorkloadJob, err)
	}
	logger.Info("Building distributed job from workload")
	if distributedBuilder == nil {
		return nil, fmt.Errorf("failed to build distributed job %s: distributed builder is nil", d.distributedWorkloadJob)
	}
	err = distributedBuilder.Build(ctx)
	if err != nil {
		logger.Error(err, "Failed to build distributed job from workload")
		return nil, fmt.Errorf("failed to build distributed job %s: %v", d.distributedWorkloadJob, err)
	}
	job := distributedBuilder.GetJob()
	if job == nil {
		logger.Info("No job was created")
		return nil, fmt.Errorf("fail to build distributed job %s: no job was created", d.distributedWorkloadJob)
	}
	if options.Owner != nil {
		err = controllerutil.SetControllerReference(options.Owner, job, d.k8sClient.Scheme())
		if err != nil {
			return nil, fmt.Errorf("failed to set controller reference for job, %v", err)
		}
	}
	return job, nil
}

func (d ComputeBuilder) GetPreview(ctx context.Context, workload *jobs.JobWorkload) ([]common.Resource, error) {
	distributedWorkload := dwv2.NewFromJobWorkload(workload)

	preview := []common.Resource{
		{
			Kind: metav1.GroupVersionKind{
				Kind: d.distributedWorkloadJob,
			},
			Name:      workload.GetChildName(),
			Namespace: distributedWorkload.Namespace,
		},
	}

	mnnvlStatus := mnnvlutils.FromContext(ctx)
	if mnnvlStatus.IsNetworkAccelerated() {
		previewComputeDomain := mnnvlcommon.PreviewComputeDomain(string(builder.KindDistributedWorkloadInitial), workload.Name, distributedWorkload.Namespace)
		previewComputeDomainResource := common.Resource{
			Kind: metav1.GroupVersionKind{
				Kind: previewComputeDomain.Kind,
			},
			Name:      previewComputeDomain.Name,
			Namespace: previewComputeDomain.Namespace,
		}
		preview = append(preview, previewComputeDomainResource)
	}

	return preview, nil
}

type ComputeBuilderFactory struct {
	builders map[string]builder.ComputeBuilder
}

func NewComputeBuilderFactory(builders map[string]builder.ComputeBuilder) *ComputeBuilderFactory {
	return &ComputeBuilderFactory{
		builders: builders,
	}
}

func (j *ComputeBuilderFactory) Build(ctx context.Context, workload *jobs.JobWorkload, options *builder.ComputeBuilderOptions) (*builder.ComputeResources, error) {
	jobType := workload.Spec.JobType
	computeBuilder, ok := j.builders[jobType]
	if !ok {
		return nil, fmt.Errorf("no computeBuilder found for job type %s", jobType)
	}
	return computeBuilder.Build(ctx, workload, options)
}

func (j *ComputeBuilderFactory) GetPreview(ctx context.Context, workload *jobs.JobWorkload) ([]common.Resource, error) {
	jobType := workload.Spec.JobType
	computeBuilder, ok := j.builders[jobType]
	if !ok {
		return nil, fmt.Errorf("no computeBuilder found for job type %s", jobType)
	}
	return computeBuilder.GetPreview(ctx, workload)
}

func CreateDistributedComputeBuilder(k8sClient client.Client,
	podLabels map[string]string, podAnnotations map[string]string, elasticEnabled bool) builder.ComputeBuilder {
	distributedJobBuilders := map[string]builder.ComputeBuilder{
		constants.KindMpiJob:     NewComputeBuilder(constants.KindMpiJob, mpi.NewMpiBuilderWithExistingNodeAffinity, k8sClient, mpi.NewMpiJobFn, podLabels, podAnnotations, elasticEnabled),
		constants.KindPyTorchJob: NewComputeBuilder(constants.KindPyTorchJob, pytorch.NewPyTorchBuilderWithExistingNodeAffinities, k8sClient, pytorch.NewPytorchJobFn, podLabels, podAnnotations, elasticEnabled),
		constants.KindTFJob:      NewComputeBuilder(constants.KindTFJob, tfjob.NewTFJobBuilderWithExistingNodeAffinities, k8sClient, tfjob.NewTFJobFn, podLabels, podAnnotations, elasticEnabled),
		constants.KindXGBoostJob: NewComputeBuilder(constants.KindXGBoostJob, xgboost.NewXGBoostJobBuilderWithExistingNodeAffinities, k8sClient, xgboost.NewXGBoostJobFn, podLabels, podAnnotations, elasticEnabled),
		constants.KindJAXJob:     NewComputeBuilder(constants.KindJAXJob, jax.NewJAXJobBuilderWithExistingNodeAffinities, k8sClient, jax.NewJAXJobFn, podLabels, podAnnotations, elasticEnabled),
	}
	return NewComputeBuilderFactory(distributedJobBuilders)
}

// calculateNumNodes returns the number of pods that has GPU requests
func calculateNumNodesForMNNVL(workload *jobs.JobWorkload) int {
	numNodes := 0
	if gpu_validator.HasRequestedGpus(&workload.Spec.ResourceAllocationFields) {
		numNodes += workload.Spec.DistributedFields.Workers.Deref()
	}

	if workload.MasterSpec != nil && gpu_validator.HasRequestedGpus(&workload.MasterSpec.ResourceAllocationFields) {
		numNodes += 1
	}

	return numNodes
}
