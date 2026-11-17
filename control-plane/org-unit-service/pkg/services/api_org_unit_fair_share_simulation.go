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
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/services/simulation"
)

const (
	orgUnitTypeDepartment = "department"
	orgUnitTypeProject    = "project"

	maxSimulatedOrgUnits = 500

	// Product decision: an org unit sent without a demand, or not sent, asks for half the node pool.
	defaultDemandShareOfCapacity = 0.5

	// The cluster agent sends a nil over-quota weight to the scheduler as 1.
	agentDefaultOverQuotaWeight = 1.0
)

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
	departments, err := loadSchedulingInputs(dao.GetDepartmentSchedulingInputs(ctx, *tenantId, nodepoolId, departmentScopes))
	if err != nil {
		return HandleDbError(ctx, fmt.Errorf("load departments of node pool %s: %w", nodepoolId, err),
			"Failed to get the node pool's departments"), nil
	}
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
		return scenario.respond(scenario.simulateDepartments(departments))
	}

	parentId := request.Parent.Id
	parentKey, err := strconv.ParseInt(parentId, 10, 64)
	if _, onNodePool := departments[parentId]; !onNodePool || err != nil {
		// An unknown parent gets the same error as an unknown project, so it can't be used to probe for departments.
		return scenario.respond(nil, notFoundInNodePool(orgUnitTypeProject, request.OrgUnits[0].Id))
	}
	projects, err := loadSchedulingInputs(dao.GetProjectSchedulingInputs(ctx, *tenantId, nodepoolId, parentKey, scopes))
	if err != nil {
		return HandleDbError(ctx, fmt.Errorf("load projects of department %s on node pool %s: %w", parentId, nodepoolId, err),
			"Failed to get the department's projects"), nil
	}
	return scenario.respond(scenario.simulateProjects(departments, projects, parentId))
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
	return value == nil || *value >= 0 || *value == simulation.Unlimited
}

type simulationScenario struct {
	capacity     float64
	capabilities api.FairShareSimulationSchedulingCapabilities
	orgUnits     []api.FairShareSimulationOrgUnit
	changes      map[string]api.FairShareSimulationOrgUnit
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

func (s simulationScenario) respond(results map[string]simulation.Result, err error) (api.ImplResponse, error) {
	if err != nil {
		return HandleError(err, http.StatusBadRequest, err.Error()), nil
	}
	response := api.FairShareSimulationResponse{OrgUnits: make([]api.FairShareSimulationResult, 0, len(s.orgUnits))}
	for _, orgUnit := range s.orgUnits {
		response.OrgUnits = append(response.OrgUnits, toFairShareSimulationResult(orgUnit.Id, results[orgUnit.Id]))
	}
	return api.Response(http.StatusOK, response), nil
}

func (s simulationScenario) demandOf(id string) float64 {
	if change, ok := s.changes[id]; ok && change.Demand != nil {
		return change.Demand.Gpu
	}
	return s.capacity * defaultDemandShareOfCapacity
}

// simulateDepartments uses each department's own demand (sent, else the default), not its projects', so it needs no
// project data.
func (s simulationScenario) simulateDepartments(departments map[string]*dao.NodePoolSchedulingInputs,
) (map[string]simulation.Result, error) {
	units, err := s.requestedLevel(orgUnitTypeDepartment, departments)
	if err != nil {
		return nil, err
	}
	return simulation.Divide(s.capacity, slices.Collect(maps.Values(units))), nil
}

// simulateProjects divides the parent's share among its projects, with the other departments idle. The parent asks for
// the sum of its projects' demand, as the scheduler does for a parent queue. The parent needn't be readable, so a
// caller with read on a project but not on its department can still simulate it.
func (s simulationScenario) simulateProjects(departments, projects map[string]*dao.NodePoolSchedulingInputs, parentId string,
) (map[string]simulation.Result, error) {
	projectUnits, err := s.requestedLevel(orgUnitTypeProject, projects)
	if err != nil {
		return nil, err
	}
	parentDemand := 0.0
	for _, project := range projectUnits {
		parentDemand += project.Demand
	}
	departmentUnits := make([]simulation.Unit, 0, len(departments))
	for id, department := range departments {
		unit := s.toSchedulerUnit(id, department)
		if id == parentId {
			unit.Demand = parentDemand
		}
		departmentUnits = append(departmentUnits, unit)
	}
	parentShare := simulation.Divide(s.capacity, departmentUnits)[parentId].FairShare
	return simulation.Divide(parentShare, slices.Collect(maps.Values(projectUnits))), nil
}

// requestedLevel applies the requested changes and maps every org unit of the level, hidden ones included. A requested
// org unit the caller can't read gets the same error as an unknown one, so the request can't be used to learn which
// org units exist.
func (s simulationScenario) requestedLevel(orgUnitType string, rows map[string]*dao.NodePoolSchedulingInputs,
) (map[string]simulation.Unit, error) {
	for _, orgUnit := range s.orgUnits {
		row, ok := rows[orgUnit.Id]
		if !ok || !row.Visible {
			return nil, notFoundInNodePool(orgUnitType, orgUnit.Id)
		}
		applyChange(row, orgUnit.Rank, orgUnit.Gpu)
	}
	units := make(map[string]simulation.Unit, len(rows))
	for id, row := range rows {
		unit := s.toSchedulerUnit(id, row)
		unit.Demand = s.demandOf(id)
		units[id] = unit
	}
	for _, orgUnit := range s.orgUnits {
		if unit := units[orgUnit.Id]; unit.Deserved >= 0 && unit.Limit != simulation.Unlimited && unit.Limit < unit.Deserved {
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

// toSchedulerUnit maps an org unit the way the cluster sync sends it to the scheduler. A capability that is off
// resets its values on every org unit, not only on the changed ones.
func (s simulationScenario) toSchedulerUnit(id string, row *dao.NodePoolSchedulingInputs) simulation.Unit {
	deserved := UnlimitedOrNil(row.GpuDeserved)
	limit := UnlimitedOrNil(row.GpuLimit)
	if !s.capabilities.Quota {
		deserved = 0
	}
	if !s.capabilities.Limits {
		limit = simulation.Unlimited
	}
	rank := &row.Rank
	if !s.capabilities.Ranks {
		rank = nil
	}
	weight := row.GpuOverQuotaWeight
	if !s.capabilities.AdjustableWeights {
		weight = nonAdjustableOverQuotaWeight(ptr.To(deserved))
	}
	return simulation.Unit{
		ID:        id,
		Priority:  int(MapRankToNumber(rank)),
		CreatedAt: row.CreatedAt,
		Deserved:  deserved,
		Limit:     limit,
		Weight:    math.Max(0, ptr.Deref(weight, agentDefaultOverQuotaWeight)),
	}
}

func toFairShareSimulationResult(id string, result simulation.Result) api.FairShareSimulationResult {
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

func loadSchedulingInputs(rows []dao.NodePoolSchedulingInputs, err error) (map[string]*dao.NodePoolSchedulingInputs, error) {
	if err != nil {
		return nil, err
	}
	byId := make(map[string]*dao.NodePoolSchedulingInputs, len(rows))
	for i := range rows {
		byId[rows[i].OrgUnitId] = &rows[i]
	}
	return byId, nil
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
