/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tests

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2alpha2"

	kaiv1 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/kai/v1"
	kaiv1alpha1 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/kai/v1alpha1"
	"github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	. "github.com/onsi/gomega"
	monitorv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/nodepool-controller/pkg/config"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	timeout              = time.Second * 4
	interval             = time.Millisecond * 250
	validateTestTimeout  = time.Second * 10
	validateTestInterval = time.Second * 5
)

func getNodePoolNodeNames(nodePoolName string, testNodes map[string]TestNode, k8sClient client.Client) (nodeNames []string) {
	nodeNames = []string{}
	nodes := getNodePoolNodes(nodePoolName, k8sClient)
	for i := range nodes.Items {
		node := &nodes.Items[i]

		// filter out the nodes that are not in testCase.Nodes -
		// they are out of the scope for the test expected values
		if _, found := testNodes[node.Name]; found {
			nodeNames = append(nodeNames, node.Name)
		}
	}
	return nodeNames
}

func getNodePoolNodes(nodePoolName string, k8sClient client.Client) (nodes *corev1.NodeList) {
	selectorStr := fmt.Sprintf("%s==%s", config.Get().NodePoolNameLabel, nodePoolName)
	if nodePoolName == config.Get().DefaultNodepoolName {
		selectorStr = fmt.Sprintf("!%v", config.Get().NodePoolNameLabel)
	}

	labelSelector, err := labels.Parse(selectorStr)
	Expect(err).To(BeNil())

	nodes = &corev1.NodeList{}
	Eventually(func() error {
		return k8sClient.List(context.Background(), nodes, &client.ListOptions{LabelSelector: labelSelector})
	}, timeout, interval).Should(BeNil())

	return nodes
}

func getNodePoolFromClient(nodePoolName string, k8sClient client.Client) (nodePool *v1alpha1.NodePool) {
	nodeKey := types.NamespacedName{Name: nodePoolName}
	nodePool = &v1alpha1.NodePool{}
	err := k8sClient.Get(context.Background(), nodeKey, nodePool)
	if err != nil {
		return nil
	}
	return nodePool
}

func eventuallyGetNodePoolFromClient(nodePoolName string, k8sClient client.Client) (nodePool *v1alpha1.NodePool) {
	Eventually(func() bool {
		nodePool = getNodePoolFromClient(nodePoolName, k8sClient)
		if nodePool == nil {
			return false
		}
		return true
	}, timeout*3, interval).Should(BeTrue())

	return nodePool
}

func getNodeFromClient(nodeName string, k8sClient client.Client) (node *corev1.Node) {
	nodeKey := types.NamespacedName{Name: nodeName}
	node = &corev1.Node{}
	err := k8sClient.Get(context.Background(), nodeKey, node)
	if err != nil {
		return nil
	}
	return node
}

func getPodFromClient(podName, namespace string, k8sClient client.Client) (pod *corev1.Pod) {
	podKey := types.NamespacedName{Name: podName, Namespace: namespace}
	pod = &corev1.Pod{}
	err := k8sClient.Get(context.Background(), podKey, pod)
	if err != nil {
		return nil
	}
	return pod
}

func updateNodeLabels(nodeName string, labelsToAdd map[string]string, labelToRemove string, k8sClient client.Client) {
	Eventually(func() error {
		nodeKey := types.NamespacedName{Name: nodeName}
		node := &corev1.Node{}
		err := k8sClient.Get(context.Background(), nodeKey, node)
		if err != nil {
			return err
		}

		if node.Labels == nil {
			node.Labels = map[string]string{}
		}
		for k, v := range labelsToAdd {
			node.Labels[k] = v
		}
		if labelToRemove != "" {
			delete(node.Labels, labelToRemove)
		}

		return k8sClient.Update(context.Background(), node)
	}, timeout*10, interval).Should(BeNil())
}

