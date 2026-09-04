/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package field_groups

import fields "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"

// StateFields defines the workload state related fields
// +k8s:deepcopy-gen=true
type StateFields struct {
	// +kubebuilder:default:={value: true}
	// Specifies whether the workload should be active or suspended.
	Active *fields.BoolField `json:"active,omitempty"`
}
