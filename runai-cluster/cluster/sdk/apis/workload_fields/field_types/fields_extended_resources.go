/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package field_types

import "sort"

// +k8s:deepcopy-gen=true
type ExtendedResourcesField struct {
	Rules     *MapRules                                 `json:"rules,omitempty"`
	ItemRules *ExtendedResourcesItemRules               `json:"itemRules,omitempty"`
	Value     *map[string]ExtendedResourcesItemAndRules `json:"items,omitempty"`
}

// +k8s:deepcopy-gen=true
type ExtendedResourcesItemAndRules struct {
	Rules   *MapItemRules               `json:"rules,omitempty"`
	Value   *ExtendedResourcesItemValue `json:"value,omitempty"`
	Deleted *bool                       `json:"deleted,omitempty"`
}

// +k8s:deepcopy-gen=true
type ExtendedResourcesItemRules struct {
	Name     *StringRules `json:"name"`
	Quantity *StringRules `json:"quantity"`
}

// +k8s:deepcopy-gen=true
type ExtendedResourcesItemValue struct {
	//+kubebuilder:validation:Required
	Name *string `json:"name"`
	// +kubebuilder:validation:Required
	Quantity *string `json:"quantity"`
}

func (self *ExtendedResourcesField) Deref() map[string]ExtendedResourcesItemAndRules {
	if self != nil && self.Value != nil {
		return *self.Value
	} else {
		return map[string]ExtendedResourcesItemAndRules{}
	}
}

func NewExtendedResourcesField(extendedResourcesMap map[string]ExtendedResourcesItemAndRules) *ExtendedResourcesField {
	if len(extendedResourcesMap) == 0 {
		return nil
	}
	return &ExtendedResourcesField{
		Value: &extendedResourcesMap,
	}
}

func SortedExtendedResourcesMapKeys(unsorted map[string]ExtendedResourcesItemAndRules) []string {
	result := make([]string, 0, len(unsorted))
	for k := range unsorted {
		result = append(result, k)
	}
	sort.Strings(result)
	return result
}
