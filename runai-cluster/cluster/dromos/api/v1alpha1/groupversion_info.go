/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	// GroupVersion is the group version used to register these objects.
	GroupVersion = schema.GroupVersion{Group: "run.ai", Version: "v1alpha1"}

	// SchemeBuilder registers the WorkloadRunner types with a runtime.Scheme.
	// Built on apimachinery's runtime.NewSchemeBuilder - no controller-runtime dependency.
	SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)

	// AddToScheme adds the WorkloadRunner types to the given scheme.
	AddToScheme = SchemeBuilder.AddToScheme
)

// SchemeGroupVersion is kept for legacy callers that expect this name.
var SchemeGroupVersion = GroupVersion

func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(GroupVersion,
		&WorkloadRunner{},
		&WorkloadRunnerList{},
	)
	metav1.AddToGroupVersion(scheme, GroupVersion)
	return nil
}
