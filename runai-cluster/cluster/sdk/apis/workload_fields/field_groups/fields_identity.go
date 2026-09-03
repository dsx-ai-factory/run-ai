/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package field_groups

import (
	flds "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
)

// +k8s:deepcopy-gen=true
type IdentityFields struct {
	//
	//   Reference to another workload. When set, this workload inherits its values from the base workload.
	//   Base workload can either reside on the same namespace of this workload (referred to as
	//   "user" template) or can reside in the runai namespace (referred to as a "global" template)
	//
	BaseWorkload *string `json:"baseWorkload,omitempty"`

	//
	//   The intended usage of this workload. possible values are
	//   "Template": this workload is used as the base for other workloads.
	//   "Submit": this workload is used for submitting a job and/or other Kubernetes resources.
	// +kubebuilder:default:=Submit
	// +kubebuilder:validation:Enum:=Template;Submit
	Usage string `json:"usage,omitempty"`

	//   Display-only field describing the user who owns the workload. The data is not used
	//   for authentication or authorization purposes.
	UserName *flds.StringField `json:"username,omitempty"`

	//   The user ID ("Subject" in the jwt-token) of the authenticated user who owns the workload. The data might be used
	//   for authentication or authorization purposes.
	UserId *flds.StringField `json:"userId,omitempty"`
}

// +k8s:deepcopy-gen=true
type IdentityDeprecatedFields struct {
	//   The specific name of the created resource.
	//   Either name of namePrefix should be provided, but not both.
	//   This field is deprecated, the created resource will use the metadata.name
	// +kubebuilder:deprecated
	Name *flds.StringField `json:"name,omitempty"`

	//   A prefix used for assigning a name to the created resource.
	//   Either name of namePrefix should be provided, but not both.
	//   This field is deprecated, use metadata.generateName instead.
	// +kubebuilder:deprecated
	NamePrefix *flds.StringField `json:"namePrefix,omitempty"`
}

func IdentityRunaiRules() IdentityFields {
	return IdentityFields{}
}

func IdentityDeprecatedRunaiRules() IdentityDeprecatedFields {
	return IdentityDeprecatedFields{}
}
