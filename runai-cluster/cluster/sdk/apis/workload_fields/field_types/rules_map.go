/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package field_types

//+kubebuilder:object:generate:=true

//	these rules apply to a value of type map (=non-primitive) as a whole
//	additionally there are rules which apply for specific items of the map
//
// +k8s:deepcopy-gen=true
type MapRules struct {
	//
	//   if the map as a whole is required
	//
	Required *bool `json:"required,omitempty"`

	//
	//   is it allowed for a workload to add items to this map
	//
	// +kubebuilder:default:=true
	CanAdd *bool `json:"canAdd,omitempty"`
}

//	these rules apply to specific item of a field of type map
//
// +k8s:deepcopy-gen=true
type MapItemRules struct {
	//
	//   is it allowed for a workload to remove this specific item from the map
	//
	// +kubebuilder:default:=true
	CanRemove *bool `json:"canRemove,omitempty"`

	//
	//   this one determines if workload can modify an item that the administrator policy
	// +kubebuilder:default:=true
	CanEdit *bool `json:"canEdit,omitempty"`
}
