package tree

import (
	"context"
	"errors"
	"fmt"
	"github.com/run-ai/runai/control-plane/policy-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/clients/cluster"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/database/dao"
	"github.com/run-ai/runai/runai-cluster/common/cluster-capabilities/checker"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
	"net/http"
)

// before doing changes to a policy, we need to check if there are clusters that need approval
// for deleting their policies. these are clusters of version < 2.21. in those clusters CP workloads
// are validates VS the cluster policy, which will cause inconsistency as we no longer sync policies
// to the cluster. to avoid this we delete the cluster policies of those clusters
func handleClusterDeletionApproval(ctx context.Context, request *GenericPolicyRequest) *api.Error {
	deletionApprovedFor, dbErr := dao.ListApprovedDeletionsByTenant(ctx)
	if dbErr != nil {
		return dbErr
	}

	clusterInfo, err := cluster.GetClustersInfo(ctx, request.PolicyMeta.ClusterId)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("[%d] get cluster info failed: %v", request.TenantId, err)
		return api.NewError(http.StatusServiceUnavailable, "failed to check cluster compatibility", errors.New(err.Message))
	}

	idsToMarkApproval := make([]string, 0, len(clusterInfo))
	idsToDeleteApproval := make([]string, 0, len(clusterInfo))

	for clusterId, clusterVersion := range clusterInfo {
		clusterCapabilities, err := checker.NewClusterCapabilitiesCheckerForVersion(clusterVersion)
		if err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("[%d] parse cluster info failed: %v", request.TenantId, err)
			return api.NewError(http.StatusServiceUnavailable, "failed to check cluster compatibility", err)
		}

		supported, checkErr := clusterCapabilities.ValidateClusterCapability(ctx, checker.MinVersionOfClusterPolicyBypassForCPWorkloads)
		if checkErr != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("[%d] check cluster policy support failed: %v", request.TenantId, checkErr)
			return api.NewError(http.StatusServiceUnavailable, "failed to check cluster compatibility", checkErr)
		}

		// if deletion was already approved for this cluster, but it has been upgraded to 2.21 and above, we need to delete
		// the approval so the cluster will start reporting "no changes" instead of simulating deletions.
		if deletionApprovedFor[clusterId] != nil {
			if supported {
				idsToDeleteApproval = append(idsToDeleteApproval, clusterId)
			}
			continue
		}

		// if deletion approval flag is sent, need to update all the old clusters that were not yet approved
		if request.ApproveClusterDeletion {
			if !supported {
				idsToMarkApproval = append(idsToMarkApproval, clusterId)
			}
			continue
		}

		if !supported {
			var errStr string
			if len(clusterInfo) > 1 {
				errStr = fmt.Sprintf("at least one afftected clusters (%s) is running version 2.20 or earlier. In such case, the platform needs your approval to delete the policy from the affected clusters.", clusterId)
			} else {
				errStr = fmt.Sprintf("cluster (%s) is running version 2.20 or earlier. In such case, the platform needs your approval to delete the policy from the affected cluster.", clusterId)
			}

			return api.NewError(http.StatusPreconditionFailed, errStr, nil)
		}
	}

	if len(idsToMarkApproval) > 0 {
		dao.UpdateApprovedDeletionByIds(ctx, idsToMarkApproval, false)
	}

	if len(idsToDeleteApproval) > 0 {
		dao.DeleteApprovedDeletionByIds(ctx, idsToDeleteApproval)
	}

	return nil
}
