/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package field_types

import "sort"

// +k8s:deepcopy-gen=true
type TolerationsField struct {
	Rules     *MapRules                           `json:"rules,omitempty"`
	ItemRules *TolerationsItemRules               `json:"itemRules,omitempty"`
	Value     *map[string]TolerationsItemAndRules `json:"items,omitempty"`
}

// +k8s:deepcopy-gen=true
type TolerationsItemAndRules struct {
	Rules   *MapItemRules         `json:"rules,omitempty"`
	Value   *TolerationsItemValue `json:"value,omitempty"`
	Deleted *bool                 `json:"deleted,omitempty"`
}

// +k8s:deepcopy-gen=true
type TolerationsItemRules struct {
	Operator *StringRules `json:"operator,omitempty"`
	Key      *StringRules `json:"tolerationKey,omitempty"`
	Value    *StringRules `json:"tolerationValue,omitempty"`
	Effect   *StringRules `json:"effect,omitempty"`
	Seconds  *IntRules    `json:"seconds,omitempty"`
}

// +k8s:deepcopy-gen=true
type TolerationsItemValue struct {
	// matching operator between the toleration rule and the taint of the node. possible operators:
	// equal: match the taint value against a given toleration value (default)
	// exist: check existance of the toleration key, regardless of the taint value
	// +kubebuilder:default:=Equal
	// +kubebuilder:validation:Enum:=Equal;Exists
	Operator *string `json:"operator,omitempty"`
	// the key to match with the taint key of a candidate node.
	// +kubebuilder:validation:Required
	Key *string `json:"tolerationKey"`
	// value to match between the toleration rule and the taint, for equal operator
	Value *string `json:"tolerationValue,omitempty"`
	// the effect of no matching between the toleration rules to the node taints.
	// NoSchedule: the system will not be able to schedule the workload to a non matching node
	// PreferNoSchedule: the system will attempt not to schedule the workload to a non matching node.
	// NoExecute: the workload will be evicted from any non-matching node (default)
	// +kubebuilder:validation:Enum:=NoSchedule;NoExecute;PreferNoSchedule
	Effect *string `json:"effect,omitempty"`
	// a toleration with NoExecute effect can specify an optional tolerationSeconds field that dictates how long
	// a pod will stay bound to a non-matching node, after the taint is added.
	Seconds *int `json:"seconds,omitempty"`
}

func (self *TolerationsField) Deref() map[string]TolerationsItemAndRules {
	if self != nil && self.Value != nil {
		return *self.Value
	} else {
		return map[string]TolerationsItemAndRules{}
	}
}

func NewTolerationsField(tolerationsMap map[string]TolerationsItemAndRules) *TolerationsField {
	if len(tolerationsMap) == 0 {
		return nil
	}
	return &TolerationsField{
		Value: &tolerationsMap,
	}
}

func SortedTolerationMapKeys(unsorted map[string]TolerationsItemAndRules) []string {
	result := make([]string, 0, len(unsorted))
	for k := range unsorted {
		result = append(result, k)
	}
	sort.Strings(result)
	return result
}
