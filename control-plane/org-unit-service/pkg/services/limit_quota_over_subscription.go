/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/checker"
	"k8s.io/utils/ptr"
)

type ResourceType string

const (
	GpuResource    ResourceType = "gpu"
	CpuResource    ResourceType = "cpu"
	MemoryResource ResourceType = "memory"

	unlimited = float64(-1)
)

var (
	resourceTypeToName = map[ResourceType]string{
		GpuResource:    "GPUs",
		CpuResource:    "CPUs",
		MemoryResource: "memory",
	}

	resourceTypeToUnit = map[ResourceType]string{
		GpuResource:    "",
		CpuResource:    " MilliCores",
		MemoryResource: " Mb",
	}
)

func ShouldValidateOQSubscription(ctx context.Context) (bool, error) {
	return IsLimitQuotaOverSubscriptionEnabled(ctx)
}

func ShouldValidateOQSubscriptionWithDepartment(ctx context.Context) (bool, error) {
	return IsLimitQuotaOverSubscriptionEnabled(ctx)
}

func ValidateProjectUpdateNotCausingOverSubscriptionV2(ctx context.Context,
	tenantId string, requestedProjectNodePoolResources []models.Resources, dbProject models.Project,
) *api.Error {
	return ValidateProjectUpdateNotCausingOverSubscription(ctx, tenantId, requestedProjectNodePoolResources, dbProject)
}

func ValidateProjectUpdateNotCausingOverSubscriptionWrapper(ctx context.Context,
	tenantId string, requestedProjectNodePoolResources []models.Resources, dbExistingResources []models.Resources, parentId *int64, clusterId uuid.UUID,
) *api.Error {
	dbProject := models.Project{
		OrgUnitData: models.OrgUnitData{
			Resources: dbExistingResources,
		},
		ParentID:  parentId,
		ClusterId: clusterId,
	}
	return ValidateProjectUpdateNotCausingOverSubscription(ctx, tenantId, requestedProjectNodePoolResources, dbProject)
}

func ValidateProjectUpdateNotCausingOverSubscription(ctx context.Context,
	tenantId string, requestedProjectNodePoolResources []models.Resources, dbProject models.Project,
) *api.Error {
	shouldValidate, err := ShouldValidateOQSubscription(ctx)
	if err != nil {
		return api.NewError(http.StatusInternalServerError, "Failed to validate project update", err)
	}
	if !shouldValidate {
		return nil
	}

	apiErr := validateProjectUpdateNotCausingOverSubscriptionOfAllocatedNonPreemtible(ctx, requestedProjectNodePoolResources, dbProject)
	if apiErr != nil {
		return apiErr
	}

	return validateProjectUpdateNotCausingOverSubscriptionOfDepartment(ctx, tenantId, requestedProjectNodePoolResources, dbProject)
}

func ValidateProjectUpdateNotCausingOverSubscriptionOfAllocatedNonPreemtible(
	ctx context.Context,
	requestedProjectNodePoolResources []models.Resources, dbProject models.Project,
) *api.Error {
	shouldValidate, err := ShouldValidateOQSubscription(ctx)
	if err != nil {
		return api.NewError(http.StatusInternalServerError, "Failed to validate project update", err)
	}
	if !shouldValidate {
		return nil
	}

	return validateProjectUpdateNotCausingOverSubscriptionOfAllocatedNonPreemtible(ctx, requestedProjectNodePoolResources, dbProject)
}

func validateProjectUpdateNotCausingOverSubscriptionOfAllocatedNonPreemtible(ctx context.Context,
	requestedProjectNodePoolResources []models.Resources, dbProject models.Project,
) *api.Error {
	isSupported, apiError := isProjectQuotaStatusesSupportedInCluster(ctx, dbProject.ClusterId)
	if apiError != nil {
		return apiError
	}
	if !isSupported {
		return nil
	}

	return validateQuotaDecreaseNotOverAllocatedNonPreemptible(requestedProjectNodePoolResources, dbProject.OrgUnitData.Resources)
}

