package projects

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"time"

	"golang.org/x/exp/slices"

	"k8s.io/utils/ptr"

	cluster_openapi "github.com/run-ai/runai/control-plane/cluster-service/pkg/client/generated"

	"github.com/google/uuid"
	openapi "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/org-unit-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/org-unit-service/pkg/database/models"
	"github.com/run-ai/runai/control-plane/org-unit-service/pkg/services"
	"github.com/run-ai/runai/runai-application-packages/convutils"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
)

const (
	additionalStatusDataConditions = "conditions"

	unlimitedResourcesValue = -1
)

func MapCreateRequestToModel(ctx context.Context, request api.ProjectCreationRequest, parentDepartment models.Department, nodePoolsMetadata map[string]cluster_openapi.NodepoolMinimal) (models.Project, error) {
	clusterId, err := uuid.Parse(request.ClusterId)
	if err != nil {
		return models.Project{}, err
	}
	data, err := services.MapRequestToOrgUnitData(ctx, request.SchedulingRules, request.NodeTypes, request.DefaultNodePools, request.Resources, false, nodePoolsMetadata, &parentDepartment.OrgUnitData, nil)
	if err != nil {
		return models.Project{}, err
	}
	parentUid, err := strconv.ParseInt(*request.ParentId, 10, 64)
	if err != nil {
		return models.Project{}, err
	}

	if request.EnforceRunaiScheduler == nil {
		request.EnforceRunaiScheduler = ptr.To(true)
	}

	return models.Project{
		Name:                  request.Name,
		Description:           request.Description,
		ClusterId:             clusterId,
		RequestedNamespace:    request.RequestedNamespace,
		EnforceRunaiScheduler: *request.EnforceRunaiScheduler,
		ParentID:              &parentUid,
		OrgUnitData:           data,
		ParentName:            &parentDepartment.Name,
		ProjectStatus: models.ProjectStatus{
			Phase: ptr.To(CreatingPhase),
			QuotaStatus: models.QuotaStatus{
				GpuAllocated: ptr.To(float64(0)),
			},
		},
	}, nil
}

func MapUpdateRequestToModel(
	ctx context.Context, request api.ProjectUpdateRequest, project *models.Project, nodePoolsMetadata map[string]cluster_openapi.NodepoolMinimal, parent *models.Department,
) (models.Project, error) {
	var parentOrgUnit *models.OrgUnitData
	if parent != nil {
		parentOrgUnit = &parent.OrgUnitData
	}
	if request.EnforceRunaiScheduler == nil {
		request.EnforceRunaiScheduler = ptr.To(project.EnforceRunaiScheduler)
	}

	data, err := services.MapRequestToOrgUnitData(ctx, request.SchedulingRules, request.NodeTypes, request.DefaultNodePools, request.Resources, false, nodePoolsMetadata, parentOrgUnit, &project.OrgUnitData)
	if err != nil {
		return models.Project{}, err
	}

	update := models.Project{
		IntBaseModel: models.IntBaseModel{
			ID:        project.ID,
			CreatedBy: project.CreatedBy,
			CreatedAt: project.CreatedAt,
		},
		TenantId:              project.TenantId,
		ClusterId:             project.ClusterId,
		Name:                  project.Name,
		RequestedNamespace:    project.RequestedNamespace,
		EnforceRunaiScheduler: *request.EnforceRunaiScheduler,
		ParentID:              project.ParentID,
		ParentName:            project.ParentName,
		OrgUnitDataID:         project.OrgUnitDataID,
		OrgUnitData:           data,
	}
	update.Description = request.Description
	update.ProjectStatus = project.ProjectStatus
	if clusterFieldsUpdated(data, project.OrgUnitData) {
		update.Phase = ptr.To(UpdatingPhase)
	}

	return update, nil
}

