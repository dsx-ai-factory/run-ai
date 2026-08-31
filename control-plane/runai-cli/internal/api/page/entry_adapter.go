/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package page

import (
	"time"

	ousc "github.com/run-ai/runai/runai-clients/org-unit-service-client/generated"
	wsc "github.com/run-ai/runai/runai-clients/workloads-client/generated"
)

type WorkloadEntryAdapter struct {
	workload *wsc.Workload
}

func NewWorkloadEntryAdapter(workload *wsc.Workload) *WorkloadEntryAdapter {
	return &WorkloadEntryAdapter{workload: workload}
}

func (wea *WorkloadEntryAdapter) GetId() string {
	return wea.workload.Id
}
func (wea *WorkloadEntryAdapter) GetCreatedAt() time.Time {
	return wea.workload.CreatedAt
}

type ProjectEntryAdapter struct {
	project *ousc.Project
}

func NewProjectEntryAdapter(project *ousc.Project) *ProjectEntryAdapter {
	return &ProjectEntryAdapter{project: project}
}

func (wpa *ProjectEntryAdapter) GetId() string {
	return wpa.project.Id
}
func (wpa *ProjectEntryAdapter) GetCreatedAt() time.Time {
	return wpa.project.CreatedAt
}

type AIApplicationEntryAdapter struct {
	app *wsc.AIApplication
}

func NewAIApplicationEntryAdapter(app *wsc.AIApplication) *AIApplicationEntryAdapter {
	return &AIApplicationEntryAdapter{app: app}
}

func (a *AIApplicationEntryAdapter) GetId() string {
	return a.app.GetId()
}

func (a *AIApplicationEntryAdapter) GetCreatedAt() time.Time {
	if a.app.CreatedAt == nil {
		return time.Time{}
	}
	return *a.app.CreatedAt
}
