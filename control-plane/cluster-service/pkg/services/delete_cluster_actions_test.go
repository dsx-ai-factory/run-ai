/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services_test

import (
	"context"
	"strconv"
	"time"

	"github.com/google/uuid"
	runai_config "github.com/run-ai/runai/runai-common-packages/config"
	"github.com/run-ai/runai/runai-common-packages/runaidb"

	"github.com/run-ai/runai/control-plane/cluster-service/pkg/config"
	cluster_delete "github.com/run-ai/runai/control-plane/cluster-service/pkg/services/cluster-delete"

	authorizationUtils "github.com/run-ai/runai/control-plane/authorization/pkg/client/helpers"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	authorizationclient "github.com/run-ai/runai/control-plane/authorization/pkg/client"
	authorizationapi "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"
	extapi "github.com/run-ai/runai/control-plane/cluster-service/pkg/api"
	"github.com/run-ai/runai/control-plane/cluster-service/pkg/database/dao"
	"github.com/run-ai/runai/control-plane/cluster-service/pkg/database/models"
	orgunitclient "github.com/run-ai/runai/runai-clients/org-unit-service-client/mocks"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
	pubsub "github.com/run-ai/runai/runai-common-packages/runaipubsub"
	pubsubmocks "github.com/run-ai/runai/runai-common-packages/runaipubsub/mocks"
	"github.com/stretchr/testify/mock"
)

type DeleteClusterMocks struct {
	orgUnitClient       *orgunitclient.IOrgUnitClient
	authorizationClient authorizationclient.IAuthorizationClient
	pubsub              *pubsubmocks.PubSuber
}

func initService() (cluster_delete.ClusterDeleter, *DeleteClusterMocks) {
	mocks := &DeleteClusterMocks{
		authorizationClient: authorizationclient.NewAuthorizationClient(),
		orgUnitClient:       &orgunitclient.IOrgUnitClient{},
		pubsub:              &pubsubmocks.PubSuber{},
	}
	pubsub.SetMock(mocks.pubsub)
	return cluster_delete.NewClusterDeleter(mocks.authorizationClient, mocks.orgUnitClient, nil), mocks
}

func prepareClusterDeleteMocks() (cluster_delete.ClusterDeleter, *DeleteClusterMocks) {
	clusterDeleter, mocks := initService()
	mocks.pubsub.On("Publish", mock.Anything, mock.Anything).Return(nil)
	// Setup service
	mocks.orgUnitClient.On("ClusterDeleted", mock.Anything, mock.Anything).Return(nil)
	return clusterDeleter, mocks
}

