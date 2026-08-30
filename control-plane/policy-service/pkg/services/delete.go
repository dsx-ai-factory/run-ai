package services

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	backend "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/policy-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/database/dao"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/tree"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/auth"
	events "github.com/run-ai/runai/runai-application-packages/async-events"
	dao_utils "github.com/run-ai/runai/runai-common-packages/dao-utils"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
)

const noNeedClusterApproval = false

// HandleProjectDeleteRequest handles an incoming project deletion request
func HandleProjectDeleteRequest(ctx context.Context, req *events.ProjectDeleteRequest) error {
	logging.Infof("Handling project delete request: %v", req)
	if req.LeaveOrphans {
		return nil
	}

	projectIdNumeric, err := projectIdToInt(req.ProjectId)
	if err != nil {
		return err
	}

	for _, workloadType := range tree.AllWorkloadTypes {
		request := tree.NewGenericPolicyRequest(workloadType, string(api.SCOPETYPE_PROJECT), "", int32(projectIdNumeric), req.ClusterId, false, noNeedClusterApproval)
		// Skip authorization since this is triggered by a pubsub delete event, and the user might already be deleted.
		delErr := internalDeletePolicy(ctx, request, tree.DeleteRequestOpt{DeleteSubTree: false, SkipAuthorization: true})
		// if the policy wasn't found, it ok, there wasn't a policy of that type, continue to the next policy type
		if delErr != nil && delErr.Message != tree.PolicyNotFoundErrMsg {
			return fmt.Errorf("failed to delete policy for workload type %s: %s", workloadType, delErr.Details)
		}
	}

	// Drop the tenant's cached breakdowns. Scoping by project would require
	// joining policy → cache; flushing the whole tenant is simpler and the
	// remaining caches will be recomputed lazily on next read.
	if err := dropTenantCaches(ctx); err != nil {
		return err
	}

	return nil
}

// HandleDeleteRequest handles an incoming cluster/tenant deletion request
func HandleDeleteRequest(ctx context.Context, scopeType api.ScopeType, scopeId string, clusterIdForSoftDeleteReq string) error {
	logging.Infof("Handling %v deleted request: %v", scopeType, scopeId) // In order to keep the policy tree structure in sync, first soft delete the policies
	for _, workloadType := range tree.AllWorkloadTypes {
		request := tree.NewGenericPolicyRequest(workloadType, string(scopeType), "", 0, clusterIdForSoftDeleteReq, false, noNeedClusterApproval)
		// Skip authorization since this is triggered by a pubsub delete event, and the user might already be deleted.
		delErr := internalDeletePolicy(ctx, request, tree.DeleteRequestOpt{DeleteSubTree: true, SkipAuthorization: true})
		// If the policy wasn't found, it ok, there wasn't a policy of that type, continue to the next policy type
		if delErr != nil && delErr.Message != tree.PolicyNotFoundErrMsg {
			return fmt.Errorf("failed to delete policy for workload type %s: %s", workloadType, delErr.Message)
		}
	}

	// After soft deleting all tenant's policies, need to evacuate them from the db.
	// in normal flow they evacuate when syncing with the cluster, however in case of cluster deletion, the cluster is
	// already disconnected, better to evacuate them now.
	err := dao.DeleteObsoletePoliciesByScope(ctx, dao_utils.DB().WithContext(ctx), scopeType, scopeId)
	if err != nil {
		return fmt.Errorf("failed to delete policies for %s %s: %s", scopeType, scopeId, err.Message)
	}
	return nil
}

// HandleClusterDeleteRequest handles an incoming cluster deletion request
func HandleClusterDeleteRequest(ctx context.Context, req *events.ClusterDelete) error {
	if err := HandleDeleteRequest(ctx, api.SCOPETYPE_CLUSTER, req.ClusterId, req.ClusterId); err != nil {
		return err
	}

	// Drop the tenant's cached breakdowns. Scoping by cluster would require
	// joining policy → cache; flushing the whole tenant is simpler and the
	// remaining caches will be recomputed lazily on next read.
	return dropTenantCaches(ctx)
}

