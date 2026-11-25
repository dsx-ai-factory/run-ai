/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pod_mutators

import (
	"fmt"
	"log"
	"strings"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/constants"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/admission-controller/pkg/mutators/utils"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/admission-controller/pkg/patch"

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
		return nodePoolLookupError(pod, fmt.Errorf("could not identify the Run:ai project for pod namespace %q", namespace))
	}

	project, err := utils.GetProjectResource(pm.client, projectName)
	if err != nil {
		return nodePoolLookupError(pod, fmt.Errorf("get project %s: %w", projectName, err))
	}

	nodePoolName, found := pod.Labels[patch.NodePoolNameLabel]
	if found {
		log.Printf("info: found node pool label on pod %s/%s, node pool: %s", pod.Namespace, pod.Name, nodePoolName)

		err := pm.setNodeAffinityForDefaultNodePools(pod, []string{nodePoolName})
		if err != nil {
			return nodePoolLookupError(pod, fmt.Errorf("apply labeled node pool %s: %w", nodePoolName, err))
		}

		return nil
	}

	var nodePoolsNames []string
	nodePoolsNames, found, err = unstructured.NestedStringSlice(project.Object, "spec",
		defaultNodePoolsFieldName)
	if err != nil {
		return nodePoolLookupError(pod, fmt.Errorf("read project %s default node pools: %w", projectName, err))
	}
	if !found || len(nodePoolsNames) == 0 {
		log.Printf("No default node pools defined for project %s, do nothing", projectName)
		return nil
	}

	hasNodePoolInPodsNodeAffinity, err := pm.hasNodePoolInPodsNodeAffinity(pod)
	if err != nil {
		return nodePoolLookupError(pod, fmt.Errorf("identify existing node pool requirements: %w", err))
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
		return nodePoolLookupError(pod, fmt.Errorf("apply project %s default node pools: %w", projectName, err))
	}

	return nil
}

// Explicit GPU placement must intersect the project rules. A failed pool lookup
// must not silently admit these pods with a broader selector.
func nodePoolLookupError(pod *v1.Pod, err error) error {
	log.Printf("error: %s", err)
	if patch.HasAcceleratorSelection(&pod.Spec, pod.Annotations) {
		return err
	}
	return nil
}

func (pm *PodMutator) hasNodePoolInPodsNodeAffinity(pod *v1.Pod) (bool, error) {
	hasGPUSelection := patch.HasAcceleratorSelection(&pod.Spec, pod.Annotations)
	// GPU placement and Knative's converted nodeSelector can both look exactly
	// like a pool selector. Only a named pool can override the project defaults.
	if hasGPUSelection && len(strings.Fields(pod.Annotations[constants.AnnotationNodepools])) == 0 {
		return false, nil
	}
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

	return hasNodePoolInNodeAffinity(terms, nodePoolsMapping, hasGPUSelection), nil
}

func hasNodePoolInNodeAffinity(
	terms []v1.NodeSelectorTerm,
	nodePoolsMapping map[KeyValuePair]bool, requireAllTerms bool) bool {
	hasPool := false
	for _, term := range terms {
		// Kubernetes never matches an empty term.
		if len(term.MatchExpressions) == 0 && len(term.MatchFields) == 0 {
			continue
		}
		if termHasNodePool(term, nodePoolsMapping, requireAllTerms) {
			hasPool = true
			if !requireAllTerms {
				return true
			}
		} else if requireAllTerms {
			// A single OR branch without a pool can escape the defaults.
			return false
		}
	}
	return hasPool
}

func termHasNodePool(term v1.NodeSelectorTerm, nodePoolsMapping map[KeyValuePair]bool, requireAllValues bool) bool {
	for _, expression := range term.MatchExpressions {
		if isMatchExpressionOfDefaultNodePool(expression) {
			return true
		}
		if expression.Operator != v1.NodeSelectorOpIn {
			continue
		}
		if isMatchExpressionOfCustomNodePool(expression, nodePoolsMapping, requireAllValues) {
			return true
		}
	}
	return false
}

func isMatchExpressionOfCustomNodePool(
	matchExpression v1.NodeSelectorRequirement,
	nodePoolsMapping map[KeyValuePair]bool, requireAllValues bool) bool {
	for _, value := range matchExpression.Values {
		found := nodePoolsMapping[KeyValuePair{Key: matchExpression.Key, Value: value}]
		if requireAllValues && !found {
			return false
		}
		if !requireAllValues && found {
			return true
		}
	}
	return requireAllValues && len(matchExpression.Values) > 0
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

	if len(matchExpressionsToAdd) == 0 && patch.HasAcceleratorSelection(&pod.Spec, pod.Annotations) {
		return fmt.Errorf("none of the specified node pools %v are active", nodePoolsNames)
	}
	if len(matchExpressionsToAdd) > 0 {
		patch.AddPlacementRequirements(&pod.Spec, pod.Annotations, matchExpressionsToAdd)
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