func validateProjectUpdateNotCausingOverSubscriptionOfDepartment(ctx context.Context, tenantId string,
	requestedProjectNodePoolResources []models.Resources, dbProject models.Project,
) *api.Error {
	if dbProject.ParentID == nil {
		return api.NewError(http.StatusInternalServerError, "Failed to validate project update", fmt.Errorf("project's parent ID is nil"))
	}

	dbDepartment, err := dao.GetDepartmentById(ctx, *dbProject.ParentID, tenantId, models.VerboseVerbosity, nil)
	if err != nil {
		return api.NewError(http.StatusInternalServerError, "Failed to validate project update", err)
	}

	return validateProjectQuotaNotOverDepartmentQuota(requestedProjectNodePoolResources, *dbDepartment, &dbProject.OrgUnitData.Resources)
}

func ValidateProjectCreateNotCausingOverSubscriptionOfDepartmentV2(ctx context.Context,
	tenantId string, requestedProjectNodePoolResources []models.Resources, parentId string,
) *api.Error {
	parentUid, err := strconv.ParseInt(parentId, 10, 64)
	if err != nil {
		return api.NewError(http.StatusInternalServerError, "Failed to validate project create", err)
	}

	return ValidateProjectCreateNotCausingOverSubscriptionOfDepartment(ctx, tenantId, requestedProjectNodePoolResources, parentUid)
}

func ValidateProjectCreateNotCausingOverSubscriptionOfDepartment(ctx context.Context,
	tenantId string, requestedProjectNodePoolResources []models.Resources, departmentId int64,
) *api.Error {
	shouldValidate, err := ShouldValidateOQSubscriptionWithDepartment(ctx)
	if err != nil {
		return api.NewError(http.StatusInternalServerError, "Failed to validate project create", err)
	}
	if !shouldValidate {
		return nil
	}

	dbDepartment, err := dao.GetDepartmentById(ctx, departmentId, tenantId, models.VerboseVerbosity, nil)
	if err != nil {
		return api.NewError(http.StatusInternalServerError, "Failed to validate project create", err)
	}

	return validateProjectQuotaNotOverDepartmentQuota(requestedProjectNodePoolResources, *dbDepartment, nil)
}

func ValidateDepartmentUpdateNotCausingOverSubscriptionV2(ctx context.Context,
	tenantId string, requestedDepartmentNodePoolResources []models.Resources, departmentId int64,
) *api.Error {
	dbDepartment, err := dao.GetDepartmentById(ctx, departmentId, tenantId, models.VerboseVerbosity, nil)
	if err != nil {
		return api.NewError(http.StatusInternalServerError, "Failed to validate department update", err)
	}
	return ValidateDepartmentUpdateNotCausingOverSubscription(ctx, requestedDepartmentNodePoolResources, *dbDepartment)
}

func ValidateDepartmentUpdateNotCausingOverSubscription(ctx context.Context,
	requestedDepartmentNodePoolResources []models.Resources, dbDepartment models.Department,
) *api.Error {
	shouldValidate, err := ShouldValidateOQSubscription(ctx)
	if err != nil {
		return api.NewError(http.StatusInternalServerError, "Failed to validate department update", err)
	}
	if !shouldValidate {
		return nil
	}

	return validateDepartmentQuotaDecreaseNotUnderProjectsSumQuota(requestedDepartmentNodePoolResources, dbDepartment)
}

// BatchProjectUpdateRequest represents a single project's original and requested resources
// used to validate a batch update against a single department.
type BatchProjectUpdateRequest struct {
	// DbProject is used for validation of allocated non-preemptible resources
	DbProject *models.Project
	// Requested desired resources for the project (per nodepool)
	Requested []models.Resources
	// Original (current) desired resources for the project (per nodepool)
	Original []models.Resources
}

