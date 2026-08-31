/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tests

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	kaiv1 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/kai/v1"
	kaiconstants "github.com/kai-scheduler/KAI-scheduler/pkg/common/constants"
	"github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/run-ai/runai/runai-cluster/cluster/nodepool-controller/pkg/common"
	"github.com/run-ai/runai/runai-cluster/cluster/nodepool-controller/pkg/config"
	managed_nodes_config "github.com/run-ai/runai/runai-cluster/cluster/nodepool-controller/pkg/managed-nodes-config"
	"github.com/run-ai/runai/runai-cluster/cluster/nodepool-controller/pkg/nodepool_controller"
	"github.com/run-ai/runai/runai-cluster/cluster/nodepool-controller/pkg/nodepool_controller/metrics"
	scheme_pkg "github.com/run-ai/runai/runai-cluster/cluster/nodepool-controller/pkg/scheme"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var scheme *runtime.Scheme

// Managed-nodes vocabulary the suite seeds into config and its fixtures.
const (
	testManagedNodesConfigName = "kai-managed-nodes-config"
	testExcludedNodepoolName   = "kai-excluded-nodes"
	testToExcludeLabel         = "kai.scheduler/to-exclude"
	testUnschedulableLabel     = "kai.scheduler/unschedulable"
)

var _ = BeforeSuite(func() {
	scheme = scheme_pkg.Scheme()
	// The run.ai Cluster CR is read as unstructured by the controller, so the
	// fake client needs its GVK registered to serve the fixture.
	registerClusterCRType(scheme)

	// Seed the package-level config singleton so production code paths (which
	// read label keys / names / finalizer domain via config.Get()) run against
	// the same values the fixtures below assert on.
	config.SetCurrent(&config.NodePoolControllerConfig{
		NodePoolNameLabel:        kaiconstants.DefaultNodePoolLabelKey,
		DefaultNodepoolName:      kaiconstants.DefaultNodePoolName,
		SchedulerName:            "runai-scheduler",
		SchedulerNamespace:       "runai",
		MetricsNamespace:         "runai",
		FinalizerDomain:          "run.ai",
		ManagedNodesConfigName:   testManagedNodesConfigName,
		ExcludedNodepoolName:     testExcludedNodepoolName,
		ShouldBeExcludedLabelKey: testToExcludeLabel,
		UnschedulableLabelKey:    testUnschedulableLabel,
	})
})

func TestNodePoolController(t *testing.T) {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	log.Info().Msgf("Started service tests")
	RegisterFailHandler(Fail)
	RunSpecs(t, "nodepool-controller Test Suite")
}

func createNodesWithMNNVLLabel(label string) map[string]TestNode {
	return map[string]TestNode{
		"kind-worker": {
			Name: "kind-worker",
			Labels: map[string]string{
				"nodePool0Key": "nodePool0Value",
				label:          "DOMAIN-X.CLIQUE-Y",
			},
		},
	}
}

func MNNVLTestCaseBase(gpuNetworkAccelerationDetection *v1alpha1.GPUNetworkAccelerationDetection, expectedGPUNetworkAccelerationDetected bool, detectionLabelKey string) *TestCase {
	return &TestCase{
		Name: "MNNVL Nodes detection",
		NodePools: []TestNodePool{
			{
				Name:       "node-pool-a",
				LabelKey:   "nodePool0Key",
				LabelValue: "nodePool0Value",
				PlacementStrategy: PlacementStrategy{
					Gpu:       "binpack",
					GpuDevice: "binpack",
					Cpu:       "binpack",
				},
				GPUNetworkAccelerationDetection: gpuNetworkAccelerationDetection,
				GPUNetworkAccelerationLabelKey:  detectionLabelKey,
			},
		},
		Nodes: map[string]TestNode{
			"kind-worker": {
				Name: "kind-worker",
				Labels: map[string]string{
					"nodePool0Key": "nodePool0Value",
				},
			},
		},
		ManagedNodesConfig: []TestManagedNodesConfig{
			getWrongNameTestManagedNodesConfig(),
			getEmptyTestManagedNodesConfig(),
		},
		ExpectedNodePoolsState: map[string]ExpectedNodePoolState{
			"node-pool-a": {
				NodeNames:                              []string{"kind-worker"},
				ExpectedGPUNetworkAccelerationDetected: expectedGPUNetworkAccelerationDetected,
				Status: v1alpha1.NodePoolStatus{
					Phase: v1alpha1.NodePoolReady,
					Nodes: []v1alpha1.NodeInNodePool{
						{
							Name:   "kind-worker",
							Status: v1alpha1.NodeReady,
						},
					},
				},
			},
			kaiconstants.DefaultNodePoolName: {},
		},
		ExpectedMetricsStateList: []ExpectedMetricsState{
			{
				NodeName:      "kind-worker",
				NodePoolsName: "default",
				Count:         0,
			},
			{
				NodeName:      "kind-worker",
				NodePoolsName: "node-pool-a",
				Count:         1,
			},
		},
		ExpectedNodeLabels: map[string]ExpectedNodeLabels{
			"kind-worker": {
				expectedExistingLabels: map[string]string{
					"nodePool0Key":                       "nodePool0Value",
					kaiconstants.DefaultNodePoolLabelKey: "node-pool-a",
				},
			},
		},
	}
}

func NetworkTopologyTestCaseBase(
	name string,
	expectedNodePoolConditions []ExpectedCondition,
	expectedMismatchNodesNames []string,
) *TestCase {
	return &TestCase{
		Name: name,
		NodePools: []TestNodePool{
			{
				Name:                "node-pool-topology",
				LabelKey:            "nodePoolTopologyKey",
				LabelValue:          "nodePoolTopologyValue",
				NetworkTopologyName: "test-topology",
				PlacementStrategy: PlacementStrategy{
					Gpu:       "binpack",
					GpuDevice: "binpack",
					Cpu:       "binpack",
				},
			},
		},
		Topologies: []TestTopology{
			{
				Name: "test-topology",
				Levels: []TestTopologyLevel{
					{NodeLabel: "topology-level-1"},
					{NodeLabel: "topology-level-2"},
					{NodeLabel: "topology-level-3"},
				},
			},
		},
		Nodes: map[string]TestNode{
			"kind-worker": {
				Name: "kind-worker",
				Labels: map[string]string{
					"nodePoolTopologyKey": "nodePoolTopologyValue",
					"topology-level-1":    "value1",
					"topology-level-2":    "value2",
					"topology-level-3":    "value3",
				},
			},
			"kind-worker2": {
				Name: "kind-worker2",
				Labels: map[string]string{
					"nodePoolTopologyKey": "nodePoolTopologyValue",
					"topology-level-1":    "value1",
					"topology-level-2":    "value2",
					"topology-level-3":    "value3",
				},
			},
			"kind-worker3": {
				Name: "kind-worker3",
				Labels: map[string]string{
					"nodePoolTopologyKey": "nodePoolTopologyValue",
					"topology-level-1":    "value1",
					"topology-level-2":    "value2",
					"topology-level-3":    "value3",
				},
			},
		},
		ExpectedNodePoolsState: map[string]ExpectedNodePoolState{
			"node-pool-topology": {
				NodeNames: []string{"kind-worker", "kind-worker2", "kind-worker3"},
				Status: v1alpha1.NodePoolStatus{
					Phase: v1alpha1.NodePoolReady,
					Nodes: []v1alpha1.NodeInNodePool{
						{
							Name:             "kind-worker",
							Status:           v1alpha1.NodeReady,
							TopologyMismatch: true,
						},
						{
							Name:   "kind-worker2",
							Status: v1alpha1.NodeReady,
						},
						{
							Name:   "kind-worker3",
							Status: v1alpha1.NodeReady,
						},
					},
				},
				ExpectedConditions: expectedNodePoolConditions,
			},
			kaiconstants.DefaultNodePoolName: {},
		},
		ExpectedMismatchNodesNames: expectedMismatchNodesNames,
		ExpectedNodeLabels: map[string]ExpectedNodeLabels{
			"kind-worker": {
				expectedExistingLabels: map[string]string{
					"nodePoolTopologyKey": "nodePoolTopologyValue",
					"topology-level-1":    "value1",
					"topology-level-2":    "value2",
					"topology-level-3":    "value3",
				},
			},
			"kind-worker2": {
				expectedExistingLabels: map[string]string{
					"nodePoolTopologyKey": "nodePoolTopologyValue",
					"topology-level-1":    "value1",
					"topology-level-2":    "value2",
					"topology-level-3":    "value3",
				},
			},
			"kind-worker3": {
				expectedExistingLabels: map[string]string{
					"nodePoolTopologyKey": "nodePoolTopologyValue",
					"topology-level-1":    "value1",
					"topology-level-2":    "value2",
					"topology-level-3":    "value3",
				},
			},
		},
	}
}

func (tc *TestCase) WithExpectedNodepoolState(nodepoolName string, state ExpectedNodePoolState) *TestCase {
	tc.ExpectedNodePoolsState[nodepoolName] = state
	return tc
}

func (tc *TestCase) WithExpectedExistingNodesLabels(tn map[string]TestNode) *TestCase {
	tc.ExpectedNodeLabels = map[string]ExpectedNodeLabels{}
	for nodeName, testNode := range tn {
		tc.ExpectedNodeLabels[nodeName] = ExpectedNodeLabels{
			expectedExistingLabels: testNode.Labels,
		}
	}
	return tc
}

func (tc *TestCase) WithNodes(nodes map[string]TestNode) *TestCase {
	tc.Nodes = nodes
	return tc
}

func (tc *TestCase) WithExpectedNodesLabels(nodeName string, expectedNodeLabels map[string]string) *TestCase {
	expectedLabels, exists := tc.ExpectedNodeLabels[nodeName]
	if !exists {
		expectedLabels = ExpectedNodeLabels{expectedExistingLabels: expectedNodeLabels}
	}
	expectedLabels.expectedExistingLabels = expectedNodeLabels
	tc.ExpectedNodeLabels[nodeName] = expectedLabels
	return tc
}

func (tc *TestCase) WithUpdatedNodesLabels(nodeName string, labels map[string]string) *TestCase {
	testNode, exists := tc.Nodes[nodeName]
	if !exists {
		testNode = TestNode{Name: nodeName}
	}
	testNode.Labels = labels
	tc.Nodes[nodeName] = testNode
	return tc
}

func (tc *TestCase) WithExpectedMismatchNodeName(nodeName string) *TestCase {
	tc.ExpectedMismatchNodesNames = append(tc.ExpectedMismatchNodesNames, nodeName)
	return tc
}

func createNodesWithMissingTopologyLabel() map[string]TestNode {
	return map[string]TestNode{
		"kind-worker": {
			Name: "kind-worker",
			Labels: map[string]string{
				"nodePoolTopologyKey": "nodePoolTopologyValue",
				"topology-level-1":    "value1",
				"topology-level-2":    "value2",
				// Missing topology-level-3
			},
		},
		"kind-worker2": {
			Name: "kind-worker2",
			Labels: map[string]string{
				"nodePoolTopologyKey": "nodePoolTopologyValue",
				"topology-level-1":    "value1",
				"topology-level-2":    "value2",
				"topology-level-3":    "value3",
			},
		},
		"kind-worker3": {
			Name: "kind-worker3",
			Labels: map[string]string{
				"nodePoolTopologyKey": "nodePoolTopologyValue",
				"topology-level-1":    "value1",
				"topology-level-2":    "value2",
				"topology-level-3":    "value3",
			},
		},
	}
}

