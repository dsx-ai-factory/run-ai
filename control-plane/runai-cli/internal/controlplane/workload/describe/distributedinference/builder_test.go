/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package distributedinference_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/workload/describe/distributedinference"
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"k8s.io/utils/ptr"
)

var _ = Describe("DistributedInference Builder", func() {
	Context("BuildFromSpec", func() {
		It("should return nil when spec is nil", func() {
			result := distributedinference.BuildFromSpec(nil)
			Expect(result).To(BeNil())
		})

		It("should build fields from spec with all values set", func() {
			spec := createFullSpec()

			result := distributedinference.BuildFromSpec(spec)

			Expect(result).NotTo(BeNil())
			Expect(result.RestartPolicy).To(Equal(ptr.To("RecreateGroupOnPodRestart")))
			Expect(result.StartupPolicy).To(Equal(ptr.To("LeaderCreated")))
			Expect(result.Replicas).To(Equal(ptr.To(int32(3))))
			Expect(result.Workers).To(Equal(ptr.To(int32(2))))
			Expect(result.ServingPort).To(Equal(ptr.To(int32(8080))))
			Expect(result.AuthorizationType).To(Equal(ptr.To("public")))
			Expect(result.AuthorizedUsers).To(Equal([]string{"user1", "user2"}))
			Expect(result.AuthorizedGroups).To(Equal([]string{"group1"}))
			Expect(result.Protocol).To(Equal(ptr.To("http")))
			Expect(result.ExposeExternally).To(Equal(ptr.To(true)))

			// Leader configuration
			Expect(result.Leader).NotTo(BeNil())
			Expect(result.Leader.Image).To(Equal(ptr.To("leader-image:latest")))
			Expect(result.Leader.ImagePullPolicy).To(Equal(ptr.To("Always")))
			Expect(result.Leader.Command).To(Equal(ptr.To("python")))
			Expect(result.Leader.Arguments).To(Equal(ptr.To("--leader")))
			Expect(result.Leader.WorkingDir).To(Equal(ptr.To("/app")))
			Expect(result.Leader.NodeType).To(Equal(ptr.To("gpu-nodes")))
			Expect(result.Leader.EnvironmentVariables).To(HaveKeyWithValue("ROLE", "leader"))
			Expect(result.Leader.Labels).To(HaveKeyWithValue("app", "leader"))
			Expect(result.Leader.RunAsUser).To(Equal(ptr.To(int64(1000))))
			Expect(result.Leader.RunAsGroup).To(Equal(ptr.To(int64(1000))))
			Expect(result.Leader.Compute).NotTo(BeNil())
			Expect(result.Leader.Compute.RequestedResources.GPUCompute.Requested).To(Equal(float64(4)))
			Expect(result.Leader.Compute.GPUComputeMode).To(Equal(ptr.To("time-slicing")))
			Expect(result.Leader.Compute.RequestedResources.CPUCores.Requested).To(Equal(float64(8)))

			// Worker configuration
			Expect(result.Worker).NotTo(BeNil())
			Expect(result.Worker.Image).To(Equal(ptr.To("worker-image:latest")))
			Expect(result.Worker.ImagePullPolicy).To(Equal(ptr.To("IfNotPresent")))
			Expect(result.Worker.Command).To(Equal(ptr.To("python")))
			Expect(result.Worker.Arguments).To(Equal(ptr.To("--worker")))
			Expect(result.Worker.WorkingDir).To(Equal(ptr.To("/app")))
			Expect(result.Worker.NodeType).To(Equal(ptr.To("gpu-nodes")))
			Expect(result.Worker.EnvironmentVariables).To(HaveKeyWithValue("ROLE", "worker"))
			Expect(result.Worker.Labels).To(HaveKeyWithValue("app", "worker"))
			Expect(result.Worker.RunAsUser).To(Equal(ptr.To(int64(2000))))
			Expect(result.Worker.Compute).NotTo(BeNil())
			Expect(result.Worker.Compute.RequestedResources.GPUCompute.Requested).To(Equal(float64(2)))
			Expect(result.Worker.Compute.GPUComputeMode).To(Equal(ptr.To("sm-sharing")))

			// Leader storage
			Expect(result.Leader.Storage).NotTo(BeNil())
			Expect(result.Leader.Storage.PVCs).To(HaveLen(1))
			Expect(result.Leader.Storage.PVCs[0].Name).To(Equal("model-cache"))
			Expect(result.Leader.Storage.PVCs[0].MountPath).To(Equal("/models"))
			Expect(result.Leader.Storage.PVCs[0].ReadOnly).To(Equal(ptr.To(true)))
			Expect(result.Leader.Storage.PVCs[0].Ephemeral).To(Equal(ptr.To(false)))
			Expect(result.Leader.Storage.PVCs[0].DataSharing).To(Equal(ptr.To(true)))
			Expect(result.Leader.Storage.ConfigMaps).To(HaveLen(1))
			Expect(result.Leader.Storage.ConfigMaps[0].Name).To(Equal("inference-config"))
			Expect(result.Leader.Storage.EmptyDirs).To(HaveLen(1))
			Expect(result.Leader.Storage.EmptyDirs[0].Name).To(Equal("tmp-cache"))
			Expect(result.Leader.Storage.EmptyDirs[0].Size).To(Equal("20Gi"))
			Expect(result.Leader.Storage.EmptyDirs[0].Ephemeral).To(Equal(ptr.To(true)))
		})

		It("should return all nil fields when spec has no values set", func() {
			spec := &openapi.DistributedInferenceSpecSpec{}

			result := distributedinference.BuildFromSpec(spec)

			Expect(result).NotTo(BeNil())
			Expect(result.RestartPolicy).To(BeNil())
			Expect(result.StartupPolicy).To(BeNil())
			Expect(result.Replicas).To(BeNil())
			Expect(result.Workers).To(BeNil())
			Expect(result.ServingPort).To(BeNil())
			Expect(result.AuthorizationType).To(BeNil())
			Expect(result.AuthorizedUsers).To(BeNil())
			Expect(result.AuthorizedGroups).To(BeNil())
			Expect(result.Protocol).To(BeNil())
			Expect(result.ExposeExternally).To(BeNil())
			Expect(result.Leader).To(BeNil())
			Expect(result.Worker).To(BeNil())
		})
	})

	Context("Builder pattern", func() {
		It("should build fields using builder methods", func() {
			builder := distributedinference.NewBuilder()

			result := builder.
				RestartPolicy("RecreateGroupOnPodRestart").
				StartupPolicy("LeaderCreated").
				Replicas(ptr.To(int32(2))).
				Workers(ptr.To(int32(3))).
				ServingPort(ptr.To(int32(8080))).
				AuthorizationType("public").
				AuthorizedUsers([]string{"user1"}).
				AuthorizedGroups([]string{"group1"}).
				Protocol("http").
				ExposeExternally(ptr.To(true)).
				Build()

			Expect(result).NotTo(BeNil())
			Expect(result.RestartPolicy).To(Equal(ptr.To("RecreateGroupOnPodRestart")))
			Expect(result.StartupPolicy).To(Equal(ptr.To("LeaderCreated")))
			Expect(result.Replicas).To(Equal(ptr.To(int32(2))))
			Expect(result.Workers).To(Equal(ptr.To(int32(3))))
			Expect(result.ServingPort).To(Equal(ptr.To(int32(8080))))
			Expect(result.AuthorizationType).To(Equal(ptr.To("public")))
			Expect(result.AuthorizedUsers).To(Equal([]string{"user1"}))
			Expect(result.AuthorizedGroups).To(Equal([]string{"group1"}))
			Expect(result.Protocol).To(Equal(ptr.To("http")))
			Expect(result.ExposeExternally).To(Equal(ptr.To(true)))
		})

		It("should build empty fields when no methods called", func() {
			builder := distributedinference.NewBuilder()
			result := builder.Build()

			Expect(result).NotTo(BeNil())
			Expect(result.RestartPolicy).To(BeNil())
			Expect(result.StartupPolicy).To(BeNil())
			Expect(result.Replicas).To(BeNil())
			Expect(result.Workers).To(BeNil())
			Expect(result.ServingPort).To(BeNil())
		})
	})

	Context("SetFields", func() {
		It("should set fields on target pointer", func() {
			fields := &distributedinference.Fields{
				RestartPolicy: ptr.To("RecreateGroupOnPodRestart"),
				Replicas:      ptr.To(int32(3)),
			}

			var target *distributedinference.Fields
			distributedinference.SetFields(&target, fields)

			Expect(target).NotTo(BeNil())
			Expect(target.RestartPolicy).To(Equal(ptr.To("RecreateGroupOnPodRestart")))
			Expect(target.Replicas).To(Equal(ptr.To(int32(3))))
		})
	})

	Context("BuildAndSetFields", func() {
		It("should build and set fields from spec", func() {
			spec := createFullSpec()

			var target *distributedinference.Fields
			distributedinference.BuildAndSetFields(&target, spec, nil, nil)

			Expect(target).NotTo(BeNil())
			Expect(target.RestartPolicy).To(Equal(ptr.To("RecreateGroupOnPodRestart")))
			Expect(target.Replicas).To(Equal(ptr.To(int32(3))))
		})

		It("should build and set fields with additionalFields and urls", func() {
			spec := createFullSpec()
			additionalFields := map[string]interface{}{
				"readyReplicas": float64(2),
			}
			urls := []string{"https://my-service.example.com"}

			var target *distributedinference.Fields
			distributedinference.BuildAndSetFields(&target, spec, additionalFields, urls)

			Expect(target).NotTo(BeNil())
			Expect(target.ReplicasReady).To(Equal(ptr.To(int32(2))))
			Expect(target.ServingURL).To(Equal(ptr.To("https://my-service.example.com")))
		})

		It("should not set fields when spec is nil", func() {
			var target *distributedinference.Fields
			distributedinference.BuildAndSetFields(&target, nil, nil, nil)

			Expect(target).To(BeNil())
		})
	})
})

