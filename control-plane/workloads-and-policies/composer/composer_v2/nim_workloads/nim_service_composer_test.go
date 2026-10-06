/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package nim_workloads

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"

	openapi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	api "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-manager-client/generated"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	nimv1alpha1 "github.com/NVIDIA/k8s-nim-operator/api/apps/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/testutils"
	"github.com/ghodss/yaml"
	"github.com/stretchr/testify/assert"
)

type NimServiceComposerTest struct {
	specFname        string // input spec filename
	expectedFname    string // expected output filename
	expectedError    string // expected error message (if any)
	subdomainEnabled bool   // whether the cluster has subdomain support enabled (default: false)
	ingressClass     string // ingress class name (default: "")
	isOpenshift      bool   // whether the cluster is an OpenShift cluster (default: false)
	gpuVendor        string // gpu vendor (default: "nvidia", supported: "nvidia", "amd")
	clusterVersion   string // cluster version for capability gating (default: "2.25.0")
}

type NimReplicaCompositionTest struct {
	autoscaling  bool
	replicas     *int32
	wantReplicas *int32
	wantError    bool
}

var NimServiceComposerTests = []TableEntry{
	Entry("valid image", NimServiceComposerTest{
		specFname:     "image_with_tag",
		expectedFname: "image_with_tag",
	}),

	Entry("invalid image", NimServiceComposerTest{
		specFname:     "image_invalid_reference",
		expectedError: "invalid image reference",
	}),

	Entry("image no tag", NimServiceComposerTest{
		specFname:     "image_no_tag",
		expectedFname: "image_no_tag",
	}),

	Entry("image missing", NimServiceComposerTest{
		specFname:     "image_missing",
		expectedError: "value for 'image' must be provided",
	}),

	Entry("image pull policy and secret", NimServiceComposerTest{
		specFname:     "image_pull_policy_and_secret",
		expectedFname: "image_pull_policy_and_secret",
	}),

	Entry("image pull secret with user credential", NimServiceComposerTest{
		specFname:     "image_pull_secret_user_credential",
		expectedError: "imagePullSecrets: userCredential is not supported",
	}),

	Entry("env vars all types", NimServiceComposerTest{
		specFname:     "env_vars_all_types",
		expectedFname: "env_vars_all_types",
	}),

	Entry("env vars missing name", NimServiceComposerTest{
		specFname:     "env_vars_missing_name",
		expectedError: "environmentVariables: name is required",
	}),

	Entry("env vars user credential not supported", NimServiceComposerTest{
		specFname:     "env_vars_user_credential",
		expectedError: "environmentVariables 'TEST_VAR': userCredential is not supported",
	}),

	Entry("env vars mutual exclusive", NimServiceComposerTest{
		specFname:     "env_vars_mutual_exclusive",
		expectedError: "environmentVariables 'MULTIPLE_SOURCES': value, configMap, podFieldRef and secret are mutually exclusive",
	}),

	Entry("env vars no source", NimServiceComposerTest{
		specFname:     "env_vars_no_source",
		expectedError: "environmentVariables 'NO_SOURCE_VAR': value, configMap, podFieldRef or secret must be set",
	}),

	Entry("env vars secret missing name", NimServiceComposerTest{
		specFname:     "env_vars_secret_missing_name",
		expectedError: "environmentVariables 'FROM_SECRET': required secret name is not provided",
	}),

	Entry("env vars secret missing key", NimServiceComposerTest{
		specFname:     "env_vars_secret_missing_key",
		expectedError: "environmentVariables 'FROM_SECRET': required secret key is not provided",
	}),

	Entry("env vars configmap missing name", NimServiceComposerTest{
		specFname:     "env_vars_configmap_missing_name",
		expectedError: "environmentVariables 'FROM_CONFIGMAP': required configMap name is not provided",
	}),

	Entry("env vars configmap missing key", NimServiceComposerTest{
		specFname:     "env_vars_configmap_missing_key",
		expectedError: "environmentVariables 'FROM_CONFIGMAP': required configMap key is not provided",
	}),

	Entry("env vars podfield missing path", NimServiceComposerTest{
		specFname:     "env_vars_podfield_missing_path",
		expectedError: "environmentVariables 'FROM_POD_FIELD': required filedRef path is not provided",
	}),

	Entry("labels and annotations", NimServiceComposerTest{
		specFname:     "labels_and_annotations",
		expectedFname: "labels_and_annotations",
	}),

	Entry("labels missing name", NimServiceComposerTest{
		specFname:     "labels_missing_name",
		expectedError: "labels: name must be provided",
	}),

	Entry("annotations missing name", NimServiceComposerTest{
		specFname:     "annotations_missing_name",
		expectedError: "annotations: name must be provided",
	}),

	Entry("tolerations", NimServiceComposerTest{
		specFname:     "tolerations",
		expectedFname: "tolerations",
	}),

	Entry("tolerations missing key", NimServiceComposerTest{
		specFname:     "tolerations_missing_key",
		expectedError: "tolerations: key must be provided",
	}),

	Entry("tolerations missing operator", NimServiceComposerTest{
		specFname:     "tolerations_missing_operator",
		expectedError: "tolerations: operator must be provided",
	}),

	Entry("tolerations missing effect", NimServiceComposerTest{
		specFname:     "tolerations_missing_effect",
		expectedError: "tolerations: effect must be provided",
	}),

	Entry("tolerations equal missing value", NimServiceComposerTest{
		specFname:     "tolerations_equal_missing_value",
		expectedError: "tolerations: value must be provided when operator is set to Equal",
	}),

	Entry("tolerations exists with value", NimServiceComposerTest{
		specFname:     "tolerations_exists_with_value",
		expectedError: "tolerations: value is not allowed when operator is set to Exists",
	}),

	Entry("tolerations seconds non noexecute", NimServiceComposerTest{
		specFname:     "tolerations_seconds_non_noexecute",
		expectedError: "tolerations: seconds is not allowed when effect is not set to NoExecute",
	}),

	Entry("ngc auth secret missing", NimServiceComposerTest{
		specFname:     "ngc_auth_secret_missing",
		expectedError: "value for 'ngcAuthSecret' must be provided",
	}),

	Entry("replicas", NimServiceComposerTest{
		specFname:     "replicas",
		expectedFname: "replicas",
	}),

	Entry("security", NimServiceComposerTest{
		specFname:     "security",
		expectedFname: "security",
	}),

	Entry("expose disabled", NimServiceComposerTest{
		specFname:     "expose_disabled",
		expectedFname: "expose_disabled",
	}),

	Entry("expose enabled", NimServiceComposerTest{
		specFname:     "expose_enabled",
		expectedFname: "expose_enabled",
	}),

	Entry("expose enabled legacy ingress (cluster < 2.25 builds Ingress in NIM CR)", NimServiceComposerTest{
		specFname:      "expose_enabled",
		expectedFname:  "expose_enabled_legacy_ingress",
		clusterVersion: "2.24.0",
	}),

	Entry("expose enabled legacy ingress with haproxy", NimServiceComposerTest{
		specFname:      "expose_enabled",
		expectedFname:  "expose_enabled_legacy_haproxy",
		ingressClass:   "haproxy",
		clusterVersion: "2.24.0",
	}),

	Entry("expose enabled with subdomain support", NimServiceComposerTest{
		specFname:        "expose_enabled",
		expectedFname:    "expose_enabled_subdomain",
		subdomainEnabled: true,
	}),

	Entry("expose enabled with custom ingress class", NimServiceComposerTest{
		specFname:     "expose_enabled",
		expectedFname: "expose_enabled_ingress_class",
		ingressClass:  "custom-ingress-class",
	}),

	Entry("expose enabled with haproxy ingress class", NimServiceComposerTest{
		specFname:     "expose_enabled",
		expectedFname: "expose_enabled_haproxy",
		ingressClass:  "haproxy",
	}),

	Entry("expose enabled for openshift", NimServiceComposerTest{
		specFname:     "expose_enabled",
		expectedFname: "expose_enabled_openshift",
		isOpenshift:   true,
	}),

	Entry("expose enabled with url", NimServiceComposerTest{
		specFname:     "expose_enabled_with_url",
		expectedFname: "expose_enabled_with_url",
	}),

	Entry("expose missing port", NimServiceComposerTest{
		specFname:     "expose_missing_port",
		expectedError: "servingPort: port must be provided",
	}),

	Entry("expose grpc with default protocol", NimServiceComposerTest{
		specFname:     "expose_grpc_ingress_default",
		expectedFname: "expose_grpc_ingress_default",
	}),

	Entry("expose grpc with grpc protocol", NimServiceComposerTest{
		specFname:     "expose_grpc_ingress_grpc",
		expectedFname: "expose_grpc_ingress_grpc",
	}),

	Entry("expose grpc with http protocol", NimServiceComposerTest{
		specFname:     "expose_grpc_ingress_http",
		expectedFname: "expose_grpc_ingress_http",
	}),

	Entry("expose url malformed", NimServiceComposerTest{
		specFname:     "expose_url_malformed",
		expectedError: "malformed url",
	}),

	Entry("expose url invalid host", NimServiceComposerTest{
		specFname:     "expose_url_invalid_host",
		expectedError: "malformed url host",
	}),

	Entry("expose url valid with path", NimServiceComposerTest{
		specFname:     "expose_url_valid_with_path",
		expectedFname: "expose_url_valid_with_path",
	}),

	Entry("expose url valid no path", NimServiceComposerTest{
		specFname:     "expose_url_valid_no_path",
		expectedFname: "expose_url_valid_no_path",
	}),

	Entry("readiness probe full", NimServiceComposerTest{
		specFname:     "readiness_probe_full",
		expectedFname: "readiness_probe_full",
	}),

	Entry("readiness probe minimal", NimServiceComposerTest{
		specFname:     "readiness_probe_minimal",
		expectedFname: "readiness_probe_minimal",
	}),

	Entry("readiness probe missing handler", NimServiceComposerTest{
		specFname:     "readiness_probe_missing_handler",
		expectedError: "probes.readiness: handler must be provided",
	}),

	Entry("readiness probe missing port", NimServiceComposerTest{
		specFname:     "readiness_probe_missing_port",
		expectedError: "probes.readiness.handler.httpGet: port must be provided",
	}),

	Entry("modelStore nim cache", NimServiceComposerTest{
		specFname:     "model_store_nim_cache",
		expectedFname: "model_store_nim_cache",
	}),

	Entry("modelStore pvc", NimServiceComposerTest{
		specFname:     "model_store_pvc",
		expectedFname: "model_store_pvc",
	}),

	Entry("modelStore multiple types", NimServiceComposerTest{
		specFname:     "model_store_multiple_types",
		expectedError: "modelStore: nimCache and pvc are mutually exclusive",
	}),

	Entry("modelStore pvc new full", NimServiceComposerTest{
		specFname:     "model_store_pvc_new_full",
		expectedFname: "model_store_pvc_new_full",
	}),

	// CPU Compute Tests - Positive Cases
	Entry("compute cpu core request", NimServiceComposerTest{
		specFname:     "compute_cpu_core_request",
		expectedFname: "compute_cpu_core_request",
	}),

	Entry("compute cpu core limit", NimServiceComposerTest{
		specFname:     "compute_cpu_core_limit",
		expectedFname: "compute_cpu_core_limit",
	}),

	Entry("compute cpu memory request", NimServiceComposerTest{
		specFname:     "compute_cpu_memory_request",
		expectedFname: "compute_cpu_memory_request",
	}),

	Entry("compute cpu memory limit", NimServiceComposerTest{
		specFname:     "compute_cpu_memory_limit",
		expectedFname: "compute_cpu_memory_limit",
	}),

	Entry("compute cpu all fields", NimServiceComposerTest{
		specFname:     "compute_cpu_all_fields",
		expectedFname: "compute_cpu_all_fields",
	}),

	// CPU Compute Tests - Negative Cases
	Entry("compute cpu core request negative", NimServiceComposerTest{
		specFname:     "compute_cpu_core_request_negative",
		expectedError: "compute.cpuCoreRequest must be no less than zero",
	}),

	Entry("compute cpu core limit negative", NimServiceComposerTest{
		specFname:     "compute_cpu_core_limit_negative",
		expectedError: "compute.cpuCoreLimit must be no less than zero",
	}),

	Entry("compute cpu core limit less than request", NimServiceComposerTest{
		specFname:     "compute_cpu_core_limit_less_than_request",
		expectedError: "compute.cpuCoreLimit (2.0) should be no less than cpuCoreRequest (4.0)",
	}),

	Entry("compute cpu memory request invalid", NimServiceComposerTest{
		specFname:     "compute_cpu_memory_request_invalid",
		expectedError: "compute.cpuMemoryRequest does not comply with k8s resource units rules",
	}),

	Entry("compute cpu memory request negative", NimServiceComposerTest{
		specFname:     "compute_cpu_memory_request_negative",
		expectedError: "compute.cpuMemoryRequest must be no less than zero",
	}),

	Entry("compute cpu memory limit invalid", NimServiceComposerTest{
		specFname:     "compute_cpu_memory_limit_invalid",
		expectedError: "compute.cpuMemoryLimit does not comply with k8s resource units rules",
	}),

	Entry("compute cpu memory limit negative", NimServiceComposerTest{
		specFname:     "compute_cpu_memory_limit_negative",
		expectedError: "compute.cpuMemoryLimit must be no less than zero",
	}),

	Entry("compute cpu memory limit less than request", NimServiceComposerTest{
		specFname:     "compute_cpu_memory_limit_less_than_request",
		expectedError: "compute.cpuMemoryLimit (4294967296) should be no less than cpuMemoryRequest (8589934592)",
	}),

	// GPU Compute Tests - Positive Cases
	Entry("compute gpu devices only", NimServiceComposerTest{
		specFname:     "compute_gpu_devices_only",
		expectedFname: "compute_gpu_devices_only",
	}),

	Entry("compute gpu devices single", NimServiceComposerTest{
		specFname:     "compute_gpu_devices_single",
		expectedFname: "compute_gpu_devices_single",
	}),

	Entry("compute gpu devices zero", NimServiceComposerTest{
		specFname:     "compute_gpu_devices_zero",
		expectedFname: "compute_gpu_devices_zero",
	}),

	Entry("compute gpu portion request", NimServiceComposerTest{
		specFname:     "compute_gpu_portion_request",
		expectedFname: "compute_gpu_portion_request",
	}),

	Entry("compute gpu portion with limit", NimServiceComposerTest{
		specFname:     "compute_gpu_portion_with_limit",
		expectedFname: "compute_gpu_portion_with_limit",
	}),

	Entry("compute gpu memory request", NimServiceComposerTest{
		specFname:     "compute_gpu_memory_request",
		expectedFname: "compute_gpu_memory_request",
	}),

	Entry("compute gpu memory with limit", NimServiceComposerTest{
		specFname:     "compute_gpu_memory_with_limit",
		expectedFname: "compute_gpu_memory_with_limit",
	}),

	Entry("compute gpu devices with portion", NimServiceComposerTest{
		specFname:     "compute_gpu_devices_with_portion",
		expectedFname: "compute_gpu_devices_with_portion",
	}),

	// Combined GPU + CPU Compute Tests
	Entry("compute gpu devices with all cpu fields", NimServiceComposerTest{
		specFname:     "compute_gpu_devices_with_cpu_all",
		expectedFname: "compute_gpu_devices_with_cpu_all",
	}),

	Entry("compute gpu devices with cpu core request and limit", NimServiceComposerTest{
		specFname:     "compute_gpu_devices_with_cpu_core",
		expectedFname: "compute_gpu_devices_with_cpu_core",
	}),

	Entry("compute gpu devices with cpu memory request and limit", NimServiceComposerTest{
		specFname:     "compute_gpu_devices_with_cpu_memory",
		expectedFname: "compute_gpu_devices_with_cpu_memory",
	}),

	Entry("compute gpu portion with all cpu fields", NimServiceComposerTest{
		specFname:     "compute_gpu_portion_with_cpu_all",
		expectedFname: "compute_gpu_portion_with_cpu_all",
	}),

	Entry("compute gpu memory with all cpu fields", NimServiceComposerTest{
		specFname:     "compute_gpu_memory_with_cpu_all",
		expectedFname: "compute_gpu_memory_with_cpu_all",
	}),

	Entry("compute gpu devices with portion and all cpu fields", NimServiceComposerTest{
		specFname:     "compute_gpu_devices_with_portion_and_cpu",
		expectedFname: "compute_gpu_devices_with_portion_and_cpu",
	}),

	// GPU Compute Tests - Negative Cases
	Entry("compute gpu devices negative", NimServiceComposerTest{
		specFname:     "compute_gpu_devices_negative",
		expectedError: "compute.gpuDevicesRequest must be no less than zero",
	}),

	Entry("compute gpu portion zero", NimServiceComposerTest{
		specFname:     "compute_gpu_portion_zero",
		expectedError: "compute.gpuPortionRequest must be greater than zero",
	}),

	Entry("compute gpu portion negative", NimServiceComposerTest{
		specFname:     "compute_gpu_portion_negative",
		expectedError: "compute.gpuPortionRequest must be greater than zero",
	}),

	Entry("compute gpu portion limit zero", NimServiceComposerTest{
		specFname:     "compute_gpu_portion_limit_zero",
		expectedError: "compute.gpuPortionLimit must be greater than zero",
	}),

	Entry("compute gpu memory request invalid", NimServiceComposerTest{
		specFname:     "compute_gpu_memory_request_invalid",
		expectedError: "compute.gpuMemoryRequest does not comply with k8s resource units rules",
	}),

	Entry("compute gpu memory zero", NimServiceComposerTest{
		specFname:     "compute_gpu_memory_zero",
		expectedError: "compute.gpuMemoryRequest must be greater than zero",
	}),

	Entry("compute gpu memory negative", NimServiceComposerTest{
		specFname:     "compute_gpu_memory_negative",
		expectedError: "compute.gpuMemoryRequest must be greater than zero",
	}),

	Entry("compute gpu memory limit invalid", NimServiceComposerTest{
		specFname:     "compute_gpu_memory_limit_invalid",
		expectedError: "compute.gpuMemoryLimit does not comply with k8s resource units rules",
	}),

	Entry("compute gpu memory limit zero", NimServiceComposerTest{
		specFname:     "compute_gpu_memory_limit_zero",
		expectedError: "compute.gpuMemoryLimit must be greater than zero",
	}),

	Entry("compute gpu portion without type", NimServiceComposerTest{
		specFname:     "compute_gpu_portion_without_type",
		expectedFname: "compute_gpu_portion_without_type",
	}),

	Entry("compute gpu memory without type", NimServiceComposerTest{
		specFname:     "compute_gpu_memory_without_type",
		expectedFname: "compute_gpu_memory_without_type",
	}),

	Entry("compute gpu portion and memory without type", NimServiceComposerTest{
		specFname:     "compute_gpu_portion_and_memory_without_type",
		expectedError: "compute.gpuPortionRequest and gpuMemoryRequest are mutually exclusive",
	}),

	Entry("compute gpu portion type without field", NimServiceComposerTest{
		specFname:     "compute_gpu_portion_type_without_field",
		expectedError: "compute.gpuPortionRequest is mandatory when gpuRequestType is portion",
	}),

	Entry("compute gpu memory type without field", NimServiceComposerTest{
		specFname:     "compute_gpu_memory_type_without_field",
		expectedError: "compute.gpuMemoryRequest is mandatory when gpuRequestType is memory",
	}),

	Entry("compute gpu portion limit less than request", NimServiceComposerTest{
		specFname:     "compute_gpu_portion_limit_less_than_request",
		expectedError: "compute.gpuPortionLimit should be no less than gpuPortionRequest",
	}),

	Entry("compute gpu memory limit less than request", NimServiceComposerTest{
		specFname:     "compute_gpu_memory_limit_less_than_request",
		expectedError: "compute.gpuMemoryLimit should be no less than gpuMemoryRequest",
	}),

	Entry("compute gpu vendor not supported", NimServiceComposerTest{
		specFname:     "compute_gpu_devices_only",
		expectedError: "compute: failed to validate gpu vendor is nvidia: nvidia gpu vendor is not available",
		gpuVendor:     "amd",
	}),

	Entry("compute gpu vendor not supported but no gpus requested", NimServiceComposerTest{
		specFname:     "compute_cpu_core_request",
		expectedFname: "compute_cpu_core_request",
		gpuVendor:     "amd",
	}),

	Entry("autoscaling minimal", NimServiceComposerTest{
		specFname:     "autoscaling_minimal",
		expectedFname: "autoscaling_minimal",
	}),

	Entry("autoscaling full", NimServiceComposerTest{
		specFname:     "autoscaling_full",
		expectedFname: "autoscaling_full",
	}),

	Entry("autoscaling memory", NimServiceComposerTest{
		specFname:     "autoscaling_memory",
		expectedFname: "autoscaling_memory",
	}),

	Entry("autoscaling custom metric", NimServiceComposerTest{
		specFname:     "autoscaling_custom_metric",
		expectedFname: "autoscaling_custom_metric",
	}),

	Entry("autoscaling with multiNode", NimServiceComposerTest{
		specFname:     "autoscaling_with_multinode",
		expectedError: "autoscaling: multiNode is mutually exclusive with autoscaling",
	}),

	Entry("autoscaling with replicas", NimServiceComposerTest{
		specFname:     "autoscaling_with_replicas",
		expectedError: "autoscaling: replicas is mutually exclusive with autoscaling",
	}),

	Entry("autoscaling missing metric", NimServiceComposerTest{
		specFname:     "autoscaling_missing_metric",
		expectedError: "autoscaling: metric is mandatory",
	}),

	Entry("autoscaling missing threshold", NimServiceComposerTest{
		specFname:     "autoscaling_missing_threshold",
		expectedError: "autoscaling: metricThreshold is mandatory",
	}),

	Entry("autoscaling max less than min", NimServiceComposerTest{
		specFname:     "autoscaling_max_less_than_min",
		expectedError: "autoscaling: maxReplicas (2) must be no less than minReplicas (5)",
	}),

	Entry("autoscaling only min replicas", NimServiceComposerTest{
		specFname:     "autoscaling_only_min_replicas",
		expectedFname: "autoscaling_only_min_replicas",
	}),

	Entry("autoscaling only max replicas", NimServiceComposerTest{
		specFname:     "autoscaling_only_max_replicas",
		expectedFname: "autoscaling_only_max_replicas",
	}),

	// Multi-Node Tests
	Entry("multinode with gpu", NimServiceComposerTest{
		specFname:     "multinode_with_gpu",
		expectedFname: "multinode_with_gpu",
	}),

	Entry("multinode multiple workers", NimServiceComposerTest{
		specFname:     "multinode_multiple_workers",
		expectedFname: "multinode_multiple_workers",
	}),

	Entry("multinode without gpu", NimServiceComposerTest{
		specFname:     "multinode_without_gpu",
		expectedFname: "multinode_without_gpu",
	}),

	Entry("multinode not set", NimServiceComposerTest{
		specFname:     "multinode_not_set",
		expectedFname: "multinode_not_set",
	}),
}

