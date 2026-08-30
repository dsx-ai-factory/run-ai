package field_types

import "sort"

// +k8s:deepcopy-gen=true
type CapabilitiesField struct {
	Rules     *MapRules                            `json:"rules,omitempty"`
	ItemRules *CapabilitiesItemRules               `json:"itemRules,omitempty"`
	Value     *map[string]CapabilitiesItemAndRules `json:"items,omitempty"`
}

// +k8s:deepcopy-gen=true
type CapabilitiesItemAndRules struct {
	Rules   *MapItemRules          `json:"rules,omitempty"`
	Value   *CapabilitiesItemValue `json:"value,omitempty"`
	Deleted *bool                  `json:"deleted,omitempty"`
}

// +k8s:deepcopy-gen=true
type CapabilitiesItemRules struct {
	Capability *StringRules `json:"capability"`
}

// +k8s:deepcopy-gen=true
type CapabilitiesItemValue struct {
	//+kubebuilder:validation:Required
	Capability *string `json:"capability"`
}

func (self *CapabilitiesField) Deref() map[string]CapabilitiesItemAndRules {
	if self != nil && self.Value != nil {
		return *self.Value
	} else {
		return map[string]CapabilitiesItemAndRules{}
	}
}

func NewCapabilitiesField(CapabilitiesMap map[string]CapabilitiesItemAndRules) *CapabilitiesField {
	if len(CapabilitiesMap) == 0 {
		return nil
	}
	return &CapabilitiesField{
		Value: &CapabilitiesMap,
	}
}

func SortedCapabilitiesMapKeys(unsorted map[string]CapabilitiesItemAndRules) []string {
	result := make([]string, 0, len(unsorted))
	for k := range unsorted {
		result = append(result, k)
	}
	sort.Strings(result)
	return result
}
