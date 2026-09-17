/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package project

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api/filter"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/api/list"
	projectcreate "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/project/create"
	projectlist "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/project/list"
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client/generated"
	"k8s.io/utils/ptr"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/consts"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/project/client"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/transaction"
)

var (
	lister  projectlist.ProjectLister
	creator projectcreate.ProjectCreator

	ErrProjectNameMissing = errors.New("project name is missing")
)

func SetLister(projLister projectlist.ProjectLister) {
	lister = projLister
}

func SetCreator(projCreator projectcreate.ProjectCreator) {
	creator = projCreator
}

func getProjectListService(domain, token string) (*projectlist.ProjectService, error) {
	// Use mock if injected (for testing)
	if lister != nil {
		return projectlist.NewProjectService(lister), nil
	}

	// Create real client
	projectClient, err := client.NewProjectClient(domain, token)
	if err != nil {
		return nil, err
	}

	return projectlist.NewProjectService(projectClient), nil
}

func GetProjectList(txn *transaction.Txn, domain, token string, pageOptions *list.PageOptions, fb *filter.Builder) (*projectlist.Output, error) {
	projService, err := getProjectListService(domain, token)
	if err != nil {
		return nil, err
	}
	return projService.List(txn, pageOptions, fb)
}

// GetProjectId returns the project ID by the project name
func GetProjectId(txn *transaction.Txn, domain, token, clusterUuid, projectName string) (string, error) {
	if projectName == "" {
		return "", ErrProjectNameMissing
	}

	fb := filter.NewBuilder().WithClusterID(clusterUuid).AppendName(projectName)
	output, err := GetProjectList(txn, domain, token, list.NewSingleArtifactPageOption(), fb)
	if err != nil {
		return "", err
	}

	for _, project := range output.GetList() {
		if project.Name == projectName {
			return project.Id, nil
		}
	}

	return "", fmt.Errorf(consts.ErrorNoProjectFound, projectName)
}

// ParseProjectIdInt32 converts a project ID string to *int32 pointer
// Returns nil if projectId is empty, or error if the string is not a valid int32
func ParseProjectIdInt32(projectId string) (*int32, error) {
	if projectId == "" {
		return nil, nil
	}

	projectIdInt, err := strconv.ParseInt(projectId, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid project ID: %s", projectId)
	}
	return ptr.To(int32(projectIdInt)), nil
}

func getProjectCreatorService(domain, token string) (*projectcreate.ProjectService, error) {
	// Use mock if injected (for testing)
	if creator != nil {
		return projectcreate.NewProjectService(creator), nil
	}

	// Create real client
	projectClient, err := client.NewProjectClient(domain, token)
	if err != nil {
		return nil, err
	}

	return projectcreate.NewProjectService(projectClient), nil
}

func CreateProject(txn *transaction.Txn, domain, token string, request *openapi.ProjectCreationRequest) (*openapi.Project, error) {
	creatorService, err := getProjectCreatorService(domain, token)
	if err != nil {
		return nil, err
	}

	return creatorService.Project(txn, request)
}