var _ = Describe("NimServiceComposer", func() {
	Context("ComposeSpec tests", func() {
		DescribeTable("should create proper nim service specs or return sensible errors", func(testDetails NimServiceComposerTest) {
			RunNimServiceComposerTest(testDetails, ".")
		}, NimServiceComposerTests)
	})

	Context("serialized replica configuration", func() {
		DescribeTable("preserves the replica configuration accepted by NIM", func(tc NimReplicaCompositionTest) {
			input, err := os.ReadFile("testing/inputs/autoscaling_minimal.yaml")
			Expect(err).To(BeNil())

			var spec api.NimServiceSpec
			Expect(yaml.Unmarshal(input, &spec)).To(Succeed())
			if !tc.autoscaling {
				spec.Autoscaling.Set(nil)
			}
			spec.Replicas.Set(tc.replicas)
			cluster := mockClustersClientInfo("test-cluster", false, "", false, "nvidia", "3.0.0")
			converter := NewNimServiceComposer("test-workload", "tester", cluster, "test-project", composer.OptionsForNimService())
			var actual nimv1alpha1.NIMServiceSpec
			composeErr := converter.Compose(context.Background(), &spec, &actual)()
			if tc.wantError {
				Expect(composeErr).NotTo(BeNil())
				Expect(composeErr.Error()).To(ContainSubstring("mutually exclusive"))
				return
			}
			Expect(composeErr).To(BeNil())

			encoded, err := json.Marshal(actual)
			Expect(err).To(BeNil())
			var wire map[string]json.RawMessage
			Expect(json.Unmarshal(encoded, &wire)).To(Succeed())
			if tc.wantReplicas == nil {
				Expect(wire).NotTo(HaveKey("replicas"), "autoscaling resource must omit replicas")
				var scale struct {
					Enabled bool `json:"enabled"`
				}
				Expect(json.Unmarshal(wire["scale"], &scale)).To(Succeed())
				Expect(scale.Enabled).To(BeTrue(), "autoscaling must remain enabled")
			} else {
				Expect(wire).To(HaveKey("replicas"))
				var replicas int32
				Expect(json.Unmarshal(wire["replicas"], &replicas)).To(Succeed())
				Expect(replicas).To(Equal(*tc.wantReplicas))
			}
		},
			Entry("autoscaling omits replicas", NimReplicaCompositionTest{autoscaling: true}),
			Entry("fixed replicas remain explicit", NimReplicaCompositionTest{replicas: convutils.Ref(int32(2)), wantReplicas: convutils.Ref(int32(2))}),
			Entry("fixed replicas default to one", NimReplicaCompositionTest{wantReplicas: convutils.Ref(int32(1))}),
			Entry("explicit replicas conflict with autoscaling", NimReplicaCompositionTest{autoscaling: true, replicas: convutils.Ref(int32(2)), wantError: true}),
		)
	})
})

