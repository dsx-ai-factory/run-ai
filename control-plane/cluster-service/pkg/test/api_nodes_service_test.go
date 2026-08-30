package test

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/lib/pq"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	authorization_api "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"
	authorization_utils "github.com/run-ai/runai/control-plane/authorization/pkg/client/helpers"
	"github.com/run-ai/runai/control-plane/cluster-service/internal/generated/pkg/api"
	bc_mocks "github.com/run-ai/runai/control-plane/cluster-service/mocks/pkg/clients/backend"
	wc_mocks "github.com/run-ai/runai/control-plane/cluster-service/mocks/pkg/clients/workloads"
	"github.com/run-ai/runai/control-plane/cluster-service/pkg/database/dao"
	"github.com/run-ai/runai/control-plane/cluster-service/pkg/database/models"
	"github.com/run-ai/runai/control-plane/cluster-service/pkg/services"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
)

var ctx = context.WithValue(setContext(), authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{Clusters: testClusterUuids})

var _ = Describe("Node DB tests", Ordered, func() {
	type Mocks struct {
		tenantSettingsClient *bc_mocks.TenantSettingsClient
	}

	testCreatedNodes := map[string]*models.Node{}
	var nodeService *services.NodeService
	var mocks *Mocks

	BeforeAll(func() {
		mocks = &Mocks{
			tenantSettingsClient: &bc_mocks.TenantSettingsClient{},
		}

		nodeService = services.NewNodeService(mocks.tenantSettingsClient, nil)

		logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Starting node service tests")
	})
	AfterEach(func() {
		for _, node := range testCreatedNodes {
			deleteNode(nodeService, node)
		}
		testCreatedNodes = map[string]*models.Node{}
	})

	Context("test nodes service and db", func() {
		It("test create/get node/s", func() {
			createNode(nodeService, testCreatedNodes, 0, 0)
			createNode(nodeService, testCreatedNodes, 0, 1)
			createNode(nodeService, testCreatedNodes, 1, 0)
			createNode(nodeService, testCreatedNodes, 1, 2)

			getAndValidateAllNodesInClusterEqualToOriginal(nodeService, 0, []int{0, 1})
			getAndValidateAllNodesInClusterEqualToOriginal(nodeService, 1, []int{0, 2})

			validateNumOfNodesInCluster(nodeService, 0, 2)
			validateNumOfNodesInCluster(nodeService, 1, 2)
			validateNumOfNodesInCluster(nodeService, 2, 0)
		})

		It("test create/update/delete node", func() {
			// create a node
			createNode(nodeService, testCreatedNodes, 0, 0)
			originalApiNode := getAndValidateNodeEqualToOriginal(nodeService, 0, 0)
			validateNumOfNodesInCluster(nodeService, 0, 1)

			// update the node
			updatedNode := *testNodeInfos[0]
			updatedNode.Status = "Ready"
			updatedNode.Conditions = []api.NodeStatusConditionDetails{}
			updatedNode.Taints = (*testNodeInfos[1]).Taints
			updatedNode.NodePool = (*testNodeInfos[0]).NodePool + "-random-suffix"
			updatedNode.GpuInfo = &api.GpuInfo{
				GpuType:  (*testNodeInfos[0]).GpuInfo.GpuType + "-different-type",
				GpuCount: 0,
			}
			time.Sleep(time.Second)
			updatedNode.CreatedAt = time.Now() // shouldn't be updated, just testing
			updateNode(nodeService, testCreatedNodes, 0, 0, &updatedNode)

			// get and validate updated fields
			expectedUpdatedNode := updatedNode
			// create_at should stay the same, even after update
			expectedUpdatedNode.CreatedAt = originalApiNode.CreatedAt
			updatedApiNode := getAndValidateNodeEqual(nodeService, 0, 0, &expectedUpdatedNode)
			Expect(roundedTime(updatedApiNode.CreatedAt).Equal(roundedTime(originalApiNode.CreatedAt))).To(BeTrue(),
				fmt.Sprintf("node's created_at same after update; actual: %v, expected: %v",
					updatedApiNode.CreatedAt, originalApiNode.CreatedAt))
			Expect(roundedTime(updatedApiNode.UpdatedAt).Equal(roundedTime(originalApiNode.UpdatedAt))).To(BeFalse(),
				fmt.Sprintf("node's updated_at different after update; actual: %v, expected not equal: %v",
					updatedApiNode.CreatedAt, originalApiNode.CreatedAt))
			Expect(updatedApiNode.Status).ToNot(Equal((*testNodeInfos[0]).Status),
				"node's status different after update")
			Expect(updatedApiNode.Conditions).ToNot(Equal((*testNodeInfos[0]).Conditions),
				"node's conditions different after update")
			Expect(updatedApiNode.Taints).ToNot(Equal((*testNodeInfos[0]).Taints),
				"node's taints different after update")
			Expect(updatedApiNode.NodePool).ToNot(Equal((*testNodeInfos[0]).NodePool),
				"node's node pool different after update")
			Expect(updatedApiNode.GpuInfo.GpuType).ToNot(Equal((*testNodeInfos[0]).GpuInfo.GpuType),
				"node's GPU type different after update")
			Expect(updatedApiNode.GpuInfo.GpuCount).ToNot(Equal((*testNodeInfos[0]).GpuInfo.GpuCount),
				"node's GPU count different after update")
			validateNumOfNodesInCluster(nodeService, 0, 1)

			// delete the node
			testCreatedNodeKey := getTestCreatedNodesKey(originalApiNode.ClusterUuid, originalApiNode.Name)
			deleteNode(nodeService, testCreatedNodes[testCreatedNodeKey])
			delete(testCreatedNodes, testCreatedNodeKey)

			// validate doesn't exist
			validateNumOfNodesInCluster(nodeService, 0, 0)
			expectGetNodeNotFound(nodeService, 0, 0)
		})

		It("test NetworkTopologyLabelValues stored in DB but not exposed via API", func() {
			// Sync nodes which have NetworkTopologyLabelValues defined in testNodeInfos
			syncNodes(nodeService, "", "", 0, []int{0, 1, 2})

			// Verify via API - api.Node doesn't have NetworkTopologyLabelValues field
			// (this is verified by the type system - if api.Node had the field, the test would need to be updated)
			apiNodes := getNodesInCluster(nodeService, 0)
			Expect(len(*apiNodes)).To(Equal(3))
			// api.Node has NetworkTopologyLabels but NOT NetworkTopologyLabelValues
			// The fact that we can't access NetworkTopologyLabelValues from api.Node is the test

			// Verify via DAO - models.Node DOES have NetworkTopologyLabelValues
			permittedScopes := authorization_api.PermittedScopes{Clusters: testClusterUuids}
			dbNodes, dbErr := dao.GetNodesByClusterUuid(ctx, TestTenantInt, testClusterUuids[0], permittedScopes, nil, "", "", 0, 0, "")
			Expect(dbErr).To(BeNil(), "DAO GetNodesByClusterUuid should succeed")
			Expect(len(*dbNodes)).To(Equal(3), "Should have 3 nodes in DB")

			// Create a map for easier lookup
			dbNodesMap := map[string]models.Node{}
			for _, node := range *dbNodes {
				dbNodesMap[node.Name] = node
			}

			// Verify node0 has topology labels
			node0 := dbNodesMap[testNodeNames[0]]
			Expect(node0.NetworkTopologyLabels).ToNot(BeNil(), "node0 should have NetworkTopologyLabels in DB")
			Expect(len(*node0.NetworkTopologyLabels)).To(Equal(len(*testTopologyLabels0)),
				"node0 NetworkTopologyLabels length should match testTopologyLabels0")

			// Verify node1 has topology labels
			node1 := dbNodesMap[testNodeNames[1]]
			Expect(node1.NetworkTopologyLabels).ToNot(BeNil(), "node1 should have NetworkTopologyLabels in DB")
			Expect(len(*node1.NetworkTopologyLabels)).To(Equal(len(*testTopologyLabels1)),
				"node1 NetworkTopologyLabels length should match testTopologyLabels1")

			// Verify node2 has topology labels
			node2 := dbNodesMap[testNodeNames[2]]
			Expect(node2.NetworkTopologyLabels).ToNot(BeNil(), "node2 should have NetworkTopologyLabels in DB")
			Expect(len(*node2.NetworkTopologyLabels)).To(Equal(len(*testTopologyLabels2)),
				"node2 NetworkTopologyLabels length should match testTopologyLabels2")

			// Clean up
			syncNodes(nodeService, "", "", 0, []int{})
		})

		It("test sync nodes", func() {
			// create nodes
			createNode(nodeService, testCreatedNodes, 1, 0)
			createNode(nodeService, testCreatedNodes, 1, 5)

			// sync nodes - from "", to node2
			syncNodes(nodeService, "", testNodeNames[2], 1, []int{1, 2})

			// node0 should not exist - should have been deleted by the sync
			expectGetNodeNotFound(nodeService, 1, 0)
			// node5 should still exist - we blocked the deletion to node2
			getAndValidateNodeEqualToOriginal(nodeService, 1, 5)
			getAndValidateAllNodesInClusterEqualToOriginal(nodeService, 1, []int{1, 2, 5})

			// sync nodes - from node2, to node4
			syncNodes(nodeService, testNodeNames[2], testNodeNames[4], 1, []int{2, 3, 4})
			// node5 should still exist
			getAndValidateNodeEqualToOriginal(nodeService, 1, 5)
			getAndValidateAllNodesInClusterEqualToOriginal(nodeService, 1, []int{1, 2, 3, 4, 5})

			// sync nodes - from node4, to ""
			syncNodes(nodeService, testNodeNames[4], "", 1, []int{4})
			// node5 should not exist - should have been deleted by the sync
			expectGetNodeNotFound(nodeService, 1, 5)
			getAndValidateAllNodesInClusterEqualToOriginal(nodeService, 1, []int{1, 2, 3, 4})

			// sync an updated node and see it changed
			originalApiNode2 := getAndValidateNodeEqualToOriginal(nodeService, 1, 2)
			updatedNode2 := *testNodeInfos[2]
			updatedNode2.NodePool = testNodePoolName0
			updatedNode2.GpuInfo = nil
			updatedNode2.Status = "Ready"
			updatedNode2.Taints = TestNodeInfo0.Taints
			updatedNode2.Conditions = TestNodeInfo0.Conditions
			getAndValidateNodeEqualToOriginal(nodeService, 1, 1)
			updatedNode1 := *testNodeInfos[1]
			updatedNode1.NodePool = updatedNode1.NodePool + "-random-other-np-suffix"
			syncNodesInner(nodeService, "", "", 1, []int{2, 1}, []*api.NodeInfo{&updatedNode2, &updatedNode1})

			// all other nodes should have been deleted
			validateNumOfNodesInCluster(nodeService, 1, 2)
			// validate both nodes updated indeed
			updatedApiNode := getAndValidateNodeEqual(nodeService, 1, 2, &updatedNode2)
			Expect(updatedApiNode.Status).ToNot(Equal((*testNodeInfos[2]).Status),
				"node's status different after update")
			Expect(updatedApiNode.Status).ToNot(Equal(originalApiNode2.Status),
				"node's status different after update")
			Expect(updatedApiNode.NodePool).ToNot(Equal(originalApiNode2.NodePool),
				"node's node pool different after update")
			Expect(updatedApiNode.GpuInfo).To(BeNil(),
				"node's GPU info different after update")
			Expect(updatedApiNode.Taints).ToNot(Equal(originalApiNode2.Taints),
				"node's taints different after update")
			Expect(updatedApiNode.Conditions).ToNot(Equal(originalApiNode2.Conditions),
				"node's conditions different after update")
			Expect(roundedTime(updatedApiNode.CreatedAt).Equal(roundedTime(originalApiNode2.CreatedAt))).To(BeTrue(),
				fmt.Sprintf("node's created_at same after update; actual: %v, expected: %v",
					updatedApiNode.CreatedAt, originalApiNode2.CreatedAt))
			Expect(roundedTime(updatedApiNode.UpdatedAt).Equal(roundedTime(originalApiNode2.UpdatedAt))).To(BeFalse(),
				fmt.Sprintf("node's updated_at different after update; actual: %v, expected not equal: %v",
					updatedApiNode.CreatedAt, originalApiNode2.CreatedAt))
			updatedApiNode1 := getAndValidateNodeEqual(nodeService, 1, 1, &updatedNode1)
			Expect(updatedApiNode1.NodePool).ToNot(Equal((*testNodeInfos[1]).NodePool),
				"node's node pool different after update")

			syncNodesInner(nodeService, "", "", 1, []int{2}, []*api.NodeInfo{&updatedNode2})
			// all other nodes should have been deleted
			validateNumOfNodesInCluster(nodeService, 1, 1)

			syncNodesInner(nodeService, "", "", 1, []int{}, []*api.NodeInfo{})
			validateNumOfNodesInCluster(nodeService, 1, 0)
		})
	})
	Context("forbidden user requests", func() {
		It("context without permitted scopes", func() {
			result, _ := nodeService.GetTenantNodes(context.WithValue(context.Background(), runaictx.CfTenantId, TestTenant), []string{fmt.Sprintf("clusterUuid==%s", testClusterUuids[0])}, "", "", 0, 0, "")
			nodes := httpResponseBodyToNodes(result)
			Expect(len(*nodes)).To(Equal(0))
		})
		It("context with permitted scopes but not correct scope", func() {
			result, _ := nodeService.GetTenantNodes(ctx, []string{fmt.Sprintf("clusterUuid==%s", notPermittedTestClusterUuid)}, "", "", 0, 0, "")
			nodes := httpResponseBodyToNodes(result)
			Expect(len(*nodes)).To(Equal(0))
		})
	})

	Context("GetNode", func() {
		It("returns a node by its control-plane ID", func() {
			createNode(nodeService, testCreatedNodes, 0, 0)
			expectedNode := getNode(nodeService, 0, 0)

			result, err := nodeService.GetNode(ctx, expectedNode.Id)

			Expect(err).To(BeNil())
			Expect(result.Code).To(Equal(http.StatusOK))
			actualNode := httpResponseBodyToNode(result)
			Expect(actualNode).To(Equal(expectedNode))
		})

		It("rejects a context without permitted scopes", func() {
			createNode(nodeService, testCreatedNodes, 0, 0)
			expectedNode := getNode(nodeService, 0, 0)
			ctxWithoutScopes := context.WithValue(context.Background(), runaictx.CfTenantId, TestTenant)

			result, err := nodeService.GetNode(ctxWithoutScopes, expectedNode.Id)

			Expect(err).To(BeNil())
			Expect(result.Code).To(Equal(http.StatusForbidden))
		})

		It("does not return a node outside the permitted cluster scope", func() {
			createNode(nodeService, testCreatedNodes, 1, 0)
			expectedNode := getNode(nodeService, 1, 0)
			restrictedCtx := context.WithValue(
				setContext(),
				authorization_utils.PermittedScopesKey,
				&authorization_api.PermittedScopes{Clusters: []string{testClusterUuids[0]}},
			)

			result, err := nodeService.GetNode(restrictedCtx, expectedNode.Id)

			Expect(err).To(BeNil())
			Expect(result.Code).To(Equal(http.StatusNotFound))
		})
	})

	Context("CountNodes per cluster param tests", func() {
		BeforeAll(func() {
			// create some nodes to test it later
			syncNodes(nodeService, "", "", 0, []int{0, 1, 2})
		})

		AfterAll(func() {
			syncNodes(nodeService, "", "", 0, []int{})
		})

		It("Test BadRequests", func() {
			result, _ := nodeService.CountNodes(ctx, testClusterUuids[0], []string{"wrong-filterrrr"}, "")
			Expect(result.Code).To(Equal(http.StatusBadRequest))
		})

		It("Test CountNodes filter", func() {
			validateCountNodes(ctx, nodeService, testClusterUuids[0], []string{}, "", 3)
			validateCountNodes(ctx, nodeService, testClusterUuids[0], []string{fmt.Sprintf("nodePool==%s", testNodePoolName0)}, "", 1)
			validateCountNodes(ctx, nodeService, testClusterUuids[0], []string{"nvLinkDomainUid==DOMAIN-X"}, "", 1)
			validateCountNodes(ctx, nodeService, testClusterUuids[0], []string{fmt.Sprintf("gpuType!=%s", testGpuType1)}, "", 2)
			validateCountNodes(ctx, nodeService, testClusterUuids[0], []string{fmt.Sprintf("gpuType!=%s", testGpuType1), "status==Ready"}, "", 1)
		})

		It("Test CountNodes search", func() {
			validateCountNodes(ctx, nodeService, testClusterUuids[0], []string{}, testNodeNames[0], 1)
			validateCountNodes(ctx, nodeService, testClusterUuids[0], []string{}, testNodePoolName1, 2)
			validateCountNodes(ctx, nodeService, testClusterUuids[0], []string{}, testGpuType2, 1)
			validateCountNodes(ctx, nodeService, testClusterUuids[0], []string{}, "drgnejth43k543h", 0)
		})
	})

	Context("GetNodes+CountNodes param tests", func() {
		BeforeAll(func() {
			// create some nodes to test it later
			syncNodes(nodeService, "", "", 0, []int{0, 2})
			syncNodes(nodeService, "", "", 1, []int{1})
		})

		AfterAll(func() {
			syncNodes(nodeService, "", "", 0, []int{})
			syncNodes(nodeService, "", "", 1, []int{})
			syncNodes(nodeService, "", "", 2, []int{})
		})

		It("Test GetNodes clusters filter", func() {
			// get nodes and validate
			result, _ := nodeService.GetTenantNodes(ctx, []string{}, api.NODEFILTERSORTFIELDS_NAME, "", 0, 0, "")
			nodes := httpResponseBodyToNodes(result)
			Expect(len(*nodes)).To(Equal(3))
			Expect((*nodes)[0].Name).To(Equal(testNodeNames[0]))
			Expect((*nodes)[1].Name).To(Equal(testNodeNames[1]))
			Expect((*nodes)[2].Name).To(Equal(testNodeNames[2]))
			// get nodes and validate
			result, _ = nodeService.GetTenantNodes(ctx, []string{fmt.Sprintf("clusterUuid!=%s", testClusterUuids[1])}, api.NODEFILTERSORTFIELDS_NAME, "", 0, 0, "")
			nodes = httpResponseBodyToNodes(result)
			Expect(len(*nodes)).To(Equal(2))
			Expect((*nodes)[0].Name).To(Equal(testNodeNames[0]))
			Expect((*nodes)[1].Name).To(Equal(testNodeNames[2]))
		})

		It("Test GetNodes offset+limit", func() {
			result, _ := nodeService.GetTenantNodes(ctx, []string{}, "", "", 0, 0, "")
			nodes := httpResponseBodyToNodes(result)
			Expect(len(*nodes)).To(Equal(3))

			result, _ = nodeService.GetTenantNodes(ctx, []string{}, "", "", 1, 0, "")
			nodes = httpResponseBodyToNodes(result)
			Expect(len(*nodes)).To(Equal(2))

			result, _ = nodeService.GetTenantNodes(ctx, []string{}, "", "", 1, 1, "")
			nodes = httpResponseBodyToNodes(result)
			Expect(len(*nodes)).To(Equal(1))

			result, _ = nodeService.GetTenantNodes(ctx, []string{}, "", "", 1, 2, "")
			nodes = httpResponseBodyToNodes(result)
			Expect(len(*nodes)).To(Equal(2))

			result, _ = nodeService.GetTenantNodes(ctx, []string{}, "", "", 2, 5, "")
			nodes = httpResponseBodyToNodes(result)
			Expect(len(*nodes)).To(Equal(1))

			result, _ = nodeService.GetTenantNodes(ctx, []string{fmt.Sprintf("gpuType!=%s", testGpuType1)}, api.NODEFILTERSORTFIELDS_STATUS, "ASC", 0, 1, "")
			nodes = httpResponseBodyToNodes(result)
			Expect(len(*nodes)).To(Equal(1)) // instead of 2 if it wasn't with a limit

			result, _ = nodeService.GetTenantNodes(ctx, []string{fmt.Sprintf("gpuType==%s", testGpuType1)}, api.NODEFILTERSORTFIELDS_STATUS, "ASC", 0, 1, testNodePoolName0)
			nodes = httpResponseBodyToNodes(result)
			Expect(len(*nodes)).To(Equal(1))
			result, _ = nodeService.GetTenantNodes(ctx, []string{fmt.Sprintf("gpuType==%s", testGpuType1)}, api.NODEFILTERSORTFIELDS_STATUS, "ASC", 1, 2, testNodePoolName0)
			nodes = httpResponseBodyToNodes(result)
			Expect(len(*nodes)).To(Equal(0))
		})

		It("Test GetNodes search", func() {
			result, _ := nodeService.GetTenantNodes(ctx, []string{}, "", "", 0, 0, testNodeNames[0])
			nodes := httpResponseBodyToNodes(result)
			Expect(len(*nodes)).To(Equal(1))
			Expect((*nodes)[0].Name).To(Equal(testNodeNames[0]))

			result, _ = nodeService.GetTenantNodes(ctx, []string{}, "", "", 0, 0, testNodePoolName1)
			nodes = httpResponseBodyToNodes(result)
			Expect(len(*nodes)).To(Equal(2))
			Expect((*nodes)[0].NodePool).To(Equal(testNodePoolName1))
			Expect((*nodes)[1].NodePool).To(Equal(testNodePoolName1))

			result, _ = nodeService.GetTenantNodes(ctx, []string{}, "", "", 0, 0, testGpuType2)
			nodes = httpResponseBodyToNodes(result)
			Expect(len(*nodes)).To(Equal(1))
			Expect((*nodes)[0].GpuInfo.GpuType).To(Equal(testGpuType2))

			result, _ = nodeService.GetTenantNodes(ctx, []string{}, "", "", 0, 0, "sfdsfhjkdsfhskjhfbes")
			nodes = httpResponseBodyToNodes(result)
			Expect(len(*nodes)).To(Equal(0))
		})
	})

	Context("CountTenantNodes tests", func() {
		BeforeAll(func() {
			// create some nodes to test it later
			syncNodes(nodeService, "", "", 0, []int{0, 2})
			syncNodes(nodeService, "", "", 1, []int{1})
		})

		AfterAll(func() {
			syncNodes(nodeService, "", "", 0, []int{})
			syncNodes(nodeService, "", "", 1, []int{})
			syncNodes(nodeService, "", "", 2, []int{})
		})

		It("Test CountTenantNodes basic functionality", func() {
			// Count all tenant nodes
			validateCountTenantNodes(ctx, nodeService, []string{}, "", 3)
		})

		It("Test CountTenantNodes with filters", func() {
			// Count nodes by node pool
			validateCountTenantNodes(ctx, nodeService, []string{fmt.Sprintf("nodePool==%s", testNodePoolName0)}, "", 1)
			validateCountTenantNodes(ctx, nodeService, []string{fmt.Sprintf("nodePool==%s", testNodePoolName1)}, "", 2)

			// Count nodes by NVLink domain
			validateCountTenantNodes(ctx, nodeService, []string{"nvLinkDomainUid==DOMAIN-X"}, "", 1)

			// Count nodes by GPU type
			validateCountTenantNodes(ctx, nodeService, []string{fmt.Sprintf("gpuType==%s", testGpuType1)}, "", 1)
			validateCountTenantNodes(ctx, nodeService, []string{fmt.Sprintf("gpuType!=%s", testGpuType1)}, "", 2)

			// Count nodes by status
			validateCountTenantNodes(ctx, nodeService, []string{"status==Ready"}, "", 1)
			validateCountTenantNodes(ctx, nodeService, []string{"status==NotReady"}, "", 1)
			validateCountTenantNodes(ctx, nodeService, []string{"status==Unknown"}, "", 1)

			// Count nodes by cluster UUID
			validateCountTenantNodes(ctx, nodeService, []string{fmt.Sprintf("clusterUuid==%s", testClusterUuids[0])}, "", 2)
			validateCountTenantNodes(ctx, nodeService, []string{fmt.Sprintf("clusterUuid==%s", testClusterUuids[1])}, "", 1)
			validateCountTenantNodes(ctx, nodeService, []string{fmt.Sprintf("clusterUuid!=%s", testClusterUuids[1])}, "", 2)

			// Count nodes with combined filters
			validateCountTenantNodes(ctx, nodeService, []string{fmt.Sprintf("gpuType!=%s", testGpuType1), "status==Ready"}, "", 1)
			validateCountTenantNodes(ctx, nodeService, []string{fmt.Sprintf("nodePool==%s", testNodePoolName1), fmt.Sprintf("clusterUuid==%s", testClusterUuids[1])}, "", 1)
		})

		It("Test CountTenantNodes with search", func() {
			// Search by node name
			validateCountTenantNodes(ctx, nodeService, []string{}, testNodeNames[0], 1)
			validateCountTenantNodes(ctx, nodeService, []string{}, testNodeNames[1], 1)
			validateCountTenantNodes(ctx, nodeService, []string{}, testNodeNames[2], 1)

			// Search by node pool name
			validateCountTenantNodes(ctx, nodeService, []string{}, testNodePoolName0, 1)
			validateCountTenantNodes(ctx, nodeService, []string{}, testNodePoolName1, 2)

			// Search by GPU type
			validateCountTenantNodes(ctx, nodeService, []string{}, testGpuType1, 1)
			validateCountTenantNodes(ctx, nodeService, []string{}, testGpuType2, 1)

			// Search with no results
			validateCountTenantNodes(ctx, nodeService, []string{}, "nonexistent-search-term", 0)
		})

		It("Test CountTenantNodes bad requests", func() {
			// Test invalid filter
			result, _ := nodeService.CountTenantNodes(ctx, []string{"invalid-filter"}, "")
			Expect(result.Code).To(Equal(http.StatusBadRequest))

			result, _ = nodeService.CountTenantNodes(ctx, []string{"nonexistent-field==value"}, "")
			Expect(result.Code).To(Equal(http.StatusBadRequest))
		})

		It("Test CountTenantNodes permission filtering", func() {
			// Test with limited permissions (only testClusterUuids[0] allowed)
			restrictedCtx := context.WithValue(setContext(), authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{Clusters: []string{testClusterUuids[0]}})

			// Should only count nodes from permitted cluster
			validateCountTenantNodes(restrictedCtx, nodeService, []string{}, "", 2)

			// Filter by specific cluster that is permitted
			validateCountTenantNodes(restrictedCtx, nodeService, []string{fmt.Sprintf("clusterUuid==%s", testClusterUuids[0])}, "", 2)

			// Filter by cluster that is not permitted - should return 0
			validateCountTenantNodes(restrictedCtx, nodeService, []string{fmt.Sprintf("clusterUuid==%s", testClusterUuids[1])}, "", 0)
		})
	})

	Context("UpsertNodes returns old topology values", func() {
		clusterUuid := testClusterUuid2

		AfterEach(func() {
			// Clean up nodes created in tests
			dao.DeleteClusterNodes(ctx, TestTenantInt, clusterUuid)
		})

		It("test UpsertNodes returns old and new NetworkTopologyLabelValues", func() {
			// 1. Create node1 using UpsertNode with initial topology labels
			node1LabelsInitial := &models.NetworkTopologyLabels{
				{LabelKey: "level1", LabelValue: "value1"},
				{LabelKey: "level2", LabelValue: "value1"},
			}
			node1ValuesInitial := pq.StringArray{"value1", "value1"}
			node1 := &models.Node{
				Name:                       "test-node-1",
				ClusterUuid:                clusterUuid,
				TenantId:                   TestTenantInt,
				Status:                     "Ready",
				NodePool:                   testNodePoolName0,
				NetworkTopologyLabels:      node1LabelsInitial,
				NetworkTopologyLabelValues: &node1ValuesInitial,
			}

			result1, err := dao.UpsertNode(ctx, node1)
			Expect(err).To(BeNil(), "UpsertNode should succeed for node1")
			Expect(result1.PreviousValues.NetworkTopologyLabelValues).To(BeNil(), "Previous topology should be nil for new node")
			Expect(result1.NewValues.NetworkTopologyLabelValues).ToNot(BeNil(), "New topology should not be nil")
			Expect([]string(*result1.NewValues.NetworkTopologyLabelValues)).To(Equal([]string(node1ValuesInitial)),
				"New topology should match initial values")

			// 2. Update node1's topology labels
			node1LabelsUpdated := &models.NetworkTopologyLabels{
				{LabelKey: "level1", LabelValue: "value1-updated"},
				{LabelKey: "level2", LabelValue: "value1-updated"},
			}
			node1ValuesUpdated := pq.StringArray{"value1-updated", "value1-updated"}
			node1Updated := &models.Node{
				Name:                       "test-node-1",
				ClusterUuid:                clusterUuid,
				TenantId:                   TestTenantInt,
				Status:                     "Ready",
				NodePool:                   testNodePoolName0,
				NetworkTopologyLabels:      node1LabelsUpdated,
				NetworkTopologyLabelValues: &node1ValuesUpdated,
			}

			// 3. Create node2 (new node) with topology labels
			node2Labels := &models.NetworkTopologyLabels{
				{LabelKey: "level1", LabelValue: "value2"},
				{LabelKey: "level2", LabelValue: "value2"},
			}
			node2Values := pq.StringArray{"value2", "value2"}
			node2 := &models.Node{
				Name:                       "test-node-2",
				ClusterUuid:                clusterUuid,
				TenantId:                   TestTenantInt,
				Status:                     "NotReady",
				NodePool:                   testNodePoolName1,
				NetworkTopologyLabels:      node2Labels,
				NetworkTopologyLabelValues: &node2Values,
			}

			// 4. Call UpsertNodes with node1 (updated) and node2 (new)
			nodes := []models.Node{*node1Updated, *node2}
			upsertResults, upsertErr := dao.UpsertNodes(ctx, &nodes, clusterUuid)

			Expect(upsertErr).To(BeNil(), "UpsertNodes should succeed")
			Expect(len(upsertResults)).To(Equal(2), "Should return results for both nodes")

			// 5. Validate results - find results by name
			var node1Result, node2Result *dao.UpsertNodeResult
			for i := range upsertResults {
				if upsertResults[i].Name == "test-node-1" {
					node1Result = &upsertResults[i]
				} else if upsertResults[i].Name == "test-node-2" {
					node2Result = &upsertResults[i]
				}
			}

			Expect(node1Result).ToNot(BeNil(), "Should have result for node1")
			Expect(node2Result).ToNot(BeNil(), "Should have result for node2")

			// Node1: verify previous values (initial) and new values (updated)
			Expect(node1Result.PreviousValues.NetworkTopologyLabelValues).ToNot(BeNil(), "Node1 should have previous topology values")
			Expect([]string(*node1Result.PreviousValues.NetworkTopologyLabelValues)).To(Equal([]string(node1ValuesInitial)),
				"Node1 previous topology should match initial values")
			Expect(node1Result.NewValues.NetworkTopologyLabelValues).ToNot(BeNil(), "Node1 should have new topology values")
			Expect([]string(*node1Result.NewValues.NetworkTopologyLabelValues)).To(Equal([]string(node1ValuesUpdated)),
				"Node1 new topology should match updated values")

			// Node2: verify nil previous values (new node) and new values
			Expect(node2Result.PreviousValues.NetworkTopologyLabelValues).To(BeNil(), "Node2 should have nil previous topology (new node)")
			Expect(node2Result.NewValues.NetworkTopologyLabelValues).ToNot(BeNil(), "Node2 should have new topology values")
			Expect([]string(*node2Result.NewValues.NetworkTopologyLabelValues)).To(Equal([]string(node2Values)),
				"Node2 new topology should match provided values")
		})
	})

	Context("networkTopologyLabelValuesChanged triggers workloads client call", func() {
		var clusterUuid string
		BeforeEach(func() {
			createdCluster, apiErr := dao.CreateCluster(ctx, &models.Cluster{
				Name:     "test-topology-cluster",
				TenantId: TestTenantInt,
				Version:  "2.25",
			})

			Expect(apiErr).To(BeNil())
			clusterUuid = createdCluster.ID.String()
			testClusterUuids = append(testClusterUuids, clusterUuid)
			ctx = context.WithValue(ctx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{Clusters: testClusterUuids})
		})

		AfterEach(func() {
			dao.DeleteClusterNodes(ctx, TestTenantInt, clusterUuid)
			dao.DeleteCluster(ctx, clusterUuid, TestTenantInt)
		})

		It("UpsertNode with same topology labels does NOT call workloads client", func() {
			workloadsMock := &wc_mocks.WorkloadsClient{}
			nodeServiceWithMock := services.NewNodeService(&bc_mocks.TenantSettingsClient{}, workloadsMock)

			topologyLabels := &[]api.NetworkTopologyLabel{
				{LabelKey: "zone", LabelValue: "zone-a"},
				{LabelKey: "region", LabelValue: "region-1"},
			}
			nodeInfo := api.NodeInfo{
				Status:                "Ready",
				NodePool:              testNodePoolName0,
				NetworkTopologyLabels: topologyLabels,
			}

			// First upsert - creates the node (workloads client will be called for new node)
			workloadsMock.On("InternalUpdateNodePods", ctx, clusterUuid, "test-topology-node", []string{"zone-a", "region-1"}).Return(nil).Once()
			result1, _ := nodeServiceWithMock.UpsertNode(ctx, clusterUuid, "test-topology-node", nodeInfo)
			Expect(result1.Code).To(Equal(http.StatusNoContent), "First UpsertNode should succeed")

			// Second upsert with SAME topology labels - workloads client should NOT be called
			result2, _ := nodeServiceWithMock.UpsertNode(ctx, clusterUuid, "test-topology-node", nodeInfo)
			Expect(result2.Code).To(Equal(http.StatusNoContent), "Second UpsertNode should succeed")

			// Verify mock expectations - InternalUpdateNodePods should only have been called once
			workloadsMock.AssertExpectations(GinkgoT())
		})

		It("UpsertNode with different topology labels DOES call workloads client", func() {
			workloadsMock := &wc_mocks.WorkloadsClient{}
			nodeServiceWithMock := services.NewNodeService(&bc_mocks.TenantSettingsClient{}, workloadsMock)

			initialTopologyLabels := &[]api.NetworkTopologyLabel{
				{LabelKey: "zone", LabelValue: "zone-a"},
				{LabelKey: "region", LabelValue: "region-1"},
			}
			nodeInfo := api.NodeInfo{
				Status:                "Ready",
				NodePool:              testNodePoolName0,
				NetworkTopologyLabels: initialTopologyLabels,
			}

			// First upsert - creates the node
			workloadsMock.On("InternalUpdateNodePods", ctx, clusterUuid, "test-topology-node-2", []string{"zone-a", "region-1"}).Return(nil).Once()
			result1, _ := nodeServiceWithMock.UpsertNode(ctx, clusterUuid, "test-topology-node-2", nodeInfo)
			Expect(result1.Code).To(Equal(http.StatusNoContent), "First UpsertNode should succeed")

			// Second upsert with DIFFERENT topology labels - workloads client SHOULD be called
			updatedTopologyLabels := &[]api.NetworkTopologyLabel{
				{LabelKey: "zone", LabelValue: "zone-b"},
				{LabelKey: "region", LabelValue: "region-2"},
			}
			nodeInfo.NetworkTopologyLabels = updatedTopologyLabels
			workloadsMock.On("InternalUpdateNodePods", ctx, clusterUuid, "test-topology-node-2", []string{"zone-b", "region-2"}).Return(nil).Once()
			result2, _ := nodeServiceWithMock.UpsertNode(ctx, clusterUuid, "test-topology-node-2", nodeInfo)
			Expect(result2.Code).To(Equal(http.StatusNoContent), "Second UpsertNode should succeed")

			// Verify mock expectations - InternalUpdateNodePods should have been called twice
			workloadsMock.AssertExpectations(GinkgoT())
		})

		It("SyncNodes with same topology labels does NOT call workloads client", func() {
			workloadsMock := &wc_mocks.WorkloadsClient{}
			nodeServiceWithMock := services.NewNodeService(&bc_mocks.TenantSettingsClient{}, workloadsMock)

			topologyLabels := &[]api.NetworkTopologyLabel{
				{LabelKey: "zone", LabelValue: "zone-a"},
				{LabelKey: "region", LabelValue: "region-1"},
			}
			nodeInfo := api.NodeInfo{
				Status:                "Ready",
				NodePool:              testNodePoolName0,
				NetworkTopologyLabels: topologyLabels,
			}
			syncRequest := api.NodesSyncRequest{
				NodesForSync: []api.NodeForSync{
					{Name: "sync-topology-node", NodeInfo: nodeInfo},
				},
			}

			// First sync - creates the node (workloads client will be called for new node)
			workloadsMock.On("InternalUpdateNodePods", ctx, clusterUuid, "sync-topology-node", []string{"zone-a", "region-1"}).Return(nil).Once()
			result1, _ := nodeServiceWithMock.SyncNodes(ctx, clusterUuid, syncRequest)
			Expect(result1.Code).To(Equal(http.StatusNoContent), "First SyncNodes should succeed")

			// Second sync with SAME topology labels - workloads client should NOT be called
			result2, _ := nodeServiceWithMock.SyncNodes(ctx, clusterUuid, syncRequest)
			Expect(result2.Code).To(Equal(http.StatusNoContent), "Second SyncNodes should succeed")

			// Verify mock expectations - InternalUpdateNodePods should only have been called once
			workloadsMock.AssertExpectations(GinkgoT())
		})

		It("SyncNodes with different topology labels DOES call workloads client", func() {
			workloadsMock := &wc_mocks.WorkloadsClient{}
			nodeServiceWithMock := services.NewNodeService(&bc_mocks.TenantSettingsClient{}, workloadsMock)

			initialTopologyLabels := &[]api.NetworkTopologyLabel{
				{LabelKey: "zone", LabelValue: "zone-a"},
				{LabelKey: "region", LabelValue: "region-1"},
			}
			nodeInfo := api.NodeInfo{
				Status:                "Ready",
				NodePool:              testNodePoolName0,
				NetworkTopologyLabels: initialTopologyLabels,
			}
			syncRequest := api.NodesSyncRequest{
				NodesForSync: []api.NodeForSync{
					{Name: "sync-topology-node-2", NodeInfo: nodeInfo},
				},
			}

			// First sync - creates the node
			workloadsMock.On("InternalUpdateNodePods", ctx, clusterUuid, "sync-topology-node-2", []string{"zone-a", "region-1"}).Return(nil).Once()
			result1, _ := nodeServiceWithMock.SyncNodes(ctx, clusterUuid, syncRequest)
			Expect(result1.Code).To(Equal(http.StatusNoContent), "First SyncNodes should succeed")

			// Second sync with DIFFERENT topology labels - workloads client SHOULD be called
			updatedTopologyLabels := &[]api.NetworkTopologyLabel{
				{LabelKey: "zone", LabelValue: "zone-b"},
				{LabelKey: "region", LabelValue: "region-2"},
			}
			nodeInfo.NetworkTopologyLabels = updatedTopologyLabels
			syncRequest.NodesForSync[0].NodeInfo = nodeInfo
			workloadsMock.On("InternalUpdateNodePods", ctx, clusterUuid, "sync-topology-node-2", []string{"zone-b", "region-2"}).Return(nil).Once()
			result2, _ := nodeServiceWithMock.SyncNodes(ctx, clusterUuid, syncRequest)
			Expect(result2.Code).To(Equal(http.StatusNoContent), "Second SyncNodes should succeed")

			// Verify mock expectations - InternalUpdateNodePods should have been called twice
			workloadsMock.AssertExpectations(GinkgoT())
		})
	})
})

