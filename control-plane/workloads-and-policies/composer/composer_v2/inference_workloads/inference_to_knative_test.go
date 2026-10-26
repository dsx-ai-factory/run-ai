/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package inference_workloads

import (
	"context"
	"fmt"
	"os"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer"
	"github.com/ghodss/yaml"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"

	clustersApi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/testutils"
	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-manager-client/generated"
	corev1 "k8s.io/api/core/v1"
	knapis "knative.dev/pkg/apis"
	knserving "knative.dev/serving/pkg/apis/serving"
)

func kindPtr(k workloadsApi.AssetKind) *workloadsApi.AssetKind { return &k }

type composerCase struct {
	name                       string
	specFile                   string
	expectedFile               string
	expectErr                  string
	assetsMapping              []workloadsApi.AssetsMapping
	dataVolumeClaimNames       map[string]string
	serviceSpecLabelsSupported *bool
	// gpuMemoryLimitAnnotationsSupported gates the >= 3.0.0 CUDA/KAI
	// gpu-memory annotations. When nil the runner defaults it to `true`
	// (annotation path). Set it to `convutils.Ref(false)` to exercise the
	// pre-3.0.0 legacy RUNAI_GPU_*_LIMIT env-var fallback against the same
	// spec fixture.
	gpuMemoryLimitAnnotationsSupported *bool
	clusterStorageConfig               composer.ClusterStorageConfig
	assert                             func(target *KnativeServiceTarget)
}

