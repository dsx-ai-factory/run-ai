package field_types

// +k8s:deepcopy-gen=true
type BoolRules struct {
	Required *bool `json:"required,omitempty"`
	// +kubebuilder:default:=true
	CanEdit *bool `json:"canEdit,omitempty"`
}

// +k8s:deepcopy-gen=true
type StringRules struct {
	Required *bool `json:"required,omitempty"`
	// +kubebuilder:default:=true
	CanEdit *bool `json:"canEdit,omitempty"`

	Options *[]OptionRule `json:"options,omitempty"`

	// +kubebuilder:validation:Enum:=envVar;password;email;volume;kubeName;name;duration;memory;imageName;gpuMemory
	Format *string `json:"format,omitempty"`
}

// +k8s:deepcopy-gen=true
type QuantityRules struct {
	Required *bool `json:"required,omitempty"`
	// +kubebuilder:default:=true
	CanEdit *bool `json:"canEdit,omitempty"`

	//+kubebuilder:validation:Pattern:=^([+-]?[0-9.]+)([eEinumkKMGTP]*[-+]?[0-9]*)$
	Min *string `json:"min,omitempty"`
	//+kubebuilder:validation:Pattern:=^([+-]?[0-9.]+)([eEinumkKMGTP]*[-+]?[0-9]*)$
	Max *string `json:"max,omitempty"`

	// Options is left for backward compatibility, since it was possible to provide options for
	// quantity fields when they were defined as strings. it can be given, and will be validated,
	// but preferably the more relevant are the min and max
	Options *[]OptionRule `json:"options,omitempty"`
}

// +k8s:deepcopy-gen=true
type OptionRule struct {
	Name     *string `json:"name,omitempty"`
	Value    *string `json:"value,omitempty"`
	Disabled *bool   `json:"disabled,omitempty"`
}

// +k8s:deepcopy-gen=true
type IntRules struct {
	Required *bool `json:"required,omitempty"`
	// +kubebuilder:default:=true
	CanEdit *bool `json:"canEdit,omitempty"`

	Min *int `json:"min,omitempty"`
	Max *int `json:"max,omitempty"`

	// +kubebuilder:validation:Enum:=port
	Format *string `json:"format,omitempty"`
}

// +k8s:deepcopy-gen=true
type NumberRules struct {
	Required *bool `json:"required,omitempty"`
	// +kubebuilder:default:=true
	CanEdit *bool `json:"canEdit,omitempty"`

	//+kubebuilder:validation:Pattern:=^[+-]?([0-9]+\.?[0-9]*|\.[0-9]+)$
	Min *string `json:"min,omitempty"`
	//+kubebuilder:validation:Pattern:=^[+-]?([0-9]+\.?[0-9]*|\.[0-9]+)$
	Max *string `json:"max,omitempty"`
	//+kubebuilder:validation:Pattern:=^[+-]?([0-9]+\.?[0-9]*|\.[0-9]+)$
	Step *string `json:"step,omitempty"`
}
