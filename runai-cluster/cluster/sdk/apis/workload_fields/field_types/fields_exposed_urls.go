package field_types

// +k8s:deepcopy-gen=true
type ExposedUrlsField struct {
	Rules     *MapRules                          `json:"rules,omitempty"`
	ItemRules *ExposedUrlItemRules               `json:"itemRules,omitempty"`
	Value     *map[string]ExposedUrlItemAndRules `json:"items,omitempty"`
}

// +k8s:deepcopy-gen=true
type ExposedUrlItemAndRules struct {
	Rules   *MapItemRules        `json:"rules,omitempty"`
	Value   *ExposedUrlItemValue `json:"value,omitempty"`
	Deleted *bool                `json:"deleted,omitempty"`
}

// +k8s:deepcopy-gen=true
type ExposedUrlItemRules struct {
	ContainerPort    *IntRules    `json:"containerPort,omitempty"`
	CustomUrl        *BoolRules   `json:"customUrl,omitempty"`
	Url              *StringRules `json:"url,omitempty"`
	AuthorizedUsers  *StringRules `json:"authorizedUsers,omitempty"`
	AuthorizedGroups *StringRules `json:"authorizedGroups,omitempty"`
}

// +k8s:deepcopy-gen=true
type ExposedUrlItemValue struct {
	//+kubebuilder:validation:Required
	ContainerPort *int32 `json:"containerPort"`
	// if enabled, it should be possible to request custom URL for this container port.
	// it is recommended to leave it disabled. in this case the system composes a unique URL for the container port.
	CustomUrl *bool `json:"customUrl,omitempty"`
	// the custom URL to assign to the container port, in case CustomUrl is enabled.
	Url *string `json:"url,omitempty"`
	// Specifies the list of users that allowed to access the exposedUrl (seperated by comma).
	// if empty, exposedUrl will be shared.
	AuthorizedUsers *string `json:"authorizedUsers,omitempty"`
	// Specifies the list of groups that allowed to access the exposedUrl (seperated by comma).
	// if empty, exposedUrl will be shared.
	AuthorizedGroups *string `json:"authorizedGroups,omitempty"`
}

func (self *ExposedUrlsField) Deref() map[string]ExposedUrlItemAndRules {
	if self != nil && self.Value != nil {
		return *self.Value
	} else {
		return map[string]ExposedUrlItemAndRules{}
	}
}

func NewExposedUrlField(urlsMap map[string]ExposedUrlItemAndRules) *ExposedUrlsField {
	if len(urlsMap) == 0 {
		return nil
	}
	return &ExposedUrlsField{
		Value: &urlsMap,
	}
}