var _ = Describe("InferenceToKnativeComposer", func() {
	DescribeTable("translates SupersetSpec into KnativeServiceTarget",
		func(tc composerCase) {
			runCase(tc)
		},
		Entry("basic public", composerCase{
			name:         "basic",
			specFile:     "basic",
			expectedFile: "basic",
			assert: func(target *KnativeServiceTarget) {
				container := &target.Service.Spec.Template.Spec.Containers[0]
				Expect(container.Image).To(Equal("my-llm:latest"))
				Expect(container.Name).To(Equal("my-llm"))

				Expect(container.Ports).To(HaveLen(1))
				Expect(container.Ports[0].ContainerPort).To(Equal(int32(8080)))

				// Public access: Knative's own routing exposes the service, so no run:ai
				// networking plugin config is emitted (a service ref without auth would be
				// rejected by the anyworkload-controller admission webhook).
				Expect(target.Config.Networking).To(BeEmpty())
			},
		}),

		Entry("autoscaling annotations", composerCase{
			specFile:     "autoscaling",
			expectedFile: "autoscaling",
			assert: func(target *KnativeServiceTarget) {
				ann := target.Service.Spec.Template.ObjectMeta.Annotations
				Expect(ann["autoscaling.knative.dev/class"]).To(Equal("kpa.autoscaling.knative.dev"))
				Expect(ann["autoscaling.knative.dev/min-scale"]).To(Equal("0"))
				Expect(ann["autoscaling.knative.dev/max-scale"]).To(Equal("5"))
				Expect(ann["autoscaling.knative.dev/metric"]).To(Equal("rps"))
				Expect(ann["autoscaling.knative.dev/target"]).To(Equal("100"))
				Expect(ann["autoscaling.knative.dev/scale-to-zero-pod-retention-period"]).To(Equal("200s"))
				Expect(target.Service.Spec.Template.Spec.ContainerConcurrency).NotTo(BeNil())
				Expect(*target.Service.Spec.Template.Spec.ContainerConcurrency).To(Equal(int64(50)))
			},
		}),

		Entry("private serving URL", composerCase{
			specFile:     "private_url",
			expectedFile: "private_url",
			assert: func(target *KnativeServiceTarget) {
				Expect(target.Service.ObjectMeta.Labels["networking.knative.dev/visibility"]).To(Equal("cluster-local"))
			},
		}),

		Entry("auth authenticatedUsers uses an oauth2-proxy sidecar", composerCase{
			specFile:             "auth_authenticated",
			clusterStorageConfig: inferenceAuthTestConfig(),
			assert: func(target *KnativeServiceTarget) {
				assertInferenceAuthSidecar(target, "--email-domain=*")
				Expect(target.Config.Networking).To(BeEmpty())
			},
		}),

		Entry("serving auth still uses the proxy when external URL auth is disabled", composerCase{
			specFile:             "auth_authenticated",
			expectedFile:         "auth_authenticated_external_auth_disabled",
			clusterStorageConfig: inferenceAuthTestConfigWithExternalAuth(false),
		}),

		Entry("authenticated NIM keeps metrics as the only networking endpoint", composerCase{
			specFile:             "auth_nim",
			clusterStorageConfig: inferenceAuthTestConfig(),
			assert: func(target *KnativeServiceTarget) {
				assertInferenceAuthSidecar(target, "--email-domain=*")
				Expect(target.Config.Networking).To(HaveLen(1))
				metricsService := target.Config.Networking[0].Services[0]
				Expect(metricsService.Authorization.IsSet()).To(BeFalse())
				metricsSpec := metricsService.Spec.Get()
				Expect(metricsSpec).NotTo(BeNil())
				Expect(metricsSpec.GetLabels()).To(HaveKeyWithValue("run.ai/workload-metrics-service", "true"))
				Expect(metricsSpec.GetLabels()).To(HaveKeyWithValue("run.ai/nim-workload", "true"))
				Expect(metricsSpec.Ports[0].GetTargetPort()).To(Equal(int32(8080)))
			},
		}),

		Entry("auth authorizedUsers adds the username loader", composerCase{
			specFile:             "auth_users",
			clusterStorageConfig: inferenceAuthTestConfig(),
			assert: func(target *KnativeServiceTarget) {
				assertInferenceAuthSidecar(target,
					"--authenticated-emails-file=/etc/runai/auth/users/c2VydmluZy1wb3J0")
				Expect(target.Service.Spec.Template.Spec.InitContainers).To(HaveLen(1))
				Expect(target.Service.Spec.Template.Spec.InitContainers[0].Name).To(Equal("runai-auth-user-loader"))
			},
		}),

		Entry("auth authorizedGroups configures every allowed group", composerCase{
			specFile:             "auth_groups",
			clusterStorageConfig: inferenceAuthTestConfig(),
			assert: func(target *KnativeServiceTarget) {
				assertInferenceAuthSidecar(target, "--allowed-group=ml-eng", "--allowed-group=platform")
			},
		}),

		Entry("serving-port groups and exposed-URL users share OAuth volumes", composerCase{
			specFile:             "auth_groups_with_exposed_url_users",
			expectedFile:         "auth_groups_with_exposed_url_users",
			clusterStorageConfig: inferenceAuthTestConfig(),
			assert: func(target *KnativeServiceTarget) {
				assertCombinedInferenceAuthorization(target)
			},
		}),

		Entry("ambiguous auth (users AND groups)", composerCase{
			specFile:  "auth_ambiguous",
			expectErr: "authorizedUsers and authorizedGroups are mutually exclusive",
		}),

		Entry("populated allowlist without authorizationType is rejected", composerCase{
			specFile:  "auth_allowlist_no_type",
			expectErr: "authorizationType must be provided",
		}),

		Entry("exposedUrls configures a routed tool endpoint", composerCase{
			specFile:             "exposed_urls",
			expectedFile:         "exposed_urls",
			clusterStorageConfig: exposedUrlsTestConfig(),
			assert: func(target *KnativeServiceTarget) {
				Expect(target.Config.Networking).To(HaveLen(1))
				endpoint := target.Config.Networking[0]
				Expect(endpoint.Services).To(HaveLen(1))
				serviceSpec := endpoint.Services[0].GetSpec()
				Expect(serviceSpec.Name).To(Equal("infw-my-llm-url"))
				Expect(serviceSpec.Selector).To(HaveKeyWithValue("release", "my-llm"))
				Expect(serviceSpec.Ports).To(HaveLen(1))
				Expect(serviceSpec.Ports[0].GetTargetPort()).To(Equal(int32(4180)))
				Expect(endpoint.HasExternalAccess()).To(BeTrue())
				Expect(endpoint.GetExternalAccess().Ingress).NotTo(BeEmpty())

				container := target.Service.Spec.Template.Spec.Containers[0]
				Expect(container.Ports).To(ContainElement(corev1.ContainerPort{
					Name:          "port-7860",
					ContainerPort: 7860,
					Protocol:      corev1.ProtocolTCP,
				}))
			},
		}),

		Entry("data volumes are mounted from their resolved PVCs", composerCase{
			specFile:             "data_volumes",
			expectedFile:         "data_volumes",
			dataVolumeClaimNames: map[string]string{"dv-models": "data-volume-models-pvc"},
		}),

		Entry("data volumes require a resolved PVC claim", composerCase{
			specFile:             "data_volumes",
			expectErr:            "dataVolume: claim name not found for data volume ID dv-models",
			dataVolumeClaimNames: map[string]string{"different-id": "some-pvc"},
		}),

		Entry("NIM label propagates to pod template", composerCase{
			specFile:     "nim",
			expectedFile: "nim",
			assert: func(target *KnativeServiceTarget) {
				Expect(target.Service.Spec.Template.ObjectMeta.Labels["run.ai/nim-workload"]).To(Equal("true"))
			},
		}),

		Entry("image-inferred NIM label propagates to Service and pod template", composerCase{
			specFile: "nim_image",
			assert: func(target *KnativeServiceTarget) {
				Expect(target.Service.ObjectMeta.Labels).To(HaveKeyWithValue("run.ai/nim-workload", "true"))
				Expect(target.Service.Spec.Template.ObjectMeta.Labels).To(HaveKeyWithValue("run.ai/nim-workload", "true"))
			},
		}),

		Entry("image-inferred vLLM label propagates to Service and pod template", composerCase{
			specFile: "vllm_image",
			assert: func(target *KnativeServiceTarget) {
				Expect(target.Service.ObjectMeta.Labels).To(HaveKeyWithValue("run.ai/vllm-workload", "true"))
				Expect(target.Service.Spec.Template.ObjectMeta.Labels).To(HaveKeyWithValue("run.ai/vllm-workload", "true"))
			},
		}),

		Entry("metrics endpoint is gated by ServiceSpec labels capability", composerCase{
			specFile:                   "nim",
			serviceSpecLabelsSupported: convutils.Ref(false),
			assert: func(target *KnativeServiceTarget) {
				Expect(target.Config.Networking).To(BeEmpty())
				Expect(target.Service.Spec.Template.ObjectMeta.Labels["run.ai/nim-workload"]).To(Equal("true"))
			},
		}),

		Entry("user-credential env var → personalcreds dest secret", composerCase{
			specFile:     "env_user_credential",
			expectedFile: "env_user_credential",
			assetsMapping: []workloadsApi.AssetsMapping{
				{
					Name: convutils.Ref("my-ngc"),
					Id:   convutils.Ref("12345678-aaaa-bbbb-cccc-deadbeef0000"),
					Kind: kindPtr(workloadsApi.ASSETKIND_NGC_API_KEY),
				},
			},
			assert: func(target *KnativeServiceTarget) {
				container := &target.Service.Spec.Template.Spec.Containers[0]
				var plainEV, credEV *corev1.EnvVar
				for i := range container.Env {
					switch container.Env[i].Name {
					case "PLAIN":
						plainEV = &container.Env[i]
					case "NGC_API_KEY":
						credEV = &container.Env[i]
					}
				}
				Expect(plainEV).NotTo(BeNil())
				Expect(plainEV.Value).To(Equal("hello"))

				Expect(credEV).NotTo(BeNil())
				Expect(credEV.ValueFrom).NotTo(BeNil())
				Expect(credEV.ValueFrom.SecretKeyRef).NotTo(BeNil())
				Expect(credEV.ValueFrom.SecretKeyRef.Name).To(Equal("pc-my-llm-ngcgs-my-ngc-12345678-a"))
				Expect(credEV.ValueFrom.SecretKeyRef.Key).To(Equal("apiKey"))

				pc := target.Config.UserCredentials.Get()
				Expect(pc).NotTo(BeNil())
				Expect(pc.Sources).To(HaveLen(1))
				Expect(pc.Sources[0].Ref).To(Equal("ngcgs-my-ngc-12345678-a"))
			},
		}),

		Entry("user-credential image pull secret → personalcreds dest secret", composerCase{
			specFile:     "image_pull_credential",
			expectedFile: "image_pull_credential",
			assetsMapping: []workloadsApi.AssetsMapping{
				{
					Name: convutils.Ref("my-ngc-docker"),
					Id:   convutils.Ref("87654321-aaaa-bbbb-cccc-deadbeef0000"),
					Kind: kindPtr(workloadsApi.ASSETKIND_NGC_API_KEY),
				},
			},
			assert: func(target *KnativeServiceTarget) {
				ips := target.Service.Spec.Template.Spec.PodSpec.ImagePullSecrets
				Expect(ips).To(HaveLen(2))
				Expect(ips[0].Name).To(Equal("regular-secret"))
				Expect(ips[1].Name).To(Equal("pc-my-llm-ngcdr-my-ngc-docker-87654321-a"))

				pc := target.Config.UserCredentials.Get()
				Expect(pc).NotTo(BeNil())
				Expect(pc.Sources).To(HaveLen(1))
				Expect(pc.Sources[0].Ref).To(Equal("ngcdr-my-ngc-docker-87654321-a"))
			},
		}),

		Entry("serving configuration timeouts", composerCase{
			specFile:     "serving_config",
			expectedFile: "serving_config",
			assert: func(target *KnativeServiceTarget) {
				ann := target.Service.Spec.Template.ObjectMeta.Annotations
				Expect(ann["serving.knative.dev/progress-deadline"]).To(Equal("300s"))
				Expect(target.Service.Spec.Template.Spec.TimeoutSeconds).NotTo(BeNil())
				Expect(*target.Service.Spec.Template.Spec.TimeoutSeconds).To(Equal(int64(120)))
			},
		}),

		Entry("empty fields → minimal valid Service", composerCase{
			specFile:     "empty_fields",
			expectedFile: "empty_fields",
		}),

		Entry("user labels + annotations propagate to Service and pod template metadata", composerCase{
			specFile:     "labels_annotations",
			expectedFile: "labels_annotations",
			assert: func(target *KnativeServiceTarget) {
				lbl := target.Service.ObjectMeta.Labels
				Expect(lbl["team"]).To(Equal("ml-platform"))
				Expect(lbl["owner"]).To(Equal("alice"))
				ann := target.Service.ObjectMeta.Annotations
				Expect(ann["example.com/owner"]).To(Equal("alice@example.com"))
				Expect(ann["example.com/cost-center"]).To(Equal("research"))
				podAnn := target.Service.Spec.Template.ObjectMeta.Annotations
				Expect(podAnn["example.com/owner"]).To(Equal("alice@example.com"))
				Expect(podAnn["example.com/cost-center"]).To(Equal("research"))
			},
		}),

		Entry("workingDir is propagated to the container", composerCase{
			specFile:     "working_dir",
			expectedFile: "working_dir",
			assert: func(target *KnativeServiceTarget) {
				Expect(target.Service.Spec.Template.Spec.Containers[0].WorkingDir).To(Equal("/home/runai"))
			},
		}),

		Entry("RUNAI placeholders in command and args are expanded", composerCase{
			specFile: "command_args_placeholders",
			assert: func(target *KnativeServiceTarget) {
				container := &target.Service.Spec.Template.Spec.Containers[0]
				Expect(container.Command).To(Equal([]string{"serve.sh"}))
				Expect(container.Args).To(Equal([]string{
					"--project=team-a",
					"--job=my-llm",
					"--uid=501",
					"--gid=20",
				}))
			},
		}),

		Entry("host IPC and host network", composerCase{
			specFile:     "host_ipc_network",
			expectedFile: "host_ipc_network",
			assert: func(target *KnativeServiceTarget) {
				podSpec := &target.Service.Spec.Template.Spec.PodSpec
				Expect(podSpec.HostIPC).To(BeTrue())
				Expect(podSpec.HostNetwork).To(BeTrue())
			},
		}),

		Entry("security context (uid/gid, capabilities, seccomp)", composerCase{
			specFile:     "security_context",
			expectedFile: "security_context",
			assert: func(target *KnativeServiceTarget) {
				sc := target.Service.Spec.Template.Spec.Containers[0].SecurityContext
				Expect(sc).NotTo(BeNil())
				Expect(sc.RunAsNonRoot).NotTo(BeNil())
				Expect(*sc.RunAsNonRoot).To(BeTrue())
				Expect(sc.ReadOnlyRootFilesystem).NotTo(BeNil())
				Expect(*sc.ReadOnlyRootFilesystem).To(BeTrue())
			},
		}),

		Entry("fractional GPU (portion request)", composerCase{
			specFile:     "fractional",
			expectedFile: "fractional",
			assert: func(target *KnativeServiceTarget) {
				ann := target.Service.Spec.Template.ObjectMeta.Annotations
				Expect(ann["gpu-fraction"]).NotTo(BeEmpty())
				Expect(target.Service.Spec.Template.Spec.SchedulerName).To(Equal(composer.RunaiScheduler))
			},
		}),

		// GPU memory limit paths: >= 3.0.0 emits the CUDA annotation
		// nvidia.com/container.<workload-name>.gpu-memory.limit on the pod
		// template; < 3.0.0 falls back to the legacy RUNAI_GPU_MEMORY_LIMIT
		// env var on the user container. Both variants compose the same
		// spec fixture. The workload container name for a knative inference
		// service is the workload name ("my-llm") so the annotation key ends
		// with "container.my-llm.gpu-memory.limit".
		Entry("gpu memory request and limit", composerCase{
			specFile:     "gpu_memory_req_and_limit",
			expectedFile: "gpu_memory_req_and_limit",
			assert: func(target *KnativeServiceTarget) {
				ann := target.Service.Spec.Template.ObjectMeta.Annotations
				Expect(ann).To(HaveKeyWithValue("nvidia.com/container.my-llm.gpu-memory.limit", "11M"))
				for _, envVar := range target.Service.Spec.Template.Spec.Containers[0].Env {
					Expect(envVar.Name).NotTo(Equal("RUNAI_GPU_MEMORY_LIMIT"))
				}
			},
		}),

		Entry("gpu memory request and limit (legacy env, cluster < 3.0)", composerCase{
			specFile:                           "gpu_memory_req_and_limit",
			expectedFile:                       "gpu_memory_req_and_limit_legacy_env",
			gpuMemoryLimitAnnotationsSupported: convutils.Ref(false),
			assert: func(target *KnativeServiceTarget) {
				ann := target.Service.Spec.Template.ObjectMeta.Annotations
				Expect(ann).NotTo(HaveKey("nvidia.com/container.my-llm.gpu-memory.limit"))
				Expect(environmentValue(target.Service.Spec.Template.Spec.Containers[0].Env, "RUNAI_GPU_MEMORY_LIMIT")).To(Equal("11M"))
			},
		}),

		// GPU portion limit paths: >= 3.0.0 emits the KAI annotation
		// kai.scheduler/container.<workload-name>.gpu-memory.portion.limit
		// (the scheduler translates the portion to an absolute size before
		// forwarding to CUDA). < 3.0.0 falls back to RUNAI_GPU_PORTION_LIMIT.
		Entry("gpu portion request and limit", composerCase{
			specFile:     "gpu_portion_req_and_limit",
			expectedFile: "gpu_portion_req_and_limit",
			assert: func(target *KnativeServiceTarget) {
				ann := target.Service.Spec.Template.ObjectMeta.Annotations
				Expect(ann).To(HaveKeyWithValue("kai.scheduler/container.my-llm.gpu-memory.portion.limit", "0.5"))
				for _, envVar := range target.Service.Spec.Template.Spec.Containers[0].Env {
					Expect(envVar.Name).NotTo(Equal("RUNAI_GPU_PORTION_LIMIT"))
				}
			},
		}),

		Entry("gpu portion request and limit (legacy env, cluster < 3.0)", composerCase{
			specFile:                           "gpu_portion_req_and_limit",
			expectedFile:                       "gpu_portion_req_and_limit_legacy_env",
			gpuMemoryLimitAnnotationsSupported: convutils.Ref(false),
			assert: func(target *KnativeServiceTarget) {
				ann := target.Service.Spec.Template.ObjectMeta.Annotations
				Expect(ann).NotTo(HaveKey("kai.scheduler/container.my-llm.gpu-memory.portion.limit"))
				Expect(environmentValue(target.Service.Spec.Template.Spec.Containers[0].Env, "RUNAI_GPU_PORTION_LIMIT")).To(Equal("0.5"))
			},
		}),

		Entry("CPU + memory requests and limits", composerCase{
			specFile:     "cpu_memory",
			expectedFile: "cpu_memory",
			assert: func(target *KnativeServiceTarget) {
				res := target.Service.Spec.Template.Spec.Containers[0].Resources
				Expect(res.Requests).NotTo(BeEmpty())
				Expect(res.Limits).NotTo(BeEmpty())
			},
		}),

		Entry("plain env vars (no credentials)", composerCase{
			specFile:     "envvars_plain",
			expectedFile: "envvars_plain",
			assert: func(target *KnativeServiceTarget) {
				env := target.Service.Spec.Template.Spec.Containers[0].Env
				names := make(map[string]string)
				for _, e := range env {
					names[e.Name] = e.Value
				}
				Expect(names["FOO"]).To(Equal("bar"))
				Expect(names["LOG_LEVEL"]).To(Equal("info"))
			},
		}),

		Entry("plain image pull secrets (no credentials)", composerCase{
			specFile:     "image_pull_secrets_plain",
			expectedFile: "image_pull_secrets_plain",
			assert: func(target *KnativeServiceTarget) {
				ips := target.Service.Spec.Template.Spec.PodSpec.ImagePullSecrets
				Expect(ips).To(HaveLen(2))
				Expect(ips[0].Name).To(Equal("ghcr-secret"))
				Expect(ips[1].Name).To(Equal("docker-hub-secret"))
			},
		}),

		Entry("existing PVC mounted as a volume", composerCase{
			specFile:     "pvc_existing",
			expectedFile: "pvc_existing",
			assert: func(target *KnativeServiceTarget) {
				vols := target.Service.Spec.Template.Spec.PodSpec.Volumes
				Expect(vols).To(HaveLen(1))
				Expect(vols[0].PersistentVolumeClaim).NotTo(BeNil())
				Expect(vols[0].PersistentVolumeClaim.ClaimName).To(Equal("existing-models-pvc"))
				Expect(vols[0].PersistentVolumeClaim.ReadOnly).To(BeTrue())
				mounts := target.Service.Spec.Template.Spec.Containers[0].VolumeMounts
				Expect(mounts).To(HaveLen(1))
				Expect(mounts[0].MountPath).To(Equal("/models"))
			},
		}),

		Entry("non-existing PVC is mounted and configured for the storage plugin", composerCase{
			specFile: "pvc_non_existing",
			assert: func(target *KnativeServiceTarget) {
				Expect(target.Service.Spec.Template.Spec.Volumes).To(HaveLen(1))
				volume := target.Service.Spec.Template.Spec.Volumes[0]
				Expect(volume.PersistentVolumeClaim).NotTo(BeNil())
				claimName := volume.PersistentVolumeClaim.ClaimName
				Expect(claimName).To(HavePrefix("my-llm-project-"))

				mounts := target.Service.Spec.Template.Spec.Containers[0].VolumeMounts
				Expect(mounts).To(ConsistOf(corev1.VolumeMount{
					Name:      volume.Name,
					MountPath: "/models",
				}))

				Expect(target.Config.Pvcs).To(HaveLen(1))
				pvc := target.Config.Pvcs[0]
				Expect(pvc.Metadata.Name).To(Equal(claimName))
				Expect(pvc.Spec.AccessModes).To(ConsistOf(openapi.WORKLOADV2UPDATECONFIGURATIONPVCACCESSMODE_READ_WRITE_ONCE))
				Expect(pvc.Spec.Resources.Requests.Storage).To(Equal("10Gi"))
				Expect(pvc.Spec.GetStorageClassName()).To(Equal("standard"))
			},
		}),

		Entry("ephemeral PVC is rejected for inference workloads", composerCase{
			specFile:  "pvc_ephemeral",
			expectErr: "Ephemeral must not be set 'true' for inference workloads",
		}),

		Entry("storage.git wires git-sync init containers into the Knative PodSpec", composerCase{
			specFile:             "storage_git",
			expectedFile:         "storage_git",
			clusterStorageConfig: gitSyncTestConfig(),
			assert: func(target *KnativeServiceTarget) {
				podSpec := &target.Service.Spec.Template.Spec.PodSpec
				Expect(podSpec.InitContainers).To(HaveLen(1))
				Expect(podSpec.InitContainers[0].Name).To(Equal("git-sync-git-code"))
				Expect(podSpec.Volumes).To(HaveLen(1))
				Expect(podSpec.Volumes[0].Name).To(Equal("code-sync-0"))
				Expect(podSpec.Volumes[0].EmptyDir).NotTo(BeNil())
				container := &target.Service.Spec.Template.Spec.Containers[0]
				mounts := container.VolumeMounts
				Expect(mounts).To(HaveLen(1))
				Expect(mounts[0].Name).To(Equal("code-sync-0"))
				Expect(mounts[0].MountPath).To(Equal("/code"))
			},
		}),

		Entry("storage.nfs is mounted", composerCase{
			specFile: "storage_nfs",
			assert: func(target *KnativeServiceTarget) {
				podSpec := &target.Service.Spec.Template.Spec.PodSpec
				Expect(podSpec.Volumes).To(ConsistOf(corev1.Volume{
					Name: "nfs-shared",
					VolumeSource: corev1.VolumeSource{
						NFS: &corev1.NFSVolumeSource{
							Server: "nfs.example.com",
							Path:   "/shared",
						},
					},
				}))
				Expect(podSpec.Containers[0].VolumeMounts).To(ConsistOf(corev1.VolumeMount{
					Name:      "nfs-shared",
					MountPath: "/mnt/shared",
				}))
			},
		}),

		Entry("Kubernetes-native storage volumes are mounted", composerCase{
			specFile: "storage_k8s_volumes",
			assert: func(target *KnativeServiceTarget) {
				podSpec := &target.Service.Spec.Template.Spec.PodSpec
				Expect(podSpec.Volumes).To(HaveLen(4))
				Expect(podSpec.Volumes).To(ContainElements(
					And(HaveField("Name", "host-path-scratch"), HaveField("VolumeSource.HostPath.Path", "/tmp/scratch")),
					And(HaveField("Name", "config-map-app-config"), HaveField("VolumeSource.ConfigMap.LocalObjectReference.Name", "my-config")),
					And(HaveField("Name", "secret-credentials"), HaveField("VolumeSource.Secret.SecretName", "my-secret")),
					And(HaveField("Name", "empty-dir-cache"), HaveField("VolumeSource.EmptyDir.Medium", corev1.StorageMediumMemory)),
				))
				Expect(podSpec.Containers[0].VolumeMounts).To(ConsistOf(
					corev1.VolumeMount{Name: "host-path-scratch", MountPath: "/scratch"},
					corev1.VolumeMount{Name: "config-map-app-config", MountPath: "/config"},
					corev1.VolumeMount{Name: "secret-credentials", MountPath: "/secrets"},
					corev1.VolumeMount{Name: "empty-dir-cache", MountPath: "/cache"},
				))

				var emptyDir *corev1.EmptyDirVolumeSource
				for _, volume := range podSpec.Volumes {
					if volume.Name == "empty-dir-cache" {
						emptyDir = volume.EmptyDir
					}
				}
				Expect(emptyDir).NotTo(BeNil())
				Expect(emptyDir.SizeLimit).NotTo(BeNil())
				Expect(emptyDir.SizeLimit.String()).To(Equal("1Gi"))
			},
		}),

		Entry("nodeType → NodeSelector entry", composerCase{
			specFile:     "node_type",
			expectedFile: "node_type",
			assert: func(target *KnativeServiceTarget) {
				Expect(target.Service.Spec.Template.Spec.NodeSelector).To(HaveKeyWithValue("run.ai/node-type", "gpu-a100"))
			},
		}),

		Entry("nodeAffinityRequired → typed NodeAffinity", composerCase{
			specFile:     "node_affinity",
			expectedFile: "node_affinity",
			assert: func(target *KnativeServiceTarget) {
				aff := target.Service.Spec.Template.Spec.Affinity
				Expect(aff).NotTo(BeNil())
				Expect(aff.NodeAffinity).NotTo(BeNil())
				Expect(aff.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution).NotTo(BeNil())
			},
		}),

		Entry("podAffinity → typed PodAffinity", composerCase{
			specFile:     "pod_affinity",
			expectedFile: "pod_affinity",
			assert: func(target *KnativeServiceTarget) {
				aff := target.Service.Spec.Template.Spec.Affinity
				Expect(aff).NotTo(BeNil())
				Expect(aff.PodAffinity).NotTo(BeNil())
			},
		}),

		Entry("tolerations propagate to pod spec", composerCase{
			specFile:     "tolerations",
			expectedFile: "tolerations",
			assert: func(target *KnativeServiceTarget) {
				tols := target.Service.Spec.Template.Spec.Tolerations
				Expect(tols).To(HaveLen(1))
				Expect(tols[0].Key).To(Equal("nvidia.com/gpu"))
			},
		}),

		Entry("readiness probe propagates to container", composerCase{
			specFile:     "probes",
			expectedFile: "probes",
			assert: func(target *KnativeServiceTarget) {
				p := target.Service.Spec.Template.Spec.Containers[0].ReadinessProbe
				Expect(p).NotTo(BeNil())
				Expect(p.HTTPGet).NotTo(BeNil())
				Expect(p.HTTPGet.Path).To(Equal("/healthz"))
			},
		}),

		Entry("preemptibility → pod template label", composerCase{
			specFile:     "preemptibility",
			expectedFile: "preemptibility",
			assert: func(target *KnativeServiceTarget) {
				lbl := target.Service.Spec.Template.ObjectMeta.Labels
				Expect(lbl).To(HaveKeyWithValue("kai.scheduler/preemptibility", "preemptible"))
			},
		}),

		Entry("semi-preemptibility → pod template label", composerCase{
			specFile:     "semi_preemptibility",
			expectedFile: "semi_preemptibility",
			assert: func(target *KnativeServiceTarget) {
				lbl := target.Service.Spec.Template.ObjectMeta.Labels
				Expect(lbl).To(HaveKeyWithValue("kai.scheduler/preemptibility", "semi-preemptible"))
			},
		}),
	)

	DescribeTable("maps API autoscaling metrics to Knative metrics",
		func(metric, expectedMetric, expectedClass string) {
			autoscaling := workloadsApi.NewAutoScaling()
			autoscaling.SetMetric(metric)
			target := NewMinimalKnativeServiceTarget()

			Expect(autoScalingToKnative(autoscaling, &target.Service, nil)).To(BeNil())
			Expect(target.Service.Spec.Template.Annotations[annotationAutoscalingMetric]).To(Equal(expectedMetric))
			Expect(target.Service.Spec.Template.Annotations[annotationAutoscalingClass]).To(Equal(expectedClass))
		},
		Entry("maps throughput to rps", "throughput", "rps", "kpa.autoscaling.knative.dev"),
		Entry("concurrency", "concurrency", "concurrency", "kpa.autoscaling.knative.dev"),
		Entry("latency", "latency", "knative_revision_request_latency", "hpa.autoscaling.knative.dev"),
		Entry("custom", "custom_metric", "custom_metric", "hpa.autoscaling.knative.dev"),
	)
})

