/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tests

import (
	"fmt"
	"strings"
	"time"

	"github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/rand"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	timeout  = time.Second * 6
	interval = time.Millisecond * 250
)

func createTestNodePool(testName string, testNodePool *TestNodePool, k8sClient client.Client) {
	createNodePool(
		testName,
		testNodePool.Name,
		testNodePool.LabelKey,
		testNodePool.LabelValue,
		testNodePool.Phase,
		k8sClient)
}

func createPodGroupAndPodsFromSpec(testName string, podGroupNamespacedName types.NamespacedName, nodePoolOptions []string,
	numOfPods int, k8sClient client.Client) {
	createPodGroupAndPodsFromSpecWithProjName(testName, podGroupNamespacedName, nodePoolOptions, numOfPods, testProject, k8sClient)
}

func createPodGroupAndPodsFromSpecWithProjName(testName string, podGroupNamespacedName types.NamespacedName, nodePoolOptions []string,
	numOfPods int, projectName string, k8sClient client.Client) {
	By(fmt.Sprintf("Test: %s, creating podgroup: %v, nodePool options: %v",
		testName, podGroupNamespacedName, nodePoolOptions))

	createPodsForPodGroupForNodePoolOptions(podGroupNamespacedName.Name, podGroupNamespacedName.Namespace, nodePoolOptions, numOfPods)

	podGroup := getPodGroupObj(podGroupNamespacedName,
		nodePoolOptions, "", projectName)
	createPodGroup(testName, podGroup, k8sClient)
}

func createPodGroupFromSpecWithSingleNPOldStyle(testName string, podGroupNamespacedName types.NamespacedName, singleNodePool, projectName string, k8sClient client.Client) {
	By(fmt.Sprintf("Test: %s, creating podgroup: %v, single nodePool (old style): %s",
		testName, podGroupNamespacedName, singleNodePool))

	createPodsForPodGroupForNodePoolOptions(podGroupNamespacedName.Name, podGroupNamespacedName.Namespace, []string{}, 1)

	podGroup := getPodGroupObj(podGroupNamespacedName, []string{}, singleNodePool, projectName)
	createPodGroup(testName, podGroup, k8sClient)
}

func createPodsForPodGroupForNodePoolOptions(podGroupName, namespace string,
	nodePoolOptions []string, numberOfPods int) []*corev1.Pod {
	matchExpressionss := getNodeSelectorTermsForNodePoolOptions(nodePoolOptions)
	return createPodsForPodGroup(podGroupName, namespace, numberOfPods, matchExpressionss)
}

func createPodsForPodGroup(podGroupName, namespace string, numberOfPods int,
	nodeMatchExpressionssForPod [][]corev1.NodeSelectorRequirement) []*corev1.Pod {
	podsCreated := []*corev1.Pod{}

	for i := 0; i < numberOfPods; i++ {
		pod := getPodObjWithNodeAffinity(
			getPodNameForPodGroupIndex(podGroupName, i),
			podGroupName,
			namespace,
			nodeMatchExpressionssForPod)
		podsCreated = append(podsCreated, pod)
		createPod(pod, k8sClient)
	}

	Eventually(func() bool {
		pods := getPodGroupPods(types.NamespacedName{Name: podGroupName, Namespace: namespace})
		if pods == nil {
			return false
		}

		return len(pods.Items) >= numberOfPods
	}, timeout, interval).Should(BeTrue())

	return podsCreated
}

func getNodeSelectorTermsForNodePoolOptions(nodePoolOptions []string) [][]corev1.NodeSelectorRequirement {
	matchExpressionss := [][]corev1.NodeSelectorRequirement{}

	allTestNodePools := append(testNodePools, extraTestNodePools...)

	for _, nodePool := range nodePoolOptions {
		if nodePool == testDefaultNodePoolName {
			matchExpressionss = append(matchExpressionss, []corev1.NodeSelectorRequirement{
				{
					Key:      testNodePoolLabelKey,
					Operator: corev1.NodeSelectorOpDoesNotExist,
				},
			})

			continue
		}

		for _, testNodePool := range allTestNodePools {
			if nodePool == testNodePool.Name {
				matchExpressionss = append(matchExpressionss, []corev1.NodeSelectorRequirement{
					{
						Key:      testNodePool.LabelKey,
						Operator: corev1.NodeSelectorOpIn,
						Values:   []string{testNodePool.LabelValue},
					},
				})
			}
		}
	}

	return matchExpressionss
}

func generatePodGroupName() string {
	return fmt.Sprintf("pg-%s", rand.String(5))
}

func generateNodePoolName() string {
	return fmt.Sprintf("np-%s", rand.String(5))
}

func generateNodePoolKey() string {
	return fmt.Sprintf("np-key-%s", rand.String(5))
}

func generateNodePoolVal() string {
	return fmt.Sprintf("np-val-%s", rand.String(5))
}

func getPodGroupQueue(projectName string, nodePoolName string) string {
	queueName := projectName
	if nodePoolName != testDefaultNodePoolName && nodePoolName != "" && nodePoolName != testUnexistingNodePool {
		queueName = queueName + "-" + nodePoolName
	}

	return queueName
}

func hackGetProjNameFromNamespace(namespace string) string {
	if strings.HasPrefix(namespace, "runai-") {
		return namespace[6:]
	}

	return namespace
}

func getPodNameForPodGroupIndex(podGroupName string, index int) string {
	return fmt.Sprintf("%s-%d", podGroupName, index)
}

// updateRandomLabelToTriggerController mutates a label and then reconciles.
// In the envtest-based suite the label change was needed to make the manager's
// watch enqueue a reconcile; here the watch is gone, so the label write is kept
// only to model an external mutation and reconcile() does the actual work.
func updateRandomLabelToTriggerController(pg types.NamespacedName, k8sClient client.Client) {
	key := generateNodePoolKey()
	val := generateNodePoolVal()

	pgFromClient := getPodGroupFromClient(pg, k8sClient)
	Expect(pgFromClient).NotTo(BeNil())
	labels := pgFromClient.GetLabels()
	labels[key] = val
	pgFromClient.SetLabels(labels)
	Expect(k8sClient.Update(apiCtx, pgFromClient)).To(Succeed())

	reconcile(pg)
}

func updateNodePoolStatus(nodePoolName string, newPhase v1alpha1.NodePoolPhase) {
	np := &v1alpha1.NodePool{}
	ExpectWithOffset(1, k8sClient.Get(apiCtx, types.NamespacedName{Name: nodePoolName}, np)).To(Succeed())
	np.Status.Phase = newPhase
	expectUpdateResourceStatus(k8sClient, np)
}
