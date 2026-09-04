/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	"context"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/runtime-rules/pkg/kubeConfig"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/runtime-rules/pkg/prometheus"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/runtime-rules/pkg/queuesManager"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/runtime-rules/pkg/util"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	"go.uber.org/zap"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	_ "k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type WorkloadHandler struct {
	K8sClient         client.Client
	QueuesManager     *queuesManager.QueuesManager
	PrometheusQuerier prometheus.Querier
	workloadCache     *util.WorkloadCache
	supportedJobKinds map[string]client.Object
	runnerConverter   *util.RunnerConverter
}

func NewWorkloadHandler(k8sClient client.Client, qm *queuesManager.QueuesManager, promQuerier prometheus.Querier, workloadCache *util.WorkloadCache, supportedJobKinds map[string]client.Object) *WorkloadHandler {
	return &WorkloadHandler{
		K8sClient:         k8sClient,
		QueuesManager:     qm,
		PrometheusQuerier: promQuerier,
		workloadCache:     workloadCache,
		supportedJobKinds: supportedJobKinds,
		runnerConverter:   util.NewRunnerConverter(k8sClient),
	}
}

// HandleRunnerEvent is the WorkloadRunner-specific entry point: it reconciles
// the runner's lifecycle annotations (so the conversion sees fresh
// start-time/completion-time on the same event), projects the runner onto a
// JobWorkload via the runner converter (which fetches the wrapped child for
// GPU + auto-delete fields), then dispatches into the runtime-limit enqueue path.
//
// Two distinct gates apply, in order:
//   - The scope gate (cast label + in-scope wrapped kind): out-of-scope runners
//     aren't ours, so we short-circuit before touching their annotations.
//   - The phase gate: an in-scope runner that isn't in an enforceable
//     (Running/Degraded) phase still has its lifecycle annotations reconciled —
//     terminal phases stamp completion-time, Suspended/Stopped clear the
//     timestamps — but is not enqueued for enforcement. Only enforceable phases
//     proceed to conversion + HandleAddUpdate.
func (workloadHandler *WorkloadHandler) HandleRunnerEvent(ctx context.Context, runner *dromosv1alpha1.WorkloadRunner) error {
	category, skipReason, ok := util.ClassifyRunner(runner)
	if !ok && !util.IsPhaseSkip(skipReason) {
		zap.S().Debugf("WorkloadRunner %s/%s skipped at runner-event entry: reason=%s", runner.Namespace, runner.Name, skipReason)
		return nil
	}
	zap.S().Debugf("WorkloadRunner %s/%s in scope at runner-event entry: category=%s", runner.Namespace, runner.Name, category)
	util.WarnIfMissingWorkloadID(runner)

	if err := workloadHandler.ReconcileTimestampAnnotations(ctx, runner); err != nil {
		zap.S().Warnf("lifecycle annotation reconcile failed for runner %s/%s: %v", runner.Namespace, runner.Name, err)
		return err
	}

	if !ok {
		zap.S().Debugf("WorkloadRunner %s/%s phase-skipped after annotation reconcile: reason=%s", runner.Namespace, runner.Name, skipReason)
		return nil
	}

	var jobWorkload jobs.JobWorkload
	if err := workloadHandler.runnerConverter.Convert(runner, &jobWorkload); err != nil {
		return err
	}
	return workloadHandler.HandleAddUpdate(ctx, &jobWorkload)
}

func (workloadHandler *WorkloadHandler) HandleAddUpdate(ctx context.Context, jobWorkload *jobs.JobWorkload) error {
	jobUuid, jobName, startTime, validationError := util.GetWorkloadValues(jobWorkload, workloadHandler.supportedJobKinds)
	if validationError != nil {
		// in case of error - just skip the workload
		zap.S().Debugf("Workload (%s) is not valid, validation error (%s)", jobWorkload.Name, validationError.Error())
		return nil
	}

	// Runner-sourced events go through the scope filter (cast label,
	// in-scope wrapped kind, enforceable phase) before they can reach the
	// project queue. Category derivation is intentionally not surfaced here —
	// the projection of the runner onto the unified internal workload model is
	// owned by a later conversion step, which will re-run ClassifyRunner against the live
	// runner when it does the full conversion. The live WorkloadRunner is read
	// once per event because the in-memory JobWorkload's labels are a
	// projected subset.
	category := ""
	if jobWorkload.Kind == constants.KindWorkloadRunner {
		runner, getErr := kubeConfig.GetWLR(ctx, workloadHandler.K8sClient, jobWorkload.Namespace, jobWorkload.Name)
		if getErr != nil {
			if apierrors.IsNotFound(getErr) {
				return nil
			}
			return getErr
		}
		derivedCategory, skipReason, ok := util.ClassifyRunner(&runner)
		if !ok {
			zap.S().Debugf("WorkloadRunner %s/%s skipped: reason=%s", jobWorkload.Namespace, jobWorkload.Name, skipReason)
			return nil
		}
		category = derivedCategory
		zap.S().Debugf("WorkloadRunner %s/%s in scope: category=%s", jobWorkload.Namespace, jobWorkload.Name, category)
		util.WarnIfMissingWorkloadID(&runner)
	}

	projectName, err := kubeConfig.ProjectNameOfNamespace(ctx, workloadHandler.K8sClient, jobWorkload.Namespace, jobWorkload.Labels)
	if err != nil {
		return err
	}

	cacheKey := util.GenerateCacheKey(jobWorkload.UID, startTime.String())
	if workloadHandler.workloadCache.WasProcessed(projectName, cacheKey) {
		// in case workload that had already processed - just skip the workload
		zap.S().Infof("Workload %s, namespace: %s, was already processed, cacheKey %s ", jobWorkload.Name, jobWorkload.Namespace, cacheKey)
		return nil
	}

	projectQueue := workloadHandler.QueuesManager.Get(projectName)
	if projectQueue == nil {
		zap.S().Warnf("project queue is nil for project %s, if this project has not time limit, ignore this warninig", projectName)
		return nil
	}

	action := queuesManager.NewActionList(projectQueue.EventList, 0, 0, nil)
	workloadId, err := util.GetWorkloadID(jobWorkload)
	if err != nil {
		return err
	}
	workloadTask := &queuesManager.WorkloadTask{ // just a dummy WorkloadTask (new one will be set)
		StartTime:            startTime.Time,
		WorkloadName:         jobWorkload.Name,
		WorkloadID:           workloadId,
		ProjectName:          projectQueue.ProjectName,
		Namespace:            jobWorkload.Namespace,
		JobUuid:              jobUuid,
		JobName:              jobName,
		WorkloadType:         jobWorkload.Kind,
		Category:             category,
		TaskAction:           *action,
		WorkloadRequestedGpu: util.HasRequestedGpu(jobWorkload),
		IsPreemptible:        util.IsJobPreemptible(jobWorkload),
	}
	err = queuesManager.HandleWorkloadTask(ctx, workloadHandler.workloadCache, workloadHandler.K8sClient, workloadHandler.PrometheusQuerier, workloadTask, projectQueue, false)
	if err != nil {
		return err
	}

	return nil
}
