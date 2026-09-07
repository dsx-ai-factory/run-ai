/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package distributed_to_kubeflow

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"time"

	clusterApi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer/composer_v2"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer/notifications"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/topology"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/testutils"
	k8sObjectsApi "github.com/dsx-ai-factory/run-ai/runai-clients/k8s-objects-tracker-client/generated"
	workloadsApi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/checker"
	"github.com/ghodss/yaml"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"
)

const assetUuid = "123e4567-e89b-42d3-a456-ab1234567890"

const testOauth2ProxyImage = "test.registry/oauth2-proxy:test"

// testClusterOpts lets per-entry tests pick a non-default cluster shape
// (OpenShift, Gateway-API enabled, etc.) so we can exercise the Route /
// HTTPRoute paths in buildExternalAccess end-to-end without changing the
// default test cluster used by the other entries.
type testClusterOpts struct {
	isOpenshift                bool
	gatewayName                string
	gatewayNamespace           string
	distGPUNetworkAcceleration bool
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
	if opts.gatewayName != "" {
		reported.SetGatewayName(opts.gatewayName)
	}
	if opts.gatewayNamespace != "" {
		reported.SetGatewayNamespace(opts.gatewayNamespace)
	}
	if opts.distGPUNetworkAcceleration {
		reported.SetDistGPUNetworkAcceleration(true)
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
	framework workloadsApi.DistributedFramework

	workerSpecFname        string
	masterSpecFname        string
	noMaster               bool
	masterSpecSameAsWorker bool
	leaveMasterSpecUnset   bool

	// the framework spec is for providing all sort of specific framework parameters
	// such as number of slots, replicas, etc.
	frameworkSpec *workloadsApi.DistributedSpecSpec

	expectedError      string
	expectedFname      string
	storageClassConfig string

	expectedRequirements composer.WorkloadRequirements

	workerAssetInfo *composer.ComplimentaryAssetInfo
	masterAssetInfo *composer.ComplimentaryAssetInfo

	clusterVersion string

	actualNodePools []clusterApi.Nodepool

	dataVolumeClaimNames map[string]string // maps data volume ID to PVC claim name

	requester string // the user who submitted the workload, defaults to "tester"

	// Optional overrides for the cluster-storage-config used by the test.
	// When zero-valued the default config (vanilla K8s, no Gateway, subdomain
	// disabled) is used. clusterSubdomain only takes effect when at least one
	// of the cluster overrides below is set, so existing entries keep using
	// the default subdomain=false.
	clusterIsOpenshift      bool
	clusterSubdomain        bool
	clusterGatewayName      string
	clusterGatewayNamespace string

	// clusterGPUNetworkAcceleration enables the workload-controller MNNVL flag
	// (distGPUNetworkAcceleration) on the test cluster config, which drives the
	// default MNNVL plugin mode when the workload does not set the label.
	clusterGPUNetworkAcceleration bool
	// mnnvlLabel, when non-empty, adds a concrete run.ai/MNNVL label to the
	// workload so the composer leaves it untouched regardless of the cluster default.
	mnnvlLabel string
}

var ComposerTests = append([]TableEntry{
	Entry("basic workspace", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_MPI,
		workerSpecFname: "basic",
		masterSpecFname: "master_no_gpu",
		expectedFname:   "basic_mpi_master_same_as",
		actualNodePools: makeTestNodePools("master-node-pool-a", "master-node-pool-b"),
	}),

	Entry("MPI master same as worker removes inherited GPU fields", ComposerTest{
		framework:              workloadsApi.DISTRIBUTEDFRAMEWORK_MPI,
		workerSpecFname:        "basic",
		masterSpecFname:        "master_no_gpu",
		masterSpecSameAsWorker: true,
		expectedFname:          "mpi_master_same_as_worker_gpu",
		actualNodePools:        makeTestNodePools("master-node-pool-a", "master-node-pool-b"),
	}),

	Entry("MPI unset master removes inherited GPU fields", ComposerTest{
		framework:            workloadsApi.DISTRIBUTEDFRAMEWORK_MPI,
		workerSpecFname:      "basic",
		leaveMasterSpecUnset: true,
		expectedFname:        "mpi_master_unset_gpu",
		actualNodePools:      makeTestNodePools("master-node-pool-a", "master-node-pool-b"),
	}),

	Entry("basic_mpi", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_MPI,
		workerSpecFname: "basic",
		masterSpecFname: "master_no_gpu",
		expectedFname:   "basic_mpi",
		actualNodePools: makeTestNodePools("master-node-pool-a", "master-node-pool-b"),
	}),

	Entry("mpi_slots", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_MPI,
		workerSpecFname: "basic",
		masterSpecFname: "master_no_gpu",
		expectedFname:   "mpi_slots",
		actualNodePools: makeTestNodePools("master-node-pool-a", "master-node-pool-b"),

		frameworkSpec: &workloadsApi.DistributedSpecSpec{
			SlotsPerWorker: *workloadsApi.NewNullableInt32(convutils.Ref(int32(12))),
		},
	}),

	Entry("mpi_ssh", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_MPI,
		workerSpecFname: "basic",
		masterSpecFname: "master_no_gpu",
		expectedFname:   "mpi_ssh",
		actualNodePools: makeTestNodePools("master-node-pool-a", "master-node-pool-b"),

		frameworkSpec: &workloadsApi.DistributedSpecSpec{
			SshAuthMountPath: *workloadsApi.NewNullableString(convutils.Ref("/home/tmp")),
		},
	}),

	Entry("mpi_x_pytorch", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_MPI,
		workerSpecFname: "basic",
		masterSpecFname: "basic_master",

		actualNodePools: makeTestNodePools("master-node-pool-a", "master-node-pool-b"),

		frameworkSpec: &workloadsApi.DistributedSpecSpec{
			MinReplicas: *workloadsApi.NewNullableInt32(convutils.Ref(int32(32))),
		},
		expectedError: "minReplicas is not supported by MPI framework",
	}),

	Entry("basic_pytorch", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_PY_TORCH,
		workerSpecFname: "basic",
		masterSpecFname: "basic_master",
		expectedFname:   "basic_pytorch",
		actualNodePools: makeTestNodePools("master-node-pool-a", "master-node-pool-b"),
	}),

	Entry("minimal_pytorch_with_master", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_PY_TORCH,
		workerSpecFname: "minimal_pytorch_worker",
		masterSpecFname: "minimal_pytorch_master",
		expectedFname:   "minimal_pytorch_with_master",
		actualNodePools: makeTestNodePools("default"),
	}),

	Entry("pytorch_replicas_range", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_PY_TORCH,
		workerSpecFname: "basic",
		masterSpecFname: "basic_master",
		expectedFname:   "pytorch_replicas_range",
		actualNodePools: makeTestNodePools("master-node-pool-a", "master-node-pool-b"),

		frameworkSpec: &workloadsApi.DistributedSpecSpec{
			MinReplicas: *workloadsApi.NewNullableInt32(convutils.Ref(int32(1))),
			MaxReplicas: *workloadsApi.NewNullableInt32(convutils.Ref(int32(32))),
		},
	}),

	Entry("pytorch_x_mpi", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_PY_TORCH,
		workerSpecFname: "basic",
		masterSpecFname: "basic_master",
		expectedFname:   "pytorch_x_mpi",
		actualNodePools: makeTestNodePools("master-node-pool-a", "master-node-pool-b"),

		frameworkSpec: &workloadsApi.DistributedSpecSpec{
			SlotsPerWorker: *workloadsApi.NewNullableInt32(convutils.Ref(int32(12))),
		},
		expectedError: "slotsPerWorker is not supported by PyTorch framework",
	}),

	Entry("pytorch_x_mpi_2", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_PY_TORCH,
		workerSpecFname: "basic",
		masterSpecFname: "basic_master",
		actualNodePools: makeTestNodePools("master-node-pool-a", "master-node-pool-b"),

		frameworkSpec: &workloadsApi.DistributedSpecSpec{
			SshAuthMountPath: *workloadsApi.NewNullableString(convutils.Ref("/home/tmp")),
		},
		expectedError: "sshAuthMountPath is not supported by PyTorch framework",
	}),

	Entry("pytorch_invalid_range", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_PY_TORCH,
		workerSpecFname: "basic",
		masterSpecFname: "basic_master",
		expectedFname:   "pytorch_x_mpi",
		actualNodePools: makeTestNodePools("master-node-pool-a", "master-node-pool-b"),

		frameworkSpec: &workloadsApi.DistributedSpecSpec{
			MinReplicas: *workloadsApi.NewNullableInt32(convutils.Ref(int32(10))),
			MaxReplicas: *workloadsApi.NewNullableInt32(convutils.Ref(int32(4))),
		},
		expectedError: "maxReplicas should be no less than minReplicas",
	}),

	Entry("pytorch_no_max_replicas", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_PY_TORCH,
		workerSpecFname: "basic",
		masterSpecFname: "basic_master",
		expectedFname:   "pytorch_no_max_replicas",
		actualNodePools: makeTestNodePools("master-node-pool-a", "master-node-pool-b"),

		frameworkSpec: &workloadsApi.DistributedSpecSpec{
			MinReplicas: *workloadsApi.NewNullableInt32(convutils.Ref(int32(10))),
		},
	}),

	Entry("distributed workload with backoff limit", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_MPI,
		workerSpecFname: "backoffLimit",
		masterSpecFname: "master_no_gpu",
		expectedFname:   "backoffLimit_dist",
		actualNodePools: makeTestNodePools("my-node-pool-a", "master-node-pool-a", "master-node-pool-b"),
	}),

	Entry("distributed workload with clean pod policy", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_MPI,
		workerSpecFname: "cleanPodPolicy",
		masterSpecFname: "master_no_gpu",
		expectedFname:   "cleanPodPolicy_dist",
		actualNodePools: makeTestNodePools("my-node-pool-a", "my-node-pool-b", "master-node-pool-a", "master-node-pool-b"),

		frameworkSpec: &workloadsApi.DistributedSpecSpec{
			CleanPodPolicy: *workloadsApi.NewNullableCleanPodPolicy(convutils.Ref(workloadsApi.CLEANPODPOLICY_NONE)),
		},
	}),

	Entry("distributed mpi with launcherCreationPolicy", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_MPI,
		workerSpecFname: "launcherCreationPolicy",
		masterSpecFname: "master_no_gpu",
		expectedFname:   "launcherCreationPolicy_dist",
		clusterVersion:  "2.22.0",
		actualNodePools: makeTestNodePools("my-node-pool-a", "my-node-pool-b", "master-node-pool-a", "master-node-pool-b"),

		frameworkSpec: &workloadsApi.DistributedSpecSpec{
			MpiLauncherCreationPolicy: *workloadsApi.NewNullableMpiLauncherCreationPolicy(convutils.Ref(workloadsApi.MPILAUNCHERCREATIONPOLICY_AT_STARTUP)),
		},
	}),

	Entry("assets_dist", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_TF,
		workerSpecFname: "assets",
		masterSpecFname: "assets_master",
		expectedFname:   "assets_dist_tf",

		workerAssetInfo: &composer.ComplimentaryAssetInfo{
			Environment: workloadsApi.Environment{
				Name: "assets-worker-name",
				Id:   "assets-worker-id",
			},
			Datasources: composer.MapAssetKindToResources{
				composer_v2.AssetKind_HOST_PATH: []workloadsApi.Datasource{
					{
						Type: convutils.Ref(composer_v2.AssetKind_HOST_PATH),
						Id:   "hp1-id",
						Name: "hp1-name",
					},
					{
						Type: convutils.Ref(composer_v2.AssetKind_HOST_PATH),
						Id:   "hp2-id",
						Name: "hp2-name",
					},
				},
			},
		},

		masterAssetInfo: &composer.ComplimentaryAssetInfo{
			Environment: workloadsApi.Environment{
				Name: "assets-master-name",
				Id:   "assets-master-id",
			},
			Datasources: composer.MapAssetKindToResources{
				composer_v2.AssetKind_HOST_PATH: []workloadsApi.Datasource{
					{
						Type: convutils.Ref(composer_v2.AssetKind_HOST_PATH),
						Id:   "hp-master-id",
						Name: "hp-master-name",
					},
				},
			},
		},
	}),

	Entry("basic workspace with email requester", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_PY_TORCH,
		workerSpecFname: "basic",
		expectedFname:   "basic_ws_email_requester_pytorch",
		requester:       "test@run.ai",
	}),

	Entry("backoffLimit", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_TF,
		workerSpecFname: "backoffLimit",
		expectedFname:   "backoffLimit_tf",
	}),

	Entry("large_shm", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_XG_BOOST,
		workerSpecFname: "large_shm",
		expectedFname:   "large_shm_xgboost",
	}),

	Entry("basic node pools", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_JAX,
		noMaster:        true,
		workerSpecFname: "basic_node_pools",
		expectedFname:   "basic_node_pools_jax",
		actualNodePools: makeTestNodePools("my-node-pool-a", "my-node-pool-b", "my-node-pool-c"),
		// extra pool "my-node-pool-c" is available but not requested - should have no effect
	}),

	Entry("node pool not found - no pools known to cluster", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_MPI,
		workerSpecFname: "basic_node_pools",
		expectedError:   `node pool "my-node-pool-a" does not exist`,
	}),

	Entry("node pool not found - one of several missing", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_PY_TORCH,
		workerSpecFname: "basic_node_pools",
		expectedError:   `node pool "my-node-pool-b" does not exist`,
		actualNodePools: makeTestNodePools("my-node-pool-a"),
	}),

	Entry("basic workspace PC", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_TF,
		workerSpecFname: "basic_pc_preemptible",
		expectedFname:   "basic_pc_preemptible_tf",
	}),

	Entry("env empty", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_XG_BOOST,
		workerSpecFname: "env_empty_value",
		expectedFname:   "env_empty_value_xgboost",
	}),

	Entry("env empty", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_JAX,
		workerSpecFname: "env_no_name",
		expectedError:   "environmentVariables: name is required",
	}),

	Entry("env no value", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_MPI,
		workerSpecFname: "env_no_value",
		expectedError:   "environmentVariables 'HOME': value, userCredential, configMap, podFieldRef or secret must be set",
	}),

	Entry("completions", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_TF,
		workerSpecFname: "completions",
		expectedFname:   "completions_tf",
	}),

	Entry("category", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_XG_BOOST,
		workerSpecFname: "category",
		expectedFname:   "category_xgboost",
	}),

	Entry("annotations", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_JAX,
		noMaster:        true,
		workerSpecFname: "annotations",
		expectedFname:   "annotations_jax",
	}),

	Entry("annotations no name", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_MPI,
		workerSpecFname: "annotations_no_name",
		expectedError:   "annotations: name must be provided",
	}),

	Entry("labels no name", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_PY_TORCH,
		workerSpecFname: "labels_no_name",
		expectedError:   "labels: name must be provided",
	}),

	Entry("auto_deletion", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_TF,
		workerSpecFname: "auto_deletion",
		expectedFname:   "auto_deletion_tf",
	}),

	Entry("labels", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_XG_BOOST,
		workerSpecFname: "labels",
		expectedFname:   "labels_xgboost",
	}),

	Entry("rich", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_JAX,
		noMaster:        true,
		workerSpecFname: "rich",
		expectedFname:   "rich_jax",
	}),

	Entry("node_type", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_MPI,
		workerSpecFname: "node_type",
		expectedFname:   "node_type_mpi",
	}),

	Entry("ports", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_PY_TORCH,
		workerSpecFname: "ports",
		expectedFname:   "ports_pytorch",
	}),

	Entry("ports_multi_worker", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_PY_TORCH,
		workerSpecFname: "ports",
		frameworkSpec: &workloadsApi.DistributedSpecSpec{
			NumWorkers: *workloadsApi.NewNullableInt32(convutils.Ref(int32(2))),
		},
		expectedError: "ports are not supported for a distributed workload with more than one worker",
	}),

	Entry("urls", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_TF,
		workerSpecFname: "urls",
		expectedFname:   "urls_tf",
	}),

	Entry("urls_multi_worker", ComposerTest{
		framework: workloadsApi.DISTRIBUTEDFRAMEWORK_PY_TORCH,
		// noMaster: worker is the primary role and carries the URLs. Multi-worker
		// + URLs on the primary role is the unsupported case the guard catches.
		noMaster:        true,
		workerSpecFname: "urls_multi_worker",
		expectedError:   "exposedUrls are not supported for a distributed workload with more than one worker",
	}),

	Entry("urls_no_master", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_PY_TORCH,
		noMaster:        true,
		workerSpecFname: "urls_no_master",
		expectedFname:   "urls_no_master_pytorch",
	}),

	Entry("urls_jax", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_JAX,
		noMaster:        true,
		workerSpecFname: "urls_no_master",
		expectedFname:   "urls_jax",
	}),

	// OpenShift cluster: each role's URL composes to a Route (route.openshift.io/v1)
	// instead of an Ingress. Master is primary (no-suffix host); worker URLs carry
	// the role suffix to avoid host collision with master's no-suffix host.
	Entry("urls_openshift", ComposerTest{
		framework:          workloadsApi.DISTRIBUTEDFRAMEWORK_PY_TORCH,
		workerSpecFname:    "urls_no_master",
		expectedFname:      "urls_openshift_pytorch",
		clusterIsOpenshift: true,
	}),

	// Vanilla Kubernetes + Gateway-API enabled: each role's URL composes to an
	// Ingress AND an additive HTTPRoute. Same primary/non-primary suffix rules
	// apply to both Ingress hosts and HTTPRoute hostnames.
	Entry("urls_gateway", ComposerTest{
		framework:               workloadsApi.DISTRIBUTEDFRAMEWORK_PY_TORCH,
		workerSpecFname:         "urls_no_master",
		expectedFname:           "urls_gateway_pytorch",
		clusterGatewayName:      "runai-gateway",
		clusterGatewayNamespace: "runai",
	}),

	Entry("urls_syntax", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_XG_BOOST,
		workerSpecFname: "urls_syntax",
		expectedError:   "malformed url: parse \"http://example.com:badport\": invalid port",
	}),

	Entry("urls_host_syntax", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_JAX,
		workerSpecFname: "urls_host_syntax",
		expectedError:   "malformed url host",
	}),

	Entry("urls_public", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_MPI,
		workerSpecFname: "urls_authenticated_users",
		expectedFname:   "urls_public_mpi",
	}),

	Entry("urls_by_group_no_groups", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_PY_TORCH,
		workerSpecFname: "urls_by_groups_with_no_groups",
		expectedError:   "exposedUrls: authorizedGroups must be set when authorizationType is authorizedGroups",
	}),

	Entry("urls_by_users_no_users", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_TF,
		workerSpecFname: "urls_by_users_no_users",
		expectedError:   "exposedUrls: authorizedUsers must be set when authorizationType is authorizedUsers",
	}),

	Entry("urls_not_abs", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_XG_BOOST,
		workerSpecFname: "urls_not_abs",
		expectedFname:   "urls_not_abs_xgboost",
		expectedError:   "Invalid url <https://compute.on-demand.io> must be an absolute path",
	}),

	Entry("urls_invalid", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_JAX,
		workerSpecFname: "urls_invalid_authorized",
		expectedError:   "authorizedUsers and authorizedGroups are mutually exclusive",
	}),

	Entry("ports_and_urls", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_MPI,
		workerSpecFname: "ports_and_urls",
		expectedFname:   "ports_and_urls_mpi",
	}),

	Entry("priority_class_build", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_PY_TORCH,
		workerSpecFname: "priority_class_build",
		expectedFname:   "priority_class_build_pytorch",
	}),

	Entry("priority_class_build_tr", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_TF,
		workerSpecFname: "priority_class_build",
		expectedFname:   "priority_class_build_tf",
	}),

	Entry("priority_class_preemptible_tr", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_XG_BOOST,
		workerSpecFname: "priority_class_preemptible",
		expectedFname:   "priority_class_preemptible_xgboost",
	}),

	Entry("priority_class_train", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_JAX,
		noMaster:        true,
		workerSpecFname: "priority_class_train",
		expectedFname:   "priority_class_train_jax",
	}),

	Entry("priority_class_preemptible", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_MPI,
		workerSpecFname: "priority_class_preemptible",
		expectedFname:   "priority_class_preemptible_mpi",
	}),

	Entry("preemptibility", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_PY_TORCH,
		workerSpecFname: "preemptibility",
		expectedFname:   "preemptibility_pytorch",
	}),

	Entry("semi-preemptibility", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_PY_TORCH,
		workerSpecFname: "semi_preemptibility",
		expectedFname:   "semi_preemptibility_pytorch",
	}),

	Entry("pod affinity - required", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_TF,
		workerSpecFname: "pod_affinity_required",
		expectedFname:   "pod_affinity_required_tf",
	}),

	Entry("privileged true", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_XG_BOOST,
		workerSpecFname: "privileged_true",
		expectedFname:   "privileged_true_xgboost",
	}),

	Entry("privileged true and allow privileged false", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_JAX,
		workerSpecFname: "privileged_and_allow_privilege_escalation",
		expectedError:   "allowPrivilegeEscalation can't be false when privileged is true",
	}),

	Entry("privileged true and run as non root true", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_MPI,
		workerSpecFname: "privileged_and_run_as_non_root",
		expectedError:   "runAsNonRoot can't be true when privileged is true",
	}),

	Entry("pod affinity - preferred", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_PY_TORCH,
		workerSpecFname: "pod_affinity_preferred",
		expectedFname:   "pod_affinity_preferred_pytorch",
	}),

	Entry("pod affinity - no type", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_TF,
		workerSpecFname: "pod_affinity_no_type",
		expectedError:   "podAffinity: type must be provided",
	}),

	Entry("pod affinity - no key", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_XG_BOOST,
		workerSpecFname: "pod_affinity_no_key",
		expectedError:   "podAffinity: key must be provided",
	}),

	Entry("tolerations", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_JAX,
		noMaster:        true,
		workerSpecFname: "tolerations",
		expectedFname:   "tolerations_jax",
	}),

	Entry("tolerations - Any effect", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_MPI,
		workerSpecFname: "tolerations_effect_any",
		expectedFname:   "tolerations_effect_any_mpi",
	}),

	Entry("tolerations - invalid - no operator", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_PY_TORCH,
		workerSpecFname: "tolerations_invalid_no_operator",
		expectedError:   "tolerations: operator must be provided",
	}),

	Entry("tolerations - invalid - no key", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_TF,
		workerSpecFname: "tolerations_invalid_no_key",
		expectedError:   "tolerations: key must be provided",
	}),

	Entry("tolerations - invalid - no effect", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_XG_BOOST,
		workerSpecFname: "tolerations_invalid_no_effect",
		expectedError:   "tolerations: effect must be provided",
	}),

	Entry("tolerations - invalid - equal operator without value", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_JAX,
		workerSpecFname: "tolerations_invalid_equal_without_value",
		expectedError:   "tolerations: value must be provided when operator is set to Equal",
	}),

	Entry("tolerations - invalid - exists operator with value", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_MPI,
		workerSpecFname: "tolerations_invalid_exists_with_value",
		expectedError:   "tolerations: value is not allowed when operator is set to Exists",
	}),

	Entry("tolerations - invalid - seconds with wrong effect", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_PY_TORCH,
		workerSpecFname: "tolerations_invalid_seconds_with_wrong_effect",
		expectedError:   "tolerations: seconds is not allowed when effect is not set to NoExecute",
	}),

	Entry("probes", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_TF,
		workerSpecFname: "probes",
		expectedFname:   "probes_tf",
	}),

	Entry("probes - only required fields", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_XG_BOOST,
		workerSpecFname: "probes_only_required",
		expectedFname:   "probes_only_required_xgboost",
	}),

	Entry("probes - invalid - no handler", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_JAX,
		workerSpecFname: "probes_invalid_no_handler",
		expectedError:   "probes.readiness: handler must be provided",
	}),

	Entry("probes - invalid - httpGet handler without port", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_MPI,
		workerSpecFname: "probes_invalid_no_httpget_port",
		expectedError:   "probes.readiness.handler.httpGet: port must be provided",
	}),

	Entry("secret_and_config_env", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_PY_TORCH,
		workerSpecFname: "secret_and_config_env",
		expectedFname:   "secret_and_config_env_pytorch",
	}),

	Entry("user_credentials_env", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_TF,
		workerSpecFname: "user_credentials_env",
		expectedFname:   "user_credentials_env_tf",
	}),

	Entry("user_credentials_env_no_asset", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_XG_BOOST,
		workerSpecFname: "user_credentials_env_no_asset",
		expectedError:   "asset mapping not found for credential name: my-user-credential-not-found",
	}),

	Entry("secret_env_err", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_JAX,
		workerSpecFname: "secret_env_err",
		expectedError:   "environmentVariables 'PASSWORD': value, userCredential, configMap, podFieldRef and secret are mutually exclusive",
	}),

	Entry("config_env_err", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_MPI,
		workerSpecFname: "config_env_err",
		expectedError:   "environmentVariables 'PASSWORD': value, userCredential, configMap, podFieldRef and secret are mutually exclusive",
	}),

	Entry("pod_field_ref_env_err", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_PY_TORCH,
		workerSpecFname: "pod_field_ref_env_err",
		expectedError:   "environmentVariables 'PASSWORD': value, userCredential, configMap, podFieldRef and secret are mutually exclusive",
	}),

	Entry("node_affinity", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_TF,
		workerSpecFname: "node_affinity",
		expectedFname:   "node_affinity_tf",
	}),

	Entry("node_affinity_invalid", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_XG_BOOST,
		workerSpecFname: "node_affinity_invalid",
		expectedError:   "nodeAffinityRequired: Values must be empty when Operator is Exists",
	}),

	Entry("image_pull_secrets", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_JAX,
		noMaster:        true,
		workerSpecFname: "image_pull_secrets",
		expectedFname:   "image_pull_secrets_jax",
	}),

	Entry("xgboost with master", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_XG_BOOST,
		workerSpecFname: "basic",
		expectedFname:   "basic_xgboost",
	}),

	//
	// noMaster positive tests — PyTorch and TF support workers-only.
	//
	Entry("no master pytorch", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_PY_TORCH,
		noMaster:        true,
		workerSpecFname: "basic",
		expectedFname:   "no_master_pytorch",
	}),

	Entry("no master tf", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_TF,
		noMaster:        true,
		workerSpecFname: "basic",
		expectedFname:   "no_master_tf",
	}),

	//
	// noMaster negative tests — MPI and XGBoost require a master/launcher.
	//
	Entry("no master mpi rejected", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_MPI,
		noMaster:        true,
		workerSpecFname: "basic",
		expectedError:   "noMaster is not supported for MPI framework",
	}),

	Entry("no master xgboost rejected", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_XG_BOOST,
		noMaster:        true,
		workerSpecFname: "basic",
		expectedError:   "noMaster is not supported for XGBoost framework",
	}),

	//
	// MPI master-with-GPU rejection. Mirrors the cluster's masterResourceGPUCheck
	// at the composer level so the violation is caught early in the CP path.
	//
	Entry("mpi master with gpu rejected", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_MPI,
		workerSpecFname: "basic",
		masterSpecFname: "basic_master",
		expectedError:   "GPU resources are not allowed on MPI master pod",
	}),

	//
	// MPI slotsPerWorker fallback from the worker's whole-GPU portion
	// (gpuPortionRequest >= 1, the first branch of the cluster MPI builder
	// fallback). The gpuDevices fallback branch is covered separately by
	// the existing compute_portion_2_devices_req test.
	//
	Entry("mpi slots per worker from gpu portion", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_MPI,
		workerSpecFname: "gpu_portion_whole",
		masterSpecFname: "master_no_gpu",
		expectedFname:   "mpi_slots_from_gpu_portion",
	}),

	//
	// PyTorch NprocPerNode from the worker's whole-GPU portion
	// (gpuPortionRequest >= 1). The gpuDevices path is already covered by
	// the existing gpu_multiple / multiple_gpus tests.
	//
	Entry("pytorch nproc per node from gpu portion", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_PY_TORCH,
		noMaster:        true,
		workerSpecFname: "gpu_portion_whole",
		expectedFname:   "pytorch_nproc_from_gpu_portion",
	}),

	//
	// MPI multi-worker — verifies the worker `replicas` field scales and
	// that the RUNAI_MPI_NUM_WORKERS env injection picks up the worker's
	// numWorkers (not the master pass's nil default).
	//
	Entry("mpi multi worker", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_MPI,
		workerSpecFname: "basic",
		masterSpecFname: "master_no_gpu",
		expectedFname:   "mpi_multi_worker",

		frameworkSpec: &workloadsApi.DistributedSpecSpec{
			NumWorkers: *workloadsApi.NewNullableInt32(convutils.Ref(int32(4))),
		},
	}),
},
	append(computeTests,
		append(storageTests,
			append(storageClassConfigTests,
				mnnvlTests...,
			)...,
		)...,
	)...,
)

