/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package builders

import (
	"context"
	"strings"
	"testing"

	"k8s.io/utils/ptr"

	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	"github.com/stretchr/testify/assert"

	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	workloadconfig "github.com/run-ai/runai/runai-cluster/cluster/workload-common/config"
	workloadproject "github.com/run-ai/runai/runai-cluster/cluster/workload-common/project"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/builders/common"
	commontesting "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/builders/common/testing"
	jobstesting "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/builders/jobs/testing"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// byte array with encrypted "private-value-1"
var encryptedValue = []byte{
	168, 221, 244, 73, 135, 225, 25, 104, 145, 32, 118, 20, 196, 29, 72, 165,
	224, 40, 97, 180, 83, 157, 205, 140, 158, 186, 60, 73, 173, 193, 146, 219,
	201, 199, 65, 36, 241, 251, 52, 131, 38, 155, 0,
}

func TestBuildInferenceWorkload(t *testing.T) {
	toTest := []common.TestConfig{
		{TestName: "basic"},
		{TestName: "basic-data-volumes"},
		{TestName: "basic_with_annotations_labels"},
		{TestName: "fractional"},
		{TestName: "multi-frac"},
		{TestName: "autoscale"},
		{TestName: "autoscale-defaults"},
		{TestName: "autoscale-hpa-metric"},
		{TestName: "autoscale-custom-metric"},
		{TestName: "empty-fields"},
		{TestName: "pvc-existing"},
		{TestName: "pvc-non-existing"},
		{TestName: "envvars", TestEnvVars: &common.TestEnvVars{ADSEncryptionKey: common.ADSEncryptionKey}},
		{TestName: "private-service-url"},
		{TestName: "node-pool"},
		{TestName: "node-type"},
		{TestName: "node_affinity_basic"},
		{TestName: "node_affinity_with_nodepools"},
		{TestName: "pod_affinity"},
		{TestName: "mps"},
		{TestName: "security-context"},
		{TestName: "nfs"},
		{TestName: "hostpath"},
		{TestName: "git-sync"},
		{TestName: "git-sync-with-ssh"},
		{TestName: "git-sync-with-password"},
		{TestName: "git-sync-custom-ca", TestEnvVars: &common.TestEnvVars{CustomCAGitEnvVarValue: "true"}},
		{TestName: "configmap-volume"},
		{TestName: "secret-volume"},
		{TestName: "emptydir-volume"},
		{TestName: "basic_ports"},
		{TestName: "basic_ports_grpc"},
		{TestName: "basic-cluster-ip"},
		{TestName: "basic-load-balancer"},
		{TestName: "basic-node-port"},
		{TestName: "basic-serving-port"},
		{TestName: "basic-priority-class"},
		{TestName: "probes"},
		{TestName: "exposed_urls"},
		{TestName: "exposed_urls_basic"},
		{TestName: "exposed_urls_with_authorized_users"},
		{TestName: "exposed_urls_with_authorized_groups"},
		{TestName: "exposed_urls_multiple_with_auth"},
		{TestName: "serving-port-auth-authenticated-users"},
		{TestName: "serving-port-auth-users"},
		{TestName: "serving-port-auth-groups"},
		{TestName: "serving-port-auth-users-with-exposed-url-auth-users"},
		{TestName: "serving-port-auth-users-with-exposed-url-auth-groups"},
		{TestName: "serving-port-auth-groups-with-exposed-url-auth-users"},
		{TestName: "serving-port-auth-groups-with-exposed-url-auth-groups"},
		{TestName: "serving-port-auth-with-protocol-grpc"},
		{TestName: "serving-port-auth-with-readiness-probe"},
		{TestName: "autoscale-advanced"},
		{TestName: "model-swap-vllm"},
		{TestName: "model-swap-vllm-with-args"},
		{TestName: "model-swap-tgi"},
		{TestName: "model-swap-nim"},
		{TestName: "model-swap-nim-with-env-vars"},
		{TestName: "nim-metrics-service"},
		{TestName: "image_pull_secrets", TestEnvVars: &common.TestEnvVars{ADSEncryptionKey: common.ADSEncryptionKey}},
		{TestName: "preemptibility"},
		{TestName: "working_dir"},
		{TestName: "host_ipc_network"},
	}

	defaultTestEnvVars := &common.TestEnvVars{RedisFQDN: "cluster-redis.runai.svc.cluster.local:6379"}

	// List of files to test
	for _, curTest := range toTest {
		runaiBuilder, ctx, commonBuilder := createBuilders(t)
		utils.ActiveUuidGenerator = &jobstesting.MockUuidGenerator{}
		if curTest.TestEnvVars == nil {
			curTest.TestEnvVars = defaultTestEnvVars
		}
		common.TestJobBuilder(ctx, t, &curTest, runaiBuilder, commonBuilder, &TestKnativeBuilder{}, commonBuilder)
	}
}

func createBuilders(t *testing.T) (common.ComputeBuilder, context.Context, *common.SharedResourcesBuilder) {
	k8sClient := initClientWithResources(t, namespace)
	builder := NewKnativeBuilder(k8sClient)
	projectName := strings.TrimPrefix(namespace, "runai-")
	ctx := workloadproject.SetProjectInContext(context.Background(), projectName)
	pvcBuilder := common.NewPvcBuilder(k8sClient)
	serviceBuilder := common.NewServiceBuilder(k8sClient)
	urlExporterBuilder := common.NewExternalsUrlsBuilder(k8sClient, workloadconfig.IsOpenshift())
	credentialsBuilder := common.NewCredentialsBuilder(k8sClient)
	commonBuilder := common.NewShardResourcesBuilder(serviceBuilder, urlExporterBuilder, pvcBuilder, credentialsBuilder)
	return builder, ctx, commonBuilder
}

func createDistributedInferenceBuilder(t *testing.T) (context.Context, *DistributedInferenceBuilder) {
	return createDistributedInferenceBuilderWithElastic(t, false)
}

func createDistributedInferenceBuilderWithElastic(t *testing.T, elasticEnabled bool) (context.Context, *DistributedInferenceBuilder) {
	k8sClient := initClientWithResources(t, namespace)
	builder := NewDistributedInferenceBuilder(k8sClient, false, elasticEnabled)
	projectName := strings.TrimPrefix(namespace, "runai-")
	ctx := workloadproject.SetProjectInContext(context.Background(), projectName)
	return ctx, builder
}

func initClientWithResources(t *testing.T, namespace string) client.Client {
	fakeDataVolumePvc := corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "data-volume-pvc",
			Namespace: namespace,
			Labels: map[string]string{
				v1alpha1.DataVolumeIdLabel: "5e0445ea-6d67-4641-a791-b4a372ef722c",
			},
		},
	}

	fakeCredential := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-user-credential",
			Namespace: "runai",
		},
		Data: map[string][]byte{
			"my-user-key": encryptedValue,
		},
	}

	kaiNodePools := []runtime.Object{
		kaiNodePool("default", "", "", ""),
		kaiNodePool("nodepoola", "test-key", "test-value", ""),
		kaiNodePool("mnnvl-node-1", "nodepool-key-1", "nodepool-value-1", "nvidia.com/gpu.imex-domain"),
		kaiNodePool("mnnvl-node-2", "nodepool-key-2", "nodepool-value-2", "custom.run.ai/mnnvl-key"),
		kaiNodePool("mnnvl-node-3", "nodepool-key-3", "nodepool-value-3", "nvidia.com/nvlink"),
	}

	objects := append([]runtime.Object{&fakeDataVolumePvc, &fakeCredential}, kaiNodePools...)

	fakeClient := commontesting.NewFakeClient(t, objects...)
	return fakeClient
}

