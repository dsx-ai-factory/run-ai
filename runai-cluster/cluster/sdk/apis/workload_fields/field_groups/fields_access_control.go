package field_groups

import flds "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"

// +k8s:deepcopy-gen=true
type AccessControlFields struct {

	// Allow the container running the workload and all launched processes to gain additional
	// privileges after the workload starts. For more information see the "User Identity in Container"
	// guide at https://docs.run.ai/admin/runai-setup/config/non-root-containers/
	AllowPrivilegeEscalation *flds.BoolField `json:"allowPrivilegeEscalation,omitempty"`

	// Allow the container running the workload to operate in privileged mode, granting it elevated
	// access to the host system. When enabled, the container receives all Linux capabilities,
	// unrestricted access to host devices, and is exempt from security constraints. This effectively
	// removes many of the isolation boundaries between the container and the host.
	Privileged *flds.BoolField `json:"privileged,omitempty"`
}
