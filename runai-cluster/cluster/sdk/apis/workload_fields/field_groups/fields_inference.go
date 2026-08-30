package field_groups

import (
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	flds "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	"k8s.io/utils/ptr"
)

// +k8s:deepcopy-gen=true
type InferenceFields struct {
	RoutingAndAuthorizationFields `json:""`
	AutoScalingFields             `json:""`
	ServingConfigurationFields    `json:""`
	MpsFields                     `json:""`
}

// +k8s:deepcopy-gen=true
type RoutingAndAuthorizationFields struct {
	// Configure the inference service to be available only on the cluster-local network, and not on the public internet
	IsPrivateServiceUrl *flds.BoolField `json:"isPrivateServiceUrl,omitempty"`
	// The serving port authorization type. Possible values are: public (default), authenticatedUsers, authorizedUsersOrGroups
	ServingPortAuthorizationType *flds.StringField `json:"servingPortAuthorizationType,omitempty"`
	// Specifies the list of users that are allowed to access the serving port (seperated by comma).
	// Should be provided if and only if the authorization type is set to authorizedUsersOrGroups.
	ServingPortAuthorizedUsers *flds.StringField `json:"servingPortAuthorizedUsers,omitempty"`
	// Specifies the list of groups that are allowed to access the serving port (seperated by comma).
	// Should be provided if and only if the authorization type is set to authorizedUsersOrGroups.
	ServingPortAuthorizedGroups *flds.StringField `json:"servingPortAuthorizedGroups,omitempty"`
}

// +k8s:deepcopy-gen=true
type ServingConfigurationFields struct {
	// The maximum time (in seconds) allowed to process an end-user request. If no response is returned within this time, the request will be ignored.
	RequestTimeoutSeconds *flds.IntField `json:"requestTimeoutSeconds,omitempty"`
	// The maximum time (in seconds) allowed for a workload to initialize and become ready. If the workload does not start within this time, it will be moved to failed state.
	InitializationTimeoutSeconds *flds.IntField `json:"initializationTimeoutSeconds,omitempty"`
}

// +k8s:deepcopy-gen=true
type AutoScalingFields struct {
	// The predefined metric to use for autoscaling. Possible values are: throughput, concurrency, latency.
	Metric *flds.StringField `json:"metric,omitempty"`
	// The exact metric name to use for autoscaling (overrides Metric field)
	MetricName *flds.StringField `json:"metricName,omitempty"`
	// The target value for the autoscaling metric
	Target *flds.IntField `json:"target,omitempty"`
	// The percentage of target value to use for autoscaling
	TargetUtilizationPercentage *flds.NumberField `json:"targetUtilizationPercentage,omitempty"`
	// The maximum number of requests allowed to flow to a single replica at any time. 0 means no limit
	ConcurrencyHardLimit *flds.IntField `json:"concurrencyHardLimit,omitempty"`
	// The autoscaler class for knative to use
	Class *flds.StringField `json:"class,omitempty"`
	// The minimum number of replicas to run
	MinScale *flds.IntField `json:"minScale,omitempty"`
	// The maximum number of replicas to run
	MaxScale *flds.IntField `json:"maxScale,omitempty"`
	// The number of replicas to run when initializing the workload (first time only)
	InitialScale *flds.IntField `json:"initialScale,omitempty"`
	// The number of replicas to run when scaling-up from zero
	ActivationScale *flds.IntField `json:"activationScale,omitempty"`
	// The minimum amount of time (in seconds) that the last replica will remain active after a scale-to-zero decision
	ScaleToZeroRetentionSeconds *flds.IntField `json:"scaleToZeroRetentionSeconds,omitempty"`
	// The minimum amount of time (in seconds) that a replica will remain active after a scale-down decision
	ScaleDownDelaySeconds *flds.IntField `json:"scaleDownDelaySeconds,omitempty"`
}

// +k8s:deepcopy-gen=true
type MpsFields struct {
	Mps *flds.BoolField `json:"mps,omitempty"`
}

var InferenceRunaiRules = InferenceFields{
	AutoScalingFields: AutoScalingFields{
		MinScale: &flds.IntField{
			Rules: &flds.IntRules{
				Min: ptr.To(0),
			},
		},
		MaxScale: &flds.IntField{
			Rules: &flds.IntRules{
				Min: ptr.To(0),
			},
		},
		Metric: &flds.StringField{
			Rules: &flds.StringRules{
				Options: &[]flds.OptionRule{
					{
						Value: ptr.To("throughput"),
						Name:  ptr.To("Throughput (Requests/sec)"),
					},
					{
						Value: ptr.To("concurrency"),
						Name:  ptr.To("Concurrency (Requests/sec)"),
					},
					{
						Value: ptr.To("latency"),
						Name:  ptr.To("Latency (ms)"),
					},
				},
			},
		},
	},
	RoutingAndAuthorizationFields: RoutingAndAuthorizationFields{
		ServingPortAuthorizationType: &flds.StringField{
			Value: ptr.To(constants.ServingPortAuthorizationTypeDefault), // RUNAI DEFAULT
			Rules: &flds.StringRules{
				Options: &[]flds.OptionRule{
					{
						Value: ptr.To("public"),
						Name:  ptr.To("Public"),
					},
					{
						Value: ptr.To("authenticatedUsers"),
						Name:  ptr.To("Authenticated users"),
					},
					{
						Value: ptr.To("authorizedUsersOrGroups"),
						Name:  ptr.To("Authorized users or groups"),
					},
				},
			},
		},
	},
}

// +k8s:deepcopy-gen=true
type SecurityFields struct {
	// Allow the container running the workload to operate in privileged mode, granting it elevated
	// access to the host system. When enabled, the container receives all Linux capabilities,
	// unrestricted access to host devices, and is exempt from security constraints. This effectively
	// removes many of the isolation boundaries between the container and the host.
	Privileged *flds.BoolField `json:"privileged,omitempty"`

	// Specifies that the created workload will use the host's ipc namespace.
	// +k8s:conversion-gen=false
	HostIpc *flds.BoolField `json:"hostIpc,omitempty"`

	// Specifies that the created workload will use the host's network stack inside its container.
	// For more information see the Docker Run Reference at
	// https://docs.docker.com/engine/reference/run/
	HostNetwork *flds.BoolField `json:"hostNetwork,omitempty"`
}
