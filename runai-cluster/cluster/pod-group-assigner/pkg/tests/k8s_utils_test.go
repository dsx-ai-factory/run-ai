/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tests

import (
	"context"
	"fmt"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/pod-group-assigner/pkg/controllers/common"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/pod-group-assigner/pkg/testbuilders"
	"github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	kaiv2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2"
	kaiv2alpha2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2alpha2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func getPodGroupObj(podGroupNamespacedName types.NamespacedName, nodePoolOptions []string, singleNodePool string, projectName string) *kaiv2alpha2.PodGroup {
	podGroup := &kaiv2alpha2.PodGroup{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podGroupNamespacedName.Name,
			Namespace: podGroupNamespacedName.Namespace,
			Labels:    map[string]string{},
			UID:       types.UID(podGroupNamespacedName.Name),
		},
		Spec: kaiv2alpha2.PodGroupSpec{
			Queue:             hackGetProjNameFromNamespace(podGroupNamespacedName.Namespace),
			MarkUnschedulable: ptr.To(false),
			SchedulingBackoff: ptr.To(int32(1)),
		},
	}

	if len(nodePoolOptions) == 0 && singleNodePool != "" && singleNodePool != testDefaultNodePoolName {
		podGroup.Labels[testNodePoolLabelKey] = singleNodePool
	} else {
		podGroup.Labels[testNodePoolLabelKey] = testUnexistingNodePool
	}

	if projectName != "" {
		podGroup.Labels["project"] = projectName
	}

	return podGroup
}

func getPodObjWithNodeAffinity(podName, podGroupName, namespace string, matchExpressionss [][]corev1.NodeSelectorRequirement) *corev1.Pod {
	return testbuilders.BuildPodWithNodeAffinity(podName, podGroupName, namespace, matchExpressionss)
}

func getNodePoolObj(nodePoolName string, labelKey string, labelVal string) *v1alpha1.NodePool {
	return testbuilders.BuildNodePool(nodePoolName, labelKey, labelVal, "")
}

func getQueueObj(queueName, projectName, nodepoolName string) *kaiv2.Queue {
	queue := &kaiv2.Queue{
		ObjectMeta: metav1.ObjectMeta{
			Name: queueName,
			Labels: map[string]string{
				RunaiProjectLabel: projectName,
			},
		},
	}

	if nodepoolName != testDefaultNodePoolName {
		queue.Labels[testNodePoolLabelKey] = nodepoolName
	}

	return queue
}

func getPodGroupFromClient(podGroupNamespacedName types.NamespacedName, k8sClient client.Client) (podGroup *kaiv2alpha2.PodGroup) {
	podGroup = &kaiv2alpha2.PodGroup{}

	err := k8sClient.Get(apiCtx, podGroupNamespacedName, podGroup)
	if err != nil {
		return nil
	}

	return podGroup
}

func getPodFromClient(podNamespacedName types.NamespacedName, k8sClient client.Client) (pod *corev1.Pod) {
	pod = &corev1.Pod{}

	err := k8sClient.Get(apiCtx, podNamespacedName, pod)
	if err != nil {
		return nil
	}

	return pod
}

func getPodGroupPods(podGroupNamespacedName types.NamespacedName) (pods *corev1.PodList) {
	podsByPodGroupFieldSelector :=
		fields.OneTermEqualSelector(common.PodByPodGroupIndexerName, podGroupNamespacedName.Name)

	pods = &corev1.PodList{}

	err := cachedClient.List(
		apiCtx,
		pods,
		&client.ListOptions{FieldSelector: podsByPodGroupFieldSelector})
	if err != nil {
		return nil
	}

	return pods
}

func getNamespaceFromClient(namespaceName string, k8sClient client.Client) *corev1.Namespace {
	namespaceKey := types.NamespacedName{Name: namespaceName}
	namespaceObj := &corev1.Namespace{}

	err := k8sClient.Get(apiCtx, namespaceKey, namespaceObj)
	if err != nil {
		return nil
	}

	return namespaceObj
}

func getQueueFromClient(queueName string, k8sClient client.Client) *kaiv2.Queue {
	queueKey := types.NamespacedName{Name: queueName}
	queueObj := &kaiv2.Queue{}

	err := k8sClient.Get(apiCtx, queueKey, queueObj)
	if err != nil {
		return nil
	}

	return queueObj
}