// ValidateBatchProjectUpdateNotCausingOverSubscriptionOfDepartment validates that applying the requested
// resources of multiple projects that belong to the same department will not cause
// the department to exceed its assigned quotas, when considered together.
//
// This function is intended for batch validations to avoid order-dependent results.
func ValidateBatchProjectUpdateNotCausingOverSubscriptionOfDepartment(
	ctx context.Context,
	tenantId string,
	departmentId int64,
	batchUpdateRequests map[int64]BatchProjectUpdateRequest,
) *api.Error {
	shouldValidate, err := ShouldValidateOQSubscriptionWithDepartment(ctx)
	if err != nil {
		return api.NewError(http.StatusInternalServerError, "Failed to validate batch projects update", err)
	}
	if !shouldValidate {
		return nil
	}

	// Load the department with its current state (assigned quotas and projects)
	dbDepartment, err := dao.GetDepartmentById(ctx, departmentId, tenantId, models.VerboseVerbosity, nil)
	if err != nil {
		return api.NewError(http.StatusInternalServerError, "Failed to validate batch projects update", err)
	}

	// Build a map of department assigned quotas per node pool
	departmentNodePoolResourcesMap := getNodePoolResourcesMap(dbDepartment.OrgUnitData.Resources)

	// Build the projects "after" state by replacing each project's resources with the requested ones (if provided)
	projectsAfter := make([]models.Project, 0, len(dbDepartment.ProjectChildren))
	for _, projectChild := range dbDepartment.ProjectChildren {
		if req, ok := batchUpdateRequests[projectChild.ID]; ok {
			projectChild.OrgUnitData.Resources = req.Requested
		}
		projectsAfter = append(projectsAfter, projectChild)
	}

	// Aggregate all projects' desired resources after applying the batch
	aggregatedAfter := getProjectAggregatedDeservedSum(projectsAfter)
	aggregatedAfterMap := getNodePoolResourcesMap(aggregatedAfter)

	// Validate once per node pool: aggregated-after must not exceed department assigned
	for nodePoolName, aggregatedNPResources := range aggregatedAfterMap {
		departmentAssigned, ok := departmentNodePoolResourcesMap[nodePoolName]
		if !ok {
			continue
		}
		for resourceType, resourceName := range resourceTypeToName {
			departmentAssignedPtr := getResourcesDeservedPtr(departmentAssigned, resourceType)
			aggregatedPtr := getResourcesDeservedPtr(aggregatedNPResources, resourceType)
			if isResourceDeservedLessThan(departmentAssignedPtr, aggregatedPtr) {
				maxAllowed := getResourceDeserved(departmentAssignedPtr)
				unit := resourceTypeToUnit[resourceType]
				return api.NewError(http.StatusBadRequest,
					fmt.Sprintf("Cannot modify the deserved %s for node pool '%s' because the requested combined amount exceeds "+
						"the department's available quota. The maximum allowed aggregated value is '%.2f%s'; "+
						"while the requested aggregated value was '%.2f%s'.",
						resourceName, nodePoolName, maxAllowed, unit, getResourceDeserved(aggregatedPtr), unit), nil)
			}
		}
	}

	return nil
}

func validateQuotaDecreaseNotOverAllocatedNonPreemptible(requestedProjectNodePoolResources []models.Resources, dbProjectResources []models.Resources) *api.Error {
	quotaStatuses := getQuotaStatusesMap(dbProjectResources)

	for _, requestedNodePoolResources := range requestedProjectNodePoolResources {
		nodePoolName := getNodePoolName(requestedNodePoolResources)
		quotaStatus, ok := quotaStatuses[nodePoolName]
		if !ok {
			continue
		}

		err := checkIfAttemptingUnderSubscription(quotaStatus, requestedNodePoolResources)
		if err != nil {
			return err
		}
	}

	return nil
}

// check if attempted under subscription - when the current allocated non-preemptible is over the requested deserved
func checkIfAttemptingUnderSubscription(resources models.Resources, requestedNodePoolResources models.Resources) *api.Error {
	for resourceType, resourceName := range resourceTypeToName {
		minAllowedAmount := getResourceAllocatedNonPreemptible(resources.ResourcesStatus, resourceType)
		if !isResourceDeservedLessThanQuotaStatus(getResourcesDeservedPtr(requestedNodePoolResources, resourceType), minAllowedAmount) {
			continue
		}

		nodePoolName := getNodePoolName(requestedNodePoolResources)
		unit := resourceTypeToUnit[resourceType]
		return api.NewError(http.StatusBadRequest,
			fmt.Sprintf("Cannot modify the deserved %s for node pool '%s' because the requested amount is below the usage of non-preemptive workloads. The minimum allowed value is '%.2f%s'.",
				resourceName, nodePoolName, minAllowedAmount, unit), nil)
	}

	return nil
}

