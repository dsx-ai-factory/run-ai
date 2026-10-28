/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package iw_tw_workloads

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"time"

	clusterApi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer/composer_v2"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer/notifications"
	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/testutils"
	k8sObjectsApi "github.com/dsx-ai-factory/run-ai/runai-clients/k8s-objects-tracker-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/checker"
	"github.com/ghodss/yaml"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
)

const assetUuid = "123e4567-e89b-42d3-a456-ab1234567890"

const testOauth2ProxyImage = "test.registry/oauth2-proxy:test"

// testClusterOpts lets per-entry tests pick a non-default cluster shape
// (OpenShift, Gateway-API enabled, etc.) so we can exercise the Route /
// HTTPRoute paths in buildExternalAccess end-to-end without changing the
// default test cluster used by the other entries.
type testClusterOpts struct {
	isOpenshift         bool
	externalAuthEnabled *bool
	ingressEnabled      *bool
	gatewayAPIEnabled   *bool
	gatewayName         string
	gatewayNamespace    string
}

func testClusterStorageConfig(domain string, subdomain bool) composer.ClusterStorageConfig {
	return testClusterStorageConfigWithOpts(domain, subdomain, testClusterOpts{})
}

func testClusterStorageConfigWithOpts(domain string, subdomain bool, opts testClusterOpts) composer.ClusterStorageConfig {
	reported := clusterApi.ClusterReportedConfig{}
	reported.SetGitSyncImage("registry.k8s.io/git-sync/git-sync:v4.7.0")
	reported.SetRunaiOauth2ProxyUsernameLoaderImage("registry.access.redhat.com/ubi9/ubi-minimal:9.8")
	reported.SetS3FileSystemImage("nvcr.io/nvidia/runai/goofys:2.26.30")
	reported.SetRunaiOauth2ProxyImage(testOauth2ProxyImage)
	reported.SetSubdomainEnabled(subdomain)
	if opts.externalAuthEnabled != nil {
		reported.SetExternalAuthUrlEnabled(*opts.externalAuthEnabled)
	}
	if opts.ingressEnabled != nil {
		reported.SetIngressEnabled(*opts.ingressEnabled)
	}
	if opts.gatewayAPIEnabled != nil {
		reported.SetGatewayAPIEnabled(*opts.gatewayAPIEnabled)
	}
	if opts.gatewayName != "" {
		reported.SetGatewayName(opts.gatewayName)
	}
	if opts.gatewayNamespace != "" {
		reported.SetGatewayNamespace(opts.gatewayNamespace)
	}

	status := clusterApi.ClusterDisplayedStatus{}
	status.SetConfig(reported)
	if opts.isOpenshift {
		platform := clusterApi.ClusterReportedStatusPlatform{}
		platform.SetType("openshift")
		status.SetPlatform(platform)
	}

	cluster := clusterApi.DisplayedCluster{}
	cluster.Status.Set(&status)
	cluster.SetDomain(domain)
	return composer.NewClusterStorageConfigFromCluster(&cluster)
}

type ComposerTest struct {
	workloadType string

	specFname          string // values of the workload OR defaults of the policy
	expectedError      string
	expectedFname      string
	storageClassConfig string

	expectedRequirements composer.WorkloadRequirements

	assetsInfo *composer.ComplimentaryAssetInfo

	clusterVersion string

	actualNodePools []clusterApi.Nodepool

	dataVolumeClaimNames map[string]string // maps data volume ID to PVC claim name

	requester string // the user who submitted the workload, defaults to "tester"

	// Optional overrides for the cluster-storage-config used by the test.
	// When zero-valued the default config (vanilla K8s, no Gateway, subdomain
	// disabled) is used. clusterSubdomain only takes effect when at least one
	// of the cluster overrides below is set, so existing entries keep using
	// the default subdomain=false.
	clusterIsOpenshift         bool
	clusterSubdomain           bool
	clusterExternalAuthEnabled *bool
	clusterIngressEnabled      *bool
	clusterGatewayAPIEnabled   *bool
	clusterGatewayName         string
	clusterGatewayNamespace    string
}

