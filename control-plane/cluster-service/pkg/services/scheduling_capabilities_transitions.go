/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"k8s.io/utils/ptr"

	authorization_api "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/database/models"
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
)

const maxBlockingOrgUnitsInMessage = 10

// validateSchedulingCapabilitiesChangeAndResetOrgUnits refuses turning quota or ranks off while org units still use
// them, and resets the org units' CPU quotas when cpuQuota is turned off and their limits when limits is turned off.
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

	// with quota turned off too, the quota check above already guarantees no org unit has a limit
	resetLimits := false
	if before.AllowLimits && !after.AllowLimits && after.AllowQuota {
		withLimits, apiErr := s.orgUnitsUsingSchedulingCapability(ctx, after, openapi.ORGUNITSNODEPOOLSUMMARYTYPE_HAS_LIMIT)
		if apiErr != nil {
			return nil, apiErr
		}
		if len(withLimits) > 0 {
			if apiErr := s.validateCallerCanResetOrgUnits(ctx, after, "limits", withLimits); apiErr != nil {
				return nil, apiErr
			}
			resetLimits = true
		}
	}

	var completedResets []string
	if before.AllowCpuQuota && !after.AllowCpuQuota {
		if apiErr := s.resetOrgUnits(ctx, after, openapi.NODEPOOLRESOURCESRESETACTION_CPU_QUOTAS_TO_DEFAULT); apiErr != nil {
			return nil, apiErr
		}
		completedResets = append(completedResets, string(openapi.NODEPOOLRESOURCESRESETACTION_CPU_QUOTAS_TO_DEFAULT))
	}
	if resetLimits {
		if apiErr := s.resetOrgUnits(ctx, after, openapi.NODEPOOLRESOURCESRESETACTION_LIMITS_TO_DEFAULT); apiErr != nil {
			apiErr.Message += describeCompletedResets(completedResets)
			return nil, apiErr
		}
		completedResets = append(completedResets, string(openapi.NODEPOOLRESOURCESRESETACTION_LIMITS_TO_DEFAULT))
	}

	return completedResets, nil
}

func (s *NodePoolsApiService) resetOrgUnits(ctx context.Context, nodepool models.Nodepool,
	resetAction openapi.NodepoolResourcesResetAction,
) *api.Error {
	err := s.orgUnitClient.NodePoolResetResources(ctx, nodepool.ClusterId.String(), strconv.Itoa(int(nodepool.ID)), resetAction)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to reset %s on nodepool %s. error: %v", resetAction, nodepool.Name, err)
		return api.NewError(http.StatusServiceUnavailable,
			fmt.Sprintf("failed to reset %s on the org units of nodepool %s", resetAction, nodepool.Name), err)
	}
	return nil
}

// validateCallerCanResetOrgUnits is needed because resets go through org-unit-service's internal API, which doesn't
// authorize the caller. It requires what the public reset endpoints require: update on every department and project
// in the cluster.
func (s *NodePoolsApiService) validateCallerCanResetOrgUnits(ctx context.Context, nodepool models.Nodepool,
	capability string, orgUnits []openapi.OrgUnitsNodepoolSummaryInner,
) *api.Error {
	for _, resourceType := range []authorization_api.ResourceType{authorization_api.RESOURCETYPE_DEPARTMENT, authorization_api.RESOURCETYPE_PROJECT} {
		scopes, _, err := s.authorizationClient.GetPermittedScopes(ctx, resourceType, ptr.To(authorization_api.ACTION_UPDATE))
		if err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get %s permitted scopes. error: %v", resourceType, err)
			return api.NewError(http.StatusServiceUnavailable, "failed to check the caller's permissions", err)
		}
		if scopes == nil || !permittedScopesCoverCluster(scopes.Update, nodepool.TenantId, nodepool.ClusterId.String()) {
			return api.NewError(http.StatusForbidden, fmt.Sprintf(
				"Cannot turn off %s on nodepool %s. It resets %d org units: %s. That requires permission to update every department and project in the cluster.",
				capability, nodepool.Name, len(orgUnits), describeOrgUnits(orgUnits)), nil)
		}
	}
	return nil
}

func permittedScopesCoverCluster(scopes authorization_api.PermittedScopes, tenantId int32, clusterId string) bool {
	return ptr.Deref(scopes.System, false) ||
		slices.Contains(scopes.Tenants, strconv.Itoa(int(tenantId))) ||
		slices.Contains(scopes.Clusters, clusterId)
}

func (s *NodePoolsApiService) validateCanTurnOffSchedulingCapability(ctx context.Context, nodepool models.Nodepool,
	capability string, summaryType openapi.OrgUnitsNodepoolSummaryType, resetAction string,
) *api.Error {
	blocking, apiErr := s.orgUnitsUsingSchedulingCapability(ctx, nodepool, summaryType)
	if apiErr != nil || len(blocking) == 0 {
		return apiErr
	}
	return api.NewError(http.StatusBadRequest, fmt.Sprintf(
		"Cannot turn off %s on nodepool %s. %d org units use it: %s. Reset them with the org-unit resources reset action %s, then try again.",
		capability, nodepool.Name, len(blocking), describeOrgUnits(blocking), resetAction), nil)
}

func (s *NodePoolsApiService) orgUnitsUsingSchedulingCapability(ctx context.Context, nodepool models.Nodepool,
	summaryType openapi.OrgUnitsNodepoolSummaryType,
) ([]openapi.OrgUnitsNodepoolSummaryInner, *api.Error) {
	orgUnits, err := s.orgUnitClient.NodePoolSummary(ctx, strconv.Itoa(int(nodepool.ID)), summaryType)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get %s summary of nodepool %s. error: %v", summaryType, nodepool.Name, err)
		return nil, api.NewError(http.StatusServiceUnavailable,
			fmt.Sprintf("failed to check the org units of nodepool %s", nodepool.Name), err)
	}
	return orgUnits, nil
}

func schedulingCapabilitiesChanged(before, after models.Nodepool) bool {
	return before.AllowQuota != after.AllowQuota ||
		before.AllowRanks != after.AllowRanks ||
		before.AllowCpuQuota != after.AllowCpuQuota ||
		before.AllowAdjustableWeights != after.AllowAdjustableWeights ||
		before.AllowLimits != after.AllowLimits
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
