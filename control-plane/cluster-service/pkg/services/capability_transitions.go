/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"k8s.io/utils/ptr"

	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/database/models"
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client/generated"
	runai_config "github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
)

const maxBlockingOrgUnitsInMessage = 10

// applyCapabilityTransitions must run after the request is merged into the node pool and before it is
// saved. All checks run before any reset, so a refused change leaves the org units untouched; resets run
// before the save, so the stored org-unit values always agree with the saved flags. The returned resets are
// already committed in org-unit-service and cannot be rolled back, so a failed save must report them.
func (s *NodePoolsApiService) applyCapabilityTransitions(ctx context.Context, before, after models.Nodepool) ([]string, *api.Error) {
	if !capabilitiesChanged(before, after) {
		return nil, nil
	}
	if !runai_config.Get[*config.EditableNodepoolCapabilities]().Enabled {
		return nil, api.NewError(http.StatusBadRequest,
			fmt.Sprintf("Changing the scheduling capabilities of nodepool %s is not supported yet", after.Name), nil)
	}

	if before.AllowQuota && !after.AllowQuota {
		if apiErr := s.validateCanTurnOff(ctx, after, "quota",
			openapi.ORGUNITSNODEPOOLSUMMARYTYPE_HAS_QUOTA, "QUOTAS_TO_DEFAULT"); apiErr != nil {
			return nil, apiErr
		}
	}

	if before.AllowRanks && !after.AllowRanks {
		if apiErr := s.validateCanTurnOff(ctx, after, "ranks",
			openapi.ORGUNITSNODEPOOLSUMMARYTYPE_NON_DEFAULT_RANK, "RANKS_TO_DEFAULT"); apiErr != nil {
			return nil, apiErr
		}
	}

	return s.resetTurnedOffCapabilities(ctx, before, after)
}

func capabilitiesChanged(before, after models.Nodepool) bool {
	return before.AllowQuota != after.AllowQuota ||
		before.AllowRanks != after.AllowRanks ||
		before.AllowCpuQuota != after.AllowCpuQuota ||
		before.AllowAdjustableWeights != after.AllowAdjustableWeights
}

func resetsAppliedHint(applied []string) string {
	if len(applied) == 0 {
		return ""
	}
	return fmt.Sprintf(". Already reset: %s. Fix the error, then send the request again", strings.Join(applied, ", "))
}

func (s *NodePoolsApiService) validateCanTurnOff(ctx context.Context, nodepool models.Nodepool, capability string,
	summaryType openapi.OrgUnitsNodepoolSummaryType, resetAction string,
) *api.Error {
	blocking, err := s.orgUnitClient.NodePoolSummary(ctx, strconv.Itoa(int(nodepool.ID)), summaryType)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get %s summary of nodepool %s. error: %v", summaryType, nodepool.Name, err)
		return api.NewError(http.StatusServiceUnavailable,
			fmt.Sprintf("failed to check the org units of nodepool %s", nodepool.Name), err)
	}
	if len(blocking) == 0 {
		return nil
	}
	return api.NewError(http.StatusBadRequest, fmt.Sprintf(
		"Cannot turn off %s on nodepool %s. %d org units use it: %s. Reset them with the org-unit resources reset action %s, then try again.",
		capability, nodepool.Name, len(blocking), describeOrgUnits(blocking), resetAction), nil)
}

func describeOrgUnits(orgUnits []openapi.OrgUnitsNodepoolSummaryInner) string {
	names := make([]string, 0, min(len(orgUnits), maxBlockingOrgUnitsInMessage))
	for _, orgUnit := range orgUnits[:min(len(orgUnits), maxBlockingOrgUnitsInMessage)] {
		names = append(names, strings.TrimSpace(ptr.Deref(orgUnit.Type, "")+" "+ptr.Deref(orgUnit.Name, "")))
	}
	if len(orgUnits) > maxBlockingOrgUnitsInMessage {
		names = append(names, fmt.Sprintf("and %d more", len(orgUnits)-maxBlockingOrgUnitsInMessage))
	}
	return strings.Join(names, ", ")
}

func (s *NodePoolsApiService) resetTurnedOffCapabilities(ctx context.Context, before, after models.Nodepool) ([]string, *api.Error) {
	resets := []struct {
		turnedOff bool
		action    openapi.NodepoolResourcesResetAction
	}{
		{before.AllowCpuQuota && !after.AllowCpuQuota, openapi.NODEPOOLRESOURCESRESETACTION_CPU_QUOTAS_TO_DEFAULT},
		// reads derive the weight from deserved while weights are off, but the stored weights stay and
		// would come back if weights are turned on again
		{before.AllowAdjustableWeights && !after.AllowAdjustableWeights, openapi.NODEPOOLRESOURCESRESETACTION_WEIGHTS_TO_DEFAULT},
	}
	var applied []string
	for _, reset := range resets {
		if !reset.turnedOff {
			continue
		}
		err := s.orgUnitClient.NodePoolResetResources(ctx, after.ClusterId.String(), strconv.Itoa(int(after.ID)), reset.action)
		if err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to reset %s on nodepool %s. error: %v", reset.action, after.Name, err)
			return applied, api.NewError(http.StatusServiceUnavailable,
				fmt.Sprintf("failed to reset %s on the org units of nodepool %s%s", reset.action, after.Name, resetsAppliedHint(applied)), err)
		}
		applied = append(applied, string(reset.action))
	}
	return applied, nil
}
