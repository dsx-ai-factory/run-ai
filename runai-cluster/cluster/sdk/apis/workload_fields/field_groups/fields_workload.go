package field_groups

//	the fields below are shared b/w all workload
//
// +k8s:deepcopy-gen=true
type WorkloadFields struct {
	IdentityFields           `json:""`
	IdentityDeprecatedFields `json:""`
	ContainerFields          `json:""`
	StorageFields            `json:""`
	NetworkingFields         `json:""`
}

// WorkloadRunaiRules setup rules that are shared b/w all workloads
func WorkloadRunaiRules() WorkloadFields {
	return WorkloadFields{
		IdentityFields:           IdentityRunaiRules(),
		IdentityDeprecatedFields: IdentityDeprecatedRunaiRules(),
		ContainerFields:          RunaiContainerRules(),
		StorageFields:            StorageRunaiRules,
		NetworkingFields:         NetworkingRunaiRules,
	}
}