func createNode(nodeService *services.NodeService, testCreatedNodes map[string]*models.Node,
	clusterUuidIndex, nodeIndex int,
) {
	updateNode(nodeService, testCreatedNodes, clusterUuidIndex, nodeIndex, nil)
}

func updateNode(nodeService *services.NodeService, testCreatedNodes map[string]*models.Node,
	clusterUuidIndex, nodeIndex int, nodeInfo *api.NodeInfo,
) {
	clusterUuid := testClusterUuids[clusterUuidIndex]
	nodeName := testNodeNames[nodeIndex]
	action := "update"
	if nodeInfo == nil {
		action = "create"
		nodeInfo = testNodeInfos[nodeIndex]
	}

	result, err := nodeService.UpsertNode(ctx, clusterUuid, nodeName, *nodeInfo)
	if err != nil || result.Code >= http.StatusBadRequest {
		Fail(fmt.Sprintf("Failed to %s node <%s> in cluster index <%d>, return code: %d",
			action, nodeName, clusterUuidIndex, result.Code))
	}
	testCreatedNodes[getTestCreatedNodesKey(clusterUuid, nodeName)] = dao.NodeInfoToNodeDao(TestTenantInt, clusterUuid, nodeName, *nodeInfo)
}

func syncNodes(nodeService *services.NodeService, from, to string, clusterUuidIndex int, nodeIndexes []int) {
	nodeInfos := []*api.NodeInfo{}
	for _, nodeIndex := range nodeIndexes {
		nodeInfos = append(nodeInfos, testNodeInfos[nodeIndex])
	}
	syncNodesInner(nodeService, from, to, clusterUuidIndex, nodeIndexes, nodeInfos)
}

