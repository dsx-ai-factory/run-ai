/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	"context"
	"errors"
	"fmt"

	kaiorg "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/runtime-rules/pkg/prometheus"
	"github.com/run-ai/runai/runai-cluster/cluster/runtime-rules/pkg/queuesManager"
	"github.com/run-ai/runai/runai-cluster/cluster/runtime-rules/pkg/util"
	"github.com/run-ai/runai/runai-cluster/cluster/runtime-rules/pkg/validation"
	"go.uber.org/zap"
	_ "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ProjectHandler struct {
	K8sClient         client.Client
	QueuesManager     *queuesManager.QueuesManager
	PrometheusQuerier prometheus.Querier
	workloadCache     *util.WorkloadCache
	Recorder          record.EventRecorder
	EventList         []int
	workloadValidator *validation.WorkloadValidator
	supportedJobKinds map[string]client.Object
}

func NewProjectHandler(k8sClient client.Client, qm *queuesManager.QueuesManager, promQuerier prometheus.Querier, workloadCache *util.WorkloadCache,
	recorder record.EventRecorder, eventList []int, workloadValidator *validation.WorkloadValidator, supportedJobKinds map[string]client.Object) *ProjectHandler {
	return &ProjectHandler{
		K8sClient:         k8sClient,
		QueuesManager:     qm,
		PrometheusQuerier: promQuerier,
		workloadCache:     workloadCache,
		Recorder:          recorder,
		EventList:         eventList,
		workloadValidator: workloadValidator,
		supportedJobKinds: supportedJobKinds,
	}
}

func (projectHandler *ProjectHandler) HandleUpdate(ctx context.Context, project *kaiorg.Project) error {
	if project == nil {
		// this should not happen - but just in case
		return errors.New("project is nil")
	}
	if project.Status.Namespace == "" {
		zap.S().Warnf("Project %s does not have namespace yet, skipping update", project.Name)
		return nil
	}

	projectQueue := projectHandler.QueuesManager.Get(project.Name)

	// in case of no rules, clean the project queue if exists and skip
	if !util.ProjectHasLimits(project) {
		zap.S().Infof("Project %s has no rules, skipping", project.Name)

		if projectQueue != nil {
			projectHandler.resetProjectQueue(project.Name, projectQueue)
		}

		return nil
	}

	isNewQueue := false
	if projectQueue == nil {
		zap.S().Debugf("Project %s not found in the QueuesManager, adding it", project.Name)
		projectQueue = projectHandler.QueuesManager.Add(ctx, projectHandler.workloadCache, projectHandler.K8sClient,
			projectHandler.PrometheusQuerier, project, projectHandler.Recorder, projectHandler.EventList,
			projectHandler.workloadValidator)
		if projectQueue == nil {
			return fmt.Errorf("Project %s not found in the QueuesManager, and failed to add it", project.Name) //nolint:staticcheck
		}

		isNewQueue = true
	}

	// only need to check if an existing queue
	if !isNewQueue && !isProjectChanged(project, projectQueue) {
		zap.S().Debugf("Project %s values did not changed, skip update ", project.Name)
		return nil
	}

	zap.S().Infof(
		"Project (%s) values changed, InteractiveJobMaxIdle (%v) sec, InteractiveJobTimeLimit (%v) sec, "+
			"TrainingJobMaxIdleDuration (%v) sec, TrainingJobTimeLimit (%v) sec, "+
			"InteractivePreemptibleJobMaxIdleDurationSecs (%v)",
		project.Name, util.InteractiveJobMaxIdleDurationSecs(project), util.InteractiveJobTimeLimitSecs(project),
		util.TrainingJobMaxIdleDurationSecs(project), util.TrainingJobTimeLimitSecs(project),
		util.InteractivePreemptibleJobMaxIdleDurationSecs(project))

	// update & cleanup queue only if queue existed before, skip if new
	if !isNewQueue {
		// update the project values in the queue manager
		projectHandler.QueuesManager.Update(project)
		zap.S().Infof("Finish update project %s with the new values", project.Name)

		// clean the project cache
		projectHandler.workloadCache.Clean(project.Name)
		zap.S().Debugf("Finish clean project %s cache", project.Name)

		// clean the project queue from workload that are not relevant anymore
		projectQueue.CleanQueue()
		zap.S().Debugf("Finish clean project %s queue", project.Name)
	}

	// list all workloads in the project
	// add them to the queue again with the update fields
	err := queuesManager.PushProjectWorkloadsToProjectQueue(ctx, projectHandler.workloadCache, projectHandler.K8sClient,
		projectHandler.PrometheusQuerier, project.Name, project.Status.Namespace, projectQueue, projectHandler.supportedJobKinds)
	if err != nil {
		zap.S().Errorf("Error while try to update project queue with project workloads, err: %+v ", err)
		return err
	}
	return nil
}

