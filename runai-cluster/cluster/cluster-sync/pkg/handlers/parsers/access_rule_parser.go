/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package parsers

import (
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/util"
	"go.uber.org/zap"

	authorization_api "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
)

func ParseAccessRuleToUpdateStatusRequest(accessRuleStatus *v1alpha1.AccessRuleStatus) authorization_api.AccessRuleStatusRequest {
	return authorization_api.AccessRuleStatusRequest{
		Phase:   accessRulePhaseApi(accessRuleStatus.Phase),
		Message: accessRuleStatus.Message,
	}
}

func GetAccessRuleDeleteRequest() authorization_api.AccessRuleStatusRequest {
	return authorization_api.AccessRuleStatusRequest{
		Phase:   authorization_api.ACCESSRULECLUSTERPHASE_DELETED,
		Message: "",
	}
}

func ParseAccessRulesToSyncRequest(accessRuleList *v1alpha1.AccessRuleList) []authorization_api.AccessRuleForSync {
	result := []authorization_api.AccessRuleForSync{}
	for _, accessRule := range accessRuleList.Items {
		accessRuleId, err := GetAccessRuleId(&accessRule)
		if err != nil {
			// error was logged - continue anyway to parse other access rules
			continue
		}

		clusterAccessRuleForSync := authorization_api.AccessRuleForSync{
			Phase:        accessRulePhaseApi(accessRule.Status.Phase),
			Message:      accessRule.Status.Message,
			AccessRuleId: accessRuleId,
		}
		result = append(result, clusterAccessRuleForSync)
	}
	return result
}

func GetAccessRuleId(accessRule *v1alpha1.AccessRule) (int32, error) {
	accessRuleId, err := util.ParseLabelInt32(accessRule.Labels, v1alpha1.AccessRuleIdLabel)
	if err != nil {
		zap.S().Errorf("Failed extracting accessRuleId from accessrule <%s>, error: %s",
			accessRule.Name, err.Error())
		return -1, err
	}

	return accessRuleId, nil
}

func accessRulePhaseApi(phase v1alpha1.AccessRulePhase) authorization_api.AccessRuleClusterPhase {
	switch phase {
	case v1alpha1.AccessRuleReady:
		return authorization_api.ACCESSRULECLUSTERPHASE_READY
	case v1alpha1.AccessRuleNotReady:
		return authorization_api.ACCESSRULECLUSTERPHASE_NOT_READY
	default:
		return authorization_api.ACCESSRULECLUSTERPHASE_READY
	}
}