func runCase(tc composerCase) {
	specYaml, err := testutils.LoadInputFile(tc.specFile, ".")
	Expect(err).To(BeNil(), fmt.Sprintf("load %s.yaml", tc.specFile))

	var src workloadsApi.InferenceSpecSpec
	Expect(yaml.Unmarshal(specYaml, &src)).To(Succeed())

	serviceSpecLabelsSupported := true
	if tc.serviceSpecLabelsSupported != nil {
		serviceSpecLabelsSupported = *tc.serviceSpecLabelsSupported
	}
	gpuMemoryLimitAnnotationsSupported := true
	if tc.gpuMemoryLimitAnnotationsSupported != nil {
		gpuMemoryLimitAnnotationsSupported = *tc.gpuMemoryLimitAnnotationsSupported
	}
	infComposer := NewInferenceToKnativeComposer(
		"my-llm",
		"team-a",
		"runai-team-a",
		"tester",
		tc.assetsMapping,
		composer.FieldsComposerOptions{
			UserCredentialsSupported:             true,
			NodeAffinitySupported:                true,
			PreemptibilitySupported:              true,
			EmptyDirAsVolumeSupported:            true,
			GpuMemoryLimitAnnotationsSupported:   gpuMemoryLimitAnnotationsSupported,
			WorkloadV2ServiceSpecLabelsSupported: serviceSpecLabelsSupported,
		},
	)
	if tc.clusterStorageConfig != nil {
		infComposer.WithClusterStorageConfig(tc.clusterStorageConfig)
	}
	infComposer.WithDataVolumeClaimNames(tc.dataVolumeClaimNames)
	target := NewMinimalKnativeServiceTarget()

	composerErr := infComposer.Compose(context.Background(), convutils.Ref(src), target)()

	if tc.expectErr != "" {
		Expect(composerErr).NotTo(BeNil(), "expected error: "+tc.expectErr)
		Expect(composerErr.Message).To(ContainSubstring(tc.expectErr))
		return
	}

	Expect(composerErr).To(BeNil(), "unexpected composer error")
	if tc.assert != nil {
		tc.assert(target)
	}
	if tc.expectedFile != "" {
		compareGolden(target, tc.expectedFile)
	}
}

