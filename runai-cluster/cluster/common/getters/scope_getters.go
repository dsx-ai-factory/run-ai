package getters

import (
	"context"

	kaiorg "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"github.com/rs/zerolog/log"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// getKaiProjectsForScopes returns the kai.resources projects for the given scopes.
// projectsDepartmentFieldIndexName must be indexed on the kai Project.
func getKaiProjectsForScopes(ctx context.Context, client client.Reader, scopes []v1alpha1.DataVolumeSharedScope,
	projectsDepartmentFieldIndexName string) ([]kaiorg.Project, error) {
	result := []kaiorg.Project{}
	for _, scope := range scopes {
		var err error
		var projects []kaiorg.Project

		switch scope.ScopeType {
		case v1alpha1.ProjectScope:
			projects, err = getKaiProjectsForProjectScope(ctx, client, scope.ScopeName)
		case v1alpha1.DepartmentScope:
			projects, err = getKaiProjectsForDepartmentScope(ctx, client, scope.ScopeName, projectsDepartmentFieldIndexName)
		case v1alpha1.ClusterScope:
			projects, err = getKaiProjectsForClusterScope(ctx, client)
		}

		if err != nil {
			return []kaiorg.Project{}, err
		}
		result = append(result, projects...)
	}

	return uniqueKaiProjects(result), nil
}

// GetNamespacesForKaiScopes returns a map of namespaces for the given scopes.
// projectsDepartmentFieldIndexName must be indexed on the kai Project.
func GetNamespacesForKaiScopes(ctx context.Context, client client.Client, scopes []v1alpha1.DataVolumeSharedScope,
	projectsDepartmentFieldIndexName string) (map[string]bool, error) {
	projects, err := getKaiProjectsForScopes(ctx, client, scopes, projectsDepartmentFieldIndexName)
	if err != nil {
		return nil, err
	}

	return getNamespacesForKaiProjects(ctx, projects, client)
}

func getKaiProjectsForProjectScope(ctx context.Context, client client.Reader, projectName string) ([]kaiorg.Project, error) {
	project, err := getKaiProject(ctx, projectName, client)
	if err != nil {
		return []kaiorg.Project{}, err
	}

	return []kaiorg.Project{*project}, nil
}

func getKaiProjectsForDepartmentScope(ctx context.Context, client client.Reader, departmentName string,
	projectsDepartmentFieldIndexName string) ([]kaiorg.Project, error) {
	// first, validate the department exists
	_, err := getKaiDepartment(ctx, client, departmentName)
	if err != nil {
		return []kaiorg.Project{}, err
	}

	fieldSelector := fields.OneTermEqualSelector(projectsDepartmentFieldIndexName, departmentName)
	projectsList, err := listKaiProjectsWithFieldSelector(ctx, client, fieldSelector)
	if err != nil {
		return []kaiorg.Project{}, err
	}
	return projectsList.Items, nil
}

func getKaiProjectsForClusterScope(ctx context.Context, client client.Reader) ([]kaiorg.Project, error) {
	projects, err := listKaiProjects(ctx, client)
	if err != nil {
		log.Ctx(ctx).Error().Msgf("Failed listing all kai projects in cluster, error: %s", err.Error())
		return []kaiorg.Project{}, err
	}

	return projects.Items, nil
}

func listKaiProjects(ctx context.Context, client client.Reader, opts ...client.ListOption) (*kaiorg.ProjectList, error) {
	projects := &kaiorg.ProjectList{}
	err := client.List(ctx, projects, opts...)
	if err != nil {
		return nil, err
	}
	return projects, nil
}

func listKaiProjectsWithFieldSelector(ctx context.Context, cachedClient client.Reader, fieldSelector fields.Selector) (*kaiorg.ProjectList, error) {
	projects, err := listKaiProjects(ctx, cachedClient, &client.ListOptions{FieldSelector: fieldSelector})
	if err != nil {
		log.Ctx(ctx).Error().Msgf("Failed listing kai projects with field selector <%s>, error: %s",
			fieldSelector.String(), err.Error())
		return nil, err
	}
	return projects, nil
}

func getKaiDepartment(ctx context.Context, client client.Reader, departmentName string) (*kaiorg.Department, error) {
	department := &kaiorg.Department{}
	err := client.Get(ctx, types.NamespacedName{Name: departmentName}, department)
	if err != nil {
		log.Ctx(ctx).Error().Msgf("Failed getting kai department <%s>, error: %s", departmentName, err.Error())
		return nil, err
	}
	return department, nil
}

func getKaiProject(ctx context.Context, projectName string, client client.Reader) (*kaiorg.Project, error) {
	project := &kaiorg.Project{}
	err := client.Get(ctx, types.NamespacedName{Name: projectName}, project)
	if err != nil {
		log.Error().Msgf("Failed to get kai project %s, error: %s", projectName, err.Error())
		return nil, err
	}
	return project, nil
}

func uniqueKaiProjects(projects []kaiorg.Project) []kaiorg.Project {
	unique := []kaiorg.Project{}
	seen := map[string]bool{}
	for _, project := range projects {
		if _, ok := seen[project.Name]; !ok {
			unique = append(unique, project)
			seen[project.Name] = true
		}
	}
	return unique
}
