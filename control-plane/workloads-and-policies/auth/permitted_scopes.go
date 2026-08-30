package auth

import (
	authClient "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"
	"github.com/run-ai/runai/runai-application-packages/datastructs"
)

// PermittedScopes is a utility struct that holds the permitted scope of an action in two ways:
// (1) as arrays (that's how they are retrieved in the "raw" response of authorization service (2) as sets
// we load the sets only when we need them
type PermittedScopes struct {
	rawScopes *authClient.PermittedScopes

	allowedClusters    datastructs.Set[string]
	allowedDepartments datastructs.Set[string]
	allowedProjects    datastructs.Set[string]
}

// NewPermittedScopes only stores the scopes in raw format and also as sets
func NewPermittedScopes(scopes *authClient.PermittedScopes) *PermittedScopes {
	return &PermittedScopes{
		rawScopes: scopes,
	}
}

func (ps *PermittedScopes) Raw() *authClient.PermittedScopes {
	return ps.rawScopes
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

func (ps *PermittedScopes) AllowedClustersArr() []string {
	return ps.rawScopes.Clusters
}

func (ps *PermittedScopes) AllowedDepartmentsArr() []string {
	return ps.rawScopes.Departments
}

func (ps *PermittedScopes) AllowedProjectsArr() []string {
	return ps.rawScopes.Projects
}

func (ps *PermittedScopes) AllowedCluster() datastructs.Set[string] {
	if ps.allowedClusters == nil {
		ps.allowedClusters = datastructs.NewSet[string](ps.rawScopes.Clusters...)
	}
	return ps.allowedClusters
}

func (ps *PermittedScopes) AllowedDepartments() datastructs.Set[string] {
	if ps.allowedDepartments == nil {
		ps.allowedDepartments = datastructs.NewSet[string](ps.rawScopes.Departments...)
	}
	return ps.allowedDepartments
}

func (ps *PermittedScopes) AllowedProjects() datastructs.Set[string] {
	if ps.allowedProjects == nil {
		ps.allowedProjects = datastructs.NewSet[string](ps.rawScopes.Projects...)
	}
	return ps.allowedProjects
}
