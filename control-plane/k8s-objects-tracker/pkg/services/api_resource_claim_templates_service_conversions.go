/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"k8s.io/utils/ptr"

	authclient "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/k8s-objects-tracker/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/k8s-objects-tracker/pkg/database/models"
)

// resourceClaimTemplateSpecEnvelope is the shape persisted inside the objects
// row's `spec` jsonb column. The DB column itself holds one JSON blob per row,
// so we fold the fields that are not part of ObjectFilters (origin, assetId,
// namespace) into this envelope. We deliberately do NOT persist the raw RCT
// spec: tracker rows are a cluster-wide inventory, and shipping every RCT's
// spec on every sync would blow up the request/response size without a
// consumer that needs the field today.
type resourceClaimTemplateSpecEnvelope struct {
	Origin    *api.Origin `json:"origin,omitempty"`
	AssetId   *string     `json:"assetId,omitempty"`
	Namespace string      `json:"namespace,omitempty"`
}

// resourceClaimTemplateApiToDb builds a persistable Object from a single RCT
// payload. Scope is read from the RCT itself: unlike the earlier per-scope
// sync endpoints, callers now carry scopeType/scopeId in the body so a single
// POST/PUT/Sync surface can accept items at any scope.
func resourceClaimTemplateApiToDb(tenantId int32, clusterUuid uuid.UUID, rct *api.ResourceClaimTemplate) (*models.Object, *api.Error) {
	scopeType := authclient.ScopeType(string(rct.ScopeType))
	filters := models.ObjectFilters{
		TenantId:  tenantId,
		Type:      models.ObjectTypeResourceClaimTemplate,
		ClusterId: clusterUuid,
		ScopeType: scopeType,
		Name:      ptr.To(rct.Name),
	}
	if apiErr := assignScopeId(&filters, clusterUuid.String(), rct.ScopeId); apiErr != nil {
		return nil, apiErr
	}

	envelope := resourceClaimTemplateSpecEnvelope{
		Origin:    rct.Origin,
		AssetId:   rct.AssetId,
		Namespace: rct.Namespace,
	}
	buf, err := json.Marshal(envelope)
	if err != nil {
		return nil, api.NewError(http.StatusInternalServerError, "failed marshaling resource claim template spec", err)
	}

	return &models.Object{
		ObjectFilters: filters,
		Spec:          string(buf),
	}, nil
}

func resourceClaimTemplateDbToApi(obj *models.Object) (*api.ResourceClaimTemplate, error) {
	envelope := resourceClaimTemplateSpecEnvelope{}
	if obj.Spec != "" {
		if err := json.Unmarshal([]byte(obj.Spec), &envelope); err != nil {
			return nil, err
		}
	}

	name := ""
	if obj.Name != nil {
		name = *obj.Name
	}

	rct := &api.ResourceClaimTemplate{
		Name:      name,
		Namespace: envelope.Namespace,
		ScopeType: api.ScopeType2(obj.ScopeType),
		Origin:    envelope.Origin,
		AssetId:   envelope.AssetId,
	}
	switch obj.ScopeType {
	case authclient.SCOPETYPE_CLUSTER:
		rct.ScopeId = obj.ClusterId.String()
	case authclient.SCOPETYPE_DEPARTMENT:
		rct.ScopeId = obj.DepartmentId
	case authclient.SCOPETYPE_PROJECT:
		rct.ScopeId = obj.ProjectId
	}
	return rct, nil
}

// assignScopeId maps the service-level scope id into the right column on the
// Object filter and enforces the invariants the schema cannot: for cluster
// scope the scopeId must equal the clusterId query parameter (so the row the
// caller writes matches the cluster they addressed), and for department /
// project scope the scopeId must be non-empty.
func assignScopeId(filter *models.ObjectFilters, clusterId string, scopeId string) *api.Error {
	if scopeId == "" {
		return api.NewValidationError("scopeId is required")
	}
	switch filter.ScopeType {
	case authclient.SCOPETYPE_CLUSTER:
		if scopeId != clusterId {
			return api.NewValidationError("scopeId must equal clusterId for cluster-scoped resource claim templates")
		}
		// cluster-scoped rows already carry cluster_id; nothing else to set.
	case authclient.SCOPETYPE_DEPARTMENT:
		filter.DepartmentId = scopeId
	case authclient.SCOPETYPE_PROJECT:
		filter.ProjectId = scopeId
	default:
		return api.NewValidationError("unsupported scope type for resource claim templates")
	}
	return nil
}

// originMatches implements the same semantics as the PVC origin filter:
// empty / "all" means no filter; otherwise match an exact origin string.
func originMatches(filter string, rowOrigin *api.Origin) bool {
	if filter == "" || filter == "all" {
		return true
	}
	if rowOrigin == nil {
		return filter == string(api.ORIGIN_EXISTING)
	}
	return filter == string(*rowOrigin)
}
