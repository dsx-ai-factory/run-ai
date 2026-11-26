/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
	k8sObjectsTracker "github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/clients/k8s-objects-tracker"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/dbmodel"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	k8sObjectsApi "github.com/dsx-ai-factory/run-ai/runai-clients/k8s-objects-tracker-client/generated"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("validateChangeOfPvcSize error messages", func() {
	const claimName = "resize-claim"

	resizeRequest := func(size string) *api.PvcUpdateRequest {
		return &api.PvcUpdateRequest{Spec: &api.PvcFieldsUpdatable{ClaimInfo: &api.ClaimInfoUpdatable{Size: convutils.Ref(size)}}}
	}

	storedSpec := func(size string) *dbmodel.PVCAssetRec {
		return &dbmodel.PVCAssetRec{PvcAssetSpec: api.PvcAssetSpec{
			ExistingPvc: convutils.Ref(false),
			ClaimName:   convutils.Ref(claimName),
			ClaimInfo:   &api.ClaimInfo{Size: convutils.Ref(size)},
		}}
	}

	AfterEach(func() {
		k8sObjectsTracker.SetMock(k8sObjectsTracker.NewSimpleMockClient())
	})

	It("reports the originally requested size when shrinking below it", func() {
		meta := &dbmodel.AssetMeta{}

		err := validateChangeOfPvcSize(context.Background(), "id", meta, storedSpec("10Gi"), resizeRequest("5Gi"))

		Expect(err).NotTo(BeNil())
		Expect(err.Message).To(Equal("requested size (5Gi) must be no less than the originally requested size (10Gi) of the PVC"))
	})

	It("reports the requested size first and the effective capacity second", func() {
		clusterId := "cluster-1"
		meta := &dbmodel.AssetMeta{Scope: api.SCOPE_CLUSTER, ClusterId: &clusterId}

		pvc := k8sObjectsApi.PVCListInfo{ScopeType: string(api.SCOPE_CLUSTER), ScopeId: clusterId, Name: claimName}
		pvc.SetPhase(k8sObjectsApi.PVCPHASEOPTIONAL_BOUND)
		pvc.SetRequestedCapacity("20Gi")
		mock := k8sObjectsTracker.NewSimpleMockClient()
		mock.PvcResponse.Pvcs = []k8sObjectsApi.PVCListInfo{pvc}
		k8sObjectsTracker.SetMock(mock)

		err := validateChangeOfPvcSize(context.Background(), "id", meta, storedSpec("10Gi"), resizeRequest("15Gi"))

		Expect(err).NotTo(BeNil())
		Expect(err.Message).To(Equal("requested size (15Gi) must be no less than the effective requested capacity (20Gi) of the PVC"))
	})
})
