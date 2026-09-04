/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dinf_to_lws

import (
	"context"
	"fmt"
	"time"

	"github.com/ghodss/yaml"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	clusterApi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/testutils"
	workloadsApi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
)

// ComposerTest declares a single distributed-inference composer scenario,
// mirroring the kubeflow `ComposerTest` shape so per-entry knobs
// (feature-gated behavior, cluster config, spec mutations) are added by
// extending the struct rather than forking the runner.
//
//   - input           : fixture name under testing/inputs (no .yaml)
//   - expected        : golden name under testing/expected (no .yaml);
//     defaults to `input` when empty. Use a distinct value when the same
//     fixture is composed under multiple feature flags.
//   - expectedError   : substring the unmarshal OR compose error must
//     contain. When set, no golden is compared.
//   - overrideOptions : mutates the default FieldsComposerOptions before
//     composing. Used to exercise feature-gated behavior against the same
//     fixture (e.g. the pre-3.0.0 legacy env-var fallback for gpu
//     memory/portion limits).
//   - expectedCode   : when non-zero, asserts the compose error's Code
//     field. Ignored on unmarshal errors (those don't carry a Code).
//   - clusterConfig   : replaces the default legacy typed cluster storage
//     config. Used by MNNVL entries to inject cluster-level flags.
//   - workloadID      : overrides the default workload ID (used to
//     demonstrate LWS-name derivation from the ID suffix).
//   - lwsName         : when non-empty, calls WithLeaderWorkerSetName
//     instead of WithWorkloadID (used to demonstrate the explicit-name
//     override on update).
//   - mnnvlLabel      : when non-empty, injects a `run.ai/MNNVL` label onto
//     the leader spec so the composer picks it up (used by MNNVL entries).
type ComposerTest struct {
	input           string
	expected        string
	expectedError   string
	expectedCode    int32
	overrideOptions func(*composer.FieldsComposerOptions)
	clusterConfig   composer.ClusterStorageConfig
	workloadID      string
	lwsName         string
	mnnvlLabel      string
}

// disableGpuMemoryLimitAnnotations turns off the >= 3.0.0 CUDA/KAI
// gpu-memory annotations so the composer falls back to the legacy
// RUNAI_GPU_MEMORY_LIMIT / RUNAI_GPU_PORTION_LIMIT env vars on each
// per-role container.
func disableGpuMemoryLimitAnnotations(o *composer.FieldsComposerOptions) {
	o.GpuMemoryLimitAnnotationsSupported = false
}

