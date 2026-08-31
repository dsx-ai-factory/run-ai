/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package field_types

import (
	"fmt"

	"github.com/run-ai/runai/runai-cluster/common/utils"
	corev1 "k8s.io/api/core/v1"
)

const (
	secretVolumePrefix = "secret-volume" // secretVolumePrefix is the prefix for naming volume base on a Secret resource
)

// SecretVolumeField contains the definition to a Secret resources use as a volume, with policy setting
// +k8s:deepcopy-gen=true
type SecretVolumeField struct {
	Rules     *MapRules                            `json:"rules,omitempty"`
	ItemRules *SecretVolumeItemRules               `json:"itemRules,omitempty"`
	Value     *map[string]SecretVolumeItemAndRules `json:"items,omitempty"`
}

// NewSecretVolumeField returns a Secret spec definition
func NewSecretVolumeField(secrets map[string]SecretVolumeItemAndRules) *SecretVolumeField {
	if len(secrets) == 0 {
		return nil
	}
	return &SecretVolumeField{
		Value: &secrets,
	}
}

// Deref returns the items and rules of a Secret volume field
func (s *SecretVolumeField) Deref() map[string]SecretVolumeItemAndRules {
	if s != nil && s.Value != nil {
		return *s.Value
	} else {
		return map[string]SecretVolumeItemAndRules{}
	}
}

// +k8s:deepcopy-gen=true
type SecretVolumeItemAndRules struct {
	Rules   *MapItemRules          `json:"rules,omitempty"`
	Value   *SecretVolumeItemValue `json:"value,omitempty"`
	Deleted *bool                  `json:"deleted,omitempty"`
}

// +k8s:deepcopy-gen=true
type SecretVolumeItemRules struct {
	Name      *StringRules `json:"name,omitempty"`
	MountPath *StringRules `json:"mountPath,omitempty"`
}

// SecretVolumeItemValue contains the definition to a Secret resources use as a volume
// +k8s:deepcopy-gen=true
type SecretVolumeItemValue struct {
	// Name is the name of the Secret resource
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names
	// +kubebuilder:validation:Required
	Name *string `json:"name"`
	// MountPath path within the container at which the volume should be mounted.  Must
	// not contain ':'.
	// +kubebuilder:validation:Required
	MountPath *string `json:"mountPath"`
	// DefaultMode sets the default file permission mode for files created from the Secret or ConfigMap.
	// Must be a 4-digit octal string (e.g., "0644", "0600").
	// This controls the permission bits when mounting the volume.
	// If unset, the default is 0644.
	// +optional
	DefaultMode *string `json:"defaultMode,omitempty"`
}

// VolumeMount returns a volume mount spec out of the Secret volume
func (s *SecretVolumeItemValue) VolumeMount(name string) corev1.VolumeMount {
	return corev1.VolumeMount{
		Name:      name,
		MountPath: utils.Deref(s.MountPath),
	}
}

// Volume returns a volume spec out of the Secret volume
// More Info: https://kubernetes.io/docs/concepts/storage/volumes/#secret
func (s *SecretVolumeItemValue) Volume(name string) corev1.Volume {
	// No need to verify s.Secret as it is required by definition
	return corev1.Volume{
		Name: name,
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName:  utils.Deref(s.Name),
				DefaultMode: toInt32PtrFromOctalString(s.DefaultMode),
			},
		},
	}
}

// GenerateSecretVolumeName returns a name of spec of the secret volume
func GenerateSecretVolumeName(index int) string {
	return fmt.Sprintf("%s-%d", secretVolumePrefix, index)
}
