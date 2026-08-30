package field_types

// +k8s:deepcopy-gen=true
type VolumesField struct {
	Rules     *MapRules                      `json:"rules,omitempty"`
	ItemRules *VolumeItemRules               `json:"itemRules,omitempty"`
	Value     *map[string]VolumeItemAndRules `json:"items,omitempty"`
}

// +k8s:deepcopy-gen=true
type VolumeItemAndRules struct {
	Rules   *MapItemRules    `json:"rules,omitempty"`
	Value   *VolumeItemValue `json:"value,omitempty"`
	Deleted *bool            `json:"deleted,omitempty"`
}

// +k8s:deepcopy-gen=true
type VolumeItemRules struct {
	SourcePath *StringRules `json:"sourcePath,omitempty"`
	TargetPath *StringRules `json:"targetPath,omitempty"`
	ReadOnly   *BoolRules   `json:"readOnly,omitempty"`
	// this attribute is obsolete. for specifying NFS volume use the Nfs fields of the workload.
	NfsServer *StringRules `json:"nfsServer,omitempty"`
}

// +k8s:deepcopy-gen=true
type VolumeItemValue struct {
	SourcePath *string `json:"sourcePath,omitempty"`
	// +kubebuilder:validation:Required
	TargetPath *string `json:"targetPath"`

	ReadOnly *bool `json:"readOnly,omitempty"`

	NfsServer *string `json:"nfsServer,omitempty"`
}

func (self *VolumesField) Deref() map[string]VolumeItemAndRules {
	if self != nil && self.Value != nil {
		return *self.Value
	} else {
		return map[string]VolumeItemAndRules{}
	}
}

func NewVolumesField(volumesMap map[string]VolumeItemAndRules) *VolumesField {
	if len(volumesMap) == 0 {
		return nil
	}
	return &VolumesField{
		Value: &volumesMap,
	}
}
