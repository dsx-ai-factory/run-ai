/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package builders

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	flds "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	v1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"
)

var _ = Describe("RestartPolicy tests", func() {
	var (
		restartPolicyField *flds.StringField
	)

	BeforeEach(func() {
		// Setup common test variables
		restartPolicyField = &flds.StringField{}
	})

	Context("GetRestartPolicy function", func() {
		It("should return nil when input is nil", func() {
			result, err := GetRestartPolicy(nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(BeNil())
		})

		It("should return nil when value is nil", func() {
			restartPolicyField.Value = nil
			result, err := GetRestartPolicy(restartPolicyField)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(BeNil())
		})

		It("should return OnFailure restart policy", func() {
			restartPolicyField.Value = ptr.To(string(v1.RestartPolicyOnFailure))
			result, err := GetRestartPolicy(restartPolicyField)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeNil())
			Expect(*result).To(Equal(v1.RestartPolicyOnFailure))
		})

		It("should return Always restart policy", func() {
			restartPolicyField.Value = ptr.To(string(v1.RestartPolicyAlways))
			result, err := GetRestartPolicy(restartPolicyField)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeNil())
			Expect(*result).To(Equal(v1.RestartPolicyAlways))
		})

		It("should return Never restart policy", func() {
			restartPolicyField.Value = ptr.To(string(v1.RestartPolicyNever))
			result, err := GetRestartPolicy(restartPolicyField)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeNil())
			Expect(*result).To(Equal(v1.RestartPolicyNever))
		})

		It("should return error for invalid restart policy", func() {
			invalidPolicy := "InvalidRestartPolicy"
			restartPolicyField.Value = ptr.To(invalidPolicy)
			result, err := GetRestartPolicy(restartPolicyField)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid restartPolicy value"))
			Expect(result).To(BeNil())
		})
	})
})
