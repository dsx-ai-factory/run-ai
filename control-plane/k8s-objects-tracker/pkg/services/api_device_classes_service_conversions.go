/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	authclient "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/k8s-objects-tracker/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/k8s-objects-tracker/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/google/uuid"
)

// deviceClassesApiToDb converts the synced device classes into objects-table records.
// The cluster comes from the (authoritative) query parameter, and only the name is stored;
// device classes are cluster-scoped and carry no spec, so we persist an empty json spec.
func deviceClassesApiToDb(tenantId int32, clusterUuid uuid.UUID, dcs api.DeviceClasses) []*models.Object {
	result := make([]*models.Object, len(dcs.Items))
	for i, dc := range dcs.Items {
		name := dc.DeviceClassName
		result[i] = &models.Object{
			ObjectFilters: models.ObjectFilters{
				Type:      models.ObjectTypeDeviceClass,
				Name:      &name,
				TenantId:  tenantId,
				ClusterId: clusterUuid,
				ScopeType: authclient.SCOPETYPE_CLUSTER,
			},
			Spec: EmptyStructStr,
		}
	}
	return result
}

func deviceClassesDbToApiV2(objects []*models.Object) *api.DeviceClasses {
	result := &api.DeviceClasses{
		Items: make([]api.DeviceClass, len(objects)),
	}
	for i, obj := range objects {
		result.Items[i] = api.DeviceClass{
			DeviceClassName: convutils.Deref(obj.Name, ""),
			ClusterId:       obj.ClusterId.String(),
		}
	}
	return result
}
