package deletion

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

// BlockersConfigMapKey is the single project-delete-blockers ConfigMap entry holding the YAML
// list of Blockers. It is the wire contract with the go-operator, which writes exactly this key.
const BlockersConfigMapKey = "blockers.yaml"

// conditionTypeSuffix / reasonSuffix are appended to a BlockerGroup's DisplayName to
// derive the project status ConditionType and Reason it reports. e.g. DisplayName
// "Workloads" -> condition type "WorkloadsReady", reason "WorkloadsDeletionHandlerFailed".
const (
	conditionTypeSuffix = "Ready"
	reasonSuffix        = "DeletionHandlerFailed"
)

// Blocker is a single project-deletion blocker as authored in the project-delete-blockers
// ConfigMap: one GVK (optionally label-filtered) tied to a DisplayName. It is the wire
// format — the ConfigMap holds a single YAML list of Blockers. Blockers that share a
// DisplayName are grouped into one BlockerGroup, and therefore report a single project
// status condition, by BlockerGroupsFromConfigMapData.
type Blocker struct {
	// DisplayName names the blocker's group (e.g. "Workloads", "Secrets") and is the single
	// source for the reported condition type and reason — see BlockerGroup.ConditionType/Reason.
	DisplayName   string                `json:"displayName"`
	Group         string                `json:"group"`
	Version       string                `json:"version"`
	Kind          string                `json:"kind"`
	LabelSelector *metav1.LabelSelector `json:"labelSelector,omitempty"`
}

// BlockerGroup is a set of queries that, together, map to a single project status
// condition. If any query matches at least one resource, the group blocks deletion
// and reports its condition (see ConditionType) as False. Groups are assembled at runtime
// from the project-delete-blockers ConfigMap (Blockers grouped by DisplayName); this package
// only defines the generic shape and engine — the runai-specific set of blockers is
// constructed by the go-operator packaging.
type BlockerGroup struct {
	// DisplayName names the group (e.g. "Workloads", "Secrets") and is the single
	// source for the reported condition type and reason — see ConditionType/Reason.
	DisplayName string
	Blockers    []Blocker
}

// ConditionType is the project status condition type this group reports
// (e.g. "Workloads" -> "WorkloadsReady").
func (g BlockerGroup) ConditionType() string {
	return g.DisplayName + conditionTypeSuffix
}

// Reason is the condition reason this group reports when it blocks deletion
// (e.g. "Workloads" -> "WorkloadsDeletionHandlerFailed").
func (g BlockerGroup) Reason() string {
	return g.DisplayName + reasonSuffix
}

// BlockerGroupsFromConfigMapData parses the project-delete-blockers ConfigMap into validated
// BlockerGroups. The ConfigMap holds a single entry (BlockersConfigMapKey) whose value is a
// YAML list of Blockers, grouped by DisplayName. Blockers that share a DisplayName collapse
// into a single group (one project condition); group order is not significant. Absent key ->
// no groups.
func BlockerGroupsFromConfigMapData(data map[string]string) ([]BlockerGroup, error) {
	raw, ok := data[BlockersConfigMapKey]
	if !ok {
		return nil, nil
	}

	var blockers []Blocker
	if err := yaml.UnmarshalStrict([]byte(raw), &blockers); err != nil {
		return nil, fmt.Errorf("parsing blockers from %q: %w", BlockersConfigMapKey, err)
	}

	return groupBlockers(blockers)
}

// groupBlockers validates each Blocker and groups them by DisplayName into BlockerGroups.
// Group order is not significant (each group maps to its own project condition by type).
func groupBlockers(blockers []Blocker) ([]BlockerGroup, error) {
	groupsMap := make(map[string][]Blocker)
	for i, blocker := range blockers {
		if err := blocker.validate(); err != nil {
			return nil, fmt.Errorf("invalid blocker %d: %w", i+1, err)
		}
		groupsMap[blocker.DisplayName] = append(groupsMap[blocker.DisplayName], blocker)
	}
	groups := make([]BlockerGroup, 0, len(groupsMap))
	for displayName, groupList := range groupsMap {
		groups = append(groups, BlockerGroup{
			DisplayName: displayName,
			Blockers:    groupList,
		})
	}
	return groups, nil
}

func (b Blocker) validate() error {
	if b.DisplayName == "" {
		return fmt.Errorf("displayName is required")
	}
	if b.Version == "" || b.Kind == "" {
		return fmt.Errorf("blocker %q: version and kind are required", b.DisplayName)
	}
	return nil
}