var _ = Describe("NodePoolController Tests", func() {
	var (
		ctx     context.Context
		cancel  context.CancelFunc
		stopper chan struct{}

		npc        *nodepool_controller.NodePoolController
		mncc       *managed_nodes_config.ManagedNodesConfigController
		fakeClient client.WithWatch
	)

	DescribeTable("",
		func(testCase *TestCase) {
			stopper = make(chan struct{}, 1)
			ctx, cancel = context.WithCancel(context.Background())
			DeferCleanup(func() {
				close(stopper)
				cancel()
			})

			fakeClient = preTestSetup(ctx, stopper, testCase, scheme)

			nodePoolControllerParams := &common.NodePoolControllerParams{}
			npc = nodepool_controller.NewNodePoolController(fakeClient, scheme, nodePoolControllerParams)
			npc.SetServiceMonitorEnabled(true)
			mncc = managed_nodes_config.NewManagedNodesConfigController(fakeClient, scheme, npc)

			reconcileAllNodePools(ctx, testCase.NodePools, npc, mncc)
			testCase.SetNodepoolController(npc)

			if testCase.TestCallback != nil {
				testCase.TestCallback(fakeClient, testCase)
				reconcileAllNodePools(ctx, testCase.NodePools, npc, mncc)
			}

			Eventually(func(g Gomega) {
				validateExpected(testCase, fakeClient, g)
			}, validateTestTimeout, validateTestInterval).Should(Succeed())
		},
		Entry("MNNVL Nodes detection - not set", MNNVLTestCaseBase(nil, false, "")),
		Entry("MNNVL Nodes detection - not set", MNNVLTestCaseBase(nil, false, "").
			WithNodes(createNodesWithMNNVLLabel(nodepool_controller.MNNVLLabel))),
		Entry("MNNVL Nodes detection - Use ", MNNVLTestCaseBase(ptr.To(v1alpha1.UseGPUNetworkAcceleration), true, "")),
		Entry("MNNVL Nodes detection - Do Not Use", MNNVLTestCaseBase(ptr.To(v1alpha1.DontUseGPUNetworkAcceleration), false, "")),
		Entry("MNNVL Nodes detection - Do Not Use With Label", MNNVLTestCaseBase(ptr.To(v1alpha1.DontUseGPUNetworkAcceleration), false, "").
			WithNodes(createNodesWithMNNVLLabel(nodepool_controller.MNNVLLabel))),
		Entry("MNNVL Nodes detection - automatic - true", MNNVLTestCaseBase(ptr.To(v1alpha1.AutoGPUNetworkAccelerationDetection), true, "").
			WithNodes(createNodesWithMNNVLLabel(nodepool_controller.MNNVLLabel))),
		Entry("MNNVL Nodes detection - automatic - true", MNNVLTestCaseBase(ptr.To(v1alpha1.AutoGPUNetworkAccelerationDetection), false, "")),
		Entry("MNNVL Nodes detection - custom label detected", MNNVLTestCaseBase(ptr.To(v1alpha1.AutoGPUNetworkAccelerationDetection), true, "GPU-NET-ACC").
			WithNodes(createNodesWithMNNVLLabel("GPU-NET-ACC"))),
		Entry("MNNVL Nodes detection - custom label not detected", MNNVLTestCaseBase(ptr.To(v1alpha1.AutoGPUNetworkAccelerationDetection), false, nodepool_controller.MNNVLLabel).
			WithNodes(createNodesWithMNNVLLabel("GPU-NET-ACC"))),
		Entry("Sanity - Nodes assigned", &TestCase{
			Name: "Sanity - Nodes assigned",
			NodePools: []TestNodePool{
				{
					Name:       "node-pool-a",
					LabelKey:   "nodePool0Key",
					LabelValue: "nodePool0Value",
					PlacementStrategy: PlacementStrategy{
						Gpu:       "binpack",
						GpuDevice: "binpack",
						Cpu:       "binpack",
					},
				},
			},
			Nodes: map[string]TestNode{
				"kind-worker": {
					Name:   "kind-worker",
					Labels: map[string]string{"nodePool0Key": "nodePool0Value"},
				},
			},
			ManagedNodesConfig: []TestManagedNodesConfig{
				getWrongNameTestManagedNodesConfig(),
				getEmptyTestManagedNodesConfig(),
			},
			ExpectedNodePoolsState: map[string]ExpectedNodePoolState{
				"node-pool-a": {
					NodeNames: []string{"kind-worker"},
					Status: v1alpha1.NodePoolStatus{
						Phase: v1alpha1.NodePoolReady,
						Nodes: []v1alpha1.NodeInNodePool{
							{
								Name:   "kind-worker",
								Status: v1alpha1.NodeReady,
							},
						},
					},
				},
				kaiconstants.DefaultNodePoolName: {},
			},
			ExpectedMetricsStateList: []ExpectedMetricsState{
				{
					NodeName:      "kind-worker",
					NodePoolsName: "default",
					Count:         0,
				},
				{
					NodeName:      "kind-worker",
					NodePoolsName: "node-pool-a",
					Count:         1,
				},
			},
			ExpectedNodeLabels: map[string]ExpectedNodeLabels{
				"kind-worker": {
					expectedExistingLabels: map[string]string{
						"nodePool0Key":                       "nodePool0Value",
						kaiconstants.DefaultNodePoolLabelKey: "node-pool-a",
					},
				},
			},
		}),
		Entry("Sanity - all nodes assigned correctly", &TestCase{
			Name: "Sanity - all nodes assigned correctly",
			NodePools: []TestNodePool{
				{
					Name:       "node-pool-a",
					LabelKey:   "nodePool0Key",
					LabelValue: "nodePool0Value",
					PlacementStrategy: PlacementStrategy{
						Gpu:       "binpack",
						GpuDevice: "binpack",
						Cpu:       "binpack",
					},
				},
			},
			Nodes: map[string]TestNode{
				"kind-worker": {
					Name:   "kind-worker",
					Labels: map[string]string{"nodePool0Key": "nodePool0Value", kaiconstants.DefaultNodePoolLabelKey: "node-pool-a"},
				},
				"kind-worker2": {
					Name:   "kind-worker2",
					Labels: map[string]string{},
				},
				"kind-worker3": {
					Name:   "kind-worker3",
					Labels: map[string]string{kaiconstants.DefaultNodePoolLabelKey: "node-pool-a"},
				},
			},
			ManagedNodesConfig: []TestManagedNodesConfig{
				getWrongNameTestManagedNodesConfig(),
			},
			ExpectedNodePoolsState: map[string]ExpectedNodePoolState{
				"node-pool-a": {
					NodeNames: []string{"kind-worker"},
					Status: v1alpha1.NodePoolStatus{
						Phase: v1alpha1.NodePoolReady,
						Nodes: []v1alpha1.NodeInNodePool{
							{
								Name:   "kind-worker",
								Status: v1alpha1.NodeReady,
							},
						},
					},
				},
				kaiconstants.DefaultNodePoolName: {
					NodeNames: []string{"kind-worker2", "kind-worker3"},
				},
			},
			ExpectedMetricsStateList: []ExpectedMetricsState{
				{
					NodeName:      "kind-worker",
					NodePoolsName: "default",
					Count:         0,
				},
				{
					NodeName:      "kind-worker",
					NodePoolsName: "node-pool-a",
					Count:         1,
				},
				{
					NodeName:      "kind-worker2",
					NodePoolsName: "default",
					Count:         1,
				},
				{
					NodeName:      "kind-worker2",
					NodePoolsName: "node-pool-a",
					Count:         0,
				},
				{
					NodeName:      "kind-worker3",
					NodePoolsName: "default",
					Count:         1,
				},
				{
					NodeName:      "kind-worker3",
					NodePoolsName: "node-pool-a",
					Count:         0,
				},
			},
			ExpectedNodeLabels: map[string]ExpectedNodeLabels{
				"kind-worker":  {},
				"kind-worker2": {},
				"kind-worker3": {},
			},
		}),
		Entry("Default NodePool - Nodes assigned", &TestCase{
			Name: "Default NodePool - Nodes assigned",
			NodePools: []TestNodePool{
				{
					Name:       "node-pool-a",
					LabelKey:   "nodePool0Key",
					LabelValue: "nodePool0Value",
					PlacementStrategy: PlacementStrategy{
						Gpu:       "binpack",
						GpuDevice: "binpack",
						Cpu:       "binpack",
					},
				},
			},
			Nodes: map[string]TestNode{
				"kind-worker": {
					Name:   "kind-worker",
					Labels: map[string]string{kaiconstants.DefaultNodePoolLabelKey: "node-pool-a"},
				},
				"kind-worker2": {
					Name:   "kind-worker2",
					Labels: map[string]string{},
				},
			},
			ManagedNodesConfig: []TestManagedNodesConfig{
				getWrongNameTestManagedNodesConfig(),
			},
			ExpectedNodePoolsState: map[string]ExpectedNodePoolState{
				"node-pool-a": {
					NodeNames: []string{},
					Status: v1alpha1.NodePoolStatus{
						Phase: v1alpha1.NodePoolEmpty,
					},
				},
				kaiconstants.DefaultNodePoolName: {
					NodeNames: []string{"kind-worker", "kind-worker2"},
				},
			},
			ExpectedMetricsStateList: []ExpectedMetricsState{
				{
					NodeName:      "kind-worker",
					NodePoolsName: "default",
					Count:         1,
				},
				{
					NodeName:      "kind-worker",
					NodePoolsName: "node-pool-a",
					Count:         0,
				},
				{
					NodeName:      "kind-worker2",
					NodePoolsName: "default",
					Count:         1,
				},
				{
					NodeName:      "kind-worker2",
					NodePoolsName: "node-pool-a",
					Count:         0,
				},
			},
			ExpectedNodeLabels: map[string]ExpectedNodeLabels{
				"kind-worker":  {},
				"kind-worker2": {},
			},
		}),
		Entry("Deleting NodePool - Nodes re-assigned", &TestCase{
			Name: "Deleting NodePool - Nodes re-assigned",
			NodePools: []TestNodePool{
				{
					Name:       "node-pool-a",
					LabelKey:   "nodePoolAKey",
					LabelValue: "nodePoolAValue",
					PlacementStrategy: PlacementStrategy{
						Gpu:       "binpack",
						GpuDevice: "binpack",
						Cpu:       "binpack",
					},
				},
				{
					Name:       "node-pool-b",
					LabelKey:   "nodePoolBKey",
					LabelValue: "nodePoolBValue",
					PlacementStrategy: PlacementStrategy{
						Gpu:       "binpack",
						GpuDevice: "binpack",
						Cpu:       "binpack",
					},
				},
			},
			Nodes: map[string]TestNode{
				"kind-worker": {
					Name:   "kind-worker",
					Labels: map[string]string{"nodePoolBKey": "nodePoolBValue", kaiconstants.DefaultNodePoolLabelKey: "node-pool-b"},
				},
				"kind-worker2": {
					Name:   "kind-worker2",
					Labels: map[string]string{"nodePoolAKey": "nodePoolAValue", kaiconstants.DefaultNodePoolLabelKey: "node-pool-a"},
				},
			},
			TestCallback: func(k8sClient client.Client, testCase *TestCase) {
				updateNodeLabels("kind-worker2", map[string]string{"nodePoolBKey": "nodePoolBValue"}, "", k8sClient)
				nodePool := eventuallyGetNodePoolFromClient("node-pool-a", k8sClient)
				_ = k8sClient.Delete(ctx, nodePool)
			},
			ExpectedNodePoolsState: map[string]ExpectedNodePoolState{
				"node-pool-a": {},
				"node-pool-b": {
					NodeNames: []string{"kind-worker", "kind-worker2"},
					Status: v1alpha1.NodePoolStatus{
						Phase: v1alpha1.NodePoolReady,
						Nodes: []v1alpha1.NodeInNodePool{
							{
								Name:   "kind-worker",
								Status: v1alpha1.NodeReady,
							},
							{
								Name:   "kind-worker2",
								Status: v1alpha1.NodeReady,
							},
						},
					},
				},
			},
			ExpectedMetricsStateList: []ExpectedMetricsState{
				{
					NodeName:      "kind-worker",
					NodePoolsName: "node-pool-a",
					Count:         0,
				},
				{
					NodeName:      "kind-worker",
					NodePoolsName: "node-pool-b",
					Count:         1,
				},
				{
					NodeName:      "kind-worker2",
					NodePoolsName: "node-pool-a",
					Count:         0,
				},
				{
					NodeName:      "kind-worker2",
					NodePoolsName: "node-pool-b",
					Count:         1,
				},
			},
		}),
		Entry("Deleting NodePool - Nodes re-assigned to default", &TestCase{
			Name: "Deleting NodePool - Nodes re-assigned to default",
			NodePools: []TestNodePool{
				{
					Name:       "node-pool-b",
					LabelKey:   "nodePoolBKey",
					LabelValue: "nodePoolBValue",
					PlacementStrategy: PlacementStrategy{
						Gpu:       "binpack",
						GpuDevice: "binpack",
						Cpu:       "binpack",
					},
				},
			},
			Nodes: map[string]TestNode{
				"kind-worker2": {
					Name:   "kind-worker2",
					Labels: map[string]string{"nodePoolBKey": "nodePoolBValue", kaiconstants.DefaultNodePoolLabelKey: "node-pool-b"},
				},
			},
			ManagedNodesConfig: []TestManagedNodesConfig{
				getEmptyTestManagedNodesConfig(),
			},
			TestCallback: func(k8sClient client.Client, testCase *TestCase) {
				nodePool := eventuallyGetNodePoolFromClient("node-pool-b", k8sClient)
				_ = k8sClient.Delete(ctx, nodePool)
			},
			ExpectedNodePoolsState: map[string]ExpectedNodePoolState{
				kaiconstants.DefaultNodePoolName: {
					NodeNames: []string{"kind-worker2"},
				},
				"node-pool-b": {},
			},
			ExpectedMetricsStateList: []ExpectedMetricsState{
				{
					NodeName:      "kind-worker2",
					NodePoolsName: "default",
					Count:         1,
				},
				{
					NodeName:      "kind-worker2",
					NodePoolsName: "node-pool-b",
					Count:         0,
				},
			},
		}),
		Entry("Updating nodepool label key - Nodes re-assigned", &TestCase{
			Name: "Updating nodepool label key - Nodes re-assigned",
			NodePools: []TestNodePool{
				{
					Name:       "node-pool-a",
					LabelKey:   "nodePoolAKey",
					LabelValue: "nodePoolAValue",
					PlacementStrategy: PlacementStrategy{
						Gpu:       "binpack",
						GpuDevice: "binpack",
						Cpu:       "binpack",
					},
				},
				{
					Name:       "node-pool-b",
					LabelKey:   "nodePoolBKey",
					LabelValue: "nodePoolBValue",
					PlacementStrategy: PlacementStrategy{
						Gpu:       "binpack",
						GpuDevice: "binpack",
						Cpu:       "binpack",
					},
				},
			},
			Nodes: map[string]TestNode{
				"kind-worker": {
					Name: "kind-worker",
					Labels: map[string]string{
						"nodePoolBKey":                       "nodePoolBValue",
						"nodePoolBKey-new":                   "nodePoolBValue",
						kaiconstants.DefaultNodePoolLabelKey: "node-pool-b"},
				},
				"kind-worker2": {
					Name:   "kind-worker2",
					Labels: map[string]string{"nodePoolBKey": "nodePoolBValue", kaiconstants.DefaultNodePoolLabelKey: "node-pool-b"},
				},
			},
			TestCallback: func(k8sClient client.Client, testCase *TestCase) {
				nodePool := eventuallyGetNodePoolFromClient("node-pool-b", k8sClient)
				updateNodePoolLabels("node-pool-b", "nodePoolBKey-new", nodePool.Spec.LabelValue, k8sClient)
			},
			ExpectedNodePoolsState: map[string]ExpectedNodePoolState{
				"node-pool-a": {
					NodeNames: []string{},
					Status: v1alpha1.NodePoolStatus{
						Phase: v1alpha1.NodePoolEmpty,
					},
				},
				"node-pool-b": {
					NodeNames: []string{"kind-worker"},
					Status: v1alpha1.NodePoolStatus{
						Phase: v1alpha1.NodePoolReady,
						Nodes: []v1alpha1.NodeInNodePool{
							{
								Name:   "kind-worker",
								Status: v1alpha1.NodeReady,
							},
						},
					},
				},
				kaiconstants.DefaultNodePoolName: {
					NodeNames: []string{"kind-worker2"},
				},
			},
			ExpectedMetricsStateList: []ExpectedMetricsState{
				{
					NodeName:      "kind-worker2",
					NodePoolsName: "default",
					Count:         1,
				},
				{
					NodeName:      "kind-worker2",
					NodePoolsName: "node-pool-a",
					Count:         0,
				},
				{
					NodeName:      "kind-worker2",
					NodePoolsName: "node-pool-b",
					Count:         0,
				},
				{
					NodeName:      "kind-worker",
					NodePoolsName: "default",
					Count:         0,
				},
				{
					NodeName:      "kind-worker",
					NodePoolsName: "node-pool-a",
					Count:         0,
				},
				{
					NodeName:      "kind-worker",
					NodePoolsName: "node-pool-b",
					Count:         1,
				},
			},
		}),
		Entry("Updating nodepool label value - Nodes re-assigned", &TestCase{
			Name: "Updating nodepool label value - Nodes re-assigned",
			NodePools: []TestNodePool{
				{
					Name:       "node-pool-a",
					LabelKey:   "nodePoolAKey",
					LabelValue: "nodePoolAValue",
					PlacementStrategy: PlacementStrategy{
						Gpu:       "binpack",
						GpuDevice: "binpack",
						Cpu:       "binpack",
					},
				},
				{
					Name:       "node-pool-b",
					LabelKey:   "nodePoolBKey",
					LabelValue: "nodePoolBValue",
					PlacementStrategy: PlacementStrategy{
						Gpu:       "binpack",
						GpuDevice: "binpack",
						Cpu:       "binpack",
					},
				},
			},
			Nodes: map[string]TestNode{
				"kind-worker": {
					Name: "kind-worker",
					Labels: map[string]string{
						"nodePoolBKey":                       "nodePoolBValue",
						"nodePoolAKey":                       "nodePoolAValue",
						kaiconstants.DefaultNodePoolLabelKey: "node-pool-b"},
				},
				"kind-worker2": {
					Name:   "kind-worker2",
					Labels: map[string]string{"nodePoolBKey": "nodePoolBValue", kaiconstants.DefaultNodePoolLabelKey: "node-pool-b"},
				},
				"kind-worker3": {
					Name: "kind-worker3",
					Labels: map[string]string{
						"nodePoolBKey":                       "nodePoolBValue-new",
						"nodePoolAKey":                       "nodePoolAValue",
						kaiconstants.DefaultNodePoolLabelKey: "node-pool-a"},
				},
				"kind-worker4": {
					Name: "kind-worker4",
					Labels: map[string]string{
						"nodePoolBKey": "nodePoolBValue-new"},
				},
			},
			TestCallback: func(k8sClient client.Client, testCase *TestCase) {
				nodePool := eventuallyGetNodePoolFromClient("node-pool-b", k8sClient)
				updateNodePoolLabels(nodePool.Name, nodePool.Spec.LabelKey, "nodePoolBValue-new", k8sClient)
			},
			ExpectedNodePoolsState: map[string]ExpectedNodePoolState{
				"node-pool-a": {
					NodeNames: []string{"kind-worker", "kind-worker3"},
					Status: v1alpha1.NodePoolStatus{
						Phase: v1alpha1.NodePoolReady,
						Nodes: []v1alpha1.NodeInNodePool{
							{
								Name:   "kind-worker",
								Status: v1alpha1.NodeReady,
							},
							{
								Name:   "kind-worker3",
								Status: v1alpha1.NodeReady,
							},
						},
					},
				},
				"node-pool-b": {
					NodeNames: []string{"kind-worker4"},
					Status: v1alpha1.NodePoolStatus{
						Phase: v1alpha1.NodePoolReady,
						Nodes: []v1alpha1.NodeInNodePool{
							{
								Name:   "kind-worker4",
								Status: v1alpha1.NodeReady,
							},
						},
					},
				},
				kaiconstants.DefaultNodePoolName: {
					NodeNames: []string{"kind-worker2"},
				},
			},
		}),
		Entry("Updating nodepool label key and value - Nodes re-assigned", &TestCase{
			Name: "Updating nodepool label key and value",
			NodePools: []TestNodePool{
				{
					Name:       "node-pool-a",
					LabelKey:   "nodePoolAKey",
					LabelValue: "nodePoolAValue",
					PlacementStrategy: PlacementStrategy{
						Gpu:       "binpack",
						GpuDevice: "binpack",
						Cpu:       "binpack",
					},
				},
				{
					Name:       "node-pool-b",
					LabelKey:   "nodePoolBKey",
					LabelValue: "nodePoolBValue",
					PlacementStrategy: PlacementStrategy{
						Gpu:       "binpack",
						GpuDevice: "binpack",
						Cpu:       "binpack",
					},
				},
			},
			Nodes: map[string]TestNode{
				"kind-worker": {
					Name: "kind-worker",
					Labels: map[string]string{
						"nodePoolBKey":                       "nodePoolBValue",
						"nodePoolBKey-new":                   "nodePoolBValue-new",
						kaiconstants.DefaultNodePoolLabelKey: "node-pool-b"},
				},
				"kind-worker2": {
					Name:   "kind-worker2",
					Labels: map[string]string{"nodePoolBKey": "nodePoolBValue", kaiconstants.DefaultNodePoolLabelKey: "node-pool-b"},
				},
				"kind-worker3": {
					Name: "kind-worker3",
					Labels: map[string]string{
						"nodePoolBKey-new":                   "nodePoolBValue-new",
						"nodePoolAKey":                       "nodePoolAValue",
						kaiconstants.DefaultNodePoolLabelKey: "node-pool-a"},
				},
				"kind-worker4": {
					Name: "kind-worker4",
					Labels: map[string]string{
						"nodePoolBKey-new": "nodePoolBValue-new"},
				},
				"kind-worker5": {
					Name: "kind-worker5",
					Labels: map[string]string{
						"nodePoolBKey-new": "nodePoolBValue-wrong"},
				},
			},
			TestCallback: func(k8sClient client.Client, testCase *TestCase) {
				updateNodePoolLabels("node-pool-b", "nodePoolBKey-new", "nodePoolBValue-new", k8sClient)
			},
			ExpectedNodePoolsState: map[string]ExpectedNodePoolState{
				"node-pool-a": {
					NodeNames: []string{"kind-worker3"},
					Status: v1alpha1.NodePoolStatus{
						Phase: v1alpha1.NodePoolReady,
						Nodes: []v1alpha1.NodeInNodePool{
							{
								Name:   "kind-worker3",
								Status: v1alpha1.NodeReady,
							},
						},
					},
				},
				"node-pool-b": {
					NodeNames: []string{"kind-worker", "kind-worker4"},
					Status: v1alpha1.NodePoolStatus{
						Phase: v1alpha1.NodePoolReady,
						Nodes: []v1alpha1.NodeInNodePool{
							{
								Name:   "kind-worker",
								Status: v1alpha1.NodeReady,
							},
							{
								Name:   "kind-worker4",
								Status: v1alpha1.NodeReady,
							},
						},
					},
				},
				kaiconstants.DefaultNodePoolName: {
					NodeNames: []string{"kind-worker2", "kind-worker5"},
				},
			},
		}),
		Entry("Node switching nodepools - all labels updated", &TestCase{
			Name: "Node switching nodepools - all labels updated",
			NodePools: []TestNodePool{
				{
					Name:       "node-pool-a",
					LabelKey:   "nodePoolAKey",
					LabelValue: "nodePoolAValue",
					PlacementStrategy: PlacementStrategy{
						Gpu:       "binpack",
						GpuDevice: "binpack",
						Cpu:       "binpack",
					},
				},
				{
					Name:       "node-pool-b",
					LabelKey:   "nodePoolBKey",
					LabelValue: "nodePoolBValue",
					PlacementStrategy: PlacementStrategy{
						Gpu:       "binpack",
						GpuDevice: "binpack",
						Cpu:       "binpack",
					},
				},
			},
			Nodes: map[string]TestNode{
				"kind-worker": {
					Name: "kind-worker",
					Labels: map[string]string{
						"nodePoolBKey":                       "nodePoolBValue",
						kaiconstants.DefaultNodePoolLabelKey: "node-pool-b"},
				},
			},
			TestCallback: func(k8sClient client.Client, testCase *TestCase) {
				updateNodeLabels("kind-worker", map[string]string{"nodePoolAKey": "nodePoolAValue"}, "nodePoolBKey", k8sClient)
			},
			ExpectedNodePoolsState: map[string]ExpectedNodePoolState{
				"node-pool-a": {
					NodeNames: []string{"kind-worker"},
					Status: v1alpha1.NodePoolStatus{
						Phase: v1alpha1.NodePoolReady,
						Nodes: []v1alpha1.NodeInNodePool{
							{
								Name:   "kind-worker",
								Status: v1alpha1.NodeReady,
							},
						},
					},
				},
				"node-pool-b": {
					NodeNames: []string{},
					Status: v1alpha1.NodePoolStatus{
						Phase: v1alpha1.NodePoolEmpty,
					},
				},
			},
			ExpectedNodeLabels: map[string]ExpectedNodeLabels{
				"kind-worker": {
					expectedExistingLabels: map[string]string{
						"nodePoolAKey":                       "nodePoolAValue",
						kaiconstants.DefaultNodePoolLabelKey: "node-pool-a",
					},
				},
			},
		}),
		Entry("Node switching nodepools - pods still running - nodepool unschedulable", &TestCase{
			Name: "Node switching nodepools - pods still running - nodepool unschedulable",
			NodePools: []TestNodePool{
				{
					Name:       "node-pool-a",
					LabelKey:   "nodePoolAKey",
					LabelValue: "nodePoolAValue",
					PlacementStrategy: PlacementStrategy{
						Gpu:       "binpack",
						GpuDevice: "binpack",
						Cpu:       "binpack",
					},
				},
				{
					Name:       "node-pool-b",
					LabelKey:   "nodePoolBKey",
					LabelValue: "nodePoolBValue",
					PlacementStrategy: PlacementStrategy{
						Gpu:       "binpack",
						GpuDevice: "binpack",
						Cpu:       "binpack",
					},
				},
			},
			Nodes: map[string]TestNode{
				"kind-worker": {
					Name: "kind-worker",
					Labels: map[string]string{
						"nodePoolBKey":                       "nodePoolBValue",
						kaiconstants.DefaultNodePoolLabelKey: "node-pool-b"},
				},
			},
			Pods: []TestPod{
				{
					Name:        "pod0",
					Labels:      map[string]string{kaiconstants.DefaultNodePoolLabelKey: "node-pool-b"},
					NodeName:    "kind-worker",
					Annotations: map[string]string{config.PodGroupAnnotationForPod: "pod-group-name-1"},
				},
			},
			PodGroups: []TestPodGroup{
				{
					Name:   "pod-group-name-1",
					Labels: map[string]string{kaiconstants.DefaultNodePoolLabelKey: "node-pool-b"},
				},
			},
			TestCallback: func(k8sClient client.Client, testCase *TestCase) {
				updateNodeLabels("kind-worker", map[string]string{"nodePoolAKey": "nodePoolAValue"}, "nodePoolBKey", k8sClient)
			},
			ExpectedNodePoolsState: map[string]ExpectedNodePoolState{
				"node-pool-a": {
					NodeNames: []string{},
					Status: v1alpha1.NodePoolStatus{
						Phase:   v1alpha1.NodePoolEmpty,
						Message: fmt.Sprintf(common.NodesAssignedToThisNodePoolWaitingForDrainMessage, "kind-worker"),
					},
				},
				"node-pool-b": {
					NodeNames: []string{"kind-worker"},
					Status: v1alpha1.NodePoolStatus{
						Phase:   v1alpha1.NodePoolUnschedulable,
						Message: fmt.Sprintf(common.NodesAssignedToDifferentNodePoolWaitingForDrainMessage, "kind-worker"),
						Nodes: []v1alpha1.NodeInNodePool{
							{
								Name:   "kind-worker",
								Status: v1alpha1.NodeUnschedulable,
							},
						},
					},
				},
			},
		}),
		Entry("Node switching nodepools - pods still running - nodepool ready but with unschedulable node message", &TestCase{
			Name: "Node switching nodepools - pods still running - nodepool ready but with unschedulable node message",
			NodePools: []TestNodePool{
				{
					Name:       "node-pool-a",
					LabelKey:   "nodePoolAKey",
					LabelValue: "nodePoolAValue",
					PlacementStrategy: PlacementStrategy{
						Gpu:       "binpack",
						GpuDevice: "binpack",
						Cpu:       "binpack",
					},
				},
				{
					Name:       "node-pool-b",
					LabelKey:   "nodePoolBKey",
					LabelValue: "nodePoolBValue",
					PlacementStrategy: PlacementStrategy{
						Gpu:       "binpack",
						GpuDevice: "binpack",
						Cpu:       "binpack",
					},
				},
			},
			Nodes: map[string]TestNode{
				"kind-worker": {
					Name: "kind-worker",
					Labels: map[string]string{
						"nodePoolBKey":                       "nodePoolBValue",
						kaiconstants.DefaultNodePoolLabelKey: "node-pool-b"},
				},
				"kind-worker2": {
					Name: "kind-worker2",
					Labels: map[string]string{
						"nodePoolBKey":                       "nodePoolBValue",
						kaiconstants.DefaultNodePoolLabelKey: "node-pool-b"},
				},
				"kind-worker3": {
					Name:   "kind-worker3",
					Labels: map[string]string{},
				},
			},
			Pods: []TestPod{
				{
					Name:        "pod0",
					Labels:      map[string]string{kaiconstants.DefaultNodePoolLabelKey: "node-pool-b"},
					NodeName:    "kind-worker",
					Annotations: map[string]string{config.PodGroupAnnotationForPod: "pod-group-name-1"},
				},
				{
					Name:        "pod1",
					Labels:      map[string]string{},
					NodeName:    "kind-worker3",
					Annotations: map[string]string{config.PodGroupAnnotationForPod: "pod-group-name-2"},
				},
			},
			PodGroups: []TestPodGroup{
				{
					Name:   "pod-group-name-1",
					Labels: map[string]string{kaiconstants.DefaultNodePoolLabelKey: "node-pool-b"},
				},
				{
					Name:   "pod-group-name-2",
					Labels: map[string]string{},
				},
			},
			TestCallback: func(k8sClient client.Client, testCase *TestCase) {
				updateNodeLabels("kind-worker", map[string]string{"nodePoolAKey": "nodePoolAValue"}, "nodePoolBKey", k8sClient)
				updateNodeLabels("kind-worker3", map[string]string{"nodePoolAKey": "nodePoolAValue"}, "", k8sClient)
			},
			ExpectedNodePoolsState: map[string]ExpectedNodePoolState{
				"node-pool-a": {
					NodeNames: []string{},
					Status: v1alpha1.NodePoolStatus{
						Phase:   v1alpha1.NodePoolEmpty,
						Message: fmt.Sprintf(common.NodesAssignedToThisNodePoolWaitingForDrainMessage, "kind-worker, kind-worker3"),
					},
				},
				"node-pool-b": {
					NodeNames: []string{"kind-worker", "kind-worker2"},
					Status: v1alpha1.NodePoolStatus{
						Phase:   v1alpha1.NodePoolReady,
						Message: fmt.Sprintf(common.NodesAssignedToDifferentNodePoolWaitingForDrainMessage, "kind-worker"),
						Nodes: []v1alpha1.NodeInNodePool{
							{
								Name:   "kind-worker",
								Status: v1alpha1.NodeUnschedulable,
							},
							{
								Name:   "kind-worker2",
								Status: v1alpha1.NodeReady,
							},
						},
					},
				},
				kaiconstants.DefaultNodePoolName: {
					NodeNames: []string{"kind-worker3"},
					Status: v1alpha1.NodePoolStatus{
						Phase: v1alpha1.NodePoolUnschedulable,
						Message: fmt.Sprintf(common.NodesAssignedToThisNodePoolWaitingForDrainMessage+"\n\n"+common.NodesAssignedToDifferentNodePoolWaitingForDrainMessage,
							"kind-worker", "kind-worker3"),
					},
				},
			},
		}),
		Entry("Node switching nodepools - pods running but then terminating - nodepool schedulable", &TestCase{
			Name: "Node switching nodepools - pods running but then terminating - nodepool schedulable",
			NodePools: []TestNodePool{
				{
					Name:       "node-pool-a",
					LabelKey:   "nodePoolAKey",
					LabelValue: "nodePoolAValue",
					PlacementStrategy: PlacementStrategy{
						Gpu:       "binpack",
						GpuDevice: "binpack",
						Cpu:       "binpack",
					},
				},
				{
					Name:       "node-pool-b",
					LabelKey:   "nodePoolBKey",
					LabelValue: "nodePoolBValue",
					PlacementStrategy: PlacementStrategy{
						Gpu:       "binpack",
						GpuDevice: "binpack",
						Cpu:       "binpack",
					},
				},
			},
			Nodes: map[string]TestNode{
				"kind-worker": {
					Name: "kind-worker",
					Labels: map[string]string{
						"nodePoolBKey":                       "nodePoolBValue",
						kaiconstants.DefaultNodePoolLabelKey: "node-pool-b"},
				},
			},
			Pods: []TestPod{
				{
					Name:        "pod0",
					Labels:      map[string]string{kaiconstants.DefaultNodePoolLabelKey: "node-pool-b"},
					NodeName:    "kind-worker",
					Annotations: map[string]string{config.PodGroupAnnotationForPod: "pod-group-name-1"},
				},
			},
			PodGroups: []TestPodGroup{
				{
					Name:   "pod-group-name-1",
					Labels: map[string]string{kaiconstants.DefaultNodePoolLabelKey: "node-pool-b"},
				},
			},
			TestCallback: func(k8sClient client.Client, testCase *TestCase) {
				node := getNodeFromClient("kind-worker", k8sClient)
				node.Labels["nodePoolAKey"] = "nodePoolAValue"
				delete(node.Labels, "nodePoolBKey")
				updateNodeLabels("kind-worker", map[string]string{"nodePoolAKey": "nodePoolAValue"}, "nodePoolBKey", k8sClient)

				pod := getPodFromClient("pod0", runaiNamespace, k8sClient)
				if pod != nil {
					DeleteImmediatelyAndPollUntilDeleted(k8sClient, pod)
				}
			},
			ExpectedNodePoolsState: map[string]ExpectedNodePoolState{
				"node-pool-a": {
					NodeNames: []string{"kind-worker"},
					Status: v1alpha1.NodePoolStatus{
						Phase: v1alpha1.NodePoolReady,
						Nodes: []v1alpha1.NodeInNodePool{
							{
								Name:   "kind-worker",
								Status: v1alpha1.NodeReady,
							},
						},
					},
				},
				"node-pool-b": {
					Status: v1alpha1.NodePoolStatus{
						Phase: v1alpha1.NodePoolEmpty,
					},
				},
			},
		}),
		Entry("Node switching nodepools - pods still running - nodepool unschedulable - then node switching back to previous nodepool", &TestCase{
			Name: "Node switching nodepools - pods still running - nodepool unschedulable - then node switching back to previous nodepool",
			NodePools: []TestNodePool{
				{
					Name:       "node-pool-a",
					LabelKey:   "nodePoolAKey",
					LabelValue: "nodePoolAValue",
					PlacementStrategy: PlacementStrategy{
						Gpu:       "binpack",
						GpuDevice: "binpack",
						Cpu:       "binpack",
					},
				},
				{
					Name:       "node-pool-b",
					LabelKey:   "nodePoolBKey",
					LabelValue: "nodePoolBValue",
					PlacementStrategy: PlacementStrategy{
						Gpu:       "binpack",
						GpuDevice: "binpack",
						Cpu:       "binpack",
					},
				},
			},
			Nodes: map[string]TestNode{
				"kind-worker": {
					Name: "kind-worker",
					Labels: map[string]string{
						"nodePoolBKey":                       "nodePoolBValue",
						kaiconstants.DefaultNodePoolLabelKey: "node-pool-b"},
				},
			},
			Pods: []TestPod{
				{
					Name:        "pod0",
					Labels:      map[string]string{kaiconstants.DefaultNodePoolLabelKey: "node-pool-b"},
					NodeName:    "kind-worker",
					Annotations: map[string]string{config.PodGroupAnnotationForPod: "pod-group-name-1"},
				},
			},
			PodGroups: []TestPodGroup{
				{
					Name:   "pod-group-name-1",
					Labels: map[string]string{kaiconstants.DefaultNodePoolLabelKey: "node-pool-a"},
				},
			},
			TestCallback: func(k8sClient client.Client, testCase *TestCase) {
				updateNodeLabels("kind-worker", map[string]string{"nodePoolAKey": "nodePoolAValue"}, "nodePoolBKey", k8sClient)
				reconcileAllNodePools(ctx, testCase.NodePools, npc, mncc)
				updateNodeLabels("kind-worker", map[string]string{"nodePoolBKey": "nodePoolBValue"}, "nodePoolAKey", k8sClient)
			},
			ExpectedNodePoolsState: map[string]ExpectedNodePoolState{
				"node-pool-a": {
					NodeNames: []string{},
					Status: v1alpha1.NodePoolStatus{
						Phase: v1alpha1.NodePoolEmpty,
					},
				},
				"node-pool-b": {
					NodeNames: []string{"kind-worker"},
					Status: v1alpha1.NodePoolStatus{
						Phase: v1alpha1.NodePoolReady,
						Nodes: []v1alpha1.NodeInNodePool{
							{
								Name:   "kind-worker",
								Status: v1alpha1.NodeReady,
							},
						},
					},
				},
			},
		}),
		Entry("Nodepool ready - pods running with different nodepool - node unschedulable", &TestCase{
			Name: "Nodepool ready - pods running with different nodepool - node unschedulable",
			NodePools: []TestNodePool{
				{
					Name:       "node-pool-a",
					LabelKey:   "nodePoolAKey",
					LabelValue: "nodePoolAValue",
					PlacementStrategy: PlacementStrategy{
						Gpu:       "binpack",
						GpuDevice: "binpack",
						Cpu:       "binpack",
					},
				},
				{
					Name:       "node-pool-b",
					LabelKey:   "nodePoolBKey",
					LabelValue: "nodePoolBValue",
					PlacementStrategy: PlacementStrategy{
						Gpu:       "binpack",
						GpuDevice: "binpack",
						Cpu:       "binpack",
					},
				},
			},
			Nodes: map[string]TestNode{
				"kind-worker": {
					Name: "kind-worker",
					Labels: map[string]string{
						"nodePoolBKey":                       "nodePoolBValue",
						kaiconstants.DefaultNodePoolLabelKey: "node-pool-b"},
				},
				"kind-worker2": {
					Name: "kind-worker2",
					Labels: map[string]string{
						"nodePoolBKey":                       "nodePoolBValue",
						kaiconstants.DefaultNodePoolLabelKey: "node-pool-b"},
				},
				"kind-worker3": {
					Name:   "kind-worker3",
					Labels: map[string]string{},
				},
				"kind-worker4": {
					Name:   "kind-worker4",
					Labels: map[string]string{},
				},
			},
			Pods: []TestPod{
				{
					Name:        "pod0",
					Labels:      map[string]string{kaiconstants.DefaultNodePoolLabelKey: "node-pool-a"},
					NodeName:    "kind-worker2",
					Annotations: map[string]string{config.PodGroupAnnotationForPod: "pod-group-name-2"},
				},
				{
					Name:        "pod1",
					Labels:      map[string]string{kaiconstants.DefaultNodePoolLabelKey: "node-pool-a"},
					NodeName:    "kind-worker3",
					Annotations: map[string]string{config.PodGroupAnnotationForPod: "pod-group-name-3"},
				},
			},
			PodGroups: []TestPodGroup{
				{
					Name:   "pod-group-name-2",
					Labels: map[string]string{kaiconstants.DefaultNodePoolLabelKey: "node-pool-a"},
				},
				{
					Name:   "pod-group-name-3",
					Labels: map[string]string{kaiconstants.DefaultNodePoolLabelKey: "node-pool-a"},
				},
			},
			ExpectedNodePoolsState: map[string]ExpectedNodePoolState{
				"node-pool-a": {
					NodeNames: []string{},
					Status: v1alpha1.NodePoolStatus{
						Phase: v1alpha1.NodePoolEmpty,
					},
				},
				"node-pool-b": {
					NodeNames: []string{"kind-worker", "kind-worker2"},
					Status: v1alpha1.NodePoolStatus{
						Phase:   v1alpha1.NodePoolReady,
						Message: fmt.Sprintf(common.NodesAssignedToDifferentNodePoolWaitingForDrainMessage, "kind-worker2"),
						Nodes: []v1alpha1.NodeInNodePool{
							{
								Name:   "kind-worker",
								Status: v1alpha1.NodeReady,
							},
							{
								Name:   "kind-worker2",
								Status: v1alpha1.NodeUnschedulable,
							},
						},
					},
				},
				kaiconstants.DefaultNodePoolName: {
					NodeNames: []string{"kind-worker3"},
					Status: v1alpha1.NodePoolStatus{
						Phase: v1alpha1.NodePoolReady,
						Message: fmt.Sprintf(common.NodesAssignedToThisNodePoolWaitingForDrainMessage+"\n\n"+common.NodesAssignedToDifferentNodePoolWaitingForDrainMessage,
							"kind-worker2", "kind-worker3"),
					},
				},
			},
		}),
		Entry("Nodepool ready - pods running with different label than podGroup - node not unschedulable", &TestCase{
			Name: "Nodepool ready - pods running with different label than podGroup - node not unschedulable",
			NodePools: []TestNodePool{
				{
					Name:       "node-pool-b",
					LabelKey:   "nodePoolBKey",
					LabelValue: "nodePoolBValue",
					PlacementStrategy: PlacementStrategy{
						Gpu:       "binpack",
						GpuDevice: "binpack",
						Cpu:       "binpack",
					},
				},
			},
			Nodes: map[string]TestNode{
				"kind-worker": {
					Name: "kind-worker",
					Labels: map[string]string{
						"nodePoolBKey":                       "nodePoolBValue",
						kaiconstants.DefaultNodePoolLabelKey: "node-pool-b"},
				},
			},
			Pods: []TestPod{
				{
					Name:        "pod0",
					Labels:      map[string]string{kaiconstants.DefaultNodePoolLabelKey: ""},
					NodeName:    "kind-worker",
					Annotations: map[string]string{config.PodGroupAnnotationForPod: "pod-group-name-1"},
				},
			},
			PodGroups: []TestPodGroup{
				{
					Name:   "pod-group-name-1",
					Labels: map[string]string{kaiconstants.DefaultNodePoolLabelKey: "node-pool-b"},
				},
			},
			ExpectedNodePoolsState: map[string]ExpectedNodePoolState{
				"node-pool-b": {
					NodeNames: []string{"kind-worker"},
					Status: v1alpha1.NodePoolStatus{
						Phase:   v1alpha1.NodePoolReady,
						Message: "",
						Nodes: []v1alpha1.NodeInNodePool{
							{
								Name:   "kind-worker",
								Status: v1alpha1.NodeReady,
							},
						},
					},
				},
			},
		}),
		Entry("Nodepool ready - pods running without nodepool label  - node not unschedulable", &TestCase{
			Name: "Nodepool ready - pods running without nodepool label  - node not unschedulable",
			NodePools: []TestNodePool{
				{
					Name:       "node-pool-b",
					LabelKey:   "nodePoolBKey",
					LabelValue: "nodePoolBValue",
					PlacementStrategy: PlacementStrategy{
						Gpu:       "binpack",
						GpuDevice: "binpack",
						Cpu:       "binpack",
					},
				},
			},
			Nodes: map[string]TestNode{
				"kind-worker": {
					Name: "kind-worker",
					Labels: map[string]string{
						"nodePoolBKey":                       "nodePoolBValue",
						kaiconstants.DefaultNodePoolLabelKey: "node-pool-b"},
				},
			},
			Pods: []TestPod{
				{
					Name:        "pod0",
					Labels:      map[string]string{},
					NodeName:    "kind-worker",
					Annotations: map[string]string{config.PodGroupAnnotationForPod: "pod-group-name-1"},
				},
			},
			PodGroups: []TestPodGroup{
				{
					Name:   "pod-group-name-1",
					Labels: map[string]string{kaiconstants.DefaultNodePoolLabelKey: "node-pool-b"},
				},
			},
			ExpectedNodePoolsState: map[string]ExpectedNodePoolState{
				"node-pool-b": {
					NodeNames: []string{"kind-worker"},
					Status: v1alpha1.NodePoolStatus{
						Phase:   v1alpha1.NodePoolReady,
						Message: "",
						Nodes: []v1alpha1.NodeInNodePool{
							{
								Name:   "kind-worker",
								Status: v1alpha1.NodeReady,
							},
						},
					},
				},
			},
		}),
		Entry("Nodepool ready - pods running with different nodepool then terminating - node schedulable", &TestCase{
			Name: "Nodepool ready - pods running with different nodepool then terminating - node schedulable",
			NodePools: []TestNodePool{
				{
					Name:       "node-pool-a",
					LabelKey:   "nodePoolAKey",
					LabelValue: "nodePoolAValue",
					PlacementStrategy: PlacementStrategy{
						Gpu:       "binpack",
						GpuDevice: "binpack",
						Cpu:       "binpack",
					},
				},
				{
					Name:       "node-pool-b",
					LabelKey:   "nodePoolBKey",
					LabelValue: "nodePoolBValue",
					PlacementStrategy: PlacementStrategy{
						Gpu:       "binpack",
						GpuDevice: "binpack",
						Cpu:       "binpack",
					},
				},
			},
			Nodes: map[string]TestNode{
				"kind-worker": {
					Name: "kind-worker",
					Labels: map[string]string{
						"nodePoolBKey":                       "nodePoolBValue",
						kaiconstants.DefaultNodePoolLabelKey: "node-pool-b"},
				},
				"kind-worker2": {
					Name: "kind-worker2",
					Labels: map[string]string{
						"nodePoolBKey":                       "nodePoolBValue",
						kaiconstants.DefaultNodePoolLabelKey: "node-pool-b"},
				},
			},
			Pods: []TestPod{
				{
					Name:        "pod0",
					Labels:      map[string]string{kaiconstants.DefaultNodePoolLabelKey: "node-pool-a"},
					NodeName:    "kind-worker2",
					Annotations: map[string]string{config.PodGroupAnnotationForPod: "pod-group-name-1"},
				},
			},
			PodGroups: []TestPodGroup{
				{
					Name:   "pod-group-name-1",
					Labels: map[string]string{kaiconstants.DefaultNodePoolLabelKey: "node-pool-a"},
				},
			},
			TestCallback: func(k8sClient client.Client, testCase *TestCase) {
				pod := getPodFromClient("pod0", runaiNamespace, k8sClient)
				if pod != nil {
					DeleteImmediatelyAndPollUntilDeleted(k8sClient, pod)
				}
			},
			ExpectedNodePoolsState: map[string]ExpectedNodePoolState{
				"node-pool-a": {
					NodeNames: []string{},
					Status: v1alpha1.NodePoolStatus{
						Phase: v1alpha1.NodePoolEmpty,
					},
				},
				"node-pool-b": {
					NodeNames: []string{"kind-worker", "kind-worker2"},
					Status: v1alpha1.NodePoolStatus{
						Phase:   v1alpha1.NodePoolReady,
						Message: "",
						Nodes: []v1alpha1.NodeInNodePool{
							{
								Name:   "kind-worker",
								Status: v1alpha1.NodeReady,
							},
							{
								Name:   "kind-worker2",
								Status: v1alpha1.NodeReady,
							},
						},
					},
				},
			},
		}),
		Entry("Deleting NodePool - pods running, nodepool in deleting state", &TestCase{
			Name: "Deleting NodePool - pods running, nodepool in deleting state",
			NodePools: []TestNodePool{
				{
					Name:       "node-pool-a",
					LabelKey:   "nodePoolAKey",
					LabelValue: "nodePoolAValue",
				},
				{
					Name:       "node-pool-b",
					LabelKey:   "nodePoolBKey",
					LabelValue: "nodePoolBValue",
				},
			},
			Nodes: map[string]TestNode{
				"kind-worker": {
					Name:   "kind-worker",
					Labels: map[string]string{"nodePoolBKey": "nodePoolBValue", kaiconstants.DefaultNodePoolLabelKey: "node-pool-b"},
				},
				"kind-worker2": {
					Name:   "kind-worker2",
					Labels: map[string]string{"nodePoolAKey": "nodePoolAValue", kaiconstants.DefaultNodePoolLabelKey: "node-pool-a"},
				},
			},
			Pods: []TestPod{
				{
					Name:        "pod0",
					Labels:      map[string]string{kaiconstants.DefaultNodePoolLabelKey: "node-pool-a"},
					NodeName:    "kind-worker2",
					Annotations: map[string]string{config.PodGroupAnnotationForPod: "pod-group-name-1"},
				},
			},
			PodGroups: []TestPodGroup{
				{
					Name:   "pod-group-name-1",
					Labels: map[string]string{kaiconstants.DefaultNodePoolLabelKey: "node-pool-a"},
				},
			},
			TestCallback: func(k8sClient client.Client, testCase *TestCase) {
				updateNodeLabels("kind-worker2", map[string]string{"nodePoolBKey": "nodePoolBValue"}, "", k8sClient)
				nodePool := getNodePoolFromClient("node-pool-a", k8sClient)
				_ = k8sClient.Delete(ctx, nodePool)
			},
			ExpectedNodePoolsState: map[string]ExpectedNodePoolState{
				"node-pool-a": {
					NodeNames: []string{"kind-worker2"},
					Status: v1alpha1.NodePoolStatus{
						Phase:   v1alpha1.NodePoolDeleting,
						Message: "Nodes Being Drained: kind-worker2",
						Nodes: []v1alpha1.NodeInNodePool{
							{
								Name:   "kind-worker2",
								Status: v1alpha1.NodeUnschedulable,
							},
						},
					},
				},
				"node-pool-b": {
					NodeNames: []string{"kind-worker"},
					Status: v1alpha1.NodePoolStatus{
						Phase:   v1alpha1.NodePoolReady,
						Message: fmt.Sprintf(common.NodesAssignedToThisNodePoolWaitingForDrainMessage, "kind-worker2"),
						Nodes: []v1alpha1.NodeInNodePool{
							{
								Name:   "kind-worker",
								Status: v1alpha1.NodeReady,
							},
						},
					},
				},
			},
		}),
		Entry("Deleting NodePool - pods running then terminating, nodepool is deleted", &TestCase{
			Name: "Deleting NodePool - pods running then terminating, nodepool is deleted",
			NodePools: []TestNodePool{
				{
					Name:       "node-pool-a",
					LabelKey:   "nodePoolAKey",
					LabelValue: "nodePoolAValue",
					PlacementStrategy: PlacementStrategy{
						Gpu:       "binpack",
						GpuDevice: "binpack",
						Cpu:       "binpack",
					},
				},
				{
					Name:       "node-pool-b",
					LabelKey:   "nodePoolBKey",
					LabelValue: "nodePoolBValue",
					PlacementStrategy: PlacementStrategy{
						Gpu:       "binpack",
						GpuDevice: "binpack",
						Cpu:       "binpack",
					},
				},
			},
			Nodes: map[string]TestNode{
				"kind-worker": {
					Name:   "kind-worker",
					Labels: map[string]string{"nodePoolBKey": "nodePoolBValue", kaiconstants.DefaultNodePoolLabelKey: "node-pool-b"},
				},
				"kind-worker2": {
					Name:   "kind-worker2",
					Labels: map[string]string{"nodePoolAKey": "nodePoolAValue", kaiconstants.DefaultNodePoolLabelKey: "node-pool-a"},
				},
			},
			Pods: []TestPod{
				{
					Name:        "pod0",
					Labels:      map[string]string{kaiconstants.DefaultNodePoolLabelKey: "node-pool-a"},
					NodeName:    "kind-worker2",
					Annotations: map[string]string{config.PodGroupAnnotationForPod: "pod-group-name-1"},
				},
			},
			PodGroups: []TestPodGroup{
				{
					Name:   "pod-group-name-1",
					Labels: map[string]string{kaiconstants.DefaultNodePoolLabelKey: "node-pool-a"},
				},
			},
			TestCallback: func(k8sClient client.Client, testCase *TestCase) {
				updateNodeLabels("kind-worker2", map[string]string{"nodePoolBKey": "nodePoolBValue"}, "", k8sClient)
				nodePool := getNodePoolFromClient("node-pool-a", k8sClient)
				_ = k8sClient.Delete(ctx, nodePool)

				pod := getPodFromClient("pod0", runaiNamespace, k8sClient)
				if pod != nil {
					DeleteImmediatelyAndPollUntilDeleted(k8sClient, pod)
				}
			},
			ExpectedNodePoolsState: map[string]ExpectedNodePoolState{
				"node-pool-a": {},
				"node-pool-b": {
					NodeNames: []string{"kind-worker", "kind-worker2"},
					Status: v1alpha1.NodePoolStatus{
						Phase: v1alpha1.NodePoolReady,
						Nodes: []v1alpha1.NodeInNodePool{
							{
								Name:   "kind-worker",
								Status: v1alpha1.NodeReady,
							},
							{
								Name:   "kind-worker2",
								Status: v1alpha1.NodeReady,
							},
						},
					},
				},
			},
		}),
		Entry("Deleting NodePool - blocked by project queue reference, stays in deleting state", &TestCase{
			Name: "Deleting NodePool - blocked by project queue reference, stays in deleting state",
			NodePools: []TestNodePool{
				{
					Name:       "node-pool-a",
					LabelKey:   "nodePoolAKey",
					LabelValue: "nodePoolAValue",
				},
			},
			Projects: []TestProject{
				{
					Name:           "proj-a",
					QueueNodePools: []string{"node-pool-a"},
				},
			},
			TestCallback: func(k8sClient client.Client, testCase *TestCase) {
				nodePool := getNodePoolFromClient("node-pool-a", k8sClient)
				_ = k8sClient.Delete(ctx, nodePool)
			},
			ExpectedNodePoolsState: map[string]ExpectedNodePoolState{
				"node-pool-a": {
					Status: v1alpha1.NodePoolStatus{
						Phase:   v1alpha1.NodePoolDeleting,
						Message: fmt.Sprintf(common.ProjectsReferencingNodePoolMessage, "proj-a"),
					},
					ExpectedConditions: []ExpectedCondition{
						{
							Type:   string(v1alpha1.ProjectReferencesExist),
							Status: "True",
							Reason: v1alpha1.ProjectReferencesExistReason,
						},
					},
				},
			},
		}),
		Entry("Deleting NodePool - project queue reference removed, nodepool is deleted", &TestCase{
			Name: "Deleting NodePool - project queue reference removed, nodepool is deleted",
			NodePools: []TestNodePool{
				{
					Name:       "node-pool-a",
					LabelKey:   "nodePoolAKey",
					LabelValue: "nodePoolAValue",
				},
			},
			Projects: []TestProject{
				{
					Name:           "proj-a",
					QueueNodePools: []string{"node-pool-a"},
				},
			},
			TestCallback: func(k8sClient client.Client, testCase *TestCase) {
				project := &v1alpha1.Project{}
				_ = k8sClient.Get(ctx, types.NamespacedName{Name: "proj-a"}, project)
				project.Spec.Queues = nil
				_ = k8sClient.Update(ctx, project)

				nodePool := getNodePoolFromClient("node-pool-a", k8sClient)
				_ = k8sClient.Delete(ctx, nodePool)
			},
			ExpectedNodePoolsState: map[string]ExpectedNodePoolState{
				"node-pool-a": {},
			},
		}),
		Entry("Deleting NodePool - default-nodepool-list reference does not block deletion", &TestCase{
			Name: "Deleting NodePool - default-nodepool-list reference does not block deletion",
			NodePools: []TestNodePool{
				{
					Name:       "node-pool-a",
					LabelKey:   "nodePoolAKey",
					LabelValue: "nodePoolAValue",
				},
			},
			Projects: []TestProject{
				{
					Name:             "proj-a",
					DefaultNodePools: []string{"node-pool-a"},
				},
			},
			TestCallback: func(k8sClient client.Client, testCase *TestCase) {
				nodePool := getNodePoolFromClient("node-pool-a", k8sClient)
				_ = k8sClient.Delete(ctx, nodePool)
			},
			ExpectedNodePoolsState: map[string]ExpectedNodePoolState{
				"node-pool-a": {},
			},
		}),
		Entry("Deleting NodePool - unrelated project reference does not block deletion", &TestCase{
			Name: "Deleting NodePool - unrelated project reference does not block deletion",
			NodePools: []TestNodePool{
				{
					Name:       "node-pool-a",
					LabelKey:   "nodePoolAKey",
					LabelValue: "nodePoolAValue",
				},
				{
					Name:       "node-pool-b",
					LabelKey:   "nodePoolBKey",
					LabelValue: "nodePoolBValue",
				},
			},
			Projects: []TestProject{
				{
					Name:           "proj-b",
					QueueNodePools: []string{"node-pool-b"},
				},
			},
			TestCallback: func(k8sClient client.Client, testCase *TestCase) {
				nodePool := getNodePoolFromClient("node-pool-a", k8sClient)
				_ = k8sClient.Delete(ctx, nodePool)
			},
			ExpectedNodePoolsState: map[string]ExpectedNodePoolState{
				"node-pool-a": {},
				"node-pool-b": {
					Status: v1alpha1.NodePoolStatus{
						Phase: v1alpha1.NodePoolEmpty,
					},
				},
			},
		}),
		Entry("Deleting NodePool - pods running, nodepool in deleting state, don't assign new nodes to it", &TestCase{
			Name: "Deleting NodePool - pods running, nodepool in deleting state, don't assign new nodes to it",
			NodePools: []TestNodePool{
				{
					Name:       "node-pool-a",
					LabelKey:   "nodePoolAKey",
					LabelValue: "nodePoolAValue",
				},
				{
					Name:       "node-pool-b",
					LabelKey:   "nodePoolBKey",
					LabelValue: "nodePoolBValue",
				},
			},
			Nodes: map[string]TestNode{
				"kind-worker": {
					Name:   "kind-worker",
					Labels: map[string]string{"nodePoolBKey": "nodePoolBValue", kaiconstants.DefaultNodePoolLabelKey: "node-pool-b"},
				},
				"kind-worker2": {
					Name:   "kind-worker2",
					Labels: map[string]string{"nodePoolAKey": "nodePoolAValue", kaiconstants.DefaultNodePoolLabelKey: "node-pool-a"},
				},
			},
			Pods: []TestPod{
				{
					Name:        "pod0",
					Labels:      map[string]string{kaiconstants.DefaultNodePoolLabelKey: "node-pool-a"},
					NodeName:    "kind-worker2",
					Annotations: map[string]string{config.PodGroupAnnotationForPod: "pod-group-name-1"},
				},
			},
			PodGroups: []TestPodGroup{
				{
					Name:   "pod-group-name-1",
					Labels: map[string]string{kaiconstants.DefaultNodePoolLabelKey: "node-pool-a"},
				},
			},
			TestCallback: func(k8sClient client.Client, testCase *TestCase) {
				updateNodeLabels("kind-worker2", map[string]string{"nodePoolBKey": "nodePoolBValue"}, "", k8sClient)
				nodePoolA := getNodePoolFromClient("node-pool-a", k8sClient)
				_ = k8sClient.Delete(ctx, nodePoolA)

				updateNodeLabels("kind-worker", map[string]string{"nodePoolAKey": "nodePoolAValue"}, "nodePoolBKey", k8sClient)
			},
			ExpectedNodePoolsState: map[string]ExpectedNodePoolState{
				"node-pool-a": {
					NodeNames: []string{"kind-worker2"},
					Status: v1alpha1.NodePoolStatus{
						Phase:   v1alpha1.NodePoolDeleting,
						Message: "Nodes Being Drained: kind-worker2",
						Nodes: []v1alpha1.NodeInNodePool{
							{
								Name:   "kind-worker2",
								Status: v1alpha1.NodeUnschedulable,
							},
						},
					},
				},
				"node-pool-b": {
					NodeNames: []string{},
					Status: v1alpha1.NodePoolStatus{
						Phase:   v1alpha1.NodePoolEmpty,
						Message: fmt.Sprintf(common.NodesAssignedToThisNodePoolWaitingForDrainMessage, "kind-worker2"),
					},
				},
				kaiconstants.DefaultNodePoolName: {
					NodeNames: []string{"kind-worker"},
				},
			},
		}),
		Entry("SchedulingShard not ready - nodepool not ready", &TestCase{
			Name: "SchedulingShard not ready - nodepool not ready",
			NodePools: []TestNodePool{
				{
					Name:       "notready-np-notready-shard",
					LabelKey:   "nodePoolAKey",
					LabelValue: "nodePoolAValue",
					PlacementStrategy: PlacementStrategy{
						Gpu:       "binpack",
						GpuDevice: "binpack",
						Cpu:       "binpack",
					},
				},
			},
			Nodes: map[string]TestNode{
				"kind-worker": {
					Name:   "kind-worker",
					Labels: map[string]string{"nodePoolAKey": "nodePoolAValue", kaiconstants.DefaultNodePoolLabelKey: "notready-np-notready-shard"},
				},
			},
			ExpectedNodePoolsState: map[string]ExpectedNodePoolState{
				"notready-np-notready-shard": {
					NodeNames: []string{"kind-worker"},
					Status: v1alpha1.NodePoolStatus{
						Phase: v1alpha1.NodePoolUnschedulable,
						Message: fmt.Sprintf("%v, reason: %v", common.SchedulerNotReadyMessage,
							"scheduler [notready-np-notready-shard] is not running yet: no status message available"),
						Nodes: []v1alpha1.NodeInNodePool{
							{
								Name:   "kind-worker",
								Status: v1alpha1.NodeReady,
							},
						},
					},
				},
			},
		}),
		Entry("SchedulingShard deleted - nodepool is Unschedulable cause SchedulingShard was created but its status was not updated", &TestCase{
			Name: "SchedulingShard deleted - nodepool is Unschedulable cause SchedulingShard was created but its status was not updated",
			NodePools: []TestNodePool{
				{
					Name:       "node-pool-a",
					LabelKey:   "nodePoolAKey",
					LabelValue: "nodePoolAValue",
					PlacementStrategy: PlacementStrategy{
						Gpu:       "binpack",
						GpuDevice: "binpack",
						Cpu:       "binpack",
					},
				},
			},
			Nodes: map[string]TestNode{
				"kind-worker": {
					Name:   "kind-worker",
					Labels: map[string]string{"nodePoolAKey": "nodePoolAValue", kaiconstants.DefaultNodePoolLabelKey: "node-pool-a"},
				},
			},
			TestCallback: func(k8sClient client.Client, testCase *TestCase) {
				shard := getSchedulingShardFromClient("node-pool-a", k8sClient)

				var gracePeriodSeconds int64 = 0
				deleteOptions := &client.DeleteOptions{GracePeriodSeconds: &gracePeriodSeconds}
				deleteFn := func() error {
					err := k8sClient.Delete(ctx, shard, deleteOptions)
					if err == nil || errors.IsNotFound(err) {
						return nil
					}
					return err
				}
				ExpectWithOffset(1, deleteFn()).To(Succeed())
			},
			ExpectedNodePoolsState: map[string]ExpectedNodePoolState{
				"node-pool-a": {
					NodeNames: []string{"kind-worker"},
					Status: v1alpha1.NodePoolStatus{
						Phase:   v1alpha1.NodePoolUnschedulable,
						Message: "Scheduler is not ready, reason: scheduler [node-pool-a] is not running yet: no status message available",
						Nodes: []v1alpha1.NodeInNodePool{
							{
								Name:   "kind-worker",
								Status: v1alpha1.NodeReady,
							},
						},
					},
				},
			},
		}),
		Entry("ServiceMonitor deleted - nodepool is Unschedulable cause ServiceMonitor was created but its status was not updated", &TestCase{
			Name: "ServiceMonitor deleted - nodepool is Unschedulable cause ServiceMonitor was created but its status was not updated",
			NodePools: []TestNodePool{
				{
					Name:       "node-pool-a",
					LabelKey:   "nodePoolAKey",
					LabelValue: "nodePoolAValue",
					PlacementStrategy: PlacementStrategy{
						Gpu:       "binpack",
						GpuDevice: "binpack",
						Cpu:       "binpack",
					},
				},
			},
			Nodes: map[string]TestNode{
				"kind-worker": {
					Name:   "kind-worker",
					Labels: map[string]string{"nodePoolAKey": "nodePoolAValue", kaiconstants.DefaultNodePoolLabelKey: "node-pool-a"},
				},
			},
			TestCallback: func(k8sClient client.Client, testCase *TestCase) {
				schedulerName := getOperandName("node-pool-a")
				serviceMonitor := getServiceMonitorFromClient(schedulerName, k8sClient)

				var gracePeriodSeconds int64 = 0
				deleteOptions := &client.DeleteOptions{GracePeriodSeconds: &gracePeriodSeconds}
				deleteFn := func() error {
					err := k8sClient.Delete(ctx, serviceMonitor, deleteOptions)
					if err == nil || errors.IsNotFound(err) {
						return nil
					}
					return err
				}
				ExpectWithOffset(1, deleteFn()).To(Succeed())
			},
			ExpectedNodePoolsState: map[string]ExpectedNodePoolState{
				"node-pool-a": {
					NodeNames: []string{"kind-worker"},
					Status: v1alpha1.NodePoolStatus{
						Phase: v1alpha1.NodePoolUnschedulable,
						// it's not really missing but go client creates it without UID so the serviceMonitor.status thinks it doesn't exist..
						Message: "Scheduler is not ready, reason: service monitor [runai-scheduler-node-pool-a] is missing",
						Nodes: []v1alpha1.NodeInNodePool{
							{
								Name:   "kind-worker",
								Status: v1alpha1.NodeReady,
							},
						},
					},
				},
			},
		}),
		Entry("Node Unschedulable (manually by outer source) - not marking it as Schedulable", &TestCase{
			Name: "Node Unschedulable (manually by outer source) - not marking it as Schedulable",
			NodePools: []TestNodePool{
				{
					Name:       "node-pool-a",
					LabelKey:   "nodePool0Key",
					LabelValue: "nodePool0Value",
					PlacementStrategy: PlacementStrategy{
						Gpu:       "binpack",
						GpuDevice: "binpack",
						Cpu:       "binpack",
					},
				},
			},
			Nodes: map[string]TestNode{
				"kind-worker": {
					Name:          "kind-worker",
					Labels:        map[string]string{"nodePool0Key": "nodePool0Value", kaiconstants.DefaultNodePoolLabelKey: "node-pool-a"},
					Unschedulable: true,
				},
				"kind-worker2": {
					Name:   "kind-worker2",
					Labels: map[string]string{"nodePool0Key": "nodePool0Value", kaiconstants.DefaultNodePoolLabelKey: "node-pool-a"},
				},
			},
			ExpectedNodePoolsState: map[string]ExpectedNodePoolState{
				"node-pool-a": {
					NodeNames: []string{"kind-worker", "kind-worker2"},
					Status: v1alpha1.NodePoolStatus{
						Phase:   v1alpha1.NodePoolReady,
						Message: fmt.Sprintf(common.NodesNotReadyMessage, "kind-worker"),
						Nodes: []v1alpha1.NodeInNodePool{
							{
								Name:   "kind-worker",
								Status: v1alpha1.NodeUnschedulable,
							},
							{
								Name:   "kind-worker2",
								Status: v1alpha1.NodeReady,
							},
						},
					},
				},
			},
		}),
		Entry("Node Unschedulable by us - marking it as Schedulable", &TestCase{
			Name: "Node Unschedulable by us - marking it as Schedulable",
			NodePools: []TestNodePool{
				{
					Name:       "node-pool-a",
					LabelKey:   "nodePool0Key",
					LabelValue: "nodePool0Value",
					PlacementStrategy: PlacementStrategy{
						Gpu:       "binpack",
						GpuDevice: "binpack",
						Cpu:       "binpack",
					},
				},
			},
			Nodes: map[string]TestNode{
				"kind-worker": {
					Name: "kind-worker",
					Labels: map[string]string{
						"nodePool0Key":                       "nodePool0Value",
						kaiconstants.DefaultNodePoolLabelKey: "node-pool-a",
						testUnschedulableLabel:               "true"},
					Unschedulable: true,
				},
				"kind-worker2": {
					Name:   "kind-worker2",
					Labels: map[string]string{"nodePool0Key": "nodePool0Value", kaiconstants.DefaultNodePoolLabelKey: "node-pool-a"},
				},
			},
			ExpectedNodePoolsState: map[string]ExpectedNodePoolState{
				"node-pool-a": {
					NodeNames: []string{"kind-worker", "kind-worker2"},
					Status: v1alpha1.NodePoolStatus{
						Phase: v1alpha1.NodePoolReady,
						Nodes: []v1alpha1.NodeInNodePool{
							{
								Name:   "kind-worker",
								Status: v1alpha1.NodeReady,
							},
							{
								Name:   "kind-worker2",
								Status: v1alpha1.NodeReady,
							},
						},
					},
				},
			},
		}),
		Entry("Node added - matching node that was in default tagged in new nodepool", &TestCase{
			Name: "Node added - matching node that was in default tagged in new nodepool",
			NodePools: []TestNodePool{
				{
					Name:                     "node-pool-a",
					LabelKey:                 "nodePool0Key",
					LabelValue:               "nodePool0Value",
					CreateDuringTestCallback: true,
					PlacementStrategy: PlacementStrategy{
						Gpu:       "binpack",
						GpuDevice: "binpack",
						Cpu:       "binpack",
					},
				},
				{
					Name:       "node-pool-b",
					LabelKey:   "nodePoolBKey",
					LabelValue: "nodePoolBValue",
					PlacementStrategy: PlacementStrategy{
						Gpu:       "binpack",
						GpuDevice: "binpack",
						Cpu:       "binpack",
					},
				},
			},
			Nodes: map[string]TestNode{
				"kind-control-plane": {
					Name:   "kind-control-plane",
					Labels: map[string]string{"nodePoolBKey": "nodePoolBValue", kaiconstants.DefaultNodePoolLabelKey: "node-pool-b"},
				},
				"kind-worker": {
					Name:   "kind-worker",
					Labels: map[string]string{"nodePool0Key": "nodePool0Value"},
				},
				"kind-worker2": {
					Name:   "kind-worker2",
					Labels: map[string]string{"nodePoolBKey": "nodePoolBValue", kaiconstants.DefaultNodePoolLabelKey: "node-pool-b"},
				},
				"kind-worker3": {
					Name:   "kind-worker3",
					Labels: map[string]string{"nodePoolBKey": "nodePoolBValue", kaiconstants.DefaultNodePoolLabelKey: "node-pool-b"},
				},
				"kind-worker4": {
					Name:   "kind-worker4",
					Labels: map[string]string{"nodePoolBKey": "nodePoolBValue", kaiconstants.DefaultNodePoolLabelKey: "node-pool-b"},
				},
				"kind-worker5": {
					Name:   "kind-worker5",
					Labels: map[string]string{"nodePoolBKey": "nodePoolBValue", kaiconstants.DefaultNodePoolLabelKey: "node-pool-b"},
				},
			},
			TestCallback: func(k8sClient client.Client, testCase *TestCase) {
				nodePoolA := TestNodePool{
					Name:       "node-pool-a",
					LabelKey:   "nodePool0Key",
					LabelValue: "nodePool0Value",
					PlacementStrategy: PlacementStrategy{
						Gpu:       "binpack",
						GpuDevice: "binpack",
						Cpu:       "binpack",
					},
					CreateDuringTestCallback: true,
				}

				By(fmt.Sprintf("Test: %v, creating nodepool: %v", "Node added - matching node that was in default tagged in new nodepool", nodePoolA.Name))
				nodePool := getTestNodePoolObj(&nodePoolA)
				ExpectCreateResource(k8sClient, nodePool)
			},
			ExpectedNodePoolsState: map[string]ExpectedNodePoolState{
				"node-pool-a": {
					NodeNames: []string{"kind-worker"},
					Status: v1alpha1.NodePoolStatus{
						Phase: v1alpha1.NodePoolReady,
						Nodes: []v1alpha1.NodeInNodePool{
							{
								Name:   "kind-worker",
								Status: v1alpha1.NodeReady,
							},
						},
					},
				},
				"node-pool-b": {
					NodeNames: []string{"kind-control-plane", "kind-worker2", "kind-worker3", "kind-worker4", "kind-worker5"},
					Status: v1alpha1.NodePoolStatus{
						Phase: v1alpha1.NodePoolReady,
						Nodes: []v1alpha1.NodeInNodePool{
							{
								Name:   "kind-control-plane",
								Status: v1alpha1.NodeReady,
							},
							{
								Name:   "kind-worker2",
								Status: v1alpha1.NodeReady,
							},
							{
								Name:   "kind-worker3",
								Status: v1alpha1.NodeReady,
							},
							{
								Name:   "kind-worker4",
								Status: v1alpha1.NodeReady,
							},
							{
								Name:   "kind-worker5",
								Status: v1alpha1.NodeReady,
							},
						},
					},
				},
				kaiconstants.DefaultNodePoolName: {},
			},
			ExpectedMetricsStateList: []ExpectedMetricsState{
				{
					NodeName:      "kind-worker",
					NodePoolsName: "default",
					Count:         0,
				},
				{
					NodeName:      "kind-worker",
					NodePoolsName: "node-pool-a",
					Count:         1,
				},
				{
					NodeName:      "kind-worker",
					NodePoolsName: "node-pool-b",
					Count:         0,
				},
				{
					NodeName:      "kind-worker2",
					NodePoolsName: "default",
					Count:         0,
				},
				{
					NodeName:      "kind-worker2",
					NodePoolsName: "node-pool-a",
					Count:         0,
				},
				{
					NodeName:      "kind-worker2",
					NodePoolsName: "node-pool-b",
					Count:         1,
				},
				{
					NodeName:      "kind-worker3",
					NodePoolsName: "default",
					Count:         0,
				},
				{
					NodeName:      "kind-worker3",
					NodePoolsName: "node-pool-a",
					Count:         0,
				},
				{
					NodeName:      "kind-worker3",
					NodePoolsName: "node-pool-b",
					Count:         1,
				},
				{
					NodeName:      "kind-worker4",
					NodePoolsName: "default",
					Count:         0,
				},
				{
					NodeName:      "kind-worker4",
					NodePoolsName: "node-pool-a",
					Count:         0,
				},
				{
					NodeName:      "kind-worker4",
					NodePoolsName: "node-pool-b",
					Count:         1,
				},
				{
					NodeName:      "kind-worker5",
					NodePoolsName: "default",
					Count:         0,
				},
				{
					NodeName:      "kind-worker5",
					NodePoolsName: "node-pool-a",
					Count:         0,
				},
				{
					NodeName:      "kind-worker5",
					NodePoolsName: "node-pool-b",
					Count:         1,
				},
			},
		}),
		Entry("Managed nodes config - Exclude some nodes", &TestCase{
			Name: "Managed nodes config - Exclude some nodes",
			NodePools: []TestNodePool{
				{
					Name:       "node-pool-a",
					LabelKey:   "nodePool0Key",
					LabelValue: "nodePool0Value",
					PlacementStrategy: PlacementStrategy{
						Gpu:       "binpack",
						GpuDevice: "binpack",
						Cpu:       "binpack",
					},
				},
			},
			Nodes: map[string]TestNode{
				"kind-worker": {
					Name:   "kind-worker",
					Labels: map[string]string{"kubernetes.io/os": "linux", "nodePool0Key": "nodePool0Value"},
				},
				"kind-worker2": {
					Name:   "kind-worker2",
					Labels: map[string]string{"kubernetes.io/os": "linux", "nodePool0Key": "nodePool0Value", kaiconstants.DefaultNodePoolLabelKey: "node-pool-a"},
				},
				"kind-worker3": {
					Name: "kind-worker3",
					// Test moving inclusion of a node that wrongly excluded
					Labels: map[string]string{"kubernetes.io/os": "linux", kaiconstants.DefaultNodePoolLabelKey: testExcludedNodepoolName},
				},
				"kind-worker4": {
					Name:   "kind-worker4",
					Labels: map[string]string{"kubernetes.io/os": "linux", kaiconstants.DefaultNodePoolLabelKey: "node-pool-a"},
				},
				"kind-worker5": {
					Name: "kind-worker5",
					// Test exclusion
					Labels: map[string]string{},
				},
				"kind-worker6": {
					Name: "kind-worker6",
					// Test exclusion of a node in different nodepool
					Labels: map[string]string{"nodePool0Key": "nodePool0Value", kaiconstants.DefaultNodePoolLabelKey: "node-pool-a"},
				},
				"kind-worker7": {
					Name: "kind-worker7",
					// Test exclusion of a node that is already excluded
					Labels: map[string]string{kaiconstants.DefaultNodePoolLabelKey: testExcludedNodepoolName},
				},
			},
			ManagedNodesConfig: []TestManagedNodesConfig{
				{
					Name: testManagedNodesConfigName,
					NodeSelector: corev1.NodeSelector{
						NodeSelectorTerms: []corev1.NodeSelectorTerm{
							{
								MatchExpressions: []corev1.NodeSelectorRequirement{
									{
										Key:      "kubernetes.io/os",
										Operator: corev1.NodeSelectorOpIn,
										Values: []string{
											"linux",
										},
									},
								},
							},
						},
					},
				},
			},
			ExpectedNodePoolsState: map[string]ExpectedNodePoolState{
				"node-pool-a": {
					NodeNames: []string{"kind-worker", "kind-worker2"},
					Status: v1alpha1.NodePoolStatus{
						Phase: v1alpha1.NodePoolReady,
						Nodes: []v1alpha1.NodeInNodePool{
							{
								Name:   "kind-worker",
								Status: v1alpha1.NodeReady,
							},
							{
								Name:   "kind-worker2",
								Status: v1alpha1.NodeReady,
							},
						},
					},
				},
				kaiconstants.DefaultNodePoolName: {
					NodeNames: []string{"kind-worker3", "kind-worker4"},
					Status: v1alpha1.NodePoolStatus{
						Phase: v1alpha1.NodePoolReady,
						Nodes: []v1alpha1.NodeInNodePool{
							{
								Name:   "kind-worker3",
								Status: v1alpha1.NodeReady,
							},
							{
								Name:   "kind-worker4",
								Status: v1alpha1.NodeReady,
							},
						},
					}},
			},
			ExpectedNodeLabels: map[string]ExpectedNodeLabels{
				"kind-worker": {
					expectedExistingLabels: map[string]string{
						"kubernetes.io/os":                   "linux",
						"nodePool0Key":                       "nodePool0Value",
						kaiconstants.DefaultNodePoolLabelKey: "node-pool-a",
					},
				},
				"kind-worker2": {
					expectedExistingLabels: map[string]string{
						"kubernetes.io/os":                   "linux",
						"nodePool0Key":                       "nodePool0Value",
						kaiconstants.DefaultNodePoolLabelKey: "node-pool-a",
					},
				},
				"kind-worker3": {
					expectedNonExistingLabels: []string{
						kaiconstants.DefaultNodePoolLabelKey,
					},
				},
				"kind-worker4": {
					expectedNonExistingLabels: []string{
						kaiconstants.DefaultNodePoolLabelKey,
					},
				},
				"kind-worker5": {
					expectedExistingLabels: map[string]string{
						kaiconstants.DefaultNodePoolLabelKey: testExcludedNodepoolName,
					},
				},
				"kind-worker6": {
					expectedExistingLabels: map[string]string{
						"nodePool0Key":                       "nodePool0Value",
						kaiconstants.DefaultNodePoolLabelKey: testExcludedNodepoolName,
					},
				},
				"kind-worker7": {
					expectedExistingLabels: map[string]string{
						kaiconstants.DefaultNodePoolLabelKey: testExcludedNodepoolName,
					},
				},
			},
		}),
		Entry("Topology validation - Nodes with all required labels - No conditions", NetworkTopologyTestCaseBase(
			"Topology validation - Nodes with all required labels - No conditions",
			nil,
			[]string{},
		).WithExpectedNodepoolState("node-pool-topology", ExpectedNodePoolState{
			NodeNames: []string{"kind-worker", "kind-worker2", "kind-worker3"},
			Status: v1alpha1.NodePoolStatus{
				Phase: v1alpha1.NodePoolReady,
				Nodes: []v1alpha1.NodeInNodePool{
					{
						Name:   "kind-worker",
						Status: v1alpha1.NodeReady,
					},
					{
						Name:   "kind-worker2",
						Status: v1alpha1.NodeReady,
					},
					{
						Name:   "kind-worker3",
						Status: v1alpha1.NodeReady,
					},
				},
			},
			ExpectedConditions: []ExpectedCondition{},
		})),
		Entry("Topology validation - Node missing one label - Conditions set", NetworkTopologyTestCaseBase(
			"Topology validation - Node missing one label - Conditions set",
			[]ExpectedCondition{
				{
					Type:   string(v1alpha1.NodeTopologyMismatch),
					Status: "True",
					Reason: v1alpha1.NodeTopologyMismatchReason,
				},
			},
			[]string{"kind-worker"},
		).WithNodes(createNodesWithMissingTopologyLabel()).WithExpectedExistingNodesLabels(createNodesWithMissingTopologyLabel())),
		Entry("Topology validation - Node missing one label - Conditions set", NetworkTopologyTestCaseBase(
			"Topology validation - Node missing one label - Conditions set",
			[]ExpectedCondition{
				{
					Type:   string(v1alpha1.NodeTopologyMismatch),
					Status: "True",
					Reason: v1alpha1.NodeTopologyMismatchReason,
				},
			},
			[]string{"kind-worker"},
		).WithNodes(createNodesWithMissingTopologyLabel()).WithExpectedExistingNodesLabels(createNodesWithMissingTopologyLabel())),

		Entry("Topology validation - Node label restored - Conditions cleared", createDynamicTopologyTestCase()),

		Entry("Topology validation - Dynamic label changes - Conditions updated", NetworkTopologyTestCaseBase(
			"Topology validation - Dynamic label changes - Conditions updated",
			[]ExpectedCondition{
				{
					Type:   string(v1alpha1.NodeTopologyMismatch),
					Status: "True",
					Reason: v1alpha1.NodeTopologyMismatchReason,
				},
			},
			[]string{"kind-worker"},
		).WithUpdatedNodesLabels("kind-worker", map[string]string{
			"nodePoolTopologyKey": "nodePoolTopologyValue",
			"topology-level-2":    "value2",
			"topology-level-3":    "value3",
		}).WithExpectedMismatchNodeName("kind-worker").
			WithExpectedNodesLabels("kind-worker", map[string]string{
				"nodePoolTopologyKey": "nodePoolTopologyValue",
				"topology-level-2":    "value2",
				"topology-level-3":    "value3",
			})),
	)
})

