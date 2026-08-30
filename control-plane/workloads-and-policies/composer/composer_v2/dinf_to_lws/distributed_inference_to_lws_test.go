package dinf_to_lws

import (
	"context"
	"encoding/json"
	"testing"

	clusterApi "github.com/run-ai/runai/control-plane/cluster-service/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/composer"
	workloadsApi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
	workloadsManagerApi "github.com/run-ai/runai/runai-clients/workloads-manager-client/generated"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	lwsv1 "sigs.k8s.io/lws/api/leaderworkerset/v1"
)

func distributedInferenceSpecFromJSON(t *testing.T, raw string) *workloadsApi.DistributedInferenceSpecSpec {
	t.Helper()
	spec := workloadsApi.NewDistributedInferenceSpecSpec()
	require.NoError(t, json.Unmarshal([]byte(raw), spec))
	setDefaultServingPort(spec)
	return spec
}

func setDefaultServingPort(spec *workloadsApi.DistributedInferenceSpecSpec) {
	if spec.ServingPort.Get() != nil {
		return
	}
	servingPort := workloadsApi.NewDistributedInferenceServingPort()
	servingPort.SetPort(8080)
	servingPort.SetExposeExternally(false)
	spec.SetServingPort(*servingPort)
}

func environmentValue(env []corev1.EnvVar, name string) string {
	for _, variable := range env {
		if variable.Name == name {
			return variable.Value
		}
	}
	return ""
}

func TestComposeTopologyAndIndependentRoleValues(t *testing.T) {
	spec := workloadsApi.NewDistributedInferenceSpecSpec()
	spec.SetWorkers(2)
	spec.SetReplicas(3)
	spec.SetStartupPolicy(workloadsApi.DISTRIBUTEDINFERENCESTARTUPPOLICY_LEADER_READY)
	spec.SetRestartPolicy(workloadsApi.DISTRIBUTEDINFERENCERESTARTPOLICY_NONE)

	leader := workloadsApi.NewDistributedInferenceLeaderWorkerSpec()
	leader.SetImage("leader:latest")
	leader.SetCommand("run-leader")
	spec.SetLeader(*leader)

	worker := workloadsApi.NewDistributedInferenceLeaderWorkerSpec()
	worker.SetImage("worker:latest")
	worker.SetCommand("run-worker")
	spec.SetWorker(*worker)
	setDefaultServingPort(spec)

	dinfComposer := NewDistributedInferenceToLWSComposer(
		"dist-inf", "project-a", "namespace-a", "user-a",
		nil, nil, composer.FieldsComposerOptions{
			FlexiblePriorityClassesSupported: true,
			PreemptibilitySupported:          true,
		}, nil,
	)
	target := NewMinimalLeaderWorkerSetTarget()

	err := dinfComposer.Compose(context.Background(), spec, target)()
	require.Nil(t, err)

	require.Equal(t, int32(3), *target.LeaderWorkerSet.Spec.LeaderWorkerTemplate.Size)
	require.Equal(t, int32(3), *target.LeaderWorkerSet.Spec.Replicas)
	require.Equal(t, "LeaderReady", string(target.LeaderWorkerSet.Spec.StartupPolicy))
	require.Equal(t, "None", string(target.LeaderWorkerSet.Spec.LeaderWorkerTemplate.RestartPolicy))

	leaderContainer := target.LeaderWorkerSet.Spec.LeaderWorkerTemplate.LeaderTemplate.Spec.Containers[0]
	workerContainer := target.LeaderWorkerSet.Spec.LeaderWorkerTemplate.WorkerTemplate.Spec.Containers[0]
	require.Equal(t, "leader:latest", leaderContainer.Image)
	require.Equal(t, []string{"run-leader"}, leaderContainer.Command)
	require.Equal(t, "worker:latest", workerContainer.Image)
	require.Equal(t, []string{"run-worker"}, workerContainer.Command)
}

