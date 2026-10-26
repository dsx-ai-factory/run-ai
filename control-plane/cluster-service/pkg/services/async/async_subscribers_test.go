/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package async

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/lib/pq"
	"github.com/stretchr/testify/mock"

	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/internal/testutils"
	extapi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/database/models"
	clusterdelete "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/services/cluster-delete"
	events "github.com/dsx-ai-factory/run-ai/runai-application-packages/async-events"
	runai_config "github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	pubsub "github.com/dsx-ai-factory/run-ai/runai-common-packages/runaipubsub"
	pubsubmocks "github.com/dsx-ai-factory/run-ai/runai-common-packages/runaipubsub/mocks"
	uuid2 "github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	authorization_client "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client"
	authorizationclient "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client"
	authorizationapi "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	authorizationUtils "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/helpers"
	orgunitclient "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client/mocks"
)

const (
	tenantId = 1
)

type AsyncMocks struct {
	orgUnitClient       *orgunitclient.IOrgUnitClient
	authorizationClient authorizationclient.IAuthorizationClient
	pubsub              *pubsubmocks.PubSuber
}

func initSubscriber() (*ClusterDeleteSubscriber, *TenantDeleteSubscriber, *AsyncMocks) {
	mocks := &AsyncMocks{
		authorizationClient: authorizationclient.NewAuthorizationClient(),
		orgUnitClient:       &orgunitclient.IOrgUnitClient{},
		pubsub:              &pubsubmocks.PubSuber{},
	}
	pubsub.SetMock(mocks.pubsub)

	serviceConfig := runai_config.Get[*config.Service]()
	serviceConfig.PubsubInitialized = true
	runai_config.Set(serviceConfig)
	deleter := clusterdelete.NewClusterDeleter(mocks.authorizationClient, mocks.orgUnitClient, nil)
	return NewClusterDeleteSubscriber(deleter), NewTenantDeleteSubscriber(deleter), mocks
}

func TestAsync(t *testing.T) {
	RegisterFailHandler(Fail)
	defer runai_config.FlushAll()
	// Register config types needed for tests
	_ = runai_config.Register[*config.Service]()

	_ = runai_config.Register[*authorization_client.Config]()

	testutils.RegisterPostgresSuite(config.Schema, "file://../../database/migrations")

	RunSpecs(t, "Cluster cleaner Suite")
}