// waitForNodePoolCondition waits for a specific condition to be set on a nodepool
func waitForNodePoolCondition(nodePoolName, conditionType, expectedStatus, expectedReason string, k8sClient client.Client) {
	Eventually(func(g Gomega) {
		nodePoolKey := types.NamespacedName{Name: nodePoolName}
		nodePool := &v1alpha1.NodePool{}
		err := k8sClient.Get(context.Background(), nodePoolKey, nodePool)
		g.Expect(err).NotTo(HaveOccurred())

		conditionFound := false
		for _, condition := range nodePool.Status.Conditions {
			if string(condition.Type) == conditionType {
				conditionFound = true
				g.Expect(string(condition.Status)).To(Equal(expectedStatus))
				if expectedReason != "" {
					g.Expect(condition.Reason).To(Equal(expectedReason))
				}
				break
			}
		}
		g.Expect(conditionFound).To(BeTrue(), "Condition %s not found on nodepool %s", conditionType, nodePoolName)
	}, timeout*10, interval).Should(Succeed())
}

func updateNodePoolLabels(nodePoolName, labelKey, labelValue string, k8sClient client.Client) {
	Eventually(func() error {
		nodePoolKey := types.NamespacedName{Name: nodePoolName}
		nodePool := &v1alpha1.NodePool{}
		err := k8sClient.Get(context.Background(), nodePoolKey, nodePool)
		if err != nil {
			return err
		}

		update := false
		if nodePool.Spec.LabelKey != labelKey {
			update = true
			nodePool.Spec.LabelKey = labelKey
		}
		if nodePool.Spec.LabelValue != labelValue {
			update = true
			nodePool.Spec.LabelValue = labelValue
		}
		if !update {
			return nil
		}
		return k8sClient.Update(context.Background(), nodePool)
	}, timeout*10, interval).Should(BeNil())
}

func updateNodePoolSchedulingShardConfig(nodePoolName string, cfg *v1alpha1.SchedulingShardConfig, k8sClient client.Client) {
	Eventually(func() error {
		nodePool := &v1alpha1.NodePool{}
		if err := k8sClient.Get(context.Background(), types.NamespacedName{Name: nodePoolName}, nodePool); err != nil {
			return err
		}
		nodePool.Spec.SchedulingShardConfig = cfg
		return k8sClient.Update(context.Background(), nodePool)
	}, timeout*10, interval).Should(BeNil())
}

func getNodeObjWithLabels(nodeName string, labels map[string]string, unschedulable *bool) *corev1.Node {
	u := false
	if unschedulable != nil {
		u = *unschedulable
	}
	return &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:   nodeName,
			Labels: labels,
		},
		Spec: corev1.NodeSpec{
			Unschedulable: u,
		},
		Status: corev1.NodeStatus{
			Conditions: []corev1.NodeCondition{
				{
					Type:   corev1.NodeReady,
					Status: corev1.ConditionTrue,
				},
			},
		},
	}
}
func getNamespaceObj(namespace string) *corev1.Namespace {
	return &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: getDefaultNamespace(namespace),
		},
	}
}

func getPodObj(podName, podNamespace, nodeName, schedulerName string, labels, annotations map[string]string) *corev1.Pod {
	if schedulerName == "" {
		schedulerName = config.Get().SchedulerName
	}

	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:        podName,
			Namespace:   getDefaultNamespace(podNamespace),
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: corev1.PodSpec{
			NodeName:      nodeName,
			SchedulerName: schedulerName,
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
		},
	}
}

func getPodGroupObj(podGroupName, podGroupNamespace string, labels map[string]string) *v2alpha2.PodGroup {

	return &v2alpha2.PodGroup{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podGroupName,
			Namespace: getDefaultNamespace(podGroupNamespace),
			Labels:    labels,
		},
		Spec:   v2alpha2.PodGroupSpec{},
		Status: v2alpha2.PodGroupStatus{},
	}
}

func getManagedNodeConfigObj(name string, selector corev1.NodeSelector) *v1alpha1.ManagedNodesConfig {
	return &v1alpha1.ManagedNodesConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: v1alpha1.ManagedNodesConfigSpec{
			InclusionCriteria: selector,
		},
	}
}