func getWrongNameTestManagedNodesConfig() TestManagedNodesConfig {
	return TestManagedNodesConfig{
		Name: "some-wrong-name",
		NodeSelector: corev1.NodeSelector{
			NodeSelectorTerms: []corev1.NodeSelectorTerm{
				{
					MatchExpressions: []corev1.NodeSelectorRequirement{
						{
							Key:      "kubernetes.io/os",
							Operator: corev1.NodeSelectorOpIn,
							Values: []string{
								"linux",
							},
						},
						{
							Key:      "kubernetes.io/arch",
							Operator: corev1.NodeSelectorOpIn,
							Values: []string{
								"amd64",
							},
						},
					},
				},
			},
		},
	}
}
func getEmptyTestManagedNodesConfig() TestManagedNodesConfig {
	return TestManagedNodesConfig{
		Name:         testManagedNodesConfigName,
		NodeSelector: corev1.NodeSelector{},
	}
}

func validateNodesNVLinkLabel(node corev1.Node, nodePool *v1alpha1.NodePool, g Gomega) {
	isSetGPUNWAccelerationLabelKey := v1alpha1.GPUNetworkAccelerationDetection(
		nodePool.Annotations[v1alpha1.AnnotationGPUNetworkAccelerationDetection]) ==
		v1alpha1.AutoGPUNetworkAccelerationDetection
	labelKey := nodepool_controller.GetMNNVLLabelOtDefault(nodePool.Annotations[v1alpha1.AnnotationGPUNetworkAccelerationLabelKey])
	if isSetGPUNWAccelerationLabelKey {
		g.Expect(node.Annotations[nodepool_controller.GPUNetworkAccelerationLabelKey]).To(Equal(labelKey))
	} else {
		_, found := node.Annotations[nodepool_controller.GPUNetworkAccelerationLabelKey]
		g.Expect(found).To(BeFalse())
	}

}

