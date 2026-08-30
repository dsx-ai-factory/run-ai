package services

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/run-ai/runai/control-plane/policy-service/pkg/database/dbmodel"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/scopes"
	"github.com/run-ai/runai/runai-application-packages/convutils"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"

	"github.com/run-ai/runai/control-plane/policy-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/auth"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/clients/orgunit"
	scope "github.com/run-ai/runai/runai-application-packages/scopeinfo"
)

// validateFallbackSameTenant verifies that the requested fallback policy resource belongs to
// the caller's tenant. For tenant scope there is no resource ID to look up — cross-tenant access
// is impossible because the API has no tenant parameter and data access (HierarchyTrees.Get) is
// always scoped to the JWT tenant. For other scopes we resolve ancestors and compare tenant IDs.
func validateFallbackSameTenant(ctx context.Context, scopeInfo *scope.ScopeInfo) *api.Error {
	if scopeInfo.Scope == scope.Tenant {
		return nil
	}

	jwtTenantId := fmt.Sprintf("%d", auth.TenantIdOf(ctx))
	ancestors, err := orgunit.Get().Client().GetAncestors(ctx, string(scopeInfo.Scope), scopeInfo.ScopeId)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("failed to obtain ancestors of %s/%s: %s", scopeInfo.Scope, scopeInfo.ScopeId, err.Message)
		return api.NewError(http.StatusForbidden, auth.ForbiddenError.Message, nil)
	}

	if ancestors.TenantId != jwtTenantId {
		return api.NewError(http.StatusForbidden, auth.ForbiddenError.Message, nil)
	}
	return nil
}

func NewScopeInfoFromMeta(meta api.PolicyCreationFields) (*scope.ScopeInfo, *api.Error) {
	switch meta.Scope {
	case api.SCOPETYPE_TENANT:
		return scope.NewTenantScope(), nil
	case api.SCOPETYPE_CLUSTER:
		if meta.ClusterId == nil {
			return nil, api.NewValidationError("mandatory clusterId is not provided")
		}
		return scope.NewClusterScope(*meta.ClusterId), nil
	case api.SCOPETYPE_DEPARTMENT:
		if meta.DepartmentId == nil {
			return nil, api.NewValidationError("mandatory departmentId is not provided")
		}
		return scope.NewDepartmentScope(convutils.Deref(meta.DepartmentId, "")), nil
	default:
		if meta.ProjectId == nil {
			return nil, api.NewValidationError("mandatory projectId is not provided")
		}
		return scope.NewProjectScope(fmt.Sprintf("%d", convutils.Deref(meta.ProjectId, 0))), nil
	}
}

// NewScopeInfoFromPolicy builds scope info from a stored Policy. A stored
// policy may carry every ancestor id (a department-scoped policy, for example,
// still has the owning ClusterId populated), but NewScopeInfo is an input
// validator that rejects any id that does not match the requested scope, so we
// must forward only the id that matches the policy's own scope. The integer
// project id is converted to the string form expected by NewScopeInfo.
func NewScopeInfoFromPolicy(ctx context.Context, p *dbmodel.Policy) (*scopes.OrgUnit, *scope.ScopeInfo, *api.Error) {
	var clusterId, departmentId, projectId *string

	switch p.Scope {
	case api.SCOPETYPE_SYSTEM:
		// no info to collect
	case api.SCOPETYPE_TENANT:
		// no info to collect
	case api.SCOPETYPE_CLUSTER:
		clusterId = p.ClusterId
	case api.SCOPETYPE_DEPARTMENT:
		departmentId = p.DepartmentId
	case api.SCOPETYPE_PROJECT:
		if p.ProjectId != nil {
			projectId = convutils.Ref(fmt.Sprintf("%d", *p.ProjectId))
		}
	default:
		return nil, nil, api.NewValidationError(fmt.Sprintf("scope type %s is not supported", p.Scope))
	}

	return NewScopeInfo(ctx, p.Scope, clusterId, departmentId, projectId)
}

func validateScopeField(field *string, required bool, fieldName, scopeName string) *api.Error {
	if required && field == nil {
		return api.NewValidationError(fmt.Sprintf("mandatory %s is not provided", fieldName))
	}
	if !required && field != nil {
		return api.NewValidationError(fmt.Sprintf("%s should not be specified for %s scope", fieldName, scopeName))
	}
	return nil
}

func NewScopeInfo(ctx context.Context, scopeType api.ScopeType, clusterId *string, departmentId *string, projectId *string) (*scopes.OrgUnit, *scope.ScopeInfo, *api.Error) {
	type scopeSpec struct {
		requireCluster, requireDept, requireProject bool

		create func() *scope.ScopeInfo
	}

	specs := map[api.ScopeType]scopeSpec{
		api.SCOPETYPE_SYSTEM:     {create: scope.NewSystemScope},
		api.SCOPETYPE_TENANT:     {create: scope.NewTenantScope},
		api.SCOPETYPE_CLUSTER:    {requireCluster: true, create: func() *scope.ScopeInfo { return scope.NewClusterScope(*clusterId) }},
		api.SCOPETYPE_DEPARTMENT: {requireDept: true, create: func() *scope.ScopeInfo { return scope.NewDepartmentScope(*departmentId) }},
		api.SCOPETYPE_PROJECT:    {requireProject: true, create: func() *scope.ScopeInfo { return scope.NewProjectScope(*projectId) }},
	}

	spec, ok := specs[scopeType]
	if !ok {
		return nil, nil, api.NewValidationError(fmt.Sprintf("unknown scope type: %s", scopeType))
	}

	scopeName := strings.ToLower(string(scopeType))
	if err := validateScopeField(clusterId, spec.requireCluster, "clusterId", scopeName); err != nil {
		return nil, nil, err
	}
	if err := validateScopeField(departmentId, spec.requireDept, "departmentId", scopeName); err != nil {
		return nil, nil, err
	}
	if err := validateScopeField(projectId, spec.requireProject, "projectId", scopeName); err != nil {
		return nil, nil, err
	}

	var projectIdNum *int32
	if scopeType == api.SCOPETYPE_PROJECT {
		asNum, numErr := projectIdToNum(*projectId)
		if numErr != nil {
			return nil, nil, numErr
		}
		projectIdNum = &asNum
	}

	orgUnit, err := scopes.MakeOrgUnit(scopeType, auth.TenantIdOf(ctx), clusterId, departmentId, projectIdNum)
	if err != nil {
		return nil, nil, err
	}

	return &orgUnit, spec.create(), nil
}
