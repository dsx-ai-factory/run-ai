/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services_test

import (
	"context"
	"fmt"

	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/services"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
)

var _ = Describe("ApiClustersService", func() {
	const testSuite = "cluster_histories"
	var (
		clustersHistoriesApiService *services.ClustersHistoriesApiService
		clusterIds                  []string
		ctx                         context.Context
	)

	BeforeEach(func() {
		ctx = context.WithValue(context.Background(), runaictx.CfTenantId, "12345")

		// Setup service
		clustersHistoriesApiService = services.NewClustersHistoriesApiService()
	})

	// Test Get Clusters
	Describe("GetClusterHistories", func() {
		Context("When there are no clusters", func() {
			It("Should return empty list", func() {
				resp, err := clustersHistoriesApiService.GetClustersHistories(ctx,
					[]string{
						"full",
					})
				Expect(err).ShouldNot(HaveOccurred())
				Expect(resp.Code).To(Equal(200))
				Expect(resp.Body).To(BeEmpty())
			})
		})

		Context("When there are clusters", func() {
			const clustersCount = 2

			BeforeEach(func() {
				clusterIds = []string{}
				for i := 0; i < clustersCount; i++ {
					cluster, _ := dao.CreateCluster(ctx, &models.Cluster{
						Name:     fmt.Sprintf(clusterNameGenerator.Generate(testSuite), i),
						Domain:   fmt.Sprintf("cluster-%d", i),
						Version:  fmt.Sprintf("2.1%d", i),
						TenantId: 12345,
					})

					clusterIds = append(clusterIds, cluster.ID.String())
				}
			})
			AfterEach(func() {
				// Returning to zero state
				for _, clusterId := range clusterIds {
					dao.DeleteCluster(ctx, clusterId, 12345)
				}
			})

			It("Should return the clusters histories correctly only not empty versions", func() {
				name := "61ee521f-ed65-4528-8ad1-89158626bc66"

				cluster, _ := dao.CreateCluster(ctx, &models.Cluster{
					Name:     name,
					Domain:   "cluster-10",
					TenantId: 12345,
				})
				dao.UpdateClusterInfo(ctx, &models.Cluster{
					ID:       cluster.ID,
					Name:     name,
					Domain:   "cluster-10",
					Version:  "2.12",
					TenantId: 12345,
				})
				clusterIds = append(clusterIds, cluster.ID.String())

				resp, err := clustersHistoriesApiService.GetClustersHistories(ctx, clusterIds)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(resp.Code).To(Equal(200))
				Expect(resp.Body).To(HaveLen(1))
			})

			It("Should not save cluster history, empty version", func() {
				name := clusterNameGenerator.Generate(testSuite)

				cluster, _ := dao.CreateCluster(ctx, &models.Cluster{
					Name:     name,
					Domain:   "cluster-10",
					TenantId: 12345,
				})
				dao.UpdateClusterInfo(ctx, &models.Cluster{
					ID:       cluster.ID,
					Name:     name,
					Domain:   "cluster-11",
					Version:  "",
					TenantId: 12345,
				})

				resp, err := clustersHistoriesApiService.GetClustersHistories(ctx, []string{name})
				Expect(err).ShouldNot(HaveOccurred())
				Expect(resp.Code).To(Equal(200))
				Expect(resp.Body).To(HaveLen(0))
			})
		})
	})
})