func createNodePool(testName, nodePoolName, labelKey, labelVal string, nodePoolPhase v1alpha1.NodePoolPhase, k8sClient client.Client) {
	By(fmt.Sprintf("Test: %s, creating nodepool: %s", testName, nodePoolName))
	expectCreateResource(k8sClient, getNodePoolObj(nodePoolName, labelKey, labelVal))

	nodePool := &v1alpha1.NodePool{}
	EventuallyWithOffset(1, func(g Gomega) {
		nodePool = &v1alpha1.NodePool{ObjectMeta: metav1.ObjectMeta{Name: nodePoolName}}
		expectGetResource(g, k8sClient, nodePool)
	}, timeout, interval).Should(Succeed())

	nodePool.Status.Phase = nodePoolPhase
	expectUpdateResourceStatus(k8sClient, nodePool)
}

func createPodGroup(testName string, podGroup *kaiv2alpha2.PodGroup, k8sClient client.Client) {
	By(fmt.Sprintf("Test: %s, creating podgroup: %s/%s", testName, podGroup.Namespace, podGroup.Name))
	expectCreateResource(k8sClient, podGroup)
	eventuallyExpectGetResource(k8sClient, podGroup)
}

func createPod(pod *corev1.Pod, k8sClient client.Client) {
	By(fmt.Sprintf("Creating pod: %s/%s", pod.Namespace, pod.Name))
	expectCreateResource(k8sClient, pod)
	eventuallyExpectGetResource(k8sClient, pod)
}

func createNamespaceIfNeeded(namespace string, k8sClient client.Client) {
	if namespace == "" {
		namespace = RunaiNamespace
	}

	namespaceObj := getNamespaceFromClient(namespace, k8sClient)
	if namespaceObj != nil {
		return
	}

	namespaceObj = &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
			Annotations: map[string]string{
				"suite": suite,
			},
		},
	}
	if namespace != RunaiNamespace {
		namespaceObj.Labels = map[string]string{
			"runai/queue": hackGetProjNameFromNamespace(namespace),
		}
	}

	expectCreateResource(k8sClient, namespaceObj)
	eventuallyExpectGetResource(k8sClient, namespaceObj)
}

func createProjectWithNodePoolsQueues(projectName string, nodePools []TestNodePool, k8sClient client.Client) {
	createProjectWithNodePoolsQueuesInner(projectName, nodePools, false, k8sClient)
}

func createProjectWithNodePoolsQueuesInner(projectName string, nodePools []TestNodePool,
	useSuffix bool, k8sClient client.Client) {
	for _, nodePool := range nodePools {
		queueName := projectName
		if nodePool.Name != defaultNodePool.Name {
			queueName = fmt.Sprintf("%s-%s", projectName, nodePool.Name)
		}

		queueNameToCreate := queueName
		if useSuffix {
			queueNameToCreate = fmt.Sprintf("%s-%s", queueName, rand.String(3))
		}

		queueObj := getQueueObj(queueNameToCreate, projectName, nodePool.Name)
		expectCreateResource(k8sClient, queueObj)
		eventuallyExpectGetResource(k8sClient, queueObj)
	}
}

func deleteNodePool(testName, nodePoolName string, k8sClient client.Client) {
	By(fmt.Sprintf("Test: <%s>, Deleting NodePool <%s>", testName, nodePoolName))

	nodePool := &v1alpha1.NodePool{}
	if err := k8sClient.Get(apiCtx, types.NamespacedName{Name: nodePoolName}, nodePool); err == nil {
		deleteAndPollUntilDeleted(k8sClient, nodePool)
	}
}

func deleteNamespace(namespace string, k8sClient client.Client) {
	By(fmt.Sprintf("Deleting Namespace <%s>", namespace))

	namespaceObj := getNamespaceFromClient(namespace, k8sClient)
	if namespaceObj != nil {
		deleteAndPollUntilDeleted(k8sClient, namespaceObj)
	}
}

func deleteProject(project string, k8sClient client.Client) {
	By(fmt.Sprintf("Deleting project <%s>", project))

	queues, err := getExistingQueuesByLabels(project, "", k8sClient)
	if err == nil {
		for _, queue := range queues {
			deleteQueue(queue.Name, k8sClient)
		}
	}
}

