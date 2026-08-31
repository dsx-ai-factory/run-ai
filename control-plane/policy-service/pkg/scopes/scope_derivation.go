/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package scopes

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/run-ai/runai/control-plane/workloads-and-policies/clients/orgunit"
	"github.com/run-ai/runai/runai-application-packages/jsonutils"

	"github.com/run-ai/runai/control-plane/workloads-and-policies/auth"

	"github.com/run-ai/runai/control-plane/policy-service/generated/pkg/api"
	"github.com/run-ai/runai/runai-application-packages/convutils"
	scope "github.com/run-ai/runai/runai-application-packages/scopeinfo"
	assetsService "github.com/run-ai/runai/runai-clients/assets-service-client/generated"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
)

const (
	FallbackPolicyAsType = "Fallback"
)

var (
	sourceOfRuleCracker   = jsonutils.NewJsonCracker(api.SourceOfRule{})
	jsonFieldScope        = sourceOfRuleCracker.NameOf(&sourceOfRuleCracker.Struct().Scope)
	jsonFieldProjectId    = sourceOfRuleCracker.NameOf(&sourceOfRuleCracker.Struct().ProjectId)
	jsonFieldClusterId    = sourceOfRuleCracker.NameOf(&sourceOfRuleCracker.Struct().ClusterId)
	jsonFieldDepartmentId = sourceOfRuleCracker.NameOf(&sourceOfRuleCracker.Struct().DepartmentId)
	jsonFieldIsFallback   = sourceOfRuleCracker.NameOf(&sourceOfRuleCracker.Struct().IsFallback)
)

// OrgUnit holds the key to an organizational unit, composed of its scope (tenant, department, project) and resource id
type OrgUnit struct {
	Scope api.ScopeType
	Id    string
}

func (ou OrgUnit) IsSameScopeAs(another OrgUnit) bool {
	return ou.Scope == another.Scope && ou.Id == another.Id
}

func (ou OrgUnit) String() string {
	return fmt.Sprintf("%s:%s", ou.Scope, ou.Id)
}

// NameForCluster composes policy name for the cluster. in the cluster, we prefer not to rely on user name
// (because user can change it all the time) but to call the policy based on the org unit.
func (ou OrgUnit) NameForCluster() string {
	if ou.Id == "" || ou.Scope == api.SCOPETYPE_CLUSTER {
		return string(ou.Scope)
	} else {
		return fmt.Sprintf("%s-%s", ou.Scope, ou.Id)
	}
}

func MakeOrgUnit(scope api.ScopeType, tenantId int32, clusterId *string, departmentId *string, projectId *int32) (OrgUnit, *api.Error) {
	result := OrgUnit{
		Scope: scope,
	}

	switch scope {
	case api.SCOPETYPE_TENANT:
		// no scope id for tenant
		result.Id = fmt.Sprintf("%d", tenantId)
	case api.SCOPETYPE_CLUSTER:
		if clusterId == nil {
			return OrgUnit{}, api.NewError(http.StatusBadRequest, "clusterId is mandatory for cluster scoped policy", nil)
		}
		result.Id = convutils.Deref(clusterId, "")
	case api.SCOPETYPE_DEPARTMENT:
		if departmentId == nil {
			return OrgUnit{}, api.NewError(http.StatusBadRequest, "departmentId is mandatory for department scoped policy", nil)
		}
		result.Id = convutils.Deref(departmentId, "")
	case api.SCOPETYPE_PROJECT:
		if projectId == nil {
			return OrgUnit{}, api.NewError(http.StatusBadRequest, "projectId is mandatory for project scoped policy", nil)
		}
		result.Id = fmt.Sprintf("%d", *projectId)
	default:
		return OrgUnit{}, api.NewError(http.StatusBadRequest, fmt.Sprintf("unsupported scope '%s'", scope), nil)
	}

	return result, nil
}

