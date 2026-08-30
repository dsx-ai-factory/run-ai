package structure

import (
	"context"
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"
	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	kartaresource "github.com/run-ai/karta/pkg/resource"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"
)

// elementIDLabelKey returns the pod label key for a element ID at the given level.
// Only levels 1 through MaxStoredComponentLevels are stored as pod labels.
//
// Example: elementIDLabelKey(1) → "run.ai/element-1-id"
func elementIDLabelKey(level int) string {
	return fmt.Sprintf("%s%d-id", ElementLabelPrefix, level)
}

// generateElementName generates the display name for an element.
// This name is deterministic: same inputs always produce the same name.
// It serves as both the human-readable display name and the key for
// matching pods to the same element across events.
//
// Examples:
//   - No instance, no replicas:        "worker"
//   - Instance only:                   "service:frontend"
//   - Instance + replica:              "service:frontend-0"
//   - No instance, single replica:     "group-0"
//   - No instance, nested replicas:    "worker-0-2" (parent replica 0, own replica 2)
func generateElementName(riComponentName string, instanceKey *string, accumulatedReplicaKeys []string) string {
	name := riComponentName
	if instanceKey != nil {
		name = fmt.Sprintf("%s:%s", name, *instanceKey)
	}
	if len(accumulatedReplicaKeys) > 0 {
		name = fmt.Sprintf("%s-%s", name, strings.Join(accumulatedReplicaKeys, "-"))
	}
	return strings.ToLower(name)
}

// findElementByName searches existing elements for one matching the
// given display name. Returns the matching StructureElement if found, nil otherwise.
// This allows reusing UUIDs for logical elements across pod events and recalculations.
func findElementByName(existingStructure map[string]*v1alpha1.StructureElement, name string) *v1alpha1.StructureElement {
	if existingStructure == nil || name == "" {
		return nil
	}
	for _, element := range existingStructure {
		if element.Name == name {
			return element
		}
	}
	return nil
}

func findElementByK8sObjectRef(existingElementsByID map[string]*v1alpha1.StructureElement, k8sObjectRef *v1alpha1.K8sObjectRef) *v1alpha1.StructureElement {
	if k8sObjectRef == nil {
		return nil
	}
	for _, element := range existingElementsByID {
		if element.K8sRef != nil && element.K8sRef.UID == k8sObjectRef.UID {
			return element
		}
	}
	return nil
}

// findPhysicalObjectFromChain finds an ObjectBacked K8s object in the owner chain
// matching the target GVK. Returns the object's K8sObjectRef - UID, Name and GVK.
//
// Returns nil if not found or target GVK is nil.
func findPhysicalObjectFromChain(targetGVK *kartav1alpha1.GroupVersionKind, ownerChain OwnerChain) *v1alpha1.K8sObjectRef {
	if targetGVK == nil {
		return nil
	}

	for _, owner := range ownerChain {
		if owner.Kind == targetGVK.Kind &&
			owner.Group == targetGVK.Group &&
			owner.Version == targetGVK.Version {
			return &v1alpha1.K8sObjectRef{
				UID:  owner.UID,
				Name: owner.Name,
				GVK: v1alpha1.GroupVersionKind{
					Group:   owner.Group,
					Version: owner.Version,
					Kind:    owner.Kind,
				},
			}
		}
	}

	return nil
}

// podMatchesComponentTypeSelector checks if pod matches component type selector.
// Uses bolt's PodQuerier for JQ-based evaluation of the selector's KeyPath.
func podMatchesComponentTypeSelector(ctx context.Context, pod *corev1.Pod, selector *kartav1alpha1.PodSelector) (bool, error) {
	if pod == nil || selector == nil || selector.ComponentTypeSelector == nil {
		return false, nil
	}

	podQuerier := kartaresource.NewPodQuerier(pod)
	matches, err := podQuerier.MatchesComponentType(ctx, selector.ComponentTypeSelector)
	if err != nil {
		return false, err
	}

	return matches, nil
}