func TestComposeExpandsRunaiTemplateVarsInCommandAndArgs(t *testing.T) {
	spec := distributedInferenceSpecFromJSON(t, `{
		"leader": {
			"image": "leader:latest",
			"command": "serve.sh",
			"args": "--project=${RUNAI_PROJECT} --job=${RUNAI_JOB_NAME} --uid=${RUNAI_UID} --gid=${RUNAI_GID}",
			"security": {"uidGidSource": "custom", "runAsUid": 501, "runAsGid": 20}
		}
	}`)

	dinfComposer := NewDistributedInferenceToLWSComposer(
		"dist-inf", "project-a", "namespace-a", "user-a",
		nil, nil, composer.FieldsComposerOptions{}, nil,
	)
	target := NewMinimalLeaderWorkerSetTarget()
	require.Nil(t, dinfComposer.Compose(context.Background(), spec, target)())

	leaderContainer := target.LeaderWorkerSet.Spec.LeaderWorkerTemplate.LeaderTemplate.Spec.Containers[0]
	require.Equal(t, []string{"serve.sh"}, leaderContainer.Command)
	require.Equal(t, []string{
		"--project=project-a",
		"--job=dist-inf",
		"--uid=501",
		"--gid=20",
	}, leaderContainer.Args)
}

func TestComposeCopiesWholeLeaderWhenWorkerIsAbsent(t *testing.T) {
	spec := workloadsApi.NewDistributedInferenceSpecSpec()
	spec.SetWorkers(1)
	leader := workloadsApi.NewDistributedInferenceLeaderWorkerSpec()
	leader.SetImage("shared:latest")
	leader.SetArgs("--model model-a")
	spec.SetLeader(*leader)
	setDefaultServingPort(spec)

	dinfComposer := NewDistributedInferenceToLWSComposer(
		"dist-inf", "project-a", "namespace-a", "user-a",
		nil, nil, composer.FieldsComposerOptions{}, nil,
	)
	target := NewMinimalLeaderWorkerSetTarget()

	err := dinfComposer.Compose(context.Background(), spec, target)()
	require.Nil(t, err)

	leaderContainer := target.LeaderWorkerSet.Spec.LeaderWorkerTemplate.LeaderTemplate.Spec.Containers[0]
	workerContainer := target.LeaderWorkerSet.Spec.LeaderWorkerTemplate.WorkerTemplate.Spec.Containers[0]
	require.Equal(t, leaderContainer.Image, workerContainer.Image)
	require.Equal(t, leaderContainer.Args, workerContainer.Args)
	require.Equal(t, "worker", workerContainer.Name)
}

func TestComposeDoesNotInheritIndividualLeaderFieldsWhenWorkerExists(t *testing.T) {
	spec := workloadsApi.NewDistributedInferenceSpecSpec()
	spec.SetWorkers(1)
	leader := workloadsApi.NewDistributedInferenceLeaderWorkerSpec()
	leader.SetImage("leader:latest")
	leader.SetArgs("--leader-only")
	spec.SetLeader(*leader)

	worker := workloadsApi.NewDistributedInferenceLeaderWorkerSpec()
	worker.SetImage("worker:latest")
	spec.SetWorker(*worker)
	setDefaultServingPort(spec)

	dinfComposer := NewDistributedInferenceToLWSComposer(
		"dist-inf", "project-a", "namespace-a", "user-a",
		nil, nil, composer.FieldsComposerOptions{}, nil,
	)
	target := NewMinimalLeaderWorkerSetTarget()

	err := dinfComposer.Compose(context.Background(), spec, target)()
	require.Nil(t, err)

	workerContainer := target.LeaderWorkerSet.Spec.LeaderWorkerTemplate.WorkerTemplate.Spec.Containers[0]
	require.Equal(t, "worker:latest", workerContainer.Image)
	require.Empty(t, workerContainer.Args)
}

func TestComposeRejectsConfiguredWorkerWhenWorkersIsZero(t *testing.T) {
	spec := workloadsApi.NewDistributedInferenceSpecSpec()
	leader := workloadsApi.NewDistributedInferenceLeaderWorkerSpec()
	leader.SetImage("leader:latest")
	spec.SetLeader(*leader)
	worker := workloadsApi.NewDistributedInferenceLeaderWorkerSpec()
	worker.SetImage("worker:latest")
	spec.SetWorker(*worker)
	setDefaultServingPort(spec)
	dinfComposer := NewDistributedInferenceToLWSComposer("dist-inf", "project-a", "namespace-a", "user-a", nil, nil, composer.FieldsComposerOptions{}, nil)

	err := dinfComposer.Compose(context.Background(), spec, NewMinimalLeaderWorkerSetTarget())()

	require.NotNil(t, err)
	require.Equal(t, int32(400), err.Code)
	require.Contains(t, err.Message, "worker spec requires workers")
}

