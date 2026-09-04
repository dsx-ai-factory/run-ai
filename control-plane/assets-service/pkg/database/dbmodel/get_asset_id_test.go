/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dbmodel

import (
	"testing"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/testutils"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestDmodel(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "dbmodel suite")
}

var _ = Describe("workload assets validation", func() {
	Context("getAssetId -> environment", func() {
		It("should return the correct asset id", func() {
			rec := EnvironmentAssetRec{AssetId: testutils.RandString(5)}
			Expect(rec.AssetId).To(Equal(rec.GetAssetId()))
		})
	})

	Context("getAssetId -> compute", func() {
		It("should return the correct asset id", func() {
			rec := ComputeAssetRec{AssetId: testutils.RandString(5)}
			Expect(rec.AssetId).To(Equal(rec.GetAssetId()))
		})
	})

	Context("getAssetId -> host path", func() {
		It("should return the correct asset id", func() {
			rec := HostPathAssetRec{AssetId: testutils.RandString(5)}
			Expect(rec.AssetId).To(Equal(rec.GetAssetId()))
		})
	})

	Context("getAssetId -> Nfs", func() {
		It("should return the correct asset id", func() {
			rec := NFSAssetRec{AssetId: testutils.RandString(5)}
			Expect(rec.AssetId).To(Equal(rec.GetAssetId()))
		})
	})

	Context("getAssetId -> git", func() {
		It("should return the correct asset id", func() {
			rec := GitAssetRec{AssetId: testutils.RandString(5)}
			Expect(rec.AssetId).To(Equal(rec.GetAssetId()))
		})
	})

	Context("getAssetId -> compute", func() {
		It("should return the correct asset id", func() {
			rec := S3AssetRec{AssetId: testutils.RandString(5)}
			Expect(rec.AssetId).To(Equal(rec.GetAssetId()))
		})
	})
})