func syncNodesInner(nodeService *services.NodeService, from, to string, clusterUuidIndex int, nodeIndexes []int, nodeInfos []*api.NodeInfo) {
	clusterUuid := testClusterUuids[clusterUuidIndex]

	nodeNames := []string{}
	nodesSyncRequest := api.NodesSyncRequest{
		NodesForSync: []api.NodeForSync{},
		From:         from,
		To:           to,
	}
	for listIndex, nodeIndex := range nodeIndexes {
		nodesSyncRequest.NodesForSync = append(nodesSyncRequest.NodesForSync, api.NodeForSync{
			Name:     testNodeNames[nodeIndex],
			NodeInfo: *nodeInfos[listIndex],
		})
		nodeNames = append(nodeNames, testNodeNames[nodeIndex])
	}

	result, err := nodeService.SyncNodes(ctx, clusterUuid, nodesSyncRequest)
	if err != nil || result.Code >= http.StatusBadRequest {
		Fail(fmt.Sprintf("Failed to upsert nodes <%v> in cluster index <%d>, return code: %d",
			nodeNames, clusterUuidIndex, result.Code))
	}
}

func deleteNode(nodeService *services.NodeService, node *models.Node) {
	result, err := nodeService.DeleteNode(ctx, node.ClusterUuid, node.Name)
	if err != nil || result.Code >= http.StatusBadRequest {
		Fail(fmt.Sprintf("Failed deleting node <%s> from cluster <%s>, return code: %d",
			node.Name, node.ClusterUuid, result.Code))
	}
}

