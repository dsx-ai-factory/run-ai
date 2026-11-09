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
	"sort"

	"k8s.io/utils/ptr"

	authorization_api "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	cluster_api "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/services/simulation"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/queryapi"
)

const (
	orgUnitTypeDepartment = "department"
	orgUnitTypeProject    = "project"
	// The cluster agent sends a nil over-quota weight to the scheduler as 1.
	agentDefaultOverQuotaWeight = 1.0
)

type simulatedOrgUnit struct {
	parentId      string
	rank          string
	resources     models.ResourcesPerNodePool
	createdAt     int64
	demand        float64
	omitsProjects bool
}

// projectSelection names the project rows a simulation loads: all of them, or a department's projects plus the
// changed ones. Other projects only add to their department's demand, which a count per department is enough for.
type projectSelection struct {
	all          bool
	departmentId string
	ids          []string
}

func (o OrgUnitApiService) SimulateNodePoolFairShare(ctx context.Context, nodepoolId string, request api.FairShareSimulationRequest,
	orgUnitType string, parentId string,
) (api.ImplResponse, error) {
	if orgUnitType != "" && orgUnitType != orgUnitTypeDepartment && orgUnitType != orgUnitTypeProject {
		return HandleError(fmt.Errorf("unknown org unit type %s", orgUnitType), http.StatusBadRequest, "Invalid orgUnitType"), nil
	}
	tenantId, err := ParseTenantId(ctx)
	if err != nil {
		return HandleError(err, http.StatusBadRequest, "Failed to parse tenant id"), nil
	}
	nodePool, err := GetNodePoolMetadata(ctx, o.clusterExternalClient, nodepoolId)
	if err != nil {
		return HandleError(err, http.StatusNotFound, fmt.Sprintf("Failed to get node pool %s", nodepoolId)), nil
	}

	capabilities := ptr.Deref(nodePool.SchedulingCapabilities, cluster_api.SchedulingCapabilities{})
	capacity := nodePoolGpuCapacity(*nodePool)
	if request.NodePool != nil {
		if request.NodePool.SchedulingCapabilities != nil {
			c := request.NodePool.SchedulingCapabilities
			capabilities = cluster_api.SchedulingCapabilities{Quota: c.Quota, Ranks: c.Ranks, AdjustableWeights: c.AdjustableWeights, CpuQuota: c.CpuQuota}
		}
		if request.NodePool.Capacity != nil {
			capacity = request.NodePool.Capacity.Gpu
		}
	}
	if capacity < 0 {
		return HandleError(errors.New("capacity must not be negative"), http.StatusBadRequest, "Invalid capacity"), nil
	}

	selection := projectSelection{all: orgUnitType != orgUnitTypeDepartment && parentId == ""}
	if orgUnitType != orgUnitTypeDepartment {
		selection.departmentId = parentId
	}
	for _, change := range request.OrgUnits {
		if change.Type == orgUnitTypeProject {
			selection.ids = append(selection.ids, change.Id)
		}
	}

	departments, projects, err := loadSimulatedOrgUnits(ctx, *tenantId, nodepoolId, selection)
	if err != nil {
		return HandleDbError(ctx, err, "Failed to get the node pool's org units"), nil
	}
	visibleDepartments, visibleProjects, err := o.visibleOrgUnits(ctx, *tenantId, nodepoolId, selection)
	if err != nil {
		return HandleError(err, http.StatusServiceUnavailable, "Failed to get permitted scopes"), nil
	}
	if err := applySimulationChanges(ctx, request.OrgUnits, departments, projects, visibleDepartments, visibleProjects,
		*nodePool, capabilities); err != nil {
		return HandleError(err, http.StatusBadRequest, err.Error()), nil
	}

	departmentShares, projectShares := simulation.Simulate(capacity,
		toSimulationOrgUnits(departments, capabilities), toSimulationOrgUnits(projects, capabilities))

	results := []api.FairShareSimulationResult{}
	if orgUnitType != orgUnitTypeProject {
		for id := range visibleDepartments {
			results = append(results, toFairShareSimulationResult(id, orgUnitTypeDepartment, nil, departmentShares[id]))
		}
	}
	if orgUnitType != orgUnitTypeDepartment {
		for id := range visibleProjects {
			projectParentId := projects[id].parentId
			if parentId != "" && projectParentId != parentId {
				continue
			}
			results = append(results, toFairShareSimulationResult(id, orgUnitTypeProject, ptr.To(projectParentId), projectShares[id]))
		}
	}
	sort.Slice(results, func(i, j int) bool {
		if results[i].Type != results[j].Type {
			return results[i].Type == orgUnitTypeDepartment
		}
		return results[i].Id < results[j].Id
	})
	return api.Response(http.StatusOK, api.FairShareSimulationResponse{
		Capacity: api.FairShareSimulationResources{Gpu: capacity},
		OrgUnits: results,
	}), nil
}

func nodePoolGpuCapacity(nodePool cluster_api.NodepoolMinimal) float64 {
	total := 0.0
	for _, gpuType := range nodePool.GpuTypes {
		total += float64(gpuType.TotalDevices)
	}
	return total
}