func validateExpected(testCase *TestCase, k8sClient client.Client, g Gomega) {
	By(fmt.Sprintf("Validating TestCase: <%v>", testCase.Name))

	for nodePoolName, expectedNodePoolState := range testCase.ExpectedNodePoolsState {
		nodePool := getNodePoolFromClient(nodePoolName, k8sClient)
		if nodePool == nil {
			g.Expect(expectedNodePoolState.NodeNames).To(BeEmpty(),
				"Test: <%v>, expected nodepool <%v> to be deleted - to have empty expectedNodePoolState.NodeNames",
				testCase.Name, nodePoolName)
			g.Expect(expectedNodePoolState.Status.Phase).To(BeEmpty(),
				"Test: <%v>, expected nodepool <%v> to be deleted - to have empty expectedNodePoolState.Status.Phase",
				testCase.Name, nodePoolName)
			g.Expect(expectedNodePoolState.Status.Message).To(BeEmpty(),
				"Test: <%v>, expected nodepool <%v> to be deleted - to have empty expectedNodePoolState.Status.Message",
				testCase.Name, nodePoolName)
			continue
		}

		// Validate Node has TopologyMismatch
		for _, node := range nodePool.Status.Nodes {
			if node.TopologyMismatch {
				g.Expect(testCase.ExpectedMismatchNodesNames).To(ContainElement(node.Name))
			}
		}

		g.Expect(nodePool).NotTo(BeNil(),
			"Test name: <%v>, expected nodepool <%v> not to be nil",
			testCase.Name, nodePoolName)

		nodes := getNodePoolNodes(nodePoolName, k8sClient)
		for _, node := range nodes.Items {
			validateNodesNVLinkLabel(node, nodePool, g)
		}
		nodeNames := getNodePoolNodeNames(nodePool.Name, testCase.Nodes, k8sClient)
		if expectedNodePoolState.NodeNames == nil {
			expectedNodePoolState.NodeNames = []string{}
		}

		validateStatus(&nodePool.Status, &expectedNodePoolState.Status, nodePoolName, testCase.Name, g)

		validateGPUNetworkAccelerationDetected(nodePool, expectedNodePoolState.ExpectedGPUNetworkAccelerationDetected,
			testCase.Name, g)

		// Validate NodePool conditions
		if expectedNodePoolState.ExpectedConditions != nil {
			validateNodePoolConditions(nodePool, expectedNodePoolState.ExpectedConditions, nodePoolName, testCase.Name, g)
		}

		if nodePoolName != config.Get().DefaultNodepoolName {
			g.Expect(nodeNames).To(ConsistOf(expectedNodePoolState.NodeNames),
				"Test name: <%v>, expected nodes assignment of nodepool <%v>",
				testCase.Name, nodePoolName)
		} else {
			g.Expect(nodeNames).To(ContainElements(expectedNodePoolState.NodeNames),
				"Test name: <%v>, expected nodes assignment of nodepool <%v> to contain nodes",
				testCase.Name, nodePoolName)
		}
	}

	if testCase.ExpectedMetricsStateList != nil {
		for _, expectedMetric := range testCase.ExpectedMetricsStateList {
			Eventually(func() error {
				metric := metrics.GetNodePoolMetric().WithLabelValues(expectedMetric.NodeName, expectedMetric.NodePoolsName)
				expectedValue := getMetricRow(expectedMetric.NodeName, expectedMetric.NodePoolsName, expectedMetric.Count)
				err := testutil.CollectAndCompare(metric, strings.NewReader(expectedValue), "runai_node_nodepool")
				return err
			}, 30*time.Second, 2*time.Second).Should(BeNil())
		}
	}

	for nodeName, expectedNodeLabels := range testCase.ExpectedNodeLabels {
		node := getNodeFromClient(nodeName, k8sClient)
		g.Expect(node).NotTo(BeNil(), "Test name: <%v>, expected node <%s> not to be nil", testCase.Name, nodeName)

		for key, value := range expectedNodeLabels.expectedExistingLabels {
			g.Expect(node.Labels[key]).To(Equal(value),
				"Test name: <%v>, expected node <%s> to have label <%s> with value <%s>",
				testCase.Name, nodeName, key, value)
		}

		for _, label := range expectedNodeLabels.expectedNonExistingLabels {
			val, found := node.Labels[label]
			g.Expect(found).To(BeFalse(),
				"Test name: <%v>, expected node <%s> not to have label <%s> (value was: <%s>)",
				testCase.Name, nodeName, label, val)
		}
	}
}

