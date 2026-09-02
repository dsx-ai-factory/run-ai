/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"context"
	"testing"
	"time"

	authorization "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/workloads-manager/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/workloads-manager/pkg/database/models"
	"github.com/run-ai/runai/runai-application-packages/convutils"
	dao_utils "github.com/run-ai/runai/runai-common-packages/dao-utils"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
	dbtesting "github.com/run-ai/runai/runai-common-packages/runaidb/testing"
)

const (
	testTenantID     int32 = 42
	testTenantIDStr        = "42"
	testWorkloadID         = "11111111-1111-1111-1111-111111111111"
	testWorkloadID2        = "22222222-2222-2222-2222-222222222222"
	testClusterID          = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	testDepartmentID       = "dept-1"
	testProjectID          = "proj-1"
	testUpdateCount  int64 = 7
)

func setupSQLMock(t *testing.T) dbtesting.Mocker {
	gormDB, mocker := dbtesting.NewGormMocker(t)
	dao_utils.SetMock(gormDB)
	t.Cleanup(func() { dao_utils.SetMock(nil) })
	return mocker
}

func ctxWithTenant() context.Context {
	return context.WithValue(context.Background(), runaictx.CfTenantId, testTenantIDStr)
}

func testQueueRecord() *models.WorkloadsQueueRecord {
	now := time.Date(2026, 8, 31, 12, 0, 0, 0, time.UTC)
	return &models.WorkloadsQueueRecord{
		ID:                testWorkloadID,
		CreatedAt:         now,
		CreatedBy:         "admin",
		UpdatedAt:         now,
		UpdatedBy:         "admin",
		TenantId:          testTenantID,
		ClusterId:         testClusterID,
		DepartmentId:      testDepartmentID,
		DepartmentName:    "research",
		ProjectId:         testProjectID,
		ProjectName:       "proj-a",
		Name:              "wl-1",
		ResourceGroup:     "batch",
		ResourceVersion:   "v1",
		ResourceKind:      "Job",
		DesiredPhase:      string(api.DESIREDPHASE_RUNNING),
		RawManifest:       dao_utils.JsonBMap{"kind": "Job"},
		EffectiveManifest: dao_utils.JsonBMap{"kind": "Job"},
		UpdateCount:       testUpdateCount,
		InputRequestPath:  "/api/v1/workloads",
		InputRequestBody:  dao_utils.JsonBMap{"name": "wl-1"},
	}
}

func testSyncStatusRecord() models.SyncStatusRecord {
	return models.SyncStatusRecord{
		WorkloadId:       testWorkloadID,
		TenantId:         testTenantID,
		ClusterId:        testClusterID,
		DepartmentId:     testDepartmentID,
		ProjectId:        testProjectID,
		Deleting:         false,
		SentToAgent:      testUpdateCount,
		ProcessedByAgent: testUpdateCount - 1,
	}
}

func tenantScope() *authorization.PermittedScopes {
	return &authorization.PermittedScopes{Tenants: []string{testTenantIDStr}}
}

func systemScope() *authorization.PermittedScopes {
	return &authorization.PermittedScopes{System: convutils.Ref(true)}
}

func projectClusterDeptScope() *authorization.PermittedScopes {
	return &authorization.PermittedScopes{
		Clusters:    []string{testClusterID},
		Departments: []string{testDepartmentID},
		Projects:    []string{testProjectID},
	}
}
