/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package jax

import (
	"errors"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	dp "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/distributedpolicy/v2alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/distributedworkload/v2alpha1"
	"k8s.io/utils/ptr"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ValidateWorkload", func() {
	var workload *v2alpha1.DistributedWorkload

	BeforeEach(func() {
		// Initialize a valid workload
		workload = &v2alpha1.DistributedWorkload{
			Spec: dp.DistributedSpec{
				DistributedFields: field_groups.DistributedFields{
					NoMaster: field_types.NewOptionalBoolField(ptr.To(true)),
				},
			},
		}
	})

	Context("validate master configuration", func() {
		It("should return MasterNotSupportedError if NoMaster is false", func() {
			workload.Spec.DistributedFields.NoMaster.Value = ptr.To(false)
			err := ValidateWorkload(workload)
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(MasterNotSupportedError))
		})

		It("should return MasterNotSupportedError if MasterSpec is not nil", func() {
			workload.MasterSpec = &dp.DistributedMasterSpec{}
			err := ValidateWorkload(workload)
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(MasterNotSupportedError))
		})
	})

	Context("validate unsupported fields", func() {
		It("should return an error if MpiJob SlotsPerWorker is not nil", func() {
			workload.Spec.MpiJob = &field_types.MpiField{
				Values: &field_types.MpiMembers{
					SlotsPerWorker: field_types.NewOptionalIntField(ptr.To(1)),
				},
			}
			err := ValidateWorkload(workload)
			Expect(err).To(HaveOccurred())
			Expect(errors.Is(err, FieldsNotSupportedError)).To(BeTrue())
			Expect(err.Error()).To(ContainSubstring("invalid workload fields"))
		})

		It("should return an error if MpiJob SshAuthMountPath is not nil", func() {
			workload.Spec.MpiJob = &field_types.MpiField{
				Values: &field_types.MpiMembers{
					SshAuthMountPath: field_types.NewOptionalStringField("/home", ""),
				},
			}
			err := ValidateWorkload(workload)
			Expect(err).To(HaveOccurred())
			Expect(errors.Is(err, FieldsNotSupportedError)).To(BeTrue())
			Expect(err.Error()).To(ContainSubstring("invalid workload fields"))
		})

		Context("PyTorchJob is not nil", func() {
			It("should return false when PyTorchJob min is not nil", func() {
				workload.Spec.PyTorchJob = &field_types.PyTorchField{
					Values: &field_types.PyTorchMembers{
						MinReplicas: field_types.NewOptionalIntField(ptr.To(1)),
					},
				}
				err := ValidateWorkload(workload)
				Expect(err).To(HaveOccurred())
				Expect(errors.Is(err, FieldsNotSupportedError)).To(BeTrue())
				Expect(err.Error()).To(ContainSubstring("invalid workload fields"))
			})
			It("should return false when PyTorchJob max is not nil", func() {
				workload.Spec.PyTorchJob = &field_types.PyTorchField{
					Values: &field_types.PyTorchMembers{
						MaxReplicas: field_types.NewOptionalIntField(ptr.To(1)),
					},
				}
				err := ValidateWorkload(workload)
				Expect(err).To(HaveOccurred())
			})
			It("should return true when PyTorchJob is not initialized", func() {
				workload.Spec.PyTorchJob = &field_types.PyTorchField{
					Values: &field_types.PyTorchMembers{},
				}
				err := ValidateWorkload(workload)
				Expect(err).ToNot(HaveOccurred())
			})
		})

	})

	Context("when everything is valid", func() {
		It("should not return any error", func() {
			workload = &v2alpha1.DistributedWorkload{
				Spec: dp.DistributedSpec{
					DistributedFields: field_groups.DistributedFields{
						NoMaster: field_types.NewOptionalBoolField(ptr.To(true)),
					},
				},
			}
			err := ValidateWorkload(workload)
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