func getMetricRow(nodeName string, nodePoolName string, count int) string {
	return fmt.Sprintf("\n\n# HELP runai_node_nodepool the node's nodepool\n# TYPE runai_node_nodepool gauge\nrunai_node_nodepool{node=\"%v\",nodepool=\"%v\"} %v\n",
		nodeName, nodePoolName, count)
}

func reconcileAllNodePools(ctx context.Context, nodePools []TestNodePool, npc *nodepool_controller.NodePoolController, mncc *managed_nodes_config.ManagedNodesConfigController) {
	// reconcile twice, since a "Reconcile" of 1 nodepool can trigger a reconcile of another..

	reconcileAllNodePoolsOnce(ctx, nodePools, npc, mncc)
	reconcileAllNodePoolsOnce(ctx, nodePools, npc, mncc)
}

func reconcileAllNodePoolsOnce(ctx context.Context, nodePools []TestNodePool, npc *nodepool_controller.NodePoolController, mncc *managed_nodes_config.ManagedNodesConfigController) {
	if mncc != nil {
		mncc.Reconcile(ctx, managed_nodes_config.MNCReconcileRequest{
			NamespacedName: types.NamespacedName{
				Name: testManagedNodesConfigName,
			},
		})
	}

	nodePools = append(nodePools, TestNodePool{Name: config.Get().DefaultNodepoolName})
	for _, nodePool := range nodePools {
		req := ctrl.Request{
			NamespacedName: types.NamespacedName{
				Name: nodePool.Name,
			},
		}
		_, _ = npc.Reconcile(ctx, req)
	}
}