// loadSimulatedOrgUnits ignores permitted scopes on purpose: every org unit of the node pool affects every other's
// fair share, and only the shares of org units the caller can read are returned.
func loadSimulatedOrgUnits(ctx context.Context, tenantId, nodepoolId string, selection projectSelection,
) (map[string]*simulatedOrgUnit, map[string]*simulatedOrgUnit, error) {
	departmentRows, err := departmentNodePoolResources(ctx, tenantId, nodepoolId, nil)
	if err != nil {
		return nil, nil, err
	}
	projectRows, err := projectNodePoolResources(ctx, tenantId, nodepoolId, selection, nil)
	if err != nil {
		return nil, nil, err
	}

	departments := make(map[string]*simulatedOrgUnit, len(*departmentRows))
	for _, row := range *departmentRows {
		departments[row.OrgUnitId] = &simulatedOrgUnit{rank: row.Rank, resources: row.ResourcesPerNodePool,
			createdAt: row.CreatedAt.UnixNano(), demand: math.Inf(1)}
	}
	projects := make(map[string]*simulatedOrgUnit, len(*projectRows))
	for _, row := range *projectRows {
		projects[row.OrgUnitId] = &simulatedOrgUnit{parentId: row.ParentId, rank: row.Rank, resources: row.ResourcesPerNodePool,
			createdAt: row.CreatedAt.UnixNano(), demand: math.Inf(1)}
	}
	if selection.all {
		return departments, projects, nil
	}

	counts, err := dao.CountProjectNodePoolResourcesPerDepartment(ctx, tenantId, nodepoolId)
	if err != nil {
		return nil, nil, err
	}
	loaded := map[string]int64{}
	for _, p := range projects {
		loaded[p.parentId]++
	}
	for _, count := range counts {
		if department, ok := departments[count.ParentId]; ok && count.Count > loaded[count.ParentId] {
			department.omitsProjects = true
		}
	}
	return departments, projects, nil
}

func departmentNodePoolResources(ctx context.Context, tenantId, nodepoolId string, scopes *authorization_api.PermittedScopes,
) (*[]models.DepartmentNodePoolResources, error) {
	filters, err := queryapi.NewParser(models.DepartmentNodePoolResources{}.FilterByFieldsToColumn()).Filters([]string{"nodePoolId==" + nodepoolId})
	if err != nil {
		return nil, err
	}
	return dao.GetDepartmentNodePoolResources(ctx, tenantId, filters, "", "", 0, 0, scopes, "")
}

func projectNodePoolResources(ctx context.Context, tenantId, nodepoolId string, selection projectSelection,
	scopes *authorization_api.PermittedScopes,
) (*[]models.ProjectNodePoolResources, error) {
	if !selection.all {
		return dao.GetProjectNodePoolResourcesOfDepartmentOrIds(ctx, tenantId, nodepoolId, selection.departmentId, selection.ids, scopes)
	}
	filters, err := queryapi.NewParser(models.ProjectNodePoolResources{}.FilterByFieldsToColumn()).Filters([]string{"nodePoolId==" + nodepoolId})
	if err != nil {
		return nil, err
	}
	return dao.GetProjectNodePoolResources(ctx, tenantId, filters, "", "", 0, 0, scopes, "")
}

func (o OrgUnitApiService) visibleOrgUnits(ctx context.Context, tenantId, nodepoolId string, selection projectSelection,
) (map[string]bool, map[string]bool, error) {
	departmentScopes, err := o.readScopes(ctx, authorization_api.RESOURCETYPE_DEPARTMENT)
	if err != nil {
		return nil, nil, err
	}
	projectScopes, err := o.readScopes(ctx, authorization_api.RESOURCETYPE_PROJECT)
	if err != nil {
		return nil, nil, err
	}

	departmentRows, err := departmentNodePoolResources(ctx, tenantId, nodepoolId, departmentScopes)
	if err != nil {
		return nil, nil, err
	}
	projectRows, err := projectNodePoolResources(ctx, tenantId, nodepoolId, selection, projectScopes)
	if err != nil {
		return nil, nil, err
	}

	departments := make(map[string]bool, len(*departmentRows))
	for _, row := range *departmentRows {
		departments[row.OrgUnitId] = true
	}
	projects := make(map[string]bool, len(*projectRows))
	for _, row := range *projectRows {
		projects[row.OrgUnitId] = true
	}
	return departments, projects, nil
}

func (o OrgUnitApiService) readScopes(ctx context.Context, resourceType authorization_api.ResourceType) (*authorization_api.PermittedScopes, error) {
	scopes, _, err := o.AuthorizationClient.GetPermittedScopes(ctx, resourceType, ptr.To(authorization_api.ACTION_READ))
	if err != nil {
		return nil, err
	}
	return &authorization_api.PermittedScopes{
		System:      scopes.Read.System,
		Tenants:     scopes.Read.Tenants,
		Clusters:    scopes.Read.Clusters,
		Departments: scopes.Read.Departments,
		Projects:    scopes.Read.Projects,
	}, nil
}

