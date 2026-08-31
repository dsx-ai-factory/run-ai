/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// +kubebuilder:object:generate:=true
package runai_node_exporter

import (
	corev1 "k8s.io/api/core/v1"

	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/common"
)

// NodeExporter is retained only to preserve RunaiConfig data required by a rollback to 2.26.
// Deprecated: runai-node-exporter is not reconciled in KAI-based releases.
type NodeExporter struct {
	// +kubebuilder:validation:Optional
	Enabled *bool `json:"enabled,omitempty"`
	// +kubebuilder:validation:Optional
	Image *common.Image `json:"image,omitempty"`
	// +kubebuilder:validation:Optional
	Resources *Resources `json:"resources,omitempty"`
	// +kubebuilder:validation:Optional
	Service *Service `json:"service,omitempty"`
	// +kubebuilder:validation:Optional
	ServiceMonitor *ServiceMonitor `json:"serviceMonitor,omitempty"`
	// +kubebuilder:validation:Optional
	Args *Args `json:"args,omitempty"`
	// +kubebuilder:validation:Optional
	RunAIProcDirectoryHostPath string `json:"runaiProcDirectoryHostPath,omitempty"`
	// +kubebuilder:validation:Optional
	RunAIProcDirectoryMountPath string `json:"runaiProcDirectoryMountPath,omitempty"`
}

type Resources corev1.ResourceRequirements

type Service struct {
	// +kubebuilder:validation:Optional
	Type corev1.ServiceType `json:"type,omitempty"`
	// +kubebuilder:validation:Optional
	Port int `json:"port,omitempty"`
	// +kubebuilder:validation:Optional
	PortName string `json:"portName,omitempty"`
}

type ServiceMonitor struct {
	// +kubebuilder:validation:Optional
	Interval string `json:"interval,omitempty"`
	// +kubebuilder:validation:Optional
	Path string `json:"path,omitempty"`
}

type Args struct {
	// +kubebuilder:validation:Optional
	CollectInterval *string `json:"collectInterval,omitempty"`
	// +kubebuilder:validation:Optional
	DebugLog *bool `json:"debugLog,omitempty"`
}