var lettersAndColon = regexp.MustCompile(`([A-Za-z]+:)`)

type ActualResults struct {
	DistributedTarget *DistributedTarget
	Notification      *workloadsApi.WorkloadCreationFields
}

func RunComposerTest(testDetails ComposerTest, basePath string) {
	testBrief := fmt.Sprintf("%s/%s, Framework: %s", testDetails.workerSpecFname, testDetails.masterSpecFname, testDetails.framework)
	workloadName := testDetails.workerSpecFname

	testDetailsStr := lettersAndColon.ReplaceAllString(fmt.Sprintf("%+v", testDetails), "\n\t$1")

	var actualResults ActualResults
	var actualRequirements composer.WorkloadRequirements

	distributedWorkload := workloadsApi.DistributedWorkload{}

	if testDetails.workerSpecFname != testutils.NA {
		specYaml, err := testutils.LoadInputFile(testDetails.workerSpecFname, basePath)
		Expect(err).To(BeNil(), fmt.Sprintf("failed to load of %s.yaml", testDetails.workerSpecFname))

		distributedWorkload.Spec = &workloadsApi.DistributedSpecSpec{}

		err = yaml.Unmarshal(specYaml, &distributedWorkload.Spec)
		if err != nil {
			if os.IsNotExist(err) {
				err = fmt.Errorf("file %s.yaml not found", testDetails.workerSpecFname)
			}
		}
		Expect(err).To(BeNil(), fmt.Sprintf("failed to parse of %s.yaml", testDetails.workerSpecFname))

		if testDetails.mnnvlLabel != "" {
			distributedWorkload.Spec.Labels = append(distributedWorkload.Spec.Labels, workloadsApi.Label{
				Name:  *workloadsApi.NewNullableString(convutils.Ref("run.ai/MNNVL")),
				Value: *workloadsApi.NewNullableString(convutils.Ref(testDetails.mnnvlLabel)),
			})
		}
	}

	if testDetails.masterSpecFname != testutils.NA {
		specYaml, err := testutils.LoadInputFile(testDetails.masterSpecFname, basePath)
		Expect(err).To(BeNil(), fmt.Sprintf("failed to load of %s.yaml", testDetails.masterSpecFname))

		masterSpec := workloadsApi.MasterSpec{}
		err = yaml.Unmarshal(specYaml, &masterSpec)
		if err != nil {
			if os.IsNotExist(err) {
				err = fmt.Errorf("file %s.yaml not found", testDetails.masterSpecFname)
			}
		}
		Expect(err).To(BeNil(), fmt.Sprintf("failed to parse of %s.yaml", testDetails.masterSpecFname))

		distributedWorkload.MasterSpec = *workloadsApi.NewNullableMasterSpec(&masterSpec)
	} else {
		if !testDetails.noMaster && !testDetails.leaveMasterSpecUnset {
			masterSpec := workloadsApi.MasterSpec{}
			copyErr := convutils.CopySpecs(&masterSpec, &distributedWorkload.Spec)
			Expect(copyErr).To(BeNil())

			distributedWorkload.MasterSpec = *workloadsApi.NewNullableMasterSpec(&masterSpec)
		}
	}
	if testDetails.masterSpecSameAsWorker {
		distributedWorkload.MasterSpecSameAsWorker = *workloadsApi.NewNullableBool(convutils.Ref(true))
	}

	// add framework-specific spec params such as slotsPerWorker, etc.
	if testDetails.frameworkSpec != nil {
		if testDetails.workerSpecFname == testutils.NA {
			distributedWorkload.Spec = testDetails.frameworkSpec
		} else {
			if convutils.HasValue(testDetails.frameworkSpec.MinReplicas) {
				distributedWorkload.Spec.MinReplicas = testDetails.frameworkSpec.MinReplicas
			}
			if convutils.HasValue(testDetails.frameworkSpec.MaxReplicas) {
				distributedWorkload.Spec.MaxReplicas = testDetails.frameworkSpec.MaxReplicas
			}
			if convutils.HasValue(testDetails.frameworkSpec.SlotsPerWorker) {
				distributedWorkload.Spec.SlotsPerWorker = testDetails.frameworkSpec.SlotsPerWorker
			}
			if convutils.HasValue(testDetails.frameworkSpec.SshAuthMountPath) {
				distributedWorkload.Spec.SshAuthMountPath = testDetails.frameworkSpec.SshAuthMountPath
			}
			if convutils.HasValue(testDetails.frameworkSpec.MpiLauncherCreationPolicy) {
				distributedWorkload.Spec.MpiLauncherCreationPolicy = testDetails.frameworkSpec.MpiLauncherCreationPolicy
			}
			if convutils.HasValue(testDetails.frameworkSpec.CleanPodPolicy) {
				distributedWorkload.Spec.CleanPodPolicy = testDetails.frameworkSpec.CleanPodPolicy
			}
			if convutils.HasValue(testDetails.frameworkSpec.NumWorkers) {
				distributedWorkload.Spec.NumWorkers = testDetails.frameworkSpec.NumWorkers
			}
			if convutils.HasValue(testDetails.frameworkSpec.BackoffLimit) {
				distributedWorkload.Spec.BackoffLimit = testDetails.frameworkSpec.BackoffLimit
			}
		}
	}

	distributedWorkload.Spec.DistributedFramework.Set(&testDetails.framework)

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

	requester := testDetails.requester
	if requester == "" {
		requester = "tester"
	}

	notificationsComposer := notifications.NewNotificationsComposer(
		string(workloadsApi.WORKLOADTYPE_DISTRIBUTED),
		workloadName,
		fmt.Sprintf("id-%s", testBrief),
		requester,
		"cluster-123-456",
		"project-654-321",
		"department-456-123",
		time.Time{},
		nil,
		options).WithConnectionsPreProcessor("project", composer.NetworkingResourcePrefixKubeflow)

	var err *workloadsApi.Error

	distComposer := NewDistributedToKubeflowComposer(
		string(workloadsApi.WORKLOADTYPE_DISTRIBUTED),
		workloadName,
		"test",
		"runai-test",
		requester,
		assetsMapping,
		assetsMapping,
		options,
		testDetails.actualNodePools)

	distComposer.WithStorageClassConfiguration(storageClassConfig)
	distComposer.WithDataVolumeClaimNames(testDetails.dataVolumeClaimNames)
	clusterOpts := testClusterOpts{
		isOpenshift:                testDetails.clusterIsOpenshift,
		gatewayName:                testDetails.clusterGatewayName,
		gatewayNamespace:           testDetails.clusterGatewayNamespace,
		distGPUNetworkAcceleration: testDetails.clusterGPUNetworkAcceleration,
	}
	distComposer.WithClusterStorageConfig(
		testClusterStorageConfigWithOpts("https://mock-q7z3k.example.com", testDetails.clusterSubdomain, clusterOpts))

	hasMasterReplica := !testDetails.noMaster
	actualResults.DistributedTarget = NewMinimalDistributedTarget(testDetails.framework, hasMasterReplica)

	err = distComposer.Compose(context.Background(), &distributedWorkload, actualResults.DistributedTarget)()
	if err == nil {
		supersetSpec := workloadsApi.SupersetSpec{}
		copyErr := convutils.CopySpecs(&supersetSpec, &distributedWorkload.Spec)
		Expect(copyErr).To(BeNil())
		err = notificationsComposer.Compose(context.Background(), nil, testDetails.workerAssetInfo, &supersetSpec)()
	}

	actualRequirements = distComposer.WorkloadRequirements

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

	// remove the global pointers to prevent duplication in the expected file
	actualResults.DistributedTarget.WorkerSpec = nil
	actualResults.DistributedTarget.MasterSpec = nil
	actualResults.DistributedTarget.TypeMeta = nil
	actualResults.DistributedTarget.ObjectMeta = nil
	actualResults.DistributedTarget.RunPolicy = nil

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

func newTestAnnotation(name, value string) workloadsApi.Annotation {
	annotation := workloadsApi.NewAnnotation()
	annotation.SetName(name)
	annotation.SetValue(value)
	return *annotation
}

func composePyTorchAnnotationTestWorkload(workerSpec *workloadsApi.DistributedSpecSpec) (*DistributedTarget, *workloadsApi.Error) {
	framework := workloadsApi.DISTRIBUTEDFRAMEWORK_PY_TORCH
	workerSpec.DistributedFramework.Set(&framework)
	workerSpec.Image.Set(convutils.Ref("ubuntu"))

	workload := &workloadsApi.DistributedWorkload{
		Spec:                   workerSpec,
		MasterSpecSameAsWorker: *workloadsApi.NewNullableBool(convutils.Ref(true)),
	}
	target := NewMinimalDistributedTarget(framework, true)
	distComposer := NewDistributedToKubeflowComposer(
		string(workloadsApi.WORKLOADTYPE_DISTRIBUTED),
		"pytorch-annotations",
		"test",
		"runai-test",
		"tester",
		nil,
		nil,
		composer.FieldsComposerOptions{},
		nil,
	)

	return target, distComposer.Compose(context.Background(), workload, target)()
}

var _ = Describe("distributed composer template variable expansion", func() {
	It("expands RUNAI placeholders in command and args", func() {
		framework := workloadsApi.DISTRIBUTEDFRAMEWORK_PY_TORCH
		var workerSpec workloadsApi.DistributedSpecSpec
		Expect(json.Unmarshal([]byte(`{
			"command": "serve.sh",
			"args": "--project=${RUNAI_PROJECT} --job=${RUNAI_JOB_NAME} --uid=${RUNAI_UID} --gid=${RUNAI_GID}",
			"security": {"uidGidSource": "custom", "runAsUid": 501, "runAsGid": 20}
		}`), &workerSpec)).To(Succeed())
		workerSpec.DistributedFramework.Set(&framework)
		workerSpec.Image.Set(convutils.Ref("ubuntu"))

		workload := &workloadsApi.DistributedWorkload{
			Spec:                   &workerSpec,
			MasterSpecSameAsWorker: *workloadsApi.NewNullableBool(convutils.Ref(true)),
		}
		target := NewMinimalDistributedTarget(framework, true)
		distComposer := NewDistributedToKubeflowComposer(
			string(workloadsApi.WORKLOADTYPE_DISTRIBUTED),
			"job-3", "team-a", "runai-test", "tester", nil, nil,
			composer.FieldsComposerOptions{}, nil,
		)
		Expect(distComposer.Compose(context.Background(), workload, target)()).To(BeNil())

		workerContainer := target.WorkerSpec.Template.Spec.Containers[0]
		Expect(workerContainer.Command).To(Equal([]string{"serve.sh"}))
		Expect(workerContainer.Args).To(Equal([]string{
			"--project=team-a",
			"--job=job-3",
			"--uid=501",
			"--gid=20",
		}))
	})
})

var _ = Describe("distributed composer validations", func() {
	Context("segmented PyTorch annotations", func() {
		DescribeTable("projects explicit workload topology to the job and keeps segment topology on workers",
			func(rootPlacementAnnotation, rootPlacementLevel string) {
				workerSpec := &workloadsApi.DistributedSpecSpec{
					Annotations: []workloadsApi.Annotation{
						newTestAnnotation(topology.KAITopologyAnnotationKey, "fabric-a"),
						newTestAnnotation(rootPlacementAnnotation, rootPlacementLevel),
						newTestAnnotation(topology.KAISegmentSizeAnnotationKey, "4"),
						newTestAnnotation(topology.KAISegmentTopologyPreferredAnnotationKey, "rack"),
						newTestAnnotation("example.com/unrelated", "kept"),
					},
				}

				target, composeErr := composePyTorchAnnotationTestWorkload(workerSpec)
				Expect(composeErr).To(BeNil())

				Expect(target.PyTorchJob.Annotations).To(Equal(map[string]string{
					composer.SubmittedByAnnotation:    "tester",
					topology.KAITopologyAnnotationKey: "fabric-a",
					rootPlacementAnnotation:           rootPlacementLevel,
				}))

				workerAnnotations := target.WorkerSpec.Template.Annotations
				Expect(workerAnnotations).To(HaveKeyWithValue(topology.KAITopologyAnnotationKey, "fabric-a"))
				Expect(workerAnnotations).NotTo(HaveKey(topology.KAITopologyAnnotationPreferredKey))
				Expect(workerAnnotations).NotTo(HaveKey(topology.KAITopologyAnnotationRequiredKey))
				Expect(workerAnnotations).To(HaveKeyWithValue(topology.KAISegmentSizeAnnotationKey, "4"))
				Expect(workerAnnotations).To(HaveKeyWithValue(topology.KAISegmentTopologyPreferredAnnotationKey, "rack"))
				Expect(workerAnnotations).To(HaveKeyWithValue("example.com/unrelated", "kept"))

				masterAnnotations := target.MasterSpec.Template.Annotations
				Expect(masterAnnotations).NotTo(HaveKey(topology.KAITopologyAnnotationKey))
				Expect(masterAnnotations).NotTo(HaveKey(topology.KAITopologyAnnotationPreferredKey))
				Expect(masterAnnotations).NotTo(HaveKey(topology.KAITopologyAnnotationRequiredKey))
				Expect(masterAnnotations).To(HaveKeyWithValue("example.com/unrelated", "kept"))
				Expect(masterAnnotations).NotTo(HaveKey(topology.KAISegmentSizeAnnotationKey))
				Expect(masterAnnotations).NotTo(HaveKey(topology.KAISegmentTopologyPreferredAnnotationKey))
				Expect(masterAnnotations).NotTo(HaveKey(topology.KAISegmentTopologyRequiredAnnotationKey))
			},
			Entry("preferred placement", topology.KAITopologyAnnotationPreferredKey, "rack"),
			Entry("required placement", topology.KAITopologyAnnotationRequiredKey, "zone"),
		)

		It("does not project a topology name without workload placement", func() {
			workerSpec := &workloadsApi.DistributedSpecSpec{
				Annotations: []workloadsApi.Annotation{
					newTestAnnotation(topology.KAITopologyAnnotationKey, "fabric-a"),
					newTestAnnotation(topology.KAISegmentSizeAnnotationKey, "4"),
					newTestAnnotation(topology.KAISegmentTopologyPreferredAnnotationKey, "rack"),
				},
			}

			target, composeErr := composePyTorchAnnotationTestWorkload(workerSpec)
			Expect(composeErr).To(BeNil())

			Expect(target.PyTorchJob.Annotations).To(Equal(map[string]string{
				composer.SubmittedByAnnotation: "tester",
			}))

			workerAnnotations := target.WorkerSpec.Template.Annotations
			Expect(workerAnnotations).To(HaveKeyWithValue(topology.KAITopologyAnnotationKey, "fabric-a"))
			Expect(workerAnnotations).To(HaveKeyWithValue(topology.KAISegmentSizeAnnotationKey, "4"))
			Expect(workerAnnotations).To(HaveKeyWithValue(topology.KAISegmentTopologyPreferredAnnotationKey, "rack"))

			masterAnnotations := target.MasterSpec.Template.Annotations
			Expect(masterAnnotations).NotTo(HaveKey(topology.KAITopologyAnnotationKey))
			Expect(masterAnnotations).NotTo(HaveKey(topology.KAISegmentSizeAnnotationKey))
			Expect(masterAnnotations).NotTo(HaveKey(topology.KAISegmentTopologyPreferredAnnotationKey))
		})

		It("propagates the AJR opt-out annotation onto the job's own metadata", func() {
			workerSpec := &workloadsApi.DistributedSpecSpec{
				Annotations: []workloadsApi.Annotation{
					newTestAnnotation(constants.DisableAJRAnnotation, "true"),
				},
			}

			target, composeErr := composePyTorchAnnotationTestWorkload(workerSpec)
			Expect(composeErr).To(BeNil())

			Expect(target.PyTorchJob.Annotations).To(HaveKeyWithValue(constants.DisableAJRAnnotation, "true"))
		})

		It("does not add the AJR opt-out annotation when it is absent", func() {
			target, composeErr := composePyTorchAnnotationTestWorkload(&workloadsApi.DistributedSpecSpec{})
			Expect(composeErr).To(BeNil())

			Expect(target.PyTorchJob.Annotations).NotTo(HaveKey(constants.DisableAJRAnnotation))
		})

		It("projects topology annotations and elasticity for segmented elastic PyTorch", func() {
			workerSpec := &workloadsApi.DistributedSpecSpec{
				Annotations: []workloadsApi.Annotation{
					newTestAnnotation(topology.KAITopologyAnnotationKey, "fabric-a"),
					newTestAnnotation(topology.KAITopologyAnnotationPreferredKey, "zone"),
					newTestAnnotation(topology.KAISegmentSizeAnnotationKey, "4"),
					newTestAnnotation(topology.KAISegmentTopologyPreferredAnnotationKey, "rack"),
				},
			}
			workerSpec.SetMinReplicas(4)
			workerSpec.SetMaxReplicas(8)

			target, composeErr := composePyTorchAnnotationTestWorkload(workerSpec)
			Expect(composeErr).To(BeNil())

			Expect(target.PyTorchJob.Annotations).To(Equal(map[string]string{
				composer.SubmittedByAnnotation:             "tester",
				topology.KAITopologyAnnotationKey:          "fabric-a",
				topology.KAITopologyAnnotationPreferredKey: "zone",
			}))

			workerAnnotations := target.WorkerSpec.Template.Annotations
			Expect(workerAnnotations).To(HaveKeyWithValue(topology.KAITopologyAnnotationKey, "fabric-a"))
			Expect(workerAnnotations).NotTo(HaveKey(topology.KAITopologyAnnotationPreferredKey))
			Expect(workerAnnotations).To(HaveKeyWithValue(topology.KAISegmentSizeAnnotationKey, "4"))
			Expect(workerAnnotations).To(HaveKeyWithValue(topology.KAISegmentTopologyPreferredAnnotationKey, "rack"))

			masterAnnotations := target.MasterSpec.Template.Annotations
			Expect(masterAnnotations).NotTo(HaveKey(topology.KAITopologyAnnotationKey))
			Expect(masterAnnotations).NotTo(HaveKey(topology.KAITopologyAnnotationPreferredKey))
			Expect(masterAnnotations).NotTo(HaveKey(topology.KAISegmentSizeAnnotationKey))
			Expect(masterAnnotations).NotTo(HaveKey(topology.KAISegmentTopologyPreferredAnnotationKey))

			Expect(target.PyTorchJob.Spec.ElasticPolicy).NotTo(BeNil())
			Expect(target.PyTorchJob.Spec.ElasticPolicy.MinReplicas).To(Equal(convutils.Ref(int32(4))))
			Expect(target.PyTorchJob.Spec.ElasticPolicy.MaxReplicas).To(Equal(convutils.Ref(int32(8))))
		})

		It("treats an empty segment size as absent", func() {
			workerSpec := &workloadsApi.DistributedSpecSpec{
				Annotations: []workloadsApi.Annotation{
					newTestAnnotation(topology.KAITopologyAnnotationKey, "fabric-a"),
					newTestAnnotation(topology.KAISegmentSizeAnnotationKey, ""),
				},
			}

			target, composeErr := composePyTorchAnnotationTestWorkload(workerSpec)
			Expect(composeErr).To(BeNil())

			Expect(target.PyTorchJob.Annotations).NotTo(HaveKey(topology.KAITopologyAnnotationKey))
			for _, replicaAnnotations := range []map[string]string{
				target.WorkerSpec.Template.Annotations,
				target.MasterSpec.Template.Annotations,
			} {
				Expect(replicaAnnotations).To(HaveKeyWithValue(topology.KAITopologyAnnotationKey, "fabric-a"))
				Expect(replicaAnnotations).To(HaveKeyWithValue(topology.KAISegmentSizeAnnotationKey, ""))
			}
		})

		It("treats explicitly null elasticity bounds as non-elastic", func() {
			workerSpec := &workloadsApi.DistributedSpecSpec{
				Annotations: []workloadsApi.Annotation{
					newTestAnnotation(topology.KAITopologyAnnotationKey, "fabric-a"),
					newTestAnnotation(topology.KAISegmentSizeAnnotationKey, "2"),
					newTestAnnotation(topology.KAISegmentTopologyPreferredAnnotationKey, "rack"),
				},
			}
			workerSpec.SetMinReplicasNil()
			workerSpec.SetMaxReplicasNil()

			target, composeErr := composePyTorchAnnotationTestWorkload(workerSpec)
			Expect(composeErr).To(BeNil())

			Expect(target.PyTorchJob.Annotations).NotTo(HaveKey(topology.KAITopologyAnnotationKey))
			Expect(target.WorkerSpec.Template.Annotations).To(HaveKeyWithValue(topology.KAITopologyAnnotationKey, "fabric-a"))
			Expect(target.WorkerSpec.Template.Annotations).To(HaveKeyWithValue(topology.KAISegmentSizeAnnotationKey, "2"))
			Expect(target.WorkerSpec.Template.Annotations).To(HaveKeyWithValue(topology.KAISegmentTopologyPreferredAnnotationKey, "rack"))
			Expect(target.MasterSpec.Template.Annotations).NotTo(HaveKey(topology.KAITopologyAnnotationKey))
			Expect(target.MasterSpec.Template.Annotations).NotTo(HaveKey(topology.KAISegmentSizeAnnotationKey))
			Expect(target.MasterSpec.Template.Annotations).NotTo(HaveKey(topology.KAISegmentTopologyPreferredAnnotationKey))
		})
	})

	Context("kubeflow networking roles", func() {
		DescribeTable("maps CP replica roles to resource names and pod selectors",
			func(framework workloadsApi.DistributedFramework, replicaType workloadsApi.ReplicaType, expectedResourceRole string, expectedSelectorReplicaType string) {
				role := kubeflowNetworkingRoleForReplica(framework, replicaType)

				Expect(role.resourceRole).To(Equal(expectedResourceRole))
				Expect(role.selectorReplicaType).To(Equal(expectedSelectorReplicaType))
			},
			Entry("TF master keeps master resource names but selects chief pods", workloadsApi.DISTRIBUTEDFRAMEWORK_TF, workloadsApi.REPLICATYPE_MASTER, "master", "chief"),
			Entry("PyTorch master selects master pods", workloadsApi.DISTRIBUTEDFRAMEWORK_PY_TORCH, workloadsApi.REPLICATYPE_MASTER, "master", "master"),
			Entry("MPI master uses launcher resources and selectors", workloadsApi.DISTRIBUTEDFRAMEWORK_MPI, workloadsApi.REPLICATYPE_MASTER, "launcher", "launcher"),
			Entry("worker uses worker resources and selectors", workloadsApi.DISTRIBUTEDFRAMEWORK_TF, workloadsApi.REPLICATYPE_WORKER, "worker", "worker"),
		)
	})

	Context("kubeflow port service options", func() {
		DescribeTable("uses resource role for names and selector role for pod selection",
			func(framework workloadsApi.DistributedFramework, replicaType workloadsApi.ReplicaType, expectedNameQualifier string, expectedSelectorKey string, expectedSelectorValue string) {
				role := kubeflowNetworkingRoleForReplica(framework, replicaType)
				options := kubeflowPortsServiceOptions("dist", framework, role, 0)

				Expect(options.NameQualifier).To(Equal(expectedNameQualifier))
				Expect(options.Selector).To(HaveKeyWithValue("training.kubeflow.org/job-name", "dist"))
				Expect(options.Selector).To(HaveKeyWithValue(expectedSelectorKey, expectedSelectorValue))
			},
			Entry("TF master keeps master resource names but selects chief pods", workloadsApi.DISTRIBUTEDFRAMEWORK_TF, workloadsApi.REPLICATYPE_MASTER, "master", "training.kubeflow.org/replica-type", "chief"),
			Entry("PyTorch master keeps master resource names and selectors", workloadsApi.DISTRIBUTEDFRAMEWORK_PY_TORCH, workloadsApi.REPLICATYPE_MASTER, "master", "training.kubeflow.org/replica-type", "master"),
			Entry("MPI master keeps launcher resource names and selectors", workloadsApi.DISTRIBUTEDFRAMEWORK_MPI, workloadsApi.REPLICATYPE_MASTER, "launcher", "training.kubeflow.org/job-role", "launcher"),
			Entry("worker keeps worker resource names and selectors", workloadsApi.DISTRIBUTEDFRAMEWORK_TF, workloadsApi.REPLICATYPE_WORKER, "worker", "training.kubeflow.org/replica-type", "worker"),
		)
	})

	Context("input VS expected tests", func() {
		DescribeTable("should create proper distributed workloads or return sensible errors", func(testDetails ComposerTest) {
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
