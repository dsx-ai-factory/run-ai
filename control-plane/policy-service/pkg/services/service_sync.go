package services

import (
	"context"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/database/dao"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/tree"
	"github.com/run-ai/runai/runai-application-packages/convutils"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
	"net/http"

	"github.com/run-ai/runai/control-plane/policy-service/generated/pkg/api"
)

type SyncApiService struct{}

func NewSyncApiService() api.SynchronizationAPIServicer {
	return &SyncApiService{}
}

func emptyResponse(updateCount int64) *api.PolicySyncResponse {
	return &api.PolicySyncResponse{
		RequestInfo: api.PolicySyncResponseRequestInfo{
			After: updateCount,
		},
		ResponseInfo: api.PolicySyncResponseResponseInfo{
			LowestUpdate:  updateCount,
			HighestUpdate: updateCount,
		},
		SystemInfo: api.PolicySyncResponseSystemInfo{
			HighestUpdate:  updateCount,
			UpdatesPerBulk: 50,
		},
		Policies: make([]api.PolicySyncResponsePoliciesInner, 0),
	}
}

func (sa *SyncApiService) SyncTrainingPolicies(ctx context.Context, clusterId string, updateCount int64) (api.ImplResponse, error) {
	return sa.SimulateDeletionOrNoChange(ctx, clusterId, tree.WorkloadTypeTraining, updateCount)
}

func (sa *SyncApiService) SyncWorkspacePolicies(ctx context.Context, clusterId string, updateCount int64) (api.ImplResponse, error) {
	return sa.SimulateDeletionOrNoChange(ctx, clusterId, tree.WorkloadTypeWorkspace, updateCount)
}

func (sa *SyncApiService) SyncDistributedPolicies(ctx context.Context, clusterId string, updateCount int64) (api.ImplResponse, error) {
	return sa.SimulateDeletionOrNoChange(ctx, clusterId, tree.WorkloadTypeDistributed, updateCount)
}

func (sa *SyncApiService) SyncInferencePolicies(ctx context.Context, clusterId string, updateCount int64) (api.ImplResponse, error) {
	return sa.SimulateDeletionOrNoChange(ctx, clusterId, tree.WorkloadTypeInference, updateCount)
}

func (sa *SyncApiService) SyncDistributedInferencePolicies(ctx context.Context, clusterId string, updateCount int64) (api.ImplResponse, error) {
	return sa.SimulateDeletionOrNoChange(ctx, clusterId, tree.WorkloadTypeDistributedInference, updateCount)
}

func (sa *SyncApiService) SyncRequestTrainingPolicies(ctx context.Context, clusterId string, updateCount int64, syncReport api.ClusterSyncReport) (api.ImplResponse, error) {
	return sa.SimulateDeletionOrNoChange(ctx, clusterId, tree.WorkloadTypeTraining, updateCount)
}

func (sa *SyncApiService) SyncRequestWorkspacePolicies(ctx context.Context, clusterId string, updateCount int64, syncReport api.ClusterSyncReport) (api.ImplResponse, error) {
	return sa.SimulateDeletionOrNoChange(ctx, clusterId, tree.WorkloadTypeWorkspace, updateCount)
}

func (sa *SyncApiService) SyncRequestDistributedPolicies(ctx context.Context, clusterId string, updateCount int64, syncReport api.ClusterSyncReport) (api.ImplResponse, error) {
	return sa.SimulateDeletionOrNoChange(ctx, clusterId, tree.WorkloadTypeDistributed, updateCount)
}

func (sa *SyncApiService) SyncRequestInferencePolicies(ctx context.Context, clusterId string, updateCount int64, syncReport api.ClusterSyncReport) (api.ImplResponse, error) {
	return sa.SimulateDeletionOrNoChange(ctx, clusterId, tree.WorkloadTypeInference, updateCount)
}

func (sa *SyncApiService) SyncRequestDistributedInferencePolicies(ctx context.Context, clusterId string, updateCount int64, syncReport api.ClusterSyncReport) (api.ImplResponse, error) {
	return sa.SimulateDeletionOrNoChange(ctx, clusterId, tree.WorkloadTypeDistributedInference, updateCount)
}