func validateNewFirstProjectQuota(departmentNodePoolResourcesMap map[string]models.Resources, requestedProjectNodePoolResourcesMap map[string]models.Resources) *api.Error {
	for nodePoolName, departmentResources := range departmentNodePoolResourcesMap {
		requestedNodePoolResources, ok := requestedProjectNodePoolResourcesMap[nodePoolName]
		if !ok {
			continue
		}
		emptyResources := models.Resources{
			GpuDeserved:    ptr.To(float64(0)),
			CpuDeserved:    ptr.To(float64(0)),
			MemoryDeserved: ptr.To(float64(0)),
		}
		err := checkIfRequestedDeservedOverDepartmentDeserved(emptyResources, departmentResources, emptyResources, requestedNodePoolResources, true)
		if err != nil {
			return err
		}
	}
	return nil
}

func validateProjectQuotaNotOverDepartmentQuota(requestedProjectNodePoolResources []models.Resources, dbDepartment models.Department, originalProjectNodePoolResources *[]models.Resources) *api.Error {
	departmentNodePoolResourcesMap := getNodePoolResourcesMap(dbDepartment.OrgUnitData.Resources)
	requestedProjectNodePoolResourcesMap := getNodePoolResourcesMap(requestedProjectNodePoolResources)

	if len(dbDepartment.ProjectChildren) == 0 {
		// this is the first project - test its quota alone
		return validateNewFirstProjectQuota(departmentNodePoolResourcesMap, requestedProjectNodePoolResourcesMap)
	}

	originalProjectNodePoolResourcesMap := map[string]models.Resources{}
	if originalProjectNodePoolResources != nil {
		originalProjectNodePoolResourcesMap = getNodePoolResourcesMap(*originalProjectNodePoolResources)
	}
	projectAggregatedDeservedSum := getProjectAggregatedDeservedSum(dbDepartment.ProjectChildren)

	for _, aggregatedNPResources := range projectAggregatedDeservedSum {
		nodePoolName := getNodePoolName(aggregatedNPResources)
		requestedNodePoolResources, ok := requestedProjectNodePoolResourcesMap[nodePoolName]
		if !ok {
			continue
		}

		departmentNodePoolResources, ok := departmentNodePoolResourcesMap[nodePoolName]
		if !ok {
			continue
		}

		originalNodePoolResourcesForProject, ok := originalProjectNodePoolResourcesMap[nodePoolName]
		if !ok {
			originalNodePoolResourcesForProject = models.Resources{
				GpuDeserved:    ptr.To(float64(0)),
				CpuDeserved:    ptr.To(float64(0)),
				MemoryDeserved: ptr.To(float64(0)),
			}
		}

		err := checkIfRequestedDeservedOverDepartmentDeserved(aggregatedNPResources, departmentNodePoolResources,
			originalNodePoolResourcesForProject, requestedNodePoolResources, originalProjectNodePoolResources == nil)
		if err != nil {
			return err
		}
	}

	return nil
}

// check if the requested deserved for the project won't exceed the department's available quota
func checkIfRequestedDeservedOverDepartmentDeserved(departmentAggregatedProjectsNodePoolResources,
	departmentNodePoolAssignedResources, originalNodePoolResourcesForProject,
	requestedNodePoolResourcesForProject models.Resources, isNewProject bool,
) *api.Error {
	for resourceType, resourceName := range resourceTypeToName {
		requestedDeserved := getResourcesDeservedPtr(requestedNodePoolResourcesForProject, resourceType)
		if requestedDeserved == nil {
			requestedDeserved = ptr.To(unlimited)
		}

		originalNodePooolResourceForProject := getResourcesDeserved(originalNodePoolResourcesForProject, resourceType)
		departmentNodePoolAssignedResource := getResourcesDeservedPtr(departmentNodePoolAssignedResources, resourceType)
		departmentAggregatedProjectsNodePoolResource := getResourcesDeserved(departmentAggregatedProjectsNodePoolResources, resourceType)

		maxAllowed, isAllowed := getMaxAllowedForProjectInDepartment(*requestedDeserved,
			originalNodePooolResourceForProject, departmentNodePoolAssignedResource, departmentAggregatedProjectsNodePoolResource)
		if isAllowed {
			continue
		}

		// isAllowed is false - construct error message

		nodePoolName := getNodePoolName(requestedNodePoolResourcesForProject)
		actionErrorMsg := "modify"

		if *requestedDeserved == 0 {
			// the specific case of: not allowed + requested deserved is 0 - means the department needs to be updated;
			// this probably means the feature flag was turned on while the state was faulty to begin with.
			// see error message.
			if isNewProject {
				actionErrorMsg = "create"
			}
			return api.NewError(http.StatusBadRequest,
				fmt.Sprintf("Cannot %s project until department's projects quotas of %s for node pool '%s' don't exceed the department's available quota.",
					actionErrorMsg, resourceName, nodePoolName), nil)
		}

		unit := resourceTypeToUnit[resourceType]
		if isNewProject {
			actionErrorMsg = "set"
		}
		return api.NewError(http.StatusBadRequest,
			fmt.Sprintf("Cannot %s the deserved %s for node pool '%s' because the requested amount exceeds the department's available quota. The maximum allowed value is '%.2f%s'.",
				actionErrorMsg, resourceName, nodePoolName, maxAllowed, unit), nil)
	}

	return nil
}