func NewOrgUnitFromCreationFields(creationFields *api.PolicyCreationFields, tenantId int32) (OrgUnit, *api.Error) {
	return MakeOrgUnit(creationFields.Scope, tenantId, creationFields.ClusterId, creationFields.DepartmentId, creationFields.ProjectId)
}

func NewOrgUnitFromAssetMeta(assetMeta *assetsService.AssetMeta, tenantId int32) (OrgUnit, *api.Error) {
	return MakeOrgUnit(api.ScopeType(assetMeta.Scope), tenantId, assetMeta.ClusterId.Get(), assetMeta.DepartmentId.Get(), assetMeta.ProjectId.Get())
}

// NewOrgUnitFromString convert org unit key generated from the String() function above back to OrgUnit object
func NewOrgUnitFromString(orgUnitKey *string) (OrgUnit, *api.Error) {
	if orgUnitKey == nil {
		return OrgUnit{}, api.NewError(http.StatusServiceUnavailable, "nil org unit key receives", nil)
	}
	parts := strings.Split(*orgUnitKey, ":")
	if len(parts) != 2 || len(parts[1]) == 0 {
		return OrgUnit{}, api.NewError(http.StatusServiceUnavailable, fmt.Sprintf("malformed org unit key %s, expecting scope:id", *orgUnitKey), nil)
	}
	switch parts[0] {
	case string(scope.Project):
		_, err := strconv.Atoi(parts[1])
		if err != nil {
			return OrgUnit{}, api.NewError(http.StatusServiceUnavailable, fmt.Sprintf("project id '%s' is not numeric", *orgUnitKey), nil)
		}
		return OrgUnit{
			Scope: api.ScopeType(parts[0]),
			Id:    parts[1],
		}, nil
	case string(scope.Tenant), string(api.SCOPETYPE_DEPARTMENT), string(api.SCOPETYPE_CLUSTER):
		return OrgUnit{
			Scope: api.ScopeType(parts[0]),
			Id:    parts[1],
		}, nil
	default:
		return OrgUnit{}, api.NewError(http.StatusServiceUnavailable, fmt.Sprintf("unsupported scope %s in %s", parts[0], *orgUnitKey), nil)
	}
}

func NewFromJson(payload map[string]interface{}) OrgUnit {
	result := OrgUnit{
		Scope: api.ScopeType(fmt.Sprintf("%v", payload[jsonFieldScope])),
	}

	switch result.Scope {
	case api.SCOPETYPE_CLUSTER:
		result.Id = payload[jsonFieldClusterId].(string)
	case api.SCOPETYPE_DEPARTMENT:
		result.Id = payload[jsonFieldDepartmentId].(string)
	case api.SCOPETYPE_PROJECT:
		result.Id = strconv.Itoa(payload[jsonFieldProjectId].(int))
	}

	return result
}

func (ou OrgUnit) Json() (map[string]interface{}, *api.Error) {
	result := map[string]interface{}{
		jsonFieldScope: ou.Scope,
	}

	var err error

	switch ou.Scope {
	case api.SCOPETYPE_PROJECT:
		result[jsonFieldProjectId], err = strconv.Atoi(ou.Id)
		if err != nil {
			return nil, api.NewError(http.StatusServiceUnavailable, fmt.Sprintf("non numeric project id %s", ou.Id), err)
		}
	case api.SCOPETYPE_CLUSTER:
		result[jsonFieldClusterId] = ou.Id
	case api.SCOPETYPE_DEPARTMENT:
		result[jsonFieldDepartmentId] = ou.Id
	}
	return result, nil
}

func (ou OrgUnit) AsSourceOfRule(workloadType string) (map[string]interface{}, *api.Error) {
	result, err := ou.Json()
	if err != nil {
		return nil, err
	}
	if workloadType == FallbackPolicyAsType {
		result[jsonFieldIsFallback] = true
	}

	return result, nil
}

