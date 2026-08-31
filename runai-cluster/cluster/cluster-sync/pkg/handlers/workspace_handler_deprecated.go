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

	"github.com/antihax/optional"
	backendclient "github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/backend-client"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/config"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/generated/swagger-runai-backend-client"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/handlers/parsers"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/interactiveworkload/v2alpha1"
	"go.uber.org/zap"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//TODO: delete when needed

type WorkspaceHandlerDeprecated struct {
	cachedReader       client.Reader
	runaiBackendClient *backendclient.BackendClient
}

func NewWorkspaceHandlerDeprecated(cachedReader client.Reader, runaiBackendClient *backendclient.BackendClient) *WorkspaceHandlerDeprecated {
	return &WorkspaceHandlerDeprecated{
		cachedReader:       cachedReader,
		runaiBackendClient: runaiBackendClient,
	}
}

func (handler *WorkspaceHandlerDeprecated) HandleUpdateItem(key client.ObjectKey) error {
	iw := &v2alpha1.InteractiveWorkload{}
	err := handler.cachedReader.Get(context.Background(), key, iw)
	if err != nil {
		return err
	}

	return handler.HandleUpdate(iw)
}

func (handler *WorkspaceHandlerDeprecated) HandleAdd(workload *v2alpha1.InteractiveWorkload) error {

	jobStatus := workload.Status.DeepCopyToJobStatus()
	statusToSend := parsers.WorkloadToStatusChangeRequest(&workload.ObjectMeta, jobStatus)
	if statusToSend == nil {
		return nil
	}

	res, err := handler.runaiBackendClient.Client.WorkspacesApi.UpdateWorkspaceStatus(
		handler.runaiBackendClient.Context(),
		*statusToSend,
		config.Config().ClusterUuid,
	)
	if res != nil && res.StatusCode == http.StatusNotFound {
		zap.S().Infof("workspace %s not found in control plane, it has probably removed already from the DB", statusToSend.Id)
		return nil
	}
	return handler.runaiBackendClient.HandleHttpResponse(
		res,
		err,
		fmt.Sprintf("Updated workspace status of %s to %s", statusToSend.Id, statusToSend.Status),
	)
}

func (handler *WorkspaceHandlerDeprecated) HandleDelete(workload *v2alpha1.InteractiveWorkload) error {
	workspaceId := workload.Annotations[parsers.AnnotationWorkloadId]
	if workspaceId == "" {
		return nil
	}

	res, err := handler.runaiBackendClient.Client.WorkspacesApi.DeleteWorkspace(
		handler.runaiBackendClient.Context(),
		config.Config().ClusterUuid,
		workspaceId,
	)
	if res != nil && res.StatusCode == http.StatusNotFound {
		zap.S().Infof("workspace %s is already removed from control plane", workspaceId)
		return nil
	}
	return handler.runaiBackendClient.HandleHttpResponse(
		res,
		err,
		fmt.Sprintf("Deleted workspace %s", workspaceId),
	)
}

func (handler *WorkspaceHandlerDeprecated) HandleUpdate(workload *v2alpha1.InteractiveWorkload) error {
	// updating status of workspace is same as in addition of workspace
	return handler.HandleAdd(workload)
}

func (handler *WorkspaceHandlerDeprecated) HandleSyncEvent(waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()
	zap.S().Infof("Start workspace Sync")

	workloadList := &v2alpha1.InteractiveWorkloadList{}
	opts := make([]client.ListOption, 0)
	kubeErr := handler.cachedReader.List(context.Background(), workloadList, opts...)
	if kubeErr != nil {
		zap.S().Errorf("%+v", kubeErr)
		return
	}

	syncToSend := make([]swagger.WorkloadStatus, 0, len(workloadList.Items))

	for _, workload := range workloadList.Items {
		jobStatus := workload.Status.DeepCopyToJobStatus()
		curStatus := parsers.WorkloadToStatusChangeRequest(&workload.ObjectMeta, jobStatus)
		if curStatus != nil {
			syncToSend = append(syncToSend, *curStatus)
		}
	}

	res, err := handler.runaiBackendClient.Client.WorkspacesApi.SyncWorkspacesStatus(
		handler.runaiBackendClient.Context(),
		config.Config().ClusterUuid,
		&swagger.WorkspacesApiSyncWorkspacesStatusOpts{
			Body: optional.NewInterface(syncToSend),
		},
	)

	err = handler.runaiBackendClient.HandleHttpResponse(res, err, "Sync workloads")
	if err != nil {
		zap.S().Errorf("%+v", err)
	}
}
