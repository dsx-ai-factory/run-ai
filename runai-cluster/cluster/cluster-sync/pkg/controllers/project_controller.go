/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package controllers

import (
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/constants"
	commonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"

	org_unit_openapi "github.com/dsx-ai-factory/run-ai/runai-clients/org-unit-service-client/generated"
	"k8s.io/utils/ptr"

	backendclient "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/backend-client"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/cachedclient"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/handlers"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/handlers/parsers"
	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"go.uber.org/zap"
	"k8s.io/client-go/tools/cache"
)

func AddProjectHandlerToCache(backendClient *backendclient.BackendClient, cachedClient cachedclient.CachedClient) {
	projectHandler := handlers.NewProjectHandler(backendClient, cachedClient)
	backendClient.RegisterSyncHandler(projectHandler)

	rateLimitedController := cachedClient.NewRateLimitedController(cachedclient.ProjectRL, projectHandler)

	projectInformer := cachedClient.GetInformer(cachedclient.ProjectType)
	_, err := projectInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(object any) {
			genericControllerErrorHandler(projectHandler.HandleAddUpdate(object.(*kaiv1alpha1.Project)), constants.AddOperation, constants.ProjectResourceName, object.(*kaiv1alpha1.Project))
		},
		UpdateFunc: func(oldObject any, newObject any) {
			newObj := newObject.(*kaiv1alpha1.Project)
			if oldObject == nil {
				rateLimitedController.AddRateLimited(objectKeyFromObject(newObj))
				return
			}

			shouldUpdateProj := shouldUpdate(oldObject.(*kaiv1alpha1.Project), newObj)
			if shouldUpdateProj {
				rateLimitedController.AddRateLimited(objectKeyFromObject(newObj))
			}
		},
		DeleteFunc: func(object any) {
			obj, ok := deletedObject[*kaiv1alpha1.Project](object)
			if !ok {
				return
			}
			rateLimitedController.MarkItemDeleted(objectKeyFromObject(obj))
			genericControllerErrorHandler(projectHandler.HandleDelete(obj), constants.DeleteOperation, constants.ProjectResourceName, obj)
		},
	})

	if err != nil {
		zap.S().Fatalf("Failed to add event handler to project informer, error: %s", err.Error())
	}

	zap.S().Debug("The Agent project controller has started")
}

func shouldUpdate(old, new *kaiv1alpha1.Project) bool {
	oldUpdatedAt := old.Annotations[commonconstants.UpdatedAtAnnotation]
	newUpdatedAt := new.Annotations[commonconstants.UpdatedAtAnnotation]
	if oldUpdatedAt == "" || newUpdatedAt == "" || oldUpdatedAt != newUpdatedAt {
		return true
	}

	toUpdateOld := parsers.ParseKaiProjectToUpdateStatusRequest(old)
	toUpdateNew := parsers.ParseKaiProjectToUpdateStatusRequest(new)

	return hasStatusChanged(toUpdateOld.Status, toUpdateNew.Status)
}

func hasStatusChanged(old, new org_unit_openapi.ProjectStatus) bool {
	if ptr.Deref(old.Phase, "") != ptr.Deref(new.Phase, "") {
		return true
	}
	if getNullableString(old.PhaseMessage) != getNullableString(new.PhaseMessage) {
		return true
	}
	if getNullableString(old.Namespace) != getNullableString(new.Namespace) {
		return true
	}
	if !quotaStatusEquals(old.QuotaStatus, new.QuotaStatus) {
		return true
	}
	if !nodePoolQuotaStatusEquals(old.NodePoolQuotaStatuses, new.NodePoolQuotaStatuses) {
		return true
	}
	return false
}

func nodePoolQuotaStatusEquals(old, new []org_unit_openapi.QuotaStatusNodePool) bool {
	if len(old) != len(new) {
		return false
	}
	nodePoolQuotaMap := make(map[string]*org_unit_openapi.QuotaStatus)
	for _, quotaStatus := range old {
		nodePoolQuotaMap[quotaStatus.NodePoolName] = &org_unit_openapi.QuotaStatus{
			Allocated:               quotaStatus.Allocated,
			AllocatedNonPreemptible: quotaStatus.AllocatedNonPreemptible,
			Requested:               quotaStatus.Requested,
		}
	}
	for _, quotaStatus := range new {
		oldQuota, found := nodePoolQuotaMap[quotaStatus.NodePoolName]
		if !found {
			return false
		}
		newQuota := &org_unit_openapi.QuotaStatus{
			Allocated:               quotaStatus.Allocated,
			AllocatedNonPreemptible: quotaStatus.AllocatedNonPreemptible,
			Requested:               quotaStatus.Requested,
		}
		if !quotaStatusEquals(oldQuota, newQuota) {
			return false
		}
	}
	return true
}

func getNullableString(str org_unit_openapi.NullableString) string {
	if !str.IsSet() || str.Get() == nil {
		return ""
	}
	return *str.Get()
}

//nolint:unused
func _getNullableTime(dateTime org_unit_openapi.NullableTime) time.Time {
	if !dateTime.IsSet() || dateTime.Get() == nil {
		return time.Time{}
	}
	return *dateTime.Get()
}

func quotaStatusEquals(old, new *org_unit_openapi.QuotaStatus) bool {
	if old == nil {
		return new == nil
	}
	if new == nil {
		return false
	}
	return quotaStatusResourcesEquals(old.Allocated, new.Allocated) &&
		quotaStatusResourcesEquals(old.AllocatedNonPreemptible, new.AllocatedNonPreemptible) &&
		quotaStatusResourcesEquals(old.Requested, new.Requested)
}

func quotaStatusResourcesEquals(old, new *org_unit_openapi.QuotaStatusResource) bool {
	if old == nil {
		return new == nil
	}
	if new == nil {
		return false
	}
	return singleQuotaEquals(old.Gpu, new.Gpu) &&
		singleQuotaEquals(old.Cpu, new.Cpu) &&
		singleQuotaEquals(old.Memory, new.Memory)
}

func singleQuotaEquals(old, new *float64) bool {
	if old == nil {
		return new == nil
	}
	if new == nil {
		return false
	}
	return *old == *new
}
