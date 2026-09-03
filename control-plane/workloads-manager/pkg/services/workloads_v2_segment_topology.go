/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"fmt"
	"math"
	"strings"

	clusterApi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/topology"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/pkg/clients/clusters"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	orgUnitApi "github.com/dsx-ai-factory/run-ai/runai-clients/org-unit-service-client/generated"
	clusterconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/checker"
	commonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	pyTorchJobVersion        = "v1"
	pyTorchReplicaTypeWorker = "Worker"
	pyTorchReplicaTypeMaster = "Master"

	unexistingNodePoolLabelValue = "runai-unexisting-node-pool"
)

var pyTorchWorkerTemplateAnnotationsPath = pyTorchReplicaTemplatePath(pyTorchReplicaTypeWorker, "metadata", "annotations")

func pyTorchReplicaSpecPath(replicaType string, fields ...string) []string {
	path := []string{"spec", "pytorchReplicaSpecs", replicaType}
	return append(path, fields...)
}

func pyTorchReplicaTemplatePath(replicaType string, fields ...string) []string {
	path := pyTorchReplicaSpecPath(replicaType, "template")
	return append(path, fields...)
}

type pyTorchSegmentAnnotationTarget struct {
	annotations map[string]string
}

func (k *K8sResourceUtil) pyTorchSegmentAnnotationTarget(resource map[string]interface{}) (*pyTorchSegmentAnnotationTarget, error) {
	gvk, err := k.GetGVK(resource)
	if err != nil {
		return nil, err
	}
	if convutils.Deref(gvk.Group, "") != clusterconstants.GroupKubeflow ||
		gvk.Version != pyTorchJobVersion ||
		gvk.Kind != clusterconstants.KindPyTorchJob {
		return nil, nil
	}

	annotations, found, err := unstructured.NestedStringMap(resource, pyTorchWorkerTemplateAnnotationsPath...)
	if err != nil {
		return nil, fmt.Errorf("PyTorch Worker template annotations must be strings: %w", err)
	}
	if !found {
		return nil, nil
	}

	if !hasSegmentTopologyIntent(annotations) {
		return nil, nil
	}
	if err := validatePyTorchSegmentConfiguration(resource, annotations); err != nil {
		return nil, err
	}

	return &pyTorchSegmentAnnotationTarget{annotations: annotations}, nil
}

func hasSegmentTopologyIntent(annotations map[string]string) bool {
	_, segmentSizeSet := annotations[topology.KAISegmentSizeAnnotationKey]
	return segmentSizeSet ||
		annotations[topology.KAISegmentTopologyPreferredAnnotationKey] != "" ||
		annotations[topology.KAISegmentTopologyRequiredAnnotationKey] != ""
}

func validatePyTorchSegmentConfiguration(
	resource map[string]interface{},
	annotations map[string]string,
) error {
	segmentSizeValue, segmentSizeSet := annotations[topology.KAISegmentSizeAnnotationKey]
	if !segmentSizeSet {
		return fmt.Errorf(
			"%s must be provided when segment placement is configured",
			topology.KAISegmentSizeAnnotationKey,
		)
	}

	segmentSize, err := topology.ParseSegmentSize(segmentSizeValue)
	if err != nil {
		return fmt.Errorf("%s %s", topology.KAISegmentSizeAnnotationKey, err)
	}
	workerReplicas, err := pyTorchWorkerReplicas(resource)
	if err != nil {
		return err
	}
	if _, err := topology.SegmentCount(workerReplicas, segmentSize); err != nil {
		return err
	}
	return nil
}

func pyTorchWorkerReplicas(resource map[string]interface{}) (int64, error) {
	replicas, found, err := unstructured.NestedNumberAsFloat64(
		resource,
		pyTorchReplicaSpecPath(pyTorchReplicaTypeWorker, "replicas")...,
	)
	if err != nil {
		return 0, fmt.Errorf("PyTorch Worker replicas must be a non-negative integer: %w", err)
	}
	if !found {
		return 1, nil
	}
	if math.IsNaN(replicas) || math.IsInf(replicas, 0) ||
		replicas < 0 || replicas > math.MaxInt32 || math.Trunc(replicas) != replicas {
		return 0, fmt.Errorf("PyTorch Worker replicas must be a non-negative integer")
	}
	return int64(replicas), nil
}

