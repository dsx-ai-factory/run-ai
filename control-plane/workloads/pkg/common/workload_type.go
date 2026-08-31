/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package common

import (
	"fmt"
	"strings"

	"github.com/run-ai/runai/control-plane/workloads/internal/generated/pkg/api"
)

// WorkloadType represents the type of workload
type WorkloadType string

const (
	WorkloadTypeWorkspace            = WorkloadType(api.WORKLOADTYPE_WORKSPACE)
	WorkloadTypeTraining             = WorkloadType(api.WORKLOADTYPE_TRAINING)
	WorkloadTypeDistributed          = WorkloadType(api.WORKLOADTYPE_DISTRIBUTED)
	WorkloadTypeInference            = WorkloadType(api.WORKLOADTYPE_INFERENCE)
	WorkloadTypeDistributedInference = WorkloadType(api.WORKLOADTYPE_DISTRIBUTED_INFERENCE)

	interactiveWorkloadTypeName = "interactive"
	workloadKindsSuffix         = "Workload"
)

// NewWorkloadType creates a new WorkloadType from a string
func NewWorkloadType(s string) WorkloadType {
	if strings.ToLower(s) == interactiveWorkloadTypeName {
		return WorkloadTypeWorkspace
	}
	return WorkloadType(s)
}

func NewWorkloadTypeFromKind(kind string) (WorkloadType, error) {
	if !strings.HasSuffix(kind, workloadKindsSuffix) {
		return "", fmt.Errorf("invalid workload kind '%s': must end with '%s'", kind, workloadKindsSuffix)
	}

	s := strings.TrimSuffix(kind, workloadKindsSuffix)
	return NewWorkloadType(s), nil
}

func (w WorkloadType) IsNative() bool {
	return api.WorkloadType(w.String()).IsValid()
}

func (w WorkloadType) String() string {
	return string(w)
}

func (w WorkloadType) ToKind() string {
	if w == WorkloadTypeWorkspace {
		return "InteractiveWorkload"
	}
	return w.String() + workloadKindsSuffix
}
