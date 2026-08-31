/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package common

import (
	"context"

	workloadbuilders "github.com/run-ai/runai/runai-cluster/cluster/workload-common/builders"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/common"
	job "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ServiceBuilder struct {
	k8sClient client.Client
}

func NewServiceBuilder(k8sClient client.Client) *ServiceBuilder {
	return &ServiceBuilder{k8sClient: k8sClient}
}

func (self *ServiceBuilder) GetPreview(ctx context.Context, workload *job.JobWorkload) ([]common.Resource, error) {
	workloadInitials, err := GetWorkloadInitials(workload.Kind)
	if err != nil {
		return nil, err
	}

	sb := workloadbuilders.NewServiceBuilder(self.k8sClient)
	resources, err := sb.GetPreview(ctx, workload.Kind, workloadInitials.String(), &workload.ObjectMeta, workload.Spec.Ports, workload.Spec.ServiceType)
	if err != nil {
		return nil, err
	}

	resourcesCommon := common.CommonResourceToResource(resources)

	return resourcesCommon, nil
}

// Build - building the resources schemas
func (self *ServiceBuilder) Build(ctx context.Context, workload *job.JobWorkload, options ResourceBuilderOptions) ([]client.Object, error) {
	jobType := GetJobType(workload)
	workloadInitials, err := GetWorkloadInitials(workload.Kind)
	if err != nil {
		return nil, err
	}

	sb := workloadbuilders.NewServiceBuilder(self.k8sClient)

	calculatedPriorityClassNameLabel := CalculatePriorityClassNameLabel(workload.Spec.PriorityClassName.Deref(), workload.Spec.Preemptibility, workload.Kind)

	services, err := sb.Build(ctx, workload.Kind, jobType, workloadInitials.String(), &workload.ObjectMeta, workload.GetChildName(),
		options.Owner, workload.Spec.Ports, workload.Spec.ServiceType, calculatedPriorityClassNameLabel)
	if err != nil {
		return nil, err
	}

	err = workloadbuilders.SetObjectMetaLabels(ctx, workload.Kind, jobType, workload.Name, &workload.ObjectMeta)
	if err != nil {
		return nil, err
	}

	resources := make([]client.Object, len(services))
	for i, service := range services {
		resources[i] = &service
	}

	return resources, nil
}