// ComposerTests is the master table for the distributed-inference composer.
// Topic-specific entry lists (currently just `mnnvlTests`) are appended in
// so every fixture flows through one runner and one DescribeTable, matching
// the pattern used by distributed_to_kubeflow.
var ComposerTests = append([]TableEntry{
	// Happy path -- one golden per fixture.
	Entry("all fields", ComposerTest{input: "dinfw_all_fields"}),
	Entry("multiple node pools", ComposerTest{input: "dinfw_node_pools_multiple"}),
	Entry("preemptibility", ComposerTest{input: "dinfw_preemptibility"}),
	Entry("semi preemptibility", ComposerTest{input: "dinfw_semi_preemptibility"}),
	Entry("multiple replicas", ComposerTest{input: "dinfw_replicas"}),
	Entry("one replica", ComposerTest{input: "dinfw_replicas_one"}),
	Entry("zero replicas", ComposerTest{input: "dinfw_replicas_zero"}),
	Entry("restart policy", ComposerTest{input: "dinfw_restart_policy"}),
	Entry("serving port authenticated users", ComposerTest{input: "dinfw_serving_port_authenticated_users"}),
	Entry("serving port authenticated with groups", ComposerTest{input: "dinfw_serving_port_authenticated_with_groups"}),
	Entry("serving port authenticated with users", ComposerTest{input: "dinfw_serving_port_authenticated_with_users"}),
	Entry("serving port authorized groups", ComposerTest{input: "dinfw_serving_port_authorized_groups"}),
	Entry("serving port authorized groups with users", ComposerTest{input: "dinfw_serving_port_authorized_groups_with_users"}),
	Entry("serving port authorized users", ComposerTest{input: "dinfw_serving_port_authorized_users"}),
	Entry("serving port authorized users with groups", ComposerTest{input: "dinfw_serving_port_authorized_users_with_groups"}),
	Entry("serving port basic", ComposerTest{input: "dinfw_serving_port_basic"}),
	Entry("serving port custom port", ComposerTest{input: "dinfw_serving_port_custom_port"}),
	Entry("serving port expose externally false", ComposerTest{input: "dinfw_serving_port_expose_externally_false"}),
	Entry("serving port expose externally true", ComposerTest{input: "dinfw_serving_port_expose_externally_true"}),
	Entry("serving port expose externally true by default", ComposerTest{input: "dinfw_serving_port_expose_externally_true_by_default"}),
	Entry("serving port exposed url default expose", ComposerTest{input: "dinfw_serving_port_exposed_url_default_expose"}),
	Entry("serving port exposed url with expose true", ComposerTest{input: "dinfw_serving_port_exposed_url_with_expose_true"}),
	Entry("serving port multiple groups", ComposerTest{input: "dinfw_serving_port_multiple_groups"}),
	Entry("serving port multiple users", ComposerTest{input: "dinfw_serving_port_multiple_users"}),
	Entry("serving port no protocol", ComposerTest{input: "dinfw_serving_port_no_protocol"}),
	Entry("serving port public explicit", ComposerTest{input: "dinfw_serving_port_public_explicit"}),
	Entry("serving port public with groups", ComposerTest{input: "dinfw_serving_port_public_with_groups"}),
	Entry("serving port public with users", ComposerTest{input: "dinfw_serving_port_public_with_users"}),
	Entry("serving port single group", ComposerTest{input: "dinfw_serving_port_single_group"}),
	Entry("serving port single user", ComposerTest{input: "dinfw_serving_port_single_user"}),
	Entry("storage multiple types", ComposerTest{input: "dinfw_storage_multiple_types"}),
	Entry("worker spec empty workers zero", ComposerTest{input: "dinfw_worker_spec_empty_workers_zero"}),
	Entry("worker spec not provided workers positive", ComposerTest{input: "dinfw_worker_spec_not_provided_workers_positive"}),
	Entry("worker spec valid config workers positive", ComposerTest{input: "dinfw_worker_spec_valid_config_workers_positive"}),
	Entry("workers not set no worker spec", ComposerTest{input: "dinfw_workers_not_set_no_worker_spec"}),
	Entry("one basic distributed worker", ComposerTest{input: "dinfw_workers_one_basic_distributed"}),
	Entry("two basic distributed workers", ComposerTest{input: "dinfw_workers_two_basic_distributed"}),
	Entry("zero workers leader only", ComposerTest{input: "dinfw_workers_zero_leader_only"}),

	// GPU memory & portion request+limit against both cluster-version
	// paths. In >= 3.0.0 the composer emits
	// nvidia.com/container.<role>.gpu-memory.limit and
	// kai.scheduler/container.<role>.gpu-memory.portion.limit on BOTH the
	// leader and worker templates (with per-role values). The `_legacy_env`
	// entries re-compose the same fixture with the annotation capability
	// disabled to lock in the pre-3.0.0 RUNAI_GPU_*_LIMIT env-var fallback.
	Entry("gpu memory request and limit", ComposerTest{
		input: "dinfw_gpu_memory_req_and_limit",
	}),
	Entry("gpu memory request and limit (legacy env, cluster < 3.0)", ComposerTest{
		input:           "dinfw_gpu_memory_req_and_limit",
		expected:        "dinfw_gpu_memory_req_and_limit_legacy_env",
		overrideOptions: disableGpuMemoryLimitAnnotations,
	}),
	Entry("gpu portion request and limit", ComposerTest{
		input: "dinfw_gpu_portion_req_and_limit",
	}),
	Entry("gpu portion request and limit (legacy env, cluster < 3.0)", ComposerTest{
		input:           "dinfw_gpu_portion_req_and_limit",
		expected:        "dinfw_gpu_portion_req_and_limit_legacy_env",
		overrideOptions: disableGpuMemoryLimitAnnotations,
	}),

	// Fold-ins from what used to be scattered `go test` unit specs. Each
	// of these previously constructed a spec programmatically, composed
	// it, and asserted a handful of scalar fields. Captured now as
	// input/expected pairs so the whole composed shape (not just the
	// author's chosen assertions) is locked in.

	// Topology-driven scalars (Size, Replicas, StartupPolicy,
	// RestartPolicy) plus independent leader/worker command+image.
	Entry("topology and independent role values", ComposerTest{
		input: "dinfw_topology_and_role_values",
	}),

	// ${RUNAI_PROJECT}/${RUNAI_JOB_NAME}/${RUNAI_UID}/${RUNAI_GID}
	// expansion in args (uid/gid come from the leader security block).
	Entry("expands run:ai template vars in command and args", ComposerTest{
		input: "dinfw_runai_template_vars",
	}),

	// When the worker spec is absent the composer copies the whole leader
	// spec (image + args + ...) into the worker template.
	Entry("copies whole leader when worker spec is absent", ComposerTest{
		input: "dinfw_leader_only_worker_absent",
	}),

	// When the worker spec is present it does NOT inherit individual
	// leader fields -- the worker's own image/args stand alone.
	Entry("does not inherit individual leader fields when worker exists", ComposerTest{
		input: "dinfw_no_inherit_when_worker_exists",
	}),

	// An empty worker spec (`worker: {}`) is allowed when workers=0; the
	// worker template still gets composed from the leader.
	Entry("allows empty worker when workers is zero", ComposerTest{
		input: "dinfw_empty_worker_zero_workers",
	}),

	// LWS-name resolution:
	// (a) derived from the workloadID suffix when only a workloadID is
	//     set -- LWS.Name = "<workload>-<last-4-of-id>"; and
	// (b) reused verbatim when an explicit LWS name is set (this is what
	//     `.WithLeaderWorkerSetName(...)` does on update).
	// Both entries share the same input fixture so the goldens only
	// differ in the resulting LWS.Name.
	Entry("lws name derived from workload id suffix", ComposerTest{
		input:      "dinfw_lws_name",
		expected:   "dinfw_lws_name_from_workload_id",
		workloadID: "workload-1234abcd",
	}),
	Entry("lws name reused from explicit override on update", ComposerTest{
		input:    "dinfw_lws_name",
		expected: "dinfw_lws_name_explicit_override",
		lwsName:  "test-lws-original",
	}),

	// Fat placement + per-role compute/security/nodeType/labels/annotations
	// test: the previous unit test asserted ~20 scalar fields; the golden
	// captures the composed shape end-to-end.
	Entry("shared placement and per-role entity fields", ComposerTest{
		input: "dinfw_shared_placement_and_entity_fields",
	}),

	// The composer-generated replica labels (app, release, workloadName)
	// win over user-provided values on the same keys.
	Entry("generated replica labels override user labels", ComposerTest{
		input: "dinfw_generated_replica_labels_override",
	}),

	// authorizedUsers serving port + MNNVL Required label -> oauth2-proxy
	// sidecar, username-loader init container, auth volume, networking
	// endpoint, and MNNVL mode all in one shot.
	Entry("authenticated users serving port with mnnvl required", ComposerTest{
		input:      "dinfw_authenticated_users_with_mnnvl",
		workloadID: "workload-abcd",
	}),

	// Public + exposeExternally -> external-access networking endpoint,
	// no auth sidecar. Uses the default (no-oauth) cluster storage config
	// to match the original unit test's setup and keep the golden focused
	// on public-serving semantics.
	Entry("public external serving port", ComposerTest{
		input:         "dinfw_public_external_serving_port",
		clusterConfig: composer.DefaultClusterStorageConfig(),
	}),

	// Mixed PVC storage: leader with an existing PVC (referenced by
	// claimName) + worker with a new PVC (composer generates the
	// PersistentVolumeClaim spec under target.Config.Pvcs).
	Entry("existing and new pvcs across roles", ComposerTest{
		input: "dinfw_existing_and_new_pvcs",
	}),

	// Error cases -- some fail at YAML->typed unmarshal (invalid enum
	// values, missing required fields), some fail at compose validation.
	// The runner asserts either error message contains `expectedError`.
	Entry("no leader", ComposerTest{
		input:         "dinfw_no_leader",
		expectedError: "leader spec is required",
		expectedCode:  400,
	}),
	Entry("no leader image", ComposerTest{
		input:         "dinfw_no_leader_image",
		expectedError: "image",
	}),
	Entry("no serving port", ComposerTest{
		input:         "dinfw_no_serving_port",
		expectedError: "servingPort is required",
		expectedCode:  400,
	}),
	Entry("no worker image", ComposerTest{
		input:         "dinfw_no_worker_image",
		expectedError: "image",
	}),
	Entry("restart policy case mismatch", ComposerTest{
		input:         "dinfw_restart_policy_case_mismatch",
		expectedError: "not a valid DistributedInferenceRestartPolicy",
	}),
	Entry("serving port authorized groups without list", ComposerTest{
		input:         "dinfw_serving_port_authorized_groups_without_list",
		expectedError: "authorizedGroups",
	}),
	Entry("serving port authorized users without list", ComposerTest{
		input:         "dinfw_serving_port_authorized_users_without_list",
		expectedError: "authorizedUsers",
	}),
	Entry("serving port empty authorized groups", ComposerTest{
		input:         "dinfw_serving_port_empty_authorized_groups",
		expectedError: "authorizedGroups",
	}),
	Entry("serving port empty authorized users", ComposerTest{
		input:         "dinfw_serving_port_empty_authorized_users",
		expectedError: "authorizedUsers",
	}),
	Entry("serving port exposed url with expose false", ComposerTest{
		input:         "dinfw_serving_port_exposed_url_with_expose_false",
		expectedError: "exposedUrl should not be set",
	}),
	Entry("serving port grpc", ComposerTest{
		input:         "dinfw_serving_port_grpc",
		expectedError: "not a valid DistributedInferenceServingPortProtocol",
	}),
	Entry("serving port tcp", ComposerTest{
		input:         "dinfw_serving_port_tcp",
		expectedError: "not a valid DistributedInferenceServingPortProtocol",
	}),
	Entry("serving port unsupported auth type", ComposerTest{
		input:         "dinfw_serving_port_unsupported_auth_type",
		expectedError: "not a valid DistributedInferenceServingPortAccessAuthorizationTypeEnum",
	}),
	Entry("worker spec provided workers not set", ComposerTest{
		input:         "dinfw_worker_spec_provided_workers_not_set",
		expectedError: "worker spec requires workers",
	}),
	Entry("worker spec provided workers zero", ComposerTest{
		input:         "dinfw_worker_spec_provided_workers_zero",
		expectedError: "worker spec requires workers",
		expectedCode:  400,
	}),
	Entry("workers not set with worker spec provided", ComposerTest{
		input:         "dinfw_workers_not_set_worker_spec_provided",
		expectedError: "worker spec requires workers",
	}),
	Entry("workers set zero with worker spec provided", ComposerTest{
		input:         "dinfw_workers_set_zero_worker_spec_provided",
		expectedError: "worker spec requires workers",
	}),
}, mnnvlTests...)