func (sa *SyncApiService) SimulateDeletionOrNoChange(ctx context.Context, clusterId string, workloadType tree.WorkloadType, fromCount int64) (api.ImplResponse, error) {
	// check if the cluster requested any policies in the past, otherwise, just return the empty response
	if fromCount == 0 {
		return api.ImplResponse{Code: http.StatusOK, Body: emptyResponse(fromCount)}, nil
	}

	//-------------------------------------------------------------------------------------------------
	// There are three cases for policy creation/update:
	// 1) approval_deletions does not contain record for this cluster ->
	//	  deletion is not needed (new cluster) or not permitted (old cluster that did not approve yet)
	// 2) approval_deletions exist for this cluster
	//    (a) avoid_deletions = false -> deletions ia permitted (user approved it in the "popup")
	//    (b) avoid_deletions = true  -> deletions are not permitted (user chose cancel in the popup, and runai CS
	//                                   manually added this record with avoid=true)
	//
	// Deletion of policies in the CP are always deleted from the cluster.
	//-------------------------------------------------------------------------------------------------

	// load the information from the table. in case of (1) the returned value is nil
	approvedDeletionsInfo, dbErr := dao.GetApprovedDeletionsInfoOfCluster(ctx, clusterId)
	if dbErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to validate admin approval for cluster deletion for cluster %s", clusterId)
		return api.ImplResponse{Code: http.StatusOK, Body: emptyResponse(fromCount)}, nil
	}

	if approvedDeletionsInfo == nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("no deletion approval for cluster %s, will not get any deletions", clusterId)
	} else {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("deletion approval record for cluster %s, avoid deletions? %v", clusterId, approvedDeletionsInfo.ApprovedDeletionsKey)
	}

	// these are the conditions for (2a).
	deleteUpdatedPoliciesFromTheCluster := approvedDeletionsInfo != nil && !approvedDeletionsInfo.AvoidDeletionsOfObsoletePolicies

	// find the policies that were deleted or updated
	modifiedPolicies, dbErr := dao.FindPoliciesForSync(ctx, string(workloadType), clusterId, fromCount)
	if dbErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to find changed policies for cluster %s", clusterId)
		return api.ImplResponse{Code: http.StatusOK, Body: emptyResponse(fromCount)}, nil
	}

	result := emptyResponse(fromCount)

	for _, policy := range modifiedPolicies {
		nameForCluster := policy.OrgUnit().NameForCluster()

		if policy.IsEmpty() {
			continue
		}

		// if the policy was updated, and we do not have permissions to delete updated policies in
		// the cluster, skip this update
		updated := !policy.Policy.IsDeleted()
		if updated {
			if !deleteUpdatedPoliciesFromTheCluster {
				logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("policy %s was updated, but we are not permitted to remove it from cluster %s", policy.PolicyId, clusterId)
				continue
			} else {
				logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("policy %s was updated, and we can delete it from cluster %s", policy.PolicyId, clusterId)
			}
		} else {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("policy %s was deleted, notify cluster %s", policy.PolicyId, clusterId)
		}

		// reaching here it means
		//    (1) policy was deleted
		//    (2) policy was updated, and we have approval to remove it
		// in such case, send deletion update to the agent
		result.Policies = append(result.Policies, api.PolicySyncResponsePoliciesInner{
			Meta: api.PolicyCreationFields{
				Scope:        policy.Scope,
				ClusterId:    policy.ClusterId,
				ProjectId:    policy.ProjectId,
				DepartmentId: policy.DepartmentId,
				Name:         nameForCluster,
			},
			IsDeleted:     convutils.Ref(true),
			UpdateCounter: policy.UpdateCount,
		})
	}

	return api.ImplResponse{Code: http.StatusOK, Body: result}, nil
}

// UpdateSyncStatusTraining (and same for all workload types) -> we ignore the cluster sync update
func (sa *SyncApiService) UpdateSyncStatusTraining(ctx context.Context, request api.SyncStatusUpdate) (api.ImplResponse, error) {
	return api.ImplResponse{Code: http.StatusNoContent}, nil
}

func (sa *SyncApiService) UpdateSyncStatusWorkspace(ctx context.Context, request api.SyncStatusUpdate) (api.ImplResponse, error) {
	return api.ImplResponse{Code: http.StatusNoContent}, nil
}

func (sa *SyncApiService) UpdateSyncStatusDistributed(ctx context.Context, request api.SyncStatusUpdate) (api.ImplResponse, error) {
	return api.ImplResponse{Code: http.StatusNoContent}, nil
}

func (sa *SyncApiService) UpdateSyncStatusInference(ctx context.Context, request api.SyncStatusUpdate) (api.ImplResponse, error) {
	return api.ImplResponse{Code: http.StatusNoContent}, nil
}

func (sa *SyncApiService) UpdateSyncStatusDistributedInference(ctx context.Context, request api.SyncStatusUpdate) (api.ImplResponse, error) {
	return api.ImplResponse{Code: http.StatusNoContent}, nil
}
