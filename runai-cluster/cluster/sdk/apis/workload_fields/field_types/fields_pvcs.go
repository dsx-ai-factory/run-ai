/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package field_types

// +k8s:deepcopy-gen=true
type PvcsField struct {
	Rules     *MapRules                    `json:"rules,omitempty"`
	ItemRules *PvcsItemRules               `json:"itemRules,omitempty"`
	Value     *map[string]PvcsItemAndRules `json:"items,omitempty"`
}

// +k8s:deepcopy-gen=true
type PvcsItemAndRules struct {
	Rules   *MapItemRules  `json:"rules,omitempty"`
	Value   *PvcsItemValue `json:"value,omitempty"`
	Deleted *bool          `json:"deleted,omitempty"`
}

// +k8s:deepcopy-gen=true
type PvcsItemRules struct {
	ClaimName     *StringRules `json:"claimName,omitempty"`
	StorageClass  *StringRules `json:"storageClass,omitempty"`
	Size          *StringRules `json:"size,omitempty"`
	Path          *StringRules `json:"path,omitempty"`
	ReadOnly      *BoolRules   `json:"readOnly,omitempty"`
	ExistingPvc   *BoolRules   `json:"existingPvc,omitempty"`
	ReadWriteOnce *BoolRules   `json:"readWriteOnce,omitempty"`
	ReadOnlyMany  *BoolRules   `json:"readOnlyMany,omitempty"`
	ReadWriteMany *BoolRules   `json:"readWriteMany,omitempty"`
}

// +k8s:deepcopy-gen=true
type PvcsItemValue struct {
	// An optional given name to the Persistent Volume Claim, allowing multiple workloads to reference it.
	// See bindExistingClaim field for further details.
	ClaimName *string `json:"claimName,omitempty"`
	// Storage class name to associate with the PVC. This parameter may be omitted if there is a single storage class in the system, or you
	// are using the default storage class. Further details: https://kubernetes.io/docs/concepts/storage/storage-classes
	StorageClass *string `json:"storageClass,omitempty"`
	// Requested size for the PVC
	Size *string `json:"size,omitempty"`
	// Local path within the workspace to which the PVC bucket will be mapped.
	//+kubebuilder:validation:Required
	Path *string `json:"path"`
	// Permits read only from the PVC, preventing additions or modifications to its content.
	ReadOnly *bool `json:"readOnly,omitempty"`
	// When set to true, the workload will attempt to connect to an existing PVC with name indicated
	// by the claimName field. Creation of the workload will fail if the PVC does not exist.
	// When set to false, the PVC will be created if it does not exist, thus sizing and optional storage class parameters
	// for creating the PVC are mandatory.
	ExistingPvc *bool `json:"existingPvc,omitempty"`
	// Requesting claim that can be mounted in read/write mode to exactly 1 host
	// If none of the mode is specified, the default is readWriteOnce.
	ReadWriteOnce *bool `json:"readWriteOnce,omitempty"`
	// Requesting claim that can be mounted in read-only mode to many hosts
	ReadOnlyMany *bool `json:"readOnlyMany,omitempty"`
	// Requesting claim that can be mounted in read/write mode to many hosts
	ReadWriteMany *bool `json:"readWriteMany,omitempty"`
	// Creates an actual PersistentVolumeClaim object in the same namespace as the Pod and ensures that the PersistentVolumeClaim gets deleted when the Pod gets deleted
	Ephemeral *bool `json:"ephemeral,omitempty"`
	// VolumeMode of the PVC. The default is Filesystem, can be set to Block.
	VolumeMode *string `json:"volumeMode,omitempty"`
}

func (self *PvcsField) Deref() map[string]PvcsItemAndRules {
	if self != nil && self.Value != nil {
		return *self.Value
	} else {
		return map[string]PvcsItemAndRules{}
	}
}

func NewPvcsField(pvcsMap map[string]PvcsItemAndRules) *PvcsField {
	if len(pvcsMap) == 0 {
		return nil
	}
	return &PvcsField{
		Value: &pvcsMap,
	}
}
