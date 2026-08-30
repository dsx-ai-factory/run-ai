package field_types

// +k8s:deepcopy-gen=true
type EmptyDirVolumeField struct {
	Rules     *MapRules                              `json:"rules,omitempty"`
	ItemRules *EmptyDirVolumeItemRules               `json:"itemRules,omitempty"`
	Value     *map[string]EmptyDirVolumeItemAndRules `json:"items,omitempty"`
}

// +k8s:deepcopy-gen=true
type EmptyDirVolumeItemAndRules struct {
	Rules   *MapItemRules            `json:"rules,omitempty"`
	Value   *EmptyDirVolumeItemValue `json:"value,omitempty"`
	Deleted *bool                    `json:"deleted,omitempty"`
}

// +k8s:deepcopy-gen=true
type EmptyDirVolumeItemRules struct {
	Name      *StringRules `json:"name,omitempty"`
	Path      *StringRules `json:"path,omitempty"`
	Medium    *StringRules `json:"medium,omitempty"`
	SizeLimit *StringRules `json:"sizeLimit,omitempty"`
}

// +k8s:deepcopy-gen=true
type EmptyDirVolumeItemValue struct {
	// Path within the container at which the emptyDir volume should be mounted. Required.
	Path *string `json:"path,omitempty"`
	// Medium specifies the type of storage medium backing this directory.
	// The default is "" (empty string), which means the node’s default storage medium is used.
	// Valid values: "" (default), "Memory".
	Medium *string `json:"medium,omitempty"`
	// For medium "Memory", the actual usage limit is the lesser of this value and the combined memory limits of all containers in the pod.
	// The default is nil which means that the limit is undefined.
	SizeLimit *string `json:"sizeLimit,omitempty"`
}

func (self *EmptyDirVolumeField) Deref() map[string]EmptyDirVolumeItemAndRules {
	if self != nil && self.Value != nil {
		return *self.Value
	} else {
		return map[string]EmptyDirVolumeItemAndRules{}
	}
}

func NewEmptyDirVolumeField(emptyDirMap map[string]EmptyDirVolumeItemAndRules) *EmptyDirVolumeField {
	if len(emptyDirMap) == 0 {
		return nil
	}
	return &EmptyDirVolumeField{
		Value: &emptyDirMap,
	}
}
