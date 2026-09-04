/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"errors"
	"math"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/audit-service/internal/database"
	"github.com/dsx-ai-factory/run-ai/control-plane/audit-service/internal/database/models"
	"github.com/dsx-ai-factory/run-ai/control-plane/audit-service/internal/services/mocks"
	auth_api "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	auth_mocks "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/mocks"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/queryapi"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var (
	start, end      = time.Now(), time.Now()
	permittedScopes = &auth_api.PermittedScopesActions{
		Read: auth_api.PermittedScopes{
			Tenants: []string{"1"},
		},
	}
)

func TestGetAuditLogs_NoTenantId(t *testing.T) {
	ctx := context.Background()
	authClient := auth_mocks.NewIAuthorizationClient(t)
	alf := ApiAuditLogsService{
		authClient: authClient,
	}

	res, err := alf.GetAuditLogs(ctx, start, end, 1, 2, "desc", "timestamp", []string{})
	require.Error(t, err)
	require.Equal(t, res.Code, http.StatusBadRequest)
}

func TestGetAuditLogs_FailedToFetch(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, "1")
	authClient := auth_mocks.NewIAuthorizationClient(t)
	authClient.On("GetPermittedScopes", ctx, mock.Anything, mock.Anything).Return(permittedScopes, http.StatusOK, nil)
	dao := mocks.NewAuditReader(t)
	alf := ApiAuditLogsService{
		dao:        dao,
		authClient: authClient,
	}
	qp := database.QueryParams{
		Offset:         1,
		Samples:        2,
		Order:          "desc",
		SortBy:         "timestamp",
		Start:          &start,
		End:            &end,
		Filters:        []queryapi.Filter{},
		WithTotalCount: true,
	}
	dao.On("Read", ctx, 1, qp).Return(nil, 0, errors.New("failed")).Once()

	res, err := alf.GetAuditLogs(ctx, start, end, 1, 2, "desc", "timestamp", []string{})
	require.Error(t, err)
	require.Equal(t, res.Code, http.StatusInternalServerError)
}

func TestGetAuditLogs_NotPermitted(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, "1")
	authClient := auth_mocks.NewIAuthorizationClient(t)

	permittedScopesResponse := &auth_api.PermittedScopesActions{
		Read: auth_api.PermittedScopes{
			Tenants: []string{"2"},
		},
	}
	authClient.On("GetPermittedScopes", ctx, auth_api.RESOURCETYPE_EVENTS_HISTORY, mock.Anything).Return(permittedScopesResponse, http.StatusOK, nil)
	alf := ApiAuditLogsService{
		authClient: authClient,
	}

	res, err := alf.GetAuditLogs(ctx, start, end, 1, 2, "desc", "timestamp", []string{})
	require.Error(t, err)
	require.Equal(t, res.Code, http.StatusForbidden)
}

func TestGetAuditLogs_FailedToGetPermissionScopes(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, "1")
	authClient := auth_mocks.NewIAuthorizationClient(t)
	authClient.On("GetPermittedScopes", ctx, auth_api.RESOURCETYPE_EVENTS_HISTORY, mock.Anything).Return(nil, http.StatusInternalServerError, errors.New("internal error"))
	alf := ApiAuditLogsService{
		authClient: authClient,
	}

	res, err := alf.GetAuditLogs(ctx, start, end, 1, 2, "desc", "timestamp", []string{})
	require.Error(t, err)
	require.Equal(t, res.Code, http.StatusInternalServerError)
}

func TestGetAuditLogs_Succeed(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, "1")
	authClient := auth_mocks.NewIAuthorizationClient(t)
	authClient.On("GetPermittedScopes", ctx, mock.Anything, mock.Anything).Return(permittedScopes, http.StatusOK, nil)
	dao := mocks.NewAuditReader(t)
	alf := ApiAuditLogsService{
		dao:        dao,
		authClient: authClient,
	}
	qp := database.QueryParams{
		Offset:         1,
		Samples:        2,
		Order:          "desc",
		SortBy:         "timestamp",
		Start:          &start,
		End:            &end,
		Filters:        []queryapi.Filter{},
		WithTotalCount: true,
	}
	dao.On("Read", ctx, 1, qp).Return([]*models.AuditLog{}, 0, nil).Once()

	res, err := alf.GetAuditLogs(ctx, start, end, 1, 2, "desc", "timestamp", []string{})
	require.NoError(t, err)
	require.Equal(t, res.Code, http.StatusOK)
}

func TestDownloadAuditLogs_Succeed(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, "1")
	authClient := auth_mocks.NewIAuthorizationClient(t)
	authClient.On("GetPermittedScopes", ctx, mock.Anything, mock.Anything).Return(permittedScopes, http.StatusOK, nil)
	dao := mocks.NewAuditReader(t)
	alf := ApiAuditLogsService{
		dao:             dao,
		deleteFilesChan: make(chan string, math.MaxInt16),
		authClient:      authClient,
	}
	qp := database.QueryParams{
		Samples: 2,
		Start:   &start,
		End:     &end,
		Filters: []queryapi.Filter{},
	}
	dao.On("Read", ctx, 1, qp).Return([]*models.AuditLog{}, 0, nil).Once()

	res, err := alf.DownloadAuditLogs(ctx, start, end, "CSV", 0, 2, []string{})
	require.NoError(t, err)
	require.Equal(t, res.Code, http.StatusOK)
	os.Remove(<-alf.deleteFilesChan)
}

func TestDeleteDownloadedFiles(t *testing.T) {
	alf := ApiAuditLogsService{
		deleteFilesChan: make(chan string, math.MaxInt16),
	}
	go alf.deleteDownloadedFiles()
	alf.deleteFilesChan <- "test"
	time.Sleep(1 * time.Second)
	require.Equal(t, 0, len(alf.deleteFilesChan))

	alf.deleteFilesChan <- "test1"
	alf.deleteFilesChan <- "test2"
	time.Sleep(1 * time.Second)
	require.Equal(t, 0, len(alf.deleteFilesChan))
}

func TestUnescapeSourceIpFilters(t *testing.T) {
	filterBy := []string{"source_ip==1.1.1.1;2.2.2.2", "name==test"}
	unescapedFilterBy := unescapeSourceIpFilters(filterBy)
	require.Equal(t, []string{"source_ip==1.1.1.1,2.2.2.2", "name==test"}, unescapedFilterBy)
}

func TestUnescapeSourceIpFilters_NoSourceIp_NoChange(t *testing.T) {
	filterBy := []string{"name==test", "id=@123"}
	unescapedFilterBy := unescapeSourceIpFilters(filterBy)
	require.Equal(t, []string{"name==test", "id=@123"}, unescapedFilterBy)
}
