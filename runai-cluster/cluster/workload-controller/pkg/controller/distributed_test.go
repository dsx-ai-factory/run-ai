/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package controller

import (
	"github.com/run-ai/runai/runai-cluster/cluster/common/extensions"
	"github.com/run-ai/runai/runai-cluster/cluster/common/extensions/distributed"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/managers"
	"k8s.io/client-go/tools/record"

	kubeflowv1 "github.com/kubeflow/training-operator/pkg/apis/kubeflow.org/v1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"

	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	dp "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedpolicy/v2alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedworkload/v2alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	logger "sigs.k8s.io/controller-runtime/pkg/log"
)

const jobTypeUnknown = "UnknownJob"

var testLoggerDw = logger.Log.WithName("test.dw")
var _ = Describe("DistributedWorkloadReconciler", func() {
	var (
		reconciler *DistributedWorkloadReconciler
		workload   *v2alpha1.DistributedWorkload
	)

	BeforeEach(func() {

		enabledFrameworks := make(map[string]DistributedJobFrameworksStatus, 3)
		enabledFrameworks[constants.KindTFJob] = DistributedJobFrameworksStatus{JobObject: &kubeflowv1.TFJob{}}
		enabledFrameworks[constants.KindPyTorchJob] = DistributedJobFrameworksStatus{JobObject: &kubeflowv1.PyTorchJob{}}
		enabledFrameworks[constants.KindMpiJob] = DistributedJobFrameworksStatus{JobObject: &kubeflowv1.MPIJob{}}

		reconciler = NewSimpleDistributedWorkloadReconciler(nil, nil, enabledFrameworks,
			DistributedJobCRDsSupportedMap, record.NewFakeRecorder(200), nil, nil, Options{FailureResourceCleanupPolicy: managers.GetDefaultFailureCleanupPolicy()})
		workload = &v2alpha1.DistributedWorkload{

			ObjectMeta: metav1.ObjectMeta{},
			Spec: dp.DistributedSpec{
				DistributedFields: field_groups.DistributedFields{
					JobType: jobTypeUnknown,
				},
			},
			MasterSpec: nil,
			Status:     v2alpha1.DistributedWorkloadStatus{},
		}
	})

	Describe("validateJobFramework", func() {
		Context("when job type is not recognized", func() {
			It("returns an error", func() {
				err := reconciler.validateJobFramework(workload, testLoggerDw, "test")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("job type UnknownJob is not recognized"))
			})
		})

		Context("when job type is recognized but not enabled", func() {
			BeforeEach(func() {

				reconciler.enabledFrameworks[constants.KindTFJob] = nil
			})

			It("returns an error when job is TFJob but object is nil", func() {
				workload.Spec.JobType = constants.KindTFJob
				err := reconciler.validateJobFramework(workload, testLoggerDw, "test")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("job type TFJob is not enabled"))
			})
			It("returns an error when job is XGBoost but not enabled", func() {
				workload.Spec.JobType = constants.KindXGBoostJob
				err := reconciler.validateJobFramework(workload, testLoggerDw, "test")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("job type XGBoostJob is not enabled"))
			})
		})

		Context("when job type is recognized and enabled", func() {
			BeforeEach(func() {
				workload.Spec.JobType = constants.KindTFJob
			})

			It("returns no error", func() {
				err := reconciler.validateJobFramework(workload, testLoggerDw, "test")
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("GetDistributedJobFrameworksStatus", func() {
		var (
			extensionsGroup extensions.Group
		)

		BeforeEach(func() {
			extensionsGroup = extensions.Group{
				distributed.MPI: extensions.Extension{
					Type: distributed.MPI,
					Status: extensions.Status{
						Enabled: true,
						Capabilities: map[extensions.CapabilityType]bool{
							distributed.Suspension: true,
						},
					},
				},
				distributed.Jax: extensions.Extension{
					Type: distributed.Jax,
					Status: extensions.Status{
						Enabled:      true,
						Capabilities: map[extensions.CapabilityType]bool{},
					},
				},
				distributed.XGBoost: extensions.Extension{
					Type: distributed.XGBoost,
					Status: extensions.Status{
						Enabled: false,
					},
				},
			}
		})

		Context("When all extensions are enabled", func() {
			It("should return the correct DistributedJobFrameworksStatus map", func() {
				statusMap, err := GetDistributedJobFrameworksStatus(extensionsGroup)
				Expect(err).To(BeNil())
				Expect(statusMap).To(HaveKey(constants.KindMpiJob))
				Expect(statusMap).To(HaveKey(constants.KindJAXJob))
				Expect(statusMap).NotTo(HaveKey(constants.KindXGBoostJob))

				// Verify properties of returned statuses
				Expect(statusMap[constants.KindMpiJob].SuspendEnabled).To(BeTrue())
				Expect(statusMap[constants.KindJAXJob].SuspendEnabled).To(BeFalse())
			})
		})

		Context("When an extension does not have a matching job kind", func() {
			BeforeEach(func() {
				// Add an unsupported extension
				extensionsGroup["Unsupported"] = extensions.Extension{
					Type: "Unsupported",
					Status: extensions.Status{
						Enabled: true,
					},
				}
			})

			It("should return an error", func() {
				_, err := GetDistributedJobFrameworksStatus(extensionsGroup)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("the job kind for the extension"))
			})
		})
	})

	Describe("materializeImplicitMasterSpec", func() {
		It("copies the worker pod fields when a master is implicit", func() {
			gpuDevices := 2
			privileged := true
			workload.Spec.Image = field_types.NewOptionalStringField("pytorch:latest", "")
			workload.Spec.Usage = "Submit"
			workload.Spec.GpuDevices = field_types.NewOptionalIntField(&gpuDevices)
			workload.Spec.NodeType = field_types.NewOptionalStringField("gpu-node", "")
			workload.Spec.Privileged = field_types.NewOptionalBoolField(&privileged)

			materializeImplicitMasterSpec(workload)

			Expect(workload.MasterSpec).NotTo(BeNil())
			Expect(workload.MasterSpec.Image.Deref()).To(Equal("pytorch:latest"))
			Expect(workload.MasterSpec.Usage).To(Equal("Submit"))
			Expect(workload.MasterSpec.GpuDevices.Deref()).To(Equal(2))
			Expect(workload.MasterSpec.NodeType.Deref()).To(Equal("gpu-node"))
			Expect(workload.MasterSpec.Privileged.Deref()).To(BeTrue())
			Expect(workload.MasterSpec.Image).NotTo(BeIdenticalTo(workload.Spec.Image))
			Expect(workload.MasterSpec.GpuDevices).NotTo(BeIdenticalTo(workload.Spec.GpuDevices))
			Expect(workload.MasterSpec.NodeType).NotTo(BeIdenticalTo(workload.Spec.NodeType))
			Expect(workload.MasterSpec.Privileged).NotTo(BeIdenticalTo(workload.Spec.Privileged))
		})

		It("preserves an explicit master", func() {
			existingMaster := &dp.DistributedMasterSpec{}
			existingMaster.Usage = "master"
			workload.MasterSpec = existingMaster
			workload.Spec.Usage = "worker"

			materializeImplicitMasterSpec(workload)

			Expect(workload.MasterSpec).To(BeIdenticalTo(existingMaster))
			Expect(workload.MasterSpec.Usage).To(Equal("master"))
		})

		It("does not create a master when noMaster is set", func() {
			noMaster := true
			workload.Spec.NoMaster = field_types.NewOptionalBoolField(&noMaster)

			materializeImplicitMasterSpec(workload)

			Expect(workload.MasterSpec).To(BeNil())
		})
	})
})