// applySimulationChanges rejects an org unit the caller can't read the same way as an unknown one, so the request
// can't be used to learn which org units exist.
func applySimulationChanges(ctx context.Context, changes []api.FairShareSimulationOrgUnit, departments, projects map[string]*simulatedOrgUnit,
	visibleDepartments, visibleProjects map[string]bool, nodePool cluster_api.NodepoolMinimal, capabilities cluster_api.SchedulingCapabilities,
) error {
	nodePool.SchedulingCapabilities = &capabilities
	nodePoolsMetadata := map[string]cluster_api.NodepoolMinimal{nodePool.Id: nodePool}
	seen := map[string]bool{}
	for _, change := range changes {
		key := change.Type + "/" + change.Id
		if seen[key] {
			return fmt.Errorf("%s %s appears more than once", change.Type, change.Id)
		}
		seen[key] = true

		var unit *simulatedOrgUnit
		switch change.Type {
		case orgUnitTypeDepartment:
			if visibleDepartments[change.Id] {
				unit = departments[change.Id]
			}
			if unit != nil && change.Demand != nil {
				return fmt.Errorf("department %s: demand is the sum of its projects' demand and can't be set", change.Id)
			}
		case orgUnitTypeProject:
			if visibleProjects[change.Id] {
				unit = projects[change.Id]
			}
		default:
			return fmt.Errorf("unknown org unit type %s", change.Type)
		}
		if unit == nil {
			return fmt.Errorf("%s %s not found in the node pool", change.Type, change.Id)
		}

		if (change.Resources == nil) != (change.Rank == "") {
			return fmt.Errorf("%s %s: rank and resources must be sent together", change.Type, change.Id)
		}
		if change.Resources != nil {
			resources := api.ResourcesNullable{
				NodePool: &api.ResourcesNodePool{Id: nodePool.Id},
				Gpu: api.ResourcesNullableGpu{
					Deserved: change.Resources.Gpu.Deserved, Limit: change.Resources.Gpu.Limit, OverQuotaWeight: change.Resources.Gpu.OverQuotaWeight,
				},
				Rank: ptr.To(change.Rank),
			}
			if err := ValidateResources(ctx, []api.ResourcesNullable{resources}, nodePoolsMetadata); err != nil {
				return fmt.Errorf("%s %s: %w", change.Type, change.Id, err)
			}
			unit.rank = change.Rank
			unit.resources.GpuDeserved = change.Resources.Gpu.Deserved
			unit.resources.GpuLimit = change.Resources.Gpu.Limit
			unit.resources.GpuOverQuotaWeight = change.Resources.Gpu.OverQuotaWeight
		}
		if change.Demand != nil {
			if change.Demand.Gpu < 0 {
				return fmt.Errorf("project %s: demand must not be negative", change.Id)
			}
			unit.demand = change.Demand.Gpu
		}
	}
	return nil
}

// toSimulationOrgUnits maps org units the way the cluster sync sends them to the scheduler. Turning quota or ranks
// off resets them, so the simulation applies those resets to every org unit, not only the changed ones.
func toSimulationOrgUnits(units map[string]*simulatedOrgUnit, capabilities cluster_api.SchedulingCapabilities) []simulation.OrgUnit {
	result := make([]simulation.OrgUnit, 0, len(units))
	for id, u := range units {
		deserved := UnlimitedOrNil(u.resources.GpuDeserved)
		limit := ptr.Deref(u.resources.GpuLimit, simulation.Unlimited)
		if !capabilities.Quota {
			deserved, limit = 0, simulation.Unlimited
		}
		rank := u.rank
		if !capabilities.Ranks || rank == "" {
			rank = DefaultResourceRank
		}
		weight := u.resources.GpuOverQuotaWeight
		if !capabilities.AdjustableWeights {
			weight = nonAdjustableOverQuotaWeight(ptr.To(deserved))
		}
		result = append(result, simulation.OrgUnit{
			ID:            id,
			ParentID:      u.parentId,
			Deserved:      deserved,
			Limit:         limit,
			Weight:        math.Max(0, ptr.Deref(weight, agentDefaultOverQuotaWeight)),
			Priority:      int(MapRankToNumber(&rank)),
			Demand:        u.demand,
			CreatedAt:     u.createdAt,
			OmitsProjects: u.omitsProjects,
		})
	}
	return result
}

func toFairShareSimulationResult(id, orgUnitType string, parentId *string, share simulation.Share) api.FairShareSimulationResult {
	return api.FairShareSimulationResult{
		Id:         id,
		Type:       orgUnitType,
		ParentId:   parentId,
		FairShare:  api.FairShareSimulationResources{Gpu: share.FairShare},
		Guaranteed: api.FairShareSimulationResources{Gpu: share.Guaranteed},
		OverQuota:  api.FairShareSimulationResources{Gpu: share.OverQuota},
		Unmet:      api.FairShareSimulationResources{Gpu: share.Unmet},
	}
}
