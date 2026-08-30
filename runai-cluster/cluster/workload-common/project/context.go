package project

import (
	"context"
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"
	logger "sigs.k8s.io/controller-runtime/pkg/log"
)

// ContextProjectName is used for injecting the project name as a string to the context
type ContextKey string

const ContextProjectName ContextKey = "projectName"

const usedForSubmit = "Submit"

// SetProjectInContext sets the project name in the context
func SetProjectInContext(ctx context.Context, project string) context.Context {
	return context.WithValue(ctx, ContextProjectName, project)
}

// SetProjectInContextFromNamespace sets the name of the project that a workload is related to in the context.
// note that a workload for submit must reside in a project. only workloads used as templates
// can reside in runai namespace (referred to as "global template")
func SetProjectInContextFromNamespace(ctx context.Context, k8sClient client.Client, namespace string, name string, usage string) (context.Context, error) {

	var project = ""
	var err error

	if namespace != "" {
		project, err = ProjectNameOfNamespace(ctx, k8sClient, namespace)
		if err != nil {
			log := logger.FromContext(ctx)
			log.Error(err, "failed to get namespace info", "ns", namespace)
			return ctx, fmt.Errorf("could not determine the project of workload %s:%s", namespace, name)
		}
	}

	if project == "" && usage == usedForSubmit {
		return ctx, fmt.Errorf("workloads with usage=Submit must reside in a project")
	}

	return context.WithValue(ctx, ContextProjectName, project), nil
}

// GetProjectFromContext retrieves the workload's project name from the context
func GetProjectFromContext(ctx context.Context) (string, error) {
	projectObj := ctx.Value(ContextProjectName)
	if projectObj == nil {
		return "", fmt.Errorf("project name not found in context")
	}

	projectName, ok := projectObj.(string)
	if !ok {
		return "", fmt.Errorf("project name is not a string")
	}

	return projectName, nil
}