func (t *pyTorchSegmentAnnotationTarget) needsPreferredPlacement() bool {
	return t.annotations[topology.KAISegmentTopologyPreferredAnnotationKey] == "" &&
		t.annotations[topology.KAISegmentTopologyRequiredAnnotationKey] == ""
}

func (t *pyTorchSegmentAnnotationTarget) setPreferredPlacement(resource map[string]interface{}, level string) error {
	t.annotations[topology.KAISegmentTopologyPreferredAnnotationKey] = level
	if err := unstructured.SetNestedStringMap(resource, t.annotations, pyTorchWorkerTemplateAnnotationsPath...); err != nil {
		return fmt.Errorf("failed to update PyTorch Worker template annotations: %w", err)
	}
	return nil
}

func (k *K8sResourceUtil) applySegmentTopologyDefaults(
	ctx context.Context,
	resource map[string]interface{},
	rootAnnotations map[string]interface{},
	project *orgUnitApi.Project,
	segmentTarget *pyTorchSegmentAnnotationTarget,
	options resourceMetadataUpdateOptions,
) error {
	_, rootTopologyNameConfigured := nonEmptyStringAnnotation(rootAnnotations, topology.KAITopologyAnnotationKey)
	_, rootPreferredConfigured := nonEmptyStringAnnotation(rootAnnotations, topology.KAITopologyAnnotationPreferredKey)
	_, rootRequiredConfigured := nonEmptyStringAnnotation(rootAnnotations, topology.KAITopologyAnnotationRequiredKey)
	rootPlacementConfigured := rootPreferredConfigured || rootRequiredConfigured

	if options.trustedConvertedFromNative {
		return validateTrustedConvertedSegmentTopology(
			rootAnnotations,
			segmentTarget,
			rootTopologyNameConfigured,
			rootPlacementConfigured,
		)
	}

	rootNeedsTopologyName := !rootTopologyNameConfigured
	rootNeedsPreferredPlacement := !rootPlacementConfigured
	segmentNeedsPreferredPlacement := segmentTarget.needsPreferredPlacement()

	projectNodePoolsNames := project.GetDefaultNodePools()
	if len(projectNodePoolsNames) == 0 {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf(
			"project (%s) doesn't have default node pools set, using default node pools",
			project.Name,
		)
		projectNodePoolsNames = defaultNodePool
	}

	actualNodePools, getErr := clusters.Get().GetNodePoolsInternalByCluster(ctx, project.ClusterId)
	if getErr != nil {
		return getErr
	}
	if actualNodePools == nil || len(actualNodePools.GetNodepools()) == 0 {
		return fmt.Errorf("no node pools found for cluster %s", project.ClusterId)
	}

	nodePoolsNames, resolveErr := requestedPyTorchWorkerNodePools(
		resource,
		actualNodePools.GetNodepools(),
		projectNodePoolsNames,
	)
	if resolveErr != nil {
		return resolveErr
	}
	// Node-pool resolution can normalize Worker annotations on the effective
	// manifest, so refresh the target before writing segment defaults.
	segmentTarget, resolveErr = k.pyTorchSegmentAnnotationTarget(resource)
	if resolveErr != nil {
		return resolveErr
	}

	masterNodePools, masterPresent, resolveErr := requestedPyTorchMasterNodePools(
		resource,
		actualNodePools.GetNodepools(),
		projectNodePoolsNames,
	)
	if resolveErr != nil {
		return resolveErr
	}
	if masterPresent {
		// Root placement covers the whole PodGroup, so Worker and Master pools
		// must resolve to the same network topology.
		nodePoolsNames = mergeNodePoolNames(nodePoolsNames, masterNodePools)
	}

	found, topologyName, topologyErr := topology.CheckNodePoolsTopology(
		ctx,
		actualNodePools.GetNodepools(),
		nodePoolsNames,
	)
	if topologyErr != nil {
		return topologyErr
	}
	if !found {
		return fmt.Errorf("segment topology-aware scheduling requires the selected node pools to have a network topology")
	}

	networkTopology, topologyErr := topology.GetByName(ctx, clusters.Get(), topologyName)
	if topologyErr != nil {
		return topologyErr
	}
	lowestTopologyLevel := networkTopology.Levels[len(networkTopology.Levels)-1]

	if requestedTopology, ok := nonEmptyStringAnnotation(rootAnnotations, topology.KAITopologyAnnotationKey); ok && requestedTopology != topologyName {
		return fmt.Errorf(
			"%s %q does not match the topology %q assigned to the selected node pools",
			topology.KAITopologyAnnotationKey,
			requestedTopology,
			topologyName,
		)
	}
	if requestedTopology := segmentTarget.annotations[topology.KAITopologyAnnotationKey]; requestedTopology != "" && requestedTopology != topologyName {
		return fmt.Errorf(
			"PyTorch Worker %s %q does not match the topology %q assigned to the selected node pools",
			topology.KAITopologyAnnotationKey,
			requestedTopology,
			topologyName,
		)
	}

	// A topology name selects the network but does not replace root placement.
	if rootNeedsTopologyName {
		rootAnnotations[topology.KAITopologyAnnotationKey] = topologyName
	}
	if rootNeedsPreferredPlacement {
		rootAnnotations[topology.KAITopologyAnnotationPreferredKey] = lowestTopologyLevel
	}
	if segmentNeedsPreferredPlacement {
		if err := segmentTarget.setPreferredPlacement(resource, lowestTopologyLevel); err != nil {
			return err
		}
	}

	return nil
}

