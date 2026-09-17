/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tree

import (
	"context"
	"fmt"
	orgUnitClient "github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/clients/org_unit"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/auth"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client/structs"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"gorm.io/gorm"
	"net/http"
	"slices"
	"strconv"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/database/dbmodel"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/scopes"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
)

type MapOrgUnitScopeToInfo structs.OrgUnitMap

// InitializeTenant is invoked on the first request for a given tenant. It reconstructs the hierarchy
// tree of the tenant from the DB
func (ht *HierarchyTree) InitializeTenant(ctx context.Context, tenantId int32, workloadType WorkloadType) *api.Error {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("[%s] restoring tree from DB started", ht.treeId.ToString())

	//
	//   load the records from the DB, note that we want also deleted policy records, because we want to include
	//   in the tree also policies that only inherit stuff but do not have given policy
	//
	records, err := dao.ListPoliciesOfTenant(context.Background(), tenantId, string(workloadType), dbmodel.IncludeDeletedPolicies)
	if err != nil {
		return api.NewError(http.StatusServiceUnavailable, "failed to initialize tenant", fmt.Errorf("%s", err.Message))
	}

	//
	//    load the orgTree of the tenant from the backend
	//
	orgTreeAsMap, _, orgTreeErr := orgUnitClient.Get().GetOrgTreeMap(ctx)
	if orgTreeErr != nil {
		return api.NewError(http.StatusServiceUnavailable, "failed to get orgTree", fmt.Errorf("%s", orgTreeErr.Message))
	}

	db := runaidb.DB()

	for _, record := range records {
		orgUnit, err := scopes.MakeOrgUnit(record.Scope, record.TenantId, record.ClusterId, record.DepartmentId, record.ProjectId)
		if err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("[%s] db record skipped: %s", ht.treeId.ToString(), err.Message)
			continue
		}

		relevantOrgUnits, orgUnitMap := GetAncestors(ctx, orgUnit, orgTreeAsMap)
		if relevantOrgUnits == nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("[%s] db record of %s skipped: ancestors not found", ht.treeId.ToString(), orgUnit.String())
			continue
		}

		//
		//  in old policy implementation, we had the hierarchy in the tree. but currently we
		//  need to have the entire hierarchy on the policy record so patch this information
		//  for any policy that has partial hierarchy
		//
		recordUpdated := false

		if clusterId, found := orgUnitMap[api.SCOPETYPE_CLUSTER]; found {
			if record.ClusterId == nil {
				record.ClusterId = convutils.Ref(clusterId)
				recordUpdated = true
			}
		}

		if departmentId, found := orgUnitMap[api.SCOPETYPE_DEPARTMENT]; found {
			if record.DepartmentId == nil {
				record.DepartmentId = convutils.Ref(departmentId)
				recordUpdated = true
			}
		}

		if recordUpdated {
			dbErr := dao.PatchPolicyScopes(ctx, db, &record.Policy)
			if dbErr != nil {
				logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to patch scopes of: %s", dbErr.Message)
			}
		}
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("[%s] restoring tree from DB completed", ht.treeId.ToString())

	return nil
}

func GetAncestors(ctx context.Context, orgUnit scopes.OrgUnit, orgUnitMap structs.OrgUnitMap) ([]scopes.OrgUnit, map[api.ScopeType]string) {
	asMap := make(map[api.ScopeType]string)

	if orgUnit.Scope == api.SCOPETYPE_TENANT && orgUnit.Id == "" {
		tenantId := auth.TenantIdOf(ctx)
		orgUnit.Id = fmt.Sprintf("%d", tenantId)

		asMap[api.SCOPETYPE_TENANT] = strconv.Itoa(int(tenantId))
	}

	ancestors := make([]scopes.OrgUnit, 0)
	gatherAncestors(orgUnitMap, structs.OrgUnitKey{
		Scope: string(orgUnit.Scope),
		Id:    orgUnit.Id,
	}, &ancestors, asMap)

	// the caller expects the ancestors ordered from lower to upper
	slices.Reverse(ancestors)

	return ancestors, asMap
}

func gatherAncestors(mapOrgUnit structs.OrgUnitMap, orgUnitKey structs.OrgUnitKey, ancestors *[]scopes.OrgUnit, asMap map[api.ScopeType]string) {
	*ancestors = append(*ancestors, scopes.OrgUnit{
		Scope: api.ScopeType(orgUnitKey.Scope),
		Id:    orgUnitKey.Id,
	})
	asMap[api.ScopeType(orgUnitKey.Scope)] = orgUnitKey.Id

	parent, foundParent := mapOrgUnit[orgUnitKey]
	if foundParent {
		gatherAncestors(mapOrgUnit, parent, ancestors, asMap)
		return
	}
}

// before running policy actions, make sure policies for all relevant org units exist. any
// policy that is missing needs to be created as deleted. we should take into account multiple
// requests running in parallel and ignore conflicts.
func (ht *HierarchyTree) makeSureAllPoliciesExist(ctx context.Context, request *GenericPolicyRequest, relevantOrgUnits []scopes.OrgUnit) *api.Error {
	db := runaidb.DB()

	for idx, orgUnit := range relevantOrgUnits {
		if orgUnit.Scope == api.SCOPETYPE_SYSTEM {
			continue
		}
		orgUnitRequest := request.Clone()

		orgUnitRequest.PolicyMeta.Scope = orgUnit.Scope
		orgUnitRequest.PolicyMeta.ClusterId = nil
		orgUnitRequest.PolicyMeta.DepartmentId = nil
		orgUnitRequest.PolicyMeta.ProjectId = nil

		emptyPolicy := NewEmptyPolicy(ctx, orgUnitRequest, relevantOrgUnits[idx:])
		emptyPolicy.DeletedAt = gorm.DeletedAt{
			Time:  time.Now(),
			Valid: true,
		}

		// attempt to add and ignore conflicts
		dbErr := dao.AddPolicy(ctx, db, emptyPolicy)
		if dbErr != nil {
			if dbErr.Code != http.StatusConflict {
				return dbErr
			}
		}
	}
	return nil
}
