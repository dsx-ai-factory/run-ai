/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package test

import (
	"context"

	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/services"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Cluster Histories DB tests", Ordered, func() {
	var clusterHistories *services.ClustersHistoriesApiService = services.NewClustersHistoriesApiService()
	ctxForHistories := context.WithValue(context.Background(), runaictx.CfTenantId, "12345")
	Context("test nodes service and db", func() {
		It("test create/get cluster histories/s", func() {
			cluster, _ := dao.CreateCluster(ctx, &models.Cluster{
				Common:   models.Common{},
				Name:     "cluster-histories-1",
				Version:  "2.11",
				TenantId: 12345,
			})
			_ = dao.UpdateClusterInfo(ctx, cluster)
			res, err := clusterHistories.GetClustersHistories(ctxForHistories, []string{cluster.ID.String()})
			body := (res.Body.([]api.ClusterHistory))[0]
			Expect(err, nil)
			Expect(res.Code, 200)
			Expect(body.ClusterId).To(Equal(cluster.ID.String()))
			Expect(body.Version).To(Equal(cluster.Version))
		})

		It("test update cluster will create new cluster history", func() {
			cluster, _ := dao.CreateCluster(ctx, &models.Cluster{
				Common:   models.Common{},
				Name:     "cluster-histories-2",
				Version:  "2.11",
				TenantId: 12345,
			})
			_ = dao.UpdateClusterInfo(ctx, cluster)
			cluster.Version = "2.12"
			_ = dao.UpdateClusterInfo(ctx, cluster)

			res, err := clusterHistories.GetClustersHistories(ctxForHistories, []string{cluster.ID.String()})
			body := (res.Body.([]api.ClusterHistory))
			Expect(err, nil)
			Expect(res.Code, 200)
			Expect(body[0].ClusterId).To(Equal(cluster.ID.String()))
			Expect(body[0].Version).To(Equal("2.11"))

			Expect(body[1].ClusterId).To(Equal(cluster.ID.String()))
			Expect(body[1].Version).To(Equal("2.12"))
		})
	})
})