func TestComposeAllowsEmptyWorkerWhenWorkersIsZero(t *testing.T) {
	spec := workloadsApi.NewDistributedInferenceSpecSpec()
	leader := workloadsApi.NewDistributedInferenceLeaderWorkerSpec()
	leader.SetImage("leader:latest")
	spec.SetLeader(*leader)
	spec.SetWorker(*workloadsApi.NewDistributedInferenceLeaderWorkerSpec())
	setDefaultServingPort(spec)
	dinfComposer := NewDistributedInferenceToLWSComposer("dist-inf", "project-a", "namespace-a", "user-a", nil, nil, composer.FieldsComposerOptions{}, nil)
	target := NewMinimalLeaderWorkerSetTarget()

	require.Nil(t, dinfComposer.Compose(context.Background(), spec, target)())
	require.Equal(t, "leader:latest", target.LeaderWorkerSet.Spec.LeaderWorkerTemplate.WorkerTemplate.Spec.Containers[0].Image)
}

func TestComposeRequiresLeader(t *testing.T) {
	dinfComposer := NewDistributedInferenceToLWSComposer(
		"dist-inf", "project-a", "namespace-a", "user-a",
		nil, nil, composer.FieldsComposerOptions{}, nil,
	)

	err := dinfComposer.Compose(context.Background(), workloadsApi.NewDistributedInferenceSpecSpec(), NewMinimalLeaderWorkerSetTarget())()

	require.NotNil(t, err)
	require.Equal(t, int32(400), err.Code)
	require.Contains(t, err.Message, "leader spec is required")
}

func TestComposeRequiresServingPort(t *testing.T) {
	spec := workloadsApi.NewDistributedInferenceSpecSpec()
	leader := workloadsApi.NewDistributedInferenceLeaderWorkerSpec()
	leader.SetImage("leader:latest")
	spec.SetLeader(*leader)
	dinfComposer := NewDistributedInferenceToLWSComposer("dist-inf", "project-a", "namespace-a", "user-a", nil, nil, composer.FieldsComposerOptions{}, nil)

	err := dinfComposer.Compose(context.Background(), spec, NewMinimalLeaderWorkerSetTarget())()

	require.NotNil(t, err)
	require.Equal(t, int32(400), err.Code)
	require.Contains(t, err.Message, "servingPort is required")
}

func TestComposeRejectsInvalidLWSPolicies(t *testing.T) {
	leader := workloadsApi.NewDistributedInferenceLeaderWorkerSpec()
	leader.SetImage("leader:latest")

	t.Run("startup policy", func(t *testing.T) {
		spec := workloadsApi.NewDistributedInferenceSpecSpec()
		spec.SetLeader(*leader)
		setDefaultServingPort(spec)
		spec.SetStartupPolicy(workloadsApi.DistributedInferenceStartupPolicy("invalid"))
		dinfComposer := NewDistributedInferenceToLWSComposer("dist-inf", "project-a", "namespace-a", "user-a", nil, nil, composer.FieldsComposerOptions{}, nil)

		err := dinfComposer.Compose(context.Background(), spec, NewMinimalLeaderWorkerSetTarget())()
		require.NotNil(t, err)
		require.Equal(t, int32(400), err.Code)
		require.Contains(t, err.Message, "invalid startup policy")
	})

	t.Run("restart policy", func(t *testing.T) {
		spec := workloadsApi.NewDistributedInferenceSpecSpec()
		spec.SetLeader(*leader)
		setDefaultServingPort(spec)
		spec.SetRestartPolicy(workloadsApi.DistributedInferenceRestartPolicy("invalid"))
		dinfComposer := NewDistributedInferenceToLWSComposer("dist-inf", "project-a", "namespace-a", "user-a", nil, nil, composer.FieldsComposerOptions{}, nil)

		err := dinfComposer.Compose(context.Background(), spec, NewMinimalLeaderWorkerSetTarget())()
		require.NotNil(t, err)
		require.Equal(t, int32(400), err.Code)
		require.Contains(t, err.Message, "invalid restart policy")
	})
}

