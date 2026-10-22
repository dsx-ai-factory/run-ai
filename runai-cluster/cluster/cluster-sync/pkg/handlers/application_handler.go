/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	"context"
	"fmt"
	"sync"

	swagger "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/config"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/handlers/parsers"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/metrics"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	"go.uber.org/zap"
	"k8s.io/utils/ptr"

	runaiBackendClient "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/backend-client"
	_ "k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ApplicationHandler struct {
	resourceName  constants.ClusterResource
	runaiClient   client.Client
	backendClient *runaiBackendClient.BackendClient
}

func NewApplicationHandler(runaiBackendClient *runaiBackendClient.BackendClient, runaiClient client.Client) *ApplicationHandler {
	return &ApplicationHandler{
		resourceName:  constants.ApplicationResourceName,
		runaiClient:   runaiClient,
		backendClient: runaiBackendClient,
	}
}

func (h *ApplicationHandler) HandleUpdateItem(key client.ObjectKey) error {
	application := &v1alpha1.Application{}
	err := h.runaiClient.Get(context.Background(), key, application)
	if err != nil {
		return err
	}

	err = h.HandleAddUpdate(application)
	if err != nil {
		return err
	}

	// The informer drops an object from the cache before calling the delete handler, which doesn't wait for queued
	// updates, so an object that's gone or replaced by now may have had its delete reach the backend before this update.
	current := &v1alpha1.Application{}
	err = h.runaiClient.Get(context.Background(), key, current)
	if client.IgnoreNotFound(err) != nil {
		return err
	}
	if err == nil && parsers.GetApplicationId(current) == parsers.GetApplicationId(application) {
		return nil
	}

	zap.S().Infof("Application %s was deleted while its update was sent - deleting it again", application.Name)
	return h.HandleDelete(application)
}

func (h *ApplicationHandler) HandleAddUpdate(application *v1alpha1.Application) error {
	defer metrics.GetMetricsExporter().MeasureExecutionTime(h.resourceName, constants.UpsertOperation)()
	backendApplication, err := parsers.ParseBackendApplication(application, false)
	if err != nil {
		return err
	}
	request := h.backendClient.WorkloadsClient.AIApplicationsAPI.
		SyncApplications(h.backendClient.Context()).
		ClusterId(config.Config().ClusterUuid).
		AIApplicationsForSync(swagger.AIApplicationsForSync{
			FullSync:     ptr.To(false),
			Applications: []swagger.AIApplicationForSync{backendApplication},
		})
	res, err := h.backendClient.WorkloadsClient.AIApplicationsAPI.SyncApplicationsExecute(request)
	if res != nil {
		_ = res.Body.Close()
	}
	return h.backendClient.HandleHttpResponse(res, err, fmt.Sprintf("Upsert application %s", application.Name))
}

func (h *ApplicationHandler) HandleDelete(application *v1alpha1.Application) error {
	defer metrics.GetMetricsExporter().MeasureExecutionTime(h.resourceName, constants.DeleteOperation)()

	backendApplication, err := parsers.ParseBackendApplication(application, true)
	if err != nil {
		return err
	}
	request := h.backendClient.WorkloadsClient.AIApplicationsAPI.
		SyncApplications(h.backendClient.Context()).
		ClusterId(config.Config().ClusterUuid).
		AIApplicationsForSync(swagger.AIApplicationsForSync{
			FullSync:     ptr.To(false),
			Applications: []swagger.AIApplicationForSync{backendApplication},
		})
	res, err := h.backendClient.WorkloadsClient.AIApplicationsAPI.SyncApplicationsExecute(request)
	if res != nil {
		_ = res.Body.Close()
	}
	return h.backendClient.HandleHttpResponse(res, err, fmt.Sprintf("Deleted application %s", application.Name))
}

func (h *ApplicationHandler) HandleSyncEvent(waitGroup *sync.WaitGroup) {
	defer metrics.GetMetricsExporter().MeasureExecutionTime(h.resourceName, constants.SyncOperation)()
	defer waitGroup.Done()
	zap.S().Infof("Start applications Sync")

	applications := &v1alpha1.ApplicationList{}
	kubeErr := h.runaiClient.List(context.Background(), applications)
	if kubeErr != nil {
		zap.S().Errorf("Failed listing applications, error: %s", kubeErr.Error())
		return
	}

	backendApplications := []swagger.AIApplicationForSync{}
	for _, application := range applications.Items {
		backendApplication, err := parsers.ParseBackendApplication(&application, false)
		if err != nil {
			return
		}
		backendApplications = append(backendApplications, backendApplication)
	}
	request := h.backendClient.WorkloadsClient.AIApplicationsAPI.
		SyncApplications(h.backendClient.Context()).
		ClusterId(config.Config().ClusterUuid).
		AIApplicationsForSync(swagger.AIApplicationsForSync{
			FullSync:     ptr.To(true),
			Applications: backendApplications,
		})
	res, err := h.backendClient.WorkloadsClient.AIApplicationsAPI.SyncApplicationsExecute(request)
	if res != nil {
		_ = res.Body.Close()
	}
	if err != nil {
		zap.S().Errorf("Failed syncing applications: %+v", err)
	}

}
