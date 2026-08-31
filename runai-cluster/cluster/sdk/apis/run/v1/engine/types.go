/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// +kubebuilder:object:generate:=true
package engine

// Engine is retained only to preserve RunaiConfig data required by a rollback to 2.26.
// Deprecated: Engine is not reconciled in KAI-based releases.
type Engine struct {
	// Enabled specifies whether Engine was enabled.
	// +kubebuilder:validation:Optional
	Enabled *bool `json:"enabled,omitempty"`
}
