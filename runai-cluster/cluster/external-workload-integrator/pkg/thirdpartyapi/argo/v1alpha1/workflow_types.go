/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// Package v1alpha1 is a minimal local stand-in for the subset of
// github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1 that
// external-workload-integrator actually consumes. The handler reads the
// upstream Workflow object via a JSON round-trip, so unknown fields on the
// wire are silently dropped. Only fields read by the handler need to live
// here. If the handler ever needs additional fields, add them with their
// upstream JSON tags.
package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// the original argo workflow dir had many many vulnerabilities, so we have this instead!
// "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"

// Workflow is the partial mirror of argoproj.io/v1alpha1 Workflow used by the
// ArgoWorkflowTopOwnerHandler.
type Workflow struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WorkflowSpec   `json:"spec"`
	Status WorkflowStatus `json:"status,omitempty"`
}

// WorkflowList is required so Workflow can be registered in a runtime.Scheme
// alongside its list type.
// +kubebuilder:object:root=true
type WorkflowList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Workflow `json:"items"`
}

type WorkflowSpec struct {
	Entrypoint           string     `json:"entrypoint,omitempty"`
	PodPriorityClassName string     `json:"podPriorityClassName,omitempty"`
	Templates            []Template `json:"templates,omitempty"`
}

type Template struct {
	Name      string            `json:"name,omitempty"`
	Container *corev1.Container `json:"container,omitempty"`
}

type WorkflowStatus struct {
	Phase      WorkflowPhase `json:"phase,omitempty"`
	Message    string        `json:"message,omitempty"`
	Conditions Conditions    `json:"conditions,omitempty"`
}

type Conditions []Condition

type ConditionType string

type Condition struct {
	Type    ConditionType          `json:"type,omitempty"`
	Status  metav1.ConditionStatus `json:"status,omitempty"`
	Message string                 `json:"message,omitempty"`
}

// DeepCopyObject lets controller-runtime's cache and client clone the object
// when handing it to consumers; without it, *Workflow does not satisfy
// runtime.Object.
func (in *Workflow) DeepCopyObject() runtime.Object {
	if in == nil {
		return nil
	}
	out := new(Workflow)
	in.DeepCopyInto(out)
	return out
}

func (in *WorkflowList) DeepCopyObject() runtime.Object {
	if in == nil {
		return nil
	}
	out := new(WorkflowList)
	in.DeepCopyInto(out)
	return out
}
