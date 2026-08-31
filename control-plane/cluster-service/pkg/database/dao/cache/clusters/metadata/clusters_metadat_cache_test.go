/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package clusters_metadata_cache

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/cluster-service/pkg/database/models"
	cacheTesting "github.com/run-ai/runai/runai-common-packages/cache/v2/testing"
	"github.com/stretchr/testify/mock"
)

var cacheMock *cacheTesting.MockCacher

func Test(t *testing.T) {
	RegisterFailHandler(Fail)
	cacheMock = cacheTesting.SetMock(t)
	RunSpecs(t, "clusters cache")
}

var _ = Describe("SetClusterHistoryOrInvalidate", Ordered, func() {
	ctx := context.Background()
	const tenant = 1
	cluster1Id, cluster2Id := uuid.New(), uuid.New()
	cluster1History1, cluster1History2 := models.ClusterHistory{
		ClusterID:       cluster1Id,
		TenantId:        tenant,
		Version:         "2.11",
		UpdateTimestamp: time.Now().Add(-48 * time.Hour),
	}, models.ClusterHistory{
		ClusterID:       cluster1Id,
		TenantId:        tenant,
		Version:         "2.12",
		UpdateTimestamp: time.Now(),
	}
	cluster2History1, cluster2History2 := models.ClusterHistory{
		ClusterID:       cluster2Id,
		TenantId:        tenant,
		Version:         "2.13",
		UpdateTimestamp: time.Now().Add(-56 * time.Hour),
	}, models.ClusterHistory{
		ClusterID:       cluster2Id,
		TenantId:        tenant,
		Version:         "2.15",
		UpdateTimestamp: time.Now().Add(-4 * time.Hour),
	}

	_ = Describe("SetClusterHistoryOrInvalidate", Ordered, func() {
		It("not in cache, should do nothing", func() {
			cacheMock.On("Get", mock.Anything, cacheTesting.KeyFromType[models.ClusterHistories](), generateClusterHistoriesKey(tenant, cluster1Id.String())).Return(nil, nil).Once()
			err := SetClusterHistoryOrInvalidate(ctx, tenant, &cluster1History1)
			Expect(err).To(BeNil())
		})

		It("in cache, should add the new history", func() {
			cacheMock.On("Get", mock.Anything, cacheTesting.KeyFromType[models.ClusterHistories](), generateClusterHistoriesKey(tenant, cluster1Id.String())).Return(&models.ClusterHistories{cluster1History1}, nil).Once()
			cacheMock.On("Set", mock.Anything, cacheTesting.KeyFromType[models.ClusterHistories](), generateClusterHistoriesKey(tenant, cluster1Id.String()), mock.Anything).Return(nil).Once()
			err := SetClusterHistoryOrInvalidate(ctx, tenant, &cluster1History2)
			Expect(err).To(BeNil())

			cacheMock.On("Get", mock.Anything, cacheTesting.KeyFromType[models.ClusterHistories](), generateClusterHistoriesKey(tenant, cluster1Id.String())).Return(&models.ClusterHistories{cluster1History1, cluster1History2}, nil).Once()
			data, err := GetAllClusterHistories(ctx, tenant, cluster1History1.ClusterID.String())
			clusterHistories := (*data)[0]
			Expect(err).To(BeNil())
			Expect(len(clusterHistories)).To(Equal(2))
			Expect(clusterHistories[0].ClusterID.String()).To(Equal(cluster1Id.String()))
			Expect(clusterHistories[1].ClusterID.String()).To(Equal(cluster1Id.String()))
		})

		It("wrong format, should throw error", func() {
			cacheMock.On("Get", mock.Anything, cacheTesting.KeyFromType[models.ClusterHistories](), generateClusterHistoriesKey(tenant, cluster1Id.String())).Return("bad format", nil).Once()
			ch, err := GetAllClusterHistories(ctx, tenant, cluster1History1.ClusterID.String())
			Expect(err).NotTo(BeNil())
			Expect(ch).To(BeNil())
		})
	})

	_ = Describe("SetAllClusterHistories", Ordered, func() {
		It("tenant not in cache, should add the new clusters history", func() {
			cacheMock.On("Get", mock.Anything, cacheTesting.KeyFromType[models.ClusterHistories](), generateClusterHistoriesKey(tenant, cluster1Id.String())).Return(nil, nil).Once()

			data, err := GetAllClusterHistories(ctx, tenant, cluster1History1.ClusterID.String())
			Expect(err).To(BeNil())
			Expect(len(*data)).To(Equal(0))
			histories := map[string]models.ClusterHistories{}
			histories[cluster1Id.String()] = models.ClusterHistories{cluster1History1}
			cacheMock.On("Set", mock.Anything, cacheTesting.KeyFromType[models.ClusterHistories](), generateClusterHistoriesKey(tenant, cluster1Id.String()), mock.Anything).Return(nil).Once()
			err = SetAllClusterHistories(ctx, tenant, histories)
			Expect(err).To(BeNil())

			cacheMock.On("Get", mock.Anything, cacheTesting.KeyFromType[models.ClusterHistories](), generateClusterHistoriesKey(tenant, cluster1Id.String())).Return(&models.ClusterHistories{cluster1History1}, nil).Once()
			data, err = GetAllClusterHistories(ctx, tenant, cluster1History1.ClusterID.String())
			Expect(err).To(BeNil())
			Expect(len((*data)[0])).To(Equal(1))
		})

		It("tenant has data in cache, should add the new cluster history", func() {
			histories := map[string]models.ClusterHistories{}
			histories[cluster2Id.String()] = models.ClusterHistories{cluster2History1}
			cacheMock.On("Set", mock.Anything, cacheTesting.KeyFromType[models.ClusterHistories](), generateClusterHistoriesKey(tenant, cluster2Id.String()), mock.Anything).Return(nil).Once()
			err := SetAllClusterHistories(ctx, tenant, histories)
			Expect(err).To(BeNil())

			cacheMock.On("Get", mock.Anything, cacheTesting.KeyFromType[models.ClusterHistories](), generateClusterHistoriesKey(tenant, cluster2Id.String())).Return(&models.ClusterHistories{cluster2History1}, nil).Once()
			data, err := GetAllClusterHistories(ctx, tenant, cluster2Id.String())
			clusterHistories := (*data)[0]
			Expect(err).To(BeNil())
			Expect(len(clusterHistories)).To(Equal(1))

			histories[cluster1Id.String()] = models.ClusterHistories{cluster1History1, cluster1History2}
			histories[cluster2Id.String()] = models.ClusterHistories{cluster2History1, cluster2History2}
			cacheMock.On("Set", mock.Anything, cacheTesting.KeyFromType[models.ClusterHistories](), generateClusterHistoriesKey(tenant, cluster1Id.String()), mock.Anything).Return(nil).Once()
			cacheMock.On("Set", mock.Anything, cacheTesting.KeyFromType[models.ClusterHistories](), generateClusterHistoriesKey(tenant, cluster2Id.String()), mock.Anything).Return(nil).Once()
			err = SetAllClusterHistories(ctx, tenant, histories)
			Expect(err).To(BeNil())

			cacheMock.On("Get", mock.Anything, cacheTesting.KeyFromType[models.ClusterHistories](), generateClusterHistoriesKey(tenant, cluster1Id.String())).Return(&models.ClusterHistories{cluster1History1, cluster1History2}, nil).Once()
			cacheMock.On("Get", mock.Anything, cacheTesting.KeyFromType[models.ClusterHistories](), generateClusterHistoriesKey(tenant, cluster2Id.String())).Return(&models.ClusterHistories{cluster2History1, cluster2History2}, nil).Once()
			data, err = GetAllClusterHistories(ctx, tenant, cluster1History1.ClusterID.String(), cluster2History1.ClusterID.String())
			Expect(err).To(BeNil())
			clusterHistories1 := (*data)[0]
			clusterHistories2 := (*data)[1]
			Expect(len(clusterHistories1)).To(Equal(2))
			Expect(len(clusterHistories2)).To(Equal(2))
		})
	})
})
