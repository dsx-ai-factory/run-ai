package migration

import (
	"context"
	"fmt"
	"strconv"

	orgUnitServiceServicesV1 "github.com/run-ai/runai/control-plane/org-unit-service/pkg/services/v1"

	logging "github.com/run-ai/runai/runai-common-packages/logging"

	clusteropenapi "github.com/run-ai/runai/control-plane/cluster-service/pkg/client/generated"

	openapi "github.com/run-ai/runai/control-plane/org-unit-service/generated/pkg/api"
	orgUnitsServiceDao "github.com/run-ai/runai/control-plane/org-unit-service/pkg/database/dao"
	orgUnitServiceModels "github.com/run-ai/runai/control-plane/org-unit-service/pkg/database/models"
	"github.com/run-ai/runai/control-plane/org-unit-service/pkg/services/projects"
	"github.com/run-ai/runai/control-plane/org-units-migrator/database"
	"github.com/run-ai/runai/control-plane/org-units-migrator/database/models"
	"github.com/run-ai/runai/runai-application-packages/convutils"
	"k8s.io/utils/ptr"
)

func CreateProjects(tenant models.Tenant, projects []models.ProjectV1Grouped, settings map[string]bool) error {
	projectsToCreate := []*orgUnitServiceModels.Project{}
	nodePools := createStubNodepoolsForProjects(projects, settings)
	for _, projectV1 := range projects {
		logging.Debugf("Migrating project %s of tenant %s.", projectV1.Name, tenant.Name)
		projectV2, err := MapProjectV1ToV2(projectV1, settings, nodePools)
		if err != nil {
			logging.Errorf("Failed to convert project %s of tenant %s to V2 struct. Err: %v", projectV1.Name, tenant.Name, err)
			return err
		}
		projectsToCreate = append(projectsToCreate, projectV2)
	}

	err := orgUnitsServiceDao.CreateProjects(context.Background(), projectsToCreate)
	if err != nil {
		logging.Errorf("Failed to create projects of tenant %s. Err: %v", tenant.Name, err)
		return err
	}

	return nil
}

func mapSchedulingRule(schedulingRuleV1 int64) *int32 {
	if schedulingRuleV1 == 0 {
		return nil
	}
	return ptr.To[int32](int32(schedulingRuleV1))
}

func MapProjectV1ToV2(projectV1 models.ProjectV1Grouped, settings map[string]bool, nodePools map[string]clusteropenapi.NodepoolMinimal) (*orgUnitServiceModels.Project, error) {
	projectV2Request := openapi.ProjectCreationRequest{
		SchedulingRules: &openapi.SchedulingRules{
			InteractiveJobTimeLimitSeconds:           mapSchedulingRule(projectV1.InteractiveJobTimeLimitSec),
			InteractiveJobMaxIdleDurationSeconds:     mapSchedulingRule(projectV1.InteractiveJobMaxIdleSec),
			InteractiveJobPreemptIdleDurationSeconds: mapSchedulingRule(projectV1.InteractivePreemptibleJobMaxIdleSec),
			TrainingJobMaxIdleDurationSeconds:        mapSchedulingRule(projectV1.TrainingJobMaxIdleSec),
			TrainingJobTimeLimitSeconds:              mapSchedulingRule(projectV1.TrainingJobTimeLimitSec),
		},
		DefaultNodePools: projectV1.DefaultNodePools,
		NodeTypes: openapi.NodeTypesPerWorkload{
			Training:  projectV1.NodeAffinityTrainingIds,
			Workspace: projectV1.NodeAffinityWorkspaceIds,
		},
		Resources:          convutils.ConvertSliceTypes(projectV1.NodepoolResources, getNodepoolResourcesToResourcesV2(settings)),
		Name:               projectV1.Name,
		ClusterId:          projectV1.ClusterUuid.String(),
		RequestedNamespace: projectV1.Namespace,
		ParentId:           ptr.To[string](fmt.Sprintf("%d", projectV1.DepartmentId)),
	}
	projectV2, err := projects.MapCreateRequestToModel(context.Background(), projectV2Request, orgUnitServiceModels.Department{Name: projectV1.DepartmentName}, nodePools)
	if err != nil {
		return nil, err
	}

	projectV2.ID = int64(projectV1.Id)
	projectV2.TenantId = fmt.Sprintf("%d", projectV1.TenantId)
	projectV2.CreatedAt = projectV1.CreatedAt
	projectV2.CreatedBy = projectV1.CreatedBy
	projectV2.UpdatedBy = projectV1.UpdatedBy
	projectV2.OrgUnitData.CreatedAt = projectV1.CreatedAt
	projectV2.OrgUnitData.CreatedBy = projectV1.CreatedBy
	projectV2.OrgUnitData.UpdatedBy = projectV1.UpdatedBy

	projectV1StatusRequest := openapi.ProjectV1StatusRequest{}
	if projectV1.Status != nil {
		projectV1StatusRequest.Status = *projectV1.Status
	}
	if projectV1.QuotaStatuses != nil {
		projectV1StatusRequest.QuotaStatuses = *projectV1.QuotaStatuses
	}

	projectStatus, resoucesStatus, err := orgUnitServiceServicesV1.MapProjectStatusApiV1ToModel(projectV1StatusRequest, &projectV2)
	if err == nil {
		projectV2.ProjectStatus = projectStatus
		for _, resources := range projectV2.OrgUnitData.Resources {
			resources.ResourcesStatus = mapResourcesStatus(resoucesStatus, resources.NodePoolName)
		}
	}
	if projectV2.Phase == nil || *projectV2.Phase == "" {
		projectV2.Phase = projectV1.Phase
	}

	return &projectV2, nil
}

