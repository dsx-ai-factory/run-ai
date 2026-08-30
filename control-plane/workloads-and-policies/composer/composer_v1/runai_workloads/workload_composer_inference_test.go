package runai_workloads

import (
	. "github.com/onsi/ginkgo/v2"
	workloadsApi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
)

var inferenceTests = []TableEntry{
	Entry("inference servingPort", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_INFERENCE),
		specFname:     "serving_port",
		expectedFname: "serving_port_inf",
	}),

	/*WIT_FOR_OFER
	  Entry("inference servingPort from defaults", ComposerTest{
	  	workloadType:  string(workloadsApi.WORKLOADTYPE_INFERENCE),
	  	defaultsFname: "serving_port",
	  	expectedFname: "serving_port_from_defaults",
	  }),
	*/

	Entry("inference servingPort without protocol", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_INFERENCE),
		specFname:     "serving_port_no_protocol",
		expectedFname: "serving_port_no_protocol_inf",
	}),

	Entry("inference servingPort and ports", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_INFERENCE),
		specFname:     "serving_port_and_ports",
		expectedFname: "serving_port_and_ports_inf",
	}),

	Entry("inference invalid no servingPort", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_INFERENCE),
		specFname:     "ports",
		expectedError: "mandatory servingPort is not provided",
	}),

	Entry("inference invalid servingPort without container", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_INFERENCE),
		specFname:     "serving_port_invalid_no_container",
		expectedError: "servingPort: mandatory container is not provided",
	}),

	Entry("inference servingPort auth - authenticated users", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_INFERENCE),
		specFname:     "serving_port_auth_authenticated_users",
		expectedFname: "serving_port_auth_authenticated_users_inf",
	}),

	Entry("inference servingPort auth - authorized users", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_INFERENCE),
		specFname:     "serving_port_auth_authorized_users",
		expectedFname: "serving_port_auth_authorized_users_inf",
	}),

	Entry("inference servingPort auth - authorized users (enhanced, specific auth type)", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_INFERENCE),
		specFname:     "serving_port_auth_authorized_users_new",
		expectedFname: "serving_port_auth_authorized_users_new",
	}),

	Entry("inference servingPort auth - authorized groups", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_INFERENCE),
		specFname:     "serving_port_auth_authorized_groups",
		expectedFname: "serving_port_auth_authorized_groups_inf",
	}),

	Entry("inference servingPort auth - authorized group (enhanced, with specific auth type)", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_INFERENCE),
		specFname:     "serving_port_auth_authorized_groups_new",
		expectedFname: "serving_port_auth_authorized_groups_new",
	}),

	Entry("inference servingPort - cluster local access only - true", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_INFERENCE),
		specFname:     "serving_port_cluster_local",
		expectedFname: "serving_port_cluster_local_inf",
	}),

	Entry("inference servingPort - cluster local access only - false", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_INFERENCE),
		specFname:     "serving_port_cluster_local_false_by_default",
		expectedFname: "serving_port_cluster_local_false_by_default_inf",
	}),

	Entry("inference servingPort - cluster local access only - false", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_INFERENCE),
		specFname:     "serving_port_cluster_local_false",
		expectedFname: "serving_port_cluster_local_false_inf",
	}),

	// enhanced behavior: just ignore the list of users and do not send it to the cluster
	Entry("inference invalid servingPort auth - default with authorizedUsers", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_INFERENCE),
		specFname:     "serving_port_auth_default_with_auth_users",
		expectedFname: "serving_port_auth_default_with_auth_users",
	}),

	// enhanced behavior: just ignore the list of groups and do not send it to the cluster
	Entry("inference invalid servingPort auth - default with authorizedGroups", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_INFERENCE),
		specFname:     "serving_port_auth_default_with_auth_groups",
		expectedFname: "serving_port_auth_default_with_auth_groups",
	}),

	// enhanced behavior: just ignore the list of users and do not send it to the cluster
	Entry("inference invalid servingPort auth - public with authorizedUsers", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_INFERENCE),
		specFname:     "serving_port_auth_public_with_auth_users",
		expectedFname: "serving_port_auth_public_with_auth_users",
	}),

	// enhanced behavior: just ignore the list of groups and do not send it to the cluster
	Entry("inference invalid servingPort auth - public with authorizedGroups", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_INFERENCE),
		specFname:     "serving_port_auth_public_with_auth_groups",
		expectedFname: "serving_port_auth_public_with_auth_groups",
	}),

	// enhanced behavior: just ignore the list of users and do not send it to the cluster
	Entry("inference invalid servingPort auth - authenticatedUsers with authorizedUsers", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_INFERENCE),
		specFname:     "serving_port_auth_authenticated_users_with_auth_users",
		expectedFname: "serving_port_auth_authenticated_users_with_auth_users",
	}),

	// enhanced behavior: just ignore the list of groups and do not send it to the cluster
	Entry("inference invalid servingPort auth - authenticatedUsers with authorizedGroups", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_INFERENCE),
		specFname:     "serving_port_auth_authenticated_users_with_auth_groups",
		expectedFname: "serving_port_auth_authenticated_users_with_auth_groups",
	}),

	Entry("inference invalid servingPort auth - authorizedUsersOrGroups without authorizedUsers or authorizedGroups", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_INFERENCE),
		specFname:     "serving_port_auth_invalid_authorized_users_or_groups_without_list",
		expectedError: "servingPort: authorizedUsers or authorizedGroups should be set when authorizationType is authorizedUsersOrGroups",
	}),

	Entry("inference invalid servingPort auth - authorizedUsersOrGroups with both authorizedUsers and authorizedGroups", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_INFERENCE),
		specFname:     "serving_port_auth_invalid_authorized_users_or_groups_with_both_lists",
		expectedError: "servingPort: authorizedUsers and authorizedGroups are mutually exclusive",
	}),

	Entry("inference basic ServingConfiguration", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_INFERENCE),
		specFname:     "serving_config_basic",
		expectedFname: "serving_config_basic_inf",
	}),

	Entry("inference autoscaling", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_INFERENCE),
		specFname:     "autoscaling",
		expectedFname: "autoscaling_inf",
	}),

	Entry("inference autoscaling vs min", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_INFERENCE),
		specFname:     "autoscaling_active_vs_min",
		expectedError: "autoscaling: activationReplicas (1) must be no less than minReplicas (2)",
	}),

	Entry("inference autoscaling vs max", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_INFERENCE),
		specFname:     "autoscaling_active_vs_max",
		expectedError: "toscaling: activationReplicas (5) must be no greater than maxReplicas (4)",
	}),

	Entry("inference autoscaling - only minReplicas is specified and is 0", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_INFERENCE),
		specFname:     "autoscaling_only_min_0",
		expectedFname: "autoscaling_only_min_0_inf",
	}),

	Entry("inference autoscaling - only minReplicas is specified and is 1", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_INFERENCE),
		specFname:     "autoscaling_only_min_1",
		expectedFname: "autoscaling_only_min_1_inf",
	}),

	Entry("inference autoscaling - only minReplicas is specified and is 2", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_INFERENCE),
		specFname:     "autoscaling_only_min_2",
		expectedFname: "autoscaling_only_min_2_inf",
	}),

	Entry("inference autoscaling - only maxReplicas is specified and is 1", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_INFERENCE),
		specFname:     "autoscaling_only_max_1",
		expectedFname: "autoscaling_only_max_1_inf",
	}),

	Entry("inference autoscaling - only maxReplicas is specified and is 2", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_INFERENCE),
		specFname:     "autoscaling_only_max_2",
		expectedFname: "autoscaling_only_max_2_inf",
	}),

	Entry("inference autoscaling scale-to-zero", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_INFERENCE),
		specFname:     "autoscaling_scale_to_zero",
		expectedFname: "autoscaling_scale_to_zero_inf",
	}),

	Entry("inference autoscaling scale-to-zero without metric", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_INFERENCE),
		specFname:     "autoscaling_scale_to_zero_no_metric",
		expectedFname: "autoscaling_scale_to_zero_no_metric_inf",
	}),

	Entry("inference autoscaling advanced", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_INFERENCE),
		specFname:     "autoscaling_advanced",
		expectedFname: "autoscaling_advanced_inf",
	}),

	Entry("inference autoscaling advanced with scale-to-zero", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_INFERENCE),
		specFname:     "autoscaling_advanced_scale_to_zero",
		expectedFname: "autoscaling_advanced_scale_to_zero_inf",
	}),

	Entry("inference invalid autoscaling - maxReplicas < minReplicas", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_INFERENCE),
		specFname:     "autoscaling_invalid_min_max_mismatch",
		expectedError: "autoscaling: maxReplicas (2) must be no less than minReplicas (5)",
	}),

	Entry("inference invalid autoscaling - initialReplicas < minReplicas", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_INFERENCE),
		specFname:     "autoscaling_invalid_initial_replicas_too_small",
		expectedError: "autoscaling: initialReplicas (1) must be no less than minReplicas (2)",
	}),

	Entry("inference invalid autoscaling - initialReplicas > maxReplicas", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_INFERENCE),
		specFname:     "autoscaling_invalid_initial_replicas_too_big",
		expectedError: "autoscaling: initialReplicas (5) must be no greater than maxReplicas (4)",
	}),

	Entry("inference invalid autoscaling - metric without metricThreshold", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_INFERENCE),
		specFname:     "autoscaling_invalid_metric_without_threshold",
		expectedError: "autoscaling: metricThreshold is mandatory when metric is provided",
	}),

	Entry("inference invalid autoscaling - metricThreshold without metric", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_INFERENCE),
		specFname:     "autoscaling_invalid_threshold_without_metric",
		expectedError: "autoscaling: metric is mandatory when metricThreshold is provided",
	}),

	Entry("inference invalid autoscaling - minReplicas < maxReplicas but no metric", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_INFERENCE),
		specFname:     "autoscaling_invalid_min_max_range_no_metric",
		expectedError: "autoscaling: metric is mandatory when minReplicas < maxReplicas",
	}),

	Entry("inference invalid autoscaling - minReplicas < maxReplicas but no metric", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_INFERENCE),
		specFname:     "autoscaling_invalid_min_max_range_min_0_no_metric",
		expectedError: "autoscaling: metric is mandatory when minReplicas < maxReplicas",
	}),

	Entry("inference invalid autoscaling - zero minReplicas with latency metric", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_INFERENCE),
		specFname:     "autoscaling_invalid_scale_to_zero_latency_metric",
		expectedError: "autoscaling: metric 'latency' is not supported when minReplicas is 0",
	}),

	Entry("inference invalid autoscaling - zero minReplicas with custom metric", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_INFERENCE),
		specFname:     "autoscaling_invalid_scale_to_zero_custom_metric",
		expectedError: "autoscaling: metric 'my_custom_metric' is not supported when minReplicas is 0",
	}),

	Entry("inference invalid autoscaling - metricThresholdPercentage with latency metric", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_INFERENCE),
		specFname:     "autoscaling_invalid_metric_threshold_percentage_with_latency_metric",
		expectedError: "autoscaling: metricThresholdPercentage is only supported when metric is either throughput or concurrency",
	}),

	Entry("inference invalid autoscaling - metricThresholdPercentage with custom metric", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_INFERENCE),
		specFname:     "autoscaling_invalid_metric_threshold_percentage_with_custom_metric",
		expectedError: "autoscaling: metricThresholdPercentage is only supported when metric is either throughput or concurrency",
	}),

	Entry("inference invalid autoscaling - scaleToZeroRetentionSeconds with positive minReplicas", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_INFERENCE),
		specFname:     "autoscaling_invalid_scale_to_zero_min",
		expectedError: "autoscaling: scaleToZeroRetentionSeconds is only supported when minReplicas is 0",
	}),

	Entry("inference pvc - ExistingPvc=true", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_INFERENCE),
		specFname:     "pvc_existing",
		expectedFname: "pvc_existing_inf",
	}),

	Entry("inference invalid pvc - ExistingPvc=false Ephemeral not specified", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_INFERENCE),
		specFname:     "pvc_not_existing",
		expectedFname: "pvc_not_existing_inf",
	}),

	Entry("inference invalid pvc - ExistingPvc=false Ephemeral=false", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_INFERENCE),
		specFname:     "pvc_not_existing_ephemeral_false",
		expectedFname: "pvc_not_existing_ephemeral_false_inf",
	}),

	Entry("inference invalid pvc - ExistingPvc=false Ephemeral=true", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_INFERENCE),
		specFname:     "pvc_not_existing_ephemeral_true",
		expectedError: "pvc: Ephemeral must not be set 'true' for inference workloads",
	}),
}
