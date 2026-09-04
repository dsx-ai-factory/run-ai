/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// +kubebuilder:object:generate:=true
package knative

import (
	"k8s.io/utils/ptr"
)

// Knative defines configurations for knative serving
type Knative struct {
	// CreateServiceMonitors specifies if service monitors should be created
	// for Knative components in the runai namespace
	// +kubebuilder:validation:Optional
	CreateServiceMonitors *bool `json:"createServiceMonitors,omitempty"`
}

// SetDefaultsWhereNeeded sets default for unset fields
func (k *Knative) SetDefaultsWhereNeeded() {
	if k.CreateServiceMonitors == nil {
		k.CreateServiceMonitors = ptr.To[bool](true)
	}
}
