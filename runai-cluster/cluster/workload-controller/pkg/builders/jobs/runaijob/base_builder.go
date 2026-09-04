/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaijob

import (
	"context"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	workloadbuilders "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/builders"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/builders/common"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/builders/factory"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// JobBuilder is common builder for interactive and training jobs
type JobBuilder struct {
	factory.Builder
	common.OldSharedBuilder

	//  special interactive tools
	IsJupyter     bool
	IsTensorboard bool

	//  the input to the builder is the merged workload
	MergedWorkload *jobs.JobWorkload
}

func (self *JobBuilder) BuildObjectMeta(ctx context.Context, om *metav1.ObjectMeta) error {

	self.OldSharedBuilder.BuildObjectMeta(ctx, om, &self.MergedWorkload.TypeMeta)

	om.Name = self.MergedWorkload.GetChildName()
	om.Annotations[constants.AnnotationTemplate] = self.MergedWorkload.Name

	calculatedPriorityClassNameLabel := common.CalculatePriorityClassNameLabel(self.MergedWorkload.Spec.PriorityClassName.Deref(), self.MergedWorkload.Spec.Preemptibility, self.MergedWorkload.Kind)
	err := workloadbuilders.SetPriorityClassLabel(ctx, self.K8sClient, calculatedPriorityClassNameLabel, om)
	if err != nil {
		return err
	}

	calculatedPreemptibilityNameLabel := common.CalculatePreemptibilityNameLabel(self.MergedWorkload.Spec.Preemptibility, self.MergedWorkload.Kind)
	workloadbuilders.SetPreemptibilityLabel(calculatedPreemptibilityNameLabel, om)

	//
	//  jupyter/tensorboard label is needed by frontend to identify jupyter/tensorboard jobs. we only display it
	//  in case the service type is ingress, otherwise the "connect to url" functionality
	//  is not working anyway (it's mainly jupyter originating from CLI)
	//
	if self.MergedWorkload.Spec.IsJupyter.Deref() {
		om.Labels[constants.LabelJupyter] = "true"
	}

	if self.MergedWorkload.Spec.IsTensorboard.Deref() {
		om.Labels[constants.LabelTensorboard] = "true"
	}

	return nil
}
