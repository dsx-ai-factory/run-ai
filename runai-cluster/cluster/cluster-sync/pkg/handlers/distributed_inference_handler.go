/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"

	openapi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/constants"
	dinfwv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/distributedinferenceworkload/v1alpha1"
	"go.uber.org/zap"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type DistributedInferenceHandler struct {
	workloadHandler *WorkloadHandler
}

func NewDistributedInferenceHandler(workloadHandler *WorkloadHandler) *DistributedInferenceHandler {
	return &DistributedInferenceHandler{
		workloadHandler: workloadHandler,
	}
}

func (h *DistributedInferenceHandler) HandleAdd(obj interface{}) error {
	workload, ok := obj.(*dinfwv1.DistributedInferenceWorkload)
	if !ok {
		zap.S().Error("Failed to cast to DistributedInferenceWorkload in HandleAdd")
		return nil
	}
	return h.HandleAddUpdate(workload)
}

func (h *DistributedInferenceHandler) HandleUpdate(newObj interface{}) error {
	workload, ok := newObj.(*dinfwv1.DistributedInferenceWorkload)
	if !ok {
		zap.S().Error("Failed to cast to DistributedInferenceWorkload in HandleUpdate")
		return nil
	}
	return h.HandleAddUpdate(workload)
}

// HandleUpdateItem implements the Handler interface for rate limited controller
func (h *DistributedInferenceHandler) HandleUpdateItem(key client.ObjectKey) error {
	workload := &dinfwv1.DistributedInferenceWorkload{}
	err := h.workloadHandler.cachedClient.Get(context.Background(), key, workload)
	if err != nil {
		return err
	}
	return h.HandleUpdate(workload)
}

func (h *DistributedInferenceHandler) HandleAddUpdate(workload *dinfwv1.DistributedInferenceWorkload) error {
	workload = workload.DeepCopy()

	zap.S().Debugf("Handling DistributedInferenceWorkload <%s/%s> add/update", workload.Namespace, workload.Name)

	// Check if parser is available
	if h.workloadHandler.parser == nil {
		zap.S().Errorf("Parser is not initialized for distributed inference workload <%s/%s>",
			workload.Namespace, workload.Name)
		return fmt.Errorf("parser is not initialized")
	}

	// Parse the distributed inference workload to server scheme
	workloadSyncFields, _, err := h.workloadHandler.parser.ParseDistributedInferenceWorkloadToServerScheme(workload)
	if err != nil {
		zap.S().Errorf("Failed to parse distributed inference workload <%s/%s>: %s",
			workload.Namespace, workload.Name, err.Error())
		return err
	}
	if isLegacyWorkloadUpToDate(workload.Status.SyncStatus, workload.Generation,
		workload.Status.LastStatusChangeTime, workloadSyncFields) {
		zap.S().Debugf("DistributedInferenceWorkload <%s/%s> is up to date - not updating",
			workload.Namespace, workload.Name)
		return nil
	}

	// Sync to control plane
	return h.workloadHandler.handleUpsert(workloadSyncFields, workload.Namespace, workload.Name, constants.DistributedInferenceWorkload)
}

func (h *WorkloadHandler) getParsedDistributedInferenceWorkloadsInNamespace(namespaceName string,
	syncedInNamespace map[types.UID]bool) ([]openapi.WorkloadSyncFields, []string, []client.Object, error) {

	fieldSelector := GetRelevantRunaiWorkloadsFieldSelector()

	workloads := &dinfwv1.DistributedInferenceWorkloadList{}
	err := h.cachedClient.List(context.Background(), workloads, &client.ListOptions{
		Namespace:     namespaceName,
		FieldSelector: fieldSelector,
	})
	if err != nil {
		zap.S().Errorf("Failed listing DistributedInferenceWorkloads in namespace <%s>, error: %s",
			namespaceName, err.Error())
		return []openapi.WorkloadSyncFields{}, []string{}, []client.Object{}, err
	}

	workloadSyncFields, upToDateIds, updateSyncStatus, _, err := h.parseDistributedInferenceWorkloads(workloads.Items, syncedInNamespace)
	return workloadSyncFields, upToDateIds, updateSyncStatus, err
}

func (h *WorkloadHandler) parseDistributedInferenceWorkloads(
	workloads []dinfwv1.DistributedInferenceWorkload, syncedInNamespace map[types.UID]bool,
) ([]openapi.WorkloadSyncFields, []string, []client.Object, []corev1.Pod, error) {
	workloadsToSync := []openapi.WorkloadSyncFields{}
	upToDateIds := []string{}
	workloadPods := []corev1.Pod{}
	for i := range workloads {
		// important: to avoid bug where we insert a pointer of for-looped map value to an array
		workload := &workloads[i]
		syncedInNamespace[workload.UID] = true

		parsed, pods, innerErr := h.parser.ParseDistributedInferenceWorkloadToServerScheme(workload)
		if innerErr != nil {
			zap.S().Warnf("Failed parsing DistributedInferenceWorkload <%s/%s>, error: %s",
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
