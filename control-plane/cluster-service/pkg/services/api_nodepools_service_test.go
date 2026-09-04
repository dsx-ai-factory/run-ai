/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services_test

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/xhit/go-str2duration/v2"

	authorization_utils "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/helpers"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/services"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	authorization_api "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/internal/generated/pkg/api"
	bc_mocks "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/mocks/pkg/clients/backend"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/database/models"
	org_unit_service_client_mock "github.com/dsx-ai-factory/run-ai/runai-clients/org-unit-service-client/mocks"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/stretchr/testify/mock"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/utils/ptr"
)

const (
	tenantIdStr    = "1"
	testName       = "np-a"
	testLabelKey   = "a"
	testLabelValue = "b"
	testUser       = "user"

	missingPrerequisitesPhaseMessage = "The following prerequisites are not met on one or more nodes in the node pool: " +
		"{NodeResourceTopology custom resource missing}. As a result, NUMA-aware scheduling may be impacted."
	missingNrtHealthyPrerequisiteCondition = "MissingNrtHealthyPrerequisite"
	missingNrtHealthyPrerequisiteReason    = "MissingNrtHealthyPrerequisiteReason"
)

var testClusterId, testClusterName string

func deleteNodepools(nodepools []api.Nodepool) {
	for _, nodepool := range nodepools {
		err := dao.DeleteNodepool(ctx, tenantId, "", nodepool.Id)
		Expect(err).To(BeNil())
	}
}

