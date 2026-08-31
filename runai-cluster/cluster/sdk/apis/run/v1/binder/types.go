/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// +kubebuilder:object:generate:=true
package binder

import (
	v1 "k8s.io/api/core/v1"

	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/common"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/global"
)

// Binder defines configurations for the binder
type Binder struct {
	// Enabled specifies if the binder should be installed
	// +kubebuilder:validation:Optional
	Enabled *bool `json:"enabled,omitempty"`

	// Image is the configuration of the binder image
	// +kubebuilder:validation:Optional
	Image *common.Image `json:"image,omitempty"`

	// Args specifies the CLI arguments for the binder
	// +kubebuilder:validation:Optional
	Args *Args `json:"args,omitempty"`

	// Service describes the service for the binder
	// +kubebuilder:validation:Optional
	Service *Service `json:"service,omitempty"`

	// Resources describes the resource requirements for the pod grouper pod
	// +kubebuilder:validation:Optional
	Resources *Resources `json:"resources,omitempty"`

	// ResourceReservation controls configuration for the resource reservation functionality
	// +kubebuilder:validation:Optional
	ResourceReservation *ResourceReservation `json:"resourceReservation,omitempty"`

	// Affinity specifies the pod's scheduling constraints for the binder
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Schemaless
	// +kubebuilder:validation:Type=object
	// +kubebuilder:pruning:PreserveUnknownFields
	Affinity *v1.Affinity `json:"affinity,omitempty"`

	// VPA specifies Vertical Pod Autoscaler configuration for the binder
	// +kubebuilder:validation:Optional
	VPA *global.VPASpec `json:"vpa,omitempty"`
}

// Resources specifies container resource requirements
type Resources v1.ResourceRequirements

// Args defines command line arguments for the binder
type Args struct {
	// Verbosity specifies the logging verbosity of the pod grouper
	// use 0 for info, -1 for debug
	// ToDo: verify
	// +kubebuilder:validation:Optional
	Verbosity *int `json:"verbosity,omitempty"`

	// Deprecated: retained as inert data for rollback to 2.26.
	// +kubebuilder:validation:Optional
	RunAIFractionalMinAllocationEnforcementBytes *int `json:"runaiFractionalMinAllocationEnforcementBytes,omitempty"`

	// Deprecated: retained as inert data for rollback to 2.26.
	// +kubebuilder:validation:Optional
	RunAIMpsDirectoryHostPath *string `json:"runaiMpsDirectoryHostPath,omitempty"`

	// Deprecated: retained as inert data for rollback to 2.26.
	// +kubebuilder:validation:Optional
	NodeLevelScheduler *bool `json:"nodeLevelScheduler,omitempty"`

	// Deprecated: retained as inert data for rollback to 2.26.
	// +kubebuilder:validation:Optional
	InjectRunaiResource *bool `json:"injectRunaiResource,omitempty"`

	// VolumeBindingTimeoutSeconds specifies the timeout for volume binding in seconds
	// +kubebuilder:validation:Optional
	VolumeBindingTimeoutSeconds *int `json:"volumeBindingTimeoutSeconds,omitempty"`

	// MaxConcurrentReconciles specifies the maximum number of concurrent reconcile workers in the binder
	// +kubebuilder:validation:Optional
	MaxConcurrentReconciles *int `json:"maxConcurrentReconciles,omitempty"`

	// QPS specifies the maximum queries per second set in the k8s client
	// +kubebuilder:validation:Optional
	QPS *int `json:"qps,omitempty"`

	// Burst specifies the maximum burst queries set in the k8s client over the QPS limit
	// +kubebuilder:validation:Optional
	Burst *int `json:"burst,omitempty"`
}

// Service defines configuration for the binder service
type Service struct {
	// Webhook specifies the webhook service spec
	// +kubebuilder:validation:Optional
	Webhook *Webhook `json:"webhook,omitempty"`
}

// Webhook defines configuration for the binder webhook
type Webhook struct {
	// Port specifies the webhook service port
	// +kubebuilder:validation:Optional
	Port *int `json:"port,omitempty"`

	// TargetPort specifies the webhook service container port
	// +kubebuilder:validation:Optional
	TargetPort *int `json:"targetPort,omitempty"`

	// ProbePort specifies the health and readiness probe port
	ProbePort *int `json:"probePort,omitempty"`
}

type ResourceReservation struct {
	// Image specifies the configuration of the resource-reservation image
	// +kubebuilder:validation:Optional
	Image *common.Image `json:"image,omitempty"`
}
