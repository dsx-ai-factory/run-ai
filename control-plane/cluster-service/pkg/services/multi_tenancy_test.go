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

	runai_config "github.com/run-ai/runai/runai-common-packages/config"
	pubsub "github.com/run-ai/runai/runai-common-packages/runaipubsub"
	pubsubmocks "github.com/run-ai/runai/runai-common-packages/runaipubsub/mocks"

	authorization_utils "github.com/run-ai/runai/control-plane/authorization/pkg/client/helpers"
	"github.com/run-ai/runai/control-plane/cluster-service/pkg/config"
	"github.com/run-ai/runai/control-plane/cluster-service/pkg/services"
	cd "github.com/run-ai/runai/control-plane/cluster-service/pkg/services/cluster-delete"
	clusterinstall "github.com/run-ai/runai/control-plane/cluster-service/pkg/services/cluster-installation/mocks"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	authorization_client "github.com/run-ai/runai/control-plane/authorization/pkg/client"
	authorization_api "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/cluster-service/internal/generated/pkg/api"
	bc_mocks "github.com/run-ai/runai/control-plane/cluster-service/mocks/pkg/clients/backend"
	csc_mocks "github.com/run-ai/runai/control-plane/cluster-service/mocks/pkg/services/cluster-state-calculator"
	cluster_config "github.com/run-ai/runai/control-plane/cluster-service/pkg/config"
	idm_client "github.com/run-ai/runai/control-plane/identity-manager/pkg/client/mocks"
	org_unit_client "github.com/run-ai/runai/runai-clients/org-unit-service-client/mocks"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
	"github.com/stretchr/testify/mock"
)