func validateTrustedConvertedSegmentTopology(
	rootAnnotations map[string]interface{},
	segmentTarget *pyTorchSegmentAnnotationTarget,
	rootTopologyNameConfigured bool,
	rootPlacementConfigured bool,
) error {
	if !rootTopologyNameConfigured || !rootPlacementConfigured || segmentTarget.needsPreferredPlacement() {
		return fmt.Errorf("converted native segmented PyTorch manifest is missing resolved root or segment topology annotations")
	}

	rootTopologyName, _ := nonEmptyStringAnnotation(rootAnnotations, topology.KAITopologyAnnotationKey)
	if workerTopologyName := segmentTarget.annotations[topology.KAITopologyAnnotationKey]; workerTopologyName != "" && workerTopologyName != rootTopologyName {
		return fmt.Errorf(
			"PyTorch Worker %s %q does not match the workload topology %q",
			topology.KAITopologyAnnotationKey,
			workerTopologyName,
			rootTopologyName,
		)
	}

	return nil
}

func nonEmptyStringAnnotation(annotations map[string]interface{}, name string) (string, bool) {
	value, ok := annotations[name].(string)
	return value, ok && value != ""
}

// requestedPyTorchWorkerNodePools resolves the Worker's effective node-pool
// intent from its pod template and the project fallback.
func requestedPyTorchWorkerNodePools(
	resource map[string]interface{},
	actualNodePools []clusterApi.Nodepool,
	fallbackNodePools []string,
) ([]string, error) {
	return requestedPyTorchReplicaNodePools(resource, actualNodePools, fallbackNodePools, pyTorchReplicaTypeWorker)
}

func requestedPyTorchMasterNodePools(
	resource map[string]interface{},
	actualNodePools []clusterApi.Nodepool,
	fallbackNodePools []string,
) ([]string, bool, error) {
	masterTemplatePath := pyTorchReplicaTemplatePath(pyTorchReplicaTypeMaster)
	_, found, err := unstructured.NestedMap(resource, masterTemplatePath...)
	if err != nil {
		return nil, false, fmt.Errorf("PyTorch Master template must be an object: %w", err)
	}
	if !found {
		return nil, false, nil
	}

	requested, err := requestedPyTorchReplicaNodePools(resource, actualNodePools, fallbackNodePools, pyTorchReplicaTypeMaster)
	return requested, true, err
}

