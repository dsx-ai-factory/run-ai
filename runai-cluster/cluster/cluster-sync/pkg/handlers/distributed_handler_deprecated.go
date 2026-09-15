/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	swagger "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client/generated"
	backendclient "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/backend-client"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/config"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/handlers/parsers"
	diwv2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/distributedworkload/v2alpha1"
	"go.uber.org/zap"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//TODO: delete and remove usage when new workloads are functioning

type DistributedHandlerDeprecated struct {
	cachedReader       client.Reader
	runaiBackendClient *backendclient.BackendClient
}

func NewDistributedHandlerDeprecated(cachedReader client.Reader, runaiBackendClient *backendclient.BackendClient) *DistributedHandlerDeprecated {
	return &DistributedHandlerDeprecated{
		cachedReader:       cachedReader,
		runaiBackendClient: runaiBackendClient,
	}
}

func (handler *DistributedHandlerDeprecated) HandleUpdateItem(key client.ObjectKey) error {
	dw := &diwv2.DistributedWorkload{}
	err := handler.cachedReader.Get(context.Background(), key, dw)
	if err != nil {
		return err
	}

	return handler.HandleUpdate(dw)
}

func (handler *DistributedHandlerDeprecated) HandleAdd(workload *diwv2.DistributedWorkload) error {
	statusToSend := parsers.WorkloadDistributedToStatusChangeRequest(&workload.ObjectMeta, workload.Status)
	if statusToSend == nil {
		return nil
	}

	_, res, err := handler.runaiBackendClient.Client.TrainingsAPI.UpdateDistributedStatus(
		handler.runaiBackendClient.Context(), config.Config().ClusterUuid,
	).WorkloadStatusUpdate(*statusToSend).Execute()
	if res != nil {
		_ = res.Body.Close()

	}
	if res != nil && res.StatusCode == http.StatusNotFound {
		zap.S().Infof("distributed %s not found in control plane, it has probably removed already from the DB", statusToSend.Id)
		return nil
	}
	return handler.runaiBackendClient.HandleHttpResponse(
		res,
		err,
		fmt.Sprintf("Updated distributed status of %s to %s", statusToSend.Id, statusToSend.Status),
	)
}

func (handler *DistributedHandlerDeprecated) HandleDelete(workload *diwv2.DistributedWorkload) error {
	workloadId := workload.Annotations[parsers.AnnotationWorkloadId]
	if workloadId == "" {
		return nil
	}

	res, err := handler.runaiBackendClient.Client.TrainingsAPI.DeleteDistributedById(
		handler.runaiBackendClient.Context(), config.Config().ClusterUuid, workloadId,
	).Execute()
	if res != nil {
		_ = res.Body.Close()

	}
	if res != nil && res.StatusCode == http.StatusNotFound {
		zap.S().Infof("distributed %s is already removed from control plane", workloadId)
		return nil
	}
	return handler.runaiBackendClient.HandleHttpResponse(
		res,
		err,
		fmt.Sprintf("Deleted distributed %s", workloadId),
	)
}

func (handler *DistributedHandlerDeprecated) HandleUpdate(workload *diwv2.DistributedWorkload) error {
	// updating status of distributed is same as in addition of distributed
	return handler.HandleAdd(workload)
}

func (handler *DistributedHandlerDeprecated) HandleSyncEvent(waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()
	zap.S().Infof("Start distributed Sync")

	workloadList := &diwv2.DistributedWorkloadList{}
	opts := make([]client.ListOption, 0)
	kubeErr := handler.cachedReader.List(context.Background(), workloadList, opts...)
	if kubeErr != nil {
		zap.S().Errorf("%+v", kubeErr)
		return
	}

	syncToSend := make([]swagger.WorkloadStatusUpdate, 0, len(workloadList.Items))

	for _, workload := range workloadList.Items {
		curStatus := parsers.WorkloadDistributedToStatusChangeRequest(&workload.ObjectMeta, workload.Status)
		if curStatus != nil {
			syncToSend = append(syncToSend, *curStatus)
		}
	}

	_, res, err := handler.runaiBackendClient.Client.TrainingsAPI.SyncDistributedStatus(
		handler.runaiBackendClient.Context(), config.Config().ClusterUuid,
	).WorkloadStatusUpdate(syncToSend).Execute()
	if res != nil {
		_ = res.Body.Close()

	}

	err = handler.runaiBackendClient.HandleHttpResponse(res, err, "Sync workloads")
	if err != nil {
		zap.S().Errorf("%+v", err)
	}
}
