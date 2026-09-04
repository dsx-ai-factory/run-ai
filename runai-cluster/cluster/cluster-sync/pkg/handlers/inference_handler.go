/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	"context"
	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/constants"
	infwv2a1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/inferenceworkload/v2alpha1"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type InferenceHandler struct {
	*WorkloadHandler
}

func (h *InferenceHandler) HandleUpdateItem(key client.ObjectKey) error {
	iw := &infwv2a1.InferenceWorkload{}
	err := h.cachedClient.Get(context.Background(), key, iw)
	if err != nil {
		return err
	}

	return h.HandleInferenceUpdate(iw)
}

func NewInferenceHandler(workloadHandler *WorkloadHandler) *InferenceHandler {
	return &InferenceHandler{WorkloadHandler: workloadHandler}
}

func (h *WorkloadHandler) HandleInferenceAdd(workload *infwv2a1.InferenceWorkload) error {
	if h.WasSynced(workload.UID, workload.Namespace) {
		zap.S().Infof("InferenceWorkload <%s/%s> was already synced - not adding",
			workload.Namespace, workload.Name)
		return nil
	}
	zap.S().Debugf("Handling InferenceWorkload <%s/%s> add", workload.Namespace, workload.Name)

	return h.handleInferenceAddUpdate(workload)
}

func (h *WorkloadHandler) HandleInferenceUpdate(workload *infwv2a1.InferenceWorkload) error {
	zap.S().Debugf("Handling InferenceWorkload <%s/%s> update", workload.Namespace, workload.Name)
	return h.handleInferenceAddUpdate(workload)
}

func (h *WorkloadHandler) handleInferenceAddUpdate(workload *infwv2a1.InferenceWorkload) error {
	workloadSyncFields, _, err := h.parser.ParseInferenceWorkloadToServerScheme(workload)
	if err != nil {
		return err
	}
	if isLegacyWorkloadUpToDate(workload.Status.SyncStatus, workload.Generation,
		workload.Status.LastStatusChangeTime, workloadSyncFields) {
		zap.S().Debugf("InferenceWorkload <%s/%s> is up to date - not updating",
			workload.Namespace, workload.Name)
		return nil
	}

	return h.handleUpsert(workloadSyncFields, workload.Namespace, workload.Name, constants.InferenceWorkload)
}

func (h *WorkloadHandler) getParsedInferenceWorkloadsInNamespace(namespaceName string,
	syncedInNamespace map[types.UID]bool) ([]openapi.WorkloadSyncFields, []string, []client.Object, error) {

	fieldSelector := GetRelevantRunaiWorkloadsFieldSelector()

	workloads := &infwv2a1.InferenceWorkloadList{}
	err := h.cachedClient.List(context.Background(), workloads, &client.ListOptions{
		Namespace:     namespaceName,
		FieldSelector: fieldSelector,
	})
	if err != nil {
		zap.S().Errorf("Failed listing InferenceWorkloads in namespace <%s>, error: %s",
			namespaceName, err.Error())
		return []openapi.WorkloadSyncFields{}, []string{}, []client.Object{}, err
	}

	workloadSyncFields, upToDateIds, updateSyncStatus, _, err := h.ParseInferenceWorkloads(workloads.Items, syncedInNamespace)
	return workloadSyncFields, upToDateIds, updateSyncStatus, err
}

func (h *WorkloadHandler) ParseInferenceWorkloads(workloads []infwv2a1.InferenceWorkload, syncedInNamespace map[types.UID]bool) ([]openapi.WorkloadSyncFields, []string, []client.Object, []corev1.Pod, error) {
	workloadsToSync := []openapi.WorkloadSyncFields{}
	upToDateIds := []string{}
	workloadPods := []corev1.Pod{}
	for i := range workloads {
		// important: to avoid bug where we insert a pointer of for-looped map value to an array
		workload := &workloads[i]
		syncedInNamespace[workload.UID] = true

		parsed, pods, innerErr := h.parser.ParseInferenceWorkloadToServerScheme(workload)
		if innerErr != nil {
			zap.S().Warnf("Failed parsing InferenceWorkload <%s/%s>, error: %s",
				workload.Namespace, workload.Name, innerErr.Error())
			continue
		}

		workloadPods = append(workloadPods, pods...)
		if isLegacyWorkloadUpToDate(workload.Status.SyncStatus, workload.Generation,
			workload.Status.LastStatusChangeTime, parsed) {
			upToDateIds = append(upToDateIds, parsed.Id)
			continue
		}
		workloadsToSync = append(workloadsToSync, *parsed)
	}

	return workloadsToSync, upToDateIds, nil, workloadPods, nil
}
