/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package v1alpha2

import (
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

var (
	GroupVersion = schema.GroupVersion{Group: constants.GroupKubeflow, Version: constants.VersionKubeflow}

	SchemeBuilder = &scheme.Builder{GroupVersion: GroupVersion}

	AddToScheme = SchemeBuilder.AddToScheme
)

func AddKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(GroupVersion,
		&MPIJob{},
		&MPIJobList{},
	)
	scheme.AddKnownTypes(GroupVersion,
		&metav1.Status{},
	)

	metav1.AddToGroupVersion(scheme, GroupVersion)
	return nil
}