// Helper function to create a full spec for testing
func createFullSpec() *openapi.DistributedInferenceSpecSpec {
	servingPort := &openapi.DistributedInferenceServingPort{
		Port:              *openapi.NewNullableInt32(ptr.To(int32(8080))),
		AuthorizationType: *openapi.NewNullableDistributedInferenceServingPortAccessAuthorizationTypeEnum(ptr.To(openapi.DISTRIBUTEDINFERENCESERVINGPORTACCESSAUTHORIZATIONTYPEENUM_PUBLIC)),
		AuthorizedUsers:   []string{"user1", "user2"},
		AuthorizedGroups:  []string{"group1"},
		Protocol:          *openapi.NewNullableDistributedInferenceServingPortProtocol(ptr.To(openapi.DISTRIBUTEDINFERENCESERVINGPORTPROTOCOL_HTTP)),
		ExposeExternally:  *openapi.NewNullableBool(ptr.To(true)),
	}

	leaderSecurity := &openapi.DistributedInferenceLeaderWorkerSpecSecurity{
		RunAsUid: *openapi.NewNullableInt64(ptr.To(int64(1000))),
		RunAsGid: *openapi.NewNullableInt64(ptr.To(int64(1000))),
	}

	leaderCompute := &openapi.WorkspaceSpecSpecAllOfCompute{
		GpuComputeMode:    *openapi.NewNullableGpuComputeMode(ptr.To(openapi.GPUCOMPUTEMODE_TIME_SLICING)),
		GpuDevicesRequest: *openapi.NewNullableInt32(ptr.To(int32(4))),
		CpuCoreRequest:    *openapi.NewNullableFloat64(ptr.To(float64(8))),
	}

	leaderStorage := &openapi.DistributedInferenceLeaderWorkerSpecStorage{
		Pvc: []openapi.PvcInstance{
			{
				ClaimName:   *openapi.NewNullableString(ptr.To("model-cache")),
				Path:        *openapi.NewNullableString(ptr.To("/models")),
				ReadOnly:    *openapi.NewNullableBool(ptr.To(true)),
				Ephemeral:   *openapi.NewNullableBool(ptr.To(false)),
				DataSharing: *openapi.NewNullableBool(ptr.To(true)),
			},
		},
		ConfigMapVolume: []openapi.ConfigMapInstance{
			{
				Name:      *openapi.NewNullableString(ptr.To("inference-config")),
				MountPath: *openapi.NewNullableString(ptr.To("/config")),
			},
		},
		EmptyDirVolume: []openapi.EmptyDirInstance{
			{
				Name:      *openapi.NewNullableString(ptr.To("tmp-cache")),
				Path:      *openapi.NewNullableString(ptr.To("/tmp/cache")),
				SizeLimit: *openapi.NewNullableString(ptr.To("20Gi")),
			},
		},
	}

	leaderSpec := &openapi.DistributedInferenceLeaderWorkerSpec{
		Image:           *openapi.NewNullableString(ptr.To("leader-image:latest")),
		ImagePullPolicy: *openapi.NewNullableImagePullPolicy(ptr.To(openapi.IMAGEPULLPOLICY_ALWAYS)),
		Command:         *openapi.NewNullableString(ptr.To("python")),
		Args:            *openapi.NewNullableString(ptr.To("--leader")),
		WorkingDir:      *openapi.NewNullableString(ptr.To("/app")),
		NodeType:        *openapi.NewNullableString(ptr.To("gpu-nodes")),
		EnvironmentVariables: []openapi.EnvironmentVariable{
			{
				Name:  *openapi.NewNullableString(ptr.To("ROLE")),
				Value: *openapi.NewNullableString(ptr.To("leader")),
			},
		},
		Labels: []openapi.Label{
			{
				Name:  *openapi.NewNullableString(ptr.To("app")),
				Value: *openapi.NewNullableString(ptr.To("leader")),
			},
		},
		Security: *openapi.NewNullableDistributedInferenceLeaderWorkerSpecSecurity(leaderSecurity),
		Compute:  *openapi.NewNullableWorkspaceSpecSpecAllOfCompute(leaderCompute),
		Storage:  *openapi.NewNullableDistributedInferenceLeaderWorkerSpecStorage(leaderStorage),
	}

	workerSecurity := &openapi.DistributedInferenceLeaderWorkerSpecSecurity{
		RunAsUid: *openapi.NewNullableInt64(ptr.To(int64(2000))),
	}

	workerCompute := &openapi.WorkspaceSpecSpecAllOfCompute{
		GpuComputeMode:    *openapi.NewNullableGpuComputeMode(ptr.To(openapi.GPUCOMPUTEMODE_SM_SHARING)),
		GpuDevicesRequest: *openapi.NewNullableInt32(ptr.To(int32(2))),
	}

	workerSpec := &openapi.DistributedInferenceLeaderWorkerSpec{
		Image:           *openapi.NewNullableString(ptr.To("worker-image:latest")),
		ImagePullPolicy: *openapi.NewNullableImagePullPolicy(ptr.To(openapi.IMAGEPULLPOLICY_IF_NOT_PRESENT)),
		Command:         *openapi.NewNullableString(ptr.To("python")),
		Args:            *openapi.NewNullableString(ptr.To("--worker")),
		WorkingDir:      *openapi.NewNullableString(ptr.To("/app")),
		NodeType:        *openapi.NewNullableString(ptr.To("gpu-nodes")),
		EnvironmentVariables: []openapi.EnvironmentVariable{
			{
				Name:  *openapi.NewNullableString(ptr.To("ROLE")),
				Value: *openapi.NewNullableString(ptr.To("worker")),
			},
		},
		Labels: []openapi.Label{
			{
				Name:  *openapi.NewNullableString(ptr.To("app")),
				Value: *openapi.NewNullableString(ptr.To("worker")),
			},
		},
		Security: *openapi.NewNullableDistributedInferenceLeaderWorkerSpecSecurity(workerSecurity),
		Compute:  *openapi.NewNullableWorkspaceSpecSpecAllOfCompute(workerCompute),
	}

	return &openapi.DistributedInferenceSpecSpec{
		RestartPolicy: *openapi.NewNullableDistributedInferenceRestartPolicy(
			ptr.To(openapi.DISTRIBUTEDINFERENCERESTARTPOLICY_RECREATE_GROUP_ON_POD_RESTART)),
		StartupPolicy: *openapi.NewNullableDistributedInferenceStartupPolicy(
			ptr.To(openapi.DISTRIBUTEDINFERENCESTARTUPPOLICY_LEADER_CREATED)),
		Replicas:    *openapi.NewNullableInt32(ptr.To(int32(3))),
		Workers:     *openapi.NewNullableInt32(ptr.To(int32(2))),
		ServingPort: *openapi.NewNullableDistributedInferenceServingPort(servingPort),
		Leader:      *openapi.NewNullableDistributedInferenceLeaderWorkerSpec(leaderSpec),
		Worker:      *openapi.NewNullableDistributedInferenceLeaderWorkerSpec(workerSpec),
	}
}
