/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package field_groups

import (
	flds "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	"k8s.io/utils/ptr"
)

const (
	Preemptible    = "preemptible"
	NonPreemptible = "non-preemptible"
)

// +k8s:deepcopy-gen=true
type SchedulingFields struct {
	// Specifies the priority class name for the pod running the created workload.
	// Do not use together with 'Preemtible' or 'NonPreemptible' fields.
	// +optional
	PriorityClassName *flds.StringField `json:"priorityClassName,omitempty"`

	// Specifies whether the workload can be preempted by higher-priority workloads.
	// Valid values are preemptible and non-preemptible.
	// If explicitly set, this value takes precedence.
	// If not set, the system derives the preemptibility from the priorityClassName field, ensuring backward compatibility.
	// Each workload type has a default preemptibility.
	// To view the default preemptibility for each workload type, use the GET /workload-types endpoint.
	// +optional
	Preemptibility *flds.StringField `json:"preemptibility,omitempty"`
}

// RunaiSchedulingRules defines the validation rules for scheduling fields
func RunaiSchedulingRules() SchedulingFields {
	return SchedulingFields{
		Preemptibility: &flds.StringField{
			Rules: &flds.StringRules{
				Options: &[]flds.OptionRule{
					{
						Value: ptr.To(Preemptible),
						Name:  ptr.To(Preemptible),
					},
					{
						Value: ptr.To(NonPreemptible),
						Name:  ptr.To(NonPreemptible),
					},
				},
			},
		},
	}
}
