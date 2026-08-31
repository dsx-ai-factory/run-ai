/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services_test

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/cluster-service/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/cluster-service/pkg/database/dao"
	clusters_metadata_cache "github.com/run-ai/runai/control-plane/cluster-service/pkg/database/dao/cache/clusters/metadata"
	"github.com/run-ai/runai/control-plane/cluster-service/pkg/database/models"
	"github.com/run-ai/runai/control-plane/cluster-service/pkg/services"
	statuspostupdatehandlers "github.com/run-ai/runai/control-plane/cluster-service/pkg/services/cluster-config/status-post-update-handlers"
	clusterdelete "github.com/run-ai/runai/control-plane/cluster-service/pkg/services/cluster-delete/mocks"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
	"github.com/stretchr/testify/mock"
)

func createTestCluster(ctx context.Context) *models.Cluster {
	clusterName := clusterNameGenerator.Generate("config")
	clusterModel := &models.Cluster{
		Name:     clusterName,
		Domain:   "cluster-11",
		TenantId: tenantId,
	}

	cluster, _ := dao.CreateCluster(ctx, clusterModel)
	return cluster
}

var _ = Describe("ApiControlPlaneConfigService", func() {
	var (
		service        api.ClusterConfigAPIServicer
		ctx            context.Context
		cluster        *models.Cluster
		clusterDeleter *clusterdelete.ClusterDeleter
	)

	const (
		tenantId = 1
	)
	BeforeEach(func() {
		clusterDeleter = clusterdelete.NewClusterDeleter(GinkgoT())
		handler := statuspostupdatehandlers.NewCompositeHandler(clusterDeleter)
		service = services.NewClusterConfigApiService(handler)
		tenantIdStr := strconv.Itoa(tenantId)
		ctx = context.WithValue(context.Background(), runaictx.CfTenantId, tenantIdStr)
		cluster = createTestCluster(ctx)
	})
	AfterEach(func() {
		dao.DeleteCluster(ctx, cluster.ID.String(), tenantId)
	})
	Describe("Get/Set config tests", func() {
		It("test config defaults", func() {
			_, err := dao.SetClusterConfig(ctx, tenantId, cluster.ID.String(), &models.ClusterConfigData{
				SystemConfig: api.ClusterSystemConfigData{
					Deactivated: false,
				},
				UserConfig: api.ClusterUserConfigData{},
			})
			Expect(err).Should(BeNil())
			config, err := service.GetClusterConfig(ctx, cluster.ID.String())
			Expect(config).ShouldNot(BeNil())
			Expect(err).Should(BeNil())
			responseBody := config.Body.(api.GetPutPatchClusterConfigResponse)
			Expect(responseBody.ClusterId).Should(Equal(cluster.ID.String()))
			Expect(responseBody.TenantId).Should(Equal(int32(tenantId)))
			Expect(responseBody.ClusterId).Should(Equal(cluster.ID.String()))
			Expect(responseBody.Status.State).Should(Equal(string(models.ClusterStateApplied)))
		})
		It("should return a valid internal config", func() {
			_, err := dao.SetClusterConfig(ctx, tenantId, cluster.ID.String(), &models.ClusterConfigData{
				SystemConfig: api.ClusterSystemConfigData{
					Deactivated: true,
				},
				UserConfig: api.ClusterUserConfigData{},
			})
			Expect(err).Should(BeNil())
			systemConfig, err := service.GetInternalClusterConfig(ctx, cluster.ID.String())
			Expect(systemConfig).ShouldNot(BeNil())
			Expect(err).Should(BeNil())
			responseBody := systemConfig.Body.(api.GetInternalClusterConfigResponse)
			Expect(responseBody.ClusterId).Should(Equal(cluster.ID.String()))
			Expect(responseBody.TenantId).Should(Equal(int32(tenantId)))
			Expect(responseBody.ClusterId).Should(Equal(cluster.ID.String()))
			Expect(responseBody.Config.SystemConfig.Deactivated).Should(BeTrue())
			Expect(responseBody.Config.UserConfig).Should(Equal(api.ClusterUserConfigData{}))

			userConf, err := service.GetClusterConfig(ctx, cluster.ID.String())
			Expect(systemConfig).ShouldNot(BeNil())
			Expect(err).Should(BeNil())
			userResponseBody := userConf.Body.(api.GetPutPatchClusterConfigResponse)
			Expect(userResponseBody.ClusterId).Should(Equal(cluster.ID.String()))
			Expect(userResponseBody.TenantId).Should(Equal(int32(tenantId)))
			Expect(userResponseBody.ClusterId).Should(Equal(cluster.ID.String()))
			Expect(userResponseBody.Config).Should(Equal(api.ClusterUserConfigData{}))
		})
		It("should return a valid user config when Cluster doesn't have a config", func() {
			config, err := service.GetClusterConfig(ctx, cluster.ID.String())
			Expect(config).ShouldNot(BeNil())
			Expect(err).Should(BeNil())
			responseBody := config.Body.(api.GetPutPatchClusterConfigResponse)
			Expect(responseBody.ClusterId).Should(Equal(cluster.ID.String()))
			Expect(responseBody.TenantId).Should(Equal(int32(tenantId)))
			Expect(responseBody.ClusterId).Should(Equal(cluster.ID.String()))
			Expect(responseBody.Config).Should(Equal(api.ClusterUserConfigData{}))
			Expect(responseBody.Status.ActualConfig).Should(Equal(api.ClusterUserConfigData{}))
			Expect(responseBody.Status.State).Should(Equal(string(models.ClusterStateApplied)))
			Expect(responseBody.Status.StateMessage).Should(Equal(string(models.ClusterStateMessageApplied)))
		})

		It("should return a user config", func() {
			_, err := dao.SetClusterConfig(ctx, tenantId, cluster.ID.String(), &models.ClusterConfigData{
				SystemConfig: api.ClusterSystemConfigData{
					Deactivated: true,
				},
				UserConfig: api.ClusterUserConfigData{},
			})

			userConf, err := service.GetClusterConfig(ctx, cluster.ID.String())
			Expect(userConf).ShouldNot(BeNil())
			Expect(err).Should(BeNil())
			userResponseBody := userConf.Body.(api.GetPutPatchClusterConfigResponse)
			Expect(userResponseBody.ClusterId).Should(Equal(cluster.ID.String()))
			Expect(userResponseBody.TenantId).Should(Equal(int32(tenantId)))
			Expect(userResponseBody.ClusterId).Should(Equal(cluster.ID.String()))
			Expect(userResponseBody.Config).Should(Equal(api.ClusterUserConfigData{}))
			Expect(userResponseBody.Status.ActualConfig).Should(Equal(api.ClusterUserConfigData{}))
		})

		// The test below needs to be re-written when user config content is implemented
		It("should update a user config", func() {
			_, err := dao.SetClusterConfig(ctx, tenantId, cluster.ID.String(), &models.ClusterConfigData{
				SystemConfig: api.ClusterSystemConfigData{
					Deactivated: true,
				},
				UserConfig: api.ClusterUserConfigData{},
			})
			clusterConfigPutRequest := api.PutClusterConfigRequest{
				Config: api.ClusterUserConfigData{},
			}
			userConf, err := service.PutClusterConfig(ctx, cluster.ID.String(), clusterConfigPutRequest)
			Expect(userConf).ShouldNot(BeNil())
			Expect(err).Should(BeNil())
			userResponseBody := userConf.Body.(api.GetPutPatchClusterConfigResponse)
			Expect(userResponseBody.ClusterId).Should(Equal(cluster.ID.String()))
			Expect(userResponseBody.TenantId).Should(Equal(int32(tenantId)))
			Expect(userResponseBody.ClusterId).Should(Equal(cluster.ID.String()))
			Expect(userResponseBody.Config).Should(Equal(api.ClusterUserConfigData{}))
		})
		// The test below needs to be re-written when user config content is implemented
		It("should update a user config partially", func() {
			_, err := dao.SetClusterConfig(ctx, tenantId, cluster.ID.String(), &models.ClusterConfigData{
				SystemConfig: api.ClusterSystemConfigData{
					Deactivated: true,
				},
				UserConfig: api.ClusterUserConfigData{},
			})
			clusterConfigPatchRequest := api.PatchClusterConfigRequest{
				Config: api.ClusterUserConfigDataPartialUpdate{},
			}
			userConf, err := service.PatchClusterConfig(ctx, cluster.ID.String(), clusterConfigPatchRequest)
			Expect(userConf).ShouldNot(BeNil())
			Expect(err).Should(BeNil())
			userResponseBody := userConf.Body.(api.GetPutPatchClusterConfigResponse)
			Expect(userResponseBody.ClusterId).Should(Equal(cluster.ID.String()))
			Expect(userResponseBody.TenantId).Should(Equal(int32(tenantId)))
			Expect(userResponseBody.ClusterId).Should(Equal(cluster.ID.String()))
			Expect(userResponseBody.Config).Should(Equal(api.ClusterUserConfigData{}))
		})
	})
	Describe("Set cluster status tests", func() {
		It("Get config fully applied state reported", func() {
			called := make(chan struct{}, 1)
			clusterDeleter.On("ForceDeleteCluster", mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
				called <- struct{}{}
			})
			_, err := dao.SetClusterConfig(ctx, tenantId, cluster.ID.String(), &models.ClusterConfigData{
				SystemConfig: api.ClusterSystemConfigData{
					Deactivated: true,
				},
				UserConfig: api.ClusterUserConfigData{},
			})
			statusRequest := api.PutClusterConfigStatusRequest{
				Status: api.PutClusterConfigStatusRequestStatus{
					ActualConfig: api.ClusterFullConfigData{
						SystemConfig: api.ClusterSystemConfigData{
							Deactivated: true,
						},
						UserConfig: api.ClusterUserConfigData{},
					},
					Conditions: []api.Condition{
						{
							Type:               "Applied",
							Status:             "True",
							LastTransitionTime: time.Now().Add(-2 * time.Minute),
						},
						{
							Type:               "Applying",
							Status:             "True",
							LastTransitionTime: time.Now().Add(-3 * time.Minute),
						},
					},
				},
			}
			resp, err := service.PutClusterConfigStatus(ctx, cluster.ID.String(), statusRequest)
			// Simulate DAO error
			Expect(err).Should(BeNil())
			Expect(resp.Code).Should(Equal(http.StatusNoContent))

			userConf, err := service.GetClusterConfig(ctx, cluster.ID.String())
			Expect(err).Should(BeNil())
			userResponseBody := userConf.Body.(api.GetPutPatchClusterConfigResponse)
			Expect(userResponseBody.Status.ActualConfig).Should(Equal(api.ClusterUserConfigData{}))
			Expect(userResponseBody.Status.State).Should(Equal(string(models.ClusterStateApplied)))
			Expect(userResponseBody.Status.StateMessage).Should(Equal(string(models.ClusterStateMessageApplied)))
			Eventually(called, 5*time.Second, time.Second).Should(Receive(), "Force delete was called since cluster in deactivated state")
		})
		It("Get Config failed state reported", func() {
			clusterDeleter.On("ForceDeleteCluster", mock.Anything, mock.Anything).Return(nil).Maybe()
			_, err := dao.SetClusterConfig(ctx, tenantId, cluster.ID.String(), &models.ClusterConfigData{
				SystemConfig: api.ClusterSystemConfigData{
					Deactivated: true,
				},
				UserConfig: api.ClusterUserConfigData{},
			})
			ctx := context.WithValue(context.Background(), runaictx.CfTenantId, strconv.Itoa(tenantId))
			statusRequest := api.PutClusterConfigStatusRequest{
				Status: api.PutClusterConfigStatusRequestStatus{
					ActualConfig: api.ClusterFullConfigData{
						SystemConfig: api.ClusterSystemConfigData{
							Deactivated: true,
						},
						UserConfig: api.ClusterUserConfigData{},
					},
					Conditions: []api.Condition{
						{
							Type:               "Applied",
							Status:             "False",
							LastTransitionTime: time.Now().Add(-2 * time.Minute),
						},
						{
							Type:               "Applying",
							Status:             "False",
							LastTransitionTime: time.Now().Add(-2 * time.Minute),
						},
					},
				},
			}
			resp, err := service.PutClusterConfigStatus(ctx, cluster.ID.String(), statusRequest)
			// Simulate DAO error
			Expect(err).Should(BeNil())
			Expect(resp.Code).Should(Equal(http.StatusNoContent))

			userConf, err := service.GetClusterConfig(ctx, cluster.ID.String())
			Expect(err).Should(BeNil())
			userResponseBody := userConf.Body.(api.GetPutPatchClusterConfigResponse)
			Expect(userResponseBody.Status.ActualConfig).Should(Equal(api.ClusterUserConfigData{}))
			Expect(userResponseBody.Status.State).Should(Equal(string(models.ClusterStateFailed)))
			Expect(userResponseBody.Status.StateMessage).Should(Equal(fmt.Sprintf(string(models.ClusterStateMessageFailed), "")))
		})
		It("Get config only applying state reported", func() {
			clusterDeleter.On("ForceDeleteCluster", mock.Anything, mock.Anything).Return(nil).Maybe()
			_, err := dao.SetClusterConfig(ctx, tenantId, cluster.ID.String(), &models.ClusterConfigData{
				SystemConfig: api.ClusterSystemConfigData{
					Deactivated: true,
				},
				UserConfig: api.ClusterUserConfigData{},
			})
			userConf1, err := service.GetClusterConfig(ctx, cluster.ID.String())
			Expect(err).Should(BeNil())

			ctx := context.WithValue(context.Background(), runaictx.CfTenantId, strconv.Itoa(tenantId))
			statusRequest := api.PutClusterConfigStatusRequest{
				Status: api.PutClusterConfigStatusRequestStatus{
					ActualConfig: api.ClusterFullConfigData{
						SystemConfig: api.ClusterSystemConfigData{
							Deactivated: true,
						},
						UserConfig: api.ClusterUserConfigData{},
					},
					Conditions: []api.Condition{
						{
							Type:               "Applying",
							Status:             "True",
							LastTransitionTime: time.Now().Add(-2 * time.Minute),
						},
						{
							Type:               "Applied",
							Status:             "False",
							LastTransitionTime: time.Now().Add(-2 * time.Minute),
						},
					},
				},
			}
			resp, err := service.PutClusterConfigStatus(ctx, cluster.ID.String(), statusRequest)
			Expect(err).Should(BeNil())
			Expect(resp.Code).Should(Equal(http.StatusNoContent))

			userConf, err := service.GetClusterConfig(ctx, cluster.ID.String())
			Expect(err).Should(BeNil())
			userResponseBody := userConf.Body.(api.GetPutPatchClusterConfigResponse)
			userResponseBody1 := userConf1.Body.(api.GetPutPatchClusterConfigResponse)
			// Check that status update is not changing updated_at
			Expect(userResponseBody.UpdatedAt).Should(Equal(userResponseBody1.UpdatedAt))
			Expect(userResponseBody.Status.ActualConfig).Should(Equal(api.ClusterUserConfigData{}))
			Expect(userResponseBody.Status.State).Should(Equal(string(models.ClusterStateApplying)))
			Expect(userResponseBody.Status.StateMessage).Should(Equal(string(models.ClusterStateMessageApplying)))
		})
		It("Should invalidate cache when updating cluster config status", func() {
			clusterDeleter.On("ForceDeleteCluster", mock.Anything, mock.Anything).Return(nil).Maybe()

			// First, populate the cache by getting cluster metadata
			emptyGetter := func() (*models.Cluster, error) {
				return nil, nil
			}
			_, err := clusters_metadata_cache.GetOrSetClusterMetadata(ctx, tenantId, cluster.ID.String(), func() (*models.Cluster, error) {
				return cluster, nil
			})
			Expect(err).Should(BeNil())

			// Verify cache is populated
			cachedCluster, _ := clusters_metadata_cache.GetOrSetClusterMetadata(ctx, tenantId, cluster.ID.String(), emptyGetter)
			Expect(cachedCluster).ShouldNot(BeNil())

			// Update cluster config status
			statusRequest := api.PutClusterConfigStatusRequest{
				Status: api.PutClusterConfigStatusRequestStatus{
					ActualConfig: api.ClusterFullConfigData{
						SystemConfig: api.ClusterSystemConfigData{
							Deactivated: false,
						},
						UserConfig: api.ClusterUserConfigData{},
					},
					Conditions: []api.Condition{
						{
							Type:               "Applied",
							Status:             "True",
							LastTransitionTime: time.Now(),
						},
					},
				},
			}
			resp, err := service.PutClusterConfigStatus(ctx, cluster.ID.String(), statusRequest)
			Expect(err).Should(BeNil())
			Expect(resp.Code).Should(Equal(http.StatusNoContent))

			// Verify cache was invalidated - should return nil with empty getter
			cachedCluster, _ = clusters_metadata_cache.GetOrSetClusterMetadata(ctx, tenantId, cluster.ID.String(), emptyGetter)
			Expect(cachedCluster).Should(BeNil())
		})
	})
})
