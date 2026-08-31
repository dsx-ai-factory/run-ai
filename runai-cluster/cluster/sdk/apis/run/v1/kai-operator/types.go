/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// +kubebuilder:object:generate:=true
package kai_operator

import "k8s.io/utils/ptr"

// KAIOperator defines configuration for the KAI operator operand.
type KAIOperator struct {
	// Enabled specifies whether the KAI operator operand is reconciled.
	// +kubebuilder:validation:Optional
	Enabled *bool `json:"enabled,omitempty"`
}

// SetDefaultsWhereNeeded sets defaults for unset fields.
func (k *KAIOperator) SetDefaultsWhereNeeded() {
	if k.Enabled == nil {
		k.Enabled = ptr.To(true)
	}
}
