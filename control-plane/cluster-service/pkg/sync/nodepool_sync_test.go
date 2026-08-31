/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package sync

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/cluster-service/internal/testutils"
	bc_mocks "github.com/run-ai/runai/control-plane/cluster-service/mocks/pkg/clients/backend"
	"github.com/run-ai/runai/control-plane/cluster-service/pkg/config"
	"github.com/run-ai/runai/control-plane/cluster-service/pkg/database/dao"
	"github.com/run-ai/runai/control-plane/cluster-service/pkg/database/models"
	"github.com/run-ai/runai/control-plane/cluster-service/pkg/services"
	org_unit_service_client_mock "github.com/run-ai/runai/runai-clients/org-unit-service-client/mocks"
	runai_config "github.com/run-ai/runai/runai-common-packages/config"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
	"github.com/run-ai/runai/runai-common-packages/runaidb"
	"github.com/stretchr/testify/mock"
	"k8s.io/utils/ptr"
)

func TestDao(t *testing.T) {
	RegisterFailHandler(Fail)
	defer runai_config.FlushAll()

	// Register config types needed for tests
	if err := runai_config.Register[*config.Sync](); err != nil {
		t.Fatalf("Failed to register sync configuration: %v", err)
	}

	testutils.RegisterPostgresSuite(config.Schema, "file://../database/migrations")

	RunSpecs(t, "Nodepool sync tests")
}