var _ = Describe("DistributedInferenceToLWSComposer", func() {
	DescribeTable("composes each input fixture into an LWS target",
		runComposer, ComposerTests)

	Context("test some enums", func() {
		newLeader := func() *workloadsApi.DistributedInferenceLeaderWorkerSpec {
			leader := workloadsApi.NewDistributedInferenceLeaderWorkerSpec()
			leader.SetImage("leader:latest")
			return leader
		}

		newComposer := func() *DistributedInferenceToLWSComposer {
			return NewDistributedInferenceToLWSComposer(
				"dist-inf", "project-a", "namespace-a", "user-a",
				nil, nil, composer.FieldsComposerOptions{}, nil,
			)
		}

		It("rejects an unsupported serving port protocol built programmatically", func() {
			spec := workloadsApi.NewDistributedInferenceSpecSpec()
			spec.SetLeader(*newLeader())
			servingPort := workloadsApi.NewDistributedInferenceServingPort()
			servingPort.SetPort(8080)
			servingPort.SetProtocol(workloadsApi.DistributedInferenceServingPortProtocol("grpc"))
			spec.SetServingPort(*servingPort)

			composeErr := newComposer().Compose(context.Background(), spec, NewMinimalLeaderWorkerSetTarget())()

			Expect(composeErr).NotTo(BeNil())
			Expect(composeErr.Code).To(Equal(int32(400)))
			Expect(composeErr.Message).To(ContainSubstring("unsupported protocol"))
		})

		DescribeTable("rejects invalid LWS policies built programmatically",
			func(mutate func(*workloadsApi.DistributedInferenceSpecSpec), expectedSubstring string) {
				spec := workloadsApi.NewDistributedInferenceSpecSpec()
				spec.SetLeader(*newLeader())
				servingPort := workloadsApi.NewDistributedInferenceServingPort()
				servingPort.SetPort(8080)
				servingPort.SetExposeExternally(false)
				spec.SetServingPort(*servingPort)
				mutate(spec)

				composeErr := newComposer().Compose(context.Background(), spec, NewMinimalLeaderWorkerSetTarget())()

				Expect(composeErr).NotTo(BeNil())
				Expect(composeErr.Code).To(Equal(int32(400)))
				Expect(composeErr.Message).To(ContainSubstring(expectedSubstring))
			},
			Entry("invalid startup policy",
				func(spec *workloadsApi.DistributedInferenceSpecSpec) {
					spec.SetStartupPolicy(workloadsApi.DistributedInferenceStartupPolicy("invalid"))
				},
				"invalid startup policy",
			),
			Entry("invalid restart policy",
				func(spec *workloadsApi.DistributedInferenceSpecSpec) {
					spec.SetRestartPolicy(workloadsApi.DistributedInferenceRestartPolicy("invalid"))
				},
				"invalid restart policy",
			),
		)

	})
})

