/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	"context"
	"fmt"
	"time"

	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	dromosv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/common"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/run-ai/runai/runai-cluster/cluster/runtime-rules/pkg/util"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
)

// AutoDeleteCompletedHandler implements the handling of workload auto-delete and is used by AutoDeleteCompletedController.
// it maintains a queue of workloads to be deleted, and workers that consume this queue and delete the workloads
type AutoDeleteCompletedHandler struct {
	K8sClient  client.Client
	recorder   record.EventRecorder
	queue      workqueue.DelayingInterface //nolint:staticcheck
	numWorkers int
}

// WorkloadKey holds a queue item, i.e. an unique identifier of a workload that is scheduled for deletion
type WorkloadKey struct {
	kind      string
	namespace string
	name      string
}

// NewAutoDeleteCompletedHandler returns a new AutoDeleteCompletedHandler struct
func NewAutoDeleteCompletedHandler(k8sClient client.Client, recorder record.EventRecorder) *AutoDeleteCompletedHandler {
	return &AutoDeleteCompletedHandler{
		K8sClient:  k8sClient,
		recorder:   recorder,
		queue:      workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "autoDeleteCompleted"), //nolint:staticcheck
		numWorkers: util.GetEnvOrDefaultInt(util.NumberOfWorkersAutoDeleteCompletedEnv, util.NumberOfWorkersAutoDeleteCompletedDefault),
	}
}

// Run runs the worker goroutines
func (autoDeleteCompletedHandler *AutoDeleteCompletedHandler) Run(ctx context.Context) {
	defer runtime.HandleCrash()
	defer autoDeleteCompletedHandler.queue.ShutDown()

	zap.S().Infof("AutoDeleteCompleted handler has started with %d workers", autoDeleteCompletedHandler.numWorkers)

	for i := 0; i < autoDeleteCompletedHandler.numWorkers; i++ {
		go wait.UntilWithContext(ctx, autoDeleteCompletedHandler.worker, time.Second)
	}

	<-ctx.Done()
}

// worker implements the worker loop of processing queue items
func (autoDeleteCompletedHandler *AutoDeleteCompletedHandler) worker(ctx context.Context) {
	for {
		autoDeleteCompletedHandler.processNextWorkItem(ctx)
	}
}

// processNextWorkItem gets an item from the queue and calls deleteWorkload with it
func (autoDeleteCompletedHandler *AutoDeleteCompletedHandler) processNextWorkItem(ctx context.Context) bool {
	key, quit := autoDeleteCompletedHandler.queue.Get()
	if quit {
		return false
	}
	defer autoDeleteCompletedHandler.queue.Done(key)

	autoDeleteCompletedHandler.deleteWorkload(ctx, key.(WorkloadKey))

	return true
}

// getDeletionTimestamp returns the DeletionTimestamp field of the workload, according to its specific kind
func (autoDeleteCompletedHandler *AutoDeleteCompletedHandler) getDeletionTimestamp(key WorkloadKey, workload client.Object) *metav1.Time {
	if key.kind != constants.KindWorkloadRunner {
		return nil
	}
	return workload.(*dromosv1alpha1.WorkloadRunner).DeletionTimestamp
}

// getPhase returns the Status.Phase field of the workload, according to its specific kind
func (autoDeleteCompletedHandler *AutoDeleteCompletedHandler) getPhase(key WorkloadKey, workload client.Object) string {
	if key.kind == constants.KindWorkloadRunner {
		return string(workload.(*dromosv1alpha1.WorkloadRunner).Status.Phase)
	}
	return ""
}

// getWorkload uses the k8s client to get a workload by its key (kind, namespace and name)
func (autoDeleteCompletedHandler *AutoDeleteCompletedHandler) getWorkload(ctx context.Context, key WorkloadKey) (client.Object, error) {
	if key.kind != constants.KindWorkloadRunner {
		return nil, fmt.Errorf("unsupported workload kind %s", key.kind)
	}

	var workload client.Object = &dromosv1alpha1.WorkloadRunner{}
	err := autoDeleteCompletedHandler.K8sClient.Get(ctx, types.NamespacedName{Name: key.name, Namespace: key.namespace}, workload)

	return workload, err
}