func (ou OrgUnit) ClusterId() *string {
	if ou.Scope == api.SCOPETYPE_CLUSTER {
		return &ou.Id
	}
	return nil
}

func (ou OrgUnit) DepartmentId() *string {
	if ou.Scope == api.SCOPETYPE_DEPARTMENT {
		return &ou.Id
	}
	return nil
}

func (ou OrgUnit) ProjectId(ctx context.Context) *int32 {
	if ou.Scope == api.SCOPETYPE_PROJECT {
		projectId, err := strconv.Atoi(ou.Id)
		if err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("[%d] invalid project id '%s' received in asset notidication", auth.TenantIdOf(ctx), ou.Id)
			return nil
		}
		return convutils.Ref(int32(projectId))
	}
	return nil
}

// Ancestors receives scope of policy and derives all the ancestors of the org unit that this policy belongs to.
// For example, policy of project 5, belonging to department "x" in tenant 10 will return:
// [ project, 5 ] ; [ department, x ] ; [ tenant 10 ]
// Note that project does not necessarily have a parent department, so it can also return [project] ; [tenant] hierarchy.
func (ou OrgUnit) Ancestors(ctx context.Context) ([]OrgUnit, *api.Error) {

	result := make([]OrgUnit, 0)
	logMessage := ""

	tenantId := auth.TenantIdOf(ctx)

	if ou.Scope != api.SCOPETYPE_TENANT {
		//  if current scope is project/department/cluster add it to the result as is
		logMessage += fmt.Sprintf(" -> %s", ou.String())
		result = append(result, ou)
	} else {
		//   for tenant scope make sure the scope id will hold the correct tenant id
		ou.Id = fmt.Sprintf("%d", tenantId)
	}

	//
	//  for project/department we need information from the backend
	//
	if ou.Scope == api.SCOPETYPE_PROJECT || ou.Scope == api.SCOPETYPE_DEPARTMENT {
		backendResponse, backendError := orgunit.Get().Client().GetAncestors(ctx, string(ou.Scope), ou.Id)
		if backendError != nil {
			return nil, &api.Error{
				Code:    backendError.Code,
				Message: backendError.Message,
			}
		}

		//  for project - add the parent department
		if ou.Scope == api.SCOPETYPE_PROJECT {
			logMessage += fmt.Sprintf("/%s:%s", ou.Scope, ou.Id)

			//
			// Add department id, if the project has a parent department
			//
			if backendResponse.DepartmentId.Get() != nil {
				department := OrgUnit{
					Scope: api.SCOPETYPE_DEPARTMENT,
					Id:    convutils.Deref(backendResponse.DepartmentId.Get(), ""),
				}
				result = append(result, department)
				logMessage += fmt.Sprintf(" -> %s", department.String())
			}
		}

		//
		//   add the cluster id
		//
		if !convutils.HasValue(backendResponse.ClusterUuid) {
			return nil, api.NewError(http.StatusServiceUnavailable, fmt.Sprintf("unknown parent cluster ID of %s", ou.String()), nil)
		}
		cluster := OrgUnit{
			Scope: api.SCOPETYPE_CLUSTER,
			Id:    convutils.Deref(backendResponse.ClusterUuid.Get(), ""),
		}
		result = append(result, cluster)
		logMessage += fmt.Sprintf(" -> %s", cluster.String())
	}

	//
	//   always add the tenant id (unless it is system)
	//
	if ou.Scope != api.SCOPETYPE_SYSTEM {
		tenant := OrgUnit{
			Scope: api.SCOPETYPE_TENANT,
			Id:    fmt.Sprintf("%d", tenantId),
		}
		result = append(result, tenant)
		logMessage += fmt.Sprintf(" -> %s", tenant.String())
	}

	result = append(result, OrgUnit{
		Scope: api.SCOPETYPE_SYSTEM,
	})

	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("--> hierarchy of tenant:%d%s", tenantId, logMessage)

	return result, nil
}