// kaiNodePool builds a kai.resources node pool fixture. An empty accelerationLabelKey
// leaves the annotation unset, which the readers treat as the schema default.
func kaiNodePool(name, labelKey, labelValue, accelerationLabelKey string) *kaiv1alpha1.NodePool {
	pool := &kaiv1alpha1.NodePool{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: kaiv1alpha1.NodePoolSpec{
			LabelKey:   labelKey,
			LabelValue: labelValue,
		},
		Status: kaiv1alpha1.NodePoolStatus{Phase: kaiv1alpha1.NodePoolReady},
	}
	if accelerationLabelKey != "" {
		pool.Annotations = map[string]string{
			kaiv1alpha1.AnnotationGPUNetworkAccelerationDetected: "true",
			kaiv1alpha1.AnnotationGPUNetworkAccelerationLabelKey: accelerationLabelKey,
		}
	}
	return pool
}

// TestBuildersInferenceNegative tests the inference builder with negative scenarios,
// where an error is expected and given as string in addition to the input yaml
func TestBuildersInferenceNegative(t *testing.T) {
	toTest := []common.TestConfig{
		{
			TestName:          "autoscale-invalid-metric",
			TestExpectedError: "is not supported",
		},
		{
			TestName:          "invalid-gpu-memory",
			TestExpectedError: "failed to parse gpu memory",
		},
		{
			TestName:          "pvc-ephemeral",
			TestExpectedError: "this workload type doesn't support ephemeral PVC",
		},
		{
			TestName:          "s3-bucket",
			TestExpectedError: "this workload type does not support s3 volumes",
		},
		{
			TestName:          "serving-port-auth-users-invalid-type1",
			TestExpectedError: "servingPortAuthorizedUsers is only allowed when servingPortAuthorizationType is set to authorizedUsersOrGroups",
		},
		{
			TestName:          "serving-port-auth-users-invalid-type2",
			TestExpectedError: "servingPortAuthorizedUsers is only allowed when servingPortAuthorizationType is set to authorizedUsersOrGroups",
		},
		{
			TestName:          "serving-port-auth-users-invalid-type3",
			TestExpectedError: "servingPortAuthorizedUsers is only allowed when servingPortAuthorizationType is set to authorizedUsersOrGroups",
		},
		{
			TestName:          "serving-port-auth-groups-invalid-type1",
			TestExpectedError: "servingPortAuthorizedGroups is only allowed when servingPortAuthorizationType is set to authorizedUsersOrGroups",
		},
		{
			TestName:          "serving-port-auth-groups-invalid-type2",
			TestExpectedError: "servingPortAuthorizedGroups is only allowed when servingPortAuthorizationType is set to authorizedUsersOrGroups",
		},
		{
			TestName:          "serving-port-auth-groups-invalid-type3",
			TestExpectedError: "servingPortAuthorizedGroups is only allowed when servingPortAuthorizationType is set to authorizedUsersOrGroups",
		},
		{
			TestName:          "serving-port-auth-invalid-both-users-and-groups",
			TestExpectedError: "servingPortAuthorizedUsers and servingPortAuthorizedGroups are mutually exclusive",
		},
		{
			TestName:          "serving-port-auth-groups-invalid-name1",
			TestExpectedError: "authorizedGroups: no valid groups found",
		},
		{
			TestName:          "serving-port-auth-groups-invalid-name2",
			TestExpectedError: "authorizedGroups: invalid group: --group-a",
		},
		{
			TestName:          "serving-port-auth-groups-invalid-name3",
			TestExpectedError: "authorizedGroups: invalid group: group a",
		},
		{
			TestName:          "model-swap-invalid-format",
			TestExpectedError: "model swap annotation contains invalid JSON: unexpected end of JSON input",
		},
		{
			TestName:          "model-swap-invalid-contents",
			TestExpectedError: "model swap annotation contains invalid model name: 'cant use spaces/    model-1'",
			// more negative tests in knative_build_model_swap_test.go
		},
	}

	for _, curTest := range toTest {
		runaiBuilder, ctx, commonBuilder := createBuilders(t)
		common.TestJobBuilder(ctx, t, &curTest, runaiBuilder, commonBuilder, &TestKnativeBuilder{}, commonBuilder)
	}
}

