/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// +kubebuilder:object:generate:=true
package device_plugin

import (
	corev1 "k8s.io/api/core/v1"

	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/common"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/global"
)

// DevicePlugin is retained only to preserve RunaiConfig data required by a rollback to 2.26.
// Deprecated: the RunAI device plugin is not reconciled in KAI-based releases.
type DevicePlugin struct {
	// +kubebuilder:validation:Optional
	Enabled *bool `json:"enabled,omitempty"`
	// +kubebuilder:validation:Optional
	Image *common.Image `json:"image,omitempty"`
	// +kubebuilder:validation:Optional
	Resources *Resources `json:"resources,omitempty"`
	// +kubebuilder:validation:Optional
	Args *Args `json:"args,omitempty"`
	// +kubebuilder:validation:Optional
	K8sClientConfig *K8sClientConfig `json:"k8sClientConfig,omitempty"`
	// +kubebuilder:validation:Optional
	VPA *global.VPASpec `json:"vpa,omitempty"`
}

type K8sClientConfig struct {
	// +kubebuilder:validation:Optional
	Burst *int `json:"burst,omitempty"`
	// +kubebuilder:validation:Optional
	QPS *int `json:"qps,omitempty"`
}

type Resources corev1.ResourceRequirements

type Args struct {
	// +kubebuilder:validation:Optional
	LogLevel *int `json:"logLevel,omitempty"`
	// +kubebuilder:validation:Optional
	DriverRoot *string `json:"driverRoot,omitempty"`
	// +kubebuilder:validation:Optional
	Slices *int `json:"slices,omitempty"`
	// +kubebuilder:validation:Optional
	DevPluginPath *string `json:"devPluginPath,omitempty"`
}
