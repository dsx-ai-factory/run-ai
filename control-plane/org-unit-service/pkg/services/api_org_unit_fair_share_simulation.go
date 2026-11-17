/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"math"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"k8s.io/utils/ptr"

	authorization_api "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	authorization_utils "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/helpers"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/services/fair_share_simulation"
)

const (
	orgUnitTypeDepartment = "department"
	orgUnitTypeProject    = "project"

	maxSimulatedOrgUnits         = 500
	defaultDemandShareOfCapacity = 0.5
	agentDefaultOverQuotaWeight  = 1.0
)

type simulationScenario struct {
	capacity     float64
	capabilities api.FairShareSimulationSchedulingCapabilities
	orgUnits     []api.FairShareSimulationOrgUnit
	changes      map[string]api.FairShareSimulationOrgUnit
}

func (o OrgUnitApiService) SimulateNodePoolFairShare(ctx context.Context, nodepoolId string, request api.FairShareSimulationRequest,
) (api.ImplResponse, error) {
	if request.OrgUnitType != orgUnitTypeDepartment && request.OrgUnitType != orgUnitTypeProject {
		err := fmt.Errorf("orgUnitType must be %s or %s", orgUnitTypeDepartment, orgUnitTypeProject)
		return HandleError(err, http.StatusBadRequest, err.Error()), nil
	}
	// The spec can't require department or project read depending on orgUnitType, so the handler checks it.
	resourceType := authorization_api.RESOURCETYPE_DEPARTMENT
	if request.OrgUnitType == orgUnitTypeProject {
		resourceType = authorization_api.RESOURCETYPE_PROJECT
	}
	scopes, err := o.readScopes(ctx, resourceType)
	if err != nil {
		return HandleError(fmt.Errorf("get permitted %s scopes: %w", resourceType, err), http.StatusServiceUnavailable,
			"Failed to get permitted scopes"), nil
	}
	if authorization_utils.IsPermittedScopesEmpty(scopes) {
		return api.Response(http.StatusForbidden, api.NewError(http.StatusForbidden,
			fmt.Sprintf("simulating %ss requires read on %ss", request.OrgUnitType, resourceType), nil)), nil
	}
	if err := validateFairShareSimulationRequest(request); err != nil {
		return HandleError(err, http.StatusBadRequest, err.Error()), nil
	}
	tenantId, err := ParseTenantId(ctx)
	if err != nil {
		return HandleError(err, http.StatusBadRequest, "Failed to parse tenant id"), nil
	}

	// Node pool ids are numbers in cluster-service, which fails on anything else instead of answering not found.
	if _, err := strconv.ParseUint(nodepoolId, 10, 64); err != nil {
		return api.Response(http.StatusNotFound, api.NewError(http.StatusNotFound,
			fmt.Sprintf("node pool %s not found", nodepoolId), nil)), nil
	}
	var departmentScopes *authorization_api.PermittedScopes
	if request.OrgUnitType == orgUnitTypeDepartment {
		departmentScopes = scopes
	}
	departmentRows, err := dao.GetDepartmentSchedulingInputs(ctx, *tenantId, nodepoolId, departmentScopes)
	if err != nil {
		return HandleDbError(ctx, fmt.Errorf("load departments of node pool %s: %w", nodepoolId, err),
			"Failed to get the node pool's departments"), nil
	}
	departments := schedulingInputsById(departmentRows)
	if len(departments) == 0 {
		if _, err := GetNodePoolMetadata(ctx, o.clusterExternalClient, nodepoolId); err != nil {
			if isNotFound(err) {
				return HandleError(err, http.StatusNotFound, fmt.Sprintf("node pool %s not found", nodepoolId)), nil
			}
			return HandleError(fmt.Errorf("get node pool %s: %w", nodepoolId, err), http.StatusServiceUnavailable,
				"Failed to get the node pool"), nil
		}
	}

	scenario := newSimulationScenario(request)
	if request.OrgUnitType == orgUnitTypeDepartment {
		results, err := scenario.simulateDepartments(departments)
		if err != nil {
			return HandleError(err, http.StatusBadRequest, err.Error()), nil
		}
		return api.Response(http.StatusOK, scenario.toResponse(results)), nil
	}

	parentId := request.Parent.Id
	parentKey, err := strconv.ParseInt(parentId, 10, 64)
	if _, onNodePool := departments[parentId]; !onNodePool || err != nil {
		// Same error as an unknown project, so the parent can't be used to probe for departments.
		err := notFoundInNodePool(orgUnitTypeProject, request.OrgUnits[0].Id)
		return HandleError(err, http.StatusBadRequest, err.Error()), nil
	}
	projectRows, err := dao.GetProjectSchedulingInputs(ctx, *tenantId, nodepoolId, parentKey, scopes)
	if err != nil {
		return HandleDbError(ctx, fmt.Errorf("load projects of department %s on node pool %s: %w", parentId, nodepoolId, err),
			"Failed to get the department's projects"), nil
	}
	results, err := scenario.simulateProjects(departments, schedulingInputsById(projectRows), parentId)
	if err != nil {
		return HandleError(err, http.StatusBadRequest, err.Error()), nil
	}
	return api.Response(http.StatusOK, scenario.toResponse(results)), nil
}