var _ = Describe("Nodepools tests", Ordered, func() {
	tenantSettingsClient := &bc_mocks.TenantSettingsClient{}
	orgClient := &org_unit_service_client_mock.IOrgUnitClient{}

	ctx := createContext()

	tenantSettingsClient.On("GetOPRSetting", mock.Anything).Return(true, nil)
	orgClient.On("NodePoolCreated", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	orgClient.On("NodePoolDeleted", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	orgClient.On("DefaultNodePoolCreated", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	nodepoolService := services.NewNodePoolsApiService(tenantSettingsClient, orgClient)
	createdNodepools := []api.Nodepool{}

	AfterAll(func() {
		deleteNodepools(createdNodepools)
		createdNodepools = []api.Nodepool{}

		dao.DeleteCluster(ctx, testClusterId, tenantId)

		res, err := nodepoolService.ClusterNotification(ctx, api.ClusterNotification{
			Operation: api.NOTIFICATIONOPERATION_DELETE,
			Id:        testClusterId,
			Name:      testClusterName,
		})
		Expect(err).To(BeNil(), "Expected cluster deletion notification to succeed")
		Expect(res.Code).To(Equal(http.StatusNoContent), "Expected cluster deletion notification to succeed")

		Eventually(func(g Gomega) {
			defaultNodePool := getDefaultNodePoolOfCluster(ctx, nodepoolService, testClusterId)
			Expect(defaultNodePool).To(BeNil(), "Expected default node pool to be deleted")
		}, 3*time.Second, 100*time.Millisecond).Should(Succeed())
	})
	BeforeAll(func() {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Running nodepool tests")

		cluster, err := dao.CreateCluster(ctx, &models.Cluster{
			Name:     testName,
			TenantId: tenantId,
			Domain:   "xyz",
			Version:  "v2.24",
		})
		Expect(err).To(BeNil())
		testClusterId = cluster.ID.String()
		testClusterName = cluster.Name

		res, err2 := nodepoolService.ClusterNotification(ctx, api.ClusterNotification{
			Operation: api.NOTIFICATIONOPERATION_CREATE,
			Id:        cluster.ID.String(),
			Name:      cluster.Name,
		})
		Expect(err2).To(BeNil(), "Expected cluster creation notification to succeed")
		Expect(res.Code).To(Equal(http.StatusNoContent), "Expected cluster creation notification to succeed")

		// validate default nodepool created and update its status to "Ready"
		var defaultNodePool *api.Nodepool
		Eventually(func(g Gomega) {
			defaultNodePool = getDefaultNodePoolOfCluster(ctx, nodepoolService, testClusterId)
			Expect(defaultNodePool).To(Not(BeNil()), "Expected default node pool to be created")
		}, 3*time.Second, 100*time.Millisecond).Should(Succeed())

		response, err2 := nodepoolService.SyncNodepoolsStatus(ctx, testClusterId, api.SyncNodepoolsStatusRequest{
			Nodepools: []api.NodepoolSyncFields{
				{
					Name:         defaultNodePool.Name,
					Phase:        api.NODEPOOLPHASE_READY,
					PhaseMessage: "Nodepool is ready",
				},
			},
			DeleteMissing: false,
		})
		Expect(err).To(BeNil())
		Expect(response.Code).To(Equal(http.StatusNoContent))

		defaultNodePool = getDefaultNodePoolOfCluster(ctx, nodepoolService, testClusterId)
		Expect(defaultNodePool).To(Not(BeNil()))
		Expect(defaultNodePool.Phase).To(Equal(api.NODEPOOLPHASE_READY))
	})

	Context("GPU Network Acceleration Detection - not Supported", Ordered, func() {
		testSuffix := "-with-gpu-acc-detection-not-supported"
		detectionClusterId := ""
		BeforeEach(func() {
			cluster, err := dao.CreateCluster(ctx, &models.Cluster{
				Name:     fmt.Sprintf("%s%s", testName, testSuffix),
				TenantId: tenantId,
				Domain:   "xyz",
				Version:  "v2.20",
			})
			Expect(err).To(BeNil())
			detectionClusterId = cluster.ID.String()
		})
		AfterEach(func() {
			deleteNodepools(createdNodepools)
			createdNodepools = []api.Nodepool{}
			dao.DeleteCluster(ctx, detectionClusterId, tenantId)
		})
		It("Create nodepool with gpu acc detection", func() {
			request := nodepoolCreateFields(detectionClusterId, testSuffix, ptr.To(api.NODEPOOLGPUNETWORKACCELERATIONDETECTION_USE))

			response, err := nodepoolService.CreateNodepool(ctx, *request)

			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusBadRequest))
		})
		It("Update nodepool with gpu acc detection", func() {
			createRequest := nodepoolCreateFields(detectionClusterId, testSuffix, nil)

			response, err := nodepoolService.CreateNodepool(ctx, *createRequest)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusCreated))

			nodepool, ok := response.Body.(api.Nodepool)
			Expect(ok).To(BeTrue())
			createdNodepools = append(createdNodepools, nodepool)
			updateRequest := nodepoolUpdateFields(testSuffix, ptr.To(api.NODEPOOLGPUNETWORKACCELERATIONDETECTION_USE))
			updateResponse, err := nodepoolService.UpdateNodepool(ctx, nodepool.Id, *updateRequest)
			Expect(err).To(BeNil())
			Expect(updateResponse.Code).To(Equal(http.StatusBadRequest))
		})
		It("Patch nodepool with gpu acc detection", func() {
			createRequest := nodepoolCreateFields(detectionClusterId, testSuffix, nil)

			response, err := nodepoolService.CreateNodepool(ctx, *createRequest)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusCreated))

			nodepool, ok := response.Body.(api.Nodepool)
			Expect(ok).To(BeTrue())
			createdNodepools = append(createdNodepools, nodepool)
			updateRequest := nodepoolUpdateFields(testSuffix, ptr.To(api.NODEPOOLGPUNETWORKACCELERATIONDETECTION_USE))

			updateResponse, err := nodepoolService.PatchNodepool(ctx, nodepool.Id, *updateRequest)
			Expect(err).To(BeNil())
			Expect(updateResponse.Code).To(Equal(http.StatusBadRequest))
		})
	})

	Context("GPU acc detection", Ordered, func() {
		testSuffix := "-with-gpu-acc-detection"
		clusterId := ""
		oldClusterId := ""
		BeforeEach(func() {
			cluster, err := dao.CreateCluster(ctx, &models.Cluster{
				Name:     fmt.Sprintf("%s%s", testName, testSuffix),
				TenantId: tenantId,
				Domain:   "xyz",
				Version:  "v2.21",
			})
			Expect(err).To(BeNil())
			clusterId = cluster.ID.String()

			cluster, err = dao.CreateCluster(ctx, &models.Cluster{
				Name:     fmt.Sprintf("%s%s-2", testName, testSuffix),
				TenantId: tenantId,
				Domain:   "xyz",
				Version:  "v2.19",
			})
			Expect(err).To(BeNil())
			oldClusterId = cluster.ID.String()
		})
		AfterEach(func() {
			deleteNodepools(createdNodepools)
			createdNodepools = []api.Nodepool{}
			dao.DeleteCluster(ctx, clusterId, tenantId)
			dao.DeleteCluster(ctx, oldClusterId, tenantId)
		})
		It("Create nodepool with gpu acc detection", func() {
			gpuDetectionLabelKey := "nvidia/gpu-acc"
			request := nodepoolCreateFields(clusterId, testSuffix, ptr.To(api.NODEPOOLGPUNETWORKACCELERATIONDETECTION_USE))
			request.GpuNetworkAccelerationLabelKey = ptr.To(gpuDetectionLabelKey)
			response, err := nodepoolService.CreateNodepool(ctx, *request)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusCreated))

			nodepool, ok := response.Body.(api.Nodepool)
			Expect(ok).To(BeTrue())
			Expect(nodepool.GpuNetworkAccelerationLabelKey).To(Equal(gpuDetectionLabelKey))
			Expect(nodepool.Id).To(Not(BeNil()))

			createdNodepools = append(createdNodepools, nodepool)
			response, err = nodepoolService.GetNodepool(ctx, nodepool.Id)
			Expect(err).To(BeNil())
			nodepool, ok = response.Body.(api.Nodepool)
			Expect(*nodepool.GpuNetworkAccelerationDetection).To(Equal(api.NODEPOOLGPUNETWORKACCELERATIONDETECTION_USE))
			Expect(nodepool.GpuNetworkAccelerationLabelKey).To(Equal(gpuDetectionLabelKey))
		})
		It("Patch nodepool with gpu acc detection and GpuNetworkAccelerationLabelKey", func() {
			gpuDetectionLabelKey := "nvidia/gpu-acc"
			request := nodepoolCreateFields(clusterId, testSuffix, ptr.To(api.NODEPOOLGPUNETWORKACCELERATIONDETECTION_USE))
			request.GpuNetworkAccelerationLabelKey = ptr.To(gpuDetectionLabelKey)
			response, err := nodepoolService.CreateNodepool(ctx, *request)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusCreated))

			nodepool, ok := response.Body.(api.Nodepool)
			Expect(ok).To(BeTrue())
			Expect(nodepool.GpuNetworkAccelerationLabelKey).To(Equal(gpuDetectionLabelKey))
			Expect(*nodepool.GpuNetworkAccelerationDetection).To(Equal(api.NODEPOOLGPUNETWORKACCELERATIONDETECTION_USE))

			createdNodepools = append(createdNodepools, nodepool)
			response, err = nodepoolService.GetNodepool(ctx, nodepool.Id)
			Expect(err).To(BeNil())
			nodepool, ok = response.Body.(api.Nodepool)
			Expect(nodepool.GpuNetworkAccelerationLabelKey).To(Equal(gpuDetectionLabelKey))

			patchFields := nodepoolUpdateFields("-swap-patch-disabled", nil)

			response, err = nodepoolService.PatchNodepool(ctx, nodepool.Id, *patchFields)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))
			nodepool, ok = response.Body.(api.Nodepool)
			Expect(ok).To(BeTrue())
			Expect(nodepool.GpuNetworkAccelerationLabelKey).To(Equal(gpuDetectionLabelKey))

			response, err = nodepoolService.GetNodepool(ctx, nodepool.Id)
			Expect(err).To(BeNil())
			nodepool, ok = response.Body.(api.Nodepool)
			Expect(nodepool.GpuNetworkAccelerationLabelKey).To(Equal(gpuDetectionLabelKey))
		})
		It("Create nodepool with gpu acc detection - old cluster version", func() {
			testSuffix := "-with-gpu-acc-detection"
			request := nodepoolCreateFields(oldClusterId, testSuffix, ptr.To(api.NODEPOOLGPUNETWORKACCELERATIONDETECTION_USE))

			response, err := nodepoolService.CreateNodepool(ctx, *request)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusBadRequest))

			_, ok := response.Body.(api.Nodepool)
			Expect(ok).To(BeFalse())
		})
	})
	Context("Happy flow", Ordered, func() {
		It("Create nodepool", func() {
			request := api.NodepoolCreateFields{
				Name:       testName,
				LabelKey:   testLabelKey,
				LabelValue: testLabelValue,
				ClusterId:  testClusterId,
				SchedulingConfiguration: &api.SchedulingConfiguration{
					PlacementStrategy: api.SchedulingConfigurationPlacementStrategy{
						Cpu: ptr.To(api.PLACEMENTSTRATEGY_BINPACK),
						Gpu: ptr.To(api.PLACEMENTSTRATEGY_BINPACK),
					},
				},
			}

			response, err := nodepoolService.CreateNodepool(ctx, request)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusCreated))

			nodepool, ok := response.Body.(api.Nodepool)
			Expect(ok).To(BeTrue())

			Expect(nodepool.Id).To(Not(BeNil()))
			createdNodepools = append(createdNodepools, nodepool)

			Expect(nodepool.TenantId).To(Equal(int32(tenantId)))
			Expect(nodepool.Phase).To(Equal(api.NODEPOOLPHASE_CREATING))
			Expect(nodepool.PhaseMessage).To(Equal(""))
			Expect(nodepool.CreatedAt).To(Not(BeNil()))
			Expect(nodepool.CreatedAt.Unix()).To(BeNumerically(">", time.Now().Unix()-5000)) // time of postgres query should not be near 5 sec
			Expect(nodepool.CreatedBy).To(Equal(testUser))
			Expect(nodepool.UpdatedAt).To(Not(BeNil()))
			Expect(nodepool.UpdatedAt.Unix()).To(BeNumerically(">", time.Now().Unix()-5000)) // time of postgres query should not be near 5 sec
			Expect(nodepool.UpdatedBy).To(Equal(testUser))
			Expect(nodepool.Name).To(Equal(request.Name))
			Expect(nodepool.LabelKey).To(Equal(request.LabelKey))
			Expect(nodepool.LabelValue).To(Equal(request.LabelValue))
			Expect(nodepool.ClusterId).To(Equal(request.ClusterId))
			Expect(nodepool.SchedulingConfiguration.PlacementStrategy.Cpu).To(Equal(request.SchedulingConfiguration.PlacementStrategy.Cpu))
			Expect(nodepool.SchedulingConfiguration.PlacementStrategy.Gpu).To(Equal(request.SchedulingConfiguration.PlacementStrategy.Gpu))

			// try to re-create with same name - should be error
			response, err = nodepoolService.CreateNodepool(ctx, request)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusConflict))
		})
		It("Sync nodepool", func() {
			Expect(len(createdNodepools)).To(BeNumerically(">", 0))

			exists := createdNodepools[0]

			response, err := nodepoolService.SyncNodepoolsStatus(ctx, exists.ClusterId, api.SyncNodepoolsStatusRequest{
				Nodepools: []api.NodepoolSyncFields{
					{
						Name:         exists.Name,
						Phase:        api.NODEPOOLPHASE_READY,
						PhaseMessage: "Nodepool is ready",
						Status: api.NodepoolSyncFieldsStatus{
							LabelKey:   ptr.To(exists.LabelKey),
							LabelValue: ptr.To(exists.LabelValue),
							SchedulingConfiguration: &api.SchedulingConfiguration{
								PlacementStrategy: api.SchedulingConfigurationPlacementStrategy{
									Cpu: exists.SchedulingConfiguration.PlacementStrategy.Cpu,
									Gpu: exists.SchedulingConfiguration.PlacementStrategy.Gpu,
								},
							},
							Nodes: []string{
								"worker-a",
								"worker-b",
							},
							GpuNetworkAccelerationDetected: true,
						},
					},
				},
				DeleteMissing: false,
			})
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusNoContent))
		})
		It("Get nodepool", func() {
			Expect(len(createdNodepools)).To(BeNumerically(">", 0))

			exists := createdNodepools[0]

			response, err := nodepoolService.GetNodepool(ctx, exists.Id)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			nodepool, ok := response.Body.(api.Nodepool)
			Expect(ok).To(BeTrue())

			Expect(nodepool.Id).To(Equal(exists.Id))

			Expect(nodepool.TenantId).To(Equal(exists.TenantId))
			Expect(nodepool.Phase).To(Equal(api.NODEPOOLPHASE_READY))
			Expect(nodepool.PhaseMessage).To(Equal("Nodepool is ready"))
			Expect(nodepool.CreatedAt.Unix()).To(Equal(exists.CreatedAt.Unix()))
			Expect(nodepool.CreatedBy).To(Equal(exists.CreatedBy))
			Expect(nodepool.UpdatedAt.Unix()).To(BeNumerically(">=", exists.UpdatedAt.Unix()))
			Expect(nodepool.UpdatedBy).To(Equal(exists.UpdatedBy))
			Expect(nodepool.Name).To(Equal(exists.Name))
			Expect(nodepool.LabelKey).To(Equal(exists.LabelKey))
			Expect(nodepool.LabelValue).To(Equal(exists.LabelValue))
			Expect(nodepool.ClusterId).To(Equal(exists.ClusterId))
			Expect(nodepool.SchedulingConfiguration.PlacementStrategy).To(Equal(exists.SchedulingConfiguration.PlacementStrategy))
			Expect(nodepool.Status).To(Not(BeNil()))
			Expect(nodepool.Status.GpuNetworkAccelerationDetected).To(Equal(true))

			Expect(len(nodepool.Status.Nodes)).To(Equal(2))
			Expect(nodepool.Status.Nodes).To(ContainElements("worker-a", "worker-b"))

			createdNodepools[0] = nodepool
		})

		// The deprecated top-level placementStrategy field is
		// re-emitted only by the plural GetNodepools when the request filters by exactly
		// `clusterId==<uuid>` (the shape used by older runai-cluster agents). The singular
		// GetNodepool, and multi-filter GetNodepools queries, should NOT populate them.
		It("Get nodepool (singular) does not populate deprecated fields", func() {
			Expect(len(createdNodepools)).To(BeNumerically(">", 0))
			exists := createdNodepools[0]

			response, err := nodepoolService.GetNodepool(ctx, exists.Id)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			nodepool, ok := response.Body.(api.Nodepool)
			Expect(ok).To(BeTrue())

			// NodepoolDeprecatedPlacementStrategy is a value-typed struct, so check the inner pointer fields.
			Expect(nodepool.PlacementStrategy.Cpu).To(BeNil())
			Expect(nodepool.PlacementStrategy.Gpu).To(BeNil())
		})

		It("Get nodepools by clusterId only populates deprecated fields", func() {
			Expect(len(createdNodepools)).To(BeNumerically(">", 0))
			exists := createdNodepools[0]

			response, err := nodepoolService.GetNodepools(ctx,
				[]string{fmt.Sprintf("clusterId==%s", exists.ClusterId)},
				"", "", 0, 0)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			list, ok := response.Body.(api.GetNodepools200Response)
			Expect(ok).To(BeTrue())

			found := false
			for _, np := range list.Nodepools {
				// check for all of them that the deprecated fields are populated (non-nil)
				Expect(np.PlacementStrategy.Cpu).ToNot(BeNil(), "nodepool %s missing PlacementStrategy.Cpu", np.Id)
				Expect(np.PlacementStrategy.Gpu).ToNot(BeNil(), "nodepool %s missing PlacementStrategy.Gpu", np.Id)

				if np.Id != exists.Id {
					continue
				}
				found = true
				Expect(np.PlacementStrategy.Cpu).To(Equal(ptr.To(api.PLACEMENTSTRATEGY_BINPACK)))
				Expect(np.PlacementStrategy.Gpu).To(Equal(ptr.To(api.PLACEMENTSTRATEGY_BINPACK)))
			}
			Expect(found).To(BeTrue(), "expected nodepool %s in the list response", exists.Id)
		})

		It("Get nodepools with multi-field filter does not populate deprecated fields", func() {
			Expect(len(createdNodepools)).To(BeNumerically(">", 0))
			exists := createdNodepools[0]

			response, err := nodepoolService.GetNodepools(ctx,
				[]string{
					fmt.Sprintf("clusterId==%s", exists.ClusterId),
					fmt.Sprintf("name==%s", exists.Name),
				},
				"", "", 0, 0)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			list, ok := response.Body.(api.GetNodepools200Response)
			Expect(ok).To(BeTrue())
			Expect(list.Nodepools).ToNot(BeEmpty())

			for _, np := range list.Nodepools {
				Expect(np.PlacementStrategy.Cpu).To(BeNil())
				Expect(np.PlacementStrategy.Gpu).To(BeNil())
			}
		})

		It("Get nodepools with no filter does not populate deprecated fields", func() {
			response, err := nodepoolService.GetNodepools(ctx, []string{}, "", "", 0, 0)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			list, ok := response.Body.(api.GetNodepools200Response)
			Expect(ok).To(BeTrue())
			Expect(list.Nodepools).ToNot(BeEmpty())

			for _, np := range list.Nodepools {
				Expect(np.PlacementStrategy.Cpu).To(BeNil())
				Expect(np.PlacementStrategy.Gpu).To(BeNil())
			}
		})

		It("Count create node pool and count filter by created node pool name", func() {
			nodepoolName := fmt.Sprintf("np-b-%s", rand.String(4))
			labelKey := "c"
			labelValue := "d"
			request := api.NodepoolCreateFields{
				Name:       nodepoolName,
				LabelKey:   labelKey,
				LabelValue: labelValue,
				ClusterId:  testClusterId,
			}

			response, err := nodepoolService.CreateNodepool(ctx, request)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusCreated))

			nodepool, ok := response.Body.(api.Nodepool)
			Expect(ok).To(BeTrue())

			Expect(nodepool.Id).To(Not(BeNil()))
			createdNodepools = append(createdNodepools, nodepool)

			countResponse, err := nodepoolService.CountNodepools(ctx, []string{fmt.Sprintf("clusterId==%s", testClusterId), fmt.Sprintf("name==%s", nodepoolName)})
			Expect(err).To(BeNil())
			Expect(countResponse.Code).To(Equal(http.StatusOK))
			count := countResponse.Body.(api.CountNodes200Response).Count
			Expect(count).To(Equal(int64(1)))
		})

		It("Update nodepool", func() {
			Expect(len(createdNodepools)).To(BeNumerically(">", 0))

			exists := createdNodepools[0]

			request := api.NodepoolUpdateFields{
				LabelKey:   ptr.To("test"),
				LabelValue: ptr.To("test-value"),
				SchedulingConfiguration: &api.SchedulingConfiguration{
					PlacementStrategy: api.SchedulingConfigurationPlacementStrategy{
						Cpu: ptr.To(api.PLACEMENTSTRATEGY_SPREAD),
						Gpu: ptr.To(api.PLACEMENTSTRATEGY_BINPACK),
					},
				},
			}

			response, err := nodepoolService.UpdateNodepool(ctx, exists.Id, request)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			nodepool, ok := response.Body.(api.Nodepool)
			Expect(ok).To(BeTrue())

			Expect(nodepool.Id).To(Equal(exists.Id))
			Expect(nodepool.TenantId).To(Equal(int32(tenantId)))
			Expect(nodepool.Phase).To(Equal(api.NODEPOOLPHASE_UPDATING))
			Expect(nodepool.PhaseMessage).To(Equal(""))
			Expect(nodepool.CreatedAt).To(Not(BeNil()))
			Expect(nodepool.CreatedAt.Unix()).To(BeNumerically(">", time.Now().Unix()-5000)) // time of postgres query should not be near 5 sec
			Expect(nodepool.CreatedBy).To(Equal(testUser))
			Expect(nodepool.UpdatedAt).To(Not(BeNil()))
			Expect(nodepool.UpdatedAt.Unix()).To(BeNumerically(">", time.Now().Unix()-5000)) // time of postgres query should not be near 5 sec
			Expect(nodepool.UpdatedBy).To(Equal(testUser))
			Expect(nodepool.Name).To(Equal(exists.Name))
			Expect(nodepool.LabelKey).To(Equal(*request.LabelKey))
			Expect(nodepool.LabelValue).To(Equal(*request.LabelValue))
			Expect(nodepool.ClusterId).To(Equal(exists.ClusterId))
			Expect(nodepool.SchedulingConfiguration.PlacementStrategy.Cpu).To(Equal(request.SchedulingConfiguration.PlacementStrategy.Cpu))
			Expect(nodepool.SchedulingConfiguration.PlacementStrategy.Gpu).To(Equal(request.SchedulingConfiguration.PlacementStrategy.Gpu))

			response, err = nodepoolService.GetNodepool(ctx, exists.Id)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			nodepool, ok = response.Body.(api.Nodepool)
			Expect(ok).To(BeTrue())

			Expect(nodepool.Id).To(Equal(exists.Id))
			Expect(nodepool.TenantId).To(Equal(int32(tenantId)))
			Expect(nodepool.Phase).To(Equal(api.NODEPOOLPHASE_UPDATING))
			Expect(nodepool.PhaseMessage).To(Equal(""))
			Expect(nodepool.CreatedAt).To(Not(BeNil()))
			Expect(nodepool.CreatedAt.Unix()).To(BeNumerically(">", time.Now().Unix()-5000)) // time of postgres query should not be near 5 sec
			Expect(nodepool.CreatedBy).To(Equal(testUser))
			Expect(nodepool.UpdatedAt).To(Not(BeNil()))
			Expect(nodepool.UpdatedAt.Unix()).To(BeNumerically(">", time.Now().Unix()-5000)) // time of postgres query should not be near 5 sec
			Expect(nodepool.UpdatedBy).To(Equal(testUser))
			Expect(nodepool.Name).To(Equal(exists.Name))
			Expect(nodepool.LabelKey).To(Equal(*request.LabelKey))
			Expect(nodepool.LabelValue).To(Equal(*request.LabelValue))
			Expect(nodepool.ClusterId).To(Equal(exists.ClusterId))
			Expect(nodepool.SchedulingConfiguration.PlacementStrategy.Cpu).To(Equal(request.SchedulingConfiguration.PlacementStrategy.Cpu))
			Expect(nodepool.SchedulingConfiguration.PlacementStrategy.Gpu).To(Equal(request.SchedulingConfiguration.PlacementStrategy.Gpu))

			createdNodepools[0] = nodepool

			// update phase to "Ready"
			response, err = nodepoolService.SyncNodepoolsStatus(ctx, nodepool.ClusterId, api.SyncNodepoolsStatusRequest{
				Nodepools: []api.NodepoolSyncFields{
					{
						Name:         nodepool.Name,
						Phase:        api.NODEPOOLPHASE_READY,
						PhaseMessage: "Nodepool is ready",
					},
				},
				DeleteMissing: false,
			})
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusNoContent))

			// "update" again with same fields - expect nodepool to stay "Ready"
			response, err = nodepoolService.UpdateNodepool(ctx, exists.Id, request)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))
			nodepool, ok = response.Body.(api.Nodepool)
			Expect(ok).To(BeTrue())
			Expect(nodepool.Id).To(Equal(exists.Id))
			Expect(nodepool.TenantId).To(Equal(int32(tenantId)))
			Expect(nodepool.Phase).To(Equal(api.NODEPOOLPHASE_READY))
			Expect(nodepool.PhaseMessage).To(Equal("Nodepool is ready"))
		})
		It("Patch nodepool", func() {
			Expect(len(createdNodepools)).To(BeNumerically(">", 0))

			exists := createdNodepools[0]

			request := api.NodepoolUpdateFields{
				LabelKey:   ptr.To("test-2"),
				LabelValue: ptr.To("test-value-2"),
				SchedulingConfiguration: &api.SchedulingConfiguration{
					PlacementStrategy: api.SchedulingConfigurationPlacementStrategy{
						Gpu: ptr.To(api.PLACEMENTSTRATEGY_SPREAD),
					},
				},
			}

			response, err := nodepoolService.PatchNodepool(ctx, exists.Id, request)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			nodepool, ok := response.Body.(api.Nodepool)
			Expect(ok).To(BeTrue())

			Expect(nodepool.Id).To(Equal(exists.Id))

			Expect(nodepool.TenantId).To(Equal(int32(tenantId)))
			Expect(nodepool.Phase).To(Equal(api.NODEPOOLPHASE_UPDATING))
			Expect(nodepool.PhaseMessage).To(BeEmpty())
			Expect(nodepool.CreatedAt).To(Not(BeNil()))
			Expect(nodepool.CreatedAt.Unix()).To(BeNumerically(">", time.Now().Unix()-5000)) // time of postgres query should not be near 5 sec
			Expect(nodepool.CreatedBy).To(Equal(testUser))
			Expect(nodepool.UpdatedAt).To(Not(BeNil()))
			Expect(nodepool.UpdatedAt.Unix()).To(BeNumerically(">", time.Now().Unix()-5000)) // time of postgres query should not be near 5 sec
			Expect(nodepool.UpdatedBy).To(Equal(testUser))
			Expect(nodepool.Name).To(Equal(exists.Name))
			Expect(nodepool.LabelKey).To(Equal(*request.LabelKey))
			Expect(nodepool.LabelValue).To(Equal(*request.LabelValue))
			Expect(nodepool.ClusterId).To(Equal(exists.ClusterId))
			Expect(nodepool.SchedulingConfiguration.PlacementStrategy.Cpu).To(Equal(exists.SchedulingConfiguration.PlacementStrategy.Cpu))
			Expect(nodepool.SchedulingConfiguration.PlacementStrategy.Gpu).To(Equal(request.SchedulingConfiguration.PlacementStrategy.Gpu))

			createdNodepools[0] = nodepool

			// update phase to "Ready"
			response, err = nodepoolService.SyncNodepoolsStatus(ctx, nodepool.ClusterId, api.SyncNodepoolsStatusRequest{
				Nodepools: []api.NodepoolSyncFields{
					{
						Name:         nodepool.Name,
						Phase:        api.NODEPOOLPHASE_READY,
						PhaseMessage: "Nodepool is ready",
					},
				},
				DeleteMissing: false,
			})
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusNoContent))

			// "patch" again with same fields - expect nodepool to stay "Ready"
			response, err = nodepoolService.PatchNodepool(ctx, exists.Id, request)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))
			nodepool, ok = response.Body.(api.Nodepool)
			Expect(ok).To(BeTrue())
			Expect(nodepool.Id).To(Equal(exists.Id))
			Expect(nodepool.TenantId).To(Equal(int32(tenantId)))
			Expect(nodepool.Phase).To(Equal(api.NODEPOOLPHASE_READY))
			Expect(nodepool.PhaseMessage).To(Equal("Nodepool is ready"))
		})
		It("Delete nodepool created", func() {
			Expect(len(createdNodepools)).To(BeNumerically(">", 0))

			exists := createdNodepools[0]

			err := dao.SetNodepoolPhase(ctx, tenantId, testUser, exists.Id, api.NODEPOOLPHASE_READY)
			Expect(err).To(BeNil())

			response, err := nodepoolService.DeleteNodepool(ctx, exists.Id)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusAccepted))

			response, err = nodepoolService.GetNodepool(ctx, exists.Id)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			nodepool, ok := response.Body.(api.Nodepool)
			Expect(ok).To(BeTrue())

			Expect(nodepool.Id).To(Equal(exists.Id))

			Expect(nodepool.TenantId).To(Equal(int32(tenantId)))
			Expect(nodepool.Phase).To(Equal(api.NODEPOOLPHASE_DELETING))
			createdNodepools[0] = nodepool
		})
		It("Delete nodepool in creating", func() {
			Expect(len(createdNodepools)).To(BeNumerically(">", 0))

			exists := createdNodepools[0]

			err := dao.SetNodepoolPhase(ctx, tenantId, testUser, exists.Id, api.NODEPOOLPHASE_CREATING)
			Expect(err).To(BeNil())

			response, err := nodepoolService.DeleteNodepool(ctx, exists.Id)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusNoContent))

			response, err = nodepoolService.GetNodepool(ctx, exists.Id)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusNotFound))
			createdNodepools = createdNodepools[1:]
		})

		It("Re-Create with same name after deletion", func() {
			request := api.NodepoolCreateFields{
				Name:       testName,
				LabelKey:   testLabelKey,
				LabelValue: testLabelValue,
				ClusterId:  testClusterId,
				SchedulingConfiguration: &api.SchedulingConfiguration{
					PlacementStrategy: api.SchedulingConfigurationPlacementStrategy{
						Cpu: ptr.To(api.PLACEMENTSTRATEGY_BINPACK),
						Gpu: ptr.To(api.PLACEMENTSTRATEGY_BINPACK),
					},
				},
			}

			response, err := nodepoolService.CreateNodepool(ctx, request)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusCreated))

			nodepool, ok := response.Body.(api.Nodepool)
			Expect(ok).To(BeTrue())

			// update phase to Ready using Sync - there was a bug related to that, so testing this specifically
			response, err = nodepoolService.SyncNodepoolsStatus(ctx, testClusterId, api.SyncNodepoolsStatusRequest{
				Nodepools: []api.NodepoolSyncFields{
					{
						Name:         testName,
						Phase:        api.NODEPOOLPHASE_READY,
						PhaseMessage: "Nodepool is ready",
					},
				},
				DeleteMissing: false,
			})
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusNoContent))

			response, err = nodepoolService.DeleteNodepool(ctx, nodepool.Id)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusAccepted))

			err = dao.SetNodepoolPhase(ctx, tenantId, testUser, nodepool.Id, api.NODEPOOLPHASE_DELETED)
			Expect(err).To(BeNil())

			response, err = nodepoolService.GetNodepool(ctx, nodepool.Id)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusNotFound))
		})
	})

	Context("Sync deletions", Ordered, func() {
		BeforeAll(func() {
			for _, nodepool := range createdNodepools {
				err := dao.DeleteNodepool(ctx, tenantId, "", nodepool.Id)
				Expect(err).To(BeNil())
			}
			createdNodepools = []api.Nodepool{}
		})
		nodepoolsCount := 5

		BeforeAll(func() {
			for i := 0; i < nodepoolsCount; i++ {
				request := api.NodepoolCreateFields{
					Name:       fmt.Sprintf("%s-%d-%s", testName, i, rand.String(4)),
					LabelKey:   fmt.Sprintf("%s-%d", testLabelKey, i),
					LabelValue: fmt.Sprintf("%s-%d", testLabelValue, i),
					ClusterId:  testClusterId,
					SchedulingConfiguration: &api.SchedulingConfiguration{
						PlacementStrategy: api.SchedulingConfigurationPlacementStrategy{
							Cpu: ptr.To(api.PLACEMENTSTRATEGY_SPREAD),
							Gpu: ptr.To(api.PLACEMENTSTRATEGY_SPREAD),
						},
					},
				}

				response, err := nodepoolService.CreateNodepool(ctx, request)
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusCreated))

				nodepool, ok := response.Body.(api.Nodepool)
				Expect(ok).To(BeTrue())

				Expect(nodepool.Id).To(Not(BeNil()))
				createdNodepools = append(createdNodepools, nodepool)
			}

			defaultNodePool := getDefaultNodePoolOfCluster(ctx, nodepoolService, testClusterId)
			Expect(defaultNodePool).To(Not(BeNil()))
		})
		It("Sync 1 nodepool", func() {
			Expect(len(createdNodepools)).To(BeNumerically(">", 0))

			exists := createdNodepools[0]

			response, err := nodepoolService.SyncNodepoolsStatus(ctx, exists.ClusterId, api.SyncNodepoolsStatusRequest{
				Nodepools: []api.NodepoolSyncFields{
					{
						Name:         exists.Name,
						Phase:        api.NODEPOOLPHASE_READY,
						PhaseMessage: "Nodepool is ready",
						Status: api.NodepoolSyncFieldsStatus{
							LabelKey:   ptr.To(exists.LabelKey),
							LabelValue: ptr.To(exists.LabelValue),
							SchedulingConfiguration: &api.SchedulingConfiguration{
								PlacementStrategy: api.SchedulingConfigurationPlacementStrategy{
									Cpu: exists.SchedulingConfiguration.PlacementStrategy.Cpu,
									Gpu: exists.SchedulingConfiguration.PlacementStrategy.Gpu,
								},
							},
							Nodes: []string{
								"worker-a",
								"worker-b",
							},
						},
					},
				},
				DeleteMissing: false,
			})
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusNoContent))

			response, err = nodepoolService.GetNodepools(ctx, []string{}, "", "", 0, 0)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			nodepools, ok := response.Body.(api.GetNodepools200Response)
			Expect(ok).To(BeTrue())

			// +1 for the default nodepool
			Expect(len(nodepools.Nodepools)).To(Equal(nodepoolsCount + 1))
		})
		It("Sync all nodepools - with deletion", func() {
			Expect(len(createdNodepools)).To(BeNumerically(">", 0))

			nodepoolSync := []api.NodepoolSyncFields{}

			for _, np := range createdNodepools {
				nodepoolSync = append(nodepoolSync, api.NodepoolSyncFields{
					Name:         np.Name,
					Phase:        api.NODEPOOLPHASE_READY,
					PhaseMessage: "Nodepool is ready",
					Status: api.NodepoolSyncFieldsStatus{
						LabelKey:   ptr.To(np.LabelKey),
						LabelValue: ptr.To(np.LabelValue),
						SchedulingConfiguration: &api.SchedulingConfiguration{
							PlacementStrategy: api.SchedulingConfigurationPlacementStrategy{
								Cpu: np.SchedulingConfiguration.PlacementStrategy.Cpu,
								Gpu: np.SchedulingConfiguration.PlacementStrategy.Gpu,
							},
						},
						Nodes: []string{
							"worker-a",
							"worker-b",
						},
					},
				})
			}

			response, err := nodepoolService.SyncNodepoolsStatus(ctx, testClusterId, api.SyncNodepoolsStatusRequest{
				Nodepools:     nodepoolSync,
				DeleteMissing: true,
			})
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusNoContent))

			response, err = nodepoolService.GetNodepools(ctx, []string{}, "", "", 0, 0)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			nodepools, ok := response.Body.(api.GetNodepools200Response)
			Expect(ok).To(BeTrue())

			// +1 for the default nodepool
			Expect(len(nodepools.Nodepools)).To(Equal(nodepoolsCount + 1))
		})
		It("Sync all nodepools - delete all in ready state", func() {
			Expect(len(createdNodepools)).To(BeNumerically(">", 0))

			nodepoolSync := []api.NodepoolSyncFields{}

			for _, np := range createdNodepools {
				nodepoolSync = append(nodepoolSync, api.NodepoolSyncFields{
					Name:         np.Name,
					Phase:        api.NODEPOOLPHASE_DELETED,
					PhaseMessage: "",
					Status: api.NodepoolSyncFieldsStatus{
						LabelKey:   ptr.To(np.LabelKey),
						LabelValue: ptr.To(np.LabelValue),
						SchedulingConfiguration: &api.SchedulingConfiguration{
							PlacementStrategy: api.SchedulingConfigurationPlacementStrategy{
								Cpu: np.SchedulingConfiguration.PlacementStrategy.Cpu,
								Gpu: np.SchedulingConfiguration.PlacementStrategy.Gpu,
							},
						},
						Nodes: []string{
							"worker-a",
							"worker-b",
						},
					},
				})
			}

			response, err := nodepoolService.SyncNodepoolsStatus(ctx, testClusterId, api.SyncNodepoolsStatusRequest{
				Nodepools: nodepoolSync,
			})
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusNoContent))

			response, err = nodepoolService.GetNodepools(ctx, []string{}, "", "", 0, 0)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			nodepools, ok := response.Body.(api.GetNodepools200Response)
			Expect(ok).To(BeTrue())

			// +1 for the default nodepool
			Expect(len(nodepools.Nodepools)).To(Equal(nodepoolsCount + 1))
		})
		It("Sync some nodepools - with delete missing not in deleting phase", func() {
			Expect(len(createdNodepools)).To(BeNumerically(">", 2))

			nodepoolSync := []api.NodepoolSyncFields{}

			for i := 0; i < nodepoolsCount-1; i++ {
				np := createdNodepools[i]
				nodepoolSync = append(nodepoolSync, api.NodepoolSyncFields{
					Name:         np.Name,
					Phase:        api.NODEPOOLPHASE_READY,
					PhaseMessage: "Nodepool is ready",
					Status: api.NodepoolSyncFieldsStatus{
						LabelKey:   ptr.To(np.LabelKey),
						LabelValue: ptr.To(np.LabelValue),
						SchedulingConfiguration: &api.SchedulingConfiguration{
							PlacementStrategy: api.SchedulingConfigurationPlacementStrategy{
								Cpu: np.SchedulingConfiguration.PlacementStrategy.Cpu,
								Gpu: np.SchedulingConfiguration.PlacementStrategy.Gpu,
							},
						},
						Nodes: []string{
							"worker-a",
							"worker-b",
						},
					},
				})
			}

			response, err := nodepoolService.SyncNodepoolsStatus(ctx, testClusterId, api.SyncNodepoolsStatusRequest{
				Nodepools:     nodepoolSync,
				DeleteMissing: true,
			})
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusNoContent))

			response, err = nodepoolService.GetNodepools(ctx, []string{}, "", "", 0, 0)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			nodepools, ok := response.Body.(api.GetNodepools200Response)
			Expect(ok).To(BeTrue())

			// +1 for the default nodepool
			Expect(len(nodepools.Nodepools)).To(Equal(nodepoolsCount + 1))

			createdNodepools = nodepools.Nodepools
		})
		It("Sync some nodepools - with delete missing in deleting phase", func() {
			Expect(len(createdNodepools)).To(BeNumerically(">", 2))

			np := createdNodepools[0]

			response, err := nodepoolService.DeleteNodepool(ctx, np.Id)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusAccepted))

			np2 := createdNodepools[1]
			response, err = nodepoolService.SyncNodepoolsStatus(ctx, np2.ClusterId, api.SyncNodepoolsStatusRequest{
				Nodepools: []api.NodepoolSyncFields{
					{
						Name:         np2.Name,
						Phase:        api.NODEPOOLPHASE_READY,
						PhaseMessage: "Nodepool is ready",
						Status: api.NodepoolSyncFieldsStatus{
							LabelKey:   ptr.To(np2.LabelKey),
							LabelValue: ptr.To(np2.LabelValue),
							SchedulingConfiguration: &api.SchedulingConfiguration{
								PlacementStrategy: api.SchedulingConfigurationPlacementStrategy{
									Cpu: np2.SchedulingConfiguration.PlacementStrategy.Cpu,
									Gpu: np2.SchedulingConfiguration.PlacementStrategy.Gpu,
								},
							},
							Nodes: []string{
								"worker-a",
								"worker-b",
							},
						},
					},
				},
				DeleteMissing: true,
			})

			response, err = nodepoolService.GetNodepools(ctx, []string{}, "", "", 0, 0)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			nodepools, ok := response.Body.(api.GetNodepools200Response)
			Expect(ok).To(BeTrue())

			// +1 for the default nodepool
			Expect(len(nodepools.Nodepools)).To(Equal(nodepoolsCount + 1 - 1))

			createdNodepools = nodepools.Nodepools
		})
		It("Try to delete default nodepool using sync API", func() {
			response, err := nodepoolService.SyncNodepoolsStatus(ctx, testClusterId, api.SyncNodepoolsStatusRequest{
				Nodepools: []api.NodepoolSyncFields{
					{
						Name:  services.NodepoolDefaultName,
						Phase: api.NODEPOOLPHASE_DELETED,
					},
				},
				DeleteMissing: false,
			})
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusNoContent))

			defaultNodePool := getDefaultNodePoolOfCluster(ctx, nodepoolService, testClusterId)
			Expect(defaultNodePool).ToNot(BeNil())
			Expect(defaultNodePool.Phase).To(Equal(api.NODEPOOLPHASE_READY))
		})
		It("Try to delete default nodepool using sync API with delete missing", func() {
			Expect(len(createdNodepools)).To(BeNumerically(">", 0))
			np := createdNodepools[0]

			response, err := nodepoolService.SyncNodepoolsStatus(ctx, testClusterId, api.SyncNodepoolsStatusRequest{
				Nodepools: []api.NodepoolSyncFields{
					{
						Name:         np.Name,
						Phase:        api.NODEPOOLPHASE_READY,
						PhaseMessage: "Nodepool is ready",
						Status: api.NodepoolSyncFieldsStatus{
							LabelKey:   ptr.To(np.LabelKey),
							LabelValue: ptr.To(np.LabelValue),
							SchedulingConfiguration: &api.SchedulingConfiguration{
								PlacementStrategy: api.SchedulingConfigurationPlacementStrategy{
									Cpu: np.SchedulingConfiguration.PlacementStrategy.Cpu,
									Gpu: np.SchedulingConfiguration.PlacementStrategy.Gpu,
								},
							},
							Nodes: []string{
								"worker-a",
								"worker-b",
							},
						},
					},
				},
				DeleteMissing: true,
			})
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusNoContent))

			defaultNodePool := getDefaultNodePoolOfCluster(ctx, nodepoolService, testClusterId)
			Expect(defaultNodePool).ToNot(BeNil())
			Expect(defaultNodePool.Phase).To(Equal(api.NODEPOOLPHASE_READY))
		})
		It("Try to delete default nodepool using sync API - change to 'Deleting' first", func() {
			response, err := nodepoolService.SyncNodepoolsStatus(ctx, testClusterId, api.SyncNodepoolsStatusRequest{
				Nodepools: []api.NodepoolSyncFields{
					{
						Name:  services.NodepoolDefaultName,
						Phase: api.NODEPOOLPHASE_DELETING,
					},
				},
				DeleteMissing: false,
			})
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusNoContent))

			defaultNodePool := getDefaultNodePoolOfCluster(ctx, nodepoolService, testClusterId)
			Expect(defaultNodePool).ToNot(BeNil())
			Expect(defaultNodePool.Phase).To(Equal(api.NODEPOOLPHASE_READY))
		})
		It("Don't allow any nodepool to be updated to 'Deleting' from cluster", func() {
			Expect(len(createdNodepools)).To(BeNumerically(">", 0))
			np := createdNodepools[0]
			response, err := nodepoolService.SyncNodepoolsStatus(ctx, testClusterId, api.SyncNodepoolsStatusRequest{
				Nodepools: []api.NodepoolSyncFields{
					{
						Name:  np.Name,
						Phase: api.NODEPOOLPHASE_DELETING,
					},
				},
				DeleteMissing: false,
			})
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusNoContent))

			response, err = nodepoolService.GetNodepool(ctx, np.Id)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))
			Expect(response.Body.(api.Nodepool).Phase).ToNot(Equal(api.NODEPOOLPHASE_DELETING))
		})
	})
	Context("Stuck on deleting", Ordered, func() {
		var existsNp api.Nodepool
		BeforeAll(func() {
			for _, nodepool := range createdNodepools {
				err := dao.DeleteNodepool(ctx, tenantId, "", nodepool.Id)
				Expect(err).To(BeNil())
			}
			createdNodepools = []api.Nodepool{}
		})

		BeforeAll(func() {
			request := api.NodepoolCreateFields{
				Name:       fmt.Sprintf("%s-%s", testName, rand.String(4)),
				LabelKey:   testLabelKey,
				LabelValue: testLabelValue,
				ClusterId:  testClusterId,
				SchedulingConfiguration: &api.SchedulingConfiguration{
					PlacementStrategy: api.SchedulingConfigurationPlacementStrategy{
						Cpu: ptr.To(api.PLACEMENTSTRATEGY_SPREAD),
						Gpu: ptr.To(api.PLACEMENTSTRATEGY_SPREAD),
					},
				},
			}

			response, err := nodepoolService.CreateNodepool(ctx, request)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusCreated))

			nodepool, ok := response.Body.(api.Nodepool)
			Expect(ok).To(BeTrue())

			Expect(nodepool.Id).To(Not(BeNil()))
			Expect(nodepool.Phase).To(Equal(api.NODEPOOLPHASE_CREATING))
			createdNodepools = append(createdNodepools, nodepool)
			existsNp = nodepool
		})
		It("Update and delete (no sync)", func() {
			Expect(existsNp).To(Not(BeNil()))
			request := api.NodepoolUpdateFields{
				LabelKey:   ptr.To("test"),
				LabelValue: ptr.To("test-value"),
				SchedulingConfiguration: &api.SchedulingConfiguration{
					PlacementStrategy: api.SchedulingConfigurationPlacementStrategy{
						Cpu: ptr.To(api.PLACEMENTSTRATEGY_SPREAD),
						Gpu: ptr.To(api.PLACEMENTSTRATEGY_BINPACK),
					},
				},
			}

			response, err := nodepoolService.UpdateNodepool(ctx, existsNp.Id, request)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			nodepool, ok := response.Body.(api.Nodepool)
			Expect(ok).To(BeTrue())

			Expect(nodepool.Id).To(Equal(existsNp.Id))
			Expect(nodepool.TenantId).To(Equal(int32(tenantId)))
			Expect(nodepool.Phase).To(Equal(api.NODEPOOLPHASE_CREATING))

			response, err = nodepoolService.PatchNodepool(ctx, existsNp.Id, request)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			nodepool, ok = response.Body.(api.Nodepool)
			Expect(ok).To(BeTrue())

			Expect(nodepool.Id).To(Equal(existsNp.Id))

			Expect(nodepool.TenantId).To(Equal(int32(tenantId)))
			Expect(nodepool.Phase).To(Equal(api.NODEPOOLPHASE_CREATING))

			response, err = nodepoolService.DeleteNodepool(ctx, existsNp.Id)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusNoContent))

			response, err = nodepoolService.GetNodepool(ctx, existsNp.Id)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusNotFound))
			createdNodepools = createdNodepools[1:]
		})
	})

	Context("SchedulingConfiguration params", Ordered, func() {
		var schedConfig *api.SchedulingConfiguration
		BeforeEach(func() {
			schedConfig = &api.SchedulingConfiguration{
				PlacementStrategy: api.SchedulingConfigurationPlacementStrategy{
					Cpu: ptr.To(api.PLACEMENTSTRATEGY_SPREAD),
					Gpu: ptr.To(api.PLACEMENTSTRATEGY_BINPACK),
				},
				MinGuaranteedRuntime: ptr.To("3600s"),
			}
		})

		It("Create nodepool with SchedulingConfiguration", func() {
			request := api.NodepoolCreateFields{
				Name:                    testName + "-sched-config",
				LabelKey:                testLabelKey + "-sched-config",
				LabelValue:              testLabelValue + "-sched-config",
				ClusterId:               testClusterId,
				SchedulingConfiguration: schedConfig,
			}
			response, err := nodepoolService.CreateNodepool(ctx, request)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusCreated))
			nodepool, ok := response.Body.(api.Nodepool)
			Expect(ok).To(BeTrue())
			createdNodepools = append(createdNodepools, nodepool)
			Expect(nodepool.SchedulingConfiguration).ToNot(BeNil())
			Expect(nodepool.SchedulingConfiguration.PlacementStrategy.Cpu).To(Equal(schedConfig.PlacementStrategy.Cpu))
			Expect(nodepool.SchedulingConfiguration.PlacementStrategy.Gpu).To(Equal(schedConfig.PlacementStrategy.Gpu))
			Expect(nodepool.SchedulingConfiguration.MinGuaranteedRuntime).To(Equal(schedConfig.MinGuaranteedRuntime))
			Expect(nodepool.SchedulingConfiguration.PlacementStrategy.Cpu).To(Equal(schedConfig.PlacementStrategy.Cpu))
			Expect(nodepool.SchedulingConfiguration.PlacementStrategy.Gpu).To(Equal(schedConfig.PlacementStrategy.Gpu))
		})

		It("Update nodepool SchedulingConfiguration", func() {
			// Create first
			createReq := api.NodepoolCreateFields{
				Name:                    testName + "-sched-update",
				LabelKey:                testLabelKey + "-sched-update",
				LabelValue:              testLabelValue + "-sched-update",
				ClusterId:               testClusterId,
				SchedulingConfiguration: schedConfig,
			}
			response, err := nodepoolService.CreateNodepool(ctx, createReq)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusCreated))
			nodepool := response.Body.(api.Nodepool)
			createdNodepools = append(createdNodepools, nodepool)
			// Update
			updateConfig := &api.SchedulingConfiguration{
				PlacementStrategy: api.SchedulingConfigurationPlacementStrategy{
					Cpu: ptr.To(api.PLACEMENTSTRATEGY_BINPACK),
					Gpu: ptr.To(api.PLACEMENTSTRATEGY_SPREAD),
				},
				MinGuaranteedRuntime: ptr.To("7200s"),
			}

			updateFields := nodepoolUpdateFields("-sched-update", nil)
			updateFields.SchedulingConfiguration = updateConfig
			updateResp, err := nodepoolService.UpdateNodepool(ctx, nodepool.Id, *updateFields)
			Expect(err).To(BeNil())
			Expect(updateResp.Code).To(Equal(http.StatusOK))
			updated := updateResp.Body.(api.Nodepool)
			Expect(updated.SchedulingConfiguration).ToNot(BeNil())
			Expect(updated.SchedulingConfiguration.PlacementStrategy.Cpu).To(Equal(updateConfig.PlacementStrategy.Cpu))
			Expect(updated.SchedulingConfiguration.PlacementStrategy.Gpu).To(Equal(updateConfig.PlacementStrategy.Gpu))
			Expect(updated.SchedulingConfiguration.MinGuaranteedRuntime).To(Equal(updateConfig.MinGuaranteedRuntime))
			Expect(updated.SchedulingConfiguration.PlacementStrategy.Cpu).To(Equal(updateConfig.PlacementStrategy.Cpu))
			Expect(updated.SchedulingConfiguration.PlacementStrategy.Gpu).To(Equal(updateConfig.PlacementStrategy.Gpu))
		})

		It("Patch nodepool SchedulingConfiguration", func() {
			// Create first
			createReq := api.NodepoolCreateFields{
				Name:                    testName + "-sched-patch",
				LabelKey:                testLabelKey + "-sched-patch",
				LabelValue:              testLabelValue + "-sched-patch",
				ClusterId:               testClusterId,
				SchedulingConfiguration: schedConfig,
			}
			response, err := nodepoolService.CreateNodepool(ctx, createReq)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusCreated))
			nodepool := response.Body.(api.Nodepool)
			createdNodepools = append(createdNodepools, nodepool)
			patchConfig := &api.SchedulingConfiguration{
				PlacementStrategy: api.SchedulingConfigurationPlacementStrategy{
					Cpu: ptr.To(api.PLACEMENTSTRATEGY_SPREAD),
					Gpu: ptr.To(api.PLACEMENTSTRATEGY_SPREAD),
				},
				MinGuaranteedRuntime: ptr.To("5d8h40m"),
			}
			patchReq := api.NodepoolUpdateFields{
				SchedulingConfiguration: patchConfig,
			}
			patchResp, err := nodepoolService.PatchNodepool(ctx, nodepool.Id, patchReq)
			Expect(err).To(BeNil())
			Expect(patchResp.Code).To(Equal(http.StatusOK))
			patched := patchResp.Body.(api.Nodepool)
			Expect(patched.SchedulingConfiguration).ToNot(BeNil())
			Expect(patched.SchedulingConfiguration.PlacementStrategy.Cpu).To(Equal(patchConfig.PlacementStrategy.Cpu))
			Expect(patched.SchedulingConfiguration.PlacementStrategy.Gpu).To(Equal(patchConfig.PlacementStrategy.Gpu))
			Expect(patched.SchedulingConfiguration.MinGuaranteedRuntime).To(Equal(patchConfig.MinGuaranteedRuntime))
			Expect(patched.SchedulingConfiguration.PlacementStrategy.Cpu).To(Equal(patchConfig.PlacementStrategy.Cpu))
			Expect(patched.SchedulingConfiguration.PlacementStrategy.Gpu).To(Equal(patchConfig.PlacementStrategy.Gpu))
		})

		It("Patch with partial fields - only cpu, only min runtime - validate rest is unchanged", func() {
			// Create with all fields
			createReq := api.NodepoolCreateFields{
				Name:       testName + "-sched-partial-patch",
				LabelKey:   testLabelKey + "-sched-partial-patch",
				LabelValue: testLabelValue + "-sched-partial-patch",
				ClusterId:  testClusterId,
				SchedulingConfiguration: &api.SchedulingConfiguration{
					PlacementStrategy: api.SchedulingConfigurationPlacementStrategy{
						Cpu: ptr.To(api.PLACEMENTSTRATEGY_SPREAD),
						Gpu: ptr.To(api.PLACEMENTSTRATEGY_BINPACK),
					},
					MinGuaranteedRuntime: ptr.To("3600s"),
				},
			}
			response, err := nodepoolService.CreateNodepool(ctx, createReq)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusCreated))
			nodepool := response.Body.(api.Nodepool)
			createdNodepools = append(createdNodepools, nodepool)

			// Patch only CPU
			patchFields := api.NodepoolUpdateFields{
				SchedulingConfiguration: &api.SchedulingConfiguration{
					PlacementStrategy: api.SchedulingConfigurationPlacementStrategy{
						Cpu: ptr.To(api.PLACEMENTSTRATEGY_BINPACK),
					},
				},
			}
			patchResp, err := nodepoolService.PatchNodepool(ctx, nodepool.Id, patchFields)
			Expect(err).To(BeNil())
			Expect(patchResp.Code).To(Equal(http.StatusOK))
			patched := patchResp.Body.(api.Nodepool)
			Expect(patched.SchedulingConfiguration.PlacementStrategy.Cpu).To(Equal(ptr.To(api.PLACEMENTSTRATEGY_BINPACK)))
			Expect(patched.SchedulingConfiguration.PlacementStrategy.Gpu).To(Equal(ptr.To(api.PLACEMENTSTRATEGY_BINPACK)))
			Expect(patched.SchedulingConfiguration.MinGuaranteedRuntime).To(Equal(ptr.To("3600s")))
			Expect(patched.SchedulingConfiguration.PlacementStrategy.Cpu).To(Equal(ptr.To(api.PLACEMENTSTRATEGY_BINPACK)))
			Expect(patched.SchedulingConfiguration.PlacementStrategy.Gpu).To(Equal(ptr.To(api.PLACEMENTSTRATEGY_BINPACK)))

			// Patch only MinGuaranteedRuntime
			patchFields = api.NodepoolUpdateFields{
				SchedulingConfiguration: &api.SchedulingConfiguration{
					MinGuaranteedRuntime: ptr.To("2d4h"),
				},
			}
			patchResp, err = nodepoolService.PatchNodepool(ctx, nodepool.Id, patchFields)
			Expect(err).To(BeNil())
			Expect(patchResp.Code).To(Equal(http.StatusOK))
			patched = patchResp.Body.(api.Nodepool)
			Expect(patched.SchedulingConfiguration.PlacementStrategy.Cpu).To(Equal(ptr.To(api.PLACEMENTSTRATEGY_BINPACK)))
			Expect(patched.SchedulingConfiguration.PlacementStrategy.Gpu).To(Equal(ptr.To(api.PLACEMENTSTRATEGY_BINPACK)))
			Expect(patched.SchedulingConfiguration.MinGuaranteedRuntime).To(Equal(ptr.To("2d4h")))
			Expect(patched.SchedulingConfiguration.PlacementStrategy.Cpu).To(Equal(ptr.To(api.PLACEMENTSTRATEGY_BINPACK)))
			Expect(patched.SchedulingConfiguration.PlacementStrategy.Gpu).To(Equal(ptr.To(api.PLACEMENTSTRATEGY_BINPACK)))

			// patch using the deprecated field - should work
			patchFields = api.NodepoolUpdateFields{
				SchedulingConfiguration: &api.SchedulingConfiguration{
					PlacementStrategy: api.SchedulingConfigurationPlacementStrategy{
						Gpu: ptr.To(api.PLACEMENTSTRATEGY_SPREAD),
					},
				},
			}
			patchResp, err = nodepoolService.PatchNodepool(ctx, nodepool.Id, patchFields)
			Expect(err).To(BeNil())
			Expect(patchResp.Code).To(Equal(http.StatusOK))
			patched = patchResp.Body.(api.Nodepool)
			Expect(patched.SchedulingConfiguration.PlacementStrategy.Cpu).To(Equal(ptr.To(api.PLACEMENTSTRATEGY_BINPACK)))
			Expect(patched.SchedulingConfiguration.PlacementStrategy.Gpu).To(Equal(ptr.To(api.PLACEMENTSTRATEGY_SPREAD)))
			Expect(patched.SchedulingConfiguration.MinGuaranteedRuntime).To(Equal(ptr.To("2d4h")))
			Expect(patched.SchedulingConfiguration.PlacementStrategy.Cpu).To(Equal(ptr.To(api.PLACEMENTSTRATEGY_BINPACK)))
			Expect(patched.SchedulingConfiguration.PlacementStrategy.Gpu).To(Equal(ptr.To(api.PLACEMENTSTRATEGY_SPREAD)))
		})

		It("MinGuaranteedRuntime - invalid field value", func() {
			invalidVal := "notaduration"
			// Create
			createReq := api.NodepoolCreateFields{
				Name:       testName + "-sched-invalid",
				LabelKey:   testLabelKey + "-sched-invalid",
				LabelValue: testLabelValue + "-sched-invalid",
				ClusterId:  testClusterId,
				SchedulingConfiguration: &api.SchedulingConfiguration{
					PlacementStrategy: api.SchedulingConfigurationPlacementStrategy{
						Cpu: ptr.To(api.PLACEMENTSTRATEGY_SPREAD),
						Gpu: ptr.To(api.PLACEMENTSTRATEGY_BINPACK),
					},
					MinGuaranteedRuntime: ptr.To(invalidVal),
				},
			}
			resp, err := nodepoolService.CreateNodepool(ctx, createReq)
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusBadRequest))
			Expect(resp.Body).To(ContainSubstring("failed to parse MinGuaranteedRuntime"))
			// Update
			createReq.SchedulingConfiguration.MinGuaranteedRuntime = ptr.To("3600s")
			resp, err = nodepoolService.CreateNodepool(ctx, createReq)
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusCreated))
			np := resp.Body.(api.Nodepool)
			createdNodepools = append(createdNodepools, np)
			updateFields := nodepoolUpdateFields("-sched-invalid", nil)
			updateFields.SchedulingConfiguration = &api.SchedulingConfiguration{
				MinGuaranteedRuntime: ptr.To(invalidVal),
			}
			resp, err = nodepoolService.UpdateNodepool(ctx, np.Id, *updateFields)
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusBadRequest))
			Expect(resp.Body).To(ContainSubstring("failed to parse MinGuaranteedRuntime"))
			// Patch
			patchFields := api.NodepoolUpdateFields{
				SchedulingConfiguration: &api.SchedulingConfiguration{
					MinGuaranteedRuntime: ptr.To(invalidVal),
				},
			}
			resp, err = nodepoolService.PatchNodepool(ctx, np.Id, patchFields)
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusBadRequest))
			Expect(resp.Body).To(ContainSubstring("failed to parse MinGuaranteedRuntime"))

			// validate negative is not allowed
			updateFields.SchedulingConfiguration = &api.SchedulingConfiguration{
				MinGuaranteedRuntime: ptr.To("-2d4h"),
			}
			resp, err = nodepoolService.UpdateNodepool(ctx, np.Id, *updateFields)
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusBadRequest))
			Expect(resp.Body).To(ContainSubstring("cannot be negative"))
		})

		It("Create nodepool with TimeBasedFairShare disabled and extra fields should fail", func() {
			request := api.NodepoolCreateFields{
				Name:       testName + "-tbf-disabled-extra",
				LabelKey:   testLabelKey + "-tbf-disabled-extra",
				LabelValue: testLabelValue + "-tbf-disabled-extra",
				ClusterId:  testClusterId,
				SchedulingConfiguration: &api.SchedulingConfiguration{
					PlacementStrategy: api.SchedulingConfigurationPlacementStrategy{
						Cpu: ptr.To(api.PLACEMENTSTRATEGY_SPREAD),
						Gpu: ptr.To(api.PLACEMENTSTRATEGY_BINPACK),
					},
					TimeBasedFairShare: &api.TimeBasedFairShare{
						Enabled:               ptr.To(false),
						HistoricalUsageWeight: ptr.To(float32(1.5)),
					},
				},
			}

			resp, err := nodepoolService.CreateNodepool(ctx, request)
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusBadRequest))
			Expect(resp.Body).To(ContainSubstring("timeBasedFairShare fields must not be set when enabled is false"))
		})

		It("Create nodepool with TimeBasedFairShare enabled and defaults applied", func() {
			request := api.NodepoolCreateFields{
				Name:       testName + "-tbf-defaults",
				LabelKey:   testLabelKey + "-tbf-defaults",
				LabelValue: testLabelValue + "-tbf-defaults",
				ClusterId:  testClusterId,
				SchedulingConfiguration: &api.SchedulingConfiguration{
					PlacementStrategy: api.SchedulingConfigurationPlacementStrategy{
						Cpu: ptr.To(api.PLACEMENTSTRATEGY_SPREAD),
						Gpu: ptr.To(api.PLACEMENTSTRATEGY_BINPACK),
					},
					TimeBasedFairShare: &api.TimeBasedFairShare{
						Enabled: ptr.To(true),
					},
				},
			}

			resp, err := nodepoolService.CreateNodepool(ctx, request)
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusCreated))

			nodepool := resp.Body.(api.Nodepool)
			createdNodepools = append(createdNodepools, nodepool)

			expected := api.TimeBasedFairShare{
				Enabled:               ptr.To(true),
				HistoricalUsageWeight: ptr.To(services.DefaultTimeBasedFairShareHistoricalUsageWeight),
				DecayHalfLife:         ptr.To(services.DefaultTimeBasedFairShareDecayHalfLife),
				WindowDuration:        ptr.To(services.DefaultTimeBasedFairShareWindowDuration),
				WindowType:            ptr.To(services.DefaultTimeBasedFairShareWindowType),
				// TumblingWindowStartTime: nil
			}

			// Validate response from Create
			validateTimeBasedFairShareConfig(nodepool, expected)

			// Validate response from GetNodepool
			getResp, err := nodepoolService.GetNodepool(ctx, nodepool.Id)
			Expect(err).To(BeNil())
			Expect(getResp.Code).To(Equal(http.StatusOK))
			fromGet := getResp.Body.(api.Nodepool)
			validateTimeBasedFairShareConfig(fromGet, expected)

			// Validate response from GetNodepools (filtering by clusterId and name)
			listResp, err := nodepoolService.GetNodepools(ctx,
				[]string{
					fmt.Sprintf("clusterId==%s", nodepool.ClusterId),
					fmt.Sprintf("name==%s", nodepool.Name),
				},
				"", "", 0, 0)
			Expect(err).To(BeNil())
			Expect(listResp.Code).To(Equal(http.StatusOK))
			list := listResp.Body.(api.GetNodepools200Response)
			Expect(list.Nodepools).To(HaveLen(1))
			validateTimeBasedFairShareConfig(list.Nodepools[0], expected)
		})

		It("Create nodepool with TimeBasedFairShare full configuration", func() {
			historicalUsageWeight := float32(2.5)
			decayHalfLife := "24h40m"
			windowDuration := "2d3h"
			windowType := api.TIMEBASEDFAIRSHAREWINDOWTYPE_TUMBLING
			tumblingStart := time.Date(2025, 5, 8, 16, 0, 0, 0, time.UTC)

			nodePoolTimeBasedFairShare := api.TimeBasedFairShare{
				Enabled:                 ptr.To(true),
				HistoricalUsageWeight:   &historicalUsageWeight,
				DecayHalfLife:           &decayHalfLife,
				WindowDuration:          &windowDuration,
				WindowType:              &windowType,
				TumblingWindowStartTime: &tumblingStart,
			}

			request := api.NodepoolCreateFields{
				Name:       testName + "-tbf-full",
				LabelKey:   testLabelKey + "-tbf-full",
				LabelValue: testLabelValue + "-tbf-full",
				ClusterId:  testClusterId,
				SchedulingConfiguration: &api.SchedulingConfiguration{
					PlacementStrategy: api.SchedulingConfigurationPlacementStrategy{
						Cpu: ptr.To(api.PLACEMENTSTRATEGY_SPREAD),
						Gpu: ptr.To(api.PLACEMENTSTRATEGY_BINPACK),
					},
					TimeBasedFairShare: &nodePoolTimeBasedFairShare,
				},
			}

			resp, err := nodepoolService.CreateNodepool(ctx, request)
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusCreated))

			nodepool := resp.Body.(api.Nodepool)
			createdNodepools = append(createdNodepools, nodepool)

			// Validate response from Create
			validateTimeBasedFairShareConfig(nodepool, nodePoolTimeBasedFairShare)

			// Validate response from GetNodepool
			getResp, err := nodepoolService.GetNodepool(ctx, nodepool.Id)
			Expect(err).To(BeNil())
			Expect(getResp.Code).To(Equal(http.StatusOK))
			fromGet := getResp.Body.(api.Nodepool)
			validateTimeBasedFairShareConfig(fromGet, nodePoolTimeBasedFairShare)

			// Validate response from GetNodepools (filtering by clusterId and name)
			listResp, err := nodepoolService.GetNodepools(ctx,
				[]string{
					fmt.Sprintf("clusterId==%s", nodepool.ClusterId),
					fmt.Sprintf("name==%s", nodepool.Name),
				},
				"", "", 0, 0)
			Expect(err).To(BeNil())
			Expect(listResp.Code).To(Equal(http.StatusOK))
			list := listResp.Body.(api.GetNodepools200Response)
			Expect(list.Nodepools).To(HaveLen(1))
			validateTimeBasedFairShareConfig(list.Nodepools[0], nodePoolTimeBasedFairShare)
		})

		It("Update nodepool TimeBasedFairShare and apply defaults for missing fields", func() {
			// Create without TimeBasedFairShare
			createReq := api.NodepoolCreateFields{
				Name:       testName + "-tbf-update",
				LabelKey:   testLabelKey + "-tbf-update",
				LabelValue: testLabelValue + "-tbf-update",
				ClusterId:  testClusterId,
				SchedulingConfiguration: &api.SchedulingConfiguration{
					PlacementStrategy: api.SchedulingConfigurationPlacementStrategy{
						Cpu: ptr.To(api.PLACEMENTSTRATEGY_SPREAD),
						Gpu: ptr.To(api.PLACEMENTSTRATEGY_BINPACK),
					},
				},
			}
			resp, err := nodepoolService.CreateNodepool(ctx, createReq)
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusCreated))
			nodepool := resp.Body.(api.Nodepool)
			createdNodepools = append(createdNodepools, nodepool)

			// Update with Enabled + HistoricalUsageWeight only; others should be set to defaults.
			newHistoricalUsage := float32(4.2)
			updateConfig := &api.SchedulingConfiguration{
				TimeBasedFairShare: &api.TimeBasedFairShare{
					Enabled:               ptr.To(true),
					HistoricalUsageWeight: &newHistoricalUsage,
				},
			}

			updateFields := nodepoolUpdateFields("-tbf-update", nil)
			updateFields.SchedulingConfiguration = updateConfig

			updateResp, err := nodepoolService.UpdateNodepool(ctx, nodepool.Id, *updateFields)
			Expect(err).To(BeNil())
			Expect(updateResp.Code).To(Equal(http.StatusOK))
			updated := updateResp.Body.(api.Nodepool)

			expected := api.TimeBasedFairShare{
				Enabled:               ptr.To(true),
				HistoricalUsageWeight: &newHistoricalUsage,
				DecayHalfLife:         ptr.To(services.DefaultTimeBasedFairShareDecayHalfLife),
				WindowDuration:        ptr.To(services.DefaultTimeBasedFairShareWindowDuration),
				WindowType:            ptr.To(services.DefaultTimeBasedFairShareWindowType),
			}

			// Update result
			validateTimeBasedFairShareConfig(updated, expected)

			// GetNodepool
			getResp, err := nodepoolService.GetNodepool(ctx, updated.Id)
			Expect(err).To(BeNil())
			Expect(getResp.Code).To(Equal(http.StatusOK))
			fromGet := getResp.Body.(api.Nodepool)
			validateTimeBasedFairShareConfig(fromGet, expected)

			// GetNodepools (filter by clusterId + name)
			listResp, err := nodepoolService.GetNodepools(ctx,
				[]string{
					fmt.Sprintf("clusterId==%s", updated.ClusterId),
					fmt.Sprintf("name==%s", updated.Name),
				},
				"", "", 0, 0)
			Expect(err).To(BeNil())
			Expect(listResp.Code).To(Equal(http.StatusOK))
			list := listResp.Body.(api.GetNodepools200Response)
			Expect(list.Nodepools).To(HaveLen(1))
			validateTimeBasedFairShareConfig(list.Nodepools[0], expected)
		})

		It("Patch nodepool TimeBasedFairShare and keep unspecified fields unchanged", func() {
			// Create with full non-default TBF config
			baseHistoricalUsage := float32(1.7)
			baseDecayHalfLife := "12h"
			baseWindowDuration := "3d"
			baseWindowType := api.TIMEBASEDFAIRSHAREWINDOWTYPE_TUMBLING
			baseTumblingStart := time.Date(2025, 1, 8, 8, 0, 0, 0, time.UTC)

			createReq := api.NodepoolCreateFields{
				Name:       testName + "-tbf-patch",
				LabelKey:   testLabelKey + "-tbf-patch",
				LabelValue: testLabelValue + "-tbf-patch",
				ClusterId:  testClusterId,
				SchedulingConfiguration: &api.SchedulingConfiguration{
					PlacementStrategy: api.SchedulingConfigurationPlacementStrategy{
						Cpu: ptr.To(api.PLACEMENTSTRATEGY_SPREAD),
						Gpu: ptr.To(api.PLACEMENTSTRATEGY_BINPACK),
					},
					TimeBasedFairShare: &api.TimeBasedFairShare{
						Enabled:                 ptr.To(true),
						HistoricalUsageWeight:   &baseHistoricalUsage,
						DecayHalfLife:           &baseDecayHalfLife,
						WindowDuration:          &baseWindowDuration,
						WindowType:              &baseWindowType,
						TumblingWindowStartTime: &baseTumblingStart,
					},
				},
			}
			resp, err := nodepoolService.CreateNodepool(ctx, createReq)
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusCreated))
			nodepool := resp.Body.(api.Nodepool)
			createdNodepools = append(createdNodepools, nodepool)

			// Patch only DecayHalfLife
			newDecayHalfLife := "36h"
			patchFields := api.NodepoolUpdateFields{
				SchedulingConfiguration: &api.SchedulingConfiguration{
					TimeBasedFairShare: &api.TimeBasedFairShare{
						DecayHalfLife: &newDecayHalfLife,
					},
				},
			}

			patchResp, err := nodepoolService.PatchNodepool(ctx, nodepool.Id, patchFields)
			Expect(err).To(BeNil())
			Expect(patchResp.Code).To(Equal(http.StatusOK))
			patched := patchResp.Body.(api.Nodepool)

			expected := api.TimeBasedFairShare{
				Enabled:                 ptr.To(true),
				HistoricalUsageWeight:   &baseHistoricalUsage,
				DecayHalfLife:           &newDecayHalfLife,
				WindowDuration:          &baseWindowDuration,
				WindowType:              &baseWindowType,
				TumblingWindowStartTime: &baseTumblingStart,
			}

			// Patch result
			validateTimeBasedFairShareConfig(patched, expected)

			// GetNodepool
			getResp, err := nodepoolService.GetNodepool(ctx, patched.Id)
			Expect(err).To(BeNil())
			Expect(getResp.Code).To(Equal(http.StatusOK))
			fromGet := getResp.Body.(api.Nodepool)
			validateTimeBasedFairShareConfig(fromGet, expected)

			// GetNodepools (filter by clusterId + name)
			listResp, err := nodepoolService.GetNodepools(ctx,
				[]string{
					fmt.Sprintf("clusterId==%s", patched.ClusterId),
					fmt.Sprintf("name==%s", patched.Name),
				},
				"", "", 0, 0)
			Expect(err).To(BeNil())
			Expect(listResp.Code).To(Equal(http.StatusOK))
			list := listResp.Body.(api.GetNodepools200Response)
			Expect(list.Nodepools).To(HaveLen(1))
			validateTimeBasedFairShareConfig(list.Nodepools[0], expected)
		})

		It("Create nodepool with tumbling TimeBasedFairShare and missing tumblingWindowStartTime should fail", func() {
			createReq := api.NodepoolCreateFields{
				Name:       testName + "-tbf-create-tumbling-missing-start",
				LabelKey:   testLabelKey + "-tbf-create-tumbling-missing-start",
				LabelValue: testLabelValue + "-tbf-create-tumbling-missing-start",
				ClusterId:  testClusterId,
				SchedulingConfiguration: &api.SchedulingConfiguration{
					PlacementStrategy: api.SchedulingConfigurationPlacementStrategy{
						Cpu: ptr.To(api.PLACEMENTSTRATEGY_SPREAD),
						Gpu: ptr.To(api.PLACEMENTSTRATEGY_BINPACK),
					},
					TimeBasedFairShare: &api.TimeBasedFairShare{
						Enabled:    ptr.To(true),
						WindowType: ptr.To(api.TIMEBASEDFAIRSHAREWINDOWTYPE_TUMBLING),
						// tumblingWindowStartTime intentionally omitted
					},
				},
			}

			resp, err := nodepoolService.CreateNodepool(ctx, createReq)
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusBadRequest))
			Expect(resp.Body).To(ContainSubstring("tumblingWindowStartTime must be set when windowType is tumbling"))
		})

		It("Update nodepool to tumbling TimeBasedFairShare without tumblingWindowStartTime should fail", func() {
			// Create with sliding TimeBasedFairShare (valid, no start required)
			createReq := api.NodepoolCreateFields{
				Name:       testName + "-tbf-update-tumbling-missing-start",
				LabelKey:   testLabelKey + "-tbf-update-tumbling-missing-start",
				LabelValue: testLabelValue + "-tbf-update-tumbling-missing-start",
				ClusterId:  testClusterId,
				SchedulingConfiguration: &api.SchedulingConfiguration{
					PlacementStrategy: api.SchedulingConfigurationPlacementStrategy{
						Cpu: ptr.To(api.PLACEMENTSTRATEGY_SPREAD),
						Gpu: ptr.To(api.PLACEMENTSTRATEGY_BINPACK),
					},
					TimeBasedFairShare: &api.TimeBasedFairShare{
						Enabled:    ptr.To(true),
						WindowType: ptr.To(api.TIMEBASEDFAIRSHAREWINDOWTYPE_SLIDING),
					},
				},
			}
			resp, err := nodepoolService.CreateNodepool(ctx, createReq)
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusCreated))
			nodepool := resp.Body.(api.Nodepool)
			createdNodepools = append(createdNodepools, nodepool)

			// Update to tumbling without start
			updateFields := nodepoolUpdateFields("-tbf-update-tumbling-missing-start", nil)
			updateFields.SchedulingConfiguration = &api.SchedulingConfiguration{
				TimeBasedFairShare: &api.TimeBasedFairShare{
					Enabled:    ptr.To(true),
					WindowType: ptr.To(api.TIMEBASEDFAIRSHAREWINDOWTYPE_TUMBLING),
					// no tumblingWindowStartTime
				},
			}

			updateResp, err := nodepoolService.UpdateNodepool(ctx, nodepool.Id, *updateFields)
			Expect(err).To(BeNil())
			Expect(updateResp.Code).To(Equal(http.StatusBadRequest))
			Expect(updateResp.Body).To(ContainSubstring("tumblingWindowStartTime must be set when windowType is tumbling"))
		})

		It("Patch nodepool to tumbling TimeBasedFairShare without tumblingWindowStartTime should fail", func() {
			// Create with sliding TimeBasedFairShare (valid, no start required)
			createReq := api.NodepoolCreateFields{
				Name:       testName + "-tbf-patch-tumbling-missing-start",
				LabelKey:   testLabelKey + "-tbf-patch-tumbling-missing-start",
				LabelValue: testLabelValue + "-tbf-patch-tumbling-missing-start",
				ClusterId:  testClusterId,
				SchedulingConfiguration: &api.SchedulingConfiguration{
					PlacementStrategy: api.SchedulingConfigurationPlacementStrategy{
						Cpu: ptr.To(api.PLACEMENTSTRATEGY_SPREAD),
						Gpu: ptr.To(api.PLACEMENTSTRATEGY_BINPACK),
					},
					TimeBasedFairShare: &api.TimeBasedFairShare{
						Enabled:    ptr.To(true),
						WindowType: ptr.To(api.TIMEBASEDFAIRSHAREWINDOWTYPE_SLIDING),
					},
				},
			}
			resp, err := nodepoolService.CreateNodepool(ctx, createReq)
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusCreated))
			nodepool := resp.Body.(api.Nodepool)
			createdNodepools = append(createdNodepools, nodepool)

			// Patch to tumbling without start
			patchFields := api.NodepoolUpdateFields{
				SchedulingConfiguration: &api.SchedulingConfiguration{
					TimeBasedFairShare: &api.TimeBasedFairShare{
						WindowType: ptr.To(api.TIMEBASEDFAIRSHAREWINDOWTYPE_TUMBLING),
						// no tumblingWindowStartTime
					},
				},
			}

			patchResp, err := nodepoolService.PatchNodepool(ctx, nodepool.Id, patchFields)
			Expect(err).To(BeNil())
			Expect(patchResp.Code).To(Equal(http.StatusBadRequest))
			Expect(patchResp.Body).To(ContainSubstring("tumblingWindowStartTime must be set when windowType is tumbling"))

			// patch successfully with tumblingWindowStartTime
			patchFields.SchedulingConfiguration.TimeBasedFairShare.TumblingWindowStartTime = ptr.To(time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC))
			patchResp, err = nodepoolService.PatchNodepool(ctx, nodepool.Id, patchFields)
			Expect(err).To(BeNil())
			Expect(patchResp.Code).To(Equal(http.StatusOK))

			// patch successfully to switch windowType
			patchFields.SchedulingConfiguration.TimeBasedFairShare = &api.TimeBasedFairShare{
				WindowType: ptr.To(api.TIMEBASEDFAIRSHAREWINDOWTYPE_SLIDING),
			}
			patchResp, err = nodepoolService.PatchNodepool(ctx, nodepool.Id, patchFields)
			Expect(err).To(BeNil())
			Expect(patchResp.Code).To(Equal(http.StatusOK))

			// now patch again with only switch to tumbling, should succeed! cause the time is already set
			patchFields.SchedulingConfiguration.TimeBasedFairShare = &api.TimeBasedFairShare{
				WindowType: ptr.To(api.TIMEBASEDFAIRSHAREWINDOWTYPE_TUMBLING),
			}
			patchResp, err = nodepoolService.PatchNodepool(ctx, nodepool.Id, patchFields)
			Expect(err).To(BeNil())
			Expect(patchResp.Code).To(Equal(http.StatusOK))
		})

		It("Create nodepool with TimeBasedFairShare in old cluster version should fail", func() {
			oldCluster, err := dao.CreateCluster(ctx, &models.Cluster{
				Name:     testName + "-tbf-old-cluster",
				TenantId: tenantId,
				Domain:   "xyz",
				Version:  "v2.23",
			})
			Expect(err).To(BeNil())
			defer dao.DeleteCluster(ctx, oldCluster.ID.String(), tenantId)

			request := api.NodepoolCreateFields{
				Name:       testName + "-tbf-old",
				LabelKey:   testLabelKey + "-tbf-old",
				LabelValue: testLabelValue + "-tbf-old",
				ClusterId:  oldCluster.ID.String(),
				SchedulingConfiguration: &api.SchedulingConfiguration{
					PlacementStrategy: api.SchedulingConfigurationPlacementStrategy{
						Cpu: ptr.To(api.PLACEMENTSTRATEGY_SPREAD),
						Gpu: ptr.To(api.PLACEMENTSTRATEGY_BINPACK),
					},
					TimeBasedFairShare: &api.TimeBasedFairShare{
						Enabled: ptr.To(true),
					},
				},
			}

			resp, apiErr := nodepoolService.CreateNodepool(ctx, request)
			Expect(apiErr).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusBadRequest))
			Expect(resp.Body).To(ContainSubstring("Time-Based Fair Share supported from cluster version 2.24"))
		})

		It("str2duration.ParseDuration - validate api regexp", func() {
			durationRegexp := "^(0|[+]?((\\d+(\\.\\d+)?)(ns|us|µs|μs|ms|s|m|h|d|w))+$)"

			validDurationsToTest := []string{"2w4d", "2d4.5h5m", "0", "2w2w", "+2d4.5h5m6s", "1h30m", "1h30m60s", "5m10m", "3600s", "2w", "5d", "3h500s", "-15m", "15m", "10s", "100s500ms", "-100us", "100us", "50ns"}
			for _, duration := range validDurationsToTest {
				parsedDuration, err := str2duration.ParseDuration(duration)
				Expect(err).To(BeNil(), "Valid duration %s should not return an error", duration)

				matched, err := regexp.MatchString(durationRegexp, duration)
				Expect(err).To(BeNil(), "Error matching duration %s against regexp", duration)
				if parsedDuration >= 0 {
					Expect(matched).To(BeTrue(), "Duration %s should match the regexp", duration)
				} else {
					// Negative durations should not match the regexp
					// They are valid in str2duration but not in the API regexp
					Expect(matched).To(BeFalse(), "Negative duration %s should not match the regexp", duration)
				}
			}

			invalidDurationsToTest := []string{"blabla", "2d-4.5h5m6s", "1h30me60s", "3600mss", "2w-4d", "1h 30m", "5d-3h", "3h500s-100ms", "-15m-10s", "10s-100ms500us", "-100us+50ns"}
			for _, duration := range invalidDurationsToTest {
				_, err := str2duration.ParseDuration(duration)
				Expect(err).ToNot(BeNil(), "Invalid duration %s should return an error", duration)

				matched, err := regexp.MatchString(durationRegexp, duration)
				Expect(err).To(BeNil(), "Error matching duration %s against regexp", duration)
				Expect(matched).To(BeFalse(), "Duration %s should not match the regexp", duration)
			}
		})

		It("Sync status with SchedulingConfiguration and validate", func() {
			// Create nodepool with initial SchedulingConfiguration
			createReq := api.NodepoolCreateFields{
				Name:       testName + "-sched-sync",
				LabelKey:   testLabelKey + "-sched-sync",
				LabelValue: testLabelValue + "-sched-sync",
				ClusterId:  testClusterId,
				SchedulingConfiguration: &api.SchedulingConfiguration{
					PlacementStrategy: api.SchedulingConfigurationPlacementStrategy{
						Cpu: ptr.To(api.PLACEMENTSTRATEGY_SPREAD),
						Gpu: ptr.To(api.PLACEMENTSTRATEGY_BINPACK),
					},
					MinGuaranteedRuntime: ptr.To("2w"),
				},
			}
			response, err := nodepoolService.CreateNodepool(ctx, createReq)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusCreated))
			nodepool := response.Body.(api.Nodepool)
			createdNodepools = append(createdNodepools, nodepool)

			// Sync status with new SchedulingConfiguration values
			newSchedConfig := &api.SchedulingConfiguration{
				PlacementStrategy: api.SchedulingConfigurationPlacementStrategy{
					Cpu: ptr.To(api.PLACEMENTSTRATEGY_SPREAD),
					Gpu: ptr.To(api.PLACEMENTSTRATEGY_BINPACK),
				},
				MinGuaranteedRuntime: ptr.To("2w"),
			}
			syncStatus := api.NodepoolSyncFieldsStatus{
				LabelKey:                ptr.To(nodepool.LabelKey),
				LabelValue:              ptr.To(nodepool.LabelValue),
				SchedulingConfiguration: newSchedConfig,
			}
			syncReq := api.SyncNodepoolsStatusRequest{
				Nodepools: []api.NodepoolSyncFields{
					{
						Name:         nodepool.Name,
						Phase:        api.NODEPOOLPHASE_READY,
						PhaseMessage: "Nodepool is ready",
						Status:       syncStatus,
					},
				},
				DeleteMissing: false,
			}
			syncResp, err := nodepoolService.SyncNodepoolsStatus(ctx, nodepool.ClusterId, syncReq)
			Expect(err).To(BeNil())
			Expect(syncResp.Code).To(Equal(http.StatusNoContent))

			// Fetch nodepool and validate status
			getResp, err := nodepoolService.GetNodepool(ctx, nodepool.Id)
			Expect(err).To(BeNil())
			Expect(getResp.Code).To(Equal(http.StatusOK))
			fetched := getResp.Body.(api.Nodepool)
			Expect(fetched.Status.SchedulingConfiguration).ToNot(BeNil())
			Expect(fetched.Status.SchedulingConfiguration.PlacementStrategy.Cpu).To(Equal(newSchedConfig.PlacementStrategy.Cpu))
			Expect(fetched.Status.SchedulingConfiguration.PlacementStrategy.Gpu).To(Equal(newSchedConfig.PlacementStrategy.Gpu))
			Expect(fetched.Status.SchedulingConfiguration.MinGuaranteedRuntime).To(Equal(newSchedConfig.MinGuaranteedRuntime))
			Expect(fetched.Status.SchedulingConfiguration.PlacementStrategy.Cpu).To(Equal(newSchedConfig.PlacementStrategy.Cpu))
			Expect(fetched.Status.SchedulingConfiguration.PlacementStrategy.Gpu).To(Equal(newSchedConfig.PlacementStrategy.Gpu))

			// validate we still support the deprecated status fields for old clusters
			syncReq.Nodepools[0].Status.SchedulingConfiguration = nil
			syncReq.Nodepools[0].Status.PlacementStrategy = api.NodepoolSyncUpdateFieldsPlacementStrategy{
				Cpu: ptr.To(api.PLACEMENTSTRATEGY_SPREAD),
				Gpu: ptr.To(api.PLACEMENTSTRATEGY_BINPACK),
			}
			syncResp, err = nodepoolService.SyncNodepoolsStatus(ctx, nodepool.ClusterId, syncReq)
			Expect(err).To(BeNil())
			Expect(syncResp.Code).To(Equal(http.StatusNoContent))
			// Fetch nodepool and validate status
			getResp, err = nodepoolService.GetNodepool(ctx, nodepool.Id)
			Expect(err).To(BeNil())
			Expect(getResp.Code).To(Equal(http.StatusOK))
			fetched = getResp.Body.(api.Nodepool)
			Expect(fetched.Status.SchedulingConfiguration).ToNot(BeNil())
			Expect(fetched.Status.SchedulingConfiguration.PlacementStrategy.Cpu).To(Equal(newSchedConfig.PlacementStrategy.Cpu))
			Expect(fetched.Status.SchedulingConfiguration.PlacementStrategy.Gpu).To(Equal(newSchedConfig.PlacementStrategy.Gpu))
			Expect(fetched.Status.SchedulingConfiguration.MinGuaranteedRuntime).To(BeNil())
			Expect(fetched.Status.SchedulingConfiguration.PlacementStrategy.Cpu).To(Equal(newSchedConfig.PlacementStrategy.Cpu))
			Expect(fetched.Status.SchedulingConfiguration.PlacementStrategy.Gpu).To(Equal(newSchedConfig.PlacementStrategy.Gpu))
		})

		It("Create nodepool with MinGuaranteedRuntime in old cluster - should fail", func() {
			// Create an old cluster (version < 2.23)
			oldCluster, err := dao.CreateCluster(ctx, &models.Cluster{
				Name:     testName + "-old-gmr-cluster",
				TenantId: tenantId,
				Domain:   "xyz",
				Version:  "v2.22",
			})
			Expect(err).To(BeNil())
			defer dao.DeleteCluster(ctx, oldCluster.ID.String(), tenantId)

			// Try to create nodepool with MinGuaranteedRuntime
			createReq := api.NodepoolCreateFields{
				Name:       testName + "-old-gmr",
				LabelKey:   testLabelKey + "-old-gmr",
				LabelValue: testLabelValue + "-old-gmr",
				ClusterId:  oldCluster.ID.String(),
				SchedulingConfiguration: &api.SchedulingConfiguration{
					PlacementStrategy: api.SchedulingConfigurationPlacementStrategy{
						Cpu: ptr.To(api.PLACEMENTSTRATEGY_SPREAD),
						Gpu: ptr.To(api.PLACEMENTSTRATEGY_BINPACK),
					},
					MinGuaranteedRuntime: ptr.To("2h"),
				},
			}
			response, apiErr := nodepoolService.CreateNodepool(ctx, createReq)
			Expect(apiErr).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusBadRequest))
			Expect(response.Body).To(ContainSubstring("MinGuaranteedRuntime supported from cluster version 2.23"))
		})

		It("Create nodepool with numaAwareScheduling on cluster older than 2.26 should fail", func() {
			oldCluster, err := dao.CreateCluster(ctx, &models.Cluster{
				Name:     testName + "-numa-old-cluster",
				TenantId: tenantId,
				Domain:   "xyz",
				Version:  "v2.25",
			})
			Expect(err).To(BeNil())
			defer dao.DeleteCluster(ctx, oldCluster.ID.String(), tenantId)

			createReq := api.NodepoolCreateFields{
				Name:       testName + "-numa-old",
				LabelKey:   testLabelKey + "-numa-old",
				LabelValue: testLabelValue + "-numa-old",
				ClusterId:  oldCluster.ID.String(),
				SchedulingConfiguration: &api.SchedulingConfiguration{
					NumaAwareScheduling: ptr.To(true),
				},
			}
			resp, apiErr := nodepoolService.CreateNodepool(ctx, createReq)
			Expect(apiErr).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusBadRequest))
			Expect(resp.Body).To(ContainSubstring("numaAwareScheduling supported from cluster version 2.26"))
		})

		Context("numaAwareScheduling on a >= 2.26 cluster", Ordered, func() {
			var numaClusterId string

			BeforeAll(func() {
				cluster, err := dao.CreateCluster(ctx, &models.Cluster{
					Name:     testName + "-numa-cluster",
					TenantId: tenantId,
					Domain:   "xyz",
					Version:  "v2.26",
				})
				Expect(err).To(BeNil())
				numaClusterId = cluster.ID.String()
			})

			AfterAll(func() {
				dao.DeleteCluster(ctx, numaClusterId, tenantId)
			})

			It("Create nodepool with numaAwareScheduling=true round-trips correctly", func() {
				createReq := api.NodepoolCreateFields{
					Name:       testName + "-numa",
					LabelKey:   testLabelKey + "-numa",
					LabelValue: testLabelValue + "-numa",
					ClusterId:  numaClusterId,
					SchedulingConfiguration: &api.SchedulingConfiguration{
						NumaAwareScheduling: ptr.To(true),
					},
				}
				resp, err := nodepoolService.CreateNodepool(ctx, createReq)
				Expect(err).To(BeNil())
				Expect(resp.Code).To(Equal(http.StatusCreated))
				created := resp.Body.(api.Nodepool)
				defer nodepoolService.DeleteNodepool(ctx, created.Id)

				Expect(created.SchedulingConfiguration).ToNot(BeNil())
				Expect(created.SchedulingConfiguration.NumaAwareScheduling).ToNot(BeNil())
				Expect(*created.SchedulingConfiguration.NumaAwareScheduling).To(BeTrue())

				getResp, err := nodepoolService.GetNodepool(ctx, created.Id)
				Expect(err).To(BeNil())
				fetched := getResp.Body.(api.Nodepool)
				Expect(fetched.SchedulingConfiguration).ToNot(BeNil())
				Expect(fetched.SchedulingConfiguration.NumaAwareScheduling).ToNot(BeNil())
				Expect(*fetched.SchedulingConfiguration.NumaAwareScheduling).To(BeTrue())

				listResp, err := nodepoolService.GetNodepools(ctx, []string{fmt.Sprintf("clusterId==%s", numaClusterId)}, "", "", 0, 0)
				Expect(err).To(BeNil())
				Expect(listResp.Code).To(Equal(http.StatusOK))
				list, ok := listResp.Body.(api.GetNodepools200Response)
				Expect(ok).To(BeTrue())
				var found *api.Nodepool
				for i := range list.Nodepools {
					if list.Nodepools[i].Id == created.Id {
						found = &list.Nodepools[i]
						break
					}
				}
				Expect(found).ToNot(BeNil())
				Expect(found.SchedulingConfiguration).ToNot(BeNil())
				Expect(found.SchedulingConfiguration.NumaAwareScheduling).ToNot(BeNil())
				Expect(*found.SchedulingConfiguration.NumaAwareScheduling).To(BeTrue())
			})

			It("Update nodepool with numaAwareScheduling changes value", func() {
				createReq := api.NodepoolCreateFields{
					Name:       testName + "-numa-update",
					LabelKey:   testLabelKey + "-numa-update",
					LabelValue: testLabelValue + "-numa-update",
					ClusterId:  numaClusterId,
				}
				createResp, err := nodepoolService.CreateNodepool(ctx, createReq)
				Expect(err).To(BeNil())
				Expect(createResp.Code).To(Equal(http.StatusCreated))
				nodepool := createResp.Body.(api.Nodepool)
				defer nodepoolService.DeleteNodepool(ctx, nodepool.Id)

				updateFields := api.NodepoolUpdateFields{
					LabelKey:   ptr.To(testLabelKey + "-numa-update"),
					LabelValue: ptr.To(testLabelValue + "-numa-update"),
					SchedulingConfiguration: &api.SchedulingConfiguration{
						NumaAwareScheduling: ptr.To(true),
					},
				}
				updateResp, err := nodepoolService.UpdateNodepool(ctx, nodepool.Id, updateFields)
				Expect(err).To(BeNil())
				Expect(updateResp.Code).To(Equal(http.StatusOK))
				updated := updateResp.Body.(api.Nodepool)
				Expect(updated.SchedulingConfiguration).ToNot(BeNil())
				Expect(updated.SchedulingConfiguration.NumaAwareScheduling).ToNot(BeNil())
				Expect(*updated.SchedulingConfiguration.NumaAwareScheduling).To(BeTrue())
			})

			It("Patch nodepool with numaAwareScheduling non-nil changes value", func() {
				createReq := api.NodepoolCreateFields{
					Name:       testName + "-numa-patch-change",
					LabelKey:   testLabelKey + "-numa-patch-change",
					LabelValue: testLabelValue + "-numa-patch-change",
					ClusterId:  numaClusterId,
					SchedulingConfiguration: &api.SchedulingConfiguration{
						NumaAwareScheduling: ptr.To(false),
					},
				}
				createResp, err := nodepoolService.CreateNodepool(ctx, createReq)
				Expect(err).To(BeNil())
				Expect(createResp.Code).To(Equal(http.StatusCreated))
				nodepool := createResp.Body.(api.Nodepool)
				defer nodepoolService.DeleteNodepool(ctx, nodepool.Id)

				patchFields := api.NodepoolUpdateFields{
					SchedulingConfiguration: &api.SchedulingConfiguration{
						NumaAwareScheduling: ptr.To(true),
					},
				}
				patchResp, err := nodepoolService.PatchNodepool(ctx, nodepool.Id, patchFields)
				Expect(err).To(BeNil())
				Expect(patchResp.Code).To(Equal(http.StatusOK))
				patched := patchResp.Body.(api.Nodepool)
				Expect(patched.SchedulingConfiguration).ToNot(BeNil())
				Expect(patched.SchedulingConfiguration.NumaAwareScheduling).ToNot(BeNil())
				Expect(*patched.SchedulingConfiguration.NumaAwareScheduling).To(BeTrue())
			})

			It("Patch nodepool: numaAwareScheduling nil leaves existing value unchanged", func() {
				createReq := api.NodepoolCreateFields{
					Name:       testName + "-numa-patch",
					LabelKey:   testLabelKey + "-numa-patch",
					LabelValue: testLabelValue + "-numa-patch",
					ClusterId:  numaClusterId,
					SchedulingConfiguration: &api.SchedulingConfiguration{
						NumaAwareScheduling: ptr.To(true),
					},
				}
				createResp, err := nodepoolService.CreateNodepool(ctx, createReq)
				Expect(err).To(BeNil())
				Expect(createResp.Code).To(Equal(http.StatusCreated))
				nodepool := createResp.Body.(api.Nodepool)
				defer nodepoolService.DeleteNodepool(ctx, nodepool.Id)

				patchFields := api.NodepoolUpdateFields{
					SchedulingConfiguration: &api.SchedulingConfiguration{
						NumaAwareScheduling: nil,
					},
				}
				patchResp, err := nodepoolService.PatchNodepool(ctx, nodepool.Id, patchFields)
				Expect(err).To(BeNil())
				Expect(patchResp.Code).To(Equal(http.StatusOK))

				getResp, err := nodepoolService.GetNodepool(ctx, nodepool.Id)
				Expect(err).To(BeNil())
				fetched := getResp.Body.(api.Nodepool)
				Expect(fetched.SchedulingConfiguration).ToNot(BeNil())
				Expect(fetched.SchedulingConfiguration.NumaAwareScheduling).ToNot(BeNil())
				Expect(*fetched.SchedulingConfiguration.NumaAwareScheduling).To(BeTrue())
			})
		})

		It("Create nodepool with invalid PlacementStrategy should fail", func() {
			invalidStrategy := api.PlacementStrategy("invalid-strategy")
			createReq := api.NodepoolCreateFields{
				Name:       testName + "-invalid-deprecated",
				LabelKey:   testLabelKey + "-invalid-deprecated",
				LabelValue: testLabelValue + "-invalid-deprecated",
				ClusterId:  testClusterId,
				SchedulingConfiguration: &api.SchedulingConfiguration{
					PlacementStrategy: api.SchedulingConfigurationPlacementStrategy{
						Cpu: &invalidStrategy,
					},
				},
			}
			resp, err := nodepoolService.CreateNodepool(ctx, createReq)
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusBadRequest))
			Expect(resp.Body).To(ContainSubstring("invalid placementStrategy.cpu"))

			createReq.SchedulingConfiguration.PlacementStrategy.Cpu = nil
			createReq.SchedulingConfiguration = &api.SchedulingConfiguration{
				PlacementStrategy: api.SchedulingConfigurationPlacementStrategy{
					Gpu: &invalidStrategy,
				},
			}
			resp, err = nodepoolService.CreateNodepool(ctx, createReq)
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusBadRequest))
			Expect(resp.Body).To(ContainSubstring("invalid placementStrategy.gpu"))
		})

		It("Update nodepool with invalid PlacementStrategy should fail", func() {
			invalidStrategy := api.PlacementStrategy("invalid-strategy")
			// Create a valid nodepool first
			validCreateReq := api.NodepoolCreateFields{
				Name:       testName + "-valid-for-update",
				LabelKey:   testLabelKey + "-valid-for-update",
				LabelValue: testLabelValue + "-valid-for-update",
				ClusterId:  testClusterId,
				SchedulingConfiguration: &api.SchedulingConfiguration{
					PlacementStrategy: api.SchedulingConfigurationPlacementStrategy{
						Cpu: ptr.To(api.PLACEMENTSTRATEGY_SPREAD),
						Gpu: ptr.To(api.PLACEMENTSTRATEGY_BINPACK),
					},
				},
			}
			resp, err := nodepoolService.CreateNodepool(ctx, validCreateReq)
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusCreated))
			nodepool := resp.Body.(api.Nodepool)
			createdNodepools = append(createdNodepools, nodepool)

			updateFields := api.NodepoolUpdateFields{
				LabelKey:   ptr.To(testLabelKey + "-valid-for-update"),
				LabelValue: ptr.To(testLabelValue + "-valid-for-update"),
				SchedulingConfiguration: &api.SchedulingConfiguration{
					PlacementStrategy: api.SchedulingConfigurationPlacementStrategy{
						Gpu: &invalidStrategy,
					},
				},
			}
			resp, err = nodepoolService.UpdateNodepool(ctx, nodepool.Id, updateFields)
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusBadRequest))
			Expect(resp.Body).To(ContainSubstring("invalid placementStrategy.gpu"))

			updateFields.SchedulingConfiguration = &api.SchedulingConfiguration{
				PlacementStrategy: api.SchedulingConfigurationPlacementStrategy{
					Cpu: &invalidStrategy,
				},
			}
			resp, err = nodepoolService.UpdateNodepool(ctx, nodepool.Id, updateFields)
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusBadRequest))
			Expect(resp.Body).To(ContainSubstring("invalid placementStrategy.cpu"))
		})

		It("Patch nodepool with invalid PlacementStrategy should fail", func() {
			invalidStrategy := api.PlacementStrategy("invalid-strategy")
			// Create a valid nodepool first
			validCreateReq := api.NodepoolCreateFields{
				Name:       testName + "-valid-for-patch",
				LabelKey:   testLabelKey + "-valid-for-patch",
				LabelValue: testLabelValue + "-valid-for-patch",
				ClusterId:  testClusterId,
				SchedulingConfiguration: &api.SchedulingConfiguration{
					PlacementStrategy: api.SchedulingConfigurationPlacementStrategy{
						Cpu: ptr.To(api.PLACEMENTSTRATEGY_SPREAD),
						Gpu: ptr.To(api.PLACEMENTSTRATEGY_BINPACK),
					},
				},
			}
			resp, err := nodepoolService.CreateNodepool(ctx, validCreateReq)
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusCreated))
			nodepool := resp.Body.(api.Nodepool)
			createdNodepools = append(createdNodepools, nodepool)

			patchFields := api.NodepoolUpdateFields{
				SchedulingConfiguration: &api.SchedulingConfiguration{
					PlacementStrategy: api.SchedulingConfigurationPlacementStrategy{
						Gpu: &invalidStrategy,
					},
				},
			}
			resp, err = nodepoolService.PatchNodepool(ctx, nodepool.Id, patchFields)
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusBadRequest))
			Expect(resp.Body).To(ContainSubstring("invalid placementStrategy.gpu"))

			patchFields.SchedulingConfiguration = &api.SchedulingConfiguration{
				PlacementStrategy: api.SchedulingConfigurationPlacementStrategy{
					Cpu: &invalidStrategy,
				},
			}
			resp, err = nodepoolService.PatchNodepool(ctx, nodepool.Id, patchFields)
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusBadRequest))
			Expect(resp.Body).To(ContainSubstring("invalid placementStrategy.cpu"))
		})
	})

	Context("Device-level GPU placement strategy on old cluster", Ordered, func() {
		// testClusterId is a v2.24 cluster — independent device-level placement
		// is not supported, so device must mirror node (cluster can't honor split).

		baseCreateReq := func(suffix string) api.NodepoolCreateFields {
			return api.NodepoolCreateFields{
				Name:       testName + suffix,
				LabelKey:   testLabelKey + suffix,
				LabelValue: testLabelValue + suffix,
				ClusterId:  testClusterId,
				SchedulingConfiguration: &api.SchedulingConfiguration{
					GpuPodsPlacementStrategy: &api.GpuPodsPlacementStrategy{
						Node: ptr.To(api.PLACEMENTSTRATEGY_BINPACK),
					},
				},
			}
		}

		It("create with deprecated gpu populates new node; device omitted on < 2.26", func() {
			req := baseCreateReq("-old-create-deprecated")
			req.SchedulingConfiguration = &api.SchedulingConfiguration{
				PlacementStrategy: api.SchedulingConfigurationPlacementStrategy{
					Gpu: ptr.To(api.PLACEMENTSTRATEGY_SPREAD),
				},
			}
			resp, err := nodepoolService.CreateNodepool(ctx, req)
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusCreated))
			np := resp.Body.(api.Nodepool)
			createdNodepools = append(createdNodepools, np)
			Expect(np.SchedulingConfiguration.GpuPodsPlacementStrategy.Node).To(Equal(ptr.To(api.PLACEMENTSTRATEGY_SPREAD)))
			Expect(np.SchedulingConfiguration.GpuPodsPlacementStrategy.Device).To(BeNil())
			// Deprecated mirror echoes node-level.
			Expect(np.SchedulingConfiguration.PlacementStrategy.Gpu).To(Equal(ptr.To(api.PLACEMENTSTRATEGY_SPREAD)))
		})

		It("create rejects device set without node on < 2.26", func() {
			req := baseCreateReq("-old-create-device-only")
			req.SchedulingConfiguration = &api.SchedulingConfiguration{
				GpuPodsPlacementStrategy: &api.GpuPodsPlacementStrategy{
					Device: ptr.To(api.PLACEMENTSTRATEGY_SPREAD),
				},
			}
			resp, err := nodepoolService.CreateNodepool(ctx, req)
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusBadRequest))
			Expect(resp.Body).To(ContainSubstring("gpuPodsPlacementStrategy.device cannot be set"))
		})

		It("create rejects device set diverging from node on < 2.26", func() {
			req := baseCreateReq("-old-create-unequal")
			req.SchedulingConfiguration = &api.SchedulingConfiguration{
				GpuPodsPlacementStrategy: &api.GpuPodsPlacementStrategy{
					Node:   ptr.To(api.PLACEMENTSTRATEGY_BINPACK),
					Device: ptr.To(api.PLACEMENTSTRATEGY_SPREAD),
				},
			}
			resp, err := nodepoolService.CreateNodepool(ctx, req)
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusBadRequest))
			Expect(resp.Body).To(ContainSubstring("gpuPodsPlacementStrategy.device cannot be set"))
		})

		It("create rejects device set even when equal to node on < 2.26", func() {
			req := baseCreateReq("-old-create-equal")
			req.SchedulingConfiguration = &api.SchedulingConfiguration{
				GpuPodsPlacementStrategy: &api.GpuPodsPlacementStrategy{
					Node:   ptr.To(api.PLACEMENTSTRATEGY_BINPACK),
					Device: ptr.To(api.PLACEMENTSTRATEGY_BINPACK),
				},
			}
			resp, err := nodepoolService.CreateNodepool(ctx, req)
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusBadRequest))
			Expect(resp.Body).To(ContainSubstring("gpuPodsPlacementStrategy.device cannot be set"))
		})

		It("patch with only new node succeeds on < 2.26; device omitted in response", func() {
			req := baseCreateReq("-old-patch-node")
			resp, err := nodepoolService.CreateNodepool(ctx, req)
			Expect(err).To(BeNil())
			np := resp.Body.(api.Nodepool)
			createdNodepools = append(createdNodepools, np)

			patchResp, err := nodepoolService.PatchNodepool(ctx, np.Id, api.NodepoolUpdateFields{
				SchedulingConfiguration: &api.SchedulingConfiguration{
					GpuPodsPlacementStrategy: &api.GpuPodsPlacementStrategy{
						Node: ptr.To(api.PLACEMENTSTRATEGY_SPREAD),
					},
				},
			})
			Expect(err).To(BeNil())
			Expect(patchResp.Code).To(Equal(http.StatusOK))
			patched := patchResp.Body.(api.Nodepool)
			Expect(patched.SchedulingConfiguration.GpuPodsPlacementStrategy.Node).To(Equal(ptr.To(api.PLACEMENTSTRATEGY_SPREAD)))
			Expect(patched.SchedulingConfiguration.GpuPodsPlacementStrategy.Device).To(BeNil())
		})

		It("update with deprecated gpu populates new node; device omitted on < 2.26", func() {
			req := baseCreateReq("-old-update-mirror")
			resp, err := nodepoolService.CreateNodepool(ctx, req)
			Expect(err).To(BeNil())
			np := resp.Body.(api.Nodepool)
			createdNodepools = append(createdNodepools, np)

			updateFields := nodepoolUpdateFields("-old-update-mirror", nil)
			updateFields.SchedulingConfiguration = &api.SchedulingConfiguration{
				PlacementStrategy: api.SchedulingConfigurationPlacementStrategy{
					Gpu: ptr.To(api.PLACEMENTSTRATEGY_SPREAD),
				},
			}
			updateResp, err := nodepoolService.UpdateNodepool(ctx, np.Id, *updateFields)
			Expect(err).To(BeNil())
			Expect(updateResp.Code).To(Equal(http.StatusOK))
			updated := updateResp.Body.(api.Nodepool)
			Expect(updated.SchedulingConfiguration.GpuPodsPlacementStrategy.Node).To(Equal(ptr.To(api.PLACEMENTSTRATEGY_SPREAD)))
			Expect(updated.SchedulingConfiguration.GpuPodsPlacementStrategy.Device).To(BeNil())
		})

		It("GET on < 2.26 omits gpuPodsPlacementStrategy.device", func() {
			req := baseCreateReq("-old-get-strips-device")
			resp, err := nodepoolService.CreateNodepool(ctx, req)
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusCreated))
			np := resp.Body.(api.Nodepool)
			createdNodepools = append(createdNodepools, np)

			getResp, err := nodepoolService.GetNodepool(ctx, np.Id)
			Expect(err).To(BeNil())
			Expect(getResp.Code).To(Equal(http.StatusOK))
			got := getResp.Body.(api.Nodepool)
			Expect(got.SchedulingConfiguration.GpuPodsPlacementStrategy.Node).To(Equal(ptr.To(api.PLACEMENTSTRATEGY_BINPACK)))
			Expect(got.SchedulingConfiguration.GpuPodsPlacementStrategy.Device).To(BeNil())
			Expect(got.SchedulingConfiguration.PlacementStrategy.Gpu).To(Equal(ptr.To(api.PLACEMENTSTRATEGY_BINPACK)))
		})

		It("round-trip GET → PUT on < 2.26 passes validation", func() {
			// Regression guard: a client that GETs a < 2.26 nodepool and PUTs the
			// returned schedulingConfiguration verbatim must not be rejected.
			req := baseCreateReq("-old-round-trip")
			resp, err := nodepoolService.CreateNodepool(ctx, req)
			Expect(err).To(BeNil())
			np := resp.Body.(api.Nodepool)
			createdNodepools = append(createdNodepools, np)

			getResp, err := nodepoolService.GetNodepool(ctx, np.Id)
			Expect(err).To(BeNil())
			got := getResp.Body.(api.Nodepool)

			updateFields := nodepoolUpdateFields("-old-round-trip", nil)
			updateFields.SchedulingConfiguration = got.SchedulingConfiguration
			updateResp, err := nodepoolService.UpdateNodepool(ctx, np.Id, *updateFields)
			Expect(err).To(BeNil())
			Expect(updateResp.Code).To(Equal(http.StatusOK))
		})
	})

	Context("Device-level GPU placement strategy on new cluster", Ordered, func() {
		var newClusterId string

		BeforeAll(func() {
			cluster, err := dao.CreateCluster(ctx, &models.Cluster{
				Name:     testName + "-new-cluster-ctx",
				TenantId: tenantId,
				Domain:   "xyz",
				Version:  "v2.26",
			})
			Expect(err).To(BeNil())
			newClusterId = cluster.ID.String()
		})

		AfterAll(func() {
			dao.DeleteCluster(ctx, newClusterId, tenantId)
		})

		baseCreateReq := func(suffix string) api.NodepoolCreateFields {
			return api.NodepoolCreateFields{
				Name:       testName + suffix,
				LabelKey:   testLabelKey + suffix,
				LabelValue: testLabelValue + suffix,
				ClusterId:  newClusterId,
				SchedulingConfiguration: &api.SchedulingConfiguration{
					GpuPodsPlacementStrategy: &api.GpuPodsPlacementStrategy{
						Node: ptr.To(api.PLACEMENTSTRATEGY_BINPACK),
					},
				},
			}
		}

		It("create with divergent node and device stores both independently", func() {
			req := baseCreateReq("-new-create-divergent")
			req.SchedulingConfiguration.GpuPodsPlacementStrategy.Device = ptr.To(api.PLACEMENTSTRATEGY_SPREAD)
			resp, err := nodepoolService.CreateNodepool(ctx, req)
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusCreated))
			np := resp.Body.(api.Nodepool)
			createdNodepools = append(createdNodepools, np)
			Expect(np.SchedulingConfiguration.GpuPodsPlacementStrategy.Node).To(Equal(ptr.To(api.PLACEMENTSTRATEGY_BINPACK)))
			Expect(np.SchedulingConfiguration.GpuPodsPlacementStrategy.Device).To(Equal(ptr.To(api.PLACEMENTSTRATEGY_SPREAD)))
		})

		It("GET on >= 2.26 preserves device", func() {
			// Negative control for the < 2.26 strip — make sure we don't accidentally
			// strip device on supported clusters.
			req := baseCreateReq("-new-get-keeps-device")
			req.SchedulingConfiguration.GpuPodsPlacementStrategy.Device = ptr.To(api.PLACEMENTSTRATEGY_SPREAD)
			resp, err := nodepoolService.CreateNodepool(ctx, req)
			Expect(err).To(BeNil())
			np := resp.Body.(api.Nodepool)
			createdNodepools = append(createdNodepools, np)

			getResp, err := nodepoolService.GetNodepool(ctx, np.Id)
			Expect(err).To(BeNil())
			Expect(getResp.Code).To(Equal(http.StatusOK))
			got := getResp.Body.(api.Nodepool)
			Expect(got.SchedulingConfiguration.GpuPodsPlacementStrategy.Node).To(Equal(ptr.To(api.PLACEMENTSTRATEGY_BINPACK)))
			Expect(got.SchedulingConfiguration.GpuPodsPlacementStrategy.Device).To(Equal(ptr.To(api.PLACEMENTSTRATEGY_SPREAD)))
		})

		It("patch only node leaves device unchanged", func() {
			req := baseCreateReq("-new-patch-only-node")
			req.SchedulingConfiguration.GpuPodsPlacementStrategy.Node = ptr.To(api.PLACEMENTSTRATEGY_BINPACK)
			req.SchedulingConfiguration.GpuPodsPlacementStrategy.Device = ptr.To(api.PLACEMENTSTRATEGY_BINPACK)
			resp, err := nodepoolService.CreateNodepool(ctx, req)
			Expect(err).To(BeNil())
			np := resp.Body.(api.Nodepool)
			createdNodepools = append(createdNodepools, np)

			patchResp, err := nodepoolService.PatchNodepool(ctx, np.Id, api.NodepoolUpdateFields{
				SchedulingConfiguration: &api.SchedulingConfiguration{
					GpuPodsPlacementStrategy: &api.GpuPodsPlacementStrategy{
						Node: ptr.To(api.PLACEMENTSTRATEGY_SPREAD),
					},
				},
			})
			Expect(err).To(BeNil())
			Expect(patchResp.Code).To(Equal(http.StatusOK))
			patched := patchResp.Body.(api.Nodepool)
			Expect(patched.SchedulingConfiguration.GpuPodsPlacementStrategy.Node).To(Equal(ptr.To(api.PLACEMENTSTRATEGY_SPREAD)))
			Expect(patched.SchedulingConfiguration.GpuPodsPlacementStrategy.Device).To(Equal(ptr.To(api.PLACEMENTSTRATEGY_BINPACK)))
		})

		It("patch only device leaves node unchanged; deprecated.gpu echoes node", func() {
			req := baseCreateReq("-new-patch-only-device")
			resp, err := nodepoolService.CreateNodepool(ctx, req)
			Expect(err).To(BeNil())
			np := resp.Body.(api.Nodepool)
			createdNodepools = append(createdNodepools, np)

			patchResp, err := nodepoolService.PatchNodepool(ctx, np.Id, api.NodepoolUpdateFields{
				SchedulingConfiguration: &api.SchedulingConfiguration{
					GpuPodsPlacementStrategy: &api.GpuPodsPlacementStrategy{
						Device: ptr.To(api.PLACEMENTSTRATEGY_SPREAD),
					},
				},
			})
			Expect(err).To(BeNil())
			Expect(patchResp.Code).To(Equal(http.StatusOK))
			patched := patchResp.Body.(api.Nodepool)
			Expect(patched.SchedulingConfiguration.GpuPodsPlacementStrategy.Node).To(Equal(ptr.To(api.PLACEMENTSTRATEGY_BINPACK)))
			Expect(patched.SchedulingConfiguration.GpuPodsPlacementStrategy.Device).To(Equal(ptr.To(api.PLACEMENTSTRATEGY_SPREAD)))
			// Deprecated legacy gpu echoes node-level only.
			Expect(patched.SchedulingConfiguration.PlacementStrategy.Gpu).To(Equal(ptr.To(api.PLACEMENTSTRATEGY_BINPACK)))
		})

		It("update with divergent node and device stores both independently", func() {
			req := baseCreateReq("-new-update-divergent")
			resp, err := nodepoolService.CreateNodepool(ctx, req)
			Expect(err).To(BeNil())
			np := resp.Body.(api.Nodepool)
			createdNodepools = append(createdNodepools, np)

			updateFields := nodepoolUpdateFields("-new-update-divergent", nil)
			updateFields.SchedulingConfiguration = &api.SchedulingConfiguration{
				GpuPodsPlacementStrategy: &api.GpuPodsPlacementStrategy{
					Node:   ptr.To(api.PLACEMENTSTRATEGY_BINPACK),
					Device: ptr.To(api.PLACEMENTSTRATEGY_SPREAD),
				},
			}
			updateResp, err := nodepoolService.UpdateNodepool(ctx, np.Id, *updateFields)
			Expect(err).To(BeNil())
			Expect(updateResp.Code).To(Equal(http.StatusOK))
			updated := updateResp.Body.(api.Nodepool)
			Expect(updated.SchedulingConfiguration.GpuPodsPlacementStrategy.Node).To(Equal(ptr.To(api.PLACEMENTSTRATEGY_BINPACK)))
			Expect(updated.SchedulingConfiguration.GpuPodsPlacementStrategy.Device).To(Equal(ptr.To(api.PLACEMENTSTRATEGY_SPREAD)))
		})

		It("new wins over deprecated when both are sent", func() {
			req := baseCreateReq("-new-precedence")
			resp, err := nodepoolService.CreateNodepool(ctx, req)
			Expect(err).To(BeNil())
			np := resp.Body.(api.Nodepool)
			createdNodepools = append(createdNodepools, np)

			updateFields := nodepoolUpdateFields("-new-precedence", nil)
			updateFields.SchedulingConfiguration = &api.SchedulingConfiguration{
				PlacementStrategy: api.SchedulingConfigurationPlacementStrategy{
					Gpu: ptr.To(api.PLACEMENTSTRATEGY_BINPACK),
				},
				GpuPodsPlacementStrategy: &api.GpuPodsPlacementStrategy{
					Node:   ptr.To(api.PLACEMENTSTRATEGY_SPREAD),
					Device: ptr.To(api.PLACEMENTSTRATEGY_BINPACK),
				},
			}
			updateResp, err := nodepoolService.UpdateNodepool(ctx, np.Id, *updateFields)
			Expect(err).To(BeNil())
			Expect(updateResp.Code).To(Equal(http.StatusOK))
			updated := updateResp.Body.(api.Nodepool)
			// New fields win; deprecated gpu is ignored.
			Expect(updated.SchedulingConfiguration.GpuPodsPlacementStrategy.Node).To(Equal(ptr.To(api.PLACEMENTSTRATEGY_SPREAD)))
			Expect(updated.SchedulingConfiguration.GpuPodsPlacementStrategy.Device).To(Equal(ptr.To(api.PLACEMENTSTRATEGY_BINPACK)))
		})

		It("create with deprecated.gpu/cpu and full new axes — new wins for all", func() {
			req := baseCreateReq("-new-create-all-fields")
			req.SchedulingConfiguration = &api.SchedulingConfiguration{
				PlacementStrategy: api.SchedulingConfigurationPlacementStrategy{
					Gpu: ptr.To(api.PLACEMENTSTRATEGY_SPREAD),
					Cpu: ptr.To(api.PLACEMENTSTRATEGY_SPREAD),
				},
				GpuPodsPlacementStrategy: &api.GpuPodsPlacementStrategy{
					Node:   ptr.To(api.PLACEMENTSTRATEGY_BINPACK),
					Device: ptr.To(api.PLACEMENTSTRATEGY_BINPACK),
				},
				CpuPodsPlacementStrategy: &api.CpuPodsPlacementStrategy{
					Node: ptr.To(api.PLACEMENTSTRATEGY_BINPACK),
				},
			}
			resp, err := nodepoolService.CreateNodepool(ctx, req)
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusCreated))
			np := resp.Body.(api.Nodepool)
			createdNodepools = append(createdNodepools, np)
			Expect(np.SchedulingConfiguration.GpuPodsPlacementStrategy.Node).To(Equal(ptr.To(api.PLACEMENTSTRATEGY_BINPACK)))
			Expect(np.SchedulingConfiguration.GpuPodsPlacementStrategy.Device).To(Equal(ptr.To(api.PLACEMENTSTRATEGY_BINPACK)))
			Expect(np.SchedulingConfiguration.CpuPodsPlacementStrategy.Node).To(Equal(ptr.To(api.PLACEMENTSTRATEGY_BINPACK)))
			// Deprecated mirror echoes node-level.
			Expect(np.SchedulingConfiguration.PlacementStrategy.Gpu).To(Equal(ptr.To(api.PLACEMENTSTRATEGY_BINPACK)))
			Expect(np.SchedulingConfiguration.PlacementStrategy.Cpu).To(Equal(ptr.To(api.PLACEMENTSTRATEGY_BINPACK)))
		})

		It("create with deprecated.gpu/cpu and only new.device — deprecated mirrors into node, new.device wins device", func() {
			req := baseCreateReq("-new-create-mixed-device-only")
			req.SchedulingConfiguration = &api.SchedulingConfiguration{
				PlacementStrategy: api.SchedulingConfigurationPlacementStrategy{
					Gpu: ptr.To(api.PLACEMENTSTRATEGY_SPREAD),
					Cpu: ptr.To(api.PLACEMENTSTRATEGY_SPREAD),
				},
				GpuPodsPlacementStrategy: &api.GpuPodsPlacementStrategy{
					Device: ptr.To(api.PLACEMENTSTRATEGY_BINPACK),
				},
			}
			resp, err := nodepoolService.CreateNodepool(ctx, req)
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusCreated))
			np := resp.Body.(api.Nodepool)
			createdNodepools = append(createdNodepools, np)
			// Per resolveRequestedPlacement: deprecated.gpu mirrors to both, then new.device overrides device.
			// Node keeps the deprecated value because new.node isn't provided.
			Expect(np.SchedulingConfiguration.GpuPodsPlacementStrategy.Node).To(Equal(ptr.To(api.PLACEMENTSTRATEGY_SPREAD)))
			Expect(np.SchedulingConfiguration.GpuPodsPlacementStrategy.Device).To(Equal(ptr.To(api.PLACEMENTSTRATEGY_BINPACK)))
			// CPU only had the deprecated value; new.cpuPodsPlacementStrategy isn't sent.
			Expect(np.SchedulingConfiguration.CpuPodsPlacementStrategy.Node).To(Equal(ptr.To(api.PLACEMENTSTRATEGY_SPREAD)))
			Expect(np.SchedulingConfiguration.PlacementStrategy.Gpu).To(Equal(ptr.To(api.PLACEMENTSTRATEGY_SPREAD)))
			Expect(np.SchedulingConfiguration.PlacementStrategy.Cpu).To(Equal(ptr.To(api.PLACEMENTSTRATEGY_SPREAD)))
		})

		It("update with deprecated.gpu and only new.device — same divergent outcome as create", func() {
			req := baseCreateReq("-new-update-mixed-device-only")
			resp, err := nodepoolService.CreateNodepool(ctx, req)
			Expect(err).To(BeNil())
			np := resp.Body.(api.Nodepool)
			createdNodepools = append(createdNodepools, np)

			updateFields := nodepoolUpdateFields("-new-update-mixed-device-only", nil)
			updateFields.SchedulingConfiguration = &api.SchedulingConfiguration{
				PlacementStrategy: api.SchedulingConfigurationPlacementStrategy{
					Gpu: ptr.To(api.PLACEMENTSTRATEGY_SPREAD),
				},
				GpuPodsPlacementStrategy: &api.GpuPodsPlacementStrategy{
					Device: ptr.To(api.PLACEMENTSTRATEGY_BINPACK),
				},
			}
			updateResp, err := nodepoolService.UpdateNodepool(ctx, np.Id, *updateFields)
			Expect(err).To(BeNil())
			Expect(updateResp.Code).To(Equal(http.StatusOK))
			updated := updateResp.Body.(api.Nodepool)
			Expect(updated.SchedulingConfiguration.GpuPodsPlacementStrategy.Node).To(Equal(ptr.To(api.PLACEMENTSTRATEGY_SPREAD)))
			Expect(updated.SchedulingConfiguration.GpuPodsPlacementStrategy.Device).To(Equal(ptr.To(api.PLACEMENTSTRATEGY_BINPACK)))
			Expect(updated.SchedulingConfiguration.PlacementStrategy.Gpu).To(Equal(ptr.To(api.PLACEMENTSTRATEGY_SPREAD)))
		})

		It("patch with deprecated.gpu and new.node — deprecated mirrors device, new.node confirms node", func() {
			req := baseCreateReq("-new-patch-deprecated-and-node")
			resp, err := nodepoolService.CreateNodepool(ctx, req)
			Expect(err).To(BeNil())
			np := resp.Body.(api.Nodepool)
			createdNodepools = append(createdNodepools, np)

			patchResp, err := nodepoolService.PatchNodepool(ctx, np.Id, api.NodepoolUpdateFields{
				SchedulingConfiguration: &api.SchedulingConfiguration{
					PlacementStrategy: api.SchedulingConfigurationPlacementStrategy{
						Gpu: ptr.To(api.PLACEMENTSTRATEGY_SPREAD),
					},
					GpuPodsPlacementStrategy: &api.GpuPodsPlacementStrategy{
						Node: ptr.To(api.PLACEMENTSTRATEGY_SPREAD),
					},
				},
			})
			Expect(err).To(BeNil())
			Expect(patchResp.Code).To(Equal(http.StatusOK))
			patched := patchResp.Body.(api.Nodepool)
			Expect(patched.SchedulingConfiguration.GpuPodsPlacementStrategy.Node).To(Equal(ptr.To(api.PLACEMENTSTRATEGY_SPREAD)))
			// Per setSchedulingConfigurationFieldsOnPatch: deprecated.gpu writes both columns first;
			// new.node confirms node; new.device isn't provided so device stays at the deprecated mirror value.
			Expect(patched.SchedulingConfiguration.GpuPodsPlacementStrategy.Device).To(Equal(ptr.To(api.PLACEMENTSTRATEGY_SPREAD)))
			Expect(patched.SchedulingConfiguration.PlacementStrategy.Gpu).To(Equal(ptr.To(api.PLACEMENTSTRATEGY_SPREAD)))
		})
	})

	Context("NetworkTopology validation", Ordered, func() {
		var testNetworkTopologyId, testNetworkTopologyName string
		var testNetworkTopologyClusterId string
		createNodepoolIndex := 0
		BeforeAll(func() {
			// Create a test cluster for network topology tests
			cluster, err := dao.CreateCluster(ctx, &models.Cluster{
				Name:     fmt.Sprintf("%s-network-topology", testName),
				TenantId: tenantId,
				Domain:   "xyz",
				Version:  "v2.23",
			})
			Expect(err).To(BeNil())
			testNetworkTopologyClusterId = cluster.ID.String()

			// Create a test network topology
			testUserStr := testUser
			networkTopology, err := dao.CreateNetworkTopology(ctx, models.NetworkTopology{
				CreatedAt: time.Now(),
				Name:      "test-network-topology",
				ClusterID: cluster.ID,
				TenantID:  tenantId,
				Levels:    []string{"topology-block", "topology-rack", "hostname"},
				Source:    dao.NetworkTopologiesRunAiSource,
				Phase:     models.NetworkTopologyPhaseReady,
				CreatedBy: &testUserStr,
			})
			Expect(err).To(BeNil())
			testNetworkTopologyId = networkTopology.ID.String()
			testNetworkTopologyName = networkTopology.Name
		})

		AfterAll(func() {
			// Clean up test network topology and cluster
			clusterUUID, _ := uuid.Parse(testNetworkTopologyClusterId)
			dao.DeleteNetworkTopology(ctx, tenantId, clusterUUID, testNetworkTopologyId)
			dao.DeleteCluster(ctx, testNetworkTopologyClusterId, tenantId)
		})

		It("Create nodepool with valid NetworkTopologyId", func() {
			request := api.NodepoolCreateFields{
				Name:              fmt.Sprintf("%s-network-topology-valid", testName),
				LabelKey:          fmt.Sprintf("%s-network-topology", testLabelKey),
				LabelValue:        fmt.Sprintf("%s-network-topology", testLabelValue),
				ClusterId:         testNetworkTopologyClusterId,
				NetworkTopologyId: &testNetworkTopologyId,
				SchedulingConfiguration: &api.SchedulingConfiguration{
					PlacementStrategy: api.SchedulingConfigurationPlacementStrategy{
						Cpu: ptr.To(api.PLACEMENTSTRATEGY_BINPACK),
						Gpu: ptr.To(api.PLACEMENTSTRATEGY_BINPACK),
					},
				},
			}

			response, err := nodepoolService.CreateNodepool(ctx, request)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusCreated))

			nodepool, ok := response.Body.(api.Nodepool)
			Expect(ok).To(BeTrue())
			Expect(*nodepool.NetworkTopologyId).To(Equal(testNetworkTopologyId))
			Expect(*nodepool.NetworkTopologyName).To(Equal(testNetworkTopologyName))
			createdNodepools = append(createdNodepools, nodepool)
		})

		It("Create nodepool with invalid NetworkTopologyId should fail", func() {
			invalidNetworkTopologyId := "invalid-uuid"
			request := api.NodepoolCreateFields{
				Name:              fmt.Sprintf("%s-network-topology-invalid", testName),
				LabelKey:          fmt.Sprintf("%s-network-topology", testLabelKey),
				LabelValue:        fmt.Sprintf("%s-network-topology", testLabelValue),
				ClusterId:         testNetworkTopologyClusterId,
				NetworkTopologyId: &invalidNetworkTopologyId,
				SchedulingConfiguration: &api.SchedulingConfiguration{
					PlacementStrategy: api.SchedulingConfigurationPlacementStrategy{
						Cpu: ptr.To(api.PLACEMENTSTRATEGY_BINPACK),
						Gpu: ptr.To(api.PLACEMENTSTRATEGY_BINPACK),
					},
				},
			}

			response, err := nodepoolService.CreateNodepool(ctx, request)
			Expect(err).To(BeNil())
			errorBody := response.Body.(*api.Error)
			Expect(response.Code).To(Equal(http.StatusBadRequest))
			Expect(errorBody.Message).To(ContainSubstring("invalid UUID"))
		})

		It("Create nodepool with non-existent NetworkTopologyId should fail", func() {
			nonExistentNetworkTopologyId := "7daba6ee-2aa9-4ab3-b735-a3d4bcdff72f"
			request := api.NodepoolCreateFields{
				Name:              fmt.Sprintf("%s-network-topology-nonexistent", testName),
				LabelKey:          fmt.Sprintf("%s-network-topology", testLabelKey),
				LabelValue:        fmt.Sprintf("%s-network-topology", testLabelValue),
				ClusterId:         testNetworkTopologyClusterId,
				NetworkTopologyId: &nonExistentNetworkTopologyId,
				SchedulingConfiguration: &api.SchedulingConfiguration{
					PlacementStrategy: api.SchedulingConfigurationPlacementStrategy{
						Cpu: ptr.To(api.PLACEMENTSTRATEGY_BINPACK),
						Gpu: ptr.To(api.PLACEMENTSTRATEGY_BINPACK),
					},
				},
			}

			response, err := nodepoolService.CreateNodepool(ctx, request)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusBadRequest))
		})

		It("Create nodepool without NetworkTopologyId should succeed", func() {
			createNodepoolIndex++
			request := api.NodepoolCreateFields{
				Name:              fmt.Sprintf("%s-network-topology-none-%d", testName, createNodepoolIndex),
				LabelKey:          fmt.Sprintf("%s-network-topology-%d", testLabelKey, createNodepoolIndex),
				LabelValue:        fmt.Sprintf("%s-network-topology-%d", testLabelValue, createNodepoolIndex),
				ClusterId:         testNetworkTopologyClusterId,
				NetworkTopologyId: nil,
				SchedulingConfiguration: &api.SchedulingConfiguration{
					PlacementStrategy: api.SchedulingConfigurationPlacementStrategy{
						Cpu: ptr.To(api.PLACEMENTSTRATEGY_BINPACK),
						Gpu: ptr.To(api.PLACEMENTSTRATEGY_BINPACK),
					},
				},
			}

			response, err := nodepoolService.CreateNodepool(ctx, request)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusCreated))

			nodepool, ok := response.Body.(api.Nodepool)
			Expect(ok).To(BeTrue())
			Expect(nodepool.NetworkTopologyId).To(BeNil())
			Expect(nodepool.NetworkTopologyName).To(BeNil())
			createdNodepools = append(createdNodepools, nodepool)
		})

		It("Update nodepool with valid NetworkTopologyId", func() {
			// Create a nodepool first
			createNodepoolIndex++
			labelKey := fmt.Sprintf("%s-network-topology-%d", testLabelKey, createNodepoolIndex)
			labelValue := fmt.Sprintf("%s-network-topology-%d", testLabelValue, createNodepoolIndex)
			createRequest := api.NodepoolCreateFields{
				Name:              fmt.Sprintf("%s-network-topology-update-%d", testName, createNodepoolIndex),
				LabelKey:          labelKey,
				LabelValue:        labelValue,
				ClusterId:         testNetworkTopologyClusterId,
				NetworkTopologyId: nil,
				SchedulingConfiguration: &api.SchedulingConfiguration{
					PlacementStrategy: api.SchedulingConfigurationPlacementStrategy{
						Cpu: ptr.To(api.PLACEMENTSTRATEGY_BINPACK),
						Gpu: ptr.To(api.PLACEMENTSTRATEGY_BINPACK),
					},
				},
			}

			response, err := nodepoolService.CreateNodepool(ctx, createRequest)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusCreated))

			nodepool := response.Body.(api.Nodepool)
			createdNodepools = append(createdNodepools, nodepool)

			// Update with NetworkTopologyId
			updateRequest := api.NodepoolUpdateFields{
				NetworkTopologyId: &testNetworkTopologyId,
				LabelKey:          &labelKey,
				LabelValue:        &labelValue,
			}

			updateResponse, err := nodepoolService.UpdateNodepool(ctx, nodepool.Id, updateRequest)
			Expect(err).To(BeNil())
			Expect(updateResponse.Code).To(Equal(http.StatusOK))

			updatedNodepool, ok := updateResponse.Body.(api.Nodepool)
			Expect(ok).To(BeTrue())
			Expect(updatedNodepool.NetworkTopologyId).To(Not(BeNil()))
			Expect(updatedNodepool.NetworkTopologyName).To(Not(BeNil()))
			Expect(*updatedNodepool.NetworkTopologyId).To(Equal(testNetworkTopologyId))
			Expect(*updatedNodepool.NetworkTopologyName).To(Equal(testNetworkTopologyName))
		})

		It("Update nodepool with invalid NetworkTopologyId should fail", func() {
			// Create a nodepool first
			createNodepoolIndex++
			createRequest := api.NodepoolCreateFields{
				Name:              fmt.Sprintf("%s-network-topology-update-invalid", testName),
				LabelKey:          fmt.Sprintf("%s-network-topology-%d", testLabelKey, createNodepoolIndex),
				LabelValue:        fmt.Sprintf("%s-network-topology-%d", testLabelValue, createNodepoolIndex),
				ClusterId:         testNetworkTopologyClusterId,
				NetworkTopologyId: nil,
				SchedulingConfiguration: &api.SchedulingConfiguration{
					PlacementStrategy: api.SchedulingConfigurationPlacementStrategy{
						Cpu: ptr.To(api.PLACEMENTSTRATEGY_BINPACK),
						Gpu: ptr.To(api.PLACEMENTSTRATEGY_BINPACK),
					},
				},
			}

			response, err := nodepoolService.CreateNodepool(ctx, createRequest)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusCreated))

			nodepool := response.Body.(api.Nodepool)
			createdNodepools = append(createdNodepools, nodepool)

			// Update with invalid NetworkTopologyId
			invalidNetworkTopologyId := "invalid-uuid"
			updateRequest := api.NodepoolUpdateFields{
				NetworkTopologyId: &invalidNetworkTopologyId,
			}

			updateResponse, err := nodepoolService.UpdateNodepool(ctx, nodepool.Id, updateRequest)
			errorBody := updateResponse.Body.(api.Error)
			Expect(err).To(BeNil())
			Expect(updateResponse.Code).To(Equal(http.StatusBadRequest))
			Expect(errorBody.Message).To(ContainSubstring("invalid UUID"))
		})

		It("Update nodepool to remove NetworkTopologyId", func() {
			createNodepoolIndex++
			labelKey := fmt.Sprintf("%s-network-topology-%d", testLabelKey, createNodepoolIndex)
			labelValue := fmt.Sprintf("%s-network-topology-%d", testLabelValue, createNodepoolIndex)
			// Create a nodepool with NetworkTopologyId first
			createRequest := api.NodepoolCreateFields{
				Name:              fmt.Sprintf("%s-network-topology-remove", testName),
				LabelKey:          labelKey,
				LabelValue:        labelValue,
				ClusterId:         testNetworkTopologyClusterId,
				NetworkTopologyId: &testNetworkTopologyId,
				SchedulingConfiguration: &api.SchedulingConfiguration{
					PlacementStrategy: api.SchedulingConfigurationPlacementStrategy{
						Cpu: ptr.To(api.PLACEMENTSTRATEGY_BINPACK),
						Gpu: ptr.To(api.PLACEMENTSTRATEGY_BINPACK),
					},
				},
			}

			response, err := nodepoolService.CreateNodepool(ctx, createRequest)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusCreated))

			nodepool := response.Body.(api.Nodepool)
			createdNodepools = append(createdNodepools, nodepool)
			Expect(nodepool.NetworkTopologyId).To(Not(BeNil()))
			Expect(*nodepool.NetworkTopologyId).To(Equal(testNetworkTopologyId))

			// Update to remove NetworkTopologyId (set to nil)
			updateRequest := api.NodepoolUpdateFields{
				NetworkTopologyId: nil,
				LabelValue:        &labelValue,
				LabelKey:          &labelKey,
			}

			updateResponse, err := nodepoolService.UpdateNodepool(ctx, nodepool.Id, updateRequest)
			Expect(err).To(BeNil())
			Expect(updateResponse.Code).To(Equal(http.StatusOK))

			updatedNodepool, ok := updateResponse.Body.(api.Nodepool)
			Expect(ok).To(BeTrue())
			Expect(updatedNodepool.NetworkTopologyId).To(BeNil())
			Expect(updatedNodepool.NetworkTopologyName).To(BeNil())

			nodepoolResponse, err := nodepoolService.GetNodepool(ctx, nodepool.Id)
			Expect(err).To(BeNil())
			Expect(nodepoolResponse.Code).To(Equal(http.StatusOK))

			getNodepool, ok := nodepoolResponse.Body.(api.Nodepool)
			Expect(ok).To(BeTrue())
			Expect(getNodepool.NetworkTopologyId).To(BeNil())
			Expect(getNodepool.NetworkTopologyName).To(BeNil())
		})

		It("Patch nodepool with valid NetworkTopologyId", func() {
			// Create a nodepool first
			createNodepoolIndex++
			createRequest := api.NodepoolCreateFields{
				Name:              fmt.Sprintf("%s-network-topology-patch", testName),
				LabelKey:          fmt.Sprintf("%s-network-topology-%d", testLabelKey, createNodepoolIndex),
				LabelValue:        fmt.Sprintf("%s-network-topology-%d", testLabelValue, createNodepoolIndex),
				ClusterId:         testNetworkTopologyClusterId,
				NetworkTopologyId: nil,
				SchedulingConfiguration: &api.SchedulingConfiguration{
					PlacementStrategy: api.SchedulingConfigurationPlacementStrategy{
						Cpu: ptr.To(api.PLACEMENTSTRATEGY_BINPACK),
						Gpu: ptr.To(api.PLACEMENTSTRATEGY_BINPACK),
					},
				},
			}

			response, err := nodepoolService.CreateNodepool(ctx, createRequest)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusCreated))

			nodepool := response.Body.(api.Nodepool)
			createdNodepools = append(createdNodepools, nodepool)

			// Patch with NetworkTopologyId
			patchRequest := api.NodepoolUpdateFields{
				NetworkTopologyId: &testNetworkTopologyId,
			}

			patchResponse, err := nodepoolService.PatchNodepool(ctx, nodepool.Id, patchRequest)
			Expect(err).To(BeNil())
			Expect(patchResponse.Code).To(Equal(http.StatusOK))

			patchedNodepool, ok := patchResponse.Body.(api.Nodepool)
			Expect(ok).To(BeTrue())
			Expect(patchedNodepool.NetworkTopologyId).To(Not(BeNil()))
			Expect(patchedNodepool.NetworkTopologyName).To(Not(BeNil()))
			Expect(*patchedNodepool.NetworkTopologyId).To(Equal(testNetworkTopologyId))
			Expect(*patchedNodepool.NetworkTopologyName).To(Equal(testNetworkTopologyName))
		})

		It("Patch nodepool with invalid NetworkTopologyId should fail", func() {
			// Create a nodepool first
			createNodepoolIndex++
			createRequest := api.NodepoolCreateFields{
				Name:              fmt.Sprintf("%s-network-topology-patch-invalid", testName),
				LabelKey:          fmt.Sprintf("%s-network-topology-%d", testLabelKey, createNodepoolIndex),
				LabelValue:        fmt.Sprintf("%s-network-topology-%d", testLabelValue, createNodepoolIndex),
				ClusterId:         testNetworkTopologyClusterId,
				NetworkTopologyId: nil,
				SchedulingConfiguration: &api.SchedulingConfiguration{
					PlacementStrategy: api.SchedulingConfigurationPlacementStrategy{
						Cpu: ptr.To(api.PLACEMENTSTRATEGY_BINPACK),
						Gpu: ptr.To(api.PLACEMENTSTRATEGY_BINPACK),
					},
				},
			}

			response, err := nodepoolService.CreateNodepool(ctx, createRequest)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusCreated))

			nodepool := response.Body.(api.Nodepool)
			createdNodepools = append(createdNodepools, nodepool)

			// Patch with invalid NetworkTopologyId
			invalidNetworkTopologyId := "invalid-uuid"
			patchRequest := api.NodepoolUpdateFields{
				NetworkTopologyId: &invalidNetworkTopologyId,
			}

			patchResponse, err := nodepoolService.PatchNodepool(ctx, nodepool.Id, patchRequest)
			errorBody := patchResponse.Body.(api.Error)
			Expect(err).To(BeNil())
			Expect(patchResponse.Code).To(Equal(http.StatusBadRequest))
			Expect(errorBody.Message).To(ContainSubstring("invalid UUID"))
		})

		It("GetNodepool should return NetworkTopologyId and NetworkTopologyName", func() {
			// Create a nodepool with NetworkTopologyId
			createNodepoolIndex++
			createRequest := api.NodepoolCreateFields{
				Name:              fmt.Sprintf("%s-network-topology-get", testName),
				LabelKey:          fmt.Sprintf("%s-network-topology-%d", testLabelKey, createNodepoolIndex),
				LabelValue:        fmt.Sprintf("%s-network-topology-%d", testLabelValue, createNodepoolIndex),
				ClusterId:         testNetworkTopologyClusterId,
				NetworkTopologyId: &testNetworkTopologyId,
				SchedulingConfiguration: &api.SchedulingConfiguration{
					PlacementStrategy: api.SchedulingConfigurationPlacementStrategy{
						Cpu: ptr.To(api.PLACEMENTSTRATEGY_BINPACK),
						Gpu: ptr.To(api.PLACEMENTSTRATEGY_BINPACK),
					},
				},
			}

			response, err := nodepoolService.CreateNodepool(ctx, createRequest)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusCreated))

			nodepool := response.Body.(api.Nodepool)
			createdNodepools = append(createdNodepools, nodepool)

			// Get the nodepool and verify NetworkTopology fields
			getResponse, err := nodepoolService.GetNodepool(ctx, nodepool.Id)
			Expect(err).To(BeNil())
			Expect(getResponse.Code).To(Equal(http.StatusOK))

			retrievedNodepool, ok := getResponse.Body.(api.Nodepool)
			Expect(ok).To(BeTrue())
			Expect(retrievedNodepool.NetworkTopologyId).To(Not(BeNil()))
			Expect(retrievedNodepool.NetworkTopologyName).To(Not(BeNil()))
			Expect(*retrievedNodepool.NetworkTopologyId).To(Equal(testNetworkTopologyId))
			Expect(*retrievedNodepool.NetworkTopologyName).To(Equal(testNetworkTopologyName))
		})

		It("GetNodepools should return NetworkTopologyId and NetworkTopologyName in list", func() {
			// Create a nodepool with NetworkTopologyId
			createNodepoolIndex++
			createRequest := api.NodepoolCreateFields{
				Name:              fmt.Sprintf("%s-network-topology-list", testName),
				LabelKey:          fmt.Sprintf("%s-network-topology-%d", testLabelKey, createNodepoolIndex),
				LabelValue:        fmt.Sprintf("%s-network-topology-%d", testLabelValue, createNodepoolIndex),
				ClusterId:         testNetworkTopologyClusterId,
				NetworkTopologyId: &testNetworkTopologyId,
				SchedulingConfiguration: &api.SchedulingConfiguration{
					PlacementStrategy: api.SchedulingConfigurationPlacementStrategy{
						Cpu: ptr.To(api.PLACEMENTSTRATEGY_BINPACK),
						Gpu: ptr.To(api.PLACEMENTSTRATEGY_BINPACK),
					},
				},
			}

			response, err := nodepoolService.CreateNodepool(ctx, createRequest)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusCreated))

			nodepool := response.Body.(api.Nodepool)
			createdNodepools = append(createdNodepools, nodepool)

			// Get nodepools list and verify NetworkTopology fields
			listResponse, err := nodepoolService.GetNodepools(ctx, []string{fmt.Sprintf("clusterId==%s", testNetworkTopologyClusterId)}, "", "", 0, 0)
			Expect(err).To(BeNil())
			Expect(listResponse.Code).To(Equal(http.StatusOK))

			nodepoolsList, ok := listResponse.Body.(api.GetNodepools200Response)
			Expect(ok).To(BeTrue())

			// Find our created nodepool in the list
			var foundNodepool *api.Nodepool
			for _, np := range nodepoolsList.Nodepools {
				if np.Id == nodepool.Id {
					foundNodepool = &np
					break
				}
			}
			Expect(foundNodepool).ToNot(BeNil())
			Expect(*foundNodepool.NetworkTopologyId).To(Equal(testNetworkTopologyId))
			Expect(*foundNodepool.NetworkTopologyName).To(Equal(testNetworkTopologyName))
		})

		It("NetworkTopologyId should trigger phase change to UPDATING when updated", func() {
			// Create a nodepool first
			createNodepoolIndex++
			createRequest := api.NodepoolCreateFields{
				Name:              fmt.Sprintf("%s-network-topology-phase", testName),
				LabelKey:          fmt.Sprintf("%s-network-topology-%d", testLabelKey, createNodepoolIndex),
				LabelValue:        fmt.Sprintf("%s-network-topology-%d", testLabelValue, createNodepoolIndex),
				ClusterId:         testNetworkTopologyClusterId,
				NetworkTopologyId: nil,
				SchedulingConfiguration: &api.SchedulingConfiguration{
					PlacementStrategy: api.SchedulingConfigurationPlacementStrategy{
						Cpu: ptr.To(api.PLACEMENTSTRATEGY_BINPACK),
						Gpu: ptr.To(api.PLACEMENTSTRATEGY_BINPACK),
					},
				},
			}

			response, err := nodepoolService.CreateNodepool(ctx, createRequest)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusCreated))

			nodepool := response.Body.(api.Nodepool)
			createdNodepools = append(createdNodepools, nodepool)

			// Set nodepool to READY phase first
			syncResponse, err := nodepoolService.SyncNodepoolsStatus(ctx, testNetworkTopologyClusterId, api.SyncNodepoolsStatusRequest{
				Nodepools: []api.NodepoolSyncFields{
					{
						Name:         nodepool.Name,
						Phase:        api.NODEPOOLPHASE_READY,
						PhaseMessage: "Nodepool is ready",
					},
				},
				DeleteMissing: false,
			})
			Expect(err).To(BeNil())
			Expect(syncResponse.Code).To(Equal(http.StatusNoContent))

			// Update NetworkTopologyId - should trigger phase change to UPDATING
			updateRequest := api.NodepoolUpdateFields{
				NetworkTopologyId: &testNetworkTopologyId,
				LabelValue:        &createRequest.LabelValue,
				LabelKey:          &createRequest.LabelKey,
			}

			updateResponse, err := nodepoolService.UpdateNodepool(ctx, nodepool.Id, updateRequest)
			Expect(err).To(BeNil())
			Expect(updateResponse.Code).To(Equal(http.StatusOK))

			updatedNodepool, ok := updateResponse.Body.(api.Nodepool)
			Expect(ok).To(BeTrue())
			Expect(updatedNodepool.Phase).To(Equal(api.NODEPOOLPHASE_UPDATING))
			Expect(*updatedNodepool.NetworkTopologyId).To(Equal(testNetworkTopologyId))
		})

		It("GetNodepools should filter by networkTopologyName", func() {
			// Create multiple nodepools with different network topology names
			createNodepoolIndex++
			createRequest1 := api.NodepoolCreateFields{
				Name:              fmt.Sprintf("%s-network-topology-filter-1", testName),
				LabelKey:          fmt.Sprintf("%s-network-topology-filter-1", testLabelKey),
				LabelValue:        fmt.Sprintf("%s-network-topology-filter-1", testLabelValue),
				ClusterId:         testNetworkTopologyClusterId,
				NetworkTopologyId: &testNetworkTopologyId,
				SchedulingConfiguration: &api.SchedulingConfiguration{
					PlacementStrategy: api.SchedulingConfigurationPlacementStrategy{
						Cpu: ptr.To(api.PLACEMENTSTRATEGY_BINPACK),
						Gpu: ptr.To(api.PLACEMENTSTRATEGY_BINPACK),
					},
				},
			}

			response1, err := nodepoolService.CreateNodepool(ctx, createRequest1)
			Expect(err).To(BeNil())
			Expect(response1.Code).To(Equal(http.StatusCreated))

			nodepool1 := response1.Body.(api.Nodepool)
			createdNodepools = append(createdNodepools, nodepool1)

			// Create another nodepool with the same network topology
			createNodepoolIndex++
			createRequest2 := api.NodepoolCreateFields{
				Name:              fmt.Sprintf("%s-network-topology-filter-2", testName),
				LabelKey:          fmt.Sprintf("%s-network-topology-filter-2", testLabelKey),
				LabelValue:        fmt.Sprintf("%s-network-topology-filter-2", testLabelValue),
				ClusterId:         testNetworkTopologyClusterId,
				NetworkTopologyId: &testNetworkTopologyId,
				SchedulingConfiguration: &api.SchedulingConfiguration{
					PlacementStrategy: api.SchedulingConfigurationPlacementStrategy{
						Cpu: ptr.To(api.PLACEMENTSTRATEGY_BINPACK),
						Gpu: ptr.To(api.PLACEMENTSTRATEGY_BINPACK),
					},
				},
			}

			response2, err := nodepoolService.CreateNodepool(ctx, createRequest2)
			Expect(err).To(BeNil())
			Expect(response2.Code).To(Equal(http.StatusCreated))

			nodepool2 := response2.Body.(api.Nodepool)
			createdNodepools = append(createdNodepools, nodepool2)

			// Filter by networkTopologyName
			filterBy := []string{fmt.Sprintf("networkTopologyName==%s", testNetworkTopologyName)}
			listResponse, err := nodepoolService.GetNodepools(ctx, filterBy, "", "", 0, 0)
			Expect(err).To(BeNil())
			Expect(listResponse.Code).To(Equal(http.StatusOK))

			nodepoolsList, ok := listResponse.Body.(api.GetNodepools200Response)
			Expect(ok).To(BeTrue())

			// Should return nodepools with the specified network topology name
			Expect(len(nodepoolsList.Nodepools)).To(Equal(8))

			// Verify all returned nodepools have the correct network topology name
			for _, np := range nodepoolsList.Nodepools {
				Expect(np.NetworkTopologyName).To(Not(BeNil()))
				Expect(*np.NetworkTopologyName).To(Equal(testNetworkTopologyName))
			}
		})

		It("GetNodepools should sort by networkTopologyName in ascending order", func() {
			// Create nodepools with different network topology names for sorting test
			createNodepoolIndex++
			createRequest1 := api.NodepoolCreateFields{
				Name:              fmt.Sprintf("%s-network-topology-sort-1", testName),
				LabelKey:          fmt.Sprintf("%s-network-topology-sort-1", testLabelKey),
				LabelValue:        fmt.Sprintf("%s-network-topology-sort-1", testLabelValue),
				ClusterId:         testNetworkTopologyClusterId,
				NetworkTopologyId: &testNetworkTopologyId,
				SchedulingConfiguration: &api.SchedulingConfiguration{
					PlacementStrategy: api.SchedulingConfigurationPlacementStrategy{
						Cpu: ptr.To(api.PLACEMENTSTRATEGY_BINPACK),
						Gpu: ptr.To(api.PLACEMENTSTRATEGY_BINPACK),
					},
				},
			}

			response1, err := nodepoolService.CreateNodepool(ctx, createRequest1)
			Expect(err).To(BeNil())
			Expect(response1.Code).To(Equal(http.StatusCreated))

			nodepool1 := response1.Body.(api.Nodepool)
			createdNodepools = append(createdNodepools, nodepool1)

			// Create another nodepool with the same network topology
			createNodepoolIndex++
			createRequest2 := api.NodepoolCreateFields{
				Name:              fmt.Sprintf("%s-network-topology-sort-2", testName),
				LabelKey:          fmt.Sprintf("%s-network-topology-sort-2", testLabelKey),
				LabelValue:        fmt.Sprintf("%s-network-topology-sort-2", testLabelValue),
				ClusterId:         testNetworkTopologyClusterId,
				NetworkTopologyId: &testNetworkTopologyId,
				SchedulingConfiguration: &api.SchedulingConfiguration{
					PlacementStrategy: api.SchedulingConfigurationPlacementStrategy{
						Cpu: ptr.To(api.PLACEMENTSTRATEGY_BINPACK),
						Gpu: ptr.To(api.PLACEMENTSTRATEGY_BINPACK),
					},
				},
			}

			response2, err := nodepoolService.CreateNodepool(ctx, createRequest2)
			Expect(err).To(BeNil())
			Expect(response2.Code).To(Equal(http.StatusCreated))

			nodepool2 := response2.Body.(api.Nodepool)
			createdNodepools = append(createdNodepools, nodepool2)

			// Sort by networkTopologyName in ascending order
			sortBy := api.NODEPOOLSORTFILTERFIELDS_NETWORK_TOPOLOGY_NAME
			sortOrder := "asc"
			listResponse, err := nodepoolService.GetNodepools(ctx, []string{}, sortBy, sortOrder, 0, 0)
			Expect(err).To(BeNil())
			Expect(listResponse.Code).To(Equal(http.StatusOK))

			nodepoolsList, ok := listResponse.Body.(api.GetNodepools200Response)
			Expect(ok).To(BeTrue())

			// Should return nodepools sorted by networkTopologyName
			Expect(len(nodepoolsList.Nodepools)).To(BeNumerically(">=", 2))

			//// Verify sorting - nodepools with network topology names should come first
			//// (assuming empty/null values come after non-empty values in ascending order)
			var lastNetworkTopologyName string

			for _, np := range nodepoolsList.Nodepools {
				if np.NetworkTopologyName != nil && *np.NetworkTopologyName != "" {
					if len(lastNetworkTopologyName) > 0 {
						lastNetworkTopologyName = *np.NetworkTopologyName
					} else {
						// Verify ascending order
						isGreaterOrEqual := *np.NetworkTopologyName >= lastNetworkTopologyName
						Expect(isGreaterOrEqual).To(BeTrue())
						lastNetworkTopologyName = *np.NetworkTopologyName
					}
				}
			}

			// 1. Find the index where the NetworkTopologyName becomes non-nil
			firstNonNilIndex := -1
			for i, np := range nodepoolsList.Nodepools {
				if np.NetworkTopologyName != nil {
					firstNonNilIndex = i
					break
				}
			}

			// 2. Verify that the remaining nodepools are sorted alphabetically
			for i := firstNonNilIndex; i < len(nodepoolsList.Nodepools); i++ {
				np := nodepoolsList.Nodepools[i]

				// The first non-nil element sets the baseline
				if i == firstNonNilIndex {
					lastNetworkTopologyName = *np.NetworkTopologyName
				} else {
					// Check for ascending order
					isGreaterOrEqual := *np.NetworkTopologyName >= lastNetworkTopologyName
					Expect(isGreaterOrEqual).To(BeTrue())
					lastNetworkTopologyName = *np.NetworkTopologyName
				}
			}
		})

		It("GetNodepools should sort by networkTopologyName in descending order", func() {
			// Create nodepools with different network topology names for sorting test
			createNodepoolIndex++
			createRequest1 := api.NodepoolCreateFields{
				Name:              fmt.Sprintf("%s-network-topology-sort-desc-1", testName),
				LabelKey:          fmt.Sprintf("%s-network-topology-sort-desc-1", testLabelKey),
				LabelValue:        fmt.Sprintf("%s-network-topology-sort-desc-1", testLabelValue),
				ClusterId:         testNetworkTopologyClusterId,
				NetworkTopologyId: &testNetworkTopologyId,
				SchedulingConfiguration: &api.SchedulingConfiguration{
					PlacementStrategy: api.SchedulingConfigurationPlacementStrategy{
						Cpu: ptr.To(api.PLACEMENTSTRATEGY_BINPACK),
						Gpu: ptr.To(api.PLACEMENTSTRATEGY_BINPACK),
					},
				},
			}

			response1, err := nodepoolService.CreateNodepool(ctx, createRequest1)
			Expect(err).To(BeNil())
			Expect(response1.Code).To(Equal(http.StatusCreated))

			nodepool1 := response1.Body.(api.Nodepool)
			createdNodepools = append(createdNodepools, nodepool1)

			// Create another nodepool with the same network topology
			createNodepoolIndex++
			createRequest2 := api.NodepoolCreateFields{
				Name:              fmt.Sprintf("%s-network-topology-sort-desc-2", testName),
				LabelKey:          fmt.Sprintf("%s-network-topology-sort-desc-2", testLabelKey),
				LabelValue:        fmt.Sprintf("%s-network-topology-sort-desc-2", testLabelValue),
				ClusterId:         testNetworkTopologyClusterId,
				NetworkTopologyId: &testNetworkTopologyId,
				SchedulingConfiguration: &api.SchedulingConfiguration{
					PlacementStrategy: api.SchedulingConfigurationPlacementStrategy{
						Cpu: ptr.To(api.PLACEMENTSTRATEGY_BINPACK),
						Gpu: ptr.To(api.PLACEMENTSTRATEGY_BINPACK),
					},
				},
			}

			response2, err := nodepoolService.CreateNodepool(ctx, createRequest2)
			Expect(err).To(BeNil())
			Expect(response2.Code).To(Equal(http.StatusCreated))

			nodepool2 := response2.Body.(api.Nodepool)
			createdNodepools = append(createdNodepools, nodepool2)

			// Sort by networkTopologyName in descending order
			sortBy := api.NODEPOOLSORTFILTERFIELDS_NETWORK_TOPOLOGY_NAME
			sortOrder := "desc"
			listResponse, err := nodepoolService.GetNodepools(ctx, []string{}, sortBy, sortOrder, 0, 0)
			Expect(err).To(BeNil())
			Expect(listResponse.Code).To(Equal(http.StatusOK))

			nodepoolsList, ok := listResponse.Body.(api.GetNodepools200Response)
			Expect(ok).To(BeTrue())

			// Should return nodepools sorted by networkTopologyName
			Expect(len(nodepoolsList.Nodepools)).To(BeNumerically(">=", 2))

			// Verify sorting - nodepools with network topology names should come first
			// (assuming empty/null values come after non-empty values in descending order)
			var hasNetworkTopology bool
			var lastNetworkTopologyName string

			for _, np := range nodepoolsList.Nodepools {
				if np.NetworkTopologyName != nil && *np.NetworkTopologyName != "" {
					if !hasNetworkTopology {
						hasNetworkTopology = true
						lastNetworkTopologyName = *np.NetworkTopologyName
					} else {
						// Verify descending order
						isLowerOrEqual := *np.NetworkTopologyName <= lastNetworkTopologyName
						Expect(isLowerOrEqual).To(BeTrue())
						lastNetworkTopologyName = *np.NetworkTopologyName
					}
				}
			}
		})

		It("GetNodepools should filter and sort by networkTopologyName", func() {
			// Create nodepools with different network topology names for combined test
			createNodepoolIndex++
			createRequest1 := api.NodepoolCreateFields{
				Name:              fmt.Sprintf("%s-network-topology-combined-1", testName),
				LabelKey:          fmt.Sprintf("%s-network-topology-combined-1", testLabelKey),
				LabelValue:        fmt.Sprintf("%s-network-topology-combined-1", testLabelValue),
				ClusterId:         testNetworkTopologyClusterId,
				NetworkTopologyId: &testNetworkTopologyId,
				SchedulingConfiguration: &api.SchedulingConfiguration{
					PlacementStrategy: api.SchedulingConfigurationPlacementStrategy{
						Cpu: ptr.To(api.PLACEMENTSTRATEGY_BINPACK),
						Gpu: ptr.To(api.PLACEMENTSTRATEGY_BINPACK),
					},
				},
			}

			response1, err := nodepoolService.CreateNodepool(ctx, createRequest1)
			Expect(err).To(BeNil())
			Expect(response1.Code).To(Equal(http.StatusCreated))

			nodepool1 := response1.Body.(api.Nodepool)
			createdNodepools = append(createdNodepools, nodepool1)

			// Create another nodepool with the same network topology
			createNodepoolIndex++
			createRequest2 := api.NodepoolCreateFields{
				Name:              fmt.Sprintf("%s-network-topology-combined-2", testName),
				LabelKey:          fmt.Sprintf("%s-network-topology-combined-2", testLabelKey),
				LabelValue:        fmt.Sprintf("%s-network-topology-combined-2", testLabelValue),
				ClusterId:         testNetworkTopologyClusterId,
				NetworkTopologyId: &testNetworkTopologyId,
				SchedulingConfiguration: &api.SchedulingConfiguration{
					PlacementStrategy: api.SchedulingConfigurationPlacementStrategy{
						Cpu: ptr.To(api.PLACEMENTSTRATEGY_BINPACK),
						Gpu: ptr.To(api.PLACEMENTSTRATEGY_BINPACK),
					},
				},
			}

			response2, err := nodepoolService.CreateNodepool(ctx, createRequest2)
			Expect(err).To(BeNil())
			Expect(response2.Code).To(Equal(http.StatusCreated))

			nodepool2 := response2.Body.(api.Nodepool)
			createdNodepools = append(createdNodepools, nodepool2)

			// Filter by networkTopologyName and sort by networkTopologyName in ascending order
			filterBy := []string{fmt.Sprintf("networkTopologyName==%s", testNetworkTopologyName)}
			sortBy := api.NODEPOOLSORTFILTERFIELDS_NETWORK_TOPOLOGY_NAME
			sortOrder := "asc"
			listResponse, err := nodepoolService.GetNodepools(ctx, filterBy, sortBy, sortOrder, 0, 0)
			Expect(err).To(BeNil())
			Expect(listResponse.Code).To(Equal(http.StatusOK))

			nodepoolsList, ok := listResponse.Body.(api.GetNodepools200Response)
			Expect(ok).To(BeTrue())

			// Should return nodepools filtered by network topology name and sorted
			Expect(len(nodepoolsList.Nodepools)).To(BeNumerically(">=", 2))

			// Verify all returned nodepools have the correct network topology name
			for _, np := range nodepoolsList.Nodepools {
				Expect(np.NetworkTopologyName).ToNot(BeNil())
				Expect(*np.NetworkTopologyName).To(Equal(testNetworkTopologyName))
			}

			// Verify sorting - should be in ascending order
			var lastNetworkTopologyName string
			for i, np := range nodepoolsList.Nodepools {
				if i == 0 {
					lastNetworkTopologyName = *np.NetworkTopologyName
				} else {
					isGreaterOrEqual := *np.NetworkTopologyName >= lastNetworkTopologyName
					Expect(isGreaterOrEqual).To(BeTrue())
					lastNetworkTopologyName = *np.NetworkTopologyName
				}
			}
		})

		It("SyncNodepools should persist the missing prerequisites phase and NRT condition", func() {
			// Create a nodepool first
			createRequest := api.NodepoolCreateFields{
				Name:       testName + "-conditions-sync-test",
				LabelKey:   testLabelKey + "-conditions-sync-test",
				LabelValue: testLabelValue + "-conditions-sync-test",
				ClusterId:  testClusterId,
				SchedulingConfiguration: &api.SchedulingConfiguration{
					PlacementStrategy: api.SchedulingConfigurationPlacementStrategy{
						Cpu: ptr.To(api.PLACEMENTSTRATEGY_BINPACK),
						Gpu: ptr.To(api.PLACEMENTSTRATEGY_BINPACK),
					},
				},
			}

			response, err := nodepoolService.CreateNodepool(ctx, createRequest)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusCreated))

			nodepool := response.Body.(api.Nodepool)
			createdNodepools = append(createdNodepools, nodepool)

			// Set one condition in the status
			singleCondition := []api.NodepoolConditionDetails{
				{
					Type:    missingNrtHealthyPrerequisiteCondition,
					Reason:  missingNrtHealthyPrerequisiteReason,
					Message: missingPrerequisitesPhaseMessage,
					Status:  "True",
				},
			}

			// Sync the nodepool status with the condition
			syncRequest := api.SyncNodepoolsStatusRequest{
				Nodepools: []api.NodepoolSyncFields{
					{
						Name:         nodepool.Name,
						Phase:        api.NODEPOOLPHASE_MISSING_PREREQUISITES,
						PhaseMessage: missingPrerequisitesPhaseMessage,
						Status: api.NodepoolSyncFieldsStatus{
							LabelKey:   ptr.To(nodepool.LabelKey),
							LabelValue: ptr.To(nodepool.LabelValue),
							SchedulingConfiguration: &api.SchedulingConfiguration{
								PlacementStrategy: api.SchedulingConfigurationPlacementStrategy{
									Cpu: nodepool.SchedulingConfiguration.PlacementStrategy.Cpu,
									Gpu: nodepool.SchedulingConfiguration.PlacementStrategy.Gpu,
								},
							},
							Nodes:      []string{"worker-node-1", "worker-node-2"},
							Conditions: singleCondition,
						},
					},
				},
				DeleteMissing: false,
			}

			syncResponse, err := nodepoolService.SyncNodepoolsStatus(ctx, testClusterId, syncRequest)
			Expect(err).To(BeNil())
			Expect(syncResponse.Code).To(Equal(http.StatusNoContent))

			// Validate that GetNodepools returns the correct data including the condition values
			response, err = nodepoolService.GetNodepools(ctx, []string{}, "", "", 0, 0)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			nodepools, ok := response.Body.(api.GetNodepools200Response)
			Expect(ok).To(BeTrue())

			// Find our updated nodepool in the list
			var foundNodepool *api.Nodepool
			for _, np := range nodepools.Nodepools {
				if np.Id == nodepool.Id {
					foundNodepool = &np
					break
				}
			}
			Expect(foundNodepool).ToNot(BeNil())
			Expect(foundNodepool.Phase).To(Equal(api.NODEPOOLPHASE_MISSING_PREREQUISITES))
			Expect(foundNodepool.PhaseMessage).To(Equal(missingPrerequisitesPhaseMessage))

			// Validate the Conditions field is properly set
			Expect(foundNodepool.Status.Conditions).To(Not(BeNil()))
			Expect(len(foundNodepool.Status.Conditions)).To(Equal(1))

			// Validate the specific condition values
			condition := foundNodepool.Status.Conditions[0]
			Expect(condition.Type).To(Equal(missingNrtHealthyPrerequisiteCondition))
			Expect(condition.Reason).To(Equal(missingNrtHealthyPrerequisiteReason))
			Expect(condition.Status).To(Equal("True"))
			Expect(condition.Message).To(Equal(missingPrerequisitesPhaseMessage))

			// Also validate using GetNodepool (single nodepool retrieval)
			getResponse, err := nodepoolService.GetNodepool(ctx, nodepool.Id)
			Expect(err).To(BeNil())
			Expect(getResponse.Code).To(Equal(http.StatusOK))

			singleNodepool, ok := getResponse.Body.(api.Nodepool)
			Expect(ok).To(BeTrue())
			Expect(singleNodepool.Phase).To(Equal(api.NODEPOOLPHASE_MISSING_PREREQUISITES))
			Expect(singleNodepool.PhaseMessage).To(Equal(missingPrerequisitesPhaseMessage))

			// Validate the Conditions field in single nodepool response
			Expect(singleNodepool.Status.Conditions).To(Not(BeNil()))
			Expect(len(singleNodepool.Status.Conditions)).To(Equal(1))

			singleConditionItem := singleNodepool.Status.Conditions[0]
			Expect(singleConditionItem.Type).To(Equal(missingNrtHealthyPrerequisiteCondition))
			Expect(singleConditionItem.Reason).To(Equal(missingNrtHealthyPrerequisiteReason))
			Expect(singleConditionItem.Message).To(Equal(missingPrerequisitesPhaseMessage))
		})
	})
})

