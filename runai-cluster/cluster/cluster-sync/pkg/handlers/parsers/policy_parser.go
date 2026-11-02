/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package parsers

import (
	"errors"

	policyswagger "github.com/dsx-ai-factory/run-ai/runai-clients/policy-service-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/config"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/constants"
	clustercommonconsts "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	dpv2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/distributedpolicy/v2alpha1"
	infpv2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/inferencepolicy/v2alpha1"
	ipv2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/interactivepolicy/v2alpha1"
	tpv2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/trainingpolicy/v2alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func ParsePolicySyncStatusToServerScheme(policy client.Object, operation string) (policyswagger.SyncStatusUpdate, error) {
	//Distributed and Inference are not taking into account right now
	policySyncStatusPayload := policyswagger.SyncStatusUpdate{}
	scope, scopeId, err := getPolicyScopeByLabels(policy)

	if err != nil {
		return policySyncStatusPayload, err
	}

	policySyncStatusPayload.ClusterId = &config.Config().ClusterUuid
	policySyncStatusPayload.ScopeType = &scope
	if scopeId != "" {
		policySyncStatusPayload.ScopeId.Set(&scopeId)
	}

	var syncErrorMsg string
	var updateCounter int64

	switch policyObject := policy.(type) {
	case *ipv2.InteractivePolicy:
		updateCounter = policyObject.Status.SyncStatus.LastSyncRevision
		syncErrorMsg = policyObject.Status.SyncStatus.SyncErrorMsg
	case *tpv2.TrainingPolicy:
		updateCounter = policyObject.Status.SyncStatus.LastSyncRevision
		syncErrorMsg = policyObject.Status.SyncStatus.SyncErrorMsg
	case *infpv2.InferencePolicy:
		updateCounter = policyObject.Status.SyncStatus.LastSyncRevision
		syncErrorMsg = policyObject.Status.SyncStatus.SyncErrorMsg
	case *dpv2.DistributedPolicy:
		updateCounter = policyObject.Status.SyncStatus.LastSyncRevision
		syncErrorMsg = policyObject.Status.SyncStatus.SyncErrorMsg
	}

	policySyncStatusPayload.UpdateCounter = &updateCounter

	if syncErrorMsg != "" {
		policySyncStatusPayload.ErrorMessage.Set(&syncErrorMsg)
	}

	action := getActionByOperation(operation)

	policySyncStatusPayload.Action = &action

	return policySyncStatusPayload, nil
}

func getPolicyScopeByLabels(policy client.Object) (policyswagger.ScopeType, string, error) {
	policyLabels := policy.GetLabels()
	scopeCluster := policyLabels[clustercommonconsts.ClusterWideResourceLabel]
	scopeDepartment := policyLabels[clustercommonconsts.DepartmentScopeLabel]
	scopeProject := policyLabels[clustercommonconsts.ProjectScopeLabel]
	scopeTenant := policyLabels[clustercommonconsts.TenantScopeLabel]

	if scopeTenant != "" {
		return policyswagger.SCOPETYPE_TENANT, "", nil
	} else if scopeCluster != "" {
		return policyswagger.SCOPETYPE_CLUSTER, config.Config().ClusterUuid, nil
	} else if scopeDepartment != "" {
		return policyswagger.SCOPETYPE_DEPARTMENT, scopeDepartment, nil
	} else if scopeProject != "" {
		return policyswagger.SCOPETYPE_PROJECT, scopeProject, nil
	}

	return "", "", errors.New("unable to identify relevant scope")
}

func getActionByOperation(operation string) string {
	applyString := constants.ApplyAction
	deleteString := constants.DeleteOperationString

	switch operation {
	case constants.AddOperationString, constants.UpdateOperationString:
		return applyString
	case constants.DeleteOperationString:
		return deleteString
	}

	return ""
}
