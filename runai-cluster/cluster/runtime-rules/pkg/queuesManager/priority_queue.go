/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package queuesManager

import (
	"fmt"
	"go.uber.org/zap"
	"reflect"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/types"
)

// WorkloadTask represents a single workloadTask to be executed.
type WorkloadTask struct {
	NextEventAt  time.Time
	StartTime    time.Time
	WorkloadName string
	WorkloadID   types.UID
	ProjectName  string
	Namespace    string
	WorkloadType string
	// Category selects the project limit bucket that governs a WorkloadRunner.
	Category             string
	JobUuid              types.UID
	JobName              string
	TaskAction           ActionList
	WorkloadRequestedGpu bool
	IsPreemptible        bool
}

// PriorityQueue implements a priority queue of workloadTasks.
type PriorityQueue struct {
	mu    sync.RWMutex
	tasks []*WorkloadTask
}

func NewPriorityQueue() *PriorityQueue {
	return &PriorityQueue{}
}

// Len returns the length of the priority queue.
func (pq *PriorityQueue) Len() int {
	pq.mu.RLock()
	defer pq.mu.RUnlock()
	return len(pq.tasks)
}

// Less compares two workloadTasks based on their NextEventAt fields.
func (pq *PriorityQueue) Less(i, j int) bool {
	pq.mu.RLock()
	defer pq.mu.RUnlock()
	return pq.tasks[i].NextEventAt.Before(pq.tasks[j].NextEventAt)
}

// Swap swaps two workloadTasks in the priority queue.
func (pq *PriorityQueue) Swap(i, j int) {
	pq.mu.Lock()
	defer pq.mu.Unlock()
	pq.tasks[i], pq.tasks[j] = pq.tasks[j], pq.tasks[i]
}

// Push adds a new workloadTask to the priority queue.
func (pq *PriorityQueue) Push(x interface{}) {
	pq.mu.Lock()
	defer pq.mu.Unlock()
	item, ok := x.(*WorkloadTask)
	if !ok {
		zap.Error(fmt.Errorf("can't push item that is not of type WorkloadTask, item type: %s", reflect.TypeOf(x)))
		return
	}
	pq.tasks = append(pq.tasks, item)
}

// Pop removes and returns the workloadTask with the highest priority.
func (pq *PriorityQueue) Pop() interface{} {
	pq.mu.Lock()
	defer pq.mu.Unlock()
	qSize := len(pq.tasks)
	if qSize == 0 {
		return nil
	}

	item := pq.tasks[qSize-1]
	pq.tasks = pq.tasks[0 : qSize-1]
	return item
}

func (pq *PriorityQueue) PopWorkloadTask() *WorkloadTask {
	pop := pq.Pop()
	if pop == nil {
		return nil
	}
	workloadTask, _ := pop.(*WorkloadTask)
	return workloadTask
}

// Clean removes all workloadTasks from the priority queue.
func (pq *PriorityQueue) Clean() {
	pqLength := pq.Len()
	for pqLength > 0 {
		pq.Pop()
		pqLength = pq.Len()
	}
}