var _ = Describe("Legacy nodepool routes tests", Ordered, func() {
	tenantSettingsClient := &bc_mocks.TenantSettingsClient{}
	orgClient := &org_unit_service_client_mock.IOrgUnitClient{}

	ctx := createContext()

	tenantSettingsClient.On("GetOPRSetting", mock.Anything).Return(true, nil)
	orgClient.On("NodePoolCreated", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	orgClient.On("NodePoolDeleted", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	orgClient.On("DefaultNodePoolCreated", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	nodepoolService := services.NewNodePoolsApiService(tenantSettingsClient, orgClient)

	var legacyClusterId, legacyClusterName string
	var createdNodepool api.Nodepool

	BeforeAll(func() {
		cluster, dbErr := dao.CreateCluster(ctx, &models.Cluster{
			Name:     "legacy-test-cluster",
			TenantId: tenantId,
			Domain:   "xyz",
			Version:  "v2.24",
		})
		Expect(dbErr).To(BeNil())
		legacyClusterId = cluster.ID.String()
		legacyClusterName = cluster.Name

		res, err := nodepoolService.ClusterNotification(ctx, api.ClusterNotification{
			Operation: api.NOTIFICATIONOPERATION_CREATE,
			Id:        cluster.ID.String(),
			Name:      cluster.Name,
		})
		Expect(err).To(BeNil())
		Expect(res.Code).To(Equal(http.StatusNoContent))

		Eventually(func(g Gomega) {
			defaultNp := getDefaultNodePoolOfCluster(ctx, nodepoolService, legacyClusterId)
			Expect(defaultNp).To(Not(BeNil()))
		}, 3*time.Second, 100*time.Millisecond).Should(Succeed())

		// Create a nodepool to use in legacy tests
		request := api.NodepoolCreateFields{
			Name:       fmt.Sprintf("legacy-np-%s", rand.String(4)),
			LabelKey:   "legacy-key",
			LabelValue: "legacy-value",
			ClusterId:  legacyClusterId,
			SchedulingConfiguration: &api.SchedulingConfiguration{
				PlacementStrategy: api.SchedulingConfigurationPlacementStrategy{
					Cpu: ptr.To(api.PLACEMENTSTRATEGY_BINPACK),
					Gpu: ptr.To(api.PLACEMENTSTRATEGY_BINPACK),
				},
			},
		}
		response, err := nodepoolService.CreateNodepool(ctx, request)
		Expect(err).To(BeNil())
		Expect(response.Code).To(Equal(http.StatusCreated))

		np, ok := response.Body.(api.Nodepool)
		Expect(ok).To(BeTrue())
		createdNodepool = np

		// Set phase to Ready
		phaseErr := dao.SetNodepoolPhase(ctx, tenantId, testUser, createdNodepool.Id, api.NODEPOOLPHASE_READY)
		Expect(phaseErr).To(BeNil())
	})

	AfterAll(func() {
		_ = dao.DeleteNodepool(ctx, tenantId, "", createdNodepool.Id)
		dao.DeleteCluster(ctx, legacyClusterId, tenantId)

		res, err := nodepoolService.ClusterNotification(ctx, api.ClusterNotification{
			Operation: api.NOTIFICATIONOPERATION_DELETE,
			Id:        legacyClusterId,
			Name:      legacyClusterName,
		})
		Expect(err).To(BeNil())
		Expect(res.Code).To(Equal(http.StatusNoContent))
	})

	Context("LegacyUpdateNodepool", func() {
		It("should update nodepool via legacy route and return same result as UpdateNodepool", func() {
			updateFields := api.LegacyUpdateNodePoolRequest{
				LabelKey:   "legacy-key-updated",
				LabelValue: "legacy-value-updated",
			}

			response, err := nodepoolService.LegacyUpdateNodepool(ctx, legacyClusterId, createdNodepool.Id, updateFields)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			nodepool, ok := response.Body.(api.Nodepool)
			Expect(ok).To(BeTrue())
			Expect(nodepool.Id).To(Equal(createdNodepool.Id))
			Expect(nodepool.LabelKey).To(Equal("legacy-key-updated"))
			Expect(nodepool.LabelValue).To(Equal("legacy-value-updated"))

			createdNodepool = nodepool
		})

		It("should ignore clusterId and work with nodepoolId only", func() {
			fakeClusterId := uuid.New().String()
			updateFields := api.LegacyUpdateNodePoolRequest{
				LabelKey:   "legacy-key-v2",
				LabelValue: "legacy-value-v2",
			}

			response, err := nodepoolService.LegacyUpdateNodepool(ctx, fakeClusterId, createdNodepool.Id, updateFields)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			nodepool, ok := response.Body.(api.Nodepool)
			Expect(ok).To(BeTrue())
			Expect(nodepool.Id).To(Equal(createdNodepool.Id))
			Expect(nodepool.LabelKey).To(Equal("legacy-key-v2"))

			createdNodepool = nodepool
		})
	})

	Context("LegacyUpdateNodepoolLabels", func() {
		It("should patch nodepool labels via legacy route", func() {
			patchFields := api.LegacyNodePoolLabelsRequest{
				LabelKey:   "legacy-key-patched",
				LabelValue: "legacy-value-patched",
			}

			response, err := nodepoolService.LegacyUpdateNodepoolLabels(ctx, legacyClusterId, createdNodepool.Id, patchFields)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			nodepool, ok := response.Body.(api.Nodepool)
			Expect(ok).To(BeTrue())
			Expect(nodepool.Id).To(Equal(createdNodepool.Id))
			Expect(nodepool.LabelKey).To(Equal("legacy-key-patched"))
			Expect(nodepool.LabelValue).To(Equal("legacy-value-patched"))

			createdNodepool = nodepool
		})

		It("should fail when updating labels of default nodepool", func() {
			defaultNp := getDefaultNodePoolOfCluster(ctx, nodepoolService, legacyClusterId)
			Expect(defaultNp).ToNot(BeNil())

			patchFields := api.LegacyNodePoolLabelsRequest{
				LabelKey:   "some-key",
				LabelValue: "some-value",
			}

			response, err := nodepoolService.LegacyUpdateNodepoolLabels(ctx, legacyClusterId, defaultNp.Id, patchFields)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusBadRequest))
		})
	})

	Context("LegacyDeleteNodepool", func() {
		It("should delete nodepool via legacy route (soft delete)", func() {
			response, err := nodepoolService.LegacyDeleteNodepool(ctx, legacyClusterId, createdNodepool.Id)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusAccepted))

			// Verify phase changed to DELETING
			getResponse, err := nodepoolService.GetNodepool(ctx, createdNodepool.Id)
			Expect(err).To(BeNil())
			Expect(getResponse.Code).To(Equal(http.StatusOK))

			nodepool, ok := getResponse.Body.(api.Nodepool)
			Expect(ok).To(BeTrue())
			Expect(nodepool.Phase).To(Equal(api.NODEPOOLPHASE_DELETING))
		})

		It("should fail when deleting default nodepool", func() {
			defaultNp := getDefaultNodePoolOfCluster(ctx, nodepoolService, legacyClusterId)
			Expect(defaultNp).ToNot(BeNil())

			response, err := nodepoolService.LegacyDeleteNodepool(ctx, legacyClusterId, defaultNp.Id)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusBadRequest))
		})
	})

	Context("LegacyGetNodepools", func() {
		It("should return nodepools for cluster in legacy format", func() {
			// Create a new nodepool for this test
			npName := fmt.Sprintf("legacy-get-np-%s", rand.String(4))
			createReq := api.NodepoolCreateFields{
				Name:       npName,
				LabelKey:   "get-key",
				LabelValue: "get-value",
				ClusterId:  legacyClusterId,
				SchedulingConfiguration: &api.SchedulingConfiguration{
					PlacementStrategy: api.SchedulingConfigurationPlacementStrategy{
						Cpu: ptr.To(api.PLACEMENTSTRATEGY_BINPACK),
						Gpu: ptr.To(api.PLACEMENTSTRATEGY_SPREAD),
					},
				},
			}
			createResp, err := nodepoolService.CreateNodepool(ctx, createReq)
			Expect(err).To(BeNil())
			Expect(createResp.Code).To(Equal(http.StatusCreated))
			np := createResp.Body.(api.Nodepool)

			response, err := nodepoolService.LegacyGetNodepools(ctx, legacyClusterId)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			nodepools, ok := response.Body.([]api.LegacyNodePool)
			Expect(ok).To(BeTrue())
			Expect(len(nodepools)).To(BeNumerically(">=", 1))

			// Find our created nodepool
			var found *api.LegacyNodePool
			for i, n := range nodepools {
				if n.Name == npName {
					found = &nodepools[i]
					break
				}
			}
			Expect(found).ToNot(BeNil())
			Expect(found.LabelKey).To(Equal("get-key"))
			Expect(found.LabelValue).To(Equal("get-value"))
			Expect(found.ClusterId).To(Equal(legacyClusterId))
			Expect(found.Status).To(Equal("Creating"))

			// Cleanup
			_ = dao.DeleteNodepool(ctx, tenantId, "", np.Id)
		})
	})

	Context("LegacyCreateNodepool", func() {
		It("should create nodepool via legacy route and return 201", func() {
			npName := fmt.Sprintf("legacy-create-np-%s", rand.String(4))
			createReq := api.LegacyNodePoolCreateRequest{
				Name:       npName,
				LabelKey:   "create-key",
				LabelValue: "create-value",
			}

			response, err := nodepoolService.LegacyCreateNodepool(ctx, legacyClusterId, createReq)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusCreated))

			// Verify nodepool was created by listing
			getResp, err := nodepoolService.LegacyGetNodepools(ctx, legacyClusterId)
			Expect(err).To(BeNil())
			nodepools := getResp.Body.([]api.LegacyNodePool)

			var found *api.LegacyNodePool
			for i, n := range nodepools {
				if n.Name == npName {
					found = &nodepools[i]
					break
				}
			}
			Expect(found).ToNot(BeNil())
			Expect(found.LabelKey).To(Equal("create-key"))
			Expect(found.LabelValue).To(Equal("create-value"))

			// Cleanup
			_ = dao.DeleteNodepool(ctx, tenantId, "", fmt.Sprintf("%d", int(found.Id)))
		})
	})

	Context("LegacyUpdateNodepoolStatus", func() {
		It("should update nodepool status via legacy route", func() {
			// Create a nodepool for this test
			npName := fmt.Sprintf("legacy-status-np-%s", rand.String(4))
			createReq := api.NodepoolCreateFields{
				Name:       npName,
				LabelKey:   "status-key",
				LabelValue: "status-value",
				ClusterId:  legacyClusterId,
				SchedulingConfiguration: &api.SchedulingConfiguration{
					PlacementStrategy: api.SchedulingConfigurationPlacementStrategy{
						Cpu: ptr.To(api.PLACEMENTSTRATEGY_BINPACK),
						Gpu: ptr.To(api.PLACEMENTSTRATEGY_BINPACK),
					},
				},
			}
			createResp, err := nodepoolService.CreateNodepool(ctx, createReq)
			Expect(err).To(BeNil())
			Expect(createResp.Code).To(Equal(http.StatusCreated))
			np := createResp.Body.(api.Nodepool)

			statusReq := api.LegacyNodePoolStatusRequest{
				Status:        "Ready",
				StatusMessage: "all nodes healthy",
				Nodes:         `["node-a","node-b"]`,
			}

			response, err := nodepoolService.LegacyUpdateNodepoolStatus(ctx, legacyClusterId, npName, statusReq)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusCreated))

			// Verify status was updated
			getResp, err := nodepoolService.GetNodepool(ctx, np.Id)
			Expect(err).To(BeNil())
			nodepool := getResp.Body.(api.Nodepool)
			Expect(nodepool.Phase).To(Equal(api.NODEPOOLPHASE_READY))

			// Cleanup
			_ = dao.DeleteNodepool(ctx, tenantId, "", np.Id)
		})
	})

	Context("LegacySyncNodepools", func() {
		It("should sync nodepools via legacy route and return 200", func() {
			syncReqs := []api.LegacyNodePoolStatusRequest{
				{
					Status:        "Ready",
					StatusMessage: "healthy",
				},
			}

			response, err := nodepoolService.LegacySyncNodepools(ctx, legacyClusterId, syncReqs)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))
		})
	})
})

