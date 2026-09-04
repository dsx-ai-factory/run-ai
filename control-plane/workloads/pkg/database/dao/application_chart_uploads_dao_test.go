/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/models"
	dbtesting "github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/testing"
	"gorm.io/gorm"
)

const (
	chartUploadInsert = "INSERT INTO application_chart_uploads (id,tenant_id,chart_bytes,created_at) VALUES (?,?,?,?)"
	chartUploadSelect = "SELECT * FROM application_chart_uploads WHERE application_chart_uploads.id = ? AND application_chart_uploads.tenant_id = ? ORDER BY application_chart_uploads.id LIMIT ?"
)

var chartUploadID = uuid.MustParse("33333333-3333-3333-3333-333333333333")

type chartUploadRow struct {
	ID         uuid.UUID
	TenantId   string
	ChartBytes []byte
}

var _ = Describe("CreateApplicationChartUpload", func() {
	It("inserts the chart upload and commits", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(chartUploadInsert, chartUploadID, "1", []byte("chart"), dbtesting.AnyTime()).Returns(1)
		m.ExpectTransaction().Commit()

		apiErr := CreateApplicationChartUpload(testCtx(1), &models.ApplicationChartUpload{
			ID:         chartUploadID,
			TenantId:   "1",
			ChartBytes: []byte("chart"),
		}, "chart.tgz")

		Expect(apiErr).To(BeNil())
	})

	It("returns service unavailable when the insert fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(chartUploadInsert, chartUploadID, "1", []byte("chart"), dbtesting.AnyTime()).ReturnsError(errors.New("boom"))
		m.ExpectTransaction().Rollback()

		apiErr := CreateApplicationChartUpload(testCtx(1), &models.ApplicationChartUpload{
			ID:         chartUploadID,
			TenantId:   "1",
			ChartBytes: []byte("chart"),
		}, "chart.tgz")

		Expect(apiErr).ToNot(BeNil())
		Expect(apiErr.Code).To(BeEquivalentTo(http.StatusServiceUnavailable))
	})
})

var _ = Describe("GetApplicationChartUploadByID", func() {
	It("returns the stored chart upload", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(chartUploadSelect, chartUploadID, "1", 1).
			Returns([]chartUploadRow{{ID: chartUploadID, TenantId: "1", ChartBytes: []byte("chart")}})

		upload, apiErr := GetApplicationChartUploadByID(testCtx(1), chartUploadID, "1")

		Expect(apiErr).To(BeNil())
		Expect(upload).ToNot(BeNil())
		Expect(upload.ID).To(Equal(chartUploadID))
		Expect(upload.ChartBytes).To(Equal([]byte("chart")))
	})

	It("returns not found when no rows are returned", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(chartUploadSelect, chartUploadID, "1", 1).Returns([]chartUploadRow{})

		upload, apiErr := GetApplicationChartUploadByID(testCtx(1), chartUploadID, "1")

		Expect(upload).To(BeNil())
		Expect(apiErr).ToNot(BeNil())
		Expect(apiErr.Code).To(BeEquivalentTo(http.StatusNotFound))
	})

	It("returns not found when the query reports a record not found error", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(chartUploadSelect, chartUploadID, "1", 1).ReturnsError(gorm.ErrRecordNotFound)

		upload, apiErr := GetApplicationChartUploadByID(testCtx(1), chartUploadID, "1")

		Expect(upload).To(BeNil())
		Expect(apiErr).ToNot(BeNil())
		Expect(apiErr.Code).To(BeEquivalentTo(http.StatusNotFound))
	})

	It("returns an internal server error when the query fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(chartUploadSelect, chartUploadID, "1", 1).ReturnsError(errors.New("boom"))

		upload, apiErr := GetApplicationChartUploadByID(testCtx(1), chartUploadID, "1")

		Expect(upload).To(BeNil())
		Expect(apiErr).ToNot(BeNil())
		Expect(apiErr.Code).To(BeEquivalentTo(http.StatusInternalServerError))
	})
})