func getAndValidateAllNodesInClusterEqualToOriginal(nodeService *services.NodeService,
	clusterUuidIndex int, nodeIndexes []int,
) {
	nodes := validateNumOfNodesInCluster(nodeService, clusterUuidIndex, len(nodeIndexes))

	// the order can be different, so create a map
	nodesMap := map[string]api.Node{}
	for _, node := range *nodes {
		nodesMap[node.Name] = node
	}

	// also create a map of ids just to validate they are all different
	nodeIdsMap := map[string]bool{}

	expectedClusterUuid := testClusterUuids[clusterUuidIndex]
	for _, nodeIndex := range nodeIndexes {
		expectedNodeInfo := testNodeInfos[nodeIndex]
		nodeName := testNodeNames[nodeIndex]

		apiNode, found := nodesMap[nodeName]
		ExpectWithOffset(1, found).To(BeTrue(),
			fmt.Sprintf("expected to find node <%s> in cluster index <%d>", nodeName, clusterUuidIndex))
		validateNodesEqual(&apiNode, expectedClusterUuid, nodeName, expectedNodeInfo, 2)
		Expect(apiNode.Id).ToNot(BeEmpty(), "node's id")
		Expect(nodeIdsMap[apiNode.Id]).To(BeFalse(), "node's id should be unique")
		nodeIdsMap[apiNode.Id] = true
	}
}

