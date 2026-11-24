/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/clients/org_unit"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/testutils"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
	dbTesting "github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/testing"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/testing"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/authorization"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/checker"
	checkerClusterClient "github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/clients/cluster"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Flat asset resolution", func() {
	router := CreateRouter()
	randoms := testutils.NewRandomValues()

	BeforeEach(func() {
		dbTesting.CleanTables()

		authMock := testing.NewRandomsBasedAuthMockClient(randoms)
		org_unit.SetMock(org_unit.NewSimpleMock(randoms))
		authorization.SetMock(authMock)

		checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
			WithCluster(randoms.ClusterId, testing.BuildDisplayedCluster(randoms, testing.UnitTestDefaultClusterVersion, false)))
		checker.InitClusterCapabilitiesChecker("", DefaultCheckerClusterConfig, "")
	})

	DescribeTable("resolves a multi-device compute asset for flexible submission", func(workloadType api.WorkloadType) {
		computeRequest := buildRandomProjectCompute(randoms)
		computeRequest.Spec = &api.ComputeAssetSpec{
			GpuDevicesRequest: convutils.Ref(int32(3)),
			GpuRequestType:    convutils.Ref(api.GPUREQUESTTYPE_PORTION),
			GpuPortionRequest: convutils.Ref(1.0),
			CpuCoreRequest:    convutils.Ref(0.5),
			CpuCoreLimit:      convutils.Ref(2.0),
			CpuMemoryRequest:  convutils.Ref("20M"),
			CpuMemoryLimit:    convutils.Ref("30M"),
		}
		asset, _, _ := createComputeAsset(computeRequest, router, headerForAdmin(randoms))

		request := api.AssetFlatRequest{
			AssetId: asset.Meta.Id, WorkloadType: workloadType, ClusterId: randoms.ClusterId, ProjectId: randoms.ProjectIdStr,
		}
		req := httptest.NewRequest(http.MethodPost, urlToFlat, makePayload(request))
		req.Header = *headerForAdmin(randoms)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		Expect(resp.Code).To(Equal(http.StatusOK), resp.Body.String())

		var flat api.AssetFlatResponse
		Expect(json.Unmarshal(resp.Body.Bytes(), &flat)).To(Succeed())
		Expect(flat.Worker.Compute).NotTo(BeNil())
		Expect(flat.Worker.Compute.GpuDevicesRequest).To(HaveValue(Equal(int32(3))))
		Expect(flat.Worker.Compute.CpuCoreRequest).To(HaveValue(Equal(0.5)))
		if workloadType == api.WORKLOADTYPE_DISTRIBUTED {
			Expect(flat.Master).NotTo(BeNil())
			Expect(flat.Master.Compute).NotTo(BeNil())
			Expect(flat.Master.Compute.GpuDevicesRequest).To(HaveValue(Equal(int32(3))))
		}
	},
		Entry("workspace", api.WORKLOADTYPE_WORKSPACE),
		Entry("training", api.WORKLOADTYPE_TRAINING),
		Entry("distributed training", api.WORKLOADTYPE_DISTRIBUTED),
	)
})
