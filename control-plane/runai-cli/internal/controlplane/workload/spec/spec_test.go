/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package spec

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/golang/mock/gomock"
	clusterOpenapi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api"
	cpmocks "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/autogenerate/mocks/controlplane"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/consts"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/api/specset/schedulig/affinity"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/api/submit"
	nodes "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/node"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/workload/spec/spectype"
	clierrors "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/errors"
	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
	"github.com/spf13/cobra"
	"k8s.io/utils/ptr"
)

func TestJobs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Spec Tests")
}

var _ = Describe("Spec", func() {
	const (
		node1 = "node1"
		node2 = "node2"
	)

	var (
		ctx              = context.Background()
		clusterId        = "cluster-id"
		mockNodesClient  *cpmocks.MockNodeClient
		mockNodesService *nodes.BackendNodeService
	)

	BeforeEach(func() {
		mockNodesClient = cpmocks.NewMockNodeClient(gomock.NewController(GinkgoT()))
		ns := []clusterOpenapi.Node{
			{Name: node1},
			{Name: node2},
		}
		mockNodes := &clusterOpenapi.Nodes{Nodes: ns}
		mockNodesClient.EXPECT().GetNodeList(gomock.Any(), clusterId).Return(mockNodes, nil).AnyTimes()
		mockNodesService = nodes.NewBackendNodeService(mockNodesClient)
	})

	Describe("GenerateMasterSpec", func() {
		var (
			gs        *GeneralSpec
			workerEnv *[]string
			masterEnv *[]string
			compute   *submit.Compute
		)

		BeforeEach(func() {
			gs = NewDistributedSpec(api.AnyDistributed)
			workerEnv = &gs.flags.Container.EnvironmentVariables
			masterEnv = &gs.flags.Container.MasterEnvironmentVariables
			compute = &gs.flags.Compute

		})

		Context("MasterEnvironment", func() {
			It("empty env worker and master", func() {
				*workerEnv = []string{}

				err := gs.generateSpec(ctx, clusterId, mockNodesService)
				Expect(err).To(BeNil())
				spec, err := GenerateMasterSpec(gs, ctx, clusterId, mockNodesService)
				Expect(err).To(BeNil())
				Expect(len(spec.GetEnvironmentVariables())).To(Equal(0))
			})

			It("worker with env, master empty", func() {
				*workerEnv = []string{"this=that"}

				err := gs.generateSpec(ctx, clusterId, mockNodesService)
				Expect(err).To(BeNil())
				spec, err := GenerateMasterSpec(gs, ctx, clusterId, mockNodesService)
				Expect(err).To(BeNil())
				e := spec.GetEnvironmentVariables()
				Expect(len(e)).To(Equal(1))
				Expect(e[0].GetName()).To(Equal("this"))
				Expect(e[0].GetValue()).To(Equal("that"))
			})

			It("worker and master has the same env", func() {
				*workerEnv = []string{"this=that"}
				*masterEnv = []string{"this=that2"}

				err := gs.generateSpec(ctx, clusterId, mockNodesService)
				Expect(err).To(BeNil())
				spec, err := GenerateMasterSpec(gs, ctx, clusterId, mockNodesService)
				Expect(err).To(BeNil())
				e := spec.GetEnvironmentVariables()
				Expect(len(e)).To(Equal(1))
				Expect(e[0].GetName()).To(Equal("this"))
				Expect(e[0].GetValue()).To(Equal("that2"))
			})

			It("worker and master has the different env", func() {
				*workerEnv = []string{"this=that"}
				*masterEnv = []string{"this2=that2"}
				err := gs.generateSpec(ctx, clusterId, mockNodesService)
				Expect(err).To(BeNil())
				spec, err := GenerateMasterSpec(gs, ctx, clusterId, mockNodesService)
				Expect(err).To(BeNil())
				e := spec.GetEnvironmentVariables()
				Expect(len(e)).To(Equal(2))
				Expect(e).Should(ContainElement(openapi.EnvironmentVariable{
					Name:  *openapi.NewNullableString(ptr.To("this")),
					Value: *openapi.NewNullableString(ptr.To("that")),
				}))
				Expect(e).Should(ContainElement(openapi.EnvironmentVariable{
					Name:  *openapi.NewNullableString(ptr.To("this2")),
					Value: *openapi.NewNullableString(ptr.To("that2")),
				}))
			})

			It("master remove first header, add new one, update third", func() {
				*workerEnv = []string{"this=that", "this2=that2", "this3=that3"}
				*masterEnv = []string{"this=-", "this2=that9", "this8=that8"}
				err := gs.generateSpec(ctx, clusterId, mockNodesService)
				Expect(err).To(BeNil())
				spec, err := GenerateMasterSpec(gs, ctx, clusterId, mockNodesService)
				Expect(err).To(BeNil())
				e := spec.GetEnvironmentVariables()
				Expect(len(e)).To(Equal(3))
				Expect(e).Should(ContainElement(openapi.EnvironmentVariable{
					Name:  *openapi.NewNullableString(ptr.To("this2")),
					Value: *openapi.NewNullableString(ptr.To("that9")),
				}))
				Expect(e).Should(ContainElement(openapi.EnvironmentVariable{
					Name:  *openapi.NewNullableString(ptr.To("this3")),
					Value: *openapi.NewNullableString(ptr.To("that3")),
				}))
				Expect(e).Should(ContainElement(openapi.EnvironmentVariable{
					Name:  *openapi.NewNullableString(ptr.To("this8")),
					Value: *openapi.NewNullableString(ptr.To("that8")),
				}))
			})

			It("master remove middle header", func() {
				*workerEnv = []string{"this=that", "this2=that2", "this3=that3"}
				*masterEnv = []string{"this2=-"}
				err := gs.generateSpec(ctx, clusterId, mockNodesService)
				Expect(err).To(BeNil())
				spec, err := GenerateMasterSpec(gs, ctx, clusterId, mockNodesService)
				Expect(err).To(BeNil())
				e := spec.GetEnvironmentVariables()
				Expect(len(e)).To(Equal(2))
				Expect(e).Should(ContainElement(openapi.EnvironmentVariable{
					Name:  *openapi.NewNullableString(ptr.To("this")),
					Value: *openapi.NewNullableString(ptr.To("that")),
				}))
				Expect(e).Should(ContainElement(openapi.EnvironmentVariable{
					Name:  *openapi.NewNullableString(ptr.To("this3")),
					Value: *openapi.NewNullableString(ptr.To("that3")),
				}))
			})

			It("master remove middle header", func() {
				*workerEnv = []string{"this=that", "this2=that2", "this3=that3"}
				*masterEnv = []string{"this3=-"}
				err := gs.generateSpec(ctx, clusterId, mockNodesService)
				Expect(err).To(BeNil())
				spec, err := GenerateMasterSpec(gs, ctx, clusterId, mockNodesService)
				Expect(err).To(BeNil())
				e := spec.GetEnvironmentVariables()
				Expect(len(e)).To(Equal(2))
				Expect(e).Should(ContainElement(openapi.EnvironmentVariable{
					Name:  *openapi.NewNullableString(ptr.To("this")),
					Value: *openapi.NewNullableString(ptr.To("that")),
				}))
				Expect(e).Should(ContainElement(openapi.EnvironmentVariable{
					Name:  *openapi.NewNullableString(ptr.To("this2")),
					Value: *openapi.NewNullableString(ptr.To("that2")),
				}))
			})
		})

		Context("updateFlags", func() {
			It("external resource for worker - same for master", func() {
				compute.ExtendedResources = []string{"res1=ext1"}

				err := gs.generateSpec(ctx, clusterId, mockNodesService)
				Expect(err).To(BeNil())
				spec, err := GenerateMasterSpec(gs, ctx, clusterId, mockNodesService)
				Expect(err).To(BeNil())
				er := spec.Compute.Get().GetExtendedResources()
				Expect(len(er)).To(Equal(1))
				Expect(er[0].GetResource()).To(Equal("res1"))
				Expect(er[0].GetQuantity()).To(Equal("ext1"))
			})

			It("external resource for worker - different for master", func() {
				compute.ExtendedResources = []string{"res1=ext1"}
				compute.MasterExtendedResources = []string{"res2=ext2"}

				err := gs.generateSpec(ctx, clusterId, mockNodesService)
				Expect(err).To(BeNil())
				spec, err := GenerateMasterSpec(gs, ctx, clusterId, mockNodesService)
				Expect(err).To(BeNil())

				er := spec.Compute.Get().GetExtendedResources()
				Expect(len(er)).To(Equal(1))
				Expect(er[0].GetResource()).To(Equal("res2"))
				Expect(er[0].GetQuantity()).To(Equal("ext2"))
			})
		})
	})

	Describe("GenerateDistributedSpec", func() {
		var (
			gs       *GeneralSpec
			distSpec *openapi.DistributedSpec
			err      error
		)

		BeforeEach(func() {
			gs = NewDistributedSpec(api.AnyDistributed)
		})

		Context("GenerateDistributedSpec", func() {
			It("empty spec", func() {
				distSpec, err = gs.GenerateDistributedSpec(ctx, clusterId, mockNodesService)
				Expect(err).To(BeNil())
				Expect(distSpec).ToNot(BeNil())
				Expect(distSpec.GetSpec()).ToNot(BeNil())
				Expect(distSpec.HasMasterSpecSameAsWorker()).To(Equal(true))
				Expect(distSpec.HasMasterSpec()).To(Equal(false))
			})

			It("no master", func() {
				gs.flags.DistributedTraining.NoMaster = true

				distSpec, err = gs.GenerateDistributedSpec(ctx, clusterId, mockNodesService)
				Expect(err).To(BeNil())
				Expect(distSpec).ToNot(BeNil())
				Expect(distSpec.GetSpec()).ToNot(BeNil())
				Expect(distSpec.GetMasterSpecSameAsWorker()).To(Equal(false))
				Expect(distSpec.HasMasterSpec()).To(Equal(false))
			})

			It("unique master - device request", func() {
				gpuDeviceRequest := int32(3)
				gs.flags.Compute.GpuDeviceRequest.Set(ptr.To(gpuDeviceRequest))
				masterGpuDeviceRequest := int32(5)
				gs.flags.Compute.MasterGpuDeviceRequest.Set(ptr.To(masterGpuDeviceRequest))

				distSpec, err = gs.GenerateDistributedSpec(ctx, clusterId, mockNodesService)
				Expect(err).To(BeNil())
				Expect(distSpec).ToNot(BeNil())
				Expect(distSpec.GetSpec()).ToNot(BeNil())
				Expect(distSpec.GetMasterSpecSameAsWorker()).To(Equal(false))
				Expect(distSpec.HasMasterSpec()).To(Equal(true))

				worker := distSpec.GetSpec()
				compute := worker.GetCompute()
				Expect(compute.GetGpuDevicesRequest()).To(Equal(int32(3)))

				master := distSpec.GetMasterSpec()
				compute = master.GetCompute()
				Expect(compute.GetGpuDevicesRequest()).To(Equal(int32(5)))
			})

			It("unique master - portion request", func() {
				gs.flags.Compute.GpuPortionRequest.Set(ptr.To(0.3))
				gs.flags.Compute.GpuPortionLimit.Set(ptr.To(0.4))
				gs.flags.Compute.MasterGpuPortionRequest.Set(ptr.To(0.5))
				gs.flags.Compute.MasterGpuPortionLimit.Set(ptr.To(0.6))

				distSpec, err = gs.GenerateDistributedSpec(ctx, clusterId, mockNodesService)
				Expect(err).To(BeNil())
				Expect(distSpec).ToNot(BeNil())
				Expect(distSpec.GetSpec()).ToNot(BeNil())
				Expect(distSpec.GetMasterSpecSameAsWorker()).To(Equal(false))
				Expect(distSpec.HasMasterSpec()).To(Equal(true))

				worker := distSpec.GetSpec()
				compute := worker.GetCompute()
				Expect(compute.GetGpuPortionRequest()).To(Equal(0.3))
				Expect(compute.GetGpuPortionLimit()).To(Equal(0.4))

				master := distSpec.GetMasterSpec()
				compute = master.GetCompute()
				Expect(compute.GetGpuPortionRequest()).To(Equal(0.5))
				Expect(compute.GetGpuPortionLimit()).To(Equal(0.6))
			})

			Context("master and worker check of args and command - portion request", func() {
				It("worker no execute, master args", func() {
					gs.flags.Container.MasterArgs = "a2"

					distSpec, err = gs.GenerateDistributedSpec(ctx, clusterId, mockNodesService)
					Expect(err).To(BeNil())
					Expect(distSpec).ToNot(BeNil())
					Expect(distSpec.GetSpec()).ToNot(BeNil())
					Expect(distSpec.GetMasterSpecSameAsWorker()).To(Equal(false))
					Expect(distSpec.HasMasterSpec()).To(Equal(true))
					Expect(*distSpec.GetMasterSpec().Args.Get()).To(Equal("a2"))
					Expect(distSpec.GetMasterSpec().Command.IsSet()).To(BeFalse())
					Expect(distSpec.GetSpec().Args.IsSet()).To(BeFalse())
				})

				It("worker no execute, master command", func() {
					gs.flags.Container.MasterCommand = "a1"

					distSpec, err = gs.GenerateDistributedSpec(ctx, clusterId, mockNodesService)
					Expect(err).To(BeNil())
					Expect(distSpec).ToNot(BeNil())
					Expect(distSpec.GetSpec()).ToNot(BeNil())
					Expect(distSpec.GetMasterSpecSameAsWorker()).To(Equal(false))
					Expect(distSpec.HasMasterSpec()).To(Equal(true))
					Expect(*distSpec.GetMasterSpec().Command.Get()).To(Equal("a1"))
					Expect(distSpec.GetMasterSpec().Args.IsSet()).To(BeFalse())
					Expect(distSpec.GetSpec().Args.IsSet()).To(BeFalse())
					Expect(distSpec.GetSpec().Command.IsSet()).To(BeFalse())
				})

				It("worker no execute, master args+command", func() {
					gs.flags.Container.MasterCommand = "a1"
					gs.flags.Container.MasterArgs = "a2"

					distSpec, err = gs.GenerateDistributedSpec(ctx, clusterId, mockNodesService)
					Expect(err).To(BeNil())
					Expect(distSpec).ToNot(BeNil())
					Expect(distSpec.GetSpec()).ToNot(BeNil())
					Expect(distSpec.GetMasterSpecSameAsWorker()).To(Equal(false))
					Expect(distSpec.HasMasterSpec()).To(Equal(true))
					Expect(*distSpec.GetMasterSpec().Command.Get()).To(Equal("a1"))
					Expect(*distSpec.GetMasterSpec().Args.Get()).To(Equal("a2"))
					Expect(distSpec.GetSpec().Args.IsSet()).To(BeFalse())
				})

				It("unique master and args worker args", func() {
					gs.flags.Container.MasterCommand = "a1"
					gs.flags.Container.MasterArgs = "a2"
					gs.flags.Container.ExecuteArgs = "b2 b1"

					distSpec, err = gs.GenerateDistributedSpec(ctx, clusterId, mockNodesService)
					Expect(err).To(BeNil())
					Expect(distSpec).ToNot(BeNil())
					Expect(distSpec.GetSpec()).ToNot(BeNil())
					Expect(distSpec.GetMasterSpecSameAsWorker()).To(Equal(false))
					Expect(distSpec.HasMasterSpec()).To(Equal(true))
					Expect(*distSpec.GetMasterSpec().Command.Get()).To(Equal("a1"))
					Expect(*distSpec.GetMasterSpec().Args.Get()).To(Equal("a2"))
					Expect(*distSpec.GetSpec().Args.Get()).To(Equal("b2 b1"))
				})

				It("unique master and args worker command", func() {
					gs.flags.Container.MasterCommand = "a1"
					gs.flags.Container.MasterArgs = "a2"
					gs.flags.Container.ExecuteCommand = "b1 b2"

					distSpec, err = gs.GenerateDistributedSpec(ctx, clusterId, mockNodesService)
					Expect(err).To(BeNil())
					Expect(distSpec).ToNot(BeNil())
					Expect(distSpec.GetSpec()).ToNot(BeNil())
					Expect(distSpec.GetMasterSpecSameAsWorker()).To(Equal(false))
					Expect(distSpec.HasMasterSpec()).To(Equal(true))
					Expect(*distSpec.GetMasterSpec().Command.Get()).To(Equal("a1"))
					Expect(*distSpec.GetMasterSpec().Args.Get()).To(Equal("a2"))
					Expect(*distSpec.GetSpec().Command.Get()).To(Equal("b1 b2"))
					Expect(distSpec.GetSpec().Args.IsSet()).To(BeFalse())
				})
			})
		})
	})

	Describe("GenerateInferenceSpec", func() {
		var (
			gs       *GeneralSpec
			infwSpec *openapi.InferenceSpecSpec
			err      error
		)

		BeforeEach(func() {
			gs = NewInferenceSpec()
		})

		Context("GenerateInferenceSpec", func() {
			It("empty spec", func() {
				infwSpec, err = gs.GenerateInferenceSpec(ctx, clusterId, mockNodesService)
				Expect(err).To(BeNil())
				Expect(infwSpec).ToNot(BeNil())
				Expect(infwSpec.GetServingPort().Container.IsSet()).To(BeFalse())
				Expect(infwSpec.Image.IsSet()).To(BeFalse())
			})

			It("spec include inference update annotations", func() {
				gs.flags.General.Annotations = []string{"runai-update-command=runai inference update abc --image=ubuntu"}

				infwSpec, err = gs.GenerateInferenceSpec(ctx, clusterId, mockNodesService)
				Expect(err).To(BeNil())
				Expect(infwSpec).ToNot(BeNil())

			})

			It("spec is set", func() {
				// non inference unique flags
				gs.flags.General.Annotations = []string{"annotationK=annotationV"}
				gs.flags.Container.Image = "image"
				gs.flags.Security.RunAsUid = 5
				gs.flags.Scheduling.NodeType = "node"

				// inference unique flags
				gs.flags.Inference.ServingPort = "3000"
				gs.flags.Inference.MinReplicas = *openapi.NewNullableInt32(ptr.To[int32](1))
				gs.flags.Inference.InitialReplicas = *openapi.NewNullableInt32(ptr.To[int32](2))
				gs.flags.Inference.MaxReplicas = *openapi.NewNullableInt32(ptr.To[int32](3))
				gs.flags.Inference.ActivationReplicas = *openapi.NewNullableInt32(ptr.To[int32](4))
				gs.flags.Inference.ConcurrencyHardLimit = *openapi.NewNullableInt32(ptr.To[int32](5))
				gs.flags.Inference.ScaleToZeroRetentionSeconds = *openapi.NewNullableInt32(ptr.To[int32](6))
				gs.flags.Inference.ScaleDownDelaySeconds = *openapi.NewNullableInt32(ptr.To[int32](7))
				gs.flags.Inference.InitializationTimeoutSeconds = *openapi.NewNullableInt32(ptr.To[int32](8))
				gs.flags.Inference.Metric = "metric"
				gs.flags.Inference.MetricThreshold = *openapi.NewNullableInt32(ptr.To[int32](9))
				gs.flags.Inference.MetricThresholdPercentage = *openapi.NewNullableFloat32(ptr.To[float32](10.0))
				gs.flags.Inference.RequestTimeoutSeconds = *openapi.NewNullableInt32(ptr.To[int32](11))

				infwSpec, err = gs.GenerateInferenceSpec(ctx, clusterId, mockNodesService)
				Expect(err).To(BeNil())
				Expect(infwSpec).ToNot(BeNil())

				// non inference unique flags
				Expect(infwSpec.Annotations).ToNot(BeNil())
				Expect(infwSpec.Image.Get()).To(Equal(ptr.To("image")))
				Expect(len(infwSpec.Ports)).To(Equal(0))
				Expect(infwSpec.Security.Get().RunAsUid.Get()).To(Equal(ptr.To(int64(5))))
				Expect(infwSpec.NodeType.Get()).To(Equal(ptr.To("node")))

				// inference unique flags
				Expect(infwSpec.GetServingPort().Container.Get()).To(Equal(ptr.To(int32(3000))))
				Expect(infwSpec.GetAutoscaling().MinReplicas.Get()).To(Equal(ptr.To(int32(1))))
				Expect(infwSpec.GetAutoscaling().InitialReplicas.Get()).To(Equal(ptr.To(int32(2))))
				Expect(infwSpec.GetAutoscaling().MaxReplicas.Get()).To(Equal(ptr.To(int32(3))))
				Expect(infwSpec.GetAutoscaling().ActivationReplicas.Get()).To(Equal(ptr.To(int32(4))))
				Expect(infwSpec.GetAutoscaling().ConcurrencyHardLimit.Get()).To(Equal(ptr.To(int32(5))))
				Expect(infwSpec.GetAutoscaling().ScaleToZeroRetentionSeconds.Get()).To(Equal(ptr.To(int32(6))))
				Expect(infwSpec.GetAutoscaling().ScaleDownDelaySeconds.Get()).To(Equal(ptr.To(int32(7))))
				Expect(infwSpec.GetAutoscaling().InitializationTimeoutSeconds.Get()).To(BeNil())
				Expect(infwSpec.GetAutoscaling().Metric.Get()).To(Equal(ptr.To("metric")))
				Expect(infwSpec.GetAutoscaling().MetricThreshold.Get()).To(Equal(ptr.To(int32(9))))
				Expect(infwSpec.GetAutoscaling().MetricThresholdPercentage.Get()).To(Equal(ptr.To(float32(10.0))))
				Expect(infwSpec.GetServingConfiguration().InitializationTimeoutSeconds.Get()).To(Equal(ptr.To(int32(8))))
				Expect(infwSpec.GetServingConfiguration().RequestTimeoutSeconds.Get()).To(Equal(ptr.To(int32(11))))

				// unset fields are not set
				Expect(infwSpec.Compute.IsSet()).To(BeFalse())
				Expect(infwSpec.EnvironmentVariables).To(BeNil())
				Expect(infwSpec.Args.IsSet()).To(BeFalse())
				Expect(infwSpec.Command.IsSet()).To(BeFalse())
				Expect(infwSpec.Probes.IsSet()).To(BeFalse())
				Expect(infwSpec.Storage.IsSet()).To(BeFalse())
			})

			It("spec is set - only few fields - to validate the rest are unset", func() {
				gs.flags.Container.Image = "image"
				gs.flags.Inference.ServingPort = "3000"

				infwSpec, err = gs.GenerateInferenceSpec(ctx, clusterId, mockNodesService)
				Expect(err).To(BeNil())
				Expect(infwSpec).ToNot(BeNil())

				// set fields are set
				Expect(infwSpec.Image.Get()).To(Equal(ptr.To("image")))
				Expect(infwSpec.GetServingPort().Container.Get()).To(Equal(ptr.To(int32(3000))))

				// unset fields are not set
				Expect(infwSpec.Compute.IsSet()).To(BeFalse())
				Expect(len(infwSpec.EnvironmentVariables)).To(BeZero())
				Expect(infwSpec.Args.IsSet()).To(BeFalse())
				Expect(infwSpec.Command.IsSet()).To(BeFalse())
				Expect(len(infwSpec.Ports)).To(BeZero())
				Expect(infwSpec.Probes.IsSet()).To(BeFalse())
				Expect(infwSpec.Storage.IsSet()).To(BeFalse())
				Expect(infwSpec.Security.IsSet()).To(BeFalse())
				Expect(infwSpec.Autoscaling.IsSet()).To(BeFalse())
			})
		})

		Context("GenerateInferenceUpdateSpec", func() {
			It("update spec sets only the fields that are included in the update", func() {
				gs.flags.Container.Image = "image"
				gs.flags.Inference.MinReplicas = *openapi.NewNullableInt32(ptr.To[int32](0))
				gs.flags.Inference.MaxReplicas = *openapi.NewNullableInt32(ptr.To[int32](1))

				infwUpdateSpec, err := gs.GenerateInferenceUpdateSpec(ctx, clusterId, mockNodesService)
				Expect(err).To(BeNil())
				Expect(infwUpdateSpec).ToNot(BeNil())

				// expect to see the fields we asked to update as set, and the fields we didn't ask to update as unset
				Expect(infwUpdateSpec.Image.IsSet()).To(BeTrue())
				Expect(infwUpdateSpec.Autoscaling.IsSet()).To(BeTrue())
				Expect(infwUpdateSpec.Autoscaling.Get().MinReplicas.IsSet()).To(BeTrue())
				Expect(infwUpdateSpec.Autoscaling.Get().MaxReplicas.IsSet()).To(BeTrue())
				Expect(infwUpdateSpec.Autoscaling.Get().Metric.IsSet()).To(BeFalse())
				Expect(infwUpdateSpec.Compute.IsSet()).To(BeFalse())
				Expect(infwUpdateSpec.NodeType.IsSet()).To(BeFalse())
				Expect(len(infwUpdateSpec.EnvironmentVariables)).To(Equal(0))
				Expect(infwUpdateSpec.Args.IsSet()).To(BeFalse())
				Expect(infwUpdateSpec.Command.IsSet()).To(BeFalse())
				Expect(infwUpdateSpec.Probes.IsSet()).To(BeFalse())
			})

			It("update spec sets only the fields that are included in the update - smaller spec - no autoscaling", func() {
				gs.flags.Container.EnvironmentVariables = []string{"a=b"}

				infwUpdateSpec, err := gs.GenerateInferenceUpdateSpec(ctx, clusterId, mockNodesService)
				Expect(err).To(BeNil())
				Expect(infwUpdateSpec).ToNot(BeNil())

				// expect to see the fields we asked to update as set, and the fields we didn't ask to update as unset
				Expect(len(infwUpdateSpec.EnvironmentVariables)).To(Equal(1))
				Expect(infwUpdateSpec.Autoscaling.IsSet()).To(BeFalse())
				Expect(infwUpdateSpec.Compute.IsSet()).To(BeFalse())
				Expect(infwUpdateSpec.NodeType.IsSet()).To(BeFalse())
				Expect(infwUpdateSpec.Image.IsSet()).To(BeFalse())
				Expect(infwUpdateSpec.Args.IsSet()).To(BeFalse())
				Expect(infwUpdateSpec.Command.IsSet()).To(BeFalse())
				Expect(infwUpdateSpec.Probes.IsSet()).To(BeFalse())
			})
		})
	})

	Describe("CheckDistributeFlags", func() {
		var (
			gs   *GeneralSpec
			dist *submit.DistributedTraining
			err  error
		)

		BeforeEach(func() {
			gs = NewDistributedSpec(api.AnyDistributed)
			dist = &gs.flags.DistributedTraining
		})

		Context("framework", func() {
			It("empty spec", func() {
				err = gs.CheckDistributeFlags()
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("invalid value '' for DistributedFramework: valid values are [MPI PyTorch TF XGBoost JAX]"))
			})

			It("framework MPI", func() {
				dist.DistributedFramework = "MPI"
				err = gs.CheckDistributeFlags()
				Expect(err).To(BeNil())
			})

			It("framework TF", func() {
				dist.DistributedFramework = "TF"
				err = gs.CheckDistributeFlags()
				Expect(err).To(BeNil())
			})

			It("framework XGBoost", func() {
				dist.DistributedFramework = "XGBoost"
				err = gs.CheckDistributeFlags()
				Expect(err).To(BeNil())
			})

			It("framework PyTorch", func() {
				dist.DistributedFramework = "PyTorch"
				err = gs.CheckDistributeFlags()
				Expect(err).To(BeNil())
			})
		})

		Context("master - true", func() {
			BeforeEach(func() {
				dist.NoMaster = true
			})

			It("MPI - fail", func() {
				dist.DistributedFramework = "MPI"
				err = gs.CheckDistributeFlags()
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal(fmt.Sprintf(FrameworkNoSupportMaster, "MPI")))
			})

			It("XGBoost - fail", func() {
				dist.DistributedFramework = "XGBoost"
				err = gs.CheckDistributeFlags()
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal(fmt.Sprintf(FrameworkNoSupportMaster, "XGBoost")))
			})

			It("TF - ok", func() {
				dist.DistributedFramework = "TF"
				err = gs.CheckDistributeFlags()
				Expect(err).To(BeNil())
			})

			It("PyTorch - ok", func() {
				dist.DistributedFramework = "PyTorch"
				err = gs.CheckDistributeFlags()
				Expect(err).To(BeNil())
			})
		})

		Context("slots-per-worker > 0", func() {
			BeforeEach(func() {
				dist.SlotsPerWorker = 2
			})

			It("MPI - ok", func() {
				dist.DistributedFramework = "MPI"
				err = gs.CheckDistributeFlags()
				Expect(err).To(BeNil())
			})

			It("XGBoost - fail", func() {
				dist.DistributedFramework = "XGBoost"
				err = gs.CheckDistributeFlags()
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal(fmt.Sprintf(FrameworkNoSupportSlotsPerWorker, "XGBoost")))
			})

			It("TF - fail", func() {
				dist.DistributedFramework = "TF"
				err = gs.CheckDistributeFlags()
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal(fmt.Sprintf(FrameworkNoSupportSlotsPerWorker, "TF")))
			})

			It("PyTorch - fail", func() {
				dist.DistributedFramework = "PyTorch"
				err = gs.CheckDistributeFlags()
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal(fmt.Sprintf(FrameworkNoSupportSlotsPerWorker, "PyTorch")))
			})
		})

		Context("ssh-auth-mount-path != \"\"", func() {
			BeforeEach(func() {
				dist.SshAuthMountPath = "/home"
			})

			It("MPI - ok", func() {
				dist.DistributedFramework = "MPI"
				err = gs.CheckDistributeFlags()
				Expect(err).To(BeNil())
			})

			It("XGBoost - fail", func() {
				dist.DistributedFramework = "XGBoost"
				err = gs.CheckDistributeFlags()
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal(fmt.Sprintf(FrameworkNoSupportSshAuthMountPath, "XGBoost")))
			})

			It("TF - fail", func() {
				dist.DistributedFramework = "TF"
				err = gs.CheckDistributeFlags()
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal(fmt.Sprintf(FrameworkNoSupportSshAuthMountPath, "TF")))
			})

			It("PyTorch - fail", func() {
				dist.DistributedFramework = "PyTorch"
				err = gs.CheckDistributeFlags()
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal(fmt.Sprintf(FrameworkNoSupportSshAuthMountPath, "PyTorch")))
			})
		})

		Context("launcher-creation-policy = AtStartup", func() {
			BeforeEach(func() {
				dist.LauncherCreationPolicy = *openapi.NewNullableMpiLauncherCreationPolicy(ptr.To(openapi.MPILAUNCHERCREATIONPOLICY_AT_STARTUP))
			})

			It("MPI - ok", func() {
				dist.DistributedFramework = "MPI"
				err = gs.CheckDistributeFlags()
				Expect(err).To(BeNil())
			})

			It("XGBoost - fail", func() {
				dist.DistributedFramework = "XGBoost"
				err = gs.CheckDistributeFlags()
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal(fmt.Sprintf(FrameworkNoSupportLauncherCreationPolicy, "XGBoost")))
			})

			It("TF - fail", func() {
				dist.DistributedFramework = "TF"
				err = gs.CheckDistributeFlags()
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal(fmt.Sprintf(FrameworkNoSupportLauncherCreationPolicy, "TF")))
			})

			It("PyTorch - fail", func() {
				dist.DistributedFramework = "PyTorch"
				err = gs.CheckDistributeFlags()
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal(fmt.Sprintf(FrameworkNoSupportLauncherCreationPolicy, "PyTorch")))
			})
		})

		Context("min/max replica > 0", func() {
			BeforeEach(func() {
				dist.MinReplica = 2
				dist.MaxReplica = 3
			})

			It("MPI - fail", func() {
				dist.DistributedFramework = "MPI"
				err = gs.CheckDistributeFlags()
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal(fmt.Sprintf(FrameworkNoSupportMinMaxRep, "MPI")))
			})

			It("XGBoost - fail", func() {
				dist.DistributedFramework = "XGBoost"
				err = gs.CheckDistributeFlags()
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal(fmt.Sprintf(FrameworkNoSupportMinMaxRep, "XGBoost")))
			})

			It("TF - fail", func() {
				dist.DistributedFramework = "TF"
				err = gs.CheckDistributeFlags()
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal(fmt.Sprintf(FrameworkNoSupportMinMaxRep, "TF")))
			})

			It("PyTorch - min higher then max", func() {
				dist.MinReplica = 7
				dist.MaxReplica = 6
				dist.DistributedFramework = "PyTorch"
				err = gs.CheckDistributeFlags()
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal(fmt.Sprintf(FrameworkNoSupportMaxEqMin, dist.MaxReplica, dist.MinReplica)))
			})

			It("PyTorch - min eq max - ok", func() {
				dist.MinReplica = 7
				dist.MaxReplica = 7
				dist.DistributedFramework = "PyTorch"
				err = gs.CheckDistributeFlags()
				Expect(err).To(BeNil())
			})

			It("PyTorch - ok", func() {
				dist.DistributedFramework = "PyTorch"
				err = gs.CheckDistributeFlags()
				Expect(err).To(BeNil())
			})
		})

		Context("segment size", func() {
			BeforeEach(func() {
				dist.SegmentSize.Set(ptr.To[int32](2))
			})

			It("PyTorch - ok", func() {
				dist.DistributedFramework = "PyTorch"
				err = gs.CheckDistributeFlags()
				Expect(err).To(BeNil())
			})

			It("non-PyTorch - fail", func() {
				dist.DistributedFramework = "TF"
				err = gs.CheckDistributeFlags()
				Expect(err).To(MatchError(fmt.Sprintf(FrameworkNoSupportSegmentSize, "TF")))
			})

			It("zero - fail", func() {
				dist.DistributedFramework = "PyTorch"
				dist.SegmentSize.Set(ptr.To[int32](0))
				err = gs.CheckDistributeFlags()
				Expect(err).To(MatchError(SegmentSizeMustBePositive))
			})

			It("elastic replicas - ok", func() {
				dist.DistributedFramework = "PyTorch"
				dist.MinReplica = 1
				dist.MaxReplica = 4
				err = gs.CheckDistributeFlags()
				Expect(err).To(BeNil())
			})
		})
	})

	Describe("Master asset submission", func() {
		It("omits master assets when the workload has no master", func() {
			gs := NewDistributedSpec(api.PyTorch)
			gs.flags.DistributedTraining.NoMaster = true
			gs.flags.Container.MasterDatasourceAssetIds = []string{"master-datasource-id"}

			Expect(gs.GetMasterAssetSubmissionIds()).To(BeNil())
		})
	})

	Describe("CheckLifeCycleFlags", func() {
		var (
			gs                     *GeneralSpec
			lc                     *submit.Lifecycle
			restartPolicyNever, _  = openapi.NewRestartPolicyFromValue("Never")
			restartPolicyAlways, _ = openapi.NewRestartPolicyFromValue("Always")
			backoffLimitOne        = *openapi.NewNullableInt32(ptr.To[int32](1))
			err                    error
		)
		BeforeEach(func() {
			gs = NewDistributedSpec(api.AnyDistributed)
			lc = &gs.flags.Lifecycle
		})
		Context("Restart policy", func() {
			Context("Negative test", func() {
				It("Should fail with backoff > 0 and restart policy set to never", func() {
					lc.BackoffLimit = backoffLimitOne

					lc.RestartPolicy.Set(restartPolicyNever)
					err = gs.Validate()
					Expect(err).ToNot(BeNil())
				})
				It("Should fail with backoff > 0 and restart policy and master restart policy set to never", func() {
					lc.BackoffLimit = backoffLimitOne
					lc.RestartPolicy.Set(restartPolicyNever)
					lc.MasterRestartPolicy.Set(restartPolicyNever)
					err = gs.Validate()
					Expect(err).ToNot(BeNil())
				})
			})
			Context("Positive test", func() {
				It("Should succeed with backoff > 0 and restart policy set to Always", func() {
					lc.BackoffLimit = backoffLimitOne
					lc.RestartPolicy.Set(restartPolicyAlways)
					err = gs.Validate()
					Expect(err).To(BeNil())
				})
				It("Should succeed with backoff > 0 and restart policy set to always and master restart policy set to never", func() {
					lc.BackoffLimit = backoffLimitOne
					lc.RestartPolicy.Set(restartPolicyAlways)
					lc.MasterRestartPolicy.Set(restartPolicyNever)
					err = gs.Validate()
					Expect(err).To(BeNil())
				})
				It("Should succeed with backoff > 0 and restart policy is not set and master restart policy set to never", func() {
					lc.BackoffLimit = backoffLimitOne
					lc.MasterRestartPolicy.Set(restartPolicyNever)
					err = gs.Validate()
					Expect(err).To(BeNil())
				})
				It("Should succeed with backoff > 0, restart policy is not set", func() {
					lc.BackoffLimit = backoffLimitOne
					err = gs.Validate()
					Expect(err).To(BeNil())
				})
				It("restart policy is set to never, no backoff limit", func() {
					lc.RestartPolicy.Set(restartPolicyAlways)
					err = gs.Validate()
					Expect(err).To(BeNil())
				})
			})
		})
		Context("Preemptibility", func() {
			var (
				preemptible, _    = openapi.NewPreemptibilityFromValue("preemptible")
				nonPreemptible, _ = openapi.NewPreemptibilityFromValue("non-preemptible")
			)
			Context("Lifecycle (workspace/training/distributed)", func() {
				It("should set preemptibility to preemptible when flag is provided", func() {
					gs = NewWorkspaceSpec()
					lc = &gs.flags.Lifecycle
					lc.Preemptibility.Set(preemptible)

					err = gs.generateSpec(ctx, clusterId, mockNodesService)
					Expect(err).To(BeNil())

					Expect(gs.fullSpec.Preemptibility.IsSet()).To(BeTrue())
					Expect(*gs.fullSpec.Preemptibility.Get()).To(Equal(openapi.PREEMPTIBILITY_PREEMPTIBLE))
				})

				It("should set preemptibility to non-preemptible when flag is provided", func() {
					gs = NewWorkspaceSpec()
					lc = &gs.flags.Lifecycle
					lc.Preemptibility.Set(nonPreemptible)

					err = gs.generateSpec(ctx, clusterId, mockNodesService)
					Expect(err).To(BeNil())

					Expect(gs.fullSpec.Preemptibility.IsSet()).To(BeTrue())
					Expect(*gs.fullSpec.Preemptibility.Get()).To(Equal(openapi.PREEMPTIBILITY_NON_PREEMPTIBLE))
				})

				It("preemptibility is not set when flag is not provided", func() {
					gs = NewWorkspaceSpec()

					err = gs.generateSpec(ctx, clusterId, mockNodesService)
					Expect(err).To(BeNil())

					Expect(gs.fullSpec.Preemptibility.IsSet()).To(BeFalse())
				})

				It("should set preemptibility for distributed workloads", func() {
					gs = NewDistributedSpec(api.MPI)
					lc = &gs.flags.Lifecycle
					lc.Preemptibility.Set(preemptible)

					err = gs.generateSpec(ctx, clusterId, mockNodesService)
					Expect(err).To(BeNil())

					Expect(gs.fullSpec.Preemptibility.IsSet()).To(BeTrue())
					Expect(*gs.fullSpec.Preemptibility.Get()).To(Equal(openapi.PREEMPTIBILITY_PREEMPTIBLE))
				})
			})

			Context("Inference", func() {
				It("should set preemptibility for inference workloads", func() {
					gs = NewInferenceSpec()
					gs.flags.Lifecycle.Preemptibility.Set(preemptible)

					infwSpec, err := gs.GenerateInferenceSpec(ctx, clusterId, mockNodesService)
					Expect(err).To(BeNil())

					Expect(infwSpec.Preemptibility.IsSet()).To(BeTrue())
					Expect(*infwSpec.Preemptibility.Get()).To(Equal(openapi.PREEMPTIBILITY_PREEMPTIBLE))
				})

				It("should set preemptibility to non-preemptible for inference", func() {
					gs = NewInferenceSpec()
					gs.flags.Lifecycle.Preemptibility.Set(nonPreemptible)

					infwSpec, err := gs.GenerateInferenceSpec(ctx, clusterId, mockNodesService)
					Expect(err).To(BeNil())

					Expect(infwSpec.Preemptibility.IsSet()).To(BeTrue())
					Expect(*infwSpec.Preemptibility.Get()).To(Equal(openapi.PREEMPTIBILITY_NON_PREEMPTIBLE))
				})

				It("preemptibility is not set for inference when flag not provided", func() {
					gs = NewInferenceSpec()

					infwSpec, err := gs.GenerateInferenceSpec(ctx, clusterId, mockNodesService)
					Expect(err).To(BeNil())

					Expect(infwSpec.Preemptibility.IsSet()).To(BeFalse())
				})
			})
		})
	})

	Describe("GeneralSpec", func() {
		var (
			generalSpec *GeneralSpec
		)

		Context("when created with NewInferenceSpec function", func() {
			BeforeEach(func() {
				generalSpec = NewInferenceSpec()
			})

			It("should have workloadType set to Inference", func() {
				Expect(generalSpec.workloadType).To(Equal(spectype.Inference))
			})

			It("should have framework set to Knative", func() {
				Expect(generalSpec.framework).To(Equal(api.Knative))
			})
		})
	})

	Describe("SetUpdateFlags", func() {
		var (
			cmd         *cobra.Command
			generalSpec *GeneralSpec
		)

		BeforeEach(func() {
			cmd = &cobra.Command{Use: "test"}
			generalSpec = NewInferenceSpec()
			generalSpec.SetUpdateFlags(cmd)
		})

		Context("when setting update flags for inference", func() {
			DescribeTable("should register expected flags",
				func(flagName string) {
					Expect(cmd.Flag(flagName)).NotTo(BeNil())
				},
				Entry("command", consts.AttributeNameCommand),
				Entry("image", consts.AttributeNameImage),
				Entry("metric", consts.AttributeNameMetric),
				Entry("cpu-core-request", consts.AttributeNameCpuCoreRequest),
				Entry("cpu-core-limit", consts.AttributeNameCpuCoreLimit),
				Entry("cpu-memory-request", consts.AttributeNameCpuMemoryRequest),
				Entry("cpu-memory-limit", consts.AttributeNameCpuMemoryLimit),
				Entry("gpu-devices-request", consts.AttributeNameGpuDeviceRequest),
				Entry("gpu-memory-request", consts.AttributeNameGpuMemoryRequest),
				Entry("gpu-memory-limit", consts.AttributeNameGpuMemoryLimit),
				Entry("large-shm", consts.AttributeNameLargeShm),
				Entry("node-pools", consts.AttributeNameNodePools),
			)
		})
	})

	Describe("Node Affinity Features", func() {
		Context("construct nodeAffinityRequired from nodes to exclude", func() {
			var (
				nodesToExclude = []string{node1, node2}
			)

			It("Training", func() {
				spec := NewTrainingSpec()
				spec.flags.Scheduling.ExcludeNodes = nodesToExclude
				generated, err := spec.GenerateTrainingSpec(ctx, clusterId, mockNodesService)
				Expect(err).NotTo(HaveOccurred())

				validateNodeExclusionAffinity(spec.flags.Scheduling.ExcludeNodes, generated.NodeAffinityRequired.Get())
			})

			It("Workspace", func() {
				spec := NewWorkspaceSpec()
				spec.flags.Scheduling.ExcludeNodes = nodesToExclude
				generated, err := spec.GenerateWorkspaceSpec(ctx, clusterId, mockNodesService)
				Expect(err).NotTo(HaveOccurred())

				validateNodeExclusionAffinity(spec.flags.Scheduling.ExcludeNodes, generated.NodeAffinityRequired.Get())
			})

			It("Distributed", func() {
				spec := NewDistributedSpec(api.MPI)
				spec.flags.Scheduling.ExcludeNodes = nodesToExclude
				spec.flags.Container.MasterCommand = "sleep"
				generated, err := spec.GenerateDistributedSpec(ctx, clusterId, mockNodesService)
				Expect(err).NotTo(HaveOccurred())

				validateNodeExclusionAffinity(spec.flags.Scheduling.ExcludeNodes, generated.Spec.NodeAffinityRequired.Get())
				validateNodeExclusionAffinity(spec.flags.Scheduling.ExcludeNodes, generated.MasterSpec.Get().NodeAffinityRequired.Get())
			})

			It("Inference", func() {
				spec := NewInferenceSpec()
				spec.flags.Scheduling.ExcludeNodes = nodesToExclude
				generated, err := spec.GenerateInferenceSpec(ctx, clusterId, mockNodesService)
				Expect(err).NotTo(HaveOccurred())

				validateNodeExclusionAffinity(spec.flags.Scheduling.ExcludeNodes, generated.NodeAffinityRequired.Get())
			})

			It("Inference - Update", func() {
				spec := NewInferenceSpec()
				spec.flags.Scheduling.ExcludeNodes = nodesToExclude
				generated, err := spec.GenerateInferenceUpdateSpec(ctx, clusterId, mockNodesService)
				Expect(err).NotTo(HaveOccurred())

				validateNodeExclusionAffinity(spec.flags.Scheduling.ExcludeNodes, generated.NodeAffinityRequired.Get())
			})

			It("validate nodes to exclude exists", func() {
				nonExistingNode1 := "blabla1"
				nonExistingNode2 := "blabla2"
				spec := NewTrainingSpec()
				spec.flags.Scheduling.ExcludeNodes = []string{node1, nonExistingNode1, nonExistingNode2}
				_, err := spec.GenerateTrainingSpec(ctx, clusterId, mockNodesService)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal(fmt.Sprintf("The following node names to exclude do not exist in the cluster: %s, %s", nonExistingNode1, nonExistingNode2)))
			})

			It("skips validation if user doesn't have permissions", func() {
				errClusterId := "errClusterId"
				forbiddenErr := &clierrors.HttpError{
					Status: http.StatusText(http.StatusForbidden),
				}
				mockNodesClient.EXPECT().GetNodeList(gomock.Any(), errClusterId).Return(nil, forbiddenErr).AnyTimes()

				nonExistingNode1 := "blabla1"
				nonExistingNode2 := "blabla2"
				spec := NewTrainingSpec()
				spec.flags.Scheduling.ExcludeNodes = []string{node1, nonExistingNode1, nonExistingNode2}
				generated, err := spec.GenerateTrainingSpec(ctx, errClusterId, mockNodesService)
				Expect(err).NotTo(HaveOccurred())

				validateNodeExclusionAffinity(spec.flags.Scheduling.ExcludeNodes, generated.NodeAffinityRequired.Get())
			})
		})
	})
})