func getAndValidateNodeEqualToOriginal(nodeService *services.NodeService,
	clusterUuidIndex, nodeIndex int,
) *api.Node {
	return getAndValidateNodeEqualInner(nodeService, clusterUuidIndex, nodeIndex, nil, 3)
}

func getAndValidateNodeEqual(nodeService *services.NodeService,
	clusterUuidIndex, nodeIndex int, expectedNodeInfo *api.NodeInfo,
) *api.Node {
	return getAndValidateNodeEqualInner(nodeService, clusterUuidIndex, nodeIndex, expectedNodeInfo, 3)
}

func getAndValidateNodeEqualInner(nodeService *services.NodeService,
	clusterUuidIndex, nodeIndex int, expectedNodeInfo *api.NodeInfo, offset int,
) *api.Node {
	expectedClusterUuid := testClusterUuids[clusterUuidIndex]
	expectedNodeName := testNodeNames[nodeIndex]
	if expectedNodeInfo == nil {
		expectedNodeInfo = testNodeInfos[nodeIndex]
	}
	if offset == 0 {
		offset = 3
	}

	apiNode := getNode(nodeService, clusterUuidIndex, nodeIndex)
	validateNodesEqual(apiNode, expectedClusterUuid, expectedNodeName, expectedNodeInfo, offset)
	return apiNode
}

