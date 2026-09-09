/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package testing

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	audit "github.com/dsx-ai-factory/run-ai/control-plane/audit-service/pkg/client"
	audit_openapi "github.com/dsx-ai-factory/run-ai/control-plane/audit-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/audit-service/pkg/client/write"
	authorization_api "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	authorization_utils "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/helpers"
	metrics_swagger "github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/pkg/client/generated"
	metrics_mocks "github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/pkg/client/mocks"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/services/reports"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
)

type recordingAuditWriter struct {
	mu        sync.Mutex
	delay     time.Duration
	err       error
	entityIds []string
}

func (w *recordingAuditWriter) Write(log audit_openapi.AuditLog) error {
	time.Sleep(w.delay)
	w.mu.Lock()
	defer w.mu.Unlock()
	w.entityIds = append(w.entityIds, log.GetEntityId())
	return w.err
}

func (w *recordingAuditWriter) written() []string {
	w.mu.Lock()
	defer w.mu.Unlock()
	return append([]string(nil), w.entityIds...)
}

var _ = Describe("BatchDeleteReports API tests", func() {
	ctx := SetContext(TestTenant, TestUser)

	makeSvc := func(mc *metrics_mocks.MetricsClientContract) reports.ReportsApiService {
		return reports.ReportsApiService{MetricsClient: mc}
	}

	makeReport := func(id, createdBy string) metrics_swagger.Report {
		return metrics_swagger.Report{Id: id, CreatedBy: createdBy}
	}

	makeBatchGetResp := func(succeeded []metrics_swagger.Report, failed []metrics_swagger.BatchResponseFailureItem) *metrics_swagger.ReportsBatchGetResponse {
		return &metrics_swagger.ReportsBatchGetResponse{
			Id:        uuid.NewString(),
			Succeeded: succeeded,
			Failed:    failed,
		}
	}

	makeBatchDeleteResp := func(batchId string, accepted []string, failed []metrics_swagger.BatchResponseFailureItem) *metrics_swagger.BatchResponseAsync {
		return &metrics_swagger.BatchResponseAsync{
			Id:       batchId,
			Accepted: accepted,
			Failed:   failed,
		}
	}

	Context("request validation", func() {
		It("rejects empty ids with 400", func() {
			mc := &metrics_mocks.MetricsClientContract{}
			svc := makeSvc(mc)

			resp, err := svc.BatchDeleteReports(ctx, api.BatchDeleteReportsRequest{Ids: []string{}})

			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusBadRequest))
			mc.AssertNumberOfCalls(GinkgoT(), "BatchGetReportsByIds", 0)
		})

		It("rejects duplicate ids with 400", func() {
			id := uuid.NewString()
			mc := &metrics_mocks.MetricsClientContract{}
			svc := makeSvc(mc)

			resp, err := svc.BatchDeleteReports(ctx, api.BatchDeleteReportsRequest{Ids: []string{id, id}})

			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusBadRequest))
			mc.AssertNumberOfCalls(GinkgoT(), "BatchGetReportsByIds", 0)
		})
	})

	Context("all reports owned by the caller", func() {
		It("returns 202 with all reports accepted", func() {
			id1, id2 := uuid.NewString(), uuid.NewString()
			batchOpId := uuid.NewString()

			mc := &metrics_mocks.MetricsClientContract{}
			mc.On("BatchGetReportsByIds", mock.Anything, []string{id1, id2}).Return(
				makeBatchGetResp(
					[]metrics_swagger.Report{makeReport(id1, TestUser), makeReport(id2, TestUser)},
					nil,
				),
				http.StatusOK, nil,
			)
			mc.On("BatchDeleteReports", mock.Anything, []string{id1, id2}).Return(
				makeBatchDeleteResp(batchOpId, []string{id1, id2}, nil),
				http.StatusAccepted, nil,
			)

			svc := makeSvc(mc)
			resp, err := svc.BatchDeleteReports(ctx, api.BatchDeleteReportsRequest{Ids: []string{id1, id2}})

			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusAccepted))
			body := resp.Body.(api.BatchDeleteReportsResponse)
			Expect(body.Id).To(Equal(batchOpId))
			Expect(body.Accepted).To(ConsistOf(id1, id2))
			Expect(body.Failed).To(BeEmpty())
			Expect(body.TotalRequested).To(Equal(int32(2)))
			Expect(body.TotalAccepted).To(Equal(int32(2)))
			Expect(body.TotalFailed).To(Equal(int32(0)))
			mc.AssertExpectations(GinkgoT())
		})
	})

	Context("some reports not found in metrics-service", func() {
		It("not-found reports appear in failed, found ones proceed to delete", func() {
			foundId := uuid.NewString()
			missingId := uuid.NewString()
			batchOpId := uuid.NewString()

			mc := &metrics_mocks.MetricsClientContract{}
			mc.On("BatchGetReportsByIds", mock.Anything, []string{foundId, missingId}).Return(
				makeBatchGetResp(
					[]metrics_swagger.Report{makeReport(foundId, TestUser)},
					[]metrics_swagger.BatchResponseFailureItem{{Id: missingId, Code: http.StatusNotFound, Message: "not found"}},
				),
				http.StatusOK, nil,
			)
			mc.On("BatchDeleteReports", mock.Anything, []string{foundId}).Return(
				makeBatchDeleteResp(batchOpId, []string{foundId}, nil),
				http.StatusAccepted, nil,
			)

			svc := makeSvc(mc)
			resp, err := svc.BatchDeleteReports(ctx, api.BatchDeleteReportsRequest{Ids: []string{foundId, missingId}})

			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusAccepted))
			body := resp.Body.(api.BatchDeleteReportsResponse)
			Expect(body.Id).To(Equal(batchOpId))
			Expect(body.Accepted).To(ConsistOf(foundId))
			Expect(body.Failed).To(HaveLen(1))
			Expect(body.Failed[0].Id).To(Equal(missingId))
			Expect(body.TotalRequested).To(Equal(int32(2)))
			Expect(body.TotalAccepted).To(Equal(int32(1)))
			Expect(body.TotalFailed).To(Equal(int32(1)))
			mc.AssertExpectations(GinkgoT())
		})
	})

	Context("permission checks", func() {
		It("rejects report owned by another user when caller has no tenant scope", func() {
			othersId := uuid.NewString()

			mc := &metrics_mocks.MetricsClientContract{}
			mc.On("BatchGetReportsByIds", mock.Anything, []string{othersId}).Return(
				makeBatchGetResp(
					[]metrics_swagger.Report{makeReport(othersId, "other-user@run.ai")},
					nil,
				),
				http.StatusOK, nil,
			)

			svc := makeSvc(mc)
			resp, err := svc.BatchDeleteReports(ctx, api.BatchDeleteReportsRequest{Ids: []string{othersId}})

			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusAccepted))
			body := resp.Body.(api.BatchDeleteReportsResponse)
			Expect(body.Id).NotTo(BeEmpty())
			Expect(body.Accepted).To(BeEmpty())
			Expect(body.Failed).To(HaveLen(1))
			Expect(body.Failed[0].Id).To(Equal(othersId))
			Expect(int(body.Failed[0].Code)).To(Equal(http.StatusForbidden))
			mc.AssertNumberOfCalls(GinkgoT(), "BatchDeleteReports", 0)
		})

		It("allows tenant admin to delete another user's report", func() {
			othersId := uuid.NewString()
			batchOpId := uuid.NewString()
			tenantAdminCtx := context.WithValue(ctx, authorization_utils.PermittedScopesKey,
				&authorization_api.PermittedScopes{Tenants: []string{TestTenant}})

			mc := &metrics_mocks.MetricsClientContract{}
			mc.On("BatchGetReportsByIds", mock.Anything, []string{othersId}).Return(
				makeBatchGetResp(
					[]metrics_swagger.Report{makeReport(othersId, "other-user@run.ai")},
					nil,
				),
				http.StatusOK, nil,
			)
			mc.On("BatchDeleteReports", mock.Anything, []string{othersId}).Return(
				makeBatchDeleteResp(batchOpId, []string{othersId}, nil),
				http.StatusAccepted, nil,
			)

			svc := makeSvc(mc)
			resp, err := svc.BatchDeleteReports(tenantAdminCtx, api.BatchDeleteReportsRequest{Ids: []string{othersId}})

			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusAccepted))
			body := resp.Body.(api.BatchDeleteReportsResponse)
			Expect(body.Id).To(Equal(batchOpId))
			Expect(body.Accepted).To(ConsistOf(othersId))
			Expect(body.Failed).To(BeEmpty())
			mc.AssertExpectations(GinkgoT())
		})
	})

	Context("all reports forbidden", func() {
		It("response Id is a generated uuid — no batch op was submitted to metrics-service", func() {
			othersId := uuid.NewString()

			mc := &metrics_mocks.MetricsClientContract{}
			mc.On("BatchGetReportsByIds", mock.Anything, []string{othersId}).Return(
				makeBatchGetResp(
					[]metrics_swagger.Report{makeReport(othersId, "other-user@run.ai")},
					nil,
				),
				http.StatusOK, nil,
			)

			svc := makeSvc(mc)
			resp, err := svc.BatchDeleteReports(ctx, api.BatchDeleteReportsRequest{Ids: []string{othersId}})

			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusAccepted))
			body := resp.Body.(api.BatchDeleteReportsResponse)
			Expect(body.Id).NotTo(BeEmpty())
			Expect(body.Accepted).To(BeEmpty())
			Expect(body.Failed).To(HaveLen(1))
			mc.AssertNumberOfCalls(GinkgoT(), "BatchDeleteReports", 0)
		})
	})

	Context("metrics-service returns 5xx during delete", func() {
		It("folds permitted ids into failed and still returns 202", func() {
			id1, id2 := uuid.NewString(), uuid.NewString()

			mc := &metrics_mocks.MetricsClientContract{}
			mc.On("BatchGetReportsByIds", mock.Anything, []string{id1, id2}).Return(
				makeBatchGetResp(
					[]metrics_swagger.Report{makeReport(id1, TestUser), makeReport(id2, TestUser)},
					nil,
				),
				http.StatusOK, nil,
			)
			mc.On("BatchDeleteReports", mock.Anything, []string{id1, id2}).Return(
				(*metrics_swagger.BatchResponseAsync)(nil),
				http.StatusInternalServerError,
				fmt.Errorf("metrics-service unavailable"),
			)

			svc := makeSvc(mc)
			resp, err := svc.BatchDeleteReports(ctx, api.BatchDeleteReportsRequest{Ids: []string{id1, id2}})

			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusAccepted))
			body := resp.Body.(api.BatchDeleteReportsResponse)
			Expect(body.Id).NotTo(BeEmpty())
			Expect(body.Accepted).To(BeEmpty())
			Expect(body.Failed).To(HaveLen(2))
			Expect(body.TotalRequested).To(Equal(int32(2)))
			Expect(body.TotalAccepted).To(Equal(int32(0)))
			Expect(body.TotalFailed).To(Equal(int32(2)))
			mc.AssertExpectations(GinkgoT())
		})
	})

	Context("audit logging of accepted deletions", func() {
		const (
			acceptedCount   = 20
			perWriteDelay   = 50 * time.Millisecond
			responseBudget  = 5 * perWriteDelay
			auditWriteLimit = 3 * acceptedCount * perWriteDelay
		)

		useAuditWriter := func(writer audit.Writer) {
			audit.SetMock(writer)
			DeferCleanup(func() { audit.SetMock(&write.EmptyWriter{}) })
		}

		mockAcceptAll := func(ids []string) *metrics_mocks.MetricsClientContract {
			reportsList := make([]metrics_swagger.Report, 0, len(ids))
			for _, id := range ids {
				reportsList = append(reportsList, makeReport(id, TestUser))
			}

			mc := &metrics_mocks.MetricsClientContract{}
			mc.On("BatchGetReportsByIds", mock.Anything, ids).Return(
				makeBatchGetResp(reportsList, nil), http.StatusOK, nil,
			)
			mc.On("BatchDeleteReports", mock.Anything, ids).Return(
				makeBatchDeleteResp(uuid.NewString(), ids, nil), http.StatusAccepted, nil,
			)
			return mc
		}

		makeIds := func(n int) []string {
			ids := make([]string, 0, n)
			for range n {
				ids = append(ids, uuid.NewString())
			}
			return ids
		}

		It("returns 202 without waiting for the audit writes, then writes them all", func() {
			writer := &recordingAuditWriter{delay: perWriteDelay}
			useAuditWriter(writer)

			ids := makeIds(acceptedCount)
			svc := makeSvc(mockAcceptAll(ids))

			start := time.Now()
			resp, err := svc.BatchDeleteReports(ctx, api.BatchDeleteReportsRequest{Ids: ids})
			elapsed := time.Since(start)

			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusAccepted))
			Expect(resp.Body.(api.BatchDeleteReportsResponse).Accepted).To(HaveLen(acceptedCount))
			Expect(elapsed).To(BeNumerically("<", responseBudget))
			Eventually(writer.written, auditWriteLimit, 10*time.Millisecond).Should(ConsistOf(ids))
		})

		// No response-time bound here: at capacity the writes are inline by design.
		It("keeps writing every audit record once the writer pool is saturated", func() {
			writer := &recordingAuditWriter{delay: perWriteDelay}
			useAuditWriter(writer)

			const concurrentRequests = 12
			batches := make([][]string, concurrentRequests)
			var wg sync.WaitGroup
			for i := range concurrentRequests {
				batches[i] = makeIds(2)
				wg.Add(1)
				go func(ids []string) {
					defer wg.Done()
					defer GinkgoRecover()
					svc := makeSvc(mockAcceptAll(ids))
					resp, err := svc.BatchDeleteReports(ctx, api.BatchDeleteReportsRequest{Ids: ids})
					Expect(err).To(BeNil())
					Expect(resp.Code).To(Equal(http.StatusAccepted))
				}(batches[i])
			}
			wg.Wait()

			allIds := []string{}
			for _, ids := range batches {
				allIds = append(allIds, ids...)
			}
			Eventually(writer.written, auditWriteLimit, 10*time.Millisecond).Should(ConsistOf(allIds))
		})

		It("still attempts every audit write when the writer keeps failing", func() {
			writer := &recordingAuditWriter{err: fmt.Errorf("audit backend unavailable")}
			useAuditWriter(writer)

			ids := makeIds(acceptedCount)
			svc := makeSvc(mockAcceptAll(ids))

			resp, err := svc.BatchDeleteReports(ctx, api.BatchDeleteReportsRequest{Ids: ids})

			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusAccepted))
			Eventually(writer.written, auditWriteLimit, 10*time.Millisecond).Should(ConsistOf(ids))
		})

		It("writes the forbidden-path audit without blocking the response", func() {
			writer := &recordingAuditWriter{}
			useAuditWriter(writer)

			othersId := uuid.NewString()
			mc := &metrics_mocks.MetricsClientContract{}
			mc.On("BatchGetReportsByIds", mock.Anything, []string{othersId}).Return(
				makeBatchGetResp([]metrics_swagger.Report{makeReport(othersId, "other-user@run.ai")}, nil),
				http.StatusOK, nil,
			)

			svc := makeSvc(mc)
			_, err := svc.BatchDeleteReports(ctx, api.BatchDeleteReportsRequest{Ids: []string{othersId}})

			Expect(err).To(BeNil())
			Eventually(writer.written, auditWriteLimit, 10*time.Millisecond).Should(ConsistOf(othersId))
		})
	})

	Context("mixed: some not found, some forbidden, some accepted", func() {
		It("partitions results correctly across all outcome categories", func() {
			ownId := uuid.NewString()
			othersId := uuid.NewString()
			missingId := uuid.NewString()
			batchOpId := uuid.NewString()

			mc := &metrics_mocks.MetricsClientContract{}
			mc.On("BatchGetReportsByIds", mock.Anything, []string{ownId, othersId, missingId}).Return(
				makeBatchGetResp(
					[]metrics_swagger.Report{
						makeReport(ownId, TestUser),
						makeReport(othersId, "other-user@run.ai"),
					},
					[]metrics_swagger.BatchResponseFailureItem{{Id: missingId, Code: http.StatusNotFound, Message: "not found"}},
				),
				http.StatusOK, nil,
			)
			mc.On("BatchDeleteReports", mock.Anything, []string{ownId}).Return(
				makeBatchDeleteResp(batchOpId, []string{ownId}, nil),
				http.StatusAccepted, nil,
			)

			svc := makeSvc(mc)
			resp, err := svc.BatchDeleteReports(ctx, api.BatchDeleteReportsRequest{Ids: []string{ownId, othersId, missingId}})

			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusAccepted))
			body := resp.Body.(api.BatchDeleteReportsResponse)
			Expect(body.Id).To(Equal(batchOpId))
			Expect(body.Accepted).To(ConsistOf(ownId))
			Expect(body.Failed).To(HaveLen(2))
			Expect(body.TotalRequested).To(Equal(int32(3)))
			Expect(body.TotalAccepted).To(Equal(int32(1)))
			Expect(body.TotalFailed).To(Equal(int32(2)))
			mc.AssertExpectations(GinkgoT())
		})
	})
})