func exposedUrlsTestConfig() composer.ClusterStorageConfig {
	config := composer.DefaultClusterStorageConfig().(composer.StaticClusterStorageConfig)
	config.SetClusterDomain("https://mock-q7z3k.example.com")
	config.SetOauth2ProxyImage("test.registry/oauth2-proxy:test")
	return config
}

func gitSyncTestConfig() composer.ClusterStorageConfig {
	config := composer.DefaultClusterStorageConfig().(composer.StaticClusterStorageConfig)
	config.SetGitSyncImage("registry.k8s.io/git-sync/git-sync:v4.7.0")
	return config
}

func inferenceAuthTestConfig() composer.ClusterStorageConfig {
	reported := clustersApi.ClusterReportedConfig{}
	reported.SetRunaiOauth2ProxyImage("test.registry/oauth2-proxy:latest")
	reported.SetRunaiOauth2ProxyUsernameLoaderImage("test.registry/username-loader:latest")
	reported.SetRunaiOauth2ProxyWorkloadAuthSecretName("runai-auth-secret")
	reported.SetRunaiOauth2ProxySessionStoreType("redis")
	reported.SetRunaiOauth2ProxyRedisFQDN("redis.runai-backend.svc.cluster.local")
	reported.SetRunaiOauth2ProxyOidcIssuerUrl("https://mock-q7z3k.example.com/auth/realms/runai")
	reported.SetSubdomainEnabled(true)

	status := clustersApi.ClusterDisplayedStatus{}
	status.SetConfig(reported)
	cluster := clustersApi.DisplayedCluster{}
	cluster.Status.Set(&status)
	cluster.SetDomain("https://mock-q7z3k.example.com")
	return composer.NewClusterStorageConfigFromCluster(&cluster)
}

