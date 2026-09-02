/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package controllers

import (
	"context"

	kaiorg "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/runtime-rules/pkg/handlers"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/runtime-rules/pkg/prometheus"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/runtime-rules/pkg/queuesManager"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/runtime-rules/pkg/util"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/runtime-rules/pkg/validation"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ProjectController struct {
	projectHandler     *handlers.ProjectHandler
	K8sClientWithWatch client.WithWatch
}

func NewProjectController(k8sClient client.Client, k8sClientWithWatch client.WithWatch, qm *queuesManager.QueuesManager, promQuerier prometheus.Querier, workloadCache *util.WorkloadCache,
	recorder record.EventRecorder, eventList []int, workloadValidator *validation.WorkloadValidator, supportedJobKinds map[string]client.Object) *ProjectController {
	projectController := &ProjectController{
		projectHandler:     handlers.NewProjectHandler(k8sClient, qm, promQuerier, workloadCache, recorder, eventList, workloadValidator, supportedJobKinds),
		K8sClientWithWatch: k8sClientWithWatch,
	}
	return projectController
}

func (controller *ProjectController) Start(ctx context.Context, stopCh <-chan struct{}) {
	zap.S().Infof("ProjectController start")

	var projectList client.ObjectList = &kaiorg.ProjectList{}
	projectWatcher, err := controller.K8sClientWithWatch.Watch(ctx, projectList)
	if err != nil {
		zap.S().Fatalf("unable to watch changes: %+v", err)
	}

loop:
	for {
		select {
		case event, ok := <-projectWatcher.ResultChan():
			if !ok {
				go controller.Start(ctx, stopCh)
				zap.S().Infof("channel of the project controller has closed, starting a new one")
				return
			}
			controller.handleEvent(ctx, event.Type, event.Object)
		case <-stopCh:
			projectWatcher.Stop()
			break loop
		}
	}
}

func (controller *ProjectController) handleEvent(ctx context.Context, eventType watch.EventType, object runtime.Object) {
	switch project := object.(type) {
	case *kaiorg.Project:
		switch eventType {
		case watch.Added:
			zap.S().Infof("projectHandler (%s) - Added ", project.Name)
			genericControllerErrorHandler(controller.projectHandler.HandleAdd(ctx, project), "update", util.ProjectResource)
		case watch.Modified:
			zap.S().Infof("projectHandler (%s) - Modified ", project.Name)
			genericControllerErrorHandler(controller.projectHandler.HandleUpdate(ctx, project), "add", util.ProjectResource)
		}
	default:
	}
}
