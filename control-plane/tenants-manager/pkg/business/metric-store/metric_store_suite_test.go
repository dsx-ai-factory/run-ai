/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package metric_store_test

import (
	"testing"

	metrics_mocks "github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/pkg/client/mocks"
	metric_store "github.com/dsx-ai-factory/run-ai/control-plane/tenants-manager/pkg/business/metric-store"
	dm "github.com/dsx-ai-factory/run-ai/control-plane/tenants-manager/pkg/database/dao/mocks"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestMetricStore(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "MetricStore Suite")
}

type ServiceTestMocks struct {
	metricStoreDaoMock    *dm.MetricStore
	tenantDaoMock         *dm.Tenant
	metricStoreClientMock *metrics_mocks.MetricStoresClienter
}

func initMocks(t FullGinkgoTInterface) (metric_store.MetricStoreService, *ServiceTestMocks) {
	metricStoreDao := dm.NewMetricStore(t)
	tenantDao := dm.NewTenant(t)
	metricStoreClientMock := metrics_mocks.NewMetricStoresClienter(t)
	testSubject := metric_store.NewMetricStoreService(metricStoreDao, tenantDao, metricStoreClientMock)
	return testSubject, &ServiceTestMocks{
		metricStoreDaoMock:    metricStoreDao,
		tenantDaoMock:         tenantDao,
		metricStoreClientMock: metricStoreClientMock,
	}
}
