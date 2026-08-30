package field_types

// +k8s:deepcopy-gen=true
type GitSyncField struct {
	Rules     *MapRules                       `json:"rules,omitempty"`
	ItemRules *GitSyncItemRules               `json:"itemRules,omitempty"`
	Value     *map[string]GitSyncItemAndRules `json:"items,omitempty"`
}

// +k8s:deepcopy-gen=true
type GitSyncItemAndRules struct {
	Rules   *MapItemRules     `json:"rules,omitempty"`
	Value   *GitSyncItemValue `json:"value,omitempty"`
	Deleted *bool             `json:"deleted,omitempty"`
}

// +k8s:deepcopy-gen=true
type GitSyncItemRules struct {
	Repository *StringRules `json:"repository"`
	Branch     *StringRules `json:"branch,omitempty"`
	Revision   *StringRules `json:"revision,omitempty"`
	Directory  *StringRules `json:"directory,omitempty"`
	ReadOnly   *BoolRules   `json:"readOnly,omitempty"`
	Image      *StringRules `json:"image,omitempty"`

	SecretName *StringRules `json:"secretName,omitempty"`
	Username   *StringRules `json:"username,omitempty"`
	Password   *StringRules `json:"password,omitempty"`
}

// +k8s:deepcopy-gen=true
type GitSyncItemValue struct {
	//+kubebuilder:validation:Required
	Repository *string `json:"repository"`

	Branch *string `json:"branch,omitempty"`

	Revision *string `json:"rev,omitempty"`
	// +kubebuilder:validation:Required
	Directory *string `json:"directory,omitempty"`

	ReadOnly *bool `json:"readOnly,omitempty"`

	Image *string `json:"image,omitempty"`

	// optional name of kubernetes secret that holds the authenticating info (git username and password or SSH key)
	SecretName *string `json:"secretName,omitempty"`
	// AuthenticatingMethod defines the authenticating method (password or SSH key) in the provided SecretName
	AuthenticatingMethod *string `json:"authenticatingMethod,omitempty"`
	// if secretName is provided, this field should contain the key, within the provided kubernetes secret, which holds the
	// value of your git username. otherwise, this field should specify your git username in plain text (example: myuser).
	Username *string `json:"username,omitempty"`
	// if secretName is provided, this field should contain the key, within the provided kubernetes secret, which holds the
	// the value of your git password. otherwise, this field should specify your git password in plain text (example: mypassword).
	Password *string `json:"password,omitempty"`
}

func (self *GitSyncField) Deref() map[string]GitSyncItemAndRules {
	if self != nil && self.Value != nil {
		return *self.Value
	} else {
		return map[string]GitSyncItemAndRules{}
	}
}

func NewGitSyncField(gitSyncMap map[string]GitSyncItemAndRules) *GitSyncField {
	if len(gitSyncMap) == 0 {
		return nil
	}
	return &GitSyncField{
		Value: &gitSyncMap,
	}
}
