/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package structure

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"
)

// GetPodElementsPath calculates the element path for a single pod.
// This is the public API that topologically sorts element definitions internally.
//
// For batch processing (multiple pods), prefer BuildElementStructure which sorts once
// and reuses the sorted definitions across all pods.
//
// existingElementByID allows reuse of UUIDs for logical elements: when a element
// with the same Name already exists, its UUID is reused instead of generating a new one.
// This ensures stable IDs across pod restarts.
func GetPodElementsPath(
	ctx context.Context,
	pod *corev1.Pod,
	karta *kartav1alpha1.Karta,
	ownerChain OwnerChain,
	existingElementsByID map[string]*v1alpha1.StructureElement,
) *PodAddress {
	sortedComponentDefs := topologicalSortComponents(karta.Spec.StructureDefinition.ChildComponents)
	return getPodElementsPathInner(ctx, pod, sortedComponentDefs, ownerChain, existingElementsByID)
}

// getPodElementsPathInner is the internal implementation that accepts pre-sorted element
// definitions, avoiding redundant topological sorts when called in a loop from
// BuildElementStructure.
//
// Pod-to-component matching strategies:
//
//	Strategy A: ComponentTypeSelector exists → use selector to match pod
//	Strategy B: Kind set, no ComponentTypeSelector → match by finding Kind in owner chain
//	Strategy C: No Kind, no selector → logical grouping, always matches
func getPodElementsPathInner(
	ctx context.Context,
	pod *corev1.Pod,
	sortedComponentDefs []kartav1alpha1.ComponentDefinition,
	ownerChain OwnerChain,
	existingElementsByID map[string]*v1alpha1.StructureElement,
) *PodAddress {
	address := &PodAddress{
		PodID:        string(pod.UID),
		ElementsPath: []v1alpha1.StructureElement{},
	}

	currentLevel := 1
	accumulatedReplicaKeys := []string{}

	var prevElementID string
	for _, componentDef := range sortedComponentDefs {
		hasTypeSelector := componentDef.PodSelector != nil && componentDef.PodSelector.ComponentTypeSelector != nil
		hasKind := componentDef.Kind != nil

		// Pre-lookup physical object for Kind-based elements
		var k8sObjectRef *v1alpha1.K8sObjectRef = nil
		if hasKind {
			k8sObjectRef = findPhysicalObjectFromChain(componentDef.Kind, ownerChain)
			existingElement := findElementByK8sObjectRef(existingElementsByID, k8sObjectRef)
			if existingElement != nil {
				address.ElementsPath = append(address.ElementsPath, *existingElement)
				prevElementID = existingElement.ID
				currentLevel++
				if existingElement.ReplicaKey != nil {
					accumulatedReplicaKeys = append(accumulatedReplicaKeys, *existingElement.ReplicaKey)
				}
				continue
			}
		}

		// Determine if pod belongs to this component
		if hasTypeSelector {
			// Strategy A: Match using ComponentTypeSelector (label/annotation match)
			matches, err := podMatchesComponentTypeSelector(ctx, pod, componentDef.PodSelector)
			if err != nil {
				log.Warn().Msgf("Failed to evaluate ComponentTypeSelector for pod <%s/%s> against component %q, error: %s",
					pod.Namespace, pod.Name, componentDef.Name, err.Error())
				continue
			}
			if !matches {
				continue
			}
		} else if hasKind {
			// Strategy B: Match by finding Kind in owner chain
			if k8sObjectRef == nil {
				continue
			}
		}
		// Strategy C (no selector, no Kind): always matches - logical grouping

		// Extract instance key from ComponentInstanceSelector (if present).
		// This splits one component DEFINITION into multiple element INSTANCES.
		instanceKey := extractInstanceKey(ctx, pod, componentDef.PodSelector)

		// Extract replica key if component definition has ReplicaSelector
		replicaKey := extractReplicaKey(ctx, pod, componentDef.PodSelector)
		if replicaKey != nil {
			accumulatedReplicaKeys = append(accumulatedReplicaKeys, *replicaKey)
		}

		elementName := generateElementName(componentDef.Name, instanceKey, accumulatedReplicaKeys)

		// Find existing element by "deterministic" name for logical elements
		if k8sObjectRef == nil {
			existingElement := findElementByName(existingElementsByID, elementName)
			if existingElement != nil {
				address.ElementsPath = append(address.ElementsPath, *existingElement)
				prevElementID = existingElement.ID
				currentLevel++
				continue
			}
		}

		// Generate element ID
		elementID := uuid.New().String()

		element := v1alpha1.StructureElement{
			ID:          elementID,
			Name:        elementName,
			K8sRef:      k8sObjectRef,
			InstanceKey: instanceKey,
			ReplicaKey:  replicaKey,
			Level:       currentLevel,
		}
		if prevElementID != "" {
			element.ParentID = ptr.To(prevElementID)
		}

		address.ElementsPath = append(address.ElementsPath, element)
		prevElementID = element.ID
		currentLevel++
	}

	return address
}