// extractInstanceKey extracts the component instance key from a pod using
// the ComponentInstanceSelector's JQ IdPath. Uses bolt's PodQuerier for evaluation.
// Returns empty string if no selector is provided or the path doesn't resolve.
func extractInstanceKey(ctx context.Context, pod *corev1.Pod, selector *kartav1alpha1.PodSelector) *string {
	if selector == nil || selector.ComponentInstanceSelector == nil {
		return nil
	}

	podQuerier := kartaresource.NewPodQuerier(pod)
	key, found, err := podQuerier.ExtractInstanceId(ctx, selector.ComponentInstanceSelector)
	if err != nil {
		log.Warn().Msgf("Failed to extract instance key from pod <%s/%s> using IdPath %q, error: %s",
			pod.Namespace, pod.Name, selector.ComponentInstanceSelector.IdPath, err.Error())
		return nil
	}
	if !found {
		return nil
	}
	return ptr.To(key)
}

// extractReplicaKey extracts replica key from pod using ReplicaSelector.
func extractReplicaKey(ctx context.Context, pod *corev1.Pod, selector *kartav1alpha1.PodSelector) *string {
	if selector == nil || selector.ReplicaSelector == nil {
		return nil
	}

	podQuerier := kartaresource.NewPodQuerier(pod)
	key, found, err := podQuerier.ExtractReplicaKey(ctx, selector.ReplicaSelector)
	if err != nil {
		log.Warn().Msgf("Failed to extract replica key from pod <%s/%s> using ReplicaSelector %q, error: %s",
			pod.Namespace, pod.Name, selector.ReplicaSelector.KeyPath, err.Error())
		return nil
	}
	if !found {
		return nil
	}
	return ptr.To(key)
}

// topologicalSortComponents sorts ChildComponents so that parents appear before
// their children. This ensures that when GetPodElementsPath iterates over
// components, it processes a parent before any of its children — which is
// required for correctly building accumulated replica keys and hierarchy levels.
//
// Each ChildComponent has an optional OwnerRef pointing to its parent's name.
// Components without an OwnerRef are direct children of the root (no dependency).
//
// Algorithm (Kahn's): Start with components that have no parent dependency,
// then iteratively release children once their parent has been processed.
// Since each component has at most one parent (OwnerRef), this is a simple
// forest traversal rather than a general DAG sort.
func topologicalSortComponents(components []kartav1alpha1.ComponentDefinition) []kartav1alpha1.ComponentDefinition {
	if len(components) == 0 {
		return nil
	}

	// Build lookup table and parent→children adjacency map
	componentByName := make(map[string]kartav1alpha1.ComponentDefinition, len(components))
	childrenOf := make(map[string][]string) // parent name → list of child names
	hasParent := make(map[string]bool, len(components))

	for _, component := range components {
		componentByName[component.Name] = component
		if component.OwnerRef != nil {
			childrenOf[*component.OwnerRef] = append(childrenOf[*component.OwnerRef], component.Name)
			hasParent[component.Name] = true
		}
	}

	// Start with root-level components. A component is root-level if:
	// - It has no OwnerRef, OR
	// - Its OwnerRef points to a name not in ChildComponents (i.e., the RootComponent)
	ready := make([]string, 0, len(components))
	for _, component := range components {
		if !hasParent[component.Name] {
			ready = append(ready, component.Name)
		} else if component.OwnerRef != nil {
			if _, parentIsChild := componentByName[*component.OwnerRef]; !parentIsChild {
				ready = append(ready, component.Name)
			}
		}
	}

	// Process queue: for each processed component, release its children.
	// Since each component has at most one parent, a child becomes ready
	// as soon as its single parent is processed.
	sorted := make([]kartav1alpha1.ComponentDefinition, 0, len(components))
	for len(ready) > 0 {
		current := ready[0]
		ready = ready[1:]
		sorted = append(sorted, componentByName[current])

		ready = append(ready, childrenOf[current]...)
	}

	return sorted
}

// hasElementLabels returns true if the pod has any run.ai/element-* labels.
func hasElementLabels(pod *corev1.Pod) bool {
	for key := range pod.Labels {
		if strings.HasPrefix(key, ElementLabelPrefix) {
			return true
		}
	}
	return false
}

// hasResolvableElementLabels returns true if the pod has element labels AND
// all referenced element IDs exist in existingElementsByID.
// A pod with labels referencing unknown element IDs (e.g., due to a race condition
// where the existing structure was lost) returns false — it should be treated as
// unlabeled and reprocessed.
func hasResolvableElementLabels(pod *corev1.Pod, existingElementsByID map[string]*v1alpha1.StructureElement) bool {
	if !hasElementLabels(pod) {
		return false
	}
	for level := 1; level <= MaxStoredComponentLevels; level++ {
		elementID := pod.Labels[elementIDLabelKey(level)]
		if elementID == "" {
			break
		}
		if _, ok := existingElementsByID[elementID]; !ok {
			return false
		}
	}
	return true
}

