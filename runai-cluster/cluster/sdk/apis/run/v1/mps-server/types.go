/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// +kubebuilder:object:generate:=true
package mps_server

import (
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/common"
	"k8s.io/utils/ptr"
)

// MPSServer defines configuration for the mps-server
type MPSServer struct {
	// RunAIMPSDirectoryHostPath defines the host path for the mps server
	// +kubebuilder:validation:Optional
	RunAIMPSDirectoryHostPath string `json:"runAIMPSDirectoryHostPath,omitempty"`

	// Image specifies the mps server image
	// +kubebuilder:validation:Optional
	Image *common.Image `json:"image,omitempty"`
}

// SetDefaultsWhereNeeded sets default for unset fields
func (m *MPSServer) SetDefaultsWhereNeeded(globalRegistry string) {
	if m.RunAIMPSDirectoryHostPath == "" {
		m.RunAIMPSDirectoryHostPath = "/tmp/runai-mps/"
	}

	if m.Image == nil {
		m.Image = &common.Image{}
	}

	m.Image.SetDefaultsWhereNeeded(&common.Image{
		Registry: ptr.To(globalRegistry),
		Name:     ptr.To("ubi-minimal"),
	})
}