func TestComposeUsesUniqueLWSNameFromWorkloadID(t *testing.T) {
	spec := workloadsApi.NewDistributedInferenceSpecSpec()
	leader := workloadsApi.NewDistributedInferenceLeaderWorkerSpec()
	leader.SetImage("leader:latest")
	spec.SetLeader(*leader)
	setDefaultServingPort(spec)

	dinfComposer := NewDistributedInferenceToLWSComposer(
		"dist-inf", "project-a", "namespace-a", "user-a",
		nil, nil, composer.FieldsComposerOptions{}, nil,
	).WithWorkloadID("workload-1234abcd")
	target := NewMinimalLeaderWorkerSetTarget()

	err := dinfComposer.Compose(context.Background(), spec, target)()
	require.Nil(t, err)
	require.Equal(t, "dist-inf-abcd", target.LeaderWorkerSet.Name)
}

func TestComposeReusesExplicitLWSNameOnUpdate(t *testing.T) {
	spec := workloadsApi.NewDistributedInferenceSpecSpec()
	leader := workloadsApi.NewDistributedInferenceLeaderWorkerSpec()
	leader.SetImage("leader:latest")
	spec.SetLeader(*leader)
	setDefaultServingPort(spec)

	dinfComposer := NewDistributedInferenceToLWSComposer(
		"dist-inf", "project-a", "namespace-a", "user-a",
		nil, nil, composer.FieldsComposerOptions{}, nil,
	).WithLeaderWorkerSetName("dist-inf-original")
	target := NewMinimalLeaderWorkerSetTarget()

	require.Nil(t, dinfComposer.Compose(context.Background(), spec, target)())
	require.Equal(t, "dist-inf-original", target.LeaderWorkerSet.Name)
}

func TestComposeSharedPlacementAndEntityFields(t *testing.T) {
	spec := distributedInferenceSpecFromJSON(t, `{
		"workers": 1,
		"nodePools": ["pool-a"],
		"preemptibility": "preemptible",
		"priorityClass": "high",
		"leader": {
			"image": "leader:latest",
			"environmentVariables": [{"name": "ROLE", "value": "leader"}],
			"compute": {"cpuCoreRequest": 1, "cpuMemoryRequest": "1Gi"},
			"security": {"hostNetwork": true, "runAsNonRoot": true},
			"nodeType": "gpu",
			"labels": [{"name": "role-label", "value": "leader"}],
			"annotations": [{"name": "example.com/role", "value": "leader"}]
		},
		"worker": {
			"image": "worker:latest",
			"environmentVariables": [{"name": "ROLE", "value": "worker"}],
			"compute": {"cpuCoreRequest": 2, "cpuMemoryRequest": "2Gi"},
			"nodeType": "cpu"
		}
	}`)

	dinfComposer := NewDistributedInferenceToLWSComposer(
		"dist-inf", "project-a", "namespace-a", "user-a",
		nil, nil, composer.FieldsComposerOptions{
			FlexiblePriorityClassesSupported: true,
			PreemptibilitySupported:          true,
		}, []clusterApi.Nodepool{{Name: "pool-a", LabelKey: "run.ai/node-pool", LabelValue: "pool-a"}},
	)
	target := NewMinimalLeaderWorkerSetTarget()
	require.Nil(t, dinfComposer.Compose(context.Background(), spec, target)())

	leaderTemplate := target.LeaderWorkerSet.Spec.LeaderWorkerTemplate.LeaderTemplate
	workerTemplate := &target.LeaderWorkerSet.Spec.LeaderWorkerTemplate.WorkerTemplate
	require.Equal(t, "1", leaderTemplate.Spec.Containers[0].Resources.Requests.Cpu().String())
	require.Equal(t, "2", workerTemplate.Spec.Containers[0].Resources.Requests.Cpu().String())
	require.Equal(t, "1Gi", leaderTemplate.Spec.Containers[0].Resources.Requests.Memory().String())
	require.Equal(t, "2Gi", workerTemplate.Spec.Containers[0].Resources.Requests.Memory().String())
	require.Equal(t, "leader", environmentValue(leaderTemplate.Spec.Containers[0].Env, "ROLE"))
	require.Equal(t, "worker", environmentValue(workerTemplate.Spec.Containers[0].Env, "ROLE"))
	require.True(t, leaderTemplate.Spec.HostNetwork)
	require.Equal(t, "ClusterFirstWithHostNet", string(leaderTemplate.Spec.DNSPolicy))
	require.Equal(t, "gpu", leaderTemplate.Spec.NodeSelector["run.ai/node-type"])
	require.Equal(t, "cpu", workerTemplate.Spec.NodeSelector["run.ai/node-type"])
	require.Equal(t, "leader", leaderTemplate.Labels["role-label"])
	require.Equal(t, "dist-inf", leaderTemplate.Labels["release"])
	require.Equal(t, "project-a", workerTemplate.Labels["project"])
	require.Equal(t, "leader", leaderTemplate.Annotations["example.com/role"])
	require.Equal(t, "preemptible", leaderTemplate.Labels["kai.scheduler/preemptibility"])
	require.NotNil(t, leaderTemplate.Spec.Affinity)
	require.NotNil(t, workerTemplate.Spec.Affinity)
}

