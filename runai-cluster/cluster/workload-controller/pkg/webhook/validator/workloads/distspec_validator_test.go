/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workloads_test

import (
	"context"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/builders/distributed/mpi"
	workloadUtils "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/resource/workload"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/webhook/validator/workloads"
	"k8s.io/utils/ptr"
)

var _ = Describe("DistSpecValidator", func() {
	var (
		distSpecValidator workloads.DistSpecValidator
		workload          *jobs.JobWorkload
		masterSpec        *jobs.JobWorkload
	)

	BeforeEach(func() {
		masterSpec = workloadUtils.NewJobWorkloadBuilder().WithJobType(constants.KindMpiJob).Build()
		workload = workloadUtils.NewJobWorkloadBuilder().WithMasterSpec(&masterSpec.Spec).WithJobType(constants.KindMpiJob).Build()
	})

	Context("MPI job with master as nil - allowed", func() {
		workload := workloadUtils.NewJobWorkloadBuilder().WithJobType(constants.KindMpiJob).Build()
		err := distSpecValidator.Validate(context.TODO(), workload)
		Expect(err).ToNot(HaveOccurred())
	})

	Context("MPI job with GPU set to 0 - allowed", func() {
		It("Master GPU is 0", func() {
			workload.MasterSpec.ResourceAllocationFields.Gpu = field_types.NewOptionalNumberField(ptr.To[float64](0))
			err := distSpecValidator.Validate(context.TODO(), workload)
			Expect(err).ToNot(HaveOccurred())
		})

		It("Master GPU devices is 0", func() {
			workload.MasterSpec.ResourceAllocationFields.GpuDevices = field_types.NewOptionalIntField(ptr.To(0))
			err := distSpecValidator.Validate(context.TODO(), workload)
			Expect(err).ToNot(HaveOccurred())
		})

		It("Master GPU limit is 0", func() {
			workload.MasterSpec.ResourceAllocationFields.GpuLimit = field_types.NewOptionalNumberField(ptr.To[float64](0))
			err := distSpecValidator.Validate(context.TODO(), workload)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	// this is added to verify that LauncherCreationPolicy is validated
	Context("MPI job very launcherCreationPolicyCheck called", func() {
		It("invalid LauncherCreationPolicy value", func() {
			workload = workloadUtils.NewJobWorkloadBuilder().WithJobType(constants.KindMpiJob).Build()
			workload.Spec.DistributedFields.MpiJob = &field_types.MpiField{
				Values: &field_types.MpiMembers{
					LauncherCreationPolicy: &field_types.StringField{
						Value: ptr.To("bla"),
					},
				},
			}
			err := distSpecValidator.Validate(context.TODO(), workload)
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError("LauncherCreationPolicy invalid value \"bla\""))
		})
	})

	Context("MPI job with GPU - rejected", func() {
		It("Master GPU is 1", func() {
			workload.MasterSpec.ResourceAllocationFields.Gpu = field_types.NewOptionalNumberField(ptr.To[float64](1))
			err := distSpecValidator.Validate(context.TODO(), workload)
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(mpi.ErrorMpiMasterNoGPUAllowed))
		})

		It("Master GPU devices is 1", func() {
			workload.MasterSpec.ResourceAllocationFields.GpuDevices = field_types.NewOptionalIntField(ptr.To(1))
			err := distSpecValidator.Validate(context.TODO(), workload)
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(mpi.ErrorMpiMasterNoGPUAllowed))
		})

		It("Master GPU limit is 1", func() {
			workload.MasterSpec.ResourceAllocationFields.GpuLimit = field_types.NewOptionalNumberField(ptr.To[float64](1))
			err := distSpecValidator.Validate(context.TODO(), workload)
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(mpi.ErrorMpiMasterNoGPUAllowed))
		})
	})

	Context("PyTorch job with GPU - allowed", func() {
		It("Master GPU is 1", func() {
			workload.Spec.JobType = constants.KindPyTorchJob
			workload.MasterSpec.ResourceAllocationFields.Gpu = field_types.NewOptionalNumberField(ptr.To[float64](1))
			err := distSpecValidator.Validate(context.TODO(), workload)
			Expect(err).ToNot(HaveOccurred())
		})

		It("Master GPU devices is 1", func() {
			workload.Spec.JobType = constants.KindPyTorchJob
			workload.MasterSpec.ResourceAllocationFields.GpuDevices = field_types.NewOptionalIntField(ptr.To(1))
			err := distSpecValidator.Validate(context.TODO(), workload)
			Expect(err).ToNot(HaveOccurred())
		})

		It("Master GPU limit is 1", func() {
			workload.Spec.JobType = constants.KindPyTorchJob
			workload.MasterSpec.ResourceAllocationFields.GpuLimit = field_types.NewOptionalNumberField(ptr.To[float64](1))
			err := distSpecValidator.Validate(context.TODO(), workload)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("Jax workload", func() {
		It("should allow creation No master spec", func() {
			workload.Spec.JobType = constants.KindJAXJob
			workload.MasterSpec = nil
			err := distSpecValidator.Validate(context.TODO(), workload)
			Expect(err).ToNot(HaveOccurred())
		})
		It("should allow creation with No master turn on", func() {
			workload.Spec.JobType = constants.KindJAXJob
			workload.Spec.NoMaster = field_types.NewOptionalBoolField(ptr.To(true))
			workload.MasterSpec = nil
			err := distSpecValidator.Validate(context.TODO(), workload)
			Expect(err).ToNot(HaveOccurred())
		})

	})

	Context("Restart policy", func() {
		It("only worker with restart policy never, back of limit check is nil", func() {
			workload.Spec.JobType = constants.KindPyTorchJob
			workload.MasterSpec = nil
			workload.Spec.RestartPolicy = &field_types.StringField{Value: ptr.To("Never")}
			err := distSpecValidator.Validate(context.TODO(), workload)
			Expect(err).ToNot(HaveOccurred())
		})

		It("only worker with restart policy never, back of limit check", func() {
			workload.Spec.JobType = constants.KindPyTorchJob
			workload.MasterSpec = nil
			workload.Spec.RestartPolicy = &field_types.StringField{Value: ptr.To("Never")}
			workload.Spec.RunPolicy = &field_types.RunPolicyField{Values: &field_types.RunPolicyMembers{
				BackoffLimit: &field_types.IntField{Value: ptr.To(1)},
			}}
			err := distSpecValidator.Validate(context.TODO(), workload)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Back of limit cannot be set with restart policy value 'Never'"))
		})

		It("worker and master with restart policy never, back of limit check is nil", func() {
			workload.Spec.JobType = constants.KindPyTorchJob
			workload.Spec.RestartPolicy = &field_types.StringField{Value: ptr.To("Never")}
			workload.MasterSpec.RestartPolicy = &field_types.StringField{Value: ptr.To("Never")}
			err := distSpecValidator.Validate(context.TODO(), workload)
			Expect(err).ToNot(HaveOccurred())
		})

		It("worker with restart policy never, back of limit check", func() {
			workload.Spec.JobType = constants.KindPyTorchJob
			workload.Spec.RestartPolicy = &field_types.StringField{Value: ptr.To("Never")}
			workload.Spec.RunPolicy = &field_types.RunPolicyField{Values: &field_types.RunPolicyMembers{
				BackoffLimit: &field_types.IntField{Value: ptr.To(1)},
			}}
			err := distSpecValidator.Validate(context.TODO(), workload)
			Expect(err).ToNot(HaveOccurred())
		})

		It("master with restart policy never, back of limit check", func() {
			workload.Spec.JobType = constants.KindPyTorchJob
			workload.MasterSpec.RestartPolicy = &field_types.StringField{Value: ptr.To("Never")}
			workload.Spec.RunPolicy = &field_types.RunPolicyField{Values: &field_types.RunPolicyMembers{
				BackoffLimit: &field_types.IntField{Value: ptr.To(1)},
			}}
			err := distSpecValidator.Validate(context.TODO(), workload)
			Expect(err).ToNot(HaveOccurred())
		})

		It("worker and master with restart policy never, back of limit check", func() {
			workload.Spec.JobType = constants.KindPyTorchJob
			workload.Spec.RestartPolicy = &field_types.StringField{Value: ptr.To("Never")}
			workload.MasterSpec.RestartPolicy = &field_types.StringField{Value: ptr.To("Never")}
			workload.Spec.RunPolicy = &field_types.RunPolicyField{Values: &field_types.RunPolicyMembers{
				BackoffLimit: &field_types.IntField{Value: ptr.To(1)},
			}}
			err := distSpecValidator.Validate(context.TODO(), workload)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Back of limit cannot be set with restart policy value 'Never'"))
		})
	})
})
