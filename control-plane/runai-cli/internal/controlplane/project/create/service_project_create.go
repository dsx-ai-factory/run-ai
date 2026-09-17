/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package create

import (
	"context"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/transaction"
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client/generated"
)

const (
	TimerCreate = "project_create_timer"
)

type ProjectCreator interface {
	CreateProject(ctx context.Context, creationRequest *openapi.ProjectCreationRequest) (*openapi.Project, error)
}

type ProjectService struct {
	projectClient ProjectCreator
}

func NewProjectService(projectCreator ProjectCreator) *ProjectService {
	return &ProjectService{projectClient: projectCreator}
}

func (ps *ProjectService) Project(txn *transaction.Txn, request *openapi.ProjectCreationRequest) (*openapi.Project, error) {
	ptnId := txn.AddTimer(TimerCreate)
	defer txn.StopTimer(ptnId)
	return ps.projectClient.CreateProject(txn.GetCtx(), request)
}