func TestComposeGeneratedReplicaLabelsOverrideUserLabels(t *testing.T) {
	spec := distributedInferenceSpecFromJSON(t, `{
		"leader": {
			"image": "leader:latest",
			"labels": [
				{"name": "app", "value": "user-app"},
				{"name": "release", "value": "user-release"},
				{"name": "workloadName", "value": "user-workload"}
			]
		}
	}`)

	dinfComposer := NewDistributedInferenceToLWSComposer(
		"dist-inf", "project-a", "namespace-a", "user-a",
		nil, nil, composer.FieldsComposerOptions{}, nil,
	)
	target := NewMinimalLeaderWorkerSetTarget()

	require.Nil(t, dinfComposer.Compose(context.Background(), spec, target)())
	labels := target.LeaderWorkerSet.Spec.LeaderWorkerTemplate.LeaderTemplate.Labels
	require.Equal(t, "leaderworkerset", labels["app"])
	require.Equal(t, "dist-inf", labels["release"])
	require.Equal(t, "dist-inf", labels["workloadName"])
}

func TestComposeServingPortAndMNNVLConfiguration(t *testing.T) {
	spec := distributedInferenceSpecFromJSON(t, `{
		"workers": 1,
		"servingPort": {
			"port": 8080,
			"authorizationType": "authorizedUsers",
			"authorizedUsers": ["user-a"],
			"exposeExternally": false
		},
		"leader": {
			"image": "leader:latest",
			"labels": [{"name": "run.ai/MNNVL", "value": "Required"}]
		},
		"worker": {"image": "worker:latest"}
	}`)

	dinfComposer := NewDistributedInferenceToLWSComposer(
		"dist-inf", "project-a", "namespace-a", "user-a",
		nil, nil, composer.FieldsComposerOptions{}, nil,
	).WithWorkloadID("workload-abcd").WithClusterStorageConfig(distributedInferenceAuthTestConfig())
	target := NewMinimalLeaderWorkerSetTarget()
	require.Nil(t, dinfComposer.Compose(context.Background(), spec, target)())

	require.Equal(t, workloadsManagerApi.MNNVLMODE_REQUIRED, target.Config.GetMnnvl())
	require.Len(t, target.Config.Networking, 1)
	endpoint := target.Config.Networking[0]
	require.False(t, endpoint.HasExternalAccess())
	require.Len(t, endpoint.Services, 1)
	service := endpoint.Services[0]
	require.True(t, service.HasSpec())
	require.Equal(t, "dinfw-dist-inf-sp-0", service.GetSpec().Name)
	require.Equal(t, "dist-inf-abcd", service.GetSpec().Selector[lwsv1.SetNameLabelKey])
	require.Equal(t, "0", service.GetSpec().Selector[lwsv1.WorkerIndexLabelKey])
	require.Equal(t, composer.Oauth2ContainerDefaultPort, service.GetSpec().Ports[0].GetTargetPort())
	require.False(t, service.HasAuthorization())

	leaderPodSpec := target.LeaderWorkerSet.Spec.LeaderWorkerTemplate.LeaderTemplate.Spec
	require.Len(t, leaderPodSpec.Containers, 2)
	require.Len(t, target.LeaderWorkerSet.Spec.LeaderWorkerTemplate.WorkerTemplate.Spec.Containers, 1)
	proxyContainer := leaderPodSpec.Containers[1]
	require.Equal(t, "runai-oauth2-proxy-0", proxyContainer.Name)
	require.Nil(t, proxyContainer.RestartPolicy)
	require.Equal(t, []corev1.ContainerPort{{ContainerPort: composer.Oauth2ContainerDefaultPort}}, proxyContainer.Ports)
	require.Contains(t, proxyContainer.Args, "--upstream=http://localhost:8080")
	require.Contains(t, proxyContainer.Args, "--redirect-url=http://dinfw-dist-inf-sp-0.namespace-a.svc.cluster.local/oauth2/callback")
	require.Contains(t, proxyContainer.Args, "--cookie-name=dinfw-dist-inf-sp-0")
	require.Contains(t, proxyContainer.Args, "--force-json-errors=true")
	require.Contains(t, proxyContainer.Args, "--authenticated-emails-file=/etc/runai/auth/users/c2VydmluZy1wb3J0")
	require.Len(t, leaderPodSpec.InitContainers, 1)
	require.Equal(t, "runai-auth-user-loader", leaderPodSpec.InitContainers[0].Name)
	require.Contains(t, leaderPodSpec.InitContainers[0].Command[2], "dXNlci1h")
	require.Contains(t, leaderPodSpec.Volumes, corev1.Volume{
		Name:         "runai-auth",
		VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}},
	})
}

