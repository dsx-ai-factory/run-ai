/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package controllers

import (
	"context"

	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	dromosv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/runtime-rules/pkg/handlers"
	"github.com/run-ai/runai/runai-cluster/cluster/runtime-rules/pkg/util"
	dw "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedworkload/v2alpha1"
	iw "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/interactiveworkload/v2alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	tw "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/trainingworkload/v2alpha1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"go.uber.org/zap"
)

// AutoDeleteCompletedController is responsible to delete finished workloads, after a specified duration has passed.
// this duration setting is a part of the workload spec.
// the auto-deletion is implemented by watching all the relevant workload resources and handling the events
type AutoDeleteCompletedController struct {
	autoDeleteCompletedHandler *handlers.AutoDeleteCompletedHandler
	K8sClientWithWatch         client.WithWatch
	workloadRunnerEnabled      bool
	legacyWorkloadsEnabled     bool
	workloadRunnerConverter    *util.RunnerConverter
}

// NewAutoDeleteCompletedController returns a new AutoDeleteCompletedController struct, and runs the
// AutoDeleteCompletedHandler of it. WorkloadRunner is present for every 3.0+
// runtime-rules instance; before 3.0 its presence follows the existing
// WorkloadRunner disable setting (see runtime_rules.go).
func NewAutoDeleteCompletedController(
	ctx context.Context,
	k8sClient client.WithWatch,
	recorder record.EventRecorder,
	supportedJobKinds map[string]client.Object,
	legacyWorkloadsEnabled bool,
) *AutoDeleteCompletedController {
	_, workloadRunnerEnabled := supportedJobKinds[constants.KindWorkloadRunner]
	autoDeleteCompletedController := &AutoDeleteCompletedController{
		autoDeleteCompletedHandler: handlers.NewAutoDeleteCompletedHandler(k8sClient, recorder),
		K8sClientWithWatch:         k8sClient,
		workloadRunnerEnabled:      workloadRunnerEnabled,
		legacyWorkloadsEnabled:     legacyWorkloadsEnabled,
	}
	if workloadRunnerEnabled {
		autoDeleteCompletedController.workloadRunnerConverter = util.NewRunnerConverter(k8sClient)
	}

	go autoDeleteCompletedController.autoDeleteCompletedHandler.Run(ctx)

	return autoDeleteCompletedController
}

// initWorkloadWatchers returns a map from workload kind (e.g. InteractiveWorkload) to a Watch object
// that allows tracking add/modify/delete operations on the resource
func (autoDeleteCompletedController *AutoDeleteCompletedController) initWorkloadWatchers(ctx context.Context) map[string]watch.Interface {
	objLists := make(map[string]client.ObjectList)
	watchers := make(map[string]watch.Interface)

	if autoDeleteCompletedController.legacyWorkloadsEnabled {
		objLists[constants.KindInteractiveWorkload] = &iw.InteractiveWorkloadList{}
		objLists[constants.KindTrainingWorkload] = &tw.TrainingWorkloadList{}
		objLists[constants.KindDistributedWorkload] = &dw.DistributedWorkloadList{}
	}
	if autoDeleteCompletedController.workloadRunnerEnabled {
		objLists[constants.KindWorkloadRunner] = &dromosv1alpha1.WorkloadRunnerList{}
	}

	for key, objList := range objLists {
		watchObj, err := autoDeleteCompletedController.K8sClientWithWatch.Watch(ctx, objList)
		if err != nil {
			zap.S().Fatalf("unable to watch changes for %s objects. Error: %+v", key, err)
		}
		watchers[key] = watchObj
	}

	return watchers
}

// stopWatchChannels stops all the watches in the given map.
// one or more of the watches may be already stopped, and it's fine
func (autoDeleteCompletedController *AutoDeleteCompletedController) stopWatchChannels(watches map[string]watch.Interface) {
	for _, watch := range watches {
		watch.Stop()
	}
}