func mapUpdateResourcesRequestToModel(ctx context.Context, requestedResources []api.ResourcesNullable, project *models.Project, nodePoolsMetadata map[string]cluster_openapi.NodepoolMinimal) error {
	originalResourcesNodePoolToId := make(map[string]uuid.UUID)
	originalResourcesNodePoolIdToRank := make(map[string]string)
	for _, originalResources := range project.OrgUnitData.Resources {
		originalResourcesNodePoolToId[originalResources.NodePoolId] = originalResources.ID
		originalResourcesNodePoolIdToRank[originalResources.NodePoolId] = ptr.Deref(originalResources.Rank, models.DefaultResourceRank)
	}
	newResources, err := services.ApiResourcesToModel(ctx, requestedResources, false, nodePoolsMetadata, true, originalResourcesNodePoolToId, originalResourcesNodePoolIdToRank)
	if err != nil {
		return err
	}

	if !resourcesEqual(newResources, project.OrgUnitData.Resources) {
		project.Phase = ptr.To(UpdatingPhase)
	}

	services.UpdateOrgUnitDataResources(newResources, &project.OrgUnitData)
	return nil
}

func mapPatchResourcesRequestToModel(ctx context.Context, requestedResources []api.ResourcesNullable, project *models.Project, nodePoolsMetadata map[string]cluster_openapi.NodepoolMinimal) error {
	originalResourcesNodePoolToId := make(map[string]uuid.UUID)
	originalResourcesNodePoolIdToRank := make(map[string]string)
	for _, originalResources := range project.OrgUnitData.Resources {
		originalResourcesNodePoolToId[originalResources.NodePoolId] = originalResources.ID
		originalResourcesNodePoolIdToRank[originalResources.NodePoolId] = ptr.Deref(originalResources.Rank, models.DefaultResourceRank)
	}
	newResources, err := services.ApiResourcesToModel(ctx, requestedResources, false, nodePoolsMetadata, false, originalResourcesNodePoolToId, originalResourcesNodePoolIdToRank)
	if err != nil {
		return err
	}

	original := services.CopyResources(project.OrgUnitData.Resources)
	err = services.PatchOrgUnitDataResources(newResources, &project.OrgUnitData)
	if err != nil {
		return err
	}

	if !resourcesEqual(original, project.OrgUnitData.Resources) {
		project.Phase = ptr.To(UpdatingPhase)
	}

	return nil
}

func clusterFieldsUpdated(first models.OrgUnitData, second models.OrgUnitData) bool {
	return !resourcesEqual(first.Resources, second.Resources) ||
		!timeLimitEqual(first.InteractiveJobTimeLimitEffective, second.InteractiveJobTimeLimitEffective) ||
		!timeLimitEqual(first.InteractiveJobPreemptIdleDurationEffective, second.InteractiveJobPreemptIdleDurationEffective) ||
		!timeLimitEqual(first.TrainingJobTimeLimitEffective, second.TrainingJobTimeLimitEffective) ||
		!timeLimitEqual(first.TrainingJobMaxIdleDurationEffective, second.TrainingJobMaxIdleDurationEffective) ||
		!timeLimitEqual(first.InteractiveJobMaxIdleDurationEffective, second.InteractiveJobMaxIdleDurationEffective) ||
		!nodeTypesEqual(first.TrainingNodeTypesEffective, second.TrainingNodeTypesEffective) ||
		!nodeTypesEqual(first.WorkspaceNodeTypesEffective, second.WorkspaceNodeTypesEffective) ||
		!reflect.DeepEqual(first.DefaultNodePoolListEffective, second.DefaultNodePoolListEffective)
}

func resourcesEqual(original []models.Resources, newData []models.Resources) bool {
	if len(original) != len(newData) {
		return false
	}
	resourcesMap := map[string]models.Resources{}
	for _, resource := range original {
		resourcesMap[resource.NodePoolId] = resource
	}

	for _, resource := range newData {
		r1Resource, found := resourcesMap[resource.NodePoolId]
		if !found {
			return false
		}
		if r1Resource.NodePoolId != resource.NodePoolId ||
			derefAndRound(r1Resource.GpuLimit) != derefAndRound(resource.GpuLimit) ||
			derefAndRound(r1Resource.GpuDeserved) != derefAndRound(resource.GpuDeserved) ||
			derefAndRound(r1Resource.GpuOverQuotaWeight) != derefAndRound(resource.GpuOverQuotaWeight) ||
			derefAndRound(r1Resource.CpuLimit) != derefAndRound(resource.CpuLimit) ||
			derefAndRound(r1Resource.CpuDeserved) != derefAndRound(resource.CpuDeserved) ||
			derefAndRound(r1Resource.CpuOverQuotaWeight) != derefAndRound(resource.CpuOverQuotaWeight) ||
			!memoryEquals(r1Resource, resource) {
			return false
		}
	}
	return true
}