func validateStatus(npStatus, expectedNpStatus *v1alpha1.NodePoolStatus,
	npName, testName string, g Gomega) {
	if (npName != config.Get().DefaultNodepoolName) ||
		(expectedNpStatus.Phase != "" || expectedNpStatus.Message != "") {
		g.Expect(npStatus.Phase).To(Equal(expectedNpStatus.Phase),
			"Test name: <%v>, expected status phase of nodepool <%v>", testName, npName)
		g.Expect(npStatus.Message).To(Equal(expectedNpStatus.Message),
			"Test name: <%v>, expected status message of nodepool <%v>", testName, npName)
	}

	if npName != config.Get().DefaultNodepoolName {
		g.Expect(npStatus.Nodes).To(ConsistOf(expectedNpStatus.Nodes),
			"Test name: <%v>, expected status nodes of nodepool <%v>", testName, npName)
	}
}

// The -detected annotation holds the expected bool when detection is configured, else is absent.
func validateGPUNetworkAccelerationDetected(nodePool *v1alpha1.NodePool, expected bool, testName string, g Gomega) {
	detectedValue, detectedFound := nodePool.Annotations[v1alpha1.AnnotationGPUNetworkAccelerationDetected]
	if nodePool.Annotations[v1alpha1.AnnotationGPUNetworkAccelerationDetection] == "" {
		g.Expect(detectedFound).To(BeFalse(),
			"Test name: <%v>, expected no detected annotation on nodepool <%v> when detection is unconfigured",
			testName, nodePool.Name)
		return
	}
	g.Expect(detectedFound).To(BeTrue(),
		"Test name: <%v>, expected detected annotation to be set on nodepool <%v>", testName, nodePool.Name)
	g.Expect(detectedValue).To(Equal(strconv.FormatBool(expected)),
		"Test name: <%v>, expected detected annotation value <%v> on nodepool <%v>", testName, expected, nodePool.Name)
}

