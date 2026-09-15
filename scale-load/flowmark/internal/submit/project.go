/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package submit

import (
	"context"
	"fmt"
	"net/http"
	"time"

	clusterclient "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	orgunitclient "github.com/dsx-ai-factory/run-ai/runai-clients/org-unit-service-client/generated"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/dsx-ai-factory/run-ai/scale-load/flowmark/internal/client"
)

const pollInterval = 2 * time.Second

func Project(ctx context.Context, clients *client.Clients, clusterID, name string, timeout time.Duration) (ProjectRef, error) {
	pools, err := nodepools(ctx, clients, clusterID)
	if err != nil {
		return ProjectRef{}, err
	}

	id, err := createProject(ctx, clients, projectRequest(name, clusterID, pools))
	if err != nil {
		return ProjectRef{}, err
	}

	namespace, err := waitProjectReady(ctx, clients, id, name, timeout)
	if err != nil {
		return ProjectRef{}, err
	}
	return ProjectRef{ID: id, Name: name, Namespace: namespace, ClusterID: clusterID}, nil
}

func projectRequest(name, clusterID string, pools []clusterclient.Nodepool) orgunitclient.ProjectCreationRequest {
	resources := make([]orgunitclient.ResourcesNullable, 0, len(pools))
	for _, pool := range pools {
		resources = append(resources, orgunitclient.ResourcesNullable{
			NodePool: *orgunitclient.NewNullableResourcesNodePool(&orgunitclient.ResourcesNodePool{
				Id:   pool.Id,
				Name: orgunitclient.PtrString(pool.Name),
			}),
			Gpu: &orgunitclient.ResourcesNullableGpu{
				Deserved: orgunitclient.PtrFloat64(0),
				Limit:    *orgunitclient.NewNullableFloat64(orgunitclient.PtrFloat64(-1)),
			},
		})
	}
	return orgunitclient.ProjectCreationRequest{
		Name:        name,
		ClusterId:   clusterID,
		Description: orgunitclient.PtrString(ownerMarker),
		Resources:   resources,
	}
}

func createProject(ctx context.Context, clients *client.Clients, request orgunitclient.ProjectCreationRequest) (string, error) {
	project, response, err := clients.OrgUnits.ProjectsAPI.CreateProject(ctx).ProjectCreationRequest(request).Execute()
	switch {
	case response != nil && response.StatusCode == http.StatusConflict:
		existing, err := managedProject(ctx, clients, request.ClusterId, request.Name)
		if err != nil {
			return "", err
		}
		return existing.Id, nil
	case err != nil:
		return "", fmt.Errorf("creating project %q: %w", request.Name, err)
	case response.StatusCode != http.StatusCreated:
		return "", fmt.Errorf("creating project %q: %s", request.Name, response.Status)
	}
	return project.Id, nil
}

func projectFilter(clusterID, name string) []string {
	return []string{"clusterId==" + clusterID, "name==" + name}
}

func managedProject(ctx context.Context, clients *client.Clients, clusterID, name string) (*orgunitclient.Project, error) {
	projects, _, err := clients.OrgUnits.ProjectsAPI.GetProjects(ctx).FilterBy(projectFilter(clusterID, name)).Execute()
	if err != nil {
		return nil, fmt.Errorf("looking up project %q: %w", name, err)
	}
	project, err := projectByName(projects.Projects, name)
	if err != nil {
		return nil, err
	}
	if err := requireManaged(project, clients.Identity); err != nil {
		return nil, err
	}
	return project, nil
}

func projectByName(projects []orgunitclient.Project, name string) (*orgunitclient.Project, error) {
	for i := range projects {
		if projects[i].Name == name {
			return &projects[i], nil
		}
	}
	return nil, fmt.Errorf("project %q: %w", name, errProjectNotFound)
}

func requireManaged(project *orgunitclient.Project, identity string) error {
	if identity == "" {
		return fmt.Errorf("cannot verify ownership of project %q: flowmark identity is unknown", project.Name)
	}
	if project.Description != ownerMarker || project.CreatedBy == nil || *project.CreatedBy != identity {
		return fmt.Errorf("project %q was not created by flowmark as %s; refusing to reuse or delete it", project.Name, identity)
	}
	switch project.Status.GetPhase() {
	case orgunitclient.PROJECTPHASE_DELETING, orgunitclient.PROJECTPHASE_DELETED:
		return fmt.Errorf("project %q is being deleted", project.Name)
	}
	return nil
}

func waitProjectReady(ctx context.Context, clients *client.Clients, id, name string, timeout time.Duration) (string, error) {
	var namespace, last string
	err := wait.PollUntilContextTimeout(ctx, pollInterval, timeout, true, func(ctx context.Context) (bool, error) {
		project, _, err := clients.OrgUnits.ProjectsAPI.GetProject(ctx, id).Execute()
		if err != nil {
			last = err.Error()
			return false, nil
		}
		phase := project.Status.GetPhase()
		switch phase {
		case orgunitclient.PROJECTPHASE_DELETING, orgunitclient.PROJECTPHASE_DELETED:
			return false, fmt.Errorf("project %q is being deleted", name)
		case orgunitclient.PROJECTPHASE_READY:
			if reported, ok := project.Status.GetNamespaceOk(); ok && reported != nil && *reported != "" {
				namespace = *reported
				return true, nil
			}
			last = "ready, namespace not reported yet"
			return false, nil
		}
		last = "phase " + string(phase)
		return false, nil
	})
	if err != nil {
		return "", fmt.Errorf("project %q not ready after %s (last: %s): %w", name, timeout, last, err)
	}
	return namespace, nil
}
