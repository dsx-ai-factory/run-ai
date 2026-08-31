/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	"context"
	"fmt"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/handlers/parsers"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/metrics"

	"go.uber.org/zap"
	"k8s.io/utils/ptr"

	"sync"

	swagger "github.com/run-ai/runai/runai-clients/workloads-client/generated"
	runaiBackendClient "github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/backend-client"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/config"
	global_consts "github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	schedulingv1 "k8s.io/api/scheduling/v1"

	_ "k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type PriorityClassHandler struct {
	resourceName  constants.ClusterResource
	runaiClient   client.Client
	backendClient *runaiBackendClient.BackendClient
}

const deletedStatus = "Deleted"
const readyStatus = "Ready"

func NewPriorityClassHandler(runaiBackendClient *runaiBackendClient.BackendClient, runaiClient client.Client) *PriorityClassHandler {
	return &PriorityClassHandler{
		resourceName:  constants.PriorityClassResourceName,
		runaiClient:   runaiClient,
		backendClient: runaiBackendClient,
	}
}

func (h *PriorityClassHandler) HandleUpdateItem(key client.ObjectKey) error {
	pc := &schedulingv1.PriorityClass{}
	err := h.runaiClient.Get(context.Background(), key, pc)
	if err != nil {
		return err
	}

	return h.HandleAddUpdate(pc)
}

func (h *PriorityClassHandler) HandleAddUpdate(pc *schedulingv1.PriorityClass) error {
	defer metrics.GetMetricsExporter().MeasureExecutionTime(h.resourceName, constants.UpsertOperation)()
	source := pc.Labels[global_consts.PriorityClassSourceLabel]
	prioritiesForStatus := swagger.PrioritiesForStatus{
		FullSync: ptr.To(false),
		Priorities: []swagger.PrioritiesForStatusPrioritiesInner{
			parsers.ParseBackendPriority(pc, source, readyStatus),
		},
	}
	if source == global_consts.PriorityClassRunaiSource {
		prioritiesForStatusRequest := h.backendClient.WorkloadsClient.WorkloadPropertiesAPI.
			WorkloadPrioritiesStatus(h.backendClient.Context()).
			ClusterId(config.Config().ClusterUuid).
			PrioritiesForStatus(prioritiesForStatus)
		res, err := h.backendClient.WorkloadsClient.WorkloadPropertiesAPI.
			WorkloadPrioritiesStatusExecute(prioritiesForStatusRequest)
		return h.backendClient.HandleHttpResponse(res, err, fmt.Sprintf("Upsert priority class %s", pc.Name))
	}
	// else - source is not control plane
	clusterPrioritiesForStatusRequest := h.backendClient.WorkloadsClient.WorkloadPropertiesAPI.
		WorkloadClusterPrioritiesStatus(h.backendClient.Context()).
		ClusterId(config.Config().ClusterUuid).
		PrioritiesForStatus(prioritiesForStatus)
	res, err := h.backendClient.WorkloadsClient.WorkloadPropertiesAPI.
		WorkloadClusterPrioritiesStatusExecute(clusterPrioritiesForStatusRequest)
	return h.backendClient.HandleHttpResponse(res, err, fmt.Sprintf("Upsert priority class %s", pc.Name))
}

func (h *PriorityClassHandler) HandleDelete(pc *schedulingv1.PriorityClass) error {
	defer metrics.GetMetricsExporter().MeasureExecutionTime(h.resourceName, constants.DeleteOperation)()

	source := pc.Labels[global_consts.PriorityClassSourceLabel]
	prioritiesForStatus := []swagger.PrioritiesForStatusPrioritiesInner{
		parsers.ParseBackendPriority(pc, source, deletedStatus),
	}
	if source == global_consts.PriorityClassRunaiSource {
		prioritiesForStatusRequest := h.backendClient.WorkloadsClient.WorkloadPropertiesAPI.
			WorkloadPrioritiesStatus(h.backendClient.Context()).
			ClusterId(config.Config().ClusterUuid).
			PrioritiesForStatus(swagger.PrioritiesForStatus{
				FullSync:   ptr.To(false),
				Priorities: prioritiesForStatus,
			})
		res, err := h.backendClient.WorkloadsClient.WorkloadPropertiesAPI.
			WorkloadPrioritiesStatusExecute(prioritiesForStatusRequest)
		return h.backendClient.HandleHttpResponse(res, err, fmt.Sprintf("Deleted priority class %s", pc.Name))
	}
	clusterPrioritiesForStatusRequest := h.backendClient.WorkloadsClient.WorkloadPropertiesAPI.
		WorkloadClusterPrioritiesStatus(h.backendClient.Context()).
		ClusterId(config.Config().ClusterUuid).
		PrioritiesForStatus(swagger.PrioritiesForStatus{
			FullSync:   ptr.To(false),
			Priorities: prioritiesForStatus,
		})
	res, err := h.backendClient.WorkloadsClient.WorkloadPropertiesAPI.
		WorkloadClusterPrioritiesStatusExecute(clusterPrioritiesForStatusRequest)
	return h.backendClient.HandleHttpResponse(res, err, fmt.Sprintf("Deleted priority class %s", pc.Name))
}

func (h *PriorityClassHandler) HandleSyncEvent(waitGroup *sync.WaitGroup) {
	defer metrics.GetMetricsExporter().MeasureExecutionTime(h.resourceName, constants.SyncOperation)()
	defer waitGroup.Done()
	zap.S().Infof("Start priority class Sync")

	pcs := &schedulingv1.PriorityClassList{}
	kubeErr := h.runaiClient.List(context.Background(), pcs)
	if kubeErr != nil {
		zap.S().Errorf("Failed listing priority classes, error: %s", kubeErr.Error())
		return
	}

	runaiPriorities := []swagger.PrioritiesForStatusPrioritiesInner{}
	clusterPriorities := []swagger.PrioritiesForStatusPrioritiesInner{}
	for _, pc := range pcs.Items {
		source := pc.Labels[global_consts.PriorityClassSourceLabel]
		priority := parsers.ParseBackendPriority(&pc, source, readyStatus)
		if source == global_consts.PriorityClassRunaiSource {
			runaiPriorities = append(runaiPriorities, priority)
		} else {
			clusterPriorities = append(clusterPriorities, priority)
		}
	}
	prioritiesForStatusRequest := h.backendClient.WorkloadsClient.WorkloadPropertiesAPI.
		WorkloadPrioritiesStatus(h.backendClient.Context()).
		ClusterId(config.Config().ClusterUuid).
		PrioritiesForStatus(swagger.PrioritiesForStatus{
			FullSync:   ptr.To(true),
			Priorities: runaiPriorities,
		})

	_, err := h.backendClient.WorkloadsClient.WorkloadPropertiesAPI.
		WorkloadPrioritiesStatusExecute(prioritiesForStatusRequest)
	if err != nil {
		zap.S().Errorf("Failed syncing priority classes: %+v", err)
	}

	clusterPrioritiesForStatusRequest := h.backendClient.WorkloadsClient.WorkloadPropertiesAPI.
		WorkloadClusterPrioritiesStatus(h.backendClient.Context()).
		ClusterId(config.Config().ClusterUuid).
		PrioritiesForStatus(swagger.PrioritiesForStatus{
			FullSync:   ptr.To(true),
			Priorities: clusterPriorities,
		})

	_, err = h.backendClient.WorkloadsClient.WorkloadPropertiesAPI.
		WorkloadClusterPrioritiesStatusExecute(clusterPrioritiesForStatusRequest)
	if err != nil {
		zap.S().Errorf("Failed syncing priority classes: %+v", err)
	}
}