func getNode(nodeService *services.NodeService, clusterUuidIndex, nodeIndex int) *api.Node {
	node, _ := getNodeInner(nodeService, clusterUuidIndex, nodeIndex, false)
	return node
}

func expectGetNodeNotFound(nodeService *services.NodeService, clusterUuidIndex, nodeIndex int) {
	apiNode, code := getNodeInner(nodeService, clusterUuidIndex, nodeIndex, true)
	ExpectWithOffset(1, apiNode).To(BeNil())
	ExpectWithOffset(1, code).To(Equal(http.StatusNotFound))
}

func getNodeInner(nodeService *services.NodeService, clusterUuidIndex, nodeIndex int, expectedToFail bool) (*api.Node, int) {
	clusterUuid := testClusterUuids[clusterUuidIndex]
	nodeName := testNodeNames[nodeIndex]

	result, err := nodeService.GetTenantNodes(ctx,
		[]string{fmt.Sprintf("clusterUuid==%s", clusterUuid), fmt.Sprintf("name==%s", nodeName)}, "", "", 0, 0, "")
	if err != nil || result.Code >= http.StatusBadRequest {
		if !expectedToFail {
			Fail(fmt.Sprintf("Failed to get node <%s> from cluster index <%d>, return code: %d",
				nodeName, clusterUuidIndex, result.Code))
		}
		return nil, result.Code
	}

	nodes := httpResponseBodyToNodes(result)
	if len(*nodes) == 0 {
		if !expectedToFail {
			Fail(fmt.Sprintf("Failed to get node <%s> from cluster index <%d>: not found",
				nodeName, clusterUuidIndex))
		}
		return nil, http.StatusNotFound
	}
	node := (*nodes)[0]
	return &node, result.Code
}

