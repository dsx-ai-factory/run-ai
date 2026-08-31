/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package migration

import (
	"context"

	clusterCommon "github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	runaijob "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	iwv2a1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/interactiveworkload/v2alpha1"
	twv2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/trainingworkload/v2alpha1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	SuspendWorkloadMigrationName = "suspend-workloads"
)

type SuspendWorkloads struct {
	k8sClient              client.Client
	MigrateExecuteCreateFn func(int, []runaijob.RunaiJob, func(context.Context, runaijob.RunaiJob) error) Executor // for tests only
	migrationMaxWorkers    int
}

func (sw *SuspendWorkloads) String() string {
	return SuspendWorkloadMigrationName
}

func NewSuspendWorkloads(k8sClient client.Client, migrationMaxWorkers int) *SuspendWorkloads {
	return &SuspendWorkloads{
		k8sClient:           k8sClient,
		migrationMaxWorkers: migrationMaxWorkers,
	}
}

func (sw *SuspendWorkloads) RequiresWebhooks() bool {
	return true
}

func (sw *SuspendWorkloads) Migrate(ctx context.Context) error {
	runaiJobGVK := runaijob.GroupVersion.WithKind("RunaiJob")
	_, err := sw.k8sClient.RESTMapper().RESTMapping(runaiJobGVK.GroupKind(), runaiJobGVK.Version)
	if meta.IsNoMatchError(err) {
		log.Info("Skipping suspend workload migration because RunaiJob API is not served")
		return nil
	}
	if err != nil {
		return err
	}

	p := NewParallelMigrator[runaijob.RunaiJob](
		SuspendWorkloadMigrationName,
		sw.k8sClient,
		sw.getNextBatch,
		NewSuspendHandler(sw.k8sClient),
		sw.MigrateExecuteCreateFn,
		sw.migrationMaxWorkers,
	)
	return p.Migrate(ctx)
}

func (sw *SuspendWorkloads) getNextBatch(continueToken string, ctx context.Context) ([]runaijob.RunaiJob, string, error) {
	runaiJobs := &runaijob.RunaiJobList{}

	pagination := client.ListOptions{
		Limit:    paginationLimit,
		Continue: continueToken,
	}

	err := sw.k8sClient.List(ctx, runaiJobs, &pagination)
	if err != nil {
		return nil, "", err
	}
	jobs := sw.filterRunaiJobs(runaiJobs)

	return jobs, runaiJobs.Continue, nil
}

func (sw *SuspendWorkloads) filterRunaiJobs(runaiJobs *runaijob.RunaiJobList) []runaijob.RunaiJob {
	var notOwnedByAnyWorkloadJobs []runaijob.RunaiJob
	for _, runaiJobItem := range runaiJobs.Items {
		cpRunaiJob := runaiJobItem
		if cpRunaiJob.OwnerReferences != nil && len(cpRunaiJob.OwnerReferences) > 0 {
			continue
		}
		if cpRunaiJob.Labels == nil || cpRunaiJob.Labels[clusterCommon.LabelWorkloadName] == "" ||
			cpRunaiJob.Labels[clusterCommon.LabelWorkloadKind] == "" {
			continue
		}
		// should be filtered by suspend field selector but the fake client does not support it, so we filter it here
		if cpRunaiJob.Spec.Suspend == nil || !(*cpRunaiJob.Spec.Suspend) {
			continue
		}
		notOwnedByAnyWorkloadJobs = append(notOwnedByAnyWorkloadJobs, cpRunaiJob)

	}
	return notOwnedByAnyWorkloadJobs
}

type SuspendHandler struct {
	k8sClient      client.Client
	activeRawPatch client.Patch
}

func NewSuspendHandler(k8sClient client.Client) *SuspendHandler {
	return &SuspendHandler{
		k8sClient:      k8sClient,
		activeRawPatch: client.RawPatch(types.MergePatchType, []byte(SuspendPatchString)),
	}
}

func (wh *SuspendHandler) Handle(ctx context.Context, runaiJob runaijob.RunaiJob) error {
	workloadName := runaiJob.Labels[clusterCommon.LabelWorkloadName]
	workloadKind := runaiJob.Labels[clusterCommon.LabelWorkloadKind]
	switch workloadKind {
	case clusterCommon.KindTrainingWorkload:
		return wh.handleTrainingWorkload(ctx, workloadName, runaiJob.Namespace)
	case clusterCommon.KindInteractiveWorkload:
		return wh.handleInteractiveWorkload(ctx, workloadName, runaiJob.Namespace)
	default:
		// not supposed to happen
		log.Info("unsupported workload kind, skipping", "workloadKind", workloadKind)
		return nil

	}
}

func (wh *SuspendHandler) handleTrainingWorkload(ctx context.Context, workloadName, namespace string) error {
	// use patch instead of get and update to avoid conflict
	trainingWorkload := &twv2.TrainingWorkload{
		ObjectMeta: metav1.ObjectMeta{
			Name:      workloadName,
			Namespace: namespace,
		},
	}
	return wh.basePatch(ctx, trainingWorkload)
}

func (wh *SuspendHandler) basePatch(ctx context.Context, workload client.Object) error {
	err := wh.k8sClient.Patch(ctx, workload, wh.activeRawPatch)
	if err != nil {
		if client.IgnoreNotFound(err) != nil {
			return err
		}
	}
	return nil
}

func (wh *SuspendHandler) handleInteractiveWorkload(ctx context.Context, workloadName, namespace string) error {
	interactiveWorkload := &iwv2a1.InteractiveWorkload{
		ObjectMeta: metav1.ObjectMeta{
			Name:      workloadName,
			Namespace: namespace,
		},
	}
	return wh.basePatch(ctx, interactiveWorkload)
}
