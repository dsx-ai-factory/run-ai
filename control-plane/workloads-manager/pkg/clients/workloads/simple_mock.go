/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workloads

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/auth"
	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
)

type SimpleMock struct {
	WorkloadTypes []workloadsApi.WorkloadTypeConfig

	ByWorkloadName      map[string][]*workloadsApi.Workload
	SyncIssuesByCluster map[string]*workloadsApi.ClusterSyncReport
	WorkloadActions     []workloadsApi.WorkloadActionRecordInput

	// cluster id to valid priorities and categories
	PriorityClassesAndCategories map[string]*workloadsApi.GetAvailablePrioritiesAndCategories200Response
}

func (sm *SimpleMock) CreateWorkloadActions(_ context.Context, request workloadsApi.CreateWorkloadActionsRequest) *workloadsApi.Error {
	sm.WorkloadActions = append(sm.WorkloadActions, request.Actions...)
	return nil
}

func NewSimpleMock() *SimpleMock {
	return &SimpleMock{
		WorkloadTypes:                make([]workloadsApi.WorkloadTypeConfig, 0),
		ByWorkloadName:               make(map[string][]*workloadsApi.Workload),
		SyncIssuesByCluster:          make(map[string]*workloadsApi.ClusterSyncReport),
		PriorityClassesAndCategories: make(map[string]*workloadsApi.GetAvailablePrioritiesAndCategories200Response),
	}
}

// CountWorkloads mock is based on the events that were sent to workloads service, which we
// collect in ByWorkloadName. Supported filters are only name and projectId
func (smc *SimpleMock) CountWorkloads(_ context.Context, filters []string, search string, includeDeleted bool, precise bool) (int64, *workloadsApi.Error) {
	projectId := ""
	name := ""

	// Parse the filters
	re := regexp.MustCompile(`^(name|projectId)==(.+)$`)
	for _, f := range filters {
		m := re.FindStringSubmatch(f)
		if len(m) == 3 {
			key, value := m[1], m[2]
			switch key {
			case "name":
				name = value
			case "projectId":
				projectId = value
			}
		}
	}

	// no name is given then return total number of names. it's not accurate but should be ok
	if name == "" {
		return int64(len(smc.ByWorkloadName)), nil
	}

	workloadsOfThisName, found := smc.ByWorkloadName[name]

	// no workload with that name
	if !found {
		return 0, nil
	}

	// if no project filter is given, return total number of workloads with that name
	if projectId == "" {
		return int64(len(workloadsOfThisName)), nil
	}

	// check if there is a workload with that name on the requested project
	for _, workload := range workloadsOfThisName {
		if projectId == workload.ProjectId {
			return int64(1), nil
		}
	}

	return 0, nil
}

// GetWorkload returns a single workload previously registered via CreateWorkload, matched by id.
// It returns a 404 error when no such workload exists, mirroring the real workloads service.
func (smc *SimpleMock) GetWorkload(_ context.Context, workloadId string) (*workloadsApi.WorkloadDetailed, *workloadsApi.Error) {
	for _, list := range smc.ByWorkloadName {
		for _, wl := range list {
			if wl.Id == workloadId {
				return &workloadsApi.WorkloadDetailed{
					Id:    wl.Id,
					Name:  wl.Name,
					Phase: wl.Phase,
				}, nil
			}
		}
	}
	return nil, workloadsApi.NewError(http.StatusNotFound, fmt.Sprintf("workload %s not found", workloadId))
}

// SetWorkloadPhase overrides the actual phase of a previously registered workload, identified
// by its id. It is used by tests to simulate cluster-driven phase transitions.
func (smc *SimpleMock) SetWorkloadPhase(workloadId string, phase workloadsApi.Phase) *SimpleMock {
	for _, list := range smc.ByWorkloadName {
		for _, wl := range list {
			if wl.Id == workloadId {
				wl.Phase = phase
			}
		}
	}
	return smc
}

func (smc *SimpleMock) ListWorkloadTypes(_ context.Context, filters []string) (*workloadsApi.ListWorkloadTypes200Response, *workloadsApi.Error) {
	filteredTypes := smc.WorkloadTypes

	// Parse filters and apply them
	if len(filters) > 0 {
		nameFilter := ""
		groupFilter := ""

		for _, filter := range filters {
			// Parse filter format: "field==value"
			if strings.HasPrefix(filter, "name==") {
				nameFilter = strings.TrimSpace(strings.TrimPrefix(filter, "name=="))
			} else if strings.HasPrefix(filter, "group==") {
				groupFilter = strings.TrimSpace(strings.TrimPrefix(filter, "group=="))
			}
		}

		// Apply filters
		var filtered []workloadsApi.WorkloadTypeConfig
		for _, workloadType := range smc.WorkloadTypes {
			matches := true

			// Filter by name (kind)
			if nameFilter != "" && workloadType.Name != nameFilter {
				matches = false
			}

			// Filter by group
			if groupFilter != "" && workloadType.Group != groupFilter {
				matches = false
			}

			if matches {
				filtered = append(filtered, workloadType)
			}
		}

		filteredTypes = filtered
	}

	return &workloadsApi.ListWorkloadTypes200Response{
		Types: filteredTypes,
	}, nil
}