var _ = Describe("ApiClustersService", Ordered, func() {
	type Mocks struct {
		clusterStateCalculator *csc_mocks.ClusterStateCalculator
		tenantSettingsClient   *bc_mocks.TenantSettingsClient
		idmClient              *idm_client.IIdm
		orgUnitClient          *org_unit_client.IOrgUnitClient
		pubsub                 *pubsubmocks.PubSuber
	}

	const (
		tenantId      = 1
		otherTenantId = 2
		mgmtTenantId  = 3
	)

	var (
		mocks              *Mocks
		clustersApiService api.ClustersAPIServicer
		nodeService        api.NodesAPIServicer
		ctx                context.Context
		otherCtx           context.Context
		mgmtCtx            context.Context

		clusters []api.DisplayedCluster
	)

	cluster1Name := clusterNameGenerator.Generate("multi-tenancy")
	cluster2Name := clusterNameGenerator.Generate("multi-tenancy")
	BeforeAll(func() {
		// Setup empty mocks
		mocks = &Mocks{
			clusterStateCalculator: &csc_mocks.ClusterStateCalculator{},
			tenantSettingsClient:   &bc_mocks.TenantSettingsClient{},
			idmClient:              &idm_client.IIdm{},
			orgUnitClient:          &org_unit_client.IOrgUnitClient{},
			pubsub:                 &pubsubmocks.PubSuber{},
		}
		pubsub.SetMock(mocks.pubsub)
		mocks.pubsub.On("Publish", mock.Anything, mock.Anything).Return(nil)
		mocks.clusterStateCalculator.On("CalculateState", mock.Anything).Return("")

		// Enable async cluster deletion
		runai_config.Set(&config.Service{PubsubInitialized: true})
		runai_config.Set(&cluster_config.AsyncClusterDelete{Enabled: true})

		tenantIdStr := strconv.Itoa(tenantId)
		ctx = context.WithValue(context.Background(), runaictx.CfTenantId, tenantIdStr)
		ctx = context.WithValue(ctx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{
			Tenants: []string{tenantIdStr},
		})
		ctx = addAuditToCtx(ctx)

		otherTenantIdStr := strconv.Itoa(otherTenantId)
		otherCtx = context.WithValue(context.Background(), runaictx.CfTenantId, otherTenantIdStr)
		otherCtx = context.WithValue(otherCtx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{
			Tenants: []string{tenantIdStr, otherTenantIdStr},
		})
		otherCtx = addAuditToCtx(otherCtx)

		mgmtTenantIdStr := strconv.Itoa(mgmtTenantId)
		mgmtCtx = context.WithValue(context.Background(), runaictx.CfTenantId, mgmtTenantIdStr)
		mgmtCtx = context.WithValue(mgmtCtx, runaictx.CfIsManagementTenant, "true")
		mgmtCtx = context.WithValue(mgmtCtx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{
			Tenants: []string{tenantIdStr, mgmtTenantIdStr},
		})
		mgmtCtx = addAuditToCtx(mgmtCtx)

		// Setup service
		clusterDeleter := cd.NewClusterDeleter(authorization_client.NewAuthorizationClient(), mocks.orgUnitClient, nil)
		clusterInstall := clusterinstall.NewClusterInstallService(GinkgoT())
		nodepoolService := services.NewNodePoolsApiService(mocks.tenantSettingsClient, mocks.orgUnitClient)

		clustersApiService = services.NewClustersApiService(mocks.clusterStateCalculator, mocks.idmClient, clusterDeleter, clusterInstall, nil, nodepoolService, mocks.orgUnitClient)
		nodeService = services.NewNodeService(mocks.tenantSettingsClient, nil)

		clusters = []api.DisplayedCluster{
			createCluster(ctx, clustersApiService, cluster1Name, "cluster-1", ""),
			createCluster(otherCtx, clustersApiService, cluster2Name, "cluster-2", ""),
		}
	})

	AfterAll(func() {
		// Returning to zero state
		mocks.orgUnitClient.On("ClusterDeleted", mock.Anything, mock.Anything).Return(nil)
		deleteAllClusters(ctx, clustersApiService)
	})

	Describe("Clusters", func() {
		It("Should return correct cluster based on authentication and authorization", func() {
			Expect(clusters[0].Name).To(Equal(cluster1Name))
			Expect(clusters[1].Name).To(Equal(cluster2Name))

			// Check that we can see the clusters in mgmt scope - only permitted once (first tenant only)
			resp, err := clustersApiService.GetClusters(mgmtCtx, "full", []string{}, false)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(resp.Code).To(Equal(200))
			Expect(resp.Body).To(HaveLen(1))

			displayedClusters, ok := resp.Body.([]api.DisplayedCluster)
			Expect(ok).To(BeTrue())
			Expect(displayedClusters[0].Name).To(Equal(cluster1Name))
			Expect(*displayedClusters[0].Domain).To(Equal("cluster-1"))

			// Check that we can see the clusters in other scope - sees only his own even if he has permitted scopes, since he is not a mgmt tenant
			resp, err = clustersApiService.GetClusters(otherCtx, "full", []string{}, false)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(resp.Code).To(Equal(200))
			Expect(resp.Body).To(HaveLen(1))

			displayedClusters, ok = resp.Body.([]api.DisplayedCluster)
			Expect(ok).To(BeTrue())
			Expect(displayedClusters[0].Name).To(Equal(cluster2Name))
			Expect(*displayedClusters[0].Domain).To(Equal("cluster-2"))
		})
		It("Filter by tenant id", func() {
			Expect(clusters[0].Name).To(Equal(cluster1Name))
			Expect(clusters[1].Name).To(Equal(cluster2Name))

			// get the tenant data using the filter
			resp, err := clustersApiService.GetClusters(mgmtCtx, "full", []string{strconv.Itoa(tenantId)}, false)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(resp.Code).To(Equal(200))
			Expect(resp.Body).To(HaveLen(1))

			displayedClusters, ok := resp.Body.([]api.DisplayedCluster)
			Expect(ok).To(BeTrue())
			Expect(displayedClusters[0].Name).To(Equal(cluster1Name))
			Expect(*displayedClusters[0].Domain).To(Equal("cluster-1"))

			// filter out the tenantId data
			resp, err = clustersApiService.GetClusters(mgmtCtx, "full", []string{strconv.Itoa(mgmtTenantId)}, false)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(resp.Code).To(Equal(200))
			Expect(resp.Body).To(HaveLen(0))
		})
		It("Tenant cannot delete cluster of another tenant but can delete their own", func() {
			// Create a new cluster for tenant 1
			clusterName := clusterNameGenerator.Generate("tenant-1")
			newCluster := createCluster(ctx, clustersApiService, clusterName, "cluster-tenant-1", "")

			// Verify the cluster exists for tenant 1
			resp, err := clustersApiService.GetClusters(ctx, "full", []string{}, false)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(resp.Code).To(Equal(200))
			displayedClusters, ok := resp.Body.([]api.DisplayedCluster)
			Expect(ok).To(BeTrue())
			Expect(len(displayedClusters)).To(Equal(2)) // Existing cluster + new cluster

			// Attempt to delete it using other tenant context
			resp, err = clustersApiService.DeleteCluster(otherCtx, newCluster.Uuid, true)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(resp.Code).To(Equal(http.StatusNotFound))

			// Verify the cluster still exists
			resp, err = clustersApiService.GetClusterByUuid(ctx, newCluster.Uuid, "full")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(resp.Code).To(Equal(http.StatusOK))

			// Attempt to delete it using correct tenant context
			mocks.orgUnitClient.On("ClusterDeleted", mock.Anything, mock.Anything).Return(nil)
			resp, err = clustersApiService.DeleteCluster(ctx, newCluster.Uuid, false)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(resp.Code).To(Equal(http.StatusAccepted))
		})
		It("Management tenant can delete cluster in its permitted scope", func() {
			// Create a new cluster for tenant 1 using ctx
			clusterName := clusterNameGenerator.Generate("tenant-1")
			newCluster := createCluster(ctx, clustersApiService, clusterName, "cluster-tenant-1", "")

			// Verify the cluster's tenantId is tenantId (1)
			Expect(newCluster.TenantId).To(Equal(int32(tenantId)))

			// Attempt to delete it using mgmtCtx
			resp, err := clustersApiService.DeleteCluster(mgmtCtx, newCluster.Uuid, false)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(resp.Code).To(Equal(http.StatusAccepted))
		})
		It("Management tenant cannot delete cluster outside its permitted scope", func() {
			// Create a new cluster for tenant 2 (otherTenantId) using otherCtx
			clusterName := clusterNameGenerator.Generate("tenant-2")
			newCluster := createCluster(otherCtx, clustersApiService, clusterName, "cluster-tenant-2", "")

			// Verify the cluster's tenantId is otherTenantId
			Expect(newCluster.TenantId).To(Equal(int32(otherTenantId)))

			// Attempt to delete it using mgmtCtx (should fail)
			resp, err := clustersApiService.DeleteCluster(mgmtCtx, newCluster.Uuid, true)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(resp.Code).To(Equal(http.StatusForbidden))

			// Attempt to delete it using otherCtx (should succeed)
			resp, err = clustersApiService.DeleteCluster(otherCtx, newCluster.Uuid, true)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(resp.Code).To(Equal(http.StatusAccepted))
		})
	})
	Describe("Nodes", func() {
		It("Should return correct nodes based on authentication and authorization", func() {
			nodeService.UpsertNode(ctx, clusters[0].Uuid, "node-1", api.NodeInfo{
				Status:    "Ready",
				NodePool:  "default",
				CreatedAt: time.Now(),
			})
			nodeService.UpsertNode(otherCtx, clusters[1].Uuid, "node-2", api.NodeInfo{
				Status:    "Ready",
				NodePool:  "default",
				CreatedAt: time.Now(),
			})

			// Check that we can see the nodes in mgmt scope - only permitted once (first tenant only)
			resp, err := nodeService.GetTenantNodes(mgmtCtx, []string{fmt.Sprintf("clusterUuid==%s", clusters[0].Uuid)}, "", "", 0, 0, "")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(resp.Code).To(Equal(200))
			nodes, ok := resp.Body.(api.Nodes)
			Expect(ok).To(BeTrue())

			Expect(nodes.Nodes).To(HaveLen(1))
			resp, err = nodeService.GetTenantNodes(otherCtx, []string{fmt.Sprintf("clusterUuid==%s", clusters[0].Uuid)}, "", "", 0, 0, "")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(resp.Code).To(Equal(200))
			nodes, ok = resp.Body.(api.Nodes)
			Expect(ok).To(BeTrue())
			Expect(nodes.Nodes).To(HaveLen(0))

			// Check that we can see the nodes in other scope - sees only his own even if he has permitted scopes, since he is not a mgmt tenant
			resp, err = nodeService.GetTenantNodes(otherCtx, []string{fmt.Sprintf("clusterUuid==%s", clusters[1].Uuid)}, "", "", 0, 0, "")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(resp.Code).To(Equal(200))
			nodes, ok = resp.Body.(api.Nodes)
			Expect(ok).To(BeTrue())
			Expect(nodes.Nodes).To(HaveLen(1))

			resp, err = nodeService.GetTenantNodes(otherCtx, []string{fmt.Sprintf("clusterUuid==%s", clusters[0].Uuid)}, "", "", 0, 0, "")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(resp.Code).To(Equal(200))
			nodes, ok = resp.Body.(api.Nodes)
			Expect(ok).To(BeTrue())
			Expect(nodes.Nodes).To(HaveLen(0))
		})
	})
})
