/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"

	"k8s.io/utils/ptr"

	authorization_api "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/services/simulation"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/queryapi"
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

type simulatedOrgUnit struct {
	id        string
	rank      string
	resources models.ResourcesPerNodePool
	unit      simulation.Unit
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
	if !hasAnyScope(*scopes) {
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
	// The ids end up in queryapi filter strings, where anything but a plain id could change the filter.
	if !isPlainId(nodepoolId) {
		return api.Response(http.StatusNotFound, api.NewError(http.StatusNotFound,
			fmt.Sprintf("node pool %s not found", nodepoolId), nil)), nil
	}
	if request.Parent != nil && !isPlainId(request.Parent.Id) {
		err := notFoundInNodePool(orgUnitTypeProject, request.OrgUnits[0].Id)
		return HandleError(err, http.StatusBadRequest, err.Error()), nil
	}

	departments, err := departmentsOfNodePool(ctx, *tenantId, nodepoolId, nil)
	if err != nil {
		return HandleDbError(ctx, err, "Failed to get the node pool's departments"), nil
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
		visibleDepartments, err := visibleDepartmentsOfNodePool(ctx, *tenantId, nodepoolId, departments, scopes)
		if err != nil {
			return HandleDbError(ctx, err, "Failed to get the node pool's departments"), nil
		}
		results, err := scenario.simulateDepartments(departments, visibleDepartments)
		if err != nil {
			return HandleError(err, http.StatusBadRequest, err.Error()), nil
		}
		return api.Response(http.StatusOK, api.FairShareSimulationResponse{OrgUnits: results}), nil
	}

	parentId := request.Parent.Id
	projects, err := projectsOfDepartment(ctx, *tenantId, nodepoolId, parentId, nil)
	if err != nil {
		return HandleDbError(ctx, err, "Failed to get the department's projects"), nil
	}
	visibleProjects := projects
	if !scopesCoverTenant(*tenantId, scopes) {
		visibleProjects, err = projectsOfDepartment(ctx, *tenantId, nodepoolId, parentId, scopes)
		if err != nil {
			return HandleDbError(ctx, err, "Failed to get the department's projects"), nil
		}
	}
	results, err := scenario.simulateProjects(departments, projects, visibleProjects)
	if err != nil {
		return HandleError(err, http.StatusBadRequest, err.Error()), nil
	}
	return api.Response(http.StatusOK, api.FairShareSimulationResponse{OrgUnits: results}), nil
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
		if _, ok := RankValuesMap[string(*rank)]; !ok {
			return fmt.Errorf("unknown rank %s", *rank)
		}
	}
	if gpu == nil {
		return nil
	}
	if gpu.Limit != nil && !capabilities.Limits {
		return errors.New("limit can't be set while limits is off")
	}
	for name, value := range map[string]*float64{"deserved": gpu.Deserved, "limit": gpu.Limit} {
		if value != nil && *value < 0 && *value != simulation.Unlimited {
			return fmt.Errorf("%s must be -1 (unlimited) or at least 0", name)
		}
	}
	if gpu.OverQuotaWeight != nil && *gpu.OverQuotaWeight < 0 {
		return errors.New("overQuotaWeight must not be negative")
	}
	return nil
}

type simulationScenario struct {
	capacity         float64
	capabilities     api.FairShareSimulationSchedulingCapabilities
	defaultDemand    float64
	parent           *api.FairShareSimulationParent
	changes          map[string]api.FairShareSimulationOrgUnit
	requestedInOrder []string
}

func newSimulationScenario(request api.FairShareSimulationRequest) simulationScenario {
	changes := make(map[string]api.FairShareSimulationOrgUnit, len(request.OrgUnits))
	order := make([]string, 0, len(request.OrgUnits))
	for _, orgUnit := range request.OrgUnits {
		changes[orgUnit.Id] = orgUnit
		order = append(order, orgUnit.Id)
	}
	return simulationScenario{
		capacity:         request.NodePool.Capacity.Gpu,
		capabilities:     *request.NodePool.SchedulingCapabilities,
		defaultDemand:    request.NodePool.Capacity.Gpu * defaultDemandShareOfCapacity,
		parent:           request.Parent,
		changes:          changes,
		requestedInOrder: order,
	}
}

// simulateDepartments uses each department's own demand (sent, else the default), not its projects', so it needs no
// project data.
func (s simulationScenario) simulateDepartments(departments map[string]*simulatedOrgUnit, visible map[string]bool,
) ([]api.FairShareSimulationResult, error) {
	if err := s.applyChanges(orgUnitTypeDepartment, departments, visible); err != nil {
		return nil, err
	}
	units := make([]simulation.Unit, 0, len(departments))
	for id, department := range departments {
		department.unit.Demand = s.defaultDemand
		if change, ok := s.changes[id]; ok && change.Demand != nil {
			department.unit.Demand = change.Demand.Gpu
		}
		units = append(units, department.unit)
	}
	return s.requestedResults(simulation.Divide(s.capacity, units)), nil
}