func validateFairShareSimulationRequest(request api.FairShareSimulationRequest) error {
	if request.NodePool == nil || request.NodePool.Capacity == nil || request.NodePool.SchedulingCapabilities == nil {
		return errors.New("nodePool.capacity and nodePool.schedulingCapabilities are required")
	}
	capabilities := *request.NodePool.SchedulingCapabilities
	switch {
	case request.OrgUnitType == orgUnitTypeDepartment && request.Parent != nil:
		return errors.New("parent can only be set for projects")
	case request.OrgUnitType == orgUnitTypeProject && (request.Parent == nil || request.Parent.Id == ""):
		return errors.New("parent is required for projects")
	case request.NodePool.Capacity.Gpu < 0:
		return errors.New("nodePool.capacity.gpu must not be negative")
	case capabilities.Limits && !capabilities.Quota:
		return errors.New("schedulingCapabilities.limits can only be on when quota is on")
	case len(request.OrgUnits) == 0 || len(request.OrgUnits) > maxSimulatedOrgUnits:
		return fmt.Errorf("orgUnits must hold 1 to %d org units", maxSimulatedOrgUnits)
	}

	seen := make(map[string]bool, len(request.OrgUnits))
	for _, orgUnit := range request.OrgUnits {
		if seen[orgUnit.Id] {
			return fmt.Errorf("%s %s appears more than once", request.OrgUnitType, orgUnit.Id)
		}
		seen[orgUnit.Id] = true
		err := validateSimulatedChange(orgUnit.Rank, orgUnit.Gpu, capabilities)
		if err == nil && orgUnit.Demand != nil && orgUnit.Demand.Gpu < 0 {
			err = errors.New("demand must not be negative")
		}
		if err != nil {
			return fmt.Errorf("%s %s: %w", request.OrgUnitType, orgUnit.Id, err)
		}
	}
	return nil
}

func validateSimulatedChange(rank *api.FairShareSimulationRank, gpu *api.FairShareSimulationGpuSettings,
	capabilities api.FairShareSimulationSchedulingCapabilities,
) error {
	if rank != nil {
		if !capabilities.Ranks {
			return errors.New("rank can't be set while ranks is off")
		}
		if _, err := ParseRank(ptr.To(string(*rank))); err != nil {
			return err
		}
	}
	if gpu == nil {
		return nil
	}
	if gpu.Limit != nil && !capabilities.Limits {
		return errors.New("limit can't be set while limits is off")
	}
	if !isQuotaValue(gpu.Deserved) {
		return errors.New("deserved must be -1 (unlimited) or at least 0")
	}
	if !isQuotaValue(gpu.Limit) {
		return errors.New("limit must be -1 (unlimited) or at least 0")
	}
	if gpu.OverQuotaWeight != nil && *gpu.OverQuotaWeight < 0 {
		return errors.New("overQuotaWeight must not be negative")
	}
	return nil
}