var ComposerTests = append([]TableEntry{
	Entry("basic workspace", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "basic",
		expectedFname: "basic_ws",
	}),

	Entry("basic workspace with email requester", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "basic",
		expectedFname: "basic_ws_email_requester",
		requester:     "test@example.com",
	}),

	Entry("backoffLimit", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "backoffLimit",
		expectedFname: "backoffLimit_ws",
	}),

	Entry("large_shm", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "large_shm",
		expectedFname: "large_shm_ws",
	}),

	Entry("basic node pools", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "basic_node_pools",
		expectedFname: "basic_node_pools_ws",
		// extra pool "my-node-pool-c" is available but not requested - should have no effect
		actualNodePools: makeTestNodePools("my-node-pool-a", "my-node-pool-b", "my-node-pool-c"),
	}),

	Entry("node pool not found - no pools known to cluster", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "basic_node_pools",
		expectedError: `node pool "my-node-pool-a" does not exist`,
	}),

	Entry("node pool not found - one of several missing", ComposerTest{
		workloadType:    string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:       "basic_node_pools",
		actualNodePools: makeTestNodePools("my-node-pool-a"),
		expectedError:   `node pool "my-node-pool-b" does not exist`,
	}),

	Entry("basic workspace PC", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "basic_pc_preemptible",
		expectedFname: "basic_pc_preemptible_ws",
	}),

	Entry("env empty", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "env_empty_value",
		expectedFname: "env_empty_value_ws",
	}),

	Entry("env empty", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "env_no_name",
		expectedError: "environmentVariables: name is required",
	}),

	Entry("env no value", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "env_no_value",
		expectedError: "environmentVariables 'HOME': value, userCredential, configMap, podFieldRef or secret must be set",
	}),

	Entry("parallelism", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "parallelism",
		expectedFname: "parallelism_ws",
	}),

	Entry("completions", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "completions",
		expectedFname: "completions_ws",
	}),

	Entry("category", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "category",
		expectedFname: "category_ws",
	}),

	Entry("annotations", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "annotations",
		expectedFname: "annotations_ws",
	}),

	Entry("annotations no name", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "annotations_no_name",
		expectedError: "annotations: name must be provided",
	}),

	Entry("labels no name", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "labels_no_name",
		expectedError: "labels: name must be provided",
	}),

	Entry("auto_deletion", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_TRAINING),
		specFname:     "auto_deletion",
		expectedFname: "auto_deletion_tr",
	}),

	Entry("labels and annotations for workspace", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "labels",
		expectedFname: "labels_ws",
	}),

	Entry("labels and annotations for training", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_TRAINING),
		specFname:     "labels",
		expectedFname: "labels_tr",
	}),

	Entry("Kubernetes Job identity labels for workspace", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "labels_job_identity",
		expectedFname: "labels_job_identity_ws",
	}),

	Entry("Kubernetes Job identity labels for training", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_TRAINING),
		specFname:     "labels_job_identity",
		expectedFname: "labels_job_identity_tr",
	}),

	Entry("generated pod identity labels for workspace", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "labels_identity_override",
		expectedFname: "labels_identity_override_ws",
	}),

	Entry("generated pod identity labels for training", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_TRAINING),
		specFname:     "labels_identity_override",
		expectedFname: "labels_identity_override_tr",
	}),

	Entry("generated user annotation without labels for workspace", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "annotations_user_override",
		expectedFname: "annotations_user_override_ws",
	}),

	Entry("generated user annotation without labels for training", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_TRAINING),
		specFname:     "annotations_user_override",
		expectedFname: "annotations_user_override_tr",
	}),

	Entry("rich", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "rich",
		expectedFname: "rich_ws",
	}),

	Entry("node_type", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_TRAINING),
		specFname:     "node_type",
		expectedFname: "node_type_tr",
	}),

	Entry("ports", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "ports",
		expectedFname: "ports_ws",
	}),

	Entry("ports_v1", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "ports_v1",
		expectedFname: "ports_v1_ws",
	}),

	Entry("urls", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "urls",
		expectedFname: "urls_ws",
	}),

	// OpenShift cluster (no subdomain): each exposed URL produces an OpenShift
	// Route (route.openshift.io/v1) instead of a vanilla Ingress, with edge TLS.
	Entry("urls_openshift", ComposerTest{
		workloadType:       string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:          "urls",
		expectedFname:      "urls_openshift_ws",
		clusterIsOpenshift: true,
	}),

	// OpenShift cluster + subdomain support: URL hosts carry the "runai."
	// namespace prefix, which the Route generator strips so traffic lands on
	// the platform's wildcard certificate (matches V1 cluster builder).
	Entry("urls_openshift_subdomain", ComposerTest{
		workloadType:       string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:          "urls",
		expectedFname:      "urls_openshift_subdomain_ws",
		clusterIsOpenshift: true,
		clusterSubdomain:   true,
	}),

	// Vanilla Kubernetes cluster with the Gateway-API Gateway configured: each
	// exposed URL produces an Ingress AND an additive HTTPRoute attached to
	// the cluster Gateway.
	Entry("urls_gateway", ComposerTest{
		workloadType:            string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:               "urls",
		expectedFname:           "urls_gateway_ws",
		clusterGatewayName:      "runai-gateway",
		clusterGatewayNamespace: "runai",
	}),

	// Gateway-API-only cluster: when Ingress is explicitly disabled and Gateway
	// API is enabled, each exposed URL produces only an HTTPRoute.
	Entry("urls_gateway_only", ComposerTest{
		workloadType:             string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:                "urls",
		expectedFname:            "urls_gateway_only_ws",
		clusterIngressEnabled:    convutils.Ref(false),
		clusterGatewayAPIEnabled: convutils.Ref(true),
		clusterGatewayName:       "runai-gateway",
		clusterGatewayNamespace:  "runai",
	}),

	Entry("urls_syntax", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "urls_syntax",
		expectedError: "malformed url: parse \"http://example.com:badport\": invalid port",
	}),

	Entry("urls_tensor_and_jupyter", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "urls_tensor_and_jupyter",
		expectedFname: "urls_tensor_and_jupyter_ws",
	}),

	Entry("urls_external_auth_disabled", ComposerTest{
		workloadType:               string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:                  "urls_tensor_and_jupyter",
		expectedFname:              "urls_external_auth_disabled_ws",
		clusterExternalAuthEnabled: convutils.Ref(false),
	}),

	Entry("urls_host_syntax", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "urls_host_syntax",
		expectedError: "malformed url host",
	}),

	Entry("urls_public", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "urls_authenticated_users",
		expectedFname: "urls_public_ws",
	}),

	Entry("urls_by_group_no_groups", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "urls_by_groups_with_no_groups",
		expectedError: "exposedUrls: authorizedGroups must be set when authorizationType is authorizedGroups",
	}),

	Entry("urls_by_users_no_users", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "urls_by_users_no_users",
		expectedError: "exposedUrls: authorizedUsers must be set when authorizationType is authorizedUsers",
	}),

	Entry("urls_not_abs", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "urls_not_abs",
		expectedFname: "urls_not_abs_ws",
		expectedError: "Invalid url <https://compute.on-demand.io> must be an absolute path",
	}),

	Entry("urls_invalid", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_TRAINING),
		specFname:     "urls_invalid_authorized",
		expectedError: "authorizedUsers and authorizedGroups are mutually exclusive",
	}),

	Entry("ports_and_urls", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "ports_and_urls",
		expectedFname: "ports_and_urls_ws",
	}),

	Entry("priority_class_build", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "priority_class_build",
		expectedFname: "priority_class_build_ws",
	}),

	Entry("priority_class_build_tr", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_TRAINING),
		specFname:     "priority_class_build",
		expectedFname: "priority_class_build_tr",
	}),

	Entry("priority_class_preemptible_tr", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_TRAINING),
		specFname:     "priority_class_preemptible",
		expectedFname: "priority_class_preemptible_tr",
	}),

	Entry("priority_class_train", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_TRAINING),
		specFname:     "priority_class_train",
		expectedFname: "priority_class_train_tr",
	}),

	Entry("priority_class_preemptible", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "priority_class_preemptible",
		expectedFname: "priority_class_preemptible_ws",
	}),

	Entry("preemptibility", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "preemptibility",
		expectedFname: "preemptibility_ws",
	}),

	Entry("semi-preemptibility", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "semi_preemptibility",
		expectedFname: "semi_preemptibility_ws",
	}),

	Entry("pod affinity - required", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "pod_affinity_required",
		expectedFname: "pod_affinity_required_ws",
	}),

	Entry("privileged true", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "privileged_true",
		expectedFname: "privileged_true",
	}),

	Entry("privileged true and allow privileged false", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "privileged_and_allow_privilege_escalation",
		expectedError: "allowPrivilegeEscalation can't be false when privileged is true",
	}),

	Entry("privileged true and run as non root true", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "privileged_and_run_as_non_root",
		expectedError: "runAsNonRoot can't be true when privileged is true",
	}),

	Entry("pod affinity - preferred", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "pod_affinity_preferred",
		expectedFname: "pod_affinity_preferred_ws",
	}),

	Entry("pod affinity - no type", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "pod_affinity_no_type",
		expectedError: "podAffinity: type must be provided",
	}),

	Entry("pod affinity - no key", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "pod_affinity_no_key",
		expectedError: "podAffinity: key must be provided",
	}),

	Entry("tolerations", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "tolerations",
		expectedFname: "tolerations_ws",
	}),

	Entry("tolerations - Any effect", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "tolerations_effect_any",
		expectedFname: "tolerations_effect_any_ws",
	}),

	Entry("tolerations - invalid - no operator", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "tolerations_invalid_no_operator",
		expectedError: "tolerations: operator must be provided",
	}),

	Entry("tolerations - invalid - no key", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "tolerations_invalid_no_key",
		expectedError: "tolerations: key must be provided",
	}),

	Entry("tolerations - invalid - no effect", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "tolerations_invalid_no_effect",
		expectedError: "tolerations: effect must be provided",
	}),

	Entry("tolerations - invalid - equal operator without value", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "tolerations_invalid_equal_without_value",
		expectedError: "tolerations: value must be provided when operator is set to Equal",
	}),

	Entry("tolerations - invalid - exists operator with value", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "tolerations_invalid_exists_with_value",
		expectedError: "tolerations: value is not allowed when operator is set to Exists",
	}),

	Entry("tolerations - invalid - seconds with wrong effect", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "tolerations_invalid_seconds_with_wrong_effect",
		expectedError: "tolerations: seconds is not allowed when effect is not set to NoExecute",
	}),

	Entry("probes", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "probes",
		expectedFname: "probes_ws",
	}),

	Entry("probes - only required fields", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "probes_only_required",
		expectedFname: "probes_only_required_ws",
	}),

	Entry("probes - invalid - no handler", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "probes_invalid_no_handler",
		expectedError: "probes.readiness: handler must be provided",
	}),

	Entry("probes - invalid - httpGet handler without port", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "probes_invalid_no_httpget_port",
		expectedError: "probes.readiness.handler.httpGet: port must be provided",
	}),

	Entry("secret_and_config_env", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "secret_and_config_env",
		expectedFname: "secret_and_config_env",
	}),

	Entry("user_credentials_env", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "user_credentials_env",
		expectedFname: "user_credentials_env",
	}),

	Entry("user_credentials_env_no_asset", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "user_credentials_env_no_asset",
		expectedError: "asset mapping not found for credential name: my-user-credential-not-found",
	}),

	Entry("secret_env_err", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "secret_env_err",
		expectedError: "environmentVariables 'PASSWORD': value, userCredential, configMap, podFieldRef and secret are mutually exclusive",
	}),

	Entry("config_env_err", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "config_env_err",
		expectedError: "environmentVariables 'PASSWORD': value, userCredential, configMap, podFieldRef and secret are mutually exclusive",
	}),

	Entry("pod_field_ref_env_err", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "pod_field_ref_env_err",
		expectedError: "environmentVariables 'PASSWORD': value, userCredential, configMap, podFieldRef and secret are mutually exclusive",
	}),

	Entry("node_affinity", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "node_affinity",
		expectedFname: "node_affinity_ws",
	}),

	Entry("node_affinity_invalid", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "node_affinity_invalid",
		expectedError: "nodeAffinityRequired: Values must be empty when Operator is Exists",
	}),

	Entry("image_pull_secrets", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "image_pull_secrets",
		expectedFname: "image_pull_secrets",
	}),
},
	append(computeTests,
		append(storageTests,
			append(storageClassConfigTests)...,
		)...,
	)...,
)

