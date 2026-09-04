/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package scope

import (
	"net/http"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Test scope validation", func() {
	Context("parameters validation", func() {
		It("should return sensible errors for missing parameters", func() {
			err := ValidateScopeParameters(&api.AssetCreationRequest{
				Scope: api.SCOPE_SYSTEM,
			})
			Expect(err).To(BeNil())

			err = ValidateScopeParameters(&api.AssetCreationRequest{
				Scope: api.SCOPE_TENANT,
			})
			Expect(err).To(BeNil())

			err = ValidateScopeParameters(&api.AssetCreationRequest{
				Scope: api.SCOPE_CLUSTER,
			})
			Expect(err).NotTo(BeNil())
			Expect(err.Code).To(Equal(int32(http.StatusBadRequest)))
			Expect(err.Message).To(ContainSubstring("clusterId is mandatory for cluster scoped assets"))

			err = ValidateScopeParameters(&api.AssetCreationRequest{
				Scope: api.SCOPE_DEPARTMENT,
			})
			Expect(err).NotTo(BeNil())
			Expect(err.Code).To(Equal(int32(http.StatusBadRequest)))
			Expect(err.Message).To(ContainSubstring("departmentId is mandatory for department scoped assets"))

			err = ValidateScopeParameters(&api.AssetCreationRequest{
				Scope: api.SCOPE_PROJECT,
			})
			Expect(err).NotTo(BeNil())
			Expect(err.Code).To(Equal(int32(http.StatusBadRequest)))
			Expect(err.Message).To(ContainSubstring("projectId is mandatory for project scoped assets"))
		})

		It("should return sensible errors for conflicting parameters", func() {
			err := ValidateScopeParameters(&api.AssetCreationRequest{
				Scope:     api.SCOPE_SYSTEM,
				ProjectId: convutils.Ref(int32(1)),
			})
			Expect(err).NotTo(BeNil())
			Expect(err.Code).To(Equal(int32(http.StatusBadRequest)))
			Expect(err.Message).To(ContainSubstring("projectId should not be provided for system scoped assets"))

			err = ValidateScopeParameters(&api.AssetCreationRequest{
				Scope:     api.SCOPE_TENANT,
				ClusterId: convutils.Ref("aaa"),
			})
			Expect(err).NotTo(BeNil())
			Expect(err.Code).To(Equal(int32(http.StatusBadRequest)))
			Expect(err.Message).To(ContainSubstring("clusterId should not be provided for tenant scoped assets"))

			err = ValidateScopeParameters(&api.AssetCreationRequest{
				Scope:        api.SCOPE_CLUSTER,
				ClusterId:    convutils.Ref("aaa"),
				DepartmentId: convutils.Ref("aaa"),
			})
			Expect(err).NotTo(BeNil())
			Expect(err.Code).To(Equal(int32(http.StatusBadRequest)))
			Expect(err.Message).To(ContainSubstring("departmentId should not be provided for cluster scoped assets"))

			err = ValidateScopeParameters(&api.AssetCreationRequest{
				Scope:        api.SCOPE_DEPARTMENT,
				ProjectId:    convutils.Ref(int32(1)),
				DepartmentId: convutils.Ref("aaa"),
			})
			Expect(err).NotTo(BeNil())
			Expect(err.Code).To(Equal(int32(http.StatusBadRequest)))
			Expect(err.Message).To(ContainSubstring("projectId should not be provided for department scoped assets"))

			err = ValidateScopeParameters(&api.AssetCreationRequest{
				Scope:        api.SCOPE_PROJECT,
				ProjectId:    convutils.Ref(int32(1)),
				ClusterId:    convutils.Ref("aaa"), // optional, can be provided
				DepartmentId: convutils.Ref("aaa"),
			})
			Expect(err).NotTo(BeNil())
			Expect(err.Code).To(Equal(int32(http.StatusBadRequest)))
			Expect(err.Message).To(ContainSubstring("departmentId should not be provided for project scoped assets"))
		})
	})
})