func validateNodeExclusionAffinity(excludedNodes []string, actualNodeAffinityRequired *openapi.NodeAffinityRequired) {
	Expect(actualNodeAffinityRequired).NotTo(BeNil())
	Expect(actualNodeAffinityRequired.NodeSelectorTerms).To(HaveLen(1))
	Expect(actualNodeAffinityRequired.NodeSelectorTerms[0].MatchExpressions).To(HaveLen(1))

	Expect(actualNodeAffinityRequired.NodeSelectorTerms[0].MatchExpressions[0].Key).To(Equal(affinity.K8sHostnameLabelName))
	Expect(actualNodeAffinityRequired.NodeSelectorTerms[0].MatchExpressions[0].Operator).To(Equal(openapi.MATCHEXPRESSIONOPERATOR_NOT_IN))
	Expect(actualNodeAffinityRequired.NodeSelectorTerms[0].MatchExpressions[0].Values).To(HaveLen(len(excludedNodes)))
	Expect(actualNodeAffinityRequired.NodeSelectorTerms[0].MatchExpressions[0].Values).To(ContainElements(excludedNodes))
}

var _ = Describe("Distributed Inference", func() {
	const clusterId = "cluster-id"

	var (
		ctx              = context.Background()
		mockNodesClient  *cpmocks.MockNodeClient
		mockNodesService *nodes.BackendNodeService
		gs               *GeneralSpec
	)

	BeforeEach(func() {
		mockNodesClient = cpmocks.NewMockNodeClient(gomock.NewController(GinkgoT()))
		mockNodes := &clusterOpenapi.Nodes{Nodes: []clusterOpenapi.Node{{Name: "node1"}}}
		mockNodesClient.EXPECT().GetNodeList(gomock.Any(), clusterId).Return(mockNodes, nil).AnyTimes()
		mockNodesService = nodes.NewBackendNodeService(mockNodesClient)

		gs = NewDistributedSpec(api.LWS)
		gs.flags.Container.Image = "vllm/vllm:latest"
	})

	Context("Worker-As-Leader Logic", func() {
		It("worker-as-leader=true with no worker flags uses same spec", func() {
			gs.flags.DistributedInference.WorkerAsLeader = true
			gs.flags.DistributedInference.Workers.Set(ptr.To(int32(3)))
			gs.flags.Compute.GpuDeviceRequest.Set(ptr.To(int32(4)))

			spec, err := gs.GenerateDistributedSpec(ctx, clusterId, mockNodesService)
			Expect(err).To(BeNil())
			Expect(spec.GetMasterSpecSameAsWorker()).To(BeTrue())
		})

		It("worker-as-leader=true with worker override creates separate specs", func() {
			gs.flags.DistributedInference.WorkerAsLeader = true
			gs.flags.DistributedInference.Workers.Set(ptr.To(int32(3)))
			gs.flags.Compute.GpuDeviceRequest.Set(ptr.To(int32(4)))
			gs.flags.Compute.WorkerGpuDevicesRequest.Set(ptr.To(int32(2)))

			spec, err := gs.GenerateDistributedSpec(ctx, clusterId, mockNodesService)
			Expect(err).To(BeNil())
			Expect(spec.GetMasterSpecSameAsWorker()).To(BeFalse())
		})

		It("worker-as-leader=false creates separate specs", func() {
			gs.flags.DistributedInference.WorkerAsLeader = false
			gs.flags.DistributedInference.Workers.Set(ptr.To(int32(3)))

			spec, err := gs.GenerateDistributedSpec(ctx, clusterId, mockNodesService)
			Expect(err).To(BeNil())
			Expect(spec.GetMasterSpecSameAsWorker()).To(BeFalse())
		})
	})

	Context("Serving Port", func() {
		It("should parse serving port with expose-externally", func() {
			gs.flags.DistributedInference.ServingPort = "container=8080,protocol=http,expose-externally"
			gs.flags.DistributedInference.Replicas = 1
			gs.flags.DistributedInference.Workers.Set(ptr.To(int32(1)))

			distInfSpec, err := gs.GenerateDistributedInferenceSpec(ctx, clusterId, mockNodesService)
			Expect(err).To(BeNil())
			Expect(distInfSpec.ServingPort.IsSet()).To(BeTrue())
			Expect(*distInfSpec.ServingPort.Get().Port.Get()).To(Equal(int32(8080)))
			Expect(*distInfSpec.ServingPort.Get().ExposeExternally.Get()).To(BeTrue())
		})

		It("should parse serving port with expose-externally=false", func() {
			gs.flags.DistributedInference.ServingPort = "container=8080,protocol=http,expose-externally=false"
			gs.flags.DistributedInference.Replicas = 1
			gs.flags.DistributedInference.Workers.Set(ptr.To(int32(1)))

			distInfSpec, err := gs.GenerateDistributedInferenceSpec(ctx, clusterId, mockNodesService)
			Expect(err).To(BeNil())
			Expect(distInfSpec.ServingPort.IsSet()).To(BeTrue())
			Expect(*distInfSpec.ServingPort.Get().Port.Get()).To(Equal(int32(8080)))
			Expect(*distInfSpec.ServingPort.Get().ExposeExternally.Get()).To(BeFalse())
		})
	})

	Context("Worker Environment Merge Logic", func() {
		It("mergeEnvVars should merge leader and worker vars with override", func() {
			leaderVars := []string{"VAR1=a", "VAR2=b"}
			workerVars := []string{"VAR2=c", "VAR3=d"}
			result := mergeEnvVars(leaderVars, workerVars)

			// Should have 3 vars: VAR1 from leader, VAR2 overridden by worker, VAR3 from worker
			Expect(len(result)).To(Equal(3))
			resultMap := make(map[string]string)
			for _, env := range result {
				parts := strings.SplitN(env, "=", 2)
				if len(parts) == 2 {
					resultMap[parts[0]] = parts[1]
				}
			}

			Expect(resultMap["VAR1"]).To(Equal("a"), "VAR1 should be inherited from leader")
			Expect(resultMap["VAR2"]).To(Equal("c"), "VAR2 should be overridden by worker")
			Expect(resultMap["VAR3"]).To(Equal("d"), "VAR3 should be from worker")
		})

		It("getEnvVarName should extract name from various formats", func() {
			Expect(getEnvVarName("MY_VAR=complex=value")).To(Equal("MY_VAR"))
			Expect(getEnvVarName("SECRET=secret,key=key")).To(Equal("SECRET"))
			Expect(getEnvVarName("invalid")).To(Equal(""))
		})
	})
})
