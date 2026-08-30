package services

import (
	"context"
	"fmt"
	"net/http"

	"github.com/run-ai/runai/control-plane/k8s-objects-tracker/internal/generated/pkg/api"
	orgUnit "github.com/run-ai/runai/control-plane/k8s-objects-tracker/pkg/clients/org-unit"
	"github.com/run-ai/runai/control-plane/k8s-objects-tracker/pkg/projects"
	"github.com/run-ai/runai/runai-application-packages/convutils"
	"github.com/run-ai/runai/runai-common-packages/cache/v2"
	"github.com/run-ai/runai/runai-common-packages/logging"
)

func namespaceToProjectId(ctx context.Context, clusterId string, namespace string) (string, *api.Error) {
	projects.InitProjectMapperCache()

	// Check if the project id is already cached in the specific cluster
	cachedResult, err := cache.Get[projects.NamespaceProjectMapping](ctx, clusterId)
	if err != nil {
		logging.Warnf("Failed to get mapping from cache for cluster %s: %v", clusterId, err)
	}
	if cachedResult != nil && cachedResult.Mapping != nil {
		projectId, found := cachedResult.Mapping[projects.BuildCacheKey(clusterId, namespace)]
		if found {
			return projectId, nil
		}
	}

	// If the project id is not cached, fetch all projects of the cluster and cache them
	projectsOfCluster, listErr := orgUnit.Get().GetProjectsMetadataByCluster(ctx, clusterId)
	if listErr != nil {
		return "", api.NewError(int(listErr.Code), listErr.Message, nil)
	}

	nsToProjIdMap := make(map[string]string)

	for _, project := range projectsOfCluster {
		if project.Namespace.Get() != nil {
			currentNamespace := convutils.Deref(project.Namespace.Get(), "")
			if currentNamespace == "" {
				currentNamespace = fmt.Sprintf("runai-%s", project.Name)
			}
			nsToProjIdMap[projects.BuildCacheKey(clusterId, currentNamespace)] = project.Id
		}
	}

	err = cache.Set[projects.NamespaceProjectMapping](ctx, clusterId, projects.NamespaceProjectMapping{Mapping: nsToProjIdMap})
	if err != nil {
		logging.Warnf("Failed to set mapping in cache for cluster %s: %v", clusterId, err)
	}

	projectId, ok := nsToProjIdMap[projects.BuildCacheKey(clusterId, namespace)]
	if !ok {
		return "", api.NewError(http.StatusNotFound, fmt.Sprintf("ProjectId for namespace %s not found in cluster %s", namespace, clusterId), nil)
	}

	return projectId, nil
}