func isQuotaValue(value *float64) bool {
	return value == nil || *value >= 0 || *value == fair_share_simulation.Unlimited
}

func newSimulationScenario(request api.FairShareSimulationRequest) simulationScenario {
	changes := make(map[string]api.FairShareSimulationOrgUnit, len(request.OrgUnits))
	for _, orgUnit := range request.OrgUnits {
		changes[orgUnit.Id] = orgUnit
	}
	return simulationScenario{
		capacity:     request.NodePool.Capacity.Gpu,
		capabilities: *request.NodePool.SchedulingCapabilities,
		orgUnits:     request.OrgUnits,
		changes:      changes,
	}
}

func (s simulationScenario) toResponse(results map[string]fair_share_simulation.Result) api.FairShareSimulationResponse {
	response := api.FairShareSimulationResponse{OrgUnits: make([]api.FairShareSimulationResult, 0, len(s.orgUnits))}
	for _, orgUnit := range s.orgUnits {
		response.OrgUnits = append(response.OrgUnits, toFairShareSimulationResult(orgUnit.Id, results[orgUnit.Id]))
	}
	return response
}

func (s simulationScenario) demandOf(id string) float64 {
	if change, ok := s.changes[id]; ok && change.Demand != nil {
		return change.Demand.Gpu
	}
	return s.capacity * defaultDemandShareOfCapacity
}

// A department asks for its own demand here, not the sum of its projects', so no project data is needed.
func (s simulationScenario) simulateDepartments(departments map[string]*dao.NodePoolSchedulingInputs,
) (map[string]fair_share_simulation.Result, error) {
	units, err := s.requestedLevel(orgUnitTypeDepartment, departments)
	if err != nil {
		return nil, err
	}
	return fair_share_simulation.Divide(s.capacity, slices.Collect(maps.Values(units))), nil
}

// The other departments are idle, and the parent asks for the sum of its projects' demand, as the scheduler does for a
// parent queue. The parent needn't be readable: a caller may read a project without reading its department.
func (s simulationScenario) simulateProjects(departments, projects map[string]*dao.NodePoolSchedulingInputs, parentId string,
) (map[string]fair_share_simulation.Result, error) {
	projectUnits, err := s.requestedLevel(orgUnitTypeProject, projects)
	if err != nil {
		return nil, err
	}
	parentDemand := 0.0
	for _, project := range projectUnits {
		parentDemand += project.Demand
	}
	departmentUnits := make([]fair_share_simulation.Unit, 0, len(departments))
	for id, department := range departments {
		unit := s.toSchedulerUnit(id, department)
		if id == parentId {
			unit.Demand = parentDemand
		}
		departmentUnits = append(departmentUnits, unit)
	}
	parentShare := fair_share_simulation.Divide(s.capacity, departmentUnits)[parentId].FairShare
	return fair_share_simulation.Divide(parentShare, slices.Collect(maps.Values(projectUnits))), nil
}

// A requested org unit the caller can't read gets the same error as an unknown one, so the request can't be used to
// learn which org units exist. Hidden org units still compete.
func (s simulationScenario) requestedLevel(orgUnitType string, rows map[string]*dao.NodePoolSchedulingInputs,
) (map[string]fair_share_simulation.Unit, error) {
	for _, orgUnit := range s.orgUnits {
		row, ok := rows[orgUnit.Id]
		if !ok || !row.Visible {
			return nil, notFoundInNodePool(orgUnitType, orgUnit.Id)
		}
		applyChange(row, orgUnit.Rank, orgUnit.Gpu)
	}
	units := make(map[string]fair_share_simulation.Unit, len(rows))
	for id, row := range rows {
		unit := s.toSchedulerUnit(id, row)
		unit.Demand = s.demandOf(id)
		units[id] = unit
	}
	for _, orgUnit := range s.orgUnits {
		if unit := units[orgUnit.Id]; unit.Deserved >= 0 && unit.Limit != fair_share_simulation.Unlimited && unit.Limit < unit.Deserved {
			return nil, fmt.Errorf("%s %s: the limit is below the quota", orgUnitType, orgUnit.Id)
		}
	}
	return units, nil
}