// simulateProjects divides the parent's share among its projects, with the other departments idle. The parent asks for
// the sum of its projects' demand, as the scheduler does for a parent queue. The parent needn't be readable, so a
// caller with read on a project but not on its department can still simulate it.
func (s simulationScenario) simulateProjects(departments map[string]*simulatedOrgUnit,
	projects map[string]*simulatedOrgUnit, visibleProjects map[string]*simulatedOrgUnit,
) ([]api.FairShareSimulationResult, error) {
	visible := make(map[string]bool, len(visibleProjects))
	for id := range visibleProjects {
		visible[id] = true
	}
	if err := s.applyChanges(orgUnitTypeProject, projects, visible); err != nil {
		return nil, err
	}
	if _, ok := departments[s.parent.Id]; !ok {
		return nil, notFoundInNodePool(orgUnitTypeProject, s.requestedInOrder[0])
	}

	projectUnits := make([]simulation.Unit, 0, len(projects))
	parentDemand := 0.0
	for id, project := range projects {
		project.unit.Demand = s.defaultDemand
		if change, ok := s.changes[id]; ok && change.Demand != nil {
			project.unit.Demand = change.Demand.Gpu
		}
		parentDemand += project.unit.Demand
		projectUnits = append(projectUnits, project.unit)
	}

	departmentUnits := make([]simulation.Unit, 0, len(departments))
	for id, department := range departments {
		department.unit = s.toSchedulerUnit(department)
		if id == s.parent.Id {
			department.unit.Demand = parentDemand
		}
		departmentUnits = append(departmentUnits, department.unit)
	}
	parentShare := simulation.Divide(s.capacity, departmentUnits)[s.parent.Id].FairShare
	return s.requestedResults(simulation.Divide(parentShare, projectUnits)), nil
}

// applyChanges rejects an org unit the caller can't read the same way as an unknown one, so the request can't be
// used to learn which org units exist.
func (s simulationScenario) applyChanges(orgUnitType string, units map[string]*simulatedOrgUnit, visible map[string]bool) error {
	for _, id := range s.requestedInOrder {
		unit, ok := units[id]
		if !ok || !visible[id] {
			return notFoundInNodePool(orgUnitType, id)
		}
		change := s.changes[id]
		applyChange(unit, change.Rank, change.Gpu)
	}
	for _, unit := range units {
		unit.unit = s.toSchedulerUnit(unit)
		if _, requested := s.changes[unit.id]; requested {
			if err := s.checkLimitAboveQuota(orgUnitType, unit); err != nil {
				return err
			}
		}
	}
	return nil
}

func applyChange(unit *simulatedOrgUnit, rank *api.FairShareSimulationRank, gpu *api.FairShareSimulationGpuSettings) {
	if rank != nil {
		unit.rank = string(*rank)
	}
	if gpu == nil {
		return
	}
	if gpu.Deserved != nil {
		unit.resources.GpuDeserved = gpu.Deserved
	}
	if gpu.Limit != nil {
		unit.resources.GpuLimit = gpu.Limit
	}
	if gpu.OverQuotaWeight != nil {
		unit.resources.GpuOverQuotaWeight = gpu.OverQuotaWeight
	}
}

func (s simulationScenario) checkLimitAboveQuota(orgUnitType string, orgUnit *simulatedOrgUnit) error {
	unit := s.toSchedulerUnit(orgUnit)
	if unit.Deserved >= 0 && unit.Limit != simulation.Unlimited && unit.Limit < unit.Deserved {
		return fmt.Errorf("%s %s: the limit is below the quota", orgUnitType, orgUnit.id)
	}
	return nil
}

// toSchedulerUnit maps an org unit the way the cluster sync sends it to the scheduler. A capability that is off
// resets its values on every org unit, not only on the changed ones.
func (s simulationScenario) toSchedulerUnit(orgUnit *simulatedOrgUnit) simulation.Unit {
	deserved := UnlimitedOrNil(orgUnit.resources.GpuDeserved)
	limit := ptr.Deref(orgUnit.resources.GpuLimit, simulation.Unlimited)
	if !s.capabilities.Quota {
		deserved = 0
	}
	if !s.capabilities.Limits {
		limit = simulation.Unlimited
	}
	rank := orgUnit.rank
	if !s.capabilities.Ranks || rank == "" {
		rank = DefaultResourceRank
	}
	weight := orgUnit.resources.GpuOverQuotaWeight
	if !s.capabilities.AdjustableWeights {
		weight = nonAdjustableOverQuotaWeight(ptr.To(deserved))
	}
	return simulation.Unit{
		ID:        orgUnit.id,
		Priority:  int(MapRankToNumber(&rank)),
		CreatedAt: orgUnit.unit.CreatedAt,
		Deserved:  deserved,
		Limit:     limit,
		Weight:    math.Max(0, ptr.Deref(weight, agentDefaultOverQuotaWeight)),
	}
}