func TestGetServingPort(t *testing.T) {
	workloadTypesToTest := []string{constants.KindInferenceWorkload}

	for _, workloadType := range workloadTypesToTest {
		_, err := GetServingPort(context.Background(), &field_groups.NetworkingFields{
			Ports: &field_types.PortsField{
				Value: &map[string]field_types.PortItemAndRules{
					"ok": {
						Value: &field_types.PortItemValue{
							ServiceType:   ptr.To(constants.ServiceTypeServingPort),
							ContainerPort: ptr.To(80),
							HostPort:      ptr.To(32767),
						},
					},
				},
			},
		}, workloadType)
		assert.NoError(t, err)
	}

	for _, workloadType := range workloadTypesToTest {
		_, err := GetServingPort(context.Background(), &field_groups.NetworkingFields{
			Ports: &field_types.PortsField{
				Value: &map[string]field_types.PortItemAndRules{
					"ok": {
						Value: &field_types.PortItemValue{
							ContainerPort: ptr.To(22),
							HostPort:      ptr.To(30000),
						},
					},
					"ok1": {
						Value: &field_types.PortItemValue{
							ContainerPort: ptr.To(80),
							HostPort:      ptr.To(32767),
						},
					},
				},
			},
			ServiceType: field_types.NewStringField(constants.ServiceTypeClusterIP),
		}, workloadType)
		assert.Contains(t, err.Error(), "inference workload should have exactly one serving port")
	}

	for _, workloadType := range workloadTypesToTest {
		_, err := GetServingPort(context.Background(), &field_groups.NetworkingFields{
			Ports: &field_types.PortsField{
				Value: &map[string]field_types.PortItemAndRules{
					"ok": {
						Value: &field_types.PortItemValue{
							ServiceType:   ptr.To(constants.ServiceTypeNodePort),
							ContainerPort: ptr.To(22),
							HostPort:      ptr.To(30000),
						},
					},
					"ok1": {
						Value: &field_types.PortItemValue{
							ServiceType:   ptr.To(constants.ServiceTypeServingPort),
							ContainerPort: ptr.To(80),
							HostPort:      ptr.To(32767),
						},
					},
				},
			},
		}, workloadType)
		assert.NoError(t, err)
	}

	for _, workloadType := range workloadTypesToTest {
		_, err := GetServingPort(context.Background(), &field_groups.NetworkingFields{
			Ports: &field_types.PortsField{
				Value: &map[string]field_types.PortItemAndRules{
					"ok": {
						Value: &field_types.PortItemValue{
							ServiceType:   ptr.To(constants.ServiceTypeServingPort),
							ContainerPort: ptr.To(22),
							HostPort:      ptr.To(30000),
						},
					},
					"ok1": {
						Value: &field_types.PortItemValue{
							ServiceType:   ptr.To(constants.ServiceTypeServingPort),
							ContainerPort: ptr.To(80),
							HostPort:      ptr.To(32767),
						},
					},
				},
			},
			ServiceType: field_types.NewStringField(constants.ServiceTypeNodePort),
		}, workloadType)
		assert.Contains(t, err.Error(), "inference workload should have exactly one serving port")
	}
}

