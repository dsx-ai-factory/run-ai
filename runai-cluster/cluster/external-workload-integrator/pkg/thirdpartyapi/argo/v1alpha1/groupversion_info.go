/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// Package v1alpha1 contains a minimal local mirror of the argoproj.io/v1alpha1
// Workflow type tree, kept just large enough for ArgoWorkflowTopOwnerHandler
// to parse the fields it consumes. See workflow_types.go for the rationale.
// +kubebuilder:object:generate=true
// +groupName=argoproj.io
package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

var (
	// GroupVersion is group version used to register these objects.
	GroupVersion = schema.GroupVersion{Group: "argoproj.io", Version: "v1alpha1"}

	SchemeGroupVersion = GroupVersion

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme.
	SchemeBuilder = &scheme.Builder{GroupVersion: GroupVersion}

	// AddToScheme adds the types in this group-version to the given scheme.
	AddToScheme = SchemeBuilder.AddToScheme

	// Resource takes an unqualified resource and returns a Group qualified GroupResource.
	Resource = func(resource string) schema.GroupResource {
		return SchemeGroupVersion.WithResource(resource).GroupResource()
	}
)

func init() {
	SchemeBuilder.Register(&Workflow{}, &WorkflowList{})
}
