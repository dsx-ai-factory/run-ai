package tests

import (
	"context"
	"fmt"
	"strings"
	"time"

	kaiv1 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/kai/v1"
	kaiconstants "github.com/kai-scheduler/KAI-scheduler/pkg/common/constants"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	monitorv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/run-ai/runai/runai-cluster/cluster/nodepool-controller/pkg/common"
	"github.com/run-ai/runai/runai-cluster/cluster/nodepool-controller/pkg/nodepool_controller"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type TestCase struct {
	Name               string
	NodePools          []TestNodePool
	Nodes              map[string]TestNode
	Pods               []TestPod
	PodGroups          []TestPodGroup
	Topologies         []TestTopology
	ManagedNodesConfig []TestManagedNodesConfig
	Projects           []TestProject
	TestCallback       TestCallbackFn
	AfterTestCallback  TestCallbackFn

	// NodePoolName to ExpectedNodePoolState
	ExpectedNodePoolsState   map[string]ExpectedNodePoolState
	ExpectedMetricsStateList []ExpectedMetricsState
	// NodeName to ExpectedNodeLabels
	ExpectedNodeLabels         map[string]ExpectedNodeLabels
	ExpectedMismatchNodesNames []string
	Npc                        *nodepool_controller.NodePoolController
}

func (tc *TestCase) SetNodepoolController(npc *nodepool_controller.NodePoolController) {
	tc.Npc = npc
}

type TestCallbackFn func(k8sClient client.Client, testCase *TestCase)

type TestNodePool struct {
	Name                            string                                    `json:"name,omitempty"`
	LabelKey                        string                                    `json:"labelKey,omitempty"`
	LabelValue                      string                                    `json:"labelValue,omitempty"`
	PlacementStrategy               PlacementStrategy                         `json:"placementStrategy,omitempty"`
	CreateDuringTestCallback        bool                                      `json:"notCreateBefore,omitempty"`
	GPUNetworkAccelerationDetection *v1alpha1.GPUNetworkAccelerationDetection `json:"gpuNetworkAccelerationDetection,omitempty"`
	GPUNetworkAccelerationLabelKey  string                                    `json:"gpuNetworkAccelerationLabelKey,omitempty"`
	NetworkTopologyName             string                                    `json:"networkTopologyName,omitempty"`
}

type PlacementStrategy struct {
	Gpu       string `json:"gpu,omitempty"`
	GpuDevice string `json:"gpuDevice,omitempty"`
	Cpu       string `json:"cpu,omitempty"`
}

type TestNode struct {
	Name          string            `json:"name,omitempty"`
	Labels        map[string]string `json:"labels,omitempty"`
	Unschedulable bool              `json:"unschedulable,omitempty"`
}

type TestManagedNodesConfig struct {
	Name         string              `json:"name,omitempty"`
	NodeSelector corev1.NodeSelector `json:"nodeSelector,omitempty"`
}

type TestPod struct {
	Name          string            `json:"name,omitempty"`
	Namespace     string            `json:"namespace,omitempty"`
	Labels        map[string]string `json:"labels,omitempty"`
	NodeName      string            `json:"nodeName,omitempty"`
	SchedulerName string            `json:"schedulerName,omitempty"`
	Annotations   map[string]string `json:"annotations,omitempty"`
}

type TestPodGroup struct {
	Name      string            `json:"name,omitempty"`
	Namespace string            `json:"namespace,omitempty"`
	Labels    map[string]string `json:"labels,omitempty"`
}

type TestTopology struct {
	Name   string              `json:"name,omitempty"`
	Levels []TestTopologyLevel `json:"levels,omitempty"`
}

type TestTopologyLevel struct {
	NodeLabel string `json:"nodeLabel,omitempty"`
}

type ExpectedNodePoolState struct {
	NodeNames                              []string                `json:"nodeNames,omitempty"`
	Status                                 v1alpha1.NodePoolStatus `json:"status,omitempty"`
	ExpectedConditions                     []ExpectedCondition     `json:"expectedConditions,omitempty"`
	ExpectedGPUNetworkAccelerationDetected bool                    `json:"expectedGpuNetworkAccelerationDetected,omitempty"`
}

type ExpectedCondition struct {
	Type   string `json:"type,omitempty"`
	Status string `json:"status,omitempty"`
	Reason string `json:"reason,omitempty"`
}

type ExpectedMetricsState struct {
	NodeName      string
	NodePoolsName string
	Count         int
}

type TestProject struct {
	Name string
	// QueueNodePools are the nodepool names referenced via the project's queues (per-nodepool resources).
	QueueNodePools []string
	// DefaultNodePools are the names referenced only via the project's default nodepool list.
	DefaultNodePools []string
}

type ExpectedNodeLabels struct {
	expectedExistingLabels    map[string]string
	expectedNonExistingLabels []string
}

var defaultTestNodePool = &TestNodePool{
	Name: kaiconstants.DefaultNodePoolName,
}

func preTestSetup(ctx context.Context, stopper <-chan struct{},
	testCase *TestCase, scheme *runtime.Scheme) client.WithWatch {
	objs := []client.Object{}

	for _, testManagedNodeConfig := range testCase.ManagedNodesConfig {
		node := getManagedNodeConfigObj(testManagedNodeConfig.Name, testManagedNodeConfig.NodeSelector)
		objs = append(objs, node)
	}

	for _, testTopology := range testCase.Topologies {
		topology := getTestTopologyObj(&testTopology)
		objs = append(objs, topology)
	}

	for _, testProject := range testCase.Projects {
		// kai.resources twin used by the suite.
		objs = append(objs, getTestProjectObj(testProject))
	}

	for testNodeName, testNode := range testCase.Nodes {
		Expect(testNodeName).To(Equal(testNode.Name))
		node := getNodeObjWithLabels(testNode.Name, testNode.Labels, &testNode.Unschedulable)
		objs = append(objs, node)
	}

	namespacesForTest := map[string]bool{}
	for _, testPod := range testCase.Pods {
		namespacesForTest[getDefaultNamespace(testPod.Namespace)] = true

		pod := getPodObj(testPod.Name, testPod.Namespace, testPod.NodeName, testPod.SchedulerName, testPod.Labels, testPod.Annotations)
		objs = append(objs, pod)
	}

	for _, testPodGroup := range testCase.PodGroups {
		namespacesForTest[getDefaultNamespace(testPodGroup.Namespace)] = true

		pod := getPodGroupObj(testPodGroup.Name, testPodGroup.Namespace, testPodGroup.Labels)
		objs = append(objs, pod)
	}

	namespacesForTest[runaiNamespace] = true
	delete(namespacesForTest, "")
	for ns := range namespacesForTest {
		nsObj := getNamespaceObj(ns)
		objs = append(objs, nsObj)
	}

	By(fmt.Sprintf("setting up test case <%v>", testCase.Name))

	nodePool := getTestNodePoolObj(defaultTestNodePool)
	objs = append(objs, nodePool)

	for _, testNodePool := range testCase.NodePools {
		Expect(testCase.ExpectedNodePoolsState).To(HaveKey(testNodePool.Name),
			"Test: <%v>, expected to have all nodepools in ExpectedNodePoolsState", testCase.Name)
		if testNodePool.CreateDuringTestCallback {
			continue
		}

		nodePool = getTestNodePoolObj(&testNodePool)
		objs = append(objs, nodePool)
	}

	if testCase.ExpectedNodePoolsState == nil {
		testCase.ExpectedNodePoolsState = make(map[string]ExpectedNodePoolState)
	}

	objs = append(objs, getSchedulerOperands(testCase.NodePools)...)
	fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithStatusSubresource(
		&v1alpha1.NodePool{}, &corev1.Node{}, &kaiv1.SchedulingShard{}).
		WithIndex(&v1alpha1.NodePool{}, common.IsDeletingPhaseField, nodepool_controller.NodePoolIsDeletingPhaseIndexer).
		WithIndex(&corev1.Pod{}, common.PodRunningWithRunaiSchedulerNodeNameField, nodepool_controller.PodRunningWithRunaiSchedulerNodeNameIndexer).
		Build()

	fakeCachedClient := NewFakeCachedClient(fakeClient)
	metricsWatch, err := nodepool_controller.InitMetricsWatch(fakeCachedClient)
	Expect(err).ToNot(HaveOccurred())
	metricsWatch.Start(ctx, stopper)

	// the metrics watch needs to get a "Added" event for each node,
	// that's why we add it after the client is created
	for _, obj := range objs {
		err = fakeCachedClient.Create(ctx, obj)
		Expect(err).ToNot(HaveOccurred())
	}

	return fakeClient
}

func getSchedulerOperands(nodePools []TestNodePool) []client.Object {
	objs := []client.Object{}

	nodePools = append(nodePools, *defaultTestNodePool)
	for _, np := range nodePools {
		status := metav1.ConditionTrue
		if strings.HasPrefix(np.Name, "notready-np") {
			status = metav1.ConditionFalse
		}

		schedulingShard := &kaiv1.SchedulingShard{
			ObjectMeta: metav1.ObjectMeta{
				Name: np.Name,
			},
			Status: kaiv1.SchedulingShardStatus{
				Conditions: []metav1.Condition{
					{
						Type:               string(kaiv1.ConditionTypeDeployed),
						Status:             status,
						Reason:             "SomeReason",
						LastTransitionTime: metav1.Date(2023, 1, 1, 1, 1, 1, 1, time.UTC),
					},
					{
						Type:               string(kaiv1.ConditionTypeAvailable),
						Status:             status,
						Reason:             "SomeReason",
						LastTransitionTime: metav1.Date(2023, 1, 1, 1, 1, 1, 1, time.UTC),
					},
				},
			},
		}

		objs = append(objs, schedulingShard)

		operandName := getOperandName(np.Name)
		schedulerMonitor := &monitorv1.ServiceMonitor{
			ObjectMeta: metav1.ObjectMeta{
				Name:      operandName,
				Namespace: runaiNamespace,
				UID:       types.UID(operandName),
			},
		}
		objs = append(objs, schedulerMonitor)
	}

	clusterCR := newClusterCR()
	objs = append(objs, clusterCR)

	return objs
}