func memoryEquals(first models.Resources, second models.Resources) bool {
	firstUnits := first.MemoryUnits
	secondUnits := second.MemoryUnits
	if firstUnits == secondUnits {
		return derefAndRound(first.MemoryLimit) == derefAndRound(second.MemoryLimit) &&
			derefAndRound(first.MemoryDeserved) == derefAndRound(second.MemoryDeserved) &&
			derefAndRound(first.MemoryOverQuotaWeight) == derefAndRound(second.MemoryOverQuotaWeight)
	}

	conversionFn := services.GetMemoryConversionFn(services.ModelUnitsToApi(firstUnits), services.ModelUnitsToApi(secondUnits))
	deservedEquals := (first.MemoryDeserved == nil || *first.MemoryDeserved == -1) && (second.MemoryDeserved == nil || *second.MemoryDeserved == -1)
	if !deservedEquals {
		deservedEquals = conversionFn(ptr.Deref(first.MemoryDeserved, -1)) == derefAndRound(second.MemoryDeserved)
	}
	limitEquals := (first.MemoryLimit == nil || *first.MemoryLimit == -1) && (second.MemoryLimit == nil || *second.MemoryLimit == -1)
	if !limitEquals {
		limitEquals = conversionFn(ptr.Deref(first.MemoryLimit, -1)) == derefAndRound(second.MemoryLimit)
	}
	return deservedEquals && limitEquals && derefAndRound(first.MemoryOverQuotaWeight) == derefAndRound(second.MemoryOverQuotaWeight)
}

func nodeTypesEqual(original *[]models.NodeType, newData *[]models.NodeType) bool {
	if original == nil && newData == nil {
		return true
	}
	if original == nil || newData == nil {
		return false
	}
	if len(*original) != len(*newData) {
		return false
	}
	nodeTypesMap := map[int64]bool{}
	for _, nodeType := range *original {
		nodeTypesMap[nodeType.ID] = true
	}

	for _, nodeType := range *newData {
		r1NodeType, found := nodeTypesMap[nodeType.ID]
		if !found || !r1NodeType {
			return false
		}
	}
	return true
}

func timeLimitEqual(original *int32, newData *int32) bool {
	if original == nil && newData == nil {
		return true
	}
	if original == nil || newData == nil {
		return false
	}
	return *original == *newData
}

func MapModelToResponse(ctx context.Context, project models.Project, nodePoolsMetadata map[string]cluster_openapi.NodepoolMinimal) api.Project {
	orgUnitData := project.OrgUnitData
	if len(orgUnitData.Resources) == 0 {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Project %v resources is empty", project.ID)
	}
	resources := services.ApiResourcesFromModel(orgUnitData.Resources, nodePoolsMetadata)
	response := api.Project{
		Id:                    strconv.FormatInt(project.ID, 10),
		CreatedAt:             project.CreatedAt,
		UpdatedAt:             project.UpdatedAt,
		CreatedBy:             project.CreatedBy,
		UpdatedBy:             project.UpdatedBy,
		Name:                  project.Name,
		Description:           project.Description,
		RequestedNamespace:    &project.RequestedNamespace,
		EnforceRunaiScheduler: project.EnforceRunaiScheduler,
		ClusterId:             project.ClusterId.String(),
		Resources:             resources,
		Status:                MapModelStatusToApi(project),
		TotalResources: api.TotalResources{
			GpuQuota:    project.OrgUnitData.TotalGpuQuota,
			CpuQuota:    project.OrgUnitData.TotalCpuQuota,
			MemoryQuota: project.OrgUnitData.TotalMemoryQuota,
		},
		OvertimeData: services.MapOvertimeDataModelToApi(orgUnitData.OvertimeData, project.Name, "project"),
	}

	if orgUnitData.DefaultNodePoolList != nil {
		apiNodePools := []string(*orgUnitData.DefaultNodePoolList)
		response.DefaultNodePools = &apiNodePools
	}
	response.NodeTypes = services.ApiNodeTypesFromModel(orgUnitData.TrainingNodeTypes, orgUnitData.WorkspaceNodeTypes)
	schedulingRules := services.MapSchedulingRules(orgUnitData)
	if (schedulingRules != api.SchedulingRules{}) {
		response.SchedulingRules = &schedulingRules
	}
	if project.ParentID != nil {
		response.ParentId = strconv.FormatInt(*project.ParentID, 10)

		response.Parent = &api.DepartmentProjectParent{
			Id: fmt.Sprintf("%d", *project.ParentID),
		}
	}
	if project.ParentName != nil {
		if response.Parent == nil {
			response.Parent = &api.DepartmentProjectParent{}
		}
		response.Parent.Name = *project.ParentName
	}
	response.Effective = mapEffectiveProject(project.OrgUnitData)

	if project.DeletionType != nil ||
		project.DeletionRequest != nil ||
		project.ForceDelete != nil {
		response.Deletion = &api.Deletion{
			Type:      ptr.Deref(project.DeletionType, ""),
			Timestamp: ptr.Deref(project.DeletionRequest, time.Time{}),
			Force:     ptr.Deref(project.ForceDelete, false),
		}
	}
	return response
}

