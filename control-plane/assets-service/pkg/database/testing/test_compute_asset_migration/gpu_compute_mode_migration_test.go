/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package test_compute_asset_migration

import (
	"context"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
	dbTesting "github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/testing"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("GPU compute mode migration", func() {
	It("preserves compute assets through upgrade, replay and rollback", func() {
		dbTesting.CleanTables()
		ctx := context.Background()
		Expect(injectComputeAssets(ctx, uuid_compute_assets_a, "selected mode", string(api.ASSETKIND_COMPUTE), "system")).To(Succeed())
		Expect(injectComputeAssets(ctx, uuid_compute_assets_b, "implicit mode", string(api.ASSETKIND_COMPUTE), "system")).To(Succeed())
		db := runaidb.DB().Actual().WithContext(ctx)

		By("upgrading existing assets without inventing a requested mode")
		migrateTo(30010005, "down 30010001")
		Expect(db.Migrator().HasColumn("assets_compute", "gpu_compute_mode")).To(BeFalse())
		migrateTo(30010001, "up 30010005")
		var unsetModes int64
		Expect(db.Raw("SELECT COUNT(*) FROM assets_compute WHERE gpu_compute_mode IS NULL").Scan(&unsetModes).Error).To(Succeed())
		Expect(unsetModes).To(Equal(int64(2)))

		By("preserving a saved mode when an older migration version is replayed")
		Expect(db.Exec("UPDATE assets_compute SET gpu_compute_mode = ? WHERE asset_id = ?", api.GPUCOMPUTEMODE_SM_SHARING, uuid_compute_assets_a).Error).To(Succeed())
		migrateTo(30010001, "up 30010005")
		var savedMode string
		Expect(db.Raw("SELECT gpu_compute_mode FROM assets_compute WHERE asset_id = ?", uuid_compute_assets_a).Scan(&savedMode).Error).To(Succeed())
		Expect(savedMode).To(Equal(string(api.GPUCOMPUTEMODE_SM_SHARING)))
		Expect(db.Raw("SELECT COUNT(*) FROM assets_compute WHERE gpu_compute_mode IS NULL").Scan(&unsetModes).Error).To(Succeed())
		Expect(unsetModes).To(Equal(int64(1)))

		By("allowing rollback to be replayed without deleting the assets")
		migrateTo(30010005, "down 30010001")
		migrateTo(30010005, "down 30010001")
		Expect(db.Migrator().HasColumn("assets_compute", "gpu_compute_mode")).To(BeFalse())
		var assetCount int64
		Expect(db.Raw("SELECT COUNT(*) FROM assets_compute").Scan(&assetCount).Error).To(Succeed())
		Expect(assetCount).To(Equal(int64(2)))

		By("reapplying the nullable column after rollback")
		migrateTo(30010001, "up 30010005")
		Expect(db.Raw("SELECT COUNT(*) FROM assets_compute WHERE gpu_compute_mode IS NULL").Scan(&unsetModes).Error).To(Succeed())
		Expect(unsetModes).To(Equal(int64(2)))
	})
})