// PodsRequiringOwnerChain returns the names of pods that need an owner chain for
// BuildElementStructure to process them. This includes:
//   - Unlabeled pods (no run.ai/element-* labels)
//   - Labeled pods with orphaned element IDs (labels reference elements not in existingElementsByID,
//     e.g., due to a race condition where the existing structure was lost)
//
// Intended usage:
//  1. Call PodsRequiringOwnerChain to find which pods need owner chains
//  2. Fetch owner chains only for those pods (via K8s API)
//  3. Call BuildElementStructure with the fetched owner chains
func PodsRequiringOwnerChain(pods []corev1.Pod, existingElementsByID map[string]*v1alpha1.StructureElement) []*corev1.Pod {
	var result []*corev1.Pod
	for i := range pods {
		if !hasResolvableElementLabels(&pods[i], existingElementsByID) {
			result = append(result, &pods[i])
		}
	}
	return result
}

// BuildElementStructure builds the intermediate element structure for a workload.
// Called from the EWI top-owner workload handler, triggered by pod CREATE/DELETE events
// and workload CREATE events (debounced 5-10 seconds).
//
// Pods are classified based on whether their element labels resolve to existing elements:
//   - Resolvable pods (have run.ai/element-{level}-id labels AND all IDs exist in
//     existingElementsByID): fast path — their StructureElements are DeepCopied from existing.
//   - Unresolvable pods (unlabeled, or labeled with orphaned element IDs): recomputed via
//     getPodElementsPathInner. Every such pod MUST have an entry in ownerChainByPodName,
//     otherwise an error is returned. Use PodsRequiringOwnerChain to determine which pods
//     need owner chains before calling this function.
//
// The returned newPodAddresses contains the computed paths for unresolvable pods only, so the
// caller can label them with GeneratePodElementLabels after storing the structure.
//
// Only intermediate child elements (levels 1-3) are built. The root is synthesized
// at API query time from workload + Karta data.
func BuildElementStructure(
	ctx context.Context,
	workloadUID string,
	karta *kartav1alpha1.Karta,
	pods []corev1.Pod,
	ownerChainByPodName map[string]*OwnerChain,
	existingElementsByID map[string]*v1alpha1.StructureElement,
) (map[string]*v1alpha1.StructureElement, []*PodAddress, error) {
	if len(pods) == 0 {
		return map[string]*v1alpha1.StructureElement{}, nil, nil
	}

	// Fast path: if all pods are already labeled and reference exactly the same
	// set of elements as existingElementsByID, the structure is unchanged.
	// Return existingElementsByID directly — no DeepCopy, no relinking, zero allocations.
	if IsStructureUnchanged(pods, existingElementsByID) {
		log.Debug().Msgf("Structure unchanged for workload %q (%d pods, %d elements), skipping recalculation",
			workloadUID, len(pods), len(existingElementsByID))
		return existingElementsByID, nil, nil
	}

	if err := validateBuildInputs(pods, ownerChainByPodName, existingElementsByID); err != nil {
		return nil, nil, err
	}

	sortedComponentDefs := topologicalSortComponents(karta.Spec.StructureDefinition.ChildComponents)

	// Process pods and set parent relationships in a single pass.
	// - Labeled pods: DeepCopy from existingElementsByID, set ParentID from label level ordering.
	// - Unlabeled pods: compute via getPodElementsPathInner, set ParentID from element path ordering.
	newElementsByID := make(map[string]*v1alpha1.StructureElement)
	var newPodAddresses []*PodAddress

	for podIdx := range pods {
		if hasResolvableElementLabels(&pods[podIdx], existingElementsByID) {
			processLabeledPod(&pods[podIdx], existingElementsByID, newElementsByID)
		} else {
			ownerChain := ownerChainByPodName[pods[podIdx].Name]
			if ownerChain == nil {
				// already validated, but just in case
				continue
			}
			addr := processUnlabeledPod(ctx, &pods[podIdx], sortedComponentDefs, *ownerChain, existingElementsByID, newElementsByID)
			if addr != nil {
				newPodAddresses = append(newPodAddresses, addr)
			}
		}
	}

	log.Debug().Msgf("Built element structure for workload %q: %d elements, %d unlabeled pods processed, %d total pods",
		workloadUID, len(newElementsByID), len(newPodAddresses), len(pods))
	return newElementsByID, newPodAddresses, nil
}

