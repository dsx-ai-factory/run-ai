package auth

import (
	authorizationClient "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"
	"github.com/run-ai/runai/runai-application-packages/datastructs"
)

// PermittedScopes is a utility struct that holds the permitted scope of an action in two ways:
// (1) as arrays (that's how they are retrieved in the "raw" response of authorization service (2) as sets
type PermittedScopes struct {
	rawScopes *authorizationClient.PermittedScopes

	allowedClusters    datastructs.Set[string]
	allowedDepartments datastructs.Set[string]
	allowedProjects    datastructs.Set[string]
}

// NewPermittedScopes only stores the scopes in raw format. we lazy-build the sets, only if we need them
// depending on the scenario, not always we need the information in the sets. for example: if the tenant
// scope is allowed, it grants access to all scopes so in this case we won't need the sets
func NewPermittedScopes(scopes *authorizationClient.PermittedScopes) *PermittedScopes {
	return &PermittedScopes{
		rawScopes: scopes,
	}
}

func (ps *PermittedScopes) isAllowedCluster(clusterId string) bool {
	if ps.allowedClusters == nil {
		ps.allowedClusters = datastructs.NewSet[string](ps.rawScopes.Clusters...)
	}
	return ps.allowedClusters.Has(clusterId)
}

func (ps *PermittedScopes) isAllowedDepartment(departmentId string) bool {
	if ps.allowedDepartments == nil {
		ps.allowedDepartments = datastructs.NewSet[string](ps.rawScopes.Departments...)
	}
	return ps.allowedDepartments.Has(departmentId)
}

func (ps *PermittedScopes) isAllowedProject(projectId string) bool {
	if ps.allowedProjects == nil {
		ps.allowedProjects = datastructs.NewSet[string](ps.rawScopes.Projects...)
	}
	return ps.allowedProjects.Has(projectId)
}