// validateBuildInputs validates the inputs to BuildElementStructure:
//   - Pods with resolvable element labels (all IDs exist in existingElementsByID): no validation needed
//   - All other pods (unlabeled, or labeled with orphaned element IDs): must have a non-nil entry in ownerChainByPodName
func validateBuildInputs(
	pods []corev1.Pod,
	ownerChainByPodName map[string]*OwnerChain,
	existingElementsByID map[string]*v1alpha1.StructureElement,
) error {
	for i := range pods {
		if hasResolvableElementLabels(&pods[i], existingElementsByID) {
			continue
		}
		chain, ok := ownerChainByPodName[pods[i].Name]
		if !ok || chain == nil {
			return fmt.Errorf(
				"pod %q requires an owner chain but has no valid entry in ownerChainByPodName",
				pods[i].Name)
		}
	}
	return nil
}

// processLabeledPod processes a labeled pod by DeepCopying its elements from
// existingElementsByID. ParentID is already set on the existing elements from
// a previous Build cycle, so DeepCopy preserves it.
func processLabeledPod(
	pod *corev1.Pod,
	existingElementsByID map[string]*v1alpha1.StructureElement,
	newElementsByID map[string]*v1alpha1.StructureElement,
) {
	for level := 1; level <= MaxStoredComponentLevels; level++ {
		elementID := pod.Labels[elementIDLabelKey(level)]
		if elementID == "" {
			break
		}
		if _, alreadyProcessed := newElementsByID[elementID]; !alreadyProcessed {
			existingElement, ok := existingElementsByID[elementID]
			if !ok || existingElement == nil {
				// already validated and shouldn't happen, but just in case
				continue
			}
			newElementsByID[elementID] = existingElement.DeepCopy()
		}
	}
}

// processUnlabeledPod processes an unlabeled pod by computing its element path via
// getPodElementsPathInner, inserting elements into newElementsByID, and setting parent
// relationships from path ordering. Returns the computed PodAddress, or nil on error.
func processUnlabeledPod(
	ctx context.Context,
	pod *corev1.Pod,
	sortedComponentDefs []kartav1alpha1.ComponentDefinition,
	ownerChain OwnerChain,
	existingElementsByID map[string]*v1alpha1.StructureElement,
	newElementsByID map[string]*v1alpha1.StructureElement,
) *PodAddress {
	// Merge prior-cycle and current-cycle elements into a single lookup map,
	// so getPodElementsPathInner can find both reusable prior-cycle elements
	// AND elements created by earlier unlabeled pods in this cycle.
	allElements := map[string]*v1alpha1.StructureElement{}
	for k, v := range existingElementsByID {
		allElements[k] = v
	}
	for k, v := range newElementsByID {
		allElements[k] = v
	}

	addr := getPodElementsPathInner(ctx, pod, sortedComponentDefs, ownerChain, allElements)

	for i := range addr.ElementsPath {
		element := addr.ElementsPath[i]
		if _, alreadyProcessed := newElementsByID[element.ID]; !alreadyProcessed {
			if existing, ok := existingElementsByID[element.ID]; ok {
				newElementsByID[element.ID] = existing.DeepCopy()
			} else {
				newElementsByID[element.ID] = element.DeepCopy()
			}
		}
	}

	return addr
}

// maxChildDepth returns the maximum nesting depth of ChildComponents below
// the given parent. A leaf node (no children) returns depth unchanged;
// each level of children adds 1.
//
// Used by ValidateKartaDepth to ensure the Karta's ChildComponents tree
// doesn't exceed MaxStoredComponentLevels (3).
//
// Example tree:
//
//	root → [a, b]    a → [a1]    a1 → [a1x]    b → [b1]
//	maxChildDepth("root", ..., 0) = 3  (root→a→a1→a1x)
func maxChildDepth(parent string, parentToChildren map[string][]string, depth int) int {
	children := parentToChildren[parent]
	if len(children) == 0 {
		return depth
	}

	maxDepth := depth
	for _, child := range children {
		d := maxChildDepth(child, parentToChildren, depth+1)
		if d > maxDepth {
			maxDepth = d
		}
	}

	return maxDepth
}
