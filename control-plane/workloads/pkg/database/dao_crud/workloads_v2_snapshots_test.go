/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao_crud

import (
	"errors"

	authorization "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/models_crud"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	snapshotUpsertUpdate = "UPDATE workloads_v2_snapshots SET created_at=?,created_by=?,updated_at=?,updated_by=?,tenant_id=?,cluster_id=?,department_id=?,project_id=?,workload_type=?,requested_name=?,name=?,namespace=?,spec=?,master_spec=?,master_same_as_worker=?,assets_mapping=? WHERE id = ?"
	snapshotGetSelect    = "SELECT * FROM workloads_v2_snapshots WHERE tenant_id = ? AND id = ? AND workload_type = ? ORDER BY workloads_v2_snapshots.id LIMIT ?"
	snapshotGetScoped    = "SELECT * FROM workloads_v2_snapshots WHERE tenant_id = ? AND id = ? AND workload_type = ? AND ((false or project_id IN ( 'p1'))) ORDER BY workloads_v2_snapshots.id LIMIT ?"
	snapshotDeleteOne    = "DELETE FROM workloads_v2_snapshots WHERE tenant_id = ? AND id = ?"
	snapshotDeleteMany   = "DELETE FROM workloads_v2_snapshots WHERE tenant_id = ? AND id IN (?,?)"
)

type snapshotRow struct {
	ID           uuid.UUID
	TenantId     int32
	WorkloadType string
	Name         string
}

var snapshotFixtureId = uuid.MustParse("11111111-1111-1111-1111-111111111111")

func newSnapshotFixture() *models_crud.WorkloadV2Snapshot {
	return &models_crud.WorkloadV2Snapshot{
		ID:           snapshotFixtureId,
		TenantId:     1,
		ClusterId:    "cluster-1",
		DepartmentId: "dep-1",
		ProjectId:    "p1",
		WorkloadType: "workspace",
		Name:         "wl",
		Namespace:    "runai-p1",
	}
}

var _ = Describe("DbUpsertWorkloadV2Snapshot", func() {
	It("upserts the snapshot", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(snapshotUpsertUpdate).Returns(1)
		m.ExpectTransaction().Commit()

		apiErr := DbUpsertWorkloadV2Snapshot(testCtx(1), newSnapshotFixture())

		Expect(apiErr).To(BeNil())
	})

	It("returns an error when the upsert fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(snapshotUpsertUpdate).ReturnsError(errors.New("boom"))
		m.ExpectTransaction().Rollback()

		apiErr := DbUpsertWorkloadV2Snapshot(testCtx(1), newSnapshotFixture())

		Expect(apiErr).ToNot(BeNil())
	})
})

var _ = Describe("DbGetWorkloadV2Snapshot", func() {
	It("returns the snapshot", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(snapshotGetSelect).Returns([]snapshotRow{{ID: snapshotFixtureId, TenantId: 1, WorkloadType: "workspace", Name: "wl"}})

		snapshot, apiErr := DbGetWorkloadV2Snapshot(testCtx(1), 1, "workspace", snapshotFixtureId.String(), nil)

		Expect(apiErr).To(BeNil())
		Expect(snapshot).ToNot(BeNil())
		Expect(snapshot.Name).To(Equal("wl"))
	})

	It("scopes the query by the permitted scopes", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(snapshotGetScoped).Returns([]snapshotRow{{ID: snapshotFixtureId, TenantId: 1, WorkloadType: "workspace", Name: "wl"}})

		scopes := &authorization.PermittedScopes{Projects: []string{"p1"}}
		snapshot, apiErr := DbGetWorkloadV2Snapshot(testCtx(1), 1, "workspace", snapshotFixtureId.String(), scopes)

		Expect(apiErr).To(BeNil())
		Expect(snapshot).ToNot(BeNil())
	})

	It("returns a not found error when no row matches", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(snapshotGetSelect).Returns([]snapshotRow{})

		snapshot, apiErr := DbGetWorkloadV2Snapshot(testCtx(1), 1, "workspace", snapshotFixtureId.String(), nil)

		Expect(snapshot).To(BeNil())
		Expect(apiErr).ToNot(BeNil())
		Expect(apiErr.Code).To(Equal(int32(404)))
	})

	It("returns a service unavailable error when the query fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(snapshotGetSelect).ReturnsError(errors.New("boom"))

		snapshot, apiErr := DbGetWorkloadV2Snapshot(testCtx(1), 1, "workspace", snapshotFixtureId.String(), nil)

		Expect(snapshot).To(BeNil())
		Expect(apiErr).ToNot(BeNil())
		Expect(apiErr.Code).To(Equal(int32(503)))
	})
})

var _ = Describe("DbDeleteWorkloadV2Snapshot", func() {
	It("deletes the snapshot", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(snapshotDeleteOne).Returns(1)
		m.ExpectTransaction().Commit()

		apiErr := DbDeleteWorkloadV2Snapshot(testCtx(1), 1, snapshotFixtureId.String(), nil)

		Expect(apiErr).To(BeNil())
	})

	It("returns an error when the delete fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(snapshotDeleteOne).ReturnsError(errors.New("boom"))
		m.ExpectTransaction().Rollback()

		apiErr := DbDeleteWorkloadV2Snapshot(testCtx(1), 1, snapshotFixtureId.String(), nil)

		Expect(apiErr).ToNot(BeNil())
	})
})

var _ = Describe("DbDeleteWorkloadV2Snapshots", func() {
	It("does nothing when no ids are given", func() {
		setupDaoMock(GinkgoT())

		apiErr := DbDeleteWorkloadV2Snapshots(testCtx(1), 1, nil, nil)

		Expect(apiErr).To(BeNil())
	})

	It("deletes all the given snapshots", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(snapshotDeleteMany).Returns(2)
		m.ExpectTransaction().Commit()

		apiErr := DbDeleteWorkloadV2Snapshots(testCtx(1), 1, []string{"a", "b"}, nil)

		Expect(apiErr).To(BeNil())
	})

	It("returns an error when the delete fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(snapshotDeleteMany).ReturnsError(errors.New("boom"))
		m.ExpectTransaction().Rollback()

		apiErr := DbDeleteWorkloadV2Snapshots(testCtx(1), 1, []string{"a", "b"}, nil)

		Expect(apiErr).ToNot(BeNil())
	})
})
