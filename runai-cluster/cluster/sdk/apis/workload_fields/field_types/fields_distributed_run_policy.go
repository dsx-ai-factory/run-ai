/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package field_types

// +k8s:deepcopy-gen=true
type RunPolicyField struct {
	Values *RunPolicyMembers `json:"values,omitempty"`
}

// +k8s:deepcopy-gen=true
type RunPolicyMembers struct {
	// Specifies the number of retries before marking a workload as failed.
	// Defaults to 6.
	BackoffLimit *IntField `json:"backoffLimit,omitempty"`

	// Pods kill policy after job completes. Supported values: All, Running, None.
	CleanPodPolicy *StringField `json:"cleanPodPolicy,omitempty"`
}

func (self *RunPolicyField) Deref() RunPolicyMembers {
	if self != nil && self.Values != nil {
		return *self.Values
	} else {
		return RunPolicyMembers{}
	}
}
