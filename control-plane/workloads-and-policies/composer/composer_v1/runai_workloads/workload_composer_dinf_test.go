package runai_workloads

import (
	. "github.com/onsi/ginkgo/v2"
	workloadsApi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
)

var distributedInferenceTests = []TableEntry{

	// positive tests
	Entry("distributed inference all fields", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_DISTRIBUTED_INFERENCE),
		specFname:     "dinfw_all_fields",
		expectedFname: "dinfw_all_fields_expected",
	}),

	Entry("distributed inference preemptibility", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_DISTRIBUTED_INFERENCE),
		specFname:     "dinfw_preemptibility",
		expectedFname: "dinfw_preemptibility_expected",
	}),

	Entry("distributed inference servingPort basic", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_DISTRIBUTED_INFERENCE),
		specFname:     "dinfw_serving_port_basic",
		expectedFname: "dinfw_serving_port_basic_expected",
	}),

	Entry("distributed inference servingPort authenticated users", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_DISTRIBUTED_INFERENCE),
		specFname:     "dinfw_serving_port_authenticated_users",
		expectedFname: "dinfw_serving_port_authenticated_users_expected",
	}),

	Entry("distributed inference servingPort authorized users", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_DISTRIBUTED_INFERENCE),
		specFname:     "dinfw_serving_port_authorized_users",
		expectedFname: "dinfw_serving_port_authorized_users_expected",
	}),

	Entry("distributed inference servingPort authorized groups", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_DISTRIBUTED_INFERENCE),
		specFname:     "dinfw_serving_port_authorized_groups",
		expectedFname: "dinfw_serving_port_authorized_groups_expected",
	}),

	Entry("distributed inference servingPort without protocol", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_DISTRIBUTED_INFERENCE),
		specFname:     "dinfw_serving_port_no_protocol",
		expectedFname: "dinfw_serving_port_no_protocol_expected",
	}),

	Entry("distributed inference servingPort with different port number", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_DISTRIBUTED_INFERENCE),
		specFname:     "dinfw_serving_port_custom_port",
		expectedFname: "dinfw_serving_port_custom_port_expected",
	}),

	Entry("distributed inference servingPort with exposeExternally true by default", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_DISTRIBUTED_INFERENCE),
		specFname:     "dinfw_serving_port_expose_externally_true_by_default",
		expectedFname: "dinfw_serving_port_expose_externally_true_by_default_expected",
	}),

	Entry("distributed inference servingPort with exposeExternally true", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_DISTRIBUTED_INFERENCE),
		specFname:     "dinfw_serving_port_expose_externally_true",
		expectedFname: "dinfw_serving_port_expose_externally_true_expected",
	}),

	Entry("distributed inference servingPort with exposeExternally false", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_DISTRIBUTED_INFERENCE),
		specFname:     "dinfw_serving_port_expose_externally_false",
		expectedFname: "dinfw_serving_port_expose_externally_false_expected",
	}),

	Entry("distributed inference servingPort with multiple authorized users", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_DISTRIBUTED_INFERENCE),
		specFname:     "dinfw_serving_port_multiple_users",
		expectedFname: "dinfw_serving_port_multiple_users_expected",
	}),

	Entry("distributed inference servingPort with multiple authorized groups", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_DISTRIBUTED_INFERENCE),
		specFname:     "dinfw_serving_port_multiple_groups",
		expectedFname: "dinfw_serving_port_multiple_groups_expected",
	}),

	Entry("distributed inference servingPort with single authorized user", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_DISTRIBUTED_INFERENCE),
		specFname:     "dinfw_serving_port_single_user",
		expectedFname: "dinfw_serving_port_single_user_expected",
	}),

	Entry("distributed inference servingPort with single authorized group", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_DISTRIBUTED_INFERENCE),
		specFname:     "dinfw_serving_port_single_group",
		expectedFname: "dinfw_serving_port_single_group_expected",
	}),

	Entry("distributed inference servingPort public explicit", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_DISTRIBUTED_INFERENCE),
		specFname:     "dinfw_serving_port_public_explicit",
		expectedFname: "dinfw_serving_port_public_explicit_expected",
	}),

	Entry("distributed inference servingPort with exposedUrl and exposeExternally true", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_DISTRIBUTED_INFERENCE),
		specFname:     "dinfw_serving_port_exposed_url_with_expose_true",
		expectedFname: "dinfw_serving_port_exposed_url_with_expose_true_expected",
	}),

	Entry("distributed inference servingPort with exposedUrl and default exposeExternally", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_DISTRIBUTED_INFERENCE),
		specFname:     "dinfw_serving_port_exposed_url_default_expose",
		expectedFname: "dinfw_serving_port_exposed_url_default_expose_expected",
	}),

	// === WORKERS FIELD VALIDATION TESTS - POSITIVE ===

	Entry("distributed inference workers = 0 (leader only mode)", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_DISTRIBUTED_INFERENCE),
		specFname:     "dinfw_workers_zero_leader_only",
		expectedFname: "dinfw_workers_zero_leader_only_expected",
	}),

	Entry("distributed inference workers = 1 (basic distributed case)", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_DISTRIBUTED_INFERENCE),
		specFname:     "dinfw_workers_one_basic_distributed",
		expectedFname: "dinfw_workers_one_basic_distributed_expected",
	}),

	Entry("distributed inference workers = 2 (basic distributed case)", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_DISTRIBUTED_INFERENCE),
		specFname:     "dinfw_workers_two_basic_distributed",
		expectedFname: "dinfw_workers_two_basic_distributed_expected",
	}),

	Entry("distributed inference workers not set and worker spec not provided (default to 0)", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_DISTRIBUTED_INFERENCE),
		specFname:     "dinfw_workers_not_set_no_worker_spec",
		expectedFname: "dinfw_workers_not_set_no_worker_spec_expected",
	}),

	// === WORKER SPEC VALIDATION TESTS - POSITIVE ===

	Entry("distributed inference worker spec empty with workers=0 (valid - leader only)", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_DISTRIBUTED_INFERENCE),
		specFname:     "dinfw_worker_spec_empty_workers_zero",
		expectedFname: "dinfw_worker_spec_empty_workers_zero_expected",
	}),

	Entry("distributed inference worker spec with valid configuration and workers>0", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_DISTRIBUTED_INFERENCE),
		specFname:     "dinfw_worker_spec_valid_config_workers_positive",
		expectedFname: "dinfw_worker_spec_valid_config_workers_positive_expected",
	}),

	// === REPLICAS VALIDATION TESTS - POSITIVE ===

	Entry("distributed inference replicas = 0", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_DISTRIBUTED_INFERENCE),
		specFname:     "dinfw_replicas_zero",
		expectedFname: "dinfw_replicas_zero_expected",
	}),

	Entry("distributed inference replicas = 1", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_DISTRIBUTED_INFERENCE),
		specFname:     "dinfw_replicas_one",
		expectedFname: "dinfw_replicas_one_expected",
	}),

	Entry("distributed inference replicas = 2", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_DISTRIBUTED_INFERENCE),
		specFname:     "dinfw_replicas",
		expectedFname: "dinfw_replicas_expected",
	}),

	// === RESTART POLICY VALIDATION TESTS - POSITIVE ===

	Entry("distributed inference restartPolicy", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_DISTRIBUTED_INFERENCE),
		specFname:     "dinfw_restart_policy",
		expectedFname: "dinfw_restart_policy_expected",
	}),

	// === NODE POOLS VALIDATION TESTS - POSITIVE ===

	Entry("distributed inference nodePools with multiple valid node pools", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_DISTRIBUTED_INFERENCE),
		specFname:     "dinfw_node_pools_multiple",
		expectedFname: "dinfw_node_pools_multiple_expected",
	}),

	// === STORAGE VALIDATION TESTS - POSITIVE ===

	Entry("distributed inference multiple storage types with different mount paths", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_DISTRIBUTED_INFERENCE),
		specFname:     "dinfw_storage_multiple_types",
		expectedFname: "dinfw_storage_multiple_types_expected",
	}),

	// negative tests

	// === WORKERS FIELD VALIDATION TESTS - NEGATIVE ===

	Entry("distributed inference workers not set but worker spec provided", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_DISTRIBUTED_INFERENCE),
		specFname:     "dinfw_workers_not_set_worker_spec_provided",
		expectedError: "worker spec requires workers field to be set to a positive value",
	}),

	Entry("distributed inference workers = 0 but worker spec provided", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_DISTRIBUTED_INFERENCE),
		specFname:     "dinfw_workers_set_zero_worker_spec_provided",
		expectedError: "worker spec requires workers field to be set to a positive value",
	}),

	Entry("distributed inference servingPort unsupported authorization type", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_DISTRIBUTED_INFERENCE),
		specFname:     "dinfw_serving_port_unsupported_auth_type",
		expectedError: "servingPort: unsupported authorizationType 'unsupportedType'",
	}),

	Entry("distributed inference servingPort empty authorized users array", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_DISTRIBUTED_INFERENCE),
		specFname:     "dinfw_serving_port_empty_authorized_users",
		expectedError: "servingPort: authorizedUsers cannot be empty",
	}),

	Entry("distributed inference servingPort empty authorized groups array", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_DISTRIBUTED_INFERENCE),
		specFname:     "dinfw_serving_port_empty_authorized_groups",
		expectedError: "servingPort: authorizedGroups cannot be empty",
	}),

	// enhanced behavior: ignore the users and groups access list, make sure not to download them to the cluster
	Entry("distributed inference servingPort public with authorized users", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_DISTRIBUTED_INFERENCE),
		specFname:     "dinfw_serving_port_public_with_users",
		expectedFname: "dinfw_serving_port_public_with_users_expected",
	}),

	Entry("distributed inference servingPort public with authorized groups", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_DISTRIBUTED_INFERENCE),
		specFname:     "dinfw_serving_port_public_with_groups",
		expectedFname: "dinfw_serving_port_public_with_groups_expected",
	}),

	// enhanced behavior: ignore users and groups, make sure not to download them to the cluster
	Entry("distributed inference servingPort authenticated users with authorized users", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_DISTRIBUTED_INFERENCE),
		specFname:     "dinfw_serving_port_authenticated_with_users",
		expectedFname: "dinfw_serving_port_authenticated_with_users_expected",
	}),

	Entry("distributed inference servingPort authenticated users with authorized groups", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_DISTRIBUTED_INFERENCE),
		specFname:     "dinfw_serving_port_authenticated_with_groups",
		expectedFname: "dinfw_serving_port_authenticated_with_groups_expected",
	}),

	Entry("distributed inference servingPort authorized users without users list", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_DISTRIBUTED_INFERENCE),
		specFname:     "dinfw_serving_port_authorized_users_without_list",
		expectedError: "servingPort: authorizedUsers must be set when authorizationType is authorizedUsers",
	}),

	// enhanced behavior: ignore the group list. as long as users list is provided, use it
	Entry("distributed inference servingPort authorized users with groups list", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_DISTRIBUTED_INFERENCE),
		specFname:     "dinfw_serving_port_authorized_users_with_groups",
		expectedFname: "dinfw_serving_port_authorized_users_with_groups_expected",
	}),

	Entry("distributed inference servingPort authorized groups without groups list", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_DISTRIBUTED_INFERENCE),
		specFname:     "dinfw_serving_port_authorized_groups_without_list",
		expectedError: "servingPort: authorizedGroups must be set when authorizationType is authorizedGroups",
	}),

	// enhanced behavior: ignore the users list. as long as groups list is provided, use it
	Entry("distributed inference servingPort authorized groups with users list", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_DISTRIBUTED_INFERENCE),
		specFname:     "dinfw_serving_port_authorized_groups_with_users",
		expectedFname: "dinfw_serving_port_authorized_groups_with_users_expected",
	}),

	Entry("distributed inference servingPort unsupported protocol TCP", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_DISTRIBUTED_INFERENCE),
		specFname:     "dinfw_serving_port_tcp",
		expectedError: "servingPort: unsupported protocol 'tcp', only 'http' is supported",
	}),

	Entry("distributed inference servingPort unsupported protocol GRPC", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_DISTRIBUTED_INFERENCE),
		specFname:     "dinfw_serving_port_grpc",
		expectedError: "servingPort: unsupported protocol 'grpc', only 'http' is supported",
	}),

	Entry("distributed inference servingPort exposedUrl with exposeExternally false", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_DISTRIBUTED_INFERENCE),
		specFname:     "dinfw_serving_port_exposed_url_with_expose_false",
		expectedError: "servingPort: exposedUrl should not be set when exposeExternally is false",
	}),

	// === WORKER SPEC VALIDATION TESTS - NEGATIVE ===

	Entry("distributed inference worker spec provided but workers=0 (should fail - no workers to configure)", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_DISTRIBUTED_INFERENCE),
		specFname:     "dinfw_worker_spec_provided_workers_zero",
		expectedError: "worker spec requires workers field to be set to a positive value",
	}),

	Entry("distributed inference worker spec provided but workers not set (should fail - need workers count)", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_DISTRIBUTED_INFERENCE),
		specFname:     "dinfw_worker_spec_provided_workers_not_set",
		expectedError: "worker spec requires workers field to be set to a positive value",
	}),

	Entry("distributed inference worker spec not provided with workers>0 (should fail - need worker spec)", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_DISTRIBUTED_INFERENCE),
		specFname:     "dinfw_worker_spec_not_provided_workers_positive",
		expectedError: "worker spec is required when workers is set to a positive value",
	}),

	// === RESTART POLICY VALIDATION TESTS - NEGATIVE ===

	Entry("distributed inference restartPolicy case-sensitive mismatch", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_DISTRIBUTED_INFERENCE),
		specFname:     "dinfw_restart_policy_case_mismatch",
		expectedError: "restartPolicy: unsupported value 'recreategrouponpodrestart'",
	}),

	Entry("distributed inference no servingPort (should fail - required field)", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_DISTRIBUTED_INFERENCE),
		specFname:     "dinfw_no_serving_port",
		expectedError: "servingPort is required for distributed inference workloads",
	}),

	Entry("distributed inference no leader (should fail - required field)", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_DISTRIBUTED_INFERENCE),
		specFname:     "dinfw_no_leader",
		expectedError: "leader spec is required for distributed inference workload",
	}),

	Entry("distributed inference no leader image (should fail - required field)", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_DISTRIBUTED_INFERENCE),
		specFname:     "dinfw_no_leader_image",
		expectedError: "Leader spec: value for 'image' must be provided",
	}),

	Entry("distributed inference no worker image (should fail - required field)", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_DISTRIBUTED_INFERENCE),
		specFname:     "dinfw_no_worker_image",
		expectedError: "Worker spec: value for 'image' must be provided",
	}),
}