func nodepoolCreateFields(clusterId, testSuffix string, gpuNetworkAccelerationDetection *api.NodepoolGpuNetworkAccelerationDetection) *api.NodepoolCreateFields {
	return &api.NodepoolCreateFields{
		Name:       fmt.Sprintf("%s%s-%s", testName, testSuffix, rand.String(4)),
		LabelKey:   fmt.Sprintf("%s%s", testLabelKey, testSuffix),
		LabelValue: fmt.Sprintf("%s%s", testLabelValue, testSuffix),
		ClusterId:  clusterId,
		SchedulingConfiguration: &api.SchedulingConfiguration{
			PlacementStrategy: api.SchedulingConfigurationPlacementStrategy{
				Cpu: ptr.To(api.PLACEMENTSTRATEGY_BINPACK),
				Gpu: ptr.To(api.PLACEMENTSTRATEGY_BINPACK),
			},
		},
		GpuNetworkAccelerationDetection: gpuNetworkAccelerationDetection,
		GpuNetworkAccelerationLabelKey:  nil,
	}
}

func nodepoolUpdateFields(testSuffix string, gpuNetworkAccelerationDetection *api.NodepoolGpuNetworkAccelerationDetection) *api.NodepoolUpdateFields {
	return &api.NodepoolUpdateFields{
		LabelKey:   ptr.To(fmt.Sprintf("%s%s", testLabelKey, testSuffix)),
		LabelValue: ptr.To(fmt.Sprintf("%s%s", testLabelValue, testSuffix)),
		SchedulingConfiguration: &api.SchedulingConfiguration{
			PlacementStrategy: api.SchedulingConfigurationPlacementStrategy{
				Cpu: ptr.To(api.PLACEMENTSTRATEGY_SPREAD),
				Gpu: ptr.To(api.PLACEMENTSTRATEGY_BINPACK),
			},
		},
		GpuNetworkAccelerationDetection: gpuNetworkAccelerationDetection,
	}
}

