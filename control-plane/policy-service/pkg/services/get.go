package services

import (
	"context"
	"errors"
	"fmt"
	backend "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/policy-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/database/dao"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/database/dbmodel"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/tree"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/auth"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
	"net/http"
	"strconv"
)

type getOptions struct {
	bypassAuth bool // for internal invocations of APIs
}

// get policy takes care both workspace and training.
func getPolicy(ctx context.Context, request *tree.GenericPolicyRequest, options getOptions) (*dbmodel.PolicyWithOptionalEffective, *api.Error) {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof(">>> get %s policy for %s started", request.WorkloadType, request.PolicyMeta.Scope)

	if request.PolicyMeta.Scope == api.SCOPETYPE_SYSTEM {
		return nil, api.NewValidationError("system scope is not supported for policies")
	}

	request.Requester = auth.SubjectIdOf(ctx)
	request.TenantId = auth.TenantIdOf(ctx)

	if !options.bypassAuth {
		scopeParams, scopeErr := NewScopeInfoFromMeta(request.PolicyMeta)
		if scopeErr != nil {
			return nil, scopeErr
		}
		// Fallback policies cascade through the hierarchy, so any user with policies:read (already
		// verified by AuthorizeMiddleware) may read them. We only verify same-tenant ownership.
		if request.WorkloadType == tree.WorkloadTypeFallback {
			if rbacErr := validateFallbackSameTenant(ctx, scopeParams); rbacErr != nil {
				return nil, rbacErr
			}
		} else { // Non-fallback policies require full scope-level RBAC
			rbacErr := auth.ValidateAccessByResourceTypeAndAction(ctx, scopeParams, backend.RESOURCETYPE_POLICIES, backend.ACTION_READ, nonPrivate)
			if rbacErr != nil {
				return nil, api.NewError(int(rbacErr.Code), rbacErr.Message, errors.New(rbacErr.Details))
			}
		}
	}

	//  in no creation mode we first look on the db and if there is no given policy for this tenant
	//  we return not found error. this is done because frontend sends a lot of get policy requests
	//  for tenant that do not have any policy and we do not want to create trees for those tenant
	//  until they create their first policy.
	if request.NoCreationMode {
		count, err := dao.CountGivenPoliciesOfTenant(ctx)
		if err != nil {
			return nil, err
		}

		if count == 0 {
			return nil, api.NewError(http.StatusNotFound, "the tenant does not have any policies yet", nil)
		}
	}

	// Get hold of the relevant hierarchy tree
	hierarchyTree, treeInitialized := tree.HierarchyTrees.Get(ctx, auth.TenantIdOf(ctx), request.WorkloadType)
	if !treeInitialized {
		err := hierarchyTree.InitializeTenant(ctx, request.TenantId, request.WorkloadType)
		if err != nil {
			return nil, err
		}
	}

	fallbackTree, err := getFallbackTree(ctx, auth.TenantIdOf(ctx), request.WorkloadType)
	if err != nil {
		return nil, err
	}

	if fallbackTree != nil {
		// make sure that all the policies of the fallback exist
		_, err = fallbackTree.HandleGetRequest(ctx, request.ToFallback())
		if err != nil {
			return nil, err
		}
	}

	return hierarchyTree.HandleGetRequest(ctx, request)
}

func projectIdToInt(projectId string) (int, error) {
	if projectId == "" {
		return 0, nil
	}

	return strconv.Atoi(projectId)
}

func projectIdToNum(projectId string) (int32, *api.Error) {
	projectIdNumeric, err := projectIdToInt(projectId)
	if err != nil {
		return 0, api.NewValidationError(fmt.Sprintf("project id '%s' is not numeric", projectId))
	}
	return int32(projectIdNumeric), nil
}

func getFallbackTree(ctx context.Context, tenantId int32, requestWorkloadType tree.WorkloadType) (*tree.HierarchyTree, *api.Error) {
	if requestWorkloadType == tree.WorkloadTypeFallback {
		return nil, nil
	}

	fallbackTree, fallbackTreeInitialized := tree.HierarchyTrees.Get(ctx, auth.TenantIdOf(ctx), tree.WorkloadTypeFallback)
	if !fallbackTreeInitialized {
		err := fallbackTree.InitializeTenant(ctx, tenantId, tree.WorkloadTypeFallback)
		if err != nil {
			return nil, err
		}
	}

	return fallbackTree, nil
}
