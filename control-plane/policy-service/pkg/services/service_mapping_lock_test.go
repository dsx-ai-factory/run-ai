/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"fmt"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/clients/cluster"
	orgUnitService "github.com/run-ai/runai/control-plane/policy-service/pkg/clients/org_unit"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/database"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/database/dao"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/database/dbmodel"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/clients/authorization"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/clients/orgunit"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
	"github.com/run-ai/runai/runai-common-packages/runaidb"
)

// Must match dao.mappingLockBusyMessage (package-private).
const expectedMappingLockBusyMessage = "mapping is being updated, please retry your update later"

func mappingLockTestTenantCtx(tenantID int32) context.Context {
	ctx := context.WithValue(context.Background(), runaictx.CfSubjectID, "lock-test-user")
	return context.WithValue(ctx, runaictx.CfTenantId, fmt.Sprintf("%d", tenantID))
}

var _ = Describe("mapping_locking table", func() {
	var backendMock *BackendMockClient
	var clusterMock cluster.SimpleMockClient

	BeforeEach(func() {
		database.DeleteAllTablesForTesting()

		backendMock = NewBackendMockClient(2, 2)
		orgUnitService.SetMock(NewPoliciesOrgUnitClient(backendMock))
		orgunit.SetMock(NewPoliciesOrgUnitClient(backendMock))
		authorization.SetMock(NewPoliciesAuthClient(backendMock))

		clusterMock = cluster.NewSimpleMockClient().WithClusters(backendMock.ClusterIdPerTenant)
		cluster.SetMock(clusterMock)
	})

	It("rejects a second lock for the same tenant until the first is released, then allows lock again", func() {
		ctx := mappingLockTestTenantCtx(1)

		release1, err1 := dao.AcquireTenantMappingLock(ctx)
		Expect(err1).To(BeNil())
		Expect(release1).ToNot(BeNil())

		_, err2 := dao.AcquireTenantMappingLock(ctx)
		Expect(err2).ToNot(BeNil())
		Expect(err2.Code).To(Equal(int32(http.StatusConflict)))
		Expect(err2.Message).To(Equal(expectedMappingLockBusyMessage))

		release1()

		release3, err3 := dao.AcquireTenantMappingLock(ctx)
		Expect(err3).To(BeNil())
		Expect(release3).ToNot(BeNil())
		release3()
	})

	It("does not block a different tenant from taking a lock", func() {
		ctxA := mappingLockTestTenantCtx(1)
		ctxB := mappingLockTestTenantCtx(2)

		releaseA, errA := dao.AcquireTenantMappingLock(ctxA)
		Expect(errA).To(BeNil())

		releaseB, errB := dao.AcquireTenantMappingLock(ctxB)
		Expect(errB).To(BeNil())

		releaseA()
		releaseB()
	})

	It("removes a stale lock older than the grace window so acquire and release succeed", func() {
		ctx := mappingLockTestTenantCtx(7)
		staleLockedAt := time.Now().UTC().Add(-10 * time.Second)

		Expect(runaidb.DB().Create(&dbmodel.MappingLocking{
			TenantId: 7,
			LockedAt: staleLockedAt,
		})).Error().ToNot(HaveOccurred())

		release, err := dao.AcquireTenantMappingLock(ctx)
		Expect(err).To(BeNil())
		Expect(release).ToNot(BeNil())
		release()
	})
})
