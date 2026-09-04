/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	"encoding/json"
	kaiorg "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"github.com/rs/zerolog/log"
	orgUnitApi "github.com/dsx-ai-factory/run-ai/runai-clients/org-unit-service-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/agent/pkg/common"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/resourcing"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// convertToKaiQueueConfigs builds the queue configs for an org unit. orgUnitName is the base of
// each queue name; projects pass the project name, departments pass the id-based base name (q-<id>)
func convertToKaiQueueConfigs(orgUnitName string, resources []orgUnitApi.ResourcesSync) []kaiorg.QueueConfig {
	queues := []kaiorg.QueueConfig{}
	for _, nodePoolsResources := range resources {
		nodePoolName := common.GetStringVal(nodePoolsResources.NodePool.Get().Name)
		queues = append(queues, kaiorg.QueueConfig{
			Name:      GetQueueName(orgUnitName, nodePoolName),
			Nodepool:  nodePoolName,
			Resources: convertToKaiQueueResources(ConvertNodePoolResources(nodePoolsResources)),
			Priority:  getPriorityOrDefault(nodePoolsResources.Rank),
		})
	}
	return queues
}

func convertToKaiQueueResources(resources *resourcing.QueueResources) *kaiorg.QueueResourcesConfig {
	if resources == nil {
		return nil
	}
	return &kaiorg.QueueResourcesConfig{
		GPU:    toKaiSystemResource(resources.GPU),
		CPU:    toKaiSystemResource(resources.CPU),
		Memory: toKaiSystemResource(resources.Memory),
	}
}

func toKaiSystemResource(resource resourcing.SystemResource) kaiorg.SystemResource {
	return kaiorg.SystemResource{
		Deserved:        resource.Deserved,
		OverQuotaWeight: resource.OverQuotaWeight,
		Limit:           resource.Limit,
	}
}

func copyOrgUnitIdLabel(from, to client.Object, idLabel string) {
	value, found := from.GetLabels()[idLabel]
	if !found {
		return
	}
	toLabels := to.GetLabels()
	if toLabels == nil {
		toLabels = map[string]string{}
	}
	toLabels[idLabel] = value
	to.SetLabels(toLabels)
}

func specJsonEqual(left, right any) bool {
	leftJson, err := json.Marshal(left)
	if err != nil {
		log.Error().Msgf("Failed to json marshal spec, error: %s", err.Error())
		return false
	}
	rightJson, err := json.Marshal(right)
	if err != nil {
		log.Error().Msgf("Failed to json marshal spec, error: %s", err.Error())
		return false
	}
	return reflect.DeepEqual(leftJson, rightJson)
}

func idLabelEqual(leftLabels, rightLabels map[string]string, idLabel string) bool {
	leftValue, foundLeft := leftLabels[idLabel]
	rightValue, foundRight := rightLabels[idLabel]
	return (foundLeft == foundRight) && (leftValue == rightValue)
}
