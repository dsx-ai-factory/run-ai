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

// simulationError carries the response a failed step answers with, so the steps can return plain errors.
type simulationError struct {
	response api.ImplResponse
}

func (e *simulationError) Error() string {
	return fmt.Sprint(e.response.Body)
}

func (o OrgUnitApiService) SimulateNodePoolFairShare(ctx context.Context, nodepoolId string, request api.FairShareSimulationRequest,
) (api.ImplResponse, error) {
	response, err := o.simulateNodePoolFairShare(ctx, nodepoolId, request)
	var failed *simulationError
	if errors.As(err, &failed) {
		return failed.response, nil
	}
	if err != nil {
		return HandleError(err, http.StatusInternalServerError, err.Error()), nil
	}
	return api.Response(http.StatusOK, response), nil
}

func (o OrgUnitApiService) simulateNodePoolFairShare(ctx context.Context, nodepoolId string, request api.FairShareSimulationRequest,
) (api.FairShareSimulationResponse, error) {
	scopes, err := o.readScopesForOrgUnitType(ctx, request.OrgUnitType)
	if err != nil {
		return api.FairShareSimulationResponse{}, err
	}
	if err := validateFairShareSimulationRequest(request); err != nil {
		return api.FairShareSimulationResponse{}, badRequest(err)
	}
	tenantId, err := ParseTenantId(ctx)
	if err != nil {
		return api.FairShareSimulationResponse{}, &simulationError{HandleError(err, http.StatusBadRequest, "Failed to parse tenant id")}
	}

	departments, err := loadDepartmentsOfNodePool(ctx, *tenantId, nodepoolId, departmentVisibilityScopes(request.OrgUnitType, scopes))
	if err != nil {
		return api.FairShareSimulationResponse{}, err
	}
	scenario := newSimulationScenario(request)

	var results map[string]fair_share_simulation.Result
	if request.OrgUnitType == orgUnitTypeDepartment {
		results, err = scenario.simulateDepartments(departments)
	} else {
		var projects map[string]*dao.NodePoolSchedulingInputs
		projects, err = loadProjectsOfParent(ctx, *tenantId, nodepoolId, request, scopes)
		if err == nil {
			results, err = scenario.simulateProjects(departments, projects, request.Parent.Id)
		}
	}
	if err != nil {
		return api.FairShareSimulationResponse{}, err
	}
	return scenario.toResponse(results), nil
}

// The spec can't require department or project read depending on orgUnitType, so the handler checks it.
func (o OrgUnitApiService) readScopesForOrgUnitType(ctx context.Context, orgUnitType string) (*authorization_api.PermittedScopes, error) {
	var resourceType authorization_api.ResourceType
	switch orgUnitType {
	case orgUnitTypeDepartment:
		resourceType = authorization_api.RESOURCETYPE_DEPARTMENT
	case orgUnitTypeProject:
		resourceType = authorization_api.RESOURCETYPE_PROJECT
	default:
		return nil, badRequest(fmt.Errorf("orgUnitType must be %s or %s", orgUnitTypeDepartment, orgUnitTypeProject))
	}
	scopes, _, err := o.AuthorizationClient.GetPermittedScopes(ctx, resourceType, ptr.To(authorization_api.ACTION_READ))
	if err != nil {
		return nil, &simulationError{HandleError(fmt.Errorf("get permitted %s scopes: %w", resourceType, err),
			http.StatusServiceUnavailable, "Failed to get permitted scopes")}
	}
	if authorization_utils.IsPermittedScopesEmpty(&scopes.Read) {
		return nil, &simulationError{api.Response(http.StatusForbidden, api.NewError(http.StatusForbidden,
			fmt.Sprintf("simulating %ss requires read on %ss", orgUnitType, resourceType), nil))}
	}
	return &scopes.Read, nil
}

// Only the department level returns departments; for projects, scopes are project scopes and departments load unscoped.
func departmentVisibilityScopes(orgUnitType string, scopes *authorization_api.PermittedScopes) *authorization_api.PermittedScopes {
	if orgUnitType == orgUnitTypeDepartment {
		return scopes
	}
	return nil
}

func loadDepartmentsOfNodePool(ctx context.Context, tenantId, nodepoolId string, scopes *authorization_api.PermittedScopes,
) (map[string]*dao.NodePoolSchedulingInputs, error) {
	rows, err := dao.GetDepartmentSchedulingInputs(ctx, tenantId, nodepoolId, scopes)
	if err != nil {
		return nil, &simulationError{HandleDbError(ctx, fmt.Errorf("load departments of node pool %s: %w", nodepoolId, err),
			"Failed to get the node pool's departments")}
	}
	// The default department has resources on every node pool, so a node pool without department rows doesn't exist.
	if len(rows) == 0 {
		return nil, &simulationError{api.Response(http.StatusNotFound, api.NewError(http.StatusNotFound,
			fmt.Sprintf("node pool %s not found", nodepoolId), nil))}
	}
	return schedulingInputsById(rows), nil
}

func loadProjectsOfParent(ctx context.Context, tenantId, nodepoolId string, request api.FairShareSimulationRequest,
	scopes *authorization_api.PermittedScopes,
) (map[string]*dao.NodePoolSchedulingInputs, error) {
	parentId, err := strconv.ParseInt(request.Parent.Id, 10, 64)
	if err != nil {
		// Same error as an unknown project, so the parent can't be used to probe for departments.
		return nil, badRequest(notFoundInNodePool(orgUnitTypeProject, request.OrgUnits[0].Id))
	}
	rows, err := dao.GetProjectSchedulingInputs(ctx, tenantId, nodepoolId, parentId, scopes)
	if err != nil {
		return nil, &simulationError{HandleDbError(ctx,
			fmt.Errorf("load projects of department %s on node pool %s: %w", request.Parent.Id, nodepoolId, err),
			"Failed to get the department's projects")}
	}
	return schedulingInputsById(rows), nil
}

func badRequest(err error) error {
	return &simulationError{HandleError(err, http.StatusBadRequest, err.Error())}
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
	units, err := s.applyRequestToLevel(orgUnitTypeDepartment, departments)
	if err != nil {
		return nil, err
	}
	return fair_share_simulation.Divide(s.capacity, slices.Collect(maps.Values(units))), nil
}

// The other departments are idle, and the parent asks for the sum of its projects' demand, as the scheduler does for a
// parent queue. The parent needn't be readable: a caller may read a project without reading its department.
func (s simulationScenario) simulateProjects(departments, projects map[string]*dao.NodePoolSchedulingInputs, parentId string,
) (map[string]fair_share_simulation.Result, error) {
	projectUnits, err := s.applyRequestToLevel(orgUnitTypeProject, projects)
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

// Hidden org units are mapped too, because they compete. A requested org unit the caller can't read gets the same error
// as an unknown one, so the request can't be used to learn which org units exist.
func (s simulationScenario) applyRequestToLevel(orgUnitType string, rows map[string]*dao.NodePoolSchedulingInputs,
) (map[string]fair_share_simulation.Unit, error) {
	for _, orgUnit := range s.orgUnits {
		row, ok := rows[orgUnit.Id]
		if !ok || !row.Visible {
			return nil, badRequest(notFoundInNodePool(orgUnitType, orgUnit.Id))
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
			return nil, badRequest(fmt.Errorf("%s %s: the limit is below the quota", orgUnitType, orgUnit.Id))
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