func deleteQueue(queueName string, k8sClient client.Client) {
	By(fmt.Sprintf("Deleting queue <%s>", queueName))

	queueObj := getQueueFromClient(queueName, k8sClient)
	if queueObj != nil {
		deleteAndPollUntilDeleted(k8sClient, queueObj)
	}
}

func deletePodGroup(podGroupNamespacedName types.NamespacedName, k8sClient client.Client) {
	By(fmt.Sprintf("Deleting PodGroup <%v>", podGroupNamespacedName))

	podGroup := getPodGroupFromClient(podGroupNamespacedName, k8sClient)
	if podGroup != nil {
		deleteAndPollUntilDeleted(k8sClient, podGroup)
	}
}

func deletePod(podNamespacedName types.NamespacedName, k8sClient client.Client) {
	By(fmt.Sprintf("Deleting Pod <%v>", podNamespacedName))

	pod := getPodFromClient(podNamespacedName, k8sClient)
	if pod != nil {
		// Force GracePeriodSeconds=0 so the pod is removed immediately rather
		// than lingering in Terminating between specs.
		deleteImmediatelyAndPollUntilDeleted(k8sClient, pod)
	}
}

func expectUpdateResourceStatus(k8sClient client.Client, resource client.Object) {
	By(fmt.Sprintf("updating status of resource %s/%s...", resource.GetNamespace(), resource.GetName()))

	ExpectWithOffset(1, k8sClient.Status().Update(apiCtx, resource)).To(Succeed())

	By(fmt.Sprintf("succesfully updated status of resource %s/%s", resource.GetNamespace(), resource.GetName()))
}

func getExistingQueuesByLabels(projectName, nodepoolName string, k8sClient client.Client) ([]kaiv2.Queue, error) {
	requirements, err := createRequirementsForQueuesList(projectName, nodepoolName)
	if err != nil {
		return nil, err
	}

	labelSelector := labels.NewSelector()
	labelSelector = labelSelector.Add(requirements...)

	queues, err := listQueuesWithLabelSelectors(k8sClient, &client.ListOptions{LabelSelector: labelSelector})
	if err != nil {
		return nil, err
	}

	return queues, err
}

func createRequirementsForQueuesList(projectName, nodepoolName string,
) ([]labels.Requirement, error) {
	result := []labels.Requirement{}

	if projectName != "" {
		projectNameRequirement, err := labels.NewRequirement(
			RunaiProjectLabel, selection.DoubleEquals, []string{projectName})
		if err != nil {
			return []labels.Requirement{}, err
		}

		result = append(result, *projectNameRequirement)
	}

	var err error

	if nodepoolName != "" {
		var nodePoolRequirement *labels.Requirement
		if nodepoolName == testDefaultNodePoolName {
			nodePoolRequirement, err = labels.NewRequirement(
				testNodePoolLabelKey, selection.DoesNotExist, []string{})
		} else {
			nodePoolRequirement, err = labels.NewRequirement(
				testNodePoolLabelKey, selection.DoubleEquals, []string{nodepoolName})
		}

		if err != nil {
			return []labels.Requirement{}, err
		}

		result = append(result, *nodePoolRequirement)
	}

	return result, nil
}

func listQueuesWithLabelSelectors(k8sClient client.Client,
	listOption client.ListOption) ([]kaiv2.Queue, error) {
	queues := &kaiv2.QueueList{}

	err := k8sClient.List(context.Background(), queues, listOption)
	if err != nil {
		return []kaiv2.Queue{}, err
	}

	return queues.Items, nil
}

func expectGetResource(g Gomega, k8sClient client.Client, resource client.Object) {
	By(fmt.Sprintf("getting resource %s/%s...", resource.GetNamespace(), resource.GetName()))

	g.ExpectWithOffset(1, k8sClient.Get(apiCtx, types.NamespacedName{
		Name:      resource.GetName(),
		Namespace: resource.GetNamespace(),
	}, resource)).To(Succeed())

	By(fmt.Sprintf("succesfully got resource %s/%s", resource.GetNamespace(), resource.GetName()))
}

func eventuallyExpectGetResource(k8sClient client.Client, resource client.Object) {
	EventuallyWithOffset(1, func(g Gomega) {
		expectGetResource(g, k8sClient, resource)
	}, timeout*10, interval).Should(Succeed())
}