func MapModelToSyncResponse(project models.Project, nodePoolsMetadata map[string]cluster_openapi.NodepoolMinimal, accessRules *[]openapi.AccessRule) api.ProjectSync {
	orgUnitData := project.OrgUnitData
	resources := services.ApiResourcesFromModelSync(orgUnitData.Resources, nodePoolsMetadata)
	response := api.ProjectSync{
		Id:                    strconv.FormatInt(project.ID, 10),
		CreatedAt:             project.CreatedAt,
		UpdatedAt:             project.UpdatedAt,
		Name:                  project.Name,
		RequestedNamespace:    &project.RequestedNamespace,
		EnforceRunaiScheduler: project.EnforceRunaiScheduler,
		ClusterId:             project.ClusterId.String(),
		Resources:             resources,
	}
	if project.DeletedAt.Valid {
		response.DeletedAt = &project.DeletedAt.Time
	}

	if orgUnitData.DefaultNodePoolList != nil {
		apiNodePools := []string(*orgUnitData.DefaultNodePoolList)
		response.DefaultNodePools = &apiNodePools
	}
	response.NodeTypes = services.ApiNodeTypesFromModel(orgUnitData.TrainingNodeTypes, orgUnitData.WorkspaceNodeTypes)
	schedulingRules := services.MapSchedulingRules(orgUnitData)
	if (schedulingRules != api.SchedulingRules{}) {
		response.SchedulingRules = &schedulingRules
	}
	if project.ParentID != nil {
		response.Parent = &api.DepartmentProjectParent{
			Id: fmt.Sprintf("%d", *project.ParentID),
		}
	}
	if project.ParentName != nil {
		if response.Parent == nil {
			response.Parent = &api.DepartmentProjectParent{}
		}
		response.Parent.Name = *project.ParentName
	}

	if accessRules != nil {
		response.Permissions = mapAccessRules(*accessRules)
	}

	if project.DeletionType != nil ||
		project.DeletionRequest != nil ||
		project.ForceDelete != nil {
		response.Deletion = &api.Deletion{
			Type:      ptr.Deref(project.DeletionType, ""),
			Timestamp: ptr.Deref(project.DeletionRequest, time.Time{}),
			Force:     ptr.Deref(project.ForceDelete, false),
		}
	}

	return response
}

func mapAccessRules(rules []openapi.AccessRule) api.ProjectSyncPermissions {
	users := []string{}
	groups := []string{}
	for _, rule := range rules {
		if rule.SubjectType == openapi.SUBJECTTYPEEXTENDED_GROUP {
			groups = append(groups, rule.SubjectId)
		} else {
			users = append(users, rule.SubjectId)
		}
	}
	slices.Sort(users)
	slices.Sort(groups)

	return api.ProjectSyncPermissions{
		Users:        users,
		Groups:       groups,
		Applications: []string{},
	}
}

