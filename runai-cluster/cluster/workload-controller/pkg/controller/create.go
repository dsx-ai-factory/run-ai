/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package controller

import (
	"fmt"

	workloadissue "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/builders/common/issue"
	"k8s.io/client-go/tools/record"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/builders/common"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/builders/distributed"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/builders/jobs/runaijob"
	workloadCommon "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/finalizers"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/managers"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/status/aggregator"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/status/update"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Options struct {
	FailureResourceCleanupPolicy  managers.FailureCleanupPolicy
	MnnvlSupportedInCluster       bool
	ElasticCDSupportedInCluster   bool
	EnableDanglingResourceCleanup bool
}

type ControllerFactory interface {
	// CreateController creates a controller for the given workload
	CreateController() (*JobController, error)
}

type InteractiveWorkloadControllerFactory struct {
	BaseWorkloadControllerFactory
}

func (f *InteractiveWorkloadControllerFactory) CreateController() (*JobController, error) {
	return f.BaseWorkloadControllerFactory.CreateController()

}

func NewInteractiveWorkloadControllerFactory(client client.Client, apiReadClient client.Reader, isOpenshift bool, eventRecorder record.EventRecorder, controllerOptions Options) *InteractiveWorkloadControllerFactory {
	return &InteractiveWorkloadControllerFactory{
		BaseWorkloadControllerFactory: BaseWorkloadControllerFactory{
			client:                   client,
			apiReadClient:            apiReadClient,
			isOpenshift:              isOpenshift,
			workloadType:             workloadCommon.InteractiveWorkload,
			computeBuilderCreationFn: runaijob.NewRunaiJobBuilder,
			eventRecorder:            eventRecorder,
			controllerOptions:        controllerOptions,
		},
	}
}

type TrainingWorkloadControllerFactory struct {
	BaseWorkloadControllerFactory
}

func (twc *TrainingWorkloadControllerFactory) CreateController() (*JobController, error) {
	return twc.BaseWorkloadControllerFactory.CreateController()
}

func NewTrainingWorkloadControllerFactory(client client.Client, apiReadClient client.Reader, isOpenshift bool,
	eventRecorder record.EventRecorder, podLabels map[string]string,
	podAnnotations map[string]string, controllerOptions Options) *TrainingWorkloadControllerFactory {
	return &TrainingWorkloadControllerFactory{
		BaseWorkloadControllerFactory: BaseWorkloadControllerFactory{
			client:                   client,
			apiReadClient:            apiReadClient,
			isOpenshift:              isOpenshift,
			workloadType:             workloadCommon.TrainingWorkload,
			computeBuilderCreationFn: runaijob.NewRunaiJobBuilder,
			eventRecorder:            eventRecorder,
			podLabels:                podLabels,
			podAnnotations:           podAnnotations,
			controllerOptions:        controllerOptions,
		},
	}
}

type DistributedWorkloadControllerFactory struct {
	BaseWorkloadControllerFactory
}

func NewDistributedWorkloadControllerFactory(client client.Client, apiReadClient client.Reader, isOpenshift bool,
	eventRecorder record.EventRecorder, podLabels map[string]string,
	podAnnotations map[string]string, controllerOptions Options) *DistributedWorkloadControllerFactory {
	return &DistributedWorkloadControllerFactory{
		BaseWorkloadControllerFactory: BaseWorkloadControllerFactory{
			client:                   client,
			apiReadClient:            apiReadClient,
			isOpenshift:              isOpenshift,
			workloadType:             workloadCommon.DistributedWorkload,
			computeBuilderCreationFn: distributed.CreateDistributedComputeBuilder,
			eventRecorder:            eventRecorder,
			podLabels:                podLabels,
			podAnnotations:           podAnnotations,
			controllerOptions:        controllerOptions,
		},
	}
}

type BaseWorkloadControllerFactory struct {
	client                   client.Client
	apiReadClient            client.Reader
	computeBuilderCreationFn func(k8sClient client.Client,
		podLabels map[string]string, podAnnotations map[string]string, elasticEnabled bool) common.ComputeBuilder
	isOpenshift       bool
	workloadType      workloadCommon.WorkloadType
	eventRecorder     record.EventRecorder
	podLabels         map[string]string
	podAnnotations    map[string]string
	controllerOptions Options
}

func NewBaseWorkloadControllerFactory(client client.Client, apiReadClient client.Reader, isOpenshift bool,
	workloadType workloadCommon.WorkloadType, jobBuilderCreationFn func(client.Client, map[string]string,
		map[string]string, bool) common.ComputeBuilder, controllerOptions Options) *BaseWorkloadControllerFactory {
	return &BaseWorkloadControllerFactory{
		client:                   client,
		apiReadClient:            apiReadClient,
		computeBuilderCreationFn: jobBuilderCreationFn,
		isOpenshift:              isOpenshift,
		workloadType:             workloadType,
		controllerOptions:        controllerOptions,
	}
}

func (bcf *BaseWorkloadControllerFactory) CreateController() (*JobController, error) {
	computeBuilder := bcf.computeBuilderCreationFn(bcf.client, bcf.podLabels, bcf.podAnnotations, bcf.controllerOptions.ElasticCDSupportedInCluster)
	serviceBuilder := common.NewServiceBuilder(bcf.client)
	externalUrlBuilder := common.NewExternalsUrlsBuilder(bcf.client, bcf.isOpenshift)
	pvcBuilder := common.NewPvcBuilder(bcf.client)
	credentialsBuilder := common.NewCredentialsBuilder(bcf.client)
	commonResourceBuilder := common.NewShardResourcesBuilder(serviceBuilder, externalUrlBuilder, pvcBuilder, credentialsBuilder)
	statusUpdaterFactory, err := update.GetStatusUpdaterFactory(bcf.workloadType)
	if err != nil {
		// this should never happen
		return nil, fmt.Errorf("failed to create status updater factory: %v", err)
	}
	statusUpdater := statusUpdaterFactory.CreateStatusUpdater(bcf.client)
	statusAggregator := aggregator.NewStatusCalculator(bcf.controllerOptions.MnnvlSupportedInCluster)
	statusManager := update.NewSimpleStatusUpdaterManager(statusUpdater)
	finalizerRemoverFactory, err := finalizers.GetRemoveFinalizerFactory(bcf.workloadType)
	if err != nil {
		// this should never happen
		return nil, fmt.Errorf("failed to create finalizer remover factory: %v", err)
	}
	finalizerRemover := finalizerRemoverFactory.CreateRemoveFinalizer(bcf.client)
	k8sManager := managers.NewResourceManager(bcf.client, bcf.apiReadClient, &workloadissue.ComputeResourceGrouper{})
	return NewJobController(bcf.client,
		computeBuilder,
		commonResourceBuilder,
		statusManager, finalizerRemover,
		statusAggregator, k8sManager, managers.ResourceOptions{}, bcf.eventRecorder, managers.NewCleanupManager(bcf.client, bcf.controllerOptions.FailureResourceCleanupPolicy, bcf.eventRecorder),
		bcf.controllerOptions.EnableDanglingResourceCleanup), nil
}