// deleteWorkload performs the resource deletion, by getting the workload, verifying it's not already deleted and finally
// deleting it using the k8s client
func (autoDeleteCompletedHandler *AutoDeleteCompletedHandler) deleteWorkload(ctx context.Context, key WorkloadKey) {
	zap.S().Infof("AutoDeleteCompleted handler about to delete workload %v", key)

	// get the workload object. it's legit to get an error here, as it means the workload has been deleted manually
	// before reaching the auto-delete time
	workload, err := autoDeleteCompletedHandler.getWorkload(ctx, key)
	if err != nil {
		zap.S().Warnf("AutoDeleteCompleted handler failed to get workload for deletion %v (might have been already deleted manually): %v", key, err)
		return
	}

	// verify that the workload wasn't deleted already
	if autoDeleteCompletedHandler.getDeletionTimestamp(key, workload) != nil {
		zap.S().Debugf("AutoDeleteCompleted handler workload %v was already deleted", key)
		return
	}

	// verify the workload is still in a finite phase. Status.CompletionTime is sticky
	// (see workload_status_calculator.go), so a workload that completed and then transitioned
	// back to a running phase (e.g. via the restart flow) would otherwise be deleted here.
	if phase := autoDeleteCompletedHandler.getPhase(key, workload); !common.IsFinitePhase(phase) {
		zap.S().Infof("AutoDeleteCompleted handler skipping deletion of %v: phase is %q, no longer finite", key, phase)
		autoDeleteCompletedHandler.recorder.Eventf(workload, corev1.EventTypeNormal, "AutoDeleteSkipped",
			"auto-delete skipped: phase is %q, no longer finite", phase)
		return
	}

	err = autoDeleteCompletedHandler.K8sClient.Delete(ctx, workload)
	if err != nil {
		zap.S().Warnf("AutoDeleteCompletedHandler failed to delete workload %v: %v", key, err)
		return
	}

	zap.S().Infof("AutoDeleteCompleted handler workload %v deleted successfully", key)
}

// HandleAddedOrModified is the main entry point of the handler, called with an event type and the
// JobWorkload projection of a WorkloadRunner.
// it first checks if there's anything to do, considering the event type and the workload fields, namely:
// DeletionTimestamp, CompletionTime and AutoDeletionTimeAfterCompletionSeconds.
// then, if deletion is relevant, it calculates the remaining time and adds the item (key) to the queue -
// either for immediate deletion or a future one
func (autoDeleteCompletedHandler *AutoDeleteCompletedHandler) HandleAddedOrModified(eventType watch.EventType, jobWorkload jobs.JobWorkload) {
	workloadKey := WorkloadKey{jobWorkload.Kind, jobWorkload.Namespace, jobWorkload.Name}
	if eventType != watch.Added && eventType != watch.Modified {
		zap.S().Debugf("AutoDeleteCompleted handler skipping %s event for %s", eventType, workloadKey)
		return
	}

	zap.S().Debugf("AutoDeleteCompleted handler handling %s event for %s", eventType, workloadKey)

	// check workload conditions before handling: not deleted, completed (and still in a
	// finite phase — CompletionTime is sticky across restarts), has auto-delete setting
	if jobWorkload.DeletionTimestamp != nil ||
		jobWorkload.Status.CompletionTime == nil ||
		jobWorkload.Spec.AutoDeletionTimeAfterCompletionSeconds == nil ||
		!common.IsFinitePhase(jobWorkload.Status.WorkloadCommonStatus.Phase) {
		return
	}

	// calculate the time left before deletion
	elapsed := time.Since(jobWorkload.Status.CompletionTime.Time)
	autoDeleteDuration := time.Duration(jobWorkload.Spec.AutoDeletionTimeAfterCompletionSeconds.Deref32()) * time.Second
	remaining := autoDeleteDuration - elapsed

	// workload should be deleted ASAP as it's completed and auto-delete time has passed (or was set to 0 in the first place)
	if remaining <= 0 {
		zap.S().Infof("AutoDeleteCompleted handler adding %v to the queue for immediate deletion", workloadKey)
		autoDeleteCompletedHandler.queue.Add(workloadKey)
	} else {
		zap.S().Infof("AutoDeleteCompleted handler adding %v to the queue for future deletion %s from now", workloadKey, remaining.String())
		autoDeleteCompletedHandler.queue.AddAfter(workloadKey, remaining)
	}

	// note: upon workload completion, 2 'modified' events occur, almost together, so two identical keys are handled,
	// but only one is actually added to the queue (no duplicates) and is eventually processed
}
