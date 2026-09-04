/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package field_types

// +k8s:deepcopy-gen=true
type NfsField struct {
	Rules     *MapRules                   `json:"rules,omitempty"`
	ItemRules *NfsItemRules               `json:"itemRules,omitempty"`
	Value     *map[string]NfsItemAndRules `json:"items,omitempty"`
}

// +k8s:deepcopy-gen=true
type NfsItemAndRules struct {
	Rules   *MapItemRules `json:"rules,omitempty"`
	Value   *NfsItemValue `json:"value,omitempty"`
	Deleted *bool         `json:"deleted,omitempty"`
}

// +k8s:deepcopy-gen=true
type NfsItemRules struct {
	SourcePath *StringRules `json:"sourcePath,omitempty"`
	TargetPath *StringRules `json:"targetPath,omitempty"`
	ReadOnly   *BoolRules   `json:"readOnly,omitempty"`
	NfsServer  *StringRules `json:"nfsServer,omitempty"`
}

// +k8s:deepcopy-gen=true
type NfsItemValue struct {
	SourcePath *string `json:"sourcePath,omitempty"`
	// +kubebuilder:validation:Required
	TargetPath *string `json:"targetPath"`

	ReadOnly *bool `json:"readOnly,omitempty"`

	// +kubebuilder:validation:Required
	NfsServer *string `json:"nfsServer"`
}

func (self *NfsField) Deref() map[string]NfsItemAndRules {
	if self != nil && self.Value != nil {
		return *self.Value
	} else {
		return map[string]NfsItemAndRules{}
	}
}

func NewNfsField(nfsVolumesMap map[string]NfsItemAndRules) *NfsField {
	if len(nfsVolumesMap) == 0 {
		return nil
	}
	return &NfsField{
		Value: &nfsVolumesMap,
	}
}