var _ = Describe("DeleteCluster", func() {
	var (
		ctx            context.Context
		clusterDeleter cluster_delete.ClusterDeleter
		mocks          *DeleteClusterMocks
	)
	BeforeEach(func() {
		tenantIdStr := strconv.Itoa(tenantId)
		ctx = context.WithValue(context.Background(), runaictx.CfTenantId, tenantIdStr)
		ctx = context.WithValue(ctx, authorizationUtils.PermittedScopesKey, &authorizationapi.PermittedScopes{
			Tenants: []string{tenantIdStr},
		})
	})
	Describe("DeleteCluster when PubsubInit=true", func() {
		BeforeEach(func() {
			serviceConfig := runai_config.Get[*config.Service]()
			serviceConfig.PubsubInitialized = true
			runai_config.Set(serviceConfig)
			clusterDeleter, mocks = prepareClusterDeleteMocks()
		})
		Describe("nonAsyncDelete", func() {
			It("nonAsyncDelete", func() {
				cluster1Name := "clusterDeleted1"
				cluster2Name := "clusterDeleted2"

				cluster1Model := &models.Cluster{
					Name:     cluster1Name,
					Domain:   "cluster-11",
					Version:  "",
					TenantId: tenantId,
				}

				cluster2Model := &models.Cluster{
					Name:     cluster2Name,
					Domain:   "cluster-22",
					Version:  "",
					TenantId: tenantId,
				}
				cluster1, _ := dao.CreateCluster(ctx, cluster1Model)
				cluster2, _ := dao.CreateCluster(ctx, cluster2Model)

				err := clusterDeleter.NonAsyncDelete(ctx, cluster1)
				Expect(err).Should(BeNil())
				cluster1FromDB, err := dao.GetClusterByIdAndTenantId(ctx, cluster1.ID.String(), tenantId, extapi.VerbosityFull)
				Expect(err.Code).Should(Equal(int32(404)))
				Expect(cluster1FromDB).Should(BeNil())
				cluster2FromDB, err := dao.GetClusterByIdAndTenantId(ctx, cluster2.ID.String(), tenantId, extapi.VerbosityMetadata)

				Expect(err).Should(BeNil())
				Expect(cluster2FromDB).ShouldNot(BeNil())
				mocks.orgUnitClient.AssertCalled(GinkgoT(), "ClusterDeleted", ctx, cluster1.ID.String())
				mocks.orgUnitClient.AssertNotCalled(GinkgoT(), "ClusterDeleted", ctx, cluster2.ID.String())
				dao.DeleteCluster(ctx, cluster1.ID.String(), tenantId)
				dao.DeleteCluster(ctx, cluster2.ID.String(), tenantId)
			})

			Context("cluster with nodes", func() {
				var (
					cluster         *models.Cluster
					node1Id         uuid.UUID
					permittedScopes = authorizationapi.PermittedScopes{
						Tenants: []string{strconv.Itoa(tenantId)},
					}
				)

				BeforeEach(func() {
					clusterModel := &models.Cluster{
						Name:     "clusterWithNodes",
						Domain:   "cluster-with-nodes",
						Version:  "2.22",
						TenantId: tenantId,
					}
					var err error
					cluster, err = dao.CreateCluster(ctx, clusterModel)
					Expect(err).Should(BeNil())
					Expect(cluster).ShouldNot(BeNil())

					node1 := &models.Node{
						Name:        "node-1",
						ClusterUuid: cluster.ID.String(),
						TenantId:    tenantId,
						Status:      "Ready",
						GpuCount:    2,
					}
					node2 := &models.Node{
						Name:        "node-2",
						ClusterUuid: cluster.ID.String(),
						TenantId:    tenantId,
						Status:      "Ready",
						GpuCount:    1,
					}
					_, err = dao.UpsertNode(ctx, node1)
					Expect(err).Should(BeNil())
					node1Id = node1.ID
					_, err = dao.UpsertNode(ctx, node2)
					Expect(err).Should(BeNil())

					_, _, syncErr := dao.SyncNumaNodes(
						ctx,
						[]dao.NodeNumaSnapshot{{
							NodeName:  node1.Name,
							NumaNodes: []models.NodeNuma{{Name: "zone-0"}},
						}},
						tenantId,
						cluster.ID,
						false,
						"",
						"",
					)
					Expect(syncErr).Should(BeNil())

					// Verify nodes exist before tests
					nodesBefore, getErr := dao.GetNodes(ctx, tenantId, &permittedScopes, nil, "", "", 0, 100, "")
					Expect(getErr).Should(BeNil())
					clusterNodesBefore := fileterClusterNodes(*nodesBefore, cluster.ID.String())
					Expect(len(clusterNodesBefore)).Should(Equal(2))

					numaNodesBefore, getNumaErr := dao.GetNumaNodes(ctx, tenantId, node1Id)
					Expect(getNumaErr).Should(BeNil())
					Expect(numaNodesBefore).Should(HaveLen(1))
				})

				It("deletes cluster nodes when cluster is deleted", func() {
					// Delete the cluster using NonAsyncDelete (the actual deletion logic)
					err := clusterDeleter.NonAsyncDelete(ctx, cluster)
					Expect(err).Should(BeNil())

					// Verify cluster is deleted
					clusterFromDB, apiErr := dao.GetClusterByIdAndTenantId(ctx, cluster.ID.String(), tenantId, extapi.VerbosityFull)
					Expect(apiErr.Code).Should(Equal(int32(404)))
					Expect(clusterFromDB).Should(BeNil())

					// Verify nodes are deleted
					nodesAfter, getErr := dao.GetNodes(ctx, tenantId, &permittedScopes, nil, "", "", 0, 100, "")
					Expect(getErr).Should(BeNil())
					clusterNodesAfter := fileterClusterNodes(*nodesAfter, cluster.ID.String())
					Expect(len(clusterNodesAfter)).Should(Equal(0))

					numaNodesAfter, getNumaErr := dao.GetNumaNodes(ctx, tenantId, node1Id)
					Expect(getNumaErr).Should(BeNil())
					Expect(numaNodesAfter).Should(BeEmpty())
				})

				It("migration query deletes orphaned nodes from deleted clusters - and only them", func() {
					// Create another cluster that will NOT be deleted (active cluster)
					anotherClusterModel := &models.Cluster{
						Name:     "activeClusterWithNodes",
						Domain:   "active-cluster-with-nodes",
						Version:  "2.22",
						TenantId: tenantId,
					}
					anotherCluster, err := dao.CreateCluster(ctx, anotherClusterModel)
					Expect(err).Should(BeNil())
					Expect(anotherCluster).ShouldNot(BeNil())

					// Create nodes for the active cluster
					anotherClusterNode1 := &models.Node{
						Name:        "active-node-1",
						ClusterUuid: anotherCluster.ID.String(),
						TenantId:    tenantId,
						Status:      "Ready",
						GpuCount:    4,
					}
					anotherClusterNode2 := &models.Node{
						Name:        "active-node-2",
						ClusterUuid: anotherCluster.ID.String(),
						TenantId:    tenantId,
						Status:      "Ready",
						GpuCount:    4,
					}
					_, err = dao.UpsertNode(ctx, anotherClusterNode1)
					Expect(err).Should(BeNil())
					_, err = dao.UpsertNode(ctx, anotherClusterNode2)
					Expect(err).Should(BeNil())

					// Delete ONLY the first cluster (not using NonAsyncDelete) - this leaves nodes orphaned
					apiErr := dao.DeleteCluster(ctx, cluster.ID.String(), tenantId)
					Expect(apiErr).Should(BeNil())

					// Verify cluster is deleted (soft delete)
					clusterFromDB, apiErr := dao.GetClusterByIdAndTenantId(ctx, cluster.ID.String(), tenantId, extapi.VerbosityFull)
					Expect(apiErr.Code).Should(Equal(int32(404)))
					Expect(clusterFromDB).Should(BeNil())

					// Verify nodes are still orphaned (not deleted yet)
					nodes, err := dao.GetNodes(ctx, tenantId, &permittedScopes, nil, "", "", 0, 100, "")
					Expect(err).Should(BeNil())
					clusterNodesOrphaned := fileterClusterNodes(*nodes, cluster.ID.String())
					Expect(len(clusterNodesOrphaned)).Should(Equal(2))
					// Verify another cluster's nodes exist
					anotherClusterNodesBefore := fileterClusterNodes(*nodes, anotherCluster.ID.String())
					Expect(len(anotherClusterNodesBefore)).Should(Equal(2))

					// Execute the migration query to delete orphaned nodes
					migrationQuery := `
						DELETE FROM node n
						WHERE NOT EXISTS (
						SELECT 1
						FROM clusters c
						WHERE c.id = n.cluster_uuid::uuid 
						AND c.deleted_at IS NULL
						)
					`
					_, execErr := runaidb.DB().Exec(migrationQuery)
					Expect(execErr).Should(BeNil())

					// Verify orphaned nodes from deleted cluster are now deleted
					nodesAfterMigration, getErr := dao.GetNodes(ctx, tenantId, &permittedScopes, nil, "", "", 0, 100, "")
					Expect(getErr).Should(BeNil())
					clusterNodesAfterMigration := fileterClusterNodes(*nodesAfterMigration, cluster.ID.String())
					Expect(len(clusterNodesAfterMigration)).Should(Equal(0))

					// Verify nodes from active cluster are NOT deleted
					anotherClusterNodesAfterMigration := fileterClusterNodes(*nodesAfterMigration, anotherCluster.ID.String())
					Expect(len(anotherClusterNodesAfterMigration)).Should(Equal(2))

					// Cleanup: delete the active cluster
					dao.DeleteCluster(ctx, anotherCluster.ID.String(), tenantId)
				})
			})
		})
		Describe("AttemptToGracefullyDeleteCluster", func() {
			It("async delete attempt new cluster", func() {
				cluster1Name := "clusterDeleted1NewVer"

				cluster1Model := &models.Cluster{
					Name:     cluster1Name,
					Domain:   "cluster-11",
					Version:  "2.22",
					TenantId: tenantId,
				}

				cluster1, _ := dao.CreateCluster(ctx, cluster1Model)
				cluster1FromDB, err := dao.GetClusterByIdAndTenantId(ctx, cluster1.ID.String(), tenantId, extapi.VerbosityFull)
				Expect(cluster1FromDB.DeleteRequestedAt).Should(BeNil())

				err = clusterDeleter.AttemptToGracefullyDeleteCluster(ctx, cluster1)
				Expect(err).Should(BeNil())
				cluster1FromDB, err = dao.GetClusterByIdAndTenantId(ctx, cluster1.ID.String(), tenantId, extapi.VerbosityFull)
				Expect(cluster1FromDB.DeleteRequestedAt).ShouldNot(BeNil())
				mocks.pubsub.AssertCalled(GinkgoT(), "Publish", "ClusterPendingDelete", mock.Anything)
				mocks.pubsub.AssertNotCalled(GinkgoT(), "Publish", "ClusterDelete", mock.Anything)
				dao.DeleteCluster(ctx, cluster1.ID.String(), tenantId)
			})

			It("async delete attempt old cluster", func() {
				cluster1Name := "clusterDeleted1OldVer"

				cluster1Model := &models.Cluster{
					Name:     cluster1Name,
					Domain:   "cluster-11",
					Version:  "2.17",
					TenantId: tenantId,
				}

				cluster1, _ := dao.CreateCluster(ctx, cluster1Model)
				cluster1FromDB, err := dao.GetClusterByIdAndTenantId(ctx, cluster1.ID.String(), tenantId, extapi.VerbosityFull)
				Expect(cluster1FromDB.DeleteRequestedAt).Should(BeNil())

				err = clusterDeleter.AttemptToGracefullyDeleteCluster(ctx, cluster1)
				Expect(err).Should(BeNil())
				cluster1FromDB, err = dao.GetClusterByIdAndTenantId(ctx, cluster1.ID.String(), tenantId, extapi.VerbosityFull)
				Expect(cluster1FromDB.DeleteRequestedAt).ShouldNot(BeNil())
				mocks.pubsub.AssertCalled(GinkgoT(), "Publish", "ClusterPendingDelete", mock.Anything)
				mocks.pubsub.AssertCalled(GinkgoT(), "Publish", "ClusterDelete", mock.Anything)
				dao.DeleteCluster(ctx, cluster1.ID.String(), tenantId)
			})
		})
		Describe("ForceDeleteCluster", func() {
			It("force delete attempt new cluster version with DeleteRequestedAt", func() {
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
				err := clusterDeleter.ForceDeleteCluster(ctx, cluster1)
				Expect(err).Should(BeNil())
				mocks.pubsub.AssertCalled(GinkgoT(), "Publish", "ClusterDelete", mock.Anything)
				dao.DeleteCluster(ctx, cluster1.ID.String(), tenantId)
			})

			It("force delete attempt new cluster version with DeleteRequestedAt=nil", func() {
				cluster1Name := "clusterForceNewVerNilTS"
				cluster1Model := &models.Cluster{
					Name:     cluster1Name,
					Domain:   "cluster-11",
					Version:  "2.22",
					TenantId: tenantId,
				}

				cluster1, _ := dao.CreateCluster(ctx, cluster1Model)
				err := clusterDeleter.ForceDeleteCluster(ctx, cluster1)
				Expect(err).Should(BeNil())
				mocks.pubsub.AssertCalled(GinkgoT(), "Publish", "ClusterDelete", mock.Anything)
				dao.DeleteCluster(ctx, cluster1.ID.String(), tenantId)
			})

			It("force delete old cluster version", func() {
				cluster1Name := "clusterForceOldVer"
				cluster1Model := &models.Cluster{
					Name:     cluster1Name,
					Domain:   "cluster-11",
					Version:  "2.17",
					TenantId: tenantId,
				}

				cluster1, _ := dao.CreateCluster(ctx, cluster1Model)
				err := clusterDeleter.ForceDeleteCluster(ctx, cluster1)
				Expect(err).Should(BeNil())
				mocks.pubsub.AssertCalled(GinkgoT(), "Publish", "ClusterDelete", mock.Anything)
				dao.DeleteCluster(ctx, cluster1.ID.String(), tenantId)
			})
		})
	})
	Describe("DeleteCluster when PubsubInit=false", func() {
		BeforeEach(func() {
			serviceConfig := runai_config.Get[*config.Service]()
			serviceConfig.PubsubInitialized = false
			runai_config.Set(serviceConfig)
			clusterDeleter, mocks = prepareClusterDeleteMocks()
		})
		Describe("nonAsyncDelete", func() {
			It("nonAsyncDelete when PubsubInitialized is false", func() {
				clusterName := "clusterDeletedPubsubFalse"
				clusterModel := &models.Cluster{
					Name:     clusterName,
					Domain:   "cluster-pubsub-false",
					Version:  "",
					TenantId: tenantId,
				}
				cluster, _ := dao.CreateCluster(ctx, clusterModel)

				err := clusterDeleter.NonAsyncDelete(ctx, cluster)
				Expect(err).Should(BeNil())

				// Verify cluster is deleted from DB
				clusterFromDB, err := dao.GetClusterByIdAndTenantId(ctx, cluster.ID.String(), tenantId, extapi.VerbosityFull)
				Expect(err.Code).Should(Equal(int32(404)))
				Expect(clusterFromDB).Should(BeNil())

				// Verify orgUnitClient.ClusterDeleted was called
				mocks.orgUnitClient.AssertCalled(GinkgoT(), "ClusterDeleted", ctx, cluster.ID.String())

				// Verify pubsub.Publish was not called
				mocks.pubsub.AssertNotCalled(GinkgoT(), "Publish", mock.Anything, mock.Anything)

				// Cleanup
				dao.DeleteCluster(ctx, cluster.ID.String(), tenantId)
			})
		})
		Describe("AttemptToGracefullyDeleteCluster", func() {
			It("async delete attempt old cluster", func() {
				cluster1Name := "clusterDeleted1OldVer"

				cluster1Model := &models.Cluster{
					Name:     cluster1Name,
					Domain:   "cluster-11",
					Version:  "2.17",
					TenantId: tenantId,
				}

				cluster1, _ := dao.CreateCluster(ctx, cluster1Model)
				cluster1FromDB, err := dao.GetClusterByIdAndTenantId(ctx, cluster1.ID.String(), tenantId, extapi.VerbosityFull)
				Expect(cluster1FromDB.DeleteRequestedAt).Should(BeNil())

				err = clusterDeleter.AttemptToGracefullyDeleteCluster(ctx, cluster1)
				Expect(err).Should(BeNil())
				clusterFromDB, err := dao.GetClusterByIdAndTenantId(ctx, cluster1.ID.String(), tenantId, extapi.VerbosityFull)
				Expect(err.Code).Should(Equal(int32(404)))
				Expect(clusterFromDB).Should(BeNil())

				// Verify orgUnitClient.ClusterDeleted was called
				mocks.orgUnitClient.AssertCalled(GinkgoT(), "ClusterDeleted", ctx, cluster1.ID.String())

				// Verify neither ClusterPendingDelete nor ClusterDelete was published
				mocks.pubsub.AssertNotCalled(GinkgoT(), "Publish", "ClusterPendingDelete", mock.Anything)
				mocks.pubsub.AssertNotCalled(GinkgoT(), "Publish", "ClusterDelete", mock.Anything)

				// Cleanup
				dao.DeleteCluster(ctx, cluster1.ID.String(), tenantId)
			})
		})
		Describe("ForceDeleteCluster", func() {
			var cluster1 *models.Cluster
			AfterEach(func() {
				clusterFromDB, err := dao.GetClusterByIdAndTenantId(ctx, cluster1.ID.String(), tenantId, extapi.VerbosityFull)
				Expect(err.Code).Should(Equal(int32(404)))
				Expect(clusterFromDB).Should(BeNil())

				// Verify orgUnitClient.ClusterDeleted was called
				mocks.orgUnitClient.AssertCalled(GinkgoT(), "ClusterDeleted", ctx, cluster1.ID.String())

				// Verify pubsub.Publish was not called
				mocks.pubsub.AssertNotCalled(GinkgoT(), "Publish", mock.Anything, mock.Anything)
				dao.DeleteCluster(ctx, cluster1.ID.String(), tenantId)
			})

			It("force delete attempt new cluster version with DeleteRequestedAt", func() {
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

				cluster1, _ = dao.CreateCluster(ctx, cluster1Model)
				err := clusterDeleter.ForceDeleteCluster(ctx, cluster1)
				Expect(err).Should(BeNil())
			})

			It("force delete attempt new cluster version with DeleteRequestedAt=nil", func() {
				cluster1Name := "clusterForceNewVerNilTS"
				cluster1Model := &models.Cluster{
					Name:     cluster1Name,
					Domain:   "cluster-11",
					Version:  "2.22",
					TenantId: tenantId,
				}

				cluster1, _ = dao.CreateCluster(ctx, cluster1Model)
				err := clusterDeleter.ForceDeleteCluster(ctx, cluster1)
				Expect(err).Should(BeNil())
			})

			It("force delete old cluster version", func() {
				cluster1Name := "clusterForceOldVer"
				cluster1Model := &models.Cluster{
					Name:     cluster1Name,
					Domain:   "cluster-11",
					Version:  "2.17",
					TenantId: tenantId,
				}

				cluster1, _ = dao.CreateCluster(ctx, cluster1Model)
				err := clusterDeleter.ForceDeleteCluster(ctx, cluster1)
				Expect(err).Should(BeNil())
			})
		})

		Describe("Test cluster soft/hard delete", func() {
			It("delete tenant after soft delete", func() {
				clusterNamePrefix := "tenantDeleted"
				cluster1Name := clusterNamePrefix + "1"
				cluster2Name := clusterNamePrefix + "2"

				cluster1Model := &models.Cluster{
					Name:     cluster1Name,
					Domain:   "cluster-11",
					Version:  "",
					TenantId: tenantId,
				}

				cluster2Model := &models.Cluster{
					Name:     cluster2Name,
					Domain:   "cluster-22",
					Version:  "",
					TenantId: tenantId,
				}
				cluster1, _ := dao.CreateCluster(ctx, cluster1Model)
				cluster2, _ := dao.CreateCluster(ctx, cluster2Model)

				err := clusterDeleter.NonAsyncDelete(ctx, cluster1)
				Expect(err).Should(BeNil())
				cluster1FromDB, err := dao.GetClusterByIdAndTenantId(ctx, cluster1.ID.String(), tenantId, extapi.VerbosityFull)
				Expect(err.Code).Should(Equal(int32(404)))
				Expect(cluster1FromDB).Should(BeNil())
				cluster2FromDB, err := dao.GetClusterByIdAndTenantId(ctx, cluster2.ID.String(), tenantId, extapi.VerbosityMetadata)

				Expect(err).Should(BeNil())
				Expect(cluster2FromDB).ShouldNot(BeNil())

				// should return 2 clusters one soft deleted and one active
				clusters, _ := getTenantClusters(strconv.Itoa(tenantId), clusterNamePrefix)
				Expect(len(clusters)).Should(Equal(2))

				daoErr := clusterDeleter.DeleteTenantClusters(ctx, tenantId)
				Expect(daoErr).Should(BeNil())
				clusters, _ = getTenantClusters(strconv.Itoa(tenantId), clusterNamePrefix)
				Expect(clusters).Should(BeEmpty(), "Expected no clusters for tenant after deletion")
			})

			It("DeleteTenantClusters success with multiple clusters and nodes", func() {
				clusterNamePrefix := "tenantDeleteSuccess"
				cluster1Name := clusterNamePrefix + "1"
				cluster2Name := clusterNamePrefix + "2"

				// Create multiple clusters
				cluster1Model := &models.Cluster{
					Name:     cluster1Name,
					Domain:   "cluster-11",
					Version:  "2.22",
					TenantId: tenantId,
				}

				cluster2Model := &models.Cluster{
					Name:     cluster2Name,
					Domain:   "cluster-22",
					Version:  "2.22",
					TenantId: tenantId,
				}
				cluster1, _ := dao.CreateCluster(ctx, cluster1Model)
				cluster2, _ := dao.CreateCluster(ctx, cluster2Model)

				// Create nodes for the clusters
				node1 := &models.Node{
					Name:        "node-1",
					ClusterUuid: cluster1.ID.String(),
					TenantId:    tenantId,
					Status:      "Ready",
					GpuCount:    2,
				}
				node2 := &models.Node{
					Name:        "node-2",
					ClusterUuid: cluster2.ID.String(),
					TenantId:    tenantId,
					Status:      "Ready",
					GpuCount:    1,
				}

				dao.UpsertNode(ctx, node1)
				dao.UpsertNode(ctx, node2)

				_, _, syncErr := dao.SyncNumaNodes(
					ctx,
					[]dao.NodeNumaSnapshot{{
						NodeName:  node1.Name,
						NumaNodes: []models.NodeNuma{{Name: "zone-0"}},
					}},
					tenantId,
					cluster1.ID,
					false,
					"",
					"",
				)
				Expect(syncErr).Should(BeNil())
				_, _, syncErr = dao.SyncNumaNodes(
					ctx,
					[]dao.NodeNumaSnapshot{{
						NodeName:  node2.Name,
						NumaNodes: []models.NodeNuma{{Name: "zone-0"}},
					}},
					tenantId,
					cluster2.ID,
					false,
					"",
					"",
				)
				Expect(syncErr).Should(BeNil())

				// Verify clusters and nodes exist before deletion
				clusters, _ := getTenantClusters(strconv.Itoa(tenantId), clusterNamePrefix)
				Expect(len(clusters)).Should(Equal(2))

				// Verify nodes exist
				permittedScopes := authorizationapi.PermittedScopes{
					Tenants: []string{strconv.Itoa(tenantId)},
				}
				nodes, _ := dao.GetNodes(ctx, tenantId, &permittedScopes, nil, "", "", 0, 100, "")
				Expect(len(*nodes)).Should(BeNumerically(">=", 2))

				// Test successful deletion
				daoErr := clusterDeleter.DeleteTenantClusters(ctx, tenantId)
				Expect(daoErr).Should(BeNil())

				// Verify all clusters are deleted
				clusters, _ = getTenantClusters(strconv.Itoa(tenantId), clusterNamePrefix)
				Expect(clusters).Should(BeEmpty(), "Expected no clusters for tenant after deletion")

				// Verify all nodes are deleted
				nodes, _ = dao.GetNodes(ctx, tenantId, &permittedScopes, nil, "", "", 0, 100, "")
				Expect(len(*nodes)).Should(Equal(0))

				node1Numa, getNumaErr := dao.GetNumaNodes(ctx, tenantId, node1.ID)
				Expect(getNumaErr).Should(BeNil())
				Expect(node1Numa).Should(BeEmpty())
				node2Numa, getNumaErr := dao.GetNumaNodes(ctx, tenantId, node2.ID)
				Expect(getNumaErr).Should(BeNil())
				Expect(node2Numa).Should(BeEmpty())
			})

			It("DeleteTenantClusters success with empty tenant", func() {
				// Test deletion of tenant with no clusters or nodes
				emptyTenantId := 999
				daoErr := clusterDeleter.DeleteTenantClusters(ctx, emptyTenantId)
				Expect(daoErr).Should(BeNil())
			})

			It("DeleteTenantClusters handles mixed cluster states", func() {
				clusterNamePrefix := "tenantDeleteMixed"
				cluster1Name := clusterNamePrefix + "1"
				cluster2Name := clusterNamePrefix + "2"

				// Create one active cluster
				cluster1Model := &models.Cluster{
					Name:     cluster1Name,
					Domain:   "cluster-11",
					Version:  "2.22",
					TenantId: tenantId,
				}

				// Create one soft-deleted cluster
				cluster2Model := &models.Cluster{
					Name:     cluster2Name,
					Domain:   "cluster-22",
					Version:  "2.22",
					TenantId: tenantId,
				}
				_, _ = dao.CreateCluster(ctx, cluster1Model)
				cluster2, _ := dao.CreateCluster(ctx, cluster2Model)

				// Soft delete the second cluster
				err := clusterDeleter.NonAsyncDelete(ctx, cluster2)
				Expect(err).Should(BeNil())

				// Verify mixed state before deletion
				clusters, _ := getTenantClusters(strconv.Itoa(tenantId), clusterNamePrefix)
				Expect(len(clusters)).Should(Equal(2))

				// Test successful deletion
				daoErr := clusterDeleter.DeleteTenantClusters(ctx, tenantId)
				Expect(daoErr).Should(BeNil())

				// Verify all clusters are deleted regardless of state
				clusters, _ = getTenantClusters(strconv.Itoa(tenantId), clusterNamePrefix)
				Expect(clusters).Should(BeEmpty(), "Expected no clusters for tenant after deletion")
			})
		})
	})

	Describe("isPubsubEnabled", func() {
		const testTenantId = 1
		var (
			testCtx            context.Context
			testClusterDeleter cluster_delete.ClusterDeleter
			testMocks          *DeleteClusterMocks
		)

		BeforeEach(func() {
			tenantIdStr := strconv.Itoa(testTenantId)
			testCtx = context.WithValue(context.Background(), runaictx.CfTenantId, tenantIdStr)
			testCtx = context.WithValue(testCtx, authorizationUtils.PermittedScopesKey, &authorizationapi.PermittedScopes{
				Tenants: []string{tenantIdStr},
			})
		})

		It("should return true when PubsubInitialized is true", func() {
			serviceConfig := runai_config.Get[*config.Service]()
			serviceConfig.PubsubInitialized = true
			runai_config.Set(serviceConfig)
			testClusterDeleter, testMocks = prepareClusterDeleteMocks()

			// Test through ForceDeleteCluster - when pubsub is enabled, it should publish
			clusterModel := &models.Cluster{
				Name:     "test-pubsub-enabled",
				Domain:   "cluster-test",
				Version:  "2.22",
				TenantId: testTenantId,
			}
			cluster, _ := dao.CreateCluster(testCtx, clusterModel)

			err := testClusterDeleter.ForceDeleteCluster(testCtx, cluster)
			Expect(err).Should(BeNil())

			// Verify pubsub.Publish was called (indicating isPubsubEnabled returned true)
			testMocks.pubsub.AssertCalled(GinkgoT(), "Publish", mock.Anything, mock.Anything)

			// Cleanup
			dao.DeleteCluster(testCtx, cluster.ID.String(), testTenantId)
		})

		It("should return false when PubsubInitialized is false", func() {
			serviceConfig := runai_config.Get[*config.Service]()
			serviceConfig.PubsubInitialized = false
			runai_config.Set(serviceConfig)
			testClusterDeleter, testMocks = prepareClusterDeleteMocks()

			// Test through ForceDeleteCluster - when pubsub is disabled, it should use NonAsyncDelete
			clusterModel := &models.Cluster{
				Name:     "test-pubsub-disabled",
				Domain:   "cluster-test",
				Version:  "2.22",
				TenantId: testTenantId,
			}
			cluster, _ := dao.CreateCluster(testCtx, clusterModel)

			err := testClusterDeleter.ForceDeleteCluster(testCtx, cluster)
			Expect(err).Should(BeNil())

			// Verify pubsub.Publish was NOT called (indicating isPubsubEnabled returned false)
			testMocks.pubsub.AssertNotCalled(GinkgoT(), "Publish", mock.Anything, mock.Anything)
			// Verify NonAsyncDelete was used (orgUnitClient.ClusterDeleted was called)
			testMocks.orgUnitClient.AssertCalled(GinkgoT(), "ClusterDeleted", testCtx, cluster.ID.String())

			// Cleanup
			dao.DeleteCluster(testCtx, cluster.ID.String(), testTenantId)
		})

		It("should dynamically reflect changes to PubsubInitialized config", func() {
			testClusterDeleter, testMocks = prepareClusterDeleteMocks()
			clusterModel := &models.Cluster{
				Name:     "test-pubsub-dynamic",
				Domain:   "cluster-test",
				Version:  "2.22",
				TenantId: testTenantId,
			}
			cluster, _ := dao.CreateCluster(testCtx, clusterModel)

			// First, set pubsub to false
			serviceConfig := runai_config.Get[*config.Service]()
			serviceConfig.PubsubInitialized = false
			runai_config.Set(serviceConfig)

			err := testClusterDeleter.ForceDeleteCluster(testCtx, cluster)
			Expect(err).Should(BeNil())
			testMocks.pubsub.AssertNotCalled(GinkgoT(), "Publish", mock.Anything, mock.Anything)

			// Recreate cluster for second test
			cluster2, _ := dao.CreateCluster(testCtx, clusterModel)

			// Now set pubsub to true
			serviceConfig.PubsubInitialized = true
			runai_config.Set(serviceConfig)

			err = testClusterDeleter.ForceDeleteCluster(testCtx, cluster2)
			Expect(err).Should(BeNil())
			// Now pubsub.Publish should be called (indicating isPubsubEnabled now returns true)
			testMocks.pubsub.AssertCalled(GinkgoT(), "Publish", mock.Anything, mock.Anything)

			// Cleanup
			dao.DeleteCluster(testCtx, cluster.ID.String(), testTenantId)
			dao.DeleteCluster(testCtx, cluster2.ID.String(), testTenantId)
		})
	})
})

func getTenantClusters(tenantId, namePrefix string) ([]models.Cluster, error) {
	tx := runaidb.DB().WithContext(context.Background()).Unscoped()
	var clusters []models.Cluster
	err := tx.Where("tenant_id = ?", tenantId).Where("name like ?", namePrefix+"%").Find(&clusters)
	return clusters, err
}

func fileterClusterNodes(nodes []models.Node, clusterId string) []models.Node {
	var result []models.Node
	for _, node := range nodes {
		if node.ClusterUuid == clusterId {
			result = append(result, node)
		}
	}
	return result
}
