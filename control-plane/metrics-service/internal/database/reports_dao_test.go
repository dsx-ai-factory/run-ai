/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package database

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/metrics-service/internal/database/dao"
	"github.com/run-ai/runai/control-plane/metrics-service/internal/database/models"
	metricerrors "github.com/run-ai/runai/control-plane/metrics-service/internal/errors"
	"github.com/run-ai/runai/control-plane/metrics-service/internal/generated/pkg/api"
	"github.com/run-ai/runai/runai-common-packages/integdb"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
	"github.com/run-ai/runai/runai-common-packages/runaidb"
	"k8s.io/utils/ptr"
)

var TENANT_ID = 1234

func TestInit(t *testing.T) {
	RegisterFailHandler(Fail)

	conf, cleanup, err := integdb.SetupPostgresDB(&integdb.IntegdbConfig{
		Schema:           "metrics",
		MigrationDirPath: "file://./migrations",
	})
	if err != nil {
		t.Fatalf("failed to set up test postgres: %v", err)
	}
	t.Cleanup(cleanup)

	if err := runaidb.Init(conf); err != nil {
		t.Fatalf("failed to init database: %v", err)
	}
	t.Cleanup(func() { _ = runaidb.Close() })

	RunSpecs(t, "Reports tests")
}

var _ = Describe("Reports tests", Ordered, func() {
	ctx := context.WithValue(context.WithValue(context.Background(), runaictx.CfTenantId, TENANT_ID), runaictx.CfSubjectID, "user")
	reportsDao := dao.NewPostgresReportsDao()
	createdReportIds := []string{}
	metrics := []*models.Metric{
		{
			DisplayName: "Metric A",
			Query:       "metric_a",
		},
	}
	report := models.Report{
		Name:     "testing",
		Start:    time.Now().Add(-1 * time.Hour),
		End:      time.Now(),
		Metrics:  metrics,
		TenantId: int32(TENANT_ID),
		Status:   string(api.REPORTSTATUS_PENDING),
		Labels: models.LabelArray{
			{
				Name:        "labelA",
				DisplayName: "Label A",
			},
		},
		AdditionalFields: models.AdditionalFields{
			"a": "B",
		},
	}

	AfterAll(func() {
		for _, reportId := range createdReportIds {
			err := reportsDao.DeleteReport(ctx, int32(TENANT_ID), reportId)
			Expect(err).To(BeNil())
		}
		createdReportIds = []string{}
	})

	BeforeAll(func() {
		reportId := uuid.New()
		r, _ := reportsDao.CreateReport(ctx, int32(TENANT_ID), models.Report{
			TenantId:         int32(TENANT_ID),
			ID:               reportId,
			Status:           string(api.REPORTSTATUS_PENDING),
			Labels:           report.Labels,
			AdditionalFields: report.AdditionalFields,
			Metrics:          report.Metrics,
		})
		createdReportIds = append(createdReportIds, r.ID.String())
	})

	It("Create and Get report", func() {
		reportResponse, err := reportsDao.CreateReport(ctx, int32(1), report)
		Expect(err).To(BeNil())
		Expect(reportResponse.ID).To(Not(BeNil()))

		reportGetResponse, err := reportsDao.GetReport(ctx, int32(1), reportResponse.ID.String())
		Expect(err).To(BeNil())
		Expect(reportGetResponse.ID).To(Equal(reportResponse.ID))
		Expect(reportGetResponse.TenantId).To(Equal(int32(1)))
		Expect(reportGetResponse.Status).To(Equal(string(api.REPORTSTATUS_PENDING)))
		Expect(reportGetResponse.StatusMessage).To(Equal(""))
		Expect(reportGetResponse.StatusUpdatedAt).To(Not(BeNil()))
		Expect(reportGetResponse.CreatedAt).To(Not(BeNil()))
		Expect(reportGetResponse.CreatedAt.Unix()).To(BeNumerically("<=", reportGetResponse.StatusUpdatedAt.Unix()))
		Expect(reportGetResponse.CreatedAt.Unix()).To(BeNumerically(">", reportGetResponse.StatusUpdatedAt.Unix()-5000)) // time of postgres query should not be near 5 sec
		Expect(reportGetResponse.CreatedBy).To(Equal("user"))
		Expect(len(reportGetResponse.Metrics)).To(Equal(1))
		Expect(len(reportGetResponse.Labels)).To(Equal(1))
		Expect(reportGetResponse.Labels[0].Name).To(Equal(reportResponse.Labels[0].Name))
		Expect(reportGetResponse.Labels[0].DisplayName).To(Equal(reportResponse.Labels[0].DisplayName))
		Expect(reportGetResponse.AdditionalFields).To(Not(BeNil()))
		_ = reportsDao.DeleteReport(ctx, int32(1), reportResponse.ID.String())
	})

	It("Get reports by ids", func() {
		tenantOneReport := report
		tenantOneReport.Name = "get-reports-by-ids-tenant-one"
		reportResponse, err := reportsDao.CreateReport(ctx, int32(1), tenantOneReport)
		Expect(err).To(BeNil())

		otherTenantReport := report
		otherTenantReport.Name = "get-reports-by-ids-other-tenant"
		otherReportResponse, err := reportsDao.CreateReport(ctx, int32(TENANT_ID), otherTenantReport)
		Expect(err).To(BeNil())

		reportsGetResponse, err := reportsDao.GetReportsByIds(ctx, int32(1), []string{
			reportResponse.ID.String(),
			otherReportResponse.ID.String(),
		})
		Expect(err).To(BeNil())
		Expect(len(reportsGetResponse)).To(Equal(1))
		Expect(reportsGetResponse[0].ID).To(Equal(reportResponse.ID))

		Expect(reportsDao.DeleteReport(ctx, int32(1), reportResponse.ID.String())).To(BeNil())
		Expect(reportsDao.DeleteReport(ctx, int32(TENANT_ID), otherReportResponse.ID.String())).To(BeNil())
	})

	It("Get report that not exist", func() {
		_, err := reportsDao.GetReport(ctx, int32(1), uuid.NewString())
		Expect(err).To(Not(BeNil()))
		Expect(err.Details).To(Equal("requested resource does not exist"))
		Expect(int(err.Code)).To(Equal(http.StatusNotFound))
	})

	It("Delete report", func() {
		deleteReport := report
		deleteReport.Name = "delete-report-test"
		reportResponse, err := reportsDao.CreateReport(ctx, int32(TENANT_ID), deleteReport)
		Expect(err).To(BeNil())

		id := reportResponse.ID.String()

		err = reportsDao.DeleteReport(ctx, int32(TENANT_ID), id)
		Expect(err).To(BeNil())

		_, err = reportsDao.GetReport(ctx, int32(TENANT_ID), id)
		Expect(err.Message).To(Equal("Failed to get report"))
	})

	It("Get all pending reports", func() {
		reports, err := reportsDao.GetPendingReports(ctx, 10)
		Expect(err).To(BeNil())
		Expect(len(reports)).To(BeNumerically(">=", len(createdReportIds)))

		report := reports[0]
		Expect(report.ID).To(Not(BeNil()))
		Expect(report.TenantId).To(Equal(int32(TENANT_ID)))
		Expect(report.Status).To(Equal(string(api.REPORTSTATUS_PENDING)))
		Expect(report.StatusMessage).To(Equal(""))
		Expect(report.StatusUpdatedAt).To(Not(BeNil()))
		Expect(len(report.Metrics)).To(Equal(1))
		Expect(len(report.Labels)).To(Equal(1))
	})

	It("Get amount of total reports by status", func() {
		statusToReportsAmountMap := map[string]int{}
		for _, reportId := range createdReportIds {
			report, apiErr := reportsDao.GetReport(ctx, int32(TENANT_ID), reportId)
			Expect(apiErr).To(BeNil())
			statusToReportsAmountMap[report.Status]++
		}

		statusMap, err := reportsDao.GetTotalReportsByStatus(ctx)
		Expect(err).To(BeNil())
		for status, amountOfReports := range statusMap {
			Expect(amountOfReports).To(Equal(statusToReportsAmountMap[status]))
		}
	})

	It("Update report status", func() {
		Expect(len(createdReportIds)).To(BeNumerically(">", 0))

		reportId := createdReportIds[0]

		err := reportsDao.UpdateReportStatus(int32(TENANT_ID), reportId, api.REPORTSTATUS_FAILED, nil)
		Expect(err).To(BeNil())

		err = reportsDao.UpdateReportStatus(int32(TENANT_ID), reportId, api.REPORTSTATUS_PROCESSING, ptr.To(api.REPORTSTATUS_FAILED))
		Expect(err).To(BeNil())
	})

	It("Update report status with message", func() {
		Expect(len(createdReportIds)).To(BeNumerically(">", 0))

		reportId := createdReportIds[0]

		err := reportsDao.UpdateReportStatusWithMessage(int32(TENANT_ID), reportId, api.REPORTSTATUS_FAILED, "failed", nil)
		Expect(err).To(BeNil())

		report, apiErr := reportsDao.GetReport(ctx, int32(TENANT_ID), reportId)
		Expect(apiErr).To(BeNil())
		Expect(report.Status).To(Equal(string(api.REPORTSTATUS_FAILED)))
		Expect(report.StatusMessage).To(Equal("failed"))
	})

	It("Update report status to processing twice", func() {
		Expect(len(createdReportIds)).To(BeNumerically(">", 0))

		reportId := createdReportIds[0]
		err := reportsDao.UpdateReportStatus(int32(TENANT_ID), reportId, api.REPORTSTATUS_PENDING, nil)
		Expect(err).To(BeNil())

		err = reportsDao.UpdateReportStatus(int32(TENANT_ID), reportId, api.REPORTSTATUS_PROCESSING, ptr.To(api.REPORTSTATUS_PENDING))
		Expect(err).To(BeNil())

		err = reportsDao.UpdateReportStatus(int32(TENANT_ID), reportId, api.REPORTSTATUS_PROCESSING, ptr.To(api.REPORTSTATUS_PENDING))
		Expect(err).To(Not(BeNil()))
	})

	Context("UpdatePendingDelete", func() {
		It("should mark non-processing reports as pending delete and fetch them", func() {
			reportResponse, err := reportsDao.CreateReport(ctx, int32(TENANT_ID), report)
			Expect(err).To(BeNil())
			existingId := reportResponse.ID.String()
			unknownId := uuid.NewString()

			updated, apiErr := reportsDao.UpdatePendingDelete(ctx, int32(TENANT_ID), []string{existingId, unknownId})
			Expect(apiErr).To(BeNil())
			Expect(updated).To(ConsistOf(existingId))

			markedReport, getErr := reportsDao.GetReport(ctx, int32(TENANT_ID), existingId)
			Expect(getErr).To(BeNil())
			Expect(markedReport.Status).To(Equal(string(api.REPORTSTATUS_PENDING_DELETE)))

			pendingDeleteReports, getPendingDeleteErr := reportsDao.GetPendingDeleteReports(ctx, 10)
			Expect(getPendingDeleteErr).To(BeNil())
			pendingDeleteReportIds := make([]uuid.UUID, 0, len(pendingDeleteReports))
			for _, pendingDeleteReport := range pendingDeleteReports {
				pendingDeleteReportIds = append(pendingDeleteReportIds, pendingDeleteReport.ID)
			}
			Expect(pendingDeleteReportIds).To(ContainElement(reportResponse.ID))

			updateErr := reportsDao.UpdateReportStatus(int32(TENANT_ID), existingId, api.REPORTSTATUS_DELETING, ptr.To(api.REPORTSTATUS_PENDING_DELETE))
			Expect(updateErr).To(BeNil())

			_ = reportsDao.DeleteReport(ctx, int32(TENANT_ID), existingId)
		})

		It("should skip processing reports", func() {
			reportResponse, err := reportsDao.CreateReport(ctx, int32(TENANT_ID), report)
			Expect(err).To(BeNil())
			existingId := reportResponse.ID.String()

			updateErr := reportsDao.UpdateReportStatus(int32(TENANT_ID), existingId, api.REPORTSTATUS_PROCESSING, ptr.To(api.REPORTSTATUS_PENDING))
			Expect(updateErr).To(BeNil())

			updated, apiErr := reportsDao.UpdatePendingDelete(ctx, int32(TENANT_ID), []string{existingId})
			Expect(apiErr).To(BeNil())
			Expect(updated).To(BeEmpty())

			processingReport, getErr := reportsDao.GetReport(ctx, int32(TENANT_ID), existingId)
			Expect(getErr).To(BeNil())
			Expect(processingReport.Status).To(Equal(string(api.REPORTSTATUS_PROCESSING)))

			_ = reportsDao.DeleteReport(ctx, int32(TENANT_ID), existingId)
		})

		It("should skip reports already in the deletion workflow", func() {
			deletingInput := report
			deletingInput.Name = "deleting-skip-test"
			deletingReport, err := reportsDao.CreateReport(ctx, int32(TENANT_ID), deletingInput)
			Expect(err).To(BeNil())
			deletingId := deletingReport.ID.String()

			_, apiErr := reportsDao.UpdatePendingDelete(ctx, int32(TENANT_ID), []string{deletingId})
			Expect(apiErr).To(BeNil())
			updateErr := reportsDao.UpdateReportStatus(int32(TENANT_ID), deletingId, api.REPORTSTATUS_DELETING, nil)
			Expect(updateErr).To(BeNil())

			updated, apiErr := reportsDao.UpdatePendingDelete(ctx, int32(TENANT_ID), []string{deletingId})
			Expect(apiErr).To(BeNil())
			Expect(updated).To(BeEmpty())

			deletingMarked, getErr := reportsDao.GetReport(ctx, int32(TENANT_ID), deletingId)
			Expect(getErr).To(BeNil())
			Expect(deletingMarked.Status).To(Equal(string(api.REPORTSTATUS_DELETING)))

			_ = reportsDao.DeleteReport(ctx, int32(TENANT_ID), deletingId)
		})
	})

	Context("Metrics store tests", func() {
		ctx := context.WithValue(context.WithValue(context.Background(), runaictx.CfTenantId, TENANT_ID), runaictx.CfSubjectID, "user")
		metricStoresDao := dao.NewMetricStoresDao()
		metricStoreEndpoint := models.MetricStoreEndpoint{
			Url: "some_url",
			Auth: &models.MetricStoreEndpointAuth{
				Type: models.ATBasic,
				Basic: &models.AuthBasic{
					Username: "test",
					Password: "secret",
				},
			},
		}
		metricStore := models.MetricStore{
			ID:           uuid.New(),
			TenantId:     int32(TENANT_ID),
			DatasourceId: "mimir_12345",
			MetricStoreDetails: models.MetricStoreDetails{
				Type:  models.MSTMimir,
				Read:  metricStoreEndpoint,
				Write: metricStoreEndpoint,
			},
		}

		Context("Create", func() {
			AfterEach(func() {
				_ = metricStoresDao.Delete(ctx, metricStore.TenantId)
			})

			It("should fail on duplicate tenant id", func() {
				metricStore.TenantId = 1
				res, err := metricStoresDao.Create(ctx, metricStore)
				Expect(err).To(BeNil())
				Expect(res.ID).To(Not(BeNil()))

				res, err = metricStoresDao.Create(ctx, metricStore)
				Expect(err).To(Not(BeNil()))
				Expect(err.Error()).To(Equal("metric store already exists"))
				Expect(errors.Is(err, metricerrors.ErrMetricStoreAlreadyExists)).To(BeTrue())

				_ = metricStoresDao.Delete(ctx, 1)
				metricStore.TenantId = int32(TENANT_ID)
			})

			It("should save all fields as expected", func() {
				res, err := metricStoresDao.GetByTenantId(ctx, int32(TENANT_ID))
				Expect(err).To(BeNil())
				Expect(res).To(BeNil())
				res, err = metricStoresDao.Create(ctx, metricStore)
				Expect(err).To(BeNil())
				Expect(res.ID).To(Not(BeNil()))

				metricStoreRes, err := metricStoresDao.GetByTenantId(ctx, int32(TENANT_ID))
				Expect(err).To(BeNil())
				Expect(metricStoreRes).To(Not(BeNil()))
				Expect(metricStoreRes.ID).To(Not(BeNil()))
				Expect(metricStoreRes.TenantId).To(Equal(int32(TENANT_ID)))
				Expect(metricStoreRes.MetricStoreDetails.Type).To(Equal(models.MSTMimir))
				Expect(metricStoreRes.MetricStoreDetails.Read.Url).To(Equal(metricStoreEndpoint.Url))
				Expect(metricStoreRes.MetricStoreDetails.Write.Url).To(Equal(metricStoreEndpoint.Url))
				Expect(metricStoreRes.MetricStoreDetails.Read.Auth.Type).To(Equal(metricStoreEndpoint.Auth.Type))
				Expect(metricStoreRes.MetricStoreDetails.Read.Auth.Basic.Password).To(Equal(metricStoreEndpoint.Auth.Basic.Password))
				Expect(metricStoreRes.MetricStoreDetails.Read.Auth.Basic.Username).To(Equal(metricStoreEndpoint.Auth.Basic.Username))
				Expect(metricStoreRes.MetricStoreDetails.Write.Auth.Type).To(Equal(metricStoreEndpoint.Auth.Type))
				Expect(metricStoreRes.MetricStoreDetails.Write.Auth.Basic.Password).To(Equal(metricStoreEndpoint.Auth.Basic.Password))
				Expect(metricStoreRes.MetricStoreDetails.Write.Auth.Basic.Username).To(Equal(metricStoreEndpoint.Auth.Basic.Username))
				_ = metricStoresDao.Delete(ctx, int32(TENANT_ID))
			})
		})

		Context("Get", func() {
			BeforeEach(func() {
				res, err := metricStoresDao.Create(ctx, metricStore)
				Expect(err).To(BeNil())
				Expect(res.ID).To(Not(BeNil()))
			})

			AfterEach(func() {
				_ = metricStoresDao.Delete(ctx, metricStore.TenantId)
			})

			It("should return nils on get deleted metric store", func() {
				metricStoreRes, err := metricStoresDao.GetByTenantId(ctx, int32(TENANT_ID))
				Expect(err).To(BeNil())
				Expect(metricStoreRes.TenantId).To(Equal(metricStore.TenantId))
				err = metricStoresDao.Delete(ctx, int32(TENANT_ID))
				Expect(err).To(BeNil())

				metricStoreRes, err = metricStoresDao.GetByTenantId(ctx, int32(TENANT_ID))
				Expect(err).To(BeNil())
				Expect(metricStoreRes).To(BeNil())
			})

			It("should return nils on not existing metric store", func() {
				res, err := metricStoresDao.GetByTenantId(ctx, int32(54321))
				Expect(err).To(BeNil())
				Expect(res).To(BeNil())
			})
		})

		Context("Delete", func() {
			It("should return nil error not existing metric store", func() {
				err := metricStoresDao.Delete(ctx, int32(TENANT_ID))
				Expect(err).To(BeNil())
			})
		})
	})
})