func validateDepartmentQuotaDecreaseNotUnderProjectsSumQuota(requestedDepartmentNodePoolResources []models.Resources, dbDepartment models.Department) *api.Error {
	requestedDepartmentNodePoolResourcesMap := getNodePoolResourcesMap(requestedDepartmentNodePoolResources)
	originalDepartmentNodePoolResourcesMap := getNodePoolResourcesMap(dbDepartment.OrgUnitData.Resources)
	projectAggregatedDeservedSum := getProjectAggregatedDeservedSum(dbDepartment.ProjectChildren)

	for _, aggregatedNPResources := range projectAggregatedDeservedSum {
		nodePoolName := getNodePoolName(aggregatedNPResources)
		requestedNodePooolResources, ok := requestedDepartmentNodePoolResourcesMap[nodePoolName]
		if !ok {
			continue
		}

		originalNodePooolResources, ok := originalDepartmentNodePoolResourcesMap[nodePoolName]
		if !ok {
			continue
		}

		err := checkIfRequestedDeservedUnderProjectsDeservedSum(aggregatedNPResources, requestedNodePooolResources, originalNodePooolResources)
		if err != nil {
			return err
		}
	}

	return nil
}

// check if the requested deserved for department, is under the aggregated sum of all projects in the department
func checkIfRequestedDeservedUnderProjectsDeservedSum(projectAggregatedNodePoolAssignedResources,
	requestedDepartmentNodePoolAssignedResources, originalNodePooolResources models.Resources,
) *api.Error {
	for resourceType, resourceName := range resourceTypeToName {
		requestedDepartmentNodePoolAssignedResource := getResourcesDeservedPtr(requestedDepartmentNodePoolAssignedResources, resourceType)
		projectAggregatedNodePoolAssignedResource := getResourcesDeservedPtr(projectAggregatedNodePoolAssignedResources, resourceType)

		if !isResourceDeservedLessThan(requestedDepartmentNodePoolAssignedResource, projectAggregatedNodePoolAssignedResource) {
			continue
		}

		// requested deserved is not allowed - construct error message

		originalNodePooolResource := getResourcesDeservedPtr(originalNodePooolResources, resourceType)

		minAllowedAmount := getResourceDeserved(projectAggregatedNodePoolAssignedResource)
		requestedDeserved := getResourceDeserved(requestedDepartmentNodePoolAssignedResource)
		originalDeserved := getResourceDeserved(originalNodePooolResource)

		nodePoolName := getNodePoolName(originalNodePooolResources)
		unit := resourceTypeToUnit[resourceType]

		minAllowedStr := "-1 (Unlimited)"
		if minAllowedAmount >= 0 {
			minAllowedStr = fmt.Sprintf("'%.2f%s'", minAllowedAmount, unit)
		}

		if requestedDeserved == originalDeserved {
			// the specific case of: not allowed + requested deserved equals original deserved - means the department needs to be updated;
			// this probably means the feature flag was turned on while the state was faulty to begin with.
			// see error message.
			return api.NewError(http.StatusBadRequest,
				fmt.Sprintf("Cannot modify department until department's deserved %s for node pool '%s' won't be below the sum of department's projects quotas, which is %s.",
					resourceName, nodePoolName, minAllowedStr), nil)
		}

		return api.NewError(http.StatusBadRequest,
			fmt.Sprintf("Cannot modify the deserved %s for node pool '%s' because the requested amount is below the sum of department's projects quotas. The minimum allowed value is %s.",
				resourceName, nodePoolName, minAllowedStr), nil)
	}

	return nil
}

