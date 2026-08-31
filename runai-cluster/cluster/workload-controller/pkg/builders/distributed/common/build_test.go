/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package common

import (
	v1 "github.com/kubeflow/training-operator/pkg/apis/kubeflow.org/v1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	v2alpha12 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedpolicy/v2alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedworkload/v2alpha1"
	"k8s.io/utils/ptr"
)

var _ = Context("Distributed Common Builder", func() {
	Describe("RunPolicy", func() {
		builder := DistributedBuilder{}
		BeforeEach(func() {

			builder.MergedWorkload = &v2alpha1.DistributedWorkload{
				Spec: v2alpha12.DistributedSpec{
					WorkloadFields:           field_groups.WorkloadFields{},
					ResourceAllocationFields: field_groups.ResourceAllocationFields{},
					RunaiSchedulerFields:     field_groups.RunaiSchedulerFields{},
					AccessControlFields:      field_groups.AccessControlFields{},
					DistributedFields: field_groups.DistributedFields{
						JobType: "",
						Workers: nil,
						RunPolicy: &field_types.RunPolicyField{Values: &field_types.RunPolicyMembers{
							BackoffLimit:   nil,
							CleanPodPolicy: nil,
						}},
						NonPreemptible: nil,
						MpiJob:         nil,
						PyTorchJob:     nil,
					},
				},
			}
		})

		It("should be able to build RunPolicy", func() {
			var kubeRunPolicy v1.RunPolicy
			Expect(builder.BuildRunPolicy(&kubeRunPolicy)).To(Succeed())
		})

		Context("RunPolicy - suspend", func() {
			It("Active = true", func() {
				builder.MergedWorkload.Spec.StateFields = field_groups.StateFields{Active: field_types.NewOptionalBoolField(ptr.To(true))}

				var kubeRunPolicy v1.RunPolicy
				Expect(builder.BuildRunPolicy(&kubeRunPolicy)).To(Succeed())
				Expect(kubeRunPolicy.Suspend).To(BeNil())
			})
			It("Active = false", func() {
				builder.MergedWorkload.Spec.StateFields = field_groups.StateFields{Active: field_types.NewOptionalBoolField(ptr.To(false))}

				var kubeRunPolicy v1.RunPolicy
				Expect(builder.BuildRunPolicy(&kubeRunPolicy)).To(Succeed())
				Expect(*kubeRunPolicy.Suspend).To(BeTrue())
			})
		})

		DescribeTable("default CleanPodPolicy",
			func(jobType string, expectedPolicy v1.CleanPodPolicy) {
				builder.MergedWorkload.Spec.JobType = jobType

				var kubeRunPolicy v1.RunPolicy
				Expect(builder.BuildRunPolicy(&kubeRunPolicy)).To(Succeed())
				Expect(*kubeRunPolicy.CleanPodPolicy).To(Equal(expectedPolicy))
			},
			Entry("TF", constants.KindTFJob, v1.CleanPodPolicyRunning),
			Entry("MPI", constants.KindMpiJob, v1.CleanPodPolicyRunning),
			Entry("PyTorch", constants.KindPyTorchJob, v1.CleanPodPolicyNone),
			Entry("XGBoost", constants.KindXGBoostJob, v1.CleanPodPolicyNone),
		)
	})

})