func RunNimServiceComposerTest(testDetails NimServiceComposerTest, basePath string) {

	testBrief := fmt.Sprintf("Spec: %s", testDetails.specFname)

	// Load input spec
	var inputSpec api.NimServiceSpec
	if testDetails.specFname != testutils.NA {
		specYaml, err := testutils.LoadInputFile(testDetails.specFname, basePath)
		Expect(err).To(BeNil(), fmt.Sprintf("failed to load %s.yaml", testDetails.specFname))

		err = yaml.Unmarshal(specYaml, &inputSpec)
		Expect(err).To(BeNil(), fmt.Sprintf("failed to parse %s.yaml", testDetails.specFname))
	}

	// Create composer and run test
	projectName := "test-project"
	workloadName := "test-workload"
	clusterId := "test-cluster"

	version := testDetails.clusterVersion
	if version == "" {
		version = "2.25.0"
	}
	clusterInfo := mockClustersClientInfo(clusterId, testDetails.subdomainEnabled, testDetails.ingressClass, testDetails.isOpenshift, testDetails.gpuVendor, version)

	options := composer.OptionsForNimService()

	nimComposer := NewNimServiceComposer(
		workloadName,
		"tester",
		clusterInfo,
		projectName,
		options)

	var nimServiceSpec nimv1alpha1.NIMServiceSpec
	err := nimComposer.Compose(context.Background(), &inputSpec, &nimServiceSpec)()

	if testDetails.expectedError != testutils.NA {
		Expect(err).NotTo(BeNil(), "expected error was not returned", testDetails.expectedError)
		Expect(err.Error()).To(ContainSubstring(testDetails.expectedError), testBrief)
		return
	} else {
		Expect(err).To(BeNil(), testBrief)
	}

	// Compare with expected results
	CompareNimServiceFiles(testBrief, &nimServiceSpec, testDetails.expectedFname, basePath)
}

