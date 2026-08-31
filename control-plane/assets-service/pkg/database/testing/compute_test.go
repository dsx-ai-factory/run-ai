/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package testing

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/assets-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dao"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/services/scope"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/auth"
	"github.com/run-ai/runai/runai-application-packages/convutils"
)

var _ = Describe("Test Compute Asset dao", func() {
	Context("simple compute gpu requests", func() {
		It("create compute with two gpu requests", func() {
			gpuRequestType := api.GPUREQUESTTYPE_PORTION

			computeCreation := api.ComputeCreationRequest{
				Spec: &api.ComputeAssetSpec{
					GpuRequestType:    &gpuRequestType,
					GpuPortionRequest: convutils.Ref(1.2),
					CpuCoreRequest:    convutils.Ref(1.0),
					CpuCoreLimit:      convutils.Ref(4.0),
					CpuMemoryRequest:  convutils.Ref("3T"),
					CpuMemoryLimit:    convutils.Ref("7P"),
					LargeShmRequest:   convutils.Ref(true),
				},
				Meta: api.AssetCreationRequest{
					Name:  "bla",
					Scope: api.SCOPE_TENANT,
				},
			}

			nameValidationQuery, _ := scope.GetNameValidationQuery(auth.MockAuthInfo(), computeCreation.Meta.Name, computeCreation.Meta.Scope, nil, nil, nil, computeCreation.Meta.Private, api.ASSETKIND_COMPUTE, nil)
			res, _ := dao.CreateComputeAsset(auth.MockAuthInfo(), computeCreation, nameValidationQuery)
			Expect(res.Spec).To(Equal(*computeCreation.Spec))
		})
	})
})