func createStubNodepoolsForProjects(projects []models.ProjectV1Grouped, settings map[string]bool) map[string]clusteropenapi.NodepoolMinimal {
	result := make(map[string]clusteropenapi.NodepoolMinimal)
	for _, project := range projects {
		for _, assignedResources := range project.NodepoolResources {
			nodePoolId := strconv.Itoa(int(assignedResources.NodePoolID))
			if _, exists := result[nodePoolId]; !exists {
				result[nodePoolId] = clusteropenapi.NodepoolMinimal{
					Id: strconv.Itoa(int(assignedResources.NodePoolID)),
					SchedulingCapabilities: &clusteropenapi.SchedulingCapabilities{
						CpuQuota:          settings[database.CPUResourcesQuota],
						AdjustableWeights: settings[database.ProjectManualOverQuota],
					},
				}
			}
		}
	}
	return result
}

func mapResourcesStatus(resoucesStatus map[string]orgUnitServiceModels.ResourcesStatus, nodePoolName string) orgUnitServiceModels.ResourcesStatus {
	statusForResources, found := resoucesStatus[nodePoolName]
	if !found {
		return orgUnitServiceModels.ResourcesStatus{}
	}
	return orgUnitServiceModels.ResourcesStatus{
		AllocatedGpu:                  statusForResources.AllocatedGpu,
		AllocatedCpu:                  statusForResources.AllocatedCpu,
		AllocatedMemory:               statusForResources.AllocatedMemory,
		AllocatedNonPreemptibleGpu:    statusForResources.AllocatedNonPreemptibleGpu,
		AllocatedNonPreemptibleCpu:    statusForResources.AllocatedNonPreemptibleCpu,
		AllocatedNonPreemptibleMemory: statusForResources.AllocatedNonPreemptibleMemory,
		RequestedGpu:                  statusForResources.RequestedGpu,
		RequestedCpu:                  statusForResources.RequestedCpu,
		RequestedMemory:               statusForResources.RequestedMemory,
	}
}

func getNodepoolResourcesToResourcesV2(settings map[string]bool) func(resource models.AssignedResource) openapi.ResourcesNullable {
	return func(assignedResources models.AssignedResource) openapi.ResourcesNullable {
		gpuDeserved := float32(-1)
		if assignedResources.GPUDeserved != nil {
			gpuDeserved = *(assignedResources.GPUDeserved)
		}
		return openapi.ResourcesNullable{
			NodePool: &openapi.ResourcesNodePool{
				Id:   fmt.Sprintf("%d", assignedResources.NodePoolID),
				Name: assignedResources.NodePoolName,
			},
			Gpu: openapi.ResourcesNullableGpu{
				Limit:           float32To64Ptr(assignedResources.GPUMaxAllowed),
				Deserved:        float64(gpuDeserved),
				OverQuotaWeight: convertProjectOQWToV2(assignedResources.GPUOverQuotaWeight, &gpuDeserved, settings),
			},
			Cpu: &openapi.ResourcesCpu{
				Limit:           float32To64Ptr(assignedResources.CPUMaxAllowed),
				Deserved:        float32To64Ptr(assignedResources.CPUDeserved),
				OverQuotaWeight: convertProjectOQWToV2(assignedResources.CPUOverQuotaWeight, assignedResources.CPUDeserved, settings),
			},
			Memory: &openapi.ResourcesMemory{
				Limit:           float32To64Ptr(assignedResources.MemoryMaxAllowed),
				Deserved:        float32To64Ptr(assignedResources.MemoryDeserved),
				OverQuotaWeight: convertProjectOQWToV2(assignedResources.MemoryOverQuotaWeight, assignedResources.MemoryDeserved, settings),
				Units:           "Mib",
			},
		}
	}
}

func float32To64Ptr(val *float32) *float64 {
	if val == nil {
		return nil
	}
	return ptr.To(float64(*val))
}

func convertProjectOQWToV2(oqw *float32, deserved *float32, settings map[string]bool) *float64 {
	if settings[database.ProjectManualOverQuota] {
		return float32To64Ptr(deserved)
	}

	if oqw != nil && *oqw == -1 {
		return ptr.To[float64](2)
	}

	return float32To64Ptr(oqw)
}