func requestedPyTorchReplicaNodePools(
	resource map[string]interface{},
	actualNodePools []clusterApi.Nodepool,
	fallbackNodePools []string,
	replicaType string,
) ([]string, error) {
	annotationsPath := pyTorchReplicaTemplatePath(replicaType, "metadata", "annotations")
	annotations, found, err := unstructured.NestedStringMap(resource, annotationsPath...)
	if err != nil {
		return nil, fmt.Errorf("PyTorch %s template annotations must be strings: %w", replicaType, err)
	}
	var annotatedNodePools []string
	if found {
		if annotation, configured := annotations[commonconstants.AnnotationNodepools]; configured {
			annotationFields := strings.Fields(annotation)
			if len(annotationFields) > 0 {
				annotatedNodePools, err = validateExplicitNodePools(annotationFields, actualNodePools, replicaType, "annotation")
			}
			if err != nil {
				return nil, err
			}
			normalized := strings.Join(annotatedNodePools, " ")
			if annotations[commonconstants.AnnotationNodepools] != normalized {
				annotations[commonconstants.AnnotationNodepools] = normalized
				if err := unstructured.SetNestedStringMap(resource, annotations, annotationsPath...); err != nil {
					return nil, fmt.Errorf("failed to normalize PyTorch %s node-pools annotation: %w", replicaType, err)
				}
			}
		}
	}

	labelsPath := pyTorchReplicaTemplatePath(replicaType, "metadata", "labels")
	labels, found, err := unstructured.NestedStringMap(resource, labelsPath...)
	if err != nil {
		return nil, fmt.Errorf("PyTorch %s template labels must be strings: %w", replicaType, err)
	}
	nodePoolFromLabel := ""
	if found {
		nodePoolFromLabel = labels[clusterconstants.NodePoolNameLabel]
		if nodePoolFromLabel == unexistingNodePoolLabelValue {
			nodePoolFromLabel = ""
		}
	}

	affinityPath := pyTorchReplicaTemplatePath(replicaType, "spec", "affinity")
	affinityResolution, err := requestedNodePoolsFromPyTorchReplicaAffinity(resource, actualNodePools, replicaType, affinityPath)
	if err != nil {
		return nil, err
	}
	affinityNodePools := affinityResolution.requestedNodePools
	effectiveNodePools := affinityNodePools
	if nodePoolFromLabel != "" {
		validatedLabel, err := validateExplicitNodePools([]string{nodePoolFromLabel}, actualNodePools, replicaType, "label")
		if err != nil {
			return nil, err
		}
		if len(affinityNodePools) > 0 && !labelSelectionOverridesAffinity(nodePoolFromLabel, affinityResolution.nodePoolSelectorKeys, actualNodePools) {
			return nil, fmt.Errorf(
				"PyTorch %s template has conflicting node-pool label %q and required node affinity %v",
				replicaType,
				nodePoolFromLabel,
				affinityNodePools,
			)
		}
		// The cluster pod mutator materializes this label as required
		// affinity and replaces expressions on the same node-label key.
		effectiveNodePools = validatedLabel
	} else if len(effectiveNodePools) == 0 {
		effectiveNodePools = append([]string(nil), fallbackNodePools...)
	}

	if len(annotatedNodePools) > 0 {
		if !sameNodePoolSet(annotatedNodePools, effectiveNodePools) {
			return nil, fmt.Errorf(
				"PyTorch %s template node-pools annotation %v conflicts with effective pod placement %v",
				replicaType,
				annotatedNodePools,
				effectiveNodePools,
			)
		}
		// The annotation controls PGA's ordered node-pool preferences.
		return annotatedNodePools, nil
	}

	return effectiveNodePools, nil
}

func sameNodePoolSet(left []string, right []string) bool {
	leftSet := make(map[string]struct{}, len(left))
	for _, nodePool := range left {
		leftSet[nodePool] = struct{}{}
	}
	rightSet := make(map[string]struct{}, len(right))
	for _, nodePool := range right {
		rightSet[nodePool] = struct{}{}
	}
	if len(leftSet) != len(rightSet) {
		return false
	}
	for nodePool := range rightSet {
		if _, found := leftSet[nodePool]; !found {
			return false
		}
	}
	return true
}

