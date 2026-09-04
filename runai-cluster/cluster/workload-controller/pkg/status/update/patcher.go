/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package update

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/common"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	"golang.org/x/exp/maps"
	logger "sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	minReadyReplicasField = "minRequiredReadyPods"
)

// Common field mapping functions that can be reused across different workload types
var (
	commonFieldMappings = map[string]func(*jobs.JobStatus) interface{}{
		"conditions":              func(s *jobs.JobStatus) interface{} { return s.Conditions },
		"completionTime":          func(s *jobs.JobStatus) interface{} { return s.WorkloadCommonStatus.CompletionTime },
		"succeededPods":           func(s *jobs.JobStatus) interface{} { return s.SucceededPods },
		"runningPods":             func(s *jobs.JobStatus) interface{} { return s.RunningPods },
		"readyPods":               func(s *jobs.JobStatus) interface{} { return s.ReadyPods },
		"pendingPods":             func(s *jobs.JobStatus) interface{} { return s.PendingPods },
		"failedPods":              func(s *jobs.JobStatus) interface{} { return s.FailedPods },
		"totalAllocatedGPU":       func(s *jobs.JobStatus) interface{} { return s.TotalAllocatedGPU },
		"totalAllocatedGPUMemory": func(s *jobs.JobStatus) interface{} { return s.TotalAllocatedGPUMemory },
		"startTime":               func(s *jobs.JobStatus) interface{} { return s.StartTime },
		"externalUrls":            func(s *jobs.JobStatus) interface{} { return s.WorkloadCommonStatus.ExternalUrls },
		"phase":                   func(s *jobs.JobStatus) interface{} { return s.WorkloadCommonStatus.Phase },
		"phaseMessage":            func(s *jobs.JobStatus) interface{} { return s.WorkloadCommonStatus.PhaseMessage },
		"observedGeneration":      func(s *jobs.JobStatus) interface{} { return s.ObservedGeneration },
		"firstAcknowledgedTime":   func(s *jobs.JobStatus) interface{} { return s.FirstAcknowledgedTime },
		"ownedResources":          func(s *jobs.JobStatus) interface{} { return s.OwnedResources },
		"lastStatusChangeTime":    func(s *jobs.JobStatus) interface{} { return s.LastStatusChangeTime },
		"services":                func(s *jobs.JobStatus) interface{} { return s.NetworkStatus.ServiceStatuses },
		"creationStatus":          func(s *jobs.JobStatus) interface{} { return s.CreationStatus },
		"issues":                  func(s *jobs.JobStatus) interface{} { return s.Issues },
	}
	getMinReadyPods = func(s *jobs.JobStatus) interface{} { return s.MinRequiredReadyPods }
)

// WorkloadPatcher handles serializing workload status fields to JSON
type WorkloadPatcher struct {
	// statusFields maps field names to functions that extract their values
	statusFields map[string]func(*jobs.JobStatus) interface{}
}

// NewWorkloadPatcher creates a patcher with the specified field mappings
func NewWorkloadPatcher(statusFields map[string]func(*jobs.JobStatus) interface{}) *WorkloadPatcher {
	return &WorkloadPatcher{
		statusFields: statusFields,
	}
}

// PatchBytes converts a workload's status into a JSON byte array
func (wp *WorkloadPatcher) PatchBytes(ctx context.Context, workload *jobs.JobWorkload) ([]byte, error) {
	if workload == nil {
		return nil, fmt.Errorf("workload cannot be nil")
	}

	log := logger.FromContext(ctx).WithValues(
		"name", workload.Name,
		"namespace", workload.Namespace,
	)

	// Set default phase if empty
	if workload.Status.WorkloadCommonStatus.Phase == "" {
		log.V(3).Info("Setting default phase", "phase", common.PhaseCreating.String())
		workload.Status.WorkloadCommonStatus.Phase = common.PhaseCreating.String()
	}

	// Build status map using configured fields
	status := make(map[string]interface{})
	for field, getValueFunc := range wp.statusFields {
		status[field] = getValueFunc(&workload.Status)
	}

	return json.Marshal(map[string]interface{}{"status": status})
}

// NewDistributedWorkloadPatcher creates a patcher with common fields
func NewDistributedWorkloadPatcher() *WorkloadPatcher {
	return NewWorkloadPatcher(commonFieldMappings)
}

// NewRunaiWorkloadPatcher creates a patcher with common fields plus MinRequiredReadyPods
func NewRunaiWorkloadPatcher() *WorkloadPatcher {
	fields := make(map[string]func(*jobs.JobStatus) interface{},
		len(commonFieldMappings)+1)

	maps.Copy(fields, commonFieldMappings)
	fields[minReadyReplicasField] = getMinReadyPods

	return NewWorkloadPatcher(fields)
}
