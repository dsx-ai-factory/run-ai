/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/handlers/parsers"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/metrics"
	clusterconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"go.uber.org/zap"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	_ "k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"

	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	runaiBackendClient "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/backend-client"
)

type DepartmentHandler struct {
	resourceName  constants.ClusterResource
	backendClient *runaiBackendClient.BackendClient
	cachedClient  client.Client
}

func NewDepartmentHandler(runaiBackendClient *runaiBackendClient.BackendClient, cachedClient client.Client) *DepartmentHandler {
	return &DepartmentHandler{
		resourceName:  constants.DepartmentResourceName,
		backendClient: runaiBackendClient,
		cachedClient:  cachedClient,
	}
}

func (departmentHandler *DepartmentHandler) HandleUpdateItem(key client.ObjectKey) error {
	department := &kaiv1alpha1.Department{}
	err := departmentHandler.cachedClient.Get(context.Background(), key, department)
	if err != nil {
		return err
	}

	return departmentHandler.HandleAddUpdate(department)
}

func (departmentHandler *DepartmentHandler) HandleAddUpdate(department *kaiv1alpha1.Department) error {
	syncTime := metav1.Now()
	updateStatusErr := departmentHandler.updateStatus(department)
	if updateStatusErr == nil {
		err := departmentHandler.updateSyncStatus(department, syncTime)
		if err != nil {
			zap.S().Warnf("Failed to update last sync time in department %s, err: %v", department.Name, err)
		}
	}
	return updateStatusErr
}

// updateSyncStatus records the last sync time as an annotation on the department.
// On the new kai.resources Department CRD the sync timestamp is no longer a status
// field; it is kept as an annotation managed exclusively by cluster-sync.
func (departmentHandler *DepartmentHandler) updateSyncStatus(department *kaiv1alpha1.Department, syncTime metav1.Time) error {
	patchBytes, err := json.Marshal(map[string]any{"metadata": map[string]any{
		"annotations": map[string]string{
			parsers.LastSyncTimestampAnnotation: syncTime.UTC().Format(time.RFC3339),
		}}})
	if err != nil {
		return err
	}

	patch := client.RawPatch(types.MergePatchType, patchBytes)
	err = departmentHandler.cachedClient.Patch(context.Background(), department, patch)
	if err != nil {
		return err
	}
	return nil
}

func (departmentHandler *DepartmentHandler) updateStatus(department *kaiv1alpha1.Department) error {
	defer metrics.GetMetricsExporter().MeasureExecutionTime(departmentHandler.resourceName, constants.UpdateOperation)()
	toUpdate := parsers.ParseKaiDepartmentToUpdateStatusRequest(department)
	res, err := departmentHandler.backendClient.OrgUnitClient.UpdateDepartmentsStatus(
		departmentHandler.backendClient.Context(),
		department.Labels[clusterconstants.SyncedDepartmentIdLabel],
		toUpdate,
	)
	return departmentHandler.backendClient.HandleHttpResponse(res, err, fmt.Sprintf("Updated status of department %s", department.Name))
}

func (departmentHandler *DepartmentHandler) HandleDelete(department *kaiv1alpha1.Department) {
	zap.S().Warnf("got delete on department name %s, id: %s", department.Name, department.Labels[clusterconstants.SyncedDepartmentIdLabel])
}
