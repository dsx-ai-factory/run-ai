/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// +kubebuilder:object:generate:=true
package pod_grouper

import (
	kaicommon "github.com/kai-scheduler/KAI-scheduler/pkg/apis/kai/v1/common"
	v1 "k8s.io/api/core/v1"

	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/common"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/global"
)

// PodGrouper defines configurations for the pod-grouper
type PodGrouper struct {
	// Enabled specifies if the pod-grouper should be installed
	// +kubebuilder:validation:Optional
	Enabled *bool `json:"enabled,omitempty"`

	// Image is the configuration of the pod-grouper image
	// +kubebuilder:validation:Optional
	Image *common.Image `json:"image,omitempty"`

	// Args specifies the CLI arguments for the pod-grouper
	// +kubebuilder:validation:Optional
	Args *Args `json:"args,omitempty"`

	// Resources describes the resource requirements for the pod grouper pod
	// +kubebuilder:validation:Optional
	Resources *Resources `json:"resources,omitempty"`
	// ClientConfig specifies the configuration of k8s client
	// +kubebuilder:validation:Optional
	K8sClientConfig *kaicommon.K8sClientConfig `json:"k8sClientConfig,omitempty"`

	//MaxConcurrentReconciles specifies the number of max concurrent reconcile workers
	MaxConcurrentReconciles *int `json:"maxConcurrentReconciles,omitempty"`

	// Affinity specifies the pod's scheduling constraints for the pod-grouper
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Schemaless
	// +kubebuilder:validation:Type=object
	// +kubebuilder:pruning:PreserveUnknownFields
	Affinity *v1.Affinity `json:"affinity,omitempty"`

	// VPA specifies Vertical Pod Autoscaler configuration for the pod-grouper
	// +kubebuilder:validation:Optional
	VPA *global.VPASpec `json:"vpa,omitempty"`
}

// Resources specifies container resource requirements
type Resources v1.ResourceRequirements

// Args defines command line arguments for the pod-grouper
type Args struct {
	// GangScheduleKnative specifies whether to enable gang scheduling for Knative revisions. Default is true.
	// +kubebuilder:validation:Optional
	GangScheduleKnative *bool `json:"gangScheduleKnative,omitempty"`
}
