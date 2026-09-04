/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package utils

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"

	kaiv2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2"
	"github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/pod-group-assigner/pkg/config"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GetNodePoolFields returns the phase and the node-affinity label of the named node pool as
// plain strings.
func GetNodePoolFields(
	ctx context.Context, client client.Client, nodePoolName string,
) (phase string, labelKey string, labelValue string, err error) {
	key := types.NamespacedName{Name: nodePoolName}

	nodePool := &v1alpha1.NodePool{}
	if err := client.Get(ctx, key, nodePool); err != nil {
		log.Ctx(ctx).Error().Msgf("Failed getting node pool <%s> object, err: <%s>", nodePoolName, err.Error())
		return "", "", "", err
	}
	return string(nodePool.Status.Phase), nodePool.Spec.LabelKey, nodePool.Spec.LabelValue, nil
}

func GetNodePoolPhase(ctx context.Context, client client.Client, nodePoolName string) (string, error) {
	phase, _, _, err := GetNodePoolFields(ctx, client, nodePoolName)
	if err != nil {
		return "", err
	}

	return phase, nil
}

func IsNodePoolAvailableForScheduling(ctx context.Context, client client.Client, nodePoolName string) (bool, string, error) {
	phase, err := GetNodePoolPhase(ctx, client, nodePoolName)
	if err != nil {
		return false, "", err
	}

	isAvailable := phase == string(v1alpha1.NodePoolReady) || phase == string(v1alpha1.NodePoolEmpty)

	return isAvailable, phase, nil
}

func GetQueueNameOfNodePool(ctx context.Context, client client.Client, projectName, nodePoolName string) (string, error) {
	queue, err := getExistingQueueByLabels(ctx, client, projectName, nodePoolName)
	if err != nil {
		return "", err
	}

	return queue.Name, nil
}

func getExistingQueueByLabels(ctx context.Context, k8sClient client.Client,
	projectName, nodePoolName string) (*kaiv2.Queue, error) {
	requirements, err := createRequirementsForQueuesList(projectName, nodePoolName)
	if err != nil {
		log.Ctx(ctx).Error().Msgf(
			"Failed creating requirements for queue list for project %s and nodepool %s, error: %s",
			projectName, nodePoolName, err.Error())

		return nil, err
	}

	labelSelector := labels.NewSelector()
	labelSelector = labelSelector.Add(requirements...)

	queues, err := listQueuesWithLabelSelectors(ctx, k8sClient, &client.ListOptions{LabelSelector: labelSelector})
	if err != nil {
		log.Ctx(ctx).Error().Msgf(
			"Failed listing queues with selector for project %s and nodepool %s, error: %s",
			projectName, nodePoolName, err.Error())

		return nil, err
	}

	if len(queues) == 0 {
		notFoundQueueName := fmt.Sprintf("%s/%s", projectName, nodePoolName)

		return nil, errors.NewNotFound(kaiv2.Resource("queue"),
			notFoundQueueName)
	}

	return &queues[0], nil
}

func createRequirementsForQueuesList(projectName, nodepoolName string,
) ([]labels.Requirement, error) {
	if projectName == "" || nodepoolName == "" {
		return []labels.Requirement{},
			fmt.Errorf(
				"wrong params for createRequirementsForQueuesList - projectName %s, nodepoolName %s",
				projectName, nodepoolName)
	}

	result := []labels.Requirement{}

	projectNameRequirement, err := labels.NewRequirement(
		config.Config().ProjectLabelKey, selection.DoubleEquals, []string{projectName})
	if err != nil {
		return []labels.Requirement{}, err
	}

	result = append(result, *projectNameRequirement)

	var nodePoolRequirement *labels.Requirement
	if nodepoolName == config.Config().DefaultNodepoolName {
		nodePoolRequirement, err = labels.NewRequirement(
			config.Config().NodePoolLabelKey, selection.DoesNotExist, []string{})
	} else {
		nodePoolRequirement, err = labels.NewRequirement(
			config.Config().NodePoolLabelKey, selection.DoubleEquals, []string{nodepoolName})
	}

	if err != nil {
		return []labels.Requirement{}, err
	}

	result = append(result, *nodePoolRequirement)

	return result, nil
}

func listQueuesWithLabelSelectors(ctx context.Context, k8sClient client.Client,
	listOption client.ListOption) ([]kaiv2.Queue, error) {
	queues := &kaiv2.QueueList{}

	err := k8sClient.List(ctx, queues, listOption)
	if err != nil {
		return []kaiv2.Queue{}, err
	}

	return queues.Items, nil
}
