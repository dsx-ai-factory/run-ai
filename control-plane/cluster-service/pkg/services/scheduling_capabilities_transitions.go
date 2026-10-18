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
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/database/models"
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
)

const maxBlockingOrgUnitsInMessage = 10

// validateSchedulingCapabilitiesChangeAndResetOrgUnits refuses turning quota or ranks off while org units still
// use them, and resets the org units' CPU quotas when cpuQuota is turned off.
func (s *NodePoolsApiService) validateSchedulingCapabilitiesChangeAndResetOrgUnits(ctx context.Context,
	before, after models.Nodepool,
) ([]string, *api.Error) {
	// with the flag off a PUT can still change cpuQuota and weights from the tenant settings, unchecked as before
	if !isEditableNodepoolCapabilitiesOn() || !schedulingCapabilitiesChanged(before, after) {
		return nil, nil
	}

	if before.AllowQuota && !after.AllowQuota {
		if apiErr := s.validateCanTurnOffSchedulingCapability(ctx, after, "quota",
			openapi.ORGUNITSNODEPOOLSUMMARYTYPE_HAS_QUOTA, "QUOTAS_TO_DEFAULT"); apiErr != nil {
			return nil, apiErr
		}
	}

	if before.AllowRanks && !after.AllowRanks {
		if apiErr := s.validateCanTurnOffSchedulingCapability(ctx, after, "ranks",
			openapi.ORGUNITSNODEPOOLSUMMARYTYPE_NON_DEFAULT_RANK, "RANKS_TO_DEFAULT"); apiErr != nil {
			return nil, apiErr
		}
	}

	// only CPU quotas are reset
	var completedResets []string
	if before.AllowCpuQuota && !after.AllowCpuQuota {
		resetAction := openapi.NODEPOOLRESOURCESRESETACTION_CPU_QUOTAS_TO_DEFAULT
		err := s.orgUnitClient.NodePoolResetResources(ctx, after.ClusterId.String(), strconv.Itoa(int(after.ID)), resetAction)
		if err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to reset %s on nodepool %s. error: %v", resetAction, after.Name, err)
			return nil, api.NewError(http.StatusServiceUnavailable,
				fmt.Sprintf("failed to reset %s on the org units of nodepool %s", resetAction, after.Name), err)
		}
		completedResets = append(completedResets, string(resetAction))
	}

	return completedResets, nil
}

func (s *NodePoolsApiService) validateCanTurnOffSchedulingCapability(ctx context.Context, nodepool models.Nodepool,
	capability string, summaryType openapi.OrgUnitsNodepoolSummaryType, resetAction string,
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

func schedulingCapabilitiesChanged(before, after models.Nodepool) bool {
	return before.AllowQuota != after.AllowQuota ||
		before.AllowRanks != after.AllowRanks ||
		before.AllowCpuQuota != after.AllowCpuQuota ||
		before.AllowAdjustableWeights != after.AllowAdjustableWeights
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

func describeCompletedResets(completedResets []string) string {
	if len(completedResets) == 0 {
		return ""
	}
	return fmt.Sprintf(". Already reset: %s. Fix the error, then send the request again", strings.Join(completedResets, ", "))
}