var lettersAndColon = regexp.MustCompile(`([A-Za-z]+:)`)

type ActualResults struct {
	BatchJobTarget *BatchJobTarget
	Notification   *workloadsApi.WorkloadCreationFields
}

func RunComposerTest(testDetails ComposerTest, basePath string) {
	testBrief := fmt.Sprintf("Value: %s, Rules: , Type: %s", testDetails.specFname, testDetails.workloadType)

	testDetailsStr := lettersAndColon.ReplaceAllString(fmt.Sprintf("%+v", testDetails), "\n\t$1")

	var actualResults ActualResults
	var actualRequirements composer.WorkloadRequirements

	var supersetSpec workloadsApi.SupersetSpec

	if testDetails.specFname != testutils.NA {
		specYaml, err := testutils.LoadInputFile(testDetails.specFname, basePath)
		Expect(err).To(BeNil(), fmt.Sprintf("failed to load of %s.yaml", testDetails.specFname))

		err = yaml.Unmarshal(specYaml, &supersetSpec)
		if err != nil {
			if os.IsNotExist(err) {
				err = fmt.Errorf("file %s.yaml not found", testDetails.specFname)
			}
		}
		Expect(err).To(BeNil(), fmt.Sprintf("failed to parse of %s.yaml", testDetails.specFname))
	}

	clusterVersion := testutils.DefaultClusterVersionForTesting
	if testDetails.clusterVersion != "" {
		clusterVersion = testDetails.clusterVersion
	}

	testutils.RandStringTestMode("p")

	var storageClassConfig map[string]k8sObjectsApi.StorageClassConfiguration

	if testDetails.storageClassConfig != "" {
		fname := "storage-class-config/" + testDetails.storageClassConfig

		configYaml, err := testutils.LoadInputFile(fname, basePath)
		Expect(err).To(BeNil(), fmt.Sprintf("expecting to load %s.yaml", fname))

		err = yaml.Unmarshal(configYaml, &storageClassConfig)
		Expect(err).To(BeNil(), fmt.Sprintf("failed to parse of %s.yaml", fname))
	}

	clusterCapabilities, checkerErr := checker.NewClusterCapabilitiesCheckerForVersion(clusterVersion)
	Expect(checkerErr).To(BeNil())

	assetsMapping := []workloadsApi.AssetsMapping{
		{
			Name: convutils.Ref("my-user-credential"),
			Id:   convutils.Ref(assetUuid),
			Kind: convutils.Ref(workloadsApi.ASSETKIND_NGC_API_KEY),
		},
	}

	options := composer.FieldsComposerOptions{
		UserCredentialsSupported:       true,
		CleanupCreatedPvcsOnCompletion: convutils.Ref(true),
	}

	checkerErr = composer_v2.SetClusterDependentOptions(context.Background(), &options, clusterCapabilities)
	Expect(checkerErr).To(BeNil(), "failed to set cluster dependent options")
	// Existing snapshots cover clusters that still embed OAuth2 templates in the image.
	options.OAuth2ProxyTemplatesSupported = false

	notificationsComposer := notifications.NewNotificationsComposer(
		testDetails.workloadType,
		testDetails.specFname,
		fmt.Sprintf("id-%s", testBrief),
		"test",
		"cluster-123-456",
		"project-654-321",
		"department-456-123",
		time.Time{},
		nil,
		options).WithConnectionsPreProcessor("test", composer.NetworkingResourcePrefixJob)

	var err *workloadsApi.Error

	requester := testDetails.requester
	if requester == "" {
		requester = "tester"
	}

	jobSpecComposer := NewInterTrainJobComposer(
		testDetails.workloadType,
		testDetails.specFname,
		"test",
		"runai-test",
		requester,
		assetsMapping,
		options,
		testDetails.actualNodePools)

	jobSpecComposer.WithStorageClassConfiguration(storageClassConfig)
	jobSpecComposer.WithDataVolumeClaimNames(testDetails.dataVolumeClaimNames)
	clusterOpts := testClusterOpts{
		isOpenshift:         testDetails.clusterIsOpenshift,
		externalAuthEnabled: testDetails.clusterExternalAuthEnabled,
		ingressEnabled:      testDetails.clusterIngressEnabled,
		gatewayAPIEnabled:   testDetails.clusterGatewayAPIEnabled,
		gatewayName:         testDetails.clusterGatewayName,
		gatewayNamespace:    testDetails.clusterGatewayNamespace,
	}
	clusterSubdomain := testDetails.clusterSubdomain
	jobSpecComposer.WithClusterStorageConfig(
		testClusterStorageConfigWithOpts("https://mock-q7z3k.example.com", clusterSubdomain, clusterOpts))

	actualResults.BatchJobTarget = NewMinimalBatchJobTarget()

	err = jobSpecComposer.Compose(context.Background(), convutils.Ref(supersetSpec), actualResults.BatchJobTarget)()
	if err == nil {
		err = notificationsComposer.Compose(context.Background(), nil, testDetails.assetsInfo, convutils.Ref(supersetSpec))()
	}
	actualRequirements = jobSpecComposer.WorkloadRequirements

	//
	//   notification
	//
	actualResults.Notification = notificationsComposer.Output

	// for testing purposes only we change the name to contain some information about the test
	actualResults.Notification.Name = testBrief

	if noResourcesRequested(actualResults.Notification.RequestedResources.Get()) {
		actualResults.Notification.RequestedResources.Unset()
	}

	//
	//   check composer error
	//
	if testDetails.expectedError != testutils.NA {
		Expect(err).NotTo(BeNil(), fmt.Sprintf("expected message was not returned: %s", testDetailsStr))
		if err != nil {
			Expect(err.Message).To(ContainSubstring(testDetails.expectedError), testDetailsStr)
		}
		return
	} else {
		Expect(err).To(BeNil(), testDetailsStr)
	}

	CompareFiles(testDetailsStr, &actualResults, testDetails.expectedFname, basePath)

	Expect(actualRequirements).To(Equal(testDetails.expectedRequirements), testDetailsStr)
}

