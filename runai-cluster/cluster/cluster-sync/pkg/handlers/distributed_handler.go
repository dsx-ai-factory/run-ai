/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	"context"
	"fmt"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/extensions"
	distirbuedExt "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/extensions/distributed"
	"golang.org/x/exp/maps"

	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/constants"
	diwv2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/distributedworkload/v2alpha1"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type DistributedHandler struct {
	*WorkloadHandler
}

func (h *DistributedHandler) HandleUpdateItem(key client.ObjectKey) error {
	dw := &diwv2.DistributedWorkload{}
	err := h.cachedClient.Get(context.Background(), key, dw)
	if err != nil {
		return err
	}

	return h.HandleDistributedUpdate(dw)
}

func NewDistributedHandler(workloadHandler *WorkloadHandler) *DistributedHandler {
	return &DistributedHandler{WorkloadHandler: workloadHandler}
}

func (h *WorkloadHandler) HandleDistributedAdd(workload *diwv2.DistributedWorkload) error {
	if h.WasSynced(workload.UID, workload.Namespace) {
		zap.S().Infof("DistributedWorkload <%s/%s> was already synced - not adding",
			workload.Namespace, workload.Name)
		return nil
	}
	zap.S().Debugf("Handling DistributedWorkload <%s/%s> add", workload.Namespace, workload.Name)

	return h.handleDistributedAddUpdate(workload)
}

func (h *WorkloadHandler) HandleDistributedUpdate(workload *diwv2.DistributedWorkload) error {
	zap.S().Debugf("Handling DistributedWorkload <%s/%s> update", workload.Namespace, workload.Name)
	return h.handleDistributedAddUpdate(workload)
}

func (h *WorkloadHandler) handleDistributedAddUpdate(workload *diwv2.DistributedWorkload) error {
	workload = workload.DeepCopy()

	clusterFrameworkExtension, depsErr := h.getClusterFrameworkExtension(context.Background())
	if depsErr != nil {
		zap.S().Errorf("Failed to get cluster dependencies, error: %s", depsErr.Error())
	}

	workloadSyncFields, _, err := h.parser.ParseDistributedWorkloadToServerScheme(workload, clusterFrameworkExtension)
	if err != nil {
		return err
	}
	if isLegacyWorkloadUpToDate(workload.Status.SyncStatus, workload.Generation,
		workload.Status.LastStatusChangeTime, workloadSyncFields) {
		zap.S().Debugf("DistributedWorkload <%s/%s> is up to date - not updating",
			workload.Namespace, workload.Name)
		return nil
	}

	return h.handleUpsert(workloadSyncFields, workload.Namespace, workload.Name, constants.DistributedWorkload)
}

func (h *WorkloadHandler) getParsedDistributedWorkloadsInNamespace(namespaceName string,
	syncedInNamespace map[types.UID]bool) ([]openapi.WorkloadSyncFields, []string, []client.Object, error) {

	fieldSelector := GetRelevantRunaiWorkloadsFieldSelector()

	workloads := &diwv2.DistributedWorkloadList{}
	err := h.cachedClient.List(context.Background(), workloads, &client.ListOptions{
		Namespace:     namespaceName,
		FieldSelector: fieldSelector,
	})
	if err != nil {
		zap.S().Errorf("Failed listing DistributedWorkloads in namespace <%s>, error: %s",
			namespaceName, err.Error())
		return []openapi.WorkloadSyncFields{}, []string{}, []client.Object{}, err
	}

	workloadSyncFields, upToDateIds, updateSyncStatus, _, err := h.parseDistributedWorkloads(workloads.Items, syncedInNamespace)
	return workloadSyncFields, upToDateIds, updateSyncStatus, err
}

func (h *WorkloadHandler) parseDistributedWorkloads(workloads []diwv2.DistributedWorkload, syncedInNamespace map[types.UID]bool) ([]openapi.WorkloadSyncFields, []string, []client.Object, []corev1.Pod, error) {
	clusterFrameworkExtension, depsErr := h.getClusterFrameworkExtension(context.Background())
	if depsErr != nil {
		zap.S().Errorf("Failed to get cluster dependencies, error: %s", depsErr.Error())
	}

	workloadsToSync := []openapi.WorkloadSyncFields{}
	upToDateIds := []string{}
	workloadPods := []corev1.Pod{}
	for i := range workloads {
		// important: to avoid bug where we insert a pointer of for-looped map value to an array
		workload := &workloads[i]
		syncedInNamespace[workload.UID] = true

		parsed, pods, innerErr := h.parser.ParseDistributedWorkloadToServerScheme(workload, clusterFrameworkExtension)
		if innerErr != nil {
			zap.S().Warnf("Failed parsing DistributedWorkload <%s/%s>, error: %s",
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

func (h *WorkloadHandler) getClusterFrameworkExtension(ctx context.Context) (extensions.Group, error) {
	extClient := distirbuedExt.NewClient(h.cachedClient)
	frameworks, err := extClient.List(ctx, maps.Keys(distirbuedExt.SupportedExtensions))
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster framework extension stauts: %w", err)
	}
	return frameworks, nil
}
