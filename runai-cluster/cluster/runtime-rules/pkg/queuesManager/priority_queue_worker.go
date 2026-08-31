/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package queuesManager

import (
	"context"
	"time"

	"github.com/run-ai/runai/runai-cluster/cluster/runtime-rules/pkg/util"
	"go.uber.org/zap"
)

// KillWorker
// we used this "KillWorker" to test the queues manager
// we can use this to kill the worker between tests
var (
	KillWorker = false
)

// Worker represents a worker that executes workloads from the priority queue.
type Worker struct {
	id            string
	queue         *PriorityQueue
	projectQueue  *ProjectQueue
	sleepTime     int
	workloadCache *util.WorkloadCache
	ctx           context.Context
}

// Start starts the worker, continuously executing workloads from the queue.
func (w *Worker) Start() {
	for { // Infinite loop
		if KillWorker {
			// used for test only
			zap.S().Debugf("* worker id (%s), KillWorker is true, return ", w.id)
			return
		}
		workloadTask := w.getNextJob(w.ctx)
		w.executeJob(w.ctx, workloadTask)
	}
}

// getQueuesManager returns the priority queue of the worker.
func (w *Worker) getProjectQueue() *ProjectQueue {
	return w.projectQueue
}

// getNextJob retrieves the next workload from the priority queue.
// Wait until there is a workload available
func (w *Worker) getNextJob(ctx context.Context) *WorkloadTask {
	for {
		// used for test only
		if KillWorker {
			zap.S().Debugf("* worker id (%s), KillWorker is true, return nil ", w.id)
			return nil
		}

		workerQueue := w.queue
		workerId := w.id
		workerSleepTime := w.sleepTime

		if workerQueue.Len() > 0 {
			// Retrieve the next workload from the queue
			// and Unlock the queue
			workloadTask, done := w.handleWorkloadQueueTask(ctx, workerQueue, workerId, workerSleepTime)
			if done {
				return workloadTask
			}
		} else {
			time.Sleep(time.Duration(workerSleepTime) * time.Second)
			continue
		}
	}
}

// handleWorkloadQueueTask retrieves the next workload from the priority queue.
func (w *Worker) handleWorkloadQueueTask(ctx context.Context, workerQueue *PriorityQueue, workerId string, workerSleepTime int) (*WorkloadTask, bool) {
	_, cancel := context.WithCancel(ctx)
	defer func() {
		time.Sleep(time.Duration(workerSleepTime) * time.Second)
		cancel()
	}()

	poppedWorkloadTask := workerQueue.PopWorkloadTask()
	if poppedWorkloadTask != nil {
		zap.S().Debugf("* worker id (%s), got next job (%s) ", workerId, poppedWorkloadTask.WorkloadName)
		if poppedWorkloadTask.NextEventAt.After(time.Now()) {
			workerQueue.Push(poppedWorkloadTask)
			zap.S().Debugf("* worker id (%s), unlocked queue , workloadTask not ready yet WorkloadName: (%s) - NextEventAt (%s)", workerId, poppedWorkloadTask.WorkloadName, poppedWorkloadTask.NextEventAt)
			return nil, false
		}
		return poppedWorkloadTask, true
	}
	return nil, false
}

// executeJob executes the given workload.
func (w *Worker) executeJob(ctx context.Context, workloadTask *WorkloadTask) {
	if workloadTask == nil {
		zap.S().Debugf("* worker id (%s), <start...> executeJob (nil) ", w.id)
		return
	}
	zap.S().Debugf("* worker id (%s), <start...> executeJob (%s) ", w.id, workloadTask.WorkloadName)

	projectQueue := w.getProjectQueue()

	if projectQueue == nil {
		zap.S().Errorf("project queue is nil for project %s", workloadTask.ProjectName)
		return
	}

	err := HandleWorkloadTask(ctx, w.workloadCache, projectQueue.K8sClient, projectQueue.PromQuerier, workloadTask, projectQueue, true)
	if err != nil {
		zap.S().Errorf("err while handling workload task %+v ", err)
	}

	zap.S().Debugf("* worker id (%s), <done!   >  executeJob (%s) ", w.id, workloadTask.WorkloadName)
}

func SetKillWorker(killStatus bool, sleepTime time.Duration) {
	KillWorker = killStatus
	time.Sleep(sleepTime)
}