func CompareFiles(testDetailsStr string, actualResults *ActualResults, expectedFname string, basePath string) {
	var actualYaml []byte
	actualYaml, marshalError := yaml.Marshal(actualResults)
	Expect(marshalError).To(BeNil(), testDetailsStr)

	expectedYaml := testutils.LoadExpectedFile(expectedFname, actualYaml, basePath)

	isOk := assert.Equal(GinkgoT(), string(expectedYaml), string(actualYaml), testDetailsStr)
	if !isOk {
		tmpPath := fmt.Sprintf("%s/%s.yaml", os.TempDir(), expectedFname)
		osError := os.WriteFile(tmpPath, actualYaml, 0600)
		if osError != nil {
			assert.Fail(GinkgoT(), fmt.Sprintf("failed to backup the result to %s: %v", tmpPath, osError.Error()))
		} else {
			assert.Fail(GinkgoT(), fmt.Sprintf("actual result backed up to %s", tmpPath))
		}
	}
}

var _ = Describe("superset composer validations", func() {
	Context("input VS expected tests", func() {
		DescribeTable("should create proper workloads or return sensible errors", func(testDetails ComposerTest) {
			RunComposerTest(testDetails, ".")
		}, ComposerTests)
	})
})

// makeTestNodePools returns a deterministic []clusterApi.Nodepool fixture
// for the given names. Each pool's label is "<name>-key" -> "<name>-value",
// which keeps generated golden files easy to read.
func makeTestNodePools(names ...string) []clusterApi.Nodepool {
	pools := make([]clusterApi.Nodepool, 0, len(names))
	for _, name := range names {
		pools = append(pools, clusterApi.Nodepool{
			Name:       name,
			LabelKey:   name + "-key",
			LabelValue: name + "-value",
		})
	}
	return pools
}

