/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/lib/pq"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	authorization_api "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/database/models"
)

var _ = Describe("network_topology_view", Ordered, func() {
	var tenantId int32 = 5678
	var cluster *models.Cluster

	BeforeAll(func() {
		var errApi *api.Error

		cluster, errApi = CreateCluster(context.Background(), &models.Cluster{
			Common:   models.Common{},
			Name:     "nt-view-cluster",
			Version:  "2.23",
			Domain:   "test-domain",
			TenantId: tenantId,
		})
		Expect(errApi).To(BeNil())
	})

	AfterAll(func() {
		_ = DeleteCluster(context.Background(), cluster.ID.String(), tenantId)
	})

	It("aggregates nodepools per topology", func() {
		ctx := context.Background()

		createdBy := "test-user"
		// Create a topology
		topology, errApi := CreateNetworkTopology(ctx, models.NetworkTopology{
			Name:      "nt-view-topology",
			Levels:    pq.StringArray([]string{"rack", "host"}),
			Source:    NetworkTopologiesRunAiSource,
			ClusterID: cluster.ID,
			TenantID:  tenantId,
			Phase:     models.NetworkTopologyPhaseReady,
			CreatedBy: &createdBy,
		})
		Expect(errApi).To(BeNil())
		Expect(topology).ToNot(BeNil())

		// Create nodepools that reference the topology
		np1, err := CreateNodepool(ctx, models.Nodepool{
			Name:                "np-a",
			TenantId:            tenantId,
			ClusterId:           cluster.ID,
			ClusterName:         cluster.Name,
			IsDefault:           false,
			NetworkTopologyId:   &topology.ID,
			NetworkTopologyName: &topology.Name,
			CreatedBy:           createdBy,
			UpdatedBy:           createdBy,
		})
		Expect(err).To(BeNil())
		Expect(np1.ID).To(BeNumerically(">", 0))

		np2, err := CreateNodepool(ctx, models.Nodepool{
			Name:                "np-b",
			TenantId:            tenantId,
			ClusterId:           cluster.ID,
			ClusterName:         cluster.Name,
			IsDefault:           false,
			NetworkTopologyId:   &topology.ID,
			NetworkTopologyName: &topology.Name,
			CreatedBy:           createdBy,
			UpdatedBy:           createdBy,
		})
		Expect(err).To(BeNil())
		Expect(np2.ID).To(BeNumerically(">", 0))

		// Nodepool without topology should not appear
		_, err = CreateNodepool(ctx, models.Nodepool{
			Name:        "np-no-topology",
			TenantId:    tenantId,
			ClusterId:   cluster.ID,
			ClusterName: cluster.Name,
			IsDefault:   false,
			CreatedBy:   createdBy,
			UpdatedBy:   createdBy,
		})
		Expect(err).To(BeNil())

		// Query via DAO
		row, daoErr := GetNetworkTopologyViewById(ctx, tenantId, topology.ID.String())
		Expect(daoErr).To(BeNil())

		Expect(row.ID).To(Equal(topology.ID))
		Expect(row.Name).To(Equal(topology.Name))
		var gotNp []string
		Expect(json.Unmarshal(row.NodePools, &gotNp)).To(BeNil())
		Expect(gotNp).To(Equal([]string{"np-a", "np-b"}))
	})

	It("updates after nodepool deletion", func() {
		ctx := context.Background()

		createdBy := "test-user"
		// Create a topology and two nodepools
		topology, errApi := CreateNetworkTopology(ctx, models.NetworkTopology{
			Name:      "nt-view-del",
			Levels:    pq.StringArray([]string{"rack", "host"}),
			Source:    NetworkTopologiesRunAiSource,
			ClusterID: cluster.ID,
			TenantID:  tenantId,
			Phase:     models.NetworkTopologyPhaseReady,
			CreatedBy: &createdBy,
		})
		Expect(errApi).To(BeNil())

		np1, err := CreateNodepool(ctx, models.Nodepool{
			Name:                "ntdel-a",
			TenantId:            tenantId,
			ClusterId:           cluster.ID,
			ClusterName:         cluster.Name,
			IsDefault:           false,
			NetworkTopologyId:   &topology.ID,
			NetworkTopologyName: &topology.Name,
			CreatedBy:           createdBy,
			UpdatedBy:           createdBy,
		})
		Expect(err).To(BeNil())

		_, err = CreateNodepool(ctx, models.Nodepool{
			Name:                "ntdel-b",
			TenantId:            tenantId,
			ClusterId:           cluster.ID,
			ClusterName:         cluster.Name,
			IsDefault:           false,
			NetworkTopologyId:   &topology.ID,
			NetworkTopologyName: &topology.Name,
			CreatedBy:           createdBy,
			UpdatedBy:           createdBy,
		})
		Expect(err).To(BeNil())

		// Validate mid-state before deletion via DAO
		rowBefore, daoErr := GetNetworkTopologyViewById(ctx, tenantId, topology.ID.String())
		Expect(daoErr).To(BeNil())
		var gotBefore []string
		Expect(json.Unmarshal(rowBefore.NodePools, &gotBefore)).To(BeNil())
		Expect(gotBefore).To(Equal([]string{"ntdel-a", "ntdel-b"}))

		// Delete one nodepool
		err = DeleteNodepool(ctx, tenantId, createdBy, fmt.Sprintf("%d", np1.ID))
		Expect(err).To(BeNil())

		// Query via DAO
		rowAfter, daoErr := GetNetworkTopologyViewById(ctx, tenantId, topology.ID.String())
		Expect(daoErr).To(BeNil())
		var gotAfter []string
		Expect(json.Unmarshal(rowAfter.NodePools, &gotAfter)).To(BeNil())
		Expect(gotAfter).To(Equal([]string{"ntdel-b"}))
	})

	It("updates after adding a nodepool", func() {
		ctx := context.Background()

		createdBy := "test-user"
		// Create a topology and one initial nodepool
		topology, errApi := CreateNetworkTopology(ctx, models.NetworkTopology{
			Name:      "nt-view-add",
			Levels:    pq.StringArray([]string{"rack", "host"}),
			Source:    NetworkTopologiesRunAiSource,
			ClusterID: cluster.ID,
			TenantID:  tenantId,
			Phase:     models.NetworkTopologyPhaseReady,
			CreatedBy: &createdBy,
		})
		Expect(errApi).To(BeNil())

		_, err := CreateNodepool(ctx, models.Nodepool{
			Name:                "ntadd-a",
			TenantId:            tenantId,
			ClusterId:           cluster.ID,
			ClusterName:         cluster.Name,
			IsDefault:           false,
			NetworkTopologyId:   &topology.ID,
			NetworkTopologyName: &topology.Name,
			CreatedBy:           createdBy,
			UpdatedBy:           createdBy,
		})
		Expect(err).To(BeNil())

		// Verify initial state via DAO
		row, daoErr := GetNetworkTopologyViewById(ctx, tenantId, topology.ID.String())
		Expect(daoErr).To(BeNil())
		var gotInit []string
		Expect(json.Unmarshal(row.NodePools, &gotInit)).To(BeNil())
		Expect(gotInit).To(Equal([]string{"ntadd-a"}))

		// Add another nodepool referencing the same topology
		_, err = CreateNodepool(ctx, models.Nodepool{
			Name:                "ntadd-b",
			TenantId:            tenantId,
			ClusterId:           cluster.ID,
			ClusterName:         cluster.Name,
			IsDefault:           false,
			NetworkTopologyId:   &topology.ID,
			NetworkTopologyName: &topology.Name,
			CreatedBy:           createdBy,
			UpdatedBy:           createdBy,
		})
		Expect(err).To(BeNil())

		// Verify updated state via DAO
		row, daoErr = GetNetworkTopologyViewById(ctx, tenantId, topology.ID.String())
		Expect(daoErr).To(BeNil())
		var gotAfter []string
		Expect(json.Unmarshal(row.NodePools, &gotAfter)).To(BeNil())
		Expect(gotAfter).To(Equal([]string{"ntadd-a", "ntadd-b"}))
	})

	It("lists topologies and includes nodepools", func() {
		ctx := context.Background()

		createdBy := "test-user"
		topology, errApi := CreateNetworkTopology(ctx, models.NetworkTopology{
			Name:      "nt-view-list",
			Levels:    pq.StringArray([]string{"rack", "host"}),
			Source:    NetworkTopologiesRunAiSource,
			ClusterID: cluster.ID,
			TenantID:  tenantId,
			Phase:     models.NetworkTopologyPhaseReady,
			CreatedBy: &createdBy,
		})
		Expect(errApi).To(BeNil())

		_, err := CreateNodepool(ctx, models.Nodepool{
			Name:                "list-a",
			TenantId:            tenantId,
			ClusterId:           cluster.ID,
			ClusterName:         cluster.Name,
			IsDefault:           false,
			NetworkTopologyId:   &topology.ID,
			NetworkTopologyName: &topology.Name,
			CreatedBy:           createdBy,
			UpdatedBy:           createdBy,
		})
		Expect(err).To(BeNil())
		_, err = CreateNodepool(ctx, models.Nodepool{
			Name:                "list-b",
			TenantId:            tenantId,
			ClusterId:           cluster.ID,
			ClusterName:         cluster.Name,
			IsDefault:           false,
			NetworkTopologyId:   &topology.ID,
			NetworkTopologyName: &topology.Name,
			CreatedBy:           createdBy,
			UpdatedBy:           createdBy,
		})
		Expect(err).To(BeNil())

		permitted := authorization_api.PermittedScopes{
			Clusters: []string{cluster.ID.String()},
		}
		rows, daoErr := GetNetworkTopologiesView(ctx, tenantId, &permitted, nil)
		Expect(daoErr).To(BeNil())

		var found *models.NetworkTopologyView
		for i := range rows {
			if rows[i].ID == topology.ID {
				found = &rows[i]
				break
			}
		}
		Expect(found).ToNot(BeNil())
		var names []string
		Expect(json.Unmarshal(found.NodePools, &names)).To(BeNil())
		Expect(names).To(Equal([]string{"list-a", "list-b"}))
	})

	It("returns empty nodepools list when no associations exist", func() {
		ctx := context.Background()

		createdBy := "test-user"
		topology, errApi := CreateNetworkTopology(ctx, models.NetworkTopology{
			Name:      "nt-view-empty",
			Levels:    pq.StringArray([]string{"rack", "host"}),
			Source:    NetworkTopologiesRunAiSource,
			ClusterID: cluster.ID,
			TenantID:  tenantId,
			Phase:     models.NetworkTopologyPhaseReady,
			CreatedBy: &createdBy,
		})
		Expect(errApi).To(BeNil())

		row, daoErr := GetNetworkTopologyViewById(ctx, tenantId, topology.ID.String())
		Expect(daoErr).To(BeNil())
		var gotEmpty []string
		Expect(json.Unmarshal(row.NodePools, &gotEmpty)).To(BeNil())
		Expect(gotEmpty).To(Equal([]string{}))
	})

	It("removes view row when topology is deleted", func() {
		ctx := context.Background()

		createdBy := "test-user"
		topology, errApi := CreateNetworkTopology(ctx, models.NetworkTopology{
			Name:      "nt-view-del-topology",
			Levels:    pq.StringArray([]string{"rack", "host"}),
			Source:    NetworkTopologiesRunAiSource,
			ClusterID: cluster.ID,
			TenantID:  tenantId,
			Phase:     models.NetworkTopologyPhaseReady,
			CreatedBy: &createdBy,
		})
		Expect(errApi).To(BeNil())

		// Attach a nodepool (to ensure join works before deletion)
		_, err := CreateNodepool(ctx, models.Nodepool{
			Name:                "ntdelrow-a",
			TenantId:            tenantId,
			ClusterId:           cluster.ID,
			ClusterName:         cluster.Name,
			IsDefault:           false,
			NetworkTopologyId:   &topology.ID,
			NetworkTopologyName: &topology.Name,
			CreatedBy:           createdBy,
			UpdatedBy:           createdBy,
		})
		Expect(err).To(BeNil())

		// Validate present via DAO
		row, daoErr := GetNetworkTopologyViewById(ctx, tenantId, topology.ID.String())
		Expect(daoErr).To(BeNil())
		Expect(row).ToNot(BeNil())

		// Delete the topology
		apiErr := DeleteNetworkTopology(ctx, tenantId, cluster.ID, topology.ID.String())
		Expect(apiErr).To(BeNil())

		// Expect row removed (DAO returns error)
		row, daoErr = GetNetworkTopologyViewById(ctx, tenantId, topology.ID.String())
		Expect(daoErr).ToNot(BeNil())
	})
})
