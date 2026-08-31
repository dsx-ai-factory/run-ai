/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package metric_store_test

import (
	"context"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metrics_api "github.com/run-ai/runai/control-plane/metrics-service/pkg/client/generated"

	"gorm.io/gorm"

	"github.com/run-ai/runai/control-plane/tenants-manager/generated/pkg/api"
	businessErrors "github.com/run-ai/runai/control-plane/tenants-manager/pkg/business/errors"
	metric_store "github.com/run-ai/runai/control-plane/tenants-manager/pkg/business/metric-store"

	"github.com/run-ai/runai/control-plane/tenants-manager/pkg/domain/models"
	"github.com/stretchr/testify/mock"
	"k8s.io/utils/ptr"
)

var _ = Describe("Test grafana lab ", Ordered, func() {
	Context("Test construction", Ordered, func() {
		It("MetricStore exists on tenant", func() {
			metricStore := &models.MetricStoreDetails{
				Read: api.MetricStoreEndpoint{
					Url: "http://read:8080",
					Auth: &api.MetricStoreEndpointAuth{
						Basic: &api.MetricStoreEndpointAuthBasic{
							Username: "a",
							Password: "b",
						},
					},
				},
				Write: api.MetricStoreEndpoint{
					Url: "http://write:8080",
					Auth: &api.MetricStoreEndpointAuth{
						Basic: &api.MetricStoreEndpointAuthBasic{
							Username: "a",
							Password: "b",
						},
					},
				},
			}

			metricStoreService, testMocks := initMocks(GinkgoT())
			testMocks.tenantDaoMock.On("GetById", mock.Anything, mock.Anything).Return(&models.Tenant{MimirId: "1"}, nil)
			apiResponse := &metrics_api.MetricStore{
				TenantId:     1,
				DatasourceId: ptr.To("1"),
				MetricStoreDetails: *metrics_api.NewNullableMetricStoreDetails(&metrics_api.MetricStoreDetails{
					Type: metrics_api.METRICSTORETYPE_THANOS,
					Read: metrics_api.MetricStoreEndpoint{
						Url: "http://read:8080",
						Auth: *metrics_api.NewNullableMetricStoreEndpointAuth(&metrics_api.MetricStoreEndpointAuth{
							Basic: *metrics_api.NewNullableMetricStoreEndpointAuthBasic(&metrics_api.MetricStoreEndpointAuthBasic{
								Username: "a",
								Password: "b",
							}),
						}),
					},
					Write: metrics_api.MetricStoreEndpoint{
						Url: "http://write:8080",
						Auth: *metrics_api.NewNullableMetricStoreEndpointAuth(&metrics_api.MetricStoreEndpointAuth{
							Basic: *metrics_api.NewNullableMetricStoreEndpointAuthBasic(&metrics_api.MetricStoreEndpointAuthBasic{
								Username: "a",
								Password: "b",
							}),
						}),
					},
				}),
			}
			testMocks.metricStoreClientMock.On("CreateMetricStore", mock.Anything, mock.Anything).Return(apiResponse, nil)
			res, err := metricStoreService.CreateMetricsStoreInfra(context.TODO(), &metric_store.MetricStoreInfraCreateRequest{
				TenantId:            1,
				TenantName:          "test",
				ExistingMetricStore: metricStore,
			})
			Expect(err).To(BeNil())
			Expect(res).To(Equal(metricStore))
		})
		It("MetricStore does not exist on tenant happy flow - delegates to metrics-service", func() {
			metricStoreService, testMocks := initMocks(GinkgoT())

			apiResponse := &metrics_api.MetricStore{
				TenantId:     1,
				DatasourceId: ptr.To("runaitest1saasmain"),
				MetricStoreDetails: *metrics_api.NewNullableMetricStoreDetails(&metrics_api.MetricStoreDetails{
					Type: metrics_api.METRICSTORETYPE_GRAFANALABS,
					Read: metrics_api.MetricStoreEndpoint{
						Url: "http://prom:8080/api/prom",
						Auth: *metrics_api.NewNullableMetricStoreEndpointAuth(&metrics_api.MetricStoreEndpointAuth{
							Basic: *metrics_api.NewNullableMetricStoreEndpointAuthBasic(&metrics_api.MetricStoreEndpointAuthBasic{
								Username: "user",
								Password: "read_token",
							}),
						}),
					},
					Write: metrics_api.MetricStoreEndpoint{
						Url: "http://prom:8080/api/prom/push",
						Auth: *metrics_api.NewNullableMetricStoreEndpointAuth(&metrics_api.MetricStoreEndpointAuth{
							Basic: *metrics_api.NewNullableMetricStoreEndpointAuthBasic(&metrics_api.MetricStoreEndpointAuthBasic{
								Username: "user",
								Password: "write_token",
							}),
						}),
					},
				}),
			}
			testMocks.metricStoreClientMock.On("CreateMetricStore", context.TODO(), &metrics_api.MetricStore{TenantId: 1}).Return(apiResponse, nil)

			res, err := metricStoreService.CreateMetricsStoreInfra(context.TODO(), &metric_store.MetricStoreInfraCreateRequest{
				TenantId:   1,
				TenantName: "test",
			})
			Expect(err).To(BeNil())
			Expect(res.Read.Url).To(Equal("http://prom:8080/api/prom"))
			Expect(res.Read.Auth.Basic.Username).To(Equal("user"))
			Expect(res.Read.Auth.Basic.Password).To(Equal("read_token"))
			Expect(res.Write.Url).To(Equal("http://prom:8080/api/prom/push"))
			Expect(res.Write.Auth.Basic.Password).To(Equal("write_token"))
		})
	})

	Context("Error handling", func() {
		It("propagates metrics-service error from CreateMetricStore", func() {
			metricStoreService, testMocks := initMocks(GinkgoT())
			testMocks.metricStoreClientMock.On("CreateMetricStore", context.TODO(), &metrics_api.MetricStore{TenantId: 1}).Return(nil, &metrics_api.Error{Code: http.StatusInternalServerError, Message: "boom"})

			res, err := metricStoreService.CreateMetricsStoreInfra(context.TODO(), &metric_store.MetricStoreInfraCreateRequest{
				TenantId:   1,
				TenantName: "test",
			})
			Expect(err).To(Not(BeNil()))
			Expect(res).To(BeNil())
		})
	})
	Context("Test get metric store", Ordered, func() {
		It("Get metrics store metric store general fetch error", func() {
			// Setup
			metricStoreService, testMocks := initMocks(GinkgoT())
			tenantId := uint(1000)
			ctx := context.Background()
			tenant := &models.Tenant{
				Model: gorm.Model{
					ID: tenantId,
				},
				Name:    "test",
				Smg:     true,
				MimirId: "32",
			}
			testMocks.tenantDaoMock.On("GetById", ctx, tenantId).Return(tenant, nil)
			testMocks.metricStoreClientMock.On("GetMetricStore", ctx, int32(tenantId)).Return(nil, &metrics_api.Error{Code: http.StatusInternalServerError, Message: "general error"})
			response, err := metricStoreService.GetTenantMetricStore(ctx, tenantId)
			Expect(err).To(Not(BeNil()))
			Expect(err.Reason).To(Equal(businessErrors.GeneralError))
			Expect(response).To(BeNil())
		})
		It("Get metrics store metric store return from db", func() {
			// Setup
			metricStoreService, testMocks := initMocks(GinkgoT())
			tenantId := uint(1000)
			ctx := context.Background()
			tenant := &models.Tenant{
				Model: gorm.Model{
					ID: tenantId,
				},
				Name:    "test",
				Smg:     true,
				MimirId: "32",
			}

			testMocks.tenantDaoMock.On("GetById", ctx, tenantId).Return(tenant, nil)
			testMocks.metricStoreClientMock.On("GetMetricStore", ctx, int32(tenantId)).Return(&metrics_api.MetricStore{
				TenantId:     int32(tenantId),
				DatasourceId: ptr.To("test"),
				MetricStoreDetails: *metrics_api.NewNullableMetricStoreDetails(&metrics_api.MetricStoreDetails{
					Read: metrics_api.MetricStoreEndpoint{
						Url: "http://read:8080",
						Auth: *metrics_api.NewNullableMetricStoreEndpointAuth(&metrics_api.MetricStoreEndpointAuth{
							Basic: *metrics_api.NewNullableMetricStoreEndpointAuthBasic(&metrics_api.MetricStoreEndpointAuthBasic{
								Username: "user_read",
								Password: "password",
							}),
						}),
					},
					Write: metrics_api.MetricStoreEndpoint{
						Url: "http://write:8080",
						Auth: *metrics_api.NewNullableMetricStoreEndpointAuth(&metrics_api.MetricStoreEndpointAuth{
							Basic: *metrics_api.NewNullableMetricStoreEndpointAuthBasic(&metrics_api.MetricStoreEndpointAuthBasic{
								Username: "user_write",
								Password: "password",
							}),
						}),
					},
				}),
			}, nil)
			response, err := metricStoreService.GetTenantMetricStore(ctx, tenantId)
			Expect(err).To(BeNil())
			Expect(response.MetricStore.Read.Url).To(Equal("http://read:8080"))
			Expect(response.MetricStore.Read.Auth.Basic.Username).To(Equal("user_read"))
			Expect(response.MetricStore.Write.Url).To(Equal("http://write:8080"))
			Expect(response.MetricStore.Write.Auth.Basic.Username).To(Equal("user_write"))
		})
		It("Get metrics store metric tenant not SMG should return not found", func() {
			// Setup
			metricStoreService, testMocks := initMocks(GinkgoT())
			tenantId := uint(1000)
			ctx := context.Background()
			tenant := &models.Tenant{
				Model: gorm.Model{
					ID: tenantId,
				},
				Name: "test",
				Smg:  false,
			}
			testMocks.tenantDaoMock.On("GetById", ctx, tenantId).Return(tenant, nil)
			testMocks.metricStoreClientMock.On("GetMetricStore", ctx, int32(tenantId)).Return(nil, &metrics_api.Error{Code: http.StatusNotFound, Message: "not found"})
			response, err := metricStoreService.GetTenantMetricStore(ctx, tenantId)
			Expect(err).To(Not(BeNil()))
			Expect(err.Reason).To(Equal(businessErrors.NotFound))
			Expect(response).To(BeNil())
		})
	})
})
