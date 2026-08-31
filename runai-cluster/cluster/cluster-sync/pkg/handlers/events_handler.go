/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	runaiBackendClient "github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/backend-client"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/config"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/handlers/parsers"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/metrics"
	"go.uber.org/zap"

	corev1 "k8s.io/api/core/v1"
)

type EventHandler struct {
	resourceName  constants.ClusterResource
	backendClient *runaiBackendClient.BackendClient
}

func NewEventHandler(runaiBackendClient *runaiBackendClient.BackendClient) *EventHandler {
	return &EventHandler{
		resourceName:  constants.EventsResourceName,
		backendClient: runaiBackendClient,
	}
}

func (h *EventHandler) HandleSync(events []corev1.Event) error {
	return h.syncEvents(events)
}

func (h *EventHandler) syncEvents(events []corev1.Event) error {
	if len(events) <= config.Config().MaxNumOfEventsPerRequest {
		return h.sendSyncEventsRequest(events)
	}

	for i := 0; i < len(events); i += config.Config().MaxNumOfEventsPerRequest {
		chunkEndIndex := i + config.Config().MaxNumOfEventsPerRequest
		if chunkEndIndex > len(events) {
			chunkEndIndex = len(events)
		}

		chunk := events[i:chunkEndIndex]
		if err := h.sendSyncEventsRequest(chunk); err != nil {
			return err
		}
	}
	return nil
}

func (h *EventHandler) sendSyncEventsRequest(events []corev1.Event) error {
	if len(events) == 0 {
		zap.S().Debugf("No events to sync - not syncing events")
		return nil
	}

	err := h.sendSyncEventsRequestInner(events)
	if err != nil {
		return err
	}

	err = h.sendAssetsSyncEventsRequestInner(events)
	if err != nil {
		return err
	}

	return nil
}

func (h *EventHandler) sendSyncEventsRequestInner(events []corev1.Event) error {
	defer metrics.GetMetricsExporter().MeasureExecutionTime(h.resourceName, constants.SyncOperation)()
	request := h.backendClient.WorkloadsClient.EventsAPI.SyncEvents(h.backendClient.Context())
	request = request.EventsSyncRequest(*parsers.ParseEventsToSyncEventsRequest(events))

	zap.S().Infof("Sending %d events to backend", len(events))

	res, err := request.Execute()
	logDetailedApiErrorMessageIfNeeded(res, err, "sync events")
	return h.backendClient.HandleHttpResponse(res, err, "Sync events")
}

func (h *EventHandler) sendAssetsSyncEventsRequestInner(events []corev1.Event) error {
	defer metrics.GetMetricsExporter().MeasureExecutionTime(h.resourceName, constants.SyncOperation)()

	zap.S().Infof("Sending %d assets events to backend", len(events))

	res, err := h.backendClient.AssetsClient.SyncEvents(h.backendClient.Context(), *parsers.ParseEventsToAssetsSyncEventsRequest(events))
	logDetailedApiErrorMessageIfNeeded(res, err, "Sync assets events")
	return h.backendClient.HandleHttpResponse(res, err, "Sync assets events")
}