func labelSelectionOverridesAffinity(
	labelNodePool string,
	affinitySelectorKeys map[string]struct{},
	actualNodePools []clusterApi.Nodepool,
) bool {
	nodePoolLabelKeys := make(map[string]string, len(actualNodePools))
	for _, nodePool := range actualNodePools {
		nodePoolLabelKeys[nodePool.GetName()] = nodePool.GetLabelKey()
	}
	nodePoolLabelKeys[composer.DefaultNodePoolName] = clusterconstants.NodePoolNameLabel

	labelKey := nodePoolLabelKeys[labelNodePool]
	if labelKey == "" {
		return false
	}
	for affinitySelectorKey := range affinitySelectorKeys {
		if affinitySelectorKey != labelKey {
			return false
		}
	}
	return true
}

func validateExplicitNodePools(
	requested []string,
	actualNodePools []clusterApi.Nodepool,
	replicaType string,
	source string,
) ([]string, error) {
	actualByName := make(map[string]clusterApi.Nodepool, len(actualNodePools))
	for _, nodePool := range actualNodePools {
		actualByName[nodePool.GetName()] = nodePool
	}

	normalized := make([]string, 0, len(requested))
	seen := make(map[string]struct{}, len(requested))
	for _, nodePoolName := range requested {
		if _, found := seen[nodePoolName]; found {
			continue
		}
		nodePool, found := actualByName[nodePoolName]
		if !found {
			return nil, fmt.Errorf("PyTorch %s template node-pool %s references unknown node pool %q", replicaType, source, nodePoolName)
		}
		if nodePool.GetPhase() == clusterApi.NODEPOOLPHASE_DELETING {
			return nil, fmt.Errorf("PyTorch %s template node-pool %s references deleting node pool %q", replicaType, source, nodePoolName)
		}
		seen[nodePoolName] = struct{}{}
		normalized = append(normalized, nodePoolName)
	}
	return normalized, nil
}

func mergeNodePoolNames(primary []string, additional []string) []string {
	merged := append([]string(nil), primary...)
	seen := make(map[string]struct{}, len(primary)+len(additional))
	for _, nodePool := range primary {
		seen[nodePool] = struct{}{}
	}
	for _, nodePool := range additional {
		if _, found := seen[nodePool]; found {
			continue
		}
		seen[nodePool] = struct{}{}
		merged = append(merged, nodePool)
	}
	return merged
}

type affinityNodePoolResolution struct {
	requestedNodePools   []string
	nodePoolSelectorKeys map[string]struct{}
}

