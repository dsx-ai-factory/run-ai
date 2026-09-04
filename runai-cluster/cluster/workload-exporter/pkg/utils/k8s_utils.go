/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package utils

import (
	"context"
	"fmt"

	kaiv2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2"
	kaiorg "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"github.com/rs/zerolog/log"
	clustercommonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	nodepoolutils "github.com/dsx-ai-factory/run-ai/runai-cluster/common/node-pool-utils/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// getKaiProject reads the kai.resources project by name.
func getKaiProject(ctx context.Context, projectName string, readerClient client.Reader) (*kaiorg.Project, error) {
	project := &kaiorg.Project{}
	err := readerClient.Get(ctx, types.NamespacedName{Name: projectName}, project)
	if err != nil {
		log.Error().Msgf("Failed getting kai project <%s>, error: %s", projectName, err.Error())
		return nil, err
	}
	return project, nil
}

func getNamespace(ctx context.Context, namespace string, readerClient client.Reader) (*corev1.Namespace, error) {
	namespaceObj := &corev1.Namespace{}
	err := readerClient.Get(ctx, types.NamespacedName{Name: namespace}, namespaceObj)
	if err != nil {
		log.Error().Msgf("Could not get namespace %s, error: %s", namespace, err.Error())
		return nil, err
	}
	return namespaceObj, nil
}

func getQueue(ctx context.Context, nodePoolName string, projectName string, readerClient client.Reader) (
	*kaiv2.Queue, error) {
	matchNodePoolRequirement, err := nodepoolutils.GetNodePoolLabelSelector(nodePoolName, v1alpha1.NodePoolNameLabel, v1alpha1.DefaultNodePoolName)
	if err != nil {
		return nil, err
	}

	matchProjectRequirement, err := getQueueOfProjectLabelSelector(projectName)
	if err != nil {
		return nil, err
	}

	queues, err := listQueuesWithRequirements(ctx, readerClient, *matchNodePoolRequirement, *matchProjectRequirement)
	if err != nil {
		return nil, err
	}

	if len(queues.Items) == 1 {
		return &queues.Items[0], nil
	}

	if len(queues.Items) == 0 {
		return nil, errors.NewNotFound(kaiv2.Resource("queue"), fmt.Sprintf("project<%s>nodepool<%s>", projectName, nodePoolName))
	}
	return nil, fmt.Errorf("error: found %d queues with node pool <%s> and project <%s> labels",
		len(queues.Items), nodePoolName, projectName)
}

func listQueuesWithRequirements(ctx context.Context, readerClient client.Reader, requirements ...labels.Requirement) (
	*kaiv2.QueueList, error) {
	labelSelector := labels.NewSelector()
	labelSelector = labelSelector.Add(requirements...)

	return listQueuesWithSelector(ctx, labelSelector, readerClient)
}

func listQueuesWithSelector(ctx context.Context, labelSelector labels.Selector, readerClient client.Reader) (
	*kaiv2.QueueList, error) {
	queues := &kaiv2.QueueList{}
	err := readerClient.List(ctx, queues, &client.ListOptions{LabelSelector: labelSelector})
	if err != nil {
		log.Error().Msgf("Failed listing queues with label selector <%v>, err: %v",
			labelSelector.String(), err.Error())
		return nil, err
	}

	return queues, nil
}

func getQueueOfProjectLabelSelector(projectName string) (requirement *labels.Requirement, err error) {
	requirement, err = labels.NewRequirement(clustercommonconstants.RunaiProjectLabel,
		selection.DoubleEquals, []string{projectName})
	if err != nil {
		log.Error().Msgf("Failed creating requirement %v/%v/%s, err: %v",
			clustercommonconstants.RunaiProjectLabel, selection.DoubleEquals, projectName, err.Error())
		return nil, err
	}

	return requirement, nil
}
