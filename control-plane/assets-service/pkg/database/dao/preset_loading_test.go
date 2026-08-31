/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"context"
	"math/rand"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-common-packages/runaidb"
)

var _ = Describe("presets cloning tests", func() {
	Context("tenant_settings", func() {
		// It("prevents multiple loading", func() {
		// 	db := runaidb.DB().Actual().WithContext(ctx)
		// 	tenantId := rand.Int31()

		// 	// Revision stored in tenant settings prevents multiple load of presets
		// 	logs := logtest.NewObservedLogsOfLevel(zap.DebugLevel)
		// 	result := setTenantPresetsRevision(db, tenantId, CurrentPresetsRevision)
		// 	Expect(result.Error).To(BeNil())
		// 	clonePresetAssetsToTenant(tenantId)
		// 	Expect(len(logs.All())).NotTo(Equal(int32(1)))
		// 	Expect(logs.TakeAll()[0].Message).To(ContainSubstring("already loaded presets"))
		// })
	})

	Context("no tenant settings", func() {
		It("should return zero revision", func() {
			db := runaidb.DB().WithContext(context.Background())
			tenantId := rand.Int31()

			revision, err := getTenantPresetsRevision(context.Background(), db, tenantId)
			Expect(err).To(BeNil())
			Expect(revision).To(Equal(int32(0)))
		})
	})

	Context("invalid tenant settings", func() {
		// It("should log error and return zero revision", func() {
		// 	db := runaidb.DB().Actual().WithContext(ctx)
		// 	tenantId := rand.Int31()

		// 	result := TenantSettingsSet(db, dbmodel.TenantSettingsRec{
		// 		TenantId: tenantId,
		// 		Key:      dbmodel.KeyPresetsInstalled,
		// 		Value:    "abc",
		// 	})
		// 	Expect(result.Error).To(BeNil())

		// 	logs := logtest.NewObservedLogsOfLevel(zap.DebugLevel)

		// 	revision, result := getTenantPresetsRevision(db, tenantId)
		// 	Expect(result.Error).To(BeNil())
		// 	Expect(revision).To(Equal(int32(0)))

		// 	Expect(len(logs.All())).NotTo(Equal(int32(1)))
		// 	Expect(logs.TakeAll()[0].Message).To(ContainSubstring("invalid presets revision 'abc'"))
		// })
	})
})