// returns maxAllowed, isAllowed
func getMaxAllowedForProjectInDepartment(requestedDeserved float64, originalNodePoolResourceDeserved float64,
	departmentNodePoolAssignedResource *float64, departmentAggregatedProjectsNodePoolResources float64,
) (float64, bool) {
	if departmentNodePoolAssignedResource == nil {
		departmentNodePoolAssignedResource = ptr.To(unlimited)
	}
	if isUnlimited(departmentNodePoolAssignedResource) {
		return unlimited, true
	}

	maxAllowed := *departmentNodePoolAssignedResource - departmentAggregatedProjectsNodePoolResources + originalNodePoolResourceDeserved
	maxAllowed = roundNumber(&maxAllowed)
	requestedDeserved = roundNumber(&requestedDeserved)

	if isUnlimited(&requestedDeserved) || requestedDeserved > maxAllowed {
		return float64(math.Max(0, float64(maxAllowed))), false
	}

	return unlimited, true
}

func isResourceDeservedLessThan(requestedDepartmentResource *float64, projectAggregatedResource *float64) bool {
	if isUnlimited(requestedDepartmentResource) {
		return false
	}
	if isUnlimited(projectAggregatedResource) {
		return true
	}

	return roundNumber(requestedDepartmentResource) < roundNumber(projectAggregatedResource)
}

func isResourceDeservedLessThanQuotaStatus(resourceDeserved *float64, minAllowedAmount float64) bool {
	if isUnlimited(resourceDeserved) {
		return false
	}
	return roundNumber(resourceDeserved) < roundNumber(&minAllowedAmount)
}

func isProjectQuotaStatusesSupportedInCluster(ctx context.Context, clusterId uuid.UUID) (bool, *api.Error) {
	clusterCapabilityChecker := checker.NewClusterCapabilitiesChecker(clusterId.String())
	clusterSupport, err := clusterCapabilityChecker.ValidateClusterCapability(ctx, checker.MinVersionOfProjectQuotaStatuses)
	if err != nil {
		return false, api.NewError(http.StatusInternalServerError, err.Error(), nil)
	}

	return clusterSupport, nil
}

func getResourceDeserved(resourceDeserved *float64) float64 {
	if resourceDeserved == nil {
		return -1
	}
	return roundNumber(resourceDeserved)
}

func getResourcesDeservedPtr(resources models.Resources, resourceType ResourceType) *float64 {
	switch resourceType {
	case GpuResource:
		return resources.GpuDeserved
	case CpuResource:
		if resources.CpuDeserved != nil {
			return resources.CpuDeserved
		}
	case MemoryResource:
		if resources.MemoryDeserved != nil {
			return getMemoryDeservedInMb(resources.MemoryDeserved, resources.MemoryUnits)
		}
	}

	return nil
}

func getResourcesDeserved(resources models.Resources, resourceType ResourceType) float64 {
	var deserved *float64 = nil
	switch resourceType {
	case GpuResource:
		deserved = resources.GpuDeserved
	case CpuResource:
		if resources.CpuDeserved != nil {
			deserved = resources.CpuDeserved
		}
	case MemoryResource:
		if resources.MemoryDeserved != nil {
			deserved = getMemoryDeservedInMb(resources.MemoryDeserved, resources.MemoryUnits)
		}
	default:
		deserved = nil
	}
	return DereferenceFloat64(deserved)
}

func getNodePoolName(nodePoolResources models.Resources) string {
	return nodePoolResources.NodePoolName
}

func getResourceAllocatedNonPreemptible(resourcesStatus models.ResourcesStatus, resourceType ResourceType) float64 {
	switch resourceType {
	case GpuResource:
		return ptr.Deref(resourcesStatus.AllocatedNonPreemptibleGpu, 0)
	case CpuResource:
		return ptr.Deref(resourcesStatus.AllocatedNonPreemptibleCpu, 0)
	case MemoryResource:
		return ptr.Deref(resourcesStatus.AllocatedNonPreemptibleMemory, 0)
	}
	return 0
}

func getMemoryDeservedInMb(memoryDeserved *float64, memoryUnits models.Units) *float64 {
	if memoryDeserved == nil {
		return nil
	}
	if memoryUnits == models.MB {
		return memoryDeserved
	}

	conversionFn := GetMemoryConversionFn(ModelUnitsToApi(memoryUnits), api.UNITS_MB)
	return ApplyMemoryConversion(memoryDeserved, conversionFn)
}