func (sm *SimpleMock) CreateWorkload(ctx context.Context, workload workloadsApi.WorkloadCreationFields) (*workloadsApi.Workload, *workloadsApi.Error) {
	currentTime := time.Now()

	notifByName := sm.ByWorkloadName[workload.Name]
	if notifByName == nil {
		notifByName = make([]*workloadsApi.Workload, 0)
	}

	// simulate the result, it does not really need to be accurate, just meaningful values
	result := workloadsApi.Workload{
		TenantId:                   auth.TenantIdOf(ctx),
		PhaseUpdatedAt:             currentTime,
		K8sPhaseUpdatedAt:          currentTime,
		CreatedAt:                  currentTime,
		UpdatedAt:                  currentTime,
		Source:                     workloadsApi.SOURCE_CONTROL_PLANE,
		Type:                       workload.Type,
		Name:                       workload.Name,
		ExternalWorkloadName:       convutils.Ref(workload.Name),
		Id:                         workload.Id,
		PriorityClassName:          "default",
		SubmittedBy:                workload.SubmittedBy,
		ClusterId:                  workload.ClusterId,
		ProjectName:                convutils.Deref(workload.ProjectName, "default"),
		ProjectId:                  workload.ProjectId,
		DepartmentName:             convutils.Deref(workload.DepartmentName, "default"),
		DepartmentId:               convutils.Deref(workload.DepartmentId, "4500000"),
		Namespace:                  fmt.Sprintf("runai-%s", convutils.Deref(workload.ProjectName, "")),
		WorkloadRequestedResources: workload.RequestedResources,
		ActionsSupport: &workloadsApi.ActionsSupport{
			Delete:  convutils.Ref(true),
			Suspend: convutils.Ref(false), // for now generic workloads do not support suspend
		},
		Phase:                workloadsApi.PHASE_CREATING,
		PhaseMessage:         convutils.Ref("Creating"),
		K8sPhase:             "Pending",
		RequestedPods:        workload.RequestedPods,
		RequestedNodePools:   workload.RequestedNodePools,
		CurrentNodePools:     workload.RequestedNodePools,
		Images:               workload.Images,
		Datasources:          workload.Datasources,
		Environments:         workload.Environments,
		ExternalConnections:  workload.ExternalConnections,
		DistributedFramework: workload.DistributedFramework,
		FullySaasWorkload:    workload.FullySaasWorkload,
		Category:             convutils.Deref(workload.Category, "default"),
	}

	if workload.RequestedPods != nil {
		result.RunningPods = convutils.Deref(workload.RequestedPods.Number.Get(), 1)
	}

	notifByName = append(notifByName, &result)
	sm.ByWorkloadName[workload.Name] = notifByName

	return &result, nil
}

func (sm *SimpleMock) ProcessSyncReport(_ context.Context, clusterId string, report workloadsApi.ClusterSyncReport) *workloadsApi.Error {
	sm.SyncIssuesByCluster[clusterId] = &report

	return nil
}

func (sm *SimpleMock) PatchWorkload(_ context.Context, _ string, _ workloadsApi.WorkloadPatchFields) *workloadsApi.Error {
	return nil
}

func (smc *SimpleMock) WithWorkloadType(workloadType workloadsApi.WorkloadTypeConfig) *SimpleMock {
	smc.WorkloadTypes = append(smc.WorkloadTypes, workloadType)
	return smc
}

func (smc *SimpleMock) WithPrioritiesAndCategories(clusterId string, response workloadsApi.GetAvailablePrioritiesAndCategories200Response) *SimpleMock {
	smc.PriorityClassesAndCategories[clusterId] = &response
	return smc
}

func (smc *SimpleMock) GetAvailablePrioritiesAndCategories(_ context.Context, clusterId *string, _ *string, _ *string, _ *[]string) (*workloadsApi.GetAvailablePrioritiesAndCategories200Response, *workloadsApi.Error) {
	if clusterId == nil {
		return nil, workloadsApi.NewError(http.StatusBadRequest, "clusterId not provided")
	}

	result := smc.PriorityClassesAndCategories[*clusterId]
	if result == nil {
		return nil, workloadsApi.NewError(http.StatusNotFound, "cluster not found in priorities and categories map")
	}

	return result, nil
}
