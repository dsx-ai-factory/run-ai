/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	"context"
	"fmt"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/dra"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/handlers/commonparsers"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/utils"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	batchv1 "k8s.io/api/batch/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type CronJobTopOwnerHandler struct {
	cachedClient client.Client
	draParser    *dra.DRAParser
}

func NewCronJobTopOwnerHandler(cachedClient client.Client, draParser *dra.DRAParser) *CronJobTopOwnerHandler {
	return &CronJobTopOwnerHandler{
		cachedClient: cachedClient,
		draParser:    draParser,
	}
}

func (h *CronJobTopOwnerHandler) GetObject(ctx context.Context, name, namespace string) (client.Object, error) {
	return utils.GetObject(ctx, h.cachedClient, name, namespace, func() client.Object { return &batchv1.CronJob{} })
}

func (h *CronJobTopOwnerHandler) Parse(ctx context.Context, object client.Object,
	externalWorkload *v1alpha1.ExternalWorkload) error {
	cronJob, ok := object.(*batchv1.CronJob)
	if !ok {
		return fmt.Errorf("failed converting object to batchv1.CronJob")
	}

	commonparsers.FillExternalWorkloadDataFromPodTemplateSpec(ctx, cronJob.ObjectMeta, cronJob.GroupVersionKind().GroupKind(),
		&cronJob.Spec.JobTemplate.Spec.Template, nil, h.cachedClient, h.draParser, externalWorkload)

	externalWorkload.Spec.RequestedPods.Parallelism = cronJob.Spec.JobTemplate.Spec.Parallelism
	externalWorkload.Spec.RequestedPods.Completions = cronJob.Spec.JobTemplate.Spec.Completions

	setCronJobObservedStatus(&cronJob.Status, &externalWorkload.Status.ObservedStatus)

	return nil
}

func setCronJobObservedStatus(cronJobStatus *batchv1.CronJobStatus,
	observedStatus *v1alpha1.ExternalWorkloadObservedStatus) {
	if len(cronJobStatus.Active) == 0 {
		observedStatus.K8sStatus = commonparsers.SuspendedStatus
		observedStatus.Phase = v1alpha1.ExWStoppedPhase
		return
	}

	observedStatus.K8sStatus = commonparsers.ActiveStatus
	observedStatus.Phase = v1alpha1.ExWRunningPhase
}

func (h *CronJobTopOwnerHandler) Filter(object client.Object) bool {
	cronJob, ok := object.(*batchv1.CronJob)
	if !ok {
		return false
	}
	return cronJob.Spec.JobTemplate.Spec.Template.Spec.SchedulerName == constants.RunaiSchedulerName
}