func inferenceAuthTestConfigWithExternalAuth(enabled bool) composer.ClusterStorageConfig {
	config := inferenceAuthTestConfig().(composer.StaticClusterStorageConfig)
	config.SetExternalAuthUrlEnabled(enabled)
	return config
}

func assertInferenceAuthSidecar(target *KnativeServiceTarget, expectedArgs ...string) {
	GinkgoHelper()

	podSpec := &target.Service.Spec.Template.Spec.PodSpec
	Expect(podSpec.Containers).To(HaveLen(2))
	userContainer := podSpec.Containers[0]
	proxyContainer := podSpec.Containers[1]

	Expect(userContainer.Ports).To(BeEmpty())
	Expect(userContainer.ReadinessProbe).To(BeNil())
	Expect(proxyContainer.Name).To(Equal("runai-oauth2-proxy-0"))
	Expect(proxyContainer.Ports).To(Equal([]corev1.ContainerPort{{ContainerPort: 4180}}))
	Expect(proxyContainer.RestartPolicy).To(BeNil())
	Expect(proxyContainer.Args).To(ContainElements(
		"--upstream=http://localhost:8080",
		"--force-json-errors=true",
		"--cookie-name=InferenceWorkload-runai-team-a-my-llm-0",
	))
	Expect(proxyContainer.Args).To(ContainElements(expectedArgs))
	Expect(proxyContainer.ReadinessProbe).NotTo(BeNil())
	Expect(proxyContainer.ReadinessProbe.TCPSocket).NotTo(BeNil())
	Expect(proxyContainer.ReadinessProbe.TCPSocket.Port.IntVal).To(Equal(int32(8080)))
	Expect(podSpec.Volumes).To(ContainElement(corev1.Volume{
		Name:         "runai-auth",
		VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}},
	}))
}