var _ = Describe("Subscribers", func() {
	var (
		ctx                     context.Context
		clusterDeleteSubscriber *ClusterDeleteSubscriber
		tenantsSubscriber       *TenantDeleteSubscriber
		pubSubMock              *AsyncMocks
	)
	BeforeEach(func() {
		tenantIdStr := strconv.Itoa(tenantId)
		ctx = context.WithValue(context.Background(), runaictx.CfTenantId, tenantIdStr)
		ctx = context.WithValue(ctx, authorizationUtils.PermittedScopesKey, &authorizationapi.PermittedScopes{
			Tenants: []string{tenantIdStr},
		})
		clusterDeleteSubscriber, tenantsSubscriber, pubSubMock = initSubscriber()
		pubSubMock.pubsub.On("Publish", mock.Anything, mock.Anything).Return(nil)
		pubSubMock.pubsub.On("SubscribeAsync", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
		serviceConfig := runai_config.Get[*config.Service]()
		serviceConfig.PubsubInitialized = true
		runai_config.Set(serviceConfig)
	})
	Describe("ClusterCallBack test", func() {
		It("happy flow", func() {
			cluster1Name := "clusterForceNewVer"
			now := time.Now()
			duration, _ := time.ParseDuration("-1.5h")
			deleteTimestamp := now.Add(duration)
			cluster1Model := &models.Cluster{
				Name:              cluster1Name,
				Domain:            "cluster-11",
				Version:           "2.22",
				TenantId:          tenantId,
				DeleteRequestedAt: &deleteTimestamp,
			}

			cluster1, _ := dao.CreateCluster(ctx, cluster1Model)
			msg := events.NewClusterDelete(cluster1.ID.String())
			err := clusterDeleteSubscriber.clusterDeleteSubscribeHandler(ctx, msg)
			Expect(err).Should(BeNil())
			cluster1, apiErr := dao.GetClusterByIdAndTenantId(ctx, cluster1.ID.String(), tenantId, extapi.VerbosityMetadata)
			Expect(apiErr.Code).Should(Equal(int32(http.StatusNotFound)))
			Expect(cluster1).Should(BeNil())

			clusterSoftDeleteValidation(cluster1Model.ID.String())
		})
		It("Delete callback cluster doesn't exist", func() {
			uuid, _ := uuid2.NewUUID()
			msg := events.NewClusterDelete(uuid.String())
			err := clusterDeleteSubscriber.clusterDeleteSubscribeHandler(ctx, msg)
			Expect(err).Should(BeNil())
		})
	})
	Describe("TenantsDeleteCallBack test", func() {
		Context("Happy Path Scenarios", func() {
			It("should successfully delete all tenant clusters and nodes", func() {
				cluster1 := &models.Cluster{
					Name:     "test-cluster-1",
					Domain:   "cluster-1.test",
					Version:  "2.22",
					TenantId: tenantId,
				}
				cluster2 := &models.Cluster{
					Name:     "test-cluster-2",
					Domain:   "cluster-2.test",
					Version:  "2.22",
					TenantId: tenantId,
				}

				createdCluster1, _ := dao.CreateCluster(ctx, cluster1)
				createdCluster2, _ := dao.CreateCluster(ctx, cluster2)

				networkTopologyLabelValues := []string{"zone-1", "region-1", "hostname-1"}
				networkTopologyLabelValuesStringArray := pq.StringArray(networkTopologyLabelValues)
				// Create test nodes for the tenant
				node1 := &models.Node{
					Name:                       "node-1",
					ClusterUuid:                createdCluster1.ID.String(),
					TenantId:                   tenantId,
					Status:                     "Ready",
					GpuCount:                   2,
					NetworkTopologyLabelValues: &networkTopologyLabelValuesStringArray,
				}
				node2 := &models.Node{
					Name:        "node-2",
					ClusterUuid: createdCluster2.ID.String(),
					TenantId:    tenantId,
					Status:      "Ready",
					GpuCount:    1,
				}

				response, _ := dao.UpsertNode(ctx, node1)
				Expect(response.PreviousValues.NetworkTopologyLabelValues).Should(BeNil())
				response, _ = dao.UpsertNode(ctx, node1)
				Expect(response.PreviousValues.NetworkTopologyLabelValues).Should(Not(BeNil()))
				Expect([]string(*response.PreviousValues.NetworkTopologyLabelValues)).Should(Equal(networkTopologyLabelValues))
				dao.UpsertNode(ctx, node2)

				// Verify clusters and nodes exist before deletion
				permittedScopes := openapi.PermittedScopes{
					Tenants: []string{strconv.Itoa(tenantId)},
				}
				clusters, _ := dao.GetClusters(ctx, tenantId, extapi.VerbosityMetadata, permittedScopes, nil, false, false, 0, 0, nil)
				Expect(len(*clusters)).Should(Equal(2))

				nodes, _ := dao.GetNodes(ctx, tenantId, &permittedScopes, nil, "", "", 0, 100, "")
				Expect(len(*nodes)).Should(Equal(2))

				// Trigger tenant deletion
				msg := events.NewTenantDelete(strconv.Itoa(tenantId))
				err := tenantsSubscriber.TenantDeleteSubscribeHandler(ctx, msg)
				Expect(err).Should(BeNil())

				// Verify all clusters are deleted
				clusters, _ = dao.GetClusters(ctx, tenantId, extapi.VerbosityMetadata, permittedScopes, nil, false, false, 0, 0, nil)
				Expect(len(*clusters)).Should(Equal(0))

				// Verify all nodes are deleted
				nodes, _ = dao.GetNodes(ctx, tenantId, &permittedScopes, nil, "", "", 0, 100, "")
				Expect(len(*nodes)).Should(Equal(0))
			})

			It("should handle tenant with no clusters or nodes", func() {
				// Create a tenant with no clusters or nodes
				emptyTenantId := 999
				tenantIdStr := strconv.Itoa(emptyTenantId)
				ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantIdStr)
				ctx = context.WithValue(ctx, authorizationUtils.PermittedScopesKey, &authorizationapi.PermittedScopes{
					Tenants: []string{tenantIdStr},
				})

				// Trigger tenant deletion
				msg := events.NewTenantDelete(strconv.Itoa(emptyTenantId))
				err := tenantsSubscriber.TenantDeleteSubscribeHandler(ctx, msg)
				Expect(err).Should(BeNil())

				// Verify no errors occurred
				permittedScopes := openapi.PermittedScopes{
					Tenants: []string{tenantIdStr},
				}
				clusters, _ := dao.GetClusters(ctx, int32(emptyTenantId), extapi.VerbosityMetadata, permittedScopes, nil, false, false, 0, 0, nil)
				Expect(len(*clusters)).Should(Equal(0))
			})
		})

		Context("Error Scenarios", func() {
			It("should handle invalid tenant ID format", func() {
				msg := events.NewTenantDelete("invalid-tenant-id")
				err := tenantsSubscriber.TenantDeleteSubscribeHandler(ctx, msg)
				Expect(err).ShouldNot(BeNil())
				Expect(err.Error()).Should(ContainSubstring("failed to convert tenant id to int"))
			})

			It("should handle empty tenant ID", func() {
				msg := events.NewTenantDelete("")
				err := tenantsSubscriber.TenantDeleteSubscribeHandler(ctx, msg)
				Expect(err).ShouldNot(BeNil())
				Expect(err.Error()).Should(ContainSubstring("failed to convert tenant id to int"))
			})

			It("should handle non-numeric tenant ID", func() {
				msg := events.NewTenantDelete("abc123")
				err := tenantsSubscriber.TenantDeleteSubscribeHandler(ctx, msg)
				Expect(err).ShouldNot(BeNil())
				Expect(err.Error()).Should(ContainSubstring("failed to convert tenant id to int"))
			})
		})

		Context("Edge Cases", func() {
			It("should handle tenant with only clusters but no nodes", func() {
				// Create only clusters without nodes
				cluster := &models.Cluster{
					Name:     "cluster-only",
					Domain:   "cluster-only.test",
					Version:  "2.22",
					TenantId: tenantId,
				}
				dao.CreateCluster(ctx, cluster)

				// Verify cluster exists
				permittedScopes := openapi.PermittedScopes{
					Tenants: []string{strconv.Itoa(tenantId)},
				}
				clusters, _ := dao.GetClusters(ctx, tenantId, extapi.VerbosityMetadata, permittedScopes, nil, false, false, 0, 0, nil)
				Expect(len(*clusters)).Should(Equal(1))

				// Trigger tenant deletion
				msg := events.NewTenantDelete(strconv.Itoa(tenantId))
				err := tenantsSubscriber.TenantDeleteSubscribeHandler(ctx, msg)
				Expect(err).Should(BeNil())

				// Verify cluster is deleted
				clusters, _ = dao.GetClusters(ctx, tenantId, extapi.VerbosityMetadata, permittedScopes, nil, false, false, 0, 0, nil)
				Expect(len(*clusters)).Should(Equal(0))
			})

			It("should handle tenant with only nodes but no clusters", func() {
				// Create a cluster first to get a valid cluster UUID
				cluster := &models.Cluster{
					Name:     "temp-cluster",
					Domain:   "temp-cluster.test",
					Version:  "2.22",
					TenantId: tenantId,
				}
				createdCluster, _ := dao.CreateCluster(ctx, cluster)

				// Create nodes
				node := &models.Node{
					Name:        "orphan-node",
					ClusterUuid: createdCluster.ID.String(),
					TenantId:    tenantId,
					Status:      "Ready",
					GpuCount:    1,
				}
				dao.UpsertNode(ctx, node)

				// Delete the cluster but leave the node
				dao.DeleteCluster(ctx, createdCluster.ID.String(), tenantId)

				// Verify node exists but cluster doesn't
				permittedScopes := openapi.PermittedScopes{
					Tenants: []string{strconv.Itoa(tenantId)},
				}
				clusters, _ := dao.GetClusters(ctx, tenantId, extapi.VerbosityMetadata, permittedScopes, nil, false, false, 0, 0, nil)
				Expect(len(*clusters)).Should(Equal(0))

				nodes, _ := dao.GetNodes(ctx, tenantId, &permittedScopes, nil, "", "", 0, 100, "")
				Expect(len(*nodes)).Should(Equal(1))

				// Trigger tenant deletion
				msg := events.NewTenantDelete(strconv.Itoa(tenantId))
				err := tenantsSubscriber.TenantDeleteSubscribeHandler(ctx, msg)
				Expect(err).Should(BeNil())

				// Verify node is deleted
				nodes, _ = dao.GetNodes(ctx, tenantId, &permittedScopes, nil, "", "", 0, 100, "")
				Expect(len(*nodes)).Should(Equal(0))
			})

			It("should handle large number of clusters and nodes", func() {
				// Create multiple clusters
				var clusterUuids []uuid2.UUID
				for i := 0; i < 5; i++ {
					cluster := &models.Cluster{
						Name:     fmt.Sprintf("cluster-%d", i),
						Domain:   fmt.Sprintf("cluster-%d.test", i),
						Version:  "2.22",
						TenantId: tenantId,
					}
					createdCluster, _ := dao.CreateCluster(ctx, cluster)
					clusterUuids = append(clusterUuids, createdCluster.ID)
				}

				// Create multiple nodes across clusters
				for i, clusterUuid := range clusterUuids {
					for j := 0; j < 3; j++ {
						node := &models.Node{
							Name:        fmt.Sprintf("node-%d-%d", i, j),
							ClusterUuid: clusterUuid.String(),
							TenantId:    tenantId,
							Status:      "Ready",
							GpuCount:    int32(j + 1),
						}
						dao.UpsertNode(ctx, node)
					}
				}

				// Verify data exists
				permittedScopes := openapi.PermittedScopes{
					Tenants: []string{strconv.Itoa(tenantId)},
				}
				clusters, _ := dao.GetClusters(ctx, tenantId, extapi.VerbosityMetadata, permittedScopes, nil, false, false, 0, 0, nil)
				Expect(len(*clusters)).Should(Equal(5))

				nodes, _ := dao.GetNodes(ctx, tenantId, &permittedScopes, nil, "", "", 0, 100, "")
				Expect(len(*nodes)).Should(Equal(15))

				// Trigger tenant deletion
				msg := events.NewTenantDelete(strconv.Itoa(tenantId))
				err := tenantsSubscriber.TenantDeleteSubscribeHandler(ctx, msg)
				Expect(err).Should(BeNil())

				// Verify all data is deleted
				clusters, _ = dao.GetClusters(ctx, tenantId, extapi.VerbosityMetadata, permittedScopes, nil, false, false, 0, 0, nil)
				Expect(len(*clusters)).Should(Equal(0))

				nodes, _ = dao.GetNodes(ctx, tenantId, &permittedScopes, nil, "", "", 0, 100, "")
				Expect(len(*nodes)).Should(Equal(0))
			})
		})

		Context("Data Integrity", func() {
			It("should not affect other tenants' data", func() {
				otherTenantId := 456

				// Create data for other tenant
				otherCluster := &models.Cluster{
					Name:     "other-tenant-cluster",
					Domain:   "other-tenant.test",
					Version:  "2.22",
					TenantId: int32(otherTenantId),
				}
				createdOtherCluster, _ := dao.CreateCluster(ctx, otherCluster)

				otherNode := &models.Node{
					Name:        "other-tenant-node",
					ClusterUuid: createdOtherCluster.ID.String(),
					TenantId:    int32(otherTenantId),
					Status:      "Ready",
					GpuCount:    1,
				}
				dao.UpsertNode(ctx, otherNode)

				// Create data for test tenant
				testCluster := &models.Cluster{
					Name:     "test-tenant-cluster",
					Domain:   "test-tenant.test",
					Version:  "2.22",
					TenantId: tenantId,
				}
				dao.CreateCluster(ctx, testCluster)

				// Trigger deletion for test tenant
				msg := events.NewTenantDelete(strconv.Itoa(tenantId))
				err := tenantsSubscriber.TenantDeleteSubscribeHandler(ctx, msg)
				Expect(err).Should(BeNil())

				// Verify other tenant's data is still intact
				otherTenantIdStr := strconv.Itoa(otherTenantId)
				otherPermittedScopes := openapi.PermittedScopes{
					Tenants: []string{otherTenantIdStr},
				}
				otherClusters, _ := dao.GetClusters(ctx, int32(otherTenantId), extapi.VerbosityMetadata, otherPermittedScopes, nil, false, false, 0, 0, nil)
				Expect(len(*otherClusters)).Should(Equal(1))

				otherNodes, _ := dao.GetNodes(ctx, int32(otherTenantId), &otherPermittedScopes, nil, "", "", 0, 100, "")
				Expect(len(*otherNodes)).Should(Equal(1))
			})

			It("should delete only targeted tenant and preserve other tenant with multiple resources", func() {
				// Create two tenants with multiple resources each
				tenant1Id := 100
				tenant2Id := 200

				// Create multiple clusters and nodes for tenant 1
				tenant1Cluster1 := &models.Cluster{
					Name:     "tenant1-cluster1",
					Domain:   "tenant1-cluster1.test",
					Version:  "2.22",
					TenantId: int32(tenant1Id),
				}
				tenant1Cluster2 := &models.Cluster{
					Name:     "tenant1-cluster2",
					Domain:   "tenant1-cluster2.test",
					Version:  "2.22",
					TenantId: int32(tenant1Id),
				}
				createdTenant1Cluster1, _ := dao.CreateCluster(ctx, tenant1Cluster1)
				createdTenant1Cluster2, _ := dao.CreateCluster(ctx, tenant1Cluster2)

				// Create nodes for tenant 1
				tenant1Node1 := &models.Node{
					Name:        "tenant1-node1",
					ClusterUuid: createdTenant1Cluster1.ID.String(),
					TenantId:    int32(tenant1Id),
					Status:      "Ready",
					GpuCount:    2,
				}
				tenant1Node2 := &models.Node{
					Name:        "tenant1-node2",
					ClusterUuid: createdTenant1Cluster2.ID.String(),
					TenantId:    int32(tenant1Id),
					Status:      "Ready",
					GpuCount:    1,
				}
				dao.UpsertNode(ctx, tenant1Node1)
				dao.UpsertNode(ctx, tenant1Node2)

				// Create multiple clusters and nodes for tenant 2
				tenant2Cluster1 := &models.Cluster{
					Name:     "tenant2-cluster1",
					Domain:   "tenant2-cluster1.test",
					Version:  "2.22",
					TenantId: int32(tenant2Id),
				}
				tenant2Cluster2 := &models.Cluster{
					Name:     "tenant2-cluster2",
					Domain:   "tenant2-cluster2.test",
					Version:  "2.22",
					TenantId: int32(tenant2Id),
				}
				tenant2Cluster3 := &models.Cluster{
					Name:     "tenant2-cluster3",
					Domain:   "tenant2-cluster3.test",
					Version:  "2.22",
					TenantId: int32(tenant2Id),
				}
				createdTenant2Cluster1, _ := dao.CreateCluster(ctx, tenant2Cluster1)
				createdTenant2Cluster2, _ := dao.CreateCluster(ctx, tenant2Cluster2)
				createdTenant2Cluster3, _ := dao.CreateCluster(ctx, tenant2Cluster3)

				// Create nodes for tenant 2
				tenant2Node1 := &models.Node{
					Name:        "tenant2-node1",
					ClusterUuid: createdTenant2Cluster1.ID.String(),
					TenantId:    int32(tenant2Id),
					Status:      "Ready",
					GpuCount:    4,
				}
				tenant2Node2 := &models.Node{
					Name:        "tenant2-node2",
					ClusterUuid: createdTenant2Cluster2.ID.String(),
					TenantId:    int32(tenant2Id),
					Status:      "Ready",
					GpuCount:    2,
				}
				tenant2Node3 := &models.Node{
					Name:        "tenant2-node3",
					ClusterUuid: createdTenant2Cluster3.ID.String(),
					TenantId:    int32(tenant2Id),
					Status:      "Ready",
					GpuCount:    1,
				}
				dao.UpsertNode(ctx, tenant2Node1)
				dao.UpsertNode(ctx, tenant2Node2)
				dao.UpsertNode(ctx, tenant2Node3)

				// Verify both tenants have their resources before deletion
				tenant1PermittedScopes := openapi.PermittedScopes{
					Tenants: []string{strconv.Itoa(tenant1Id)},
				}
				tenant2PermittedScopes := openapi.PermittedScopes{
					Tenants: []string{strconv.Itoa(tenant2Id)},
				}

				tenant1Clusters, _ := dao.GetClusters(ctx, int32(tenant1Id), extapi.VerbosityFull, tenant1PermittedScopes, nil, false, false, 0, 0, nil)
				Expect(len(*tenant1Clusters)).Should(Equal(2))

				tenant1Nodes, _ := dao.GetNodes(ctx, int32(tenant1Id), &tenant1PermittedScopes, nil, "", "", 0, 100, "")
				Expect(len(*tenant1Nodes)).Should(Equal(2))

				tenant2Clusters, _ := dao.GetClusters(ctx, int32(tenant2Id), extapi.VerbosityFull, tenant2PermittedScopes, nil, false, false, 0, 0, nil)
				Expect(len(*tenant2Clusters)).Should(Equal(3))

				tenant2Nodes, _ := dao.GetNodes(ctx, int32(tenant2Id), &tenant2PermittedScopes, nil, "", "", 0, 100, "")
				Expect(len(*tenant2Nodes)).Should(Equal(3))

				// Delete only tenant 1
				msg := events.NewTenantDelete(strconv.Itoa(tenant1Id))
				err := tenantsSubscriber.TenantDeleteSubscribeHandler(ctx, msg)
				Expect(err).Should(BeNil())

				// Verify tenant 1 is completely deleted
				tenant1ClustersAfterDelete, _ := dao.GetClusters(ctx, int32(tenant1Id), extapi.VerbosityFull, tenant1PermittedScopes, nil, false, false, 0, 0, nil)
				Expect(len(*tenant1ClustersAfterDelete)).Should(Equal(0))

				tenant1NodesAfterDelete, _ := dao.GetNodes(ctx, int32(tenant1Id), &tenant1PermittedScopes, nil, "", "", 0, 100, "")
				Expect(len(*tenant1NodesAfterDelete)).Should(Equal(0))

				// Verify tenant 2 is completely intact with all resources
				tenant2ClustersAfterDelete, _ := dao.GetClusters(ctx, int32(tenant2Id), extapi.VerbosityFull, tenant2PermittedScopes, nil, false, false, 0, 0, nil)
				Expect(len(*tenant2ClustersAfterDelete)).Should(Equal(3))

				tenant2NodesAfterDelete, _ := dao.GetNodes(ctx, int32(tenant2Id), &tenant2PermittedScopes, nil, "", "", 0, 100, "")
				Expect(len(*tenant2NodesAfterDelete)).Should(Equal(3))

				// Matched by name rather than by position: neither GetClusters nor GetNodes
				// applies an ORDER BY here, so the row order is whatever the planner yields.
				Expect(*tenant2ClustersAfterDelete).Should(ContainElement(SatisfyAll(
					HaveField("Name", "tenant2-cluster1"),
					HaveField("Domain", "tenant2-cluster1.test"),
					HaveField("Version", "2.22"),
				)))

				Expect(*tenant2NodesAfterDelete).Should(ContainElement(SatisfyAll(
					HaveField("Name", "tenant2-node1"),
					HaveField("GpuCount", int32(4)),
					HaveField("Status", "Ready"),
				)))
			})
		})

		Describe("SubscribeTenantDelete", func() {
			It("should successfully subscribe to tenant delete events", func() {
				// Set up mock for SubscribeAsync method - use mock.Anything for all parameters
				pubSubMock.pubsub.On("SubscribeAsync", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

				errChan := tenantsSubscriber.SubscribeTenantDelete()
				Expect(errChan).ShouldNot(BeNil())
			})
		})
	})
})

func clusterSoftDeleteValidation(clusterId string) {
	err, cluster := clusterEntity(clusterId)
	Expect(err).Should(BeNil())
	Expect(cluster.DeletedAt.Valid).Should(BeTrue())
	Expect(wasWithinLast10Minutes(cluster.DeletedAt.Time)).Should(BeTrue())
	err, clusterConfig := clusterConfigEntity(clusterId)
	Expect(err).Should(BeNil())
	Expect(clusterConfig.DeletedAt.Valid).Should(BeTrue())
	Expect(wasWithinLast10Minutes(cluster.DeletedAt.Time)).Should(BeTrue())
}

func clusterHardDeletedValidation(clusterId string) {
	err, cluster := clusterEntity(clusterId)
	Expect(err.Error()).Should(ContainSubstring("record not found"))
	Expect(cluster.DeletedAt.Valid).Should(BeFalse())
	Expect(cluster.DeletedAt.Time.IsZero()).Should(BeTrue())
	err, clusterConfig := clusterConfigEntity(clusterId)
	Expect(err.Error()).Should(ContainSubstring("record not found"))
	Expect(clusterConfig.DeletedAt.Valid).Should(BeFalse())
	Expect(clusterConfig.DeletedAt.Time.IsZero()).Should(BeTrue())
}

func clusterConfigEntity(clusterId string) (error, *models.ClusterConfig) {
	tx := runaidb.DB().WithContext(context.Background()).Unscoped()
	var clusterConfig models.ClusterConfig
	err := tx.Where("cluster_id = ?", clusterId).First(&clusterConfig)
	return err, &clusterConfig
}

func clusterEntity(id string) (error, *models.Cluster) {
	tx := runaidb.DB().WithContext(context.Background()).Unscoped()
	var cluster models.Cluster
	err := tx.Where("id = ?", id).First(&cluster)
	return err, &cluster
}

func wasWithinLast10Minutes(t time.Time) bool {
	return time.Since(t) <= 10*time.Minute && time.Since(t) >= 0
}
