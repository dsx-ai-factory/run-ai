package testing

import (
	"context"
	"fmt"
	"strings"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/authorization"
	scope "github.com/dsx-ai-factory/run-ai/runai-application-packages/scopeinfo"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/testutils"
	orgUnitApi "github.com/dsx-ai-factory/run-ai/runai-clients/org-unit-service-client/generated"

	authClient "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/auth"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/datastructs"
)

// RandomsBasedAuthMockClient is a mock client that returns clusterId + departmentId based on
// the set of random values created for the test
type RandomsBasedAuthMockClient struct {
	TenantId   int32
	ClusterIds map[string]string

	AllClusters    datastructs.Set[string]
	AllDepartments datastructs.Set[string]
	AllProjects    map[string]*orgUnitApi.ProjectMetadata

	DepartmentOfProject map[string]string

	AllowedClusters    datastructs.Set[string]
	AllowedDepartments datastructs.Set[string]
	AllowedProjects    map[string]*orgUnitApi.ProjectMetadata

	TranslationTable map[string]*testutils.RandomValues
}

func NewRandomsBasedAuthMockClient(randoms ...*testutils.RandomValues) *RandomsBasedAuthMockClient {
	result := RandomsBasedAuthMockClient{
		TranslationTable: map[string]*testutils.RandomValues{},

		AllClusters:    datastructs.NewSet[string](),
		AllDepartments: datastructs.NewSet[string](),
		AllProjects:    map[string]*orgUnitApi.ProjectMetadata{},

		AllowedClusters:    datastructs.NewSet[string](),
		AllowedDepartments: datastructs.NewSet[string](),
		AllowedProjects:    map[string]*orgUnitApi.ProjectMetadata{},

		DepartmentOfProject: map[string]string{},

		ClusterIds: make(map[string]string),
	}

	clusterIds := datastructs.NewSet[string]()

	for idx, randomsSet := range randoms {
		if idx == 0 {
			result.TenantId = int32(randomsSet.TenantId)
		}

		result.TranslationTable[randomsSet.ProjectIdStr] = randomsSet
		result.TranslationTable[randomsSet.DepartmentId] = randomsSet

		projectInfo := orgUnitApi.ProjectMetadata{
			Id:        randomsSet.ProjectIdStr,
			Name:      fmt.Sprintf("%d", randomsSet.ProjectId), // in testing we assume project id = project name
			ClusterId: randomsSet.ClusterId,
		}

		result.DepartmentOfProject[randomsSet.ProjectIdStr] = randomsSet.DepartmentId

		result.AllClusters.Add(randomsSet.ClusterId)
		result.AllDepartments.Add(randomsSet.DepartmentId)
		result.AllProjects[randomsSet.ProjectIdStr] = &projectInfo

		if !randomsSet.IsForbidden {
			result.AllowedClusters.Add(randomsSet.ClusterId)
			result.AllowedDepartments.Add(randomsSet.DepartmentId)
			result.AllowedProjects[randomsSet.ProjectIdStr] = &projectInfo
		}

		if !clusterIds.Has(randomsSet.ClusterId) {
			result.ClusterIds[randomsSet.ClusterId] = fmt.Sprintf("cluster-%s", randomsSet.ClusterId)
		}
	}

	return &result
}

func (mock *RandomsBasedAuthMockClient) Client() authorization.ClientInterface {
	return mock
}

// GetAncestors returns the organizational hierarchy (tenant, cluster, department) for a resource
func (mock *RandomsBasedAuthMockClient) GetAncestors(_ context.Context, _ scope.Scope, resourceId string) (*orgUnitApi.Ancestors, *orgUnitApi.Error) {
	response := orgUnitApi.Ancestors{
		TenantId: fmt.Sprintf("%d", mock.TenantId),
	}

	if randoms, ok := mock.TranslationTable[resourceId]; ok {
		response.SetClusterUuid(randoms.ClusterId)
		response.SetDepartmentId(randoms.DepartmentId)
	}

	return &response, nil
}

func isAdmin(ctx context.Context) bool {
	return strings.HasPrefix(auth.SubjectIdOf(ctx), "admin")
}

func (mock *RandomsBasedAuthMockClient) CheckPermittedScopes(ctx context.Context, _ authClient.ResourceType, _ authClient.Action) (*authClient.PermittedScopes, *authClient.Error) {
	tenantIdStr := fmt.Sprintf("%d", auth.TenantIdOf(ctx))
	result := authClient.PermittedScopes{}

	if isAdmin(ctx) {
		result.Tenants = []string{tenantIdStr}
	} else {
		for departmentId := range mock.AllowedDepartments {
			result.Departments = append(result.Departments, departmentId)
		}
		for projectId := range mock.AllowedProjects {
			result.Projects = append(result.Projects, projectId)
		}
	}
	return &result, nil
}
