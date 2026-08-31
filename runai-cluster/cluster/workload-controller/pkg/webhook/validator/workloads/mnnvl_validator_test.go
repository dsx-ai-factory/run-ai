/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workloads

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/mnnvl"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

var _ = Describe("MnnvlValidator", func() {
	var (
		validator *MnnvlValidator
		workload  *jobs.JobWorkload
	)
	ctx := context.Background()

	BeforeEach(func() {
		workload = &jobs.JobWorkload{
			TypeMeta: metav1.TypeMeta{
				Kind: constants.KindDistributedWorkload,
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:   "w1",
				Labels: make(map[string]string),
			},
		}

		validator = NewMnnvlValidator(false, false)
	})

	Context("when MNNVL is not supported by the cluster", func() {
		It("not distributed", func() {
			workload.Kind = constants.KindTrainingWorkload
			err := validator.Validate(ctx, workload)
			Expect(err).NotTo(HaveOccurred())
		})

		It("has not feature label", func() {
			err := validator.Validate(ctx, workload)
			Expect(err).NotTo(HaveOccurred())
		})

		It("has None label", func() {
			workload.Labels[mnnvl.FeatureLabelKey] = mnnvl.None
			err := validator.Validate(ctx, workload)
			Expect(err).NotTo(HaveOccurred())
		})

		DescribeTable("should deny if feature label is not None",
			func(value string) {
				workload.Labels[mnnvl.FeatureLabelKey] = value
				err := validator.Validate(ctx, workload)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("is not enabled in the cluster"))
			},
			Entry("Required", mnnvl.Required),
			Entry("Preferred", mnnvl.Preferred),
		)
	})

	Context("when MNNVL is supported in the cluster and compute domain is static", func() {
		BeforeEach(func() {
			validator = NewMnnvlValidator(true, false)
		})

		Context("with Required or Preferred feature label", func() {
			BeforeEach(func() {
				// Add GPU resources to make tests pass validation
				workload.Spec = jobs.JobSpec{
					ResourceAllocationFields: field_groups.ResourceAllocationFields{
						GpuDevices: field_types.NewOptionalIntField(ptr.To(1)),
					},
					RunaiSchedulerFields: field_groups.RunaiSchedulerFields{
						// Add node pools to pass node pool validation
						NodePools: field_types.NewOptionalStringField("some-node-pool", ""),
					},
				}

				// Also set up master spec with node pools
				if workload.MasterSpec == nil {
					workload.MasterSpec = &jobs.JobSpec{}
				}
				workload.MasterSpec.RunaiSchedulerFields = field_groups.RunaiSchedulerFields{
					NodePools: field_types.NewOptionalStringField("master-node-pool", ""),
				}
			})

			DescribeTable("validate regardless of label value with compute domain static",
				func(value string) {
					workload.Labels[mnnvl.FeatureLabelKey] = value

					// Should pass with NVIDIA GPU and non-elastic job with GPU resources
					err := validator.Validate(ctx, workload)
					Expect(err).NotTo(HaveOccurred())

					// Should fail with elastic PyTorch job
					elasticJob := getMnnvlRequiredPytorchJob(ptr.To(1), ptr.To(3))
					elasticJob.Labels[mnnvl.FeatureLabelKey] = value

					err = validator.Validate(ctx, elasticJob)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("elastic PyTorchJob"))

					// Should fail with no GPU resources
					noGpuWorkload := &jobs.JobWorkload{
						TypeMeta: metav1.TypeMeta{
							Kind: constants.KindDistributedWorkload,
						},
						ObjectMeta: metav1.ObjectMeta{
							Name:   "no-gpu",
							Labels: map[string]string{mnnvl.FeatureLabelKey: value},
						},
					}
					err = validator.Validate(ctx, noGpuWorkload)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("requested resources"))
				},
				Entry("Required", mnnvl.Required),
				Entry("Preferred", mnnvl.Preferred),
			)
		})

		Context("with None feature label or no label", func() {
			It("should pass validation with None label", func() {
				workload.Labels[mnnvl.FeatureLabelKey] = mnnvl.None
				err := validator.Validate(ctx, workload)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should pass validation with no label", func() {
				err := validator.Validate(ctx, workload)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		DescribeTable("with invalid feature label value",
			func(value string) {
				workload.Labels[mnnvl.FeatureLabelKey] = value
				err := validator.Validate(ctx, workload)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("invalid value for the label"))
			},
			Entry("value is not a mode", "blabla"),
			Entry("empty value", ""),
		)
	})

	Context("when MNNVL is supported in the cluster and compute domain is elastic", func() {
		BeforeEach(func() {
			validator = NewMnnvlValidator(true, true)
		})

		Context("with Required or Preferred feature label", func() {
			BeforeEach(func() {
				// Add GPU resources to make tests pass validation
				workload.Spec = jobs.JobSpec{
					ResourceAllocationFields: field_groups.ResourceAllocationFields{
						GpuDevices: field_types.NewOptionalIntField(ptr.To(1)),
					},
					RunaiSchedulerFields: field_groups.RunaiSchedulerFields{
						// Add node pools to pass node pool validation
						NodePools: field_types.NewOptionalStringField("some-node-pool", ""),
					},
				}

				// Also set up master spec with node pools
				if workload.MasterSpec == nil {
					workload.MasterSpec = &jobs.JobSpec{}
				}
				workload.MasterSpec.RunaiSchedulerFields = field_groups.RunaiSchedulerFields{
					NodePools: field_types.NewOptionalStringField("master-node-pool", ""),
				}
			})

			DescribeTable("validate regardless of label value with compute domain static",
				func(value string) {
					workload.Labels[mnnvl.FeatureLabelKey] = value

					// Should pass with NVIDIA GPU and non-elastic job with GPU resources
					err := validator.Validate(ctx, workload)
					Expect(err).NotTo(HaveOccurred())

					// Should pass with elastic PyTorch job
					elasticJob := getMnnvlRequiredPytorchJob(ptr.To(1), ptr.To(3))
					elasticJob.Labels[mnnvl.FeatureLabelKey] = value

					err = validator.Validate(ctx, elasticJob)
					Expect(err).ToNot(HaveOccurred())

					// Should fail with no GPU resources
					noGpuWorkload := &jobs.JobWorkload{
						TypeMeta: metav1.TypeMeta{
							Kind: constants.KindDistributedWorkload,
						},
						ObjectMeta: metav1.ObjectMeta{
							Name:   "no-gpu",
							Labels: map[string]string{mnnvl.FeatureLabelKey: value},
						},
					}
					err = validator.Validate(ctx, noGpuWorkload)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("requested resources"))
				},
				Entry("Required", mnnvl.Required),
				Entry("Preferred", mnnvl.Preferred),
			)
		})

		Context("with None feature label or no label", func() {
			It("should pass validation with None label", func() {
				workload.Labels[mnnvl.FeatureLabelKey] = mnnvl.None
				err := validator.Validate(ctx, workload)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should pass validation with no label", func() {
				err := validator.Validate(ctx, workload)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		DescribeTable("with invalid feature label value",
			func(value string) {
				workload.Labels[mnnvl.FeatureLabelKey] = value
				err := validator.Validate(ctx, workload)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("invalid value for the label"))
			},
			Entry("value is not a mode", "blabla"),
			Entry("empty value", ""),
		)
	})
})

func getMnnvlRequiredPytorchJob(minReplicas, maxReplicas *int) *jobs.JobWorkload {
	job := resource.CreateElasticPyTorchJob(minReplicas, maxReplicas)
	job.Labels = map[string]string{
		mnnvl.FeatureLabelKey: mnnvl.Required,
	}

	return job
}
