/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package reports_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/run-ai/runai/control-plane/metrics-service/internal/database/models"
	"github.com/run-ai/runai/control-plane/metrics-service/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/metrics-service/internal/reports"
	"github.com/run-ai/runai/control-plane/metrics-service/internal/reports/mocks"
	"github.com/run-ai/runai/runai-common-packages/config"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func initDeleterTestConfig(t *testing.T, runPeriodSec, fetchLimit int) {
	t.Helper()
	conf := &reports.ReportsConfiguration{}
	conf.SetDefaults()
	conf.RunPeriodSec = runPeriodSec
	conf.DeleteFetchLimit = fetchLimit
	config.Set(conf)
}

func TestDeleterStart_FailedToGetPendingDeleteReports(t *testing.T) {
	initDeleterTestConfig(t, 1, 10)

	ctx, cancel := context.WithCancel(context.Background())
	reportsDeleter := mocks.NewReportsDeleter(t)
	wg := mocks.NewWaitGrouper(t)
	deleter := reports.NewTestDeleter(wg, 10, mocks.NewReportsUpdater(t), reportsDeleter, mocks.NewS3Deleter(t))

	reportsDeleter.On("GetPendingDeleteReports", ctx, 10).Return(nil, &api.Error{Code: 500})
	wg.On("Wait")

	go deleter.Start(ctx)
	defer deleter.WaitStop()
	defer cancel()

	time.Sleep(time.Millisecond * 1_100)
	require.Equal(t, 0, deleter.ConcurrencyInUse())
}

func TestDeleterStart_NoReports(t *testing.T) {
	initDeleterTestConfig(t, 1, 10)

	ctx, cancel := context.WithCancel(context.Background())
	reportsDeleter := mocks.NewReportsDeleter(t)
	wg := mocks.NewWaitGrouper(t)
	deleter := reports.NewTestDeleter(wg, 10, mocks.NewReportsUpdater(t), reportsDeleter, mocks.NewS3Deleter(t))

	reportsDeleter.On("GetPendingDeleteReports", ctx, 10).Return([]models.Report{}, nil)
	wg.On("Wait")

	go deleter.Start(ctx)
	defer deleter.WaitStop()
	defer cancel()

	time.Sleep(time.Millisecond * 1_100)
	require.Equal(t, 0, deleter.ConcurrencyInUse())
}

func TestDeleterStart_WithReport(t *testing.T) {
	initDeleterTestConfig(t, 1, 10)

	called := make(chan bool, 1)
	ctx, cancel := context.WithCancel(context.Background())
	reportsUpdater := mocks.NewReportsUpdater(t)
	reportsDeleter := mocks.NewReportsDeleter(t)
	s3Client := mocks.NewS3Deleter(t)
	wg := mocks.NewWaitGrouper(t)

	reportId := uuid.New()
	report := models.Report{ID: reportId, TenantId: 1}
	deleter := reports.NewTestDeleter(wg, 10, reportsUpdater, reportsDeleter, s3Client)

	reportsDeleter.On("GetPendingDeleteReports", ctx, 10).Return([]models.Report{report}, nil).Once()
	reportsUpdater.On("UpdateReportStatus", int32(1), reportId.String(), api.REPORTSTATUS_DELETING, mock.MatchedBy(func(previousStatus *api.ReportStatus) bool {
		return previousStatus != nil && *previousStatus == api.REPORTSTATUS_PENDING_DELETE
	})).Return(nil).Once()
	s3Client.On("Delete", reportId.String()).Return(nil).Once()
	reportsDeleter.On("DeleteReport", mock.Anything, int32(1), reportId.String()).Return(nil).Once()

	wg.On("Add", 1).Once()
	wg.On("Done").Run(func(args mock.Arguments) {
		called <- true
	}).Once()
	wg.On("Wait")

	go deleter.Start(ctx)
	defer deleter.WaitStop()
	defer cancel()

	<-called
	require.Equal(t, 0, deleter.ConcurrencyInUse())
}

func TestDeleterStart_S3Error(t *testing.T) {
	initDeleterTestConfig(t, 1, 10)

	called := make(chan bool, 1)
	ctx, cancel := context.WithCancel(context.Background())
	reportsUpdater := mocks.NewReportsUpdater(t)
	reportsDeleter := mocks.NewReportsDeleter(t)
	s3Client := mocks.NewS3Deleter(t)
	wg := mocks.NewWaitGrouper(t)

	reportId := uuid.New()
	report := models.Report{ID: reportId, TenantId: 1}
	s3Err := fmt.Errorf("s3 failure")
	deleter := reports.NewTestDeleter(wg, 10, reportsUpdater, reportsDeleter, s3Client)

	reportsDeleter.On("GetPendingDeleteReports", ctx, 10).Return([]models.Report{report}, nil).Once()
	reportsUpdater.On("UpdateReportStatus", int32(1), reportId.String(), api.REPORTSTATUS_DELETING, mock.MatchedBy(func(previousStatus *api.ReportStatus) bool {
		return previousStatus != nil && *previousStatus == api.REPORTSTATUS_PENDING_DELETE
	})).Return(nil).Once()
	s3Client.On("Delete", reportId.String()).Return(s3Err).Once()
	reportsUpdater.On("UpdateReportStatusWithMessage", int32(1), reportId.String(), api.REPORTSTATUS_PENDING_DELETE, s3Err.Error(), mock.MatchedBy(func(previousStatus *api.ReportStatus) bool {
		return previousStatus != nil && *previousStatus == api.REPORTSTATUS_DELETING
	})).Return(nil).Once()

	wg.On("Add", 1).Once()
	wg.On("Done").Run(func(args mock.Arguments) {
		called <- true
	}).Once()
	wg.On("Wait")

	go deleter.Start(ctx)
	defer deleter.WaitStop()
	defer cancel()

	<-called
}

