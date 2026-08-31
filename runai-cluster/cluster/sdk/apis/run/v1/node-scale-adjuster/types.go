/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// +kubebuilder:object:generate:=true
package node_scale_adjuster

import (
	v1 "k8s.io/api/core/v1"

	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/common"
)

// NodeScaleAdjuster defines configurations for the node-scale-adjuster
type NodeScaleAdjuster struct {
	// Enabled specifies if the node-scale-adjuster should be installed
	// +kubebuilder:validation:Optional
	Enabled *bool `json:"enabled,omitempty"`

	// Image is the configuration of the node scale adjuster image
	// +kubebuilder:validation:Optional
	Image *common.Image `json:"image,omitempty"`

	// Resources describes the resource requirements for the node scale adjuster pod
	// +kubebuilder:validation:Optional
	Resources *Resources `json:"resources,omitempty"`

	// ScalingPodImage is the image to use for the scaling pod
	// +kubebuilder:validation:Optional
	ScalingPodImage *common.Image `json:"scalingPodImage,omitempty"`

	// Args specifies the CLI arguments for node-scale-adjuster
	// +kubebuilder:validation:Optional
	Args *Args `json:"args,omitempty"`

	// Affinity specifies the pod's scheduling constraints for the node-scale-adjuster
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Schemaless
	// +kubebuilder:validation:Type=object
	// +kubebuilder:pruning:PreserveUnknownFields
	Affinity *v1.Affinity `json:"affinity,omitempty"`
}

// Resources specifies container resource requirements
type Resources v1.ResourceRequirements

// Args specifies the CLI arguments for node-scale-adjuster
type Args struct {
	// GPUMemoryToFractionRatio is the ratio of GPU memory to fraction conversion
	// +kubebuilder:validation:Optional
	GPUMemoryToFractionRatio *float64 `json:"gpuMemoryToFractionRatio,omitempty"`
}
