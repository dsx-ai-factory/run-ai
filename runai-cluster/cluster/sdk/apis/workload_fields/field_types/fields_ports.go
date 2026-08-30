package field_types

import "sort"

// +k8s:deepcopy-gen=true
type PortsField struct {
	Rules     *MapRules                    `json:"rules,omitempty"`
	ItemRules *PortItemRules               `json:"itemRules,omitempty"`
	Value     *map[string]PortItemAndRules `json:"items,omitempty"`
}

// +k8s:deepcopy-gen=true
type PortItemAndRules struct {
	Rules   *MapItemRules  `json:"rules,omitempty"`
	Value   *PortItemValue `json:"value,omitempty"`
	Deleted *bool          `json:"deleted,omitempty"`
}

// +k8s:deepcopy-gen=true
type PortItemRules struct {
	ServiceType     *StringRules `json:"serviceType,omitempty"`
	ContainerPort   *IntRules    `json:"container,omitempty"`
	HostPort        *IntRules    `json:"external,omitempty"`
	AutoGenHostPort *BoolRules   `json:"autoGenerate,omitempty"`
	Protocol        *StringRules `json:"protocol,omitempty"`
}

// +k8s:deepcopy-gen=true
type PortItemValue struct {
	// Specifies the service exposure method for this port. if not specified, the default
	// serviceType specified for the workload will be used.
	// Options are: LoadBalancer, NodePort or ClusterIP.
	// For more information see the External Access to Containers guide on
	// https://docs.run.ai/admin/runai-setup/config/allow-external-access-to-containers/
	ServiceType *string `json:"serviceType,omitempty"`
	// +optional
	ContainerPort *int `json:"container"`

	HostPort *int `json:"external,omitempty"`

	AutoGenHostPort *bool `json:"autoGenerate,omitempty"`

	Protocol *string `json:"protocol,omitempty"`
}

func (self *PortsField) Deref() map[string]PortItemAndRules {
	if self != nil && self.Value != nil {
		return *self.Value
	} else {
		return map[string]PortItemAndRules{}
	}
}

func NewPortsField(portsMap map[string]PortItemAndRules) *PortsField {
	if len(portsMap) == 0 {
		return nil
	}
	return &PortsField{
		Value: &portsMap,
	}
}

func SortedPortsMapKeys(unsorted map[string]PortItemAndRules) []string {
	result := make([]string, 0, len(unsorted))
	for k := range unsorted {
		result = append(result, k)
	}
	sort.Strings(result)
	return result
}