func noResourcesRequested(resources *workloadsApi.WorkloadRequestResources) bool {
	if resources == nil {
		return true
	}
	if resources.GpuRequestType.IsSet() || resources.Gpu.IsSet() || resources.GpuMemory.IsSet() {
		return false
	}
	if resources.Cpu.IsSet() || resources.CpuMemory.IsSet() {
		return false
	}
	if len(resources.ExtendedResources) != 0 {
		return false
	}
	return true
}

var _ = Describe("native Job composer template variable expansion", func() {
	const (
		projectName  = "team-a"
		workloadName = "job-3"
	)

	loadSpec := func(name string) workloadsApi.SupersetSpec {
		GinkgoHelper()
		specYaml, err := testutils.LoadInputFile(name, ".")
		Expect(err).NotTo(HaveOccurred())
		var spec workloadsApi.SupersetSpec
		Expect(yaml.Unmarshal(specYaml, &spec)).To(Succeed())
		return spec
	}

	composeContainer := func(workloadType string, spec workloadsApi.SupersetSpec) *corev1.Container {
		GinkgoHelper()
		ctx := context.Background()
		caps, err := checker.NewClusterCapabilitiesCheckerForVersion(testutils.DefaultClusterVersionForTesting)
		Expect(err).NotTo(HaveOccurred())
		options := composer.FieldsComposerOptions{}
		Expect(composer_v2.SetClusterDependentOptions(ctx, &options, caps)).To(Succeed())

		iwTwComposer := NewInterTrainJobComposer(workloadType, workloadName, projectName, "runai-test", "tester", nil, options, nil)
		iwTwComposer.WithClusterStorageConfig(testClusterStorageConfig("https://mock-q7z3k.example.com", false))

		target := NewMinimalBatchJobTarget()
		Expect(iwTwComposer.Compose(ctx, convutils.Ref(spec), target)()).To(BeNil())
		return &target.Job.Spec.Template.Spec.Containers[0]
	}

	DescribeTable("expands the RUNAI placeholders in command and args",
		func(workloadType string) {
			container := composeContainer(workloadType, loadSpec("jupyter_placeholders"))

			Expect(container.Command).To(Equal([]string{"start-notebook.sh", "--serve-project=team-a"}))
			Expect(container.Args).To(Equal([]string{
				"--NotebookApp.base_url=/team-a/job-3",
				"--project=team-a",
				"--job=job-3",
				"--uid=1000",
				"--gid=2000",
			}))
		},
		Entry("workspace", string(workloadsApi.WORKLOADTYPE_WORKSPACE)),
		Entry("training", string(workloadsApi.WORKLOADTYPE_TRAINING)),
	)
})