func validateNodePoolConditions(nodePool *v1alpha1.NodePool, expectedConditions []ExpectedCondition, nodePoolName, testName string, g Gomega) {
	for _, expectedCondition := range expectedConditions {
		found := false
		for _, condition := range nodePool.Status.Conditions {
			if condition.Type == v1alpha1.NodePoolConditionType(expectedCondition.Type) {
				found = true
				g.Expect(string(condition.Status)).To(Equal(expectedCondition.Status),
					"Test name: <%v>, expected nodepool <%v> condition <%v> status to be <%v>, got <%v>",
					testName, nodePoolName, expectedCondition.Type, expectedCondition.Status, string(condition.Status))
				if expectedCondition.Reason != "" {
					g.Expect(condition.Reason).To(Equal(expectedCondition.Reason),
						"Test name: <%v>, expected nodepool <%v> condition <%v> reason to be <%v>, got <%v>",
						testName, nodePoolName, expectedCondition.Type, expectedCondition.Reason, condition.Reason)
				}
				break
			}
		}
		hasExpectConditions := expectedConditions != nil && len(expectedConditions) > 0
		g.Expect(found).To(Equal(hasExpectConditions),
			"Test name: <%v>, expected nodepool <%v> to have condition <%v>",
			testName, nodePoolName, expectedCondition.Type)
	}
}

