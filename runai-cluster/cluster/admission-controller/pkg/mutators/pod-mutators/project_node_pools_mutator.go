/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pod_mutators

import (
	"fmt"
	"log"
	"strings"

	"github.com/run-ai/runai/runai-cluster/cluster/admission-controller/pkg/mutators/utils"
	"github.com/run-ai/runai/runai-cluster/cluster/admission-controller/pkg/patch"

	"k8s.io/api/admission/v1beta1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const (
	defaultNodePoolsFieldName = "defaultNodePools"
)

type KeyValuePair struct {
	Key   string
	Value string
}

func (pm *PodMutator) mutateNodeAffinityForDefaultNodePools(pod *v1.Pod, request *v1beta1.AdmissionRequest) error {
	namespace := getDestinationNamespace(pod, request)
	projectName := utils.GetDesiredProjectName(pm.client, namespace, nil)
	if projectName == "" {
		log.Printf("error: Could not identify the runai project related to this pod")
		return nil
	}

	project, err := utils.GetProjectResource(pm.client, projectName)
	if err != nil {
		log.Printf("error: Could get project resources for project %s, err: %s", projectName, err.Error())
		return nil
	}

	nodePoolName, found := pod.Labels[patch.NodePoolNameLabel]
	if found {
		log.Printf("info: found node pool label on pod %s/%s, node pool: %s", pod.Namespace, pod.Name, nodePoolName)

		err := pm.setNodeAffinityForDefaultNodePools(pod, []string{nodePoolName})
		if err != nil {
			log.Printf(
				"error: failed to add nodepool %s from label to the pod's node affinity, err: %s",
				nodePoolName, err.Error())
		}

		return nil
	}

	var nodePoolsNames []string
	nodePoolsNames, found, err = unstructured.NestedStringSlice(project.Object, "spec",
		defaultNodePoolsFieldName)
	if err != nil {
		log.Printf("error: got an error trying to fetch the default node pool list from the project, err:  %s", err.Error())
		return nil
	}
	if !found || len(nodePoolsNames) == 0 {
		log.Printf("No default node pools defined for project %s, do nothing", projectName)
		return nil
	}

	hasNodePoolInPodsNodeAffinity, err := pm.hasNodePoolInPodsNodeAffinity(pod)
	if err != nil {
		log.Printf("error: failed to calculate current pod's node pools, err: %s", err.Error())
		return nil
	}
	if hasNodePoolInPodsNodeAffinity {
		log.Printf(
			"There are existing node pools in pod's node affinity, pod <%v/%v> should not be mutated - do nothing",
			pod.Namespace,
			pod.Name)
		return nil
	}

	err = pm.setNodeAffinityForDefaultNodePools(pod, nodePoolsNames)
	if err != nil {
		log.Printf(
			"error: failed to add project %s default node pools to the pod's node affinity, err: %s",
			projectName, err.Error())
	}

	return nil
}

func (pm *PodMutator) hasNodePoolInPodsNodeAffinity(pod *v1.Pod) (bool, error) {
	if pod.Spec.Affinity == nil ||
		pod.Spec.Affinity.NodeAffinity == nil ||
		pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution == nil ||
		pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms == nil ||
		len(pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms) == 0 {
		return false, nil
	}
	terms := pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms
	allNodePools, err := pm.client.GetNodePoolsList()
	if err != nil {
		return false, err
	}

	nodePoolsMapping, err := pm.getNodePoolsMapping(allNodePools)
	if err != nil {
		return false, err
	}

	return hasNodePoolInNodeAffinity(terms, nodePoolsMapping), nil
}

func hasNodePoolInNodeAffinity(
	terms []v1.NodeSelectorTerm,
	nodePoolsMapping map[KeyValuePair]bool) bool {
	for _, term := range terms {
		for _, matchExpression := range term.MatchExpressions {
			if matchExpression.Operator != v1.NodeSelectorOpIn {
				if isMatchExpressionOfDefaultNodePool(matchExpression) {
					return true
				}
				continue
			}

			if isMatchExpressionOfCustomNodePool(matchExpression, nodePoolsMapping) {
				return true
			}
		}
	}

	return false
}