func TestBuildDistributedInferenceWorkload(t *testing.T) {
	toTest := []TestConfig{
		{TestName: "lws/basic"},
		{TestName: "lws/no_spec"},
		{TestName: "lws/no_worker_spec"},
		{TestName: "lws/zero_workers_and_replicas"},
		{TestName: "lws/serving_port_only_port"},
		{TestName: "lws/serving_port_expose_externally_true"},
		{TestName: "lws/serving_port_expose_externally_false"},
		{TestName: "lws/serving_port_with_custom_url"},
		{TestName: "lws/serving_port_public_auth"},
		{TestName: "lws/serving_port_auth_external"},
		{TestName: "lws/serving_port_auth_internal"},
		{TestName: "lws/serving_port_auth_users"},
		{TestName: "lws/serving_port_auth_groups"},
		{TestName: "lws/serving_port_all_fields"},
		{TestName: "lws/startup_policy_leadercreated"},
		{TestName: "lws/startup_policy_leaderready"},
		{TestName: "lws/restart_policy_recreate"},
		{TestName: "lws/restart_policy_none"},
		{TestName: "lws/container_resources"},
		{TestName: "lws/container_resources_multi_frac"},
		{TestName: "lws/container_environment"},
		{TestName: "lws/container_security_context"},
		{TestName: "lws/container_probes"},
		{TestName: "lws/container_probes_with_serving_port"},
		{TestName: "lws/annotations_labels"},
		{TestName: "lws/backend_workload_counter_annotation"},
		{TestName: "lws/priority_class"},
		{TestName: "lws/tolerations"},
		{TestName: "lws/pod_affinity"},
		{TestName: "lws/image_pull_policy"},
		{TestName: "lws/run_as_non_root"},
		{TestName: "lws/read_only_root_filesystem"},
		{TestName: "lws/create_home_dir"},
		{TestName: "lws/node_pools"},
		{TestName: "lws/node_type"},
		{TestName: "lws/node_affinity"},
		{TestName: "lws/node_affinity_with_node_pools"},
		{TestName: "lws/working_dir"},
		{TestName: "lws/stdin_tty"},
		{TestName: "lws/seccomp_profile_type"},
		{TestName: "lws/restart_policy_entities"},
		{TestName: "lws/pvc_existing"},
		{TestName: "lws/pvc_new"},
		{TestName: "lws/pvc_mix"},
		{TestName: "lws/configmap_volume"},
		{TestName: "lws/secret_volume"},
		{TestName: "lws/emptydir_volume"},
		{TestName: "lws/image_pull_secrets"},
		{TestName: "lws/mnnvl_basic"},
		{TestName: "lws/mnnvl_none"},
		{TestName: "lws/mnnvl_leader_gpu_only"},
		{TestName: "lws/mnnvl_worker_gpu_only"},
		{TestName: "lws/mnnvl_multiple_replicas"},
		{TestName: "lws/mnnvl_gpu_memory_and_fraction"},
		{TestName: "lws/mnnvl_pod_affinity"},
		{TestName: "lws/privileged"},
		{TestName: "lws/host_ipc_network"},
		{TestName: "lws/preemptibility"},
		{TestName: "lws/path-rewrite-nginx-http"},
		{TestName: "lws/path-rewrite-nginx-http"},
		{TestName: "lws/path-rewrite-haproxy-http", TestEnvVars: &TestEnvVars{IngressClassEnvVarValue: "haproxy"}},
		{TestName: "lws/path-rewrite-no-path"},

		// negative tests
		{
			TestName:          "lws/serving_port_missing_port",
			TestExpectedError: "serving port: mandatory field 'port' is not set",
		},
		{
			TestName:          "lws/serving_port_multiple_ports",
			TestExpectedError: "only a single serving port is supported",
		},
		{
			TestName:          "lws/serving_port_auth_public_with_users",
			TestExpectedError: "authorizedUsers is only allowed when authorizationType is set to authorizedUsers",
		},
		{
			TestName:          "lws/serving_port_auth_public_with_groups",
			TestExpectedError: "authorizedGroups is only allowed when authorizationType is set to authorizedGroups",
		},
		{
			TestName:          "lws/serving_port_auth_authenticated_with_users",
			TestExpectedError: "authorizedUsers is only allowed when authorizationType is set to authorizedUsers",
		},
		{
			TestName:          "lws/serving_port_auth_authenticated_with_groups",
			TestExpectedError: "authorizedGroups is only allowed when authorizationType is set to authorizedGroups",
		},
		{
			TestName:          "lws/serving_port_auth_authorized_users_with_groups",
			TestExpectedError: "authorizedUsers is only allowed when authorizationType is set to authorizedUsers",
		},
		{
			TestName:          "lws/serving_port_auth_authorized_groups_with_users",
			TestExpectedError: "authorizedGroups is only allowed when authorizationType is set to authorizedGroups",
		},
		{
			TestName:          "lws/serving_port_auth_authorized_users_empty",
			TestExpectedError: "authorizedUsers is only allowed when authorizationType is set to authorizedUsers",
		},
		{
			TestName:          "lws/serving_port_auth_authorized_groups_empty",
			TestExpectedError: "authorizedGroups is only allowed when authorizationType is set to authorizedGroups",
		},
		{
			TestName:          "lws/serving_port_auth_authorized_users_both",
			TestExpectedError: "authorizedGroups is only allowed when authorizationType is set to authorizedGroups",
		},
		{
			TestName:          "lws/serving_port_auth_authorized_groups_both",
			TestExpectedError: "authorizedUsers is only allowed when authorizationType is set to authorizedUsers",
		},
		{
			TestName:          "lws/serving_port_auth_authorized_users_or_groups_invalid",
			TestExpectedError: "invalid authorizationType: authorizedUsersOrGroups",
		},
		{
			TestName:          "lws/startup_policy_invalid",
			TestExpectedError: "invalid startup policy",
		},
		{
			TestName:          "lws/restart_policy_invalid",
			TestExpectedError: "invalid restart policy",
		},
		{
			TestName:          "lws/pvc_ephemeral",
			TestExpectedError: "this workload type doesn't support ephemeral PVC",
		},
		{
			TestName:          "lws/mnnvl_required_without_gpus",
			TestExpectedError: "ComputeDomain is not supported for CPU-only workloads",
		},
		{
			TestName:          "lws/mnnvl_no_replicas",
			TestExpectedError: "ComputeDomain is not supported for CPU-only workloads",
		},
	}

	// List of files to test
	for _, curTest := range toTest {
		ctx, builder := createDistributedInferenceBuilder(t)
		utils.ActiveUuidGenerator = &jobstesting.MockUuidGenerator{}

		TestDistributedInferenceWorkloadBuilder(ctx, t, &curTest, builder)
	}
}

func TestBuildDistributedInferenceWorkloadElasticEnabled(t *testing.T) {
	toTest := []TestConfig{
		{TestName: "lws/mnnvl_elastic_enabled"},
	}

	// List of files to test with elastic enabled
	for _, curTest := range toTest {
		ctx, builder := createDistributedInferenceBuilderWithElastic(t, true)
		utils.ActiveUuidGenerator = &jobstesting.MockUuidGenerator{}

		TestDistributedInferenceWorkloadBuilder(ctx, t, &curTest, builder)
	}
}