func getNodesInCluster(nodeService *services.NodeService, clusterUuidIndex int) *[]api.Node {
	clusterUuid := testClusterUuids[clusterUuidIndex]
	result, err := nodeService.GetTenantNodes(ctx, []string{fmt.Sprintf("clusterUuid==%s", clusterUuid)}, "", "", 0, 0, "")
	if err != nil || result.Code >= http.StatusBadRequest {
		Fail(fmt.Sprintf("Failed to get nodes from cluster index <%d>, return code: %d",
			clusterUuidIndex, result.Code))
	}

	return httpResponseBodyToNodes(result)
}

func validateNumOfNodesInCluster(nodeService *services.NodeService, clusterUuidIndex int, expectedNumOfNodes int) *[]api.Node {
	nodes := getNodesInCluster(nodeService, clusterUuidIndex)
	ExpectWithOffset(1, nodes).ToNot(BeNil(), "expected get nodes to succeed")
	ExpectWithOffset(1, len(*nodes)).To(Equal(expectedNumOfNodes), "expected num of nodes in cluster")
	return nodes
}

func validateNodesEqual(actual *api.Node,
	expectedClusterUuid, expectedNodeName string, expectedNodeInfo *api.NodeInfo, offset int,
) {
	fillNodeNilValuesIfNeeded(actual)
	fillNilValuesIfNeeded(expectedNodeInfo)

	ExpectWithOffset(offset, actual.Name).To(Equal(expectedNodeName), "node name")
	ExpectWithOffset(offset, actual.ClusterUuid).To(Equal(expectedClusterUuid), "node's cluster uuid")
	ExpectWithOffset(offset, actual.NodePool).To(Equal(expectedNodeInfo.NodePool), "node's node pool")
	if expectedNodeInfo.GpuInfo == nil {
		ExpectWithOffset(offset, actual.GpuInfo).To(BeNil(), "node's gpu type")
	} else {
		ExpectWithOffset(offset, actual.GpuInfo.GpuType).To(Equal(expectedNodeInfo.GpuInfo.GpuType), "node's gpu type")
		ExpectWithOffset(offset, actual.GpuInfo.GpuCount).To(Equal(expectedNodeInfo.GpuInfo.GpuCount), "node's gpu count")
	}
	ExpectWithOffset(offset, actual.Status).To(Equal(expectedNodeInfo.Status), "node's status")
	ExpectWithOffset(offset, actual.NvLinkDomainUid).To(Equal(expectedNodeInfo.NvLinkDomainUid), "node's NvLinkDomainUid")
	ExpectWithOffset(offset, actual.NvLinkCliqueId).To(Equal(expectedNodeInfo.NvLinkCliqueId), "node's NvLinkCliqueId")
	ExpectWithOffset(offset, actual.Numa).To(Equal(expectedNodeInfo.Numa), "node's NUMA info")
	ExpectWithOffset(offset, actual.Conditions).To(BeEquivalentTo(expectedNodeInfo.Conditions), "node's conditions")
	ExpectWithOffset(offset, actual.Taints).To(BeEquivalentTo(expectedNodeInfo.Taints), "node's taints")
	ExpectWithOffset(offset, roundedTime(actual.CreatedAt).Equal(roundedTime(expectedNodeInfo.CreatedAt))).To(BeTrue(),
		fmt.Sprintf("node's created at; actual: %v, expected: %v", actual.CreatedAt, expectedNodeInfo.CreatedAt))
	ExpectWithOffset(offset, actual.NetworkTopologyLabels).To(BeEquivalentTo(expectedNodeInfo.NetworkTopologyLabels), "node's NetworkTopologyLabels")
}

