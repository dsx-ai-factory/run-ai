/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package org_unit

import (
	"context"
	"strconv"

	"github.com/dsx-ai-factory/run-ai/control-plane/k8s-objects-tracker/pkg/common"
	orgUnitApi "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
)

// AncestorsResolver is responsible for obtaining the ancestors of a given scope.
// we use it because the ancestors may be needed a few times in the same API, but
// they are not always needed. so the resolver makes sure to do it only once, and
// only if needed.
type AncestorsResolver struct {
	scopeType orgUnitApi.ScopeType
	scopeId   string

	ancestors *orgUnitApi.Ancestors
}

func NewAncestorsResolverOfCluster(clusterId string) *AncestorsResolver {
	return &AncestorsResolver{
		scopeType: orgUnitApi.SCOPETYPE_CLUSTER,
		scopeId:   clusterId,
	}
}

func NewAncestorsResolver(scopeType orgUnitApi.ScopeType, scopeId string) *AncestorsResolver {
	return &AncestorsResolver{
		scopeType: scopeType,
		scopeId:   scopeId,
	}
}

func (ar *AncestorsResolver) Get(ctx context.Context) (*orgUnitApi.Ancestors, *orgUnitApi.Error) {
	//
	//  if not loaded yet, load it. otherwise, use the same one
	//
	if ar.ancestors == nil {
		tenantId, ctxError := common.TenantIdFromContext(ctx)
		if ctxError != nil {
			return nil, orgUnitApi.NewError(ctxError.Code, ctxError.Message)
		}

		if ar.ScopeType() == orgUnitApi.SCOPETYPE_TENANT {
			ar.ancestors = &orgUnitApi.Ancestors{
				TenantId: strconv.Itoa(int(tenantId)),
			}
			return ar.ancestors, nil
		}

		if ar.ScopeType() == orgUnitApi.SCOPETYPE_CLUSTER {
			ar.ancestors = &orgUnitApi.Ancestors{
				TenantId:    strconv.Itoa(int(tenantId)),
				ClusterUuid: *orgUnitApi.NewNullableString(&ar.scopeId),
			}
			return ar.ancestors, nil
		}

		ancestors, getErr := Get().GetAncestorsOfScope(ctx, ar.scopeType, ar.scopeId)
		if getErr != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("failed to obtain ancestors of %s/%s: %v", ar.scopeType, ar.scopeId, getErr.Message)
			return nil, getErr
		}

		ar.ancestors = ancestors
	}

	return ar.ancestors, nil
}

func (ar *AncestorsResolver) ScopeType() orgUnitApi.ScopeType {
	return ar.scopeType
}

func (ar *AncestorsResolver) ScopeId() string {
	return ar.scopeId
}