func TestComposePublicExternalServingPort(t *testing.T) {
	spec := distributedInferenceSpecFromJSON(t, `{
		"servingPort": {
			"port": 8080,
			"authorizationType": "public",
			"exposeExternally": true,
			"exposedUrl": "https://model.example.com/predict"
		},
		"leader": {"image": "leader:latest"}
	}`)

	dinfComposer := NewDistributedInferenceToLWSComposer(
		"dist-inf", "project-a", "namespace-a", "user-a",
		nil, nil, composer.FieldsComposerOptions{}, nil,
	).WithClusterStorageConfig(composer.DefaultClusterStorageConfig())
	target := NewMinimalLeaderWorkerSetTarget()
	require.Nil(t, dinfComposer.Compose(context.Background(), spec, target)())

	require.Len(t, target.Config.Networking, 1)
	endpoint := target.Config.Networking[0]
	require.True(t, endpoint.HasExternalAccess())
	require.NotEmpty(t, endpoint.GetExternalAccess().Ingress)
	require.False(t, endpoint.Services[0].HasAuthorization())
	require.Equal(t, int32(8080), endpoint.Services[0].GetSpec().Ports[0].GetTargetPort())
	require.Len(t, target.LeaderWorkerSet.Spec.LeaderWorkerTemplate.LeaderTemplate.Spec.Containers, 1)
}

func TestComposeRejectsUnsupportedServingProtocol(t *testing.T) {
	spec := workloadsApi.NewDistributedInferenceSpecSpec()
	leader := workloadsApi.NewDistributedInferenceLeaderWorkerSpec()
	leader.SetImage("leader:latest")
	spec.SetLeader(*leader)
	servingPort := workloadsApi.NewDistributedInferenceServingPort()
	servingPort.SetPort(8080)
	servingPort.SetProtocol(workloadsApi.DistributedInferenceServingPortProtocol("grpc"))
	spec.SetServingPort(*servingPort)

	dinfComposer := NewDistributedInferenceToLWSComposer("dist-inf", "project-a", "namespace-a", "user-a", nil, nil, composer.FieldsComposerOptions{}, nil)
	err := dinfComposer.Compose(context.Background(), spec, NewMinimalLeaderWorkerSetTarget())()

	require.NotNil(t, err)
	require.Equal(t, int32(400), err.Code)
	require.Contains(t, err.Message, "unsupported protocol")
}