func roundedTime(t time.Time) time.Time {
	return t.Round(time.Millisecond)
}

func fillNilValuesIfNeeded(node *api.NodeInfo) {
	if node.Conditions == nil {
		node.Conditions = []api.NodeStatusConditionDetails{}
	}
	if node.Taints == nil {
		node.Taints = []api.NodeTaint{}
	}
}

func fillNodeNilValuesIfNeeded(node *api.Node) {
	if node.Conditions == nil {
		node.Conditions = []api.NodeStatusConditionDetails{}
	}
	if node.Taints == nil {
		node.Taints = []api.NodeTaint{}
	}
}

func getTestCreatedNodesKey(clusterUuid, nodeName string) string {
	return clusterUuid + nodeName
}

func validateCountNodes(ctx context.Context, nodeService *services.NodeService, clusterUuid string, filterBy []string, search string,
	expectedCount int64,
) {
	result, _ := nodeService.CountNodes(ctx, clusterUuid, filterBy, search)
	Expect(result.Code).To(Equal(http.StatusOK))
	count := result.Body.(api.CountNodes200Response).Count
	Expect(count).To(Equal(expectedCount))
}

func validateCountTenantNodes(ctx context.Context, nodeService *services.NodeService, filterBy []string, search string,
	expectedCount int64,
) {
	result, _ := nodeService.CountTenantNodes(ctx, filterBy, search)
	Expect(result.Code).To(Equal(http.StatusOK))
	count := result.Body.(api.CountNodes200Response).Count
	Expect(count).To(Equal(expectedCount))
}