func getTestProjectObj(testProject TestProject) *v1alpha1.Project {
	queues := make([]v1alpha1.QueueConfig, 0, len(testProject.QueueNodePools))
	for _, nodePoolName := range testProject.QueueNodePools {
		queues = append(queues, v1alpha1.QueueConfig{
			Name:     fmt.Sprintf("%s-%s", testProject.Name, nodePoolName),
			Nodepool: nodePoolName,
		})
	}

	return &v1alpha1.Project{
		ObjectMeta: metav1.ObjectMeta{
			Name: testProject.Name,
		},
		Spec: v1alpha1.ProjectSpec{
			Queues:           queues,
			DefaultNodePools: testProject.DefaultNodePools,
		},
	}
}

func getTestNodePoolObj(testNodePool *TestNodePool) *v1alpha1.NodePool {
	phase := v1alpha1.NodePoolReady
	if !strings.HasPrefix(testNodePool.Name, "notready-np") {
		phase = v1alpha1.NodePoolUnschedulable
	}

	// Detection is driven by annotations; nil detection means no annotation (feature disabled).
	annotations := map[string]string{}
	if testNodePool.GPUNetworkAccelerationDetection != nil {
		annotations[v1alpha1.AnnotationGPUNetworkAccelerationDetection] =
			string(*testNodePool.GPUNetworkAccelerationDetection)
	}
	if testNodePool.GPUNetworkAccelerationLabelKey != "" {
		annotations[v1alpha1.AnnotationGPUNetworkAccelerationLabelKey] =
			testNodePool.GPUNetworkAccelerationLabelKey
	}

	var schedulingShardConfig *v1alpha1.SchedulingShardConfig
	if testNodePool.PlacementStrategy.Gpu != "" || testNodePool.PlacementStrategy.Cpu != "" {
		schedulingShardConfig = &v1alpha1.SchedulingShardConfig{
			PlacementStrategy: &kaiv1.PlacementStrategy{
				GPU: ptr.To(testNodePool.PlacementStrategy.Gpu),
				CPU: ptr.To(testNodePool.PlacementStrategy.Cpu),
			},
		}
	}

	nodePool := &v1alpha1.NodePool{
		ObjectMeta: metav1.ObjectMeta{
			Name:        testNodePool.Name,
			Labels:      map[string]string{},
			Annotations: annotations,
		},
		Spec: v1alpha1.NodePoolSpec{
			LabelKey:                     testNodePool.LabelKey,
			LabelValue:                   testNodePool.LabelValue,
			PreferredNetworkTopologyName: testNodePool.NetworkTopologyName,
			SchedulingShardConfig:        schedulingShardConfig,
		},
		Status: v1alpha1.NodePoolStatus{
			Phase: phase,
		},
	}
	return nodePool
}

func getDefaultNamespace(namespace string) string {
	if namespace == "" {
		return runaiNamespace
	}
	return namespace
}

func getOperandName(ownerNodePoolName string) string {
	return fmt.Sprintf("%v-%v", "runai-scheduler", ownerNodePoolName)
}

func getServiceMonitorFromClient(deploymentName string, k8sClient client.Client) (serviceMonitor *monitorv1.ServiceMonitor) {
	nodeKey := types.NamespacedName{Name: deploymentName, Namespace: runaiNamespace}
	serviceMonitor = &monitorv1.ServiceMonitor{}
	err := k8sClient.Get(context.Background(), nodeKey, serviceMonitor)
	if err != nil {
		return nil
	}
	return serviceMonitor
}

func getSchedulingShardFromClient(shardName string, k8sClient client.Client) *kaiv1.SchedulingShard {
	shard := &kaiv1.SchedulingShard{}
	err := k8sClient.Get(context.Background(), client.ObjectKey{Name: shardName}, shard)
	if err != nil {
		return nil
	}
	return shard
}

func getTestTopologyObj(testTopology *TestTopology) *kaiv1alpha1.Topology {
	levels := []kaiv1alpha1.TopologyLevel{}
	for _, level := range testTopology.Levels {
		levels = append(levels, kaiv1alpha1.TopologyLevel{
			NodeLabel: level.NodeLabel,
		})
	}

	return &kaiv1alpha1.Topology{
		ObjectMeta: metav1.ObjectMeta{
			Name: testTopology.Name,
		},
		Spec: kaiv1alpha1.TopologySpec{
			Levels: levels,
		},
	}
}