var _ = Describe("Nodepool sync tests", Ordered, func() {
	clusterCountPerTenant := 3
	tenantsCount := 3

	// Configure sync settings for tests
	syncConfig := runai_config.Get[*config.Sync]()
	syncConfig.TenantsInChunk = tenantsCount
	syncConfig.IntervalMinutes = 1
	syncConfig.TenantsIntervalSeconds = 5
	runai_config.Set(syncConfig)

	tenantSettingsClient := &bc_mocks.TenantSettingsClient{}
	orgClient := &org_unit_service_client_mock.IOrgUnitClient{}

	ctx := context.Background()
	ctx = context.WithValue(ctx, runaictx.CfSubjectID, "internal-NodePoolsSync")

	orgClient.On("NodePoolCreated", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	orgClient.On("NodePoolDeleted", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	orgClient.On("DefaultNodePoolCreated", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	nodepoolService := services.NewNodePoolsApiService(tenantSettingsClient, orgClient)
	nodepoolSyncService := NewNodePoolsSync(nodepoolService)

	tenantIds := []string{}
	createdClusters := []*models.Cluster{}
	createdDefaultNodepools := []*models.Nodepool{}
	createdNotDefaultNodepools := []*models.Nodepool{}

	BeforeAll(func() {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Starting nodepool sync tests")

		for i := 0; i < tenantsCount; i++ {
			tenantIds = append(tenantIds, fmt.Sprintf("%d", i))
			for j := 0; j < clusterCountPerTenant; j++ {
				created, err := dao.CreateCluster(ctx, &models.Cluster{
					ID:           uuid.New(),
					Name:         fmt.Sprintf("Cluster-%d", j),
					TenantId:     int32(i),
					LastLiveness: ptr.To(time.Now()),
				})

				Expect(err).To(BeNil())
				createdClusters = append(createdClusters, created)

				createdNp, createErr := dao.CreateNodepool(ctx, models.Nodepool{
					Name:        "default",
					TenantId:    int32(i),
					ClusterId:   created.ID,
					ClusterName: created.Name,
					IsDefault:   true,
				})
				Expect(createErr).To(BeNil())
				createdDefaultNodepools = append(createdDefaultNodepools, &createdNp)

				createdNotDefaultNp, createErr := dao.CreateNodepool(ctx, models.Nodepool{
					Name:        "not-default",
					TenantId:    int32(i),
					LabelKey:    ptr.To("a"),
					LabelValue:  ptr.To("b"),
					ClusterId:   created.ID,
					ClusterName: created.Name,
					IsDefault:   false,
				})
				Expect(createErr).To(BeNil())
				createdNotDefaultNodepools = append(createdNotDefaultNodepools, &createdNotDefaultNp)
			}
		}
	})

	AfterAll(func() {
		for _, np := range createdDefaultNodepools {
			_, _ = runaidb.DB().Unscoped().Delete(np)
		}
		createdDefaultNodepools = []*models.Nodepool{}
		for _, np := range createdNotDefaultNodepools {
			_, _ = runaidb.DB().Unscoped().Delete(np)
		}
		createdNotDefaultNodepools = []*models.Nodepool{}

		for _, cluster := range createdClusters {
			_, _ = runaidb.DB().Unscoped().Delete(cluster)
		}
		createdClusters = []*models.Cluster{}
	})

	It("Sanity, all clusters are well configured", func() {
		clusters, err := getClustersWithoutDefaultNodepool(ctx, runaidb.DB(), tenantIds)
		Expect(err).To(BeNil())
		Expect(len(clusters)).To(Equal(0))

		err = nodepoolSyncService.syncClustersDefaultNodepool(ctx)
		Expect(err).To(BeNil())
	})

	It("1 cluster to be fixed, no default nodepool", func() {
		newCluster, apiErr := dao.CreateCluster(ctx, &models.Cluster{
			ID:           uuid.New(),
			Name:         "test-cluster",
			TenantId:     1,
			LastLiveness: ptr.To(time.Now()),
		})
		Expect(apiErr).To(BeNil())
		createdClusters = append(createdClusters, newCluster)

		clusters, err := getClustersWithoutDefaultNodepool(ctx, runaidb.DB(), tenantIds)
		Expect(err).To(BeNil())
		Expect(len(clusters)).To(Equal(1))

		err = nodepoolSyncService.syncClustersDefaultNodepool(ctx)
		Expect(err).To(BeNil())

		createdNp := &models.Nodepool{}
		findErr := runaidb.DB().Table(models.NodepoolTable).
			Where("name = 'default'").
			Where("is_default").
			Where("cluster_id = ?", newCluster.ID.String()).
			First(createdNp)
		Expect(findErr).To(BeNil())
		Expect(createdNp.ID).To(Not(Equal(int32(0))))
		createdDefaultNodepools = append(createdDefaultNodepools, createdNp)
		orgClient.AssertCalled(GinkgoT(), "DefaultNodePoolCreated", mock.Anything, newCluster.ID.String(), fmt.Sprintf("%d", createdNp.ID))

		clusters, err = getClustersWithoutDefaultNodepool(ctx, runaidb.DB(), tenantIds)
		Expect(err).To(BeNil())
		Expect(len(clusters)).To(Equal(0))
	})

	It("1 cluster to be fixed, default nodepool deleted", func() {
		lastCreatedDefaultNodepool := createdDefaultNodepools[len(createdDefaultNodepools)-1]
		err := dao.DeleteNodepool(ctx, lastCreatedDefaultNodepool.TenantId, "", fmt.Sprintf("%d", lastCreatedDefaultNodepool.ID))
		Expect(err).To(BeNil())
		createdDefaultNodepools = createdDefaultNodepools[0 : len(createdDefaultNodepools)-1]

		clusters, err := getClustersWithoutDefaultNodepool(ctx, runaidb.DB(), tenantIds)
		Expect(err).To(BeNil())
		Expect(len(clusters)).To(Equal(1))

		err = nodepoolSyncService.syncClustersDefaultNodepool(ctx)
		Expect(err).To(BeNil())

		createdNp := &models.Nodepool{}
		findErr := runaidb.DB().Table(models.NodepoolTable).
			Where("name = 'default'").
			Where("is_default").
			Where("cluster_id = ?", lastCreatedDefaultNodepool.ClusterId.String()).
			First(createdNp)
		Expect(findErr).To(BeNil())
		Expect(createdNp.ID).To(Not(Equal(int32(0))))
		createdDefaultNodepools = append(createdDefaultNodepools, createdNp)
		orgClient.AssertCalled(GinkgoT(), "DefaultNodePoolCreated", mock.Anything, lastCreatedDefaultNodepool.ClusterId.String(), fmt.Sprintf("%d", createdNp.ID))

		clusters, err = getClustersWithoutDefaultNodepool(ctx, runaidb.DB(), tenantIds)
		Expect(err).To(BeNil())
		Expect(len(clusters)).To(Equal(0))
	})
	It("All clusters need to be fixed", func() {
		for i, np := range createdDefaultNodepools {
			// test both cases of hard and soft delete
			if i%2 == 0 {
				_, delErr := runaidb.DB().Unscoped().Delete(np, np.ID)
				Expect(delErr).To(BeNil())
			} else {
				err := dao.DeleteNodepool(ctx, np.TenantId, "", fmt.Sprintf("%d", np.ID))
				Expect(err).To(BeNil())
			}
		}

		clusters, err := getClustersWithoutDefaultNodepool(ctx, runaidb.DB(), tenantIds)
		Expect(err).To(BeNil())
		Expect(len(clusters)).To(Equal(len(createdClusters))) // for each tenant there is a not connected cluster
		createdDefaultNodepools = []*models.Nodepool{}

		err = nodepoolSyncService.syncClustersDefaultNodepool(ctx)
		Expect(err).To(BeNil())

		for _, cluster := range createdClusters {
			if cluster.LastLiveness != nil {
				orgClient.AssertCalled(GinkgoT(), "DefaultNodePoolCreated", mock.Anything, cluster.ID.String(), mock.Anything)
			}
		}

		clusters, err = getClustersWithoutDefaultNodepool(ctx, runaidb.DB(), tenantIds)
		Expect(err).To(BeNil())
		Expect(len(clusters)).To(Equal(0))
	})
})
