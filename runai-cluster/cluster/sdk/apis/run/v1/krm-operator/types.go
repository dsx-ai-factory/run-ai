/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// +kubebuilder:object:generate:=true
package krm_operator

import "k8s.io/utils/ptr"

// KRMOperator defines configuration for the KRM operator operand.
type KRMOperator struct {
	// Enabled specifies whether the KRMConfig singleton is reconciled.
	// +kubebuilder:validation:Optional
	Enabled *bool `json:"enabled,omitempty"`
}

// SetDefaultsWhereNeeded sets defaults for unset fields.
func (k *KRMOperator) SetDefaultsWhereNeeded() {
	if k.Enabled == nil {
		k.Enabled = ptr.To(true)
	}
}
