package field_types

// +k8s:deepcopy-gen=true
type DataVolumeField struct {
	Rules     *MapRules                          `json:"rules,omitempty"`
	ItemRules *DataVolumesItemRules              `json:"itemRules,omitempty"`
	Value     *map[string]DataVolumeItemAndRules `json:"items,omitempty"`
}

// +k8s:deepcopy-gen=true
type DataVolumeItemAndRules struct {
	Rules   *MapItemRules        `json:"rules,omitempty"`
	Value   *DataVolumeItemValue `json:"value,omitempty"`
	Deleted *bool                `json:"deleted,omitempty"`
}

// +k8s:deepcopy-gen=true
type DataVolumesItemRules struct {
	Id        *StringRules `json:"id,omitempty"`
	MountPath *StringRules `json:"mountPath,omitempty"`
}

// +k8s:deepcopy-gen=true
type DataVolumeItemValue struct {
	// the identifier of the data volume
	Id *string `json:"id,omitempty"`
	// the path to mount the data volume
	MountPath *string `json:"mountPath,omitempty"`
}

func (self *DataVolumeField) Deref() map[string]DataVolumeItemAndRules {
	if self != nil && self.Value != nil {
		return *self.Value
	} else {
		return map[string]DataVolumeItemAndRules{}
	}
}
