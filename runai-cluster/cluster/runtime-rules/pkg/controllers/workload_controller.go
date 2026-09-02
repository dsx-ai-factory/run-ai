/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package controllers

import (
	"context"
	"fmt"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/runtime-rules/pkg/handlers"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/runtime-rules/pkg/util"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/watch"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type WorkloadController struct {
	workloadHandler    *handlers.WorkloadHandler
	K8sClientWithWatch client.WithWatch
	workloadKind       string
}

func NewWorkloadController(k8sClient client.WithWatch, workloadHandler *handlers.WorkloadHandler, workloadKind string) *WorkloadController {
	workloadController := &WorkloadController{
		workloadHandler:    workloadHandler,
		K8sClientWithWatch: k8sClient,
		workloadKind:       workloadKind,
	}

	return workloadController
}

func (controller *WorkloadController) Start(ctx context.Context, stopCh <-chan struct{}) {
	zap.S().Infof(fmt.Sprintf("%s Controller start", controller.workloadKind))
	workloadList, err := controller.getObjectList()
	if err != nil {
		zap.S().Fatalf("unable to init workload list: %+v", err)
	}

	workloadWatcher, err := controller.K8sClientWithWatch.Watch(ctx, workloadList)
	if err != nil {
		zap.S().Fatalf("unable to watch changes: %+v", err)
	}

loop:
	for {
		select {
		case event, ok := <-workloadWatcher.ResultChan():
			if !ok {
				go controller.Start(ctx, stopCh)
				zap.S().Infof(fmt.Sprintf("Channel of the %s Controller was closed, starting a new one", controller.workloadKind))
				return
			}
			object := event.Object
			eventType := event.Type

			runner, ok := object.(*dromosv1alpha1.WorkloadRunner)
			if !ok {
				zap.S().Warnf("%s Controller received unsupported object type %T", controller.workloadKind, object)
				continue
			}

			switch eventType {
			case watch.Added:
				controllerErrorHandler(controller.workloadHandler.HandleRunnerEvent(ctx, runner), "add", constants.KindWorkloadRunner)
			case watch.Modified:
				controllerErrorHandler(controller.workloadHandler.HandleRunnerEvent(ctx, runner), "update", constants.KindWorkloadRunner)
			default:
				zap.S().Debugf("WorkloadHandler (%s) , (%s) - Deleted ", controller.workloadKind, runner.Name)
			}
		case <-stopCh:
			workloadWatcher.Stop()
			break loop
		}
	}
}

func controllerErrorHandler(err error, operation, resourceName string) {
	if err != nil {
		zap.S().Errorf("Had an error while trying to %s %s: %v ", operation, resourceName, err)
	}
}

func (controller *WorkloadController) getObjectList() (client.ObjectList, error) {
	switch controller.workloadKind {
	case constants.KindWorkloadRunner:
		return &dromosv1alpha1.WorkloadRunnerList{}, nil
	default:
		return nil, fmt.Errorf(util.UnknownKindMsg, controller.workloadKind)
	}
}