func mapEffectiveProject(orgUnitData models.OrgUnitData) api.ProjectEffective {
	effectiveProject := api.ProjectEffective{}
	if orgUnitData.DefaultNodePoolListEffective != nil {
		apiNodePools := []string(*orgUnitData.DefaultNodePoolListEffective)
		effectiveProject.DefaultNodePools = &apiNodePools
	}
	effectiveProject.NodeTypes = services.ApiNodeTypesFromModel(orgUnitData.TrainingNodeTypesEffective, orgUnitData.WorkspaceNodeTypesEffective)
	schedulingRules := services.MapSchedulingRulesEffective(orgUnitData)
	if (schedulingRules != api.SchedulingRules{}) {
		effectiveProject.SchedulingRules = &schedulingRules
	}
	return effectiveProject
}

func MapModelToGetMetadataResponse(project models.Project) api.ProjectMetadata {
	return api.ProjectMetadata{
		Id:                    strconv.FormatInt(project.ID, 10),
		Name:                  project.Name,
		Namespace:             project.StatusNamespace,
		EnforceRunaiScheduler: project.EnforceRunaiScheduler,
		Phase:                 project.Phase,
		ClusterId:             project.ClusterId.String(),
		CreatedAt:             project.CreatedAt,
		Allocated: api.QuotaStatusResource{
			Gpu:    services.DereferenceFloat64(project.QuotaStatus.GpuAllocated),
			Cpu:    services.DereferenceFloat64(project.QuotaStatus.CpuAllocated),
			Memory: services.DereferenceFloat64(project.QuotaStatus.MemoryAllocated),
		},
		Deserved: api.QuotaStatusResource{
			Gpu:    services.DereferenceFloat64(project.OrgUnitData.TotalGpuQuota),
			Cpu:    services.DereferenceFloat64(project.OrgUnitData.TotalCpuQuota),
			Memory: services.DereferenceFloat64(project.OrgUnitData.TotalMemoryQuota),
		},
	}
}

func resourcesToQuotaStatus(resources models.Resources) api.QuotaStatusNodePool {
	return api.QuotaStatusNodePool{
		Allocated: api.QuotaStatusResource{
			Gpu:    ptr.Deref(resources.AllocatedGpu, 0),
			Cpu:    ptr.Deref(resources.AllocatedCpu, 0),
			Memory: ptr.Deref(resources.AllocatedMemory, 0),
		},
		AllocatedNonPreemptible: api.QuotaStatusResource{
			Gpu:    ptr.Deref(resources.AllocatedNonPreemptibleGpu, 0),
			Cpu:    ptr.Deref(resources.AllocatedNonPreemptibleCpu, 0),
			Memory: ptr.Deref(resources.AllocatedNonPreemptibleMemory, 0),
		},
		Requested: api.QuotaStatusResource{
			Gpu:    ptr.Deref(resources.RequestedGpu, 0),
			Cpu:    ptr.Deref(resources.RequestedCpu, 0),
			Memory: ptr.Deref(resources.RequestedMemory, 0),
		},
		NodePoolName: resources.NodePoolName,
		NodePoolId:   resources.NodePoolId,
	}
}

func MapModelStatusToApi(project models.Project) api.ProjectStatus {
	status := project.ProjectStatus
	quotaStatusNodePools := convutils.ConvertSliceTypes(project.OrgUnitData.Resources, resourcesToQuotaStatus)
	return api.ProjectStatus{
		Phase:                 mapModelPhaseToApi(status.Phase),
		PhaseMessage:          status.PhaseMessage,
		Namespace:             status.StatusNamespace,
		NodePoolQuotaStatuses: quotaStatusNodePools,
		QuotaStatus:           services.MapQuotaStatusModelToApi(status.QuotaStatus),
		AdditionalStatusData: &map[string]interface{}{
			additionalStatusDataConditions: status.Conditions,
		},
	}
}