// TestComposeMNNVLLabelMapping asserts how the shared FieldsComposer.ComposeMNNVL
// pipeline maps the leader's run.ai/MNNVL label onto the emitted MNNVL plugin
// mode for distributed-inference workloads. The dinf composer defaults to None
// when the cluster's inference-workload-controller GPU network acceleration flag
// is off (the case exercised here), and the WorkloadRunner plugin treats
// Preferred the same as Required, so composition normalizes Preferred → Required.
func TestComposeMNNVLLabelMapping(t *testing.T) {
	for _, tc := range []struct {
		name     string
		value    string // empty means don't set the label on the leader
		expected workloadsManagerApi.MnnvlMode
	}{
		{name: "no label defaults to None", value: "", expected: workloadsManagerApi.MNNVLMODE_NONE},
		{name: "Preferred normalizes to Required", value: "Preferred", expected: workloadsManagerApi.MNNVLMODE_REQUIRED},
		{name: "explicit None is honored", value: "None", expected: workloadsManagerApi.MNNVLMODE_NONE},
	} {
		t.Run(tc.name, func(t *testing.T) {
			spec := workloadsApi.NewDistributedInferenceSpecSpec()
			leader := workloadsApi.NewDistributedInferenceLeaderWorkerSpec()
			leader.SetImage("leader:latest")
			if tc.value != "" {
				leader.SetLabels([]workloadsApi.Label{{
					Name:  *workloadsApi.NewNullableString(stringPointer("run.ai/MNNVL")),
					Value: *workloadsApi.NewNullableString(stringPointer(tc.value)),
				}})
			}
			spec.SetLeader(*leader)
			setDefaultServingPort(spec)
			dinfComposer := NewDistributedInferenceToLWSComposer("dist-inf", "project-a", "namespace-a", "user-a", nil, nil, composer.FieldsComposerOptions{}, nil)
			target := NewMinimalLeaderWorkerSetTarget()

			require.Nil(t, dinfComposer.Compose(context.Background(), spec, target)())
			require.Equal(t, tc.expected, target.Config.GetMnnvl())
		})
	}
}

func stringPointer(value string) *string {
	return &value
}

func distributedInferenceAuthTestConfig() composer.ClusterStorageConfig {
	reported := clusterApi.ClusterReportedConfig{}
	reported.SetRunaiOauth2ProxyImage("test.registry/oauth2-proxy:latest")
	reported.SetRunaiOauth2ProxyUsernameLoaderImage("test.registry/username-loader:latest")
	reported.SetRunaiOauth2ProxyWorkloadAuthSecretName("runai-auth-secret")
	reported.SetRunaiOauth2ProxySessionStoreType("redis")
	reported.SetRunaiOauth2ProxyRedisFQDN("redis.runai-backend.svc.cluster.local")
	reported.SetRunaiOauth2ProxyOidcIssuerUrl("https://test.run.ai/auth/realms/runai")

	status := clusterApi.ClusterDisplayedStatus{}
	status.SetConfig(reported)
	cluster := clusterApi.DisplayedCluster{}
	cluster.Status.Set(&status)
	cluster.SetDomain("https://test.run.ai")
	return composer.NewClusterStorageConfigFromCluster(&cluster)
}

func TestComposeExistingAndNewPVCs(t *testing.T) {
	spec := distributedInferenceSpecFromJSON(t, `{
		"workers": 1,
		"leader": {
			"image": "leader:latest",
			"storage": {"pvc": [{
				"name": "shared",
				"path": "/shared",
				"existingPvc": true,
				"claimName": "shared-claim"
			}]}
		},
		"worker": {
			"image": "worker:latest",
			"storage": {"pvc": [{
				"name": "models",
				"path": "/models",
				"existingPvc": false,
				"ephemeral": false,
				"claimName": "models-claim",
				"claimInfo": {"size": "1Gi"}
			}]}
		}
	}`)

	dinfComposer := NewDistributedInferenceToLWSComposer(
		"dist-inf", "project-a", "namespace-a", "user-a",
		nil, nil, composer.FieldsComposerOptions{}, nil,
	)
	target := NewMinimalLeaderWorkerSetTarget()
	require.Nil(t, dinfComposer.Compose(context.Background(), spec, target)())

	leaderVolumes := target.LeaderWorkerSet.Spec.LeaderWorkerTemplate.LeaderTemplate.Spec.Volumes
	workerVolumes := target.LeaderWorkerSet.Spec.LeaderWorkerTemplate.WorkerTemplate.Spec.Volumes
	require.Equal(t, "shared-claim", leaderVolumes[0].PersistentVolumeClaim.ClaimName)
	require.Equal(t, "models-claim", workerVolumes[0].PersistentVolumeClaim.ClaimName)
	require.Len(t, target.Config.Pvcs, 1)
	require.Equal(t, "models-claim", target.Config.Pvcs[0].Metadata.Name)
	require.Equal(t, "1Gi", target.Config.Pvcs[0].Spec.Resources.Requests.Storage)
}
