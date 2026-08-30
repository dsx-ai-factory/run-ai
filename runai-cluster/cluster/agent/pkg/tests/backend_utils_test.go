package test

import (
	"time"

	auth_openapi "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"
	openapi2 "github.com/run-ai/runai/control-plane/cluster-service/pkg/client/generated"
	workloads_openapi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"k8s.io/utils/ptr"

	"github.com/onsi/gomega/ghttp"
	org_unit_openapi "github.com/run-ai/runai/runai-clients/org-unit-service-client/generated"
	"github.com/run-ai/runai/runai-cluster/cluster/agent/pkg/generated/datavolumes_openapi"

	"github.com/run-ai/runai/runai-cluster/cluster/agent/pkg/generated/backend_openapi"
)

// update in test so project2 will be of department1 instead of department2
var Project2InDepartment2 = true

var NodePoolsFeatureTurnedOff = false

func CreateBasicRunaiBackend() *ghttp.Server {
	return ghttp.NewTLSServer()
}

func createProjectsOrgUnitGetResponse(projectNamesToReturn ...string) *org_unit_openapi.GetSyncProjects200Response {
	result := []org_unit_openapi.ProjectSync{}
	for _, projectToReturn := range projectNamesToReturn {
		if !Project2InDepartment2 && projectToReturn == Project2Name {
			orgUnitProjToReturn := *OrgUnitProjectsMap[projectToReturn]
			orgUnitProjToReturn.Parent = *org_unit_openapi.NewNullableDepartmentProjectParent(&org_unit_openapi.DepartmentProjectParent{
				Id:   Dep1IdStr,
				Name: Department1Name,
			})
			result = append(result, orgUnitProjToReturn)
			continue
		}
		result = append(result, *OrgUnitProjectsMap[projectToReturn])
	}
	return &org_unit_openapi.GetSyncProjects200Response{
		Projects: result,
	}
}

func createDepartmentsOrgUnitGetResponse(departmentsNamesToReturn ...string) *org_unit_openapi.GetSyncDepartments200Response {
	result := []org_unit_openapi.DepartmentSync{}
	for _, departmentToReturn := range departmentsNamesToReturn {
		result = append(result, *OrgUnitDepartmentsMap[departmentToReturn])
	}
	return &org_unit_openapi.GetSyncDepartments200Response{
		Departments: result,
	}
}

func createProjectsOrgUnitGetDelResponse() *org_unit_openapi.GetSyncProjects200Response {
	return &org_unit_openapi.GetSyncProjects200Response{
		Projects: []org_unit_openapi.ProjectSync{OrgUnitProject3del, OrgUnitProject4delOtherId},
	}
}

func createNodePoolsGetResponse(nodepoolNamesToReturn ...string) []openapi2.Nodepool {
	result := []openapi2.Nodepool{}
	if !NodePoolsFeatureTurnedOff {
		for _, nodepoolToReturn := range nodepoolNamesToReturn {
			result = append(result, *OpenApiNodepoolsMap[nodepoolToReturn])
		}
	} else {
		result = append(result, *OpenApiNodepoolsMap[NodePoolDName])
	}
	return result
}

func createDatavolumesGetResponse(datavolumeNamesToReturn ...string) *datavolumes_openapi.GetDatavolumesInternal200Response {
	dvs := []datavolumes_openapi.DatavolumeInternal{}
	for _, datavolumeToReturn := range datavolumeNamesToReturn {
		dvs = append(dvs, *BackendDatavolumesMap[datavolumeToReturn])
	}
	result := &datavolumes_openapi.GetDatavolumesInternal200Response{
		Datavolumes: dvs,
	}

	return result
}

func createAccessRulesGetResponse(accessrulesToReturn ...auth_openapi.AccessRuleSync) *auth_openapi.GetAccessRulesSync200Response {
	return &auth_openapi.GetAccessRulesSync200Response{
		AccessRules: accessrulesToReturn,
	}
}

func createUsersGetResponse(userNamesToReturn ...string) *backend_openapi.Users {
	result := []backend_openapi.User{}
	for _, userNameToReturn := range userNamesToReturn {
		result = append(result, *BackendUsersGroupsMap[userNameToReturn])
	}
	return &backend_openapi.Users{Users: result}
}

func createGroupsGetResponse(groupNamesToReturn ...string) *backend_openapi.Groups {
	result := []backend_openapi.User{}
	for _, groupNameToReturn := range groupNamesToReturn {
		result = append(result, *BackendUsersGroupsMap[groupNameToReturn])
	}
	return &backend_openapi.Groups{Groups: result}
}

func createListWorkloadsTypesResponse() *workloads_openapi.ListWorkloadTypes200Response {
	return &workloads_openapi.ListWorkloadTypes200Response{
		Types: backendWorkloadTypeConfigs,
	}
}

func createListWorkloadsCategories() *workloads_openapi.ListCategories200Response {
	categories := []workloads_openapi.WorkloadCategory{{
		Id:          "cat1",
		Name:        "Train",
		Description: ptr.To("DescriptionA"),
	}, {
		Id:          "cat2",
		Name:        "Deploy",
		Description: ptr.To("DescriptionB"),
	}}
	return &workloads_openapi.ListCategories200Response{
		Categories: categories,
	}
}

func createListWorkloadsPrioritiesResponse() *[]workloads_openapi.SyncPriority {
	return &[]workloads_openapi.SyncPriority{
		{
			Id:     priorityClassAddedId,
			Source: ptr.To(constants.PriorityClassRunaiSource),
			Name:   priorityClassAddedName,
			Value:  10,
		},
		{
			Id:     priorityClassUpdatedId,
			Source: ptr.To(constants.PriorityClassRunaiSource),
			Name:   priorityClassUpdatedName,
			Value:  20,
		},
		{
			Id:     priorityClassNonRunaiId,
			Source: ptr.To(constants.PriorityClassRunaiSource),
			Name:   priorityClassNonRunaiName,
			Value:  200,
		},
		{
			Id:        priorityClassDeletedId,
			Source:    ptr.To(constants.PriorityClassRunaiSource),
			Name:      priorityClassDeletedName,
			Value:     30,
			DeletedAt: *workloads_openapi.NewNullableTime(ptr.To(time.Now())),
		},
	}
}