func MapApiStatusToModel(status api.ProjectStatus, project *models.Project) (models.ProjectStatus, map[string]models.ResourcesStatus) {
	var phase *string = nil
	if status.Phase != "" {
		phase = ptr.To(string(status.Phase))
	}
	projectStatus := models.ProjectStatus{
		Phase:                 phase,
		PhaseMessage:          status.PhaseMessage,
		StatusNamespace:       status.Namespace,
		Conditions:            ConditionsApiToModel(status.AdditionalStatusData),
		StatusLastUpdatedTime: status.LastUpdatedTime,
		QuotaStatus: models.QuotaStatus{
			GpuAllocated:                  ptr.To(status.QuotaStatus.Allocated.Gpu),
			GpuAllocatedNonPreemptible:    ptr.To(status.QuotaStatus.AllocatedNonPreemptible.Gpu),
			GpuRequested:                  ptr.To(status.QuotaStatus.Requested.Gpu),
			CpuAllocated:                  ptr.To(status.QuotaStatus.Allocated.Cpu),
			CpuAllocatedNonPreemptible:    ptr.To(status.QuotaStatus.AllocatedNonPreemptible.Cpu),
			CpuRequested:                  ptr.To(status.QuotaStatus.Requested.Cpu),
			MemoryAllocated:               ptr.To(status.QuotaStatus.Allocated.Memory),
			MemoryAllocatedNonPreemptible: ptr.To(status.QuotaStatus.AllocatedNonPreemptible.Memory),
			MemoryRequested:               ptr.To(status.QuotaStatus.Requested.Memory),
		},
	}
	resourcesStatus := make(map[string]models.ResourcesStatus)
	for _, nodePoolStatus := range status.NodePoolQuotaStatuses {
		resourcesStatus[nodePoolStatus.NodePoolName] = createResourcesStatus(nodePoolStatus)
	}
	for _, nodePoolResources := range project.OrgUnitData.Resources {
		_, found := resourcesStatus[nodePoolResources.NodePoolName]
		if !found {
			resourcesStatus[nodePoolResources.NodePoolName] = models.ResourcesStatus{}
		}
	}
	return projectStatus, resourcesStatus
}

func createResourcesStatus(nodePoolStatus api.QuotaStatusNodePool) models.ResourcesStatus {
	return models.ResourcesStatus{
		AllocatedGpu:                  services.MapAllocationStatus(nodePoolStatus.Allocated.Gpu),
		AllocatedCpu:                  services.MapAllocationStatus(nodePoolStatus.Allocated.Cpu),
		AllocatedMemory:               services.MapAllocationStatus(nodePoolStatus.Allocated.Memory),
		AllocatedNonPreemptibleGpu:    services.MapAllocationStatus(nodePoolStatus.AllocatedNonPreemptible.Gpu),
		AllocatedNonPreemptibleCpu:    services.MapAllocationStatus(nodePoolStatus.AllocatedNonPreemptible.Cpu),
		AllocatedNonPreemptibleMemory: services.MapAllocationStatus(nodePoolStatus.AllocatedNonPreemptible.Memory),
		RequestedGpu:                  services.MapAllocationStatus(nodePoolStatus.Requested.Gpu),
		RequestedCpu:                  services.MapAllocationStatus(nodePoolStatus.Requested.Cpu),
		RequestedMemory:               services.MapAllocationStatus(nodePoolStatus.Requested.Memory),
	}
}

func mapModelPhaseToApi(phase *string) api.ProjectPhase {
	var apiPhase api.ProjectPhase = ""
	if phase != nil {
		apiPhase = api.ProjectPhase(*phase)
	}
	if apiPhase == "" {
		return api.PROJECTPHASE_UNKNOWN
	}
	return apiPhase
}

func ConditionsApiToModel(additionalData *map[string]interface{}) []models.Condition {
	if additionalData == nil {
		return []models.Condition{}
	}

	conditions, ok := (*additionalData)[additionalStatusDataConditions]
	if !ok {
		return []models.Condition{}
	}

	marshaled, err := json.Marshal(conditions)
	if err != nil {
		return []models.Condition{}
	}

	result := []models.Condition{}
	err = json.Unmarshal(marshaled, &result)
	if err != nil {
		return []models.Condition{}
	}
	return result
}

func derefAndRound(resourcesVal *float64) float64 {
	if resourcesVal == nil {
		return unlimitedResourcesValue
	}
	return roundResources(*resourcesVal)
}

func roundResources(resourcesVal float64) float64 {
	// the resources db columns are defined as DECIMAL(15,3)
	factor := float64(1000)
	return math.Round(resourcesVal*factor) / factor
}