// GeneratePodElementLabels generates labels for a pod based on its element path.
// Only generates labels for levels 1-3 (intermediate child elements).
//
// For each level, emits a single label:
//   - run.ai/element-{level}-id: The element ID (UUID for logical, K8s UID for ObjectBacked)
func GeneratePodElementLabels(elementsPath []v1alpha1.StructureElement) map[string]string {
	labels := make(map[string]string)

	for _, element := range elementsPath {
		if element.Level < 1 || element.Level > MaxStoredComponentLevels {
			continue
		}
		labels[elementIDLabelKey(element.Level)] = element.ID
	}

	return labels
}

// IsStructureUnchanged returns true if all pods are labeled and the set of
// element IDs they reference matches exactly the keys of existingElementsByID.
// When true, the structure hasn't changed — the caller can skip owner chain fetching,
// BuildElementStructure, and any K8s updates entirely.
func IsStructureUnchanged(pods []corev1.Pod, existingElementsByID map[string]*v1alpha1.StructureElement) bool {
	if len(existingElementsByID) == 0 {
		return len(pods) == 0
	}

	referencedIDs := make(map[string]struct{})
	for i := range pods {
		if !hasResolvableElementLabels(&pods[i], existingElementsByID) {
			return false
		}
		for level := 1; level <= MaxStoredComponentLevels; level++ {
			elementID := pods[i].Labels[elementIDLabelKey(level)]
			if elementID == "" {
				break
			}
			referencedIDs[elementID] = struct{}{}
		}
	}

	return len(referencedIDs) == len(existingElementsByID)
}

// ValidateKartaDepth validates that the ChildComponents tree in an Karta definition
// does not exceed MaxStoredComponentLevels (3).
func ValidateKartaDepth(karta *kartav1alpha1.Karta) error {
	if karta == nil {
		return nil
	}

	childComponents := karta.Spec.StructureDefinition.ChildComponents
	if len(childComponents) == 0 {
		return nil
	}

	rootName := karta.Spec.StructureDefinition.RootComponent.Name
	childNamesByParent := make(map[string][]string)

	for _, comp := range childComponents {
		parentName := rootName
		if comp.OwnerRef != nil {
			parentName = *comp.OwnerRef
		}
		childNamesByParent[parentName] = append(childNamesByParent[parentName], comp.Name)
	}

	depth := maxChildDepth(rootName, childNamesByParent, 0)

	if depth > MaxStoredComponentLevels {
		return fmt.Errorf(
			"Karta ChildComponents nesting depth %d exceeds maximum allowed depth of %d "+
				"(total hierarchy would be %d levels, max is %d). "+
				"The hierarchy supports: Level 0 (root) + up to %d levels of child components + pods (leaf)",
			depth, MaxStoredComponentLevels,
			depth+2, MaxComponentLevels,
			MaxStoredComponentLevels,
		)
	}

	return nil
}
