/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dinf_to_lws

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/ghodss/yaml"
	clusterApi "github.com/run-ai/runai/control-plane/cluster-service/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/composer"
	workloadsApi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
)

const (
	distInfGoldenInputDir    = "testing/inputs"
	distInfGoldenExpectedDir = "testing/expected"
	distInfLegacyCaseCount   = 54
)

var distInfExpectedErrors = map[string]string{
	"dinfw_no_leader":                                   "leader spec is required",
	"dinfw_no_leader_image":                             "image",
	"dinfw_no_serving_port":                             "servingPort is required",
	"dinfw_no_worker_image":                             "image",
	"dinfw_restart_policy_case_mismatch":                "not a valid DistributedInferenceRestartPolicy",
	"dinfw_serving_port_authorized_groups_without_list": "authorizedGroups",
	"dinfw_serving_port_authorized_users_without_list":  "authorizedUsers",
	"dinfw_serving_port_empty_authorized_groups":        "authorizedGroups",
	"dinfw_serving_port_empty_authorized_users":         "authorizedUsers",
	"dinfw_serving_port_exposed_url_with_expose_false":  "exposedUrl should not be set",
	"dinfw_serving_port_grpc":                           "not a valid DistributedInferenceServingPortProtocol",
	"dinfw_serving_port_tcp":                            "not a valid DistributedInferenceServingPortProtocol",
	"dinfw_serving_port_unsupported_auth_type":          "not a valid DistributedInferenceServingPortAccessAuthorizationTypeEnum",
	"dinfw_worker_spec_provided_workers_not_set":        "worker spec requires workers",
	"dinfw_worker_spec_provided_workers_zero":           "worker spec requires workers",
	"dinfw_workers_not_set_worker_spec_provided":        "worker spec requires workers",
	"dinfw_workers_set_zero_worker_spec_provided":       "worker spec requires workers",
}

func TestDistributedInferenceLegacyTypedGoldenCases(t *testing.T) {
	entries, err := os.ReadDir(distInfGoldenInputDir)
	require.NoError(t, err)

	var names []string
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".yaml" {
			names = append(names, strings.TrimSuffix(entry.Name(), ".yaml"))
		}
	}
	sort.Strings(names)
	require.Len(t, names, distInfLegacyCaseCount, "all existing typed DistInf cases must be carried into the new composer")

	for _, name := range names {
		name := name
		t.Run(name, func(t *testing.T) {
			input, readErr := os.ReadFile(filepath.Join(distInfGoldenInputDir, name+".yaml"))
			require.NoError(t, readErr)

			spec := workloadsApi.NewDistributedInferenceSpecSpec()
			unmarshalErr := yaml.Unmarshal(input, spec)
			if unmarshalErr != nil {
				expectedError, found := distInfExpectedErrors[name]
				require.True(t, found, "unexpected typed input validation error: %v", unmarshalErr)
				require.Contains(t, unmarshalErr.Error(), expectedError)
				return
			}

			dinfComposer := NewDistributedInferenceToLWSComposer(
				"test-lws",
				"team-a",
				"runai-team-a",
				"tester",
				nil,
				nil,
				composer.FieldsComposerOptions{
					EmptyDirAsVolumeSupported:        true,
					FlexiblePriorityClassesSupported: true,
					NodeAffinitySupported:            true,
					PreemptibilitySupported:          true,
					UserCredentialsSupported:         true,
					VolumeDefaultModeSupported:       true,
					WorkloadV2NetworkingSupported:    true,
				},
				legacyTypedTestNodePools(),
			).WithWorkloadID("00000000-0000-0000-0000-000000001234").WithClusterStorageConfig(legacyTypedTestClusterConfig())
			target := NewMinimalLeaderWorkerSetTarget()
			composeErr := dinfComposer.Compose(context.Background(), spec, target)()

			if expectedError, found := distInfExpectedErrors[name]; found {
				require.NotNil(t, composeErr)
				require.Contains(t, composeErr.Message, expectedError)
				return
			}

			require.Nil(t, composeErr)
			normalizeGeneratedPVCNames(target)
			actual, marshalErr := yaml.Marshal(target)
			require.NoError(t, marshalErr)
			compareDistInfGolden(t, name, actual)
		})
	}
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

func compareDistInfGolden(t *testing.T, name string, actual []byte) {
	t.Helper()
	expectedPath := filepath.Join(distInfGoldenExpectedDir, name+".yaml")
	if os.Getenv("UPDATE_GOLDENS") == "1" {
		require.NoError(t, os.MkdirAll(distInfGoldenExpectedDir, 0755))
		require.NoError(t, os.WriteFile(expectedPath, actual, 0644))
		return
	}
	expected, err := os.ReadFile(expectedPath)
	if os.IsNotExist(err) {
		require.NoError(t, os.MkdirAll(distInfGoldenExpectedDir, 0755))
		require.NoError(t, os.WriteFile(expectedPath, actual, 0644))
		t.Fatalf("expected file %s was missing; generated it for review", expectedPath)
	}
	require.NoError(t, err)
	require.Equal(t, string(expected), string(actual), fmt.Sprintf("golden mismatch for %s", name))
}
