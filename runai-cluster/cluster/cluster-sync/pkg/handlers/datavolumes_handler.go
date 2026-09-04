/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	"context"
	"fmt"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/metrics"

	"go.uber.org/zap"
	"k8s.io/utils/ptr"

	"net/http"
	"sort"
	"sync"

	swagger "github.com/dsx-ai-factory/run-ai/runai-clients/datavolumes-service-client/generated"
	runaiBackendClient "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/backend-client"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/config"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/handlers/parsers"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"

	_ "k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type DatavolumeHandler struct {
	resourceName                  constants.ClusterResource
	runaiClient                   client.Client
	backendClient                 *runaiBackendClient.BackendClient
	maxNumOfDatavolumesPerRequest int
}

func NewDatavolumeHandler(runaiBackendClient *runaiBackendClient.BackendClient, runaiClient client.Client) *DatavolumeHandler {
	return &DatavolumeHandler{
		resourceName:                  constants.DataVolumeResourceName,
		runaiClient:                   runaiClient,
		backendClient:                 runaiBackendClient,
		maxNumOfDatavolumesPerRequest: config.Config().MaxNumOfDatavolumesPerRequest,
	}
}

func (h *DatavolumeHandler) HandleUpdateItem(key client.ObjectKey) error {
	datavolume := &v1alpha1.DataVolume{}
	err := h.runaiClient.Get(context.Background(), key, datavolume)
	if err != nil {
		return err
	}

	return h.HandleAddUpdate(datavolume)
}

func (h *DatavolumeHandler) HandleAddUpdate(datavolume *v1alpha1.DataVolume) error {
	return h.updateDatavolumeStatus(datavolume)
}

func (h *DatavolumeHandler) HandleDelete(datavolume *v1alpha1.DataVolume) error {
	defer metrics.GetMetricsExporter().MeasureExecutionTime(h.resourceName, constants.DeleteOperation)()
	datavolumeStatusInternal := parsers.ParseDatavolumeToDeleteStatusRequest()

	res, err := h.backendClient.DatavolumesClient.UpdateDatavolumeStatus(
		h.backendClient.Context(),
		datavolume.Spec.DatavolumeId,
		datavolumeStatusInternal,
	)
	return h.backendClient.HandleHttpResponse(res, err, fmt.Sprintf("Deleted datavolume %s", datavolume.Name))
}

func (h *DatavolumeHandler) updateDatavolumeStatus(datavolume *v1alpha1.DataVolume) error {
	datavolumeStatusInternal := parsers.ParseDatavolumeToUpdateStatusRequest(&datavolume.Status)
	defer metrics.GetMetricsExporter().MeasureExecutionTime(h.resourceName, constants.UpdateOperation)()
	res, err := h.backendClient.DatavolumesClient.UpdateDatavolumeStatus(
		h.backendClient.Context(),
		datavolume.Spec.DatavolumeId,
		datavolumeStatusInternal,
	)
	return h.backendClient.HandleHttpResponse(res, err, fmt.Sprintf("Updated datavolume %s", datavolume.Name))
}

func (h *DatavolumeHandler) HandleSyncEvent(waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()
	zap.S().Infof("Start datavolumes Sync")

	datavolumes := &v1alpha1.DataVolumeList{}
	kubeErr := h.runaiClient.List(context.Background(), datavolumes)
	if kubeErr != nil {
		zap.S().Errorf("%+v", kubeErr)
		return
	}

	dvsToSync := parsers.ParseDatavolumesToStatusesForSync(datavolumes.Items)
	if len(dvsToSync) == 0 {
		zap.S().Infof("No Datavolumes to sync")
		return
	}

	err := h.handleSyncWithPagination(dvsToSync)
	if err != nil {
		zap.S().Errorf("Had an error while syncing datavolumes, error: %s", err.Error())
	}
}

func (h *DatavolumeHandler) handleSyncWithPagination(dvsToSync []swagger.DatavolumeStatusForSync) error {
	if len(dvsToSync) <= h.maxNumOfDatavolumesPerRequest {
		return h.sendSyncDatavolumesRequests(dvsToSync, nil, nil)
	}

	// sort all datavolumes first
	sortApiDatavolumesById(dvsToSync)

	for i := 0; i < len(dvsToSync); i += h.maxNumOfDatavolumesPerRequest - 1 {
		chunkEndIndex := i + h.maxNumOfDatavolumesPerRequest
		if chunkEndIndex > len(dvsToSync) {
			chunkEndIndex = len(dvsToSync)
		}

		chunk := dvsToSync[i:chunkEndIndex]

		var from *string = nil
		var to *string = nil
		if i > 0 {
			from = ptr.To(chunk[0].DatavolumeId)
		}
		if chunkEndIndex < len(dvsToSync) {
			to = ptr.To(chunk[len(chunk)-1].DatavolumeId)
		}

		if err := h.sendSyncDatavolumesRequests(chunk, from, to); err != nil {
			return err
		}
	}
	return nil
}

func (h *DatavolumeHandler) sendSyncDatavolumesRequests(dvsToSync []swagger.DatavolumeStatusForSync,
	from, to *string) error {
	msg := "Sync Datavolumes"
	if from != nil && *from != "" {
		msg = fmt.Sprintf("%s; from <%s>", msg, *from)
	}
	if to != nil && *to != "" {
		msg = fmt.Sprintf("%s; to <%s>", msg, *to)
	} else {
		msg = fmt.Sprintf("Finished %s", msg)
	}

	defer metrics.GetMetricsExporter().MeasureExecutionTime(h.resourceName, constants.SyncOperation)()
	res, err := h.backendClient.DatavolumesClient.SyncDatavolumes(h.backendClient.Context(),
		dvsToSync,
		config.Config().ClusterUuid,
		from,
		to,
	)

	logDetailedApiErrorMessageIfNeeded(res, err, msg)
	statusCode, err := h.backendClient.HandleHttpResponseGetStatus(res, err, msg)
	if statusCode == http.StatusRequestEntityTooLarge && h.maxNumOfDatavolumesPerRequest > 2 {
		// reduce by 20%
		h.maxNumOfDatavolumesPerRequest = int(h.maxNumOfDatavolumesPerRequest * 4 / 5)
		zap.S().Warnf("Changing maxNumOfDatavolumesPerRequest to: %d", h.maxNumOfDatavolumesPerRequest)
	}
	return err
}

func sortApiDatavolumesById(datavolumes []swagger.DatavolumeStatusForSync) {
	sort.Slice(datavolumes, func(i, j int) bool {
		return datavolumes[i].DatavolumeId < datavolumes[j].DatavolumeId
	})
}
