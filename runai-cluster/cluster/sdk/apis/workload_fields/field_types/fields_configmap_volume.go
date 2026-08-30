package field_types

import (
	"fmt"
	"k8s.io/utils/ptr"
	"strconv"

	"github.com/run-ai/runai/runai-cluster/common/utils"
	corev1 "k8s.io/api/core/v1"
)

const (
	configMapVolumePrefix = "cm-volume" // configMapVolumePrefix is the prefix for naming volume base on ConfigMap resource
)

// ConfigMapVolumeField contains the definition to ConfigMap resources use as a volume, with policy setting
// +k8s:deepcopy-gen=true
type ConfigMapVolumeField struct {
	Rules     *MapRules                               `json:"rules,omitempty"`
	ItemRules *ConfigMapVolumeItemRules               `json:"itemRules,omitempty"`
	Value     *map[string]ConfigMapVolumeItemAndRules `json:"items,omitempty"`
}

// NewConfigMapVolumeField returns ConfigMap spec definition
func NewConfigMapVolumeField(configMaps map[string]ConfigMapVolumeItemAndRules) *ConfigMapVolumeField {
	if len(configMaps) == 0 {
		return nil
	}
	return &ConfigMapVolumeField{
		Value: &configMaps,
	}
}

// Deref returns the items and rules of ConfigMap volume field
func (c *ConfigMapVolumeField) Deref() map[string]ConfigMapVolumeItemAndRules {
	if c != nil && c.Value != nil {
		return *c.Value
	} else {
		return map[string]ConfigMapVolumeItemAndRules{}
	}
}

// +k8s:deepcopy-gen=true
type ConfigMapVolumeItemAndRules struct {
	Rules   *MapItemRules             `json:"rules,omitempty"`
	Value   *ConfigMapVolumeItemValue `json:"value,omitempty"`
	Deleted *bool                     `json:"deleted,omitempty"`
}

// +k8s:deepcopy-gen=true
type ConfigMapVolumeItemRules struct {
	Name      *StringRules `json:"name,omitempty"`
	MountPath *StringRules `json:"mountPath,omitempty"`
}

// ConfigMapVolumeItemValue contains the definition to ConfigMap resources use as a volume
// +k8s:deepcopy-gen=true
type ConfigMapVolumeItemValue struct {
	// Name is the name of the ConfigMap resource
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names
	// +kubebuilder:validation:Required
	Name *string `json:"name"`
	// MountPath path within the container at which the volume should be mounted.  Must
	// not contain ':'.
	// +kubebuilder:validation:Required
	MountPath *string `json:"mountPath"`
	// Path within the volume from which the container's volume should be mounted.
	// Defaults to "" (volume's root).
	// +optional
	SubPath *string `json:"subPath,omitempty"`
	// DefaultMode sets the default file permission mode for files created from the Secret or ConfigMap.
	// Must be a 4-digit octal string (e.g., "0644", "0600").
	// This controls the permission bits when mounting the volume.
	// If unset, the default is 0644.
	// +optional
	DefaultMode *string `json:"defaultMode,omitempty"`
}

// VolumeMount returns a volume mount spec out of the ConfigMap volume
func (c *ConfigMapVolumeItemValue) VolumeMount(name string) corev1.VolumeMount {
	return corev1.VolumeMount{
		Name:      name,
		MountPath: utils.Deref(c.MountPath),
		SubPath:   ptr.Deref(c.SubPath, ""),
	}
}

// Volume returns a volume spec out of the ConfigMap volume
// More Info: https://kubernetes.io/docs/concepts/storage/volumes/#configmap
func (c *ConfigMapVolumeItemValue) Volume(name string) corev1.Volume {
	// No need to verify c.ConfigMap as it is required by definition
	return corev1.Volume{
		Name: name,
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: utils.Deref(c.Name),
				},
				DefaultMode: toInt32PtrFromOctalString(c.DefaultMode),
			},
		},
	}
}

func toInt32PtrFromOctalString(s *string) *int32 {
	if s == nil {
		return nil
	}
	val, err := strconv.ParseInt(*s, 8, 32)
	if err != nil {
		return nil
	}
	i32 := int32(val)
	return &i32
}

// GenerateConfigMapVolumeName returns a name of spec of the ConfigMap volume
func GenerateConfigMapVolumeName(index int) string {
	return fmt.Sprintf("%s-%d", configMapVolumePrefix, index)
}