func createDynamicTopologyTestCase() *TestCase {
	return &TestCase{
		Name: "Topology validation - Node label restored - Conditions cleared",
		NodePools: []TestNodePool{
			{
				Name:                "node-pool-topology",
				LabelKey:            "nodePoolTopologyKey",
				LabelValue:          "nodePoolTopologyValue",
				NetworkTopologyName: "test-topology",
				PlacementStrategy: PlacementStrategy{
					Gpu:       "binpack",
					GpuDevice: "binpack",
					Cpu:       "binpack",
				},
			},
		},
		Topologies: []TestTopology{
			{
				Name: "test-topology",
				Levels: []TestTopologyLevel{
					{NodeLabel: "topology-level-1"},
					{NodeLabel: "topology-level-2"},
					{NodeLabel: "topology-level-3"},
				},
			},
		},
		Nodes: map[string]TestNode{
			"kind-worker": {
				Name: "kind-worker",
				Labels: map[string]string{
					"nodePoolTopologyKey": "nodePoolTopologyValue",
					"topology-level-1":    "value1",
					"topology-level-2":    "value2",
					"topology-level-3":    "value3",
				},
			},
			"kind-worker2": {
				Name: "kind-worker2",
				Labels: map[string]string{
					"nodePoolTopologyKey": "nodePoolTopologyValue",
					"topology-level-1":    "value1",
					"topology-level-2":    "value2",
					"topology-level-3":    "value3",
				},
			},
			"kind-worker3": {
				Name: "kind-worker3",
				Labels: map[string]string{
					"nodePoolTopologyKey": "nodePoolTopologyValue",
					"topology-level-1":    "value1",
					"topology-level-2":    "value2",
					"topology-level-3":    "value3",
				},
			},
		},
		TestCallback: func(k8sClient client.Client, testCase *TestCase) {
			// First, remove a label to create mismatch
			updateNodeLabels("kind-worker", map[string]string{}, "topology-level-3", k8sClient)
			reconcileAllNodePools(context.Background(), testCase.NodePools, testCase.Npc, nil)

			// Wait for conditions to be set
			waitForNodePoolCondition("node-pool-topology", "NodeTopologyMismatch", "True", v1alpha1.NodeTopologyMismatchReason, k8sClient)

			// Then restore the label to clear the condition
			updateNodeLabels("kind-worker", map[string]string{"topology-level-3": "value3"}, "", k8sClient)
			reconcileAllNodePools(context.Background(), testCase.NodePools, testCase.Npc, nil)
		},
		ExpectedNodePoolsState: map[string]ExpectedNodePoolState{
			"node-pool-topology": {
				NodeNames: []string{"kind-worker", "kind-worker2", "kind-worker3"},
				Status: v1alpha1.NodePoolStatus{
					Phase: v1alpha1.NodePoolReady,
					Nodes: []v1alpha1.NodeInNodePool{
						{
							Name:   "kind-worker",
							Status: v1alpha1.NodeReady,
						},
						{
							Name:   "kind-worker2",
							Status: v1alpha1.NodeReady,
						},
						{
							Name:   "kind-worker3",
							Status: v1alpha1.NodeReady,
						},
					},
				},
				ExpectedConditions: []ExpectedCondition{
					{
						Type:   "NodeTopologyMismatch",
						Status: "False",
						Reason: v1alpha1.NodeTopologyMismatchReason,
					},
				},
			},
			kaiconstants.DefaultNodePoolName: {},
		},
		ExpectedMismatchNodesNames: []string{"kind-worker"},
		ExpectedNodeLabels: map[string]ExpectedNodeLabels{
			"kind-worker": {
				expectedExistingLabels: map[string]string{
					"nodePoolTopologyKey": "nodePoolTopologyValue",
					"topology-level-1":    "value1",
					"topology-level-2":    "value2",
					"topology-level-3":    "value3",
				},
			},
			"kind-worker2": {
				expectedExistingLabels: map[string]string{
					"nodePoolTopologyKey": "nodePoolTopologyValue",
					"topology-level-1":    "value1",
					"topology-level-2":    "value2",
					"topology-level-3":    "value3",
				},
			},
			"kind-worker3": {
				expectedExistingLabels: map[string]string{
					"nodePoolTopologyKey": "nodePoolTopologyValue",
					"topology-level-1":    "value1",
					"topology-level-2":    "value2",
					"topology-level-3":    "value3",
				},
			},
		},
	}
}

var _ = Describe("Updating a nodepool's SchedulingShardConfig while a pod is running", func() {
	var (
		ctx     context.Context
		cancel  context.CancelFunc
		stopper chan struct{}
	)

	AfterEach(func() {
		close(stopper)
		cancel()
	})

	It("keeps the nodepool Ready and propagates the change to its SchedulingShard", func() {
		stopper = make(chan struct{}, 1)
		ctx, cancel = context.WithCancel(context.Background())

		const nodePoolName = "node-pool-a"
		testCase := &TestCase{
			Name: "SchedulingShardConfig update - pod running - stays Ready",
			NodePools: []TestNodePool{{
				Name:              nodePoolName,
				LabelKey:          "nodePoolAKey",
				LabelValue:        "nodePoolAValue",
				PlacementStrategy: PlacementStrategy{Gpu: "spread", Cpu: "binpack"},
			}},
			Nodes: map[string]TestNode{
				"kind-worker": {
					Name:   "kind-worker",
					Labels: map[string]string{"nodePoolAKey": "nodePoolAValue", config.Get().NodePoolNameLabel: nodePoolName},
				},
			},
			Pods: []TestPod{{
				Name:        "pod0",
				Labels:      map[string]string{config.Get().NodePoolNameLabel: nodePoolName},
				NodeName:    "kind-worker",
				Annotations: map[string]string{config.PodGroupAnnotationForPod: "pod-group-name-1"},
			}},
			PodGroups: []TestPodGroup{{
				Name:   "pod-group-name-1",
				Labels: map[string]string{config.Get().NodePoolNameLabel: nodePoolName},
			}},
			// Required by preTestSetup (every nodepool must have an entry). This spec drives its
			// own assertions below rather than the table validator, so this states the baseline.
			ExpectedNodePoolsState: map[string]ExpectedNodePoolState{
				nodePoolName: {
					NodeNames: []string{"kind-worker"},
					Status:    v1alpha1.NodePoolStatus{Phase: v1alpha1.NodePoolReady},
				},
			},
		}

		k8sClient := preTestSetup(ctx, stopper, testCase, scheme)
		npc := nodepool_controller.NewNodePoolController(k8sClient, scheme, &common.NodePoolControllerParams{})
		npc.SetServiceMonitorEnabled(true)
		mncc := managed_nodes_config.NewManagedNodesConfigController(k8sClient, scheme, npc)

		reconcileAllNodePools(ctx, testCase.NodePools, npc, mncc)

		// Baseline: the nodepool is Ready and its shard carries the seeded placement strategy.
		Expect(eventuallyGetNodePoolFromClient(nodePoolName, k8sClient).Status.Phase).To(Equal(v1alpha1.NodePoolReady))
		baselineShard := getSchedulingShardFromClient(nodePoolName, k8sClient)
		Expect(baselineShard).NotTo(BeNil())
		Expect(baselineShard.Spec.PlacementStrategy).NotTo(BeNil())
		Expect(baselineShard.Spec.PlacementStrategy.GPU).To(Equal(ptr.To("spread")))
		Expect(baselineShard.Spec.PlacementStrategy.CPU).To(Equal(ptr.To("binpack")))

		// Change the nodepool's scheduler config while the pod keeps running.
		updateNodePoolSchedulingShardConfig(nodePoolName, &v1alpha1.SchedulingShardConfig{
			PlacementStrategy: &kaiv1.PlacementStrategy{GPU: ptr.To("binpack"), CPU: ptr.To("spread")},
		}, k8sClient)

		reconcileAllNodePools(ctx, testCase.NodePools, npc, mncc)

		Eventually(func(g Gomega) {
			np := getNodePoolFromClient(nodePoolName, k8sClient)
			g.Expect(np).NotTo(BeNil())
			// The running workload is undisturbed: the nodepool stays Ready with its node.
			g.Expect(np.Status.Phase).To(Equal(v1alpha1.NodePoolReady))
			g.Expect(np.Status.Nodes).To(HaveLen(1))
			g.Expect(np.Status.Nodes[0].Name).To(Equal("kind-worker"))

			// ...and the change propagated to the shard: the strategy flipped from the seeded values.
			shard := getSchedulingShardFromClient(nodePoolName, k8sClient)
			g.Expect(shard).NotTo(BeNil())
			g.Expect(shard.Spec.PlacementStrategy).NotTo(BeNil())
			g.Expect(shard.Spec.PlacementStrategy.GPU).To(Equal(ptr.To("binpack")))
			g.Expect(shard.Spec.PlacementStrategy.CPU).To(Equal(ptr.To("spread")))
		}, validateTestTimeout, validateTestInterval).Should(Succeed())
	})
})
