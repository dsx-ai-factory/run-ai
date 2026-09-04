/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"context"
	"fmt"
	"net/http"

	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/orm"
)

func DbUpdateAccessRulePhase(ctx context.Context, accessRuleId uint, tenantId uint, clusterId string, accessRulesStatusRequest api.AccessRuleStatusRequest, dbAccessRule *models.AccessRule) *api.Error {
	if dbAccessRule.ScopeType == api.SCOPETYPE_SYSTEM {
		// not recording the phase of system scoped access rules
		return nil
	}

	if accessRulesStatusRequest.Phase == api.ACCESSRULECLUSTERPHASE_DELETED {
		// at this phase, we are not reporting deletion
		return nil
		//if dbAccessRule.Phase != string(api.ACCESSRULEPHASE_DELETING) && dbAccessRule.Phase != string(api.ACCESSRULEPHASE_CREATING) {
		//	return api.NewError(http.StatusBadRequest, fmt.Sprintf("Access rule %d is not in Deleting or Creating phase; cannot delete it", accessRuleId), nil)
		//}
		//return dbUpdateAccessRulePhaseDeleted(ctx, accessRuleId, tenantId, dbAccessRule, clusterId)
	}

	if dbAccessRule.ScopeType == api.SCOPETYPE_TENANT {
		return dbUpdateAccessRulePhaseAndPhaseByCluster(ctx, accessRuleId, tenantId, dbAccessRule, clusterId, accessRulesStatusRequest.Phase)
	}

	return DbUpdateAccessRulePhaseAndMessage(ctx, accessRuleId, tenantId, clusterId, accessRulesStatusRequest)
}

func DbUpdateAccessRulePhaseAndMessage(ctx context.Context, accessRuleId uint, tenantId uint, clusterId string, accessRulesStatusRequest api.AccessRuleStatusRequest) *api.Error {
	dbErr := api.NewError(http.StatusInternalServerError, fmt.Sprintf("Failed to update access rule %d phase", accessRuleId), nil)

	err := runaidb.DB().WithContext(ctx).Transaction(func(tx orm.DB) error {
		_, err := tx.
			Model(&models.AccessRule{}).
			Where("tenant_id = ?", tenantId).
			Where("cluster_uuid = ?", clusterId).
			Where("id = ?", accessRuleId).
			Updates(map[string]interface{}{
				"phase":         string(accessRulesStatusRequest.Phase),
				"phase_message": accessRulesStatusRequest.Message,
			})
		if err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to update access rule %d phase in db: %v", accessRuleId, err)
		}
		return err
	})
	if err != nil {
		return dbErr.WithDetails(err)
	}

	return nil
}

func dbUpdateAccessRulePhaseAndPhaseByCluster(ctx context.Context, accessRuleId uint, tenantId uint, dbAccessRule *models.AccessRule, clusterId string, phaseInCluster api.AccessRuleClusterPhase) *api.Error {
	dbPhaseByCluster := dbAccessRule.PhaseByCluster
	if dbPhaseByCluster == nil {
		dbPhaseByCluster = models.PhaseByCluster{}
	}
	dbPhaseByCluster[clusterId] = string(phaseInCluster)
	calculatedPhase := calculateAccumulativePhaseAfterPhaseUpdateFromCluster(dbPhaseByCluster, dbAccessRule.Phase)

	dbErr := api.NewError(http.StatusInternalServerError, fmt.Sprintf("Failed to update access rule %d phase and phase detailed", accessRuleId), nil)

	err := runaidb.DB().WithContext(ctx).Transaction(func(tx orm.DB) error {
		_, err := tx.
			Model(&models.AccessRule{}).
			Where("tenant_id = ?", tenantId).
			Where("id = ?", accessRuleId).
			Updates(map[string]interface{}{
				"phase":            calculatedPhase,
				"phase_by_cluster": dbPhaseByCluster,
			})
		if err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to update access rule %d phase and phase detailed in db: %v", accessRuleId, err)
		}
		return err
	})
	if err != nil {
		return dbErr.WithDetails(err)
	}

	return nil
}

// calculateAccumulativePhaseAfterPhaseUpdateFromCluster calculates the accumulative phase of the access rule.
// this func should be called after dbPhaseByCluster is already set with the new phase reported from the cluster.
func calculateAccumulativePhaseAfterPhaseUpdateFromCluster(dbPhaseByCluster models.PhaseByCluster, dbPhase string) string {
	if dbPhase == string(api.ACCESSRULEPHASE_DELETING) {
		return dbPhase
	}

	hasOneNotReady := false
	for _, clusterPhase := range dbPhaseByCluster {
		if clusterPhase == string(api.ACCESSRULECLUSTERPHASE_NOT_READY) {
			hasOneNotReady = true
			break
		}
	}

	if hasOneNotReady {
		return string(api.ACCESSRULECLUSTERPHASE_NOT_READY)
	}

	// dbPhaseByCluster should already be updated with the new phase from the cluster -
	// which is either Ready or NotReady
	// (since "Deleted" has a different flow).
	// So - if no cluster is not ready, then the access rule is ready
	return string(api.ACCESSRULECLUSTERPHASE_READY)
}
