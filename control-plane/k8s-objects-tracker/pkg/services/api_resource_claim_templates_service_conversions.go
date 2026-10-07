/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"encoding/json"

	authclient "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/k8s-objects-tracker/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/k8s-objects-tracker/pkg/database/models"
	"k8s.io/utils/ptr"
)

// resourceClaimTemplateSpecEnvelope is the shape persisted inside the objects
// row's `spec` jsonb column. The DB column itself is scoped to one JSON blob
// per row, so we fold the fields that are not part of ObjectFilters (origin,
// assetId, namespace, raw RCT spec) into this envelope. Keeping it an exported
// Go type (lowercase name but exported fields) makes json round-trips stable.
type resourceClaimTemplateSpecEnvelope struct {
	Origin    *api.Origin            `json:"origin,omitempty"`
	AssetId   *string                `json:"assetId,omitempty"`
	Namespace string                 `json:"namespace,omitempty"`
	Spec      map[string]interface{} `json:"spec,omitempty"`
}

func resourceClaimTemplateApiToDb(scopeFilter *models.ObjectFilters, rct *api.ResourceClaimTemplate) (*models.Object, error) {
	envelope := resourceClaimTemplateSpecEnvelope{
		Origin:    rct.Origin,
		AssetId:   rct.AssetId,
		Namespace: rct.Namespace,
		Spec:      rct.Spec,
	}
	buf, err := json.Marshal(envelope)
	if err != nil {
		return nil, err
	}

	// Clone the scope filter rather than share a pointer: downstream callers
	// use the same filter object as a sync-wide WHERE clause and should not
	// pick up this row's Name as an accidental extra predicate.
	filters := *scopeFilter
	filters.Name = ptr.To(rct.Name)

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

	return &api.ResourceClaimTemplate{
		Name:      name,
		Namespace: envelope.Namespace,
		Origin:    envelope.Origin,
		AssetId:   envelope.AssetId,
		Spec:      envelope.Spec,
	}, nil
}

// assignScopeId maps the service-level scope id into the right column on the
// Object filter. The Object model stores scope id in different columns per
// scope (cluster_id / department_id / project_id), so we must dispatch here
// rather than cram everything into a single "scope_id" column.
func assignScopeId(filter *models.ObjectFilters, scopeId string) *api.Error {
	switch filter.ScopeType {
	case authclient.SCOPETYPE_CLUSTER:
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
