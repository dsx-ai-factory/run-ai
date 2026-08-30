package dao

import (
	"context"
	"fmt"
	"net/http"

	"github.com/run-ai/runai/control-plane/cluster-service/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/cluster-service/pkg/database/models"
	"github.com/run-ai/runai/runai-common-packages/runaidb"

	"github.com/google/uuid"
	"github.com/lib/pq"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("NetworkTopology DAO tests", Ordered, func() {
	var testNetworkTopologyClusterId uuid.UUID
	testTenantId := int32(12345)

	BeforeAll(func() {
		cluster, err := CreateCluster(context.Background(), &models.Cluster{
			Common:   models.Common{},
			Name:     "cluster-1",
			Version:  "2.23",
			Domain:   "test-domain",
			TenantId: testTenantId,
		})

		Expect(err).To(BeNil())
		testNetworkTopologyClusterId = cluster.ID
	})

	Context("when both name and topology ID are nil", func() {
		It("should return bad request error", func() {
			ctx := context.Background()

			err := deleteNetworkTopology(ctx, nil, testTenantId, testNetworkTopologyClusterId, nil, nil)

			Expect(err).NotTo(BeNil())
			Expect(err.Code).To(Equal(int32(http.StatusBadRequest)))
			Expect(err.Message).To(ContainSubstring("Failed to delete network topology for clusterId"))
			Expect(err.Details).To(ContainSubstring("either name or networkTopologyId must be specified"))
		})
	})

	Context("when deleting network topology with attached nodepools", func() {
		var testNetworkTopology *models.NetworkTopology
		var testNodepools []models.Nodepool

		BeforeEach(func() {
			ctx := context.Background()

			// Create a test network topology
			testNetworkTopology = &models.NetworkTopology{
				Name:      "test-topology-for-nodepools",
				Levels:    []string{"level1", "level2"},
				Source:    NetworkTopologiesRunAiSource,
				ClusterID: testNetworkTopologyClusterId,
				TenantID:  testTenantId,
				Phase:     models.NetworkTopologyPhaseReady,
			}

			createdTopology, err := CreateNetworkTopology(ctx, *testNetworkTopology)
			Expect(err).To(BeNil())
			testNetworkTopology = createdTopology

			// Create test nodepools with network topology references
			nodepool1 := models.Nodepool{
				Name:                "nodepool-with-topology-1",
				TenantId:            testTenantId,
				ClusterId:           testNetworkTopologyClusterId,
				ClusterName:         "cluster-1",
				IsDefault:           false,
				NetworkTopologyId:   &testNetworkTopology.ID,
				NetworkTopologyName: &testNetworkTopology.Name,
				CreatedBy:           "test-user",
				UpdatedBy:           "test-user",
			}

			nodepool2 := models.Nodepool{
				Name:                "nodepool-with-topology-2",
				TenantId:            testTenantId,
				ClusterId:           testNetworkTopologyClusterId,
				ClusterName:         "cluster-1",
				IsDefault:           false,
				NetworkTopologyId:   &testNetworkTopology.ID,
				NetworkTopologyName: &testNetworkTopology.Name,
				CreatedBy:           "test-user",
				UpdatedBy:           "test-user",
			}

			// Create nodepool without network topology (should not be affected)
			nodepool3 := models.Nodepool{
				Name:        "nodepool-without-topology",
				TenantId:    testTenantId,
				ClusterId:   testNetworkTopologyClusterId,
				ClusterName: "cluster-1",
				IsDefault:   false,
				CreatedBy:   "test-user",
				UpdatedBy:   "test-user",
			}

			createdNodepool1, createErr := CreateNodepool(ctx, nodepool1)
			Expect(createErr).To(BeNil())
			testNodepools = append(testNodepools, createdNodepool1)

			createdNodepool2, createErr := CreateNodepool(ctx, nodepool2)
			Expect(createErr).To(BeNil())
			testNodepools = append(testNodepools, createdNodepool2)

			createdNodepool3, createErr := CreateNodepool(ctx, nodepool3)
			Expect(createErr).To(BeNil())
			testNodepools = append(testNodepools, createdNodepool3)

			// Verify nodepools were created with network topology references
			for i := 0; i < 2; i++ {
				Expect(testNodepools[i].NetworkTopologyId).To(Not(BeNil()))
				Expect(testNodepools[i].NetworkTopologyName).To(Not(BeNil()))
				Expect(*testNodepools[i].NetworkTopologyId).To(Equal(testNetworkTopology.ID))
				Expect(*testNodepools[i].NetworkTopologyName).To(Equal(testNetworkTopology.Name))
			}

			// Verify third nodepool has no network topology references
			Expect(testNodepools[2].NetworkTopologyId).To(BeNil())
			Expect(testNodepools[2].NetworkTopologyName).To(BeNil())
		})

		AfterEach(func() {
			ctx := context.Background()

			// Clean up nodepools
			for _, nodepool := range testNodepools {
				err := DeleteNodepool(ctx, testTenantId, "test-user", fmt.Sprintf("%d", nodepool.ID))
				Expect(err).To(BeNil())
			}
			testNodepools = nil
		})

		It("should clear network topology references from nodepools when deleting by ID", func() {
			ctx := context.Background()
			tx := runaidb.DB().WithContext(ctx).Begin()

			topologyIdStr := testNetworkTopology.ID.String()

			// Delete the network topology
			err := deleteNetworkTopology(ctx, tx, testTenantId, testNetworkTopologyClusterId, &topologyIdStr, nil)
			Expect(err).To(BeNil())

			// Commit the transaction
			commitErr := tx.Commit()
			Expect(commitErr).To(BeNil())

			// Verify network topology is deleted
			deletedTopology, getErr := GetNetworkTopologyById(ctx, testTenantId, testNetworkTopology.ID.String())
			Expect(getErr).NotTo(BeNil())
			Expect(deletedTopology).To(BeNil())

			// Verify nodepools have network topology references cleared
			for i := 0; i < 2; i++ {
				updatedNodepool, getErr := GetNodepoolById(ctx, testTenantId, fmt.Sprintf("%d", testNodepools[i].ID))
				Expect(getErr).To(BeNil())
				Expect(updatedNodepool.NetworkTopologyId).To(BeNil())
				Expect(updatedNodepool.NetworkTopologyName).To(BeNil())
			}

			// Verify nodepool without topology is unchanged
			unchangedNodepool, unchangedErr := GetNodepoolById(ctx, testTenantId, fmt.Sprintf("%d", testNodepools[2].ID))
			Expect(unchangedErr).To(BeNil())
			Expect(unchangedNodepool.NetworkTopologyId).To(BeNil())
			Expect(unchangedNodepool.NetworkTopologyName).To(BeNil())
		})

		It("should clear network topology references from nodepools when deleting by name", func() {
			ctx := context.Background()
			tx := runaidb.DB().WithContext(ctx).Begin()

			topologyName := testNetworkTopology.Name

			// Delete the network topology by name
			err := deleteNetworkTopology(ctx, tx, testTenantId, testNetworkTopologyClusterId, nil, &topologyName)
			Expect(err).To(BeNil())

			// Commit the transaction
			commitErr := tx.Commit()
			Expect(commitErr).To(BeNil())

			// Verify network topology is deleted
			deletedTopology, getErr := GetNetworkTopologyById(ctx, testTenantId, testNetworkTopology.ID.String())
			Expect(getErr).NotTo(BeNil())
			Expect(deletedTopology).To(BeNil())

			// Verify nodepools have network topology references cleared
			for i := 0; i < 2; i++ {
				updatedNodepool, getErr := GetNodepoolById(ctx, testTenantId, fmt.Sprintf("%d", testNodepools[i].ID))
				Expect(getErr).To(BeNil())
				Expect(updatedNodepool.NetworkTopologyId).To(BeNil())
				Expect(updatedNodepool.NetworkTopologyName).To(BeNil())
			}

			// Verify nodepool without topology is unchanged
			unchangedNodepool, unchangedErr := GetNodepoolById(ctx, testTenantId, fmt.Sprintf("%d", testNodepools[2].ID))
			Expect(unchangedErr).To(BeNil())
			Expect(unchangedNodepool.NetworkTopologyId).To(BeNil())
			Expect(unchangedNodepool.NetworkTopologyName).To(BeNil())
		})

		It("should clear network topology references from nodepools when bulk deleting", func() {
			ctx := context.Background()
			tx := runaidb.DB().WithContext(ctx).Begin()

			topologyIDs := []uuid.UUID{testNetworkTopology.ID}

			// Bulk delete the network topology
			err := bulkDeleteNetworkTopologies(ctx, tx, testTenantId, testNetworkTopologyClusterId, topologyIDs)
			Expect(err).To(BeNil())

			// Commit the transaction
			commitErr := tx.Commit()
			Expect(commitErr).To(BeNil())

			// Verify network topology is deleted
			deletedTopology, getErr := GetNetworkTopologyById(ctx, testTenantId, testNetworkTopology.ID.String())
			Expect(getErr).NotTo(BeNil())
			Expect(deletedTopology).To(BeNil())

			// Verify nodepools have network topology references cleared
			for i := 0; i < 2; i++ {
				updatedNodepool, getErr := GetNodepoolById(ctx, testTenantId, fmt.Sprintf("%d", testNodepools[i].ID))
				Expect(getErr).To(BeNil())
				Expect(updatedNodepool.NetworkTopologyId).To(BeNil())
				Expect(updatedNodepool.NetworkTopologyName).To(BeNil())
			}

			// Verify nodepool without topology is unchanged
			unchangedNodepool, unchangedErr := GetNodepoolById(ctx, testTenantId, fmt.Sprintf("%d", testNodepools[2].ID))
			Expect(unchangedErr).To(BeNil())
			Expect(unchangedNodepool.NetworkTopologyId).To(BeNil())
			Expect(unchangedNodepool.NetworkTopologyName).To(BeNil())
		})
	})

	Context("when updating a cluster network topology", func() {
		var existingTopology *models.NetworkTopology
		originalLevels := []string{"original-level1", "original-level2"}

		BeforeEach(func() {
			ctx := context.Background()
			existingTopology = &models.NetworkTopology{
				Name:      fmt.Sprintf("update-test-topology-%s", uuid.NewString()),
				Levels:    pq.StringArray(originalLevels),
				Source:    NetworkTopologiesClusterSource,
				ClusterID: testNetworkTopologyClusterId,
				TenantID:  testTenantId,
				Phase:     models.NetworkTopologyPhaseReady,
				Grove: &models.Grove{
					Name:    "original-grove",
					Domains: pq.StringArray{"original-domain"},
					Conditions: models.GroveTopologyConditions{
						{Type: "Ready", Status: "False", Reason: "Initial", Message: "starting", ObservedGeneration: 1},
					},
				},
			}
			created, err := CreateNetworkTopology(ctx, *existingTopology)
			Expect(err).To(BeNil())
			existingTopology = created
		})

		AfterEach(func() {
			ctx := context.Background()
			tx := runaidb.DB().WithContext(ctx).Begin()
			topologyId := existingTopology.ID.String()
			_ = deleteNetworkTopology(ctx, tx, testTenantId, testNetworkTopologyClusterId, &topologyId, nil)
			Expect(tx.Commit()).To(BeNil())
		})

		It("should return bad request when name is empty", func() {
			ctx := context.Background()
			tx := runaidb.DB().WithContext(ctx).Begin()
			defer func() { _ = tx.Rollback() }()

			err := updateClusterNetworkTopology(ctx, tx, testTenantId, testNetworkTopologyClusterId, api.NetworkTopologySyncFields{})
			Expect(err).NotTo(BeNil())
			Expect(err.Code).To(Equal(int32(http.StatusBadRequest)))
		})

		It("should replace grove fields and leave levels untouched", func() {
			ctx := context.Background()
			tx := runaidb.DB().WithContext(ctx).Begin()

			fields := api.NetworkTopologySyncFields{
				Name:   existingTopology.Name,
				Source: NetworkTopologiesClusterSource,
				Levels: []string{"ignored-level"},
				Grove: &api.NetworkTopologySyncGrove{
					Name:    "updated-grove",
					Domains: []string{"updated-d1", "updated-d2"},
					Conditions: []api.NetworkTopologySyncGroveCondition{
						{Type: "Ready", Status: "True", Reason: "Reconciled", Message: "ok", ObservedGeneration: 7},
					},
				},
			}

			err := updateClusterNetworkTopology(ctx, tx, testTenantId, testNetworkTopologyClusterId, fields)
			Expect(err).To(BeNil())
			Expect(tx.Commit()).To(BeNil())

			updated, getErr := GetNetworkTopologyById(ctx, testTenantId, existingTopology.ID.String())
			Expect(getErr).To(BeNil())
			Expect(updated).NotTo(BeNil())
			Expect([]string(updated.Levels)).To(Equal(originalLevels))
			Expect(updated.Source).To(Equal(NetworkTopologiesClusterSource))
			Expect(updated.Grove).NotTo(BeNil())
			Expect(updated.Grove.Name).To(Equal("updated-grove"))
			Expect([]string(updated.Grove.Domains)).To(Equal([]string{"updated-d1", "updated-d2"}))
			Expect(updated.Grove.Conditions).To(HaveLen(1))
			Expect(updated.Grove.Conditions[0].Type).To(Equal("Ready"))
			Expect(updated.Grove.Conditions[0].Status).To(Equal("True"))
			Expect(updated.Grove.Conditions[0].Reason).To(Equal("Reconciled"))
			Expect(updated.Grove.Conditions[0].Message).To(Equal("ok"))
			Expect(updated.Grove.Conditions[0].ObservedGeneration).To(Equal(int64(7)))
		})

		It("should clear grove fields when grove is nil in the request", func() {
			ctx := context.Background()
			tx := runaidb.DB().WithContext(ctx).Begin()

			fields := api.NetworkTopologySyncFields{
				Name:   existingTopology.Name,
				Source: NetworkTopologiesClusterSource,
				Levels: originalLevels,
				Grove:  nil,
			}

			err := updateClusterNetworkTopology(ctx, tx, testTenantId, testNetworkTopologyClusterId, fields)
			Expect(err).To(BeNil())
			Expect(tx.Commit()).To(BeNil())

			updated, getErr := GetNetworkTopologyById(ctx, testTenantId, existingTopology.ID.String())
			Expect(getErr).To(BeNil())
			Expect(updated).NotTo(BeNil())
			Expect([]string(updated.Levels)).To(Equal(originalLevels))
			Expect(updated.Grove).To(BeNil())
		})

		It("should not update topologies of a different source", func() {
			ctx := context.Background()
			runaiTopology := &models.NetworkTopology{
				Name:      fmt.Sprintf("runai-topology-%s", uuid.NewString()),
				Levels:    pq.StringArray{"level1"},
				Source:    NetworkTopologiesRunAiSource,
				ClusterID: testNetworkTopologyClusterId,
				TenantID:  testTenantId,
				Phase:     models.NetworkTopologyPhaseReady,
			}
			created, createErr := CreateNetworkTopology(ctx, *runaiTopology)
			Expect(createErr).To(BeNil())
			runaiTopology = created
			defer func() {
				tx := runaidb.DB().WithContext(ctx).Begin()
				topoId := runaiTopology.ID.String()
				_ = deleteNetworkTopology(ctx, tx, testTenantId, testNetworkTopologyClusterId, &topoId, nil)
				Expect(tx.Commit()).To(BeNil())
			}()

			tx := runaidb.DB().WithContext(ctx).Begin()

			fields := api.NetworkTopologySyncFields{
				Name:   runaiTopology.Name,
				Source: NetworkTopologiesClusterSource,
				Grove: &api.NetworkTopologySyncGrove{
					Name: "should-not-stick",
				},
			}

			err := updateClusterNetworkTopology(ctx, tx, testTenantId, testNetworkTopologyClusterId, fields)
			Expect(err).To(BeNil())
			Expect(tx.Commit()).To(BeNil())

			fetched, getErr := GetNetworkTopologyById(ctx, testTenantId, runaiTopology.ID.String())
			Expect(getErr).To(BeNil())
			Expect(fetched.Grove).To(BeNil())
		})

		It("should be a no-op when the topology does not exist", func() {
			ctx := context.Background()
			tx := runaidb.DB().WithContext(ctx).Begin()

			fields := api.NetworkTopologySyncFields{
				Name:   "non-existent-topology",
				Source: NetworkTopologiesClusterSource,
				Grove:  &api.NetworkTopologySyncGrove{Name: "ignored"},
			}

			err := updateClusterNetworkTopology(ctx, tx, testTenantId, testNetworkTopologyClusterId, fields)
			Expect(err).To(BeNil())
			Expect(tx.Commit()).To(BeNil())
		})
	})

})
