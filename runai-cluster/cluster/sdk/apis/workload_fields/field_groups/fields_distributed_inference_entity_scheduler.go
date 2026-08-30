package field_groups

import (
	flds "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
)

// +k8s:deepcopy-gen=true
type DistributedInferenceEntitySchedulerFields struct {

	// Specifies nodes (machines) or a group of nodes on which the workload will run. To use this
	// feature, your Administrator will need to label nodes as explained in the Group Nodes guide at
	// https://docs.run.ai/admin/researcher-setup/limit-to-node-group.
	// This flag can be used in conjunction with Project-based affinity. In this case, the flag is
	// used to refine the list of allowable node groups set in the Project. For more information see
	// the Projects setup guide at https://docs.run.ai/admin/admin-ui-setup/project-setup.
	NodeType *flds.StringField `json:"nodeType,omitempty"`

	// Specifies the node affinity requirements for the job's pods.
	// https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/#node-affinity
	flds.NodeAffinityFields `json:""`

	// Indicates whether pod affinity scheduling rules apply.
	PodAffinity *flds.BoolField `json:"podAffinity,omitempty"`

	// Specifies the Pod Affinity Topology to be used for scheduling the job
	// This field can be specified only if PodAffinity is set to true
	PodAffinityTopology *flds.StringField `json:"podAffinityTopology,omitempty"`

	// Indicates if we want to use the Pod affinity rule as :
	// the "hard" (required) or the "soft" (preferred)
	// This field can be specified only if PodAffinity is set to true
	PodAffinitySchedulingRule *flds.StringField `json:"podAffinitySchedulingRule,omitempty"`
}

func DistributedInferenceEntitySchedulerRunaiRules() DistributedInferenceEntitySchedulerFields {
	var result DistributedInferenceEntitySchedulerFields

	result.PodAffinityTopology = getPodAffinityTopologyRules()
	result.PodAffinitySchedulingRule = getPodAffinitySchedulingRules()

	return result
}