func applyChange(row *dao.NodePoolSchedulingInputs, rank *api.FairShareSimulationRank, gpu *api.FairShareSimulationGpuSettings) {
	if rank != nil {
		row.Rank = string(*rank)
	}
	if gpu == nil {
		return
	}
	if gpu.Deserved != nil {
		row.GpuDeserved = gpu.Deserved
	}
	if gpu.Limit != nil {
		row.GpuLimit = gpu.Limit
	}
	if gpu.OverQuotaWeight != nil {
		row.GpuOverQuotaWeight = gpu.OverQuotaWeight
	}
}

// Mirrors what the cluster sync sends to the scheduler: a capability that is off resets its values on every org unit,
// not only on the changed ones.
func (s simulationScenario) toSchedulerUnit(id string, row *dao.NodePoolSchedulingInputs) fair_share_simulation.Unit {
	deserved := UnlimitedOrNil(row.GpuDeserved)
	limit := UnlimitedOrNil(row.GpuLimit)
	if !s.capabilities.Quota {
		deserved = 0
	}
	if !s.capabilities.Limits {
		limit = fair_share_simulation.Unlimited
	}
	rank := &row.Rank
	if !s.capabilities.Ranks {
		rank = nil
	}
	weight := row.GpuOverQuotaWeight
	if !s.capabilities.AdjustableWeights {
		// From the stored quota, as the sync does: no quota gives no weight, and the agent then sends the default.
		storedDeserved := row.GpuDeserved
		if !s.capabilities.Quota {
			storedDeserved = ptr.To(0.0)
		}
		weight = nonAdjustableOverQuotaWeight(storedDeserved)
	}
	return fair_share_simulation.Unit{
		ID:        id,
		Priority:  int(MapRankToNumber(rank)),
		CreatedAt: row.CreatedAt,
		Deserved:  deserved,
		Limit:     limit,
		Weight:    math.Max(0, ptr.Deref(weight, agentDefaultOverQuotaWeight)),
	}
}

func toFairShareSimulationResult(id string, result fair_share_simulation.Result) api.FairShareSimulationResult {
	return api.FairShareSimulationResult{
		Id:                id,
		Demand:            api.FairShareSimulationResources{Gpu: result.Demand},
		FairShare:         api.FairShareSimulationResources{Gpu: result.FairShare},
		Guaranteed:        api.FairShareSimulationResources{Gpu: result.Guaranteed},
		OverQuota:         api.FairShareSimulationResources{Gpu: result.OverQuota},
		Unmet:             api.FairShareSimulationResources{Gpu: result.Unmet},
		WeightShareInRank: result.WeightShareInRank,
	}
}

func notFoundInNodePool(orgUnitType, id string) error {
	return fmt.Errorf("%s %s not found in the node pool", orgUnitType, id)
}

func schedulingInputsById(rows []dao.NodePoolSchedulingInputs) map[string]*dao.NodePoolSchedulingInputs {
	byId := make(map[string]*dao.NodePoolSchedulingInputs, len(rows))
	for i := range rows {
		byId[rows[i].OrgUnitId] = &rows[i]
	}
	return byId
}

// isNotFound reads the status the cluster-service client puts at the start of its errors, as in "404 Not Found".
func isNotFound(err error) bool {
	return strings.HasPrefix(err.Error(), strconv.Itoa(http.StatusNotFound)+" ")
}

func (o OrgUnitApiService) readScopes(ctx context.Context, resourceType authorization_api.ResourceType) (*authorization_api.PermittedScopes, error) {
	scopes, _, err := o.AuthorizationClient.GetPermittedScopes(ctx, resourceType, ptr.To(authorization_api.ACTION_READ))
	if err != nil {
		return nil, err
	}
	return &scopes.Read, nil
}