func (projectHandler *ProjectHandler) resetProjectQueue(projectName string, projectQueue *queuesManager.ProjectQueue) {
	zap.S().Infof("Cleaning project %s cache", projectName)
	projectHandler.workloadCache.Clean(projectName)

	// clean the project queue if exist
	if projectQueue != nil && !projectQueue.IsEmpty() {
		zap.S().Infof("Cleaning project %s queue", projectName)
		projectQueue.CleanQueue()
		zap.S().Infof("Finished cleaning project %s queue", projectName)
	}

	projectHandler.QueuesManager.Remove(projectName)
	zap.S().Infof("Removed project queue for project %s", projectName)
}

// HandleAdd seeds the project queue for a kai.resources project (limits read from
// annotations).
func (projectHandler *ProjectHandler) HandleAdd(ctx context.Context, project *kaiorg.Project) error {
	if project == nil {
		// this should not happen - but just in case
		return errors.New("project is nil")
	}
	if project.Status.Namespace == "" {
		zap.S().Warnf("Project %s does not have namespace yet, skipping adding", project.Name)
		return nil
	}

	projectQueue := projectHandler.QueuesManager.Get(project.Name)

	// in case of no rules, clean the project queue if exists and skip
	if !util.ProjectHasLimits(project) {
		zap.S().Infof("Project %s has no rules, skipping", project.Name)

		if projectQueue != nil {
			projectHandler.resetProjectQueue(project.Name, projectQueue)
		}

		return nil
	}

	if projectQueue == nil {
		zap.S().Debugf("Project %s not found in the QueuesManager, adding it", project.Name)
		projectQueue = projectHandler.QueuesManager.Add(ctx, projectHandler.workloadCache, projectHandler.K8sClient,
			projectHandler.PrometheusQuerier, project, projectHandler.Recorder, projectHandler.EventList,
			projectHandler.workloadValidator)
		if projectQueue == nil {
			return fmt.Errorf("project %s not found in the QueuesManager, and failed to add it", project.Name)
		}
	}

	zap.S().Infof(
		"Project (%s) values changed, InteractiveJobMaxIdle (%v) sec, InteractiveJobTimeLimit (%v) sec, "+
			"TrainingJobMaxIdleDuration (%v) sec, TrainingJobTimeLimit (%v) sec, "+
			"InteractivePreemptibleJobMaxIdleDurationSecs (%v)",
		project.Name, util.InteractiveJobMaxIdleDurationSecs(project), util.InteractiveJobTimeLimitSecs(project),
		util.TrainingJobMaxIdleDurationSecs(project), util.TrainingJobTimeLimitSecs(project),
		util.InteractivePreemptibleJobMaxIdleDurationSecs(project))

	// list all workloads in the project
	// add them to the queue again with the update fields
	err := queuesManager.PushProjectWorkloadsToProjectQueue(ctx, projectHandler.workloadCache, projectHandler.K8sClient,
		projectHandler.PrometheusQuerier, project.Name, project.Status.Namespace, projectQueue, projectHandler.supportedJobKinds)
	if err != nil {
		zap.S().Errorf("error while try to init project queue with project workloads, err: %+v ", err)
		return err
	}
	return nil
}

func isProjectChanged(project *kaiorg.Project, projectQueue *queuesManager.ProjectQueue) bool {
	if projectQueue.InteractiveJobMaxIdleDurationSecs == util.InteractiveJobMaxIdleDurationSecs(project) &&
		projectQueue.InteractiveJobTimeLimitSecs == util.InteractiveJobTimeLimitSecs(project) &&
		projectQueue.TrainingJobMaxIdleDurationSecs == util.TrainingJobMaxIdleDurationSecs(project) &&
		projectQueue.TrainingJobTimeLimitSecs == util.TrainingJobTimeLimitSecs(project) &&
		projectQueue.InteractivePreemptibleJobMaxIdleDurationSecs == util.InteractivePreemptibleJobMaxIdleDurationSecs(project) {
		return false
	}
	return true
}
