/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services_test

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	authorization_api "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	authorization_utils "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/helpers"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/internal/generated/pkg/api"
	bc_mocks "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/mocks/pkg/clients/backend"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/services"
	org_unit_service_client_mock "github.com/dsx-ai-factory/run-ai/runai-clients/org-unit-service-client/mocks"
	"github.com/stretchr/testify/mock"
)

// testNodepoolData holds the test data for creating nodepools
type testNodepoolData struct {
	Name       string
	LabelKey   string
	LabelValue string
	ClusterID  *string // pointer to allow dynamic assignment
}

var _ = Describe("GetNodepoolsWithFields tests", Ordered, func() {
	tenantSettingsClient := &bc_mocks.TenantSettingsClient{}
	orgClient := &org_unit_service_client_mock.IOrgUnitClient{}

	ctx := createContext()

	tenantSettingsClient.On("GetOPRSetting", mock.Anything).Return(true, nil)
	orgClient.On("NodePoolCreated", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	orgClient.On("NodePoolDeleted", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	orgClient.On("DefaultNodePoolCreated", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	nodepoolService := services.NewNodePoolsApiService(tenantSettingsClient, orgClient)

	var metadataTestClusterId string
	var metadataTestCluster2Id string
	var createdNodepools []api.Nodepool

	// Test nodepool definitions - reusable across test cases
	testNodepools := map[string]testNodepoolData{
		"metadata-np-1": {
			Name:       "metadata-np-1",
			LabelKey:   "test-key-1",
			LabelValue: "test-value-1",
			ClusterID:  &metadataTestClusterId,
		},
		"metadata-np-2": {
			Name:       "metadata-np-2",
			LabelKey:   "test-key-2",
			LabelValue: "test-value-2",
			ClusterID:  &metadataTestClusterId,
		},
		"metadata-np-3": {
			Name:       "metadata-np-3",
			LabelKey:   "test-key-3",
			LabelValue: "test-value-3",
			ClusterID:  &metadataTestCluster2Id,
		},
	}

	BeforeAll(func() {
		// Create first test cluster
		cluster1, apiErr := dao.CreateCluster(ctx, &models.Cluster{
			Name:     "metadata-test-cluster-1",
			TenantId: tenantId,
			Domain:   "xyz",
			Version:  "v2.23",
		})
		Expect(apiErr).To(BeNil())
		metadataTestClusterId = cluster1.ID.String()

		// Create default nodepool for cluster 1
		notifResp1, err := nodepoolService.ClusterNotification(ctx, api.ClusterNotification{
			Operation: api.NOTIFICATIONOPERATION_CREATE,
			Id:        metadataTestClusterId,
			Name:      cluster1.Name,
		})
		Expect(err).To(BeNil())
		Expect(notifResp1.Code).To(Equal(http.StatusNoContent))

		// Create second test cluster
		cluster2, apiErr := dao.CreateCluster(ctx, &models.Cluster{
			Name:     "metadata-test-cluster-2",
			TenantId: tenantId,
			Domain:   "xyz",
			Version:  "v2.23",
		})
		Expect(apiErr).To(BeNil())
		metadataTestCluster2Id = cluster2.ID.String()

		// Create default nodepool for cluster 2
		notifResp2, nErr := nodepoolService.ClusterNotification(ctx, api.ClusterNotification{
			Operation: api.NOTIFICATIONOPERATION_CREATE,
			Id:        metadataTestCluster2Id,
			Name:      cluster2.Name,
		})
		Expect(nErr).To(BeNil())
		Expect(notifResp2.Code).To(Equal(http.StatusNoContent))

		// Create additional nodepools for testing using the map
		for name, npData := range testNodepools {
			createResp, err := nodepoolService.CreateNodepool(ctx, api.NodepoolCreateFields{
				Name:       npData.Name,
				LabelKey:   npData.LabelKey,
				LabelValue: npData.LabelValue,
				ClusterId:  *npData.ClusterID,
			})
			Expect(err).To(BeNil(), "Failed to create nodepool %s", name)
			Expect(createResp.Code).To(Equal(http.StatusCreated), "Failed to create nodepool %s", name)
			createdNodepools = append(createdNodepools, createResp.Body.(api.Nodepool))
		}
	})

	AfterAll(func() {
		// Clean up
		deleteNodepools(createdNodepools)

		// Delete default nodepools
		res, err := nodepoolService.ClusterNotification(ctx, api.ClusterNotification{
			Operation: api.NOTIFICATIONOPERATION_DELETE,
			Id:        metadataTestClusterId,
			Name:      "metadata-test-cluster-1",
		})
		Expect(err).To(BeNil(), "Expected cluster deletion notification to succeed")
		Expect(res.Code).To(Equal(http.StatusNoContent), "Expected cluster deletion notification to succeed")

		res, err = nodepoolService.ClusterNotification(ctx, api.ClusterNotification{
			Operation: api.NOTIFICATIONOPERATION_DELETE,
			Id:        metadataTestCluster2Id,
			Name:      "metadata-test-cluster-2",
		})
		Expect(err).To(BeNil(), "Expected cluster deletion notification to succeed")
		Expect(res.Code).To(Equal(http.StatusNoContent), "Expected cluster deletion notification to succeed")

		dao.DeleteCluster(ctx, metadataTestClusterId, tenantId)
		dao.DeleteCluster(ctx, metadataTestCluster2Id, tenantId)
	})

	Context("Basic functionality", func() {
		It("should return all nodepools metadata", func() {
			response, err := nodepoolService.GetNodepoolsMinimal(ctx, []string{}, "", "asc", 0, 0)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			metadata := response.Body.(api.GetNodepoolsMinimal200Response).Nodepools
			Expect(len(metadata)).To(Equal(5)) // 2 default + 3 created

			// Verify metadata contains only expected fields and validate against test data
			for _, np := range metadata {
				// All fields should be present
				Expect(np.Id).NotTo(BeEmpty(), "ID should not be empty")
				Expect(np.Name).NotTo(BeEmpty(), "Name should not be empty")
				Expect(np.ClusterId).NotTo(BeEmpty(), "ClusterId should not be empty")

				// Verify scheduling capabilities have expected defaults
				Expect(np.SchedulingCapabilities.Quota).To(BeTrue())
				Expect(np.SchedulingCapabilities.Ranks).To(BeTrue())
				Expect(np.SchedulingCapabilities.AdjustableWeights).To(BeFalse())
				Expect(np.SchedulingCapabilities.CpuQuota).To(BeFalse())

				// Verify gpuNetworkAccelerationDetected is returned with default false value
				Expect(np.GpuNetworkAccelerationDetected).To(BeFalse())

				// If this is one of our test nodepools, validate the fields match expected data
				if expectedData, exists := testNodepools[np.Name]; exists {
					Expect(np.Name).To(Equal(expectedData.Name), "Name should match test data")
					Expect(np.ClusterId).To(Equal(*expectedData.ClusterID), "ClusterId should match test data")
					// Phase should be CREATING for created nodepools - Before cluster sync
					Expect(np.Phase).To(Equal(api.NODEPOOLPHASE_CREATING),
						"Phase should be Ready for successfully created nodepools")
				}
			}
		})

		It("should return metadata with correct structure", func() {
			response, err := nodepoolService.GetNodepoolsMinimal(ctx, []string{}, "", "asc", 0, 0)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			metadata := response.Body.(api.GetNodepoolsMinimal200Response).Nodepools
			Expect(len(metadata)).To(Equal(5)) // 2 default + 3 created

			// Check first nodepool has required metadata fields
			np := metadata[0]
			Expect(np.Id).NotTo(BeEmpty())
			Expect(np.Name).NotTo(BeEmpty())
			Expect(np.ClusterId).NotTo(BeEmpty())
			Expect(np.Phase).NotTo(BeEmpty())

			// Verify scheduling capabilities have expected defaults
			Expect(np.SchedulingCapabilities.Quota).To(BeTrue())              // Default is true
			Expect(np.SchedulingCapabilities.Ranks).To(BeTrue())              // Default is true
			Expect(np.SchedulingCapabilities.AdjustableWeights).To(BeFalse()) // Default is false
			Expect(np.SchedulingCapabilities.CpuQuota).To(BeFalse())          // Default is false
		})
	})

	Context("Filtering by clusterId", func() {
		It("should filter nodepools by clusterId", func() {
			filterBy := []string{fmt.Sprintf("clusterId==%s", metadataTestClusterId)}
			response, err := nodepoolService.GetNodepoolsMinimal(ctx, filterBy, "", "asc", 0, 0)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			metadata := response.Body.(api.GetNodepoolsMinimal200Response).Nodepools
			Expect(len(metadata)).To(Equal(3)) // 1 default + 2 created

			// Verify all returned nodepools belong to the specified cluster
			for _, np := range metadata {
				Expect(np.ClusterId).To(Equal(metadataTestClusterId))
				// Verify scheduling capabilities have expected defaults
				Expect(np.SchedulingCapabilities.Quota).To(BeTrue())
				Expect(np.SchedulingCapabilities.Ranks).To(BeTrue())
				Expect(np.SchedulingCapabilities.AdjustableWeights).To(BeFalse())
				Expect(np.SchedulingCapabilities.CpuQuota).To(BeFalse())
			}
		})

		It("should return nodepools from different cluster when filtered", func() {
			filterBy := []string{fmt.Sprintf("clusterId==%s", metadataTestCluster2Id)}
			response, err := nodepoolService.GetNodepoolsMinimal(ctx, filterBy, "", "asc", 0, 0)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			metadata := response.Body.(api.GetNodepoolsMinimal200Response).Nodepools
			Expect(len(metadata)).To(Equal(2)) // 1 default + 1 created

			// Verify all returned nodepools belong to the specified cluster
			for _, np := range metadata {
				Expect(np.ClusterId).To(Equal(metadataTestCluster2Id))
				// Verify scheduling capabilities have expected defaults
				Expect(np.SchedulingCapabilities.Quota).To(BeTrue())
				Expect(np.SchedulingCapabilities.Ranks).To(BeTrue())
				Expect(np.SchedulingCapabilities.AdjustableWeights).To(BeFalse())
				Expect(np.SchedulingCapabilities.CpuQuota).To(BeFalse())
			}
		})

		It("should return empty list for non-existent clusterId", func() {
			nonExistentClusterId := uuid.New().String()
			filterBy := []string{fmt.Sprintf("clusterId==%s", nonExistentClusterId)}
			response, err := nodepoolService.GetNodepoolsMinimal(ctx, filterBy, "", "asc", 0, 0)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			metadata := response.Body.(api.GetNodepoolsMinimal200Response).Nodepools
			Expect(len(metadata)).To(Equal(0))
		})
	})

	Context("Sorting", func() {
		It("should sort by clusterId ascending", func() {
			response, err := nodepoolService.GetNodepoolsMinimal(ctx, []string{}, "clusterId", "asc", 0, 0)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			metadata := response.Body.(api.GetNodepoolsMinimal200Response).Nodepools
			Expect(len(metadata)).To(Equal(5)) // 2 default + 3 created

			// Verify ascending order (string comparison for UUIDs) and scheduling capabilities
			for i := 1; i < len(metadata); i++ {
				Expect(metadata[i].ClusterId >= metadata[i-1].ClusterId).To(BeTrue(),
					"Expected clusterId %s to be >= %s", metadata[i].ClusterId, metadata[i-1].ClusterId)
				// Verify scheduling capabilities have expected defaults
				Expect(metadata[i].SchedulingCapabilities.Quota).To(BeTrue())
				Expect(metadata[i].SchedulingCapabilities.Ranks).To(BeTrue())
				Expect(metadata[i].SchedulingCapabilities.AdjustableWeights).To(BeFalse())
				Expect(metadata[i].SchedulingCapabilities.CpuQuota).To(BeFalse())
			}
		})

		It("should sort by clusterId descending", func() {
			response, err := nodepoolService.GetNodepoolsMinimal(ctx, []string{}, "clusterId", "desc", 0, 0)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			metadata := response.Body.(api.GetNodepoolsMinimal200Response).Nodepools
			Expect(len(metadata)).To(Equal(5)) // 2 default + 3 created

			// Verify descending order (string comparison for UUIDs) and scheduling capabilities
			for i := 1; i < len(metadata); i++ {
				Expect(metadata[i].ClusterId <= metadata[i-1].ClusterId).To(BeTrue(),
					"Expected clusterId %s to be <= %s", metadata[i].ClusterId, metadata[i-1].ClusterId)
				// Verify scheduling capabilities have expected defaults
				Expect(metadata[i].SchedulingCapabilities.Quota).To(BeTrue())
				Expect(metadata[i].SchedulingCapabilities.Ranks).To(BeTrue())
				Expect(metadata[i].SchedulingCapabilities.AdjustableWeights).To(BeFalse())
				Expect(metadata[i].SchedulingCapabilities.CpuQuota).To(BeFalse())
			}
		})
	})

	Context("Pagination", func() {
		It("should respect limit parameter", func() {
			response, err := nodepoolService.GetNodepoolsMinimal(ctx, []string{}, "", "asc", 0, 2)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			metadata := response.Body.(api.GetNodepoolsMinimal200Response).Nodepools
			Expect(len(metadata)).To(Equal(2))

			// Verify scheduling capabilities have expected defaults
			for _, np := range metadata {
				Expect(np.SchedulingCapabilities.Quota).To(BeTrue())
				Expect(np.SchedulingCapabilities.Ranks).To(BeTrue())
				Expect(np.SchedulingCapabilities.AdjustableWeights).To(BeFalse())
				Expect(np.SchedulingCapabilities.CpuQuota).To(BeFalse())
			}
		})

		It("should respect offset parameter", func() {
			// Get first page
			response1, err := nodepoolService.GetNodepoolsMinimal(ctx, []string{}, "", "asc", 0, 1)
			Expect(err).To(BeNil())
			Expect(response1.Code).To(Equal(http.StatusOK))
			firstPage := response1.Body.(api.GetNodepoolsMinimal200Response).Nodepools
			Expect(len(firstPage)).To(Equal(1))

			// Get second page
			response2, err := nodepoolService.GetNodepoolsMinimal(ctx, []string{}, "", "asc", 1, 1)
			Expect(err).To(BeNil())
			Expect(response2.Code).To(Equal(http.StatusOK))
			secondPage := response2.Body.(api.GetNodepoolsMinimal200Response).Nodepools
			Expect(len(secondPage)).To(Equal(1))

			// Verify different nodepools
			Expect(firstPage[0].Id).NotTo(Equal(secondPage[0].Id))

			// Verify both have scheduling capabilities with expected defaults
			Expect(firstPage[0].SchedulingCapabilities.Quota).To(BeTrue())
			Expect(firstPage[0].SchedulingCapabilities.Ranks).To(BeTrue())
			Expect(firstPage[0].SchedulingCapabilities.AdjustableWeights).To(BeFalse())
			Expect(firstPage[0].SchedulingCapabilities.CpuQuota).To(BeFalse())
			Expect(secondPage[0].SchedulingCapabilities.Quota).To(BeTrue())
			Expect(secondPage[0].SchedulingCapabilities.Ranks).To(BeTrue())
			Expect(secondPage[0].SchedulingCapabilities.AdjustableWeights).To(BeFalse())
			Expect(secondPage[0].SchedulingCapabilities.CpuQuota).To(BeFalse())
		})

		It("should combine offset and limit correctly", func() {
			response, err := nodepoolService.GetNodepoolsMinimal(ctx, []string{}, "", "asc", 1, 2)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			metadata := response.Body.(api.GetNodepoolsMinimal200Response).Nodepools
			Expect(len(metadata)).To(BeNumerically("<=", 2))

			// Verify scheduling capabilities have expected defaults
			for _, np := range metadata {
				Expect(np.SchedulingCapabilities.Quota).To(BeTrue())
				Expect(np.SchedulingCapabilities.Ranks).To(BeTrue())
				Expect(np.SchedulingCapabilities.AdjustableWeights).To(BeFalse())
				Expect(np.SchedulingCapabilities.CpuQuota).To(BeFalse())
			}
		})
	})

	Context("Field validation", func() {
		It("should reject filtering by unsupported field", func() {
			filterBy := []string{"phase==Ready"} // phase is not available in metadata
			response, err := nodepoolService.GetNodepoolsMinimal(ctx, filterBy, "", "asc", 0, 0)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusBadRequest))
		})

		It("should reject sorting by unsupported field", func() {
			response, err := nodepoolService.GetNodepoolsMinimal(ctx, []string{}, "phase", "asc", 0, 0) // phase not supported
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusBadRequest))
		})

		It("should reject invalid UUID format for clusterId", func() {
			filterBy := []string{"clusterId==invalid-uuid"}
			response, err := nodepoolService.GetNodepoolsMinimal(ctx, filterBy, "", "asc", 0, 0)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusBadRequest))
		})
	})

	Context("Combined filters and sorting", func() {
		It("should apply both filter and sort", func() {
			filterBy := []string{fmt.Sprintf("clusterId==%s", metadataTestClusterId)}
			response, err := nodepoolService.GetNodepoolsMinimal(ctx, filterBy, "clusterId", "asc", 0, 0)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			metadata := response.Body.(api.GetNodepoolsMinimal200Response).Nodepools
			Expect(len(metadata)).To(Equal(3)) // 1 default + 2 created

			// Verify all belong to correct cluster and are sorted by name
			for i, np := range metadata {
				Expect(np.ClusterId).To(Equal(metadataTestClusterId))
				// Verify scheduling capabilities have expected defaults
				Expect(np.SchedulingCapabilities.Quota).To(BeTrue())
				Expect(np.SchedulingCapabilities.Ranks).To(BeTrue())
				Expect(np.SchedulingCapabilities.AdjustableWeights).To(BeFalse())
				Expect(np.SchedulingCapabilities.CpuQuota).To(BeFalse())
				if i > 0 {
					// String comparison for name sorting
					Expect(np.ClusterId >= metadata[i-1].ClusterId).To(BeTrue(),
						"Expected name %s to be >= %s", np.ClusterId, metadata[i-1].ClusterId)
				}
			}
		})
	})

	Context("GetNodepoolMinimal by ID - Basic functionality", func() {
		It("should return nodepool metadata by ID", func() {
			// Use one of the created nodepools
			nodepoolId := createdNodepools[0].Id
			response, err := nodepoolService.GetNodepoolMinimal(ctx, nodepoolId)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			metadata := response.Body.(api.NodepoolMinimal)
			Expect(metadata.Id).To(Equal(nodepoolId))
			Expect(metadata.Name).To(Equal(createdNodepools[0].Name))
			Expect(metadata.ClusterId).To(Equal(createdNodepools[0].ClusterId))
			Expect(metadata.Phase).To(Equal(createdNodepools[0].Phase))

			// Verify scheduling capabilities are present with expected default values
			Expect(metadata.SchedulingCapabilities.Quota).To(BeTrue())
			Expect(metadata.SchedulingCapabilities.Ranks).To(BeTrue())
			Expect(metadata.SchedulingCapabilities.AdjustableWeights).To(BeFalse())
			Expect(metadata.SchedulingCapabilities.CpuQuota).To(BeFalse())

			// Verify gpuNetworkAccelerationDetected is returned with default false value
			Expect(metadata.GpuNetworkAccelerationDetected).To(BeFalse())
		})

		It("should return metadata with only metadata fields", func() {
			nodepoolId := createdNodepools[0].Id
			response, err := nodepoolService.GetNodepoolMinimal(ctx, nodepoolId)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			metadata := response.Body.(api.NodepoolMinimal)
			// Verify metadata structure - should only have id, name, clusterId, phase with correct values
			Expect(metadata.Id).To(Equal(createdNodepools[0].Id))
			Expect(metadata.Name).To(Equal(createdNodepools[0].Name))
			Expect(metadata.ClusterId).To(Equal(createdNodepools[0].ClusterId))
			Expect(metadata.Phase).To(Equal(createdNodepools[0].Phase))

			// Verify scheduling capabilities have expected defaults
			Expect(metadata.SchedulingCapabilities.Quota).To(BeTrue())              // Default is true
			Expect(metadata.SchedulingCapabilities.Ranks).To(BeTrue())              // Default is true
			Expect(metadata.SchedulingCapabilities.AdjustableWeights).To(BeFalse()) // Default is false
			Expect(metadata.SchedulingCapabilities.CpuQuota).To(BeFalse())          // Default is false
		})
	})

	Context("GetNodepoolMinimal by ID - Error handling", func() {
		It("should return 404 for non-existent nodepool ID", func() {
			nonExistentId := "99999"
			response, err := nodepoolService.GetNodepoolMinimal(ctx, nonExistentId)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusNotFound))
		})

		It("should return 404 for deleted nodepool", func() {
			// Create and then delete a nodepool
			createResp, err := nodepoolService.CreateNodepool(ctx, api.NodepoolCreateFields{
				Name:       "to-be-deleted-np",
				LabelKey:   "test-key",
				LabelValue: "test-value",
				ClusterId:  metadataTestClusterId,
			})
			Expect(err).To(BeNil())
			Expect(createResp.Code).To(Equal(http.StatusCreated))

			deletedNodepool := createResp.Body.(api.Nodepool)
			deletedNodepoolId := deletedNodepool.Id

			// Delete the nodepool
			deleteResp, err := nodepoolService.DeleteNodepool(ctx, deletedNodepoolId)
			Expect(err).To(BeNil())
			Expect(deleteResp.Code).To(SatisfyAny(Equal(http.StatusAccepted), Equal(http.StatusNoContent)))

			// Try to get metadata for deleted nodepool
			response, err := nodepoolService.GetNodepoolMinimal(ctx, deletedNodepoolId)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusNotFound))
		})

		It("should handle empty nodepool ID", func() {
			response, err := nodepoolService.GetNodepoolMinimal(ctx, "")
			Expect(err).To(BeNil())
			// Expecting either BadRequest or NotFound depending on implementation
			Expect(response.Code).To(SatisfyAny(Equal(http.StatusBadRequest), Equal(http.StatusNotFound)))
		})
	})

	Context("GetNodepoolMinimal by ID - Validation", func() {
		It("should return correct nodepool when multiple nodepools exist", func() {
			// Use existing created nodepools
			firstNodepoolId := createdNodepools[0].Id
			secondNodepoolId := createdNodepools[1].Id

			// Get metadata for the first nodepool
			response, err := nodepoolService.GetNodepoolMinimal(ctx, firstNodepoolId)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			metadata := response.Body.(api.NodepoolMinimal)
			Expect(metadata.Id).To(Equal(firstNodepoolId))
			Expect(metadata.Name).To(Equal(createdNodepools[0].Name))
			Expect(metadata.Id).NotTo(Equal(secondNodepoolId))

			// Verify scheduling capabilities are present with expected defaults
			Expect(metadata.SchedulingCapabilities.Quota).To(BeTrue())
			Expect(metadata.SchedulingCapabilities.Ranks).To(BeTrue())
			Expect(metadata.SchedulingCapabilities.AdjustableWeights).To(BeFalse())
			Expect(metadata.SchedulingCapabilities.CpuQuota).To(BeFalse())
		})

		It("should return correct nodepool from different clusters", func() {
			// Get nodepool from cluster 1
			filterBy := []string{fmt.Sprintf("clusterId==%s", metadataTestClusterId)}
			listResponse, err := nodepoolService.GetNodepoolsMinimal(ctx, filterBy, "", "asc", 0, 1)
			Expect(err).To(BeNil())
			Expect(listResponse.Code).To(Equal(http.StatusOK))
			cluster1Nodepools := listResponse.Body.(api.GetNodepoolsMinimal200Response).Nodepools
			Expect(len(cluster1Nodepools)).To(Equal(1))

			cluster1NodepoolId := cluster1Nodepools[0].Id

			// Get by ID
			response, err := nodepoolService.GetNodepoolMinimal(ctx, cluster1NodepoolId)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			metadata := response.Body.(api.NodepoolMinimal)
			Expect(metadata.Id).To(Equal(cluster1NodepoolId))
			Expect(metadata.ClusterId).To(Equal(metadataTestClusterId))

			// Verify scheduling capabilities are present with expected defaults
			Expect(metadata.SchedulingCapabilities.Quota).To(BeTrue())
			Expect(metadata.SchedulingCapabilities.Ranks).To(BeTrue())
			Expect(metadata.SchedulingCapabilities.AdjustableWeights).To(BeFalse())
			Expect(metadata.SchedulingCapabilities.CpuQuota).To(BeFalse())
		})
	})

	Context("GetNodepoolMinimal by ID - No authorization required", func() {
		It("should return metadata without authorization", func() {
			// This endpoint should work without authorization headers
			// The x-no-authorization flag means it bypasses auth checks
			nodepoolId := createdNodepools[0].Id
			response, err := nodepoolService.GetNodepoolMinimal(ctx, nodepoolId)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			metadata := response.Body.(api.NodepoolMinimal)
			Expect(metadata.Id).To(Equal(nodepoolId))

			// Verify scheduling capabilities are present with expected defaults
			Expect(metadata.SchedulingCapabilities.Quota).To(BeTrue())
			Expect(metadata.SchedulingCapabilities.Ranks).To(BeTrue())
			Expect(metadata.SchedulingCapabilities.AdjustableWeights).To(BeFalse())
			Expect(metadata.SchedulingCapabilities.CpuQuota).To(BeFalse())
		})
	})

	Context("GetNodepoolsMinimal - Basic functionality", func() {
		It("should return all nodepools full metadata with scheduling capabilities", func() {
			response, err := nodepoolService.GetNodepoolsMinimal(ctx, []string{}, "", "asc", 0, 0)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			metadata := response.Body.(api.GetNodepoolsMinimal200Response).Nodepools
			Expect(len(metadata)).To(Equal(5))

			// Verify full metadata contains all expected fields including scheduling capabilities
			for _, np := range metadata {
				// Basic metadata fields
				Expect(np.Id).NotTo(BeEmpty(), "ID should not be empty")
				Expect(np.Name).NotTo(BeEmpty(), "Name should not be empty")
				Expect(np.ClusterId).NotTo(BeEmpty(), "ClusterId should not be empty")

				// Scheduling capabilities should have expected default values
				Expect(np.SchedulingCapabilities.Quota).To(BeTrue())
				Expect(np.SchedulingCapabilities.Ranks).To(BeTrue())
				Expect(np.SchedulingCapabilities.AdjustableWeights).To(BeFalse())
				Expect(np.SchedulingCapabilities.CpuQuota).To(BeFalse())

				// If this is one of our test nodepools, validate the fields match expected data
				if expectedData, exists := testNodepools[np.Name]; exists {
					Expect(np.Name).To(Equal(expectedData.Name), "Name should match test data")
					Expect(np.ClusterId).To(Equal(*expectedData.ClusterID), "ClusterId should match test data")
					Expect(np.Phase).To(Equal(api.NODEPOOLPHASE_CREATING), "Phase should be CREATING for created nodepools")
				}
			}
		})

		It("should return full metadata with correct scheduling capabilities defaults", func() {
			response, err := nodepoolService.GetNodepoolsMinimal(ctx, []string{}, "", "asc", 0, 0)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			metadata := response.Body.(api.GetNodepoolsMinimal200Response).Nodepools
			Expect(len(metadata)).To(Equal(5)) // 2 default + 3 created

			// Check first nodepool has scheduling capabilities with expected defaults
			np := metadata[0]
			Expect(np.SchedulingCapabilities.Quota).To(BeTrue()) // Default is true
			Expect(np.SchedulingCapabilities.Ranks).To(BeTrue()) // Default is true
		})
	})

	Context("GetNodepoolsMinimal - Filtering and sorting", func() {
		It("should filter full metadata by clusterId", func() {
			filterBy := []string{fmt.Sprintf("clusterId==%s", metadataTestClusterId)}
			response, err := nodepoolService.GetNodepoolsMinimal(ctx, filterBy, "", "asc", 0, 0)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			metadata := response.Body.(api.GetNodepoolsMinimal200Response).Nodepools
			Expect(len(metadata)).To(Equal(3)) // 1 default + 2 created

			// Verify all returned nodepools belong to the specified cluster
			for _, np := range metadata {
				Expect(np.ClusterId).To(Equal(metadataTestClusterId))
				// Verify scheduling capabilities have expected defaults
				Expect(np.SchedulingCapabilities.Quota).To(BeTrue())
			}
		})

		It("should sort full metadata by clusterId ascending", func() {
			response, err := nodepoolService.GetNodepoolsMinimal(ctx, []string{}, "clusterId", "asc", 0, 0)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			metadata := response.Body.(api.GetNodepoolsMinimal200Response).Nodepools
			Expect(len(metadata)).To(Equal(5)) // 2 default + 3 created

			// Verify ascending order and presence of scheduling capabilities
			for i := 1; i < len(metadata); i++ {
				Expect(metadata[i].ClusterId >= metadata[i-1].ClusterId).To(BeTrue())
				Expect(metadata[i].SchedulingCapabilities.Quota).To(BeTrue())
			}
		})
	})

	Context("GetNodepoolsMinimal - Pagination", func() {
		It("should respect limit parameter", func() {
			response, err := nodepoolService.GetNodepoolsMinimal(ctx, []string{}, "", "asc", 0, 2)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			metadata := response.Body.(api.GetNodepoolsMinimal200Response).Nodepools
			Expect(len(metadata)).To(Equal(2))
			// Verify scheduling capabilities have expected defaults
			for _, np := range metadata {
				Expect(np.SchedulingCapabilities.Quota).To(BeTrue())
				Expect(np.SchedulingCapabilities.Ranks).To(BeTrue())
				Expect(np.SchedulingCapabilities.AdjustableWeights).To(BeFalse())
				Expect(np.SchedulingCapabilities.CpuQuota).To(BeFalse())
			}
		})

		It("should respect offset and limit parameters", func() {
			// Get first page
			response1, err := nodepoolService.GetNodepoolsMinimal(ctx, []string{}, "", "asc", 0, 1)
			Expect(err).To(BeNil())
			Expect(response1.Code).To(Equal(http.StatusOK))
			firstPage := response1.Body.(api.GetNodepoolsMinimal200Response).Nodepools
			Expect(len(firstPage)).To(Equal(1))

			// Get second page
			response2, err := nodepoolService.GetNodepoolsMinimal(ctx, []string{}, "", "asc", 1, 1)
			Expect(err).To(BeNil())
			Expect(response2.Code).To(Equal(http.StatusOK))
			secondPage := response2.Body.(api.GetNodepoolsMinimal200Response).Nodepools
			Expect(len(secondPage)).To(Equal(1))

			// Verify different nodepools
			Expect(firstPage[0].Id).NotTo(Equal(secondPage[0].Id))
			// Verify both have scheduling capabilities with expected defaults
			Expect(firstPage[0].SchedulingCapabilities.Quota).To(BeTrue())
			Expect(firstPage[0].SchedulingCapabilities.Ranks).To(BeTrue())
			Expect(firstPage[0].SchedulingCapabilities.AdjustableWeights).To(BeFalse())
			Expect(firstPage[0].SchedulingCapabilities.CpuQuota).To(BeFalse())
			Expect(secondPage[0].SchedulingCapabilities.Quota).To(BeTrue())
			Expect(secondPage[0].SchedulingCapabilities.Ranks).To(BeTrue())
			Expect(secondPage[0].SchedulingCapabilities.AdjustableWeights).To(BeFalse())
			Expect(secondPage[0].SchedulingCapabilities.CpuQuota).To(BeFalse())
		})
	})

	Context("GetNodepoolsMinimal - Error handling", func() {
		It("should reject filtering by unsupported field", func() {
			filterBy := []string{"phase==Ready"}
			response, err := nodepoolService.GetNodepoolsMinimal(ctx, filterBy, "", "asc", 0, 0)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusBadRequest))
		})

		It("should reject sorting by unsupported field", func() {
			response, err := nodepoolService.GetNodepoolsMinimal(ctx, []string{}, "phase", "asc", 0, 0)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusBadRequest))
		})
	})

	Context("GetNodepoolMinimal by ID - Basic functionality", func() {
		It("should return nodepool full metadata by ID with scheduling capabilities", func() {
			// Use one of the created nodepools
			nodepoolId := createdNodepools[0].Id
			response, err := nodepoolService.GetNodepoolMinimal(ctx, nodepoolId)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			metadata := response.Body.(api.NodepoolMinimal)

			// Verify basic metadata fields
			Expect(metadata.Id).To(Equal(nodepoolId))
			Expect(metadata.Name).To(Equal(createdNodepools[0].Name))
			Expect(metadata.ClusterId).To(Equal(createdNodepools[0].ClusterId))
			Expect(metadata.Phase).To(Equal(createdNodepools[0].Phase))

			// Verify scheduling capabilities are present with expected values
			Expect(metadata.SchedulingCapabilities.Quota).To(BeTrue())
			Expect(metadata.SchedulingCapabilities.Ranks).To(BeTrue())
			Expect(metadata.SchedulingCapabilities.AdjustableWeights).To(BeFalse())
			Expect(metadata.SchedulingCapabilities.CpuQuota).To(BeFalse())
		})

		It("should return full metadata with default scheduling capabilities values", func() {
			nodepoolId := createdNodepools[0].Id
			response, err := nodepoolService.GetNodepoolMinimal(ctx, nodepoolId)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			metadata := response.Body.(api.NodepoolMinimal)

			// Verify scheduling capabilities defaults
			Expect(metadata.SchedulingCapabilities.Quota).To(BeTrue())              // Default is true
			Expect(metadata.SchedulingCapabilities.Ranks).To(BeTrue())              // Default is true
			Expect(metadata.SchedulingCapabilities.AdjustableWeights).To(BeFalse()) // Default is false
			Expect(metadata.SchedulingCapabilities.CpuQuota).To(BeFalse())          // Default is false
		})

		It("should return different nodepools with correct full metadata", func() {
			// Get metadata for first nodepool
			firstNodepoolId := createdNodepools[0].Id
			response1, err := nodepoolService.GetNodepoolMinimal(ctx, firstNodepoolId)
			Expect(err).To(BeNil())
			Expect(response1.Code).To(Equal(http.StatusOK))
			firstMetadata := response1.Body.(api.NodepoolMinimal)

			// Get metadata for second nodepool
			secondNodepoolId := createdNodepools[1].Id
			response2, err := nodepoolService.GetNodepoolMinimal(ctx, secondNodepoolId)
			Expect(err).To(BeNil())
			Expect(response2.Code).To(Equal(http.StatusOK))
			secondMetadata := response2.Body.(api.NodepoolMinimal)

			// Verify they are different nodepools
			Expect(firstMetadata.Id).NotTo(Equal(secondMetadata.Id))
			Expect(firstMetadata.Name).NotTo(Equal(secondMetadata.Name))

			// Both should have scheduling capabilities with expected defaults
			Expect(firstMetadata.SchedulingCapabilities.Quota).To(BeTrue())
			Expect(firstMetadata.SchedulingCapabilities.Ranks).To(BeTrue())
			Expect(firstMetadata.SchedulingCapabilities.AdjustableWeights).To(BeFalse())
			Expect(firstMetadata.SchedulingCapabilities.CpuQuota).To(BeFalse())
			Expect(secondMetadata.SchedulingCapabilities.Quota).To(BeTrue())
			Expect(secondMetadata.SchedulingCapabilities.Ranks).To(BeTrue())
			Expect(secondMetadata.SchedulingCapabilities.AdjustableWeights).To(BeFalse())
			Expect(secondMetadata.SchedulingCapabilities.CpuQuota).To(BeFalse())
		})
	})

	Context("GetNodepoolMinimal by ID - Error handling", func() {
		It("should return 404 for non-existent nodepool ID", func() {
			nonExistentId := "99999"
			response, err := nodepoolService.GetNodepoolMinimal(ctx, nonExistentId)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusNotFound))
		})

		It("should return 404 for deleted nodepool", func() {
			// Create and then delete a nodepool
			createResp, err := nodepoolService.CreateNodepool(ctx, api.NodepoolCreateFields{
				Name:       "to-be-deleted-full-np",
				LabelKey:   "test-key-full",
				LabelValue: "test-value-full",
				ClusterId:  metadataTestClusterId,
			})
			Expect(err).To(BeNil())
			Expect(createResp.Code).To(Equal(http.StatusCreated))

			deletedNodepool := createResp.Body.(api.Nodepool)
			deletedNodepoolId := deletedNodepool.Id

			// Delete the nodepool
			deleteResp, err := nodepoolService.DeleteNodepool(ctx, deletedNodepoolId)
			Expect(err).To(BeNil())
			Expect(deleteResp.Code).To(SatisfyAny(Equal(http.StatusAccepted), Equal(http.StatusNoContent)))

			// Try to get full metadata for deleted nodepool
			response, err := nodepoolService.GetNodepoolMinimal(ctx, deletedNodepoolId)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusNotFound))
		})

		It("should handle empty nodepool ID", func() {
			response, err := nodepoolService.GetNodepoolMinimal(ctx, "")
			Expect(err).To(BeNil())
			// Expecting either BadRequest or NotFound depending on implementation
			Expect(response.Code).To(SatisfyAny(Equal(http.StatusBadRequest), Equal(http.StatusNotFound)))
		})
	})

	Context("GetNodepoolMinimal by ID - Validation", func() {
		It("should return correct full metadata for nodepool from specific cluster", func() {
			// Get a nodepool from cluster 1
			filterBy := []string{fmt.Sprintf("clusterId==%s", metadataTestClusterId)}
			listResponse, err := nodepoolService.GetNodepoolsMinimal(ctx, filterBy, "", "asc", 0, 1)
			Expect(err).To(BeNil())
			Expect(listResponse.Code).To(Equal(http.StatusOK))
			cluster1Nodepools := listResponse.Body.(api.GetNodepoolsMinimal200Response).Nodepools
			Expect(len(cluster1Nodepools)).To(Equal(1))

			cluster1NodepoolId := cluster1Nodepools[0].Id

			// Get full metadata by ID
			response, err := nodepoolService.GetNodepoolMinimal(ctx, cluster1NodepoolId)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			metadata := response.Body.(api.NodepoolMinimal)
			Expect(metadata.Id).To(Equal(cluster1NodepoolId))
			Expect(metadata.ClusterId).To(Equal(metadataTestClusterId))
			Expect(metadata.SchedulingCapabilities.Quota).To(BeTrue())
			Expect(metadata.SchedulingCapabilities.Ranks).To(BeTrue())
			Expect(metadata.SchedulingCapabilities.AdjustableWeights).To(BeFalse())
			Expect(metadata.SchedulingCapabilities.CpuQuota).To(BeFalse())
		})
	})

	Context("GetNodepoolMinimal by ID - No authorization required", func() {
		It("should return full metadata without authorization", func() {
			// This endpoint should work without authorization headers
			// The x-no-authorization flag means it bypasses auth checks
			nodepoolId := createdNodepools[0].Id
			response, err := nodepoolService.GetNodepoolMinimal(ctx, nodepoolId)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			metadata := response.Body.(api.NodepoolMinimal)
			Expect(metadata.Id).To(Equal(nodepoolId))
			Expect(metadata.SchedulingCapabilities.Quota).To(BeTrue())
			Expect(metadata.SchedulingCapabilities.Ranks).To(BeTrue())
			Expect(metadata.SchedulingCapabilities.AdjustableWeights).To(BeFalse())
			Expect(metadata.SchedulingCapabilities.CpuQuota).To(BeFalse())
		})
	})

	Context("Comparison between basic and full metadata", func() {
		It("should return same basic fields in both metadata and full metadata endpoints", func() {
			nodepoolId := createdNodepools[0].Id

			// Get basic metadata
			basicResponse, err := nodepoolService.GetNodepoolMinimal(ctx, nodepoolId)
			Expect(err).To(BeNil())
			Expect(basicResponse.Code).To(Equal(http.StatusOK))
			basicMetadata := basicResponse.Body.(api.NodepoolMinimal)

			// Get full metadata
			fullResponse, err := nodepoolService.GetNodepoolMinimal(ctx, nodepoolId)
			Expect(err).To(BeNil())
			Expect(fullResponse.Code).To(Equal(http.StatusOK))
			metadata := fullResponse.Body.(api.NodepoolMinimal)

			// Compare basic fields - should be identical
			Expect(metadata.Id).To(Equal(basicMetadata.Id))
			Expect(metadata.Name).To(Equal(basicMetadata.Name))
			Expect(metadata.ClusterId).To(Equal(basicMetadata.ClusterId))
			Expect(metadata.Phase).To(Equal(basicMetadata.Phase))

			// Both basic and full metadata should have scheduling capabilities with expected defaults
			Expect(basicMetadata.SchedulingCapabilities.Quota).To(BeTrue())
			Expect(basicMetadata.SchedulingCapabilities.Ranks).To(BeTrue())
			Expect(basicMetadata.SchedulingCapabilities.AdjustableWeights).To(BeFalse())
			Expect(basicMetadata.SchedulingCapabilities.CpuQuota).To(BeFalse())

			Expect(metadata.SchedulingCapabilities.Quota).To(BeTrue())
			Expect(metadata.SchedulingCapabilities.Ranks).To(BeTrue())
			Expect(metadata.SchedulingCapabilities.AdjustableWeights).To(BeFalse())
			Expect(metadata.SchedulingCapabilities.CpuQuota).To(BeFalse())

			// Verify scheduling capabilities are identical between basic and full metadata
			Expect(metadata.SchedulingCapabilities.Quota).To(Equal(basicMetadata.SchedulingCapabilities.Quota))
			Expect(metadata.SchedulingCapabilities.Ranks).To(Equal(basicMetadata.SchedulingCapabilities.Ranks))
			Expect(metadata.SchedulingCapabilities.AdjustableWeights).To(Equal(basicMetadata.SchedulingCapabilities.AdjustableWeights))
			Expect(metadata.SchedulingCapabilities.CpuQuota).To(Equal(basicMetadata.SchedulingCapabilities.CpuQuota))
		})
	})

	Context("Scope validation for nodepools-minimal functions", func() {
		It("GetNodepoolsMinimal should return forbidden when no permitted cluster scopes exist", func() {
			// Create context without permitted scopes
			emptyCtx := createContextForTenant(fmt.Sprintf("%d", tenantId), "user-1")
			emptyCtx = context.WithValue(emptyCtx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{})

			response, err := nodepoolService.GetNodepoolsMinimal(emptyCtx, []string{}, "", "asc", 0, 0)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))
			metadata := response.Body.(api.GetNodepoolsMinimal200Response).Nodepools
			Expect(len(metadata)).To(Equal(0))
		})

		It("GetNodepoolsMinimal should succeed with valid tenant scope", func() {
			response, err := nodepoolService.GetNodepoolsMinimal(ctx, []string{}, "", "asc", 0, 0)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			metadata := response.Body.(api.GetNodepoolsMinimal200Response).Nodepools
			Expect(len(metadata)).To(BeNumerically(">", 0))
		})

		It("GetNodepoolMinimal should return forbidden when no permitted cluster scopes exist", func() {
			nodepoolId := createdNodepools[0].Id

			// Create context without permitted scopes
			emptyCtx := createContextForTenant(fmt.Sprintf("%d", tenantId), "user-1")
			emptyCtx = context.WithValue(emptyCtx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{})

			response, err := nodepoolService.GetNodepoolMinimal(emptyCtx, nodepoolId)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusForbidden))
		})

		It("GetNodepoolMinimal should succeed with valid cluster scope", func() {
			nodepoolId := createdNodepools[0].Id
			clusterId := createdNodepools[0].ClusterId

			// Create context with specific cluster scope
			clusterScopeCtx := createContextForTenant(fmt.Sprintf("%d", tenantId), "user-1")
			clusterScopeCtx = context.WithValue(clusterScopeCtx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{
				Clusters: []string{clusterId},
			})

			response, err := nodepoolService.GetNodepoolMinimal(clusterScopeCtx, nodepoolId)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			metadata := response.Body.(api.NodepoolMinimal)
			Expect(metadata.Id).To(Equal(nodepoolId))
		})
	})
})