func createContext() context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, runaictx.CfTenantId, tenantIdStr)
	ctx = context.WithValue(ctx, runaictx.CfSubjectID, testUser)
	ctx = context.WithValue(ctx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{
		Tenants: []string{tenantIdStr},
	})
	ctx = addAuditToCtx(ctx)
	return ctx
}

func getDefaultNodePoolOfCluster(ctx context.Context, nodepoolService api.NodePoolsAPIServicer, clusterId string) *api.Nodepool {
	response, err := nodepoolService.GetNodepools(ctx, []string{fmt.Sprintf("clusterId==%s", clusterId)}, "", "", 0, 0)
	ExpectWithOffset(1, err).To(BeNil())
	ExpectWithOffset(1, response.Code).To(Equal(http.StatusOK))

	nodepools := response.Body.(api.GetNodepools200Response).Nodepools
	for _, np := range nodepools {
		if np.Name == services.NodepoolDefaultName {
			return ptr.To(np)
		}
	}
	return nil
}

func validateTimeBasedFairShareConfig(nodepool api.Nodepool, expected api.TimeBasedFairShare) {
	Expect(nodepool.SchedulingConfiguration).ToNot(BeNil())

	actual := nodepool.SchedulingConfiguration.TimeBasedFairShare
	Expect(actual).ToNot(BeNil())

	Expect(actual.Enabled).ToNot(BeNil())
	if expected.Enabled == nil {
		Expect(*actual.Enabled).To(BeFalse())
	} else {
		Expect(*actual.Enabled).To(Equal(*expected.Enabled))
	}

	if expected.HistoricalUsageWeight == nil {
		Expect(actual.HistoricalUsageWeight).To(BeNil())
	} else {
		Expect(actual.HistoricalUsageWeight).ToNot(BeNil())
		Expect(*actual.HistoricalUsageWeight).To(Equal(*expected.HistoricalUsageWeight))
	}

	if expected.DecayHalfLife == nil {
		Expect(actual.DecayHalfLife).To(BeNil())
	} else {
		Expect(actual.DecayHalfLife).ToNot(BeNil())
		Expect(*actual.DecayHalfLife).To(Equal(*expected.DecayHalfLife))
	}

	if expected.WindowDuration == nil {
		Expect(actual.WindowDuration).To(BeNil())
	} else {
		Expect(actual.WindowDuration).ToNot(BeNil())
		Expect(*actual.WindowDuration).To(Equal(*expected.WindowDuration))
	}

	if expected.WindowType == nil {
		Expect(actual.WindowType).To(BeNil())
	} else {
		Expect(actual.WindowType).ToNot(BeNil())
		Expect(*actual.WindowType).To(Equal(*expected.WindowType))
	}

	if expected.TumblingWindowStartTime == nil {
		Expect(actual.TumblingWindowStartTime).To(BeNil())
	} else {
		Expect(actual.TumblingWindowStartTime).ToNot(BeNil())
		Expect(*actual.TumblingWindowStartTime).To(Equal(*expected.TumblingWindowStartTime))
	}
}
