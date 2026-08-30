package structure

import (
	commonconstants "github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	"k8s.io/apimachinery/pkg/types"
)

// ==============================================================================
// Constants: Hierarchy Depth Limits
// ==============================================================================

// MaxStoredComponentLevels is the number of intermediate component levels
// stored as rows in the workload_components table and as pod labels.
// Only levels 1 through MaxStoredComponentLevels are stored. Level 0 (root)
// is synthesized at API query time from workload + Karta data. Pods are
// referenced via podIds.
// This also serves as the maximum allowed nesting depth of ChildComponents
// in an Karta definition.
const MaxStoredComponentLevels = commonconstants.MaxStoredElementLevels

// MaxComponentLevels is the total number of hierarchy levels conceptually
// (root + stored components + pods), but only intermediate levels 1-3 are
// actually stored as component rows. Root (level 0) is synthesized, and
// pods are referenced via podIds.
const MaxComponentLevels = MaxStoredComponentLevels + 2

// ElementLabelPrefix is the prefix for element-related pod labels.
const ElementLabelPrefix = commonconstants.ElementLabelPrefix

// ==============================================================================
// Core Data Structures
// ==============================================================================

// PodAddress represents a pod's location in the component hierarchy.
type PodAddress struct {
	ElementsPath []v1alpha1.StructureElement
	PodID        string
}

// OwnerChain represents the full owner reference chain from pod to top owner.
// Built by StructureUpdater.buildOwnerChain which walks OwnerReferences via
// topowner.GetOwnerInstance.
type OwnerChain []OwnerChainEntry

// OwnerChainEntry represents a single owner in the chain (from pod up to top owner).
type OwnerChainEntry struct {
	UID     types.UID // K8s object UID
	Name    string
	Kind    string
	Group   string // e.g., "apps" or "leaderworkerset.x-k8s.io"
	Version string // e.g., "v1"
}