func requestedNodePoolsFromPyTorchReplicaAffinity(
	resource map[string]interface{},
	actualNodePools []clusterApi.Nodepool,
	replicaType string,
	affinityPath []string,
) (affinityNodePoolResolution, error) {
	affinityMap, found, err := unstructured.NestedMap(resource, affinityPath...)
	if err != nil {
		return affinityNodePoolResolution{}, fmt.Errorf("PyTorch %s template affinity must be an object: %w", replicaType, err)
	}
	if !found {
		return affinityNodePoolResolution{}, nil
	}

	affinity := &corev1.Affinity{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(affinityMap, affinity); err != nil {
		return affinityNodePoolResolution{}, fmt.Errorf("invalid PyTorch %s template affinity: %w", replicaType, err)
	}
	if affinity.NodeAffinity == nil ||
		affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution == nil {
		return affinityNodePoolResolution{}, nil
	}

	nodePoolByLabel := make(map[string]map[string]string)
	for _, nodePool := range actualNodePools {
		if nodePool.GetPhase() == clusterApi.NODEPOOLPHASE_DELETING || nodePool.GetLabelKey() == "" {
			continue
		}
		values, ok := nodePoolByLabel[nodePool.GetLabelKey()]
		if !ok {
			values = make(map[string]string)
			nodePoolByLabel[nodePool.GetLabelKey()] = values
		}
		values[nodePool.GetLabelValue()] = nodePool.GetName()
	}

	requested := make([]string, 0)
	seen := make(map[string]struct{})
	selectorKeys := make(map[string]struct{})
	for _, term := range affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms {
		foundNodePoolsInTerm := false
		for _, expression := range term.MatchExpressions {
			nodePools, matched := nodePoolsForSelectorRequirement(expression, nodePoolByLabel)
			if !matched {
				continue
			}
			selectorKeys[expression.Key] = struct{}{}
			if foundNodePoolsInTerm {
				continue
			}
			for _, nodePool := range nodePools {
				if _, exists := seen[nodePool]; exists {
					continue
				}
				seen[nodePool] = struct{}{}
				requested = append(requested, nodePool)
			}
			// Match the cluster resolver by using only the first recognized
			// node-pool expression in each term, but keep scanning so label
			// conflict validation sees later expressions retained by admission.
			foundNodePoolsInTerm = true
		}
	}

	return affinityNodePoolResolution{
		requestedNodePools:   requested,
		nodePoolSelectorKeys: selectorKeys,
	}, nil
}

func nodePoolsForSelectorRequirement(
	expression corev1.NodeSelectorRequirement,
	nodePoolByLabel map[string]map[string]string,
) ([]string, bool) {
	if expression.Operator == corev1.NodeSelectorOpDoesNotExist && expression.Key == clusterconstants.NodePoolNameLabel {
		return []string{composer.DefaultNodePoolName}, true
	}
	if expression.Operator != corev1.NodeSelectorOpIn {
		return nil, false
	}

	poolNamesByValue, found := nodePoolByLabel[expression.Key]
	if !found {
		return nil, false
	}

	requested := make([]string, 0, len(expression.Values))
	for _, value := range expression.Values {
		if nodePoolName, found := poolNamesByValue[value]; found {
			requested = append(requested, nodePoolName)
		}
	}
	return requested, len(requested) > 0
}

func (w *WorkloadsV2Impl) validateSegmentTopologyClusterSupport(
	ctx context.Context,
	clusterID string,
	manifest map[string]interface{},
) *api.SubmissionErrorV2 {
	target, err := w.k8sUtil.pyTorchSegmentAnnotationTarget(manifest)
	if err != nil {
		return NewManifestError(err.Error())
	}
	if target == nil {
		return nil
	}

	capabilitiesChecker := checker.NewClusterCapabilitiesChecker(clusterID)
	supported, capabilityErr := capabilitiesChecker.ValidateClusterCapability(ctx, checker.MinVersionOfSegmentTopology)
	if capabilityErr != nil {
		return NewClusterError(fmt.Sprintf(
			"failed to verify cluster support for segment topology-aware scheduling: %v",
			capabilityErr,
		))
	}
	if !supported {
		return NewClusterError(fmt.Sprintf(
			"segment topology-aware scheduling requires cluster version %s or later",
			checker.MinVersionForSupporting(checker.MinVersionOfSegmentTopology),
		))
	}

	return nil
}

func (w *WorkloadsV2Impl) validateSegmentTopologyForDryRun(
	ctx context.Context,
	manifest map[string]interface{},
	project *orgUnitApi.Project,
	trustedConvertedFromNative bool,
) *api.SubmissionErrorV2 {
	target, err := w.k8sUtil.pyTorchSegmentAnnotationTarget(manifest)
	if err != nil {
		return NewManifestError(err.Error())
	}
	if target == nil {
		return nil
	}

	// Metadata enrichment works on a deep copy. Running it for a segmented
	// dry-run exercises the same topology lookup and validation as a real create
	// without changing the submitted manifest or creating a workload.
	if _, err := w.k8sUtil.updateResourceMetadataWithOptions(
		ctx,
		manifest,
		"",
		project,
		resourceMetadataUpdateOptions{trustedConvertedFromNative: trustedConvertedFromNative},
	); err != nil {
		return NewManifestError(fmt.Sprintf("failed to apply segment topology defaults: %v", err))
	}
	return nil
}
