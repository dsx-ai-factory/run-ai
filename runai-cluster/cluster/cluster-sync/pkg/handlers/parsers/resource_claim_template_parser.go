/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package parsers

import (
	"fmt"

	auth_client "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	trackerclient "github.com/dsx-ai-factory/run-ai/runai-clients/k8s-objects-tracker-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/config"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/util"
	commonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	resourcev1 "k8s.io/api/resource/v1"
)

// ParsedResourceClaimTemplate carries both the server-bound payload and the
// scope the parser inferred for this RCT; the handler uses the scope to drop
// replicas (project-scope templates living outside their project namespace,
// non-project templates living outside the Run:ai default namespace).
type ParsedResourceClaimTemplate struct {
	Scope           ScopeTypeAndId
	CreationRequest trackerclient.ResourceClaimTemplate
}

// ParseResourceClaimTemplateToServerScheme converts a cluster-local RCT into
// the single body shape used by the collapsed /resource-claim-templates/sync
// endpoint: scope and scopeId travel inside each item rather than in the URL.
// Returns an error when the RCT cannot be scoped (which, for the sync
// pipeline, means the handler will skip it).
func ParseResourceClaimTemplateToServerScheme(rct *resourcev1.ResourceClaimTemplate) (*ParsedResourceClaimTemplate, error) {
	var origin trackerclient.NullableOrigin
	var assetId trackerclient.NullableString

	if util.IsAssetRelatedResource(&rct.ObjectMeta) {
		origin.Set(&originAssets)
		// RCTs created by the assets-sync pipeline stamp the owning asset
		// id so the control plane can join tracker rows back to the asset
		// record; propagate it when present so the server-side filter has
		// something to key on.
		if id, ok := rct.Labels[commonconstants.RunaiAssetIdLabel]; ok && id != "" {
			s := id
			assetId.Set(&s)
		}
	} else {
		// Everything that isn't asset-managed is treated as a user-created
		// template surfaced into the tracker as origin=existing, mirroring
		// how PVCs are classified.
		origin.Set(&originExisting)
	}

	result := &ParsedResourceClaimTemplate{
		CreationRequest: trackerclient.ResourceClaimTemplate{
			Name:      rct.Name,
			Namespace: rct.Namespace,
			Origin:    origin,
			AssetId:   assetId,
			Issues:    util.ExtractIssuesFromObject(rct.GetAnnotations()),
		},
	}

	if err := parseScope(&result.Scope, &rct.ObjectMeta); err != nil {
		return nil, fmt.Errorf("parse RCT scope %s/%s: %w", rct.Namespace, rct.Name, err)
	}
	result.CreationRequest.ScopeType = trackerclient.ResourceScopeType(result.Scope.Type)
	// Mirror the control-plane invariant assignScopeId enforces on the
	// server: cluster-scope rows always carry the cluster id as scopeId so
	// the server-side equality check (scopeId == clusterId) holds.
	if result.Scope.Type == auth_client.SCOPETYPE_CLUSTER {
		result.CreationRequest.ScopeId = config.Config().ClusterUuid
	} else {
		result.CreationRequest.ScopeId = result.Scope.Id
	}

	return result, nil
}