func runComposer(tc ComposerTest) {
	specYaml, err := testutils.LoadInputFile(tc.input, ".")
	Expect(err).To(BeNil(), "read fixture "+tc.input)

	spec := workloadsApi.NewDistributedInferenceSpecSpec()
	unmarshalErr := yaml.Unmarshal(specYaml, spec)
	if unmarshalErr != nil {
		Expect(tc.expectedError).NotTo(BeEmpty(),
			fmt.Sprintf("unexpected typed input validation error for %s: %v", tc.input, unmarshalErr))
		Expect(unmarshalErr.Error()).To(ContainSubstring(tc.expectedError))
		return
	}

	if tc.mnnvlLabel != "" {
		leader := spec.Leader.Get()
		Expect(leader).NotTo(BeNil(), "mnnvlLabel requires a leader spec in the fixture")
		leader.Labels = append(leader.Labels, workloadsApi.Label{
			Name:  *workloadsApi.NewNullableString(convutils.Ref("run.ai/MNNVL")),
			Value: *workloadsApi.NewNullableString(convutils.Ref(tc.mnnvlLabel)),
		})
	}

	options := composer.FieldsComposerOptions{
		EmptyDirAsVolumeSupported:          true,
		FlexiblePriorityClassesSupported:   true,
		GpuMemoryLimitAnnotationsSupported: true,
		NodeAffinitySupported:              true,
		PreemptibilitySupported:            true,
		UserCredentialsSupported:           true,
		VolumeDefaultModeSupported:         true,
		WorkloadV2NetworkingSupported:      true,
	}
	if tc.overrideOptions != nil {
		tc.overrideOptions(&options)
	}

	clusterConfig := tc.clusterConfig
	if clusterConfig == nil {
		clusterConfig = legacyTypedTestClusterConfig()
	}

	dinfComposer := NewDistributedInferenceToLWSComposer(
		"test-lws", "team-a", "runai-team-a", "tester",
		nil, nil, options, legacyTypedTestNodePools(),
	).WithClusterStorageConfig(clusterConfig)

	// LWS name resolution: an explicit lwsName wins over any workloadID
	// (the composer's WithLeaderWorkerSetName is used on update to reuse
	// the existing resource name); otherwise the LWS name is derived from
	// the workloadID's suffix.
	if tc.lwsName != "" {
		dinfComposer.WithLeaderWorkerSetName(tc.lwsName)
	} else {
		workloadID := tc.workloadID
		if workloadID == "" {
			workloadID = "00000000-0000-0000-0000-000000001234"
		}
		dinfComposer.WithWorkloadID(workloadID)
	}

	target := NewMinimalLeaderWorkerSetTarget()
	composeErr := dinfComposer.Compose(context.Background(), spec, target)()

	if tc.expectedError != "" {
		Expect(composeErr).NotTo(BeNil(), "expected compose error for "+tc.input)
		Expect(composeErr.Message).To(ContainSubstring(tc.expectedError))
		if tc.expectedCode != 0 {
			Expect(composeErr.Code).To(Equal(tc.expectedCode))
		}
		return
	}
	Expect(composeErr).To(BeNil(), "unexpected compose error for "+tc.input)

	normalizeGeneratedPVCNames(target)
	actualYaml, marshalErr := yaml.Marshal(target)
	Expect(marshalErr).To(BeNil())

	expectedFname := tc.expected
	if expectedFname == "" {
		expectedFname = tc.input
	}
	expectedYaml := testutils.LoadExpectedFile(expectedFname, actualYaml, ".")
	assert.Equal(GinkgoT(), string(expectedYaml), string(actualYaml), expectedFname)
}