// HandleTenantDeleteRequest handles an incoming tenant deletion request
func HandleTenantDeleteRequest(ctx context.Context, req *events.TenantDelete) error {
	if err := HandleDeleteRequest(ctx, api.SCOPETYPE_TENANT, req.TenantId, ""); err != nil {
		return err
	}

	tenantId := auth.TenantIdOf(ctx)
	db := dao_utils.DB().WithContext(ctx)

	if apiErr := dao.DeleteMappingsOfTenant(ctx, db, tenantId); apiErr != nil {
		return fmt.Errorf("failed to delete mappings for tenant %s: %s", req.TenantId, apiErr.Message)
	}

	if apiErr := dao.DeleteEntityCachesOfTenant(ctx, db, tenantId); apiErr != nil {
		return fmt.Errorf("failed to delete entity caches for tenant %s: %s", req.TenantId, apiErr.Message)
	}

	return nil
}

// dropTenantCaches removes every cached breakdown (and its dependency rows)
// owned by the tenant resolved from ctx. used by cluster and project deletion
// handlers as a coarse but simple invalidation: any policy whose breakdown was
// cached will recompute it on the next read.
func dropTenantCaches(ctx context.Context) error {
	tenantId := auth.TenantIdOf(ctx)
	db := dao_utils.DB().WithContext(ctx)
	if apiErr := dao.DeleteEntityCachesOfTenant(ctx, db, tenantId); apiErr != nil {
		return fmt.Errorf("failed to delete entity caches for tenant %d: %s", tenantId, apiErr.Message)
	}
	return nil
}

// delete policy takes care both workspace and training.
func deletePolicy(ctx context.Context, request *tree.GenericPolicyRequest) api.ImplResponse {
	err := internalDeletePolicy(ctx, request, tree.DeleteRequestOpt{DeleteSubTree: false})
	if err != nil {
		return NewErrorResponse(err)
	}
	return api.Response(http.StatusNoContent, nil)
}

func internalDeletePolicy(ctx context.Context, request *tree.GenericPolicyRequest, opt tree.DeleteRequestOpt) *api.Error {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof(">>> delete %s policy for %s started", request.WorkloadType, request.PolicyMeta.Scope)

	if request.PolicyMeta.Scope == api.SCOPETYPE_SYSTEM {
		return api.NewValidationError("system scope is not supported for policies")
	}

	request.Requester = auth.SubjectIdOf(ctx)
	request.TenantId = auth.TenantIdOf(ctx)

	scopeParams, scopeErr := NewScopeInfoFromMeta(request.PolicyMeta)
	if scopeErr != nil {
		return scopeErr
	}

	if !opt.SkipAuthorization {
		rbacErr := auth.ValidateAccessByResourceTypeAndAction(ctx, scopeParams, backend.RESOURCETYPE_POLICIES, backend.ACTION_DELETE, nonPrivate)
		if rbacErr != nil {
			return api.NewError(int(rbacErr.Code), rbacErr.Message, errors.New(rbacErr.Details))
		}
	}

	// Get hold of the relevant hierarchy tree
	hierarchyTree, treeInitialized := tree.HierarchyTrees.Get(ctx, auth.TenantIdOf(ctx), request.WorkloadType)
	if !treeInitialized {
		err := hierarchyTree.InitializeTenant(ctx, request.TenantId, request.WorkloadType)
		if err != nil {
			return err
		}
	}

	fallbackTree, err := getFallbackTree(ctx, auth.TenantIdOf(ctx), request.WorkloadType)
	if err != nil {
		return err
	}

	if fallbackTree != nil {
		// make sure that all the policies of the fallback exist
		_, err = fallbackTree.HandleGetRequest(ctx, request.ToFallback())
		if err != nil {
			if err.Code == http.StatusNotFound &&
				err.Message != tree.PolicyNotFoundErrMsg {
				logging.WithCtx(runaictx.LoggerHook(ctx)).Infof(
					"scope entity not found during fallback get, skipping fallback for orphaned policy delete")
			} else {
				return err
			}
		}
	}

	err = hierarchyTree.HandleDeleteRequest(ctx, request, opt)
	if err != nil {
		return err
	}

	return nil
}