// Start starts the controller by first initializing watchers for all supported workload kinds,
// and then waiting on all of their channels (plus the given stop channel), using select.
// restarts the controller by calling itself in case any of the channels is closed.
// calls the handler on workload events of types added and modified, ignores the rest
func (autoDeleteCompletedController *AutoDeleteCompletedController) Start(ctx context.Context, stopCh <-chan struct{}) {
	zap.S().Debug("AutoDeleteCompleted controller has started")

	// init watchers and defer their stop, for the case of restarting this function
	watchers := autoDeleteCompletedController.initWorkloadWatchers(ctx)
	defer autoDeleteCompletedController.stopWatchChannels(watchers)

	// A receive on a nil channel blocks forever, so disabled workload kinds have inert select cases.
	var interactiveWorkloadCh <-chan watch.Event
	var trainingWorkloadCh <-chan watch.Event
	var distributedWorkloadCh <-chan watch.Event
	if w, ok := watchers[constants.KindInteractiveWorkload]; ok {
		interactiveWorkloadCh = w.ResultChan()
	}
	if w, ok := watchers[constants.KindTrainingWorkload]; ok {
		trainingWorkloadCh = w.ResultChan()
	}
	if w, ok := watchers[constants.KindDistributedWorkload]; ok {
		distributedWorkloadCh = w.ResultChan()
	}
	var workloadRunnerCh <-chan watch.Event
	if w, ok := watchers[constants.KindWorkloadRunner]; ok {
		workloadRunnerCh = w.ResultChan()
	}

	for {
		select {
		case event, ok := <-interactiveWorkloadCh:
			if !ok {
				autoDeleteCompletedController.handleStoppedController(ctx, stopCh, constants.KindInteractiveWorkload)
				return
			}
			autoDeleteCompletedController.handleWorkloadEvent(event)
		case event, ok := <-trainingWorkloadCh:
			if !ok {
				autoDeleteCompletedController.handleStoppedController(ctx, stopCh, constants.KindTrainingWorkload)
				return
			}
			autoDeleteCompletedController.handleWorkloadEvent(event)
		case event, ok := <-distributedWorkloadCh:
			if !ok {
				autoDeleteCompletedController.handleStoppedController(ctx, stopCh, constants.KindDistributedWorkload)
				return
			}
			autoDeleteCompletedController.handleWorkloadEvent(event)
		case event, ok := <-workloadRunnerCh:
			if !ok {
				autoDeleteCompletedController.handleStoppedController(ctx, stopCh, constants.KindWorkloadRunner)
				return
			}
			autoDeleteCompletedController.handleWorkloadEvent(event)
		case <-stopCh:
			zap.S().Infof("AutoDeleteCompleted controller has stopped")
			return
		}
	}
}

func (autoDeleteCompletedController *AutoDeleteCompletedController) handleStoppedController(ctx context.Context, stopCh <-chan struct{}, kind string) {
	go autoDeleteCompletedController.Start(ctx, stopCh)
	zap.S().Infof("channel of type %s of the AutoDeleteCompleted controller has closed, starting a new one", kind)
}

func (autoDeleteCompletedController *AutoDeleteCompletedController) handleWorkloadEvent(event watch.Event) {
	var jobWorkload jobs.JobWorkload

	switch workloadObject := event.Object.(type) {
	case *iw.InteractiveWorkload:
		workloadObject.CopyTo(&jobWorkload)
	case *tw.TrainingWorkload:
		workloadObject.CopyTo(&jobWorkload)
	case *dw.DistributedWorkload:
		workloadObject.CopyTo(&jobWorkload)
	case *dromosv1alpha1.WorkloadRunner:
		category, skipReason, ok := util.ClassifyRunner(workloadObject)
		if !ok && !util.IsPhaseSkip(skipReason) {
			zap.S().Debugf("auto-delete: skipping out-of-scope runner %s/%s: %s", workloadObject.Namespace, workloadObject.Name, skipReason)
			return
		}
		zap.S().Debugf("auto-delete %s/%s in scope at runner-event entry: category=%s", workloadObject.Namespace, workloadObject.Name, category)

		if err := autoDeleteCompletedController.workloadRunnerConverter.Convert(workloadObject, &jobWorkload); err != nil {
			zap.S().Warnf("auto-delete: failed to convert runner %s/%s: %v", workloadObject.Namespace, workloadObject.Name, err)
			return
		}
	default:
		zap.S().Warnf("unknown type %s, skipping", workloadObject.GetObjectKind())
		return
	}

	autoDeleteCompletedController.autoDeleteCompletedHandler.HandleAddedOrModified(event.Type, jobWorkload)
}