func normalizeGeneratedPVCNames(target *LeaderWorkerSetTarget) {
	names := make(map[string]string, len(target.Config.Pvcs))
	for index := range target.Config.Pvcs {
		oldName := target.Config.Pvcs[index].Metadata.Name
		stableName := fmt.Sprintf("generated-pvc-%d", index+1)
		names[oldName] = stableName
		target.Config.Pvcs[index].Metadata.Name = stableName
	}
	normalizeTemplatePVCNames(&target.LeaderWorkerSet.Spec.LeaderWorkerTemplate.LeaderTemplate.Spec, names)
	normalizeTemplatePVCNames(&target.LeaderWorkerSet.Spec.LeaderWorkerTemplate.WorkerTemplate.Spec, names)
}

func normalizeTemplatePVCNames(spec *corev1.PodSpec, names map[string]string) {
	for index := range spec.Volumes {
		claim := spec.Volumes[index].PersistentVolumeClaim
		if claim == nil {
			continue
		}
		if stableName, found := names[claim.ClaimName]; found {
			claim.ClaimName = stableName
		}
	}
}

func legacyTypedTestClusterConfig() composer.ClusterStorageConfig {
	reported := clusterApi.ClusterReportedConfig{}
	reported.SetRunaiOauth2ProxyImage("test.registry/oauth2-proxy:latest")
	reported.SetRunaiOauth2ProxyUsernameLoaderImage("test.registry/username-loader:latest")
	reported.SetRunaiOauth2ProxyWorkloadAuthSecretName("runai-auth-secret")
	reported.SetRunaiOauth2ProxySessionStoreType("redis")
	reported.SetRunaiOauth2ProxyRedisFQDN("redis.runai-backend.svc.cluster.local")
	reported.SetRunaiOauth2ProxyOidcIssuerUrl("https://test.run.ai/auth/realms/runai")

	status := clusterApi.ClusterDisplayedStatus{}
	status.SetConfig(reported)
	cluster := clusterApi.NewDisplayedCluster("cluster-a", 1, "cluster-a", time.Time{})
	cluster.Status.Set(&status)
	cluster.SetDomain("https://test.run.ai")
	return composer.NewClusterStorageConfigFromCluster(cluster)
}

func legacyTypedTestNodePools() []clusterApi.Nodepool {
	names := []string{"gpu-pool-a", "gpu-pool-b", "cpu-pool"}
	pools := make([]clusterApi.Nodepool, 0, len(names))
	for _, name := range names {
		pools = append(pools, clusterApi.Nodepool{Name: name, LabelKey: "run.ai/node-pool", LabelValue: name})
	}
	return pools
}
