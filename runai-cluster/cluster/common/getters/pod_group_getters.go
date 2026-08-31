/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package getters

import (
	"context"

	kaiv2alpha2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2alpha2"
	commonconstants "github.com/run-ai/runai/runai-cluster/common/constants"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetPodOfWorkload(ctx context.Context, workloadId, workloadNamespace, podName string, readerClient client.Reader) (
	*corev1.Pod, error) {
	pod := &corev1.Pod{}
	err := readerClient.Get(ctx, types.NamespacedName{Name: podName, Namespace: workloadNamespace}, pod)
	if err != nil {
		return nil, err
	}
	if pod.Labels[commonconstants.ClusterTopOwnerUidLabel] != workloadId {
		return nil, errors.NewNotFound(corev1.Resource("Pod"), "pod not found for workload")
	}
	return pod, nil
}

func GetPodsOfWorkloadWithSelector(ctx context.Context, workloadId, workloadNamespace string,
	readerClient client.Reader, selector *client.ListOptions) (
	[]corev1.Pod, error) {

	var podList *corev1.PodList
	var err error
	options := getPodsOfWorkloadLabelSelector(workloadId, workloadNamespace)
	if selector != nil {
		options.ApplyToList(selector)
	}
	podList, err = listPods(ctx, readerClient, options)
	if err != nil {
		if errors.IsNotFound(err) {
			return []corev1.Pod{}, nil
		}
		return nil, err
	}
	return podList.Items, nil
}

func getPodsWithSelector(ctx context.Context, workloadNamespace string, readerClient client.Reader,
	selector *client.ListOptions) ([]corev1.Pod, error) {
	var podList *corev1.PodList
	var err error

	if selector == nil {
		return nil, errors.NewBadRequest("getPodsWithSelector: selector is required")
	}

	selector.Namespace = workloadNamespace
	podList, err = listPods(ctx, readerClient, selector)
	if err != nil {
		if client.IgnoreNotFound(err) != nil {
			return []corev1.Pod{}, nil
		}
		return nil, err
	}
	return podList.Items, nil
}

func GetPodsOfWorkload(ctx context.Context, workloadId, workloadNamespace string, readerClient client.Reader) (
	[]corev1.Pod, error) {
	return GetPodsOfWorkloadWithSelector(ctx, workloadId, workloadNamespace, readerClient, nil)
}

func GetPodsOfRevision(ctx context.Context, revisionName, workloadNamespace string, readerClient client.Reader) (
	[]corev1.Pod, error) {

	// Use labels.Set to create a label selector from a map
	labelSelector := labels.Set{
		commonconstants.KnativeRevisionLabel: revisionName,
	}

	// Create a ListOptions object with the label selector
	selector := &client.ListOptions{
		LabelSelector: labels.SelectorFromSet(labelSelector),
	}

	return getPodsWithSelector(ctx, workloadNamespace, readerClient, selector)
}

func GetPodGroupsOfPods(ctx context.Context, namespace string, podsOfWorkload []corev1.Pod,
	readerClient client.Reader) ([]*kaiv2alpha2.PodGroup, error) {
	podGroupNames := map[string]bool{}
	for i := range podsOfWorkload {
		pod := &podsOfWorkload[i]
		podGroupName := getPodGroupNameOfPod(pod)
		if podGroupName != "" {
			podGroupNames[podGroupName] = true
		}
	}

	result := []*kaiv2alpha2.PodGroup{}
	for podGroupName := range podGroupNames {
		podGroup, err := getPodGroup(ctx, podGroupName, namespace, readerClient)
		if err != nil {
			if errors.IsNotFound(err) {
				continue
			}
			return nil, err
		}
		result = append(result, podGroup)
	}

	return result, nil
}

func GetPodGroupsOfWorkload(ctx context.Context, workloadId, workloadNamespace string, readerClient client.Reader,
) ([]*kaiv2alpha2.PodGroup, error) {
	podsOfWorkload, err := GetPodsOfWorkload(ctx, workloadId, workloadNamespace, readerClient)
	if err != nil {
		return nil, err
	}
	podGroupsOfWorkload, err := GetPodGroupsOfPods(ctx, workloadNamespace, podsOfWorkload, readerClient)
	if err != nil {
		return nil, err
	}
	return podGroupsOfWorkload, nil
}

func listPods(ctx context.Context, readerClient client.Reader, opts ...client.ListOption) (
	*corev1.PodList, error) {
	pods := &corev1.PodList{}
	err := readerClient.List(ctx, pods, opts...)
	if err != nil {
		return nil, err
	}

	return pods, nil
}

func getPodGroup(ctx context.Context, name, namespace string, readerClient client.Reader) (*kaiv2alpha2.PodGroup, error) {
	podGroup := &kaiv2alpha2.PodGroup{}
	err := readerClient.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, podGroup)
	if err != nil {
		return nil, err
	}
	return podGroup, nil
}

func getPodGroupNameOfPod(pod *corev1.Pod) string {
	podGroupName, found := pod.Annotations[commonconstants.PodGroupAnnotationForPod]
	if !found {
		return ""
	}

	return podGroupName
}

func getPodsOfWorkloadLabelSelector(workloadId, workloadNamespace string) *client.ListOptions {
	labelSelector := client.MatchingLabels{
		commonconstants.TopOwnerUidLabel: workloadId,
	}
	listOpts := &client.ListOptions{Namespace: workloadNamespace}
	labelSelector.ApplyToList(listOpts)
	return listOpts
}