func CompareNimServiceFiles(testBrief string, actualResults *nimv1alpha1.NIMServiceSpec, expectedFname string, basePath string) {
	var actualYaml []byte
	actualYaml, marshalError := yaml.Marshal(actualResults)
	Expect(marshalError).To(BeNil(), testBrief)

	expectedYaml := testutils.LoadExpectedFile(expectedFname, actualYaml, basePath)

	isOk := assert.Equal(GinkgoT(), string(expectedYaml), string(actualYaml), testBrief)
	if !isOk {
		tmpPath := fmt.Sprintf("%s/%s.yaml", os.TempDir(), expectedFname)
		osError := os.WriteFile(tmpPath, actualYaml, 0755)
		if osError != nil {
			assert.Fail(GinkgoT(), fmt.Sprintf("failed to backup the result to %s: %v", tmpPath, osError.Error()))
		} else {
			assert.Fail(GinkgoT(), fmt.Sprintf("actual result backed up to %s", tmpPath))
		}
	}
}

func mockClustersClientInfo(clusterId string, subdomainSupportEnabled bool, ingressClass string, isOpenshift bool, gpuVendor string, version string) *openapi.DisplayedCluster {
	displayedCluster := &openapi.DisplayedCluster{
		Uuid: clusterId,
	}
	displayedCluster.Domain.Set(convutils.Ref("https://example.com"))
	displayedCluster.Version.Set(convutils.Ref(version))
	clusterDisplayedStatus := &openapi.ClusterDisplayedStatus{}
	clusterDisplayedStatus.Config.Set(&openapi.ClusterReportedConfig{
		SubdomainEnabled: *openapi.NewNullableBool(&subdomainSupportEnabled),
		IngressClass:     *openapi.NewNullableString(&ingressClass),
	})

	if isOpenshift {
		clusterDisplayedStatus.Platform.Set(&openapi.ClusterReportedStatusPlatform{
			Type: *openapi.NewNullableString(convutils.Ref("openshift")),
		})
	}

	if gpuVendor == "" {
		gpuVendor = "nvidia"
	}
	clusterDisplayedStatus.Dependencies = &openapi.ClusterDependenciesStatus{
		Required: map[string]openapi.ClusterDependencyStatus{
			"gpu-stack": {
				Available: true,
				Components: &map[string]openapi.ClusterDependencyStatus{
					gpuVendor: {
						Available: true,
					},
				},
			},
		},
	}

	displayedCluster.Status.Set(clusterDisplayedStatus)

	return displayedCluster
}