func (s simulationScenario) requestedResults(results map[string]simulation.Result) []api.FairShareSimulationResult {
	response := make([]api.FairShareSimulationResult, 0, len(s.requestedInOrder))
	for _, id := range s.requestedInOrder {
		response = append(response, toFairShareSimulationResult(id, results[id]))
	}
	return response
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

// departmentsOfNodePool loads every department when scopes is nil: hidden departments compete too.
func departmentsOfNodePool(ctx context.Context, tenantId, nodepoolId string, scopes *authorization_api.PermittedScopes,
) (map[string]*simulatedOrgUnit, error) {
	filters, err := queryapi.NewParser(models.DepartmentNodePoolResources{}.FilterByFieldsToColumn()).Filters([]string{"nodePoolId==" + nodepoolId})
	if err != nil {
		return nil, fmt.Errorf("filter departments of node pool %s: %w", nodepoolId, err)
	}
	rows, err := dao.GetDepartmentNodePoolResources(ctx, tenantId, filters, "", "", 0, 0, scopes, "", models.SlimVerbosity)
	if err != nil {
		return nil, fmt.Errorf("load departments of node pool %s: %w", nodepoolId, err)
	}
	units := make(map[string]*simulatedOrgUnit, len(*rows))
	for _, row := range *rows {
		units[row.OrgUnitId] = &simulatedOrgUnit{id: row.OrgUnitId, rank: row.Rank, resources: row.ResourcesPerNodePool,
			unit: simulation.Unit{CreatedAt: row.CreatedAt}}
	}
	return units, nil
}

// projectsOfDepartment loads every project of the department when scopes is nil: hidden projects compete too.
func projectsOfDepartment(ctx context.Context, tenantId, nodepoolId, departmentId string, scopes *authorization_api.PermittedScopes,
) (map[string]*simulatedOrgUnit, error) {
	filters, err := queryapi.NewParser(models.ProjectNodePoolResources{}.FilterByFieldsToColumn()).
		Filters([]string{"nodePoolId==" + nodepoolId, "parentId==" + departmentId})
	if err != nil {
		return nil, fmt.Errorf("filter projects of department %s on node pool %s: %w", departmentId, nodepoolId, err)
	}
	rows, err := dao.GetProjectNodePoolResources(ctx, tenantId, filters, "", "", 0, 0, scopes, "", models.SlimVerbosity)
	if err != nil {
		return nil, fmt.Errorf("load projects of department %s on node pool %s: %w", departmentId, nodepoolId, err)
	}
	units := make(map[string]*simulatedOrgUnit, len(*rows))
	for _, row := range *rows {
		units[row.OrgUnitId] = &simulatedOrgUnit{id: row.OrgUnitId, rank: row.Rank, resources: row.ResourcesPerNodePool,
			unit: simulation.Unit{CreatedAt: row.CreatedAt}}
	}
	return units, nil
}

func visibleDepartmentsOfNodePool(ctx context.Context, tenantId, nodepoolId string, all map[string]*simulatedOrgUnit,
	scopes *authorization_api.PermittedScopes,
) (map[string]bool, error) {
	departments := all
	var err error
	if !scopesCoverTenant(tenantId, scopes) {
		if departments, err = departmentsOfNodePool(ctx, tenantId, nodepoolId, scopes); err != nil {
			return nil, fmt.Errorf("load visible departments: %w", err)
		}
	}
	visible := make(map[string]bool, len(departments))
	for id := range departments {
		visible[id] = true
	}
	return visible, nil
}

func isPlainId(id string) bool {
	_, err := strconv.ParseUint(id, 10, 64)
	return err == nil
}

// isNotFound reads the status the cluster-service client puts at the start of its errors, as in "404 Not Found".
func isNotFound(err error) bool {
	return strings.HasPrefix(err.Error(), strconv.Itoa(http.StatusNotFound)+" ")
}

func hasAnyScope(scopes authorization_api.PermittedScopes) bool {
	return ptr.Deref(scopes.System, false) || len(scopes.Tenants) > 0 || len(scopes.Clusters) > 0 ||
		len(scopes.Departments) > 0 || len(scopes.Projects) > 0
}

func scopesCoverTenant(tenantId string, scopes *authorization_api.PermittedScopes) bool {
	return ptr.Deref(scopes.System, false) || dao.IsTenantPermitted(tenantId, scopes)
}

func (o OrgUnitApiService) readScopes(ctx context.Context, resourceType authorization_api.ResourceType) (*authorization_api.PermittedScopes, error) {
	scopes, _, err := o.AuthorizationClient.GetPermittedScopes(ctx, resourceType, ptr.To(authorization_api.ACTION_READ))
	if err != nil {
		return nil, err
	}
	return &scopes.Read, nil
}