func TestDeleterStart_DBDeleteError(t *testing.T) {
	initDeleterTestConfig(t, 1, 10)

	called := make(chan bool, 1)
	ctx, cancel := context.WithCancel(context.Background())
	reportsUpdater := mocks.NewReportsUpdater(t)
	reportsDeleter := mocks.NewReportsDeleter(t)
	s3Client := mocks.NewS3Deleter(t)
	wg := mocks.NewWaitGrouper(t)

	reportId := uuid.New()
	report := models.Report{ID: reportId, TenantId: 1}
	deleter := reports.NewTestDeleter(wg, 10, reportsUpdater, reportsDeleter, s3Client)

	reportsDeleter.On("GetPendingDeleteReports", ctx, 10).Return([]models.Report{report}, nil).Once()
	reportsUpdater.On("UpdateReportStatus", int32(1), reportId.String(), api.REPORTSTATUS_DELETING, mock.MatchedBy(func(previousStatus *api.ReportStatus) bool {
		return previousStatus != nil && *previousStatus == api.REPORTSTATUS_PENDING_DELETE
	})).Return(nil).Once()
	s3Client.On("Delete", reportId.String()).Return(nil).Once()
	reportsDeleter.On("DeleteReport", mock.Anything, int32(1), reportId.String()).
		Return(&api.Error{Code: 500, Message: "db failure"}).Once()

	wg.On("Add", 1).Once()
	wg.On("Done").Run(func(args mock.Arguments) {
		called <- true
	}).Once()
	wg.On("Wait")

	go deleter.Start(ctx)
	defer deleter.WaitStop()
	defer cancel()

	<-called
}

func TestDeleterStart_TransitionToDeletingError(t *testing.T) {
	initDeleterTestConfig(t, 1, 10)

	called := make(chan bool, 1)
	ctx, cancel := context.WithCancel(context.Background())
	reportsUpdater := mocks.NewReportsUpdater(t)
	reportsDeleter := mocks.NewReportsDeleter(t)
	s3Client := mocks.NewS3Deleter(t)
	wg := mocks.NewWaitGrouper(t)

	reportId := uuid.New()
	report := models.Report{ID: reportId, TenantId: 1}
	deleter := reports.NewTestDeleter(wg, 10, reportsUpdater, reportsDeleter, s3Client)

	reportsDeleter.On("GetPendingDeleteReports", ctx, 10).Return([]models.Report{report}, nil).Once()
	reportsUpdater.On("UpdateReportStatus", int32(1), reportId.String(), api.REPORTSTATUS_DELETING, mock.MatchedBy(func(previousStatus *api.ReportStatus) bool {
		return previousStatus != nil && *previousStatus == api.REPORTSTATUS_PENDING_DELETE
	})).Return(fmt.Errorf("transition failed")).Once()

	wg.On("Add", 1).Once()
	wg.On("Done").Run(func(args mock.Arguments) {
		called <- true
	}).Once()
	wg.On("Wait")

	go deleter.Start(ctx)
	defer deleter.WaitStop()
	defer cancel()

	<-called
	s3Client.AssertNotCalled(t, "Delete", mock.Anything)
	reportsDeleter.AssertNotCalled(t, "DeleteReport", mock.Anything, mock.Anything, mock.Anything)
}

func TestDeleterStart_GettingToConcurrencyLimit(t *testing.T) {
	initDeleterTestConfig(t, 2, 2)

	called := make(chan bool, 2)
	ctx, cancel := context.WithCancel(context.Background())
	reportsUpdater := mocks.NewReportsUpdater(t)
	reportsDeleter := mocks.NewReportsDeleter(t)
	s3Client := mocks.NewS3Deleter(t)
	wg := mocks.NewWaitGrouper(t)

	report1Id := uuid.New()
	report2Id := uuid.New()
	deleter := reports.NewTestDeleter(wg, 1, reportsUpdater, reportsDeleter, s3Client)

	reportsDeleter.On("GetPendingDeleteReports", ctx, 2).Return([]models.Report{
		{ID: report1Id, TenantId: 1},
		{ID: report2Id, TenantId: 1},
	}, nil)

	reportsUpdater.On("UpdateReportStatus", mock.Anything, mock.Anything, api.REPORTSTATUS_DELETING, mock.Anything).
		Return(nil)
	s3Client.On("Delete", mock.Anything).Run(func(args mock.Arguments) {
		time.Sleep(time.Millisecond * 1000)
	}).Return(nil)
	reportsDeleter.On("DeleteReport", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	wg.On("Add", 1)
	wg.On("Done").Run(func(args mock.Arguments) {
		called <- true
	}).Times(1)
	wg.On("Wait")

	go deleter.Start(ctx)
	defer cancel()
	defer deleter.WaitStop()

	<-called
	require.Equal(t, 0, len(called))
	require.Equal(t, 0, deleter.ConcurrencyInUse())
}
