/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package client

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	openapi "github.com/run-ai/runai/control-plane/metrics-service/pkg/client/generated"
)

type MetricsClientContract interface {
	GetMetrics(ctx context.Context, metricInfo *openapi.MetricInformation, overrideTenantId int32) (*openapi.GetMetrics200Response, int, error)
	GetReport(ctx context.Context, id string) (*openapi.Report, int, error)
	ListReports(ctx context.Context, filterBy []string, sortBy openapi.ReportFilterAndSortFields, sortOrder string, offset int32, limit int32, search string) ([]openapi.Report, int, error)
	CountReports(ctx context.Context, filterBy []string, search string) (int64, int, error)
	DeleteReport(ctx context.Context, id string) (int, error)
	BatchDeleteReports(ctx context.Context, ids []string) (*openapi.BatchResponseAsync, int, error)
	BatchGetReportsByIds(ctx context.Context, ids []string) (*openapi.ReportsBatchGetResponse, int, error)
	CreateReport(ctx context.Context, report *openapi.ReportRequestFields) (*openapi.Report, int, error)
}

type metricsClient struct {
	client *openapi.APIClient
	config *Config
}

func (c *metricsClient) GetMetrics(ctx context.Context, metricInfo *openapi.MetricInformation, overrideTenantId int32) (*openapi.GetMetrics200Response, int, error) {
	data, response, err := c.client.MetricsAPI.GetMetrics(ctx).MetricInformation(*metricInfo).OverrideTenantId(overrideTenantId).Execute()
	if err != nil {
		return nil, 0, fmt.Errorf("request for metric: %+v failed with error: %+v", metricInfo, err)
	}
	if response != nil && response.StatusCode != http.StatusOK {
		return nil, response.StatusCode, fmt.Errorf("request for metric: %+v failed with status code: %v and error: %v", metricInfo, response.StatusCode, err)
	}
	return data, response.StatusCode, nil
}

func (c *metricsClient) GetReport(ctx context.Context, id string) (*openapi.Report, int, error) {
	data, response, err := c.client.ReportsAPI.GetReportById(ctx, id).Execute()
	if response != nil && response.StatusCode != http.StatusOK {
		return nil, response.StatusCode, fmt.Errorf("get report request failed with status code: %v and error: %v", response.StatusCode, err)
	}
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("get report request failed with error: %v", err)
	}

	return data, response.StatusCode, nil
}

func (c *metricsClient) ListReports(ctx context.Context, filterBy []string, sortBy openapi.ReportFilterAndSortFields, sortOrder string, offset int32, limit int32, search string) ([]openapi.Report, int, error) {
	data, response, err := c.client.ReportsAPI.ListReports(ctx).
		FilterBy(filterBy).
		SortBy(sortBy).
		SortOrder(sortOrder).
		Offset(offset).
		Limit(limit).
		Search(search).
		Execute()
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("list reports request failed with error: %v", err)
	}
	if response != nil && response.StatusCode != http.StatusOK {
		return nil, response.StatusCode, fmt.Errorf("list reports request failed with status code: %v and error: %v", response.StatusCode, err)
	}

	return data.Reports, response.StatusCode, nil
}

func (c *metricsClient) CountReports(ctx context.Context, filterBy []string, search string) (int64, int, error) {
	data, response, err := c.client.ReportsAPI.CountReports(ctx).
		FilterBy(filterBy).
		Search(search).
		Execute()
	if err != nil {
		return 0, response.StatusCode, fmt.Errorf("list reports request failed with error: %v", err)
	}
	if response != nil && response.StatusCode != http.StatusOK {
		return 0, response.StatusCode, fmt.Errorf("list reports request failed with status code: %v and error: %v", response.StatusCode, err)
	}

	return data.Count, response.StatusCode, nil
}

func (c *metricsClient) DeleteReport(ctx context.Context, id string) (int, error) {
	response, err := c.client.ReportsAPI.DeleteReportById(ctx, id).Execute()
	if err != nil {
		return response.StatusCode, fmt.Errorf("delete report request failed with error: %v", err)
	}
	if response != nil && response.StatusCode != http.StatusNoContent {
		return response.StatusCode, fmt.Errorf("delete report request failed with status code: %v and error: %v", response.StatusCode, err)
	}

	return response.StatusCode, nil
}

func (c *metricsClient) BatchDeleteReports(ctx context.Context, ids []string) (*openapi.BatchResponseAsync, int, error) {
	data, response, err := c.client.ReportsAPI.BatchDeleteReports(ctx).
		ReportsIdsBatchRequest(openapi.ReportsIdsBatchRequest{Ids: ids}).
		Execute()
	if err != nil {
		return nil, responseStatusCode(response), fmt.Errorf("batch delete reports request failed with error: %v", err)
	}
	if response.StatusCode != http.StatusAccepted {
		return nil, response.StatusCode, fmt.Errorf("batch delete reports request failed with status code: %v", response.StatusCode)
	}

	return data, response.StatusCode, nil
}

func (c *metricsClient) BatchGetReportsByIds(ctx context.Context, ids []string) (*openapi.ReportsBatchGetResponse, int, error) {
	data, response, err := c.client.ReportsAPI.BatchGetReportsByIds(ctx).
		ReportsIdsBatchRequest(openapi.ReportsIdsBatchRequest{Ids: ids}).
		Execute()
	if err != nil {
		return nil, responseStatusCode(response), fmt.Errorf("batch get reports request failed with error: %v", err)
	}
	if response.StatusCode != http.StatusOK {
		return nil, response.StatusCode, fmt.Errorf("batch get reports request failed with status code: %v", response.StatusCode)
	}

	return data, response.StatusCode, nil
}

func responseStatusCode(response *http.Response) int {
	if response == nil {
		return 0
	}
	return response.StatusCode
}

func (c *metricsClient) CreateReport(ctx context.Context, report *openapi.ReportRequestFields) (*openapi.Report, int, error) {
	data, response, err := c.client.ReportsAPI.CreateReport(ctx).ReportRequestFields(*report).Execute()
	if err != nil {
		return nil, response.StatusCode, fmt.Errorf("create report request failed with error: %v", err)
	}
	if response != nil && response.StatusCode != http.StatusAccepted {
		return nil, response.StatusCode, fmt.Errorf("create report request failed with status code: %v and error: %v", response.StatusCode, err)
	}

	return data, response.StatusCode, nil
}

var once sync.Once

var singleInstance MetricsClientContract

type GetClientInstancer func() MetricsClientContract

var GetClientInstance = func() MetricsClientContract {
	once.Do(func() {
		config := readConfig()
		clientConfig := openapi.NewConfiguration()
		clientConfig.Host = config.MetricsServerHost
		clientConfig.Scheme = config.MetricsServerScheme
		singleInstance = &metricsClient{
			client: openapi.NewAPIClient(clientConfig),
			config: config,
		}
	})
	return singleInstance
}

func SetMock(mock MetricsClientContract) {
	GetClientInstance = func() MetricsClientContract {
		return mock
	}
}