func isMatchExpressionOfCustomNodePool(
	matchExpression v1.NodeSelectorRequirement,
	nodePoolsMapping map[KeyValuePair]bool) bool {
	for _, value := range matchExpression.Values {
		_, foundNodePool := nodePoolsMapping[KeyValuePair{Key: matchExpression.Key, Value: value}]
		if foundNodePool {
			return true
		}
	}
	return false
}

func isMatchExpressionOfDefaultNodePool(expression v1.NodeSelectorRequirement) bool {
	return expression.Operator == v1.NodeSelectorOpDoesNotExist && expression.Key == patch.NodePoolNameLabel
}

func (pm *PodMutator) setNodeAffinityForDefaultNodePools(pod *v1.Pod, nodePoolsNames []string) error {
	matchExpressionsToAdd := []v1.NodeSelectorRequirement{}
	for _, nodePoolName := range nodePoolsNames {
		matchExpressionToAdd := v1.NodeSelectorRequirement{}
		created, err := pm.createMatchExpressionForNodePool(nodePoolName, &matchExpressionToAdd)
		if err != nil {
			log.Printf("Failed to create a match expression for node pool %s, err: %s", nodePoolName, err.Error())
			return err
		}
		if created {
			matchExpressionsToAdd = append(matchExpressionsToAdd, matchExpressionToAdd)
		}
	}

	if len(matchExpressionsToAdd) > 0 {
		patch.AddNodeAffinityToPodSpec(&pod.Spec, matchExpressionsToAdd)
	}

	return nil
}

func (pm *PodMutator) getNodePoolsMapping(
	nodePools *unstructured.UnstructuredList) (map[KeyValuePair]bool, error) {
	mapping := map[KeyValuePair]bool{}
	for _, nodePool := range nodePools.Items {
		if nodePool.GetName() == patch.DefaultNodePoolName {
			continue
		}

		nodePoolKeyValue, err := getNodePoolKeyValuePair(nodePool.Object)
		if err != nil {
			return nil, err
		}

		mapping[nodePoolKeyValue] = true
	}

	return mapping, nil
}

func (pm *PodMutator) createMatchExpressionForNodePool(
	nodePoolName string,
	matchExpression *v1.NodeSelectorRequirement) (bool, error) {
	if nodePoolName == patch.DefaultNodePoolName {
		*matchExpression = v1.NodeSelectorRequirement{
			Key:      patch.NodePoolNameLabel,
			Operator: v1.NodeSelectorOpDoesNotExist,
		}
		return true, nil
	}

	nodePool, err := pm.client.GetNodePool(nodePoolName)
	if err != nil {
		if errors.IsNotFound(err) {
			return false, fmt.Errorf("error: node pool %s does not exist", nodePoolName)
		}

		return false, err
	}

	nodePoolObject := nodePool.Object
	phase, err := extractNestedString(nodePoolObject, "status", "phase")
	if err != nil {
		return false, err
	}
	if phase == patch.NodePoolDeleting {
		log.Printf("error: node pool %s is %s - can't submit to this node pool", nodePoolName, patch.NodePoolDeleting)
		return false, nil
	}

	keyValue, err := getNodePoolKeyValuePair(nodePoolObject)
	if err != nil {
		return false, err
	}

	*matchExpression = v1.NodeSelectorRequirement{
		Key:      keyValue.Key,
		Operator: v1.NodeSelectorOpIn,
		Values:   []string{keyValue.Value},
	}

	return true, nil
}

func getNodePoolKeyValuePair(nodePoolObject map[string]interface{}) (KeyValuePair, error) {
	keyValuePair := KeyValuePair{}

	labelKey, err := extractNestedString(nodePoolObject, "spec", "labelKey")
	if err != nil {
		return keyValuePair, err
	}
	labelValue, err := extractNestedString(nodePoolObject, "spec", "labelValue")
	if err != nil {
		return keyValuePair, err
	}

	keyValuePair.Key = labelKey
	keyValuePair.Value = labelValue

	return keyValuePair, nil
}

func extractNestedString(object map[string]interface{}, fields ...string) (string, error) {
	value, found, err := unstructured.NestedString(object, fields...)

	if err != nil {
		return "", err
	}
	if !found {
		return "", fmt.Errorf("could not find the requested field %s", strings.Join(fields, "."))
	}

	return value, nil
}
