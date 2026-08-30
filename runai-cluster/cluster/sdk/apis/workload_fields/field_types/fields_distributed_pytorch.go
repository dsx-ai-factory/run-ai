package field_types

// +k8s:deepcopy-gen=true
type PyTorchField struct {
	Values *PyTorchMembers `json:"values,omitempty"`
}

// +k8s:deepcopy-gen=true
type PyTorchMembers struct {
	// minReplicas is the lower limit for the number of worker pods to which the training job can scale down.
	// +optional
	MinReplicas *IntField `json:"minReplicas,omitempty"`
	// upper limit for the number of worker pods that can be set by the autoscaler. Cannot be smaller than MinReplicas.
	// +optional
	MaxReplicas *IntField `json:"maxReplicas,omitempty"`
}

func (self *PyTorchField) Deref() PyTorchMembers {
	if self != nil && self.Values != nil {
		return *self.Values
	} else {
		return PyTorchMembers{}
	}
}
