/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package queuesManager

import (
	"context"
	"fmt"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/runtime-rules/pkg/prometheus"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/runtime-rules/pkg/util"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/runtime-rules/pkg/validation"
	kaiorg "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"go.uber.org/zap"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ProjectQueue struct {
	NumberOfWorkersPerQueue                      int // specifies how many workers should be spawned per queue
	WorkerSleepSeconds                           int // specifies how long the worker should sleep before checking for new work
	ProjectName                                  string
	queue                                        *PriorityQueue
	Workers                                      []*Worker
	InteractiveJobMaxIdleDurationSecs            int64
	InteractivePreemptibleJobMaxIdleDurationSecs int64
	TrainingJobMaxIdleDurationSecs               int64
	InteractiveJobTimeLimitSecs                  int64
	TrainingJobTimeLimitSecs                     int64

	K8sClient         client.Client // this is nil when trying to execute a workload task
	PromQuerier       prometheus.Querier
	Recorder          record.EventRecorder
	EventList         []int
	WorkloadValidator *validation.WorkloadValidator
}

func (pqm *ProjectQueue) CleanQueue() {
	zap.S().Debugf("start clean Queue for project: %s", pqm.ProjectName)
	pqm.queue.Clean()
	zap.S().Debugf("end clean Queue for project: %s", pqm.ProjectName)
}

func (pqm *ProjectQueue) PushToQueue(task *WorkloadTask) {
	zap.S().Debugf("start task push for project: %s", pqm.ProjectName)
	pqm.queue.Push(task)
	zap.S().Debugf("end task push for project: %s", pqm.ProjectName)
}

func (pqm *ProjectQueue) IsEmpty() bool {
	if pqm.queue != nil && pqm.queue.Len() > 0 {
		return false
	}
	return true
}

// QueuesManager is a map of project name to project queue
// [p1] -> project queue { projectName, queue, workers ... }
// [p2] -> project queue { projectName, queue, workers, ... }
type QueuesManager map[string]*ProjectQueue

func (qm QueuesManager) Get(projectName string) *ProjectQueue {
	return qm[projectName]
}

func (qm QueuesManager) Remove(projectName string) {
	delete(qm, projectName)
}

// Update refreshes the queue's limits from a kai.resources project (limits read from annotations).
func (qm QueuesManager) Update(project *kaiorg.Project) {
	currentProject := qm[project.Name]
	if currentProject == nil {
		zap.S().Warnf("project %s not found in the QueuesManager", project.Name)
		return
	}

	currentProject.InteractiveJobMaxIdleDurationSecs = util.InteractiveJobMaxIdleDurationSecs(project)
	currentProject.InteractivePreemptibleJobMaxIdleDurationSecs = util.InteractivePreemptibleJobMaxIdleDurationSecs(project)
	currentProject.TrainingJobMaxIdleDurationSecs = util.TrainingJobMaxIdleDurationSecs(project)
	currentProject.InteractiveJobTimeLimitSecs = util.InteractiveJobTimeLimitSecs(project)
	currentProject.TrainingJobTimeLimitSecs = util.TrainingJobTimeLimitSecs(project)
}

// Add registers a queue for a kai.resources project, taking its limits from annotations.
func (qm QueuesManager) Add(ctx context.Context, workloadCache *util.WorkloadCache, k8sClient client.Client,
	promQuerier prometheus.Querier, project *kaiorg.Project, recorder record.EventRecorder, eventList []int, workloadValidator *validation.WorkloadValidator) *ProjectQueue {
	if project == nil {
		zap.S().Errorf("project is nil, cannot add project to the QueuesManager")
		return nil
	}

	projectQueue := &ProjectQueue{
		NumberOfWorkersPerQueue: util.GetEnvOrDefaultInt(util.NumberOfWorkersPerQueueEnv, util.NumberOfWorkersPerQueueDefault),
		WorkerSleepSeconds:      util.GetEnvOrDefaultInt(util.WorkerSleepSecondsEnv, util.WorkerSleepSecondsDefault),
		ProjectName:             project.Name,
		queue:                   NewPriorityQueue(),
		InteractivePreemptibleJobMaxIdleDurationSecs: util.InteractivePreemptibleJobMaxIdleDurationSecs(project),
		InteractiveJobMaxIdleDurationSecs:            util.InteractiveJobMaxIdleDurationSecs(project),
		TrainingJobMaxIdleDurationSecs:               util.TrainingJobMaxIdleDurationSecs(project),
		InteractiveJobTimeLimitSecs:                  util.InteractiveJobTimeLimitSecs(project),
		TrainingJobTimeLimitSecs:                     util.TrainingJobTimeLimitSecs(project),
		K8sClient:                                    k8sClient,
		PromQuerier:                                  promQuerier,
		Recorder:                                     recorder,
		EventList:                                    eventList,
		WorkloadValidator:                            workloadValidator,
	}

	return qm.startQueue(ctx, workloadCache, projectQueue)
}

// startQueue spawns the queue's workers and registers it in the manager. Shared by the
// kai and legacy Add twins, which differ only in where they read the project's limits.
func (qm QueuesManager) startQueue(ctx context.Context, workloadCache *util.WorkloadCache, projectQueue *ProjectQueue) *ProjectQueue {
	WorkersPointer := make([]*Worker, 0)

	numWorkers := projectQueue.NumberOfWorkersPerQueue

	for i := 1; i <= numWorkers; i++ {
		newWorker := Worker{
			id:            fmt.Sprintf("%s-%d", projectQueue.ProjectName, i),
			projectQueue:  projectQueue,
			queue:         projectQueue.queue,
			sleepTime:     projectQueue.WorkerSleepSeconds,
			ctx:           ctx,
			workloadCache: workloadCache,
		}
		go newWorker.Start()
		WorkersPointer = append(projectQueue.Workers, &newWorker)
	}

	projectQueue.Workers = WorkersPointer
	qm[projectQueue.ProjectName] = projectQueue

	return projectQueue
}