func assertCombinedInferenceAuthorization(target *KnativeServiceTarget) {
	GinkgoHelper()

	podSpec := &target.Service.Spec.Template.Spec.PodSpec
	for _, container := range append(podSpec.Containers, podSpec.InitContainers...) {
		Expect(knapis.CheckDisallowedFields(container, *knserving.ContainerMask(&container))).To(BeNil())
	}

	Expect(podSpec.Volumes).To(ConsistOf(corev1.Volume{
		Name:         "runai-auth",
		VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}},
	}))

	allNames := make([]string, 0, len(podSpec.Containers)+len(podSpec.InitContainers))
	for _, container := range podSpec.Containers {
		allNames = append(allNames, container.Name)
	}
	for _, container := range podSpec.InitContainers {
		allNames = append(allNames, container.Name)
	}
	Expect(allNames).To(ConsistOf(
		"my-llm",
		"runai-oauth2-proxy-0",
		"runai-auth-user-loader",
		"runai-oauth2-proxy-1",
	))

	Expect(podSpec.Containers).To(HaveLen(3))
	Expect(podSpec.Containers[0].Ports).To(BeEmpty())
	servingProxy := &podSpec.Containers[1]
	Expect(servingProxy.Name).To(Equal("runai-oauth2-proxy-0"))
	Expect(servingProxy.Ports).To(ConsistOf(corev1.ContainerPort{ContainerPort: 4180}))

	var externalProxy corev1.Container
	Expect(podSpec.Containers).To(ContainElement(HaveField("Name", "runai-oauth2-proxy-1"), &externalProxy))
	Expect(externalProxy.RestartPolicy).To(BeNil())
	Expect(externalProxy.Ports).To(BeEmpty())
	Expect(externalProxy.Args).To(ContainElement("--http-address=0.0.0.0:4181"))
	var servingRedirect, externalRedirect string
	Expect(servingProxy.Args).To(ContainElement(HavePrefix("--redirect-url="), &servingRedirect))
	Expect(externalProxy.Args).To(ContainElement(HavePrefix("--redirect-url="), &externalRedirect))
	Expect(externalRedirect).NotTo(Equal(servingRedirect))
	Expect(externalProxy.Args).To(ContainElement(HavePrefix("--authenticated-emails-file=")))
	Expect(target.Config.Networking[0].Services[0].GetSpec().Ports[0].GetTargetPort()).To(Equal(int32(4181)))
	Expect(servingProxy.Args).To(ContainElement("--allowed-group=serving-port-group"))
}

// compareGolden self-bootstraps: missing files are written + the test fails so the dev reviews + commits.
// environmentValue returns the value of the first env var in `env` with
// the given name, or "" if the name is not present.
func environmentValue(env []corev1.EnvVar, name string) string {
	for _, variable := range env {
		if variable.Name == name {
			return variable.Value
		}
	}
	return ""
}

func compareGolden(target *KnativeServiceTarget, name string) {
	actualYaml, err := yaml.Marshal(target)
	Expect(err).To(BeNil(), "marshal target to YAML")

	expectedYaml := testutils.LoadExpectedFile(name, actualYaml, ".")

	ok := assert.Equal(GinkgoT(), string(expectedYaml), string(actualYaml), "golden file mismatch for "+name)
	if !ok {
		tmp := fmt.Sprintf("%s/%s.yaml", os.TempDir(), name)
		if werr := os.WriteFile(tmp, actualYaml, 0600); werr == nil {
			assert.Fail(GinkgoT(), fmt.Sprintf("actual result backed up to %s", tmp))
		}
	}
}
